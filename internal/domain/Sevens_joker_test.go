package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSevensGetJokerPlacements(t *testing.T) {
	game := NewSevens(NewTrumpCards(0), []*SevensPlayer{NewSevensPlayer(true)}, DefaultSevensConfig())
	assert.Equal(t, []SevensPosition{
		{Suit: 1, Value: 6}, {Suit: 1, Value: 8}, {Suit: 2, Value: 6}, {Suit: 2, Value: 8},
		{Suit: 3, Value: 6}, {Suit: 3, Value: 8}, {Suit: 4, Value: 6}, {Suit: 4, Value: 8},
	}, game.GetJokerPlacements())
	assert.NotNil(t, game.GetJokerPlacements())
}

func TestSevensGetJokerPlacementsBoardRules(t *testing.T) {
	cfg := DefaultSevensConfig()
	cfg.TunnelEnabled = true
	cfg.TunnelSkipWidth = 2
	cfg.EndStopEnabled = true
	game := NewSevens(NewTrumpCards(0), []*SevensPlayer{NewSevensPlayer(true)}, cfg)
	game.placePosition(1, 13)
	game.placePosition(1, 1)
	game.placePosition(2, 1)
	game.placePosition(2, 13)
	game.placePosition(3, 13)
	placements := game.GetJokerPlacements()
	assert.Contains(t, placements, SevensPosition{Suit: 3, Value: 11})   // tunnel skip reaches K
	assert.NotContains(t, placements, SevensPosition{Suit: 3, Value: 1}) // K end-stop blocks A
	assert.NotContains(t, placements, SevensPosition{Suit: 1, Value: 8}) // A end-stop blocks upper side
	assert.NotContains(t, placements, SevensPosition{Suit: 2, Value: 6}) // K end-stop blocks lower side
	assert.NotContains(t, placements, SevensPosition{Suit: 1, Value: 1}) // placed positions excluded
	assert.NotContains(t, placements, SevensPosition{Suit: 1, Value: 13})
	assert.NotContains(t, placements, SevensPosition{Suit: 1, Value: 7})
}

func TestSevensGetJokerPlacementsCustomSkipWidthThree(t *testing.T) {
	for _, tc := range []struct {
		name   string
		width  int
		value  int
		placed []int
		want   bool
	}{
		{name: "skip low", width: 3, value: 4, placed: []int{7}, want: true},
		{name: "skip high", width: 3, value: 10, placed: []int{7}, want: true},
		{name: "not a skip", width: 3, value: 5, placed: []int{7}},
		{name: "normal adjacency", width: 3, value: 6, placed: []int{7}, want: true},
		{name: "width one disabled", width: 1, value: 5, placed: []int{7}},
		{name: "width zero disabled", width: 0, value: 4, placed: []int{7}},
		{name: "skip does not go out of range without tunnel", width: 3, value: 11, placed: []int{1, 7}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := DefaultSevensConfig()
			cfg.TunnelSkipWidth = tc.width
			game := NewSevens(NewTrumpCards(0), []*SevensPlayer{NewSevensPlayer(true)}, cfg)
			game.tablePlaced = [5]uint16{}
			for _, value := range tc.placed {
				game.placePosition(CardDesignSpade, value)
			}
			assert.Equal(t, tc.want, containsSevensPosition(game.GetJokerPlacements(), CardDesignSpade, tc.value))
		})
	}
}

func TestSevensGetJokerPlacementsCustomSkipWrap(t *testing.T) {
	for _, tc := range []struct {
		name, placedValue string
		value             int
	}{
		{name: "low direction wraps", placedValue: "1", value: 11},
		{name: "high direction wraps", placedValue: "13", value: 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := DefaultSevensConfig()
			cfg.TunnelEnabled = true
			cfg.TunnelSkipWidth = 3
			game := NewSevens(NewTrumpCards(0), []*SevensPlayer{NewSevensPlayer(true)}, cfg)
			game.tablePlaced = [5]uint16{}
			game.placePosition(CardDesignSpade, 7)
			if tc.placedValue == "1" {
				game.placePosition(CardDesignSpade, 1)
			} else {
				game.placePosition(CardDesignSpade, 13)
			}
			assert.Contains(t, game.GetJokerPlacements(), SevensPosition{Suit: CardDesignSpade, Value: tc.value})
		})
	}
}

func TestSevensGetJokerPlacementsTunnelEndStop(t *testing.T) {
	for _, tc := range []struct {
		name       string
		placed     int
		target     int
		wantTunnel bool
	}{
		{name: "K opens A", placed: 13, target: 1, wantTunnel: true},
		{name: "A opens K", placed: 1, target: 13, wantTunnel: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, tunnel := range []bool{false, true} {
				cfg := DefaultSevensConfig()
				cfg.TunnelEnabled = tunnel
				cfg.EndStopEnabled = false
				game := NewSevens(NewTrumpCards(0), []*SevensPlayer{NewSevensPlayer(true)}, cfg)
				game.tablePlaced = [5]uint16{}
				game.placePosition(CardDesignSpade, tc.placed)
				got := containsSevensPosition(game.GetJokerPlacements(), CardDesignSpade, tc.target)
				assert.Equal(t, tunnel && tc.wantTunnel, got, "tunnel enabled: %v", tunnel)
			}
		})
	}
}

func TestSevensGetJokerPlacementsSingleOpenedSuitAndNoAnchors(t *testing.T) {
	game := NewSevens(NewTrumpCards(0), []*SevensPlayer{NewSevensPlayer(true)}, DefaultSevensConfig())
	game.tablePlaced = [5]uint16{}
	game.placePosition(CardDesignSpade, 7)
	assert.Equal(t, []SevensPosition{
		{Suit: CardDesignSpade, Value: 6}, {Suit: CardDesignSpade, Value: 8},
	}, game.GetJokerPlacements())

	game.tablePlaced = [5]uint16{}
	assert.NotNil(t, game.GetJokerPlacements())
	assert.Empty(t, game.GetJokerPlacements())
}

func containsSevensPosition(positions []SevensPosition, suit, value int) bool {
	for _, position := range positions {
		if position.Suit == suit && position.Value == value {
			return true
		}
	}
	return false
}

func TestSevensGetJokerPlacementsCustomSkipWithoutTunnel(t *testing.T) {
	cfg := DefaultSevensConfig()
	cfg.TunnelSkipWidth = 2
	game := NewSevens(NewTrumpCards(0), []*SevensPlayer{NewSevensPlayer(true)}, cfg)
	placements := game.GetJokerPlacements()
	assert.Contains(t, placements, SevensPosition{Suit: 1, Value: 5})
	assert.Contains(t, placements, SevensPosition{Suit: 1, Value: 9})
	assert.NotContains(t, placements, SevensPosition{Suit: 1, Value: 1})
}

func TestSevensGetJokerPlacementsEmpty(t *testing.T) {
	game := NewSevens(NewTrumpCards(0), []*SevensPlayer{NewSevensPlayer(true)}, DefaultSevensConfig())
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		for value := 1; value <= 13; value++ {
			game.placePosition(suit, value)
		}
	}
	placements := game.GetJokerPlacements()
	assert.NotNil(t, placements)
	assert.Empty(t, placements)
}
