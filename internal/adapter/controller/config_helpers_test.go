//go:build test

package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

func intPtr(v int) *int    { return &v }
func boolPtr(v bool) *bool { return &v }

func TestApplyIntIfGte(t *testing.T) {
	tests := []struct {
		name     string
		dst      int
		src      *int
		minVal   int
		expected int
	}{
		{"nil src unchanged", 10, nil, 1, 10},
		{"below min unchanged", 10, intPtr(0), 1, 10},
		{"at min applied", 10, intPtr(1), 1, 1},
		{"above min applied", 10, intPtr(50), 1, 50},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := tt.dst
			applyIntIfGte(&dst, tt.src, tt.minVal)
			assert.Equal(t, tt.expected, dst)
		})
	}
}

func TestApplyBool(t *testing.T) {
	tests := []struct {
		name     string
		dst      bool
		src      *bool
		expected bool
	}{
		{"nil src unchanged", false, nil, false},
		{"set true", false, boolPtr(true), true},
		{"set false", true, boolPtr(false), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := tt.dst
			applyBool(&dst, tt.src)
			assert.Equal(t, tt.expected, dst)
		})
	}
}

func TestApplyBettingLimit(t *testing.T) {
	tests := []struct {
		name     string
		src      *int
		expected domain.BettingLimitType
	}{
		{"nil unchanged", nil, domain.BettingLimitFixed},
		{"clamp negative to 0", intPtr(-1), domain.BettingLimitFixed},
		{"valid 1", intPtr(1), domain.BettingLimitPotLimit},
		{"clamp above 2 to 2", intPtr(5), domain.BettingLimitNoLimit},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := domain.BettingLimitFixed
			applyBettingLimit(&dst, tt.src)
			assert.Equal(t, tt.expected, dst)
		})
	}
}

func TestApplyRebuyConfig(t *testing.T) {
	enabled, maxCount, chips, period := false, 0, 0, 0
	applyRebuyConfig(&enabled, &maxCount, &chips, &period,
		boolPtr(true), intPtr(5), intPtr(1000), intPtr(20))
	assert.True(t, enabled)
	assert.Equal(t, 5, maxCount)
	assert.Equal(t, 1000, chips)
	assert.Equal(t, 20, period)
}

func TestApplyAddonConfig(t *testing.T) {
	enabled, chips, afterHand := false, 0, 0
	applyAddonConfig(&enabled, &chips, &afterHand,
		boolPtr(true), intPtr(1500), intPtr(25))
	assert.True(t, enabled)
	assert.Equal(t, 1500, chips)
	assert.Equal(t, 25, afterHand)
}

func TestValidateAndApplyBlinds(t *testing.T) {
	tests := []struct {
		name      string
		sbPtr     *int
		bbPtr     *int
		defaultBB int
		wantSB    int
		wantBB    int
		wantErr   bool
	}{
		{"both nil unchanged", nil, nil, 10, 5, 10, false},
		{"sb only below default bb", intPtr(3), nil, 10, 3, 10, false},
		{"sb only at default bb auto-adjusts bb", intPtr(10), nil, 10, 10, 20, false},
		{"bb only auto-adjusts sb", nil, intPtr(20), 10, 10, 20, false},
		{"both provided valid", intPtr(5), intPtr(20), 10, 5, 20, false},
		{"sb >= bb error", intPtr(10), intPtr(10), 10, 0, 0, true},
		{"sb > bb error", intPtr(20), intPtr(10), 10, 0, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb, bb := 5, 10
			err := validateAndApplyBlinds(&sb, &bb, tt.sbPtr, tt.bbPtr, tt.defaultBB)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantSB, sb)
				assert.Equal(t, tt.wantBB, bb)
			}
		})
	}
}
