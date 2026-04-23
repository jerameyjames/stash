package actions

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alash3al/stash/internal/store"
)

// ParseFilterDSL parses a simple filter DSL and returns a store.Predicate.
// Format: "field=value,field>=value,field<value"
// Supported operators: =, !=, <, >, <=, >=
// Multiple filters are AND-ed together.
// Returns an error if the DSL is malformed.
func ParseFilterDSL(dsl string) (*store.Predicate, error) {
	if dsl == "" {
		return nil, nil
	}

	parts := strings.Split(dsl, ",")
	if len(parts) == 0 {
		return nil, nil
	}

	predicates := make([]store.Predicate, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		pred, err := parseFilterExpression(part)
		if err != nil {
			return nil, err
		}
		predicates = append(predicates, pred)
	}

	if len(predicates) == 0 {
		return nil, nil
	}

	if len(predicates) == 1 {
		return &predicates[0], nil
	}

	// Multiple predicates: AND them together
	return &store.Predicate{
		And: predicates,
	}, nil
}

// parseFilterExpression parses a single filter expression like "severity=high" or "level>=3".
// Returns a Predicate or an error.
func parseFilterExpression(expr string) (store.Predicate, error) {
	// Try operators in order of length (longest first to avoid matching <= as < then =)
	operators := []struct {
		op   string
		storeOp store.Op
	}{
		{"<=", store.OpLte},
		{">=", store.OpGte},
		{"!=", store.OpNe},
		{"<", store.OpLt},
		{">", store.OpGt},
		{"=", store.OpEq},
	}

	var field, value string
	var op store.Op

	found := false
	for _, o := range operators {
		idx := strings.Index(expr, o.op)
		if idx > 0 {
			field = strings.TrimSpace(expr[:idx])
			value = strings.TrimSpace(expr[idx+len(o.op):])
			op = o.storeOp
			found = true
			break
		}
	}

	if !found {
		return store.Predicate{}, fmt.Errorf("invalid filter expression %q: must contain operator (=, !=, <, >, <=, >=)", expr)
	}

	if field == "" {
		return store.Predicate{}, fmt.Errorf("invalid filter expression %q: missing field name", expr)
	}

	if value == "" {
		return store.Predicate{}, fmt.Errorf("invalid filter expression %q: missing value", expr)
	}

	// Convert field to dotted metadata path
	metadataField := "metadata." + field

	// Try to parse value as number if the operator suggests numeric comparison
	var finalValue any = value
	if op == store.OpGt || op == store.OpGte || op == store.OpLt || op == store.OpLte {
		if num, err := strconv.ParseFloat(value, 64); err == nil {
			finalValue = num
		}
		// If it's not a number, keep it as string (comparison will still work)
	}

	return store.Predicate{
		Field: metadataField,
		Op:    op,
		Value: finalValue,
	}, nil
}
