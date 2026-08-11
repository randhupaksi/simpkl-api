package service

import (
	"context"
	"sort"
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

func (s *Service) Dashboard(ctx context.Context, periodID string, scopes map[string]string) (*entity.Dashboard, error) {
	result := &entity.Dashboard{}
	period, err := s.dashboardPeriod(ctx, periodID)
	if err != nil {
		return nil, err
	}
	result.Period = period.dashboardPeriod()

	studentQuery := scopedStudents(s.db.WithContext(ctx).Table("students s").Where("s.deleted_at IS NULL"), scopes)
	if period.ID != "" {
		studentQuery = studentQuery.Where("s.cohort = ?", period.Cohort)
	}
	if err := studentQuery.Count(&result.TotalStudents).Error; err != nil {
		return nil, err
	}
	result.UnplacedStudents = count(studentQuery.Session(&gorm.Session{}).Where("s.pkl_status IN ?", []string{"unplaced", "unregistered", "placement_process"}))
	result.PlacedStudents = count(studentQuery.Session(&gorm.Session{}).Where("s.pkl_status NOT IN ?", []string{"unplaced", "unregistered", "placement_process", "not_participating", "cancelled"}))
	result.ActivePlacements = count(studentQuery.Session(&gorm.Session{}).Where("s.pkl_status = ?", "active"))
	result.ActiveCompanies = count(s.db.WithContext(ctx).Table("companies").Where("deleted_at IS NULL AND status = ?", "active"))

	placementQuery := scopedPlacements(s.db.WithContext(ctx).Table("placements p").Where("p.deleted_at IS NULL"), scopes)
	if period.ID != "" {
		placementQuery = placementQuery.Where("p.period_id = ?", period.ID)
	}
	today := time.Now()
	result.StartingSoon = count(placementQuery.Session(&gorm.Session{}).Where("p.start_date BETWEEN ? AND ?", today, today.AddDate(0, 0, 7)))
	result.EndingSoon = count(placementQuery.Session(&gorm.Session{}).Where("p.end_date BETWEEN ? AND ?", today, today.AddDate(0, 0, 14)))

	readinessQuery := scopedReadiness(s.db.WithContext(ctx).Table("administrative_readiness r").Where("r.deleted_at IS NULL"), scopes)
	if period.ID != "" {
		readinessQuery = readinessQuery.Where("r.period_id = ?", period.ID)
	}
	result.Readiness.Total = count(readinessQuery.Session(&gorm.Session{}))
	result.Readiness.Ready = count(readinessQuery.Session(&gorm.Session{}).Where("r.status IN ?", []string{"ready", "started", "completed"}))
	result.Readiness.Attention = count(readinessQuery.Session(&gorm.Session{}).Where("r.status = ?", "attention"))
	result.Readiness.Incomplete = count(readinessQuery.Session(&gorm.Session{}).Where("r.status = ?", "incomplete"))
	result.IncompleteDocuments = result.Readiness.Attention + result.Readiness.Incomplete
	var readinessAverage struct{ Average float64 }
	if err := readinessQuery.Session(&gorm.Session{}).Select("COALESCE(AVG(r.percentage), 0) AS average").Scan(&readinessAverage).Error; err != nil {
		return nil, err
	}
	result.Readiness.Average = readinessAverage.Average

	result.PlacementStatuses = dashboardStatuses(studentQuery)
	result.MajorProgress = s.dashboardMajorProgress(ctx, period, scopes)
	result.CompanyCapacity = s.dashboardCapacity(ctx, period, scopes)
	result.Agenda = s.dashboardAgenda(ctx, period, scopes)
	result.Priorities = dashboardPriorities(result, companyExpirations(ctx, s.db.WithContext(ctx)))
	result.RecentActivities = s.dashboardActivities(ctx, scopes)
	return result, nil
}

type dashboardPeriodRecord struct {
	ID           string
	Name         string
	AcademicYear string
	Semester     string
	Status       string
	StartDate    time.Time
	EndDate      time.Time
	Cohort       int
}

func (s *Service) dashboardPeriod(ctx context.Context, periodID string) (dashboardPeriodRecord, error) {
	var period dashboardPeriodRecord
	query := s.db.WithContext(ctx).Table("periods").Where("deleted_at IS NULL")
	if periodID != "" {
		err := query.Where("id = ?", periodID).Take(&period).Error
		if err == gorm.ErrRecordNotFound {
			return period, nil
		}
		return period, err
	}
	err := query.Where("status = ?", "active").Order("start_date DESC").Take(&period).Error
	if err == gorm.ErrRecordNotFound {
		err = query.Order("start_date DESC").Take(&period).Error
	}
	if err == gorm.ErrRecordNotFound {
		return period, nil
	}
	return period, err
}

func (p dashboardPeriodRecord) dashboardPeriod() entity.DashboardPeriod {
	return entity.DashboardPeriod{ID: p.ID, Name: p.Name, AcademicYear: p.AcademicYear, Semester: p.Semester, Status: p.Status, StartDate: formatDate(p.StartDate), EndDate: formatDate(p.EndDate)}
}

func scopedStudents(query *gorm.DB, scopes map[string]string) *gorm.DB {
	if scopes["major_id"] != "" {
		query = query.Where("s.major_id = ?", scopes["major_id"])
	}
	if scopes["class_id"] != "" {
		query = query.Where("s.class_id = ?", scopes["class_id"])
	}
	return query
}

func scopedPlacements(query *gorm.DB, scopes map[string]string) *gorm.DB {
	if scopes["major_id"] == "" && scopes["class_id"] == "" {
		return query
	}
	query = query.Joins("JOIN students dashboard_students ON dashboard_students.id = p.student_id AND dashboard_students.deleted_at IS NULL")
	if scopes["major_id"] != "" {
		query = query.Where("dashboard_students.major_id = ?", scopes["major_id"])
	}
	if scopes["class_id"] != "" {
		query = query.Where("dashboard_students.class_id = ?", scopes["class_id"])
	}
	return query
}

func scopedReadiness(query *gorm.DB, scopes map[string]string) *gorm.DB {
	if scopes["major_id"] == "" && scopes["class_id"] == "" {
		return query
	}
	query = query.Joins("JOIN students dashboard_students ON dashboard_students.id = r.student_id AND dashboard_students.deleted_at IS NULL")
	if scopes["major_id"] != "" {
		query = query.Where("dashboard_students.major_id = ?", scopes["major_id"])
	}
	if scopes["class_id"] != "" {
		query = query.Where("dashboard_students.class_id = ?", scopes["class_id"])
	}
	return query
}

func dashboardStatuses(students *gorm.DB) []entity.DashboardBreakdown {
	statuses := []struct {
		key, label string
		values     []string
	}{
		{"unplaced", "Belum ditempatkan", []string{"unplaced", "unregistered", "placement_process"}},
		{"ready", "Siap berangkat", []string{"awaiting_documents", "ready"}},
		{"active", "Sedang PKL", []string{"active"}},
		{"completed", "Selesai PKL", []string{"completed"}},
	}
	result := make([]entity.DashboardBreakdown, 0, len(statuses))
	for _, status := range statuses {
		result = append(result, entity.DashboardBreakdown{
			Key: status.key, Label: status.label,
			Value: count(students.Session(&gorm.Session{}).Where("s.pkl_status IN ?", status.values)),
		})
	}
	return result
}

func (s *Service) dashboardMajorProgress(ctx context.Context, period dashboardPeriodRecord, scopes map[string]string) []entity.DashboardMajor {
	query := scopedStudents(s.db.WithContext(ctx).Table("students s").Joins("JOIN majors m ON m.id = s.major_id AND m.deleted_at IS NULL").Where("s.deleted_at IS NULL"), scopes)
	if period.ID != "" {
		query = query.Where("s.cohort = ?", period.Cohort)
	}
	items := make([]entity.DashboardMajor, 0)
	_ = query.Select(`m.id AS major_id, m.name AS major_name, COUNT(s.id) AS total_students,
		SUM(CASE WHEN s.pkl_status NOT IN ('unplaced', 'unregistered', 'placement_process', 'not_participating', 'cancelled') THEN 1 ELSE 0 END) AS placed_students,
		SUM(CASE WHEN s.pkl_status = 'active' THEN 1 ELSE 0 END) AS active_students`).
		Group("m.id, m.name").Order("total_students DESC, m.name ASC").Scan(&items).Error
	return items
}

func (s *Service) dashboardCapacity(ctx context.Context, period dashboardPeriodRecord, scopes map[string]string) entity.DashboardCapacity {
	result := entity.DashboardCapacity{}
	_ = s.db.WithContext(ctx).Table("companies").Where("deleted_at IS NULL AND status = ?", "active").Select("COALESCE(SUM(capacity), 0) AS total").Scan(&result).Error
	placements := scopedPlacements(s.db.WithContext(ctx).Table("placements p").Where("p.deleted_at IS NULL AND p.status NOT IN ?", []string{"cancelled", "draft"}), scopes)
	if period.ID != "" {
		placements = placements.Where("p.period_id = ?", period.ID)
	}
	result.Used = count(placements)
	return result
}

func (s *Service) dashboardAgenda(ctx context.Context, period dashboardPeriodRecord, scopes map[string]string) []entity.DashboardAgendaItem {
	now := time.Now()
	items := make([]entity.DashboardAgendaItem, 0, 8)
	placements := scopedPlacements(s.db.WithContext(ctx).Table("placements p").Joins("JOIN students s ON s.id = p.student_id").Joins("JOIN companies c ON c.id = p.company_id").Where("p.deleted_at IS NULL"), scopes)
	if period.ID != "" {
		placements = placements.Where("p.period_id = ?", period.ID)
	}
	var starting []struct {
		ID, StudentName, CompanyName string
		StartDate                    time.Time
	}
	_ = placements.Session(&gorm.Session{}).Select("p.id, s.name AS student_name, c.name AS company_name, p.start_date").Where("p.start_date BETWEEN ? AND ?", now, now.AddDate(0, 0, 14)).Order("p.start_date ASC").Limit(3).Scan(&starting).Error
	for _, item := range starting {
		items = append(items, entity.DashboardAgendaItem{ID: "placement-start-" + item.ID, Type: "placement_start", Title: "PKL akan dimulai", Description: item.StudentName + " · " + item.CompanyName, Date: formatDate(item.StartDate), DaysLeft: daysUntil(now, item.StartDate), Tone: "info", Path: "/placements/" + item.ID})
	}
	var ending []struct {
		ID, StudentName, CompanyName string
		EndDate                      time.Time
	}
	_ = placements.Session(&gorm.Session{}).Select("p.id, s.name AS student_name, c.name AS company_name, p.end_date").Where("p.end_date BETWEEN ? AND ?", now, now.AddDate(0, 0, 14)).Order("p.end_date ASC").Limit(3).Scan(&ending).Error
	for _, item := range ending {
		items = append(items, entity.DashboardAgendaItem{ID: "placement-end-" + item.ID, Type: "placement_end", Title: "PKL akan berakhir", Description: item.StudentName + " · " + item.CompanyName, Date: formatDate(item.EndDate), DaysLeft: daysUntil(now, item.EndDate), Tone: "warning", Path: "/placements/" + item.ID})
	}
	var expiring []struct {
		ID, Name       string
		CooperationEnd time.Time
	}
	_ = s.db.WithContext(ctx).Table("companies").Where("deleted_at IS NULL AND cooperation_end BETWEEN ? AND ?", now, now.AddDate(0, 0, 30)).Select("id, name, cooperation_end").Order("cooperation_end ASC").Limit(3).Scan(&expiring).Error
	for _, item := range expiring {
		items = append(items, entity.DashboardAgendaItem{ID: "company-expiration-" + item.ID, Type: "company_expiration", Title: "Kerja sama akan berakhir", Description: item.Name, Date: formatDate(item.CooperationEnd), DaysLeft: daysUntil(now, item.CooperationEnd), Tone: "danger", Path: "/companies/" + item.ID})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].DaysLeft < items[j].DaysLeft })
	if len(items) > 6 {
		return items[:6]
	}
	return items
}

