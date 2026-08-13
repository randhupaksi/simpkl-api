package service

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"simpkl-api/internal/modules/documentautomation/entity"
	"simpkl-api/internal/platform/documentgen"
	excelreport "simpkl-api/internal/platform/report/excel"
	"simpkl-api/internal/platform/storage"
	apperrors "simpkl-api/internal/shared/errors"
	"simpkl-api/internal/shared/types"
)

type Service struct {
	db      *gorm.DB
	storage storage.Storage
	auditor types.Auditor
	excel   *excelreport.Generator
}

func New(db *gorm.DB, fileStorage storage.Storage, auditor types.Auditor) *Service {
	return &Service{db: db, storage: fileStorage, auditor: auditor, excel: excelreport.New()}
}

type Filters struct {
	PeriodID     string   `json:"period_id" form:"period_id"`
	ClassID      string   `json:"class_id" form:"class_id"`
	MajorID      string   `json:"major_id" form:"major_id"`
	CompanyID    string   `json:"company_id" form:"company_id"`
	SupervisorID string   `json:"supervisor_id" form:"supervisor_id"`
	PlacementIDs []string `json:"placement_ids" form:"placement_ids"`
}

type GenerateInput struct {
	Name          string    `json:"name"`
	Filters       Filters   `json:"filters"`
	TemplateCodes []string  `json:"template_codes"`
	Formats       []string  `json:"formats"`
	SignatoryID   string    `json:"signatory_id"`
	LetterDate    time.Time `json:"letter_date"`
	RequestedBy   string    `json:"-"`
}

type PlacementData struct {
	PlacementID              string    `json:"placement_id"`
	PeriodID                 string    `json:"period_id"`
	PeriodName               string    `json:"period_name"`
	AcademicYear             string    `json:"academic_year"`
	StudentID                string    `json:"student_id"`
	StudentNIS               string    `json:"student_nis"`
	StudentNISN              string    `json:"student_nisn"`
	StudentName              string    `json:"student_name"`
	ParentName               string    `json:"parent_name"`
	ParentPhone              string    `json:"parent_phone"`
	ClassName                string    `json:"class_name"`
	MajorName                string    `json:"major_name"`
	CompanyName              string    `json:"company_name"`
	CompanyAddress           string    `json:"company_address"`
	CompanyCity              string    `json:"company_city"`
	ContactName              string    `json:"contact_name"`
	ContactPosition          string    `json:"contact_position"`
	SupervisorName           string    `json:"supervisor_name"`
	SupervisorEmployeeNumber string    `json:"supervisor_employee_number"`
	SupervisorPosition       string    `json:"supervisor_position"`
	PlacementDivision        string    `json:"placement_division"`
	PlacementPosition        string    `json:"placement_position"`
	PlacementAddress         string    `json:"placement_address"`
	PlacementStart           time.Time `json:"placement_start"`
	PlacementEnd             time.Time `json:"placement_end"`
	PlacementStatus          string    `json:"placement_status"`
}

type ValidationIssue struct {
	PlacementID string `json:"placement_id,omitempty"`
	StudentName string `json:"student_name,omitempty"`
	Field       string `json:"field"`
	Message     string `json:"message"`
	Severity    string `json:"severity"`
}

type Preview struct {
	PlacementCount int               `json:"placement_count"`
	DocumentCount  int               `json:"document_count"`
	Ready          bool              `json:"ready"`
	Issues         []ValidationIssue `json:"issues"`
	Placements     []PlacementData   `json:"placements"`
}

type artifact struct {
	name string
	mime string
	data []byte
}

func (s *Service) Profile(ctx context.Context) (*entity.SchoolProfile, error) {
	var profile entity.SchoolProfile
	err := s.db.WithContext(ctx).Order("created_at ASC").First(&profile).Error
	return &profile, err
}

