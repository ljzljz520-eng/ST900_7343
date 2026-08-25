package query

import (
	"sort"

	"studio-console/domain"
	"studio-console/store"
)

type Dashboard struct {
	Total       int                   `json:"total"`
	ByRole      map[domain.Role]int   `json:"by_role"`
	ByStatus    map[domain.Status]int `json:"by_status"`
	Contactable int                   `json:"contactable"`
	NeedsReview int                   `json:"needs_review"`
}

type RoleSummary struct {
	Role     domain.Role `json:"role"`
	Display  string      `json:"display"`
	Total    int         `json:"total"`
	Active   int         `json:"active"`
	Disabled int         `json:"disabled"`
}

func BuildDashboard(accounts []domain.StaffAccount) Dashboard {
	result := Dashboard{ByRole: make(map[domain.Role]int), ByStatus: make(map[domain.Status]int)}
	for _, account := range accounts {
		result.Total++
		result.ByRole[account.Role]++
		result.ByStatus[account.Status]++
		if account.Phone != "" || account.Email != "" {
			result.Contactable++
		}
		if account.Status == domain.StatusInvited || account.Name == "" {
			result.NeedsReview++
		}
	}
	return result
}

func BuildRoleSummaries(accounts []domain.StaffAccount) []RoleSummary {
	indexed := make(map[domain.Role]*RoleSummary)
	for _, role := range domain.SupportedRoles() {
		indexed[role] = &RoleSummary{Role: role, Display: role.DisplayName()}
	}
	for _, account := range accounts {
		summary := indexed[account.Role]
		if summary == nil {
			continue
		}
		summary.Total++
		if account.Status == domain.StatusActive {
			summary.Active++
		}
		if account.Status == domain.StatusDisabled {
			summary.Disabled++
		}
	}
	result := make([]RoleSummary, 0, len(indexed))
	for _, summary := range indexed {
		result = append(result, *summary)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Role < result[j].Role })
	return result
}

func LoadDashboard(database *store.Store) (Dashboard, error) {
	accounts, err := database.ListStaff()
	if err != nil {
		return Dashboard{}, err
	}
	return BuildDashboard(accounts), nil
}
