//go:build test
// +build test

package domain

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestADR0040MigratedGameLogsCarryCodesAndParams(t *testing.T) {
	tests := []struct {
		name   string
		code   string
		params map[string]string
		add    func()
		logs   func() []*ActionLogEntry
	}{
		{"Guandan", "guandan.log.deal", map[string]string{"hand": "2", "level": "5"}, func() {
			g := NewDefaultGuandan()
			g.addLog(-1, "deal", "guandan.log.deal", map[string]string{"hand": "2", "level": "5"}, nil)
			guandanMigrationLogs = g.GetActionLog()
		}, func() []*ActionLogEntry { return guandanMigrationLogs }},
		{"ShengJi", "shengji.log.trickWon", map[string]string{"points": "10"}, func() {
			g := NewDefaultShengJi()
			g.addLog(1, "trickWon", "shengji.log.trickWon", map[string]string{"points": "10"}, nil)
			shengJiMigrationLogs = g.GetActionLog()
		}, func() []*ActionLogEntry { return shengJiMigrationLogs }},
		{"SixBidSolo", "sixbidsolo.log.settle", map[string]string{"points": "61", "target": "60"}, func() {
			g := NewDefaultSixBidSolo()
			g.addLog(0, "settle", "sixbidsolo.log.settle", map[string]string{"points": "61", "target": "60"}, nil)
			sixBidSoloMigrationLogs = g.GetActionLog()
		}, func() []*ActionLogEntry { return sixBidSoloMigrationLogs }},
		{"Karnoffel", "karnoffel.log.deal", map[string]string{"hand": "1", "suit": "3"}, func() {
			g := NewDefaultKarnoffel()
			g.addLog(-1, "deal", "karnoffel.log.deal", map[string]string{"hand": "1", "suit": "3"}, nil)
			karnoffelMigrationLogs = g.GetActionLog()
		}, func() []*ActionLogEntry { return karnoffelMigrationLogs }},
		{"Literature", "literature.log.ask", map[string]string{"to": "1", "suit": "2", "rank": "11", "got": "true"}, func() {
			g := NewDefaultLiterature()
			g.addLog(0, "ask", "literature.log.ask", map[string]string{"to": "1", "suit": "2", "rank": "11", "got": "true"}, nil)
			literatureMigrationLogs = g.GetActionLog()
		}, func() []*ActionLogEntry { return literatureMigrationLogs }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.add()
			entry := requireLogByCode(t, tt.logs(), tt.code)
			require.Equal(t, tt.params, entry.DetailParams)
		})
	}
}

func TestADR0040TargetGamesOperationsCarryCodesAndParams(t *testing.T) {
	t.Run("Binokel bid", func(t *testing.T) {
		g := newTestBinokel()
		g.Reset()
		g.bidPlayerIdx = 0
		require.NoError(t, g.PlayerBid(BinokelMinBid))
		entry := requireLogByCode(t, g.GetActionLog(), "binokel.log.bid")
		require.Equal(t, fmt.Sprintf("%d", BinokelMinBid), entry.DetailParams["amount"])
	})
	t.Run("Cribbage discard", func(t *testing.T) {
		g := newTestCribbage()
		setupDiscardPhase(g)
		require.NoError(t, g.PlayerDiscard([]int{0, 1}))
		entry := requireLogByCode(t, g.GetActionLog(), "cribbage.log.discard")
		require.Empty(t, entry.DetailParams)
	})
	t.Run("Pinochle bid", func(t *testing.T) {
		g := newTestPinochle()
		g.Reset()
		g.bidPlayerIdx = 0
		require.NoError(t, g.PlayerBid(PinochleMinBid))
		entry := requireLogByCode(t, g.GetActionLog(), "pinochle.log.bid")
		require.Equal(t, fmt.Sprintf("%d", PinochleMinBid), entry.DetailParams["amount"])
	})
	t.Run("ChinesePoker bet", func(t *testing.T) {
		g := NewDefaultChinesePoker()
		require.NoError(t, g.Bet(ChinesePokerMinBet))
		entry := requireLogByCode(t, g.GetActionLog(), "chinesepoker.log.bet")
		require.Equal(t, fmt.Sprintf("%d", ChinesePokerMinBet), entry.DetailParams["amount"])
	})
	t.Run("LaughAndLieDown deal", func(t *testing.T) {
		g := NewDefaultLaughAndLieDown()
		g.Reset()
		entry := requireLogByCode(t, g.GetActionLog(), "laughandliedown.log.deal")
		require.Empty(t, entry.DetailParams)
	})
	t.Run("GoFish ask", func(t *testing.T) {
		g := newTestGoFish()
		addGoFishCards(g.players[0], 3)
		addGoFishCards(g.players[1], 3)
		require.NoError(t, g.PlayerAsk(1, 3))
		entry := requireLogByCode(t, g.GetActionLog(), "gofish.log.askHit")
		require.Equal(t, map[string]string{"asker": "0", "target": "1", "rank": "3", "count": "1"}, entry.DetailParams)
	})
}

var (
	guandanMigrationLogs, shengJiMigrationLogs, sixBidSoloMigrationLogs []*ActionLogEntry
	karnoffelMigrationLogs, literatureMigrationLogs                     []*ActionLogEntry
)

func requireLogByCode(t *testing.T, logs []*ActionLogEntry, code string) *ActionLogEntry {
	t.Helper()
	for _, entry := range logs {
		if entry.DetailCode == code {
			return entry
		}
	}
	t.Fatalf("missing detail code %q", code)
	return nil
}
