//go:build test

package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGongZhuExposeActionLogSplitsEmptyAndCardLabels(t *testing.T) {
	for _, tt := range []struct {
		name       string
		exposure   GongZhuExposure
		code       string
		wantParams map[string]string
	}{
		{"none", GongZhuExposure{}, "gongzhu.log.exposeNone", map[string]string{"round": "3"}},
		{"cards", GongZhuExposure{Pig: true, Sheep: true}, "gongzhu.log.exposeCards", map[string]string{"round": "3", "cards": "♠Q, ♦J"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := NewDefaultGongZhu()
			g.SetRoundNumber(3)
			g.exposed = tt.exposure
			g.exposeReady = [GongZhuPlayerCnt]bool{true, true, true, true}
			g.ExecuteExpose()
			entry := g.actionLog[len(g.actionLog)-1]
			assert.Equal(t, tt.code, entry.DetailCode)
			assert.Equal(t, tt.wantParams, entry.DetailParams)
		})
	}
}
