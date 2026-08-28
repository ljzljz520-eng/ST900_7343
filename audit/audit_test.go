package audit

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"studio-console/domain"
)

func TestAuditSummaryAndCSV(t *testing.T) {
	now := time.Date(2026, 8, 26, 9, 0, 0, 0, time.UTC)
	one, _ := domain.NewAuditEntry("audit-1", "admin", "staff.created", "staff-1", now, map[string]string{"role": "photographer"})
	two, _ := domain.NewAuditEntry("audit-2", "admin", "staff.updated", "staff-1", now.Add(time.Minute), nil)
	entries := []domain.AuditEntry{one, two}
	summary := Summarize(entries)
	if summary.Total != 2 || summary.UniqueStaff != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	var output bytes.Buffer
	if err := WriteCSV(&output, entries); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "创建人员") {
		t.Fatalf("unexpected CSV: %s", output.String())
	}
}
