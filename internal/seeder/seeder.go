package seeder

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"gorm.io/gorm"
)

func SeedDatabase(db *gorm.DB) error {
	fmt.Println("🌱 Starting database seeding...")

	if err := seedUsers(db); err != nil {
		return fmt.Errorf("error seeding users: %w", err)
	}

	if err := seedDepartments(db); err != nil {
		return fmt.Errorf("error seeding departments: %w", err)
	}

	if err := seedServiceTypes(db); err != nil {
		return fmt.Errorf("error seeding service types: %w", err)
	}

	if err := seedCitizenProfiles(db); err != nil {
		return fmt.Errorf("error seeding citizen profiles: %w", err)
	}

	if err := seedStaffProfiles(db); err != nil {
		return fmt.Errorf("error seeding staff profiles: %w", err)
	}

	if err := seedApplications(db); err != nil {
		return fmt.Errorf("error seeding applications: %w", err)
	}

	if err := seedNotifications(db); err != nil {
		return fmt.Errorf("error seeding notifications: %w", err)
	}

	fmt.Println("✅ Database seeding completed successfully!")
	return nil
}

func findUserByEmail(db *gorm.DB, email string) (models.User, error) {
	var u models.User
	err := db.Where("email = ?", email).First(&u).Error
	return u, err
}

func findDeptByCode(db *gorm.DB, code string) (models.Department, error) {
	var d models.Department
	err := db.Where("code = ?", code).First(&d).Error
	return d, err
}

func findServiceByCode(db *gorm.DB, code string) (models.ServiceType, error) {
	var s models.ServiceType
	err := db.Where("code = ?", code).First(&s).Error
	return s, err
}

