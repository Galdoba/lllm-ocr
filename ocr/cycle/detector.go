package cycle

import (
	"log"
	"sync"
)

var maxTokensZeroWarn sync.Once

var checkIntervalZeroWarn sync.Once

// Detector monitors whitespace-delimited tokens for repeating patterns.
// Thread-safe via sync.RWMutex.
type Detector struct {
	mu            sync.RWMutex
	tokens        []string
	checkInterval int
	minTokens     int
	maxPatternLen int
	minRepeats    int
	maxTokens     int
	cycling       bool
	patternLen    int
	repeats       int
	cycleStart    int
	lastCheckedIdx int
}

// NewDetector creates a Detector initialized with the given Config.
func NewDetector(cfg Config) *Detector {
	return &Detector{
		tokens:        make([]string, 0, 2000),
		checkInterval: cfg.CheckInterval,
		minTokens:     cfg.MinTokens,
		maxPatternLen: cfg.MaxPatternLen,
		minRepeats:    cfg.MinRepeats,
		maxTokens:     cfg.MaxTokens,
	}
}

// AddToken appends a token to the detector's buffer and triggers cycle detection at configured intervals.
func (d *Detector) AddToken(token string) {
	d.mu.Lock()

	if d.maxTokens == 0 {
		d.mu.Unlock()
		maxTokensZeroWarn.Do(func() {
			log.Println("cycle detection disabled: maxTokens is 0")
		})
		return
	}

	if d.checkInterval <= 0 {
		d.mu.Unlock()
		checkIntervalZeroWarn.Do(func() {
			log.Println("cycle detection disabled: checkInterval is 0 or negative")
		})
		return
	}

	if len(d.tokens) >= d.maxTokens {
		d.mu.Unlock()
		return
	}

	d.tokens = append(d.tokens, token)

	if len(d.tokens) >= d.minTokens && len(d.tokens)%d.checkInterval == 0 {
		if d.lastCheckedIdx != len(d.tokens) {
			d.lastCheckedIdx = len(d.tokens)
			// Release lock before detection to avoid deadlock (detectCycling needs RLock).
			d.mu.Unlock()
			d.detectCycling()
			d.mu.Lock()
		}
	}

	d.mu.Unlock()
}

// findCycle searches tokens for repeating patterns of length 3 to maxPatternLen with at least minRepeats consecutive occurrences.
func findCycle(tokens []string, minTokens, maxPatternLen, minRepeats int) (found bool, patternLen, repeats, cycleStart int) {
	if len(tokens) < minTokens {
		return false, 0, 0, 0
	}

	actualMax := maxPatternLen
	if len(tokens)/2 < actualMax {
		actualMax = len(tokens) / 2
	}

	for pLen := 3; pLen <= actualMax; pLen++ {
		r := countRepeats(tokens, pLen)
		if r >= minRepeats {
			return true, pLen, r, len(tokens) - pLen*r
		}
	}
	return false, 0, 0, 0
}

// detectCycling runs findCycle on a snapshot of tokens and updates the detector state if a cycle is found.
func (d *Detector) detectCycling() {
	d.mu.RLock()
	tokens := make([]string, len(d.tokens))
	copy(tokens, d.tokens)
	d.mu.RUnlock()

	found, pLen, reps, cStart := findCycle(tokens, d.minTokens, d.maxPatternLen, d.minRepeats)
	if found {
		d.mu.Lock()
		d.cycling = true
		d.patternLen = pLen
		d.repeats = reps
		d.cycleStart = cStart
		d.mu.Unlock()
	}
}

// Detect synchronously runs cycle detection on a snapshot of the current tokens
// and updates the detector state if a cycle is found. Useful for tests that need
// deterministic behavior without async goroutines.
func (d *Detector) Detect() (found bool, patternLen, repeats, cycleStart int) {
	d.mu.Lock()
	tokens := make([]string, len(d.tokens))
	copy(tokens, d.tokens)
	d.mu.Unlock()

	found, pLen, reps, cStart := findCycle(tokens, d.minTokens, d.maxPatternLen, d.minRepeats)
	if found {
		d.mu.Lock()
		d.cycling = true
		d.patternLen = pLen
		d.repeats = reps
		d.cycleStart = cStart
		d.mu.Unlock()
	}
	return found, pLen, reps, cStart
}

// countRepeats counts how many consecutive times the last patternLen tokens repeat backwards from the end of the slice.
func countRepeats(tokens []string, patternLen int) int {
	end := len(tokens)
	if end < patternLen*2 {
		return 0
	}

	pattern := tokens[end-patternLen : end]
	repeats := 1

	for i := end - patternLen*2; i >= 0; i -= patternLen {
		chunk := tokens[i : i+patternLen]
		if !equalChunks(chunk, pattern) {
			break
		}
		repeats++
	}

	return repeats
}

// equalChunks returns true if two string slices have identical contents in the same order.
func equalChunks(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// IsCycling returns true if a repeating token pattern has been detected.
func (d *Detector) IsCycling() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.cycling
}

// GetCleanTokens returns all tokens, removing the repeating cycle portion if one was detected.
func (d *Detector) GetCleanTokens() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if !d.cycling {
		result := make([]string, len(d.tokens))
		copy(result, d.tokens)
		return result
	}

	// cycling is true; if cycleStart == 0, all tokens are part of the cycle.
	result := make([]string, d.cycleStart)
	copy(result, d.tokens[:d.cycleStart])
	return result
}

// GetCycleInfo returns cycle detection status and details (pattern length, repeat count, cycle start index).
func (d *Detector) GetCycleInfo() (found bool, patternLen, repeats, cycleStart int) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.cycling, d.patternLen, d.repeats, d.cycleStart
}
