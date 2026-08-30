package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// pitchLastTrick は直前に解決されたトリック（誰が何を出し誰が取ったか）を
// アクションログから再構築する。
//
// ドメインは専用の lastTrick フィールドを持たないが、各トリックの "play" ログ
// （プレイヤーと札）と "trick_win" ログ（勝者）から復元する。Web と CUI の両方が
// この振り返りを出すため、復元ロジックを共通化している (#6380)。
//
// ラウンド最初のトリックのプレイ中（この局にまだ確定済みトリックが無い）や
// 確定済みトリックのログが不足している場合は nil と -1 を返す。
func pitchLastTrick(s interfaces.PitchGame) ([]*domain.ActionLogEntry, int) {
	// ラウンド最初のトリックがプレイ中は、この局に確定済みトリックが無いため空を返す。
	if s.GetPhase() == domain.PitchPhasePlay && s.GetTrickNumber() <= 1 {
		return nil, -1
	}

	log := s.GetActionLog()
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
	if len(plays) < domain.PitchPlayerCnt {
		return nil, -1
	}
	plays = plays[len(plays)-domain.PitchPlayerCnt:]
	return plays, log[winIdx].PlayerIdx
}
