package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"studio-console/domain"
	"studio-console/store"
	"studio-console/validation"
)

type Clock interface {
	Now() time.Time
}

type FixedClock struct {
	Value time.Time
}

func (c FixedClock) Now() time.Time { return c.Value }

type Manager struct {
	store *store.Store
	clock Clock
	mu    sync.Mutex
	seq   uint64
}

func NewManager(database *store.Store, clock Clock) (*Manager, error) {
	if database == nil {
		return nil, errors.New("store is required")
	}
	if clock == nil {
		return nil, errors.New("clock is required")
	}
	return &Manager{store: database, clock: clock}, nil
}

type CreateStaffCommand struct {
	ActorID string
	Input   validation.StaffInput
}

type UpdateStaffCommand struct {
	ActorID         string
	StaffID         string
	ExpectedVersion uint64
	Input           validation.StaffInput
}

type ChangeStatusCommand struct {
	ActorID         string
	StaffID         string
	Status          domain.Status
	ExpectedVersion uint64
}

func (m *Manager) nextID(prefix string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.seq++
	return prefix + "-" + fmt.Sprintf("%06d", m.seq)
}

func (m *Manager) CreateStaff(command CreateStaffCommand) (domain.StaffAccount, error) {
	if strings.TrimSpace(command.ActorID) == "" {
		return domain.StaffAccount{}, errors.New("操作人不能为空")
	}
	input, err := validation.ValidateStaffInput(command.Input)
	if err != nil {
		return domain.StaffAccount{}, err
	}
	role, _ := domain.ParseRole(input.Role)
	now := m.clock.Now()
	account, err := domain.NewStaffAccount(m.nextID("staff"), input.Name, input.Phone, input.Email, role, now)
	if err != nil {
		return domain.StaffAccount{}, err
	}
	if input.Status != "" && input.Status != string(domain.StatusInvited) {
		status, _ := domain.ParseStatus(input.Status)
		if err := account.Transition(status, now); err != nil {
			return domain.StaffAccount{}, err
		}
	}
	if err := m.store.CreateStaff(account); err != nil {
		return domain.StaffAccount{}, translateStoreError(err)
	}
	contacts, err := m.contactsFromInput(account.ID, input, now)
	if err != nil {
		_ = m.store.DeleteStaff(account.ID)
		return domain.StaffAccount{}, err
	}
	for _, contact := range contacts {
		if err := m.store.PutContact(contact); err != nil {
			_ = m.store.DeleteStaff(account.ID)
			return domain.StaffAccount{}, fmt.Errorf("保存联系方式失败: %w", err)
		}
	}
	entry, _ := domain.NewAuditEntry(m.nextID("audit"), command.ActorID, "staff.created", account.ID, now, map[string]string{"role": string(account.Role), "name": account.Name})
	if err := m.store.AppendAudit(entry); err != nil {
		return domain.StaffAccount{}, fmt.Errorf("保存审计记录失败: %w", err)
	}
	return account, nil
}

func (m *Manager) contactsFromInput(staffID string, input validation.StaffInput, now time.Time) ([]domain.ContactMethod, error) {
	contacts := make([]domain.ContactMethod, 0, 2)
	if input.Phone != "" {
		method, err := domain.NewContactMethod(m.nextID("contact"), staffID, domain.ContactPhone, input.Phone, true, now)
		if err != nil {
			return nil, err
		}
		contacts = append(contacts, method)
	}
	if input.Email != "" {
		method, err := domain.NewContactMethod(m.nextID("contact"), staffID, domain.ContactEmail, input.Email, true, now)
		if err != nil {
			return nil, err
		}
		contacts = append(contacts, method)
	}
	return contacts, nil
}

func (m *Manager) UpdateStaff(command UpdateStaffCommand) (domain.StaffAccount, error) {
	if strings.TrimSpace(command.ActorID) == "" || strings.TrimSpace(command.StaffID) == "" {
		return domain.StaffAccount{}, errors.New("操作人和人员编号不能为空")
	}
	input, err := validation.ValidateStaffInput(command.Input)
	if err != nil {
		return domain.StaffAccount{}, err
	}
	account, err := m.store.GetStaff(command.StaffID)
	if err != nil {
		return domain.StaffAccount{}, translateStoreError(err)
	}
	if account == nil {
		return domain.StaffAccount{}, MissingRecordError{Entity: "人员", ID: command.StaffID}
	}
	role, _ := domain.ParseRole(input.Role)
	previousRole := account.Role
	if err := account.UpdateProfile(input.Name, input.Phone, input.Email, role, m.clock.Now()); err != nil {
		return domain.StaffAccount{}, err
	}
	if err := m.store.UpdateStaff(*account, command.ExpectedVersion); err != nil {
		return domain.StaffAccount{}, translateStoreError(err)
	}
	entry, _ := domain.NewAuditEntry(m.nextID("audit"), command.ActorID, "staff.updated", account.ID, m.clock.Now(), map[string]string{"previous_role": string(previousRole), "role": string(account.Role), "version": strconv.FormatUint(account.Version, 10)})
	if err := m.store.AppendAudit(entry); err != nil {
		return domain.StaffAccount{}, fmt.Errorf("保存审计记录失败: %w", err)
	}
	return *account, nil
}

func (m *Manager) ChangeStatus(command ChangeStatusCommand) (domain.StaffAccount, error) {
	if command.ActorID == "" || command.StaffID == "" {
		return domain.StaffAccount{}, errors.New("操作人和人员编号不能为空")
	}
	account, err := m.store.GetStaff(command.StaffID)
	if err != nil {
		return domain.StaffAccount{}, translateStoreError(err)
	}
	if account == nil {
		return domain.StaffAccount{}, MissingRecordError{Entity: "人员", ID: command.StaffID}
	}
	previous := account.Status
	if err := account.Transition(command.Status, m.clock.Now()); err != nil {
		return domain.StaffAccount{}, err
	}
	if err := m.store.UpdateStaff(*account, command.ExpectedVersion); err != nil {
		return domain.StaffAccount{}, translateStoreError(err)
	}
	entry, _ := domain.NewAuditEntry(m.nextID("audit"), command.ActorID, "staff.status_changed", account.ID, m.clock.Now(), map[string]string{"from": string(previous), "to": string(command.Status)})
	if err := m.store.AppendAudit(entry); err != nil {
		return domain.StaffAccount{}, fmt.Errorf("保存审计记录失败: %w", err)
	}
	return *account, nil
}

func (m *Manager) ActivateStaff(actorID, staffID string, version uint64) (domain.StaffAccount, error) {
	return m.ChangeStatus(ChangeStatusCommand{ActorID: actorID, StaffID: staffID, Status: domain.StatusActive, ExpectedVersion: version})
}

func (m *Manager) DisableStaff(actorID, staffID string, version uint64) (domain.StaffAccount, error) {
	return m.ChangeStatus(ChangeStatusCommand{ActorID: actorID, StaffID: staffID, Status: domain.StatusDisabled, ExpectedVersion: version})
}

func translateStoreError(err error) error {
	if errors.Is(err, store.ErrNotFound) {
		return MissingRecordError{Entity: "记录", ID: extractTrailingID(err.Error())}
	}
	if errors.Is(err, store.ErrConflict) {
		return ConflictError{Message: "联系方式已被占用或数据版本已更新"}
	}
	return err
}

func extractTrailingID(message string) string {
	parts := strings.Fields(message)
	if len(parts) == 0 {
		return "unknown"
	}
	return parts[len(parts)-1]
}