func seedUsers(db *gorm.DB) error {
	hash := "$2a$10$nFfLjD3IXTX8j45eYDNF5.zxZzxoUTBU4Wklu6M3NFUojJtnYkukS"
	users := []models.User{
		// Admin & Manager
		{Name: "Nguyễn Văn An", Email: "admin@example.com", PasswordHash: hash, Phone: "0912345678", Address: "123 Đường Lê Lợi, Quận 1, TP. Hồ Chí Minh", Role: models.UserRoleSuperAdmin, Status: models.UserStatusActive},
		{Name: "Trần Thị Bình", Email: "manager@example.com", PasswordHash: hash, Phone: "0912345679", Address: "456 Đường Nguyễn Huệ, Quận 1, TP. Hồ Chí Minh", Role: models.UserRoleManager, Status: models.UserStatusActive},

		// Staff 1–8 (existing)
		{Name: "Phạm Minh Tuấn", Email: "staff1@example.com", PasswordHash: hash, Phone: "0912345680", Address: "789 Đường Tôn Đức Thắng, Quận 1, TP. Hồ Chí Minh", Role: models.UserRoleStaff, Status: models.UserStatusActive},
		{Name: "Hoàng Thị Liên", Email: "staff2@example.com", PasswordHash: hash, Phone: "0912345681", Address: "321 Đường Độc Lập, Quận 1, TP. Hồ Chí Minh", Role: models.UserRoleStaff, Status: models.UserStatusActive},
		{Name: "Lê Quốc Việt", Email: "staff3@example.com", PasswordHash: hash, Phone: "0912345684", Address: "12 Trần Hưng Đạo, Quận 1, TP. Hồ Chí Minh", Role: models.UserRoleStaff, Status: models.UserStatusActive},
		{Name: "Nguyễn Gia Hân", Email: "staff4@example.com", PasswordHash: hash, Phone: "0912345685", Address: "45 Võ Văn Tần, Quận 3, TP. Hồ Chí Minh", Role: models.UserRoleStaff, Status: models.UserStatusActive},
		{Name: "Trần Minh Khoa", Email: "staff5@example.com", PasswordHash: hash, Phone: "0912345686", Address: "88 Nguyễn Đình Chiểu, Quận 1, TP. Hồ Chí Minh", Role: models.UserRoleStaff, Status: models.UserStatusActive},
		{Name: "Đỗ Thị Mai", Email: "staff6@example.com", PasswordHash: hash, Phone: "0912345687", Address: "19 Lê Văn Sỹ, Quận Phú Nhuận, TP. Hồ Chí Minh", Role: models.UserRoleStaff, Status: models.UserStatusActive},
		{Name: "Phạm Đức Long", Email: "staff7@example.com", PasswordHash: hash, Phone: "0912345688", Address: "201 Cách Mạng Tháng 8, Quận 10, TP. Hồ Chí Minh", Role: models.UserRoleStaff, Status: models.UserStatusActive},
		{Name: "Bùi Thanh Nhân", Email: "staff8@example.com", PasswordHash: hash, Phone: "0912345689", Address: "77 Cộng Hòa, Quận Tân Bình, TP. Hồ Chí Minh", Role: models.UserRoleStaff, Status: models.UserStatusActive},

		// Staff 9–15 (new)
		{Name: "Lý Thị Hoa", Email: "staff9@example.com", PasswordHash: hash, Phone: "0912345690", Address: "33 Đinh Tiên Hoàng, Bình Thạnh, TP. Hồ Chí Minh", Role: models.UserRoleStaff, Status: models.UserStatusActive},
		{Name: "Võ Văn Hùng", Email: "staff10@example.com", PasswordHash: hash, Phone: "0912345691", Address: "67 Phan Văn Hân, Bình Thạnh, TP. Hồ Chí Minh", Role: models.UserRoleStaff, Status: models.UserStatusActive},
		{Name: "Ngô Thị Kim Anh", Email: "staff11@example.com", PasswordHash: hash, Phone: "0912345692", Address: "15 Xô Viết Nghệ Tĩnh, Bình Thạnh, TP. Hồ Chí Minh", Role: models.UserRoleStaff, Status: models.UserStatusActive},
		{Name: "Đặng Văn Tú", Email: "staff12@example.com", PasswordHash: hash, Phone: "0912345693", Address: "28 Ung Văn Khiêm, Bình Thạnh, TP. Hồ Chí Minh", Role: models.UserRoleStaff, Status: models.UserStatusActive},
		{Name: "Trịnh Thị Lan", Email: "staff13@example.com", PasswordHash: hash, Phone: "0912345694", Address: "90 Lê Quang Định, Bình Thạnh, TP. Hồ Chí Minh", Role: models.UserRoleStaff, Status: models.UserStatusActive},
		{Name: "Huỳnh Minh Đức", Email: "staff14@example.com", PasswordHash: hash, Phone: "0912345695", Address: "44 Phạm Văn Đồng, Gò Vấp, TP. Hồ Chí Minh", Role: models.UserRoleStaff, Status: models.UserStatusActive},
		{Name: "Phan Thị Tuyết", Email: "staff15@example.com", PasswordHash: hash, Phone: "0912345696", Address: "102 Nguyễn Oanh, Gò Vấp, TP. Hồ Chí Minh", Role: models.UserRoleStaff, Status: models.UserStatusActive},

		// Citizens 1–15
		{Name: "Vũ Thành Công", Email: "citizen1@example.com", PasswordHash: hash, Phone: "0912345682", Address: "555 Đường Lạc Long Quân, Quận 5, TP. Hồ Chí Minh", Role: models.UserRoleCitizen, Status: models.UserStatusActive},
		{Name: "Đặng Hữu Minh", Email: "citizen2@example.com", PasswordHash: hash, Phone: "0912345683", Address: "666 Đường Phan Đình Phùng, Quận 3, TP. Hồ Chí Minh", Role: models.UserRoleCitizen, Status: models.UserStatusActive},
		{Name: "Lê Thị Thu Hương", Email: "citizen3@example.com", PasswordHash: hash, Phone: "0901111001", Address: "10 Hoa Mai, Phú Nhuận, TP. Hồ Chí Minh", Role: models.UserRoleCitizen, Status: models.UserStatusActive},
		{Name: "Nguyễn Hồng Phúc", Email: "citizen4@example.com", PasswordHash: hash, Phone: "0901111002", Address: "22 Trường Chinh, Tân Bình, TP. Hồ Chí Minh", Role: models.UserRoleCitizen, Status: models.UserStatusActive},
		{Name: "Phạm Quốc Toàn", Email: "citizen5@example.com", PasswordHash: hash, Phone: "0901111003", Address: "58 Lý Thường Kiệt, Quận 10, TP. Hồ Chí Minh", Role: models.UserRoleCitizen, Status: models.UserStatusActive},
		{Name: "Trần Thị Ngọc Mai", Email: "citizen6@example.com", PasswordHash: hash, Phone: "0901111004", Address: "9 Bà Huyện Thanh Quan, Quận 3, TP. Hồ Chí Minh", Role: models.UserRoleCitizen, Status: models.UserStatusActive},
		{Name: "Hoàng Văn Dũng", Email: "citizen7@example.com", PasswordHash: hash, Phone: "0901111005", Address: "71 Nguyễn Thị Minh Khai, Quận 1, TP. Hồ Chí Minh", Role: models.UserRoleCitizen, Status: models.UserStatusActive},
		{Name: "Bùi Thị Lan Anh", Email: "citizen8@example.com", PasswordHash: hash, Phone: "0901111006", Address: "34 Cô Giang, Quận 1, TP. Hồ Chí Minh", Role: models.UserRoleCitizen, Status: models.UserStatusActive},
		{Name: "Đỗ Minh Tuấn", Email: "citizen9@example.com", PasswordHash: hash, Phone: "0901111007", Address: "5 Đinh Tiên Hoàng, Quận 1, TP. Hồ Chí Minh", Role: models.UserRoleCitizen, Status: models.UserStatusActive},
		{Name: "Lý Văn Bình", Email: "citizen10@example.com", PasswordHash: hash, Phone: "0901111008", Address: "88 Hai Bà Trưng, Quận 1, TP. Hồ Chí Minh", Role: models.UserRoleCitizen, Status: models.UserStatusActive},
		{Name: "Võ Thị Kim Linh", Email: "citizen11@example.com", PasswordHash: hash, Phone: "0901111009", Address: "120 Điện Biên Phủ, Bình Thạnh, TP. Hồ Chí Minh", Role: models.UserRoleCitizen, Status: models.UserStatusActive},
		{Name: "Ngô Quang Hải", Email: "citizen12@example.com", PasswordHash: hash, Phone: "0901111010", Address: "47 Nơ Trang Long, Bình Thạnh, TP. Hồ Chí Minh", Role: models.UserRoleCitizen, Status: models.UserStatusActive},
		{Name: "Trương Thị Phương", Email: "citizen13@example.com", PasswordHash: hash, Phone: "0901111011", Address: "63 Nguyễn Xí, Bình Thạnh, TP. Hồ Chí Minh", Role: models.UserRoleCitizen, Status: models.UserStatusActive},
		{Name: "Lưu Thanh Sơn", Email: "citizen14@example.com", PasswordHash: hash, Phone: "0901111012", Address: "18 Bạch Đằng, Bình Thạnh, TP. Hồ Chí Minh", Role: models.UserRoleCitizen, Status: models.UserStatusActive},
		{Name: "Mai Thị Bích Ngọc", Email: "citizen15@example.com", PasswordHash: hash, Phone: "0901111013", Address: "200 Phan Xích Long, Phú Nhuận, TP. Hồ Chí Minh", Role: models.UserRoleCitizen, Status: models.UserStatusActive},
	}

	for i := range users {
		users[i].CreatedAt = time.Now()
		users[i].UpdatedAt = time.Now()
		if err := db.Where("email = ?", users[i].Email).FirstOrCreate(&users[i]).Error; err != nil {
			return err
		}
	}

	fmt.Println("✓ Users seeded")
	return nil
}

