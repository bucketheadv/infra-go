package numx

import (
	"math"
	"testing"
)

func TestClampAndInRange(t *testing.T) {
	if Clamp(5, 1, 10) != 5 || Clamp(0, 1, 10) != 1 || Clamp(99, 1, 10) != 10 {
		t.Fatal("Clamp() int failed")
	}
	if Clamp(5.0, 10.0, 1.0) != 5.0 {
		t.Fatal("Clamp() should normalize reversed bounds")
	}
	if !InRange(5, 1, 10) || InRange(0, 1, 10) || !InRange(10, 1, 10) {
		t.Fatal("InRange() failed")
	}
}

func TestRound(t *testing.T) {
	if Round(1.234, 2) != 1.23 || Round(1.235, 2) != 1.24 {
		t.Fatal("Round() failed")
	}
	if Round(1.2, -1) != 1.2 {
		t.Fatal("Round() negative places")
	}
}

func TestPercent(t *testing.T) {
	if Percent(25, 100) != 25 || Percent(1, 4) != 25 {
		t.Fatal("Percent() failed")
	}
	if Percent(1, 0) != 0 {
		t.Fatal("Percent() zero total")
	}
}

func TestApproximatelyEqual(t *testing.T) {
	if !ApproximatelyEqual(1.0, 1.0000001, 0.001) {
		t.Fatal("ApproximatelyEqual() should be true")
	}
	if ApproximatelyEqual(1.0, 2.0, 0.001) {
		t.Fatal("ApproximatelyEqual() should be false")
	}
	if ApproximatelyEqual(math.NaN(), math.NaN(), 0.001) {
		t.Fatal("ApproximatelyEqual(NaN, NaN) should be false")
	}
}

func TestFloatSpecials(t *testing.T) {
	if !math.IsNaN(Clamp(math.NaN(), 0.0, 1.0)) {
		t.Fatal("Clamp(NaN) should return NaN")
	}
	if InRange(math.NaN(), 0.0, 1.0) {
		t.Fatal("InRange(NaN) should be false")
	}
}
