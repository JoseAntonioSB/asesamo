## 2026-10-01T22:15:51Z

You are teamwork_preview_spec_miner_survey_1.
Your working directory is: /Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_spec_miner_survey_1

MANDATORY: You MUST read /Users/jasbotello/devel/asesamo/.agents/teamwork/ORIGINAL_REQUEST.md before starting work.
Also read:
- /Users/jasbotello/devel/asesamo/DEVELOPMENT_RULES.md
- /Users/jasbotello/devel/asesamo/db/migrations/000001_init.up.sql

Your task:
Conduct a comprehensive backend specification survey for Phase 1 MVP of ASESAMO.
Extract and document:
1. All required backend endpoints (including POST /api/v1/auth/login, POST /api/v1/access/check-in, health check, etc.), exact request/response envelopes, status codes, error codes/messages in Spanish per rules.
2. Clean Architecture layer structure for asesamo-api (domain/entity, domain/repository, usecase, adapter/http, adapter/postgres with sqlc, infra/config, database, etc.).
3. Database schema mappings, sqlc queries needed, migration runner requirements.
4. Business logic for access point check-in verification:
   - Worker lookup and active status
   - Contractor lookup, relationship to plant (plant_contractors status)
   - Document evaluation against plant policy (plant_document_requirements, block_on_expired_docs, required documents, expiry dates)
   - Immutable access_logs creation with JSONB document_snapshot
   - Return status: granted, granted_with_warning, denied
5. Multi-tenancy and tenant isolation enforcement (tenant.FromContext, WHERE tenant_id = $1).
6. JWT authentication rules, claims (sub, tenant_id, role, exp), bcrypt password hashing, guard login flow.
7. Unit and integration testing requirements (mocked interfaces for unit tests, isolated DB tests for integration).

Write your detailed findings to:
/Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_spec_miner_survey_1/survey_report.md
Also write handoff.md in your working directory.
When finished, call send_message to report completion.
