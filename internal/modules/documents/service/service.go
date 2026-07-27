package service

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"

	"simpkl-api/internal/modules/documents/entity"
	"simpkl-api/internal/platform/storage"
	"simpkl-api/internal/shared/crud"
	apperrors "simpkl-api/internal/shared/errors"
	"simpkl-api/internal/shared/pagination"
)

const defaultMaxUpload = 10 << 20

type Service struct {
	db      *gorm.DB
	storage storage.Storage
	base    *crud.GormRepository[entity.Document]
}

func New(db *gorm.DB, fileStorage storage.Storage) *Service {
	return &Service{
		db: db, storage: fileStorage,
		base: crud.NewGormRepository[entity.Document](
			db,
			[]string{"number", "original_name", "notes"},
			map[string]string{
				"document_type_id": "document_type_id", "owner_type": "owner_type",
				"owner_id": "owner_id", "period_id": "period_id",
				"placement_id": "placement_id", "status": "status",
			},
		),
	}
}

func (s *Service) List(ctx context.Context, query pagination.Query, filters map[string]string) ([]entity.Document, pagination.Meta, error) {
	query.Normalize()
	items, total, err := s.base.List(ctx, query, filters)
	for index := range items {
		if items[index].ValidUntil != nil && items[index].ValidUntil.Before(time.Now()) && items[index].Status == "valid" {
			items[index].Status = "expired"
			_ = s.db.WithContext(ctx).Model(&items[index]).Update("status", "expired").Error
		}
	}
	return items, pagination.NewMeta(query, total), err
}

func (s *Service) Get(ctx context.Context, id string) (*entity.Document, error) {
	return s.base.Get(ctx, id)
}

func (s *Service) Upload(
	ctx context.Context,
	input *entity.Document,
	fileHeader *multipart.FileHeader,
) (*entity.Document, error) {
	if fileHeader == nil || fileHeader.Size <= 0 {
		return nil, invalid("FILE_REQUIRED", "File dokumen wajib dipilih")
	}
	var documentType entity.DocumentType
	if err := s.db.WithContext(ctx).First(&documentType, "id = ? AND status = ?", input.DocumentTypeID, "active").Error; err != nil {
		return nil, invalid("DOCUMENT_TYPE_NOT_FOUND", "Jenis dokumen tidak ditemukan atau tidak aktif")
	}
	maxSize := documentType.MaxSize
	if maxSize <= 0 {
		maxSize = defaultMaxUpload
	}
	if fileHeader.Size > maxSize {
		return nil, invalid("FILE_TOO_LARGE", "Ukuran file maksimal 10 MB")
	}
	source, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer source.Close()

	buffer := make([]byte, 512)
	read, err := source.Read(buffer)
	if err != nil && err != io.EOF {
		return nil, err
	}
	mimeType := http.DetectContentType(buffer[:read])
	if !allowedMIME(mimeType) || (documentType.AllowedMIME != "" && !strings.Contains(documentType.AllowedMIME, mimeType)) {
		return nil, invalid("UNSUPPORTED_FILE_TYPE", "File harus berupa PDF, JPG, JPEG, atau PNG")
	}
	if _, err := source.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	saved, err := s.storage.Save(ctx, source, fileHeader.Filename)
	if err != nil {
		return nil, err
	}
	input.OriginalName = filepath.Base(fileHeader.Filename)
	input.StoredName = saved.StoredName
	input.Path = saved.Path
	input.MimeType = mimeType
	input.Size = saved.Size
	var previous entity.Document
	previousErr := s.db.WithContext(ctx).
		Where("document_type_id = ? AND owner_type = ? AND owner_id = ? AND deleted_at IS NULL", input.DocumentTypeID, input.OwnerType, input.OwnerID).
		Order("version DESC").
		First(&previous).Error
	input.Version = 1
	if previousErr == nil {
		input.Version = previous.Version + 1
	}
	if input.Status == "" {
		input.Status = "pending"
	}
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(input).Error; err != nil {
			return err
		}
		if previousErr == nil {
			return tx.Model(&previous).Update("status", "superseded").Error
		}
		return nil
	}); err != nil {
		_ = s.storage.Delete(ctx, saved.Path)
		return nil, err
	}
	return input, nil
}

func (s *Service) Verify(ctx context.Context, id, status, verifierID, notes string) (*entity.Document, error) {
	if status != "valid" && status != "revision_required" && status != "rejected" {
		return nil, invalid("INVALID_DOCUMENT_STATUS", "Status verifikasi dokumen tidak valid")
	}
	now := time.Now()
	if err := s.db.WithContext(ctx).Model(&entity.Document{}).Where("id = ?", id).Updates(map[string]any{
		"status": status, "verified_by": verifierID, "verified_at": now, "notes": notes,
	}).Error; err != nil {
		return nil, err
	}
	return s.base.Get(ctx, id)
}

func (s *Service) Open(ctx context.Context, id string) (*entity.Document, io.ReadCloser, error) {
	document, err := s.base.Get(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	file, err := s.storage.Open(ctx, document.Path)
	return document, file, err
}

func (s *Service) Delete(ctx context.Context, id string) error {
	document, err := s.base.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := s.db.WithContext(ctx).Delete(document).Error; err != nil {
		return err
	}
	return s.storage.Delete(ctx, document.Path)
}

func allowedMIME(mimeType string) bool {
	return mimeType == "application/pdf" || mimeType == "image/jpeg" || mimeType == "image/png"
}

func invalid(code, message string) error {
	return &apperrors.AppError{Status: http.StatusUnprocessableEntity, Code: code, Message: message, Cause: fmt.Errorf("%s", code)}
}

var _ = strings.TrimSpace
