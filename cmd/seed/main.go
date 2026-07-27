package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
	"gorm.io/gorm"

	"simpkl-api/internal/app"
	roleentity "simpkl-api/internal/modules/roles/entity"
	userentity "simpkl-api/internal/modules/users/entity"
	platformauth "simpkl-api/internal/platform/auth"
)

func main() {
	settings := loadSettings()
	name := required(settings, "SEED_ADMIN_NAME")
	email := strings.ToLower(required(settings, "SEED_ADMIN_EMAIL"))
	username := strings.ToLower(required(settings, "SEED_ADMIN_USERNAME"))
	password := required(settings, "SEED_ADMIN_PASSWORD")
	if len(password) < 8 {
		log.Fatal("SEED_ADMIN_PASSWORD minimal 8 karakter")
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
		if err := tx.Omit("MajorID", "ClassID").Create(&user).Error; err != nil {
			return err
		}
		return tx.Create(&roleentity.UserRole{UserID: user.ID, RoleID: role.ID}).Error
	})
	if err != nil {
		log.Fatalf("seed admin: %v", err)
	}
	log.Printf("super admin %s berhasil dibuat", email)
}

func loadSettings() *viper.Viper {
	settings := viper.New()
	settings.SetConfigName(".env")
	settings.SetConfigType("env")
	settings.AddConfigPath(".")
	settings.AutomaticEnv()

	if err := settings.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			log.Fatalf("read environment file: %v", err)
		}
	}

	return settings
}

func required(settings *viper.Viper, key string) string {
	value := strings.TrimSpace(settings.GetString(key))
	if value == "" {
		log.Fatalf("%s wajib diisi", key)
	}
	return value
}
