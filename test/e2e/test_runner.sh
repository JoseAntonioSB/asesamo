#!/usr/bin/env bash
# ==============================================================================
# ASESAMO E2E Master Test Runner (test_runner.sh)
# ==============================================================================
# Master orchestrator for ASESAMO Phase 1 MVP End-to-End Testing Suite.
#
# Usage:
#   ./test/e2e/test_runner.sh [OPTIONS]
#
# Options:
#   -t, --tier <1|2|3|4>    Execute a single test tier
#   -a, --all               Execute all tiers in sequence (default)
#   -u, --url <URL>         Set target API base URL (default: http://localhost:8080)
#   -v, --verbose           Enable verbose HTTP request/response debugging
#   -h, --help              Display this help message
#
# Exit Codes:
#   0 = All executed test tiers passed completely
#   1 = One or more test assertions failed
#   2 = Environment / pre-flight failure (e.g. API unreachable, missing dependencies)
# ==============================================================================

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

SELECTED_TIER="all"

usage() {
  echo -e "${BOLD}ASESAMO E2E Test Suite Runner${NC}

${BOLD}USAGE:${NC}
  $0 [OPTIONS]

${BOLD}OPTIONS:${NC}
  -t, --tier <1|2|3|4|5>  Execute only the specified test tier:
                            1: Feature Coverage (Login, Check-in, Access points, Logs)
                            2: Boundary & Corner Cases (Expired JWT, Malformed UUIDs, Isolation)
                            3: Cross-Feature Combinations (Full audit lifecycles)
                            4: Industrial Shift Workload (15 rapid worker arrivals, SLA check)
                            5: White-Box Adversarial Hardening (SQLi, Nil UUID, Concurrency, Isolation)
  -a, --all               Run all 5 tiers in succession (Default)
  -u, --url <URL>         Target API Base URL (Default: ${API_BASE_URL})
  -v, --verbose           Enable verbose curl logging
  -h, --help              Show this help message

${BOLD}EXAMPLES:${NC}
  $0                      # Run full test suite
  $0 -t 1                 # Run only Tier 1
  $0 -t 2 --verbose       # Run Tier 2 with debug output
  API_BASE_URL=\"http://api:8080\" $0
"
  exit 0
}

# Parse Command-Line Arguments
while [[ $# -gt 0 ]]; do
  case "$1" in
    -t|--tier)
      SELECTED_TIER="$2"
      shift 2
      ;;
    -a|--all)
      SELECTED_TIER="all"
      shift
      ;;
    -u|--url)
      export API_BASE_URL="$2"
      shift 2
      ;;
    -v|--verbose)
      export VERBOSE=1
      shift
      ;;
    -h|--help)
      usage
      ;;
    *)
      echo -e "${RED}[ERROR] Unknown option: $1${NC}"
      usage
      ;;
  esac
done

# Pre-flight Checks
log_header "ASESAMO E2E TEST RUNNER — PRE-FLIGHT VERIFICATION"

log_info "Verifying required tools (curl, jq)..."
check_dependencies
log_pass "Tools verified (curl, jq available)"

log_info "Connecting to ASESAMO API at: ${API_BASE_URL}..."
if ! check_api_connectivity 2>/dev/null; then
  echo -e "\n${RED}==============================================================================${NC}"
  echo -e "${RED}  PRE-FLIGHT FAILURE: ASESAMO API IS UNREACHABLE${NC}"
  echo -e "${RED}==============================================================================${NC}"
  echo -e "Could not establish HTTP connection to: ${BOLD}${API_BASE_URL}${NC}"
  echo -e "Please ensure the local development environment or Docker Compose stack is running:"
  echo -e "  ${YELLOW}docker compose up -d${NC}   or   ${YELLOW}cd asesamo-api && go run cmd/api/main.go${NC}\n"
  exit 2
fi
log_pass "API connectivity established at ${API_BASE_URL}"

# Execution Engine
RUNNER_START_TIME=$(date +%s)
OVERALL_EXIT_CODE=0

