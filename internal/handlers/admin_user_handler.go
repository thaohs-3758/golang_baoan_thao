package handlers

import (
	"encoding/csv"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	"github.com/awesome-academy/golang_baoan_thao/internal/utils"
	"github.com/labstack/echo/v5"
)

type AdminUserService interface {
	ListUsers(filter repositories.UserFilter, page, limit int) ([]models.User, int64, error)
	GetUser(id string) (*models.User, error)
	CreateUser(req *dtos.AdminCreateUserRequest, createdBy string) (*models.User, error)
	UpdateUser(id string, req *dtos.AdminUpdateUserRequest, updatedBy string) (*models.User, error)
	BlockUser(id string, updatedBy string) error
	UnblockUser(id string, updatedBy string) error
	DeleteUser(id string, deletedBy string) error
}

type StaffImportExportService interface {
	ImportStaff(rows []services.StaffImportRow, createdBy string) []string
}

type AdminCitizenProfileLookup interface {
	GetByUserID(userID string) (*models.CitizenProfile, error)
}

type AdminUserHandler struct {
	svc                AdminUserService
	importSvc          StaffImportExportService
	deptRepo           repositories.DepartmentRepository
	staffRepo          repositories.StaffProfileRepository
	citizenProfileRepo AdminCitizenProfileLookup
}

func NewAdminUserHandler(svc AdminUserService) *AdminUserHandler {
	return &AdminUserHandler{svc: svc}
}

func (h *AdminUserHandler) WithImportExport(svc StaffImportExportService) *AdminUserHandler {
	h.importSvc = svc
	return h
}

func (h *AdminUserHandler) WithDeptAndStaffRepos(deptRepo repositories.DepartmentRepository, staffRepo repositories.StaffProfileRepository) *AdminUserHandler {
	h.deptRepo = deptRepo
	h.staffRepo = staffRepo
	return h
}

func (h *AdminUserHandler) WithCitizenProfileRepo(profileRepo AdminCitizenProfileLookup) *AdminUserHandler {
	h.citizenProfileRepo = profileRepo
	return h
}

func adminCurrentUser(c *echo.Context) *configs.JwtCustomClaims {
	v := c.Get("user")
	if v == nil {
		return nil
	}
	claims, _ := v.(*configs.JwtCustomClaims)
	return claims
}

func adminFlashURL(flash, msg string) string {
	return "/admin/users?" + url.Values{"flash": {flash}, "msg": {msg}}.Encode()
}

func actorID(c *echo.Context) string {
	if u := adminCurrentUser(c); u != nil {
		return u.ID
	}
	return ""
}

func extractFieldErrors(c *echo.Context, err error) map[string]string {
	var ve *configs.ValidatorError
	if !errors.As(err, &ve) {
		return nil
	}
	out := make(map[string]string, len(ve.Messages))
	for _, m := range ve.Messages {
		params := make(map[string]string, len(m.Params)+1)
		for k, v := range m.Params {
			params[k] = v
		}
		params["field"] = m.Field
		out[m.Field] = configs.T(c, m.Key, params)
	}
	return out
}

func (h *AdminUserHandler) ListUsers(c *echo.Context) error {
	search := c.QueryParam("search")
	role := c.QueryParam("role")
	page, limit := parsePagination(c)

	filter := repositories.UserFilter{Search: search, Role: role}
	users, total, err := h.svc.ListUsers(filter, page, limit)
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"Title":          configs.T(c, "ui.users.title", nil),
		"CurrentPath":    "/admin/users",
		"CurrentUser":    adminCurrentUser(c),
		"Users":          users,
		"Pagination":     utils.NewPagination(page, limit, total),
		"Search":         search,
		"RoleFilter":     role,
		"ImportAction":   "/admin/users/import",
		"TemplateAction": "/admin/users/template",
		"Flash":          flashFromQuery(c),
	}
	return c.Render(http.StatusOK, "admin/pages/users/list.html", data)
}

