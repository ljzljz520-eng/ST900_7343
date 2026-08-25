package store

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	bolt "go.etcd.io/bbolt"

	"studio-console/domain"
)

func (s *Store) AppendAudit(entry domain.AuditEntry) error {
	if err := entry.Validate(); err != nil {
		return err
	}
	payload, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("encode audit: %w", err)
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(auditBucket)
		if bucket.Get([]byte(entry.ID)) != nil {
			return fmt.Errorf("%w: audit %s", ErrConflict, entry.ID)
		}
		return bucket.Put([]byte(entry.ID), payload)
	})
}

func (s *Store) ListAudit(targetID string, since time.Time) ([]domain.AuditEntry, error) {
	entries := make([]domain.AuditEntry, 0)
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(auditBucket).ForEach(func(_, value []byte) error {
			var entry domain.AuditEntry
			if err := json.Unmarshal(value, &entry); err != nil {
				return fmt.Errorf("decode audit: %w", err)
			}
			if targetID != "" && entry.TargetID != targetID {
				return nil
			}
			if !since.IsZero() && entry.OccurredAt.Before(since) {
				return nil
			}
			entries = append(entries, entry)
			return nil
		})
	})
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].OccurredAt.Equal(entries[j].OccurredAt) {
			return entries[i].ID < entries[j].ID
		}
		return entries[i].OccurredAt.Before(entries[j].OccurredAt)
	})
	return entries, err
}

func (s *Store) PutSetting(setting domain.StudioSetting) error {
	if setting.Key == "" || setting.Value == "" || setting.UpdatedAt.IsZero() {
		return fmt.Errorf("invalid studio setting")
	}
	payload, err := json.Marshal(setting)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(settingBucket).Put([]byte(setting.Key), payload)
	})
}

func (s *Store) GetSetting(key string) (domain.StudioSetting, error) {
	var setting domain.StudioSetting
	err := s.db.View(func(tx *bolt.Tx) error {
		payload := tx.Bucket(settingBucket).Get([]byte(key))
		if payload == nil {
			return fmt.Errorf("%w: setting %s", ErrNotFound, key)
		}
		return json.Unmarshal(payload, &setting)
	})
	return setting, err
}

func (s *Store) ListSettings() ([]domain.StudioSetting, error) {
	settings := make([]domain.StudioSetting, 0)
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(settingBucket).ForEach(func(_, value []byte) error {
			var setting domain.StudioSetting
			if err := json.Unmarshal(value, &setting); err != nil {
				return err
			}
			settings = append(settings, setting)
			return nil
		})
	})
	sort.Slice(settings, func(i, j int) bool { return settings[i].Key < settings[j].Key })
	return settings, err
}

func (s *Store) Backup(destination string) error {
	if destination == "" {
		return fmt.Errorf("backup destination is required")
	}
	return s.db.View(func(tx *bolt.Tx) error {
		return tx.CopyFile(destination, 0o600)
	})
}

func (s *Store) SnapshotCounts() (map[string]int, error) {
	result := make(map[string]int)
	err := s.db.View(func(tx *bolt.Tx) error {
		for label, name := range map[string][]byte{"staff": staffBucket, "contacts": contactBucket, "audit": auditBucket, "settings": settingBucket} {
			bucket := tx.Bucket(name)
			if bucket == nil {
				return fmt.Errorf("bucket %s is missing", name)
			}
			result[label] = bucket.Stats().KeyN
		}
		return nil
	})
	return result, err
}
