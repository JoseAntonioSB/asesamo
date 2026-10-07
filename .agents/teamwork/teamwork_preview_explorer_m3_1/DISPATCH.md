## 2026-10-02T05:56:48Z
Milestone M3: Compliance Evaluation Engine & JSONB Snapshots.
Task:
1. Analyze the Check-In Verification Engine:
   - Worker active status (`users.is_active`, `contractor_workers.is_active`).
   - Contractor active status (`contractors.is_active`, `plant_contractors.status = 'active'`).
   - Plant policy: `plants.block_on_expired_docs`.
   - Plant document requirements (`plant_document_requirements`): Which document types are mandatory for this plant?
   - Document evaluation:
     * Check worker documents and contractor documents against each requirement.
     * Document status must be 'approved'.
     * Expiration date: `expires_at > now()`.
     * Warning threshold: expiring within warning window (e.g. <= 15 days).
     * Result determination:
       - If any required document is missing, rejected, or expired (when block_on_expired_docs is true) -> `denied`, with explicit Spanish denial reason.
       - If all required documents are approved and valid, but some expire soon -> `granted_with_warning`.
       - If all required documents are approved, valid, and not expiring soon -> `granted`.
2. Analyze the immutable `access_logs` entry:
   - Columns: `id`, `tenant_id`, `plant_id`, `access_point_id`, `user_id`, `contractor_id`, `processed_by`, `checked_in_at`, `result`, `denial_reason`, `document_snapshot`.
   - `document_snapshot` structure in JSONB: Snapshot of all evaluated documents, requirements, statuses, expiration dates, contractor info, and worker details.
3. Formulate the exact domain entity, usecase logic, and snapshot JSON schema.

Write analysis and proposed code structures to analysis.md and handoff.md in your working directory. Call send_message when done.
