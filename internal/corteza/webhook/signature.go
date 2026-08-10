package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
	"time"
)

const ReplayWindow = 5 * time.Minute

var (
	ErrMissingSignature   = errors.New("webhook signature header is missing")
	ErrInvalidSignature   = errors.New("invalid webhook signature")
	ErrMissingTimestamp   = errors.New("webhook timestamp header is missing")
	ErrInvalidTimestamp   = errors.New("invalid webhook timestamp")
	ErrRequestExpired     = errors.New("webhook request has expired (replay protection)")
	ErrEmptySecret        = errors.New("webhook secret cannot be empty")
)

// ValidateSecureSignature checks the HMAC signature and enforces replay protection
func ValidateSecureSignature(rawBody []byte, signatureHeader, timestampHeader, secretKey string, clock Clock) error {
	if secretKey == "" {
		return ErrEmptySecret 
	}

	if signatureHeader == "" {
		return ErrMissingSignature
	}

	if timestampHeader == "" {
		return ErrMissingTimestamp
	}

	ts, err := strconv.ParseInt(timestampHeader, 10, 64)
	if err != nil {
		return ErrInvalidTimestamp
	}
	
	eventTime := time.Unix(ts, 0)
	now := time.Now()
	if clock != nil {
		now = clock.Now()
	}

	if now.Sub(eventTime) > ReplayWindow || eventTime.Sub(now) > ReplayWindow {
		return ErrRequestExpired
	}

	if !strings.HasPrefix(signatureHeader, "sha256=") {
		return ErrInvalidSignature
	}
	receivedMAC := strings.TrimPrefix(signatureHeader, "sha256=")

	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(timestampHeader + "."))
	mac.Write(rawBody)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expectedMAC), []byte(receivedMAC)) {
		return ErrInvalidSignature
	}

	return nil
}