func seedDepartments(db *gorm.DB) error {
	staff1, err := findUserByEmail(db, "staff1@example.com")
	if err != nil {
		return fmt.Errorf("lookup staff1: %w", err)
	}
	staff2, err := findUserByEmail(db, "staff2@example.com")
	if err != nil {
		return fmt.Errorf("lookup staff2: %w", err)
	}
	staff7, err := findUserByEmail(db, "staff7@example.com")
	if err != nil {
		return fmt.Errorf("lookup staff7: %w", err)
	}
	staff10, err := findUserByEmail(db, "staff10@example.com")
	if err != nil {
		return fmt.Errorf("lookup staff10: %w", err)
	}
	staff12, err := findUserByEmail(db, "staff12@example.com")
	if err != nil {
		return fmt.Errorf("lookup staff12: %w", err)
	}
	staff14, err := findUserByEmail(db, "staff14@example.com")
	if err != nil {
		return fmt.Errorf("lookup staff14: %w", err)
	}

	departments := []models.Department{
		{
			Name:         "Phòng Cấp Giấy Tờ Tùy Thân",
			Code:         "PGTT",
			Address:      "123 Đường Lê Lợi, Quận 1, TP. Hồ Chí Minh",
			LeaderUserID: &staff1.ID,
		},
		{
			Name:         "Phòng Đăng Ký Xe Cơ Giới",
			Code:         "PDXCG",
			Address:      "456 Đường Nguyễn Huệ, Quận 1, TP. Hồ Chí Minh",
			LeaderUserID: &staff2.ID,
		},
		{
			Name:         "Phòng Cấp Giấy Phép Lái Xe",
			Code:         "PGPLX",
			Address:      "789 Đường Tôn Đức Thắng, Quận 1, TP. Hồ Chí Minh",
			LeaderUserID: &staff7.ID,
		},
		{
			Name:         "Phòng Đăng Ký Kinh Doanh",
			Code:         "PDKKD",
			Address:      "100 Đường Pasteur, Quận 3, TP. Hồ Chí Minh",
			LeaderUserID: &staff10.ID,
		},
		{
			Name:         "Phòng Tư Pháp",
			Code:         "PTP",
			Address:      "55 Đường Lý Tự Trọng, Quận 1, TP. Hồ Chí Minh",
			LeaderUserID: &staff12.ID,
		},
		{
			Name:         "Phòng Lao Động Thương Binh",
			Code:         "PLDTB",
			Address:      "88 Đường Nguyễn Thị Minh Khai, Quận 3, TP. Hồ Chí Minh",
			LeaderUserID: &staff14.ID,
		},
	}

	for i := range departments {
		departments[i].CreatedAt = time.Now()
		departments[i].UpdatedAt = time.Now()
		if err := db.Where("code = ?", departments[i].Code).FirstOrCreate(&departments[i]).Error; err != nil {
			return err
		}
	}

	fmt.Println("✓ Departments seeded")
	return nil
}

