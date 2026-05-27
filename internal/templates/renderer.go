package templates

import (
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/labstack/echo/v5"
)

type TemplateRenderer struct {
	templates *template.Template
	raw       map[string]string
}

func NewRenderer(root string) (*TemplateRenderer, error) {
	t := template.New("").Funcs(template.FuncMap{
		"t": func(key string, params ...map[string]string) string {
			return key
		},
		"currentLocale": func() string {
			return "vi"
		},
		// provide a simple default function: use pipeline value if non-empty, otherwise use provided default
		"default": func(v interface{}, d string) string {
			if v == nil {
				return d
			}
			s, ok := v.(string)
			if !ok {
				return d
			}
			if strings.TrimSpace(s) == "" {
				return d
			}
			return s
		},
		// stub for fmtTime so templates can be parsed before real func is injected
		"fmtTime": func(v interface{}, layout string) string { return "" },
		"concat":  func(a, b string) string { return a + b },
	})
	raw := map[string]string{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".html" {
			relPath, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			templateName := filepath.ToSlash(relPath)
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			s := strings.TrimSpace(string(content))
			raw[templateName] = s
			if _, err := t.New(templateName).Parse(s); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &TemplateRenderer{templates: t, raw: raw}, nil
}

func (r *TemplateRenderer) Render(c *echo.Context, w io.Writer, name string, data any) error {
	funcMap := template.FuncMap{
		"t": func(key string, params ...map[string]string) string {
			paramMap := map[string]string{}
			if len(params) > 0 {
				paramMap = params[0]
			}
			return configs.T(c, key, paramMap)
		},
		"currentLocale": func() string {
			return configs.LocaleFromContext(c)
		},
		"fmtTime": func(v interface{}, layout string) string {
			if v == nil {
				return ""
			}
			switch t := v.(type) {
			case time.Time:
				if t.IsZero() {
					return ""
				}
				return t.Format(layout)
			case *time.Time:
				if t == nil || t.IsZero() {
					return ""
				}
				return t.Format(layout)
			case string:
				s := strings.TrimSpace(t)
				if s == "" {
					return ""
				}
				// try parse common RFC formats
				if tm, err := time.Parse(time.RFC3339Nano, s); err == nil {
					return tm.Format(layout)
				}
				if tm, err := time.Parse(time.RFC3339, s); err == nil {
					return tm.Format(layout)
				}
				// fallback: if looks like 'YYYY-MM-DD HH:MM:SS...' just take prefix
				if len(s) >= 16 {
					return s[:16]
				}
				return s
			default:
				return ""
			}
		},
	}

	cloned, err := r.templates.Funcs(funcMap).Clone()
	if err != nil {
		return err
	}

	// Re-parse the requested template last so its block definitions (e.g., "content")
	// override any other parsed definitions. This avoids cross-page block collisions.
	if pageContent, ok := r.raw[name]; ok {
		if _, err := cloned.New(name).Parse(pageContent); err != nil {
			return err
		}
	}

	return cloned.ExecuteTemplate(w, name, data)
}
