-- ============================================================
-- ASESAMO — Phase 3 Operational Safety Modules Seed Data
-- File: db/seeds/000002_phase3_operational_seeds.sql
-- Modules:
--   1. Safety Inspections & EHS Audits (NOM-002, NOM-009, NOM-017)
--   2. Safety Incidents & Near-Misses (NOM-021-STPS)
--   3. EPP Matrix & Delivery Control (NOM-017-STPS)
--   4. Emergency Evacuation & Roll Call at Muster Points (NOM-002-STPS)
-- ============================================================

BEGIN;

-- Common IDs from 000001_seed_dev_data.sql:
-- Tenant:     11111111-1111-1111-1111-111111111111
-- Plant:      22222222-2222-2222-2222-222222222222
-- Admin/EHS:  44444444-4444-4444-4444-444444444443 (Ing. Laura Sanchez - EHS)
-- Contractor: 55555555-5555-5555-5555-555555555555 (Mantenimiento Industrial del Norte)
-- Worker 1:   66666666-6666-6666-6666-666666666661 (Carlos Mendoza)
-- Worker 2:   66666666-6666-6666-6666-666666666662 (Juan Perez)
-- Worker 3:   66666666-6666-6666-6666-666666666663 (Roberto Garcia)

-- ------------------------------------------------------------
-- 1. INSPECTION TEMPLATES & ITEMS
-- ------------------------------------------------------------
-- Template 1: NOM-002 Extintores y Protección Contra Incendios
INSERT INTO inspection_templates (
  id, tenant_id, title, description, category, normative_reference, is_active
) VALUES (
  '77777777-1111-1111-1111-000000000001',
  '11111111-1111-1111-1111-111111111111',
  'Inspección Mensual de Extintores y Sistemas Contra Incendio',
  'Verificación del estado operativo, presión, señalización y acceso a extintores portátiles en nave industrial.',
  'fire_safety',
  'NOM-002-STPS-2010',
  TRUE
) ON CONFLICT (id) DO NOTHING;

INSERT INTO inspection_template_items (
  id, template_id, category, question, guidance_notes, weight, critical_risk, sort_order
) VALUES
  (
    '77777777-1111-1111-2222-000000000001',
    '77777777-1111-1111-1111-000000000001',
    'Ubicación y Acceso',
    '¿El extintor cuenta con libre acceso, sin obstáculos a menos de 1 metro de claro frontal?',
    'No deben existir tarimas, cajas o maquinaria bloqueando el equipo.',
    2, FALSE, 1
  ),
  (
    '77777777-1111-1111-2222-000000000002',
    '77777777-1111-1111-1111-000000000001',
    'Operatividad',
    '¿El manómetro indica presión en rango operativo (zona verde) y el precinto de seguridad está intacto?',
    'Si la aguja está en recarga o sobrepresión, o el pasador está roto, marcar no conforme.',
    3, TRUE, 2
  ),
  (
    '77777777-1111-1111-2222-000000000003',
    '77777777-1111-1111-1111-000000000001',
    'Señalización',
    '¿Cuenta con señalización fotoluminiscente visible a no menos de 5 metros conforme a NOM-026-STPS?',
    'Letrero de tipo de fuego (A, B, C) colocado a altura normativa de 1.80m.',
    1, FALSE, 3
  ),
  (
    '77777777-1111-1111-2222-000000000004',
    '77777777-1111-1111-1111-000000000001',
    'Mantenimiento',
    '¿El collarín y etiqueta de recarga anual se encuentran vigentes por taller acreditado?',
    'La vigencia máxima es de 12 meses a partir de la fecha perforada en la etiqueta.',
    2, TRUE, 4
  )
ON CONFLICT (id) DO NOTHING;

-- Template 2: NOM-009 Trabajos en Altura y Andamios
INSERT INTO inspection_templates (
  id, tenant_id, title, description, category, normative_reference, is_active
) VALUES (
  '77777777-1111-1111-1111-000000000002',
  '11111111-1111-1111-1111-111111111111',
  'Auditoría Previa de Andamios y Trabajo en Altura (>1.80m)',
  'Inspección obligatoria antes del inicio de maniobras de altura por personal o contratistas.',
  'work_at_height',
  'NOM-009-STPS-2011',
  TRUE
) ON CONFLICT (id) DO NOTHING;

