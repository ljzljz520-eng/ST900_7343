package httpapi

import (
	"net/http"
	"time"

	"studio-console/audit"
	"studio-console/query"
)

func (s *Server) handleDashboard(writer http.ResponseWriter, _ *http.Request) {
	dashboard, err := query.LoadDashboard(s.store)
	if err != nil {
		writeError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, dashboard)
}

func (s *Server) handleAudit(writer http.ResponseWriter, request *http.Request) {
	limit, err := parsePositive(request.URL.Query().Get("limit"), 50)
	if err != nil {
		writeJSON(writer, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	var since time.Time
	if value := request.URL.Query().Get("since"); value != "" {
		since, err = time.Parse(time.RFC3339, value)
		if err != nil {
			writeJSON(writer, http.StatusBadRequest, errorResponse{Error: "since 必须是 RFC3339 时间"})
			return
		}
	}
	entries, err := s.audit.Search(audit.Filter{TargetID: request.URL.Query().Get("target_id"), ActorID: request.URL.Query().Get("actor_id"), Action: request.URL.Query().Get("action"), Since: since, Limit: limit})
	if err != nil {
		writeError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, map[string]any{"items": entries, "summary": audit.Summarize(entries)})
}

func (s *Server) handleListSettings(writer http.ResponseWriter, _ *http.Request) {
	settings, err := s.manager.ListSettings()
	if err != nil {
		writeError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, map[string]any{"items": settings})
}

func (s *Server) handleSetting(writer http.ResponseWriter, request *http.Request) {
	actor, err := actorID(request)
	if err != nil {
		writeJSON(writer, http.StatusUnauthorized, errorResponse{Error: err.Error()})
		return
	}
	var payload struct {
		Value string `json:"value"`
	}
	if err := s.decode(writer, request, &payload); err != nil {
		writeJSON(writer, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	setting, err := s.manager.SetSetting(actor, request.PathValue("key"), payload.Value)
	if err != nil {
		writeError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, setting)
}
