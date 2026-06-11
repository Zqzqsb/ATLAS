#!/bin/bash

# ATLAS Log Collection Script
# Collects lucid-frontend / lucid-backend / lucid-mariadb container logs
# Automatically splits backend logs into per-agent files
#
# Usage:
#   bash scripts/collect-logs.sh          # Real-time follow mode
#   bash scripts/collect-logs.sh -n       # Snapshot mode (collect and exit)
#   bash scripts/collect-logs.sh -c       # Preserve ANSI color codes
#   bash scripts/collect-logs.sh -t 500   # Tail last 500 lines per container

set -euo pipefail

# ============================================================
# Parse arguments
# ============================================================
KEEP_COLOR=false
FOLLOW_MODE=true
TAIL_LINES=200
SHOW_HELP=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --color|-c)
            KEEP_COLOR=true
            shift
            ;;
        --no-follow|-n)
            FOLLOW_MODE=false
            shift
            ;;
        --tail|-t)
            TAIL_LINES="${2:-200}"
            shift 2
            ;;
        --help|-h)
            SHOW_HELP=true
            shift
            ;;
        *)
            shift
            ;;
    esac
done

if [ "$SHOW_HELP" = true ]; then
    echo "Usage: $0 [options]"
    echo ""
    echo "Options:"
    echo "  --color, -c        Preserve ANSI color codes"
    echo "  --no-follow, -n    Snapshot mode (collect current logs and exit)"
    echo "  --tail, -t <N>     Tail last N lines per container (default: 200)"
    echo "  --help, -h         Show this help"
    echo ""
    echo "Output structure:"
    echo "  logs/<timestamp>/"
    echo "    all.log           - Merged log from all containers"
    echo "    backend.log       - Full backend log"
    echo "    frontend.log      - Full frontend log"
    echo "    mariadb.log       - Full MariaDB log"
    echo "    agents/"
    echo "      grounding.log   - [Ground] Semantic Grounding agent"
    echo "      linking.log     - [Link] Schema Linking agent"
    echo "      react.log       - ReAct agent (schema linking & SQL gen)"
    echo "      sqlgen.log      - [Execute] SQL generation + [Connect/ExecuteQuery] DB adapter"
    echo "      lakebase.log    - [Lakebase] Lake-Base storage operations"
    echo "      embedding.log   - Embedding operations"
    echo "      agent.log       - [Agent] Self-maintenance agent"
    echo "      gin.log         - [GIN] HTTP request logs"
    echo "      other.log       - Unmatched backend logs"
    echo ""
    echo "Examples:"
    echo "  $0                 # Real-time follow"
    echo "  $0 -n              # Snapshot: collect and exit"
    echo "  $0 -n -t 1000     # Snapshot last 1000 lines"
    echo "  $0 -c              # Preserve colors"
    exit 0
fi

# ============================================================
# Prepare log directory
# ============================================================
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
LOG_DIR="$PROJECT_DIR/logs/$TIMESTAMP"
AGENT_DIR="$LOG_DIR/agents"
mkdir -p "$AGENT_DIR"

echo "📝 ATLAS Log Collection"
echo "   Output: $LOG_DIR"
if [ "$KEEP_COLOR" = true ]; then
    echo "   🎨 Color mode: ANSI codes preserved (use less -R to view)"
else
    echo "   📄 Plain text mode (editor-friendly)"
fi
echo ""

# ============================================================
# Discover containers
# ============================================================
EXPECTED_CONTAINERS=(
    "lucid-mariadb"
    "lucid-backend"
    "lucid-frontend"
)

CONTAINERS=()
MISSING=()

for container in "${EXPECTED_CONTAINERS[@]}"; do
    if docker ps -a --format '{{.Names}}' | grep -q "^${container}$"; then
        CONTAINERS+=("$container")
    else
        MISSING+=("$container")
    fi
done

