// Package mapx provides small typed helpers over builtin maps, layered on the
// standard library (maps, slices, cmp, iter) rather than replacing it.
//
// Deliberately function-only with no wrapper type and no generic methods:
// Catalyst and the stdlib pass builtin maps (map[string]int tallies,
// map[string]any envelopes) everywhere, and interop beats chaining here. For
// Keys/Values/Collect/Clone reach for the stdlib maps package directly. All
// helpers are pure: inputs are never mutated, and nil inputs yield empty
// non-nil maps.
//
// Which function?
//
// | When you want to ...           | Reach for ... |
// |--------------------------------|---------------|
// | Index a slice by key           | KeyBy         |
// | Group values by key            | GroupByMapped |
// | Keep only matching entries     | Filter        |
// | Keep only matching keys        | FilterKeys    |
// | Keep only matching values      | FilterValues  |
// | Transform every value          | MapValues     |
// | Remap every key                | MapKeys       |
// | Combine two maps               | Merge         |
// | Keep only listed keys          | Pick          |
// | Drop listed keys               | Omit          |
// | Flip values into key groups    | Invert        |
// | Read with an explicit fallback | ValueOr       |
// | Bridge pairs into a map          | FromEntries |
// | Split a map into pairs           | ToEntries   |
// | Keep only listed values          | PickByValues |
package mapx

import "github.com/seinra/axiom/slice"

// KeyBy indexes elements by keyOf. On key collision the last element wins.
func KeyBy[Element any, Key comparable](elements []Element, keyOf func(Element) Key) map[Key]Element {
	indexed := make(map[Key]Element, len(elements))
	for _, element := range elements {
		indexed[keyOf(element)] = element
	}
	return indexed
}

// Filter keeps the entries where keep reports true, returning a new map.
func Filter[Key comparable, Value any](entries map[Key]Value, keep func(Key, Value) bool) map[Key]Value {
	kept := make(map[Key]Value, len(entries))
	for key, value := range entries {
		if keep(key, value) {
			kept[key] = value
		}
	}
	return kept
}

// FilterKeys keeps the entries where keep reports true for the key, returning a new map.
// Nil input yields an empty non-nil map, and entries is never mutated.
// For example, FilterKeys(scores, isShortName) keeps only the short-named scores.
func FilterKeys[Key comparable, Value any](entries map[Key]Value, keep func(Key) bool) map[Key]Value {
	kept := make(map[Key]Value, len(entries))
	for heldKey, heldValue := range entries {
		if keep(heldKey) {
			kept[heldKey] = heldValue
		}
	}
	return kept
}

// FilterValues keeps the entries where keep reports true for the value, returning a new map.
// Nil input yields an empty non-nil map, and entries is never mutated.
// For example, FilterValues(scores, isPassing) keeps only the passing scores.
func FilterValues[Key comparable, Value any](entries map[Key]Value, keep func(Value) bool) map[Key]Value {
	kept := make(map[Key]Value, len(entries))
	for heldKey, heldValue := range entries {
		if keep(heldValue) {
			kept[heldKey] = heldValue
		}
	}
	return kept
}

// MapValues transforms every value with transform, keeping keys unchanged.
func MapValues[Key comparable, Value any, Transformed any](entries map[Key]Value, transform func(Value) Transformed) map[Key]Transformed {
	mapped := make(map[Key]Transformed, len(entries))
	for key, value := range entries {
		mapped[key] = transform(value)
	}
	return mapped
}

// MapKeys remaps every key through keyOf, keeping values unchanged.
// On key collision the last write wins: colliding inputs collapse into a single
// entry, and because map iteration order is unspecified the surviving value is
// one of the colliding values, never a merge. Nil input yields an empty non-nil
// map, and entries is never mutated.
// For example, MapKeys(map[string]int{"a": 1}, strings.ToUpper) yields map[string]int{"A": 1}.
func MapKeys[Key comparable, Value any, NewKey comparable](entries map[Key]Value, keyOf func(Key) NewKey) map[NewKey]Value {
	remapped := make(map[NewKey]Value, len(entries))
	for heldKey, heldValue := range entries {
		remapped[keyOf(heldKey)] = heldValue
	}
	return remapped
}

// Merge combines first and second into a new map. Keys held by one side copy
// over; keys held by both resolve through combine receiving the colliding key
// and both values.
func Merge[Key comparable, Value any](first map[Key]Value, second map[Key]Value, combine func(Key, Value, Value) Value) map[Key]Value {
	combined := make(map[Key]Value, len(first)+len(second))
	for key, value := range first {
		combined[key] = value
	}
	for key, value := range second {
		if held, collides := combined[key]; collides {
			combined[key] = combine(key, held, value)
		} else {
			combined[key] = value
		}
	}
	return combined
}