func (s *Service) UpdateProfile(ctx context.Context, input entity.SchoolProfile) (*entity.SchoolProfile, error) {
	if strings.TrimSpace(input.InstitutionName) == "" || strings.TrimSpace(input.Address) == "" {
		return nil, validation("Nama dan alamat institusi wajib diisi")
	}
	profile, err := s.Profile(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		input.ID = uuid.NewString()
		if input.Timezone == "" {
			input.Timezone = "Asia/Jakarta"
		}
		if err := s.db.WithContext(ctx).Create(&input).Error; err != nil {
			return nil, err
		}
		return &input, nil
	}
	if err != nil {
		return nil, err
	}
	updates := map[string]any{
		"institution_name": input.InstitutionName, "institution_type": input.InstitutionType,
		"npsn": input.NPSN, "address": input.Address, "village": input.Village,
		"district": input.District, "city": input.City, "province": input.Province,
		"postal_code": input.PostalCode, "phone": input.Phone, "email": input.Email,
		"website": input.Website, "letterhead_tagline": input.LetterheadTagline,
		"timezone": fallback(input.Timezone, "Asia/Jakarta"),
	}
	if err := s.db.WithContext(ctx).Model(profile).Updates(updates).Error; err != nil {
		return nil, err
	}
	return s.Profile(ctx)
}

func (s *Service) Signatories(ctx context.Context) ([]entity.Signatory, error) {
	var items []entity.Signatory
	return items, s.db.WithContext(ctx).Order("is_default DESC, name ASC").Find(&items).Error
}

func (s *Service) SaveSignatory(ctx context.Context, id string, input entity.Signatory) (*entity.Signatory, error) {
	if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Title) == "" {
		return nil, validation("Nama dan jabatan penandatangan wajib diisi")
	}
	if input.RoleCode == "" {
		input.RoleCode = "principal"
	}
	if input.Status == "" {
		input.Status = "active"
	}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if input.IsDefault {
			if err := tx.Model(&entity.Signatory{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
				return err
			}
		}
		if id == "" {
			input.ID = uuid.NewString()
			return tx.Create(&input).Error
		}
		return tx.Model(&entity.Signatory{}).Where("id = ?", id).Updates(map[string]any{
			"name": input.Name, "title": input.Title, "employee_number": input.EmployeeNumber,
			"role_code": input.RoleCode, "is_default": input.IsDefault, "status": input.Status,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	var result entity.Signatory
	if id == "" {
		id = input.ID
	}
	return &result, s.db.WithContext(ctx).First(&result, "id = ?", id).Error
}

func (s *Service) DeleteSignatory(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&entity.Signatory{}, "id = ?", id).Error
}

func (s *Service) Templates(ctx context.Context) ([]entity.DocumentTemplate, error) {
	var items []entity.DocumentTemplate
	err := s.db.WithContext(ctx).Order("code ASC, version DESC").Find(&items).Error
	return items, err
}

func (s *Service) SaveTemplate(ctx context.Context, id string, input entity.DocumentTemplate) (*entity.DocumentTemplate, error) {
	if strings.TrimSpace(input.Code) == "" || strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.BodyTemplate) == "" {
		return nil, validation("Kode, nama, dan isi template wajib diisi")
	}
	input.Code = sanitizeCode(input.Code)
	if input.Category == "" {
		input.Category = "letter"
	}
	if input.NumberPattern == "" {
		input.NumberPattern = "{{sequence}}/{{code}}/{{month_roman}}/{{year}}"
	}
	if input.SubjectTemplate == "" {
		input.SubjectTemplate = input.Name
	}
	input.IsActive = true
	if id == "" {
		var maxVersion int
		_ = s.db.WithContext(ctx).Model(&entity.DocumentTemplate{}).Where("code = ?", input.Code).Select("COALESCE(MAX(version), 0)").Scan(&maxVersion).Error
		input.ID, input.Version = uuid.NewString(), maxVersion+1
		if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&entity.DocumentTemplate{}).Where("code = ? AND is_active = ?", input.Code, true).Update("is_active", false).Error; err != nil {
				return err
			}
			return tx.Create(&input).Error
		}); err != nil {
			return nil, err
		}
		return &input, nil
	}
	var previous entity.DocumentTemplate
	if err := s.db.WithContext(ctx).First(&previous, "id = ?", id).Error; err != nil {
		return nil, err
	}
	input.ID = uuid.NewString()
	input.Code = previous.Code
	input.Version = previous.Version + 1
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&previous).Update("is_active", false).Error; err != nil {
			return err
		}
		return tx.Create(&input).Error
	})
	return &input, err
}

