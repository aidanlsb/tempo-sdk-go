package rng_test

import (
	"reflect"
	"testing"

	"github.com/aidanlsb/tempo-sdk-go/rng"
)

func TestGeneratorIsDeterministicForSeed(t *testing.T) {
	a := rng.New(0xDEADBEEF)
	b := rng.New(0xDEADBEEF)
	for i := 0; i < 16; i++ {
		if a.Uint64() != b.Uint64() {
			t.Fatalf("identical seeds diverged at draw %d", i)
		}
	}
	if !reflect.DeepEqual(a.Draws(), b.Draws()) {
		t.Fatalf("identical seeds produced different draw logs")
	}
}

func TestDifferentSeedsDiverge(t *testing.T) {
	a := rng.New(1)
	b := rng.New(2)
	if a.Uint64() == b.Uint64() {
		t.Fatal("distinct seeds produced the same first word")
	}
}

func TestIntNIsBoundedAndRecorded(t *testing.T) {
	g := rng.New(42)
	for i := 0; i < 100; i++ {
		v := g.IntN(7)
		if v < 0 || v >= 7 {
			t.Fatalf("IntN(7) = %d out of range", v)
		}
	}
	draws := g.Draws()
	if len(draws) != 100 {
		t.Fatalf("recorded %d draws, want 100", len(draws))
	}
	for i, draw := range draws {
		if draw.Method != rng.MethodIntN || draw.Bound != 7 {
			t.Fatalf("draw %d = %#v, want intn bound 7", i, draw)
		}
		// The recorded raw word reduces to a result in range, so an auditor can
		// recompute the observed draw from the seed alone.
		if result := draw.Value % draw.Bound; result >= 7 {
			t.Fatalf("draw %d recomputes to %d, out of range", i, result)
		}
	}
}

func TestFloat64InUnitInterval(t *testing.T) {
	g := rng.New(7)
	for i := 0; i < 100; i++ {
		v := g.Float64()
		if v < 0 || v >= 1 {
			t.Fatalf("Float64() = %v out of [0,1)", v)
		}
	}
	if len(g.Draws()) != 100 {
		t.Fatalf("recorded %d draws, want 100", len(g.Draws()))
	}
}

func TestIntNPanicsOnNonPositive(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("IntN(0) did not panic")
		}
	}()
	rng.New(1).IntN(0)
}

func TestDrawsIsNilBeforeAnyDraw(t *testing.T) {
	if draws := rng.New(1).Draws(); draws != nil {
		t.Fatalf("Draws() = %#v before drawing, want nil", draws)
	}
}
