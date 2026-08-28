package httpapi

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"studio-console/service"
	"studio-console/store"
)

func TestCreateStaffHTTP(t *testing.T) {
	database, _ := store.Open(filepath.Join(t.TempDir(), "studio.db"))
	defer database.Close()
	manager, _ := service.NewManager(database, service.FixedClock{Value: time.Date(2026, 8, 26, 9, 0, 0, 0, time.UTC)})
	server, _ := NewServer(database, manager, 1<<20)
	request := httptest.NewRequest(http.MethodPost, "/api/staff", strings.NewReader(`{"name":"林青","phone":"13800138000","role":"photographer"}`))
	request.Header.Set("X-Admin-ID", "admin")
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("unexpected response %d: %s", response.Code, response.Body.String())
	}
}
