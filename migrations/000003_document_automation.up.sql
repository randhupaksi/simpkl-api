CREATE TABLE school_profiles (
    id CHAR(36) PRIMARY KEY,
    institution_name VARCHAR(180) NOT NULL,
    institution_type VARCHAR(80) NOT NULL DEFAULT 'Sekolah Menengah Kejuruan',
    npsn VARCHAR(30) NULL,
    address TEXT NOT NULL,
    village VARCHAR(100) NULL,
    district VARCHAR(100) NULL,
    city VARCHAR(100) NULL,
    province VARCHAR(100) NULL,
    postal_code VARCHAR(15) NULL,
    phone VARCHAR(30) NULL,
    email VARCHAR(150) NULL,
    website VARCHAR(255) NULL,
    letterhead_tagline VARCHAR(255) NULL,
    logo_path VARCHAR(500) NULL,
    timezone VARCHAR(50) NOT NULL DEFAULT 'Asia/Jakarta',
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    deleted_at DATETIME(3) NULL,
    INDEX idx_school_profiles_deleted_at (deleted_at)
);

CREATE TABLE signatories (
    id CHAR(36) PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    title VARCHAR(150) NOT NULL,
    employee_number VARCHAR(60) NULL,
    role_code VARCHAR(60) NOT NULL DEFAULT 'principal',
    signature_path VARCHAR(500) NULL,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(30) NOT NULL DEFAULT 'active',
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    deleted_at DATETIME(3) NULL,
    INDEX idx_signatories_status (status),
    INDEX idx_signatories_deleted_at (deleted_at)
);

CREATE TABLE document_templates (
    id CHAR(36) PRIMARY KEY,
    code VARCHAR(80) NOT NULL,
    name VARCHAR(180) NOT NULL,
    category VARCHAR(50) NOT NULL DEFAULT 'letter',
    subject_template VARCHAR(255) NOT NULL,
    body_template LONGTEXT NOT NULL,
    number_pattern VARCHAR(180) NOT NULL DEFAULT '{{sequence}}/{{code}}/{{month_roman}}/{{year}}',
    version INT NOT NULL DEFAULT 1,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    deleted_at DATETIME(3) NULL,
    UNIQUE INDEX uq_document_templates_code_version (code, version),
    INDEX idx_document_templates_active (is_active),
    INDEX idx_document_templates_deleted_at (deleted_at)
);

CREATE TABLE letter_sequences (
    id CHAR(36) PRIMARY KEY,
    template_code VARCHAR(80) NOT NULL,
    sequence_year INT NOT NULL,
    sequence_month INT NOT NULL,
    last_number INT NOT NULL DEFAULT 0,
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    UNIQUE INDEX uq_letter_sequences_scope (template_code, sequence_year, sequence_month)
);

CREATE TABLE document_generation_batches (
    id CHAR(36) PRIMARY KEY,
    period_id CHAR(36) NULL,
    requested_by CHAR(36) NULL,
    name VARCHAR(180) NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'processing',
    requested_count INT NOT NULL DEFAULT 0,
    generated_count INT NOT NULL DEFAULT 0,
    failed_count INT NOT NULL DEFAULT 0,
    filters_json JSON NULL,
    error_summary TEXT NULL,
    archive_name VARCHAR(255) NULL,
    archive_path VARCHAR(500) NULL,
    archive_size BIGINT NOT NULL DEFAULT 0,
    completed_at DATETIME(3) NULL,
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    deleted_at DATETIME(3) NULL,
    CONSTRAINT fk_generation_batch_period FOREIGN KEY (period_id) REFERENCES periods(id) ON DELETE SET NULL,
    CONSTRAINT fk_generation_batch_user FOREIGN KEY (requested_by) REFERENCES users(id) ON DELETE SET NULL,
    INDEX idx_generation_batches_status (status),
    INDEX idx_generation_batches_period (period_id),
    INDEX idx_generation_batches_deleted_at (deleted_at)
);

