// Package repogorm adapts GORM onto the repo.Repository contract.
//
// Generic code cannot divine schema: orderColumn (e.g. "id") is the explicit
// primary-key column List orders by for deterministic pages. Models use
// normal GORM struct tags: the adapter follows GORM conventions, and the
// core's no-tags rule stays core-side. YAGNI: no raw SQL strings beyond the
// order column identifier.
package repogorm

import (
	"errors"
	"fmt"
	"regexp"

	"gorm.io/gorm"

	"github.com/seinra/axiom/repo"
)

// Store is a GORM-backed repo.Repository. List order is deterministic: pages
// always sort by the order column fixed at construction. The zero value is
// not usable: construct with NewStore. Store is safe for concurrent use
// through the shared *gorm.DB handle.
type Store[Record any, Identity comparable] struct {
	database    *gorm.DB
	orderColumn string
}

// NewStore migrates models once and returns the store reading them.
// A nil database, an empty orderColumn, or a migration failure yields a
// descriptive failure: construction problems are programmer errors surfaced
// eagerly, never control flow.
func NewStore[Record any, Identity comparable](database *gorm.DB, models []any, orderColumn string) (*Store[Record, Identity], error) {
	if database == nil {
		return nil, errors.New("axiom repogorm: NewStore needs a non-nil database")
	}
	if orderColumn == "" {
		return nil, errors.New("axiom repogorm: NewStore needs a non-empty order column")
	}
	if !validOrderColumn(orderColumn) {
		return nil, fmt.Errorf("axiom repogorm: order column %q is not a plain identifier", orderColumn)
	}
	if migrateFailure := database.AutoMigrate(models...); migrateFailure != nil {
		return nil, fmt.Errorf("axiom repogorm: migrate %d models: %w", len(models), migrateFailure)
	}
	return &Store[Record, Identity]{database: database, orderColumn: orderColumn}, nil
}

// orderColumnPattern matches a plain dotted identifier: a leading letter or
// underscore, then letters, digits, underscores, or dots.
var orderColumnPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.]*$`)

// validOrderColumn reports whether name is a plain dotted identifier, keeping
// caller-supplied order columns out of SQL-injection shape. Best-effort gate,
// not a grammar: GORM still owns the final word at query time.
func validOrderColumn(name string) bool {
	return orderColumnPattern.MatchString(name)
}

// Store persists candidate, overwriting any record with the same identity.
func (current *Store[Record, Identity]) Store(candidate Record) error {
	outcome := current.database.Save(&candidate)
	if outcome.Error != nil {
		return fmt.Errorf("axiom repogorm: store record: %w", outcome.Error)
	}
	return nil
}

// Fetch returns the record held under identity, or a wrapped ErrRecordNotFound.
func (current *Store[Record, Identity]) Fetch(identity Identity) (Record, error) {
	var found Record
	outcome := current.database.First(&found, identity)
	if outcome.Error != nil {
		if errors.Is(outcome.Error, gorm.ErrRecordNotFound) {
			return found, fmt.Errorf("axiom repogorm: fetch %v: %w", identity, repo.ErrRecordNotFound)
		}
		return found, fmt.Errorf("axiom repogorm: fetch %v: %w", identity, outcome.Error)
	}
	return found, nil
}

// Remove drops the record held under identity, or a wrapped ErrRecordNotFound.
func (current *Store[Record, Identity]) Remove(identity Identity) error {
	var candidate Record
	outcome := current.database.Delete(&candidate, identity)
	if outcome.Error != nil {
		return fmt.Errorf("axiom repogorm: remove %v: %w", identity, outcome.Error)
	}
	if outcome.RowsAffected == 0 {
		return fmt.Errorf("axiom repogorm: remove %v: %w", identity, repo.ErrRecordNotFound)
	}
	return nil
}

// List returns at most Size records for the requested page, ordered by the
// order column. An invalid page (Number < 1 or Size < 1) yields a wrapped
// ErrInvalidPage. Listing past the end yields an empty non-nil slice.
func (current *Store[Record, Identity]) List(page repo.Page) ([]Record, error) {
	if page.Number < 1 {
		return nil, fmt.Errorf("axiom repogorm: list page number %d: %w (numbers start at 1)", page.Number, repo.ErrInvalidPage)
	}
	if page.Size < 1 {
		return nil, fmt.Errorf("axiom repogorm: list page size %d: %w (sizes start at 1)", page.Size, repo.ErrInvalidPage)
	}
	listed := make([]Record, 0, page.Size)
	outcome := current.database.Order(current.orderColumn).Offset(page.Offset()).Limit(page.Size).Find(&listed)
	if outcome.Error != nil {
		return nil, fmt.Errorf("axiom repogorm: list page %+v: %w", page, outcome.Error)
	}
	return listed, nil
}
