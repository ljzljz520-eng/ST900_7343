package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"studio-console/audit"
	"studio-console/domain"
	"studio-console/query"
	"studio-console/service"
	"studio-console/store"
	"studio-console/validation"
)

type Server struct {
	manager *service.Manager
	staff   *query.StaffReader
	audit   *audit.Reader
	store   *store.Store
	maxBody int64
	mux     *http.ServeMux
}

func NewServer(database *store.Store, manager *service.Manager, maxBody int64) (*Server, error) {
	if database == nil || manager == nil {
		return nil, errors.New("store and manager are required")
	}
	if maxBody < 1024 {
		return nil, errors.New("body limit must be at least 1024")
	}
	server := &Server{manager: manager, staff: query.NewStaffReader(database), audit: audit.NewReader(database), store: database, maxBody: maxBody, mux: http.NewServeMux()}
	server.routes()
	return server, nil
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /api/staff", s.handleListStaff)
	s.mux.HandleFunc("POST /api/staff", s.handleCreateStaff)
	s.mux.HandleFunc("GET /api/staff/{id}", s.handleStaffDetails)
	s.mux.HandleFunc("PUT /api/staff/{id}", s.handleUpdateStaff)
	s.mux.HandleFunc("POST /api/staff/{id}/status", s.handleChangeStatus)
	s.mux.HandleFunc("GET /api/dashboard", s.handleDashboard)
	s.mux.HandleFunc("GET /api/audit", s.handleAudit)
	s.mux.HandleFunc("GET /api/settings", s.handleListSettings)
	s.mux.HandleFunc("PUT /api/settings/{key}", s.handleSetting)
}

func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.mux.ServeHTTP(writer, request)
}

type errorResponse struct {
	Error  string                  `json:"error"`
	Fields []validation.FieldError `json:"fields,omitempty"`
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}

func writeError(writer http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	message := "操作失败，请稍后重试"
	response := errorResponse{}
	var missing service.MissingRecordError
	var conflict service.ConflictError
	var input validation.InputErrors
	switch {
	case errors.As(err, &missing):
		status, message = http.StatusNotFound, missing.Error()
	case errors.As(err, &conflict):
		status, message = http.StatusConflict, conflict.Error()
	case errors.As(err, &input):
		status, message, response.Fields = http.StatusBadRequest, "提交内容有误", input.Fields
	case errors.Is(err, domain.ErrInvalidTransition), errors.Is(err, domain.ErrInvalidRole), errors.Is(err, domain.ErrInvalidStatus):
		status, message = http.StatusBadRequest, err.Error()
	}
	response.Error = message
	writeJSON(writer, status, response)
}

func (s *Server) decode(writer http.ResponseWriter, request *http.Request, destination any) error {
	request.Body = http.MaxBytesReader(writer, request.Body, s.maxBody)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		if errors.Is(err, io.EOF) {
			return errors.New("请求内容不能为空")
		}
		return fmt.Errorf("请求格式不正确: %w", err)
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return errors.New("请求只能包含一个 JSON 对象")
	}
	return nil
}

func actorID(request *http.Request) (string, error) {
	value := strings.TrimSpace(request.Header.Get("X-Admin-ID"))
	if value == "" {
		return "", errors.New("缺少 X-Admin-ID 请求头")
	}
	return value, nil
}

func parseVersion(request *http.Request, fallback uint64) (uint64, error) {
	value := strings.TrimSpace(request.Header.Get("If-Match"))
	if value == "" {
		return fallback, nil
	}
	version, err := strconv.ParseUint(strings.Trim(value, `"`), 10, 64)
	if err != nil || version == 0 {
		return 0, errors.New("If-Match 必须是有效版本号")
	}
	return version, nil
}
