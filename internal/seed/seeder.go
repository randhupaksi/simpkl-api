package seed

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	archiveentity "simpkl-api/internal/modules/archives/entity"
	auditentity "simpkl-api/internal/modules/auditlogs/entity"
	classentity "simpkl-api/internal/modules/classes/entity"
	companyentity "simpkl-api/internal/modules/companies/entity"
	contactentity "simpkl-api/internal/modules/companycontacts/entity"
	automationentity "simpkl-api/internal/modules/documentautomation/entity"
	documententity "simpkl-api/internal/modules/documents/entity"
	majorentity "simpkl-api/internal/modules/majors/entity"
	periodentity "simpkl-api/internal/modules/periods/entity"
	permissionentity "simpkl-api/internal/modules/permissions/entity"
	placemententity "simpkl-api/internal/modules/placements/entity"
	readinessentity "simpkl-api/internal/modules/readiness/entity"
	roleentity "simpkl-api/internal/modules/roles/entity"
	studententity "simpkl-api/internal/modules/students/entity"
	supervisorentity "simpkl-api/internal/modules/supervisors/entity"
	userentity "simpkl-api/internal/modules/users/entity"
	platformauth "simpkl-api/internal/platform/auth"
)

type Options struct {
	RecordCount   int
	ResetLegacy   bool
	AdminName     string
	AdminEmail    string
	AdminUsername string
	AdminPassword string
}

func Run(ctx context.Context, db *gorm.DB, options Options) error {
	// Jumlah data mengikuti kebutuhan domain PKL, bukan dipaksa sama untuk
	// setiap tabel. SEED_RECORD_COUNT tetap dipakai sebagai fallback jumlah user
	// demo agar env lama tetap kompatibel.
	if options.RecordCount < 1 {
		options.RecordCount = 5
	}
	if strings.TrimSpace(options.AdminPassword) == "" {
		return fmt.Errorf("seed admin password is required")
	}

	passwordHash, err := platformauth.HashPassword(options.AdminPassword)
	if err != nil {
		return fmt.Errorf("hash seed password: %w", err)
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if options.ResetLegacy {
			if err := cleanupLegacyFixtures(tx); err != nil {
				return err
			}
		}
		roles, err := seedRoles(tx)
		if err != nil {
			return err
		}
		majors, err := seedMajors(tx, 6)
		if err != nil {
			return err
		}
		classes, err := seedClasses(tx, majors, 12)
		if err != nil {
			return err
		}
		users, err := seedUsers(tx, options, passwordHash, roles, majors, classes)
		if err != nil {
			return err
		}
		periods, err := seedPeriods(tx, 3)
		if err != nil {
			return err
		}
		students, err := seedStudents(tx, majors, classes, 30)
		if err != nil {
			return err
		}
		companies, err := seedCompanies(tx, majors, 6)
		if err != nil {
			return err
		}
		contacts, err := seedContacts(tx, companies, 6)
		if err != nil {
			return err
		}
		supervisors, err := seedSupervisors(tx, majors, 8)
		if err != nil {
			return err
		}
		placements, err := seedPlacements(tx, periods, students, companies, contacts, supervisors, 30)
		if err != nil {
			return err
		}
		documentTypes, err := seedDocumentTypes(tx)
		if err != nil {
			return err
		}
		if err := seedDocumentAutomation(tx); err != nil {
			return err
		}
		if err := seedDocuments(tx, documentTypes, periods, placements, users, 30); err != nil {
			return err
		}
		if err := seedReadiness(tx, periods, students, placements, 30); err != nil {
			return err
		}
		if err := seedArchives(tx, periods[:1], users[0], 1); err != nil {
			return err
		}
		return seedAuditLogs(tx, users[0], periods, 30)
	})
}

