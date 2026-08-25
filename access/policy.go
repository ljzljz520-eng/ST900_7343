package access

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type Permission string

const (
	PermissionStaffRead    Permission = "staff.read"
	PermissionStaffCreate  Permission = "staff.create"
	PermissionStaffUpdate  Permission = "staff.update"
	PermissionStaffDisable Permission = "staff.disable"
	PermissionAuditRead    Permission = "audit.read"
	PermissionSettingRead  Permission = "setting.read"
	PermissionSettingWrite Permission = "setting.write"
	PermissionReportRead   Permission = "report.read"
)

type AdminRole string

const (
	AdminOwner   AdminRole = "owner"
	AdminManager AdminRole = "manager"
	AdminViewer  AdminRole = "viewer"
)

type Principal struct {
	ID          string       `json:"id"`
	Role        AdminRole    `json:"role"`
	Permissions []Permission `json:"permissions"`
}

func ParseAdminRole(value string) (AdminRole, error) {
	role := AdminRole(strings.ToLower(strings.TrimSpace(value)))
	switch role {
	case AdminOwner, AdminManager, AdminViewer:
		return role, nil
	default:
		return "", fmt.Errorf("unknown administrator role: %s", value)
	}
}

func PermissionsForRole(role AdminRole) []Permission {
	switch role {
	case AdminOwner:
		return []Permission{PermissionStaffRead, PermissionStaffCreate, PermissionStaffUpdate, PermissionStaffDisable, PermissionAuditRead, PermissionSettingRead, PermissionSettingWrite, PermissionReportRead}
	case AdminManager:
		return []Permission{PermissionStaffRead, PermissionStaffCreate, PermissionStaffUpdate, PermissionStaffDisable, PermissionAuditRead, PermissionSettingRead, PermissionReportRead}
	case AdminViewer:
		return []Permission{PermissionStaffRead, PermissionSettingRead, PermissionReportRead}
	default:
		return nil
	}
}

func NewPrincipal(id string, role AdminRole) (Principal, error) {
	principal := Principal{ID: strings.TrimSpace(id), Role: role, Permissions: PermissionsForRole(role)}
	if principal.ID == "" {
		return Principal{}, errors.New("administrator id is required")
	}
	if len(principal.Permissions) == 0 {
		return Principal{}, errors.New("administrator role has no permissions")
	}
	return principal, nil
}

func (p Principal) Allows(permission Permission) bool {
	for _, granted := range p.Permissions {
		if granted == permission {
			return true
		}
	}
	return false
}

func (p Principal) Validate() error {
	if p.ID == "" {
		return errors.New("administrator id is required")
	}
	if _, err := ParseAdminRole(string(p.Role)); err != nil {
		return err
	}
	if len(p.Permissions) == 0 {
		return errors.New("administrator permissions are required")
	}
	return nil
}

func (p Principal) PermissionNames() []string {
	result := make([]string, 0, len(p.Permissions))
	for _, permission := range p.Permissions {
		result = append(result, string(permission))
	}
	sort.Strings(result)
	return result
}

type DeniedError struct {
	PrincipalID string
	Permission  Permission
}

func (e DeniedError) Error() string {
	return fmt.Sprintf("管理员 %s 没有 %s 权限", e.PrincipalID, e.Permission)
}

func Require(principal Principal, permission Permission) error {
	if err := principal.Validate(); err != nil {
		return err
	}
	if !principal.Allows(permission) {
		return DeniedError{PrincipalID: principal.ID, Permission: permission}
	}
	return nil
}
