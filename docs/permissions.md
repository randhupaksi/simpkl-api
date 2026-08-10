# Permission

Permission selalu diperiksa di backend. Middleware membaca user ID dari access token, lalu memeriksa relasi role-permission pada database.

Kelompok permission:

- `period.*`, `major.*`, `class.*`, `student.*`
- `company.*`, `supervisor.*`, `placement.*`
- `document.*`, `document_type.*`, `readiness.*`
- `automation.view`, `automation.generate`, `automation.download`, `automation.manage`
- `report.view`, `archive.view`, `audit.view`
- `user.*`, `role.*`, `permission.*`

Permission `*` hanya diberikan kepada `super_admin`. Pembatasan data berdasarkan jurusan, kelas, atau siswa bimbingan dapat ditambahkan sebagai data scope di service saat akun staf mulai digunakan.
