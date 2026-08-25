package store

import (
	"path/filepath"
	"testing"
	"time"

	"studio-console/domain"
)

func TestPersistenceSurvivesReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "studio.db")
	now := time.Date(2026, 8, 26, 9, 0, 0, 0, time.UTC)
	database, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	staff, _ := domain.NewStaffAccount("staff-1", "林青", "13800138000", "lin@example.com", domain.RolePhotographer, now)
	contact, _ := domain.NewContactMethod("contact-1", staff.ID, domain.ContactPhone, staff.Phone, true, now)
	audit, _ := domain.NewAuditEntry("audit-1", "admin", "staff.created", staff.ID, now, map[string]string{"role": string(staff.Role)})
	setting, _ := domain.NewStudioSetting("studio_name", "光影工作室", now)
	if err := database.CreateStaff(staff); err != nil {
		t.Fatal(err)
	}
	if err := database.PutContact(contact); err != nil {
		t.Fatal(err)
	}
	if err := database.AppendAudit(audit); err != nil {
		t.Fatal(err)
	}
	if err := database.PutSetting(setting); err != nil {
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}
	database, err = Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	counts, err := database.SnapshotCounts()
	if err != nil {
		t.Fatal(err)
	}
	for _, entity := range []string{"staff", "contacts", "audit", "settings"} {
		if counts[entity] != 1 {
			t.Fatalf("%s was not persisted: %+v", entity, counts)
		}
	}
}