TIER_NAMES=(
  "Tier 1: Feature Coverage Tests"
  "Tier 2: Boundary & Corner Cases"
  "Tier 3: Cross-Feature Combinations"
  "Tier 4: Industrial Shift Workload"
  "Tier 5: White-Box Adversarial Hardening"
)

TIER_SCRIPTS=(
  "${SCRIPT_DIR}/tier1_feature_test.sh"
  "${SCRIPT_DIR}/tier2_boundary_test.sh"
  "${SCRIPT_DIR}/tier3_combination_test.sh"
  "${SCRIPT_DIR}/tier4_workload_test.sh"
  "${SCRIPT_DIR}/tier5_whitebox_adversarial_probe.sh"
)

TIER_RESULTS=(0 0 0 0 0)
TIER_EXECUTED=(false false false false false)

run_single_tier() {
  local tier_idx="$1" # 0-indexed (0 to 4)
  local tier_num=$((tier_idx + 1))
  local script="${TIER_SCRIPTS[$tier_idx]}"
  local name="${TIER_NAMES[$tier_idx]}"

  TIER_EXECUTED[$tier_idx]=true

  if [ ! -f "$script" ]; then
    echo -e "${RED}[ERROR] Test script not found: ${script}${NC}"
    TIER_RESULTS[$tier_idx]=1
    OVERALL_EXIT_CODE=1
    return 1
  fi

  chmod +x "$script"

  echo -e "\n${BOLD}${MAGENTA}>>> Launching ${name}...${NC}"
  if bash "$script"; then
    TIER_RESULTS[$tier_idx]=0
    echo -e "${GREEN}>>> ${name} PASSED${NC}\n"
  else
    TIER_RESULTS[$tier_idx]=1
    OVERALL_EXIT_CODE=1
    echo -e "${RED}>>> ${name} FAILED${NC}\n"
  fi
}

case "$SELECTED_TIER" in
  1) run_single_tier 0 ;;
  2) run_single_tier 1 ;;
  3) run_single_tier 2 ;;
  4) run_single_tier 3 ;;
  5) run_single_tier 4 ;;
  all)
    for i in 0 1 2 3 4; do
      run_single_tier "$i"
    done
    ;;
  *)
    echo -e "${RED}[ERROR] Invalid tier: ${SELECTED_TIER}. Choose 1, 2, 3, 4, 5, or 'all'.${NC}"
    exit 2
    ;;
esac

RUNNER_END_TIME=$(date +%s)
RUNNER_DURATION=$((RUNNER_END_TIME - RUNNER_START_TIME))

# ==============================================================================
# Final Master Summary Report
# ==============================================================================
echo -e "\n${BOLD}${CYAN}==============================================================================${NC}"
echo -e "${BOLD}${CYAN}  ASESAMO E2E TEST SUITE — MASTER EXECUTION REPORT${NC}"
echo -e "${BOLD}${CYAN}==============================================================================${NC}"
echo -e "  Target API URL:    ${API_BASE_URL}"
echo -e "  Total Duration:    ${RUNNER_DURATION}s"
echo -e "  Executed Tiers:"

for i in 0 1 2 3 4; do
  if [ "${TIER_EXECUTED[$i]}" = true ]; then
    if [ "${TIER_RESULTS[$i]}" -eq 0 ]; then
      echo -e "    ${GREEN}✔ ${TIER_NAMES[$i]}: PASSED${NC}"
    else
      echo -e "    ${RED}✖ ${TIER_NAMES[$i]}: FAILED${NC}"
    fi
  fi
done

echo -e "${BOLD}${CYAN}==============================================================================${NC}"

if [ "$OVERALL_EXIT_CODE" -eq 0 ]; then
  echo -e "${BOLD}${GREEN}✔ ALL EXECUTED TEST TIERS PASSED SUCCESSFULLY! (Exit Code: 0)${NC}\n"
else
  echo -e "${BOLD}${RED}✖ ONE OR MORE TEST TIERS ENCOUNTERED FAILURES. (Exit Code: 1)${NC}\n"
fi

exit "$OVERALL_EXIT_CODE"
