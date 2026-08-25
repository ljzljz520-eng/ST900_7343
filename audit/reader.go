package audit

import (
	"sort"
	"strings"
	"time"

	"studio-console/domain"
	"studio-console/store"
)

type Filter struct {
	TargetID string
	ActorID  string
	Action   string
	Since    time.Time
	Limit    int
}

type Reader struct {
	store *store.Store
}

func NewReader(database *store.Store) *Reader {
	return &Reader{store: database}
}

func (r *Reader) Search(filter Filter) ([]domain.AuditEntry, error) {
	entries, err := r.store.ListAudit(filter.TargetID, filter.Since)
	if err != nil {
		return nil, err
	}
	result := make([]domain.AuditEntry, 0, len(entries))
	for _, entry := range entries {
		if filter.ActorID != "" && entry.ActorID != filter.ActorID {
			continue
		}
		if filter.Action != "" && entry.Action != filter.Action {
			continue
		}
		result = append(result, entry)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].OccurredAt.Equal(result[j].OccurredAt) {
			return result[i].ID > result[j].ID
		}
		return result[i].OccurredAt.After(result[j].OccurredAt)
	})
	if filter.Limit > 0 && len(result) > filter.Limit {
		result = result[:filter.Limit]
	}
	return result, nil
}

type ActivitySummary struct {
	Total       int            `json:"total"`
	ByAction    map[string]int `json:"by_action"`
	ByActor     map[string]int `json:"by_actor"`
	LatestAt    time.Time      `json:"latest_at"`
	UniqueStaff int            `json:"unique_staff"`
}

func Summarize(entries []domain.AuditEntry) ActivitySummary {
	result := ActivitySummary{ByAction: make(map[string]int), ByActor: make(map[string]int)}
	targets := make(map[string]bool)
	for _, entry := range entries {
		result.Total++
		result.ByAction[entry.Action]++
		result.ByActor[entry.ActorID]++
		if strings.HasPrefix(entry.TargetID, "staff-") {
			targets[entry.TargetID] = true
		}
		if entry.OccurredAt.After(result.LatestAt) {
			result.LatestAt = entry.OccurredAt
		}
	}
	result.UniqueStaff = len(targets)
	return result
}

func HumanAction(action string) string {
	switch action {
	case "staff.created":
		return "创建人员"
	case "staff.updated":
		return "更新人员"
	case "staff.status_changed":
		return "变更账号状态"
	case "setting.updated":
		return "更新系统设置"
	default:
		return "其他操作"
	}
}
