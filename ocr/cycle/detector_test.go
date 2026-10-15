package cycle

import (
	"bytes"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"
)

func TestDetectCycling_NoCycle(t *testing.T) {
	tokens := make([]string, 500)
	for i := range tokens {
		tokens[i] = string(rune('a' + i%26))
	}

	repeats := countRepeats(tokens, 10)
	if repeats >= 3 {
		t.Errorf("expected no cycle, got repeats=%d", repeats)
	}
}

func TestDetectCycling_SimpleCycle(t *testing.T) {
	tokens := make([]string, 400)
	for i := 0; i < 370; i++ {
		tokens[i] = string(rune('a' + i%26))
	}
	pattern := []string{"cycle", "token", "1", "2", "3", "4", "5", "6", "7", "8"}
	for i := 0; i < 3; i++ {
		copy(tokens[370+i*10:370+(i+1)*10], pattern)
	}

	repeats := countRepeats(tokens, 10)
	if repeats < 3 {
		t.Errorf("expected >=3 repeats, got %d", repeats)
	}
}

func TestDetectCycling_BelowMinTokens(t *testing.T) {
	tokens := make([]string, 50)
	for i := 0; i < 50; i++ {
		tokens[i] = "repeated"
	}

	repeats := countRepeats(tokens, 1)
	if repeats != 50 {
		t.Errorf("expected 50 repeats for identical tokens, got %d", repeats)
	}
}

func TestDetectCycling_MaxPatternLen(t *testing.T) {
	totalTokens := 500
	tokens := make([]string, totalTokens)
	longPattern := make([]string, 130)
	for i := range longPattern {
		longPattern[i] = string(rune('a' + i%26))
	}

	for i := 0; i < 110; i++ {
		tokens[i] = string(rune('z' - i%26))
	}

	for i := 0; i < 3; i++ {
		copy(tokens[110+i*130:110+(i+1)*130], longPattern)
	}

	repeats := countRepeats(tokens, 130)
	// countRepeats doesn't enforce maxPatternLen (that's done in findCycle),
	// but with our data the first chunk at index 0 won't match the pattern,
	// so we get exactly 3 repeats (the three occurrences at the end).
	if repeats != 3 {
		t.Errorf("expected 3 repeats for pattern at end of buffer, got %d", repeats)
	}
}

func TestDetectCycling_PartialCycle(t *testing.T) {
	tokens := make([]string, 400)
	for i := 0; i < 394; i++ {
		tokens[i] = string(rune('a' + i%26))
	}
	pattern := []string{"x", "y", "z"}

	copy(tokens[394:397], pattern)
	copy(tokens[397:400], pattern)

	repeats := countRepeats(tokens, 3)
	if repeats >= 3 {
		t.Errorf("expected no detection for 2 repeats, got %d", repeats)
	}
}

func TestDetector_AddToken(t *testing.T) {
	cfg := DefaultConfig()
	cfg.CheckInterval = 100
	cfg.MinTokens = 400
	detector := NewDetector(cfg)

	for i := 0; i < 300; i++ {
		detector.AddToken("x")
	}

	if detector.IsCycling() {
		t.Error("should not be cycling below minTokens")
	}
}

func TestDetector_DetectCycle(t *testing.T) {
	cfg := DefaultConfig()
	detector := NewDetector(cfg)

	for i := 0; i < 370; i++ {
		detector.AddToken(string(rune('a' + i%26)))
	}

	pattern := []string{"cycle", "token", "1", "2", "3", "4", "5", "6", "7", "8"}
	for i := 0; i < 3; i++ {
		for _, tok := range pattern {
			detector.AddToken(tok)
		}
	}

	time.Sleep(200 * time.Millisecond)

	if !detector.IsCycling() {
		t.Error("expected cycle detection, but IsCycling() is false")
	}

	found, patternLen, repeats, cycleStart := detector.GetCycleInfo()
	if !found {
		t.Error("expected cycle to be found")
	}
	if patternLen != 10 {
		t.Errorf("expected patternLen=10, got %d", patternLen)
	}
	if repeats < 3 {
		t.Errorf("expected repeats>=3, got %d", repeats)
	}
	if cycleStart != 370 {
		t.Errorf("expected cycleStart=370, got %d", cycleStart)
	}
}

