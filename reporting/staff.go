package reporting

import (
	"sort"
	"strings"
	"time"

	"studio-console/domain"
)

type StaffRow struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	RoleDisplay string `json:"role_display"`
	Status      string `json:"status"`
	StatusLabel string `json:"status_label"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	Contact     string `json:"contact"`
	NeedsAction bool   `json:"needs_action"`
}

func BuildStaffRows(accounts []domain.StaffAccount, location *time.Location) []StaffRow {
	if location == nil {
		location = time.UTC
	}
	result := make([]StaffRow, 0, len(accounts))
	for _, account := range accounts {
		contact := account.Phone
		if contact == "" {
			contact = account.Email
		}
		result = append(result, StaffRow{
			ID: account.ID, Name: account.Name, Phone: account.Phone, Email: account.Email,
			Role: string(account.Role), RoleDisplay: account.Role.DisplayName(), Status: string(account.Status),
			StatusLabel: account.Status.DisplayName(), CreatedAt: account.CreatedAt.In(location).Format("2006-01-02 15:04"),
			UpdatedAt: account.UpdatedAt.In(location).Format("2006-01-02 15:04"), Contact: contact,
			NeedsAction: account.Status == domain.StatusInvited || contact == "",
		})
	}
	return result
}

type ContactCoverage struct {
	Total      int     `json:"total"`
	WithPhone  int     `json:"with_phone"`
	WithEmail  int     `json:"with_email"`
	WithBoth   int     `json:"with_both"`
	WithoutAny int     `json:"without_any"`
	Percent    float64 `json:"percent"`
}

func CalculateContactCoverage(accounts []domain.StaffAccount) ContactCoverage {
	result := ContactCoverage{Total: len(accounts)}
	for _, account := range accounts {
		hasPhone := strings.TrimSpace(account.Phone) != ""
		hasEmail := strings.TrimSpace(account.Email) != ""
		if hasPhone {
			result.WithPhone++
		}
		if hasEmail {
			result.WithEmail++
		}
		if hasPhone && hasEmail {
			result.WithBoth++
		}
		if !hasPhone && !hasEmail {
			result.WithoutAny++
		}
	}
	if result.Total > 0 {
		result.Percent = float64(result.Total-result.WithoutAny) * 100 / float64(result.Total)
	}
	return result
}

type TenureBand struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

func BuildTenureBands(accounts []domain.StaffAccount, asOf time.Time) []TenureBand {
	bands := []TenureBand{{Label: "0-30天"}, {Label: "31-90天"}, {Label: "91-365天"}, {Label: "365天以上"}}
	for _, account := range accounts {
		days := int(asOf.Sub(account.CreatedAt).Hours() / 24)
		switch {
		case days <= 30:
			bands[0].Count++
		case days <= 90:
			bands[1].Count++
		case days <= 365:
			bands[2].Count++
		default:
			bands[3].Count++
		}
	}
	return bands
}

func SortRows(rows []StaffRow, field string, descending bool) {
	sort.SliceStable(rows, func(i, j int) bool {
		var less bool
		switch field {
		case "name":
			less = rows[i].Name < rows[j].Name
		case "role":
			less = rows[i].Role < rows[j].Role
		case "status":
			less = rows[i].Status < rows[j].Status
		default:
			less = rows[i].CreatedAt < rows[j].CreatedAt
		}
		if descending {
			return !less
		}
		return less
	})
}