func (h *AdminUserHandler) ShowCreateForm(c *echo.Context) error {
	data := map[string]interface{}{
		"Title":       configs.T(c, "ui.users.form.create_title", nil),
		"CurrentPath": "/admin/users",
		"CurrentUser": adminCurrentUser(c),
		"IsEdit":      false,
	}
	return c.Render(http.StatusOK, "admin/pages/users/form.html", data)
}

func (h *AdminUserHandler) CreateUser(c *echo.Context) error {
	req := new(dtos.AdminCreateUserRequest)
	if err := c.Bind(req); err != nil {
		return h.renderFormErrors(c, false, nil, req, nil, configs.T(c, "common.invalid_data", nil))
	}
	if err := c.Validate(req); err != nil {
		return h.renderFormErrors(c, false, nil, req, extractFieldErrors(c, err), "")
	}

	_, err := h.svc.CreateUser(req, actorID(c))
	if err != nil {
		msg := configs.T(c, "common.internal_error", nil)
		if errors.Is(err, services.ErrEmailAlreadyExists) {
			msg = configs.T(c, "auth.email_exists", nil)
		}
		return h.renderFormErrors(c, false, nil, req, nil, msg)
	}

	return c.Redirect(http.StatusSeeOther, adminFlashURL("success", configs.T(c, "ui.msg.user_created", nil)))
}

func (h *AdminUserHandler) ShowUser(c *echo.Context) error {
	id := c.Param("id")
	user, err := h.svc.GetUser(id)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, adminFlashURL("error", configs.T(c, "auth.user_not_found", nil)))
	}

	var citizenProfile *models.CitizenProfile
	if user != nil && user.Role == models.UserRoleCitizen && h.citizenProfileRepo != nil {
		citizenProfile, _ = h.citizenProfileRepo.GetByUserID(user.ID)
	}

	data := map[string]interface{}{
		"Title":          configs.T(c, "ui.users.detail.title", nil),
		"CurrentPath":    "/admin/users",
		"CurrentUser":    adminCurrentUser(c),
		"User":           user,
		"CitizenProfile": citizenProfile,
	}
	return c.Render(http.StatusOK, "admin/pages/users/detail.html", data)
}

func (h *AdminUserHandler) ShowEditForm(c *echo.Context) error {
	id := c.Param("id")
	user, err := h.svc.GetUser(id)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, adminFlashURL("error", configs.T(c, "auth.user_not_found", nil)))
	}

	var depts []models.Department
	if h.deptRepo != nil {
		depts, _, _ = h.deptRepo.List(c.Request().Context(), repositories.DepartmentFilter{}, 0, 1000)
	}

	currentDeptID := ""
	if h.staffRepo != nil {
		if sp, _ := h.staffRepo.FindByUserID(id); sp != nil && sp.DepartmentID != nil {
			currentDeptID = *sp.DepartmentID
		}
	}

	data := map[string]interface{}{
		"Title":         configs.T(c, "ui.users.form.edit_title", nil),
		"CurrentPath":   "/admin/users",
		"CurrentUser":   adminCurrentUser(c),
		"IsEdit":        true,
		"User":          user,
		"Departments":   depts,
		"CurrentDeptID": currentDeptID,
		"Flash":         flashFromQuery(c),
	}
	return c.Render(http.StatusOK, "admin/pages/users/form.html", data)
}

