package validation

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"

	"studio-console/domain"
)

var phonePattern = regexp.MustCompile(`^\+?[0-9][0-9 -]{5,19}$`)

type StaffInput struct {
	Name   string `json:"name"`
	Phone  string `json:"phone"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Status string `json:"status,omitempty"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type InputErrors struct {
	Fields []FieldError `json:"fields"`
}

func (e InputErrors) Error() string {
	parts := make([]string, 0, len(e.Fields))
	for _, field := range e.Fields {
		parts = append(parts, field.Field+": "+field.Message)
	}
	return strings.Join(parts, "; ")
}

func NormalizeStaffInput(input StaffInput) StaffInput {
	input.Name = strings.Join(strings.Fields(input.Name), " ")
	input.Phone = strings.TrimSpace(input.Phone)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.Role = strings.ToLower(strings.TrimSpace(input.Role))
	input.Status = strings.ToLower(strings.TrimSpace(input.Status))
	return input
}

func ValidateStaffInput(input StaffInput) (StaffInput, error) {
	input = NormalizeStaffInput(input)
	problems := make([]FieldError, 0)
	if input.Name == "" {
		problems = append(problems, FieldError{Field: "name", Message: "姓名不能为空"})
	} else if utf8.RuneCountInString(input.Name) > 80 {
		problems = append(problems, FieldError{Field: "name", Message: "姓名不能超过80个字符"})
	}
	if input.Phone == "" && input.Email == "" {
		problems = append(problems, FieldError{Field: "contact", Message: "手机和邮箱至少填写一项"})
	}
	if input.Phone != "" && !phonePattern.MatchString(input.Phone) {
		problems = append(problems, FieldError{Field: "phone", Message: "手机号格式不正确"})
	}
	if input.Email != "" {
		if _, err := mail.ParseAddress(input.Email); err != nil {
			problems = append(problems, FieldError{Field: "email", Message: "邮箱格式不正确"})
		}
	}
	if _, err := domain.ParseRole(input.Role); err != nil {
		problems = append(problems, FieldError{Field: "role", Message: "角色必须是摄影师、修图师、化妆师或客服"})
	}
	if input.Status != "" {
		if _, err := domain.ParseStatus(input.Status); err != nil {
			problems = append(problems, FieldError{Field: "status", Message: "账号状态无效"})
		}
	}
	if len(problems) > 0 {
		return StaffInput{}, InputErrors{Fields: problems}
	}
	return input, nil
}

func ValidatePage(page, pageSize int) error {
	if page < 1 {
		return errors.New("页码必须大于0")
	}
	if pageSize < 1 || pageSize > 100 {
		return errors.New("每页数量必须在1到100之间")
	}
	return nil
}

func ValidateSetting(key, value string) error {
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)
	if key == "" || value == "" {
		return errors.New("设置名称和值不能为空")
	}
	allowed := map[string]bool{"studio_name": true, "timezone": true, "default_country_code": true, "booking_notice": true}
	if !allowed[key] {
		return fmt.Errorf("不支持的设置项: %s", key)
	}
	if len(value) > 500 {
		return errors.New("设置值不能超过500个字符")
	}
	return nil
}
