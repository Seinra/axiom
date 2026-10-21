// Package either provides an unbiased union of two typed sides.
//
// Use Left to hold the first side and Right to hold the second side. Neither
// side is favoured: Left is not an error and Right is not a success. For a
// success-or-failure outcome use the result package; for a maybe-present
// value use the opt package. Match splits both sides through one outcome,
// MapRight transforms the Right side while carrying Left untouched, MapLeft
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
// | Split both sides into one outcome     | Match           |
// | Transform the Right side              | MapRight        |
// | Transform the Left side               | MapLeft         |
// | Exchange the two sides                | Swap            |
// | Bridge into result.Result             | ToResult        |
//
// Worked example: a course store checks whether an expansion pack is granted.
// The check yields Either holding either the lock reason (Left) or the
// granted pack title (Right); Match renders one label for both paths:
//
//	locked := either.Left[string, string]("season pass missing")
//	granted := either.Right[string, string]("Frozen Depths")
//	render := func(reason string) string { return "locked: " + reason }
//	describe := func(packTitle string) string { return "ready: " + packTitle }
//	either.Match(locked, render, describe)   // "locked: season pass missing"
//	either.Match(granted, render, describe)  // "ready: Frozen Depths"
package either

import (
	"github.com/seinra/axiom/result"
)

// Either holds exactly one of two typed sides: a Left of type LeftT or a
// Right of type RightT. The zero value is a Left holding the zero LeftT value,
// so a declared Either without a constructor still reads as Left.
type Either[LeftT, RightT any] struct {
	leftSide  LeftT
	rightSide RightT
	isRight   bool
}

// Left holds leftSide as the Left side.
func Left[LeftT, RightT any](leftSide LeftT) Either[LeftT, RightT] {
	return Either[LeftT, RightT]{leftSide: leftSide}
}

// Right holds rightSide as the Right side.
func Right[LeftT, RightT any](rightSide RightT) Either[LeftT, RightT] {
	return Either[LeftT, RightT]{rightSide: rightSide, isRight: true}
}

// IsLeft reports whether current holds the Left side.
func (current Either[LeftT, RightT]) IsLeft() bool {
	return !current.isRight
}

// IsRight reports whether current holds the Right side.
func (current Either[LeftT, RightT]) IsRight() bool {
	return current.isRight
}

// Match splits current into one outcome: onLeft runs for a Left,
// onRight runs for a Right. Exactly one branch runs.
func Match[LeftT, RightT any, Outcome any](current Either[LeftT, RightT], onLeft func(LeftT) Outcome, onRight func(RightT) Outcome) Outcome {
	if current.isRight {
		return onRight(current.rightSide)
	}
	return onLeft(current.leftSide)
}

// MapRight transforms the Right side with transform, carrying a Left through untouched.
// The Rust map precedent for the favoured side, without favouring either side here.
func MapRight[LeftT, RightT any, Transformed any](current Either[LeftT, RightT], transform func(RightT) Transformed) Either[LeftT, Transformed] {
	if current.isRight {
		return Right[LeftT, Transformed](transform(current.rightSide))
	}
	return Left[LeftT, Transformed](current.leftSide)
}

// MapLeft transforms the Left side with transform, carrying a Right through untouched.
// Mirrors MapRight for the other side: a Right never runs transform.
func MapLeft[LeftT, RightT any, Transformed any](current Either[LeftT, RightT], transform func(LeftT) Transformed) Either[Transformed, RightT] {
	if current.isRight {
		return Right[Transformed, RightT](current.rightSide)
	}
	return Left[Transformed, RightT](transform(current.leftSide))
}

// Swap exchanges the two sides: a Left becomes a Right holding the same
// payload and a Right becomes a Left holding the same payload.
func (current Either[LeftT, RightT]) Swap() Either[RightT, LeftT] {
	if current.isRight {
		return Left[RightT, LeftT](current.rightSide)
	}
	return Right[RightT, LeftT](current.leftSide)
}

// ToResult bridges current into a result.Result: a Right becomes Ok, a Left
// becomes Fail with failureOf applied to the held Left side.
func ToResult[LeftT, RightT any](current Either[LeftT, RightT], failureOf func(LeftT) error) result.Result[RightT] {
	if current.isRight {
		return result.Ok(current.rightSide)
	}
	return result.Fail[RightT](failureOf(current.leftSide))
}
