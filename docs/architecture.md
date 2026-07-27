# Arsitektur

SIMPKL API memakai modular monolith. `internal/app` menjadi composition root; setiap domain berada di `internal/modules`; kebutuhan teknis berada di `internal/platform`; kontrak lintas domain berada di `internal/shared`.

Alur request standar:

```text
Gin route -> auth/permission middleware -> handler -> service -> repository -> MySQL
```

Handler hanya mengurus HTTP dan serialisasi. Service menangani validasi serta aturan bisnis. Repository menangani query. Dependensi dibuat melalui constructor. CRUD standar menggunakan generic repository/service/handler, sedangkan auth, perpindahan, dokumen, readiness, laporan, dan arsip memiliki service khusus.

Dokumen tidak pernah disajikan sebagai URL publik. Storage adapter memvalidasi path dan endpoint download selalu melalui autentikasi serta permission.