func dashboardPriorities(result *entity.Dashboard, expiringCompanies int64) []entity.DashboardPriority {
	items := []entity.DashboardPriority{}
	if result.UnplacedStudents > 0 {
		items = append(items, entity.DashboardPriority{Key: "unplaced", Title: "Siswa belum ditempatkan", Description: "Tentukan perusahaan dan jadwal PKL untuk peserta yang masih menunggu.", Value: result.UnplacedStudents, Tone: "warning", Path: "/students"})
	}
	if result.IncompleteDocuments > 0 {
		items = append(items, entity.DashboardPriority{Key: "readiness", Title: "Kesiapan administrasi perlu ditinjau", Description: "Ada persyaratan PKL yang belum lengkap atau membutuhkan perhatian.", Value: result.IncompleteDocuments, Tone: "danger", Path: "/administrative-readiness"})
	}
	if result.EndingSoon > 0 {
		items = append(items, entity.DashboardPriority{Key: "ending", Title: "PKL akan segera berakhir", Description: "Siapkan evaluasi, laporan, dan dokumen penyelesaian PKL.", Value: result.EndingSoon, Tone: "info", Path: "/placements"})
	}
	if expiringCompanies > 0 {
		items = append(items, entity.DashboardPriority{Key: "company-expiration", Title: "Kerja sama perusahaan mendekati akhir", Description: "Tinjau perpanjangan kerja sama sebelum masa berlaku berakhir.", Value: expiringCompanies, Tone: "danger", Path: "/companies"})
	}
	return items
}

