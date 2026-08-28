package reporting

import (
	"sort"
	"strings"
	"time"

	"studio-console/domain"
)

type QualityIssue struct {
	StaffID string `json:"staff_id"`
	Field   string `json:"field"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

func InspectQuality(accounts []domain.StaffAccount, asOf time.Time) []QualityIssue {
	issues := make([]QualityIssue, 0)
	phoneOwners := make(map[string]string)
	emailOwners := make(map[string]string)
	for _, account := range accounts {
		if strings.TrimSpace(account.Name) == "" {
			issues = append(issues, QualityIssue{StaffID: account.ID, Field: "name", Level: "error", Message: "姓名缺失"})
		}
		if account.Phone == "" && account.Email == "" {
			issues = append(issues, QualityIssue{StaffID: account.ID, Field: "contact", Level: "error", Message: "联系方式缺失"})
		}
		if owner := phoneOwners[account.Phone]; account.Phone != "" && owner != "" {
			issues = append(issues, QualityIssue{StaffID: account.ID, Field: "phone", Level: "warning", Message: "手机号与 " + owner + " 重复"})
		} else if account.Phone != "" {
			phoneOwners[account.Phone] = account.ID
		}
		key := strings.ToLower(account.Email)
		if owner := emailOwners[key]; key != "" && owner != "" {
			issues = append(issues, QualityIssue{StaffID: account.ID, Field: "email", Level: "warning", Message: "邮箱与 " + owner + " 重复"})
		} else if key != "" {
			emailOwners[key] = account.ID
		}
		if account.Status == domain.StatusInvited && asOf.Sub(account.CreatedAt) > 7*24*time.Hour {
			issues = append(issues, QualityIssue{StaffID: account.ID, Field: "status", Level: "warning", Message: "邀请超过7天仍未激活"})
		}
		if account.UpdatedAt.Before(account.CreatedAt) {
			issues = append(issues, QualityIssue{StaffID: account.ID, Field: "updated_at", Level: "error", Message: "更新时间早于创建时间"})
		}
	}
	sort.Slice(issues, func(i, j int) bool {
		if issues[i].StaffID == issues[j].StaffID {
			return issues[i].Field < issues[j].Field
		}
		return issues[i].StaffID < issues[j].StaffID
	})
	return issues
}

type QualitySummary struct {
	Total    int            `json:"total"`
	Errors   int            `json:"errors"`
	Warnings int            `json:"warnings"`
	ByField  map[string]int `json:"by_field"`
}

func SummarizeQuality(issues []QualityIssue) QualitySummary {
	result := QualitySummary{Total: len(issues), ByField: make(map[string]int)}
	for _, issue := range issues {
		result.ByField[issue.Field]++
		switch issue.Level {
		case "error":
			result.Errors++
		case "warning":
			result.Warnings++
		}
	}
	return result
}

func StaffWithIssues(issues []QualityIssue) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)
	for _, issue := range issues {
		if !seen[issue.StaffID] {
			seen[issue.StaffID] = true
			result = append(result, issue.StaffID)
		}
	}
	sort.Strings(result)
	return result
}
