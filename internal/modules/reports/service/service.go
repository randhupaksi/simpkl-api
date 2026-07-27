package service

import (
	"context"
	"time"

	"gorm.io/gorm"

	"simpkl-api/internal/modules/reports/entity"
	excelreport "simpkl-api/internal/platform/report/excel"
	pdfreport "simpkl-api/internal/platform/report/pdf"
)

type Service struct {
	db    *gorm.DB
	excel *excelreport.Generator
	pdf   *pdfreport.Generator
}

func New(db *gorm.DB) *Service {
	return &Service{db: db, excel: excelreport.New(), pdf: pdfreport.New()}
}

func (s *Service) Dashboard(ctx context.Context, periodID string) (*entity.Dashboard, error) {
	result := &entity.Dashboard{}
	studentQuery := s.db.WithContext(ctx).Table("students").Where("deleted_at IS NULL")
	if err := studentQuery.Count(&result.TotalStudents).Error; err != nil {
		return nil, err
	}
	result.UnplacedStudents = count(s.db.WithContext(ctx).Table("students").Where("deleted_at IS NULL AND pkl_status = ?", "unplaced"))
	placementQuery := s.db.WithContext(ctx).Table("placements").Where("deleted_at IS NULL")
	if periodID != "" {
		placementQuery = placementQuery.Where("period_id = ?", periodID)
	}
	result.PlacedStudents = count(placementQuery.Session(&gorm.Session{}).Where("status NOT IN ?", []string{"cancelled"}).Distinct("student_id"))
	result.ActivePlacements = count(placementQuery.Session(&gorm.Session{}).Where("status = ?", "active"))
	result.ActiveCompanies = count(s.db.WithContext(ctx).Table("companies").Where("deleted_at IS NULL AND status = ?", "active"))
	result.IncompleteDocuments = count(s.db.WithContext(ctx).Table("administrative_readiness").Where("deleted_at IS NULL AND status IN ?", []string{"incomplete", "attention"}))
	today := time.Now()
	result.StartingSoon = count(placementQuery.Session(&gorm.Session{}).Where("start_date BETWEEN ? AND ?", today, today.AddDate(0, 0, 7)))
	result.EndingSoon = count(placementQuery.Session(&gorm.Session{}).Where("end_date BETWEEN ? AND ?", today, today.AddDate(0, 0, 14)))
	return result, nil
}

func (s *Service) Placements(ctx context.Context, filters map[string]string) ([]entity.PlacementRow, error) {
	query := s.db.WithContext(ctx).Table("placements p").
		Select(`students.nis, students.name AS student_name, classes.name AS class_name,
			majors.name AS major_name, companies.name AS company_name,
			supervisors.name AS supervisor_name,
			DATE_FORMAT(p.start_date, '%Y-%m-%d') AS start_date,
			DATE_FORMAT(p.end_date, '%Y-%m-%d') AS end_date, p.status`).
		Joins("JOIN students ON students.id = p.student_id").
		Joins("LEFT JOIN classes ON classes.id = students.class_id").
		Joins("LEFT JOIN majors ON majors.id = students.major_id").
		Joins("JOIN companies ON companies.id = p.company_id").
		Joins("LEFT JOIN supervisors ON supervisors.id = p.supervisor_id").
		Where("p.deleted_at IS NULL")
	columns := map[string]string{
		"period_id": "p.period_id", "major_id": "students.major_id", "class_id": "students.class_id",
		"company_id": "p.company_id", "supervisor_id": "p.supervisor_id", "status": "p.status",
	}
	for key, value := range filters {
		if column, ok := columns[key]; ok && value != "" {
			query = query.Where(column+" = ?", value)
		}
	}
	var rows []entity.PlacementRow
	return rows, query.Order("students.name ASC").Scan(&rows).Error
}

func (s *Service) ExportPlacements(ctx context.Context, format string, filters map[string]string) ([]byte, string, error) {
	rows, err := s.Placements(ctx, filters)
	if err != nil {
		return nil, "", err
	}
	headers := []string{"NIS", "Nama Siswa", "Kelas", "Jurusan", "Perusahaan", "Guru Pembimbing", "Mulai", "Selesai", "Status"}
	values := make([][]any, len(rows))
	for index, row := range rows {
		values[index] = []any{row.NIS, row.StudentName, row.ClassName, row.MajorName, row.CompanyName, row.SupervisorName, row.StartDate, row.EndDate, row.Status}
	}
	if format == "pdf" {
		data, err := s.pdf.Generate("Rekap Penempatan PKL", headers, values)
		return data, "application/pdf", err
	}
	data, err := s.excel.Generate("Penempatan PKL", headers, values)
	return data, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", err
}

func (s *Service) Expirations(ctx context.Context, days int) ([]entity.ExpirationItem, error) {
	if days < 1 {
		days = 30
	}
	var result []entity.ExpirationItem
	end := time.Now().AddDate(0, 0, days)
	var companies []struct {
		ID, Name       string
		CooperationEnd time.Time
	}
	if err := s.db.WithContext(ctx).Table("companies").Select("id, name, cooperation_end").
		Where("deleted_at IS NULL AND cooperation_end BETWEEN ? AND ?", time.Now(), end).Scan(&companies).Error; err != nil {
		return nil, err
	}
	for _, item := range companies {
		result = append(result, entity.ExpirationItem{
			Type: "company_cooperation", ID: item.ID, Name: item.Name,
			ExpiresAt: item.CooperationEnd.Format("2006-01-02"),
			DaysLeft:  int(time.Until(item.CooperationEnd).Hours() / 24),
		})
	}
	return result, nil
}

func count(query *gorm.DB) int64 {
	var total int64
	_ = query.Count(&total).Error
	return total
}
