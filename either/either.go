// Package either provides an unbiased union of two typed sides.
//
// Use Left to hold the first side and Right to hold the second side. Neither
// side is favoured: Left is not an error and Right is not a success. For a
// success-or-failure outcome use the result package; for a maybe-present
// value use the opt package. Bifold splits both sides through one outcome,
// Map transforms the Right side while carrying Left untouched, MapLeft
// transforms the Left side while carrying Right untouched, Swap exchanges the
// sides, and ToResult bridges into result.Result.
//
// Which function?
//
// | When you want to ...                  | Reach for ... |
// |---------------------------------------|-----------------|
// | Hold the first side                   | Left            |
// | Hold the second side                  | Right           |
// | Ask which side is held                | IsLeft/IsRight  |
// | Split both sides into one outcome     | Bifold          |
// | Transform the Right side              | Map             |
// | Transform the Left side               | MapLeft         |
// | Exchange the two sides                | Swap            |
// | Bridge into result.Result             | ToResult        |
//
// Worked example: a course store checks whether an expansion pack is granted.
// The check yields Either holding either the lock reason (Left) or the
// granted pack title (Right); Bifold renders one label for both paths:
//
//	locked := either.Left[string, string]("season pass missing")
//	granted := either.Right[string, string]("Frozen Depths")
//	render := func(reason string) string { return "locked: " + reason }
//	describe := func(packTitle string) string { return "ready: " + packTitle }
//	either.Bifold(locked, render, describe)   // "locked: season pass missing"
//	either.Bifold(granted, render, describe)  // "ready: Frozen Depths"
package either

import (
	"github.com/seinra/axiom/result"
)

// Either holds exactly one of two typed sides: a Left of type L or a
// Right of type R. The zero value is a Left holding the zero L value,
// so a declared Either without a constructor still reads as Left.
type Either[L, R any] struct {
	leftSide  L
	rightSide R
	isRight   bool
}

// Left holds leftSide as the Left side.
func Left[L, R any](leftSide L) Either[L, R] {
	return Either[L, R]{leftSide: leftSide}
}

// Right holds rightSide as the Right side.
func Right[L, R any](rightSide R) Either[L, R] {
	return Either[L, R]{rightSide: rightSide, isRight: true}
}

// IsLeft reports whether current holds the Left side.
func (current Either[L, R]) IsLeft() bool {
	return !current.isRight
}

// IsRight reports whether current holds the Right side.
func (current Either[L, R]) IsRight() bool {
	return current.isRight
}

// Bifold splits current into one outcome: onLeft runs for a Left,
// onRight runs for a Right. Exactly one branch runs.
func Bifold[L, R any, Outcome any](current Either[L, R], onLeft func(L) Outcome, onRight func(R) Outcome) Outcome {
	if current.isRight {
		return onRight(current.rightSide)
	}
	return onLeft(current.leftSide)
}

// Map transforms the Right side with transform, carrying a Left through untouched.
// The Rust map precedent for the favoured side, without favouring either side here.
func Map[L, R any, Transformed any](current Either[L, R], transform func(R) Transformed) Either[L, Transformed] {
	if current.isRight {
		return Right[L, Transformed](transform(current.rightSide))
	}
	return Left[L, Transformed](current.leftSide)
}

// MapLeft transforms the Left side with transform, carrying a Right through untouched.
// Mirrors Map for the other side: a Right never runs transform.
func MapLeft[L, R any, Transformed any](current Either[L, R], transform func(L) Transformed) Either[Transformed, R] {
	if current.isRight {
		return Right[Transformed, R](current.rightSide)
	}
	return Left[Transformed, R](transform(current.leftSide))
}

// Swap exchanges the two sides: a Left becomes a Right holding the same
// payload and a Right becomes a Left holding the same payload.
func (current Either[L, R]) Swap() Either[R, L] {
	if current.isRight {
		return Left[R, L](current.rightSide)
	}
	return Right[R, L](current.leftSide)
}

// ToResult bridges current into a result.Result: a Right becomes Ok, a Left
// becomes Fail with failureOf applied to the held Left side.
func ToResult[L, R any](current Either[L, R], failureOf func(L) error) result.Result[R] {
	if current.isRight {
		return result.Ok(current.rightSide)
	}
	return result.Fail[R](failureOf(current.leftSide))
}
