package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	"github.com/labstack/echo/v5"
)

// stubRenderer is already declared in admin_user_handler_test.go (same package).
// newAdminEcho() (admin_user_handler_test.go) sets e.Renderer = &stubRenderer{}.

// --- fake serviceCatalogSvc ---

type fakeCatalogSvc struct {
	listFn     func(ctx context.Context, filter repositories.ListFilter) (*repositories.ListResult, error)
	getByIDFn  func(ctx context.Context, id string) (*models.ServiceType, error)
}

func (f *fakeCatalogSvc) List(ctx context.Context, filter repositories.ListFilter) (*repositories.ListResult, error) {
	if f.listFn != nil {
		return f.listFn(ctx, filter)
	}
	return &repositories.ListResult{Items: nil, Total: 0}, nil
}
func (f *fakeCatalogSvc) GetByID(ctx context.Context, id string) (*models.ServiceType, error) {
	if f.getByIDFn != nil {
		return f.getByIDFn(ctx, id)
	}
	return &models.ServiceType{
		ID:         id,
		Name:       "Test Service",
		FormSchema: json.RawMessage(`{"fields":["name"],"required_fields":["name"]}`),
	}, nil
}
func (f *fakeCatalogSvc) GetByIDForAdmin(ctx context.Context, id string) (*models.ServiceType, error) {
	return f.GetByID(ctx, id)
}
func (f *fakeCatalogSvc) ListDepartments(ctx context.Context) ([]models.Department, error) { return nil, nil }
func (f *fakeCatalogSvc) ListCategories(ctx context.Context) ([]models.Category, error)    { return nil, nil }
func (f *fakeCatalogSvc) Create(ctx context.Context, st *models.ServiceType) error         { return nil }
func (f *fakeCatalogSvc) Update(ctx context.Context, st *models.ServiceType) error         { return nil }
func (f *fakeCatalogSvc) Delete(ctx context.Context, id string) error                      { return nil }

// --- fake citizenAppSvc ---

type fakeAppWebSvc struct {
	submitFn     func(uid string, req *dtos.SubmitApplicationRequest, files []*multipart.FileHeader) (*dtos.ApplicationResponse, error)
	listFn       func(uid string, page, limit int) ([]models.Application, int64, error)
	getFn        func(uid, appID string) (*dtos.ApplicationResponse, error)
	supplementFn func(uid, appID string, files []*multipart.FileHeader) ([]dtos.ApplicationAttachmentResponse, error)
}

func (f *fakeAppWebSvc) SubmitApplication(uid string, req *dtos.SubmitApplicationRequest, files []*multipart.FileHeader) (*dtos.ApplicationResponse, error) {
	if f.submitFn != nil {
		return f.submitFn(uid, req, files)
	}
	return &dtos.ApplicationResponse{ID: "app-id", ApplicationCode: "CODE-001"}, nil
}
func (f *fakeAppWebSvc) ListMyApplications(uid string, page, limit int) ([]models.Application, int64, error) {
	if f.listFn != nil {
		return f.listFn(uid, page, limit)
	}
	return nil, 0, nil
}
func (f *fakeAppWebSvc) GetMyApplication(uid, appID string) (*dtos.ApplicationResponse, error) {
	if f.getFn != nil {
		return f.getFn(uid, appID)
	}
	return &dtos.ApplicationResponse{ID: appID, ApplicationCode: "CODE-001", Status: "received"}, nil
}
func (f *fakeAppWebSvc) ListMyApplicationStatusHistory(uid, appID string, page, limit int, since *time.Time) ([]models.ApplicationStatusLog, int64, error) {
	return nil, 0, nil
}
func (f *fakeAppWebSvc) UploadMyApplicationSupplements(uid, appID string, files []*multipart.FileHeader) ([]dtos.ApplicationAttachmentResponse, error) {
	if f.supplementFn != nil {
		return f.supplementFn(uid, appID, files)
	}
	return nil, nil
}

type fakeCitizenProfileWebSvc struct {
	getFn    func(userID string) (*dtos.CitizenProfileResponse, error)
	updateFn func(userID string, req *dtos.UpdateCitizenProfileRequest) (*dtos.CitizenProfileResponse, error)
	passFn   func(userID string, req *dtos.ChangeMyPasswordRequest) error
}

func (f *fakeCitizenProfileWebSvc) GetProfile(userID string) (*dtos.CitizenProfileResponse, error) {
	if f.getFn != nil {
		return f.getFn(userID)
	}
	return &dtos.CitizenProfileResponse{
		UserID:          userID,
		Name:            "Citizen A",
		Email:           "citizen@test.com",
		Phone:           "0900000000",
		Address:         "HCM",
		CitizenIDNumber: "012345678901",
	}, nil
}

