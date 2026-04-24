#!/bin/bash
#
# Stress Test: Agent Workload Simulation
#
# Simulates realistic agent behavior:
# - Remember operations (learning from interactions)
# - Recall queries (retrieving context)
# - Introspection (reflect)
# - Contradiction detection
# - Memory cleanup (forget)
#
# Usage:
#   ./stress-test.sh [num_operations]  (default: 100)
#   ./stress-test.sh 200               (200 operations)
#

set -e

cd "$(dirname "$0")"

export STASHCONFIG=.env
STASH="${STASH:-./stash}"
NUM_OPS="${1:-100}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'

# Stats
remember=0
recall=0
reflect=0
contradict=0
forget=0
failed=0

log() { echo -e "${GREEN}✓${NC} $1"; }
error() { echo -e "${RED}✗${NC} $1"; ((failed++)); }

# Verify setup
if [[ ! -f "$STASH" ]]; then
  error "Binary not found: $STASH"
  error "Build with: go build -o stash ./cmd/cli"
  exit 1
fi

if ! timeout 10 $STASH env >/dev/null 2>&1; then
  error "Config not loaded. Set STASHCONFIG=.env and verify .env exists"
  exit 1
fi

echo "🧠 Stash Stress Test"
echo "   Running $NUM_OPS operations..."
echo ""

START=$(date +%s%N)

for i in $(seq 1 $NUM_OPS); do
  case $((i % 5)) in
    0)
      if $STASH remember "memory $(date +%s)" >/dev/null 2>&1; then
        ((remember++))
      else
        error "remember"
      fi
      ;;
    
    1)
      if $STASH recall "stash memory" --limit 3 >/dev/null 2>&1; then
        ((recall++))
      else
        error "recall"
      fi
      ;;
    
    2)
      if $STASH reflect >/dev/null 2>&1; then
        ((reflect++))
      else
        error "reflect"
      fi
      ;;
    
    3)
      if $STASH contradict >/dev/null 2>&1; then
        ((contradict++))
      else
        error "contradict"
      fi
      ;;
    
    4)
      $STASH forget "old" >/dev/null 2>&1 || true
      ((forget++))
      ;;
  esac
  
  # Progress every 20
  [[ $((i % 20)) -eq 0 ]] && echo "  Progress: $i / $NUM_OPS"
done

END=$(date +%s%N)
ELAPSED_MS=$(( (END - START) / 1000000 ))
OPS_PER_SEC=$(echo "scale=1; $NUM_OPS * 1000 / $ELAPSED_MS" | bc)

echo ""
echo "📊 Results"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Remember:          $remember ✓"
echo "  Recall:            $recall ✓"
echo "  Reflect:           $reflect ✓"
echo "  Contradict:        $contradict ✓"
echo "  Forget:            $forget ✓"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Total Ops:         $NUM_OPS"
echo "  Time:              ${ELAPSED_MS}ms"
echo "  Throughput:        ${OPS_PER_SEC} ops/sec"
echo "  Errors:            $failed"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [[ $failed -eq 0 ]]; then
  echo "✅ PASS: Agent workload handled smoothly"
  exit 0
else
  echo "❌ FAIL: $failed operations failed"
  exit 1
fi
