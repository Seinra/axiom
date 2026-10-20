// Package di provides a lazy-singleton Injector built on Provide and Invoke.
//
// Register a constructor with Provide, then resolve the shared singleton with
// Invoke: the constructor runs once on first Invoke and the cached value
// serves later calls. Invoke is safe for concurrent use; failing builds are
// discarded so the next Invoke retries.
//
// One type per name: a service name maps to exactly one Service type, and a
// second Provide with the same name reports ErrDuplicateService.
//
// Dependency cycles are a programmer error (infinite recursion), same as the
// reference samber/do model: a constructor that (transitively) Invokes its
// own service never settles, so the container performs no cycle detection.
//
// No scopes, no eager build, no cycle detection, no struct-tag injection by
// design (YAGNI).
//
// Boundary: the table type-erases INTERNALLY (one private map) but any never
// leaks through the public API — every crossing is gated by a generic Invoke
// returning T or an error. This is the single sanctioned exception to the
// no-any rule, and it stays inside this package.
package di

import (
	"errors"
	"fmt"
	"sync"
)

// ErrDuplicateService reports a second Provide with an already-used service name.
var ErrDuplicateService = errors.New("axiom di: duplicate service name")

// ErrServiceNotFound reports an Invoke for a service name with no Provide.
var ErrServiceNotFound = errors.New("axiom di: service not found")

// Injector is a lazy-singleton container of named service constructors.
// Do not copy an Injector after use: it holds a mutex-guarded table, so share
// the pointer from New instead of copying the value.
type Injector struct {
	mutex sync.Mutex
	table map[string]*serviceEntry
}

// serviceEntry holds one lazy singleton: the type-erased constructor, the
// cached build once it succeeds, and the in-flight signal concurrent Invokes
// wait on while a build runs outside the table lock.
type serviceEntry struct {
	build     func(*Injector) (any, error)
	cached    any
	hasCached bool
	building  bool
	settled   chan struct{}
}

// New returns an empty Injector. Share the returned pointer and never copy it:
// the Injector holds a mutex-guarded table that must not be copied after use.
func New() *Injector {
	return &Injector{table: make(map[string]*serviceEntry)}
}

// Provide registers constructor as the lazy builder for serviceName.
// The constructor runs once on the first Invoke for serviceName, never at
// Provide time. One type per name: serviceName maps to exactly one Service
// type. A second Provide with the same name reports ErrDuplicateService and
// keeps the first registration.
func Provide[Service any](injector *Injector, serviceName string, constructor func(*Injector) (Service, error)) error {
	if injector == nil {
		return fmt.Errorf("axiom di: Provide service %q on nil Injector", serviceName)
	}
	if serviceName == "" {
		return fmt.Errorf("axiom di: Provide with empty service name")
	}
	if constructor == nil {
		return fmt.Errorf("axiom di: Provide service %q with nil constructor", serviceName)
	}
	injector.mutex.Lock()
	defer injector.mutex.Unlock()
	if _, found := injector.table[serviceName]; found {
		return fmt.Errorf("axiom di: Provide duplicate service %q: %w", serviceName, ErrDuplicateService)
	}
	injector.table[serviceName] = &serviceEntry{
		build: func(active *Injector) (any, error) {
			return constructor(active)
		},
	}
	return nil
}

// Invoke resolves the singleton for serviceName, running its constructor once
// on first use and returning the cached value afterwards. A missing name
// reports ErrServiceNotFound. Invoking on a nil Injector reports the same
// ErrServiceNotFound sentinel (no container holds no service): distinguish
// the two cases with errors.Is plus the message text ("nil Injector" versus
// "unknown service"). A failing build propagates the constructor
// failure without caching, so the next Invoke retries. Concurrent Invokes for
// one service share a single build: the table lock guards only the lookup and
// the publish, never the constructor run.
//
// A panicking build resets the entry so the next Invoke retries, and every
// concurrent waiter retries and re-runs the constructor instead of receiving
// the panic payload directly: each caller still panics with the same payload,
// but the constructor may execute once per caller under concurrency.
func Invoke[Service any](injector *Injector, serviceName string) (Service, error) {
	var empty Service
	if injector == nil {
		return empty, fmt.Errorf("axiom di: Invoke service %q on nil Injector: %w", serviceName, ErrServiceNotFound)
	}
	for {
		injector.mutex.Lock()
		entry, found := injector.table[serviceName]
		if !found {
			injector.mutex.Unlock()
			return empty, fmt.Errorf("axiom di: Invoke unknown service %q: %w", serviceName, ErrServiceNotFound)
		}
		if entry.hasCached {
			cached := entry.cached
			injector.mutex.Unlock()
			return castCached[Service](serviceName, cached)
		}
		if entry.building {
			settled := entry.settled
			injector.mutex.Unlock()
			<-settled
			continue
		}
		entry.building = true
		settled := make(chan struct{})
		entry.settled = settled
		build := entry.build
		injector.mutex.Unlock()
		built, constructionFailure := invokeBuild(injector, entry, settled, build)
		if constructionFailure != nil {
			abandonBuild(injector, entry, settled)
			return empty, fmt.Errorf("axiom di: build service %q: %w", serviceName, constructionFailure)
		}
		injector.mutex.Lock()
		entry.cached = built
		entry.hasCached = true
		entry.building = false
		entry.settled = nil
		injector.mutex.Unlock()
		close(settled)
		return castCached[Service](serviceName, built)
	}
}

// abandonBuild resets entry for the next attempt and wakes its waiters by
// closing settled exactly once. Shared by the failure and panic paths so the
// reset-and-wake pairing cannot drift between them.
func abandonBuild(injector *Injector, entry *serviceEntry, settled chan struct{}) {
	injector.mutex.Lock()
	entry.building = false
	entry.settled = nil
	injector.mutex.Unlock()
	close(settled)
}

// invokeBuild runs build outside the table lock, resetting the entry and
// waking waiters before re-panicking when build panics. Without the reset,
// waiters would block forever on a channel that never closes: every caller
// panics with the same payload instead of hanging. Waiters wake, retry, and
// re-run the constructor (retry-then-panic contract), so under concurrency
// the constructor may execute once per caller; the entry stays reusable and
// the next Invoke after the storm retries instead of sticking.
func invokeBuild(injector *Injector, entry *serviceEntry, settled chan struct{}, build func(*Injector) (any, error)) (built any, constructionFailure error) {
	defer func() {
		if payload := recover(); payload != nil {
			abandonBuild(injector, entry, settled)
			panic(payload)
		}
	}()
	return build(injector)
}

// MustInvoke resolves the singleton for serviceName like Invoke but panics on
// a missing name or a failed build. Use only for services known at
// construction time (tests, static seeds): panic signals programmer error,
// never control flow, the same convention as result.Must.
func MustInvoke[Service any](injector *Injector, serviceName string) Service {
	built, invokeFailure := Invoke[Service](injector, serviceName)
	if invokeFailure != nil {
		panic(invokeFailure)
	}
	return built
}

// castCached asserts cached back to the Service type the caller asked for.
// Provide enforces one type per name, so a mismatch reports a descriptive
// failure instead of panicking on the type assertion.
func castCached[Service any](serviceName string, cached any) (Service, error) {
	var empty Service
	typed, holds := cached.(Service)
	if !holds {
		return empty, fmt.Errorf("axiom di: service %q holds %T, not %T", serviceName, cached, empty)
	}
	return typed, nil
}
