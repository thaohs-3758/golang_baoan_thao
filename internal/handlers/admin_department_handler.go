package handlers

import (
	"encoding/csv"
	"errors"
	"net/http"
	"net/url"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/dtos"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
	"github.com/awesome-academy/golang_baoan_thao/internal/services"
	"github.com/awesome-academy/golang_baoan_thao/internal/utils"
	"github.com/labstack/echo/v5"
)

type DepartmentService interface {
	ListDepartments(filter repositories.DepartmentFilter, page, limit int) ([]models.Department, int64, error)
	GetDepartment(id string) (*models.Department, error)
	CreateDepartment(req *dtos.DepartmentCreateRequest, createdBy string) (*models.Department, error)
	UpdateDepartment(id string, req *dtos.DepartmentUpdateRequest, updatedBy string) (*models.Department, error)
	DeleteDepartment(id string, deletedBy string) error
}

type DepartmentImportExportService interface {
	ImportDepartments(rows []services.DepartmentImportRow, createdBy string) []string
}

type StaffProfileService interface {
	ListStaffByDepartment(deptID string, page, limit int) ([]models.StaffProfile, int64, error)
	FindStaffProfileByUserID(userID string) (*models.StaffProfile, error)
	AssignStaffToDepartment(userID string, deptID string, updatedBy string) error
	RemoveStaffFromDepartment(userID string, updatedBy string) error
}

type AdminDepartmentHandler struct {
	svc        DepartmentService
	userSvc    AdminUserService
	profileSvc StaffProfileService
	importSvc  DepartmentImportExportService
}

func NewAdminDepartmentHandler(svc DepartmentService, userSvc AdminUserService, profileSvc StaffProfileService) *AdminDepartmentHandler {
	return &AdminDepartmentHandler{svc: svc, userSvc: userSvc, profileSvc: profileSvc}
}

func (h *AdminDepartmentHandler) WithImportExport(svc DepartmentImportExportService) *AdminDepartmentHandler {
	h.importSvc = svc
	return h
}

func deptFlashURL(flash, msg string) string {
	return "/admin/departments?" + url.Values{"flash": {flash}, "msg": {msg}}.Encode()
}

func deptStaffFlashURL(deptID, flash, msg string) string {
	return "/admin/departments/" + deptID + "/staff?" + url.Values{"flash": {flash}, "msg": {msg}}.Encode()
}

func deptAssignWarnURL(deptID, userID string) string {
	return "/admin/departments/" + deptID + "/staff/assign?" + url.Values{
		"warn":    {"department_transfer_confirm_required"},
		"user_id": {userID},
	}.Encode()
}

func (h *AdminDepartmentHandler) ListDepartments(c *echo.Context) error {
	search := c.QueryParam("search")
	page, limit := parsePagination(c)

	filter := repositories.DepartmentFilter{Search: search}
	depts, total, err := h.svc.ListDepartments(filter, page, limit)
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"Title":          configs.T(c, "ui.departments.title", nil),
		"CurrentPath":    "/admin/departments",
		"CurrentUser":    adminCurrentUser(c),
		"Departments":    depts,
		"Pagination":     utils.NewPagination(page, limit, total),
		"Search":         search,
		"ImportAction":   "/admin/departments/import",
		"TemplateAction": "/admin/departments/template",
		"Flash":          flashFromQuery(c),
	}
	return c.Render(http.StatusOK, "admin/pages/departments/list.html", data)
}

func (h *AdminDepartmentHandler) ShowCreateForm(c *echo.Context) error {
	staffUsers, err := h.listAssignableStaffUsers()
	if err != nil {
		return err
	}
	data := map[string]interface{}{
		"Title":           configs.T(c, "ui.departments.form.create_title", nil),
		"CurrentPath":     "/admin/departments",
		"CurrentUser":     adminCurrentUser(c),
		"IsEdit":          false,
		"StaffUsers":      staffUsers,
		"CurrentLeaderID": "",
	}
	return c.Render(http.StatusOK, "admin/pages/departments/form.html", data)
}

