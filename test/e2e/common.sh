#!/usr/bin/env bash
# ==============================================================================
# ASESAMO E2E Test Suite — Common Utilities & Harness
# ==============================================================================

set -uo pipefail

# ANSI Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Configurable Test Environment Variables
export API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
export GUARD_EMAIL="${GUARD_EMAIL:-guard@asesamo.com}"
export GUARD_PASSWORD="${GUARD_PASSWORD:-Password123!}"
export TEST_TENANT_ID="${TEST_TENANT_ID:-11111111-1111-1111-1111-111111111111}"
export TEST_PLANT_ID="${TEST_PLANT_ID:-22222222-2222-2222-2222-222222222222}"
export TEST_ACCESS_POINT_ID="${TEST_ACCESS_POINT_ID:-33333333-3333-3333-3333-333333333331}"
export WORKER_COMPLIANT_ID="${WORKER_COMPLIANT_ID:-66666666-6666-6666-6666-666666666661}"
export WORKER_NON_COMPLIANT_ID="${WORKER_NON_COMPLIANT_ID:-66666666-6666-6666-6666-666666666662}"
export WORKER_WARNING_ID="${WORKER_WARNING_ID:-66666666-6666-6666-6666-666666666663}"
export VERBOSE="${VERBOSE:-0}"

# Test Counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

# Cached Auth Token
CACHED_TOKEN=""

# ------------------------------------------------------------------------------
# Logging Functions
# ------------------------------------------------------------------------------
log_header() {
  local title="$1"
  echo -e "\n${BOLD}${CYAN}==============================================================================${NC}"
  echo -e "${BOLD}${CYAN}  ${title}${NC}"
  echo -e "${BOLD}${CYAN}==============================================================================${NC}\n"
}

log_sub_header() {
  local title="$1"
  echo -e "\n${BOLD}${BLUE}--- ${title} ---${NC}"
}

log_info() {
  echo -e "${BLUE}[INFO]${NC} $1"
}

log_pass() {
  TOTAL_TESTS=$((TOTAL_TESTS + 1))
  PASSED_TESTS=$((PASSED_TESTS + 1))
  echo -e "  ${GREEN}✔ PASS:${NC} $1"
}

log_fail() {
  TOTAL_TESTS=$((TOTAL_TESTS + 1))
  FAILED_TESTS=$((FAILED_TESTS + 1))
  echo -e "  ${RED}✖ FAIL:${NC} $1"
  if [ -n "${2:-}" ]; then
    echo -e "         ${MAGENTA}Detail:${NC} $2"
  fi
}

log_skip() {
  TOTAL_TESTS=$((TOTAL_TESTS + 1))
  SKIPPED_TESTS=$((SKIPPED_TESTS + 1))
  echo -e "  ${YELLOW}⊘ SKIP:${NC} $1"
}

print_summary() {
  local suite_name="$1"
  echo -e "\n${BOLD}------------------------------------------------------------------------------${NC}"
  echo -e "${BOLD}Summary for: ${suite_name}${NC}"
  echo -e "  Total Assertions: ${TOTAL_TESTS}"
  echo -e "  ${GREEN}Passed:${NC}           ${PASSED_TESTS}"
  if [ "${FAILED_TESTS}" -gt 0 ]; then
    echo -e "  ${RED}Failed:${NC}           ${FAILED_TESTS}"
  else
    echo -e "  ${GREEN}Failed:${NC}           0"
  fi
  if [ "${SKIPPED_TESTS}" -gt 0 ]; then
    echo -e "  ${YELLOW}Skipped:${NC}          ${SKIPPED_TESTS}"
  fi
  echo -e "${BOLD}------------------------------------------------------------------------------${NC}\n"

  if [ "${FAILED_TESTS}" -gt 0 ]; then
    return 1
  fi
  return 0
}

# ------------------------------------------------------------------------------
# Pre-flight Checks
# ------------------------------------------------------------------------------
check_dependencies() {
  local missing=0
  for cmd in curl jq; do
    if ! command -v "$cmd" &>/dev/null; then
      echo -e "${RED}[ERROR] Required command '${cmd}' is not installed or not in PATH.${NC}"
      missing=1
    fi
  done
  if [ "$missing" -ne 0 ]; then
    exit 2
  fi
}

check_api_connectivity() {
  local url="${API_BASE_URL}"
  # Try both /api/v1/health and /health
  local res
  res=$(curl -s -m 3 "${url}/api/v1/health" || curl -s -m 3 "${url}/health" || true)
  if [ -z "$res" ]; then
    echo -e "${RED}[ERROR] Cannot connect to ASESAMO API at ${url}${NC}"
    echo -e "${YELLOW}Please ensure the backend is running (e.g. docker compose up or go run cmd/api/main.go).${NC}"
    exit 2
  fi
}

