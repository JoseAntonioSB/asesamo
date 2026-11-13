-- 000004_create_safety_inspections.up.sql
-- Module: Safety Inspections & EHS Audits (NOM-002, NOM-004, NOM-009, NOM-017, NOM-033)

CREATE TYPE inspection_type AS ENUM (
    'routine',       -- Rutinaria / Diaria
    'planned',       -- Planeada / Mensual / Trimestral
    'special',       -- Extraordinaria / Post-incidente
    'regulatory'     -- Auditoría de Cumplimiento Normativo (STPS)
);

CREATE TYPE inspection_status AS ENUM (
    'draft',         -- En borrador / Edición
    'in_progress',   -- En ejecución en campo
    'completed',     -- Finalizada con scoring calculado
    'requires_capa', -- Finalizada pero con hallazgos críticos que exigen plan de acción
    'cancelled'      -- Cancelada
);

CREATE TYPE item_compliance_status AS ENUM (
    'compliant',     -- Conforme (cumple estándar)
    'non_compliant', -- No conforme (desvío / riesgo detectado)
    'not_applicable' -- No aplica para esta área o proceso
);

CREATE TYPE finding_severity AS ENUM (
    'low',           -- Riesgo leve / desvío menor
    'medium',        -- Riesgo moderado
    'high',          -- Riesgo alto
    'critical'       -- Riesgo inminente / paro de trabajo
);

CREATE TYPE capa_status AS ENUM (
    'open',          -- Acción correctiva abierta
    'in_progress',   -- En atención
    'resolved',      -- Resuelta (evidencia enviada)
    'verified',      -- Verificada por EHS (cerrada formalmente)
    'overdue'        -- Vencida
);

-- 1. Plantillas de Inspección / Checklists
CREATE TABLE inspection_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    title VARCHAR(150) NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL, -- e.g. 'fire_safety', 'work_at_height', 'general_ehs', 'machinery'
    normative_reference VARCHAR(100), -- e.g. 'NOM-002-STPS-2010', 'NOM-009-STPS-2011'
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 2. Items/Preguntas de la Plantilla
CREATE TABLE inspection_template_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id UUID NOT NULL REFERENCES inspection_templates(id) ON DELETE CASCADE,
    category VARCHAR(50) NOT NULL, -- e.g. 'Señalización', 'Estado de Equipos', 'EPP'
    question TEXT NOT NULL,
    guidance_notes TEXT,
    weight INTEGER NOT NULL DEFAULT 1, -- Ponderación para el scoring (1 a 5)
    critical_risk BOOLEAN NOT NULL DEFAULT false, -- Si es 'non_compliant', dispara CAPA inmediata
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 3. Registro de Inspección en Campo
CREATE TABLE safety_inspections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    template_id UUID NOT NULL REFERENCES inspection_templates(id) ON DELETE RESTRICT,
    inspection_number VARCHAR(50) NOT NULL, -- e.g. 'INSP-2026-0001'
    inspection_type inspection_type NOT NULL DEFAULT 'routine',
    status inspection_status NOT NULL DEFAULT 'draft',
    area_or_location VARCHAR(120) NOT NULL,
    contractor_id UUID REFERENCES contractors(id) ON DELETE SET NULL, -- Opcional, si inspecciona a contratista
    inspector_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    score_percentage NUMERIC(5,2), -- 0.00% a 100.00%
    total_items INTEGER NOT NULL DEFAULT 0,
    compliant_items INTEGER NOT NULL DEFAULT 0,
    non_compliant_items INTEGER NOT NULL DEFAULT 0,
    na_items INTEGER NOT NULL DEFAULT 0,
    notes TEXT,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_tenant_inspection_number UNIQUE (tenant_id, inspection_number)
);

-- 4. Respuestas a los Items en la Inspección
CREATE TABLE safety_inspection_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    inspection_id UUID NOT NULL REFERENCES safety_inspections(id) ON DELETE CASCADE,
    template_item_id UUID NOT NULL REFERENCES inspection_template_items(id) ON DELETE RESTRICT,
    status item_compliance_status NOT NULL DEFAULT 'compliant',
    finding_notes TEXT,
    evidence_photo_url VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_inspection_item UNIQUE (inspection_id, template_item_id)
);

-- 5. Plan de Acciones Correctivas y Preventivas (CAPA)
CREATE TABLE inspection_action_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    inspection_id UUID NOT NULL REFERENCES safety_inspections(id) ON DELETE CASCADE,
    inspection_item_id UUID REFERENCES safety_inspection_items(id) ON DELETE SET NULL,
    description TEXT NOT NULL,
    severity finding_severity NOT NULL DEFAULT 'medium',
    status capa_status NOT NULL DEFAULT 'open',
    assigned_to_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    contractor_id UUID REFERENCES contractors(id) ON DELETE SET NULL,
    due_date DATE NOT NULL,
    resolved_at TIMESTAMP WITH TIME ZONE,
    verified_at TIMESTAMP WITH TIME ZONE,
    resolution_notes TEXT,
    resolution_photo_url VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Índices de consulta frecuente y aislamiento tenant
CREATE INDEX idx_safety_inspections_tenant_status ON safety_inspections(tenant_id, status);
CREATE INDEX idx_safety_inspections_contractor ON safety_inspections(contractor_id);
CREATE INDEX idx_safety_inspections_inspector ON safety_inspections(inspector_id);
CREATE INDEX idx_safety_inspection_items_inspection ON safety_inspection_items(inspection_id);
CREATE INDEX idx_inspection_action_plans_tenant_status ON inspection_action_plans(tenant_id, status);
CREATE INDEX idx_inspection_action_plans_due_date ON inspection_action_plans(due_date);