func (h *AdminUserHandler) UpdateUser(c *echo.Context) error {
	id := c.Param("id")

	user, err := h.svc.GetUser(id)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, adminFlashURL("error", configs.T(c, "auth.user_not_found", nil)))
	}

	req := new(dtos.AdminUpdateUserRequest)
	if err := c.Bind(req); err != nil {
		return h.renderFormErrors(c, true, user, req, nil, configs.T(c, "common.invalid_data", nil))
	}
	if err := c.Validate(req); err != nil {
		return h.renderFormErrors(c, true, user, req, extractFieldErrors(c, err), "")
	}

	if _, err := h.svc.UpdateUser(id, req, actorID(c)); err != nil {
		return h.renderFormErrors(c, true, user, req, nil, configs.T(c, "common.internal_error", nil))
	}

	if h.staffRepo != nil {
		var deptID *string
		if req.DepartmentID != "" {
			deptID = &req.DepartmentID
		}
		sp, _ := h.staffRepo.FindByUserID(id)
		if sp == nil {
			now := time.Now()
			_, _ = h.staffRepo.Create(&models.StaffProfile{
				UserID:       id,
				DepartmentID: deptID,
				CreatedAt:    now,
				UpdatedAt:    now,
			})
		} else {
			_ = h.staffRepo.UpdateDepartment(id, deptID, actorID(c))
		}
	}

	return c.Redirect(http.StatusSeeOther, adminFlashURL("success", configs.T(c, "ui.msg.user_updated", nil)))
}

func (h *AdminUserHandler) BlockUser(c *echo.Context) error {
	id := c.Param("id")
	if err := h.svc.BlockUser(id, actorID(c)); err != nil {
		return c.Redirect(http.StatusSeeOther, adminFlashURL("error", configs.T(c, "ui.msg.block_failed", nil)))
	}
	return c.Redirect(http.StatusSeeOther, adminFlashURL("success", configs.T(c, "ui.msg.user_blocked", nil)))
}

func (h *AdminUserHandler) UnblockUser(c *echo.Context) error {
	id := c.Param("id")
	if err := h.svc.UnblockUser(id, actorID(c)); err != nil {
		return c.Redirect(http.StatusSeeOther, adminFlashURL("error", configs.T(c, "ui.msg.unblock_failed", nil)))
	}
	return c.Redirect(http.StatusSeeOther, adminFlashURL("success", configs.T(c, "ui.msg.user_unblocked", nil)))
}

func (h *AdminUserHandler) DeleteUser(c *echo.Context) error {
	id := c.Param("id")
	if err := h.svc.DeleteUser(id, actorID(c)); err != nil {
		return c.Redirect(http.StatusSeeOther, adminFlashURL("error", configs.T(c, "ui.msg.delete_failed", nil)))
	}
	return c.Redirect(http.StatusSeeOther, adminFlashURL("success", configs.T(c, "ui.msg.user_deleted", nil)))
}

// DownloadTemplate handles GET /admin/users/template
func (h *AdminUserHandler) DownloadTemplate(c *echo.Context) error {
	return writeXLSXTemplate(c, "template_can_bo.xlsx", []XLSXColumn{
		{Header: "ho_ten", Hint: "Họ và tên đầy đủ. VD: Trần Thị B"},
		{Header: "email", Hint: "Email hợp lệ. VD: tranthib@email.com"},
		{Header: "so_cccd", Hint: "12 chữ số (không bắt buộc). VD: 012345678901"},
		{Header: "so_dien_thoai", Hint: "Số điện thoại. VD: 0901234567"},
		{Header: "vai_tro", Hint: "staff | manager | super_admin (hoặc: cán bộ | quản lý)"},
		{Header: "ma_phong_ban", Hint: "Mã phòng ban (tùy chọn). VD: IT"},
	})
}

// ExportCitizens handles GET /admin/users/export/citizens
func (h *AdminUserHandler) ExportCitizens(c *echo.Context) error {
	setCSVHeaders(c, "cong_dan.csv")
	writeCSVBOM(c)
	w := csv.NewWriter(c.Response())
	_ = w.Write([]string{"ho_ten", "email", "so_dien_thoai", "dia_chi", "trang_thai"})
	h.exportUserRows(c, w, repositories.UserFilter{Role: string(models.UserRoleCitizen)})
	w.Flush()
	return nil
}