func companyExpirations(ctx context.Context, db *gorm.DB) int64 {
	return count(db.WithContext(ctx).Table("companies").Where("deleted_at IS NULL AND cooperation_end BETWEEN ? AND ?", time.Now(), time.Now().AddDate(0, 0, 30)))
}

func (s *Service) dashboardActivities(ctx context.Context, scopes map[string]string) []entity.DashboardActivity {
	if scopes["major_id"] != "" || scopes["class_id"] != "" {
		return []entity.DashboardActivity{}
	}
	var rows []struct {
		ID, Action, Resource, ActorName string
		CreatedAt                       time.Time
	}
	err := s.db.WithContext(ctx).Table("audit_logs a").Joins("LEFT JOIN users u ON u.id = a.actor_id").Select("a.id, a.action, a.resource, COALESCE(u.name, 'Sistem') AS actor_name, a.created_at").Order("a.created_at DESC").Limit(6).Scan(&rows).Error
	if err != nil {
		return []entity.DashboardActivity{}
	}
	items := make([]entity.DashboardActivity, 0, len(rows))
	for _, row := range rows {
		items = append(items, entity.DashboardActivity{ID: row.ID, Action: row.Action, Resource: row.Resource, ActorName: row.ActorName, CreatedAt: row.CreatedAt.Format(time.RFC3339)})
	}
	return items
}

func formatDate(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02")
}
func daysUntil(now, date time.Time) int { return int(date.Sub(now).Hours() / 24) }

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
