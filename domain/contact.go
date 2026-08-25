package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type ContactKind string

const (
	ContactPhone  ContactKind = "phone"
	ContactEmail  ContactKind = "email"
	ContactWeChat ContactKind = "wechat"
)

type ContactMethod struct {
	ID        string      `json:"id"`
	StaffID   string      `json:"staff_id"`
	Kind      ContactKind `json:"kind"`
	Value     string      `json:"value"`
	Primary   bool        `json:"primary"`
	CreatedAt time.Time   `json:"created_at"`
}

func ParseContactKind(value string) (ContactKind, error) {
	kind := ContactKind(strings.ToLower(strings.TrimSpace(value)))
	switch kind {
	case ContactPhone, ContactEmail, ContactWeChat:
		return kind, nil
	default:
		return "", fmt.Errorf("unsupported contact kind: %s", value)
	}
}

func NewContactMethod(id, staffID string, kind ContactKind, value string, primary bool, now time.Time) (ContactMethod, error) {
	method := ContactMethod{ID: id, StaffID: staffID, Kind: kind, Value: strings.TrimSpace(value), Primary: primary, CreatedAt: now.UTC()}
	if err := method.Validate(); err != nil {
		return ContactMethod{}, err
	}
	return method, nil
}

func (c ContactMethod) Validate() error {
	if strings.TrimSpace(c.ID) == "" || strings.TrimSpace(c.StaffID) == "" {
		return errors.New("contact id and staff id are required")
	}
	if _, err := ParseContactKind(string(c.Kind)); err != nil {
		return err
	}
	if strings.TrimSpace(c.Value) == "" {
		return errors.New("contact value is required")
	}
	if c.CreatedAt.IsZero() {
		return errors.New("contact creation time is required")
	}
	return nil
}

func PrimaryContact(methods []ContactMethod, kind ContactKind) (ContactMethod, bool) {
	for _, method := range methods {
		if method.Kind == kind && method.Primary {
			return method, true
		}
	}
	for _, method := range methods {
		if method.Kind == kind {
			return method, true
		}
	}
	return ContactMethod{}, false
}

func NormalizePrimary(methods []ContactMethod) []ContactMethod {
	seen := make(map[ContactKind]bool)
	result := make([]ContactMethod, len(methods))
	copy(result, methods)
	for i := range result {
		if result[i].Primary && seen[result[i].Kind] {
			result[i].Primary = false
		}
		if result[i].Primary {
			seen[result[i].Kind] = true
		}
	}
	return result
}