func seedServiceTypes(db *gorm.DB) error {
	deptGTTT, err := findDeptByCode(db, "PGTT")
	if err != nil {
		return fmt.Errorf("lookup dept PGTT: %w", err)
	}
	deptXCG, err := findDeptByCode(db, "PDXCG")
	if err != nil {
		return fmt.Errorf("lookup dept PDXCG: %w", err)
	}
	deptGPLX, err := findDeptByCode(db, "PGPLX")
	if err != nil {
		return fmt.Errorf("lookup dept PGPLX: %w", err)
	}
	deptKD, err := findDeptByCode(db, "PDKKD")
	if err != nil {
		return fmt.Errorf("lookup dept PDKKD: %w", err)
	}
	deptTP, err := findDeptByCode(db, "PTP")
	if err != nil {
		return fmt.Errorf("lookup dept PTP: %w", err)
	}
	deptLD, err := findDeptByCode(db, "PLDTB")
	if err != nil {
		return fmt.Errorf("lookup dept PLDTB: %w", err)
	}

	schemaID, _ := json.Marshal(map[string]interface{}{
		"name":            "Mẫu đơn cấp CCCD",
		"fields":          []string{"full_name", "date_of_birth", "gender", "nationality"},
		"required_fields": []string{"full_name", "date_of_birth", "gender"},
	})
	schemaLicense, _ := json.Marshal(map[string]interface{}{
		"name":            "Mẫu đơn cấp Giấy phép lái xe",
		"fields":          []string{"license_class", "experience_years", "medical_exam_date"},
		"required_fields": []string{"license_class", "medical_exam_date"},
	})
	schemaBike, _ := json.Marshal(map[string]interface{}{
		"name":            "Mẫu đơn đăng ký xe",
		"fields":          []string{"vehicle_type", "vehicle_brand", "chassis_no", "engine_no"},
		"required_fields": []string{"vehicle_type", "chassis_no", "engine_no"},
	})
	schemaEdu, _ := json.Marshal(map[string]interface{}{
		"name":            "Mẫu đơn đăng ký nhập học",
		"fields":          []string{"student_name", "date_of_birth", "school_name", "grade"},
		"required_fields": []string{"student_name", "date_of_birth", "school_name", "grade"},
	})
	schemaHealth, _ := json.Marshal(map[string]interface{}{
		"name":            "Mẫu đơn đăng ký khám sức khỏe",
		"fields":          []string{"full_name", "date_of_birth", "health_insurance_no", "preferred_date"},
		"required_fields": []string{"full_name", "date_of_birth"},
	})
	schemaBiz, _ := json.Marshal(map[string]interface{}{
		"name":            "Mẫu đơn đăng ký kinh doanh",
		"fields":          []string{"business_name", "business_type", "capital", "address"},
		"required_fields": []string{"business_name", "business_type", "address"},
	})
	schemaMarriage, _ := json.Marshal(map[string]interface{}{
		"name":            "Mẫu đơn đăng ký kết hôn",
		"fields":          []string{"groom_name", "bride_name", "wedding_date", "ceremony_address"},
		"required_fields": []string{"groom_name", "bride_name", "wedding_date"},
	})
	schemaBirth, _ := json.Marshal(map[string]interface{}{
		"name":            "Mẫu đơn đăng ký khai sinh",
		"fields":          []string{"child_name", "date_of_birth", "father_name", "mother_name"},
		"required_fields": []string{"child_name", "date_of_birth", "mother_name"},
	})
	schemaJob, _ := json.Marshal(map[string]interface{}{
		"name":            "Mẫu đơn trợ cấp thất nghiệp",
		"fields":          []string{"full_name", "last_employer", "termination_date", "bank_account"},
		"required_fields": []string{"full_name", "last_employer", "termination_date"},
	})

	serviceTypes := []models.ServiceType{
		{
			Name: "Cấp CCCD lần đầu", Code: "CCCD_NEW",
			Description:             "Cấp Căn cước công dân lần đầu cho công dân đủ 14 tuổi",
			RequiredDocuments:       "Giấy khai sinh, Chứng minh thư hoặc Hộ chiếu, Ảnh màu 3x4",
			FormSchema:              schemaID, ProcessingTime: intPtr(3), Fee: 0,
			ResponsibleDepartmentID: &deptGTTT.ID, IsActive: true,
		},
		{
			Name: "Đổi CCCD (cập nhật thông tin)", Code: "CCCD_RENEW",
			Description:             "Đổi Căn cước công dân do thay đổi thông tin cá nhân",
			RequiredDocuments:       "CCCD cũ, Giấy tờ chứng minh thay đổi thông tin",
			FormSchema:              schemaID, ProcessingTime: intPtr(5), Fee: 50000,
			ResponsibleDepartmentID: &deptGTTT.ID, IsActive: true,
		},
		{
			Name: "Cấp Giấy phép lái xe hạng A", Code: "LICENSE_CLASS_A",
			Description:             "Cấp Giấy phép lái xe hạng A (xe máy)",
			RequiredDocuments:       "CCCD/Hộ chiếu, Giấy chứng nhận sức khỏe, 4 ảnh 3x4",
			FormSchema:              schemaLicense, ProcessingTime: intPtr(5), Fee: 70000,
			ResponsibleDepartmentID: &deptXCG.ID, IsActive: true,
		},
		{
			Name: "Cấp Giấy phép lái xe hạng C", Code: "LICENSE_CLASS_C",
			Description:             "Cấp Giấy phép lái xe hạng C (ô tô nhỏ)",
			RequiredDocuments:       "CCCD/Hộ chiếu, Giấy chứng nhận sức khỏe, 4 ảnh 3x4",
			FormSchema:              schemaLicense, ProcessingTime: intPtr(7), Fee: 150000,
			ResponsibleDepartmentID: &deptXCG.ID, IsActive: true,
		},
		{
			Name: "Cấp Giấy phép lái xe hạng B", Code: "LICENSE_CLASS_B",
			Description:             "Cấp Giấy phép lái xe hạng B (ô tô dưới 9 chỗ)",
			RequiredDocuments:       "CCCD/Hộ chiếu, Giấy chứng nhận sức khỏe, 4 ảnh 3x4",
			FormSchema:              schemaLicense, ProcessingTime: intPtr(7), Fee: 135000,
			ResponsibleDepartmentID: &deptGPLX.ID, IsActive: true,
		},
		{
			Name: "Đăng ký xe máy", Code: "REGISTER_BIKE",
			Description:       "Đăng ký xe máy tại Cục Đăng ký Lái xe và Xe cơ giới",
			RequiredDocuments: "Hóa đơn bán hàng, CCCD, Bảng kiểm tra kỹ thuật",
			FormSchema:        schemaBike, ProcessingTime: intPtr(1), Fee: 50000,
			IsActive: true,
		},
		{
			Name: "Đăng ký xe ô tô", Code: "REGISTER_CAR",
			Description:       "Đăng ký xe ô tô tại Cục Đăng ký Lái xe và Xe cơ giới",
			RequiredDocuments: "Hóa đơn bán hàng, CCCD, Bảng kiểm tra kỹ thuật, Bảo hiểm",
			FormSchema:        schemaBike, ProcessingTime: intPtr(3), Fee: 100000,
			IsActive: true,
		},
		{
			Name: "Đăng ký nhập học trường công lập", Code: "EDU_ENROLL",
			Description:       "Đăng ký nhập học cho học sinh vào trường tiểu học và trung học công lập",
			RequiredDocuments: "Giấy khai sinh, Hộ khẩu hoặc Giấy xác nhận cư trú, Ảnh 3x4",
			FormSchema:        schemaEdu, ProcessingTime: intPtr(2), Fee: 0,
			IsActive: true,
		},
		{
			Name: "Đăng ký khám sức khỏe định kỳ", Code: "HEALTH_CHECK",
			Description:       "Đăng ký dịch vụ khám sức khỏe định kỳ tại cơ sở y tế công lập",
			RequiredDocuments: "CCCD, Thẻ bảo hiểm y tế",
			FormSchema:        schemaHealth, ProcessingTime: intPtr(1), Fee: 30000,
			IsActive: true,
		},
		{
			Name: "Đăng ký kinh doanh hộ cá thể", Code: "BIZ_REGISTER",
			Description:             "Đăng ký kinh doanh hộ cá thể và cấp giấy phép kinh doanh",
			RequiredDocuments:       "CCCD, Hộ khẩu, Đơn đăng ký kinh doanh, Hợp đồng thuê mặt bằng",
			FormSchema:              schemaBiz, ProcessingTime: intPtr(5), Fee: 200000,
			ResponsibleDepartmentID: &deptKD.ID, IsActive: true,
		},
		{
			Name: "Gia hạn giấy phép kinh doanh", Code: "BIZ_RENEW",
			Description:             "Gia hạn giấy phép kinh doanh hộ cá thể",
			RequiredDocuments:       "CCCD, Giấy phép kinh doanh cũ, Biên lai nộp thuế",
			FormSchema:              schemaBiz, ProcessingTime: intPtr(3), Fee: 100000,
			ResponsibleDepartmentID: &deptKD.ID, IsActive: true,
		},
		{
			Name: "Đăng ký kết hôn", Code: "MARRIAGE_REG",
			Description:             "Đăng ký kết hôn tại cơ quan hộ tịch",
			RequiredDocuments:       "CCCD hai bên, Giấy xác nhận tình trạng hôn nhân, Ảnh 4x6",
			FormSchema:              schemaMarriage, ProcessingTime: intPtr(5), Fee: 0,
			ResponsibleDepartmentID: &deptTP.ID, IsActive: true,
		},
		{
			Name: "Đăng ký khai sinh", Code: "BIRTH_REG",
			Description:             "Đăng ký khai sinh cho trẻ em",
			RequiredDocuments:       "CCCD bố/mẹ, Giấy chứng sinh, Hộ khẩu gia đình",
			FormSchema:              schemaBirth, ProcessingTime: intPtr(3), Fee: 0,
			ResponsibleDepartmentID: &deptTP.ID, IsActive: true,
		},
		{
			Name: "Trợ cấp thất nghiệp", Code: "UNEMPLOYMENT",
			Description:             "Đăng ký hưởng trợ cấp thất nghiệp",
			RequiredDocuments:       "CCCD, Quyết định thôi việc, Sổ bảo hiểm xã hội",
			FormSchema:              schemaJob, ProcessingTime: intPtr(10), Fee: 0,
			ResponsibleDepartmentID: &deptLD.ID, IsActive: true,
		},
	}

	for i := range serviceTypes {
		serviceTypes[i].CreatedAt = time.Now()
		serviceTypes[i].UpdatedAt = time.Now()
		if err := db.Where("code = ?", serviceTypes[i].Code).FirstOrCreate(&serviceTypes[i]).Error; err != nil {
			return err
		}
	}

	fmt.Println("✓ Service Types seeded")
	return nil
}

