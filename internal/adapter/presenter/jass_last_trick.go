//go:build !js || !wasm || extra3

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// jassLastTrickPlay は直前トリックで 1 人が出した札。
type jassLastTrickPlay struct {
	// PlayerIdx は出したプレイヤー。
	PlayerIdx int
	// Card は出された札。
	Card *domain.Card
}

// jassLastTrick は直前に解決されたトリック（誰が何を出し、誰が取ったか）を
// アクションログから再構築する。
//
// **ドメインは専用の lastTrick フィールドを持たない。**各トリックの "play" ログ
// （プレイヤーと札）と "trick_win" ログ（勝者）から復元する。Web と CUI の両方が
// この振り返りを出すので、復元を presenter ごとに書き写さない (#5685)。
//
// ビッドフェーズおよびラウンド最初のトリックのプレイ中（この局にまだ確定済み
// トリックが無い）は nil と -1。アクションログはゲーム全体で累積されるため、
// フェーズガードにより前ラウンドのトリックを誤って表示しないようにする。
func jassLastTrick(g interfaces.JassGame) ([]jassLastTrickPlay, int) {
	switch g.GetPhase() {
	case domain.JassPhaseBidTrump, domain.JassPhaseBidPartner:
		return nil, -1
	case domain.JassPhasePlay:
		if g.GetTrickNumber() <= 1 {
			return nil, -1
		}
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
	if len(plays) < domain.JassPlayerCnt {
		return nil, -1
	}
	plays = plays[len(plays)-domain.JassPlayerCnt:]

	out := make([]jassLastTrickPlay, 0, len(plays))
	for _, e := range plays {
		out = append(out, jassLastTrickPlay{PlayerIdx: e.PlayerIdx, Card: e.Cards[0]})
	}
	return out, log[winIdx].PlayerIdx
}
