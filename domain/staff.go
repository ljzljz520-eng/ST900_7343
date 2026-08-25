package domain

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

type Role string

const (
	RolePhotographer Role = "photographer"
	RoleRetoucher    Role = "retoucher"
	RoleMakeupArtist Role = "makeup_artist"
	RoleSupport      Role = "support"
)

type Status string

const (
	StatusInvited  Status = "invited"
	StatusActive   Status = "active"
	StatusDisabled Status = "disabled"
)

var (
	ErrInvalidRole       = errors.New("unsupported staff role")
	ErrInvalidStatus     = errors.New("invalid staff status")
	ErrInvalidTransition = errors.New("invalid status transition")
	ErrMissingName       = errors.New("staff name is required")
)

type StaffAccount struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	Role      Role      `json:"role"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Version   uint64    `json:"version"`
}

func SupportedRoles() []Role {
	return []Role{RolePhotographer, RoleRetoucher, RoleMakeupArtist, RoleSupport}
}

func ParseRole(value string) (Role, error) {
	normalized := Role(strings.ToLower(strings.TrimSpace(value)))
	for _, role := range SupportedRoles() {
		if normalized == role {
			return role, nil
		}
	}
	return "", fmt.Errorf("%w: %q", ErrInvalidRole, value)
}

func (r Role) DisplayName() string {
	switch r {
	case RolePhotographer:
		return "摄影师"
	case RoleRetoucher:
		return "修图师"
	case RoleMakeupArtist:
		return "化妆师"
	case RoleSupport:
		return "客服"
	default:
		return "未知角色"
	}
}

func ParseStatus(value string) (Status, error) {
	status := Status(strings.ToLower(strings.TrimSpace(value)))
	switch status {
	case StatusInvited, StatusActive, StatusDisabled:
		return status, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrInvalidStatus, value)
	}
}

func (s Status) DisplayName() string {
	switch s {
	case StatusInvited:
		return "待激活"
	case StatusActive:
		return "正常"
	case StatusDisabled:
		return "已停用"
	default:
		return "未知状态"
	}
}

func NewStaffAccount(id, name, phone, email string, role Role, now time.Time) (StaffAccount, error) {
	account := StaffAccount{
		ID: id, Name: strings.TrimSpace(name), Phone: phone, Email: email,
		Role: role, Status: StatusInvited, CreatedAt: now.UTC(), UpdatedAt: now.UTC(), Version: 1,
	}
	if err := account.Validate(); err != nil {
		return StaffAccount{}, err
	}
	return account, nil
}

func (s StaffAccount) Validate() error {
	if strings.TrimSpace(s.ID) == "" {
		return errors.New("staff id is required")
	}
	if strings.TrimSpace(s.Name) == "" {
		return ErrMissingName
	}
	if _, err := ParseRole(string(s.Role)); err != nil {
		return err
	}
	if _, err := ParseStatus(string(s.Status)); err != nil {
		return err
	}
	if s.CreatedAt.IsZero() || s.UpdatedAt.IsZero() {
		return errors.New("staff timestamps are required")
	}
	return nil
}

func (s StaffAccount) CanTransition(next Status) bool {
	if s.Status == next {
		return true
	}
	switch s.Status {
	case StatusInvited:
		return next == StatusActive || next == StatusDisabled
	case StatusActive:
		return next == StatusDisabled
	case StatusDisabled:
		return next == StatusActive
	default:
		return false
	}
}

func (s *StaffAccount) Transition(next Status, now time.Time) error {
	if !s.CanTransition(next) {
		return fmt.Errorf("%w: %s to %s", ErrInvalidTransition, s.Status, next)
	}
	s.Status = next
	s.UpdatedAt = now.UTC()
	s.Version++
	return nil
}

func (s *StaffAccount) UpdateProfile(name, phone, email string, role Role, now time.Time) error {
	previous := *s
	s.Name = strings.TrimSpace(name)
	s.Phone = strings.TrimSpace(phone)
	s.Email = strings.ToLower(strings.TrimSpace(email))
	s.Role = role
	s.UpdatedAt = now.UTC()
	s.Version++
	if err := s.Validate(); err != nil {
		*s = previous
		return err
	}
	return nil
}

func SortStaff(accounts []StaffAccount, field string, descending bool) {
	sort.SliceStable(accounts, func(i, j int) bool {
		left, right := accounts[i], accounts[j]
		var less bool
		switch field {
		case "name":
			less = strings.ToLower(left.Name) < strings.ToLower(right.Name)
		case "role":
			less = left.Role < right.Role
		case "status":
			less = left.Status < right.Status
		default:
			less = left.CreatedAt.Before(right.CreatedAt)
		}
		if descending {
			return !less
		}
		return less
	})
}
