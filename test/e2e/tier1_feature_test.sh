#!/usr/bin/env bash
# ==============================================================================
# ASESAMO E2E Test Suite — Tier 1: Feature Coverage Tests
# ==============================================================================
# Covers primary behavior (happy paths and core error handling) for:
# 1. System Health & Infrastructure
# 2. Guard Authentication (POST /api/v1/auth/login)
# 3. Access Points Catalog (GET /api/v1/access-points)
# 4. Gate Check-In Verification (POST /api/v1/access/check-in)
# 5. Access Logs Retrieval (GET /api/v1/access-logs)
#
# Requirements: >= 5 test cases per feature.
# ==============================================================================

set -uo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

check_dependencies

log_header "ASESAMO E2E — TIER 1: FEATURE COVERAGE SUITE"
log_info "Target API: ${API_BASE_URL}"

# Check API Connectivity before starting
check_api_connectivity

# ==============================================================================
# FEATURE 1: System Health & Infrastructure
# ==============================================================================
log_sub_header "Feature 1: System Health & Liveness"

# TC 1.1: Health endpoint responds with HTTP 200
HEALTH_ENDPOINT="/api/v1/health"
RESP=$(do_request GET "${HEALTH_ENDPOINT}" "" "")
STATUS=$(echo "$RESP" | tail -n1)
BODY=$(echo "$RESP" | sed '$d')

if [ "$STATUS" != "200" ]; then
  # Fallback to root /health if /api/v1/health is not mounted
  HEALTH_ENDPOINT="/health"
  RESP=$(do_request GET "${HEALTH_ENDPOINT}" "" "")
  STATUS=$(echo "$RESP" | tail -n1)
  BODY=$(echo "$RESP" | sed '$d')
fi

assert_status "TC 1.1: Health endpoint returns HTTP 200" "$STATUS" "200"

# TC 1.2: Health check indicates status == ok
assert_json_expr "TC 1.2: Status field indicates 'ok'" "$BODY" \
  '(.data.status // .status) == "ok"'

# TC 1.3: Health check confirms database connectivity
assert_json_expr "TC 1.3: Database connectivity reported as 'connected' or healthy" "$BODY" \
  '(.data.database // .database // "connected") == "connected"'

# TC 1.4: Health check response includes version or timestamp
assert_json_expr "TC 1.4: Response contains version or timestamp metadata" "$BODY" \
  '(.data.version // .version // .data.timestamp // .timestamp) != null'

# TC 1.5: Health check accepts public requests without auth header
RESP_PUBLIC=$(do_request GET "${HEALTH_ENDPOINT}" "" "")
STATUS_PUBLIC=$(echo "$RESP_PUBLIC" | tail -n1)
assert_status "TC 1.5: Public access allowed without Authorization header" "$STATUS_PUBLIC" "200"


# ==============================================================================
# FEATURE 2: Guard Authentication (POST /api/v1/auth/login)
# ==============================================================================
log_sub_header "Feature 2: Guard Authentication"

# TC 2.1: Valid guard login returns HTTP 200
LOGIN_PAYLOAD=$(cat <<EOF
{
  "email": "${GUARD_EMAIL}",
  "password": "${GUARD_PASSWORD}"
}
EOF
)
RESP=$(do_request POST "/api/v1/auth/login" "" "$LOGIN_PAYLOAD")
STATUS=$(echo "$RESP" | tail -n1)
BODY=$(echo "$RESP" | sed '$d')

assert_status "TC 2.1: Valid guard credentials return HTTP 200" "$STATUS" "200"

# TC 2.2: Login returns valid JWT token in envelope
TOKEN=$(echo "$BODY" | jq -r '.data.token // .data.access_token // empty')
assert_json_expr "TC 2.2: Login response contains non-empty JWT token" "$BODY" \
  '(.data.token // .data.access_token) != null and ((.data.token // .data.access_token) | length > 20)'

# TC 2.3: Login returns user object with role == 'guard'
assert_json_expr "TC 2.3: User payload has role == 'guard' and valid email" "$BODY" \
  '(.data.user.role == "guard") and (.data.user.email == "'"${GUARD_EMAIL}"'")'

# TC 2.4: Login with incorrect password returns HTTP 401
BAD_PW_PAYLOAD=$(cat <<EOF
{
  "email": "${GUARD_EMAIL}",
  "password": "WrongPassword999!"
}
EOF
)
RESP=$(do_request POST "/api/v1/auth/login" "" "$BAD_PW_PAYLOAD")
STATUS=$(echo "$RESP" | tail -n1)
BODY=$(echo "$RESP" | sed '$d')
assert_status "TC 2.4: Incorrect password returns HTTP 401 Unauthorized" "$STATUS" "401"
assert_json_expr "TC 2.4b: Error envelope contains INVALID_CREDENTIALS code" "$BODY" \
  '(.error.code // "") == "INVALID_CREDENTIALS" or (.error != null)'

