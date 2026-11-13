-- 000006_create_epp_module.up.sql
-- Module: Personal Protective Equipment (EPP) Management & Matrix (NOM-017-STPS)

CREATE TYPE epp_category AS ENUM (
    'head',          -- Cabeza (Cascos clase E/G, barboquejos)
    'eyes_face',     -- Ojos y Cara (Lentes de seguridad, caretas de soldador/esmeril)
    'hearing',       -- Auditiva (Tapones auditivos, conchas/orejeras)
    'respiratory',   -- Respiratoria (Mascarillas N95, media cara con cartuchos para vapores/humos)
    'hands',         -- Extremidades superiores (Guantes de carnaza, nitrilo, dieléctricos)
    'body',          -- Tronco y cuerpo (Chaleco reflejante clase 2, mandiles de carnaza, overoles)
    'feet',          -- Extremidades inferiores (Calzado con casquillo de poliamida/acero, dieléctrico)
    'fall_arrest'    -- Alturas (Arnés de cuerpo completo, línea de vida con amortiguador)
);

CREATE TYPE epp_delivery_status AS ENUM (
    'delivered_active', -- Entregado y vigente
    'replaced',         -- Reemplazado por desgaste / daño
    'expired',          -- Caducado
    'returned'          -- Devuelto a almacén
);

-- 1. Catálogo de Artículos de EPP
CREATE TABLE epp_catalog (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    code VARCHAR(50) NOT NULL, -- e.g. 'EPP-CAS-001'
    name VARCHAR(120) NOT NULL,
    category epp_category NOT NULL,
    normative_standard VARCHAR(100), -- e.g. 'NOM-115-STPS-2009', 'ANSI/ISEA Z87.1'
    lifespan_days INTEGER NOT NULL DEFAULT 180, -- Vida útil en días antes de requerir recambio
    stock_quantity INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_tenant_epp_code UNIQUE (tenant_id, code)
);

-- 2. Bitácora de Asignación y Entrega de EPP al Trabajador
CREATE TABLE epp_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
    epp_item_id UUID NOT NULL REFERENCES epp_catalog(id) ON DELETE RESTRICT,
    worker_id UUID NOT NULL REFERENCES workers(id) ON DELETE RESTRICT,
    contractor_id UUID REFERENCES contractors(id) ON DELETE SET NULL,
    delivery_date DATE NOT NULL,
    expiration_date DATE NOT NULL,
    status epp_delivery_status NOT NULL DEFAULT 'delivered_active',
    delivered_by_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    signature_data TEXT, -- Firma digital o token de recepción del trabajador
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_epp_catalog_tenant_category ON epp_catalog(tenant_id, category);
CREATE INDEX idx_epp_deliveries_tenant_worker ON epp_deliveries(tenant_id, worker_id);
CREATE INDEX idx_epp_deliveries_expiration ON epp_deliveries(expiration_date);
