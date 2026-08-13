package main

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/spf13/viper"

	"simpkl-api/internal/app"
	seed "simpkl-api/internal/seed"
)

func main() {
	settings := loadSettings()
	if !settings.GetBool("SEED_ENABLED") {
		log.Fatal("SEED_ENABLED=false; set SEED_ENABLED=true untuk menjalankan seeder")
	}
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

	err = seed.Run(ctx, dependencies.Database.GORM, seed.Options{
		RecordCount: settings.GetInt("SEED_RECORD_COUNT"), ResetLegacy: settings.GetBool("SEED_RESET_LEGACY"), AdminName: name,
		AdminEmail: email, AdminUsername: username, AdminPassword: password,
	})
	if err != nil {
		log.Fatalf("seed data: %v", err)
	}
	log.Printf("seeder berhasil: dataset PKL 2026/2027 siap, admin %s siap digunakan", email)
}

func loadSettings() *viper.Viper {
	settings := viper.New()
	settings.SetConfigName(".env")
	settings.SetConfigType("env")
	settings.AddConfigPath(".")
	settings.AutomaticEnv()
	settings.SetDefault("SEED_ENABLED", true)
	settings.SetDefault("SEED_RECORD_COUNT", 5)
	settings.SetDefault("SEED_RESET_LEGACY", false)

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
