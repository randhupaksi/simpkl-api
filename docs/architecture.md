# Arsitektur

SIMPKL API memakai modular monolith. `internal/app` menjadi composition root; setiap domain berada di `internal/modules`; kebutuhan teknis berada di `internal/platform`; kontrak lintas domain berada di `internal/shared`.

Alur request standar:

```text
Gin route -> auth/permission middleware -> handler -> service -> repository -> MySQL
```

Handler hanya mengurus HTTP dan serialisasi. Service menangani validasi serta aturan bisnis. Repository menangani query. Dependensi dibuat melalui constructor. CRUD standar menggunakan generic repository/service/handler, sedangkan auth, perpindahan, dokumen, readiness, laporan, dan arsip memiliki service khusus.

Dokumen tidak pernah disajikan sebagai URL publik. Storage adapter memvalidasi path dan endpoint download selalu melalui autentikasi serta permission.

Database migrations are embedded into the API binary and run before the HTTP
server starts when `MIGRATIONS_AUTO_APPLY=true`. The runner records the current
version in `schema_migrations`, acquires a MySQL advisory lock, applies only
new `.up.sql` files in numeric order, and fails closed when a dirty migration
is detected. Teams that use a dedicated deployment migration job can disable
the startup runner and keep using the existing CLI.

## Otomasi dokumen

Modul `documentautomation` mengorkestrasi data lintas domain tanpa memindahkan
kepemilikan data master. Service mengambil snapshot penempatan, siswa, kelas,
jurusan, perusahaan, PIC, pembimbing, periode, profil institusi, penandatangan,
dan versi template aktif. Generator di `internal/platform/documentgen`
menghasilkan OOXML/DOCX dan PDF A4; generator laporan Excel menghasilkan rekap
XLSX; standard library membentuk paket ZIP.

Riwayat lama tetap mereferensikan versi template dan snapshot saat generate
sehingga perubahan data master tidak mengubah dokumen final yang sudah dibuat.
