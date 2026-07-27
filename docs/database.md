# Database

Schema dikelola oleh SQL migration, bukan AutoMigrate produksi.

Relasi utama:

```text
periods -> placements <- students
companies -> company_contacts
companies -> placements <- supervisors
placements -> documents
students + periods -> administrative_readiness
periods -> archives
users -> user_roles -> roles -> role_permissions -> permissions
users -> refresh_sessions
users -> audit_logs
```

Penempatan lama tidak dihapus ketika siswa pindah. Statusnya menjadi `transferred`, lalu penempatan baru menyimpan `previous_placement_id`. Soft delete digunakan untuk data master dan operasional; audit log menyimpan ringkasan perubahan.