func (f *fakeCitizenProfileWebSvc) UpdateProfile(userID string, req *dtos.UpdateCitizenProfileRequest) (*dtos.CitizenProfileResponse, error) {
	if f.updateFn != nil {
		return f.updateFn(userID, req)
	}
	phone := ""
	if req.Phone != nil {
		phone = *req.Phone
	}
	address := ""
	if req.Address != nil {
		address = *req.Address
	}
	return &dtos.CitizenProfileResponse{
		UserID:          userID,
		Name:            "Citizen A",
		Email:           "citizen@test.com",
		Phone:           phone,
		Address:         address,
		CitizenIDNumber: "012345678901",
	}, nil
}

func (f *fakeCitizenProfileWebSvc) ChangeMyPassword(userID string, req *dtos.ChangeMyPasswordRequest) error {
	if f.passFn != nil {
		return f.passFn(userID, req)
	}
	return nil
}

// --- helpers ---

func newCitizenCtx(e *echo.Echo, method, target string) (*echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, target, nil)
	req.Header.Set("Accept-Language", "vi")
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func setCitizenUser(c *echo.Context, id string) {
	c.Set("user", &configs.JwtCustomClaims{ID: id, Email: "citizen@test.com", Role: "citizen"})
}

func setPathParam(c *echo.Context, name, value string) {
	c.SetPathValues(echo.PathValues{{Name: name, Value: value}})
}

// --- ShowServiceCatalog tests ---

func TestShowServiceCatalog_ReturnsErrorWhenCatalogNil(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{})

	c, _ := newCitizenCtx(e, http.MethodGet, "/citizen/services")
	setCitizenUser(c, "u1")

	err := h.ShowServiceCatalog(c)
	var he *echo.HTTPError
	if !errors.As(err, &he) || he.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 HTTPError, got %v", err)
	}
}