// Pick returns a new map holding only the requested keys that are present.
// Missing keys are ignored.
func Pick[Key comparable, Value any](entries map[Key]Value, keys ...Key) map[Key]Value {
	picked := make(map[Key]Value, len(keys))
	for _, key := range keys {
		if value, held := entries[key]; held {
			picked[key] = value
		}
	}
	return picked
}

// Omit returns a new map without the listed keys. Missing keys are ignored.
// It is the complement of Pick.
func Omit[Key comparable, Value any](entries map[Key]Value, keys ...Key) map[Key]Value {
	excluded := make(map[Key]struct{}, len(keys))
	for _, key := range keys {
		excluded[key] = struct{}{}
	}
	omitted := make(map[Key]Value, len(entries))
	for key, value := range entries {
		if _, skip := excluded[key]; !skip {
			omitted[key] = value
		}
	}
	return omitted
}

// Invert groups keys by their shared value. Colliding values collect into a
// slice instead of last-wins overwriting, so no key is silently dropped.
func Invert[Key comparable, Value comparable](entries map[Key]Value) map[Value][]Key {
	grouped := make(map[Value][]Key, len(entries))
	for key, value := range entries {
		grouped[value] = append(grouped[value], key)
	}
	return grouped
}

// ValueOr returns the value held for key, or fallback when key is absent.
// It mirrors the Kotlin getOrDefault precedent.
func ValueOr[Key comparable, Value any](entries map[Key]Value, key Key, fallback Value) Value {
	if held, present := entries[key]; present {
		return held
	}
	return fallback
}

// GroupByMapped groups valueOf(element) by keyOf(element).
// It mirrors the Kotlin groupBy precedent with separate key and value transforms.
func GroupByMapped[Element any, Key comparable, Value any](elements []Element, keyOf func(Element) Key, valueOf func(Element) Value) map[Key][]Value {
	grouped := make(map[Key][]Value)
	for _, element := range elements {
		key := keyOf(element)
		grouped[key] = append(grouped[key], valueOf(element))
	}
	return grouped
}

// FromEntries builds a map from pairs, the inverse of ToEntries and the
// map-side cousin of KeyBy. Zip output feeds it directly:
// FromEntries(Zip(keys, values)) indexes two slices the way KeyBy indexes one.
//
// Later pairs win on duplicate keys, mirroring the KeyBy last-wins rule.
// Nil or empty input yields an empty non-nil map.
//
// The stdlib has no pairs-to-map bridge (maps has Collect for iter.Seq2,
// slices has no pair type of its own); Zip plus a loop is the only route.
// mapx imports slice for Pair; the dependency is acyclic (either imports
// result by the same precedent).
func FromEntries[Key comparable, Value any](pairs []slice.Pair[Key, Value]) map[Key]Value {
	entries := make(map[Key]Value, len(pairs))
	for _, pair := range pairs {
		entries[pair.First] = pair.Second
	}
	return entries
}

// ToEntries splits a map back into pairs, the inverse of FromEntries.
// Pair order is unspecified: maps iterate in random order, so callers assert
// membership (or round-trip through FromEntries), never order.
//
// Nil or empty input yields an empty non-nil slice.
//
// The stdlib has no map-to-pairs helper; maps.All plus append is the only
// route, and it leaves the Pair element type to the caller.
func ToEntries[Key comparable, Value any](entries map[Key]Value) []slice.Pair[Key, Value] {
	bridged := make([]slice.Pair[Key, Value], 0, len(entries))
	for key, value := range entries {
		bridged = append(bridged, slice.Pair[Key, Value]{First: key, Second: value})
	}
	return bridged
}

// PickByValues returns a new map holding only the entries whose value is
// one of the listed values. It is the value-channel complement of Pick
// (which selects by key); duplicate listed values collapse into one
// membership test. Missing values are ignored.
//
// Nil input, or no listed values, yields an empty non-nil map. The input is
// never mutated.
//
// The stdlib has no value-membership helper; maps plus a wanted-set is the
// only route, which this folds into one call.
func PickByValues[Key comparable, Value comparable](entries map[Key]Value, values ...Value) map[Key]Value {
	wanted := make(map[Value]struct{}, len(values))
	for _, value := range values {
		wanted[value] = struct{}{}
	}
	picked := make(map[Key]Value)
	for key, value := range entries {
		if _, keep := wanted[value]; keep {
			picked[key] = value
		}
	}
	return picked
}
