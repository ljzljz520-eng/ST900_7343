package domain

import (
	"errors"
	"strings"
	"time"
)

type AuditEntry struct {
	ID         string            `json:"id"`
	ActorID    string            `json:"actor_id"`
	Action     string            `json:"action"`
	TargetID   string            `json:"target_id"`
	OccurredAt time.Time         `json:"occurred_at"`
	Details    map[string]string `json:"details"`
}

func NewAuditEntry(id, actorID, action, targetID string, now time.Time, details map[string]string) (AuditEntry, error) {
	entry := AuditEntry{ID: id, ActorID: actorID, Action: action, TargetID: targetID, OccurredAt: now.UTC(), Details: cloneDetails(details)}
	if err := entry.Validate(); err != nil {
		return AuditEntry{}, err
	}
	return entry, nil
}

func (a AuditEntry) Validate() error {
	if strings.TrimSpace(a.ID) == "" {
		return errors.New("audit id is required")
	}
	if strings.TrimSpace(a.ActorID) == "" {
		return errors.New("audit actor is required")
	}
	if strings.TrimSpace(a.Action) == "" || strings.TrimSpace(a.TargetID) == "" {
		return errors.New("audit action and target are required")
	}
	if a.OccurredAt.IsZero() {
		return errors.New("audit occurrence time is required")
	}
	return nil
}

func cloneDetails(source map[string]string) map[string]string {
	result := make(map[string]string, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

type StudioSetting struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewStudioSetting(key, value string, now time.Time) (StudioSetting, error) {
	setting := StudioSetting{Key: strings.TrimSpace(key), Value: strings.TrimSpace(value), UpdatedAt: now.UTC()}
	if setting.Key == "" {
		return StudioSetting{}, errors.New("setting key is required")
	}
	if setting.Value == "" {
		return StudioSetting{}, errors.New("setting value is required")
	}
	return setting, nil
}