func seedCitizenProfiles(db *gorm.DB) error {
	type profileSeed struct {
		email    string
		idNum    string
		dob      time.Time
		gender   string
		address  string
	}
	seeds := []profileSeed{
		{"citizen1@example.com", "123456789012", time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC), "Nam", "555 Đường Lạc Long Quân, Phường 9, Quận 5, TP. Hồ Chí Minh"},
		{"citizen2@example.com", "987654321098", time.Date(1985, 10, 20, 0, 0, 0, 0, time.UTC), "Nam", "666 Đường Phan Đình Phùng, Phường 1, Quận 3, TP. Hồ Chí Minh"},
		{"citizen3@example.com", "111222333444", time.Date(1992, 3, 8, 0, 0, 0, 0, time.UTC), "Nữ", "10 Hoa Mai, Phường 2, Phú Nhuận, TP. Hồ Chí Minh"},
		{"citizen4@example.com", "222333444555", time.Date(1988, 7, 22, 0, 0, 0, 0, time.UTC), "Nam", "22 Trường Chinh, Phường 12, Tân Bình, TP. Hồ Chí Minh"},
		{"citizen5@example.com", "333444555666", time.Date(1995, 1, 30, 0, 0, 0, 0, time.UTC), "Nam", "58 Lý Thường Kiệt, Phường 7, Quận 10, TP. Hồ Chí Minh"},
		{"citizen6@example.com", "444555666777", time.Date(1993, 11, 14, 0, 0, 0, 0, time.UTC), "Nữ", "9 Bà Huyện Thanh Quan, Phường 6, Quận 3, TP. Hồ Chí Minh"},
		{"citizen7@example.com", "555666777888", time.Date(1980, 4, 5, 0, 0, 0, 0, time.UTC), "Nam", "71 Nguyễn Thị Minh Khai, Phường 6, Quận 1, TP. Hồ Chí Minh"},
		{"citizen8@example.com", "666777888999", time.Date(1998, 9, 17, 0, 0, 0, 0, time.UTC), "Nữ", "34 Cô Giang, Phường Cô Giang, Quận 1, TP. Hồ Chí Minh"},
		{"citizen9@example.com", "777888999000", time.Date(1975, 6, 12, 0, 0, 0, 0, time.UTC), "Nam", "5 Đinh Tiên Hoàng, Phường Đa Kao, Quận 1, TP. Hồ Chí Minh"},
		{"citizen10@example.com", "888999000111", time.Date(2000, 2, 28, 0, 0, 0, 0, time.UTC), "Nam", "88 Hai Bà Trưng, Phường Bến Nghé, Quận 1, TP. Hồ Chí Minh"},
		{"citizen11@example.com", "999000111222", time.Date(1991, 8, 3, 0, 0, 0, 0, time.UTC), "Nữ", "120 Điện Biên Phủ, Phường 15, Bình Thạnh, TP. Hồ Chí Minh"},
		{"citizen12@example.com", "000111222333", time.Date(1987, 12, 25, 0, 0, 0, 0, time.UTC), "Nam", "47 Nơ Trang Long, Phường 14, Bình Thạnh, TP. Hồ Chí Minh"},
		{"citizen13@example.com", "112233445566", time.Date(1996, 5, 19, 0, 0, 0, 0, time.UTC), "Nữ", "63 Nguyễn Xí, Phường 26, Bình Thạnh, TP. Hồ Chí Minh"},
		{"citizen14@example.com", "223344556677", time.Date(1983, 9, 7, 0, 0, 0, 0, time.UTC), "Nam", "18 Bạch Đằng, Phường 2, Bình Thạnh, TP. Hồ Chí Minh"},
		{"citizen15@example.com", "334455667788", time.Date(1999, 3, 15, 0, 0, 0, 0, time.UTC), "Nữ", "200 Phan Xích Long, Phường 2, Phú Nhuận, TP. Hồ Chí Minh"},
	}

	for _, s := range seeds {
		user, err := findUserByEmail(db, s.email)
		if err != nil {
			return fmt.Errorf("lookup %s: %w", s.email, err)
		}
		dob := s.dob
		profile := models.CitizenProfile{
			UserID:                   user.ID,
			CitizenIDNumber:          s.idNum,
			DateOfBirth:              &dob,
			Gender:                   s.gender,
			PermanentAddress:         s.address,
			EmailNotificationEnabled: true,
			CreatedAt:                time.Now(),
			UpdatedAt:                time.Now(),
		}
		if err := db.Where("user_id = ?", profile.UserID).FirstOrCreate(&profile).Error; err != nil {
			return err
		}
	}

	fmt.Println("✓ Citizen Profiles seeded")
	return nil
}

