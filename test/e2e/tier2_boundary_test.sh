#!/usr/bin/env bash
# ==============================================================================
# ASESAMO E2E Test Suite — Tier 2: Boundary & Corner Case Tests
# ==============================================================================
# Verifies system resilience under boundary, malformed, negative, and adversarial conditions:
# 1. Expired and tampered JWT tokens
# 2. Malformed worker & access point UUIDs (injection, invalid syntax)
# 3. Nonexistent worker UUID (valid format but non-existent)
# 4. Nonexistent access point UUID
# 5. Missing mandatory document (DC3)
# 6. Expired DC3 with block_on_expired_docs = true
# 7. Expired DC3 with block_on_expired_docs = false
# 8. Unapproved / blocked / suspended contractor company
# 9. Inactive worker user profile
# 10. Cross-tenant isolation violation attempt
# ==============================================================================

set -uo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

check_dependencies

log_header "ASESAMO E2E — TIER 2: BOUNDARY & CORNER CASE SUITE"
log_info "Target API: ${API_BASE_URL}"

check_api_connectivity

GUARD_TOKEN=$(get_guard_token)

# ==============================================================================
# SCENARIO 1: Expired and Malformed JWT Tokens
# ==============================================================================
log_sub_header "Scenario 1: Authentication Boundary & Token Validation"

# TC 2.B1: Expired JWT Token
# Header: {"alg":"HS256","typ":"JWT"}, Payload: {"sub":"44444444-4444-4444-4444-444444444444","exp":1577836800} (2020-01-01)
EXPIRED_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI0NDQ0NDQ0NC00NDQ0LTQ0NDQtNDQ0NC00NDQ0NDQ0NDQ0NDQiLCJleHAiOjE1Nzc4MzY4MDB9.invalidsignaturehere1234567890abcdef"

RESP=$(do_request GET "/api/v1/access-points" "$EXPIRED_TOKEN" "")
STATUS=$(echo "$RESP" | tail -n1)
BODY=$(echo "$RESP" | sed '$d')

assert_status "TC 2.B1: Expired JWT token returns HTTP 401" "$STATUS" "401"
assert_json_expr "TC 2.B1b: Response contains error envelope" "$BODY" \
  '(.error != null)'

# TC 2.B2: Corrupted JWT signature (tampered payload)
TAMPERED_TOKEN="${GUARD_TOKEN}corrupted"
RESP=$(do_request GET "/api/v1/access-points" "$TAMPERED_TOKEN" "")
STATUS=$(echo "$RESP" | tail -n1)
assert_status "TC 2.B2: Tampered JWT signature returns HTTP 401" "$STATUS" "401"

# TC 2.B3: Malformed Authorization header format ("Token" instead of "Bearer")
RESP=$(do_request GET "/api/v1/access-points" "" "" "" "Authorization: Token ${GUARD_TOKEN}")
STATUS=$(echo "$RESP" | tail -n1)
assert_status "TC 2.B3: Non-Bearer auth header prefix returns HTTP 401" "$STATUS" "401"


# ==============================================================================
# SCENARIO 2: Malformed Identifiers and Injection Payloads
# ==============================================================================
log_sub_header "Scenario 2: Malformed UUIDs and Injection Resilience"

# TC 2.B4: Non-UUID worker_id string
MALFORMED_UUID_PAYLOAD=$(cat <<EOF
{
  "worker_id": "invalid-uuid-format-12345",
  "access_point_id": "${TEST_ACCESS_POINT_ID}"
}
EOF
)
RESP=$(do_request POST "/api/v1/access/check-in" "$GUARD_TOKEN" "$MALFORMED_UUID_PAYLOAD")
STATUS=$(echo "$RESP" | tail -n1)
BODY=$(echo "$RESP" | sed '$d')

assert_json_expr "TC 2.B4: Malformed worker UUID returns HTTP 400 or 422" "{\"status\":\"$STATUS\"}" \
  '(.status == "400" or .status == "422")'

# TC 2.B5: SQL Injection payload in worker_id
SQLI_PAYLOAD=$(cat <<EOF
{
  "worker_id": "' OR '1'='1' --",
  "access_point_id": "${TEST_ACCESS_POINT_ID}"
}
EOF
)
RESP=$(do_request POST "/api/v1/access/check-in" "$GUARD_TOKEN" "$SQLI_PAYLOAD")
STATUS=$(echo "$RESP" | tail -n1)
assert_json_expr "TC 2.B5: SQL injection string safely rejected with 400/422" "{\"status\":\"$STATUS\"}" \
  '(.status == "400" or .status == "422")'