if [ ${#CONTAINERS[@]} -eq 0 ]; then
    echo "❌ No atlas/lucid containers found. Start services first:"
    echo "   make rebuild   or   make up"
    exit 1
fi

echo "📋 Containers:"
for container in "${CONTAINERS[@]}"; do
    STATUS=$(docker inspect -f '{{.State.Status}}' "$container" 2>/dev/null || echo "unknown")
    if [ "$STATUS" = "running" ]; then
        echo "   ✅ $container (running)"
    else
        echo "   ⚠️  $container ($STATUS)"
    fi
done

if [ ${#MISSING[@]} -gt 0 ]; then
    echo ""
    echo "⚠️  Missing containers:"
    for container in "${MISSING[@]}"; do
        echo "   - $container"
    done
fi
echo ""

# ============================================================
# ANSI color strip pattern
# ============================================================
STRIP_ANSI='s/\x1b\[[0-9;]*m//g'

# ============================================================
# Per-container tail lines
# ============================================================
get_tail_lines() {
    local container=$1
    case "$container" in
        *backend*)  echo "$((TAIL_LINES * 5))" ;;
        *)          echo "$TAIL_LINES" ;;
    esac
}

# ============================================================
# Log filename from container name
# ============================================================
get_log_filename() {
    local container=$1
    echo "${container#lucid-}.log"
}

# ============================================================
# Backend agent log router
# Routes each backend log line to the appropriate agent file
# ============================================================
route_backend_line() {
    local line="$1"
    local agent_file=""

    # Match by log prefix tags in the message
    case "$line" in
        *"[Ground]"*|*"[SmallScale]"*|*"[LargeScale]"*|*"[performGrounding]"*)
            agent_file="$AGENT_DIR/grounding.log"
            ;;
        *"[Link]"*|*"Schema linking"*|*"schema_linking"*)
            agent_file="$AGENT_DIR/linking.log"
            ;;
        *"ReAct"*|*"react"*|*"iteration"*|*"[Thought]"*|*"[Action]"*|*"[Observation]"*|*"final answer"*)
            agent_file="$AGENT_DIR/react.log"
            ;;
        *"[Execute]"*|*"SQL gen"*|*"sql_gen"*|*"Generated SQL"*)
            agent_file="$AGENT_DIR/sqlgen.log"
            ;;
        *"[Connect]"*|*"[ExecuteQuery]"*|*"mysql_adapter"*)
            agent_file="$AGENT_DIR/sqlgen.log"
            ;;
        *"[NewService]"*|*"grounding_service"*)
            agent_file="$AGENT_DIR/grounding.log"
            ;;
        *"[Lakebase]"*|*"[SyncSchema]"*|*"[SyncAllSchemas]"*|*"Lake-Base"*|*"Rich Context"*|*"rich context"*|*"rc_"*)
            agent_file="$AGENT_DIR/lakebase.log"
            ;;
        *"embedding"*|*"Embedding"*|*"vector"*)
            agent_file="$AGENT_DIR/embedding.log"
            ;;
        *"[Agent]"*|*"[DDL]"*|*"maintenance"*|*"Maintenance"*|*"evolution"*|*"Evolution"*|*"[Stage"*)
            agent_file="$AGENT_DIR/agent.log"
            ;;
        *"[GIN]"*|*"GET"*|*"POST"*|*"PUT"*|*"DELETE"*|*"/health"*|*"/api/"*)
            agent_file="$AGENT_DIR/gin.log"
            ;;
        *"[AutoMigrate]"*|*"Server starting"*|*"API endpoint"*|*"LLM"*|*"Log level"*|*"━"*)
            agent_file="$AGENT_DIR/startup.log"
            ;;
        *)
            agent_file="$AGENT_DIR/other.log"
            ;;
    esac

    echo "$line" >> "$agent_file"
}

# ============================================================
# Track background PIDs
# ============================================================
declare -A PIDS

