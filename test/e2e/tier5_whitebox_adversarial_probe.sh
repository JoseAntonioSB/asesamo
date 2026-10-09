#!/usr/bin/env bash
# ==============================================================================
# ASESAMO E2E Test Suite — Tier 5: White-Box Adversarial Coverage Hardening
# ==============================================================================
# Comprehensive white-box adversarial verification probing edge cases,
# payload boundaries, SQL injection resilience, path traversals, nil UUIDs,
# cross-tenant isolation, concurrent check-in races, and repeat checkouts.
# ==============================================================================

set -uo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

check_dependencies

log_header "ASESAMO E2E — TIER 5: WHITE-BOX ADVERSARIAL COVERAGE HARDENING"
log_info "Target API: ${API_BASE_URL}"

check_api_connectivity

GUARD_TOKEN=$(get_guard_token)
JWT_SECRET="${JWT_SECRET:-super_secret_jwt_key_min_32_characters_long_12345}"

# Helper to synthesize custom JWTs
mint_jwt() {
  local user_id="$1"
  local tenant_id="$2"
  local role="$3"
  local exp_offset="${4:-3600}"
  local secret="${5:-$JWT_SECRET}"
  local alg="${6:-HS256}"

  python3 -c "
import hmac, hashlib, base64, json, time

secret = '${secret}'
alg = '${alg}'
header_data = {'alg': alg, 'typ': 'JWT'}
header = base64.urlsafe_b64encode(json.dumps(header_data).encode()).decode().rstrip('=')

payload_data = {
    'sub': '${user_id}',
    'tenant_id': '${tenant_id}',
    'role': '${role}',
    'iat': int(time.time()),
    'exp': int(time.time()) + int('${exp_offset}')
}
payload = base64.urlsafe_b64encode(json.dumps(payload_data).encode()).decode().rstrip('=')

if alg == 'none':
    print(f'{header}.{payload}.')
else:
    sig = base64.urlsafe_b64encode(hmac.new(secret.encode(), f'{header}.{payload}'.encode(), hashlib.sha256).digest()).decode().rstrip('=')
    print(f'{header}.{payload}.{sig}')
"
}

# ==============================================================================
# SCENARIO 1: Extreme Payload Boundaries & Parsing Resilience
# ==============================================================================
log_sub_header "Scenario 1: Extreme Payload Boundaries & Body Parsing"

# TC 5.1: Completely empty request body to POST /api/v1/auth/login
RESP=$(curl -s -w "\n%{http_code}" -X POST "${API_BASE_URL}/api/v1/auth/login" \
  -H "Content-Type: application/json" --data-binary "")
STATUS=$(echo "$RESP" | tail -n1)
BODY=$(echo "$RESP" | sed '$d')
assert_status "TC 5.1: Empty body on auth login returns HTTP 400" "$STATUS" "400"
assert_json_expr "TC 5.1b: Error envelope has INVALID_INPUT" "$BODY" '(.error.code == "INVALID_INPUT")'

# TC 5.2: Empty JSON object {} on auth login
RESP=$(do_request POST "/api/v1/auth/login" "" "{}")
STATUS=$(echo "$RESP" | tail -n1)
assert_status "TC 5.2: Empty JSON {} on auth login returns HTTP 400" "$STATUS" "400"

# TC 5.3: Completely empty request body to POST /api/v1/access/check-in
RESP=$(curl -s -w "\n%{http_code}" -X POST "${API_BASE_URL}/api/v1/access/check-in" \
  -H "Authorization: Bearer ${GUARD_TOKEN}" \
  -H "Content-Type: application/json" --data-binary "")
STATUS=$(echo "$RESP" | tail -n1)
assert_status "TC 5.3: Empty body on check-in returns HTTP 400" "$STATUS" "400"

