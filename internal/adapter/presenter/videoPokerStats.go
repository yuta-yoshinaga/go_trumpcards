//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

type videoPokerStats struct {
	hands   int
	winRate int
	net     int
}

func videoPokerSessionStats(vp interfaces.VideoPokerGame) videoPokerStats {
	hands := vp.GetHands()
	winRate := 0
	if hands > 0 {
		winRate = (vp.GetWins()*200 + hands) / (hands * 2)
	}
	return videoPokerStats{
		hands:   hands,
		winRate: winRate,
		net:     vp.GetTotalPayout() - vp.GetTotalBet(),
	}
}
