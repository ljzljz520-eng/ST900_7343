package service

import "fmt"

type MissingRecordError struct {
	Entity string
	ID     string
}

func (e MissingRecordError) Error() string {
	return fmt.Sprintf("%s不存在: %s", e.Entity, e.ID)
}

type ConflictError struct {
	Message string
}

func (e ConflictError) Error() string {
	if e.Message == "" {
		return "数据冲突"
	}
	return e.Message
}

type OperationError struct {
	Operation string
	Cause     error
}

func (e OperationError) Error() string {
	return fmt.Sprintf("%s失败: %v", e.Operation, e.Cause)
}

func (e OperationError) Unwrap() error { return e.Cause }
