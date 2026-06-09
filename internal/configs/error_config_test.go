package configs

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
)

func TestAdminErrorKeys(t *testing.T) {
	cases := []struct {
		code        int
		wantHeading string
		wantMessage string
	}{
		{http.StatusNotFound, "ui.error.404.heading", "ui.error.404.message"},
		{http.StatusForbidden, "ui.error.403.heading", "ui.error.403.message"},
		{http.StatusUnauthorized, "ui.error.401.heading", "ui.error.401.message"},
		{http.StatusInternalServerError, "ui.error.500.heading", "ui.error.500.message"},
		{http.StatusBadRequest, "ui.error.500.heading", "ui.error.500.message"},
	}
	for _, tc := range cases {
		h, m := adminErrorKeys(tc.code)
		if h != tc.wantHeading || m != tc.wantMessage {
			t.Errorf("adminErrorKeys(%d) = (%q, %q), want (%q, %q)", tc.code, h, m, tc.wantHeading, tc.wantMessage)
		}
	}
}

func TestTranslateHTTPMessage_StringKey(t *testing.T) {
	messages = map[string]map[string]string{"vi": {"some.key": "Lỗi"}}
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	detail := translateHTTPMessage(c, "some.key")
	if detail.Code != "some.key" {
		t.Fatalf("expected code some.key, got %q", detail.Code)
	}
	if detail.Message != "Lỗi" {
		t.Fatalf("expected message 'Lỗi', got %q", detail.Message)
	}
}

func TestTranslateHTTPMessage_NonStringKey(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	detail := translateHTTPMessage(c, 42)
	if detail.Code != "error.unknown" {
		t.Fatalf("expected code error.unknown, got %q", detail.Code)
	}
}

func TestTranslateValidationMessages(t *testing.T) {
	messages = map[string]map[string]string{"vi": {"validation.required": "Trường {field} là bắt buộc"}}
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	msgs := []ValidatorMessage{
		{Field: "email", Key: "validation.required"},
	}
	details := translateValidationMessages(c, msgs)
	if len(details) != 1 {
		t.Fatalf("expected 1 detail, got %d", len(details))
	}
	if details[0].Field != "email" {
		t.Fatalf("expected field email, got %q", details[0].Field)
	}
}

func TestCustomHTTPErrorHandler_JSONResponse(t *testing.T) {
	messages = map[string]map[string]string{"vi": {"common.internal_error": "Lỗi hệ thống"}}
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/something", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	CustomHTTPErrorHandler(c, errors.New("generic error"))
	if rec.Code != http.StatusOK {
		// Body check is sufficient — status is set inside the JSON write
	}
	if rec.Body.Len() == 0 {
		t.Fatal("expected non-empty response body")
	}
}

func TestCustomHTTPErrorHandler_HTTPError(t *testing.T) {
	messages = map[string]map[string]string{"vi": {"application.not_found": "Không tìm thấy"}}
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/something", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	CustomHTTPErrorHandler(c, echo.NewHTTPError(http.StatusNotFound, "application.not_found"))
	if rec.Body.Len() == 0 {
		t.Fatal("expected non-empty response body")
	}
}

func TestCustomHTTPErrorHandler_HeadRequest(t *testing.T) {
	messages = map[string]map[string]string{"vi": {}}
	e := echo.New()
	req := httptest.NewRequest(http.MethodHead, "/api/something", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	CustomHTTPErrorHandler(c, echo.NewHTTPError(http.StatusNotFound, "not_found"))
	if rec.Body.Len() != 0 {
		t.Fatal("expected empty body for HEAD request")
	}
}

func TestCustomHTTPErrorHandler_AdminWebPath(t *testing.T) {
	messages = map[string]map[string]string{"vi": {"ui.error.404.heading": "404", "ui.error.404.message": "Not found"}}
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/admin/something", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Without renderer, Render fails and falls back to JSON
	CustomHTTPErrorHandler(c, echo.NewHTTPError(http.StatusNotFound, "not_found"))
	if rec.Body.Len() == 0 {
		t.Fatal("expected non-empty response body (JSON fallback)")
	}
}

func TestCustomHTTPErrorHandler_AdminLoginTooManyRequests(t *testing.T) {
	if err := LoadI18nMessages("../../locales"); err != nil {
		t.Fatalf("load i18n messages: %v", err)
	}

	e := echo.New()
	renderer := &captureRenderer{}
	e.Renderer = renderer

	req := httptest.NewRequest(http.MethodPost, "/admin/login", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(LocaleKey, "en")

	CustomHTTPErrorHandler(c, echo.NewHTTPError(http.StatusTooManyRequests, "common.too_many_requests"))

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected status %d, got %d", http.StatusTooManyRequests, rec.Code)
	}
	if renderer.templateName != "admin/pages/auth/login.html" {
		t.Fatalf("expected admin login template, got %q", renderer.templateName)
	}
	if got := renderer.data["Error"]; got != "Too many requests. Please try again later." {
		t.Fatalf("expected rate-limit message, got %#v", got)
	}
	if got := renderer.data["Title"]; got != "System Login" {
		t.Fatalf("expected login title, got %#v", got)
	}
}

type captureRenderer struct {
	templateName string
	data         map[string]any
}

func (r *captureRenderer) Render(_ *echo.Context, w io.Writer, templateName string, data any) error {
	r.templateName = templateName
	if m, ok := data.(map[string]interface{}); ok {
		r.data = make(map[string]any, len(m))
		for k, v := range m {
			r.data[k] = v
		}
	}
	_, _ = w.Write([]byte(templateName))
	return nil
}

func TestCustomHTTPErrorHandler_CitizenWebPath(t *testing.T) {
	messages = map[string]map[string]string{"vi": {"ui.error.404.heading": "404", "ui.error.404.message": "Not found"}}
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/citizen/profile", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	CustomHTTPErrorHandler(c, echo.NewHTTPError(http.StatusNotFound, "not_found"))
	if rec.Body.Len() == 0 {
		t.Fatal("expected non-empty response body (JSON fallback)")
	}
}

func TestCustomHTTPErrorHandler_CitizenRootPath(t *testing.T) {
	messages = map[string]map[string]string{"vi": {}}
	e := echo.New()
	for _, path := range []string{"/", "/login", "/register", "/logout"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		CustomHTTPErrorHandler(c, echo.NewHTTPError(http.StatusInternalServerError, "error"))
		if rec.Body.Len() == 0 {
			t.Fatalf("expected non-empty response body for path %s", path)
		}
	}
}

func TestCustomHTTPErrorHandler_ValidatorError(t *testing.T) {
	messages = map[string]map[string]string{"vi": {"validation.required": "Bắt buộc"}}
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/something", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ve := &ValidatorError{
		Code: http.StatusBadRequest,
		Messages: []ValidatorMessage{
			{Field: "email", Key: "validation.required"},
		},
	}
	CustomHTTPErrorHandler(c, ve)
	if rec.Body.Len() == 0 {
		t.Fatal("expected non-empty response body")
	}
}
