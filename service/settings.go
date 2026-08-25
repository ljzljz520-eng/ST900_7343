package service

import (
	"errors"
	"strings"

	"studio-console/domain"
	"studio-console/validation"
)

func (m *Manager) SetSetting(actorID, key, value string) (domain.StudioSetting, error) {
	if strings.TrimSpace(actorID) == "" {
		return domain.StudioSetting{}, errors.New("操作人不能为空")
	}
	if err := validation.ValidateSetting(key, value); err != nil {
		return domain.StudioSetting{}, err
	}
	setting, err := domain.NewStudioSetting(key, value, m.clock.Now())
	if err != nil {
		return domain.StudioSetting{}, err
	}
	if err := m.store.PutSetting(setting); err != nil {
		return domain.StudioSetting{}, OperationError{Operation: "保存设置", Cause: err}
	}
	entry, _ := domain.NewAuditEntry(m.nextID("audit"), actorID, "setting.updated", key, m.clock.Now(), map[string]string{"value": value})
	if err := m.store.AppendAudit(entry); err != nil {
		return domain.StudioSetting{}, OperationError{Operation: "保存设置审计", Cause: err}
	}
	return setting, nil
}

func (m *Manager) GetSetting(key string) (domain.StudioSetting, error) {
	setting, err := m.store.GetSetting(strings.TrimSpace(key))
	if err != nil {
		return domain.StudioSetting{}, translateStoreError(err)
	}
	return setting, nil
}

func (m *Manager) ListSettings() ([]domain.StudioSetting, error) {
	settings, err := m.store.ListSettings()
	if err != nil {
		return nil, OperationError{Operation: "读取设置", Cause: err}
	}
	return settings, nil
}

func (m *Manager) StaffDetails(id string) (domain.StaffAccount, []domain.ContactMethod, error) {
	account, err := m.store.GetStaff(strings.TrimSpace(id))
	if err != nil {
		return domain.StaffAccount{}, nil, translateStoreError(err)
	}
	if account == nil {
		return domain.StaffAccount{}, nil, MissingRecordError{Entity: "人员", ID: id}
	}
	contacts, err := m.store.ContactsForStaff(id)
	if err != nil {
		return domain.StaffAccount{}, nil, OperationError{Operation: "读取联系方式", Cause: err}
	}
	return *account, contacts, nil
}
