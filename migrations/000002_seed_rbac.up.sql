INSERT INTO permissions (id, code, name, module, created_at, updated_at) VALUES
(UUID(), '*', 'Akses penuh sistem', 'system', NOW(3), NOW(3)),
(UUID(), 'period.view', 'Lihat periode', 'period', NOW(3), NOW(3)),
(UUID(), 'period.create', 'Buat periode', 'period', NOW(3), NOW(3)),
(UUID(), 'period.update', 'Ubah periode', 'period', NOW(3), NOW(3)),
(UUID(), 'period.delete', 'Nonaktifkan periode', 'period', NOW(3), NOW(3)),
(UUID(), 'period.archive', 'Arsipkan periode', 'period', NOW(3), NOW(3)),
(UUID(), 'major.view', 'Lihat jurusan', 'major', NOW(3), NOW(3)),
(UUID(), 'major.create', 'Buat jurusan', 'major', NOW(3), NOW(3)),
(UUID(), 'major.update', 'Ubah jurusan', 'major', NOW(3), NOW(3)),
(UUID(), 'major.delete', 'Nonaktifkan jurusan', 'major', NOW(3), NOW(3)),
(UUID(), 'class.view', 'Lihat kelas', 'class', NOW(3), NOW(3)),
(UUID(), 'class.create', 'Buat kelas', 'class', NOW(3), NOW(3)),
(UUID(), 'class.update', 'Ubah kelas', 'class', NOW(3), NOW(3)),
(UUID(), 'class.delete', 'Nonaktifkan kelas', 'class', NOW(3), NOW(3)),
(UUID(), 'student.view', 'Lihat siswa', 'student', NOW(3), NOW(3)),
(UUID(), 'student.create', 'Buat siswa', 'student', NOW(3), NOW(3)),
(UUID(), 'student.update', 'Ubah siswa', 'student', NOW(3), NOW(3)),
(UUID(), 'student.delete', 'Nonaktifkan siswa', 'student', NOW(3), NOW(3)),
(UUID(), 'student.import', 'Impor siswa dari Excel', 'student', NOW(3), NOW(3)),
(UUID(), 'company.view', 'Lihat perusahaan', 'company', NOW(3), NOW(3)),
(UUID(), 'company.create', 'Buat perusahaan/PIC', 'company', NOW(3), NOW(3)),
(UUID(), 'company.update', 'Ubah perusahaan/PIC', 'company', NOW(3), NOW(3)),
(UUID(), 'company.delete', 'Nonaktifkan perusahaan/PIC', 'company', NOW(3), NOW(3)),
(UUID(), 'supervisor.view', 'Lihat pembimbing', 'supervisor', NOW(3), NOW(3)),
(UUID(), 'supervisor.create', 'Buat pembimbing', 'supervisor', NOW(3), NOW(3)),
(UUID(), 'supervisor.update', 'Ubah pembimbing', 'supervisor', NOW(3), NOW(3)),
(UUID(), 'supervisor.delete', 'Nonaktifkan pembimbing', 'supervisor', NOW(3), NOW(3)),
(UUID(), 'placement.view', 'Lihat penempatan', 'placement', NOW(3), NOW(3)),
(UUID(), 'placement.create', 'Buat penempatan', 'placement', NOW(3), NOW(3)),
(UUID(), 'placement.update', 'Ubah/validasi penempatan', 'placement', NOW(3), NOW(3)),
(UUID(), 'placement.delete', 'Nonaktifkan penempatan', 'placement', NOW(3), NOW(3)),
(UUID(), 'placement.transfer', 'Pindahkan siswa', 'placement', NOW(3), NOW(3)),
(UUID(), 'document.view', 'Lihat dokumen', 'document', NOW(3), NOW(3)),
(UUID(), 'document.upload', 'Unggah dokumen', 'document', NOW(3), NOW(3)),
(UUID(), 'document.verify', 'Verifikasi dokumen', 'document', NOW(3), NOW(3)),
(UUID(), 'document.download', 'Unduh dokumen', 'document', NOW(3), NOW(3)),
(UUID(), 'document.delete', 'Hapus dokumen', 'document', NOW(3), NOW(3)),
(UUID(), 'document_type.view', 'Lihat jenis dokumen', 'document', NOW(3), NOW(3)),
(UUID(), 'document_type.create', 'Buat jenis dokumen', 'document', NOW(3), NOW(3)),
(UUID(), 'document_type.update', 'Ubah jenis dokumen', 'document', NOW(3), NOW(3)),
(UUID(), 'document_type.delete', 'Nonaktifkan jenis dokumen', 'document', NOW(3), NOW(3)),
(UUID(), 'readiness.view', 'Lihat kesiapan', 'readiness', NOW(3), NOW(3)),
(UUID(), 'readiness.update', 'Hitung kesiapan', 'readiness', NOW(3), NOW(3)),
(UUID(), 'readiness.override', 'Pengecualian kesiapan', 'readiness', NOW(3), NOW(3)),
(UUID(), 'report.view', 'Lihat dan ekspor laporan', 'report', NOW(3), NOW(3)),
(UUID(), 'archive.view', 'Lihat arsip', 'archive', NOW(3), NOW(3)),
(UUID(), 'audit.view', 'Lihat audit log', 'audit', NOW(3), NOW(3)),
(UUID(), 'user.view', 'Lihat pengguna', 'user', NOW(3), NOW(3)),
(UUID(), 'user.create', 'Buat pengguna', 'user', NOW(3), NOW(3)),
(UUID(), 'user.update', 'Ubah pengguna', 'user', NOW(3), NOW(3)),
(UUID(), 'user.delete', 'Nonaktifkan pengguna', 'user', NOW(3), NOW(3)),
(UUID(), 'user.manage', 'Kelola role pengguna', 'user', NOW(3), NOW(3)),
(UUID(), 'role.view', 'Lihat role', 'role', NOW(3), NOW(3)),
(UUID(), 'role.create', 'Buat role', 'role', NOW(3), NOW(3)),
(UUID(), 'role.update', 'Ubah role', 'role', NOW(3), NOW(3)),
(UUID(), 'role.delete', 'Nonaktifkan role', 'role', NOW(3), NOW(3)),
(UUID(), 'role.manage', 'Kelola permission role', 'role', NOW(3), NOW(3)),
(UUID(), 'permission.view', 'Lihat permission', 'permission', NOW(3), NOW(3)),
(UUID(), 'permission.create', 'Buat permission', 'permission', NOW(3), NOW(3)),
(UUID(), 'permission.update', 'Ubah permission', 'permission', NOW(3), NOW(3)),
(UUID(), 'permission.delete', 'Hapus permission', 'permission', NOW(3), NOW(3));