func TestShowServiceCatalog_RendersListPage(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	catalog := &fakeCatalogSvc{}
	h := NewCitizenWebHandler(&fakeAuthService{}).WithCatalogService(catalog)

	c, rec := newCitizenCtx(e, http.MethodGet, "/citizen/services")
	setCitizenUser(c, "u1")

	if err := h.ShowServiceCatalog(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// --- ShowApplicationsList tests ---

func TestShowApplicationsList_ReturnsErrorWhenNoUser(t *testing.T) {
	e := newAdminEcho()
	appSvc := &fakeAppWebSvc{}
	h := NewCitizenWebHandler(&fakeAuthService{}).WithApplicationService(appSvc)

	c, _ := newCitizenCtx(e, http.MethodGet, "/citizen/applications")
	// no user set in context

	err := h.ShowApplicationsList(c)
	var he *echo.HTTPError
	if !errors.As(err, &he) || he.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 HTTPError, got %v", err)
	}
}

func TestShowApplicationsList_RendersListPage(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	appSvc := &fakeAppWebSvc{}
	h := NewCitizenWebHandler(&fakeAuthService{}).WithApplicationService(appSvc)

	c, rec := newCitizenCtx(e, http.MethodGet, "/citizen/applications")
	setCitizenUser(c, "u1")

	if err := h.ShowApplicationsList(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestShowProfilePage_RendersProfile(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{}).WithProfileService(&fakeCitizenProfileWebSvc{})

	c, rec := newCitizenCtx(e, http.MethodGet, "/citizen/profile")
	setCitizenUser(c, "u1")

	if err := h.ShowProfilePage(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestUpdateProfile_RedirectsOnSuccess(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{}).WithProfileService(&fakeCitizenProfileWebSvc{})

	req := httptest.NewRequest(http.MethodPost, "/citizen/profile", strings.NewReader("phone=0911&address=HN"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setCitizenUser(c, "u1")

	if err := h.UpdateProfile(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", rec.Code)
	}
}

func TestUpdateProfile_PassesPersonalFieldsToService(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	var capturedReq *dtos.UpdateCitizenProfileRequest
	h := NewCitizenWebHandler(&fakeAuthService{}).WithProfileService(&fakeCitizenProfileWebSvc{
		updateFn: func(userID string, req *dtos.UpdateCitizenProfileRequest) (*dtos.CitizenProfileResponse, error) {
			capturedReq = req
			return &dtos.CitizenProfileResponse{UserID: userID}, nil
		},
	})

	form := "phone=0911&address=HN&gender=male&permanent_address=Quan+1&date_of_birth=1999-10-20"
	req := httptest.NewRequest(http.MethodPost, "/citizen/profile", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setCitizenUser(c, "u1")

	if err := h.UpdateProfile(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedReq == nil {
		t.Fatal("expected profile update request to be captured")
	}
	if capturedReq.Gender == nil || *capturedReq.Gender != "male" {
		t.Fatalf("expected gender=male, got %+v", capturedReq.Gender)
	}
	if capturedReq.PermanentAddress == nil || *capturedReq.PermanentAddress != "Quan 1" {
		t.Fatalf("expected permanent_address='Quan 1', got %+v", capturedReq.PermanentAddress)
	}
	if capturedReq.DateOfBirth == nil || capturedReq.DateOfBirth.Format("2006-01-02") != "1999-10-20" {
		t.Fatalf("expected date_of_birth=1999-10-20, got %+v", capturedReq.DateOfBirth)
	}
}

func TestChangePassword_RedirectsOnSuccess(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{}).WithProfileService(&fakeCitizenProfileWebSvc{})

	req := httptest.NewRequest(http.MethodPost, "/citizen/profile/password", strings.NewReader("current_password=old12345&new_password=new12345&confirm_new_password=new12345"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setCitizenUser(c, "u1")

	if err := h.ChangePassword(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected 303, got %d", rec.Code)
	}
}

// --- SubmitApplication tests ---

func TestSubmitApplication_RedirectsOnSuccess(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	catalog := &fakeCatalogSvc{}
	appSvc := &fakeAppWebSvc{}
	h := NewCitizenWebHandler(&fakeAuthService{}).
		WithCatalogService(catalog).
		WithApplicationService(appSvc)

	req := httptest.NewRequest(http.MethodPost, "/citizen/applications", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept-Language", "vi")
	req.ParseForm()
	req.Form.Set("service_type_id", "svc-id")
	req.Form.Set("name", "Nguyen Van A")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setCitizenUser(c, "u1")

	if err := h.SubmitApplication(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect 303, got %d", rec.Code)
	}
	if rec.Header().Get("Location") == "" {
		t.Fatal("expected Location header on redirect")
	}
}

func TestSubmitApplication_RendersFormOnServiceError(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	catalog := &fakeCatalogSvc{}
	appSvc := &fakeAppWebSvc{
		submitFn: func(uid string, req *dtos.SubmitApplicationRequest, files []*multipart.FileHeader) (*dtos.ApplicationResponse, error) {
			return nil, services.ErrMissingRequiredField
		},
	}
	h := NewCitizenWebHandler(&fakeAuthService{}).
		WithCatalogService(catalog).
		WithApplicationService(appSvc)

	req := httptest.NewRequest(http.MethodPost, "/citizen/applications", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept-Language", "vi")
	req.ParseForm()
	req.Form.Set("service_type_id", "svc-id")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setCitizenUser(c, "u1")

	if err := h.SubmitApplication(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
}

// --- ShowApplicationDetail tests ---

func TestShowApplicationDetail_Returns404WhenNotFound(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	appSvc := &fakeAppWebSvc{
		getFn: func(uid, appID string) (*dtos.ApplicationResponse, error) {
			return nil, services.ErrApplicationNotFound
		},
	}
	h := NewCitizenWebHandler(&fakeAuthService{}).WithApplicationService(appSvc)

	c, _ := newCitizenCtx(e, http.MethodGet, "/citizen/applications/unknown-id")
	setPathParam(c, "id", "unknown-id")
	setCitizenUser(c, "u1")

	err := h.ShowApplicationDetail(c)
	var he *echo.HTTPError
	if !errors.As(err, &he) || he.Code != http.StatusNotFound {
		t.Fatalf("expected 404 HTTPError, got %v", err)
	}
}

func TestShowApplicationDetail_RendersDetailPage(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	appSvc := &fakeAppWebSvc{}
	h := NewCitizenWebHandler(&fakeAuthService{}).WithApplicationService(appSvc)

	c, rec := newCitizenCtx(e, http.MethodGet, "/citizen/applications/app-id")
	setPathParam(c, "id", "app-id")
	setCitizenUser(c, "u1")

	if err := h.ShowApplicationDetail(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// --- ShowLoginPage ---

func TestShowLoginPage_Renders(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{})
	c, rec := newCitizenCtx(e, http.MethodGet, "/login")
	if err := h.ShowLoginPage(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// --- WebLogin ---

func TestWebLogin_Success_CitizenRedirects(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{
		loginFn: func(req *dtos.LoginRequest) (*models.User, string, string, error) {
			return &models.User{ID: "u1", Role: models.UserRoleCitizen}, "access", "refresh", nil
		},
	})
	form := "email=citizen%40example.com&password=pass"
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.WebLogin(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
}

func TestWebLogin_NonCitizen_ReturnsUnauthorized(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{
		loginFn: func(req *dtos.LoginRequest) (*models.User, string, string, error) {
			return &models.User{ID: "a1", Role: models.UserRoleSuperAdmin}, "access", "refresh", nil
		},
	})
	form := "email=admin%40example.com&password=pass"
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.WebLogin(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestWebLogin_NilUser_Returns500(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{
		loginFn: func(req *dtos.LoginRequest) (*models.User, string, string, error) {
			return nil, "", "", nil
		},
	})
	form := "email=x%40x.com&password=pass"
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	err := h.WebLogin(c)
	if err == nil {
		t.Fatal("expected error for nil user")
	}
}

func TestWebLogin_LoginError_RendersForm(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{
		loginFn: func(req *dtos.LoginRequest) (*models.User, string, string, error) {
			return nil, "", "", services.ErrUserNotFound
		},
	})
	form := "email=x%40x.com&password=wrong"
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.WebLogin(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestWebLogin_PasswordMismatch(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{
		loginFn: func(req *dtos.LoginRequest) (*models.User, string, string, error) {
			return nil, "", "", services.ErrPasswordMismatch
		},
	})
	form := "email=x%40x.com&password=wrong"
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = h.WebLogin(c)
}

func TestWebLogin_UserBlocked(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{
		loginFn: func(req *dtos.LoginRequest) (*models.User, string, string, error) {
			return nil, "", "", services.ErrUserBlocked
		},
	})
	form := "email=x%40x.com&password=pass"
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = h.WebLogin(c)
}

func TestWebLogin_GenericError(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{
		loginFn: func(req *dtos.LoginRequest) (*models.User, string, string, error) {
			return nil, "", "", errors.New("generic error")
		},
	})
	form := "email=x%40x.com&password=pass"
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	_ = h.WebLogin(c)
}

// --- ShowRegisterPage ---

func TestShowRegisterPage_Renders(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{})
	c, rec := newCitizenCtx(e, http.MethodGet, "/register")
	if err := h.ShowRegisterPage(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// --- WebRegister ---

func TestWebRegister_Success(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{
		registerFn: func(req *dtos.RegisterRequest) (*models.User, error) {
			return &models.User{ID: "u1", Role: models.UserRoleCitizen}, nil
		},
	})
	// All required fields: name, email, citizen_id_number (12 digits), password, confirm_password
	form := "email=new%40example.com&password=pass123&confirm_password=pass123&name=User&citizen_id_number=123456789012"
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.WebRegister(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
}

func TestWebRegister_Error_RendersForm(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{
		registerFn: func(req *dtos.RegisterRequest) (*models.User, error) {
			return nil, errors.New("email already exists")
		},
	})
	form := "email=existing%40example.com&password=pass123&name=User"
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.WebRegister(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
}

// --- WebLogout ---

func TestWebLogout_Redirects(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{})
	c, rec := newCitizenCtx(e, http.MethodPost, "/logout")
	if err := h.WebLogout(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
}

func TestWebLogout_WithUser_Redirects(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{})
	c, rec := newCitizenCtx(e, http.MethodPost, "/logout")
	setCitizenUser(c, "u1")
	if err := h.WebLogout(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
}

// --- ShowDashboard ---

func TestShowDashboard_Renders(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{})
	c, rec := newCitizenCtx(e, http.MethodGet, "/citizen")
	setCitizenUser(c, "u1")
	if err := h.ShowDashboard(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// --- ListNotifications ---

func TestListNotifications_NoUser_Returns500(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{})
	c, _ := newCitizenCtx(e, http.MethodGet, "/citizen/notifications")
	err := h.ListNotifications(c)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

type fakeNotifSvc struct {
	items      []dtos.NotificationResponse
	total      int64
	listErr    error
	markErr    error
	markAllErr error
	countVal   int64
	countErr   error
}

func (s *fakeNotifSvc) List(_ string, _ repositories.NotificationFilter, _, _ int) ([]dtos.NotificationResponse, int64, error) {
	return s.items, s.total, s.listErr
}
func (s *fakeNotifSvc) MarkAsRead(_, _ string) error { return s.markErr }
func (s *fakeNotifSvc) MarkAllAsRead(_ string) error { return s.markAllErr }
func (s *fakeNotifSvc) CountUnread(_ string) (int64, error) {
	return s.countVal, s.countErr
}

func TestListNotifications_Success(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{}).WithNotificationService(&fakeNotifSvc{})
	c, rec := newCitizenCtx(e, http.MethodGet, "/citizen/notifications")
	setCitizenUser(c, "u1")
	if err := h.ListNotifications(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestListNotifications_WithIsReadFilter(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{}).WithNotificationService(&fakeNotifSvc{})
	req := httptest.NewRequest(http.MethodGet, "/citizen/notifications?is_read=true&type=system", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setCitizenUser(c, "u1")
	if err := h.ListNotifications(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListNotifications_WithIsReadFalse(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{}).WithNotificationService(&fakeNotifSvc{})
	req := httptest.NewRequest(http.MethodGet, "/citizen/notifications?is_read=false", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setCitizenUser(c, "u1")
	_ = h.ListNotifications(c)
}

func TestListNotifications_ServiceError(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{}).WithNotificationService(&fakeNotifSvc{listErr: errors.New("db error")})
	c, _ := newCitizenCtx(e, http.MethodGet, "/citizen/notifications")
	setCitizenUser(c, "u1")
	err := h.ListNotifications(c)
	if err == nil {
		t.Fatal("expected error")
	}
}

// --- MarkNotificationRead ---

func TestMarkNotificationRead_NoUser_Redirects(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{})
	c, rec := newCitizenCtx(e, http.MethodPost, "/citizen/notifications/n1/read")
	_ = h.MarkNotificationRead(c)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
}

func TestMarkNotificationRead_Success(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{}).WithNotificationService(&fakeNotifSvc{})
	c, rec := newCitizenCtx(e, http.MethodPost, "/citizen/notifications/n1/read")
	setCitizenUser(c, "u1")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "n1"}})
	_ = h.MarkNotificationRead(c)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
}

// --- MarkAllNotificationsRead ---

func TestMarkAllNotificationsRead_NoUser_Redirects(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{})
	c, rec := newCitizenCtx(e, http.MethodPost, "/citizen/notifications/read-all")
	_ = h.MarkAllNotificationsRead(c)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
}

func TestMarkAllNotificationsRead_Success(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{}).WithNotificationService(&fakeNotifSvc{})
	c, rec := newCitizenCtx(e, http.MethodPost, "/citizen/notifications/read-all")
	setCitizenUser(c, "u1")
	_ = h.MarkAllNotificationsRead(c)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
}

func TestMarkAllNotificationsRead_Error_Redirects(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{}).WithNotificationService(&fakeNotifSvc{markAllErr: errors.New("fail")})
	c, rec := newCitizenCtx(e, http.MethodPost, "/citizen/notifications/read-all")
	setCitizenUser(c, "u1")
	_ = h.MarkAllNotificationsRead(c)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
}

// --- ShowServiceDetail ---

func TestShowServiceDetail_NoService_Returns500(t *testing.T) {
	e := newAdminEcho()
	// catalogSvc is nil → should return 500
	h := NewCitizenWebHandler(&fakeAuthService{})
	c, _ := newCitizenCtx(e, http.MethodGet, "/citizen/services/svc-1")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "svc-1"}})
	err := h.ShowServiceDetail(c)
	if err == nil {
		t.Fatal("expected error for nil catalogSvc")
	}
}

func TestShowServiceDetail_GetByIDError_Returns404(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{}).WithCatalogService(&fakeCatalogSvc{
		getByIDFn: func(_ context.Context, _ string) (*models.ServiceType, error) {
			return nil, errors.New("not found")
		},
	})
	c, _ := newCitizenCtx(e, http.MethodGet, "/citizen/services/svc-1")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "svc-1"}})
	err := h.ShowServiceDetail(c)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestShowServiceDetail_Renders(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{}).WithCatalogService(&fakeCatalogSvc{
		getByIDFn: func(_ context.Context, _ string) (*models.ServiceType, error) {
			return &models.ServiceType{ID: "svc-1", Name: "Service"}, nil
		},
	})
	c, rec := newCitizenCtx(e, http.MethodGet, "/citizen/services/svc-1")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "svc-1"}})
	setCitizenUser(c, "u1")
	if err := h.ShowServiceDetail(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestShowServiceCatalog_UnreadCountError(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	catalog := &fakeCatalogSvc{}
	notifSvc := &fakeNotifSvc{countErr: errors.New("count db error")}
	h := NewCitizenWebHandler(&fakeAuthService{}).
		WithCatalogService(catalog).
		WithNotificationService(notifSvc)

	c, rec := newCitizenCtx(e, http.MethodGet, "/citizen/services")
	setCitizenUser(c, "u1")

	if err := h.ShowServiceCatalog(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// --- WithActivityLogger / WithNotificationService ---

func TestCitizenWebHandler_WithActivityLogger_Returns(t *testing.T) {
	h := NewCitizenWebHandler(&fakeAuthService{})
	logger := &fakeActivityLogger{}
	result := h.WithActivityLogger(logger)
	if result == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestCitizenWebHandler_WithNotificationService_Returns(t *testing.T) {
	h := NewCitizenWebHandler(&fakeAuthService{})
	result := h.WithNotificationService(&fakeNotifSvc{})
	if result == nil {
		t.Fatal("expected non-nil handler")
	}
}

// --- isAllDigits ---

func TestIsAllDigits_True(t *testing.T) {
	if !isAllDigits("12345") {
		t.Fatal("expected true for all digits")
	}
}

func TestIsAllDigits_False(t *testing.T) {
	if isAllDigits("123abc") {
		t.Fatal("expected false for non-digits")
	}
}

func TestIsAllDigits_Empty(t *testing.T) {
	if isAllDigits("") {
		t.Fatal("expected false for empty string")
	}
}

// --- ShowApplyForm tests ---

func TestShowApplyForm_NilServices(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{})

	c, _ := newCitizenCtx(e, http.MethodGet, "/citizen/applications/new")
	setCitizenUser(c, "u1")

	err := h.ShowApplyForm(c)
	var he *echo.HTTPError
	if !errors.As(err, &he) || he.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 HTTPError, got %v", err)
	}
}

func TestShowApplyForm_NoClaims(t *testing.T) {
	e := newAdminEcho()
	catalog := &fakeCatalogSvc{}
	appSvc := &fakeAppWebSvc{}
	h := NewCitizenWebHandler(&fakeAuthService{}).WithCatalogService(catalog).WithApplicationService(appSvc)

	c, _ := newCitizenCtx(e, http.MethodGet, "/citizen/applications/new")
	// no user set

	err := h.ShowApplyForm(c)
	var he *echo.HTTPError
	if !errors.As(err, &he) || he.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 HTTPError, got %v", err)
	}
}

func TestShowApplyForm_NoServiceTypeID(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	catalog := &fakeCatalogSvc{}
	appSvc := &fakeAppWebSvc{}
	h := NewCitizenWebHandler(&fakeAuthService{}).WithCatalogService(catalog).WithApplicationService(appSvc)

	c, rec := newCitizenCtx(e, http.MethodGet, "/citizen/applications/new")
	setCitizenUser(c, "u1")

	err := h.ShowApplyForm(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
}

func TestShowApplyForm_ServiceTypeNotFound(t *testing.T) {
	e := newAdminEcho()
	catalog := &fakeCatalogSvc{
		getByIDFn: func(_ context.Context, id string) (*models.ServiceType, error) {
			return nil, errors.New("not found")
		},
	}
	appSvc := &fakeAppWebSvc{}
	h := NewCitizenWebHandler(&fakeAuthService{}).WithCatalogService(catalog).WithApplicationService(appSvc)

	req := httptest.NewRequest(http.MethodGet, "/citizen/applications/new?service_type_id=st-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setCitizenUser(c, "u1")

	err := h.ShowApplyForm(c)
	var he *echo.HTTPError
	if !errors.As(err, &he) || he.Code != http.StatusNotFound {
		t.Fatalf("expected 404 HTTPError, got %v", err)
	}
}

func TestShowApplyForm_OK(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	catalog := &fakeCatalogSvc{}
	appSvc := &fakeAppWebSvc{}
	h := NewCitizenWebHandler(&fakeAuthService{}).WithCatalogService(catalog).WithApplicationService(appSvc)

	req := httptest.NewRequest(http.MethodGet, "/citizen/applications/new?service_type_id=st-1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setCitizenUser(c, "u1")

	if err := h.ShowApplyForm(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

// --- UploadApplicationSupplements tests ---

func TestUploadApplicationSupplements_NoClaims(t *testing.T) {
	e := newAdminEcho()
	appSvc := &fakeAppWebSvc{}
	h := NewCitizenWebHandler(&fakeAuthService{}).WithApplicationService(appSvc)

	c, _ := newCitizenCtx(e, http.MethodPost, "/citizen/applications/app1/supplements")
	// no user

	err := h.UploadApplicationSupplements(c)
	var he *echo.HTTPError
	if !errors.As(err, &he) || he.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %v", err)
	}
}

func TestUploadApplicationSupplements_NilAppSvc(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{})

	c, _ := newCitizenCtx(e, http.MethodPost, "/citizen/applications/app1/supplements")
	setCitizenUser(c, "u1")

	err := h.UploadApplicationSupplements(c)
	var he *echo.HTTPError
	if !errors.As(err, &he) || he.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %v", err)
	}
}

func TestUploadApplicationSupplements_NotFound(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	appSvc := &fakeAppWebSvc{
		supplementFn: func(_, _ string, _ []*multipart.FileHeader) ([]dtos.ApplicationAttachmentResponse, error) {
			return nil, services.ErrApplicationNotFound
		},
	}
	h := NewCitizenWebHandler(&fakeAuthService{}).WithApplicationService(appSvc)

	body := &strings.Builder{}
	body.WriteString("--boundary\r\nContent-Disposition: form-data; name=\"attachments[]\"; filename=\"a.pdf\"\r\n\r\ndata\r\n--boundary--\r\n")
	req := httptest.NewRequest(http.MethodPost, "/citizen/applications/app1/supplements", strings.NewReader(body.String()))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=boundary")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "app1"}})
	setCitizenUser(c, "u1")

	err := h.UploadApplicationSupplements(c)
	var he *echo.HTTPError
	if !errors.As(err, &he) || he.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %v", err)
	}
}

func TestUploadApplicationSupplements_Success(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	appSvc := &fakeAppWebSvc{}
	h := NewCitizenWebHandler(&fakeAuthService{}).WithApplicationService(appSvc)

	body := &strings.Builder{}
	body.WriteString("--boundary\r\nContent-Disposition: form-data; name=\"attachments[]\"; filename=\"a.pdf\"\r\n\r\ndata\r\n--boundary--\r\n")
	req := httptest.NewRequest(http.MethodPost, "/citizen/applications/app1/supplements", strings.NewReader(body.String()))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=boundary")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "app1"}})
	setCitizenUser(c, "u1")

	if err := h.UploadApplicationSupplements(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
}

func TestUploadApplicationSupplements_ServiceErrors(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	cases := []struct {
		name string
		err  error
	}{
		{"supplement_not_allowed", services.ErrSupplementNotAllowed},
		{"too_many_attachments", services.ErrTooManyAttachments},
		{"attachment_too_large", services.ErrAttachmentTooLarge},
		{"internal", errors.New("internal")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := newAdminEcho()
			appSvc := &fakeAppWebSvc{
				supplementFn: func(_, _ string, _ []*multipart.FileHeader) ([]dtos.ApplicationAttachmentResponse, error) {
					return nil, tc.err
				},
			}
			h := NewCitizenWebHandler(&fakeAuthService{}).WithApplicationService(appSvc)

			body := &strings.Builder{}
			body.WriteString("--boundary\r\nContent-Disposition: form-data; name=\"attachments[]\"; filename=\"a.pdf\"\r\n\r\ndata\r\n--boundary--\r\n")
			req := httptest.NewRequest(http.MethodPost, "/citizen/applications/app1/supplements", strings.NewReader(body.String()))
			req.Header.Set("Content-Type", "multipart/form-data; boundary=boundary")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPathValues(echo.PathValues{{Name: "id", Value: "app1"}})
			setCitizenUser(c, "u1")

			err := h.UploadApplicationSupplements(c)
			if err != nil {
				t.Fatalf("expected nil error (redirect), got %v", err)
			}
			if rec.Code != http.StatusSeeOther {
				t.Fatalf("expected redirect, got %d", rec.Code)
			}
		})
	}
}

// --- WebRegister validation branches ---

func TestWebRegister_PasswordMismatch(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{})
	form := "name=User&email=a%40b.com&citizen_id_number=123456789012&password=pass123&confirm_password=different"
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.WebRegister(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
}

func TestWebRegister_InvalidCCCD(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{})
	form := "name=User&email=a%40b.com&citizen_id_number=123abc&password=pass123&confirm_password=pass123"
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.WebRegister(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
}

func TestWebRegister_PasswordTooShort(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{})
	form := "name=User&email=a%40b.com&citizen_id_number=123456789012&password=abc&confirm_password=abc"
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.WebRegister(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
}

func TestWebRegister_EmailExists(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{
		registerFn: func(_ *dtos.RegisterRequest) (*models.User, error) {
			return nil, services.ErrEmailAlreadyExists
		},
	})
	form := "name=User&email=exists%40b.com&citizen_id_number=123456789012&password=pass123&confirm_password=pass123"
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.WebRegister(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", rec.Code)
	}
}

func TestWebRegister_WithActivityLogger(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	logger := &fakeActivityLogger{}
	h := NewCitizenWebHandler(&fakeAuthService{
		registerFn: func(_ *dtos.RegisterRequest) (*models.User, error) {
			return &models.User{ID: "u1"}, nil
		},
	}).WithActivityLogger(logger)
	form := "name=User&email=new%40b.com&citizen_id_number=123456789012&password=pass123&confirm_password=pass123"
	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if err := h.WebRegister(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
	if logger.calls != 1 {
		t.Fatalf("expected 1 log call, got %d", logger.calls)
	}
}

// --- ShowApplicationDetail missing branches ---

func TestShowApplicationDetail_NoClaims(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{}).WithApplicationService(&fakeAppWebSvc{})
	c, _ := newCitizenCtx(e, http.MethodGet, "/citizen/applications/app1")
	// no user set
	err := h.ShowApplicationDetail(c)
	var he *echo.HTTPError
	if !errors.As(err, &he) || he.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %v", err)
	}
}

