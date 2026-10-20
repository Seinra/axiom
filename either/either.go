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

// Either holds exactly one of two typed sides: a Left of type Left or a
// Right of type Right. The zero value is a Left holding the zero Left value,
// so a declared Either without a constructor still reads as Left.
type Either[Left, Right any] struct {
	leftSide  Left
	rightSide Right
	isRight   bool
}

// Left holds leftSide as the Left side.
func Left[Left, Right any](leftSide Left) Either[Left, Right] {
	return Either[Left, Right]{leftSide: leftSide}
}

// Right holds rightSide as the Right side.
func Right[Left, Right any](rightSide Right) Either[Left, Right] {
	return Either[Left, Right]{rightSide: rightSide, isRight: true}
}

// IsLeft reports whether current holds the Left side.
func (current Either[Left, Right]) IsLeft() bool {
	return !current.isRight
}

// IsRight reports whether current holds the Right side.
func (current Either[Left, Right]) IsRight() bool {
	return current.isRight
}

// Bifold splits current into one outcome: onLeft runs for a Left,
// onRight runs for a Right. Exactly one branch runs.
func Bifold[Left, Right any, Outcome any](current Either[Left, Right], onLeft func(Left) Outcome, onRight func(Right) Outcome) Outcome {
	if current.isRight {
		return onRight(current.rightSide)
	}
	return onLeft(current.leftSide)
}

// Map transforms the Right side with transform, carrying a Left through untouched.
// The Rust map precedent for the favoured side, without favouring either side here.
func Map[Left, Right any, Transformed any](current Either[Left, Right], transform func(Right) Transformed) Either[Left, Transformed] {
	if current.isRight {
		return Either[Left, Transformed]{rightSide: transform(current.rightSide), isRight: true}
	}
	return Either[Left, Transformed]{leftSide: current.leftSide}
}

// MapLeft transforms the Left side with transform, carrying a Right through untouched.
// Mirrors Map for the other side: a Right never runs transform.
func MapLeft[Left, Right any, Transformed any](current Either[Left, Right], transform func(Left) Transformed) Either[Transformed, Right] {
	if current.isRight {
		return Either[Transformed, Right]{rightSide: current.rightSide, isRight: true}
	}
	return Either[Transformed, Right]{leftSide: transform(current.leftSide)}
}

// Swap exchanges the two sides: a Left becomes a Right holding the same
// payload and a Right becomes a Left holding the same payload.
func (current Either[Left, Right]) Swap() Either[Right, Left] {
	if current.isRight {
		return Either[Right, Left]{leftSide: current.rightSide}
	}
	return Either[Right, Left]{rightSide: current.leftSide, isRight: true}
}

// ToResult bridges current into a result.Result: a Right becomes Ok, a Left
// becomes Fail with failureOf applied to the held Left side.
func ToResult[Left, Right any](current Either[Left, Right], failureOf func(Left) error) result.Result[Right] {
	if current.isRight {
		return result.Ok(current.rightSide)
	}
	return result.Fail[Right](failureOf(current.leftSide))
}
