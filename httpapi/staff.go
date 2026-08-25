package httpapi

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"studio-console/domain"
	"studio-console/query"
	"studio-console/service"
	"studio-console/validation"
)

func (s *Server) handleHealth(writer http.ResponseWriter, _ *http.Request) {
	if err := s.store.Health(); err != nil {
		writeJSON(writer, http.StatusServiceUnavailable, errorResponse{Error: "存储不可用"})
		return
	}
	writeJSON(writer, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleCreateStaff(writer http.ResponseWriter, request *http.Request) {
	actor, err := actorID(request)
	if err != nil {
		writeJSON(writer, http.StatusUnauthorized, errorResponse{Error: err.Error()})
		return
	}
	var input validation.StaffInput
	if err := s.decode(writer, request, &input); err != nil {
		writeJSON(writer, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	account, err := s.manager.CreateStaff(service.CreateStaffCommand{ActorID: actor, Input: input})
	if err != nil {
		writeError(writer, err)
		return
	}
	writeJSON(writer, http.StatusCreated, account)
}

func (s *Server) handleUpdateStaff(writer http.ResponseWriter, request *http.Request) {
	actor, err := actorID(request)
	if err != nil {
		writeJSON(writer, http.StatusUnauthorized, errorResponse{Error: err.Error()})
		return
	}
	var input validation.StaffInput
	if err := s.decode(writer, request, &input); err != nil {
		writeJSON(writer, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	version, err := parseVersion(request, 1)
	if err != nil {
		writeJSON(writer, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	account, err := s.manager.UpdateStaff(service.UpdateStaffCommand{ActorID: actor, StaffID: request.PathValue("id"), ExpectedVersion: version, Input: input})
	if err != nil {
		writeError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, account)
}

func (s *Server) handleStaffDetails(writer http.ResponseWriter, request *http.Request) {
	account, contacts, err := s.manager.StaffDetails(request.PathValue("id"))
	if err != nil {
		writeError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, map[string]any{"staff": account, "contacts": contacts})
}

func (s *Server) handleListStaff(writer http.ResponseWriter, request *http.Request) {
	page, _ := strconv.Atoi(request.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(request.URL.Query().Get("page_size"))
	filter := query.StaffFilter{Search: request.URL.Query().Get("search"), SortField: request.URL.Query().Get("sort"), Descending: request.URL.Query().Get("direction") == "desc", Page: page, PageSize: pageSize}
	for _, value := range request.URL.Query()["role"] {
		role, err := domain.ParseRole(value)
		if err != nil {
			writeJSON(writer, http.StatusBadRequest, errorResponse{Error: "角色筛选值无效"})
			return
		}
		filter.Roles = append(filter.Roles, role)
	}
	for _, value := range request.URL.Query()["status"] {
		status, err := domain.ParseStatus(value)
		if err != nil {
			writeJSON(writer, http.StatusBadRequest, errorResponse{Error: "状态筛选值无效"})
			return
		}
		filter.Statuses = append(filter.Statuses, status)
	}
	result, err := s.staff.List(filter)
	if err != nil {
		writeJSON(writer, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	writeJSON(writer, http.StatusOK, result)
}

func (s *Server) handleChangeStatus(writer http.ResponseWriter, request *http.Request) {
	actor, err := actorID(request)
	if err != nil {
		writeJSON(writer, http.StatusUnauthorized, errorResponse{Error: err.Error()})
		return
	}
	var payload struct {
		Status string `json:"status"`
	}
	if err := s.decode(writer, request, &payload); err != nil {
		writeJSON(writer, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	status, err := domain.ParseStatus(payload.Status)
	if err != nil {
		writeError(writer, err)
		return
	}
	version, err := parseVersion(request, 1)
	if err != nil {
		writeJSON(writer, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	account, err := s.manager.ChangeStatus(service.ChangeStatusCommand{ActorID: actor, StaffID: request.PathValue("id"), Status: status, ExpectedVersion: version})
	if err != nil {
		writeError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, account)
}

func parsePositive(value string, fallback int) (int, error) {
	if strings.TrimSpace(value) == "" {
		return fallback, nil
	}
	number, err := strconv.Atoi(value)
	if err != nil || number < 1 {
		return 0, errors.New("参数必须是正整数")
	}
	return number, nil
}
