#!/bin/bash
#
# Stress Test: Agent Workload Simulation
#
# Simulates realistic agent behavior:
# - Continuous remember operations (learning from interactions)
# - Frequent recall queries (retrieving context)
# - Periodic consolidation (synthesizing memories)
# - Graph traversal (reasoning about relationships)
# - Relationship extraction (building knowledge model)
#
# Run this to verify Stash can handle agent-like workload:
#   ./stress-test.sh [duration_seconds] [concurrency]
#

set -e

STASH="${STASH:-./stash}"
DURATION="${1:-60}"           # Default 60 seconds
CONCURRENCY="${2:-4}"         # Default 4 parallel agents
NAMESPACE="stress-test-$$"    # Unique namespace per run

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Statistics
REMEMBER_COUNT=0
RECALL_COUNT=0
CONSOLIDATE_COUNT=0
GRAPH_COUNT=0
EXTRACT_COUNT=0
ERROR_COUNT=0

# Sample facts for agent to remember
FACTS=(
    "Alice works at TechCorp as an engineer"
    "Bob manages Alice's team"
    "TechCorp is located in San Francisco"
    "Alice and Bob work on authentication system"
    "The authentication system uses OAuth2"
    "OAuth2 was chosen for security reasons"
    "Alice prefers TypeScript for backend work"
    "Bob prefers Go for infrastructure"
    "TechCorp uses PostgreSQL for persistence"
    "PostgreSQL was chosen for ACID guarantees"
    "Alice completed the OAuth2 migration last week"
    "The migration improved security score by 40%"
    "Bob discovered a race condition in cache layer"
    "Alice fixed the race condition in two hours"
    "The team meets every Monday for planning"
    "TechCorp has 200 employees"
    "Alice has been at TechCorp for 3 years"
    "Bob joined TechCorp 2 years ago"
    "The authentication system processes 10K requests/sec"
    "Response time SLA is 95th percentile under 100ms"
)

QUERIES=(
    "what do we know about Alice?"
    "what systems do we have?"
    "who works on authentication?"
    "what companies are involved?"
    "what technologies do we use?"
    "who manages who?"
    "what happened recently?"
    "why was OAuth2 chosen?"
    "what are team preferences?"
    "performance characteristics we track"
)

ENTITIES=(
    "Alice"
    "Bob"
    "TechCorp"
    "OAuth2"
    "PostgreSQL"
    "authentication"
)

# Log helper
log() {
    echo -e "${GREEN}[$(date '+%H:%M:%S')]${NC} $1"
}

error() {
    echo -e "${RED}[$(date '+%H:%M:%S')] ERROR: $1${NC}"
    ((ERROR_COUNT++))
}

warn() {
    echo -e "${YELLOW}[$(date '+%H:%M:%S')] WARN: $1${NC}"
}

# Simulate single agent in background
agent_workload() {
    local agent_id=$1
    local end_time=$(($(date +%s) + DURATION))
    
    while [[ $(date +%s) -lt $end_time ]]; do
        # 50% remember (learning)
        if (( RANDOM % 100 < 50 )); then
            local fact="${FACTS[$((RANDOM % ${#FACTS[@]}))]}"
            if $STASH remember "$fact" --namespace "$NAMESPACE" &>/dev/null; then
                ((REMEMBER_COUNT++))
            else
                error "remember failed for agent $agent_id"
            fi
        fi
        
        # 30% recall (retrieving context)
        if (( RANDOM % 100 < 30 )); then
            local query="${QUERIES[$((RANDOM % ${#QUERIES[@]}))]}"
            if $STASH recall "$query" --namespace "$NAMESPACE" --limit 5 &>/dev/null; then
                ((RECALL_COUNT++))
            else
                error "recall failed for agent $agent_id"
            fi
        fi
        
        # 10% consolidate (synthesizing beliefs)
        if (( RANDOM % 100 < 10 )); then
            if $STASH facts consolidate --namespace "$NAMESPACE" --window 5m &>/dev/null; then
                ((CONSOLIDATE_COUNT++))
            else
                error "consolidate failed for agent $agent_id"
            fi
        fi
        
        # 5% extract relationships (building knowledge model)
        if (( RANDOM % 100 < 5 )); then
            if $STASH facts extract-relationships --namespace "$NAMESPACE" --limit 50 &>/dev/null; then
                ((EXTRACT_COUNT++))
            else
                error "extract-relationships failed for agent $agent_id"
            fi
        fi
        
        # 5% graph traversal (reasoning about connections)
        if (( RANDOM % 100 < 5 )); then
            local entity="${ENTITIES[$((RANDOM % ${#ENTITIES[@]}))]}"
            if $STASH facts graph --entity "$entity" --namespace "$NAMESPACE" --depth 2 &>/dev/null 2>&1; then
                ((GRAPH_COUNT++))
            fi
        fi
        
        # Small delay to avoid hammering
        sleep 0.1
    done
}

