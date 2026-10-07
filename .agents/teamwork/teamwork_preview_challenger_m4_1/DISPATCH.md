## 2026-10-07T16:50:18Z
You are teamwork_preview_challenger_m4_1.
Your working directory is: /Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_challenger_m4_1

MANDATORY: You MUST read /Users/jasbotello/devel/asesamo/.agents/teamwork/ORIGINAL_REQUEST.md before starting work.
Also read:
- /Users/jasbotello/devel/asesamo/PROJECT.md
- /Users/jasbotello/devel/asesamo/DEVELOPMENT_RULES.md
- /Users/jasbotello/devel/asesamo/.agents/teamwork/teamwork_preview_worker_m4_1/handoff.md
- Code in /Users/jasbotello/devel/asesamo/asesamo-web/

Milestone M4 Gate Adversarial Challenge: Functional & Interaction Stress.
Adversarially challenge and stress-test the web application:
1. Route protection:
   - Verify Edge middleware redirects unauthenticated users from `/caseta` to `/login`.
   - Verify Edge middleware redirects authenticated users from `/login` to `/caseta`.
   - Verify `/` redirects properly.
2. API client resilience:
   - Test envelope unwrapping and typed error throwing.
   - Verify error mapping for `INVALID_CREDENTIALS`, `TOO_MANY_REQUESTS`, and network timeouts.
3. Scanner & Presets:
   - Verify `ScannerInput` keyboard barcode handling (Enter submits without reload, clears buffer, retains focus).
   - Verify quick-fill presets match seed records (`...6661` Granted, `...6662` Denied, `...6663` Warning).
   - Verify 3-tier banners render correct colors and text.
4. Write test harness scripts or assertions, execute them, and document results in handoff.md. State your final verdict explicitly as APPROVE or REJECT. Call send_message when done.
