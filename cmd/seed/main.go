package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"gorm.io/gorm"

	"simpkl-api/internal/app"
	roleentity "simpkl-api/internal/modules/roles/entity"
	userentity "simpkl-api/internal/modules/users/entity"
	platformauth "simpkl-api/internal/platform/auth"
)

func main() {
	name := required("SEED_ADMIN_NAME")
	email := strings.ToLower(required("SEED_ADMIN_EMAIL"))
	username := strings.ToLower(required("SEED_ADMIN_USERNAME"))
	password := required("SEED_ADMIN_PASSWORD")
	if len(password) < 12 {
		log.Fatal("SEED_ADMIN_PASSWORD minimal 12 karakter")
	}

	ctx := context.Background()
	dependencies, err := app.BuildDependencies(ctx)
	if err != nil {
		log.Fatalf("initialize dependencies: %v", err)
	}
	defer dependencies.Close()

	hash, err := platformauth.HashPassword(password)
	if err != nil {
		log.Fatalf("hash password: %v", err)
	}
	err = dependencies.Database.GORM.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var role roleentity.Role
		if err := tx.Where("code = ?", "super_admin").First(&role).Error; err != nil {
			return fmt.Errorf("super_admin role is missing; run migrations first: %w", err)
		}
		var user userentity.User
		err := tx.Where("email = ? OR username = ?", email, username).First(&user).Error
		if err == nil {
			return fmt.Errorf("admin user already exists")
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}
		user = userentity.User{
			Name: name, Email: email, Username: username,
			PasswordHash: hash, Status: "active",
		}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		return tx.Create(&roleentity.UserRole{UserID: user.ID, RoleID: role.ID}).Error
	})
	if err != nil {
		log.Fatalf("seed admin: %v", err)
	}
	log.Printf("super admin %s berhasil dibuat", email)
}

func required(key string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		log.Fatalf("%s wajib diisi", key)
	}
	return value
}
