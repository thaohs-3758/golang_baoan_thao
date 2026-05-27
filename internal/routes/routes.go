package routes

import (
	"net/http"

	"github.com/awesome-academy/golang_baoan_thao/internal/handlers"
	"github.com/awesome-academy/golang_baoan_thao/internal/middlewares"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

type ApiHandler struct {
	AdminAuthHandler        *handlers.AdminAuthHandler
	AdminDashboardHandler   *handlers.AdminDashboardHandler
	AdminUserHandler        *handlers.AdminUserHandler
	AdminDepartmentHandler  *handlers.AdminDepartmentHandler
	AdminCategoryHandler    *handlers.AdminCategoryHandler
	AdminApplicationHandler *handlers.AdminApplicationHandler
	AdminLogHandler         *handlers.AdminLogHandler
	AdminCitizenHandler     *handlers.AdminCitizenHandler
	AdminProfileHandler     *handlers.AdminProfileHandler
	AuthHandler             *handlers.AuthHandler
	CitizenProfileHandler   *handlers.CitizenProfileHandler
	ServiceCatalogHandler   *handlers.ServiceCatalogHandler
	ApplicationHandler      *handlers.ApplicationHandler
	NotificationHandler     *handlers.NotificationHandler
	CitizenWebHandler       *handlers.CitizenWebHandler
}

func SetupRoutes(e *echo.Echo, handler *ApiHandler) {
	e.GET("/", func(c *echo.Context) error {
		return c.Redirect(http.StatusSeeOther, "/login")
	})

	// Citizen web auth (public)
	e.GET("/login", handler.CitizenWebHandler.ShowLoginPage)
	e.POST("/login", handler.CitizenWebHandler.WebLogin)
	e.GET("/register", handler.CitizenWebHandler.ShowRegisterPage)
	e.POST("/register", handler.CitizenWebHandler.WebRegister)
	e.GET("/logout", handler.CitizenWebHandler.WebLogout)

	// Citizen web routes (protected by cookie auth, role=citizen)
	citizenWeb := e.Group("/citizen", middlewares.CitizenWebMiddleware, middleware.CSRF())
	citizenWeb.GET("", handler.CitizenWebHandler.ShowDashboard)
	citizenWeb.GET("/notifications", handler.CitizenWebHandler.ListNotifications)
	citizenWeb.POST("/notifications/read-all", handler.CitizenWebHandler.MarkAllNotificationsRead)
	citizenWeb.POST("/notifications/toggle-email", handler.CitizenWebHandler.ToggleEmailNotification)
	citizenWeb.POST("/notifications/:id/read", handler.CitizenWebHandler.MarkNotificationRead)

	// Citizen service catalog
	citizenWeb.GET("/services", handler.CitizenWebHandler.ShowServiceCatalog)
	citizenWeb.GET("/services/:id", handler.CitizenWebHandler.ShowServiceDetail)

	// Citizen applications — /new must be before /:id to avoid route conflict
	citizenWeb.GET("/applications", handler.CitizenWebHandler.ShowApplicationsList)
	citizenWeb.GET("/applications/new", handler.CitizenWebHandler.ShowApplyForm)
	citizenWeb.POST("/applications", handler.CitizenWebHandler.SubmitApplication)
	citizenWeb.GET("/applications/:id", handler.CitizenWebHandler.ShowApplicationDetail)
	citizenWeb.POST("/applications/:id/supplements", handler.CitizenWebHandler.UploadApplicationSupplements)
	citizenWeb.GET("/profile", handler.CitizenWebHandler.ShowProfilePage)
	citizenWeb.POST("/profile", handler.CitizenWebHandler.UpdateProfile)
	citizenWeb.POST("/profile/password", handler.CitizenWebHandler.ChangePassword)

	// Admin auth (public)
	e.GET("/admin/login", handler.AdminAuthHandler.ShowLoginPage)
	e.POST("/admin/login", handler.AdminAuthHandler.WebLogin)
	e.GET("/admin/logout", handler.AdminAuthHandler.WebLogout)
	e.GET("/set-locale", handler.AdminAuthHandler.SetLocale)

	// Admin web routes (protected by cookie auth)
	admin := e.Group("/admin", middlewares.AdminWebMiddleware)
	admin.GET("", handler.AdminDashboardHandler.ShowDashboard)

	// Admin self-profile (staff + manager + super_admin)
	adminProfile := admin.Group("", middlewares.AdminWebRequireRoles(models.UserRoleStaff, models.UserRoleManager, models.UserRoleSuperAdmin))
	adminProfile.GET("/profile", handler.AdminProfileHandler.ShowProfilePage)
	adminProfile.POST("/profile", handler.AdminProfileHandler.UpdateProfile)
	adminProfile.POST("/profile/password", handler.AdminProfileHandler.ChangePassword)

	// Citizens admin (Super Admin only)
	citizens := admin.Group("/citizens", middlewares.AdminWebRequireRoles(models.UserRoleSuperAdmin))
	citizens.GET("", handler.AdminCitizenHandler.ListCitizens)
	citizens.GET("/template", handler.AdminCitizenHandler.DownloadTemplate)
	citizens.POST("/import", handler.AdminCitizenHandler.ImportCSV)

	// Service types: list (Manager + Super Admin)
	stRead := admin.Group("/service-types", middlewares.AdminWebRequireRoles(models.UserRoleManager, models.UserRoleSuperAdmin))
	stRead.GET("", handler.ServiceCatalogHandler.ListServiceTypesAdmin)
	stRead.GET("/:id", handler.ServiceCatalogHandler.ShowServiceType)

	// Service types: write (Super Admin only)
	stWrite := admin.Group("/service-types", middlewares.AdminWebRequireRoles(models.UserRoleSuperAdmin))
	stWrite.GET("/export", handler.ServiceCatalogHandler.ExportCSV)
	stWrite.GET("/template", handler.ServiceCatalogHandler.DownloadTemplate)
	stWrite.POST("/import", handler.ServiceCatalogHandler.ImportCSV)
	stWrite.GET("/new", handler.ServiceCatalogHandler.CreateServiceTypeForm)
	stWrite.POST("", handler.ServiceCatalogHandler.CreateServiceType)
	stWrite.GET("/:id/edit", handler.ServiceCatalogHandler.EditServiceTypeForm)
	stWrite.POST("/:id", handler.ServiceCatalogHandler.UpdateServiceType)
	stWrite.POST("/:id/delete", handler.ServiceCatalogHandler.DeleteServiceType)

	// Users: list (Manager + Super Admin)
	usersRead := admin.Group("/users", middlewares.AdminWebRequireRoles(models.UserRoleManager, models.UserRoleSuperAdmin))
	usersRead.GET("", handler.AdminUserHandler.ListUsers)
	usersRead.GET("/:id", handler.AdminUserHandler.ShowUser)

	// Users: write (Super Admin only)
	usersWrite := admin.Group("/users", middlewares.AdminWebRequireRoles(models.UserRoleSuperAdmin))
	usersWrite.GET("/export/citizens", handler.AdminUserHandler.ExportCitizens)
	usersWrite.GET("/export/staff", handler.AdminUserHandler.ExportStaff)
	usersWrite.GET("/template", handler.AdminUserHandler.DownloadTemplate)
	usersWrite.POST("/import", handler.AdminUserHandler.ImportCSV)
	usersWrite.GET("/new", handler.AdminUserHandler.ShowCreateForm)
	usersWrite.POST("", handler.AdminUserHandler.CreateUser)
	usersWrite.GET("/:id/edit", handler.AdminUserHandler.ShowEditForm)
	usersWrite.POST("/:id/edit", handler.AdminUserHandler.UpdateUser)
	usersWrite.POST("/:id/block", handler.AdminUserHandler.BlockUser)
	usersWrite.POST("/:id/unblock", handler.AdminUserHandler.UnblockUser)
	usersWrite.POST("/:id/delete", handler.AdminUserHandler.DeleteUser)

	// Department list (Manager + Super Admin)
	deptList := admin.Group("/departments", middlewares.AdminWebRequireRoles(models.UserRoleManager, models.UserRoleSuperAdmin))
	deptList.GET("", handler.AdminDepartmentHandler.ListDepartments)
	deptList.GET("/export", handler.AdminDepartmentHandler.ExportCSV)
	deptList.GET("/template", handler.AdminDepartmentHandler.DownloadTemplate)

	// Departments admin actions (Super Admin only)
	depts := admin.Group("/departments", middlewares.AdminWebRequireRoles(models.UserRoleSuperAdmin))
	depts.POST("/import", handler.AdminDepartmentHandler.ImportCSV)
	depts.GET("/new", handler.AdminDepartmentHandler.ShowCreateForm)
	depts.POST("", handler.AdminDepartmentHandler.CreateDepartment)
	depts.GET("/:id/edit", handler.AdminDepartmentHandler.ShowEditForm)
	depts.POST("/:id/edit", handler.AdminDepartmentHandler.UpdateDepartment)
	depts.POST("/:id/delete", handler.AdminDepartmentHandler.DeleteDepartment)

	// Categories: list (Manager + Super Admin)
	catRead := admin.Group("/categories", middlewares.AdminWebRequireRoles(models.UserRoleManager, models.UserRoleSuperAdmin))
	catRead.GET("", handler.AdminCategoryHandler.ListCategories)

	// Categories: write (Super Admin only)
	catWrite := admin.Group("/categories", middlewares.AdminWebRequireRoles(models.UserRoleSuperAdmin))
	catWrite.GET("/new", handler.AdminCategoryHandler.ShowCreateForm)
	catWrite.POST("", handler.AdminCategoryHandler.CreateCategory)
	catWrite.GET("/:id/edit", handler.AdminCategoryHandler.ShowEditForm)
	catWrite.POST("/:id/edit", handler.AdminCategoryHandler.UpdateCategory)
	catWrite.POST("/:id/delete", handler.AdminCategoryHandler.DeleteCategory)
	// Department staff management (Manager only)
	deptStaff := admin.Group("/departments/:id/staff", middlewares.AdminWebRequireRoles(models.UserRoleManager))
	deptStaff.GET("", handler.AdminDepartmentHandler.ListDepartmentStaff)
	deptStaff.GET("/assign", handler.AdminDepartmentHandler.ShowAssignStaffForm)
	deptStaff.POST("/assign", handler.AdminDepartmentHandler.AssignStaffToDept)
	deptStaff.POST("/:user_id/remove", handler.AdminDepartmentHandler.RemoveStaffFromDept)

	// Admin applications: read (Manager + Super Admin + Staff)
	appsRead := admin.Group("/applications", middlewares.AdminWebRequireRoles(models.UserRoleManager, models.UserRoleSuperAdmin, models.UserRoleStaff))
	appsRead.GET("", handler.AdminApplicationHandler.ListApplications)
	appsRead.GET("/export", handler.AdminApplicationHandler.ExportCSV)
	appsRead.GET("/:id", handler.AdminApplicationHandler.ShowApplication)

	// Admin applications: process (Staff only)
	appsProcess := admin.Group("/applications", middlewares.AdminWebRequireRoles(models.UserRoleStaff))
	appsProcess.POST("/:id/process", handler.AdminApplicationHandler.ProcessApplication)

	// Admin applications: assign (Manager only)
	appsAssign := admin.Group("/applications", middlewares.AdminWebRequireRoles(models.UserRoleManager))
	appsAssign.GET("/:id/assign", handler.AdminApplicationHandler.ShowAssignForm)
	appsAssign.POST("/:id/assign", handler.AdminApplicationHandler.AssignToStaff)

	// Activity logs: read (Manager + Super Admin)
	logsRead := admin.Group("/logs", middlewares.AdminWebRequireRoles(models.UserRoleManager, models.UserRoleSuperAdmin))
	logsRead.GET("", handler.AdminLogHandler.ListLogs)

	// Activity logs cleanup: write (Super Admin only)
	logsWrite := admin.Group("/logs", middlewares.AdminWebRequireRoles(models.UserRoleSuperAdmin))
	logsWrite.POST("/cleanup", handler.AdminLogHandler.CleanupLogs)

	api := e.Group("/api")

	auth := api.Group("/auth")
	auth.POST("/login", handler.AuthHandler.Login)
	auth.POST("/register", handler.AuthHandler.Register)
	auth.POST("/refresh", handler.AuthHandler.RefreshTokenHandler)
	auth.POST("/logout", handler.AuthHandler.Logout)

	citizen := api.Group("/citizens")
	citizen.Use(middlewares.JWTMiddleware)
	citizen.Use(middlewares.RequireRoles(models.UserRoleCitizen))
	citizen.GET("/me", handler.CitizenProfileHandler.GetMe)
	citizen.PUT("/me", handler.CitizenProfileHandler.UpdateMe)
	citizen.PUT("/me/password", handler.CitizenProfileHandler.ChangeMyPassword)
	citizen.GET("/services", handler.ServiceCatalogHandler.ListServices)
	citizen.GET("/services/:id", handler.ServiceCatalogHandler.GetService)

	// Applications
	citizen.POST("/me/applications", handler.ApplicationHandler.Submit)
	citizen.GET("/me/applications", handler.ApplicationHandler.ListMine)
	citizen.GET("/me/applications/:id", handler.ApplicationHandler.GetMine)
	citizen.GET("/me/applications/:id/status-history", handler.ApplicationHandler.ListMyStatusHistory)
	citizen.POST("/me/applications/:id/supplements", handler.ApplicationHandler.UploadSupplements)

	// Notifications
	citizen.GET("/me/notifications", handler.NotificationHandler.List)
	citizen.PUT("/me/notifications/read-all", handler.NotificationHandler.MarkAllAsRead)
	citizen.PUT("/me/notifications/:id/read", handler.NotificationHandler.MarkAsRead)

	staff := api.Group("/staff")
	staff.Use(middlewares.JWTMiddleware)
	staff.Use(middlewares.RequireRoles(models.UserRoleStaff))

	manager := api.Group("/managers")
	manager.Use(middlewares.JWTMiddleware)
	manager.Use(middlewares.RequireRoles(models.UserRoleManager))
}
