package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestOldMaid_cpuSelectCardIdx_AllBranches covers all branches of cpuSelectCardIdx:
// - size <= 1 returns 0
// - 30% chance of edge selection (first or last card)
// - 70% chance of random selection
func TestOldMaid_cpuSelectCardIdx_AllBranches(t *testing.T) {
	// Test size <= 1 branch
	assert.Equal(t, 0, cpuSelectCardIdx(0))
	assert.Equal(t, 0, cpuSelectCardIdx(1))

	// Run many times with size > 1 to statistically hit all branches
	size := 10
	hitFirst := false  // return 0 from edge select (line 260)
	hitLast := false   // return size-1 from edge select (line 262)
	hitRandom := false // return rand.Intn(size) (line 264)

	for i := 0; i < 5000; i++ {
		idx := cpuSelectCardIdx(size)
		if idx == 0 {
			hitFirst = true
		} else if idx == size-1 {
			hitLast = true
		} else {
			hitRandom = true
		}
		if hitFirst && hitLast && hitRandom {
			break
		}
	}
	assert.True(t, hitFirst, "should hit first card (edge select)")
	assert.True(t, hitLast, "should hit last card (edge select)")
	assert.True(t, hitRandom, "should hit random middle card")
}
