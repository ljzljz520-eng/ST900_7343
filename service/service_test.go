package service

import (
	"path/filepath"
	"testing"
	"time"

	"studio-console/domain"
	"studio-console/store"
	"studio-console/validation"
)

func TestCreateActivateAndConfigure(t *testing.T) {
	database, err := store.Open(filepath.Join(t.TempDir(), "studio.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	clock := FixedClock{Value: time.Date(2026, 8, 26, 9, 0, 0, 0, time.UTC)}
	manager, _ := NewManager(database, clock)
	staff, err := manager.CreateStaff(CreateStaffCommand{ActorID: "admin", Input: validation.StaffInput{Name: "周岚", Phone: "13800138000", Role: string(domain.RoleMakeupArtist)}})
	if err != nil {
		t.Fatal(err)
	}
	staff, err = manager.ActivateStaff("admin", staff.ID, staff.Version)
	if err != nil {
		t.Fatal(err)
	}
	if staff.Status != domain.StatusActive {
		t.Fatalf("unexpected status: %s", staff.Status)
	}
	if _, err := manager.SetSetting("admin", "studio_name", "光影工作室"); err != nil {
		t.Fatal(err)
	}
}