INSERT INTO inspection_template_items (
  id, template_id, category, question, guidance_notes, weight, critical_risk, sort_order
) VALUES
  (
    '77777777-1111-1111-2222-000000000005',
    '77777777-1111-1111-1111-000000000002',
    'Estructura',
    '¿El andamio está nivelado con husillos ajustables sobre placas base sólidas y crucetas completas?',
    'Prohibido usar tabiques o trozos de madera para nivelar.',
    3, TRUE, 1
  ),
  (
    '77777777-1111-1111-2222-000000000006',
    '77777777-1111-1111-1111-000000000002',
    'Plataforma',
    '¿La plataforma cuenta con barandal superior (1.10m), intermedio (0.50m) y rodapié de 15cm?',
    'Tablones metálicos antideslizantes amarrados sin aberturas mayores a 2.5cm.',
    2, TRUE, 2
  ),
  (
    '77777777-1111-1111-2222-000000000007',
    '77777777-1111-1111-1111-000000000002',
    'EPP Anticaídas',
    '¿Los trabajadores portan arnés certificado de cuerpo completo anclado a punto estructural independiente?',
    'Resistencia mínima requerida en anclaje: 2,268 kg (5,000 lbs).',
    3, TRUE, 3
  ),
  (
    '77777777-1111-1111-2222-000000000008',
    '77777777-1111-1111-1111-000000000002',
    'Delimitación',
    '¿El área inferior está delimitada con cinta perimetral de peligro y letreros de advertencia?',
    'Evitar paso de personas en el radio de caída potencial.',
    1, FALSE, 4
  )
ON CONFLICT (id) DO NOTHING;

-- Template 3: NOM-017 EPP Básico en Planta
INSERT INTO inspection_templates (
  id, tenant_id, title, description, category, normative_reference, is_active
) VALUES (
  '77777777-1111-1111-1111-000000000003',
  '11111111-1111-1111-1111-111111111111',
  'Auditoría en Piso: Uso Correcto de EPP Obligatorio',
  'Monitoreo rutinario de cumplimiento de equipo de protección personal por cuadrilla.',
  'general_ehs',
  'NOM-017-STPS-2008',
  TRUE
) ON CONFLICT (id) DO NOTHING;

INSERT INTO inspection_template_items (
  id, template_id, category, question, guidance_notes, weight, critical_risk, sort_order
) VALUES
  (
    '77777777-1111-1111-2222-000000000009',
    '77777777-1111-1111-1111-000000000003',
    'Protección a la Cabeza',
    '¿El personal porta casco Clase E/G con barboquejo colocado correctamente?',
    'Sin fracturas, calcomanías que cubran fisuras o suspensiones vencidas.',
    2, TRUE, 1
  ),
  (
    '77777777-1111-1111-2222-000000000010',
    '77777777-1111-1111-1111-000000000003',
    'Protección Ocular',
    '¿Portan lentes de seguridad con protección lateral certificados ANSI Z87.1?',
    'Micas limpias y libres de rayaduras profundas.',
    2, FALSE, 2
  ),
  (
    '77777777-1111-1111-2222-000000000011',
    '77777777-1111-1111-1111-000000000003',
    'Calzado de Seguridad',
    '¿Calzado industrial dieléctrico con casquillo protector certificado NOM-113-STPS?',
    'Suela antiderrapante limpia y en buenas condiciones.',
    2, TRUE, 3
  )
ON CONFLICT (id) DO NOTHING;

