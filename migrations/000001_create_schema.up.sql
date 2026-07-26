CREATE TABLE permissions (
    id CHAR(36) PRIMARY KEY,
    code VARCHAR(120) NOT NULL UNIQUE,
    name VARCHAR(150) NOT NULL,
    module VARCHAR(80) NOT NULL,
    description TEXT NULL,
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    deleted_at DATETIME(3) NULL,
    INDEX idx_permissions_module (module),
    INDEX idx_permissions_deleted_at (deleted_at)
);

CREATE TABLE roles (
    id CHAR(36) PRIMARY KEY,
    code VARCHAR(80) NOT NULL UNIQUE,
    name VARCHAR(120) NOT NULL,
    description TEXT NULL,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(30) NOT NULL DEFAULT 'active',
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    deleted_at DATETIME(3) NULL,
    INDEX idx_roles_status (status),
    INDEX idx_roles_deleted_at (deleted_at)
);

CREATE TABLE role_permissions (
    role_id CHAR(36) NOT NULL,
    permission_id CHAR(36) NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    CONSTRAINT fk_role_permissions_role FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    CONSTRAINT fk_role_permissions_permission FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);

CREATE TABLE majors (
    id CHAR(36) PRIMARY KEY,
    code VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(150) NOT NULL,
    abbreviation VARCHAR(30) NOT NULL,
    head_name VARCHAR(150) NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'active',
    description TEXT NULL,
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    deleted_at DATETIME(3) NULL,
    INDEX idx_majors_status (status),
    INDEX idx_majors_deleted_at (deleted_at)
);

CREATE TABLE classes (
    id CHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    level INT NOT NULL,
    major_id CHAR(36) NOT NULL,
    homeroom_teacher VARCHAR(150) NULL,
    academic_year VARCHAR(20) NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'active',
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    deleted_at DATETIME(3) NULL,
    CONSTRAINT fk_classes_major FOREIGN KEY (major_id) REFERENCES majors(id),
    INDEX idx_classes_name (name),
    INDEX idx_classes_major (major_id),
    INDEX idx_classes_academic_year (academic_year),
    INDEX idx_classes_deleted_at (deleted_at)
);

CREATE TABLE users (
    id CHAR(36) PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    email VARCHAR(180) NOT NULL UNIQUE,
    username VARCHAR(80) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    major_id CHAR(36) NULL,
    class_id CHAR(36) NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'active',
    last_login_at DATETIME(3) NULL,
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    deleted_at DATETIME(3) NULL,
    CONSTRAINT fk_users_major FOREIGN KEY (major_id) REFERENCES majors(id) ON DELETE SET NULL,
    CONSTRAINT fk_users_class FOREIGN KEY (class_id) REFERENCES classes(id) ON DELETE SET NULL,
    INDEX idx_users_status (status),
    INDEX idx_users_deleted_at (deleted_at)
);

CREATE TABLE user_roles (
    user_id CHAR(36) NOT NULL,
    role_id CHAR(36) NOT NULL,
    PRIMARY KEY (user_id, role_id),
    CONSTRAINT fk_user_roles_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_user_roles_role FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);

CREATE TABLE refresh_sessions (
    id CHAR(36) PRIMARY KEY,
    user_id CHAR(36) NOT NULL,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    expires_at DATETIME(3) NOT NULL,
    revoked_at DATETIME(3) NULL,
    ip_address VARCHAR(60) NULL,
    user_agent VARCHAR(500) NULL,
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    deleted_at DATETIME(3) NULL,
    CONSTRAINT fk_refresh_sessions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_refresh_sessions_user (user_id),
    INDEX idx_refresh_sessions_expiry (expires_at),
    INDEX idx_refresh_sessions_deleted_at (deleted_at)
);

CREATE TABLE periods (
    id CHAR(36) PRIMARY KEY,
    name VARCHAR(180) NOT NULL,
    academic_year VARCHAR(20) NOT NULL,
    semester VARCHAR(20) NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    cohort INT NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'draft',
    notes TEXT NULL,
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    deleted_at DATETIME(3) NULL,
    INDEX idx_periods_academic_year (academic_year),
    INDEX idx_periods_status (status),
    INDEX idx_periods_dates (start_date, end_date),
    INDEX idx_periods_deleted_at (deleted_at)
);

