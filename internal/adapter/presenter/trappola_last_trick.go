//go:build !js || !wasm || extra2

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// trappolaLastTrickPlay は直前トリックで 1 人が出した札。
type trappolaLastTrickPlay struct {
	// PlayerIdx は出したプレイヤー。
	PlayerIdx int
	// Card は出された札。
	Card *domain.Card
}

// trappolaLastTrick は直前に解決されたトリック（誰が何を出し誰が取ったか）を
// アクションログから再構築する。ドメインは専用の lastTrick フィールドを持たないが、
// 各トリックの "play" ログ（プレイヤーと札）と "trick_win" ログ（勝者）から
// 現ラウンドの直近トリックを復元できる。ラウンド開始直後（プレイフェーズのトリック 1
// で、この局のトリックがまだ確定していない）は nil と -1 を返す。
func trappolaLastTrick(g interfaces.TrappolaGame) ([]trappolaLastTrickPlay, int) {
	// ラウンド最初のトリックがプレイ中は、この局に確定済みトリックが無いため空を返す。
	if g.GetPhase() == domain.TrappolaPhasePlay && g.GetTrickNumber() <= 1 {
		return nil, -1
	}

	log := g.GetActionLog()
	winIdx := -1
	for i := len(log) - 1; i >= 0; i-- {
		if log[i] != nil && log[i].ActionType == "trick_win" {
			winIdx = i
			break
		}
	}
	if winIdx < 0 {
		return nil, -1
	}

	// trick_win 直前の "play" ログ（プレイ順）が、そのトリックの各札に対応する。
	var plays []*domain.ActionLogEntry
	for i := 0; i < winIdx; i++ {
		if e := log[i]; e != nil && e.ActionType == "play" && len(e.Cards) > 0 {
			plays = append(plays, e)
		}
	}
	if len(plays) < domain.TrappolaPlayerCnt {
		return nil, -1
	}
	plays = plays[len(plays)-domain.TrappolaPlayerCnt:]

	out := make([]trappolaLastTrickPlay, 0, len(plays))
	for _, e := range plays {
		out = append(out, trappolaLastTrickPlay{
			PlayerIdx: e.PlayerIdx,
			Card:      e.Cards[0],
		})
	}
	return out, log[winIdx].PlayerIdx
}