INSERT INTO roles (id, code, name, description, is_system, status, created_at, updated_at) VALUES
(UUID(), 'super_admin', 'Super Admin', 'Akses penuh terhadap seluruh sistem', TRUE, 'active', NOW(3), NOW(3)),
(UUID(), 'admin_pkl', 'Admin PKL', 'Pengelola operasional PKL', TRUE, 'active', NOW(3), NOW(3)),
(UUID(), 'coordinator_pkl', 'Koordinator PKL', 'Pengawas dan pemberi persetujuan', TRUE, 'active', NOW(3), NOW(3)),
(UUID(), 'program_head', 'Kepala Program', 'Pemantauan berdasarkan jurusan', TRUE, 'active', NOW(3), NOW(3)),
(UUID(), 'homeroom_teacher', 'Wali Kelas', 'Pemantauan berdasarkan kelas', TRUE, 'active', NOW(3), NOW(3)),
(UUID(), 'supervisor_teacher', 'Guru Pembimbing', 'Pemantauan siswa bimbingan', TRUE, 'active', NOW(3), NOW(3)),
(UUID(), 'leadership', 'Pimpinan', 'Akses dashboard dan laporan', TRUE, 'active', NOW(3), NOW(3));

INSERT INTO role_permissions (role_id, permission_id)
SELECT roles.id, permissions.id FROM roles JOIN permissions
WHERE roles.code = 'super_admin' AND permissions.code = '*';

INSERT INTO role_permissions (role_id, permission_id)
SELECT roles.id, permissions.id FROM roles JOIN permissions
WHERE roles.code = 'admin_pkl'
  AND permissions.code <> '*'
  AND permissions.module NOT IN ('role', 'permission');

INSERT INTO role_permissions (role_id, permission_id)
SELECT roles.id, permissions.id FROM roles JOIN permissions
WHERE roles.code = 'coordinator_pkl'
  AND (
    permissions.code LIKE '%.view'
    OR permissions.code IN (
      'period.update','period.archive','company.update','placement.update',
      'placement.transfer','document.verify','readiness.update','readiness.override'
    )
  );

INSERT INTO role_permissions (role_id, permission_id)
SELECT roles.id, permissions.id FROM roles JOIN permissions
WHERE roles.code = 'program_head'
  AND permissions.code IN (
    'period.view','major.view','class.view','student.view','company.view',
    'supervisor.view','placement.view','document.view','readiness.view',
    'report.view','archive.view'
  );

INSERT INTO role_permissions (role_id, permission_id)
SELECT roles.id, permissions.id FROM roles JOIN permissions
WHERE roles.code = 'homeroom_teacher'
  AND permissions.code IN (
    'period.view','class.view','student.view','company.view',
    'placement.view','document.view','readiness.view','report.view'
  );

INSERT INTO role_permissions (role_id, permission_id)
SELECT roles.id, permissions.id FROM roles JOIN permissions
WHERE roles.code = 'supervisor_teacher'
  AND permissions.code IN (
    'period.view','student.view','company.view','supervisor.view',
    'placement.view','document.view','document.download','readiness.view'
  );

INSERT INTO role_permissions (role_id, permission_id)
SELECT roles.id, permissions.id FROM roles JOIN permissions
WHERE roles.code = 'leadership'
  AND permissions.code IN ('period.view','report.view','archive.view');

INSERT INTO document_types (id, code, name, category, required, has_expiry, max_size, allowed_mime, status, created_at, updated_at) VALUES
(UUID(), 'acceptance_letter', 'Surat Penerimaan Perusahaan', 'company', TRUE, FALSE, 10485760, 'application/pdf,image/jpeg,image/png', 'active', NOW(3), NOW(3)),
(UUID(), 'parent_permission', 'Surat Izin Orang Tua', 'student', TRUE, FALSE, 10485760, 'application/pdf,image/jpeg,image/png', 'active', NOW(3), NOW(3)),
(UUID(), 'introduction_letter', 'Surat Pengantar Sekolah', 'school', TRUE, FALSE, 10485760, 'application/pdf', 'active', NOW(3), NOW(3)),
(UUID(), 'mou', 'MoU/Perjanjian Kerja Sama', 'company', FALSE, TRUE, 10485760, 'application/pdf', 'active', NOW(3), NOW(3)),
(UUID(), 'supervisor_assignment', 'Surat Tugas Guru Pembimbing', 'school', FALSE, FALSE, 10485760, 'application/pdf', 'active', NOW(3), NOW(3));
