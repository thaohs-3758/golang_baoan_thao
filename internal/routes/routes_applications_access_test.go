package routes

import (
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/handlers"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func init() {
	_ = os.Setenv("JWT_SECRET", "test-secret-for-routes")
}

type fakeRenderer struct{}

func (r *fakeRenderer) Render(_ *echo.Context, w io.Writer, _ string, _ interface{}) error {
	_, _ = w.Write([]byte("ok"))
	return nil
}

type fakeAdminApplicationsSvc struct{}

func (s *fakeAdminApplicationsSvc) ListApplications(_ repositories.ApplicationFilter, _, _ int) ([]models.Application, int64, error) {
	return []models.Application{{ID: "a1", ApplicationCode: "APP-1", Status: models.ApplicationStatusReceived}}, 1, nil
}

func (s *fakeAdminApplicationsSvc) ListApplicationsForActor(_ repositories.ApplicationFilter, _, _ int, _ models.UserRole, _ string) ([]models.Application, int64, error) {
	return []models.Application{{ID: "a1", ApplicationCode: "APP-1", Status: models.ApplicationStatusReceived}}, 1, nil
}

func (s *fakeAdminApplicationsSvc) GetApplication(_ string) (*models.Application, error) {
	return &models.Application{
		ID:              "a1",
		ApplicationCode: "APP-1",
		Status:          models.ApplicationStatusReceived,
		SubmittedAt:     time.Now(),
		ServiceType:     models.ServiceType{Name: "Svc"},
		CitizenUser:     models.User{Name: "Citizen"},
	}, nil
}

func (s *fakeAdminApplicationsSvc) AssignToStaff(_ string, _ *string, _ string) error {
	return nil
}

func (s *fakeAdminApplicationsSvc) ProcessApplication(_ string, _ models.ApplicationStatus, _ string, _ []*multipart.FileHeader, _ string) error {
	return nil
}

type fakeAdminProfileSvc struct{}

func (s *fakeAdminProfileSvc) GetSelf(_ string) (*models.User, error) {
	return &models.User{ID: "u1", Name: "Admin", Role: models.UserRoleManager}, nil
}

func (s *fakeAdminProfileSvc) UpdateSelfContact(_ string, _ *dtos.UpdateAdminProfileRequest) (*models.User, error) {
	return &models.User{ID: "u1", Name: "Admin", Role: models.UserRoleManager}, nil
}

func (s *fakeAdminProfileSvc) ChangeSelfPassword(_ string, _ *dtos.ChangeAdminPasswordRequest) error {
	return nil
}

func makeRefreshTokenForRole(role models.UserRole) string {
	user := &models.User{ID: "u1", Email: "u1@test.com", Role: role}
	token, _ := configs.GenerateRefreshToken(user)
	return token
}

func newRoutesEchoForAppsAccess() *echo.Echo {
	e := echo.New()
	e.Renderer = &fakeRenderer{}

	appHandler := handlers.NewAdminApplicationHandler(&fakeAdminApplicationsSvc{}, nil, nil)
	apiHandler := &ApiHandler{
		AdminApplicationHandler: appHandler,
		AdminProfileHandler:     handlers.NewAdminProfileHandler(&fakeAdminProfileSvc{}),
		AdminAuthHandler:        &handlers.AdminAuthHandler{},
		AdminDashboardHandler:   &handlers.AdminDashboardHandler{},
		AdminUserHandler:        &handlers.AdminUserHandler{},
		AdminDepartmentHandler:  &handlers.AdminDepartmentHandler{},
		AdminCategoryHandler:    &handlers.AdminCategoryHandler{},
		AdminLogHandler:         &handlers.AdminLogHandler{},
		AdminCitizenHandler:     &handlers.AdminCitizenHandler{},
		AuthHandler:             &handlers.AuthHandler{},
		CitizenProfileHandler:   &handlers.CitizenProfileHandler{},
		ServiceCatalogHandler:   &handlers.ServiceCatalogHandler{},
		ApplicationHandler:      &handlers.ApplicationHandler{},
	}
	SetupRoutes(e, apiHandler)
	return e
}