-- ------------------------------------------------------------
-- 2. SAMPLE SAFETY INSPECTION & CAPA RECORD
-- ------------------------------------------------------------
INSERT INTO safety_inspections (
  id, tenant_id, template_id, inspection_number, inspection_type, status,
  area_or_location, contractor_id, inspector_id, score_percentage,
  total_items, compliant_items, non_compliant_items, na_items, notes,
  started_at, completed_at
) VALUES (
  '77777777-2222-1111-1111-000000000001',
  '11111111-1111-1111-1111-111111111111',
  '77777777-1111-1111-1111-000000000001',
  'INSP-2026-0001',
  'planned',
  'requires_capa',
  'Nave B - Línea de Ensamble y Almacén de Químicos',
  '55555555-5555-5555-5555-555555555555',
  '44444444-4444-4444-4444-444444444443',
  75.00,
  4, 3, 1, 0,
  'Se detectó un extintor con etiqueta de mantenimiento anual vencida hace 15 días.',
  now() - INTERVAL '2 days',
  now() - INTERVAL '2 days' + INTERVAL '45 minutes'
) ON CONFLICT (id) DO NOTHING;

INSERT INTO safety_inspection_items (
  id, inspection_id, template_item_id, status, finding_notes
) VALUES
  (
    '77777777-3333-1111-1111-000000000001',
    '77777777-2222-1111-1111-000000000001',
    '77777777-1111-1111-2222-000000000001',
    'compliant',
    'Acceso libre y despejado.'
  ),
  (
    '77777777-3333-1111-1111-000000000002',
    '77777777-2222-1111-1111-000000000001',
    '77777777-1111-1111-2222-000000000002',
    'compliant',
    'Presión en verde, seguro intacto.'
  ),
  (
    '77777777-3333-1111-1111-000000000003',
    '77777777-2222-1111-1111-000000000001',
    '77777777-1111-1111-2222-000000000003',
    'compliant',
    'Señalética en altura adecuada.'
  ),
  (
    '77777777-3333-1111-1111-000000000004',
    '77777777-2222-1111-1111-000000000001',
    '77777777-1111-1111-2222-000000000004',
    'non_compliant',
    'Etiqueta de recarga vencida en octubre 2026. Requiere reemplazo de extintor de reserva.'
  )
ON CONFLICT (id) DO NOTHING;

INSERT INTO inspection_action_plans (
  id, tenant_id, inspection_id, inspection_item_id, description, severity,
  status, assigned_to_user_id, contractor_id, due_date
) VALUES (
  '77777777-4444-1111-1111-000000000001',
  '11111111-1111-1111-1111-111111111111',
  '77777777-2222-1111-1111-000000000001',
  '77777777-3333-1111-1111-000000000004',
  'Reemplazar extintor PQS de 9kg en Almacén Químicos por unidad recargada y certificada.',
  'high',
  'open',
  '44444444-4444-4444-4444-444444444443',
  '55555555-5555-5555-5555-555555555555',
  CURRENT_DATE + INTERVAL '3 days'
) ON CONFLICT (id) DO NOTHING;

-- ------------------------------------------------------------
-- 3. EPP CATALOG & DELIVERIES (NOM-017-STPS)
-- ------------------------------------------------------------
INSERT INTO epp_catalog (
  id, tenant_id, code, name, category, normative_standard, lifespan_days, stock_quantity, is_active
) VALUES
  (
    '88888888-1111-1111-1111-000000000001',
    '11111111-1111-1111-1111-111111111111',
    'EPP-CAS-001',
    'Casco de Seguridad Industrial Clase E Dielectrico 20kV',
    'head',
    'NOM-115-STPS-2009',
    365, 45, TRUE
  ),
  (
    '88888888-1111-1111-1111-000000000002',
    '11111111-1111-1111-1111-111111111111',
    'EPP-LEN-001',
    'Gafas de Seguridad Anti-impacto y Filtro UV 99.9%',
    'eyes_face',
    'ANSI/ISEA Z87.1-2020',
    180, 120, TRUE
  ),
  (
    '88888888-1111-1111-1111-000000000003',
    '11111111-1111-1111-1111-111111111111',
    'EPP-CAL-001',
    'Bota de Seguridad con Casquillo de Poliamida Dieléctrico',
    'feet',
    'NOM-113-STPS-2009',
    365, 30, TRUE
  ),
  (
    '88888888-1111-1111-1111-000000000004',
    '11111111-1111-1111-1111-111111111111',
    'EPP-ARN-001',
    'Arnés de Cuerpo Completo 4 Anillos con Línea de Vida Doble',
    'fall_arrest',
    'NOM-009-STPS-2011',
    730, 18, TRUE
  ),
  (
    '88888888-1111-1111-1111-000000000005',
    '11111111-1111-1111-1111-111111111111',
    'EPP-GUA-001',
    'Guantes de Carnaza Reforzada para Soldador y Maniobras',
    'hands',
    'NMX-S-040-1987',
    90, 85, TRUE
  ),
  (
    '88888888-1111-1111-1111-000000000006',
    '11111111-1111-1111-1111-111111111111',
    'EPP-RES-001',
    'Respirador Media Cara con Filtros para Vapores Orgánicos',
    'respiratory',
    'NOM-116-STPS-2009',
    120, 24, TRUE
  )
