// Package rng provides Tempo's fiction-neutral, recordable pseudo-random number
// generator.
//
// A Generator is a deterministic PRNG keyed by a single uint64 Action-level
// seed. It is the shared draw contract used by both the in-process physics Host
// and Go-authored WASM guests, so an identical seed always yields an identical,
// ordered sequence of Draws regardless of which transport executed the handler.
//
// Every draw is recorded as a Draw for audit. Tempo commits the seed and the
// ordered Draws onto the resulting Action record; replay never re-rolls and
// never re-executes guest RNG, it only folds the already-recorded Effects. Live
// turns stay non-deterministic (a fresh seed is minted per Action from OS
// entropy unless a test injects one); committed history stays deterministic.
package rng

// Draw is one recorded pull from a Generator, ordered by draw sequence.
//
// Value is the raw 64-bit PRNG word that was consumed; Method records how it
// was interpreted and Bound carries the argument for bounded methods (IntN).
// Recording the raw word plus the interpretation makes each derived result
// independently recomputable during an audit.
type Draw struct {
	Method string `json:"method"`
	Bound  uint64 `json:"bound,omitempty"`
	Value  uint64 `json:"value"`
}

// Draw method identifiers.
const (
	MethodUint64  = "uint64"
	MethodIntN    = "intn"
	MethodFloat64 = "float64"
)

// Generator is a deterministic, recordable PRNG seeded by a uint64.
//
// The underlying algorithm is SplitMix64: portable, fixed-width, and identical
// across the in-process and WASM paths. A Generator is not safe for concurrent
// use; each handler invocation gets its own Generator.
type Generator struct {
	state uint64
	seed  uint64
	draws []Draw
}

// New returns a Generator keyed by seed.
func New(seed uint64) *Generator {
	return &Generator{state: seed, seed: seed}
}

// Seed returns the seed the Generator was created with.
func (g *Generator) Seed() uint64 {
	return g.seed
}

// next advances the SplitMix64 state and returns the next 64-bit word.
func (g *Generator) next() uint64 {
	g.state += 0x9E3779B97F4A7C15
	z := g.state
	z = (z ^ (z >> 30)) * 0xBF58476D1CE4E5B9
	z = (z ^ (z >> 27)) * 0x94D049BB133111EB
	return z ^ (z >> 31)
}

// Uint64 returns the next 64-bit word and records the draw.
func (g *Generator) Uint64() uint64 {
	value := g.next()
	g.draws = append(g.draws, Draw{Method: MethodUint64, Value: value})
	return value
}

// IntN returns a pseudo-random integer in [0, n) and records the draw.
//
// n must be positive; IntN panics otherwise, matching math/rand/v2 semantics.
// The reduction uses a simple modulo, which carries a negligible bias for the
// small bounds v1 dice-style draws use; the raw word is recorded so an auditor
// can recompute the result exactly.
func (g *Generator) IntN(n int64) int64 {
	if n <= 0 {
		panic("rng: IntN requires n > 0")
	}
	value := g.next()
	g.draws = append(g.draws, Draw{
		Method: MethodIntN,
		Bound:  uint64(n),
		Value:  value,
	})
	return int64(value % uint64(n))
}

// Float64 returns a pseudo-random float in [0, 1) and records the draw.
//
// It uses the top 53 bits of the next word so the result has full float64
// mantissa precision.
func (g *Generator) Float64() float64 {
	value := g.next()
	g.draws = append(g.draws, Draw{Method: MethodFloat64, Value: value})
	return float64(value>>11) / (1 << 53)
}

// Draws returns the ordered draws recorded so far. The returned slice is a copy
// and is safe for the caller to retain or serialize.
func (g *Generator) Draws() []Draw {
	if len(g.draws) == 0 {
		return nil
	}
	out := make([]Draw, len(g.draws))
	copy(out, g.draws)
	return out
}