func (s *Service) SetTemplateActive(ctx context.Context, id string, active bool) error {
	var template entity.DocumentTemplate
	if err := s.db.WithContext(ctx).First(&template, "id = ?", id).Error; err != nil {
		return err
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if active {
			if err := tx.Model(&entity.DocumentTemplate{}).Where("code = ?", template.Code).Update("is_active", false).Error; err != nil {
				return err
			}
		}
		return tx.Model(&template).Update("is_active", active).Error
	})
}

func (s *Service) Preview(ctx context.Context, filters Filters, templateCodes []string, formats []string) (*Preview, error) {
	placements, err := s.placements(ctx, filters)
	if err != nil {
		return nil, err
	}
	issues := s.validate(ctx, placements, templateCodes)
	if issues == nil {
		issues = make([]ValidationIssue, 0)
	}
	if placements == nil {
		placements = make([]PlacementData, 0)
	}
	count := 0
	for _, code := range unique(templateCodes) {
		if code == "placement_recap" {
			count += 1
			continue
		}
		count += len(placements) * max(1, len(unique(formats)))
	}
	return &Preview{PlacementCount: len(placements), DocumentCount: count, Ready: !hasErrors(issues), Issues: issues, Placements: placements}, nil
}

func (s *Service) Generate(ctx context.Context, input GenerateInput) (*entity.GenerationBatch, error) {
	if len(input.TemplateCodes) == 0 {
		return nil, validation("Pilih minimal satu jenis dokumen")
	}
	if len(input.Formats) == 0 {
		input.Formats = []string{"docx", "pdf"}
	}
	input.Formats = unique(input.Formats)
	for _, format := range input.Formats {
		if format != "docx" && format != "pdf" {
			return nil, validation("Format surat hanya boleh DOCX atau PDF")
		}
	}
	if input.LetterDate.IsZero() {
		input.LetterDate = time.Now()
	}
	placements, err := s.placements(ctx, input.Filters)
	if err != nil {
		return nil, err
	}
	if len(placements) == 0 {
		return nil, validation("Tidak ada penempatan yang cocok dengan filter")
	}
	issues := s.validate(ctx, placements, input.TemplateCodes)
	if hasErrors(issues) {
		return nil, &apperrors.AppError{Status: 422, Code: "INCOMPLETE_DOCUMENT_DATA", Message: "Data belum lengkap untuk membuat dokumen"}
	}
	profile, _ := s.Profile(ctx)
	signatory, err := s.resolveSignatory(ctx, input.SignatoryID)
	if err != nil {
		return nil, validation("Penandatangan aktif belum tersedia")
	}
	templates, err := s.activeTemplates(ctx, input.TemplateCodes)
	if err != nil {
		return nil, err
	}
	filterJSON, _ := json.Marshal(input.Filters)
	periodID := pointer(placements[0].PeriodID)
	requestedBy := optionalPointer(input.RequestedBy)
	batch := entity.GenerationBatch{
		PeriodID: periodID, RequestedBy: requestedBy, Name: fallback(input.Name, "Paket Dokumen PKL "+input.LetterDate.Format("2006-01-02")),
		Status: "processing", FiltersJSON: string(filterJSON),
	}
	if err := s.db.WithContext(ctx).Create(&batch).Error; err != nil {
		return nil, err
	}
	created := make([]artifact, 0)
	failures := make([]string, 0)
	for _, template := range templates {
		if template.Code == "placement_recap" || template.Category == "spreadsheet" {
			batch.RequestedCount++
			file, fileErr := s.generateRecap(placements, input.LetterDate)
			if fileErr != nil {
				batch.FailedCount++
				failures = append(failures, fileErr.Error())
				continue
			}
			if _, fileErr = s.persistArtifact(ctx, &batch, template, nil, nil, signatory, input.RequestedBy, "REKAP-PKL-"+input.LetterDate.Format("2006"), template.Name, file, placements); fileErr != nil {
				batch.FailedCount++
				failures = append(failures, fileErr.Error())
				continue
			}
			batch.GeneratedCount++
			created = append(created, file)
			continue
		}
		for index := range placements {
			placement := &placements[index]
			number, numberErr := s.nextNumber(ctx, template, input.LetterDate)
			if numberErr != nil {
				failures = append(failures, numberErr.Error())
				batch.RequestedCount += len(input.Formats)
				batch.FailedCount += len(input.Formats)
				continue
			}
			letter := makeLetter(*profile, *signatory, template, *placement, number, input.LetterDate)
			for _, format := range unique(input.Formats) {
				if format != "docx" && format != "pdf" {
					continue
				}
				batch.RequestedCount++
				file, fileErr := renderLetter(letter, template, *placement, format)
				if fileErr == nil {
					_, fileErr = s.persistArtifact(ctx, &batch, template, placement, pointer(placement.StudentID), signatory, input.RequestedBy, number, letter.Subject, file, placement)
				}
				if fileErr != nil {
					batch.FailedCount++
					failures = append(failures, fileErr.Error())
					continue
				}
				batch.GeneratedCount++
				created = append(created, file)
			}
		}
	}
	if len(created) > 0 {
		zipArtifact, zipErr := makeZIP(batch.Name, created)
		if zipErr == nil {
			saved, saveErr := s.storage.Save(ctx, bytes.NewReader(zipArtifact.data), zipArtifact.name)
			if saveErr == nil {
				batch.ArchiveName, batch.ArchivePath, batch.ArchiveSize = zipArtifact.name, saved.Path, saved.Size
			}
		}
	}
	now := time.Now()
	batch.CompletedAt = &now
	batch.Status = "completed"
	if batch.FailedCount > 0 {
		batch.Status = "completed_with_errors"
	}
	if batch.GeneratedCount == 0 {
		batch.Status = "failed"
	}
	batch.ErrorSummary = strings.Join(unique(failures), "; ")
	if err := s.db.WithContext(ctx).Save(&batch).Error; err != nil {
		return nil, err
	}
	_ = s.auditor.Record(types.AuditEvent{ActorID: input.RequestedBy, Action: "generate", Resource: "document_automation", ResourceID: batch.ID, After: batch})
	return &batch, nil
}