# ============================================================
# Collect logs for a single container
# ============================================================
collect_log() {
    local container=$1
    local log_file="$LOG_DIR/$(get_log_filename "$container")"
    local all_log="$LOG_DIR/all.log"
    local tail_n
    tail_n=$(get_tail_lines "$container")
    local short_name="${container#lucid-}"
    local is_backend=false
    [[ "$container" == *"backend"* ]] && is_backend=true

    while true; do
        # Check if container exists
        if ! docker ps -a --format '{{.Names}}' | grep -q "^${container}$"; then
            if [ "$FOLLOW_MODE" = true ]; then
                sleep 5
                continue
            else
                break
            fi
        fi

        # Check if container is running (for follow mode)
        if ! docker ps --format '{{.Names}}' | grep -q "^${container}$"; then
            if [ "$FOLLOW_MODE" = true ]; then
                sleep 5
                continue
            fi
        fi

        # Build docker logs command
        local docker_opts="--tail $tail_n"
        if [ "$FOLLOW_MODE" = true ]; then
            docker_opts="-f --tail $tail_n"
        fi

        # Collect logs
        if [ "$KEEP_COLOR" = true ]; then
            docker logs $docker_opts "$container" 2>&1 | \
                while IFS= read -r line; do
                    local tagged="[$short_name] $line"
                    echo "$tagged" >> "$log_file"
                    echo "$tagged" >> "$all_log"
                    # Route backend lines to agent-specific files
                    if [ "$is_backend" = true ]; then
                        route_backend_line "$tagged"
                    fi
                done
        else
            docker logs $docker_opts "$container" 2>&1 | \
                sed -u "$STRIP_ANSI" | \
                while IFS= read -r line; do
                    local tagged="[$short_name] $line"
                    echo "$tagged" >> "$log_file"
                    echo "$tagged" >> "$all_log"
                    # Route backend lines to agent-specific files
                    if [ "$is_backend" = true ]; then
                        route_backend_line "$tagged"
                    fi
                done
        fi

        # If not follow mode, exit after first pass
        if [ "$FOLLOW_MODE" = false ]; then
            break
        fi

        sleep 2
    done
}

# ============================================================
# Start log collection
# ============================================================
echo "🚀 Starting log collection..."
for container in "${CONTAINERS[@]}"; do
    collect_log "$container" &
    PIDS[$container]=$!
done

echo ""
echo "✅ Collecting logs..."
echo ""
echo "   Container logs:"
for container in "${CONTAINERS[@]}"; do
    echo "     $(get_log_filename "$container")"
done
echo ""
echo "   Agent logs (auto-split from backend):"
echo "     agents/grounding.log   - Semantic Grounding"
echo "     agents/linking.log     - Schema Linking"
echo "     agents/react.log       - ReAct Agent"
echo "     agents/sqlgen.log      - SQL Generation"
echo "     agents/lakebase.log    - Lake-Base Storage"
echo "     agents/embedding.log   - Embedding"
echo "     agents/agent.log       - Self-Maintenance"
echo "     agents/gin.log         - HTTP Requests"
echo "     agents/startup.log     - Server Startup"
echo "     agents/other.log       - Other"
echo ""
echo "   Merged: all.log"
echo ""

if [ "$FOLLOW_MODE" = true ]; then
    echo "🔄 Follow mode — press Ctrl+C to stop"
    echo "================================================"
fi

# ============================================================
# Cleanup on exit
# ============================================================
cleanup() {
    echo ""
    echo "🛑 Stopping log collection..."
    for container in "${!PIDS[@]}"; do
        if [ -n "${PIDS[$container]}" ]; then
            kill "${PIDS[$container]}" 2>/dev/null || true
        fi
    done
    sleep 1

    # Remove empty agent log files
    find "$AGENT_DIR" -name "*.log" -empty -delete 2>/dev/null || true
    # Remove agents dir if empty
    rmdir "$AGENT_DIR" 2>/dev/null || true

    echo "✅ Done"
    echo ""
    echo "📂 Logs saved to: $LOG_DIR"
    echo ""
    echo "   Container logs:"
    ls -lh "$LOG_DIR"/*.log 2>/dev/null | awk '{print "     "$NF" ("$5")"}'
    if [ -d "$AGENT_DIR" ]; then
        echo ""
        echo "   Agent logs:"
        ls -lh "$AGENT_DIR"/*.log 2>/dev/null | awk '{print "     "$NF" ("$5")"}'
    fi
    exit 0
}

trap cleanup INT TERM

# Wait for all background jobs
wait

# In snapshot mode, run cleanup manually (wait returns normally)
if [ "$FOLLOW_MODE" = false ]; then
    cleanup
fi
