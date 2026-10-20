// Package set provides a typed Set[Member] for holding distinct values.
//
// Use Of to seed members and New for an empty set; the zero value is also an
// empty set ready to use. Add, Remove, and Clear mutate through a pointer;
// queries (Has, Len, IsEmpty, Equal) and combinations (Union, Intersection,
// Difference, MapValue) never mutate. Members snapshots come unordered: never
// assert order, only membership.
package set

// Set holds distinct members of type Member.
// The zero value is an empty set ready to use.
type Set[Member comparable] struct {
	members map[Member]struct{}
}

// New returns an empty set.
func New[Member comparable]() Set[Member] {
	return Set[Member]{members: make(map[Member]struct{})}
}

// Of returns a set holding members, dropping duplicates.
func Of[Member comparable](members ...Member) Set[Member] {
	present := New[Member]()
	for _, member := range members {
		present.Add(member)
	}
	return present
}

// Add holds candidate. Adding a present member changes nothing.
func (current *Set[Member]) Add(candidate Member) {
	if current.members == nil {
		current.members = make(map[Member]struct{})
	}
	current.members[candidate] = struct{}{}
}

// Remove drops candidate. Removing an absent member changes nothing.
func (current *Set[Member]) Remove(candidate Member) {
	delete(current.members, candidate)
}

// Clear drops every member, leaving an empty set.
func (current *Set[Member]) Clear() {
	current.members = make(map[Member]struct{})
}

// Has reports whether candidate is held.
func (current Set[Member]) Has(candidate Member) bool {
	_, held := current.members[candidate]
	return held
}

// Len returns the count of held members.
func (current Set[Member]) Len() int {
	return len(current.members)
}

// IsEmpty reports whether no member is held.
func (current Set[Member]) IsEmpty() bool {
	return len(current.members) == 0
}

// Equal reports whether current and other hold exactly the same members.
func (current Set[Member]) Equal(other Set[Member]) bool {
	if len(current.members) != len(other.members) {
		return false
	}
	for member := range current.members {
		if !other.Has(member) {
			return false
		}
	}
	return true
}

// Union returns the set of members held by current or other.
func (current Set[Member]) Union(other Set[Member]) Set[Member] {
	combined := New[Member]()
	for member := range current.members {
		combined.Add(member)
	}
	for member := range other.members {
		combined.Add(member)
	}
	return combined
}

// Intersection returns the set of members held by both current and other.
func (current Set[Member]) Intersection(other Set[Member]) Set[Member] {
	shared := New[Member]()
	for member := range current.members {
		if other.Has(member) {
			shared.Add(member)
		}
	}
	return shared
}

// Difference returns the set of members held by current but not by other.
func (current Set[Member]) Difference(other Set[Member]) Set[Member] {
	remaining := New[Member]()
	for member := range current.members {
		if !other.Has(member) {
			remaining.Add(member)
		}
	}
	return remaining
}

// Members returns a snapshot slice of held members in unspecified order.
func (current Set[Member]) Members() []Member {
	members := make([]Member, 0, len(current.members))
	for member := range current.members {
		members = append(members, member)
	}
	return members
}

// MapValue transforms every member, dropping transform collisions.
// Go 1.27 generic method: the method declares its own type params, so calls
// chain left to right: Of(1, 2).MapValue(describe).
func (current Set[Member]) MapValue[Transformed comparable](transform func(Member) Transformed) Set[Transformed] {
	mapped := New[Transformed]()
	for member := range current.members {
		mapped.Add(transform(member))
	}
	return mapped
}