func (s *Service) Batches(ctx context.Context, limit int) ([]entity.GenerationBatch, error) {
	if limit < 1 || limit > 100 {
		limit = 30
	}
	var items []entity.GenerationBatch
	return items, s.db.WithContext(ctx).Order("created_at DESC").Limit(limit).Find(&items).Error
}

func (s *Service) Batch(ctx context.Context, id string) (*entity.GenerationBatch, []entity.GeneratedDocument, error) {
	var batch entity.GenerationBatch
	if err := s.db.WithContext(ctx).First(&batch, "id = ?", id).Error; err != nil {
		return nil, nil, err
	}
	documents, err := s.Documents(ctx, map[string]string{"batch_id": id}, 200)
	return &batch, documents, err
}

func (s *Service) Documents(ctx context.Context, filters map[string]string, limit int) ([]entity.GeneratedDocument, error) {
	if limit < 1 || limit > 500 {
		limit = 50
	}
	query := s.db.WithContext(ctx).Table("generated_documents gd").
		Select("gd.*, students.name AS student_name, periods.name AS period_name").
		Joins("LEFT JOIN students ON students.id = gd.student_id").
		Joins("LEFT JOIN periods ON periods.id = gd.period_id").Where("gd.deleted_at IS NULL")
	columns := map[string]string{"batch_id": "gd.batch_id", "student_id": "gd.student_id", "period_id": "gd.period_id", "template_code": "gd.template_code", "format": "gd.format"}
	for key, value := range filters {
		if column, ok := columns[key]; ok && value != "" {
			query = query.Where(column+" = ?", value)
		}
	}
	var items []entity.GeneratedDocument
	return items, query.Order("gd.generated_at DESC").Limit(limit).Scan(&items).Error
}

