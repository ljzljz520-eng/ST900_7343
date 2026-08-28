package studio_console_test

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"studio-console/audit"
	"studio-console/domain"
	"studio-console/query"
	"studio-console/service"
	"studio-console/store"
	"studio-console/validation"
)

func workflowManager(t *testing.T) (*store.Store, *service.Manager) {
	t.Helper()
	database, err := store.Open(filepath.Join(t.TempDir(), "studio.db"))
	if err != nil {
		t.Fatal(err)
	}
	clock := service.FixedClock{Value: time.Date(2026, 8, 26, 9, 0, 0, 0, time.UTC)}
	manager, err := service.NewManager(database, clock)
	if err != nil {
		t.Fatal(err)
	}
	return database, manager
}

func TestWorkflowCreateStaff(t *testing.T) {
	database, manager := workflowManager(t)
	defer database.Close()
	created, err := manager.CreateStaff(service.CreateStaffCommand{ActorID: "admin", Input: validation.StaffInput{Name: "林青", Phone: "13800138000", Email: "lin@example.com", Role: "photographer"}})
	if err != nil {
		t.Fatal(err)
	}
	page, err := query.NewStaffReader(database).List(query.StaffFilter{Search: "林青", Page: 1, PageSize: 20})
	if err != nil {
		t.Fatal(err)
	}
	if page.TotalItems != 1 || page.Items[0].ID != created.ID {
		t.Fatalf("created staff not visible: %+v", page)
	}
}

func TestBusinessChain02(t *testing.T) {
	database, manager := workflowManager(t)
	defer database.Close()
	_, err := manager.UpdateStaff(service.UpdateStaffCommand{ActorID: "admin", StaffID: "staff-missing", ExpectedVersion: 1, Input: validation.StaffInput{Name: "不存在", Phone: "13800138000", Role: "retoucher"}})
	var missing service.MissingRecordError
	if !errors.As(err, &missing) {
		t.Fatalf("expected explicit missing record error, got %v", err)
	}
}

func TestWorkflowDisableAndAudit(t *testing.T) {
	database, manager := workflowManager(t)
	defer database.Close()
	created, err := manager.CreateStaff(service.CreateStaffCommand{ActorID: "admin", Input: validation.StaffInput{Name: "周岚", Phone: "13900139000", Role: "makeup_artist", Status: "active"}})
	if err != nil {
		t.Fatal(err)
	}
	disabled, err := manager.DisableStaff("admin", created.ID, created.Version)
	if err != nil {
		t.Fatal(err)
	}
	if disabled.Status != domain.StatusDisabled {
		t.Fatalf("unexpected status: %s", disabled.Status)
	}
	entries, err := audit.NewReader(database).Search(audit.Filter{TargetID: created.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 || entries[0].Action != "staff.status_changed" {
		t.Fatalf("unexpected audit trail: %+v", entries)
	}
}