func seedDocumentAutomation(tx *gorm.DB) error {
	profile := automationentity.SchoolProfile{
		InstitutionName: "SMK Citra Negara Depok", InstitutionType: "Sekolah Menengah Kejuruan",
		NPSN: "20232537", Address: "Jl. Raya Tanah Baru No. 99, Kemiri Jaya", Village: "Kemiri Jaya", District: "Beji",
		City: "Depok", Province: "Jawa Barat", PostalCode: "16421", Phone: "(021) 77201052",
		Email: "info@citranegara.sch.id", Website: "https://smk.citranegara.sch.id",
		LetterheadTagline: "Religius, Disiplin, dan Terampil", Timezone: "Asia/Jakarta",
	}
	var existingProfile automationentity.SchoolProfile
	if err := tx.First(&existingProfile).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		if err := tx.Create(&profile).Error; err != nil {
			return fmt.Errorf("seed school profile: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("find school profile: %w", err)
	} else if strings.TrimSpace(existingProfile.InstitutionName) == "" || strings.TrimSpace(existingProfile.Address) == "" || existingProfile.InstitutionName == "Nama Institusi" || existingProfile.InstitutionName == "SMK Nusantara Teknologi" {
		profile.ID = existingProfile.ID
		if err := tx.Model(&existingProfile).Updates(profile).Error; err != nil {
			return fmt.Errorf("update placeholder school profile: %w", err)
		}
	}

	signatory := automationentity.Signatory{Name: "Abdul Kodir Zaelani, S.Pd.I.", Title: "Kepala Sekolah", EmployeeNumber: "", RoleCode: "principal", IsDefault: true, Status: "active"}
	var existingSignatory automationentity.Signatory
	if err := tx.Where("name = ? AND role_code = ?", signatory.Name, signatory.RoleCode).First(&existingSignatory).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		if err := tx.Create(&signatory).Error; err != nil {
			return fmt.Errorf("seed signatory: %w", err)
		}
		existingSignatory = signatory
	} else if err != nil {
		return fmt.Errorf("find signatory: %w", err)
	}
	if err := tx.Model(&automationentity.Signatory{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
		return fmt.Errorf("clear default signatories: %w", err)
	}
	if err := tx.Model(&existingSignatory).Updates(map[string]any{
		"name": signatory.Name, "title": signatory.Title,
		"employee_number": signatory.EmployeeNumber, "role_code": signatory.RoleCode,
		"is_default": true, "status": signatory.Status,
	}).Error; err != nil {
		return fmt.Errorf("activate default signatory: %w", err)
	}

	templates := []automationentity.DocumentTemplate{
		{Code: "introduction_letter", Name: "Surat Pengantar PKL", Category: "letter", SubjectTemplate: "Permohonan Praktik Kerja Lapangan", BodyTemplate: "Dengan hormat,\n\nDalam rangka pelaksanaan program Praktik Kerja Lapangan (PKL) tahun ajaran {{academic_year}}, kami memohon kesediaan {{company_name}} untuk menerima peserta didik berikut:\n\nNama: {{student_name}}\nNIS: {{student_nis}}\nKelas: {{class_name}}\nProgram Keahlian: {{major_name}}\nPeriode: {{placement_start}} sampai {{placement_end}}\n\nKami berharap peserta didik tersebut memperoleh kesempatan belajar dan pengalaman kerja sesuai bidang keahliannya. Atas perhatian dan kerja sama yang baik, kami sampaikan terima kasih.", NumberPattern: "{{sequence}}/PKL/{{month_roman}}/{{year}}", Version: 1, IsActive: true},
		{Code: "placement_letter", Name: "Surat Keterangan Penempatan", Category: "letter", SubjectTemplate: "Keterangan Penempatan Praktik Kerja Lapangan", BodyTemplate: "Yang bertanda tangan di bawah ini menerangkan bahwa:\n\nNama: {{student_name}}\nNIS: {{student_nis}}\nKelas: {{class_name}}\nProgram Keahlian: {{major_name}}\n\ntelah ditempatkan untuk melaksanakan Praktik Kerja Lapangan di {{company_name}}, {{company_address}}, pada bagian {{placement_division}} sebagai {{placement_position}} selama {{placement_start}} sampai {{placement_end}}.\n\nSurat keterangan ini dibuat untuk dipergunakan sebagaimana mestinya.", NumberPattern: "{{sequence}}/KET-PKL/{{month_roman}}/{{year}}", Version: 1, IsActive: true},
		{Code: "supervisor_assignment", Name: "Surat Tugas Guru Pembimbing", Category: "letter", SubjectTemplate: "Penugasan Guru Pembimbing PKL", BodyTemplate: "Dengan ini menugaskan:\n\nNama: {{supervisor_name}}\nNomor Pegawai: {{supervisor_employee_number}}\nJabatan: {{supervisor_position}}\n\nuntuk melaksanakan pembimbingan, pemantauan, dan evaluasi Praktik Kerja Lapangan bagi peserta didik {{student_name}} dari kelas {{class_name}} yang ditempatkan di {{company_name}} selama {{placement_start}} sampai {{placement_end}}.\n\nDemikian surat tugas ini diberikan untuk dilaksanakan dengan penuh tanggung jawab.", NumberPattern: "{{sequence}}/ST-PKL/{{month_roman}}/{{year}}", Version: 1, IsActive: true},
		{Code: "parent_consent", Name: "Surat Persetujuan Orang Tua", Category: "letter", SubjectTemplate: "Persetujuan Mengikuti Praktik Kerja Lapangan", BodyTemplate: "Saya yang bertanda tangan di bawah ini:\n\nNama Orang Tua/Wali: {{parent_name}}\nOrang tua/wali dari: {{student_name}}\nKelas: {{class_name}}\nProgram Keahlian: {{major_name}}\n\ndengan ini menyetujui peserta didik tersebut mengikuti Praktik Kerja Lapangan di {{company_name}} pada {{placement_start}} sampai {{placement_end}}. Kami memahami dan bersedia mendukung ketentuan pelaksanaan PKL yang berlaku.\n\nDemikian surat persetujuan ini dibuat dengan sebenarnya.", NumberPattern: "{{sequence}}/IZIN-PKL/{{month_roman}}/{{year}}", Version: 1, IsActive: true},
		{Code: "placement_recap", Name: "Rekap Penempatan PKL", Category: "spreadsheet", SubjectTemplate: "Rekap Penempatan Praktik Kerja Lapangan", BodyTemplate: "Rekap penempatan peserta didik berdasarkan periode dan filter yang dipilih.", NumberPattern: "REKAP-PKL-{{year}}", Version: 1, IsActive: true},
	}
	for i := range templates {
		if err := tx.Where("code = ? AND version = ?", templates[i].Code, templates[i].Version).FirstOrCreate(&templates[i]).Error; err != nil {
			return fmt.Errorf("seed document automation template %s: %w", templates[i].Code, err)
		}
	}
	return nil
}

// cleanupLegacyFixtures hanya menghapus marker yang dibuat oleh seeder awal.
// Data sekolah/non-seeder tidak cocok dengan marker ini dan tetap dipertahankan.
func cleanupLegacyFixtures(tx *gorm.DB) error {
	var oldStudents []studententity.Student
	var oldCompanies []companyentity.Company
	var oldPeriods []periodentity.Period
	var oldUsers []userentity.User
	if err := tx.Unscoped().Where("nis LIKE ?", "SEED%").Find(&oldStudents).Error; err != nil {
		return fmt.Errorf("find legacy students: %w", err)
	}
	if err := tx.Unscoped().Where("name LIKE ? OR name IN ?", "Mitra Industri Seed%", []string{
		"CV Beji Solusi Digital", "PT Margonda Teknologi Kreatif", "CV Depok Network Center",
		"PT Citra Aplikasi Nusantara", "Studio Visual Kukusan", "CV Arsip Bisnis Depok",
		"PT Layanan Data Margonda", "Toko Komputer Tanah Baru", "CV Promosi Kreatif Beji",
		"PT Ritel Sejahtera Depok", "CV Kantor Digital Kemiri Jaya", "Hotel Mitra Depok",
	}).Find(&oldCompanies).Error; err != nil {
		return fmt.Errorf("find legacy companies: %w", err)
	}
	if err := tx.Unscoped().Where("name LIKE ?", "PKL Seed%").Find(&oldPeriods).Error; err != nil {
		return fmt.Errorf("find legacy periods: %w", err)
	}
	if err := tx.Unscoped().Where("email LIKE ?", "demo.seed.%").Find(&oldUsers).Error; err != nil {
		return fmt.Errorf("find legacy users: %w", err)
	}

	studentIDs := idsOfStudents(oldStudents)
	companyIDs := idsOfCompanies(oldCompanies)
	periodIDs := idsOfPeriods(oldPeriods)
	userIDs := idsOfUsers(oldUsers)
	if len(studentIDs) > 0 {
		if err := tx.Unscoped().Where("student_id IN ?", studentIDs).Delete(&placemententity.Placement{}).Error; err != nil {
			return fmt.Errorf("delete legacy placements by student: %w", err)
		}
		if err := tx.Unscoped().Where("student_id IN ?", studentIDs).Delete(&readinessentity.Readiness{}).Error; err != nil {
			return fmt.Errorf("delete legacy readiness: %w", err)
		}
	}
	if len(periodIDs) > 0 {
		if err := tx.Unscoped().Where("period_id IN ?", periodIDs).Delete(&placemententity.Placement{}).Error; err != nil {
			return fmt.Errorf("delete legacy placements by period: %w", err)
		}
		if err := tx.Unscoped().Where("period_id IN ?", periodIDs).Delete(&archiveentity.Archive{}).Error; err != nil {
			return fmt.Errorf("delete legacy archives: %w", err)
		}
	}
	if len(companyIDs) > 0 {
		if err := tx.Unscoped().Where("company_id IN ?", companyIDs).Delete(&placemententity.Placement{}).Error; err != nil {
			return fmt.Errorf("delete legacy placements by company: %w", err)
		}
		if err := tx.Unscoped().Where("company_id IN ?", companyIDs).Delete(&contactentity.CompanyContact{}).Error; err != nil {
			return fmt.Errorf("delete legacy company contacts: %w", err)
		}
		if err := tx.Unscoped().Where("company_id IN ?", companyIDs).Delete(&companyentity.MajorCapacity{}).Error; err != nil {
			return fmt.Errorf("delete legacy company capacities: %w", err)
		}
	}
	if err := tx.Unscoped().Where("number LIKE ?", "DOC-SEED-%").Delete(&documententity.Document{}).Error; err != nil {
		return fmt.Errorf("delete legacy documents: %w", err)
	}
	if err := tx.Unscoped().Where("request_id LIKE ?", "seed-request-%").Delete(&auditentity.AuditLog{}).Error; err != nil {
		return fmt.Errorf("delete legacy audit logs: %w", err)
	}
	if err := tx.Unscoped().Where("email LIKE ?", "pic.seed.%").Delete(&contactentity.CompanyContact{}).Error; err != nil {
		return fmt.Errorf("delete legacy contacts: %w", err)
	}
	if err := tx.Unscoped().Where("employee_number LIKE ?", "GURU-SEED-%").Delete(&supervisorentity.Supervisor{}).Error; err != nil {
		return fmt.Errorf("delete legacy supervisors: %w", err)
	}
	if len(userIDs) > 0 {
		if err := tx.Unscoped().Where("user_id IN ?", userIDs).Delete(&roleentity.UserRole{}).Error; err != nil {
			return fmt.Errorf("delete legacy user roles: %w", err)
		}
	}
	if err := tx.Unscoped().Where("nis LIKE ?", "SEED%").Delete(&studententity.Student{}).Error; err != nil {
		return fmt.Errorf("delete legacy students: %w", err)
	}
	legacyClassIDs := tx.Model(&classentity.Class{}).Select("id").Where("major_id IN (?)", tx.Model(&majorentity.Major{}).Select("id").Where("code LIKE ?", "SEED-%"))
	legacyClassStudentIDs := tx.Model(&studententity.Student{}).Select("id").Where("class_id IN (?)", legacyClassIDs)
	if err := tx.Unscoped().Where("student_id IN (?)", legacyClassStudentIDs).Delete(&placemententity.Placement{}).Error; err != nil {
		return fmt.Errorf("delete placements from legacy classes: %w", err)
	}
	if err := tx.Unscoped().Where("student_id IN (?)", legacyClassStudentIDs).Delete(&readinessentity.Readiness{}).Error; err != nil {
		return fmt.Errorf("delete readiness from legacy classes: %w", err)
	}
	if err := tx.Unscoped().Where("owner_id IN (?) OR number LIKE ? OR number LIKE ?", legacyClassStudentIDs, "DOC-SEED-%", "DOC-CN-2627-%").Delete(&documententity.Document{}).Error; err != nil {
		return fmt.Errorf("delete documents from legacy classes: %w", err)
	}
	if err := tx.Unscoped().Where("class_id IN (?)", legacyClassIDs).Delete(&studententity.Student{}).Error; err != nil {
		return fmt.Errorf("delete students from legacy classes: %w", err)
	}
	if err := tx.Unscoped().Where("major_id IN (?)", tx.Model(&majorentity.Major{}).Select("id").Where("code LIKE ?", "SEED-%")).Delete(&classentity.Class{}).Error; err != nil {
		return fmt.Errorf("delete legacy classes by major: %w", err)
	}
	if err := tx.Unscoped().Where("name LIKE ?", "XII-SEED-%").Delete(&classentity.Class{}).Error; err != nil {
		return fmt.Errorf("delete legacy classes: %w", err)
	}
	if err := tx.Unscoped().Where("code LIKE ?", "SEED-%").Delete(&majorentity.Major{}).Error; err != nil {
		return fmt.Errorf("delete legacy majors: %w", err)
	}
	if err := tx.Unscoped().Where("name LIKE ?", "Mitra Industri Seed%").Delete(&companyentity.Company{}).Error; err != nil {
		return fmt.Errorf("delete legacy companies: %w", err)
	}
	if err := tx.Unscoped().Where("name LIKE ?", "PKL Seed%").Delete(&periodentity.Period{}).Error; err != nil {
		return fmt.Errorf("delete legacy periods: %w", err)
	}
	if err := tx.Unscoped().Where("email LIKE ?", "demo.seed.%").Delete(&userentity.User{}).Error; err != nil {
		return fmt.Errorf("delete legacy users: %w", err)
	}
	return nil
}

func idsOfStudents(items []studententity.Student) []string {
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	return ids
}

func idsOfCompanies(items []companyentity.Company) []string {
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	return ids
}

func idsOfPeriods(items []periodentity.Period) []string {
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	return ids
}

func idsOfUsers(items []userentity.User) []string {
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	return ids
}

func seedRoles(tx *gorm.DB) (map[string]roleentity.Role, error) {
	result := make(map[string]roleentity.Role)
	for _, code := range []string{"super_admin", "admin_pkl", "coordinator_pkl", "program_head", "homeroom_teacher", "supervisor_teacher", "leadership"} {
		var role roleentity.Role
		if err := tx.Where("code = ?", code).First(&role).Error; err != nil {
			return nil, fmt.Errorf("find role %s: %w; run migrations first", code, err)
		}
		result[code] = role
	}
	var wildcard permissionentity.Permission
	if err := tx.Where("code = ?", "*").First(&wildcard).Error; err != nil {
		return nil, fmt.Errorf("find wildcard permission: %w; run RBAC migrations first", err)
	}
	superAdmin := result["super_admin"]
	if err := tx.Model(&superAdmin).Update("status", "active").Error; err != nil {
		return nil, fmt.Errorf("activate super admin role: %w", err)
	}
	if err := tx.Where("role_id = ? AND permission_id = ?", superAdmin.ID, wildcard.ID).FirstOrCreate(&permissionentity.RolePermission{RoleID: superAdmin.ID, PermissionID: wildcard.ID}).Error; err != nil {
		return nil, fmt.Errorf("repair super admin permissions: %w", err)
	}
	return result, nil
}

func seedMajors(tx *gorm.DB, count int) ([]majorentity.Major, error) {
	majors := make([]majorentity.Major, count)
	programs := []struct {
		code, name, abbreviation, description string
	}{
		{"CN-TJKT", "Teknik Jaringan Komputer dan Telekomunikasi", "TJKT", "Instalasi jaringan, fiber optik, perangkat komputer, keamanan jaringan, dan dukungan teknis."},
		{"CN-PPLG", "Pengembangan Perangkat Lunak dan Gim", "PPLG", "Analisis kebutuhan, pengembangan aplikasi, website, basis data, gim, dan pengujian perangkat lunak."},
		{"CN-DKV", "Desain Komunikasi Visual", "DKV", "Desain grafis, fotografi, produksi konten, branding, dan komunikasi visual digital."},
		{"CN-PH", "Perhotelan", "PH", "Layanan akomodasi, tata graha, pelayanan tamu, front office, dan operasional hospitality."},
		{"CN-MPLB", "Manajemen Perkantoran dan Layanan Bisnis", "MPLB", "Administrasi perkantoran, pengarsipan, input data, layanan publik, dan pengelolaan dokumen."},
		{"CN-PM", "Pemasaran", "PM", "Penjualan, layanan pelanggan, promosi, pemasaran digital, dan pengelolaan kanal bisnis."},
	}
	for i := range majors {
		program := programs[i%len(programs)]
		item := majorentity.Major{Code: program.code, Name: program.name, Abbreviation: program.abbreviation, HeadName: "", Status: "active", Description: program.description}
		if err := tx.Where("code = ?", item.Code).Assign(item).FirstOrCreate(&item).Error; err != nil {
			return nil, fmt.Errorf("seed major %d: %w", i+1, err)
		}
		majors[i] = item
	}
	return majors, nil
}

func seedClasses(tx *gorm.DB, majors []majorentity.Major, count int) ([]classentity.Class, error) {
	classes := make([]classentity.Class, count)
	for i := range classes {
		major := majors[i%len(majors)]
		item := classentity.Class{Name: fmt.Sprintf("XII %s %d", major.Abbreviation, i/len(majors)+1), Level: 12, MajorID: major.ID, HomeroomTeacher: []string{"Rina Kurniawati, S.Pd.", "Dedi Hermawan, S.Pd.", "Siti Maemunah, S.Pd.", "Agus Setiawan, S.Pd."}[i%4], AcademicYear: "2026/2027", Status: "active"}
		if err := tx.Where("name = ? AND academic_year = ?", item.Name, item.AcademicYear).Assign(item).FirstOrCreate(&item).Error; err != nil {
			return nil, fmt.Errorf("seed class %d: %w", i+1, err)
		}
		classes[i] = item
	}
	return classes, nil
}

func seedUsers(tx *gorm.DB, options Options, passwordHash string, roles map[string]roleentity.Role, majors []majorentity.Major, classes []classentity.Class) ([]userentity.User, error) {
	adminEmail := strings.ToLower(strings.TrimSpace(options.AdminEmail))
	adminUsername := strings.ToLower(strings.TrimSpace(options.AdminUsername))
	admin := userentity.User{Name: options.AdminName, Email: adminEmail, Username: adminUsername, PasswordHash: passwordHash, Status: "active"}
	if err := tx.Where("email = ?", adminEmail).FirstOrCreate(&admin).Error; err != nil {
		return nil, fmt.Errorf("seed admin: %w", err)
	}
	if err := tx.Where("user_id = ? AND role_id = ?", admin.ID, roles["super_admin"].ID).FirstOrCreate(&roleentity.UserRole{UserID: admin.ID, RoleID: roles["super_admin"].ID}).Error; err != nil {
		return nil, fmt.Errorf("assign admin role: %w", err)
	}

	users := make([]userentity.User, options.RecordCount)
	roleCodes := []string{"admin_pkl", "coordinator_pkl", "program_head", "homeroom_teacher", "supervisor_teacher"}
	for i := range users {
		item := userentity.User{Name: fmt.Sprintf("Pengguna Demo %d", i+1), Email: fmt.Sprintf("demo.seed.%d@example.sch.id", i+1), Username: fmt.Sprintf("demo_seed_%d", i+1), PasswordHash: passwordHash, MajorID: majors[i%len(majors)].ID, ClassID: classes[i%len(classes)].ID, Status: "active"}
		if err := tx.Where("email = ?", item.Email).FirstOrCreate(&item).Error; err != nil {
			return nil, fmt.Errorf("seed demo user %d: %w", i+1, err)
		}
		role := roles[roleCodes[i%len(roleCodes)]]
		if err := tx.Where("user_id = ? AND role_id = ?", item.ID, role.ID).FirstOrCreate(&roleentity.UserRole{UserID: item.ID, RoleID: role.ID}).Error; err != nil {
			return nil, fmt.Errorf("assign demo user role %d: %w", i+1, err)
		}
		users[i] = item
	}
	return append([]userentity.User{admin}, users...), nil
}

func seedPeriods(tx *gorm.DB, count int) ([]periodentity.Period, error) {
	periods := make([]periodentity.Period, count)
	periodDefinitions := []struct {
		name, academicYear, semester string
		start, end                   time.Time
		cohort                       int
		status                       string
	}{
		{"PKL Semester Genap 2025/2026", "2025/2026", "even", time.Date(2026, 1, 5, 0, 0, 0, 0, time.Local), time.Date(2026, 6, 30, 0, 0, 0, 0, time.Local), 2025, "completed"},
		{"PKL Semester Ganjil 2026/2027", "2026/2027", "odd", time.Date(2026, 7, 13, 0, 0, 0, 0, time.Local), time.Date(2026, 12, 19, 0, 0, 0, 0, time.Local), 2026, "active"},
		{"PKL Semester Ganjil 2027/2028", "2027/2028", "odd", time.Date(2027, 7, 12, 0, 0, 0, 0, time.Local), time.Date(2027, 12, 18, 0, 0, 0, 0, time.Local), 2027, "active"},
	}
	for i := range periods {
		definition := periodDefinitions[i%len(periodDefinitions)]
		item := periodentity.Period{Name: definition.name, AcademicYear: definition.academicYear, Semester: definition.semester, StartDate: definition.start, EndDate: definition.end, Cohort: definition.cohort, Status: definition.status, Notes: "Fixture lokal SIMPKL untuk simulasi pengelolaan PKL SMK Citra Negara Depok."}
		if err := tx.Where("name = ?", item.Name).Assign(item).FirstOrCreate(&item).Error; err != nil {
			return nil, fmt.Errorf("seed period %d: %w", i+1, err)
		}
		periods[i] = item
	}
	return periods, nil
}

func seedStudents(tx *gorm.DB, majors []majorentity.Major, classes []classentity.Class, count int) ([]studententity.Student, error) {
	students := make([]studententity.Student, count)
	names := []string{"Aditya Pratama", "Aisyah Nuraini", "Bagas Ramadhan", "Bella Maharani", "Cahyo Nugroho", "Citra Lestari", "Daffa Alfarizi", "Dinda Permata", "Fajar Maulana", "Farah Nabila", "Galang Saputra", "Hana Fitriani", "Iqbal Maulana", "Intan Safitri", "Jovan Kurniawan", "Keisya Amalia", "M. Rizky Firmansyah", "Nadia Putri", "Oka Wijaya", "Putri Ayuningtyas", "Rafi Hidayat", "Salma Azzahra", "Tegar Firmansyah", "Ulfa Rahmawati", "Vino Adiputra", "Wahyu Setiawan", "Yasmin Khairunnisa", "Zaki Pranata", "Andika Saputra", "Novi Lestari"}
	for i := range students {
		item := studententity.Student{NIS: fmt.Sprintf("2627%04d", i+1), NISN: fmt.Sprintf("0068%08d", i+1), Name: names[i%len(names)], Nickname: strings.Split(names[i%len(names)], " ")[0], Gender: []string{"male", "female"}[i%2], ClassID: classes[i%len(classes)].ID, MajorID: majors[i%len(majors)].ID, Phone: fmt.Sprintf("081200000%03d", i+1), Email: fmt.Sprintf("siswa.%d@example.sch.id", i+1), Address: fmt.Sprintf("Jl. Raya Tanah Baru No. %d, Beji, Depok", 12+i), ParentName: fmt.Sprintf("Orang Tua %s", names[i%len(names)]), ParentPhone: fmt.Sprintf("081300000%03d", i+1), Status: "active", PKLStatus: []string{"active", "active", "ready", "awaiting_documents"}[i%4], Notes: "Data siswa sintetis untuk simulasi administrasi PKL."}
		if err := tx.Where("nis = ?", item.NIS).Assign(item).FirstOrCreate(&item).Error; err != nil {
			return nil, fmt.Errorf("seed student %d: %w", i+1, err)
		}
		students[i] = item
	}
	return students, nil
}

func seedCompanies(tx *gorm.DB, majors []majorentity.Major, count int) ([]companyentity.Company, error) {
	companies := make([]companyentity.Company, count)
	definitions := []struct {
		name, businessType, industry, description, address, district, city, postalCode, website, notes string
		status                                                                                         string
		capacity                                                                                       int
		majorIndexes                                                                                   []int
	}{
		{"Fakultas Ilmu Komputer Universitas Indonesia", "Perguruan tinggi", "Pendidikan tinggi dan teknologi informasi", "Lokasi PKL yang terdokumentasi untuk siswa SMK Citra Negara pada bidang administrasi perkantoran di lingkungan Fasilkom UI.", "Kampus UI, Depok", "Beji", "Depok", "16424", "https://cs.ui.ac.id", "Tercatat sebagai pengalaman PKL siswa SMK Citra Negara dalam publikasi Fasilkom UI; bukan klaim bahwa seluruh program studi menerima PKL setiap periode.", "active", 12, []int{0, 1, 4}},
		{"Sakubi Teknologi Indonesia", "PT", "Teknologi pendidikan dan layanan bisnis", "Perusahaan teknologi yang menjadi lokasi PKL siswa MPLB SMK Citra Negara pada laporan kegiatan tahun 2025.", "Kampus UI, Depok", "Beji", "Depok", "16424", "", "Tercatat sebagai instansi pada laporan PKL siswa SMK Citra Negara tahun 2025; detail PIC tidak dipublikasikan.", "active", 10, []int{1, 4}},
		{"PT Propertree Investa Cendekia", "PT", "Teknologi finansial dan properti", "Perusahaan teknologi finansial yang menjadi salah satu lokasi PKL siswa DKV SMK Citra Negara pada laporan tahun 2025.", "Jl. K.H.M. Yusuf Raya, Mekar Jaya", "Sukmajaya", "Depok", "16418", "", "Tercatat sebagai lokasi PKL siswa DKV SMK Citra Negara pada laporan kegiatan publik; detail PIC tidak dipublikasikan.", "active", 8, []int{1, 2, 4}},
		{"Raden Village", "Unit usaha", "Properti dan layanan hunian", "Lingkungan usaha properti di Pancoran Mas yang menjadi lokasi PKL siswa DKV SMK Citra Negara pada laporan tahun 2025.", "Jl. Al-Muttaqin, RT 003/RW 008", "Pancoran Mas", "Depok", "16434", "", "Tercatat sebagai lokasi PKL siswa DKV SMK Citra Negara pada laporan kegiatan publik; detail PIC tidak dipublikasikan.", "active", 8, []int{2, 3, 4, 5}},
		{"PT CQU Executive Business Training Centre", "PT", "Pelatihan bisnis dan pengembangan profesional", "Pusat pelatihan bisnis yang menjadi salah satu lokasi PKL siswa DKV SMK Citra Negara pada laporan tahun 2025.", "Mall Kuningan City LG Unit 24-25, Jl. Prof. Dr. Satrio Kav. 18", "Setiabudi", "Jakarta Selatan", "12940", "", "Tercatat sebagai lokasi PKL siswa DKV SMK Citra Negara pada laporan kegiatan publik; detail PIC tidak dipublikasikan.", "active", 8, []int{2, 4, 5}},
		{"PT Lencana Gaya Indonesia", "PT", "Manufaktur garmen dan industri kreatif", "Rumah produksi kreatif bidang pakaian dan mode yang menjadi lokasi PKL siswa DKV SMK Citra Negara pada 2024.", "Jl. Siliwangi No. 1", "Pancoran Mas", "Depok", "16431", "", "Tercatat sebagai lokasi PKL siswa DKV SMK Citra Negara pada laporan kegiatan 2024; detail PIC tidak dipublikasikan.", "active", 8, []int{2, 4, 5}},
	}
	for i := range companies {
		definition := definitions[i%len(definitions)]
		item := companyentity.Company{Name: definition.name, BusinessType: definition.businessType, Industry: definition.industry, Description: definition.description, Address: definition.address, District: definition.district, City: definition.city, Province: "Jawa Barat", PostalCode: definition.postalCode, Website: definition.website, Status: definition.status, Capacity: definition.capacity, Notes: definition.notes}
		if err := tx.Where("name = ?", item.Name).FirstOrCreate(&item).Error; err != nil {
			return nil, fmt.Errorf("seed company %d: %w", i+1, err)
		}
		for _, majorIndex := range definition.majorIndexes {
			if err := tx.Where("company_id = ? AND major_id = ?", item.ID, majors[majorIndex].ID).FirstOrCreate(&companyentity.MajorCapacity{CompanyID: item.ID, MajorID: majors[majorIndex].ID, Capacity: 4}).Error; err != nil {
				return nil, fmt.Errorf("seed company capacity %d/%d: %w", i+1, majorIndex, err)
			}
		}
		companies[i] = item
	}
	return companies, nil
}

func seedContacts(tx *gorm.DB, companies []companyentity.Company, count int) ([]contactentity.CompanyContact, error) {
	contacts := make([]contactentity.CompanyContact, count)
	roles := []string{"Sekretariat PKL Fasilkom UI", "Koordinator PKL Sakubi", "Koordinator PKL Propertree", "Pengelola Raden Village", "Student Engagement CQU", "Pembimbing Instansi Lencana"}
	for i := range contacts {
		item := contactentity.CompanyContact{CompanyID: companies[i%len(companies)].ID, Name: roles[i%len(roles)], Position: "Koordinator penerimaan PKL", Division: "Hubungan industri", Email: fmt.Sprintf("pic.seed.%d@example.sch.id", i+1), IsPrimary: true, Notes: "Placeholder operasional untuk demo lokal; nama dan kontak PIC resmi tidak dipublikasikan pada sumber yang digunakan."}
		if err := tx.Where("email = ?", item.Email).Assign(item).FirstOrCreate(&item).Error; err != nil {
			return nil, fmt.Errorf("seed company contact %d: %w", i+1, err)
		}
		contacts[i] = item
	}
	return contacts, nil
}

func seedSupervisors(tx *gorm.DB, majors []majorentity.Major, count int) ([]supervisorentity.Supervisor, error) {
	supervisors := make([]supervisorentity.Supervisor, count)
	names := []string{"Muhammad Fajar Sidiq, S.E.", "Syta Choirunnisa, S.Kom.", "Dewi Halimah Shoolihah, S.Ds.", "Neneng Supartiwi, S.Pd., M.M.", "Nurhakim Wirasena, S.I.Kom.", "Riki Yakub, S.Kom., Gr.", "Deni Romandono, S.T.", "Decky Riyansyah, M.Kom."}
	for i := range supervisors {
		item := supervisorentity.Supervisor{EmployeeNumber: fmt.Sprintf("CN-GPKL-%03d", i+1), Name: names[i%len(names)], Phone: fmt.Sprintf("081500000%03d", i+1), Email: fmt.Sprintf("pembimbing.%d@example.sch.id", i+1), MajorID: majors[i%len(majors)].ID, Position: "Guru Pembimbing PKL", Status: "active", MaxStudents: 20}
		if err := tx.Where("employee_number = ?", item.EmployeeNumber).Assign(item).FirstOrCreate(&item).Error; err != nil {
			return nil, fmt.Errorf("seed supervisor %d: %w", i+1, err)
		}
		supervisors[i] = item
	}
	return supervisors, nil
}

func seedPlacements(tx *gorm.DB, periods []periodentity.Period, students []studententity.Student, companies []companyentity.Company, contacts []contactentity.CompanyContact, supervisors []supervisorentity.Supervisor, count int) ([]placemententity.Placement, error) {
	placements := make([]placemententity.Placement, 0, count)
	placementPeriods := periods
	currentPeriodID := periods[len(periods)-1].ID
	// Setiap pola siswa memakai satu perusahaan berbeda. Urutan ini menjaga
	// kombinasi jurusan dan perusahaan tetap sesuai kapasitas yang diseed,
	// sekaligus memastikan setiap perusahaan bisa diuji dari document automation.
	companyByMajor := []int{0, 1, 4, 3, 2, 5}
	divisions := []string{"Administrasi dan layanan akademik", "Administrasi dan teknologi bisnis", "Desain dan konten visual", "Layanan operasional properti", "Administrasi perkantoran", "Pemasaran dan layanan pelanggan"}
	positions := []string{"Staf administrasi PKL", "Asisten administrasi digital", "Desainer konten junior", "Asisten layanan operasional", "Staf pengarsipan dan input data", "Asisten pemasaran"}
	for periodIndex, period := range placementPeriods {
		for i := 0; i < count; i++ {
			majorIndex := i % 6
			companyIndex := companyByMajor[majorIndex]
			status := []string{"active", "active", "ready", "pending_verification", "approved", "ready"}[i%6]
			switch period.Status {
			case "completed", "archived":
				status = "completed"
			case "draft":
				status = "draft"
			case "preparation":
				status = "pending_verification"
			}
			item := placemententity.Placement{PeriodID: period.ID, StudentID: students[i%len(students)].ID, CompanyID: companies[companyIndex].ID, CompanyContactID: contacts[companyIndex].ID, SupervisorID: supervisors[majorIndex%len(supervisors)].ID, Division: divisions[majorIndex], Position: positions[majorIndex], WorkSystem: []string{"wfo", "hybrid", "wfo", "wfo", "hybrid", "wfo"}[majorIndex], Address: companies[companyIndex].Address, StartDate: period.StartDate, EndDate: period.EndDate, Status: status, Source: []string{"school", "teacher_recommendation", "self_submission"}[i%3], Notes: "Fixture lokal berbasis konteks PKL SMK Citra Negara Depok; verifikasi lapangan diperlukan untuk penggunaan nyata."}
			if err := tx.Omit("PreviousPlacementID").Where("student_id = ? AND period_id = ?", item.StudentID, item.PeriodID).Assign(item).FirstOrCreate(&item).Error; err != nil {
				return nil, fmt.Errorf("seed placement %d/%d: %w", periodIndex+1, i+1, err)
			}
			if period.ID == currentPeriodID {
				placements = append(placements, item)
			}
		}
	}
	return placements, nil
}

func seedDocumentTypes(tx *gorm.DB) ([]documententity.DocumentType, error) {
	definitions := []documententity.DocumentType{
		{Code: "acceptance_letter", Name: "Surat Penerimaan Perusahaan", Category: "company", Required: true, AllowedMIME: "application/pdf,image/jpeg,image/png"},
		{Code: "parent_permission", Name: "Surat Izin Orang Tua", Category: "student", Required: true, AllowedMIME: "application/pdf,image/jpeg,image/png"},
		{Code: "introduction_letter", Name: "Surat Pengantar Sekolah", Category: "school", Required: true, AllowedMIME: "application/pdf"},
		{Code: "mou", Name: "MoU/Perjanjian Kerja Sama", Category: "company", HasExpiry: true, AllowedMIME: "application/pdf"},
		{Code: "supervisor_assignment", Name: "Surat Tugas Guru Pembimbing", Category: "school", AllowedMIME: "application/pdf"},
	}
	result := make([]documententity.DocumentType, len(definitions))
	for i, item := range definitions {
		if err := tx.Where("code = ?", item.Code).FirstOrCreate(&item).Error; err != nil {
			return nil, fmt.Errorf("seed document type %s: %w", item.Code, err)
		}
		result[i] = item
	}
	return result, nil
}

func seedDocuments(tx *gorm.DB, types []documententity.DocumentType, periods []periodentity.Period, placements []placemententity.Placement, users []userentity.User, count int) error {
	for i := 0; i < count; i++ {
		item := documententity.Document{DocumentTypeID: types[i%len(types)].ID, OwnerType: "student", OwnerID: placements[i%len(placements)].StudentID, PeriodID: periods[len(periods)-1].ID, PlacementID: placements[i%len(placements)].ID, Number: fmt.Sprintf("DOC-CN-2627-%03d", i+1), OriginalName: fmt.Sprintf("berkas-pkl-%03d.pdf", i+1), StoredName: fmt.Sprintf("seed-document-%d.pdf", i+1), Path: fmt.Sprintf("seed/documents/%d.pdf", i+1), MimeType: "application/pdf", Size: 2048, Status: []string{"valid", "valid", "pending", "revision_required"}[i%4], Version: 1, VerifiedBy: users[0].ID, Notes: "Berkas dummy untuk simulasi administrasi PKL."}
		if err := tx.Where("number = ?", item.Number).FirstOrCreate(&item).Error; err != nil {
			return fmt.Errorf("seed document %d: %w", i+1, err)
		}
	}
	return nil
}

func seedReadiness(tx *gorm.DB, periods []periodentity.Period, students []studententity.Student, placements []placemententity.Placement, count int) error {
	for i := 0; i < count; i++ {
		item := readinessentity.Readiness{StudentID: students[i%len(students)].ID, PeriodID: periods[len(periods)-1].ID, PlacementID: placements[i%len(placements)].ID, DataComplete: true, CompanyAssigned: true, ContactAvailable: true, SupervisorAssigned: true, DatesSet: true, AcceptanceLetterValid: i%4 != 3, ParentPermissionValid: true, IntroductionLetterValid: true, RequiredCount: 8, CompletedCount: 7 + boolToInt(i%4 != 3), Percentage: 87.5 + float64(boolToInt(i%4 != 3))*12.5, Status: []string{"attention", "ready", "ready", "attention"}[i%4]}
		if err := tx.Where("student_id = ? AND period_id = ?", item.StudentID, item.PeriodID).FirstOrCreate(&item).Error; err != nil {
			return fmt.Errorf("seed readiness %d: %w", i+1, err)
		}
	}
	return nil
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func seedArchives(tx *gorm.DB, periods []periodentity.Period, admin userentity.User, count int) error {
	for i := 0; i < count; i++ {
		snapshot, _ := json.Marshal(map[string]any{"source": "seed", "period": periods[i%len(periods)].Name, "record_count": 30})
		item := archiveentity.Archive{PeriodID: periods[i%len(periods)].ID, ArchivedBy: admin.ID, ArchivedAt: time.Now(), Reason: "Arsip fixture untuk pengembangan lokal.", Snapshot: string(snapshot)}
		if err := tx.Where("period_id = ?", item.PeriodID).FirstOrCreate(&item).Error; err != nil {
			return fmt.Errorf("seed archive %d: %w", i+1, err)
		}
	}
	return nil
}

func seedAuditLogs(tx *gorm.DB, admin userentity.User, periods []periodentity.Period, count int) error {
	for i := 0; i < count; i++ {
		item := auditentity.AuditLog{ActorID: admin.ID, Action: "seed", Resource: "period", ResourceID: periods[i%len(periods)].ID, RequestID: fmt.Sprintf("seed-request-%d", i+1), AfterJSON: `{"source":"seed"}`, Reason: "Data fixture untuk pengembangan lokal.", IPAddress: "127.0.0.1", UserAgent: "SIMPKL Seeder"}
		if err := tx.Where("request_id = ?", item.RequestID).FirstOrCreate(&item).Error; err != nil {
			return fmt.Errorf("seed audit log %d: %w", i+1, err)
		}
	}
	return nil
}