func (s *Service) OpenDocument(ctx context.Context, id string) (*entity.GeneratedDocument, io.ReadCloser, error) {
	var item entity.GeneratedDocument
	if err := s.db.WithContext(ctx).First(&item, "id = ?", id).Error; err != nil {
		return nil, nil, err
	}
	file, err := s.storage.Open(ctx, item.Path)
	return &item, file, err
}

func (s *Service) OpenBatch(ctx context.Context, id string) (*entity.GenerationBatch, io.ReadCloser, error) {
	var batch entity.GenerationBatch
	if err := s.db.WithContext(ctx).First(&batch, "id = ?", id).Error; err != nil {
		return nil, nil, err
	}
	if batch.ArchivePath == "" {
		return nil, nil, gorm.ErrRecordNotFound
	}
	file, err := s.storage.Open(ctx, batch.ArchivePath)
	return &batch, file, err
}

func (s *Service) placements(ctx context.Context, filters Filters) ([]PlacementData, error) {
	query := s.db.WithContext(ctx).Table("placements p").Select(`
		p.id AS placement_id, p.period_id, periods.name AS period_name, periods.academic_year,
		students.id AS student_id, students.nis AS student_nis, students.nisn AS student_nisn,
		students.name AS student_name, students.parent_name, students.parent_phone,
		classes.name AS class_name, majors.name AS major_name,
		companies.name AS company_name, companies.address AS company_address, companies.city AS company_city,
		company_contacts.name AS contact_name, company_contacts.position AS contact_position,
		supervisors.name AS supervisor_name, supervisors.employee_number AS supervisor_employee_number,
		supervisors.position AS supervisor_position, p.division AS placement_division,
		p.position AS placement_position, p.address AS placement_address,
		p.start_date AS placement_start, p.end_date AS placement_end, p.status AS placement_status`).
		Joins("JOIN periods ON periods.id = p.period_id").Joins("JOIN students ON students.id = p.student_id").
		Joins("LEFT JOIN classes ON classes.id = students.class_id").Joins("LEFT JOIN majors ON majors.id = students.major_id").
		Joins("JOIN companies ON companies.id = p.company_id").Joins("LEFT JOIN company_contacts ON company_contacts.id = p.company_contact_id").
		Joins("LEFT JOIN supervisors ON supervisors.id = p.supervisor_id").Where("p.deleted_at IS NULL")
	if filters.PeriodID != "" {
		query = query.Where("p.period_id = ?", filters.PeriodID)
	}
	if filters.ClassID != "" {
		query = query.Where("students.class_id = ?", filters.ClassID)
	}
	if filters.MajorID != "" {
		query = query.Where("students.major_id = ?", filters.MajorID)
	}
	if filters.CompanyID != "" {
		query = query.Where("p.company_id = ?", filters.CompanyID)
	}
	if filters.SupervisorID != "" {
		query = query.Where("p.supervisor_id = ?", filters.SupervisorID)
	}
	if len(filters.PlacementIDs) > 0 {
		query = query.Where("p.id IN ?", filters.PlacementIDs)
	}
	rows := make([]PlacementData, 0)
	return rows, query.Order("students.name ASC").Scan(&rows).Error
}