CREATE TABLE students (
    id CHAR(36) PRIMARY KEY,
    nis VARCHAR(30) NOT NULL UNIQUE,
    nisn VARCHAR(30) NULL,
    name VARCHAR(150) NOT NULL,
    nickname VARCHAR(80) NULL,
    gender VARCHAR(20) NULL,
    class_id CHAR(36) NOT NULL,
    major_id CHAR(36) NOT NULL,
    cohort INT NOT NULL,
    phone VARCHAR(30) NULL,
    email VARCHAR(150) NULL,
    address TEXT NULL,
    parent_name VARCHAR(150) NULL,
    parent_phone VARCHAR(30) NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'active',
    pkl_status VARCHAR(50) NOT NULL DEFAULT 'unplaced',
    notes TEXT NULL,
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    deleted_at DATETIME(3) NULL,
    CONSTRAINT fk_students_class FOREIGN KEY (class_id) REFERENCES classes(id),
    CONSTRAINT fk_students_major FOREIGN KEY (major_id) REFERENCES majors(id),
    INDEX idx_students_name (name),
    INDEX idx_students_class (class_id),
    INDEX idx_students_major (major_id),
    INDEX idx_students_status (status, pkl_status),
    INDEX idx_students_deleted_at (deleted_at)
);

CREATE TABLE companies (
    id CHAR(36) PRIMARY KEY,
    name VARCHAR(180) NOT NULL,
    business_type VARCHAR(80) NULL,
    industry VARCHAR(150) NOT NULL,
    description TEXT NULL,
    address TEXT NOT NULL,
    district VARCHAR(100) NULL,
    city VARCHAR(100) NOT NULL,
    province VARCHAR(100) NULL,
    postal_code VARCHAR(15) NULL,
    phone VARCHAR(30) NULL,
    email VARCHAR(150) NULL,
    website VARCHAR(255) NULL,
    maps_url VARCHAR(500) NULL,
    status VARCHAR(40) NOT NULL DEFAULT 'candidate',
    capacity INT NOT NULL DEFAULT 0,
    cooperation_start DATE NULL,
    cooperation_end DATE NULL,
    notes TEXT NULL,
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    deleted_at DATETIME(3) NULL,
    INDEX idx_companies_name (name),
    INDEX idx_companies_industry (industry),
    INDEX idx_companies_city (city),
    INDEX idx_companies_status (status),
    INDEX idx_companies_expiry (cooperation_end),
    INDEX idx_companies_deleted_at (deleted_at)
);

CREATE TABLE company_contacts (
    id CHAR(36) PRIMARY KEY,
    company_id CHAR(36) NOT NULL,
    name VARCHAR(150) NOT NULL,
    position VARCHAR(100) NULL,
    division VARCHAR(100) NULL,
    phone VARCHAR(30) NULL,
    email VARCHAR(150) NULL,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    notes TEXT NULL,
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    deleted_at DATETIME(3) NULL,
    CONSTRAINT fk_company_contacts_company FOREIGN KEY (company_id) REFERENCES companies(id),
    INDEX idx_company_contacts_company (company_id),
    INDEX idx_company_contacts_primary (is_primary),
    INDEX idx_company_contacts_deleted_at (deleted_at)
);

CREATE TABLE company_major_capacities (
    company_id CHAR(36) NOT NULL,
    major_id CHAR(36) NOT NULL,
    capacity INT NOT NULL DEFAULT 0,
    PRIMARY KEY (company_id, major_id),
    CONSTRAINT fk_company_major_company FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE CASCADE,
    CONSTRAINT fk_company_major_major FOREIGN KEY (major_id) REFERENCES majors(id) ON DELETE CASCADE
);

CREATE TABLE supervisors (
    id CHAR(36) PRIMARY KEY,
    employee_number VARCHAR(50) NULL UNIQUE,
    name VARCHAR(150) NOT NULL,
    phone VARCHAR(30) NULL,
    email VARCHAR(150) NULL,
    major_id CHAR(36) NULL,
    position VARCHAR(100) NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'active',
    max_students INT NOT NULL DEFAULT 20,
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    deleted_at DATETIME(3) NULL,
    CONSTRAINT fk_supervisors_major FOREIGN KEY (major_id) REFERENCES majors(id) ON DELETE SET NULL,
    INDEX idx_supervisors_name (name),
    INDEX idx_supervisors_status (status),
    INDEX idx_supervisors_deleted_at (deleted_at)
);