func requestWithRole(e *echo.Echo, method, path string, role models.UserRole) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: makeRefreshTokenForRole(role)})
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestApplicationsRouteAccessPolicy(t *testing.T) {
	e := newRoutesEchoForAppsAccess()

	tests := []struct {
		name        string
		role        models.UserRole
		method      string
		path        string
		wantAllowed bool
	}{
		{name: "super_admin can read list", role: models.UserRoleSuperAdmin, method: http.MethodGet, path: "/admin/applications", wantAllowed: true},
		{name: "super_admin can export", role: models.UserRoleSuperAdmin, method: http.MethodGet, path: "/admin/applications/export", wantAllowed: true},
		{name: "super_admin can view detail", role: models.UserRoleSuperAdmin, method: http.MethodGet, path: "/admin/applications/a1", wantAllowed: true},
		{name: "staff can read list", role: models.UserRoleStaff, method: http.MethodGet, path: "/admin/applications", wantAllowed: true},
		{name: "staff can view detail", role: models.UserRoleStaff, method: http.MethodGet, path: "/admin/applications/a1", wantAllowed: true},
		{name: "staff can process", role: models.UserRoleStaff, method: http.MethodPost, path: "/admin/applications/a1/process", wantAllowed: true},
		{name: "super_admin cannot process", role: models.UserRoleSuperAdmin, method: http.MethodPost, path: "/admin/applications/a1/process", wantAllowed: false},
		{name: "manager cannot process", role: models.UserRoleManager, method: http.MethodPost, path: "/admin/applications/a1/process", wantAllowed: false},
		{name: "manager can assign", role: models.UserRoleManager, method: http.MethodPost, path: "/admin/applications/a1/assign", wantAllowed: true},
		{name: "super_admin cannot open assign form", role: models.UserRoleSuperAdmin, method: http.MethodGet, path: "/admin/applications/a1/assign", wantAllowed: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := requestWithRole(e, tt.method, tt.path, tt.role)
			if tt.wantAllowed {
				if rec.Code == http.StatusSeeOther {
					assert.NotEqual(t, "/admin/login", rec.Header().Get("Location"))
					return
				}
				assert.NotEqual(t, http.StatusForbidden, rec.Code)
				return
			}
			assert.NotEqual(t, http.StatusOK, rec.Code)
			assert.Contains(t, []int{http.StatusSeeOther, http.StatusForbidden}, rec.Code)
		})
	}
}

func TestApplicationsAssignAccessPolicy_OnlyManagerAllowed(t *testing.T) {
	e := newRoutesEchoForAppsAccess()

	tests := []struct {
		name        string
		role        models.UserRole
		wantAllowed bool
	}{
		{name: "manager allowed", role: models.UserRoleManager, wantAllowed: true},
		{name: "staff forbidden", role: models.UserRoleStaff, wantAllowed: false},
		{name: "super_admin forbidden", role: models.UserRoleSuperAdmin, wantAllowed: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := requestWithRole(e, http.MethodGet, "/admin/applications/a1/assign", tt.role)
			if tt.wantAllowed {
				assert.Equal(t, http.StatusOK, rec.Code)
				return
			}
			assert.NotEqual(t, http.StatusOK, rec.Code)
			assert.Contains(t, []int{http.StatusSeeOther, http.StatusForbidden}, rec.Code)
		})
	}
}

func TestAdminProfileRouteAccessPolicy(t *testing.T) {
	e := newRoutesEchoForAppsAccess()

	tests := []struct {
		name        string
		role        models.UserRole
		wantAllowed bool
	}{
		{name: "staff allowed", role: models.UserRoleStaff, wantAllowed: true},
		{name: "manager denied", role: models.UserRoleManager, wantAllowed: false},
		{name: "super_admin denied", role: models.UserRoleSuperAdmin, wantAllowed: false},
		{name: "citizen denied", role: models.UserRoleCitizen, wantAllowed: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := requestWithRole(e, http.MethodGet, "/admin/profile", tt.role)
			if tt.wantAllowed {
				assert.Equal(t, http.StatusOK, rec.Code)
				return
			}
			assert.NotEqual(t, http.StatusOK, rec.Code)
			assert.Contains(t, []int{http.StatusSeeOther, http.StatusForbidden}, rec.Code)
		})
	}
}

func TestDepartmentStaffRouteAccessPolicy_ManagerOnly(t *testing.T) {
	e := newRoutesEchoForAppsAccess()

	tests := []struct {
		name   string
		role   models.UserRole
		method string
		path   string
	}{
		{name: "super_admin cannot view department staff list", role: models.UserRoleSuperAdmin, method: http.MethodGet, path: "/admin/departments/d1/staff"},
		{name: "super_admin cannot open assign form", role: models.UserRoleSuperAdmin, method: http.MethodGet, path: "/admin/departments/d1/staff/assign"},
		{name: "super_admin cannot assign staff", role: models.UserRoleSuperAdmin, method: http.MethodPost, path: "/admin/departments/d1/staff/assign"},
		{name: "super_admin cannot remove staff", role: models.UserRoleSuperAdmin, method: http.MethodPost, path: "/admin/departments/d1/staff/u1/remove"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := requestWithRole(e, tt.method, tt.path, tt.role)
			assert.NotEqual(t, http.StatusOK, rec.Code)
			assert.Contains(t, []int{http.StatusSeeOther, http.StatusForbidden}, rec.Code)
		})
	}
}