func (h *AdminUserHandler) ExportStaff(c *echo.Context) error {
	setCSVHeaders(c, "can_bo.csv")
	writeCSVBOM(c)
	w := csv.NewWriter(c.Response())
	_ = w.Write([]string{"ho_ten", "email", "so_dien_thoai", "dia_chi", "vai_tro", "trang_thai"})
	h.exportUserRows(c, w, repositories.UserFilter{Roles: []string{
		string(models.UserRoleStaff),
		string(models.UserRoleManager),
		string(models.UserRoleSuperAdmin),
	}})
	w.Flush()
	return nil
}

func (h *AdminUserHandler) exportUserRows(c *echo.Context, w *csv.Writer, filter repositories.UserFilter) {
	batchSize := 1000
	for page := 1; ; page++ {
		users, _, err := h.svc.ListUsers(filter, page, batchSize)
		if err != nil {
			break
		}
		for _, u := range users {
			_ = w.Write([]string{u.Name, u.Email, u.Phone, u.Address, string(u.Role), string(u.Status)})
		}
		if len(users) < batchSize {
			break
		}
	}
}

// ImportCSV handles POST /admin/users/import
func (h *AdminUserHandler) ImportCSV(c *echo.Context) error {
	if h.importSvc == nil {
		return echo.NewHTTPError(http.StatusNotImplemented, "Chức năng import chưa được kích hoạt")
	}
	dataRows, err := parseUploadedCSV(c)
	if err != nil {
		return err
	}

	importRows := make([]services.StaffImportRow, 0, len(dataRows))
	for _, cols := range dataRows {
		importRows = append(importRows, services.StaffImportRow{
			HoTen:       safeCol(cols, 0),
			Email:       safeCol(cols, 1),
			SoCCCD:      safeCol(cols, 2),
			SoDienThoai: safeCol(cols, 3),
			VaiTro:      safeCol(cols, 4),
			MaPhongBan:  safeCol(cols, 5),
		})
	}

	errs := h.importSvc.ImportStaff(importRows, actorID(c))
	if len(errs) > 0 {
		search := c.QueryParam("search")
		role := c.QueryParam("role")
		users, total, _ := h.svc.ListUsers(repositories.UserFilter{Search: search, Role: role}, 1, 20)
		data := map[string]interface{}{
			"Title":          configs.T(c, "ui.users.title", nil),
			"CurrentPath":    "/admin/users",
			"CurrentUser":    adminCurrentUser(c),
			"Users":          users,
			"Pagination":     utils.NewPagination(1, 20, total),
			"Search":         search,
			"RoleFilter":     role,
			"ImportErrors":   errs,
			"ImportAction":   "/admin/users/import",
			"TemplateAction": "/admin/users/template",
			"Flash":          flashFromQuery(c),
		}
		return c.Render(http.StatusUnprocessableEntity, "admin/pages/users/list.html", data)
	}

	return c.Redirect(http.StatusSeeOther, adminFlashURL("success", configs.T(c, "ui.msg.import_success", nil)))
}

func (h *AdminUserHandler) renderFormErrors(c *echo.Context, isEdit bool, user *models.User, req interface{}, fieldErrors map[string]string, globalError string) error {
	titleKey := "ui.users.form.create_title"
	if isEdit {
		titleKey = "ui.users.form.edit_title"
	}
	data := map[string]interface{}{
		"Title":       configs.T(c, titleKey, nil),
		"CurrentPath": "/admin/users",
		"CurrentUser": adminCurrentUser(c),
		"IsEdit":      isEdit,
		"User":        user,
		"Req":         req,
		"FieldErrors": fieldErrors,
		"Error":       globalError,
	}
	return c.Render(http.StatusUnprocessableEntity, "admin/pages/users/form.html", data)
}

func flashFromQuery(c *echo.Context) map[string]string {
	t := c.QueryParam("flash")
	msg := c.QueryParam("msg")
	if t == "" || msg == "" {
		return nil
	}
	return map[string]string{"Type": t, "Message": msg}
}