CREATE TABLE placements (
    id CHAR(36) PRIMARY KEY,
    period_id CHAR(36) NOT NULL,
    student_id CHAR(36) NOT NULL,
    company_id CHAR(36) NOT NULL,
    company_contact_id CHAR(36) NULL,
    supervisor_id CHAR(36) NULL,
    previous_placement_id CHAR(36) NULL,
    division VARCHAR(120) NULL,
    position VARCHAR(120) NULL,
    work_system VARCHAR(30) NOT NULL,
    address TEXT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status VARCHAR(40) NOT NULL DEFAULT 'draft',
    source VARCHAR(40) NOT NULL DEFAULT 'school',
    override_reason TEXT NULL,
    transfer_reason TEXT NULL,
    notes TEXT NULL,
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    deleted_at DATETIME(3) NULL,
    CONSTRAINT fk_placements_period FOREIGN KEY (period_id) REFERENCES periods(id),
    CONSTRAINT fk_placements_student FOREIGN KEY (student_id) REFERENCES students(id),
    CONSTRAINT fk_placements_company FOREIGN KEY (company_id) REFERENCES companies(id),
    CONSTRAINT fk_placements_contact FOREIGN KEY (company_contact_id) REFERENCES company_contacts(id) ON DELETE SET NULL,
    CONSTRAINT fk_placements_supervisor FOREIGN KEY (supervisor_id) REFERENCES supervisors(id) ON DELETE SET NULL,
    CONSTRAINT fk_placements_previous FOREIGN KEY (previous_placement_id) REFERENCES placements(id) ON DELETE SET NULL,
    INDEX idx_placements_period (period_id),
    INDEX idx_placements_student (student_id),
    INDEX idx_placements_company (company_id),
    INDEX idx_placements_supervisor (supervisor_id),
    INDEX idx_placements_status (status),
    INDEX idx_placements_dates (start_date, end_date),
    INDEX idx_placements_deleted_at (deleted_at)
);

CREATE TABLE document_types (
    id CHAR(36) PRIMARY KEY,
    code VARCHAR(60) NOT NULL UNIQUE,
    name VARCHAR(150) NOT NULL,
    category VARCHAR(40) NOT NULL,
    required BOOLEAN NOT NULL DEFAULT FALSE,
    has_expiry BOOLEAN NOT NULL DEFAULT FALSE,
    max_size BIGINT NOT NULL DEFAULT 10485760,
    allowed_mime VARCHAR(500) NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'active',
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    deleted_at DATETIME(3) NULL,
    INDEX idx_document_types_category (category),
    INDEX idx_document_types_status (status),
    INDEX idx_document_types_deleted_at (deleted_at)
);

CREATE TABLE documents (
    id CHAR(36) PRIMARY KEY,
    document_type_id CHAR(36) NOT NULL,
    owner_type VARCHAR(30) NOT NULL,
    owner_id CHAR(36) NOT NULL,
    period_id CHAR(36) NULL,
    placement_id CHAR(36) NULL,
    number VARCHAR(100) NULL,
    original_name VARCHAR(255) NOT NULL,
    stored_name VARCHAR(255) NOT NULL,
    path VARCHAR(500) NOT NULL,
    mime_type VARCHAR(120) NOT NULL,
    size BIGINT NOT NULL,
    issued_at DATETIME(3) NULL,
    valid_from DATETIME(3) NULL,
    valid_until DATETIME(3) NULL,
    status VARCHAR(40) NOT NULL DEFAULT 'pending',
    version INT NOT NULL DEFAULT 1,
    verified_by CHAR(36) NULL,
    verified_at DATETIME(3) NULL,
    notes TEXT NULL,
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    deleted_at DATETIME(3) NULL,
    CONSTRAINT fk_documents_type FOREIGN KEY (document_type_id) REFERENCES document_types(id),
    CONSTRAINT fk_documents_period FOREIGN KEY (period_id) REFERENCES periods(id) ON DELETE SET NULL,
    CONSTRAINT fk_documents_placement FOREIGN KEY (placement_id) REFERENCES placements(id) ON DELETE SET NULL,
    CONSTRAINT fk_documents_verifier FOREIGN KEY (verified_by) REFERENCES users(id) ON DELETE SET NULL,
    INDEX idx_documents_owner (owner_type, owner_id),
    INDEX idx_documents_period (period_id),
    INDEX idx_documents_placement (placement_id),
    INDEX idx_documents_status (status),
    INDEX idx_documents_expiry (valid_until),
    INDEX idx_documents_deleted_at (deleted_at)
);