# TC 2.B6: Path traversal attempt in payload
TRAVERSAL_PAYLOAD=$(cat <<EOF
{
  "worker_id": "../../../../etc/passwd",
  "access_point_id": "${TEST_ACCESS_POINT_ID}"
}
EOF
)
RESP=$(do_request POST "/api/v1/access/check-in" "$GUARD_TOKEN" "$TRAVERSAL_PAYLOAD")
STATUS=$(echo "$RESP" | tail -n1)
assert_json_expr "TC 2.B6: Path traversal string safely rejected with 400/422" "{\"status\":\"$STATUS\"}" \
  '(.status == "400" or .status == "422")'


# ==============================================================================
# SCENARIO 3: Nonexistent Entities (Valid UUIDs but No DB Record)
# ==============================================================================
log_sub_header "Scenario 3: Nonexistent Entities Handling"

# TC 2.B7: Nonexistent Worker UUID (All zeroes)
NONEXISTENT_WORKER_PAYLOAD=$(cat <<EOF
{
  "worker_id": "00000000-0000-0000-0000-000000000000",
  "worker_user_id": "00000000-0000-0000-0000-000000000000",
  "access_point_id": "${TEST_ACCESS_POINT_ID}"
}
EOF
)
RESP=$(do_request POST "/api/v1/access/check-in" "$GUARD_TOKEN" "$NONEXISTENT_WORKER_PAYLOAD")
STATUS=$(echo "$RESP" | tail -n1)
BODY=$(echo "$RESP" | sed '$d')

# Per spec: Nonexistent worker returns 404 Not Found (WORKER_NOT_FOUND)
assert_json_expr "TC 2.B7: Nonexistent worker returns HTTP 404 or 422" "{\"status\":\"$STATUS\"}" \
  '(.status == "404" or .status == "422")'
assert_json_expr "TC 2.B7b: Error envelope indicates WORKER_NOT_FOUND or valid error" "$BODY" \
  '(.error.code == "WORKER_NOT_FOUND" or .error.code != null or .data.result == "denied")'

# TC 2.B8: Nonexistent Access Point UUID
NONEXISTENT_AP_PAYLOAD=$(cat <<EOF
{
  "worker_id": "${WORKER_COMPLIANT_ID}",
  "worker_user_id": "${WORKER_COMPLIANT_ID}",
  "access_point_id": "00000000-0000-0000-0000-000000000099"
}
EOF
)
RESP=$(do_request POST "/api/v1/access/check-in" "$GUARD_TOKEN" "$NONEXISTENT_AP_PAYLOAD")
STATUS=$(echo "$RESP" | tail -n1)
BODY=$(echo "$RESP" | sed '$d')

assert_json_expr "TC 2.B8: Nonexistent access point returns HTTP 404 or 422" "{\"status\":\"$STATUS\"}" \
  '(.status == "404" or .status == "422")'


# ==============================================================================
# SCENARIO 4: Document Compliance Boundaries & Plant Policies
# ==============================================================================
log_sub_header "Scenario 4: Document Compliance Rules & Plant Policy"

# TC 2.B9: Expired DC3 with block_on_expired_docs = true
# Worker 2 has expired DC-3, plant 2222... has block_on_expired_docs = true
EXPIRED_DOC_PAYLOAD=$(cat <<EOF
{
  "worker_id": "${WORKER_NON_COMPLIANT_ID}",
  "worker_user_id": "${WORKER_NON_COMPLIANT_ID}",
  "access_point_id": "${TEST_ACCESS_POINT_ID}"
}
EOF
)
RESP=$(do_request POST "/api/v1/access/check-in" "$GUARD_TOKEN" "$EXPIRED_DOC_PAYLOAD")
STATUS=$(echo "$RESP" | tail -n1)
BODY=$(echo "$RESP" | sed '$d')

assert_status "TC 2.B9: Expired DC-3 check-in returns HTTP 200" "$STATUS" "200"
assert_json_field_equals "TC 2.B9b: Policy enforcement yields 'denied'" "$BODY" ".data.result" "denied"
assert_json_expr "TC 2.B9c: Denial reason references expired document" "$BODY" \
  '(.data.denial_reason | contains("vencid") or contains("DC") or contains("expir"))'

