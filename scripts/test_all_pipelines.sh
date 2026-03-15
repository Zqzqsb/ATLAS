#!/bin/bash
# Test all grounding pipeline combinations
# Usage: bash scripts/test_all_pipelines.sh

BASE_URL="http://localhost:19001/api/v1"
PASS=0
FAIL=0
SKIP=0

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

run_test() {
  local TEST_NAME="$1"
  local DB_ID="$2"
  local LINKING_MODE="$3"
  local FORCE_SMALL="$4"
  local GROUNDING_ONLY="$5"
  local QUESTION="$6"

  echo -e "\n${CYAN}═══════════════════════════════════════════════════${NC}"
  echo -e "${CYAN}TEST: ${TEST_NAME}${NC}"
  echo -e "${CYAN}  DB=${DB_ID}  linking_mode=${LINKING_MODE}  force_small=${FORCE_SMALL}${NC}"
  echo -e "${CYAN}═══════════════════════════════════════════════════${NC}"

  # Build JSON
  local JSON=$(cat <<EOF
{
  "question": "${QUESTION}",
  "database_id": "${DB_ID}",
  "options": {
    "linking_mode": "${LINKING_MODE}",
    "force_small_scale": ${FORCE_SMALL},
    "grounding_only": ${GROUNDING_ONLY},
    "use_grounding": true,
    "use_rich_context": true,
    "stream": true
  }
}
EOF
)

  # Call SSE endpoint, capture full response
  local RESPONSE
  RESPONSE=$(curl -s -N --max-time 120 \
    -H "Content-Type: application/json" \
    -H "Accept: text/event-stream" \
    -d "$JSON" \
    "${BASE_URL}/text2sql/stream" 2>&1)

  if [ $? -ne 0 ]; then
    echo -e "${RED}✗ FAIL: curl error${NC}"
    FAIL=$((FAIL + 1))
    return
  fi

  # Extract key SSE events
  local EVENTS=$(echo "$RESPONSE" | grep '^event:')
  local HAS_ERROR=$(echo "$EVENTS" | grep -c 'event: error')
  local HAS_GROUNDING_START=$(echo "$EVENTS" | grep -c 'event: grounding_start')
  local HAS_RETRIEVAL_COMPLETE=$(echo "$EVENTS" | grep -c 'event: retrieval_complete')
  local HAS_LINKING_COMPLETE=$(echo "$EVENTS" | grep -c 'event: linking_complete')
  local HAS_GROUNDING_COMPLETE=$(echo "$EVENTS" | grep -c 'event: grounding_complete')
  local HAS_COMPLETE=$(echo "$EVENTS" | grep -c 'event: complete')

  # Check for linking_step ReAct events (these are emitted as thought/action/observation with phase=schema_linking)
  # But we need to distinguish: linking_step events from grounding vs inference pipeline schema_linking events.
  # Grounding linking_steps come BEFORE grounding_complete.
  # Inference schema_linking events come AFTER grounding_complete.
  
  # Count linking_step events = thought events with phase=schema_linking that appear between grounding_start and grounding_complete
  # Simpler approach: check for linking_agent_direct vs linking_agent_async in execution_logs
  local HAS_AGENT_DIRECT=$(echo "$RESPONSE" | grep -c '"linking_agent_direct"')
  local HAS_AGENT_ASYNC=$(echo "$RESPONSE" | grep -c '"linking_agent_async"')

  # Extract strategy
  local STRATEGY=$(echo "$RESPONSE" | grep -o '"strategy":"[^"]*"' | head -1 | cut -d'"' -f4)
  
  # Extract selected table count from linking_complete
  local REASONING_LATENCY=$(echo "$RESPONSE" | grep 'event: linking_complete' -A 1 | grep -o '"reasoning_latency_ms":[0-9]*' | cut -d: -f2)
  local RETRIEVAL_LATENCY=$(echo "$RESPONSE" | grep 'event: linking_complete' -A 1 | grep -o '"retrieval_latency_ms":[0-9]*' | cut -d: -f2)

  echo "  Strategy: ${STRATEGY:-N/A}"
  echo "  Events: grounding_start=${HAS_GROUNDING_START} retrieval_complete=${HAS_RETRIEVAL_COMPLETE} linking_complete=${HAS_LINKING_COMPLETE} grounding_complete=${HAS_GROUNDING_COMPLETE} complete=${HAS_COMPLETE}"
  echo "  Agent path: direct=${HAS_AGENT_DIRECT} async=${HAS_AGENT_ASYNC}"
  echo "  Latency: retrieval=${RETRIEVAL_LATENCY:-N/A}ms reasoning=${REASONING_LATENCY:-N/A}ms"
  echo "  Error: ${HAS_ERROR}"

  # Validate
  if [ "$HAS_ERROR" -gt 0 ]; then
    local ERROR_MSG=$(echo "$RESPONSE" | grep 'event: error' -A 2 | grep 'data:' | head -1)
    echo -e "${RED}✗ FAIL: Error event received: ${ERROR_MSG}${NC}"
    FAIL=$((FAIL + 1))
    return
  fi

  if [ "$HAS_GROUNDING_START" -eq 0 ]; then
    echo -e "${RED}✗ FAIL: No grounding_start event (use_grounding not enabled?)${NC}"
    FAIL=$((FAIL + 1))
    return
  fi

  if [ "$HAS_GROUNDING_COMPLETE" -eq 0 ]; then
    echo -e "${RED}✗ FAIL: No grounding_complete event${NC}"
    FAIL=$((FAIL + 1))
    return
  fi

  # Mode-specific validations
  case "$LINKING_MODE" in
    "off")
      # Should have no linking agent at all (no direct, no async)
      if [ "$HAS_AGENT_DIRECT" -gt 0 ] || [ "$HAS_AGENT_ASYNC" -gt 0 ]; then
        echo -e "${RED}✗ FAIL: linking_mode=off should skip linking agent entirely${NC}"
        FAIL=$((FAIL + 1))
        return
      fi
      ;;
    "one-shot")
      # Should use LinkDirect (linking_agent_direct), NOT LinkAsync (linking_agent_async)
      if [ "$HAS_AGENT_ASYNC" -gt 0 ]; then
        echo -e "${RED}✗ FAIL: one-shot should use LinkDirect, but found linking_agent_async!${NC}"
        echo "  ↑ This is the core bug: one-shot should NOT use ReAct/LinkAsync"
        FAIL=$((FAIL + 1))
        return
      fi
      if [ "$HAS_AGENT_DIRECT" -eq 0 ] && [ "$HAS_LINKING_COMPLETE" -gt 0 ]; then
        echo -e "${YELLOW}⚠ WARN: one-shot linked but no linking_agent_direct log found${NC}"
      fi
      ;;
    "react")
      # Should use LinkAsync (linking_agent_async)
      if [ "$HAS_AGENT_ASYNC" -eq 0 ]; then
        echo -e "${YELLOW}⚠ WARN: react mode expected linking_agent_async but not found${NC}"
      fi
      ;;
  esac

  echo -e "${GREEN}✓ PASS${NC}"
  PASS=$((PASS + 1))
}