func (s *Service) validate(ctx context.Context, placements []PlacementData, templateCodes []string) []ValidationIssue {
	issues := make([]ValidationIssue, 0)
	codes := unique(templateCodes)
	if len(codes) == 0 {
		issues = append(issues, ValidationIssue{Field: "template_codes", Message: "Pilih minimal satu template dokumen", Severity: "error"})
	} else {
		var activeCount int64
		if err := s.db.WithContext(ctx).Model(&entity.DocumentTemplate{}).Where("code IN ? AND is_active = ?", codes, true).Count(&activeCount).Error; err != nil || activeCount != int64(len(codes)) {
			issues = append(issues, ValidationIssue{Field: "template_codes", Message: "Salah satu template tidak tersedia atau memiliki konfigurasi versi aktif yang tidak valid", Severity: "error"})
		}
	}
	profile, err := s.Profile(ctx)
	if err != nil || strings.TrimSpace(profile.InstitutionName) == "" || profile.InstitutionName == "Nama Institusi" {
		issues = append(issues, ValidationIssue{Field: "school_profile", Message: "Lengkapi profil institusi sebelum membuat dokumen", Severity: "error"})
	}
	signatory, signatoryErr := s.resolveSignatory(ctx, "")
	if signatoryErr != nil || strings.TrimSpace(signatory.Name) == "" || signatory.Name == "Nama Kepala Sekolah" {
		issues = append(issues, ValidationIssue{Field: "signatory", Message: "Tetapkan penandatangan aktif utama", Severity: "error"})
	}
	if len(placements) == 0 {
		issues = append(issues, ValidationIssue{Field: "placements", Message: "Tidak ada data penempatan yang dipilih", Severity: "error"})
	}
	for _, row := range placements {
		add := func(field, message string) {
			issues = append(issues, ValidationIssue{PlacementID: row.PlacementID, StudentName: row.StudentName, Field: field, Message: message, Severity: "error"})
		}
		if row.StudentName == "" {
			add("student_name", "Nama siswa belum lengkap")
		}
		if row.StudentNIS == "" {
			add("student_nis", "NIS siswa belum lengkap")
		}
		if row.ClassName == "" {
			add("class_name", "Kelas siswa belum lengkap")
		}
		if row.MajorName == "" {
			add("major_name", "Jurusan siswa belum lengkap")
		}
		if row.CompanyName == "" {
			add("company_name", "Perusahaan penempatan belum lengkap")
		}
		if row.CompanyAddress == "" && row.PlacementAddress == "" {
			add("company_address", "Alamat perusahaan/penempatan belum lengkap")
		}
		if contains(codes, "supervisor_assignment") && row.SupervisorName == "" {
			add("supervisor_name", "Guru pembimbing belum dipilih")
		}
		if contains(codes, "parent_consent") && row.ParentName == "" {
			add("parent_name", "Nama orang tua/wali belum diisi")
		}
	}
	return issues
}

func (s *Service) resolveSignatory(ctx context.Context, id string) (*entity.Signatory, error) {
	var item entity.Signatory
	query := s.db.WithContext(ctx).Where("status = ?", "active")
	if id != "" {
		query = query.Where("id = ?", id)
	} else {
		query = query.Order("is_default DESC, created_at ASC")
	}
	return &item, query.First(&item).Error
}

func (s *Service) activeTemplates(ctx context.Context, codes []string) ([]entity.DocumentTemplate, error) {
	var items []entity.DocumentTemplate
	err := s.db.WithContext(ctx).Where("code IN ? AND is_active = ?", unique(codes), true).Order("code ASC").Find(&items).Error
	if err == nil && len(items) != len(unique(codes)) {
		return nil, validation("Salah satu template tidak tersedia atau tidak aktif")
	}
	return items, err
}

func (s *Service) nextNumber(ctx context.Context, template entity.DocumentTemplate, date time.Time) (string, error) {
	number := 0
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var sequence entity.LetterSequence
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("template_code = ? AND sequence_year = ? AND sequence_month = ?", template.Code, date.Year(), int(date.Month())).First(&sequence).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			sequence = entity.LetterSequence{ID: uuid.NewString(), TemplateCode: template.Code, SequenceYear: date.Year(), SequenceMonth: int(date.Month()), LastNumber: 1}
			if err := tx.Create(&sequence).Error; err != nil {
				return err
			}
			number = 1
			return nil
		}
		if err != nil {
			return err
		}
		sequence.LastNumber++
		number = sequence.LastNumber
		return tx.Save(&sequence).Error
	})
	if err != nil {
		return "", err
	}
	values := map[string]string{"sequence": fmt.Sprintf("%03d", number), "code": strings.ToUpper(template.Code), "month": fmt.Sprintf("%02d", int(date.Month())), "month_roman": romanMonth(date.Month()), "year": strconv.Itoa(date.Year())}
	return render(template.NumberPattern, values), nil
}