func TestDetector_GetCleanTokens(t *testing.T) {
	cfg := DefaultConfig()
	detector := NewDetector(cfg)

	for i := 0; i < 480; i++ {
		detector.AddToken(fmt.Sprintf("unique_%d", i))
	}

	// Use synchronous Detect() instead of async AddToken + sleep to avoid race.
	detector.Detect()

	clean := detector.GetCleanTokens()

	if len(clean) != 480 {
		t.Errorf("expected 480 clean tokens, got %d", len(clean))
	}
	if clean[0] != "unique_0" {
		t.Errorf("expected first token to be 'unique_0', got '%s'", clean[0])
	}
}

func TestDetector_NoCycleGetAllTokens(t *testing.T) {
	cfg := DefaultConfig()
	detector := NewDetector(cfg)

	for i := 0; i < 500; i++ {
		detector.AddToken(fmt.Sprintf("token_%d", i))
	}

	time.Sleep(300 * time.Millisecond)

	if detector.IsCycling() {
		t.Error("should not be cycling")
	}

	clean := detector.GetCleanTokens()
	if len(clean) != 500 {
		t.Errorf("expected 500 tokens, got %d", len(clean))
	}
}

func TestGetCleanTokens_AllTokensAreCycle(t *testing.T) {
	// Issue #16: when cycling=true and cycleStart==0, GetCleanTokens
	// should return empty slice (all tokens are part of the cycle).
	detector := NewDetector(DefaultConfig())

	// 30 tokens: 10 unique + 20 tokens forming 2 repeats of a 10-token pattern.
	for i := 0; i < 10; i++ {
		detector.AddToken(fmt.Sprintf("unique_%d", i))
	}
	pattern := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	for i := 0; i < 2; i++ {
		for _, tok := range pattern {
			detector.AddToken(tok)
		}
	}

	// Manually set cycling state to simulate full-buffer cycle.
	detector.mu.Lock()
	detector.cycling = true
	detector.cycleStart = 0 // all tokens are cycle
	detector.patternLen = 10
	detector.repeats = 2
	detector.mu.Unlock()

	clean := detector.GetCleanTokens()
	if len(clean) != 0 {
		t.Errorf("expected 0 clean tokens when cycleStart==0, got %d", len(clean))
	}
}

func TestDetector_MaxTokensCap(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxTokens = 1000
	detector := NewDetector(cfg)

	for i := 0; i < 2000; i++ {
		detector.AddToken("x")
	}

	detector.mu.RLock()
	count := len(detector.tokens)
	detector.mu.RUnlock()

	if count != 1000 {
		t.Errorf("expected max 1000 tokens, got %d", count)
	}
}

func TestDetector_ConcurrentAdd(t *testing.T) {
	cfg := DefaultConfig()
	detector := NewDetector(cfg)

	var wg sync.WaitGroup
	for goroutine := 0; goroutine < 10; goroutine++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				detector.AddToken(string(rune('a' + i%26)))
			}
		}()
	}

	wg.Wait()

	detector.mu.RLock()
	count := len(detector.tokens)
	detector.mu.RUnlock()

	if count == 0 {
		t.Error("expected some tokens to be collected")
	}
}

func TestCountRepeats_FirstChunkEvaluated(t *testing.T) {
	// Regression test for off-by-one: when i reaches 0, the first chunk
	// must still be compared against the pattern (issue #4).
	tokens := []string{"x", "y", "z", "x", "y", "z"}
	repeats := countRepeats(tokens, 3)
	if repeats != 2 {
		t.Errorf("expected 2 repeats for [xyz][xyz], got %d", repeats)
	}
}

