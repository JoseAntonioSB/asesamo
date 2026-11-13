-- 000005_create_incidents_module.up.sql
-- Module: Incident & Near-Miss Investigation (NOM-021-STPS)

CREATE TYPE incident_type AS ENUM (
    'unsafe_act',           -- Acto Inseguro
    'unsafe_condition',     -- Condición Insegura
    'near_miss',            -- Cuasi-accidente (Riesgo potencial sin lesión)
    'first_aid',            -- Primeros Auxilios (Lesión menor sin incapacidad)
    'medical_treatment',    -- Atención Médica Externa
    'lost_time_injury',     -- Accidente con Pérdida de Días (LTI)
    'fatality'              -- Fatalidad
);

CREATE TYPE incident_severity AS ENUM (
    'minor',                -- Leve
    'moderate',             -- Moderado
    'major',                -- Grave
    'catastrophic'          -- Crítico / Catastrófico
);

CREATE TYPE incident_status AS ENUM (
    'reported',             -- Notificado / En espera de investigación
    'investigating',        -- Investigación activa de causa raíz
    'actions_pending',      -- Acciones correctivas abiertas
    'closed',               -- Cerrado y verificado por EHS
    'cancelled'             -- Desestimado / Duplicado
);

CREATE TABLE safety_incidents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    incident_number VARCHAR(50) NOT NULL, -- e.g. 'INC-2026-0001'
    incident_type incident_type NOT NULL,
    severity incident_severity NOT NULL DEFAULT 'minor',
    status incident_status NOT NULL DEFAULT 'reported',
    
    title VARCHAR(150) NOT NULL,
    description TEXT NOT NULL,
    location VARCHAR(120) NOT NULL,
    event_timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    
    reported_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    involved_worker_id UUID REFERENCES users(id) ON DELETE SET NULL,
    contractor_id UUID REFERENCES contractors(id) ON DELETE SET NULL,
    
    days_lost INTEGER NOT NULL DEFAULT 0,
    evidence_photo_url VARCHAR(255),
    
    -- Análisis de Causa Raíz (5 Porqués / Metodología Causal)
    root_cause_summary TEXT,
    immediate_cause TEXT,  -- Falla de equipo, falta de EPP, etc.
    basic_cause TEXT,      -- Falta de capacitación, mantenimiento inadecuado
    
    investigated_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    investigated_at TIMESTAMP WITH TIME ZONE,
    closed_at TIMESTAMP WITH TIME ZONE,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT uk_tenant_incident_number UNIQUE (tenant_id, incident_number)
);

CREATE INDEX idx_safety_incidents_tenant_status ON safety_incidents(tenant_id, status);
CREATE INDEX idx_safety_incidents_type ON safety_incidents(incident_type);
CREATE INDEX idx_safety_incidents_contractor ON safety_incidents(contractor_id);
CREATE INDEX idx_safety_incidents_timestamp ON safety_incidents(event_timestamp);
