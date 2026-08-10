package entity

import (
	"time"

	"simpkl-api/internal/shared/types"
)

type SchoolProfile struct {
	types.BaseModel
	InstitutionName   string `gorm:"size:180;not null" json:"institution_name"`
	InstitutionType   string `gorm:"size:80;not null" json:"institution_type"`
	NPSN              string `gorm:"size:30" json:"npsn"`
	Address           string `gorm:"type:text;not null" json:"address"`
	Village           string `gorm:"size:100" json:"village"`
	District          string `gorm:"size:100" json:"district"`
	City              string `gorm:"size:100" json:"city"`
	Province          string `gorm:"size:100" json:"province"`
	PostalCode        string `gorm:"size:15" json:"postal_code"`
	Phone             string `gorm:"size:30" json:"phone"`
	Email             string `gorm:"size:150" json:"email"`
	Website           string `gorm:"size:255" json:"website"`
	LetterheadTagline string `gorm:"size:255" json:"letterhead_tagline"`
	LogoPath          string `gorm:"size:500" json:"logo_path"`
	Timezone          string `gorm:"size:50;not null" json:"timezone"`
}

func (SchoolProfile) TableName() string { return "school_profiles" }

type Signatory struct {
	types.BaseModel
	Name           string `gorm:"size:150;not null" json:"name"`
	Title          string `gorm:"size:150;not null" json:"title"`
	EmployeeNumber string `gorm:"size:60" json:"employee_number"`
	RoleCode       string `gorm:"size:60;not null" json:"role_code"`
	SignaturePath  string `gorm:"size:500" json:"signature_path"`
	IsDefault      bool   `gorm:"not null" json:"is_default"`
	Status         string `gorm:"size:30;not null" json:"status"`
}

func (Signatory) TableName() string { return "signatories" }

type DocumentTemplate struct {
	types.BaseModel
	Code            string `gorm:"size:80;not null" json:"code"`
	Name            string `gorm:"size:180;not null" json:"name"`
	Category        string `gorm:"size:50;not null" json:"category"`
	SubjectTemplate string `gorm:"size:255;not null" json:"subject_template"`
	BodyTemplate    string `gorm:"type:longtext;not null" json:"body_template"`
	NumberPattern   string `gorm:"size:180;not null" json:"number_pattern"`
	Version         int    `gorm:"not null" json:"version"`
	IsActive        bool   `gorm:"not null" json:"is_active"`
}

func (DocumentTemplate) TableName() string { return "document_templates" }

type LetterSequence struct {
	ID            string    `gorm:"type:char(36);primaryKey" json:"id"`
	TemplateCode  string    `gorm:"size:80;not null" json:"template_code"`
	SequenceYear  int       `gorm:"not null" json:"sequence_year"`
	SequenceMonth int       `gorm:"not null" json:"sequence_month"`
	LastNumber    int       `gorm:"not null" json:"last_number"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (LetterSequence) TableName() string { return "letter_sequences" }

type GenerationBatch struct {
	types.BaseModel
	PeriodID       *string    `gorm:"type:char(36)" json:"period_id"`
	RequestedBy    *string    `gorm:"type:char(36)" json:"requested_by"`
	Name           string     `gorm:"size:180;not null" json:"name"`
	Status         string     `gorm:"size:30;not null" json:"status"`
	RequestedCount int        `gorm:"not null" json:"requested_count"`
	GeneratedCount int        `gorm:"not null" json:"generated_count"`
	FailedCount    int        `gorm:"not null" json:"failed_count"`
	FiltersJSON    string     `gorm:"type:json" json:"filters_json"`
	ErrorSummary   string     `gorm:"type:text" json:"error_summary"`
	ArchiveName    string     `gorm:"size:255" json:"archive_name"`
	ArchivePath    string     `gorm:"size:500" json:"-"`
	ArchiveSize    int64      `gorm:"not null" json:"archive_size"`
	CompletedAt    *time.Time `json:"completed_at"`
}

func (GenerationBatch) TableName() string { return "document_generation_batches" }

type GeneratedDocument struct {
	types.BaseModel
	BatchID         *string   `gorm:"type:char(36)" json:"batch_id"`
	TemplateID      *string   `gorm:"type:char(36)" json:"template_id"`
	TemplateCode    string    `gorm:"size:80;not null" json:"template_code"`
	TemplateVersion int       `gorm:"not null" json:"template_version"`
	PlacementID     *string   `gorm:"type:char(36)" json:"placement_id"`
	StudentID       *string   `gorm:"type:char(36)" json:"student_id"`
	PeriodID        *string   `gorm:"type:char(36)" json:"period_id"`
	SignatoryID     *string   `gorm:"type:char(36)" json:"signatory_id"`
	GeneratedBy     *string   `gorm:"type:char(36)" json:"generated_by"`
	DocumentNumber  string    `gorm:"size:180" json:"document_number"`
	Title           string    `gorm:"size:255;not null" json:"title"`
	Format          string    `gorm:"size:20;not null" json:"format"`
	OriginalName    string    `gorm:"size:255;not null" json:"original_name"`
	StoredName      string    `gorm:"size:255;not null" json:"stored_name"`
	Path            string    `gorm:"size:500;not null" json:"-"`
	MimeType        string    `gorm:"size:150;not null" json:"mime_type"`
	Size            int64     `gorm:"not null" json:"size"`
	Status          string    `gorm:"size:30;not null" json:"status"`
	DataSnapshot    string    `gorm:"type:json;not null" json:"data_snapshot"`
	ChecksumSHA256  string    `gorm:"size:64;not null" json:"checksum_sha256"`
	GeneratedAt     time.Time `json:"generated_at"`
	StudentName     string    `gorm:"-" json:"student_name,omitempty"`
	PeriodName      string    `gorm:"-" json:"period_name,omitempty"`
}

func (GeneratedDocument) TableName() string { return "generated_documents" }
