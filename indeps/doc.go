// Package indeps is a small library for analysing a single Go package to
// discover the internal dependencies between its top-level types and functions.
//
// The intended use-case is to look for clusters of declarations that could
// potentially be factored out into a separate package without creating cycles.
package indeps