func seedStaffProfiles(db *gorm.DB) error {
	type profileSeed struct {
		email    string
		deptCode *string
		position string
	}

	deptPGTT := "PGTT"
	deptPDXCG := "PDXCG"
	deptPGPLX := "PGPLX"
	deptPDKKD := "PDKKD"
	deptPTP := "PTP"
	deptPLDTB := "PLDTB"

	seeds := []profileSeed{
		// PGTT: staff1 (trưởng), staff3, staff4
		{"staff1@example.com", &deptPGTT, "Trưởng phòng"},
		{"staff3@example.com", &deptPGTT, "Chuyên viên"},
		{"staff4@example.com", &deptPGTT, "Chuyên viên"},
		// PDXCG: staff2 (trưởng), staff5, staff6
		{"staff2@example.com", &deptPDXCG, "Trưởng phòng"},
		{"staff5@example.com", &deptPDXCG, "Chuyên viên"},
		{"staff6@example.com", &deptPDXCG, "Chuyên viên"},
		// PGPLX: staff7 (trưởng), staff8, staff9
		{"staff7@example.com", &deptPGPLX, "Trưởng phòng"},
		{"staff8@example.com", &deptPGPLX, "Chuyên viên"},
		{"staff9@example.com", &deptPGPLX, "Chuyên viên"},
		// PDKKD: staff10 (trưởng), staff11
		{"staff10@example.com", &deptPDKKD, "Trưởng phòng"},
		{"staff11@example.com", &deptPDKKD, "Chuyên viên"},
		// PTP: staff12 (trưởng), staff13
		{"staff12@example.com", &deptPTP, "Trưởng phòng"},
		{"staff13@example.com", &deptPTP, "Chuyên viên"},
		// PLDTB: staff14 (trưởng), staff15
		{"staff14@example.com", &deptPLDTB, "Trưởng phòng"},
		{"staff15@example.com", &deptPLDTB, "Chuyên viên"},
	}

	for _, s := range seeds {
		user, err := findUserByEmail(db, s.email)
		if err != nil {
			return fmt.Errorf("lookup %s: %w", s.email, err)
		}

		var deptID *string
		if s.deptCode != nil {
			dept, err := findDeptByCode(db, *s.deptCode)
			if err != nil {
				return fmt.Errorf("lookup dept %s: %w", *s.deptCode, err)
			}
			deptID = &dept.ID
		}

		profile := models.StaffProfile{
			UserID:       user.ID,
			DepartmentID: deptID,
			Position:     s.position,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if err := db.Where("user_id = ?", profile.UserID).FirstOrCreate(&profile).Error; err != nil {
			return err
		}
	}

	fmt.Println("✓ Staff Profiles seeded")
	return nil
}

func seedApplications(db *gorm.DB) error {
	type appSeed struct {
		code        string
		citizenEmail string
		serviceCode  string
		staffEmail   *string
		status       models.ApplicationStatus
		resultNote   string
		rejectedNote string
		submittedAt  time.Time
		startedAt    *time.Time
		completedAt  *time.Time
	}

	s1 := "staff1@example.com"
	s2 := "staff2@example.com"
	s3 := "staff3@example.com"
	s5 := "staff5@example.com"
	s7 := "staff7@example.com"
	s10 := "staff10@example.com"
	s12 := "staff12@example.com"
	s14 := "staff14@example.com"

	now := time.Now()
	d := func(days int) time.Time { return now.AddDate(0, 0, days) }
	dp := func(days int) *time.Time { t := d(days); return &t }

	seeds := []appSeed{
		{"HCM-2026-001", "citizen1@example.com", "CCCD_NEW", &s1, models.ApplicationStatusProcessing, "Đã tiếp nhận và đang xử lý hồ sơ.", "", d(-7), dp(-1), nil},
		{"HCM-2026-002", "citizen2@example.com", "LICENSE_CLASS_A", nil, models.ApplicationStatusReceived, "", "", d(-2), nil, nil},
		{"HCM-2026-003", "citizen1@example.com", "LICENSE_CLASS_C", &s1, models.ApplicationStatusNeedMoreInfo, "Vui lòng bổ sung giấy khám sức khỏe bản gốc.", "", d(-4), dp(-3), nil},
		{"HCM-2026-004", "citizen2@example.com", "LICENSE_CLASS_C", &s2, models.ApplicationStatusApproved, "Đã phê duyệt hồ sơ.", "", d(-30), dp(-25), dp(-5)},
		{"HCM-2026-005", "citizen1@example.com", "CCCD_NEW", &s2, models.ApplicationStatusRejected, "", "Thông tin giấy tờ không khớp hồ sơ gốc.", d(-10), dp(-9), dp(-8)},
		{"HCM-2026-006", "citizen3@example.com", "CCCD_RENEW", &s3, models.ApplicationStatusProcessing, "Đang xác minh thông tin.", "", d(-5), dp(-2), nil},
		{"HCM-2026-007", "citizen4@example.com", "REGISTER_BIKE", nil, models.ApplicationStatusReceived, "", "", d(-1), nil, nil},
		{"HCM-2026-008", "citizen5@example.com", "BIZ_REGISTER", &s10, models.ApplicationStatusProcessing, "Đang thẩm định hồ sơ đăng ký kinh doanh.", "", d(-6), dp(-3), nil},
		{"HCM-2026-009", "citizen6@example.com", "MARRIAGE_REG", &s12, models.ApplicationStatusApproved, "Đã xác nhận và cấp giấy đăng ký kết hôn.", "", d(-20), dp(-18), dp(-15)},
		{"HCM-2026-010", "citizen7@example.com", "BIRTH_REG", &s12, models.ApplicationStatusApproved, "Khai sinh đã được đăng ký thành công.", "", d(-15), dp(-14), dp(-12)},
		{"HCM-2026-011", "citizen8@example.com", "UNEMPLOYMENT", &s14, models.ApplicationStatusProcessing, "Đang xét duyệt hồ sơ thất nghiệp.", "", d(-8), dp(-5), nil},
		{"HCM-2026-012", "citizen9@example.com", "HEALTH_CHECK", nil, models.ApplicationStatusReceived, "", "", d(-1), nil, nil},
		{"HCM-2026-013", "citizen10@example.com", "EDU_ENROLL", nil, models.ApplicationStatusReceived, "", "", d(-3), nil, nil},
		{"HCM-2026-014", "citizen11@example.com", "LICENSE_CLASS_B", &s7, models.ApplicationStatusProcessing, "Đang kiểm tra hồ sơ bằng lái B.", "", d(-9), dp(-6), nil},
		{"HCM-2026-015", "citizen12@example.com", "REGISTER_CAR", &s5, models.ApplicationStatusApproved, "Xe đã đăng ký thành công.", "", d(-25), dp(-23), dp(-20)},
		{"HCM-2026-016", "citizen13@example.com", "BIZ_RENEW", &s10, models.ApplicationStatusNeedMoreInfo, "Cần bổ sung biên lai thuế năm ngoái.", "", d(-12), dp(-10), nil},
		{"HCM-2026-017", "citizen14@example.com", "CCCD_NEW", &s3, models.ApplicationStatusApproved, "CCCD đã được cấp.", "", d(-35), dp(-33), dp(-28)},
		{"HCM-2026-018", "citizen15@example.com", "MARRIAGE_REG", nil, models.ApplicationStatusReceived, "", "", d(-2), nil, nil},
		{"HCM-2026-019", "citizen3@example.com", "HEALTH_CHECK", nil, models.ApplicationStatusReceived, "", "", d(-1), nil, nil},
		{"HCM-2026-020", "citizen4@example.com", "BIRTH_REG", &s12, models.ApplicationStatusProcessing, "Đang xử lý hồ sơ khai sinh.", "", d(-4), dp(-2), nil},
		{"HCM-2026-021", "citizen5@example.com", "LICENSE_CLASS_A", &s5, models.ApplicationStatusRejected, "", "Giấy khám sức khỏe hết hạn.", d(-14), dp(-13), dp(-11)},
		{"HCM-2026-022", "citizen6@example.com", "CCCD_NEW", &s1, models.ApplicationStatusApproved, "CCCD lần đầu đã được cấp.", "", d(-40), dp(-38), dp(-35)},
		{"HCM-2026-023", "citizen7@example.com", "UNEMPLOYMENT", &s14, models.ApplicationStatusNeedMoreInfo, "Cần bổ sung quyết định thôi việc có xác nhận.", "", d(-7), dp(-5), nil},
		{"HCM-2026-024", "citizen8@example.com", "EDU_ENROLL", nil, models.ApplicationStatusReceived, "", "", d(-1), nil, nil},
		{"HCM-2026-025", "citizen9@example.com", "BIZ_REGISTER", &s10, models.ApplicationStatusRejected, "", "Địa chỉ kinh doanh không đáp ứng yêu cầu quy hoạch.", d(-18), dp(-16), dp(-14)},
		{"HCM-2026-026", "citizen10@example.com", "LICENSE_CLASS_C", nil, models.ApplicationStatusReceived, "", "", d(-2), nil, nil},
		{"HCM-2026-027", "citizen11@example.com", "REGISTER_BIKE", &s5, models.ApplicationStatusApproved, "Xe máy đã được đăng ký.", "", d(-22), dp(-21), dp(-19)},
		{"HCM-2026-028", "citizen12@example.com", "MARRIAGE_REG", &s12, models.ApplicationStatusApproved, "Đăng ký kết hôn hoàn tất.", "", d(-28), dp(-26), dp(-24)},
		{"HCM-2026-029", "citizen13@example.com", "CCCD_RENEW", nil, models.ApplicationStatusReceived, "", "", d(-1), nil, nil},
		{"HCM-2026-030", "citizen14@example.com", "HEALTH_CHECK", nil, models.ApplicationStatusReceived, "", "", d(-1), nil, nil},
	}

	submittedData, _ := json.Marshal(map[string]interface{}{"submitted": true})

	for _, s := range seeds {
		citizen, err := findUserByEmail(db, s.citizenEmail)
		if err != nil {
			return fmt.Errorf("lookup citizen %s: %w", s.citizenEmail, err)
		}
		svc, err := findServiceByCode(db, s.serviceCode)
		if err != nil {
			return fmt.Errorf("lookup service %s: %w", s.serviceCode, err)
		}

		var staffID *string
		if s.staffEmail != nil {
			staff, err := findUserByEmail(db, *s.staffEmail)
			if err != nil {
				return fmt.Errorf("lookup staff %s: %w", *s.staffEmail, err)
			}
			staffID = &staff.ID
		}

		app := models.Application{
			ApplicationCode:     s.code,
			CitizenUserID:       citizen.ID,
			ServiceTypeID:       svc.ID,
			AssignedStaffUserID: staffID,
			Status:              s.status,
			SubmittedData:       submittedData,
			ResultNote:          s.resultNote,
			RejectedReason:      s.rejectedNote,
			SubmittedAt:         s.submittedAt,
			ProcessingStartedAt: s.startedAt,
			CompletedAt:         s.completedAt,
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}
		if err := db.Where("application_code = ?", app.ApplicationCode).FirstOrCreate(&app).Error; err != nil {
			return err
		}
	}

	fmt.Println("✓ Applications seeded")
	return nil
}

func seedNotifications(db *gorm.DB) error {
	citizen1, err := findUserByEmail(db, "citizen1@example.com")
	if err != nil {
		return fmt.Errorf("lookup citizen1: %w", err)
	}
	citizen2, err := findUserByEmail(db, "citizen2@example.com")
	if err != nil {
		return fmt.Errorf("lookup citizen2: %w", err)
	}
	citizen3, err := findUserByEmail(db, "citizen3@example.com")
	if err != nil {
		return fmt.Errorf("lookup citizen3: %w", err)
	}

	var app1 models.Application
	db.Where("application_code = ?", "HCM-2026-001").First(&app1)
	var app4 models.Application
	db.Where("application_code = ?", "HCM-2026-004").First(&app4)

	notifications := []models.Notification{
		{
			UserID: citizen1.ID, ApplicationID: &app1.ID,
			Title:     "Cập nhật trạng thái hồ sơ",
			Message:   "Hồ sơ HCM-2026-001 của bạn đã chuyển sang giai đoạn xử lý. Vui lòng chờ kết quả.",
			Type:      models.NotificationTypeReceived,
			IsRead:    false,
			CreatedAt: time.Now().AddDate(0, 0, -1),
		},
		{
			UserID: citizen2.ID, ApplicationID: &app4.ID,
			Title:     "Hồ sơ đã được phê duyệt",
			Message:   "Hồ sơ HCM-2026-004 của bạn đã được phê duyệt. Vui lòng đến nhận kết quả.",
			Type:      models.NotificationTypeResult,
			IsRead:    true,
			ReadAt:    timePtr(time.Now().AddDate(0, 0, -5)),
			CreatedAt: time.Now().AddDate(0, 0, -5),
		},
		{
			UserID:    citizen2.ID,
			Title:     "Thông báo hệ thống",
			Message:   "Chào mừng bạn đến với Hệ thống Quản lý Dịch vụ Công. Bạn có thể đăng ký các dịch vụ công trực tuyến.",
			Type:      models.NotificationTypeSystem,
			IsRead:    true,
			ReadAt:    timePtr(time.Now().AddDate(0, 0, -10)),
			CreatedAt: time.Now().AddDate(0, 0, -10),
		},
		{
			UserID:    citizen3.ID,
			Title:     "Thông báo hệ thống",
			Message:   "Chào mừng bạn đến với Hệ thống Quản lý Dịch vụ Công.",
			Type:      models.NotificationTypeSystem,
			IsRead:    false,
			CreatedAt: time.Now().AddDate(0, 0, -2),
		},
		{
			UserID:    citizen1.ID,
			Title:     "Nhắc nhở bổ sung hồ sơ",
			Message:   "Hồ sơ HCM-2026-003 của bạn cần bổ sung tài liệu. Vui lòng liên hệ phòng ban phụ trách.",
			Type:      models.NotificationTypeNeedMoreInfo,
			IsRead:    false,
			CreatedAt: time.Now().AddDate(0, 0, -3),
		},
	}

	for i := range notifications {
		if err := db.Where("user_id = ? AND title = ?", notifications[i].UserID, notifications[i].Title).
			FirstOrCreate(&notifications[i]).Error; err != nil {
			return err
		}
	}

	fmt.Println("✓ Notifications seeded")
	return nil
}

func intPtr(i int) *int {
	return &i
}

func timePtr(t time.Time) *time.Time {
	return &t
}
