// Package repo provides a pure-generics repository core: the Repository
// contract, the Page convention future adapters reuse, and an in-memory
// implementation for tests and teaching.
//
// The core depends only on the standard library: no GORM or SQL imports, no
// struct tags, and no reflection (YAGNI — identity extraction stays a
// caller-supplied keyOf function, never tags or scanners). A GORM adapter
// will arrive as its own slice and reuse the Page convention defined here.
// Reach for slices and maps directly when the data fits in memory for one
// call; reach for this package when records need identity, paging, and a
// swappable Repository behind the contract.
//
// Which function?
//
// | When you want to ...                 | Reach for ...                         |
// |---------------------------------------|---------------------------------------|
// | Persist / fetch / drop one record     | Repository.Store / Fetch / Remove     |
// | List one page                         | Repository.List                       |
// | In-memory Repository for tests        | NewMemoryRepository (MemoryRepository) |
// | Select a page; count skipped records  | Page / Page.Offset                    |
// | Count pages; check for a next page    | TotalPages / HasNextPage              |
// | Match missing record / bad page       | ErrRecordNotFound / ErrInvalidPage    |
package repo

import (
	"errors"
	"fmt"
	"sync"
)

// ErrRecordNotFound reports a Fetch or Remove for an unknown identity.
// Compare with errors.Is; the repository wraps it with the missing identity.
var ErrRecordNotFound = errors.New("axiom repo: record not found")

// ErrInvalidPage reports a List call whose page falls outside the 1-based
// convention. Compare with errors.Is.
var ErrInvalidPage = errors.New("axiom repo: invalid page")

// Page selects one slice of a listing. Number is 1-based: the first page is
// Number 1, never 0. Size caps how many records List returns for the page.
//
// Size is application policy, not a repository clamp: List rejects Size < 1
// with ErrInvalidPage but never clamps an oversized Size, so cap Size at the
// handler edge before listing.
type Page struct {
	Number int
	Size   int
}

// Offset returns how many records precede the page: (Number-1)*Size.
// Offset assumes a valid page; List rejects Number < 1 or Size < 1 with
// ErrInvalidPage before applying the offset.
func (current Page) Offset() int {
	return (current.Number - 1) * current.Size
}

// TotalPages returns how many pages hold totalCount records at pageSize per page.
// It divides upward: a remainder needs one more page. A non-positive totalCount
// or pageSize yields 0: there is nothing to paginate or no page size to paginate with.
func TotalPages(totalCount int, pageSize int) int {
	if totalCount <= 0 {
		return 0
	}
	if pageSize <= 0 {
		return 0
	}
	pages := totalCount / pageSize
	if totalCount%pageSize != 0 {
		pages++
	}
	return pages
}

// HasNextPage reports whether a further page after current still holds records.
// It holds when current.Number*current.Size falls short of totalCount, checked
// without multiplying to stay safe on huge pages. An invalid current page
// (Number < 1 or Size < 1) or a non-positive totalCount yields false: an empty
// store never has a next page.
func HasNextPage(current Page, totalCount int) bool {
	if current.Number < 1 {
		return false
	}
	if current.Size < 1 {
		return false
	}
	if totalCount <= 0 {
		return false
	}
	return current.Number <= (totalCount-1)/current.Size
}

// Repository stores and retrieves records addressed by an identity.
// Implementations decide ordering: List order is unspecified unless the
// implementation documents otherwise.
type Repository[Record any, Identity comparable] interface {
	// Store persists candidate, overwriting any record with the same identity.
	Store(candidate Record) error
	// Fetch returns the record held under identity, or ErrRecordNotFound.
	Fetch(identity Identity) (Record, error)
	// Remove drops the record held under identity, or ErrRecordNotFound.
	Remove(identity Identity) error
	// List returns at most Size records for the requested page.
	List(page Page) ([]Record, error)
}

// MemoryRepository is a mutex-guarded in-memory Repository for tests and
// teaching. The zero value is not usable: construct with NewMemoryRepository.
// Store overwrites records that share an identity (upsert). List order is
// unspecified: map iteration randomizes, so callers assert membership, never
// order. Deterministic ordering is the later adapter's job.
// Do not copy after use: the mutex must stay with its map.
type MemoryRepository[Record any, Identity comparable] struct {
	mutex  sync.Mutex
	stored map[Identity]Record
	keyOf  func(Record) Identity
}

// NewMemoryRepository returns an empty repository keyed by keyOf.
// keyOf extracts the identity callers address records by. A nil keyOf panics:
// supplying the extractor is a construction-time programmer error, never
// control flow.
func NewMemoryRepository[Record any, Identity comparable](keyOf func(Record) Identity) *MemoryRepository[Record, Identity] {
	if keyOf == nil {
		panic("axiom repo: NewMemoryRepository needs a non-nil keyOf")
	}
	return &MemoryRepository[Record, Identity]{stored: make(map[Identity]Record), keyOf: keyOf}
}

// Store persists candidate, overwriting any record with the same identity.
func (current *MemoryRepository[Record, Identity]) Store(candidate Record) error {
	current.mutex.Lock()
	defer current.mutex.Unlock()
	current.stored[current.keyOf(candidate)] = candidate
	return nil
}

// Fetch returns the record held under identity, or a wrapped ErrRecordNotFound.
func (current *MemoryRepository[Record, Identity]) Fetch(identity Identity) (Record, error) {
	current.mutex.Lock()
	defer current.mutex.Unlock()
	candidate, held := current.stored[identity]
	if !held {
		var empty Record
		return empty, fmt.Errorf("axiom repo: fetch %v: %w", identity, ErrRecordNotFound)
	}
	return candidate, nil
}

// Remove drops the record held under identity, or a wrapped ErrRecordNotFound.
func (current *MemoryRepository[Record, Identity]) Remove(identity Identity) error {
	current.mutex.Lock()
	defer current.mutex.Unlock()
	if _, held := current.stored[identity]; !held {
		return fmt.Errorf("axiom repo: remove %v: %w", identity, ErrRecordNotFound)
	}
	delete(current.stored, identity)
	return nil
}

// List returns the records for page in unspecified order.
// An invalid page (Number < 1 or Size < 1) yields a wrapped ErrInvalidPage.
// Listing past the end yields an empty non-nil slice and a nil failure.
func (current *MemoryRepository[Record, Identity]) List(page Page) ([]Record, error) {
	current.mutex.Lock()
	defer current.mutex.Unlock()
	collected := make([]Record, 0, len(current.stored))
	for _, candidate := range current.stored {
		collected = append(collected, candidate)
	}
	return slicePage(collected, page)
}

// slicePage carves page out of records, holding the pagination convention GORM
// and SQL adapters will reuse: 1-based Number, Offset skips, Size caps, past
// the end yields an empty non-nil slice. Arithmetic overflow lands past the
// end as well, never in a negative-index panic.
func slicePage[Record any](records []Record, page Page) ([]Record, error) {
	if page.Number < 1 {
		return nil, fmt.Errorf("axiom repo: list page number %d: %w (numbers start at 1)", page.Number, ErrInvalidPage)
	}
	if page.Size < 1 {
		return nil, fmt.Errorf("axiom repo: list page size %d: %w (sizes start at 1)", page.Size, ErrInvalidPage)
	}
	start := page.Offset()
	if start < 0 || start >= len(records) {
		return []Record{}, nil
	}
	finish := start + page.Size
	if finish > len(records) {
		finish = len(records)
	}
	return records[start:finish], nil
}