# ------------------------------------------------------------------------------
# HTTP Request Utilities
# ------------------------------------------------------------------------------
# Executes curl and outputs status code on the last line, body on preceding lines
# Usage:
#   RESP=$(do_request GET "/api/v1/health" "" "")
#   STATUS=$(echo "$RESP" | tail -n1)
#   BODY=$(echo "$RESP" | sed '$d')
do_request() {
  local method="$1"
  local endpoint="$2"
  local token="${3:-}"
  local data="${4:-}"
  local extra_headers="${5:-}"

  local url="${API_BASE_URL}${endpoint}"
  local curl_args=(-s -w "\n%{http_code}" -X "$method")

  if [ -n "$token" ]; then
    curl_args+=(-H "Authorization: Bearer ${token}")
  fi

  if [ -n "$data" ]; then
    curl_args+=(-H "Content-Type: application/json" -d "$data")
  fi

  if [ -n "$extra_headers" ]; then
    IFS=';' read -ra HEADERS <<< "$extra_headers"
    for h in "${HEADERS[@]}"; do
      curl_args+=(-H "$h")
    done
  fi

  if [ "${VERBOSE}" = "1" ]; then
    echo -e "${BLUE}[CURL]${NC} ${method} ${url}" >&2
    if [ -n "$data" ]; then
      echo -e "${BLUE}[DATA]${NC} ${data}" >&2
    fi
  fi

  curl "${curl_args[@]}" "${url}"
}

# ------------------------------------------------------------------------------
# Authentication Helper
# ------------------------------------------------------------------------------
get_guard_token() {
  if [ -n "${CACHED_TOKEN}" ]; then
    echo "${CACHED_TOKEN}"
    return 0
  fi

  local login_payload
  login_payload=$(cat <<EOF
{
  "email": "${GUARD_EMAIL}",
  "password": "${GUARD_PASSWORD}"
}
EOF
)

  local resp
  resp=$(do_request POST "/api/v1/auth/login" "" "${login_payload}")
  local status
  status=$(echo "$resp" | tail -n1)
  local body
  body=$(echo "$resp" | sed '$d')

  if [ "$status" != "200" ]; then
    echo -e "${RED}[ERROR] Guard login failed with HTTP ${status}: ${body}${NC}" >&2
    return 1
  fi

  local token
  token=$(echo "$body" | jq -r '.data.token // .data.access_token // empty')
  if [ -z "$token" ] || [ "$token" = "null" ]; then
    echo -e "${RED}[ERROR] Failed to extract token from login response: ${body}${NC}" >&2
    return 1
  fi

  CACHED_TOKEN="$token"
  echo "$token"
  return 0
}

# ------------------------------------------------------------------------------
# Assertion Helpers
# ------------------------------------------------------------------------------
assert_equals() {
  local desc="$1"
  local actual="$2"
  local expected="$3"

  if [ "$actual" = "$expected" ]; then
    log_pass "$desc (expected: '${expected}', got: '${actual}')"
    return 0
  else
    log_fail "$desc" "Expected '${expected}', but got '${actual}'"
    return 1
  fi
}

assert_status() {
  local desc="$1"
  local actual_status="$2"
  local expected_status="$3"

  if [ "$actual_status" = "$expected_status" ]; then
    log_pass "$desc (HTTP ${actual_status})"
    return 0
  else
    log_fail "$desc" "Expected HTTP ${expected_status}, but got HTTP ${actual_status}"
    return 1
  fi
}

assert_json_expr() {
  local desc="$1"
  local json_str="$2"
  local jq_expr="$3"

  local result
  result=$(echo "$json_str" | jq -e "$jq_expr" 2>/dev/null || true)
  if [ "$result" = "true" ] || ([ -n "$result" ] && [ "$result" != "false" ] && [ "$result" != "null" ]); then
    log_pass "$desc"
    return 0
  else
    log_fail "$desc" "Condition failed: ${jq_expr}. Response body: ${json_str}"
    return 1
  fi
}

assert_json_field_equals() {
  local desc="$1"
  local json_str="$2"
  local field_path="$3"
  local expected="$4"

  local actual
  actual=$(echo "$json_str" | jq -r "${field_path} // empty" 2>/dev/null || true)
  if [ "$actual" = "$expected" ]; then
    log_pass "$desc (${field_path} == '${expected}')"
    return 0
  else
    log_fail "$desc" "Field '${field_path}' expected '${expected}', got '${actual}'. Response: ${json_str}"
    return 1
  fi
}

assert_json_field_not_empty() {
  local desc="$1"
  local json_str="$2"
  local field_path="$3"

  local val
  val=$(echo "$json_str" | jq -r "${field_path} // empty" 2>/dev/null || true)
  if [ -n "$val" ] && [ "$val" != "null" ]; then
    log_pass "$desc (${field_path} is non-empty: '${val}')"
    return 0
  else
    log_fail "$desc" "Field '${field_path}' is null or empty. Response: ${json_str}"
    return 1
  fi
}