func TestShowApplicationDetail_NilSvc(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{})
	c, _ := newCitizenCtx(e, http.MethodGet, "/citizen/applications/app1")
	setCitizenUser(c, "u1")
	err := h.ShowApplicationDetail(c)
	var he *echo.HTTPError
	if !errors.As(err, &he) || he.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %v", err)
	}
}

func TestShowApplicationDetail_InternalError(t *testing.T) {
	e := newAdminEcho()
	appSvc := &fakeAppWebSvc{
		getFn: func(_, _ string) (*dtos.ApplicationResponse, error) {
			return nil, errors.New("db error")
		},
	}
	h := NewCitizenWebHandler(&fakeAuthService{}).WithApplicationService(appSvc)
	c, _ := newCitizenCtx(e, http.MethodGet, "/citizen/applications/app1")
	setCitizenUser(c, "u1")
	setPathParam(c, "id", "app1")
	err := h.ShowApplicationDetail(c)
	var he *echo.HTTPError
	if !errors.As(err, &he) || he.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %v", err)
	}
}

// --- SubmitApplication missing branches ---

func TestSubmitApplication_NoClaims(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{}).
		WithCatalogService(&fakeCatalogSvc{}).
		WithApplicationService(&fakeAppWebSvc{})
	c, _ := newCitizenCtx(e, http.MethodPost, "/citizen/applications")
	// no user
	err := h.SubmitApplication(c)
	var he *echo.HTTPError
	if !errors.As(err, &he) || he.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %v", err)
	}
}

