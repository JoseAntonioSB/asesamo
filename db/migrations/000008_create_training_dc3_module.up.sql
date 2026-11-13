-- 000008_create_training_dc3_module.up.sql
-- Module: Training & STPS DC-3 / DC-4 Competency Certificates (NOM-019, NOM-030, STPS)

CREATE TYPE course_thematic_area AS ENUM (
    'safety',            -- Seguridad e Higiene en el Trabajo (6000)
    'technical',         -- Operación y Mantenimiento Técnico (5000)
    'environmental',     -- Gestión y Protección Ambiental
    'administrative',    -- Administración y Organización
    'specialized'        -- Trabajos Críticos de Alto Riesgo (Alturas, Confinados, Fuego, Eléctrico)
);

CREATE TYPE certificate_status AS ENUM (
    'draft',             -- Borrador
    'issued',            -- Emitida formalmente
    'revoked',           -- Revocada por incumplimiento
    'expired'            -- Vencida por cumplimiento de vigencia
);

-- 1. Catálogo de Cursos de Capacitación y Adiestramiento
CREATE TABLE training_courses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    code VARCHAR(50) NOT NULL, -- e.g. 'CUR-ALT-001'
    title VARCHAR(150) NOT NULL,
    description TEXT,
    thematic_area course_thematic_area NOT NULL DEFAULT 'safety',
    stps_thematic_code VARCHAR(10) NOT NULL DEFAULT '6000', -- Código de área temática STPS (e.g. 6000 para Seguridad)
    duration_hours INTEGER NOT NULL DEFAULT 8,
    validity_years INTEGER NOT NULL DEFAULT 1, -- Vigencia recomendada en años antes de recertificación
    normative_reference VARCHAR(100), -- e.g. 'NOM-009-STPS-2011'
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_tenant_course_code UNIQUE (tenant_id, code)
);

-- 2. Edición o Sesión Impartida de un Curso
CREATE TABLE training_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    course_id UUID NOT NULL REFERENCES training_courses(id) ON DELETE RESTRICT,
    session_number VARCHAR(50) NOT NULL, -- e.g. 'SES-2026-001'
    instructor_name VARCHAR(150) NOT NULL,
    instructor_stps_register VARCHAR(100), -- Número de registro de Agente Capacitador Externo (STPS)
    location VARCHAR(120) NOT NULL DEFAULT 'Sala de Capacitación Planta',
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_tenant_session_number UNIQUE (tenant_id, session_number)
);

-- 3. Constancias Oficiales de Competencias Laborales (Formato DC-3)
CREATE TABLE training_certificates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    session_id UUID REFERENCES training_sessions(id) ON DELETE SET NULL,
    course_id UUID NOT NULL REFERENCES training_courses(id) ON DELETE RESTRICT,
    worker_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    contractor_id UUID REFERENCES contractors(id) ON DELETE SET NULL,
    
    folio_number VARCHAR(50) NOT NULL, -- Folio único impreso en formato DC-3 e.g. 'DC3-2026-00001'
    status certificate_status NOT NULL DEFAULT 'issued',
    
    -- Datos requeridos en el formato oficial DC-3 de la STPS:
    worker_occupation_code VARCHAR(20) NOT NULL DEFAULT '8100', -- Ocupación específica según catálogo nacional
    worker_curp VARCHAR(18) NOT NULL,
    worker_full_name VARCHAR(150) NOT NULL,
    
    company_name VARCHAR(150) NOT NULL,
    company_rfc VARCHAR(13) NOT NULL,
    
    legal_rep_name VARCHAR(150) NOT NULL,
    worker_rep_name VARCHAR(150) NOT NULL,
    instructor_name VARCHAR(150) NOT NULL,
    
    issue_date DATE NOT NULL,
    expiration_date DATE NOT NULL,
    score_percentage NUMERIC(5,2) DEFAULT 100.00,
    
    digital_signature_token VARCHAR(255),
    document_r2_key VARCHAR(255),
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_tenant_certificate_folio UNIQUE (tenant_id, folio_number)
);

CREATE INDEX idx_training_courses_tenant ON training_courses(tenant_id);
CREATE INDEX idx_training_sessions_tenant ON training_sessions(tenant_id);
CREATE INDEX idx_training_certificates_tenant_worker ON training_certificates(tenant_id, worker_id);
CREATE INDEX idx_training_certificates_expiration ON training_certificates(expiration_date);
CREATE INDEX idx_training_certificates_course ON training_certificates(course_id);