func TestCountRepeats_FirstChunkNonMatching(t *testing.T) {
	// First chunk doesn't match pattern - should return 1 (pattern only).
	tokens := []string{"a", "b", "c", "d", "e", "f"}
	repeats := countRepeats(tokens, 3)
	if repeats != 1 {
		t.Errorf("expected 1 repeat when first chunk doesn't match, got %d", repeats)
	}
}

func TestCountRepeats_ExactTwoPatterns(t *testing.T) {
	// Two identical patterns covering the entire buffer.
	tokens := []string{"x", "y", "z", "x", "y", "z"}
	repeats := countRepeats(tokens, 3)
	if repeats != 2 {
		t.Errorf("expected 2 repeats for exact two patterns, got %d", repeats)
	}
}

func TestEqualChunks(t *testing.T) {
	a := []string{"a", "b", "c"}
	b := []string{"a", "b", "c"}
	c := []string{"a", "b", "d"}
	d := []string{"a", "b"}

	if !equalChunks(a, b) {
		t.Error("equal slices should return true")
	}
	if equalChunks(a, c) {
		t.Error("different slices should return false")
	}
	if equalChunks(a, d) {
		t.Error("different lengths should return false")
	}
}

func TestCompleteHTMLWithContent_UnclosedTags(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-unclosed-*.html")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	unclosedHTML := `<body><div data-bbox="100 200 900 300"><p>Hello</p><div data-bbox="100 400 900 500"><p>World</p></div>`
	if _, err := tmpFile.WriteString(unclosedHTML); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	if err := CompleteHTMLWithContent(tmpFile.Name(), false); err != nil {
		t.Fatalf("completeHTMLWithContent failed: %v", err)
	}

	completed, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	completedStr := string(completed)
	if !bytes.Contains(completed, []byte("</body>")) {
		t.Errorf("expected </body> in completed HTML, got: %s", completedStr)
	}
}

func TestCompleteHTMLWithContent_ValidHTML(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-valid-*.html")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	validHTML := `<html><body><div data-bbox="100 200 900 300"><p>Hello</p></div></body></html>`
	if _, err := tmpFile.WriteString(validHTML); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	if err := CompleteHTMLWithContent(tmpFile.Name(), false); err != nil {
		t.Fatalf("completeHTMLWithContent failed: %v", err)
	}

	completed, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	if len(completed) == 0 {
		t.Error("expected non-empty completed HTML")
	}
}

func TestCompleteHTMLWithContent_BareElements(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-bare-*.html")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	bareHTML := `<div data-bbox="100 200 900 300"><p>Hello</p></div>`
	if _, err := tmpFile.WriteString(bareHTML); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	if err := CompleteHTMLWithContent(tmpFile.Name(), false); err != nil {
		t.Fatalf("completeHTMLWithContent failed: %v", err)
	}

	completed, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	completedStr := string(completed)
	if !bytes.Contains(completed, []byte("<body>")) {
		t.Errorf("expected <body> wrapper in completed HTML, got: %s", completedStr)
	}
}

func TestDetector_DetectionTimeout(t *testing.T) {
	cfg := DefaultConfig()
	detector := NewDetector(cfg)

	// AddToken is now synchronous (no goroutine-per-check), so 1000 calls should complete quickly.
	start := time.Now()
	for i := 0; i < 1000; i++ {
		detector.AddToken(string(rune('a' + i%26)))
	}
	elapsed := time.Since(start)

	if elapsed > 5*time.Second {
		t.Fatalf("AddToken took too long: %v", elapsed)
	}
}

func BenchmarkCountRepeats(b *testing.B) {
	tokens := make([]string, 1000)
	for i := range tokens {
		tokens[i] = string(rune('a' + i%26))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		countRepeats(tokens, 10)
	}
}

func BenchmarkDetector_AddToken(b *testing.B) {
	cfg := DefaultConfig()
	detector := NewDetector(cfg)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.AddToken("test-token")
	}
}
