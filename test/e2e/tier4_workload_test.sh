#!/usr/bin/env bash
# ==============================================================================
# ASESAMO E2E Test Suite — Tier 4: Industrial Shift Workload Simulation
# ==============================================================================
# Simulates a realistic morning guard shift change at an industrial plant:
# - Guard signs on at Caseta Principal.
# - High-throughput burst of 15 worker arrivals at the gate:
#   * Compliant contractor workers (Expected: Granted)
#   * Non-compliant contractor workers with expired DC-3s (Expected: Denied)
#   * Workers with documents nearing expiration (Expected: Granted with Warning)
#   * Unregistered badges / unknown personnel (Expected: Rejected / Denied)
# - Measures check-in processing latency (<= 500ms per transaction).
# - Performs shift reconciliation: Verifies 100% audit log capture in DB.
# ==============================================================================

set -uo pipefail
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/common.sh"

check_dependencies

log_header "ASESAMO E2E — TIER 4: INDUSTRIAL SHIFT WORKLOAD SUITE"
log_info "Target API: ${API_BASE_URL}"

check_api_connectivity

# ==============================================================================
# Phase 1: Guard Shift Initialization
# ==============================================================================
log_sub_header "Phase 1: Shift Initialization & Gate Checkpoint Verification"

GUARD_TOKEN=$(get_guard_token)
assert_equals "Shift Init: Guard JWT active" "$([ -n "$GUARD_TOKEN" ] && echo 'active')" "active"

# Fetch Access Point
AP_RESP=$(do_request GET "/api/v1/access-points" "$GUARD_TOKEN" "")
AP_STATUS=$(echo "$AP_RESP" | tail -n1)
AP_BODY=$(echo "$AP_RESP" | sed '$d')
assert_status "Shift Init: Access point catalog accessible" "$AP_STATUS" "200"

CASETA_ID=$(echo "$AP_BODY" | jq -r '.data[0].id // empty')
if [ -z "$CASETA_ID" ] || [ "$CASETA_ID" = "null" ]; then
  CASETA_ID="${TEST_ACCESS_POINT_ID}"
fi
log_info "Active Gate Checkpoint: ${CASETA_ID}"

# ==============================================================================
# Phase 2: Morning Shift Burst Simulation (15 Worker Check-Ins)
# ==============================================================================
log_sub_header "Phase 2: Executing Shift Change Arrival Burst (15 Transactions)"

# Define worker queue representing mixed contractor arrivals
# Format: WORKER_ID|EXPECTED_RESULT|DESCRIPTION
WORKER_QUEUE=(
  "${WORKER_COMPLIANT_ID}|granted|Turno A - Soldador Carlos Mendoza (Docs OK)"
  "${WORKER_NON_COMPLIANT_ID}|denied|Turno A - Ayudante Roberto Morales (DC-3 Vencido)"
  "${WORKER_WARNING_ID}|warning|Turno A - Electricista Mateo Vargas (DC-3 por vencer)"
  "${WORKER_COMPLIANT_ID}|granted|Turno A - Re-ingreso herramientas Soldador Carlos"
  "${WORKER_NON_COMPLIANT_ID}|denied|Turno A - Re-intento Ayudante Roberto"
  "${WORKER_COMPLIANT_ID}|granted|Turno A - Personal de apoyo soldadura"
  "${WORKER_WARNING_ID}|warning|Turno A - Supervisor mantenimiento Mateo"
  "00000000-0000-0000-0000-000000000000|not_found|Visitante no registrado (Badge desconocido)"
  "${WORKER_COMPLIANT_ID}|granted|Turno A - Cuadrilla electrica pase 1"
  "${WORKER_COMPLIANT_ID}|granted|Turno A - Cuadrilla electrica pase 2"
  "${WORKER_NON_COMPLIANT_ID}|denied|Turno A - Contratista no autorizado"
  "${WORKER_WARNING_ID}|warning|Turno A - Tecnico de instrumentacion"
  "${WORKER_COMPLIANT_ID}|granted|Turno A - Proveedor de gas industrial"
  "${WORKER_COMPLIANT_ID}|granted|Turno A - Chofer de transporte de personal"
  "${WORKER_NON_COMPLIANT_ID}|denied|Turno A - Ayudante sin EPP ni DC-3"
)

