# SIMPKL Citra Negara API

Backend internal untuk manajemen administrasi PKL SMK Citra Negara. Ruang lingkupnya mencakup periode, siswa, kelas, jurusan, perusahaan/PIC, penempatan, guru pembimbing, dokumen privat, kesiapan administrasi, laporan, arsip, pengguna, RBAC, dan audit log. API ini tidak memuat absensi, jurnal harian, GPS, atau portal operasional perusahaan.

## Stack

Go, Gin, GORM, MySQL, JWT, Viper, Zap, Excelize, Swagger UI, dan SQL migration.

## Menjalankan lokal

1. Salin `.env.example` menjadi `.env` dan ganti seluruh secret.
2. Jalankan MySQL: `docker compose up -d mysql`.
3. Set `DATABASE_URL`, contoh:
   `mysql://simpkl:simpkl_dev_password@tcp(localhost:3306)/simpkl?multiStatements=true`
4. Jalankan migration:
   `go run github.com/golang-migrate/migrate/v4/cmd/migrate -path migrations -database "$env:DATABASE_URL" up`
5. Pastikan migration selesai, `SEED_ENABLED=true`, dan isi variabel `SEED_ADMIN_*`, lalu jalankan `go run ./cmd/seed`.
6. Jalankan API: `go run ./cmd/api`.

Health check tersedia di `GET /health`, API menggunakan prefix `/api/v1`, dan dokumentasi interaktif tersedia di `/swagger/index.html`.

## Validasi

```text
gofmt -w ./cmd ./internal
go vet ./...
go test ./...
go build -o ./bin/simpkl-api.exe ./cmd/api
```

## Struktur

- `cmd/api`: entrypoint HTTP.
- `cmd/seed`: bootstrap super admin dan fixture sintetis seluruh domain untuk PKL SMK Citra Negara 2026/2027 tanpa kredensial hardcoded. Dataset mencakup PPLG, TJKT, Pemasaran, MPLB, DKV, mitra dummy sekitar Depok, penempatan aktif, dokumen, readiness, arsip, dan audit log. Seeder idempotent dengan jumlah data yang mengikuti kebutuhan masing-masing konteks.
- `internal/app`: composition root dan lifecycle server.
- `internal/modules`: domain aplikasi.
- `internal/shared/crud`: pola repository–service–handler yang digunakan ulang.
- `internal/platform`: database, JWT, logging, storage, Excel, dan PDF.
- `migrations`: schema dan seed referensi.
- `docs`: arsitektur, database, permission, konvensi API, dan OpenAPI.

File dokumen disimpan di storage privat dan hanya dapat diunduh melalui endpoint yang telah melewati autentikasi serta permission.
