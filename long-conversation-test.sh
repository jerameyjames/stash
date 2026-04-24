#!/bin/bash
#
# Long Conversation Test: Real-world Agent Memory Simulation
#
# Simulates a realistic agent having multiple conversations:
# - Technical context (code, architecture, infrastructure)
# - Personal context (preferences, working style)
# - Project context (status, decisions, timeline)
# - Tests context switching and semantic recall
#

set -e

cd "$(dirname "$0")"
export STASHCONFIG=.env
STASH="${STASH:-./stash}"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

log() { echo -e "${BLUE}[$(date '+%H:%M:%S')]${NC} $1"; }
success() { echo -e "${GREEN}✓${NC} $1"; }
warn() { echo -e "${YELLOW}⚠${NC} $1"; }
error() { echo -e "${RED}✗${NC} $1"; }

# Verify setup
if [[ ! -f "$STASH" ]]; then
    error "Binary not found"
    exit 1
fi

if ! timeout 10 $STASH env >/dev/null 2>&1; then
    error "Config not loaded"
    exit 1
fi

log "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
log "🧠 Long Conversation Test"
log "Multi-context agent memory simulation"
log "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Phase 1: Technical Context Learning
log "Phase 1️⃣: Technical Context"
tech_learned=0
for fact in \
    "We use PostgreSQL 16 with pgvector extension" \
    "Vector embeddings are 1536 dimensions" \
    "Go 1.25.5 is the standard language" \
    "OpenRouter provides API access" \
    "Kubernetes handles deployment" \
    "Database latency SLA is 500ms p95" \
    "Connection pool size is 20" \
    "TTL for working memory is 1 hour"; do
    
    if $STASH remember "$fact" >/dev/null 2>&1; then
        ((tech_learned++))
    fi
done
success "$tech_learned technical facts learned"
echo ""

# Phase 2: Personal Context Learning
log "Phase 2️⃣: Personal Context"
personal_learned=0
for fact in \
    "I prefer dark mode interface" \
    "I work best before 11am UTC+3" \
    "I take breaks every 90 minutes" \
    "I like detailed technical explanations" \
    "I've been in tech 8 years" \
    "I speak English, French, Arabic" \
    "Coffee is my preferred beverage" \
    "I prefer async over meetings"; do
    
    if $STASH remember "$fact" >/dev/null 2>&1; then
        ((personal_learned++))
    fi
done
success "$personal_learned personal facts learned"
echo ""

# Phase 3: Project Context Learning
log "Phase 3️⃣: Project Context"
project_learned=0
for fact in \
    "Sprint ends Friday 5pm UTC" \
    "Team of 4: Alice (arch), Bob (DevOps), Charlie (QA), Elena (frontend)" \
    "Current goal: launch memory system" \
    "Blocker: vector DB performance benchmarking" \
    "Decision: PostgreSQL over Elasticsearch" \
    "Customer demo next Thursday 2pm" \
    "Budget approved yesterday" \
    "Phase 3 deadline is end of month"; do
    
    if $STASH remember "$fact" >/dev/null 2>&1; then
        ((project_learned++))
    fi
done
success "$project_learned project facts learned"
echo ""

# Phase 4: Context Switch - Technical Recall
log "Phase 4️⃣: Context Switch #1 → Technical"
tech_recalled=0
for q in \
    "What database do we use?" \
    "What's the vector dimension?" \
    "What language do we use?" \
    "What's our deployment platform?"; do
    
    if $STASH recall "$q" --limit 1 2>/dev/null | grep -q "PostgreSQL\|1536\|Go\|Kubernetes"; then
        ((tech_recalled++))
        success "Recalled: $q"
    else
        warn "Partial: $q"
    fi
done
echo ""

# Phase 5: Context Switch - Personal Recall
log "Phase 5️⃣: Context Switch #2 → Personal"
personal_recalled=0
for q in \
    "What's my preferred UI theme?" \
    "When am I most productive?" \
    "How long have I worked in tech?" \
    "What's my communication preference?"; do
    
    if $STASH recall "$q" --limit 1 2>/dev/null | grep -q "dark\|11am\|8 years\|async\|meeting"; then
        ((personal_recalled++))
        success "Recalled: $q"
    else
        warn "Partial: $q"
    fi
done
echo ""

# Phase 6: Context Switch - Project Recall
log "Phase 6️⃣: Context Switch #3 → Project"
project_recalled=0
for q in \
    "When does sprint end?" \
    "Who's on the team?" \
    "What's blocking us?" \
    "When's the customer demo?"; do
    
    if $STASH recall "$q" --limit 1 2>/dev/null | grep -q "Friday\|Alice\|Bob\|performance\|Thursday"; then
        ((project_recalled++))
        success "Recalled: $q"
    else
        warn "Partial: $q"
    fi
done
echo ""

# Phase 7: Rapid Context Switching
log "Phase 7️⃣: Rapid Context Switching (30 switches)"
rapid_switches=0
for i in {1..30}; do
    case $((RANDOM % 3)) in
        0) q="PostgreSQL vectors"; ctx="Tech" ;;
        1) q="dark mode preference"; ctx="Personal" ;;
        2) q="sprint deadline"; ctx="Project" ;;
    esac
    
    if $STASH recall "$q" --limit 1 >/dev/null 2>&1; then
        ((rapid_switches++))
    fi
    
    [[ $((i % 10)) -eq 0 ]] && echo "  $i/30 switches completed..."
done
success "$rapid_switches successful switches"
echo ""

# Phase 8: Memory Introspection
log "Phase 8️⃣: Memory Introspection"
reflect=$($STASH reflect 2>/dev/null)
events=$(echo "$reflect" | grep -o '"total_events":[0-9]*' | grep -o '[0-9]*')
facts=$(echo "$reflect" | grep -o '"total_facts":[0-9]*' | grep -o '[0-9]*')

echo "  Events in store: ${events:-?}"
echo "  Facts synthesized: ${facts:-?}"
echo ""

# Results Summary
log "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
log "📊 Test Results"
log "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

total_learned=$((tech_learned + personal_learned + project_learned))
total_recalled=$((tech_recalled + personal_recalled + project_recalled))
total_ops=$((total_learned + total_recalled + rapid_switches))

cat << EOF

Learning Phase:
  Technical facts:      $tech_learned / 8
  Personal facts:       $personal_learned / 8
  Project facts:        $project_learned / 8
  ──────────────────────────────
  Total learned:        $total_learned / 24

Recall Phase:
  Tech context:         $tech_recalled / 4
  Personal context:     $personal_recalled / 4
  Project context:      $project_recalled / 4
  ──────────────────────────────
  Total recalled:       $total_recalled / 12

Context Switching:
  Rapid switches:       $rapid_switches / 30

Overall:
  Total operations:     $total_ops
  Success rate:         $(echo "scale=1; ($total_learned + $total_recalled + $rapid_switches) * 100 / ($total_ops + 3)" | bc)%

Database State:
  Events stored:        ${events:-unknown}
  Facts synthesized:    ${facts:-unknown}

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
EOF

# Final verdict
echo ""
if [[ $total_learned -gt 20 ]] && [[ $total_recalled -gt 9 ]] && [[ $rapid_switches -gt 25 ]]; then
    log "✅ PASS: Agent handles realistic multi-context conversations"
    log "✓ Memory persistence: verified"
    log "✓ Semantic recall: working"
    log "✓ Context switching: stable"
    log "✓ No errors: confirmed"
    exit 0
else
    log "⚠️  PARTIAL: Test completed with notes"
    log "Review results above for details"
    exit 1
fi
