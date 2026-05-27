package services

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/awesome-academy/golang_baoan_thao/internal/repositories"
)

var ErrServiceTypeHasApplications = errors.New("service_type.applications_exist")
var ErrServiceTypeCodeExists = errors.New("service_type.code_duplicate")

type ServiceCatalogService struct {
	repo           repositories.ServiceTypeRepository
	activityLogger activityLogger
}

func NewServiceCatalogService(repo repositories.ServiceTypeRepository, loggers ...activityLogger) *ServiceCatalogService {
	var logger activityLogger
	if len(loggers) > 0 {
		logger = loggers[0]
	}
	return &ServiceCatalogService{repo: repo, activityLogger: logger}
}

func (s *ServiceCatalogService) List(ctx context.Context, filter repositories.ListFilter) (*repositories.ListResult, error) {
	return s.repo.List(ctx, filter)
}
func (s *ServiceCatalogService) GetByID(ctx context.Context, id string) (*models.ServiceType, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ServiceCatalogService) GetByIDForAdmin(ctx context.Context, id string) (*models.ServiceType, error) {
	return s.repo.GetByIDForAdmin(ctx, id)
}

func (s *ServiceCatalogService) ListDepartments(ctx context.Context) ([]models.Department, error) {
	return s.repo.ListDepartments(ctx)
}

func (s *ServiceCatalogService) ListCategories(ctx context.Context) ([]models.Category, error) {
	return s.repo.ListCategories(ctx)
}

func (s *ServiceCatalogService) Create(ctx context.Context, st *models.ServiceType) error {
	if err := s.repo.Create(ctx, st); err != nil {
		if isDuplicateServiceTypeError(err) {
			return ErrServiceTypeCodeExists
		}
		return err
	}
	actorID := createdByFromContext(ctx)
	s.logActivity(&models.ActivityLog{
		ActorUserID: actorID,
		Action:      "service_type.create",
		EntityType:  "service_type",
		EntityID:    &st.ID,
		Description: "Tạo loại dịch vụ: " + st.Name,
		Result:      "success",
		CreatedAt:   time.Now(),
	})
	return nil
}

func (s *ServiceCatalogService) Update(ctx context.Context, st *models.ServiceType) error {
	var before *models.ServiceType
	if strings.TrimSpace(st.ID) != "" {
		before, _ = s.repo.GetByIDForAdmin(ctx, st.ID)
	}
	if err := s.repo.Update(ctx, st); err != nil {
		if isDuplicateServiceTypeError(err) {
			return ErrServiceTypeCodeExists
		}
		return err
	}
	actorID := createdByFromContext(ctx)
	metadata := map[string]any{}
	if before != nil {
		metadata["changes"] = map[string]any{
			"name":      map[string]any{"before": before.Name, "after": st.Name},
			"code":      map[string]any{"before": before.Code, "after": st.Code},
			"is_active": map[string]any{"before": before.IsActive, "after": st.IsActive},
		}
	}
	s.logActivity(&models.ActivityLog{
		ActorUserID:  actorID,
		Action:       "service_type.update",
		EntityType:   "service_type",
		EntityID:     &st.ID,
		Description:  "Cập nhật loại dịch vụ: " + st.Name,
		Result:       "success",
		MetadataJSON: mustJSON(metadata),
		CreatedAt:    time.Now(),
	})
	return nil
}

func (s *ServiceCatalogService) Delete(ctx context.Context, id string) error {
	serviceType, err := s.repo.GetByIDForAdmin(ctx, id)
	if err != nil {
		return err
	}

	count, err := s.repo.CountApplications(ctx, serviceType.ID)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrServiceTypeHasApplications
	}

	if err := s.repo.Delete(ctx, serviceType.ID); err != nil {
		return err
	}
	actorID := createdByFromContext(ctx)
	s.logActivity(&models.ActivityLog{
		ActorUserID: actorID,
		Action:      "service_type.delete",
		EntityType:  "service_type",
		EntityID:    &serviceType.ID,
		Description: "Xóa loại dịch vụ: " + serviceType.Name,
		Result:      "success",
		CreatedAt:   time.Now(),
	})
	return nil
}

func isDuplicateServiceTypeError(err error) bool {
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate key value violates unique constraint") || strings.Contains(message, "service_types_code_key")
}

func (s *ServiceCatalogService) logActivity(entry *models.ActivityLog) {
	if s.activityLogger == nil || entry == nil {
		return
	}
	if err := s.activityLogger.Log(entry); err != nil {
		log.Printf("activity log write failed for action %s: %v", entry.Action, err)
	}
}

func createdByFromContext(ctx context.Context) *string {
	if ctx == nil {
		return nil
	}
	actor, ok := ctx.Value("actor_user_id").(string)
	if !ok || strings.TrimSpace(actor) == "" {
		return nil
	}
	return &actor
}
