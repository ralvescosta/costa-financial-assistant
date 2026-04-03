// Package contracts defines transport-agnostic contract primitives for BFF
// service boundaries. Domain-specific contract types are added in companion
// files under the same package.
package contracts

// Empty is a semantic placeholder used by service contracts that return no
// payload while still requiring a typed response boundary.
type Empty struct{}

// Page captures pagination information passed through service contracts.
type Page struct {
	Size  int32
	Token string
}

// PageResult captures the cursor returned by paginated service calls.
type PageResult struct {
	NextToken string
}