func (h *AdminDepartmentHandler) CreateDepartment(c *echo.Context) error {
	req := new(dtos.DepartmentCreateRequest)
	if err := c.Bind(req); err != nil {
		return h.renderFormErrors(c, false, nil, req, nil, configs.T(c, "common.invalid_data", nil))
	}
	if err := c.Validate(req); err != nil {
		return h.renderFormErrors(c, false, nil, req, extractFieldErrors(c, err), "")
	}
	if req.LeaderUserID != "" {
		user, err := h.userSvc.GetUser(req.LeaderUserID)
		if err != nil || user == nil || user.Role != models.UserRoleStaff {
			return h.renderFormErrors(c, false, nil, req, nil, configs.T(c, "validation.invalid", nil))
		}
	}

	_, err := h.svc.CreateDepartment(req, actorID(c))
	if err != nil {
		msg := configs.T(c, "common.internal_error", nil)
		if errors.Is(err, services.ErrDepartmentCodeExists) {
			msg = configs.T(c, "department.code_exists", nil)
		}
		if errors.Is(err, services.ErrDepartmentLeaderAlreadyAssigned) {
			msg = configs.T(c, "department.leader_already_assigned", nil)
		}
		return h.renderFormErrors(c, false, nil, req, nil, msg)
	}

	return c.Redirect(http.StatusSeeOther, deptFlashURL("success", configs.T(c, "ui.msg.department_created", nil)))
}

func (h *AdminDepartmentHandler) ShowEditForm(c *echo.Context) error {
	id := c.Param("id")
	dept, err := h.svc.GetDepartment(id)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, deptFlashURL("error", configs.T(c, "department.not_found", nil)))
	}

	staffUsers, err := h.listAssignableStaffUsers()
	if err != nil {
		return err
	}
	data := map[string]interface{}{
		"Title":           configs.T(c, "ui.departments.form.edit_title", nil),
		"CurrentPath":     "/admin/departments",
		"CurrentUser":     adminCurrentUser(c),
		"IsEdit":          true,
		"Department":      dept,
		"StaffUsers":      staffUsers,
		"CurrentLeaderID": derefStr(dept.LeaderUserID),
	}
	return c.Render(http.StatusOK, "admin/pages/departments/form.html", data)
}

func (h *AdminDepartmentHandler) UpdateDepartment(c *echo.Context) error {
	id := c.Param("id")

	dept, err := h.svc.GetDepartment(id)
	if err != nil {
		return c.Redirect(http.StatusSeeOther, deptFlashURL("error", configs.T(c, "department.not_found", nil)))
	}

	req := new(dtos.DepartmentUpdateRequest)
	if err := c.Bind(req); err != nil {
		return h.renderFormErrors(c, true, dept, req, nil, configs.T(c, "common.invalid_data", nil))
	}
	if err := c.Validate(req); err != nil {
		return h.renderFormErrors(c, true, dept, req, extractFieldErrors(c, err), "")
	}
	if req.LeaderUserID != "" {
		user, err := h.userSvc.GetUser(req.LeaderUserID)
		if err != nil || user == nil || user.Role != models.UserRoleStaff {
			return h.renderFormErrors(c, true, dept, req, nil, configs.T(c, "validation.invalid", nil))
		}
	}

	if _, err := h.svc.UpdateDepartment(id, req, actorID(c)); err != nil {
		msg := configs.T(c, "common.internal_error", nil)
		if errors.Is(err, services.ErrDepartmentCodeExists) {
			msg = configs.T(c, "department.code_exists", nil)
		}
		if errors.Is(err, services.ErrDepartmentLeaderAlreadyAssigned) {
			msg = configs.T(c, "department.leader_already_assigned", nil)
		}
		return h.renderFormErrors(c, true, dept, req, nil, msg)
	}

	return c.Redirect(http.StatusSeeOther, deptFlashURL("success", configs.T(c, "ui.msg.department_updated", nil)))
}

func (h *AdminDepartmentHandler) DeleteDepartment(c *echo.Context) error {
	id := c.Param("id")
	if err := h.svc.DeleteDepartment(id, actorID(c)); err != nil {
		return c.Redirect(http.StatusSeeOther, deptFlashURL("error", configs.T(c, "ui.msg.department_delete_failed", nil)))
	}
	return c.Redirect(http.StatusSeeOther, deptFlashURL("success", configs.T(c, "ui.msg.department_deleted", nil)))
}