CREATE TABLE generated_documents (
    id CHAR(36) PRIMARY KEY,
    batch_id CHAR(36) NULL,
    template_id CHAR(36) NULL,
    template_code VARCHAR(80) NOT NULL,
    template_version INT NOT NULL DEFAULT 1,
    placement_id CHAR(36) NULL,
    student_id CHAR(36) NULL,
    period_id CHAR(36) NULL,
    signatory_id CHAR(36) NULL,
    generated_by CHAR(36) NULL,
    document_number VARCHAR(180) NULL,
    title VARCHAR(255) NOT NULL,
    format VARCHAR(20) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    stored_name VARCHAR(255) NOT NULL,
    path VARCHAR(500) NOT NULL,
    mime_type VARCHAR(150) NOT NULL,
    size BIGINT NOT NULL DEFAULT 0,
    status VARCHAR(30) NOT NULL DEFAULT 'final',
    data_snapshot JSON NOT NULL,
    checksum_sha256 CHAR(64) NOT NULL,
    generated_at DATETIME(3) NOT NULL,
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    deleted_at DATETIME(3) NULL,
    CONSTRAINT fk_generated_document_batch FOREIGN KEY (batch_id) REFERENCES document_generation_batches(id) ON DELETE SET NULL,
    CONSTRAINT fk_generated_document_template FOREIGN KEY (template_id) REFERENCES document_templates(id) ON DELETE SET NULL,
    CONSTRAINT fk_generated_document_placement FOREIGN KEY (placement_id) REFERENCES placements(id) ON DELETE SET NULL,
    CONSTRAINT fk_generated_document_student FOREIGN KEY (student_id) REFERENCES students(id) ON DELETE SET NULL,
    CONSTRAINT fk_generated_document_period FOREIGN KEY (period_id) REFERENCES periods(id) ON DELETE SET NULL,
    CONSTRAINT fk_generated_document_signatory FOREIGN KEY (signatory_id) REFERENCES signatories(id) ON DELETE SET NULL,
    CONSTRAINT fk_generated_document_user FOREIGN KEY (generated_by) REFERENCES users(id) ON DELETE SET NULL,
    INDEX idx_generated_documents_batch (batch_id),
    INDEX idx_generated_documents_student (student_id),
    INDEX idx_generated_documents_period (period_id),
    INDEX idx_generated_documents_code (template_code),
    INDEX idx_generated_documents_generated_at (generated_at),
    INDEX idx_generated_documents_deleted_at (deleted_at)
);

INSERT INTO permissions (id, code, name, module, created_at, updated_at) VALUES
(UUID(), 'automation.view', 'Lihat pusat otomasi dokumen', 'automation', NOW(3), NOW(3)),
(UUID(), 'automation.generate', 'Buat dokumen otomatis', 'automation', NOW(3), NOW(3)),
(UUID(), 'automation.download', 'Unduh dokumen otomatis', 'automation', NOW(3), NOW(3)),
(UUID(), 'automation.manage', 'Kelola profil, penandatangan, dan template', 'automation', NOW(3), NOW(3));

INSERT INTO role_permissions (role_id, permission_id)
SELECT roles.id, permissions.id FROM roles JOIN permissions
WHERE roles.code = 'admin_pkl' AND permissions.module = 'automation';

INSERT INTO role_permissions (role_id, permission_id)
SELECT roles.id, permissions.id FROM roles JOIN permissions
WHERE roles.code = 'coordinator_pkl' AND permissions.code IN ('automation.view','automation.generate','automation.download','automation.manage');

INSERT INTO role_permissions (role_id, permission_id)
SELECT roles.id, permissions.id FROM roles JOIN permissions
WHERE roles.code IN ('program_head','homeroom_teacher','supervisor_teacher')
  AND permissions.code IN ('automation.view','automation.generate','automation.download');

INSERT INTO role_permissions (role_id, permission_id)
SELECT roles.id, permissions.id FROM roles JOIN permissions
WHERE roles.code = 'leadership' AND permissions.code IN ('automation.view','automation.download');

INSERT INTO school_profiles (
    id, institution_name, institution_type, address, city, province, letterhead_tagline,
    timezone, created_at, updated_at
) VALUES (
    UUID(), 'Nama Institusi', 'Sekolah Menengah Kejuruan',
    'Lengkapi alamat institusi pada menu Pengaturan Dokumen', '', '',
    'Praktik Kerja Lapangan', 'Asia/Jakarta', NOW(3), NOW(3)
);

