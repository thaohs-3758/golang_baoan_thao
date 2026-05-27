package repositories

import (
	"errors"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"gorm.io/gorm"
)

type CitizenExportRow struct {
	SoCCCD    string
	HoTen     string
	Email     string
	SoDienThoai string
	DiaChi    string
	NgaySinh  string
	GioiTinh  string
	DiaChiThuongTru string
	TongHoSo  int64
}

type CitizenProfileRepository interface {
	GetByUserID(userID string) (*models.CitizenProfile, error)
	Update(profile *models.CitizenProfile) error
	CreateInTx(tx *gorm.DB, profile *models.CitizenProfile) error
	FindByCitizenIDNumber(cccd string) (*models.CitizenProfile, error)
	ListAllForExport(offset, limit int) ([]CitizenExportRow, int64, error)
}

type citizenProfileRepo struct {
	db *gorm.DB
}

func NewCitizenProfileRepository(db *gorm.DB) CitizenProfileRepository {
	return &citizenProfileRepo{db: db}
}

func (r *citizenProfileRepo) GetByUserID(userID string) (*models.CitizenProfile, error) {
	var p models.CitizenProfile
	if err := r.db.Where("user_id = ?", userID).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *citizenProfileRepo) Update(profile *models.CitizenProfile) error {
	return r.db.Save(profile).Error
}

func (r *citizenProfileRepo) CreateInTx(tx *gorm.DB, profile *models.CitizenProfile) error {
	return tx.Create(profile).Error
}

func (r *citizenProfileRepo) FindByCitizenIDNumber(cccd string) (*models.CitizenProfile, error) {
	var p models.CitizenProfile
	if err := r.db.Where("citizen_id_number = ? AND deleted_at IS NULL", cccd).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *citizenProfileRepo) ListAllForExport(offset, limit int) ([]CitizenExportRow, int64, error) {
	type rawRow struct {
		SoCCCD      string
		HoTen       string
		Email       string
		SoDienThoai string
		DiaChi      string
		NgaySinh    string
		GioiTinh    string
		DiaChiThuongTru string
		TongHoSo    int64
	}

	var total int64
	if err := r.db.Model(&models.CitizenProfile{}).
		Joins("JOIN users ON users.id = citizen_profiles.user_id AND users.deleted_at IS NULL").
		Where("citizen_profiles.deleted_at IS NULL").
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := make([]rawRow, 0)
	err := r.db.Table("citizen_profiles cp").
		Select(`cp.citizen_id_number AS so_cccd,
			u.name AS ho_ten,
			u.email AS email,
			u.phone AS so_dien_thoai,
			u.address AS dia_chi,
			COALESCE(TO_CHAR(cp.date_of_birth, 'DD/MM/YYYY'), '') AS ngay_sinh,
			COALESCE(cp.gender, '') AS gioi_tinh,
			COALESCE(cp.permanent_address, '') AS dia_chi_thuong_tru,
			COUNT(a.id) AS tong_ho_so`).
		Joins("JOIN users u ON u.id = cp.user_id AND u.deleted_at IS NULL").
		Joins("LEFT JOIN applications a ON a.citizen_user_id = cp.user_id AND a.deleted_at IS NULL").
		Where("cp.deleted_at IS NULL").
		Group("cp.citizen_id_number, u.name, u.email, u.phone, u.address, cp.date_of_birth, cp.gender, cp.permanent_address").
		Order("u.name ASC").
		Offset(offset).Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	out := make([]CitizenExportRow, len(rows))
	for i, r := range rows {
		out[i] = CitizenExportRow{
			SoCCCD:      r.SoCCCD,
			HoTen:       r.HoTen,
			Email:       r.Email,
			SoDienThoai: r.SoDienThoai,
			DiaChi:      r.DiaChi,
			NgaySinh:    r.NgaySinh,
			GioiTinh:    r.GioiTinh,
			DiaChiThuongTru: r.DiaChiThuongTru,
			TongHoSo:    r.TongHoSo,
		}
	}
	return out, total, nil
}
