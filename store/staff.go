package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	bolt "go.etcd.io/bbolt"

	"studio-console/domain"
)

func (s *Store) CreateStaff(account domain.StaffAccount) error {
	if err := account.Validate(); err != nil {
		return fmt.Errorf("validate staff: %w", err)
	}
	key, err := encodeKey(account.ID)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(account)
	if err != nil {
		return fmt.Errorf("encode staff: %w", err)
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(staffBucket)
		if bucket.Get(key) != nil {
			return fmt.Errorf("%w: staff %s", ErrConflict, account.ID)
		}
		if duplicate := findDuplicateContact(bucket, account, ""); duplicate != "" {
			return fmt.Errorf("%w: contact already belongs to %s", ErrConflict, duplicate)
		}
		return bucket.Put(key, payload)
	})
}

func (s *Store) UpdateStaff(account domain.StaffAccount, expectedVersion uint64) error {
	if err := account.Validate(); err != nil {
		return fmt.Errorf("validate staff: %w", err)
	}
	payload, err := json.Marshal(account)
	if err != nil {
		return fmt.Errorf("encode staff: %w", err)
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(staffBucket)
		currentPayload := bucket.Get([]byte(account.ID))
		if currentPayload == nil {
			return fmt.Errorf("%w: staff %s", ErrNotFound, account.ID)
		}
		var current domain.StaffAccount
		if err := json.Unmarshal(currentPayload, &current); err != nil {
			return fmt.Errorf("decode current staff: %w", err)
		}
		if current.Version != expectedVersion {
			return fmt.Errorf("%w: expected version %d, got %d", ErrConflict, expectedVersion, current.Version)
		}
		if duplicate := findDuplicateContact(bucket, account, account.ID); duplicate != "" {
			return fmt.Errorf("%w: contact already belongs to %s", ErrConflict, duplicate)
		}
		return bucket.Put([]byte(account.ID), payload)
	})
}

func findDuplicateContact(bucket *bolt.Bucket, candidate domain.StaffAccount, ignoredID string) string {
	var duplicate string
	cursor := bucket.Cursor()
	for key, value := cursor.First(); key != nil; key, value = cursor.Next() {
		if string(key) == ignoredID {
			continue
		}
		var account domain.StaffAccount
		if json.Unmarshal(value, &account) != nil {
			continue
		}
		phoneMatches := candidate.Phone != "" && account.Phone == candidate.Phone
		emailMatches := candidate.Email != "" && strings.EqualFold(account.Email, candidate.Email)
		if phoneMatches || emailMatches {
			duplicate = account.ID
			break
		}
	}
	return duplicate
}

func (s *Store) GetStaff(id string) (*domain.StaffAccount, error) {
	key, err := encodeKey(id)
	if err != nil {
		return nil, err
	}
	var account *domain.StaffAccount
	err = s.db.View(func(tx *bolt.Tx) error {
		payload := tx.Bucket(staffBucket).Get(key)
		if payload == nil {
			return nil
		}
		var decoded domain.StaffAccount
		if err := json.Unmarshal(payload, &decoded); err != nil {
			return fmt.Errorf("decode staff: %w", err)
		}
		account = &decoded
		return nil
	})
	return account, err
}

func (s *Store) RequireStaff(id string) (domain.StaffAccount, error) {
	account, err := s.GetStaff(id)
	if err != nil {
		return domain.StaffAccount{}, err
	}
	if account == nil {
		return domain.StaffAccount{}, fmt.Errorf("%w: staff %s", ErrNotFound, id)
	}
	return *account, nil
}

func (s *Store) DeleteStaff(id string) error {
	key, err := encodeKey(id)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(staffBucket)
		if bucket.Get(key) == nil {
			return fmt.Errorf("%w: staff %s", ErrNotFound, id)
		}
		if err := bucket.Delete(key); err != nil {
			return err
		}
		contacts := tx.Bucket(contactBucket)
		keys := make([][]byte, 0)
		contacts.ForEach(func(contactKey, value []byte) error {
			var contact domain.ContactMethod
			if json.Unmarshal(value, &contact) == nil && contact.StaffID == id {
				keys = append(keys, append([]byte(nil), contactKey...))
			}
			return nil
		})
		for _, contactKey := range keys {
			if err := contacts.Delete(contactKey); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *Store) ListStaff() ([]domain.StaffAccount, error) {
	accounts := make([]domain.StaffAccount, 0)
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(staffBucket).ForEach(func(_, value []byte) error {
			var account domain.StaffAccount
			if err := json.Unmarshal(value, &account); err != nil {
				return fmt.Errorf("decode staff list item: %w", err)
			}
			accounts = append(accounts, account)
			return nil
		})
	})
	sort.Slice(accounts, func(i, j int) bool { return accounts[i].ID < accounts[j].ID })
	return accounts, err
}

func (s *Store) CountStaff() (int, error) {
	count := 0
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(staffBucket)
		if bucket == nil {
			return errors.New("staff bucket is missing")
		}
		count = bucket.Stats().KeyN
		return nil
	})
	return count, err
}

func (s *Store) PutContact(method domain.ContactMethod) error {
	if err := method.Validate(); err != nil {
		return err
	}
	payload, err := json.Marshal(method)
	if err != nil {
		return err
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		if tx.Bucket(staffBucket).Get([]byte(method.StaffID)) == nil {
			return fmt.Errorf("%w: staff %s", ErrNotFound, method.StaffID)
		}
		return tx.Bucket(contactBucket).Put([]byte(method.ID), payload)
	})
}

func (s *Store) ContactsForStaff(staffID string) ([]domain.ContactMethod, error) {
	result := make([]domain.ContactMethod, 0)
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket(contactBucket).ForEach(func(_, value []byte) error {
			var method domain.ContactMethod
			if err := json.Unmarshal(value, &method); err != nil {
				return err
			}
			if method.StaffID == staffID {
				result = append(result, method)
			}
			return nil
		})
	})
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, err
}