// DownloadTemplate handles GET /admin/departments/template
func (h *AdminDepartmentHandler) DownloadTemplate(c *echo.Context) error {
	return writeXLSXTemplate(c, "template_phong_ban.xlsx", []XLSXColumn{
		{Header: "ten", Hint: "Tên phòng ban. VD: Phòng Công nghệ thông tin"},
		{Header: "mo_ta", Hint: "Mô tả (tùy chọn). VD: Phụ trách hạ tầng CNTT"},
		{Header: "ma_code", Hint: "Mã phòng ban, viết liền không dấu. VD: IT"},
	})
}

// ExportCSV handles GET /admin/departments/export
func (h *AdminDepartmentHandler) ExportCSV(c *echo.Context) error {
	setCSVHeaders(c, "phong_ban.csv")
	writeCSVBOM(c)

	w := csv.NewWriter(c.Response())
	_ = w.Write([]string{"ten", "mo_ta", "ma_code"})

	batchSize := 1000
	for page := 1; ; page++ {
		depts, _, err := h.svc.ListDepartments(repositories.DepartmentFilter{}, page, batchSize)
		if err != nil {
			break
		}
		for _, d := range depts {
			_ = w.Write([]string{d.Name, d.Address, d.Code})
		}
		if len(depts) < batchSize {
			break
		}
	}
	w.Flush()
	return nil
}

// ImportCSV handles POST /admin/departments/import
func (h *AdminDepartmentHandler) ImportCSV(c *echo.Context) error {
	if h.importSvc == nil {
		return echo.NewHTTPError(http.StatusNotImplemented, "Chức năng import chưa được kích hoạt")
	}
	dataRows, err := parseUploadedCSV(c)
	if err != nil {
		return err
	}

	importRows := make([]services.DepartmentImportRow, 0, len(dataRows))
	for _, cols := range dataRows {
		importRows = append(importRows, services.DepartmentImportRow{
			Ten:    safeCol(cols, 0),
			MoTa:   safeCol(cols, 1),
			MaCode: safeCol(cols, 2),
		})
	}

	errs := h.importSvc.ImportDepartments(importRows, actorID(c))
	if len(errs) > 0 {
		search := c.QueryParam("search")
		depts, total, _ := h.svc.ListDepartments(repositories.DepartmentFilter{Search: search}, 1, 20)
		data := map[string]interface{}{
			"Title":          configs.T(c, "ui.departments.title", nil),
			"CurrentPath":    "/admin/departments",
			"CurrentUser":    adminCurrentUser(c),
			"Departments":    depts,
			"Pagination":     utils.NewPagination(1, 20, total),
			"Search":         search,
			"ImportErrors":   errs,
			"ImportAction":   "/admin/departments/import",
			"TemplateAction": "/admin/departments/template",
			"Flash":          flashFromQuery(c),
		}
		return c.Render(http.StatusUnprocessableEntity, "admin/pages/departments/list.html", data)
	}

	return c.Redirect(http.StatusSeeOther, deptFlashURL("success", configs.T(c, "ui.msg.import_success", nil)))
}

