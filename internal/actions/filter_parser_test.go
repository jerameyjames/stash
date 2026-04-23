package actions

import (
	"testing"

	"github.com/alash3al/stash/internal/store"
)

func TestParseFilterDSL(t *testing.T) {
	tests := []struct {
		name    string
		dsl     string
		want    *store.Predicate
		wantErr bool
	}{
		{
			name:    "empty",
			dsl:     "",
			want:    nil,
			wantErr: false,
		},
		{
			name: "single equality",
			dsl:  "severity=high",
			want: &store.Predicate{
				Field: "metadata.severity",
				Op:    store.OpEq,
				Value: "high",
			},
			wantErr: false,
		},
		{
			name: "single not-equals",
			dsl:  "status!=pending",
			want: &store.Predicate{
				Field: "metadata.status",
				Op:    store.OpNe,
				Value: "pending",
			},
			wantErr: false,
		},
		{
			name: "numeric greater-than",
			dsl:  "level>5",
			want: &store.Predicate{
				Field: "metadata.level",
				Op:    store.OpGt,
				Value: 5.0,
			},
			wantErr: false,
		},
		{
			name: "numeric greater-than-or-equal",
			dsl:  "count>=10",
			want: &store.Predicate{
				Field: "metadata.count",
				Op:    store.OpGte,
				Value: 10.0,
			},
			wantErr: false,
		},
		{
			name: "numeric less-than",
			dsl:  "priority<3",
			want: &store.Predicate{
				Field: "metadata.priority",
				Op:    store.OpLt,
				Value: 3.0,
			},
			wantErr: false,
		},
		{
			name: "numeric less-than-or-equal",
			dsl:  "age<=100",
			want: &store.Predicate{
				Field: "metadata.age",
				Op:    store.OpLte,
				Value: 100.0,
			},
			wantErr: false,
		},
		{
			name: "multiple filters AND",
			dsl:  "severity=high,component=gateway",
			want: &store.Predicate{
				And: []store.Predicate{
					{
						Field: "metadata.severity",
						Op:    store.OpEq,
						Value: "high",
					},
					{
						Field: "metadata.component",
						Op:    store.OpEq,
						Value: "gateway",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "three filters AND",
			dsl:  "severity=high,component=api,status!=resolved",
			want: &store.Predicate{
				And: []store.Predicate{
					{
						Field: "metadata.severity",
						Op:    store.OpEq,
						Value: "high",
					},
					{
						Field: "metadata.component",
						Op:    store.OpEq,
						Value: "api",
					},
					{
						Field: "metadata.status",
						Op:    store.OpNe,
						Value: "resolved",
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "missing operator",
			dsl:     "severity",
			wantErr: true,
		},
		{
			name:    "missing field",
			dsl:     "=high",
			wantErr: true,
		},
		{
			name:    "missing value",
			dsl:     "severity=",
			wantErr: true,
		},
		{
			name: "whitespace handling",
			dsl:  " severity = high , component = api ",
			want: &store.Predicate{
				And: []store.Predicate{
					{
						Field: "metadata.severity",
						Op:    store.OpEq,
						Value: "high",
					},
					{
						Field: "metadata.component",
						Op:    store.OpEq,
						Value: "api",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFilterDSL(tt.dsl)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseFilterDSL(%q) error = %v, wantErr %v", tt.dsl, err, tt.wantErr)
			}
			if err != nil {
				return // Expected error case
			}

			// Compare predicates
			if !predicatesEqual(got, tt.want) {
				t.Errorf("ParseFilterDSL(%q) = %#v, want %#v", tt.dsl, got, tt.want)
			}
		})
	}
}

// predicatesEqual compares two predicates for testing.
// This is a simplified comparison that handles basic cases.
func predicatesEqual(a, b *store.Predicate) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Compare basic fields
	if a.Field != b.Field || a.Op != b.Op {
		return false
	}

	// Compare values (handle numeric comparisons)
	if !valuesEqual(a.Value, b.Value) {
		return false
	}

	// Compare And predicates
	if len(a.And) != len(b.And) {
		return false
	}
	for i := range a.And {
		if !predicatesEqual(&a.And[i], &b.And[i]) {
			return false
		}
	}

	// Compare Or predicates
	if len(a.Or) != len(b.Or) {
		return false
	}
	for i := range a.Or {
		if !predicatesEqual(&a.Or[i], &b.Or[i]) {
			return false
		}
	}

	// Compare Not predicate
	if (a.Not == nil) != (b.Not == nil) {
		return false
	}
	if a.Not != nil && !predicatesEqual(a.Not, b.Not) {
		return false
	}

	return true
}

// valuesEqual compares two values, handling numeric type conversions.
func valuesEqual(a, b any) bool {
	if a == b {
		return true
	}

	// Handle float/int comparisons
	switch av := a.(type) {
	case float64:
		switch bv := b.(type) {
		case float64:
			return av == bv
		case int:
			return av == float64(bv)
		}
	case int:
		switch bv := b.(type) {
		case float64:
			return float64(av) == bv
		case int:
			return av == bv
		}
	}

	return false
}