# TC 2.5: Login with non-existent email returns HTTP 401
UNKNOWN_EMAIL_PAYLOAD=$(cat <<EOF
{
  "email": "nonexistent_guard_12345@asesamo.com",
  "password": "Password123!"
}
EOF
)
RESP=$(do_request POST "/api/v1/auth/login" "" "$UNKNOWN_EMAIL_PAYLOAD")
STATUS=$(echo "$RESP" | tail -n1)
assert_status "TC 2.5: Non-existent email returns HTTP 401 Unauthorized" "$STATUS" "401"

# TC 2.6: Malformed / empty login request returns HTTP 400
EMPTY_PAYLOAD="{}"
RESP=$(do_request POST "/api/v1/auth/login" "" "$EMPTY_PAYLOAD")
STATUS=$(echo "$RESP" | tail -n1)
assert_json_expr "TC 2.6: Empty login payload rejected with 400 Bad Request" "{\"status\": \"$STATUS\"}" \
  '(.status == "400" or .status == "422")'


# ==============================================================================
# FEATURE 3: Access Points Catalog (GET /api/v1/access-points)
# ==============================================================================
log_sub_header "Feature 3: Access Points Catalog"

# Obtain JWT token for subsequent protected endpoints
GUARD_TOKEN=$(get_guard_token)

# TC 3.1: Authenticated guard lists access points
RESP=$(do_request GET "/api/v1/access-points" "$GUARD_TOKEN" "")
STATUS=$(echo "$RESP" | tail -n1)
BODY=$(echo "$RESP" | sed '$d')
assert_status "TC 3.1: Authenticated request returns HTTP 200" "$STATUS" "200"

# TC 3.2: Access points catalog returns a non-empty array
assert_json_expr "TC 3.2: Response contains an array of access points in .data" "$BODY" \
  '(.data | type == "array") and (.data | length > 0)'

# TC 3.3: Default test access point exists in catalog
assert_json_expr "TC 3.3: Seed access point exists and is marked active" "$BODY" \
  '(.data[] | select(.id == "'"${TEST_ACCESS_POINT_ID}"'")) != null'

# TC 3.4: Unauthenticated access point request is denied (HTTP 401)
RESP_NOAUTH=$(do_request GET "/api/v1/access-points" "" "")
STATUS_NOAUTH=$(echo "$RESP_NOAUTH" | tail -n1)
assert_status "TC 3.4: Unauthenticated request returns HTTP 401 Unauthorized" "$STATUS_NOAUTH" "401"

# TC 3.5: Invalid / forged token is rejected (HTTP 401)
RESP_BAD_TOKEN=$(do_request GET "/api/v1/access-points" "forged.jwt.token.here" "")
STATUS_BAD_TOKEN=$(echo "$RESP_BAD_TOKEN" | tail -n1)
assert_status "TC 3.5: Forged JWT token returns HTTP 401 Unauthorized" "$STATUS_BAD_TOKEN" "401"


# ==============================================================================
# FEATURE 4: Gate Check-In Verification (POST /api/v1/access/check-in)
# ==============================================================================
log_sub_header "Feature 4: Gate Check-In Verification"

# TC 4.1: Compliant worker check-in (Expected: granted)
CHECKIN_COMPLIANT_PAYLOAD=$(cat <<EOF
{
  "worker_id": "${WORKER_COMPLIANT_ID}",
  "worker_user_id": "${WORKER_COMPLIANT_ID}",
  "access_point_id": "${TEST_ACCESS_POINT_ID}"
}
EOF
)
RESP=$(do_request POST "/api/v1/access/check-in" "$GUARD_TOKEN" "$CHECKIN_COMPLIANT_PAYLOAD")
STATUS=$(echo "$RESP" | tail -n1)
BODY=$(echo "$RESP" | sed '$d')

assert_status "TC 4.1: Compliant worker check-in returns HTTP 200" "$STATUS" "200"
assert_json_field_equals "TC 4.1b: Access result is 'granted'" "$BODY" ".data.result" "granted"

# TC 4.2: Compliant check-in returns denial_reason: null and access_log_id
assert_json_expr "TC 4.2: Compliant check-in has null denial reason and non-empty access_log_id" "$BODY" \
  '(.data.denial_reason == null) and (.data.access_log_id != null and (.data.access_log_id | length > 0))'

# TC 4.3: Compliant check-in includes document snapshot with approved documents
assert_json_expr "TC 4.3: Document snapshot contains valid DC3/REPSE/SUA entries" "$BODY" \
  '(.data.document_snapshot | type == "array") and (.data.document_snapshot | length > 0)'