ON CONFLICT (id) DO NOTHING;

-- Entregas activas a trabajadores de contratista
INSERT INTO epp_deliveries (
  id, tenant_id, epp_item_id, worker_id, contractor_id, delivery_date,
  expiration_date, status, delivered_by_user_id, signature_data, notes
) VALUES
  (
    '88888888-2222-1111-1111-000000000001',
    '11111111-1111-1111-1111-111111111111',
    '88888888-1111-1111-1111-000000000001', -- Casco Clase E
    '66666666-6666-6666-6666-666666666661', -- Carlos Mendoza
    '55555555-5555-5555-5555-555555555555',
    CURRENT_DATE - INTERVAL '30 days',
    CURRENT_DATE + INTERVAL '335 days',
    'delivered_active',
    '44444444-4444-4444-4444-444444444443',
    'data:image/png;base64,demo_sig_carlos_mendoza',
    'Entrega por ingreso a proyecto de mantenimiento'
  ),
  (
    '88888888-2222-1111-1111-000000000002',
    '11111111-1111-1111-1111-111111111111',
    '88888888-1111-1111-1111-000000000004', -- Arnés
    '66666666-6666-6666-6666-666666666661', -- Carlos Mendoza
    '55555555-5555-5555-5555-555555555555',
    CURRENT_DATE - INTERVAL '15 days',
    CURRENT_DATE + INTERVAL '715 days',
    'delivered_active',
    '44444444-4444-4444-4444-444444444443',
    'data:image/png;base64,demo_sig_carlos_arnes',
    'Asignado para maniobras en techumbre'
  ),
  (
    '88888888-2222-1111-1111-000000000003',
    '11111111-1111-1111-1111-111111111111',
    '88888888-1111-1111-1111-000000000002', -- Lentes
    '66666666-6666-6666-6666-666666666662', -- Juan Perez
    '55555555-5555-5555-5555-555555555555',
    CURRENT_DATE - INTERVAL '10 days',
    CURRENT_DATE + INTERVAL '170 days',
    'delivered_active',
    '44444444-4444-4444-4444-444444444443',
    'data:image/png;base64,demo_sig_juan_lentes',
    'Reposición por mica desgastada'
  )
ON CONFLICT (id) DO NOTHING;

-- ------------------------------------------------------------
-- 4. SAFETY INCIDENTS & NEAR-MISSES (NOM-021-STPS)
-- ------------------------------------------------------------
INSERT INTO safety_incidents (
  id, tenant_id, incident_number, incident_type, severity, status,
  title, description, location, event_timestamp, reported_by_user_id,
  involved_worker_id, contractor_id, days_lost,
  root_cause_summary, immediate_cause, basic_cause
) VALUES
  (
    '99999999-1111-1111-1111-000000000001',
    '11111111-1111-1111-1111-111111111111',
    'INC-2026-0001',
    'near_miss',
    'moderate',
    'investigating',
    'Caída de herramienta manual desde plataforma de andamio a 4 metros',
    'Durante el ajuste de bridas en torre de enfriamiento, una llave española de 3/4" se deslizó de la mano del operador y cayó al suelo. No hubo personal lesionado debido a que la zona inferior estaba acordonada.',
    'Torre de Enfriamiento 2, Nivel +4.00m',
    now() - INTERVAL '3 days',
    '44444444-4444-4444-4444-444444444443',
    '66666666-6666-6666-6666-666666666661',
    '55555555-5555-5555-5555-555555555555',
    0,
    'Falta de sujeción (lanyard / cuerda de seguridad) en herramientas de mano en trabajos de altura.',
    'Herramienta operada sin muñequera de retención anticaída.',
    'Procedimiento de trabajo seguro en altura no contemplaba checklist específico de amarre de herramienta.'
  ),
  (
    '99999999-1111-1111-1111-000000000002',
    '11111111-1111-1111-1111-111111111111',
    'INC-2026-0002',
    'unsafe_condition',
    'minor',
    'closed',
    'Derrame de líquido refrigerante en pasillo de tránsito peatonal',
    'Fuga leve en manguera de retorno de centro de maquinado CNC 3. El piso presentaba charco resbaloso sin señalización preventiva.',
    'Nave Industrial A - Pasillo Principal de Maquinados',
    now() - INTERVAL '5 days',
    '44444444-4444-4444-4444-444444444444',
    NULL,
    NULL,
    0,
    'Abrazadera de manguera floja por vibración natural de la máquina.',
    'Falla mecánica menor en conducción de refrigerante.',
    'Falta de inspección preventiva semanal en mangueras flexibles.'
  )
