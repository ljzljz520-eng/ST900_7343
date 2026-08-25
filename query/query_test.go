package query

import (
	"path/filepath"
	"testing"
	"time"

	"studio-console/domain"
	"studio-console/store"
)

func TestStaffFilterAndDashboard(t *testing.T) {
	database, _ := store.Open(filepath.Join(t.TempDir(), "studio.db"))
	defer database.Close()
	now := time.Date(2026, 8, 26, 9, 0, 0, 0, time.UTC)
	first, _ := domain.NewStaffAccount("staff-1", "林青", "13800138000", "", domain.RolePhotographer, now)
	second, _ := domain.NewStaffAccount("staff-2", "周岚", "13900139000", "", domain.RoleMakeupArtist, now)
	_ = database.CreateStaff(first)
	_ = database.CreateStaff(second)
	page, err := NewStaffReader(database).List(StaffFilter{Roles: []domain.Role{domain.RolePhotographer}, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	if page.TotalItems != 1 || page.Items[0].Name != "林青" {
		t.Fatalf("unexpected page: %+v", page)
	}
	dashboard, err := LoadDashboard(database)
	if err != nil || dashboard.Total != 2 {
		t.Fatalf("unexpected dashboard: %+v, %v", dashboard, err)
	}
}