func (h *AdminDepartmentHandler) renderFormErrors(c *echo.Context, isEdit bool, dept *models.Department, req interface{}, fieldErrors map[string]string, globalError string) error {
	titleKey := "ui.departments.form.create_title"
	if isEdit {
		titleKey = "ui.departments.form.edit_title"
	}
	staffUsers, _ := h.listAssignableStaffUsers()
	curLeaderID := ""
	if dept != nil {
		curLeaderID = derefStr(dept.LeaderUserID)
	}
	data := map[string]interface{}{
		"Title":           configs.T(c, titleKey, nil),
		"CurrentPath":     "/admin/departments",
		"CurrentUser":     adminCurrentUser(c),
		"IsEdit":          isEdit,
		"Department":      dept,
		"Req":             req,
		"FieldErrors":     fieldErrors,
		"Error":           globalError,
		"StaffUsers":      staffUsers,
		"CurrentLeaderID": curLeaderID,
	}
	return c.Render(http.StatusUnprocessableEntity, "admin/pages/departments/form.html", data)
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func departmentIDParam(c *echo.Context) (string, error) {
	id := c.Param("id")
	if id == "" || id == ":id" {
		return "", echo.NewHTTPError(http.StatusNotFound, "department.not_found")
	}
	return id, nil
}

func (h *AdminDepartmentHandler) listStaffUsers() ([]models.User, error) {
	users, _, err := h.userSvc.ListUsers(repositories.UserFilter{}, 1, 1000)
	if err != nil {
		return nil, err
	}
	var staff []models.User
	for _, u := range users {
		if u.Role == models.UserRoleStaff || u.Role == models.UserRoleManager || u.Role == models.UserRoleSuperAdmin {
			staff = append(staff, u)
		}
	}
	return staff, nil
}

// filterOutOtherDeptLeaders removes staff who are leaders of a department other
// than targetDeptID. Managers are not allowed to reassign such staff.
func (h *AdminDepartmentHandler) filterOutOtherDeptLeaders(users []models.User, targetDeptID string) []models.User {
	depts, _, err := h.svc.ListDepartments(repositories.DepartmentFilter{}, 1, 10000)
	if err != nil {
		return []models.User{}
	}
	blocked := make(map[string]struct{})
	for _, d := range depts {
		if d.ID == targetDeptID || d.LeaderUserID == nil {
			continue
		}
		blocked[*d.LeaderUserID] = struct{}{}
	}
	if len(blocked) == 0 {
		return users
	}
	filtered := make([]models.User, 0, len(users))
	for _, u := range users {
		if _, ok := blocked[u.ID]; !ok {
			filtered = append(filtered, u)
		}
	}
	return filtered
}

func (h *AdminDepartmentHandler) listAssignableStaffUsers() ([]models.User, error) {
	users, _, err := h.userSvc.ListUsers(repositories.UserFilter{}, 1, 1000)
	if err != nil {
		return nil, err
	}
	staff := make([]models.User, 0)
	for _, u := range users {
		if u.Role == models.UserRoleStaff {
			staff = append(staff, u)
		}
	}
	return staff, nil
}

func (h *AdminDepartmentHandler) ListDepartmentStaff(c *echo.Context) error {
	if h.profileSvc == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "service.unavailable")
	}
	id, err := departmentIDParam(c)
	if err != nil {
		return err
	}
	page, limit := parsePagination(c)
	// load department to get leader info
	dept, err := h.svc.GetDepartment(id)
	if err != nil {
		return err
	}

	profiles, total, err := h.profileSvc.ListStaffByDepartment(id, page, limit)
	if err != nil {
		return err
	}

	var leaderProfile *models.StaffProfile
	filtered := make([]models.StaffProfile, 0, len(profiles))
	for _, p := range profiles {
		if dept != nil && dept.LeaderUserID != nil && p.UserID == *dept.LeaderUserID {
			// capture leader's profile for separate display and remove from list
			leaderProfile = &p
			continue
		}
		filtered = append(filtered, p)
	}

	// adjust total if leader was present in this page
	if leaderProfile != nil && total > 0 {
		total = total - 1
	}

	data := map[string]interface{}{
		"Title":         configs.T(c, "ui.departments.staff_list_title", nil),
		"CurrentPath":   "/admin/departments",
		"CurrentUser":   adminCurrentUser(c),
		"DepartmentID":  id,
		"LeaderUser":    dept.LeaderUser,
		"LeaderProfile": leaderProfile,
		"StaffProfiles": filtered,
		"Pagination":    utils.NewPagination(page, limit, total),
		"Flash":         flashFromQuery(c),
	}
	return c.Render(http.StatusOK, "admin/pages/departments/staff_list.html", data)
}