func TestSubmitApplication_NilServices(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{})
	c, _ := newCitizenCtx(e, http.MethodPost, "/citizen/applications")
	setCitizenUser(c, "u1")
	err := h.SubmitApplication(c)
	var he *echo.HTTPError
	if !errors.As(err, &he) || he.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %v", err)
	}
}

func TestSubmitApplication_EmptyServiceTypeID(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{}).
		WithCatalogService(&fakeCatalogSvc{}).
		WithApplicationService(&fakeAppWebSvc{})
	form := "service_type_id="
	req := httptest.NewRequest(http.MethodPost, "/citizen/applications", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setCitizenUser(c, "u1")
	if err := h.SubmitApplication(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
}

func TestSubmitApplication_ServiceTypeNotFound(t *testing.T) {
	e := newAdminEcho()
	h := NewCitizenWebHandler(&fakeAuthService{}).
		WithCatalogService(&fakeCatalogSvc{
			getByIDFn: func(_ context.Context, _ string) (*models.ServiceType, error) {
				return nil, errors.New("not found")
			},
		}).
		WithApplicationService(&fakeAppWebSvc{})
	form := "service_type_id=st-1"
	req := httptest.NewRequest(http.MethodPost, "/citizen/applications", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	setCitizenUser(c, "u1")
	err := h.SubmitApplication(c)
	var he *echo.HTTPError
	if !errors.As(err, &he) || he.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %v", err)
	}
}

func TestSubmitApplication_ServiceErrors(t *testing.T) {
	_ = configs.LoadI18nMessages("../../locales")
	cases := []struct {
		name string
		err  error
	}{
		{"missing_required_field", services.ErrMissingRequiredField},
		{"attachment_required", services.ErrAttachmentRequired},
		{"too_many_attachments", services.ErrTooManyAttachments},
		{"attachment_too_large", services.ErrAttachmentTooLarge},
		{"attachment_invalid_type", services.ErrAttachmentInvalidType},
		{"internal_error", errors.New("db error")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := newAdminEcho()
			h := NewCitizenWebHandler(&fakeAuthService{}).
				WithCatalogService(&fakeCatalogSvc{}).
				WithApplicationService(&fakeAppWebSvc{
					submitFn: func(_ string, _ *dtos.SubmitApplicationRequest, _ []*multipart.FileHeader) (*dtos.ApplicationResponse, error) {
						return nil, tc.err
					},
				})
			form := "service_type_id=st-1"
			req := httptest.NewRequest(http.MethodPost, "/citizen/applications", strings.NewReader(form))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			setCitizenUser(c, "u1")
			if err := h.SubmitApplication(c); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if rec.Code != http.StatusUnprocessableEntity {
				t.Fatalf("expected 422, got %d", rec.Code)
			}
		})
	}
}

