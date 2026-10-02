## 2026-10-01T22:19:25Z
You are teamwork_preview_explorer_m1_2.
Your working directory is: /Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_explorer_m1_2

MANDATORY: You MUST read /Users/jasbotello/devel/asesamo/.agents/teamwork/ORIGINAL_REQUEST.md before starting work.
Also read:
- /Users/jasbotello/devel/asesamo/PROJECT.md
- /Users/jasbotello/devel/asesamo/DEVELOPMENT_RULES.md
- /Users/jasbotello/devel/asesamo/db/migrations/000001_init.up.sql

Milestone M1: Database & Environment Scaffolding.
Your task:
Formulate the complete, exact SQL seed script: `db/seeds/000001_seed_dev_data.sql`.
Requirements:
- Must use native PostgreSQL `pgcrypto` function: `crypt('Password123!', gen_salt('bf', 10))` for guard password hash.
- Must insert deterministic UUIDs for:
  * Tenant: `11111111-1111-1111-1111-111111111111` ("Grupo Industrial Norteño SA de CV")
  * Plant: `22222222-2222-2222-2222-222222222222` ("Planta Querétaro Norte", block_on_expired_docs = true)
  * Access Point: `33333333-3333-3333-3333-333333333331` ("Caseta Principal Norte")
  * Guard user: `44444444-4444-4444-4444-444444444444` (`guard@asesamo.com`, role = 'guard', password = Password123!)
  * Contractor company: `55555555-5555-5555-5555-555555555555` ("Mantenimiento Industrial del Norte SA de CV", status active at plant)
  * Worker 1 (Compliant -> granted): `66666666-6666-6666-6666-666666666661` ("Carlos Mendoza", approved DC3, approved INE, approved Medical)
  * Worker 2 (Non-compliant -> denied): `66666666-6666-6666-6666-666666666662` ("Juan Perez", expired DC3, rejected Medical)
  * Worker 3 (Warning -> granted_with_warning): `66666666-6666-6666-6666-666666666663` ("Roberto Garcia", approved DC3 expiring in 5 days)
  * Contractor company required documents (REPSE, SUA approved).
- Script must be idempotent (`ON CONFLICT DO NOTHING`).
Write your analysis and exact SQL recommendations to your working directory in analysis.md and handoff.md. Call send_message when done.
