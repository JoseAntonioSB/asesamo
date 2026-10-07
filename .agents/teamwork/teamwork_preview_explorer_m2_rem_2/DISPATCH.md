## 2026-10-02T05:10:06Z
Milestone M2 Remediation: Auth Security, Timing Side-Channel & Rate Limiting.
Task:
1. Analyze `LoginUseCase.Execute` in `login.go`:
   - Timing Disparity (CWE-208): If user does not exist, comparing no password takes ~50µs whereas existing user bcrypt compare takes ~80ms (8,000x difference). Formulate dummy bcrypt hash comparison on user not found (e.g. comparing password against a valid bcrypt dummy hash) so timing is indistinguishable.
   - Account Status Enumeration: Currently checks `!user.IsActive` before `bcrypt.CompareHashAndPassword`, allowing unauthenticated attackers to test if an email exists and is inactive without knowing the password. Formulate fix: verify password first, or return generic invalid credentials error.
2. Analyze `router.go`:
   - Rate Limiting: Mount `httprate` middleware (`github.com/go-chi/httprate`) on `/api/v1/auth/login` (e.g. 10 requests per minute per IP per DEVELOPMENT_RULES.md line 323). Check how `httprate` is configured in Chi.
3. Formulate the exact code changes and patch specifications.

Write your findings and patch proposals to analysis.md and handoff.md in your working directory. Call send_message when done.
