package query

import (
	"strings"

	"studio-console/domain"
	"studio-console/store"
	"studio-console/validation"
)

type StaffFilter struct {
	Search     string
	Roles      []domain.Role
	Statuses   []domain.Status
	SortField  string
	Descending bool
	Page       int
	PageSize   int
}

type StaffPage struct {
	Items      []domain.StaffAccount `json:"items"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	TotalItems int                   `json:"total_items"`
	TotalPages int                   `json:"total_pages"`
}

type StaffReader struct {
	store *store.Store
}

func NewStaffReader(database *store.Store) *StaffReader {
	return &StaffReader{store: database}
}

func (r *StaffReader) List(filter StaffFilter) (StaffPage, error) {
	if filter.Page == 0 {
		filter.Page = 1
	}
	if filter.PageSize == 0 {
		filter.PageSize = 20
	}
	if err := validation.ValidatePage(filter.Page, filter.PageSize); err != nil {
		return StaffPage{}, err
	}
	accounts, err := r.store.ListStaff()
	if err != nil {
		return StaffPage{}, err
	}
	filtered := make([]domain.StaffAccount, 0, len(accounts))
	for _, account := range accounts {
		if !matchesSearch(account, filter.Search) {
			continue
		}
		if !matchesRole(account.Role, filter.Roles) {
			continue
		}
		if !matchesStatus(account.Status, filter.Statuses) {
			continue
		}
		filtered = append(filtered, account)
	}
	domain.SortStaff(filtered, filter.SortField, filter.Descending)
	total := len(filtered)
	start := (filter.Page - 1) * filter.PageSize
	if start > total {
		start = total
	}
	end := start + filter.PageSize
	if end > total {
		end = total
	}
	pages := 0
	if total > 0 {
		pages = (total + filter.PageSize - 1) / filter.PageSize
	}
	return StaffPage{Items: filtered[start:end], Page: filter.Page, PageSize: filter.PageSize, TotalItems: total, TotalPages: pages}, nil
}

func matchesSearch(account domain.StaffAccount, search string) bool {
	search = strings.ToLower(strings.TrimSpace(search))
	if search == "" {
		return true
	}
	values := []string{account.ID, account.Name, account.Phone, account.Email, account.Role.DisplayName(), account.Status.DisplayName()}
	for _, value := range values {
		if strings.Contains(strings.ToLower(value), search) {
			return true
		}
	}
	return false
}

func matchesRole(role domain.Role, expected []domain.Role) bool {
	if len(expected) == 0 {
		return true
	}
	for _, candidate := range expected {
		if role == candidate {
			return true
		}
	}
	return false
}

func matchesStatus(status domain.Status, expected []domain.Status) bool {
	if len(expected) == 0 {
		return true
	}
	for _, candidate := range expected {
		if status == candidate {
			return true
		}
	}
	return false
}