# Main
main() {
    log "🧠 Stash Stress Test"
    log "Duration: ${DURATION}s | Concurrency: ${CONCURRENCY} agents"
    log "Namespace: $NAMESPACE"
    log "---"
    
    # Verify binary exists
    if [[ ! -f "$STASH" ]]; then
        error "Stash binary not found at $STASH"
        echo "Build it with: go build -o stash ./cmd/cli"
        exit 1
    fi
    
    # Verify environment
    if ! $STASH env &>/dev/null; then
        error "Stash environment not configured properly"
        exit 1
    fi
    
    log "Starting ${CONCURRENCY} parallel agents for ${DURATION} seconds..."
    log "---"
    
    START_TIME=$(date +%s%N)
    
    # Launch concurrent agents
    for ((i = 1; i <= CONCURRENCY; i++)); do
        agent_workload $i &
    done
    
    # Monitor progress
    while [[ $(jobs -r -p | wc -l) -gt 0 ]]; do
        sleep 5
        elapsed=$(($(date +%s) - $(echo $START_TIME | cut -c1-10)))
        log "Progress [$elapsed/$DURATION] - Remember: $REMEMBER_COUNT | Recall: $RECALL_COUNT | Consolidate: $CONSOLIDATE_COUNT | Extract: $EXTRACT_COUNT | Graph: $GRAPH_COUNT | Errors: $ERROR_COUNT"
    done
    
    END_TIME=$(date +%s%N)
    ELAPSED=$(($(echo "$END_TIME - $START_TIME" | bc) / 1000000000))
    
    log "---"
    log "✅ Stress test complete"
    
    # Calculate rates
    REMEMBER_RATE=$(echo "scale=2; $REMEMBER_COUNT / $ELAPSED" | bc)
    RECALL_RATE=$(echo "scale=2; $RECALL_COUNT / $ELAPSED" | bc)
    TOTAL_OPS=$((REMEMBER_COUNT + RECALL_COUNT + CONSOLIDATE_COUNT + EXTRACT_COUNT + GRAPH_COUNT))
    OPS_RATE=$(echo "scale=2; $TOTAL_OPS / $ELAPSED" | bc)
    
    # Results
    cat << EOF

📊 RESULTS
──────────────────────────────────────────
Remember Operations:        $REMEMBER_COUNT ops (${REMEMBER_RATE} ops/sec)
Recall Operations:          $RECALL_COUNT ops (${RECALL_RATE} ops/sec)
Consolidations:             $CONSOLIDATE_COUNT
Relationship Extractions:   $EXTRACT_COUNT
Graph Traversals:           $GRAPH_COUNT
──────────────────────────────────────────
Total Operations:           $TOTAL_OPS
Total Time:                 ${ELAPSED}s
Overall Throughput:         ${OPS_RATE} ops/sec
──────────────────────────────────────────
Errors:                     $ERROR_COUNT
Status:                     $([ $ERROR_COUNT -eq 0 ] && echo "✅ PASS" || echo "❌ FAIL")

EOF
    
    # Verify data was actually stored
    log "---"
    log "Verifying stored data..."
    
    STORED=$(timeout 10 $STASH facts list --namespace "$NAMESPACE" --limit 1000 2>/dev/null | wc -l)
    log "Facts stored: $STORED"
    
    RELATIONSHIPS=$(timeout 10 $STASH facts relationships --entity Alice --namespace "$NAMESPACE" 2>/dev/null | wc -l)
    log "Relationships found: $RELATIONSHIPS"
    
    # Exit code based on results
    if [[ $ERROR_COUNT -eq 0 ]] && [[ $TOTAL_OPS -gt 100 ]]; then
        log "---"
        log "✅ Stress test PASSED: Agent workload stable"
        exit 0
    else
        log "---"
        log "❌ Stress test FAILED: Too many errors or low throughput"
        exit 1
    fi
}

main "$@"