INSERT INTO signatories (
    id, name, title, role_code, is_default, status, created_at, updated_at
) VALUES (
    UUID(), 'Nama Kepala Sekolah', 'Kepala Sekolah', 'principal', TRUE, 'active', NOW(3), NOW(3)
);

INSERT INTO document_templates (
    id, code, name, category, subject_template, body_template, number_pattern,
    version, is_active, created_at, updated_at
) VALUES
(UUID(), 'introduction_letter', 'Surat Pengantar PKL', 'letter', 'Permohonan Praktik Kerja Lapangan',
'Dengan hormat,\n\nDalam rangka pelaksanaan program Praktik Kerja Lapangan (PKL) tahun ajaran {{academic_year}}, kami memohon kesediaan {{company_name}} untuk menerima peserta didik berikut:\n\nNama: {{student_name}}\nNIS: {{student_nis}}\nKelas: {{class_name}}\nProgram Keahlian: {{major_name}}\nPeriode: {{placement_start}} sampai {{placement_end}}\n\nKami berharap peserta didik tersebut memperoleh kesempatan belajar dan pengalaman kerja sesuai bidang keahliannya. Atas perhatian dan kerja sama yang baik, kami sampaikan terima kasih.',
'{{sequence}}/PKL/{{month_roman}}/{{year}}', 1, TRUE, NOW(3), NOW(3)),
(UUID(), 'placement_letter', 'Surat Keterangan Penempatan', 'letter', 'Keterangan Penempatan Praktik Kerja Lapangan',
'Yang bertanda tangan di bawah ini menerangkan bahwa:\n\nNama: {{student_name}}\nNIS: {{student_nis}}\nKelas: {{class_name}}\nProgram Keahlian: {{major_name}}\n\ntelah ditempatkan untuk melaksanakan Praktik Kerja Lapangan di {{company_name}}, {{company_address}}, pada bagian {{placement_division}} sebagai {{placement_position}} selama {{placement_start}} sampai {{placement_end}}.\n\nSurat keterangan ini dibuat untuk dipergunakan sebagaimana mestinya.',
'{{sequence}}/KET-PKL/{{month_roman}}/{{year}}', 1, TRUE, NOW(3), NOW(3)),
(UUID(), 'supervisor_assignment', 'Surat Tugas Guru Pembimbing', 'letter', 'Penugasan Guru Pembimbing PKL',
'Dengan ini menugaskan:\n\nNama: {{supervisor_name}}\nNomor Pegawai: {{supervisor_employee_number}}\nJabatan: {{supervisor_position}}\n\nuntuk melaksanakan pembimbingan, pemantauan, dan evaluasi Praktik Kerja Lapangan bagi peserta didik {{student_name}} dari kelas {{class_name}} yang ditempatkan di {{company_name}} selama {{placement_start}} sampai {{placement_end}}.\n\nDemikian surat tugas ini diberikan untuk dilaksanakan dengan penuh tanggung jawab.',
'{{sequence}}/ST-PKL/{{month_roman}}/{{year}}', 1, TRUE, NOW(3), NOW(3)),
(UUID(), 'parent_consent', 'Surat Persetujuan Orang Tua', 'letter', 'Persetujuan Mengikuti Praktik Kerja Lapangan',
'Saya yang bertanda tangan di bawah ini:\n\nNama Orang Tua/Wali: {{parent_name}}\nOrang tua/wali dari: {{student_name}}\nKelas: {{class_name}}\nProgram Keahlian: {{major_name}}\n\ndengan ini menyetujui peserta didik tersebut mengikuti Praktik Kerja Lapangan di {{company_name}} pada {{placement_start}} sampai {{placement_end}}. Kami memahami dan bersedia mendukung ketentuan pelaksanaan PKL yang berlaku.\n\nDemikian surat persetujuan ini dibuat dengan sebenarnya.',
'{{sequence}}/IZIN-PKL/{{month_roman}}/{{year}}', 1, TRUE, NOW(3), NOW(3)),
(UUID(), 'placement_recap', 'Rekap Penempatan PKL', 'spreadsheet', 'Rekap Penempatan Praktik Kerja Lapangan',
'Rekap penempatan peserta didik berdasarkan periode dan filter yang dipilih.',
'REKAP-PKL-{{year}}', 1, TRUE, NOW(3), NOW(3));