func (h *AdminDepartmentHandler) ShowAssignStaffForm(c *echo.Context) error {
	id, err := departmentIDParam(c)
	if err != nil {
		return err
	}
	staffUsers, err := h.listAssignableStaffUsers()
	if err != nil {
		return err
	}
	// Managers cannot transfer a staff who is a leader of another department.
	// Filter them out of the dropdown so the UI never offers an unselectable choice.
	currentUser := adminCurrentUser(c)
	if currentUser != nil && currentUser.Role == string(models.UserRoleManager) {
		staffUsers = h.filterOutOtherDeptLeaders(staffUsers, id)
	}
	warn := c.QueryParam("warn")
	selectedUserID := c.QueryParam("user_id")
	warningActionURL := "/admin/departments/" + id + "/staff/assign"
	selectedStaffName := ""
	currentDepartmentName := ""
	targetDepartmentName := ""
	if warn == "department_transfer_confirm_required" && selectedUserID != "" {
		warningActionURL = deptAssignWarnURL(id, selectedUserID)
		if user, uErr := h.userSvc.GetUser(selectedUserID); uErr == nil && user != nil {
			selectedStaffName = user.Name
		}
		if dept, dErr := h.svc.GetDepartment(id); dErr == nil && dept != nil {
			targetDepartmentName = dept.Name
		}
		if h.profileSvc != nil {
			if sp, pErr := h.profileSvc.FindStaffProfileByUserID(selectedUserID); pErr == nil && sp != nil && sp.DepartmentID != nil {
				if fromDept, fromErr := h.svc.GetDepartment(*sp.DepartmentID); fromErr == nil && fromDept != nil {
					currentDepartmentName = fromDept.Name
				}
			}
		}
	}

	data := map[string]interface{}{
		"Title":                 configs.T(c, "ui.departments.assign_staff_title", nil),
		"CurrentPath":           "/admin/departments",
		"CurrentUser":           adminCurrentUser(c),
		"DepartmentID":          id,
		"StaffUsers":            staffUsers,
		"WarningType":           warn,
		"SelectedUserID":        selectedUserID,
		"ConfirmTransferValue":  "1",
		"WarningActionURL":      warningActionURL,
		"SelectedStaffName":     selectedStaffName,
		"CurrentDepartmentName": currentDepartmentName,
		"TargetDepartmentName":  targetDepartmentName,
	}
	return c.Render(http.StatusOK, "admin/pages/departments/assign_staff_form.html", data)
}

func (h *AdminDepartmentHandler) AssignStaffToDept(c *echo.Context) error {
	if h.profileSvc == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "service.unavailable")
	}
	deptID, err := departmentIDParam(c)
	if err != nil {
		return err
	}
	userID := c.FormValue("user_id")
	if userID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "validation.invalid")
	}
	confirmTransfer := c.FormValue("confirm_transfer") == "1"
	warn := c.QueryParam("warn")
	warnUserID := c.QueryParam("user_id")
	if confirmTransfer && (warn != "department_transfer_confirm_required" || warnUserID == "" || warnUserID != userID) {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "validation.invalid")
	}
	user, err := h.userSvc.GetUser(userID)
	if err != nil || user == nil || user.Role != models.UserRoleStaff {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "validation.invalid")
	}
	sp, err := h.profileSvc.FindStaffProfileByUserID(userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}
	if sp != nil && sp.DepartmentID != nil && *sp.DepartmentID != deptID && !confirmTransfer {
		return c.Redirect(http.StatusSeeOther, deptAssignWarnURL(deptID, userID))
	}
	if err := h.profileSvc.AssignStaffToDepartment(userID, deptID, actorID(c)); err != nil {
		if errors.Is(err, services.ErrLeaderTransferForbiddenForManager) {
			return echo.NewHTTPError(http.StatusUnprocessableEntity, "department.leader_transfer_forbidden_for_manager")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}
	return c.Redirect(http.StatusSeeOther, deptStaffFlashURL(deptID, "success", configs.T(c, "ui.msg.department_staff_assigned", nil)))
}

func (h *AdminDepartmentHandler) RemoveStaffFromDept(c *echo.Context) error {
	if h.profileSvc == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "service.unavailable")
	}
	userID := c.Param("user_id")
	deptID, err := departmentIDParam(c)
	if err != nil {
		return err
	}
	if userID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "validation.invalid")
	}
	if err := h.profileSvc.RemoveStaffFromDepartment(userID, actorID(c)); err != nil {
		if errors.Is(err, services.ErrLeaderTransferForbiddenForManager) {
			return echo.NewHTTPError(http.StatusUnprocessableEntity, "department.leader_transfer_forbidden_for_manager")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "common.internal_error")
	}
	return c.Redirect(http.StatusSeeOther, deptStaffFlashURL(deptID, "success", configs.T(c, "ui.msg.department_staff_removed", nil)))
}
