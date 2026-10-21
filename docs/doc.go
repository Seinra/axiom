// Package docs holds Axiom's narrative documentation (quickstart, cookbook, guides).
//
// The Markdown files under docs/ are prose; the Go files in this package are
// tests that execute the cookbook recipes, so `go test ./docs/` keeps the
// documented outputs honest. If a recipe changes, update the prose and the
// test together.
package docs