echo "============================================"
echo "  LUCID Pipeline Test Suite"
echo "  Testing all grounding path combinations"
echo "============================================"

# ─────────────────────────────────────────────────
# Group 1: SmallScale (spider_tvshow, 3 tables)
# ─────────────────────────────────────────────────

run_test "SmallScale + off" \
  "spider_tvshow" "off" "false" "false" \
  "Show all TV channels"

run_test "SmallScale + one-shot (LinkDirect)" \
  "spider_tvshow" "one-shot" "false" "false" \
  "Which channel has the most TV series?"

run_test "SmallScale + react" \
  "spider_tvshow" "react" "false" "false" \
  "Find the TV series with highest rating on each channel"

# ─────────────────────────────────────────────────
# Group 2: LargeScale (tpch_enterprise, 517 tables)
# ─────────────────────────────────────────────────

run_test "LargeScale + off (retrieval only)" \
  "tpch_enterprise" "off" "false" "false" \
  "Show total revenue by region"

run_test "LargeScale + one-shot (LinkDirect) ★ CORE FIX" \
  "tpch_enterprise" "one-shot" "false" "false" \
  "What is the total order amount by customer segment?"

run_test "LargeScale + react" \
  "tpch_enterprise" "react" "false" "false" \
  "Find the top suppliers by part count in each nation"

# ─────────────────────────────────────────────────
# Group 3: ForceSmallScale (tpch_enterprise + force_small_scale)
# ─────────────────────────────────────────────────

run_test "ForceSmallScale + off (517 tables, no linking)" \
  "tpch_enterprise" "off" "true" "false" \
  "List all regions"

run_test "ForceSmallScale + one-shot (517 tables, LinkDirect)" \
  "tpch_enterprise" "one-shot" "true" "false" \
  "Show customers by nation"

# Skip react with 517 tables + force_small_scale (too slow, >2min)
echo -e "\n${YELLOW}SKIP: ForceSmallScale + react (517 tables → too slow for test)${NC}"
SKIP=$((SKIP + 1))

# ─────────────────────────────────────────────────
# Summary
# ─────────────────────────────────────────────────
echo -e "\n${CYAN}════════════════════════════════════════════${NC}"
echo -e "${CYAN}  TEST SUMMARY${NC}"
echo -e "${CYAN}════════════════════════════════════════════${NC}"
echo -e "  ${GREEN}PASS: ${PASS}${NC}"
echo -e "  ${RED}FAIL: ${FAIL}${NC}"
echo -e "  ${YELLOW}SKIP: ${SKIP}${NC}"
echo -e "  Total: $((PASS + FAIL + SKIP))"
echo ""

if [ "$FAIL" -gt 0 ]; then
  echo -e "${RED}Some tests FAILED!${NC}"
  exit 1
else
  echo -e "${GREEN}All tests PASSED!${NC}"
  exit 0
fi
