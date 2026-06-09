package configs

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
)

type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func CustomHTTPErrorHandler(c *echo.Context, err error) {
	if resp, uErr := echo.UnwrapResponse(c.Response()); uErr == nil {
		if resp.Committed {
			return
		}
	}

	code := http.StatusInternalServerError
	var errorDetails []ErrorDetail

	var ve *ValidatorError
	if errors.As(err, &ve) {
		code = ve.Code
		errorDetails = translateValidationMessages(c, ve.Messages)
	} else if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		errorDetails = []ErrorDetail{translateHTTPMessage(c, he.Message)}
	} else {
		errorDetails = []ErrorDetail{
			{
				Code:    "common.internal_error",
				Message: T(c, "common.internal_error", nil),
			},
		}
	}

	path := c.Request().URL.Path
	isAdminWeb := strings.HasPrefix(path, "/admin") && !strings.HasPrefix(path, "/admin/api")
	isCitizenWeb := strings.HasPrefix(path, "/citizen") ||
		path == "/" || path == "/login" || path == "/register" || path == "/logout"

	if c.Request().Method == http.MethodHead {
		c.NoContent(code)
		return
	}

	if code == http.StatusTooManyRequests {
		if templateName, data, ok := authTooManyRequestsPage(c); ok {
			if renderErr := c.Render(code, templateName, data); renderErr != nil {
				c.JSON(code, map[string]interface{}{"errors": errorDetails, "code": code})
			}
			return
		}
	}

	if isAdminWeb {
		headingKey, messageKey := adminErrorKeys(code)
		data := map[string]interface{}{
			"Title":   T(c, headingKey, nil),
			"Code":    code,
			"Heading": T(c, headingKey, nil),
			"Message": T(c, messageKey, nil),
		}
		if renderErr := c.Render(code, "admin/pages/error.html", data); renderErr != nil {
			c.JSON(code, map[string]interface{}{"errors": errorDetails, "code": code})
		}
		return
	}

	if isCitizenWeb {
		headingKey, messageKey := adminErrorKeys(code)
		data := map[string]interface{}{
			"Title":    T(c, headingKey, nil),
			"Code":     code,
			"Heading":  T(c, headingKey, nil),
			"Message":  T(c, messageKey, nil),
			"FullPage": true,
		}
		if renderErr := c.Render(code, "citizen/pages/error.html", data); renderErr != nil {
			c.JSON(code, map[string]interface{}{"errors": errorDetails, "code": code})
		}
		return
	}

	c.JSON(code, map[string]interface{}{"errors": errorDetails, "code": code})
}

func authTooManyRequestsPage(c *echo.Context) (string, map[string]interface{}, bool) {
	path := c.Request().URL.Path
	message := T(c, "common.too_many_requests", nil)

	switch {
	case path == "/admin/login":
		return "admin/pages/auth/login.html", map[string]interface{}{
			"Title":    T(c, "ui.form.admin_login", nil),
			"Error":    message,
			"FullPage": true,
		}, true
	case path == "/login":
		return "citizen/pages/auth/login.html", map[string]interface{}{
			"Title": T(c, "ui.form.citizen_login", nil),
			"Error": message,
		}, true
	case path == "/register":
		return "citizen/pages/auth/register.html", map[string]interface{}{
			"Title": T(c, "ui.form.citizen_register", nil),
			"Error": message,
		}, true
	default:
		return "", nil, false
	}
}

func adminErrorKeys(code int) (headingKey, messageKey string) {
	switch code {
	case http.StatusNotFound:
		return "ui.error.404.heading", "ui.error.404.message"
	case http.StatusForbidden:
		return "ui.error.403.heading", "ui.error.403.message"
	case http.StatusUnauthorized:
		return "ui.error.401.heading", "ui.error.401.message"
	default:
		return "ui.error.500.heading", "ui.error.500.message"
	}
}

func translateHTTPMessage(c *echo.Context, message interface{}) ErrorDetail {
	key, ok := message.(string)
	if !ok {
		return ErrorDetail{
			Code:    "error.unknown",
			Message: http.StatusText(http.StatusInternalServerError),
		}
	}

	return ErrorDetail{
		Code:    key,
		Message: T(c, key, nil),
	}
}

func translateValidationMessages(c *echo.Context, messages []ValidatorMessage) []ErrorDetail {
	translatedMessages := make([]ErrorDetail, 0, len(messages))
	for _, message := range messages {
		params := make(map[string]string, len(message.Params)+1)
		for k, v := range message.Params {
			params[k] = v
		}
		params["field"] = message.Field
		translatedMessages = append(translatedMessages, ErrorDetail{
			Field:   message.Field,
			Code:    message.Key,
			Message: T(c, message.Key, params),
		})
	}

	return translatedMessages
}