CREATE TABLE administrative_readiness (
    id CHAR(36) PRIMARY KEY,
    student_id CHAR(36) NOT NULL,
    period_id CHAR(36) NOT NULL,
    placement_id CHAR(36) NULL,
    data_complete BOOLEAN NOT NULL DEFAULT FALSE,
    company_assigned BOOLEAN NOT NULL DEFAULT FALSE,
    contact_available BOOLEAN NOT NULL DEFAULT FALSE,
    supervisor_assigned BOOLEAN NOT NULL DEFAULT FALSE,
    dates_set BOOLEAN NOT NULL DEFAULT FALSE,
    acceptance_letter_valid BOOLEAN NOT NULL DEFAULT FALSE,
    parent_permission_valid BOOLEAN NOT NULL DEFAULT FALSE,
    introduction_letter_valid BOOLEAN NOT NULL DEFAULT FALSE,
    required_count INT NOT NULL DEFAULT 8,
    completed_count INT NOT NULL DEFAULT 0,
    percentage DECIMAL(5,2) NOT NULL DEFAULT 0,
    status VARCHAR(30) NOT NULL DEFAULT 'incomplete',
    override_reason TEXT NULL,
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    deleted_at DATETIME(3) NULL,
    CONSTRAINT fk_readiness_student FOREIGN KEY (student_id) REFERENCES students(id),
    CONSTRAINT fk_readiness_period FOREIGN KEY (period_id) REFERENCES periods(id),
    CONSTRAINT fk_readiness_placement FOREIGN KEY (placement_id) REFERENCES placements(id) ON DELETE SET NULL,
    UNIQUE KEY uq_readiness_student_period (student_id, period_id),
    INDEX idx_readiness_status (status),
    INDEX idx_readiness_deleted_at (deleted_at)
);

CREATE TABLE archives (
    id CHAR(36) PRIMARY KEY,
    period_id CHAR(36) NOT NULL UNIQUE,
    archived_by CHAR(36) NOT NULL,
    archived_at DATETIME(3) NOT NULL,
    reason TEXT NULL,
    snapshot LONGTEXT NULL,
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    deleted_at DATETIME(3) NULL,
    CONSTRAINT fk_archives_period FOREIGN KEY (period_id) REFERENCES periods(id),
    CONSTRAINT fk_archives_user FOREIGN KEY (archived_by) REFERENCES users(id),
    INDEX idx_archives_archived_at (archived_at),
    INDEX idx_archives_deleted_at (deleted_at)
);

CREATE TABLE audit_logs (
    id CHAR(36) PRIMARY KEY,
    actor_id CHAR(36) NULL,
    action VARCHAR(80) NOT NULL,
    resource VARCHAR(100) NOT NULL,
    resource_id CHAR(36) NULL,
    request_id VARCHAR(100) NULL,
    before_json LONGTEXT NULL,
    after_json LONGTEXT NULL,
    reason TEXT NULL,
    ip_address VARCHAR(60) NULL,
    user_agent VARCHAR(500) NULL,
    created_at DATETIME(3) NOT NULL,
    updated_at DATETIME(3) NOT NULL,
    deleted_at DATETIME(3) NULL,
    CONSTRAINT fk_audit_logs_actor FOREIGN KEY (actor_id) REFERENCES users(id) ON DELETE SET NULL,
    INDEX idx_audit_logs_actor (actor_id),
    INDEX idx_audit_logs_action (action),
    INDEX idx_audit_logs_resource (resource, resource_id),
    INDEX idx_audit_logs_request (request_id),
    INDEX idx_audit_logs_created (created_at)
);
