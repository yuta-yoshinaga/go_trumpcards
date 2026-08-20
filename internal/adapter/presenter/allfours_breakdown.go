package presenter

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// allFoursBreakdown は High / Low / Jack / Game の獲得者と、それが暫定かどうか。
//
// **Web と CUI の両方がこの内訳を出す** (#4771 / #5683)。判定を presenter ごとに
// 書き写すと、同じ局面が画面によって別の結果に見える。
type allFoursBreakdown struct {
	// HighIdx は最高トランプの捕獲者 (-1 = 該当なし)。
	HighIdx int
	// HighCard は最高トランプの札 (nil = 該当なし)。
	HighCard *domain.Card
	// LowIdx は最低トランプの捕獲者 (-1 = 該当なし)。
	LowIdx int
	// LowCard は最低トランプの札 (nil = 該当なし)。
	LowCard *domain.Card
	// JackIdx はトランプ J の捕獲者 (-1 = 場に出ていない)。
	JackIdx int
	// GameIdx はピップ合計が単独最大のプレイヤー (-1 = 同点または全員 0)。
	GameIdx int
	// GamePoints はプレイヤーごとのピップ合計。
	GamePoints []int
	// Provisional は途中経過かどうか。**まだ出ていないトランプで High も Low も
	// 引っくり返る**ので、確定値と区別できるようにする。
	Provisional bool
}

// allFoursComputeBreakdown は現在の盤面から内訳を求める。プレイ中と確定後
// (ROUND_END / GAME_END) 以外では nil。
//
// High/Low = 捕獲済みトランプの最高/最低ランク札の捕獲者。Jack = J トランプの
// 捕獲者 (場に出なければ -1)。Game = ピップ合計が単独最大のプレイヤー
// (同点・全員 0 なら -1)。domain.pegAwards の判定を adapter 層で再現する。
func allFoursComputeBreakdown(s interfaces.AllFoursGame) *allFoursBreakdown {
	phase := s.GetPhase()
	final := phase == domain.AllFoursPhaseRoundEnd || phase == domain.AllFoursPhaseGameEnd
	if !final && phase != domain.AllFoursPhasePlay {
		return nil
	}
	playerCnt := s.GetPlayerCnt()
	bd := &allFoursBreakdown{
		HighIdx: -1, LowIdx: -1, JackIdx: -1, GameIdx: -1,
		GamePoints:  make([]int, playerCnt),
		Provisional: !final,
	}
	trump := s.GetTrumpSuit()
	if trump == domain.AllFoursTrumpUnset {
		return bd
	}

	highRank, lowRank := -1, math.MaxInt32
	for i := 0; i < playerCnt; i++ {
		pl := s.GetPlayer(i)
		if pl == nil {
			continue
		}
		for _, trick := range pl.GetTricksTaken() {
			for _, card := range trick {
				if card == nil {
					continue
				}
				bd.GamePoints[i] += allFoursPipValue(card.GetValue())
				if card.GetDesign() != trump {
					continue
				}
				rank := allFoursRankValue(card.GetValue())
				if rank > highRank {
					highRank, bd.HighCard, bd.HighIdx = rank, card, i
				}
				if rank < lowRank {
					lowRank, bd.LowCard, bd.LowIdx = rank, card, i
				}
				if card.GetValue() == 11 {
					bd.JackIdx = i
				}
			}
		}
	}

	gw, maxTotal, tied := -1, -1, false
	for i, total := range bd.GamePoints {
		switch {
		case total > maxTotal:
			maxTotal, gw, tied = total, i, false
		case total == maxTotal:
			tied = true
		}
	}
	if !tied && gw >= 0 && maxTotal > 0 {
		bd.GameIdx = gw
	}
	return bd
}
