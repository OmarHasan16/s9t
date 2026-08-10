package outbox

import (
	"encoding/json"
	"time"
)

type EventStatus string

const (
	StatusPending   EventStatus = "pending"
	StatusPublished EventStatus = "published"
	StatusFailed    EventStatus = "failed"
)

// Event represents a durable domain event stored in the local database
type Event struct {
	ID           string
	EventType    string
	TenantID     string
	AggregateID  string          // e.g., ContactID or DealID
	Payload      json.RawMessage // JSON serialized event data
	Status       EventStatus     // pending, published, failed
	CreatedAt    time.Time
	PublishedAt  *time.Time
	AttemptCount int
}
