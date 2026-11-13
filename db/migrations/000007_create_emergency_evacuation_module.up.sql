-- 000007_create_emergency_evacuation_module.up.sql
-- Module: Emergency Response, Evacuation & Roll Call at Muster Points (NOM-002-STPS)

CREATE TYPE evacuation_type AS ENUM (
    'drill',        -- Simulacro programado
    'fire',         -- Incendio real
    'gas_leak',     -- Fuga de gas / químico
    'earthquake',   -- Sismo / Terremoto
    'spill',        -- Derrame mayor
    'other'         -- Otra contingencia
);

CREATE TYPE evacuation_status AS ENUM (
    'active',       -- Evacuación en curso
    'all_clear',    -- Contingencia superada / Todos a salvo
    'cancelled'     -- Falsa alarma / Cancelada
);

CREATE TYPE roll_call_status AS ENUM (
    'safe',         -- Presente y a salvo en el Punto de Reunión
    'missing',      -- Faltante / No localizado
    'hospitalized'  -- Trasladado a centro médico / lesionado
);

-- 1. Puntos de Reunión / Zonas de Seguridad de la Planta
CREATE TABLE muster_points (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    code VARCHAR(30) NOT NULL, -- e.g. 'PR-01'
    name VARCHAR(100) NOT NULL, -- e.g. 'Punto de Reunión 1 - Estacionamiento Principal'
    location_description TEXT,
    capacity INTEGER NOT NULL DEFAULT 100,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 2. Registro de Evento de Evacuación / Simulacro
CREATE TABLE emergency_evacuations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    evacuation_number VARCHAR(50) NOT NULL, -- e.g. 'EVAC-2026-0001'
    evacuation_type evacuation_type NOT NULL,
    status evacuation_status NOT NULL DEFAULT 'active',
    description TEXT,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMP WITH TIME ZONE,
    evacuation_time_seconds INTEGER, -- Tiempo total de desalojo
    initiated_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_tenant_evac_number UNIQUE (tenant_id, evacuation_number)
);

-- 3. Pase de Lista en Punto de Reunión
CREATE TABLE evacuation_roll_call (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    evacuation_id UUID NOT NULL REFERENCES emergency_evacuations(id) ON DELETE CASCADE,
    worker_id UUID NOT NULL REFERENCES workers(id) ON DELETE RESTRICT,
    muster_point_id UUID REFERENCES muster_points(id) ON DELETE SET NULL,
    status roll_call_status NOT NULL DEFAULT 'missing',
    checked_in_at TIMESTAMP WITH TIME ZONE,
    checked_in_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    notes TEXT,
    CONSTRAINT uk_evac_worker UNIQUE (evacuation_id, worker_id)
);

CREATE INDEX idx_emergency_evacuations_tenant_status ON emergency_evacuations(tenant_id, status);
CREATE INDEX idx_evacuation_roll_call_evac_status ON evacuation_roll_call(evacuation_id, status);
