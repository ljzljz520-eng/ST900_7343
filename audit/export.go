package audit

import (
	"encoding/csv"
	"io"
	"sort"
	"strings"
	"time"

	"studio-console/domain"
)

func WriteCSV(writer io.Writer, entries []domain.AuditEntry) error {
	csvWriter := csv.NewWriter(writer)
	if err := csvWriter.Write([]string{"记录编号", "操作人", "操作", "目标", "发生时间", "详情"}); err != nil {
		return err
	}
	for _, entry := range entries {
		row := []string{entry.ID, entry.ActorID, HumanAction(entry.Action), entry.TargetID, entry.OccurredAt.UTC().Format(time.RFC3339), flattenDetails(entry.Details)}
		if err := csvWriter.Write(row); err != nil {
			return err
		}
	}
	csvWriter.Flush()
	return csvWriter.Error()
}

func flattenDetails(details map[string]string) string {
	keys := make([]string, 0, len(details))
	for key := range details {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+details[key])
	}
	return strings.Join(parts, ";")
}

func GroupByDay(entries []domain.AuditEntry) map[string][]domain.AuditEntry {
	result := make(map[string][]domain.AuditEntry)
	for _, entry := range entries {
		day := entry.OccurredAt.UTC().Format("2006-01-02")
		result[day] = append(result[day], entry)
	}
	return result
}

func ActionsForTarget(entries []domain.AuditEntry, targetID string) []string {
	result := make([]string, 0)
	for _, entry := range entries {
		if entry.TargetID == targetID {
			result = append(result, entry.Action)
		}
	}
	return result
}