# TC 5.4: Truncated / malformed JSON on check-in
RESP=$(curl -s -w "\n%{http_code}" -X POST "${API_BASE_URL}/api/v1/access/check-in" \
  -H "Authorization: Bearer ${GUARD_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"worker_user_id":"66666666-6666-6666-6666-666666666661",')
STATUS=$(echo "$RESP" | tail -n1)
assert_status "TC 5.4: Malformed truncated JSON returns HTTP 400" "$STATUS" "400"

# TC 5.5: Deeply nested JSON structure in body
NESTED_PAYLOAD='{"worker_user_id":"66666666-6666-6666-6666-666666666661","access_point_id":"33333333-3333-3333-3333-333333333331","notes":{"nested":{"level1":{"level2":{"level3":true}}}}}'
RESP=$(do_request POST "/api/v1/access/check-in" "$GUARD_TOKEN" "$NESTED_PAYLOAD")
STATUS=$(echo "$RESP" | tail -n1)
assert_json_expr "TC 5.5: Deeply nested non-string notes rejected or handled safely (400/422)" "{\"status\":\"$STATUS\"}" \
  '(.status == "400" or .status == "422")'

# TC 5.6: Oversized payload (100 KB payload with long dummy notes)
OVERSIZED_NOTES=$(python3 -c "print('X' * 100000)")
OVERSIZED_PAYLOAD=$(cat <<EOF
{
  "worker_user_id": "${WORKER_COMPLIANT_ID}",
  "access_point_id": "${TEST_ACCESS_POINT_ID}",
  "notes": "${OVERSIZED_NOTES}"
}
EOF
)
RESP=$(do_request POST "/api/v1/access/check-in" "$GUARD_TOKEN" "$OVERSIZED_PAYLOAD")
STATUS=$(echo "$RESP" | tail -n1)
assert_json_expr "TC 5.6: 100KB payload handled without server crash (200, 400, or 413)" "{\"status\":\"$STATUS\"}" \
  '(.status == "200" or .status == "400" or .status == "413")'

# TC 5.7: Special characters, unicode, emojis, and quotes in notes
SPECIAL_NOTES='Aviso: Trabajador ingresa con equipo soldador 👨‍🏭⚡ & "casco" especial <OK>!'
SPECIAL_PAYLOAD=$(jq -n \
  --arg w "$WORKER_COMPLIANT_ID" \
  --arg ap "$TEST_ACCESS_POINT_ID" \
  --arg n "$SPECIAL_NOTES" \
  '{worker_user_id: $w, access_point_id: $ap, notes: $n}')
RESP=$(do_request POST "/api/v1/access/check-in" "$GUARD_TOKEN" "$SPECIAL_PAYLOAD")
STATUS=$(echo "$RESP" | tail -n1)
BODY=$(echo "$RESP" | sed '$d')
assert_status "TC 5.7: Unicode and emoji notes accepted with HTTP 200" "$STATUS" "200"
assert_json_field_equals "TC 5.7b: Check-in result is granted" "$BODY" ".data.result" "granted"

# TC 5.8: Content-Type text/plain gracefully rejected with 400
RESP=$(curl -s -w "\n%{http_code}" -X POST "${API_BASE_URL}/api/v1/access/check-in" \
  -H "Authorization: Bearer ${GUARD_TOKEN}" \
  -H "Content-Type: text/plain" \
  -d 'worker_user_id=66666666-6666-6666-6666-666666666661')
STATUS=$(echo "$RESP" | tail -n1)
assert_status "TC 5.8: Non-JSON Content-Type text/plain returns HTTP 400" "$STATUS" "400"

# TC 5.9: Schema pollution / unexpected fields ignored safely
POLLUTED_PAYLOAD=$(cat <<EOF
{
  "worker_user_id": "${WORKER_COMPLIANT_ID}",
  "access_point_id": "${TEST_ACCESS_POINT_ID}",
  "isAdmin": true,
  "role": "super_admin",
  "result": "granted",
  "denial_reason": null
}
EOF
)
RESP=$(do_request POST "/api/v1/access/check-in" "$GUARD_TOKEN" "$POLLUTED_PAYLOAD")
STATUS=$(echo "$RESP" | tail -n1)
assert_status "TC 5.9: Extra fields in body handled cleanly without corruption (HTTP 200)" "$STATUS" "200"


# ==============================================================================
# SCENARIO 2: SQL Injection Vectors & Parameter Escaping
# ==============================================================================
log_sub_header "Scenario 2: SQL Injection Vectors & Parameter Hardening"

# TC 5.10: SQL injection bypass attempt on login
SQLI_AUTH_PAYLOAD='{"email":"'\'' OR '\''1'\''='\''1'\'' --","password":"Password123!"}'
RESP=$(do_request POST "/api/v1/auth/login" "" "$SQLI_AUTH_PAYLOAD")
STATUS=$(echo "$RESP" | tail -n1)
assert_status "TC 5.10: SQL injection on login email returns HTTP 401" "$STATUS" "401"

# TC 5.11: SQL injection stacked query on login
STACKED_AUTH_PAYLOAD='{"email":"guard@asesamo.com'\''; DROP TABLE users; --","password":"Password123!"}'
RESP=$(do_request POST "/api/v1/auth/login" "" "$STACKED_AUTH_PAYLOAD")
STATUS=$(echo "$RESP" | tail -n1)
assert_status "TC 5.11: Stacked SQL injection on login returns HTTP 401" "$STATUS" "401"

# TC 5.12: SQL injection in check-in notes field
SQLI_NOTES_PAYLOAD=$(cat <<EOF
{
  "worker_user_id": "${WORKER_COMPLIANT_ID}",
  "access_point_id": "${TEST_ACCESS_POINT_ID}",
  "notes": "'; DROP TABLE access_logs; --"
}
EOF
)
RESP=$(do_request POST "/api/v1/access/check-in" "$GUARD_TOKEN" "$SQLI_NOTES_PAYLOAD")
STATUS=$(echo "$RESP" | tail -n1)
assert_status "TC 5.12: SQL injection in notes stored safely as string (HTTP 200)" "$STATUS" "200"

# Verify access_logs table was not dropped
RESP_CHECK=$(do_request GET "/api/v1/access-logs?per_page=1" "$GUARD_TOKEN" "")
STATUS_CHECK=$(echo "$RESP_CHECK" | tail -n1)
assert_status "TC 5.12b: access_logs table intact after SQLi test (HTTP 200)" "$STATUS_CHECK" "200"

# TC 5.13: SQL injection vectors in GET /api/v1/access-logs parameters
SQLI_PARAMS=(
  "plant_id=%27+OR+%271%27%3D%271"
  "access_point_id=%27%3B+DROP+TABLE+access_logs%3B+--"
  "worker_user_id=%27+UNION+SELECT+NULL--"
  "result=granted%27+OR+1%3D1--"
  "from=2026-01-01%27+UNION+SELECT+NULL--"
  "to=2026-12-31%27+OR+1%3D1--"
)
for param in "${SQLI_PARAMS[@]}"; do
  RESP=$(do_request GET "/api/v1/access-logs?${param}" "$GUARD_TOKEN" "")
  STATUS=$(echo "$RESP" | tail -n1)
  assert_status "TC 5.13: SQLi query parameter '${param}' rejected with HTTP 400" "$STATUS" "400"
done

# TC 5.14: Negative and massive pagination parameters
RESP=$(do_request GET "/api/v1/access-logs?page=-99&per_page=999999999" "$GUARD_TOKEN" "")
STATUS=$(echo "$RESP" | tail -n1)
BODY=$(echo "$RESP" | sed '$d')
assert_status "TC 5.14: Out-of-bounds pagination returns HTTP 200" "$STATUS" "200"
assert_json_expr "TC 5.14b: Pagination capped safely at per_page <= 100" "$BODY" \
  '(.meta.per_page <= 100 and .meta.page >= 1)'


# ==============================================================================
# SCENARIO 3: Path Traversals & Route Integrity
# ==============================================================================
log_sub_header "Scenario 3: Path Traversals & Route Integrity"

# TC 5.15: Path traversal attempt in URL
RESP=$(do_request GET "/api/v1/access/../../../../etc/passwd" "$GUARD_TOKEN" "")
STATUS=$(echo "$RESP" | tail -n1)
assert_json_expr "TC 5.15: Path traversal in URL yields 404 or 400" "{\"status\":\"$STATUS\"}" \
  '(.status == "404" or .status == "400")'

# TC 5.16: Unsupported HTTP methods
RESP=$(do_request PUT "/api/v1/auth/login" "" '{"email":"guard@asesamo.com","password":"Password123!"}')
STATUS=$(echo "$RESP" | tail -n1)
assert_status "TC 5.16: PUT on auth login returns HTTP 405 Method Not Allowed" "$STATUS" "405"

RESP=$(do_request DELETE "/api/v1/access/check-in" "$GUARD_TOKEN" "")
STATUS=$(echo "$RESP" | tail -n1)
assert_status "TC 5.16b: DELETE on access check-in returns HTTP 405 Method Not Allowed" "$STATUS" "405"

# TC 5.17: Trailing slash request normalization
RESP=$(do_request GET "/api/v1/health/" "" "")
STATUS=$(echo "$RESP" | tail -n1)
assert_json_expr "TC 5.17: Health with trailing slash handled gracefully (200 or 404)" "{\"status\":\"$STATUS\"}" \
  '(.status == "200" or .status == "404")'


# ==============================================================================
# SCENARIO 4: Nil UUID & Nonexistent Entity Probing
# ==============================================================================
log_sub_header "Scenario 4: Nil UUID & Nonexistent Entity Verification"

# TC 5.18: All-zeroes UUID for worker_user_id
NIL_WORKER_PAYLOAD=$(cat <<EOF
{
  "worker_user_id": "00000000-0000-0000-0000-000000000000",
  "access_point_id": "${TEST_ACCESS_POINT_ID}"
}
EOF
)
RESP=$(do_request POST "/api/v1/access/check-in" "$GUARD_TOKEN" "$NIL_WORKER_PAYLOAD")
STATUS=$(echo "$RESP" | tail -n1)
BODY=$(echo "$RESP" | sed '$d')
assert_status "TC 5.18: Nil UUID worker returns HTTP 404 WORKER_NOT_FOUND" "$STATUS" "404"
assert_json_expr "TC 5.18b: Error code is WORKER_NOT_FOUND" "$BODY" '(.error.code == "WORKER_NOT_FOUND")'

# TC 5.19: All-zeroes UUID for access_point_id
NIL_AP_PAYLOAD=$(cat <<EOF
{
  "worker_user_id": "${WORKER_COMPLIANT_ID}",
  "access_point_id": "00000000-0000-0000-0000-000000000000"
}
EOF
)
RESP=$(do_request POST "/api/v1/access/check-in" "$GUARD_TOKEN" "$NIL_AP_PAYLOAD")
STATUS=$(echo "$RESP" | tail -n1)
BODY=$(echo "$RESP" | sed '$d')
assert_status "TC 5.19: Nil UUID access point returns HTTP 404 ACCESS_POINT_NOT_FOUND" "$STATUS" "404"
assert_json_expr "TC 5.19b: Error code is ACCESS_POINT_NOT_FOUND" "$BODY" '(.error.code == "ACCESS_POINT_NOT_FOUND")'

# TC 5.20: Nonexistent valid UUID for worker_user_id
NONEXISTENT_WORKER_PAYLOAD=$(cat <<EOF
{
  "worker_user_id": "ffffffff-ffff-ffff-ffff-ffffffffffff",
  "access_point_id": "${TEST_ACCESS_POINT_ID}"
}
EOF
)
RESP=$(do_request POST "/api/v1/access/check-in" "$GUARD_TOKEN" "$NONEXISTENT_WORKER_PAYLOAD")
STATUS=$(echo "$RESP" | tail -n1)
BODY=$(echo "$RESP" | sed '$d')
assert_status "TC 5.20: Nonexistent valid UUID worker returns HTTP 404" "$STATUS" "404"
assert_json_expr "TC 5.20b: Error code is WORKER_NOT_FOUND" "$BODY" '(.error.code == "WORKER_NOT_FOUND")'

# TC 5.21: Nonexistent valid UUID for access_point_id
NONEXISTENT_AP_PAYLOAD=$(cat <<EOF
{
  "worker_user_id": "${WORKER_COMPLIANT_ID}",
  "access_point_id": "ffffffff-ffff-ffff-ffff-ffffffffffff"
}
EOF
)
RESP=$(do_request POST "/api/v1/access/check-in" "$GUARD_TOKEN" "$NONEXISTENT_AP_PAYLOAD")
STATUS=$(echo "$RESP" | tail -n1)
BODY=$(echo "$RESP" | sed '$d')
assert_status "TC 5.21: Nonexistent valid UUID access point returns HTTP 404" "$STATUS" "404"
assert_json_expr "TC 5.21b: Error code is ACCESS_POINT_NOT_FOUND" "$BODY" '(.error.code == "ACCESS_POINT_NOT_FOUND")'

# TC 5.22: Empty string UUIDs
EMPTY_UUID_PAYLOAD='{"worker_user_id":"","access_point_id":""}'
RESP=$(do_request POST "/api/v1/access/check-in" "$GUARD_TOKEN" "$EMPTY_UUID_PAYLOAD")
STATUS=$(echo "$RESP" | tail -n1)
assert_status "TC 5.22: Empty string UUIDs return HTTP 400" "$STATUS" "400"

# TC 5.23: Explicit null values for UUIDs
NULL_UUID_PAYLOAD='{"worker_user_id":null,"access_point_id":null}'
RESP=$(do_request POST "/api/v1/access/check-in" "$GUARD_TOKEN" "$NULL_UUID_PAYLOAD")
STATUS=$(echo "$RESP" | tail -n1)
assert_status "TC 5.23: Null UUIDs return HTTP 400" "$STATUS" "400"


# ==============================================================================
# SCENARIO 5: Multi-Tenancy & Cross-Tenant Leak Probing
# ==============================================================================
log_sub_header "Scenario 5: Multi-Tenancy & Cross-Tenant Boundary Probing"

FOREIGN_TENANT_ID="99999999-9999-9999-9999-999999999999"
FOREIGN_USER_ID="88888888-8888-8888-8888-888888888888"

FOREIGN_TOKEN=$(mint_jwt "$FOREIGN_USER_ID" "$FOREIGN_TENANT_ID" "guard")

# TC 5.24: Foreign tenant requesting access points catalog
RESP=$(do_request GET "/api/v1/access-points" "$FOREIGN_TOKEN" "")
STATUS=$(echo "$RESP" | tail -n1)
BODY=$(echo "$RESP" | sed '$d')
assert_status "TC 5.24: Foreign tenant gets HTTP 200" "$STATUS" "200"
assert_json_expr "TC 5.24b: Zero access points leaked across tenant boundary" "$BODY" \
  '(.data | length == 0)'

# TC 5.25: Foreign tenant attempting check-in at Tenant 1's Caseta
RESP=$(do_request POST "/api/v1/access/check-in" "$FOREIGN_TOKEN" "$(cat <<EOF
{
  "worker_user_id": "${WORKER_COMPLIANT_ID}",
  "access_point_id": "${TEST_ACCESS_POINT_ID}"
}
EOF
)")
STATUS=$(echo "$RESP" | tail -n1)
BODY=$(echo "$RESP" | sed '$d')
assert_status "TC 5.25: Cross-tenant check-in rejected with HTTP 404" "$STATUS" "404"
assert_json_expr "TC 5.25b: Cross-tenant access point not found" "$BODY" \
  '(.error.code == "ACCESS_POINT_NOT_FOUND")'

# TC 5.26: Foreign tenant querying access logs
RESP=$(do_request GET "/api/v1/access-logs" "$FOREIGN_TOKEN" "")
STATUS=$(echo "$RESP" | tail -n1)
BODY=$(echo "$RESP" | sed '$d')
assert_status "TC 5.26: Foreign tenant access logs query returns HTTP 200" "$STATUS" "200"
assert_json_expr "TC 5.26b: Zero access logs leaked across tenant boundary" "$BODY" \
  '((.data | length) == 0 and .meta.total == 0)'

# TC 5.27: Token with Nil Tenant UUID (00000000-0000-0000-0000-000000000000)
NIL_TENANT_TOKEN=$(mint_jwt "$FOREIGN_USER_ID" "00000000-0000-0000-0000-000000000000" "guard")
RESP=$(do_request GET "/api/v1/access-points" "$NIL_TENANT_TOKEN" "")
STATUS=$(echo "$RESP" | tail -n1)
assert_status "TC 5.27: Nil tenant token fails closed with HTTP 401" "$STATUS" "401"

# TC 5.28: Alg 'none' token (unsigned bypass attempt)
ALG_NONE_TOKEN=$(mint_jwt "$FOREIGN_USER_ID" "$TEST_TENANT_ID" "guard" 3600 "$JWT_SECRET" "none")
RESP=$(do_request GET "/api/v1/access-points" "$ALG_NONE_TOKEN" "")
STATUS=$(echo "$RESP" | tail -n1)
assert_status "TC 5.28: Alg 'none' token rejected with HTTP 401" "$STATUS" "401"

# TC 5.29: Token signed with incorrect secret
WRONG_SECRET_TOKEN=$(mint_jwt "$FOREIGN_USER_ID" "$TEST_TENANT_ID" "guard" 3600 "wrong_secret_123456789012345678901234567890")
RESP=$(do_request GET "/api/v1/access-points" "$WRONG_SECRET_TOKEN" "")
STATUS=$(echo "$RESP" | tail -n1)
assert_status "TC 5.29: Token with wrong secret rejected with HTTP 401" "$STATUS" "401"


# ==============================================================================
# SCENARIO 6: Concurrency, Race Conditions & Rapid Checkouts
# ==============================================================================
log_sub_header "Scenario 6: Concurrency, Race Conditions & Rapid Checkouts"

# TC 5.30: 25 Concurrent Check-Ins for the same compliant worker
CONCURRENT_COUNT=25
log_info "Launching ${CONCURRENT_COUNT} parallel check-in requests for Worker 1..."

CONCURRENT_TMP=$(mktemp -d)
CHECKIN_BODY=$(cat <<EOF
{
  "worker_user_id": "${WORKER_COMPLIANT_ID}",
  "access_point_id": "${TEST_ACCESS_POINT_ID}",
  "notes": "Concurrent stress check-in probe"
}
EOF
)

for i in $(seq 1 $CONCURRENT_COUNT); do
  (
    RESP=$(do_request POST "/api/v1/access/check-in" "$GUARD_TOKEN" "$CHECKIN_BODY")
    STATUS=$(echo "$RESP" | tail -n1)
    echo "$STATUS" > "${CONCURRENT_TMP}/status_${i}.txt"
  ) &
done

wait

SUCCESS_COUNT=0
for i in $(seq 1 $CONCURRENT_COUNT); do
  if [ -f "${CONCURRENT_TMP}/status_${i}.txt" ]; then
    CODE=$(cat "${CONCURRENT_TMP}/status_${i}.txt")
    if [ "$CODE" = "200" ]; then
      SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
    fi
  fi
done
rm -rf "$CONCURRENT_TMP"

assert_equals "TC 5.30: All ${CONCURRENT_COUNT} concurrent check-ins succeeded" "$SUCCESS_COUNT" "$CONCURRENT_COUNT"

# TC 5.31: Repeat Checkouts
# A: Checkout when never checked in
NONEXISTENT_USER="ffffffff-ffff-ffff-ffff-ffffffffffff"
RESP=$(do_request POST "/api/v1/access/check-out" "$GUARD_TOKEN" "$(cat <<EOF
{
  "worker_user_id": "${NONEXISTENT_USER}",
  "access_point_id": "${TEST_ACCESS_POINT_ID}"
}
EOF
)")
STATUS=$(echo "$RESP" | tail -n1)
assert_status "TC 5.31a: Checkout for worker without prior check-in returns HTTP 404" "$STATUS" "404"

# B: First Checkout for Compliant Worker
RESP=$(do_request POST "/api/v1/access/check-out" "$GUARD_TOKEN" "$(cat <<EOF
{
  "worker_user_id": "${WORKER_COMPLIANT_ID}",
  "access_point_id": "${TEST_ACCESS_POINT_ID}"
}
EOF
)")
STATUS=$(echo "$RESP" | tail -n1)
assert_status "TC 5.31b: First checkout returns HTTP 200" "$STATUS" "200"

# C: Immediate Repeat Checkout for the same Compliant Worker
RESP=$(do_request POST "/api/v1/access/check-out" "$GUARD_TOKEN" "$(cat <<EOF
{
  "worker_user_id": "${WORKER_COMPLIANT_ID}",
  "access_point_id": "${TEST_ACCESS_POINT_ID}"
}
EOF
)")
STATUS=$(echo "$RESP" | tail -n1)
assert_status "TC 5.31c: Immediate repeat checkout returns HTTP 200 (idempotent fallback)" "$STATUS" "200"

# TC 5.32: Mixed Concurrent Check-in and Check-out Operations
MIXED_TMP=$(mktemp -d)
log_info "Launching 10 concurrent check-ins + 10 concurrent check-outs..."

for i in $(seq 1 10); do
  (
    RESP=$(do_request POST "/api/v1/access/check-in" "$GUARD_TOKEN" "$CHECKIN_BODY")
    STATUS=$(echo "$RESP" | tail -n1)
    echo "$STATUS" > "${MIXED_TMP}/checkin_${i}.txt"
  ) &
  (
    RESP=$(do_request POST "/api/v1/access/check-out" "$GUARD_TOKEN" "$(cat <<EOF
{
  "worker_user_id": "${WORKER_COMPLIANT_ID}",
  "access_point_id": "${TEST_ACCESS_POINT_ID}"
}
EOF
)")
    STATUS=$(echo "$RESP" | tail -n1)
    echo "$STATUS" > "${MIXED_TMP}/checkout_${i}.txt"
  ) &
done

wait

MIXED_PASS=0
for i in $(seq 1 10); do
  CI_CODE=$(cat "${MIXED_TMP}/checkin_${i}.txt" 2>/dev/null || echo "000")
  CO_CODE=$(cat "${MIXED_TMP}/checkout_${i}.txt" 2>/dev/null || echo "000")
  if [ "$CI_CODE" = "200" ] && [ "$CO_CODE" = "200" ]; then
    MIXED_PASS=$((MIXED_PASS + 1))
  fi
done
rm -rf "$MIXED_TMP"

assert_equals "TC 5.32: All 10 parallel check-in/check-out pairs succeeded with HTTP 200" "$MIXED_PASS" "10"


# ==============================================================================
# SCENARIO 7: Rate Limiting & Denial of Service Protection
# ==============================================================================
log_sub_header "Scenario 7: Rate Limiting & Brute Force Protection"

# TC 5.33: Verify live rate limiting triggers HTTP 429, emits localized error envelope,
# and keeps unthrottled localhost IP accessible.
# Chi router applies httprate limiter (100 req/min in dev, 10 req/min in prod per IP) on POST /api/v1/auth/login.
# We execute a rapid burst using an isolated synthetic test IP (via X-Forwarded-For) so localhost and subsequent tests are never throttled.
log_info "Verifying live rate limiting and standard HTTP 429 error response envelope..."

RATE_TEST_IP="203.0.113.$(( (RANDOM % 200) + 10 ))"
RATE_LIMIT_RESULT=$(python3 -c "
import urllib.request, urllib.error, json, sys

url = '${API_BASE_URL}/api/v1/auth/login'
test_ip = '${RATE_TEST_IP}'
headers = {'Content-Type': 'application/json', 'X-Forwarded-For': test_ip}
data = b'{}'

got_429 = False
body_429 = {}
attempts = 0

for i in range(110):
    attempts += 1
    req = urllib.request.Request(url, data=data, headers=headers, method='POST')
    try:
        with urllib.request.urlopen(req) as resp:
            pass
    except urllib.error.HTTPError as e:
        if e.code == 429:
            got_429 = True
            try:
                body_429 = json.loads(e.read().decode('utf-8'))
            except Exception:
                pass
            break

err_code = body_429.get('error', {}).get('code', '')
err_msg = body_429.get('error', {}).get('message', '')

# Verify clean request from localhost (no X-Forwarded-For) remains unthrottled
clean_empty_req = urllib.request.Request(
    url,
    data=b'{}',
    headers={'Content-Type': 'application/json'},
    method='POST'
)
clean_empty_status = 0
try:
    with urllib.request.urlopen(clean_empty_req) as resp:
        clean_empty_status = resp.getcode()
except urllib.error.HTTPError as e:
    clean_empty_status = e.code

clean_guard_req = urllib.request.Request(
    url,
    data=json.dumps({'email': '${GUARD_EMAIL}', 'password': '${GUARD_PASSWORD}'}).encode('utf-8'),
    headers={'Content-Type': 'application/json'},
    method='POST'
)
clean_guard_status = 0
try:
    with urllib.request.urlopen(clean_guard_req) as resp:
        clean_guard_status = resp.getcode()
except urllib.error.HTTPError as e:
    clean_guard_status = e.code

if not got_429:
    print(f'FAIL: Did not receive HTTP 429 after {attempts} requests')
elif err_code != 'TOO_MANY_REQUESTS':
    print(f'FAIL: Expected .error.code == TOO_MANY_REQUESTS, got {err_code}')
elif err_msg != 'Demasiadas solicitudes. Intente más tarde.':
    print(f'FAIL: Expected .error.message == \"Demasiadas solicitudes. Intente más tarde.\", got \"{err_msg}\"')
elif clean_empty_status == 429 or clean_guard_status == 429:
    print(f'FAIL: Clean localhost request was throttled (empty status: {clean_empty_status}, guard status: {clean_guard_status})')
elif clean_empty_status != 400 or clean_guard_status != 200:
    print(f'FAIL: Unexpected clean localhost statuses (empty: {clean_empty_status}, guard: {clean_guard_status})')
else:
    print('PASS')
")

assert_equals "TC 5.33: Rate limiting emits HTTP 429 with standard error envelope and leaves localhost unthrottled" "$RATE_LIMIT_RESULT" "PASS"

# ==============================================================================
# Print Final Summary & Exit
# ==============================================================================
print_summary "Tier 5: White-Box Adversarial Coverage Hardening"
exit $?