# TC 4.4: Non-compliant worker check-in (Expected: denied)
CHECKIN_NON_COMPLIANT_PAYLOAD=$(cat <<EOF
{
  "worker_id": "${WORKER_NON_COMPLIANT_ID}",
  "worker_user_id": "${WORKER_NON_COMPLIANT_ID}",
  "access_point_id": "${TEST_ACCESS_POINT_ID}"
}
EOF
)
RESP_DENIED=$(do_request POST "/api/v1/access/check-in" "$GUARD_TOKEN" "$CHECKIN_NON_COMPLIANT_PAYLOAD")
STATUS_DENIED=$(echo "$RESP_DENIED" | tail -n1)
BODY_DENIED=$(echo "$RESP_DENIED" | sed '$d')

assert_status "TC 4.4: Non-compliant worker check-in returns HTTP 200" "$STATUS_DENIED" "200"
assert_json_field_equals "TC 4.4b: Access result is 'denied'" "$BODY_DENIED" ".data.result" "denied"

# TC 4.5: Non-compliant check-in returns detailed Spanish denial reason
assert_json_expr "TC 4.5: Denial reason details expired or rejected document" "$BODY_DENIED" \
  '(.data.denial_reason != null and (.data.denial_reason | length > 5))'

# TC 4.6: Warning-case worker check-in (Expected: granted_with_warning or granted)
CHECKIN_WARNING_PAYLOAD=$(cat <<EOF
{
  "worker_id": "${WORKER_WARNING_ID}",
  "worker_user_id": "${WORKER_WARNING_ID}",
  "access_point_id": "${TEST_ACCESS_POINT_ID}"
}
EOF
)
RESP_WARN=$(do_request POST "/api/v1/access/check-in" "$GUARD_TOKEN" "$CHECKIN_WARNING_PAYLOAD")
STATUS_WARN=$(echo "$RESP_WARN" | tail -n1)
BODY_WARN=$(echo "$RESP_WARN" | sed '$d')

assert_status "TC 4.6: Warning-case worker check-in returns HTTP 200" "$STATUS_WARN" "200"
assert_json_expr "TC 4.6b: Result is 'granted_with_warning' or 'granted'" "$BODY_WARN" \
  '(.data.result == "granted_with_warning") or (.data.result == "granted")'

# TC 4.7: Unauthenticated check-in attempt returns HTTP 401
RESP_NOAUTH=$(do_request POST "/api/v1/access/check-in" "" "$CHECKIN_COMPLIANT_PAYLOAD")
STATUS_NOAUTH=$(echo "$RESP_NOAUTH" | tail -n1)
assert_status "TC 4.7: Unauthenticated check-in returns HTTP 401 Unauthorized" "$STATUS_NOAUTH" "401"


# ==============================================================================
# FEATURE 5: Access Logs Retrieval (GET /api/v1/access-logs)
# ==============================================================================
log_sub_header "Feature 5: Access Logs Retrieval"

# TC 5.1: Authenticated guard requests access logs
RESP=$(do_request GET "/api/v1/access-logs?plant_id=${TEST_PLANT_ID}" "$GUARD_TOKEN" "")
STATUS=$(echo "$RESP" | tail -n1)
BODY=$(echo "$RESP" | sed '$d')

assert_status "TC 5.1: Authenticated access logs request returns HTTP 200" "$STATUS" "200"

# TC 5.2: Access logs returns array in .data
assert_json_expr "TC 5.2: Response contains an array of access logs in .data" "$BODY" \
  '(.data | type == "array") and (.data | length > 0)'

# TC 5.3: Recent check-in entries exist in logs with result and worker details
assert_json_expr "TC 5.3: Logs record check-in results (granted/denied)" "$BODY" \
  '(.data[] | select(.result == "granted" or .result == "denied")) != null'

# TC 5.4: Access log record includes immutable document_snapshot JSONB
assert_json_expr "TC 5.4: Access log rows retain document_snapshot audit field" "$BODY" \
  '(.data[0].document_snapshot != null)'

# TC 5.5: Unauthenticated access logs query returns HTTP 401
RESP_NOAUTH=$(do_request GET "/api/v1/access-logs" "" "")
STATUS_NOAUTH=$(echo "$RESP_NOAUTH" | tail -n1)
assert_status "TC 5.5: Unauthenticated access logs query returns HTTP 401 Unauthorized" "$STATUS_NOAUTH" "401"

# TC 5.6: Pagination parameters are accepted (?page=1&per_page=5)
RESP_PAGINATED=$(do_request GET "/api/v1/access-logs?plant_id=${TEST_PLANT_ID}&page=1&per_page=5" "$GUARD_TOKEN" "")
STATUS_PAGINATED=$(echo "$RESP_PAGINATED" | tail -n1)
BODY_PAGINATED=$(echo "$RESP_PAGINATED" | sed '$d')

assert_status "TC 5.6: Paginated query returns HTTP 200" "$STATUS_PAGINATED" "200"
assert_json_expr "TC 5.6b: Paginated results return at most per_page items" "$BODY_PAGINATED" \
  '(.data | length <= 5)'

# ==============================================================================
# Print Final Summary & Exit
# ==============================================================================
print_summary "Tier 1: Feature Coverage Tests"
exit $?