ON CONFLICT (id) DO NOTHING;

-- ------------------------------------------------------------
-- 5. MUSTER POINTS & EVACUATION DRILLS (NOM-002-STPS)
-- ------------------------------------------------------------
INSERT INTO muster_points (
  id, tenant_id, code, name, location_description, capacity, is_active
) VALUES
  (
    'aaaaaaaa-1111-1111-1111-000000000001',
    '11111111-1111-1111-1111-111111111111',
    'PR-01',
    'Punto de Reunión 1 - Estacionamiento Principal Norte',
    'Área despejada frente a Caseta 1, libre de tendido eléctrico aéreo y tránsito de vehículos.',
    200, TRUE
  ),
  (
    'aaaaaaaa-1111-1111-1111-000000000002',
    '11111111-1111-1111-1111-111111111111',
    'PR-02',
    'Punto de Reunión 2 - Canchas Deportivas Sur',
    'Zona deportiva y áreas verdes posteriores a Planta de Tratamiento de Aguas.',
    250, TRUE
  )
ON CONFLICT (id) DO NOTHING;

INSERT INTO emergency_evacuations (
  id, tenant_id, evacuation_number, evacuation_type, status, description,
  started_at, ended_at, evacuation_time_seconds, initiated_by_user_id
) VALUES (
  'aaaaaaaa-2222-1111-1111-000000000001',
  '11111111-1111-1111-1111-111111111111',
  'EVAC-2026-0001',
  'drill',
  'all_clear',
  'Simulacro Nacional de Evacuación por Sismo. Desalojo total de naves operativas y oficinas administrativas.',
  now() - INTERVAL '7 days',
  now() - INTERVAL '7 days' + INTERVAL '3 minutes 45 seconds',
  225,
  '44444444-4444-4444-4444-444444444443'
) ON CONFLICT (id) DO NOTHING;

INSERT INTO evacuation_roll_call (
  id, evacuation_id, worker_id, muster_point_id, status, checked_in_at, checked_in_by_user_id, notes
) VALUES
  (
    'aaaaaaaa-3333-1111-1111-000000000001',
    'aaaaaaaa-2222-1111-1111-000000000001',
    '66666666-6666-6666-6666-666666666661',
    'aaaaaaaa-1111-1111-1111-000000000001',
    'safe',
    now() - INTERVAL '7 days' + INTERVAL '1 minute 30 seconds',
    '44444444-4444-4444-4444-444444444443',
    'Presente en tiempo normativo'
  ),
  (
    'aaaaaaaa-3333-1111-1111-000000000002',
    'aaaaaaaa-2222-1111-1111-000000000001',
    '66666666-6666-6666-6666-666666666662',
    'aaaaaaaa-1111-1111-1111-000000000001',
    'safe',
    now() - INTERVAL '7 days' + INTERVAL '2 minutes 10 seconds',
    '44444444-4444-4444-4444-444444444443',
    'Presente sin incidentes'
  )
ON CONFLICT (id) DO NOTHING;

COMMIT;