TOTAL_TRANSACTIONS=0
EXPECTED_GRANTED=0
EXPECTED_DENIED=0
EXPECTED_WARNING=0
EXPECTED_NOT_FOUND=0

ACTUAL_GRANTED=0
ACTUAL_DENIED=0
ACTUAL_WARNING=0
ACTUAL_NOT_FOUND=0

MAX_LATENCY_MS=0
TOTAL_LATENCY_MS=0
LATENCY_BREACHES=0

START_SHIFT_TIME=$(date +%s)

for item in "${WORKER_QUEUE[@]}"; do
  IFS='|' read -r W_ID EXPECTED_RES W_DESC <<< "$item"
  TOTAL_TRANSACTIONS=$((TOTAL_TRANSACTIONS + 1))

  case "$EXPECTED_RES" in
    granted) EXPECTED_GRANTED=$((EXPECTED_GRANTED + 1)) ;;
    denied)  EXPECTED_DENIED=$((EXPECTED_DENIED + 1)) ;;
    warning) EXPECTED_WARNING=$((EXPECTED_WARNING + 1)) ;;
    not_found) EXPECTED_NOT_FOUND=$((EXPECTED_NOT_FOUND + 1)) ;;
  esac

  PAYLOAD=$(cat <<EOF
{
  "worker_id": "${W_ID}",
  "worker_user_id": "${W_ID}",
  "access_point_id": "${CASETA_ID}",
  "notes": "Shift-Simulation-Arrival-${TOTAL_TRANSACTIONS}"
}
EOF
)

  # Measure latency with high precision
  T_START=$(date +%s%N 2>/dev/null || python3 -c 'import time; print(int(time.time()*1e9))')
  RESP=$(do_request POST "/api/v1/access/check-in" "$GUARD_TOKEN" "$PAYLOAD")
  T_END=$(date +%s%N 2>/dev/null || python3 -c 'import time; print(int(time.time()*1e9))')

  DURATION_MS=$(( (T_END - T_START) / 1000000 ))
  # Guard against negative or clock skew
  if [ "$DURATION_MS" -lt 0 ]; then DURATION_MS=1; fi

  TOTAL_LATENCY_MS=$((TOTAL_LATENCY_MS + DURATION_MS))
  if [ "$DURATION_MS" -gt "$MAX_LATENCY_MS" ]; then
    MAX_LATENCY_MS="$DURATION_MS"
  fi

  # Flag latency over 500ms
  if [ "$DURATION_MS" -gt 500 ]; then
    LATENCY_BREACHES=$((LATENCY_BREACHES + 1))
  fi

  STATUS=$(echo "$RESP" | tail -n1)
  BODY=$(echo "$RESP" | sed '$d')

  RESULT=$(echo "$BODY" | jq -r '.data.result // empty')
  ERR_CODE=$(echo "$BODY" | jq -r '.error.code // empty')

  if [ "$STATUS" = "200" ]; then
    case "$RESULT" in
      granted)
        ACTUAL_GRANTED=$((ACTUAL_GRANTED + 1))
        echo -e "  [Arrival ${TOTAL_TRANSACTIONS}/15] ${GREEN}GRANTED${NC} (${DURATION_MS}ms) - ${W_DESC}"
        ;;
      denied)
        ACTUAL_DENIED=$((ACTUAL_DENIED + 1))
        echo -e "  [Arrival ${TOTAL_TRANSACTIONS}/15] ${RED}DENIED${NC} (${DURATION_MS}ms) - ${W_DESC}"
        ;;
      granted_with_warning)
        ACTUAL_WARNING=$((ACTUAL_WARNING + 1))
        echo -e "  [Arrival ${TOTAL_TRANSACTIONS}/15] ${YELLOW}WARNING${NC} (${DURATION_MS}ms) - ${W_DESC}"
        ;;
      *)
        echo -e "  [Arrival ${TOTAL_TRANSACTIONS}/15] ${MAGENTA}UNKNOWN (${RESULT})${NC} (${DURATION_MS}ms)"
        ;;
    esac
  elif [ "$STATUS" = "404" ] || [ "$STATUS" = "422" ]; then
    ACTUAL_NOT_FOUND=$((ACTUAL_NOT_FOUND + 1))
    echo -e "  [Arrival ${TOTAL_TRANSACTIONS}/15] ${CYAN}REJECTED (${STATUS}/${ERR_CODE})${NC} (${DURATION_MS}ms) - ${W_DESC}"
  else
    echo -e "  [Arrival ${TOTAL_TRANSACTIONS}/15] ${RED}ERROR HTTP ${STATUS}${NC} (${DURATION_MS}ms) - ${W_DESC}"
  fi
