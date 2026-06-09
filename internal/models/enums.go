package models

type UserRole string

const (
	UserRoleCitizen    UserRole = "citizen"
	UserRoleStaff      UserRole = "staff"
	UserRoleManager    UserRole = "manager"
	UserRoleSuperAdmin UserRole = "super_admin"
)

type UserStatus string

const (
	UserStatusActive  UserStatus = "active"
	UserStatusBlocked UserStatus = "blocked"
)

type ApplicationStatus string

const (
	ApplicationStatusReceived     ApplicationStatus = "received"
	ApplicationStatusProcessing   ApplicationStatus = "processing"
	ApplicationStatusNeedMoreInfo ApplicationStatus = "need_more_info"
	ApplicationStatusApproved     ApplicationStatus = "approved"
	ApplicationStatusRejected     ApplicationStatus = "rejected"
)

type AttachmentType string

const (
	AttachmentTypeSubmitted  AttachmentType = "submitted"
	AttachmentTypeSupplement AttachmentType = "supplement"
	AttachmentTypeResult     AttachmentType = "result"
)

type NotificationType string

const (
	NotificationTypeReceived     NotificationType = "received"
	NotificationTypeNeedMoreInfo NotificationType = "need_more_info"
	NotificationTypeResult       NotificationType = "result"
	NotificationTypeSystem       NotificationType = "system"
	NotificationTypeDeadlineReminder NotificationType = "deadline_reminder"
)

type AssignmentAction string

const (
	AssignmentActionAssigned    AssignmentAction = "assigned"
	AssignmentActionTransferred AssignmentAction = "transferred"
	AssignmentActionUnassigned  AssignmentAction = "unassigned"
)

type ImportExportType string

const (
	ImportExportTypeImport ImportExportType = "import"
	ImportExportTypeExport ImportExportType = "export"
)

type ImportExportStatus string

const (
	ImportExportStatusPending    ImportExportStatus = "pending"
	ImportExportStatusProcessing ImportExportStatus = "processing"
	ImportExportStatusSuccess    ImportExportStatus = "success"
	ImportExportStatusFailed     ImportExportStatus = "failed"
)

type ServiceCategory string

const (
	ServiceCategoryAdministrative ServiceCategory = "hanh_chinh_cong"
	ServiceCategoryEducation      ServiceCategory = "giao_duc"
	ServiceCategoryHealth         ServiceCategory = "y_te"
	ServiceCategoryConstruction   ServiceCategory = "xay_dung"
	ServiceCategoryResources      ServiceCategory = "tai_nguyen"
)
