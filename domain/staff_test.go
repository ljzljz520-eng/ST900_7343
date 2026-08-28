package domain

import (
	"errors"
	"testing"
	"time"
)

func TestStaffLifecycle(t *testing.T) {
	now := time.Date(2026, 8, 26, 9, 0, 0, 0, time.UTC)
	staff, err := NewStaffAccount("staff-1", "林青", "13800138000", "lin@example.com", RolePhotographer, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := staff.Transition(StatusActive, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if staff.Status != StatusActive || staff.Version != 2 {
		t.Fatalf("unexpected staff state: %+v", staff)
	}
	if err := staff.Transition(StatusInvited, now.Add(2*time.Hour)); !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("expected invalid transition, got %v", err)
	}
}

func TestContactPrimaryNormalization(t *testing.T) {
	now := time.Date(2026, 8, 26, 9, 0, 0, 0, time.UTC)
	one, _ := NewContactMethod("one", "staff-1", ContactPhone, "13800138000", true, now)
	two, _ := NewContactMethod("two", "staff-1", ContactPhone, "13900139000", true, now)
	result := NormalizePrimary([]ContactMethod{one, two})
	if !result[0].Primary || result[1].Primary {
		t.Fatalf("unexpected primary contacts: %+v", result)
	}
}
