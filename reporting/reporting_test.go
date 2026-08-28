package reporting

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"studio-console/domain"
)

func TestStaffReportAndImport(t *testing.T) {
	now := time.Date(2026, 8, 26, 9, 0, 0, 0, time.UTC)
	staff, _ := domain.NewStaffAccount("staff-1", "林青", "13800138000", "lin@example.com", domain.RolePhotographer, now)
	coverage := CalculateContactCoverage([]domain.StaffAccount{staff})
	if coverage.Percent != 100 || coverage.WithBoth != 1 {
		t.Fatalf("unexpected coverage: %+v", coverage)
	}
	var output bytes.Buffer
	if err := WriteStaffCSV(&output, []domain.StaffAccount{staff}, DefaultCSVOptions()); err != nil {
		t.Fatal(err)
	}
	rows, problems, err := ReadStaffCSV(strings.NewReader("姓名,角色,手机\n周岚,makeup_artist,13900139000\n"), 10)
	if err != nil || len(rows) != 1 || len(problems) != 0 {
		t.Fatalf("unexpected import: %+v %+v %v", rows, problems, err)
	}
}