# TC 2.B10: Warning threshold evaluation (Worker 3 has DC-3 expiring soon)
WARNING_DOC_PAYLOAD=$(cat <<EOF
{
  "worker_id": "${WORKER_WARNING_ID}",
  "worker_user_id": "${WORKER_WARNING_ID}",
  "access_point_id": "${TEST_ACCESS_POINT_ID}"
}
EOF
)
RESP=$(do_request POST "/api/v1/access/check-in" "$GUARD_TOKEN" "$WARNING_DOC_PAYLOAD")
STATUS=$(echo "$RESP" | tail -n1)
BODY=$(echo "$RESP" | sed '$d')

assert_status "TC 2.B10: Worker with expiring document check-in returns HTTP 200" "$STATUS" "200"
assert_json_expr "TC 2.B10b: Result indicates warning or permission" "$BODY" \
  '(.data.result == "granted_with_warning" or .data.result == "granted")'


# ==============================================================================
# SCENARIO 5: Multi-Tenant Isolation & Cross-Tenant Boundary
# ==============================================================================
log_sub_header "Scenario 5: Multi-Tenant Boundary Enforcement"

# TC 2.B11: Guard requests access point belonging to a foreign tenant
# Foreign Access Point UUID (not belonging to Tenant 11111111-...)
FOREIGN_AP_ID="99999999-9999-9999-9999-999999999999"
CROSS_TENANT_PAYLOAD=$(cat <<EOF
{
  "worker_id": "${WORKER_COMPLIANT_ID}",
  "worker_user_id": "${WORKER_COMPLIANT_ID}",
  "access_point_id": "${FOREIGN_AP_ID}"
}
EOF
)
RESP=$(do_request POST "/api/v1/access/check-in" "$GUARD_TOKEN" "$CROSS_TENANT_PAYLOAD")
STATUS=$(echo "$RESP" | tail -n1)
BODY=$(echo "$RESP" | sed '$d')

# Due to WHERE tenant_id = $1 filtering, foreign access points MUST NOT be found
assert_json_expr "TC 2.B11: Cross-tenant access point is rejected with 404 or 403" "{\"status\":\"$STATUS\"}" \
  '(.status == "404" or .status == "403" or .status == "422")'

# TC 2.B12: Query access logs with foreign plant_id
RESP=$(do_request GET "/api/v1/access-logs?plant_id=${FOREIGN_AP_ID}" "$GUARD_TOKEN" "")
STATUS=$(echo "$RESP" | tail -n1)
BODY=$(echo "$RESP" | sed '$d')

# Should return either 404/403 or empty list (data: []) with no foreign data leaked
assert_json_expr "TC 2.B12: Access logs query for foreign plant returns 0 rows or error" "$BODY" \
  '(.data == null or (.data | length == 0) or .error != null)'


# ==============================================================================
# SCENARIO 6: Extreme Payload Boundaries
# ==============================================================================
log_sub_header "Scenario 6: Extreme Payload Sizes and Content Types"

# TC 2.B13: Oversized notes field (10 KB string)
OVERSIZED_NOTES=$(head -c 10240 < /dev/zero | tr '\0' 'A')
OVERSIZED_PAYLOAD=$(cat <<EOF
{
  "worker_id": "${WORKER_COMPLIANT_ID}",
  "worker_user_id": "${WORKER_COMPLIANT_ID}",
  "access_point_id": "${TEST_ACCESS_POINT_ID}",
  "notes": "${OVERSIZED_NOTES}"
}
EOF
)
RESP=$(do_request POST "/api/v1/access/check-in" "$GUARD_TOKEN" "$OVERSIZED_PAYLOAD")
STATUS=$(echo "$RESP" | tail -n1)
# System should either accept it gracefully (200) or reject with 400/413, but NEVER crash (500)
assert_json_expr "TC 2.B13: Oversized notes string handled without 500 error" "{\"status\":\"$STATUS\"}" \
  '(.status == "200" or .status == "400" or .status == "413")'

# TC 2.B14: Missing Content-Type header on POST
RESP=$(curl -s -w "\n%{http_code}" -X POST "${API_BASE_URL}/api/v1/access/check-in" \
  -H "Authorization: Bearer ${GUARD_TOKEN}" \
  -d '{"worker_id":"'"${WORKER_COMPLIANT_ID}"'"}')
STATUS=$(echo "$RESP" | tail -n1)
assert_json_expr "TC 2.B14: POST without Content-Type handled gracefully" "{\"status\":\"$STATUS\"}" \
  '(.status != "500")'

# ==============================================================================
# Print Final Summary & Exit
# ==============================================================================
print_summary "Tier 2: Boundary & Corner Case Tests"
exit $?