// --- parseFormSchema ---

func TestParseFormSchema_EmptyRaw(t *testing.T) {
	result := parseFormSchema(nil)
	if result.Required == nil {
		t.Error("expected non-nil Required map")
	}
	if len(result.Fields) != 0 {
		t.Errorf("expected empty Fields, got %v", result.Fields)
	}
}

func TestParseFormSchema_InvalidJSON(t *testing.T) {
	result := parseFormSchema([]byte("{invalid"))
	if result.Required == nil {
		t.Error("expected non-nil Required map")
	}
}

// --- SetLocale with same-host referer ---

func TestAdminAuthHandlerSetLocale_SameHostReferer(t *testing.T) {
	e := newTestEcho()
	handler := NewAdminAuthHandler(&fakeAuthService{})

	req := httptest.NewRequest(http.MethodGet, "/admin/locale?lang=vi", nil)
	req.Host = "example.com"
	req.Header.Set("Referer", "http://example.com/admin/dashboard?tab=1")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler.SetLocale(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
	loc := rec.Header().Get(echo.HeaderLocation)
	if loc != "/admin/dashboard?tab=1" {
		t.Fatalf("expected same-host redirect, got %q", loc)
	}
}

func TestAdminAuthHandlerSetLocale_RelativeReferer(t *testing.T) {
	e := newTestEcho()
	handler := NewAdminAuthHandler(&fakeAuthService{})

	req := httptest.NewRequest(http.MethodGet, "/admin/locale?lang=en", nil)
	req.Header.Set("Referer", "/admin/users")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := handler.SetLocale(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", rec.Code)
	}
	loc := rec.Header().Get(echo.HeaderLocation)
	if loc != "/admin/users" {
		t.Fatalf("expected relative path redirect, got %q", loc)
	}
}