func makeLetter(profile entity.SchoolProfile, signatory entity.Signatory, template entity.DocumentTemplate, row PlacementData, number string, date time.Time) documentgen.Letter {
	values := placementValues(row, date)
	return documentgen.Letter{
		Profile:   documentgen.SchoolProfile{Name: profile.InstitutionName, Type: profile.InstitutionType, NPSN: profile.NPSN, Address: profile.Address, City: profile.City, Province: profile.Province, Phone: profile.Phone, Email: profile.Email, Website: profile.Website, Tagline: profile.LetterheadTagline},
		Signatory: documentgen.Signatory{Name: signatory.Name, Title: signatory.Title, EmployeeNumber: signatory.EmployeeNumber},
		Number:    number, Date: indonesianDate(date), Subject: render(template.SubjectTemplate, values), Recipient: fallback(row.ContactName, "Pimpinan "+row.CompanyName), RecipientAddress: fallback(row.CompanyCity, row.CompanyAddress), Body: render(template.BodyTemplate, values),
	}
}

func renderLetter(letter documentgen.Letter, template entity.DocumentTemplate, row PlacementData, format string) (artifact, error) {
	var data []byte
	var err error
	mime := "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	if format == "pdf" {
		data, err = documentgen.PDF(letter)
		mime = "application/pdf"
	} else {
		data, err = documentgen.DOCX(letter)
	}
	name := sanitizeFilename(template.Code+"_"+row.StudentName+"_"+row.StudentNIS) + "." + format
	return artifact{name: name, mime: mime, data: data}, err
}

func (s *Service) generateRecap(rows []PlacementData, date time.Time) (artifact, error) {
	headers := []string{"NIS", "NISN", "Nama Siswa", "Kelas", "Jurusan", "Perusahaan", "Alamat", "Guru Pembimbing", "Bagian", "Mulai", "Selesai", "Status"}
	values := make([][]any, len(rows))
	for i, row := range rows {
		values[i] = []any{row.StudentNIS, row.StudentNISN, row.StudentName, row.ClassName, row.MajorName, row.CompanyName, fallback(row.PlacementAddress, row.CompanyAddress), row.SupervisorName, row.PlacementDivision, row.PlacementStart.Format("2006-01-02"), row.PlacementEnd.Format("2006-01-02"), row.PlacementStatus}
	}
	data, err := s.excel.Generate("Rekap PKL", headers, values)
	return artifact{name: "rekap_penempatan_pkl_" + date.Format("20060102") + ".xlsx", mime: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data: data}, err
}

func (s *Service) persistArtifact(ctx context.Context, batch *entity.GenerationBatch, template entity.DocumentTemplate, placement *PlacementData, studentID *string, signatory *entity.Signatory, userID, number, title string, file artifact, snapshot any) (*entity.GeneratedDocument, error) {
	saved, err := s.storage.Save(ctx, bytes.NewReader(file.data), file.name)
	if err != nil {
		return nil, err
	}
	snapshotJSON, _ := json.Marshal(snapshot)
	checksum := sha256.Sum256(file.data)
	item := entity.GeneratedDocument{BatchID: pointer(batch.ID), TemplateID: pointer(template.ID), TemplateCode: template.Code, TemplateVersion: template.Version, StudentID: studentID, PeriodID: batch.PeriodID, SignatoryID: pointer(signatory.ID), GeneratedBy: optionalPointer(userID), DocumentNumber: number, Title: title, Format: strings.TrimPrefix(file.name[strings.LastIndex(file.name, "."):], "."), OriginalName: file.name, StoredName: saved.StoredName, Path: saved.Path, MimeType: file.mime, Size: saved.Size, Status: "final", DataSnapshot: string(snapshotJSON), ChecksumSHA256: hex.EncodeToString(checksum[:]), GeneratedAt: time.Now()}
	if placement != nil {
		item.PlacementID = pointer(placement.PlacementID)
		item.StudentID = pointer(placement.StudentID)
		item.PeriodID = pointer(placement.PeriodID)
	}
	if err := s.db.WithContext(ctx).Create(&item).Error; err != nil {
		_ = s.storage.Delete(ctx, saved.Path)
		return nil, err
	}
	return &item, nil
}

