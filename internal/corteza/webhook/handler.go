package webhook

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"s9t.os/internal/app/http/middleware"
)

var allowedEventTypes = map[string]struct{}{
	"crm.contact.updated":           {},
	"crm.contact.lifecycle_changed": {},
	"crm.deal.stage_changed":        {},
}

var (
	ErrUnknownEventType = errors.New("unknown event type")
	ErrInvalidPayload   = errors.New("missing required payload fields")
)

type Clock interface {
	Now() time.Time
}

type EventPayload struct {
	EventID         string `json:"event_id"`
	TenantID        string `json:"tenant_id"`
	EventType       string `json:"event_type"`
	CortezaRecordID string `json:"corteza_record_id"`
}

type BusinessDispatcher interface {
	Dispatch(ctx context.Context, tx pgx.Tx, payload EventPayload) error
}

type WebhookProcessor struct {
	dbPool     *pgxpool.Pool
	secretKey  string
	clock      Clock
	dispatcher BusinessDispatcher
}

func NewWebhookProcessor(pool *pgxpool.Pool, secret string, clock Clock, dispatcher BusinessDispatcher) *WebhookProcessor {
	return &WebhookProcessor{
		dbPool:     pool,
		secretKey:  secret,
		clock:      clock,
		dispatcher: dispatcher,
	}
}

func GeneratePayloadHash(body []byte) string {
	h := sha256.New()
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func (p *WebhookProcessor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	authTenantID, ok := middleware.TenantIDFromContext(r.Context())
	if !ok || authTenantID == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized tenant context"})
		return
	}

	rawBody, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		writeJSON(w, http.StatusRequestEntityTooLarge, map[string]string{"error": "payload too large"})
		return
	}
	defer r.Body.Close()

	if err := ValidateSecureSignature(rawBody, r.Header.Get("X-Corteza-Signature"), r.Header.Get("X-Corteza-Timestamp"), p.secretKey, p.clock); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid signature or expired"})
		return
	}

	var payload EventPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	if _, ok := allowedEventTypes[payload.EventType]; !ok {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "unknown event type"})
		return
	}
	if payload.EventID == "" || payload.CortezaRecordID == "" {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "missing event_id or record_id"})
		return
	}
	if err := uuid.Validate(payload.TenantID); err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "invalid tenant uuid format"})
		return
	}
	if string(authTenantID) != payload.TenantID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "tenant mismatch"})
		return
	}

	payloadHash := GeneratePayloadHash(rawBody)
	ctx := r.Context()
	tx, err := p.dbPool.Begin(ctx)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "db error"})
		return
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `SELECT set_config('app.current_tenant', $1, true)`, string(authTenantID))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "rls context failed"})
		return
	}

	err = TryLockEvent(ctx, tx, authTenantID, payload.EventID, payload.EventType, payloadHash, p.clock.Now())
	if err != nil {
		if errors.Is(err, ErrEventAlreadyProcessed) {
			writeJSON(w, http.StatusOK, map[string]string{"status": "already_processed"})
			return
		}
		if errors.Is(err, ErrTamperedPayload) {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "payload tampered"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "idempotency check failed"})
		return
	}

	err = p.dispatcher.Dispatch(ctx, tx, payload)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"error": "business logic failed: " + err.Error()})
		return 
	}

	_, err = tx.Exec(ctx, `UPDATE platform.webhook_events SET status = 'completed', processed_at = $1 WHERE tenant_id = $2 AND event_id = $3`, p.clock.Now(), authTenantID, payload.EventID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to mark as completed"})
		return
	}

	if err := tx.Commit(ctx); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "commit failed"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "success"})
}
