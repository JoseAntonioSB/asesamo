-- 000009_create_occupational_health_module.up.sql
-- Module: Occupational Health & Periodic Medical Exams (NOM-030-STPS, NOM-009, NOM-033)

CREATE TYPE medical_fitness_status AS ENUM (
    'fit',                   -- Apto médico integral sin restricciones
    'fit_with_restrictions', -- Apto con restricciones laborales (e.g. lentes con graduación, límite de carga < 15kg)
    'unfit_temporary',       -- No apto temporalmente (bajo tratamiento o incapacidad)
    'unfit_permanent'        -- No apto definitivo para tareas de alto riesgo
);

CREATE TYPE medical_exam_type AS ENUM (
    'pre_placement',         -- Examen médico de ingreso
    'periodic_annual',       -- Periódico anual reglamentario NOM-030
    'height_clearance',      -- Dictamen de aptitud para trabajos en altura (NOM-009)
    'confined_space',        -- Dictamen de aptitud para espacios confinados (NOM-033)
    'post_incident',         -- Valoración post-accidente / reintegro laboral
    'antidoping'             -- Tamizaje toxicológico antidoping
);

-- 1. Expediente Médico Ocupacional Base del Trabajador
CREATE TABLE worker_medical_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    worker_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    
    blood_type VARCHAR(10) NOT NULL, -- e.g. 'O+', 'A+', 'B+', 'O-'
    allergies TEXT,
    chronic_conditions TEXT,
    emergency_contact_name VARCHAR(150) NOT NULL,
    emergency_contact_phone VARCHAR(50) NOT NULL,
    emergency_contact_relationship VARCHAR(50) NOT NULL,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_tenant_worker_medical_profile UNIQUE (tenant_id, worker_id)
);

-- 2. Evaluaciones y Exámenes Médicos Periódicos
CREATE TABLE occupational_medical_exams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    worker_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    contractor_id UUID REFERENCES contractors(id) ON DELETE SET NULL,
    
    exam_number VARCHAR(50) NOT NULL, -- e.g. 'MED-2026-0001'
    exam_type medical_exam_type NOT NULL DEFAULT 'periodic_annual',
    fitness_status medical_fitness_status NOT NULL DEFAULT 'fit',
    
    exam_date DATE NOT NULL,
    expiration_date DATE NOT NULL,
    
    doctor_name VARCHAR(150) NOT NULL,
    doctor_license VARCHAR(50) NOT NULL, -- Cédula Profesional del Médico Dictaminador
    clinic_or_institution VARCHAR(150) NOT NULL DEFAULT 'Servicio Médico Ocupacional de Planta',
    
    restrictions_notes TEXT,
    recommendations TEXT,
    
    -- Resultados de pruebas paraclínicas estándar STPS
    visual_acuity_result VARCHAR(50), -- e.g. '20/20 corregida con lentes'
    audiometry_result VARCHAR(50),    -- e.g. 'Normal / Sin trauma acústico'
    spirometry_result VARCHAR(50),    -- e.g. 'Patrón normal FVC > 80%'
    antidoping_result VARCHAR(30),    -- e.g. 'Negativo 5 paneles'
    blood_pressure VARCHAR(20),       -- e.g. '120/80 mmHg'
    
    certificate_r2_key VARCHAR(255),
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_tenant_exam_number UNIQUE (tenant_id, exam_number)
);

CREATE INDEX idx_worker_medical_profiles_tenant_worker ON worker_medical_profiles(tenant_id, worker_id);
CREATE INDEX idx_occupational_medical_exams_tenant_worker ON occupational_medical_exams(tenant_id, worker_id);
CREATE INDEX idx_occupational_medical_exams_expiration ON occupational_medical_exams(expiration_date);
CREATE INDEX idx_occupational_medical_exams_fitness ON occupational_medical_exams(fitness_status);