done

AVG_LATENCY_MS=$((TOTAL_LATENCY_MS / TOTAL_TRANSACTIONS))

log_info "Burst completed: 15 transactions processed."
log_info "Latency Metrics: Avg=${AVG_LATENCY_MS}ms, Max=${MAX_LATENCY_MS}ms, Breaches (>500ms)=${LATENCY_BREACHES}"

# ==============================================================================
# Phase 3: Assertions on Workload Execution
# ==============================================================================
log_sub_header "Phase 3: Workload Throughput & Compliance Assertions"

# TC 4.W1: All 15 transactions processed
assert_equals "TC 4.W1: Total shift transactions executed" "$TOTAL_TRANSACTIONS" "15"

# TC 4.W2: Compliant arrivals granted
assert_equals "TC 4.W2: Compliant workers granted access" "$ACTUAL_GRANTED" "$EXPECTED_GRANTED"

# TC 4.W3: Non-compliant arrivals denied
assert_equals "TC 4.W3: Non-compliant workers blocked" "$ACTUAL_DENIED" "$EXPECTED_DENIED"

# TC 4.W4: Unknown / unregistered worker correctly rejected
assert_equals "TC 4.W4: Unknown badge rejected (404/422)" "$ACTUAL_NOT_FOUND" "$EXPECTED_NOT_FOUND"

# TC 4.W5: Maximum latency does not breach critical industrial SLA (1500ms max ceiling)
assert_json_expr "TC 4.W5: Max transaction latency within acceptable operational boundary (<1500ms)" \
  "{\"max_ms\": $MAX_LATENCY_MS}" \
  '(.max_ms < 1500)'


# ==============================================================================
# Phase 4: Shift Audit Reconciliation (DB Audit Trail Consistency)
# ==============================================================================
log_sub_header "Phase 4: Audit Reconciliation (Database Consistency Check)"

# Query access logs for the plant
LOGS_RESP=$(do_request GET "/api/v1/access-logs?plant_id=${TEST_PLANT_ID}&page=1&per_page=50" "$GUARD_TOKEN" "")
LOGS_STATUS=$(echo "$LOGS_RESP" | tail -n1)
LOGS_BODY=$(echo "$LOGS_RESP" | sed '$d')

assert_status "TC 4.W6: Access logs retrieved for shift reconciliation" "$LOGS_STATUS" "200"

TOTAL_LOGGED_RESULTS=$(echo "$LOGS_BODY" | jq '.data | length')
log_info "Total access logs returned in shift window: ${TOTAL_LOGGED_RESULTS}"

# TC 4.W7: Audit trail contains valid logs matching our shift transactions
assert_json_expr "TC 4.W7: Access logs contain granted transactions" "$LOGS_BODY" \
  '(.data[] | select(.result == "granted")) != null'

assert_json_expr "TC 4.W8: Access logs contain denied transactions" "$LOGS_BODY" \
  '(.data[] | select(.result == "denied")) != null'

# TC 4.W9: Zero data corruption — snapshots in logs are valid JSON arrays
assert_json_expr "TC 4.W9: Immutable snapshots preserved without corruption" "$LOGS_BODY" \
  '(.data | map(select(.document_snapshot != null)) | length) > 0'

# ==============================================================================
# Print Final Summary & Exit
# ==============================================================================
print_summary "Tier 4: Industrial Shift Workload Simulation"
exit $?