func makeZIP(name string, files []artifact) (artifact, error) {
	var buffer bytes.Buffer
	archive := zip.NewWriter(&buffer)
	used := map[string]int{}
	for _, file := range files {
		filename := file.name
		used[filename]++
		if used[filename] > 1 {
			extension := filename[strings.LastIndex(filename, "."):]
			filename = strings.TrimSuffix(filename, extension) + fmt.Sprintf("_%d", used[filename]) + extension
		}
		entry, err := archive.Create(filename)
		if err != nil {
			return artifact{}, err
		}
		if _, err := entry.Write(file.data); err != nil {
			return artifact{}, err
		}
	}
	if err := archive.Close(); err != nil {
		return artifact{}, err
	}
	return artifact{name: sanitizeFilename(name) + ".zip", mime: "application/zip", data: buffer.Bytes()}, nil
}

func placementValues(row PlacementData, date time.Time) map[string]string {
	return map[string]string{
		"academic_year": row.AcademicYear, "student_name": row.StudentName, "student_nis": row.StudentNIS,
		"student_nisn": row.StudentNISN, "class_name": row.ClassName, "major_name": row.MajorName,
		"parent_name": row.ParentName, "parent_phone": row.ParentPhone, "company_name": row.CompanyName,
		"company_address": fallback(row.PlacementAddress, row.CompanyAddress), "company_city": row.CompanyCity,
		"company_contact_name": row.ContactName, "company_contact_position": row.ContactPosition,
		"supervisor_name": row.SupervisorName, "supervisor_employee_number": row.SupervisorEmployeeNumber,
		"supervisor_position": row.SupervisorPosition, "placement_division": fallback(row.PlacementDivision, "-"),
		"placement_position": fallback(row.PlacementPosition, "Peserta PKL"), "placement_start": indonesianDate(row.PlacementStart),
		"placement_end": indonesianDate(row.PlacementEnd), "letter_date": indonesianDate(date), "period_name": row.PeriodName,
	}
}

var placeholderPattern = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_]+)\s*\}\}`)

func render(template string, values map[string]string) string {
	return placeholderPattern.ReplaceAllStringFunc(template, func(token string) string {
		match := placeholderPattern.FindStringSubmatch(token)
		if value, ok := values[match[1]]; ok {
			return value
		}
		return token
	})
}
func sanitizeCode(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = regexp.MustCompile(`[^a-z0-9_]+`).ReplaceAllString(value, "_")
	return strings.Trim(value, "_")
}
func sanitizeFilename(value string) string {
	value = strings.TrimSpace(value)
	value = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]+`).ReplaceAllString(value, "_")
	value = regexp.MustCompile(`\s+`).ReplaceAllString(value, "_")
	if len(value) > 150 {
		value = value[:150]
	}
	return strings.Trim(value, "._")
}
func romanMonth(month time.Month) string {
	values := []string{"I", "II", "III", "IV", "V", "VI", "VII", "VIII", "IX", "X", "XI", "XII"}
	return values[int(month)-1]
}
func indonesianDate(value time.Time) string {
	months := []string{"Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
	return fmt.Sprintf("%d %s %d", value.Day(), months[int(value.Month())-1], value.Year())
}
func fallback(value, alternative string) string {
	if strings.TrimSpace(value) == "" {
		return alternative
	}
	return value
}
func pointer(value string) *string { return &value }
func optionalPointer(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}
func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
func unique(values []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" && !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}
	sort.Strings(result)
	return result
}
func hasErrors(issues []ValidationIssue) bool {
	for _, issue := range issues {
		if issue.Severity == "error" {
			return true
		}
	}
	return false
}
func validation(message string) error {
	return &apperrors.AppError{Status: 422, Code: "VALIDATION_ERROR", Message: message}
}
