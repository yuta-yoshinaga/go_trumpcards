//go:build !js || !wasm || solo

package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SnapWebPresenter スナップWebプレゼンタークラス
type SnapWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *SnapWebPresenter) Output(s interfaces.SnapGame, lastErr error) string {
	resObj := p.buildBase(s)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(s, lastErr)
	// 受動ヒントは Output() でも埋める (#4483)。
	resObj.Hint = snapHintOutput(s)
	return marshalOrError(resObj)
}

// snapHintOutput はヒントを出力形に変換する。
func snapHintOutput(s interfaces.SnapGame) *controller.SnapWebOutputHint {
	hint := s.GetHint()
	if hint == nil {
		return nil
	}
	return &controller.SnapWebOutputHint{Snap: hint.Snap, Reason: hint.Reason}
}

// buildBase 共通フィールドを構築
func (p *SnapWebPresenter) buildBase(s interfaces.SnapGame) *controller.SnapWebOutput {
	cfg := s.GetConfig()
	pending := s.GetPending()
	last := s.GetLastEvent()

	resObj := new(controller.SnapWebOutput)
	resObj.Phase = int(s.GetPhase())
	resObj.GameEndFlag = s.GetGameEndFlag()
	resObj.WinnerIdx = s.GetWinnerIdx()
	resObj.CurrentTurnIdx = s.GetCurrentTurnIdx()
	resObj.IsHumanTurn = !s.GetGameEndFlag() && s.GetCurrentTurnIdx() == 0
	// **上 2 枚が同ランクのときだけ真。** ページはこれを見るだけでよい。
	resObj.SnapAvailable = s.IsSnapAvailable()
	resObj.CenterPileSize = s.GetCenterPileSize()
	if top := s.GetTopCard(); top != nil {
		resObj.TopCard = cardToOutput(top)
	}
	resObj.Players = p.buildPlayersOutput(s)
	resObj.PlayerCnt = cfg.PlayerCnt
	resObj.CpuDifficulty = int(cfg.CpuDifficulty)
	resObj.PendingKind = int(pending.Kind)
	resObj.PendingDeadlineMs = pending.DeadlineMs
	resObj.LastEventKind = int(last.Kind)
	resObj.LastEventPlayerIdx = last.PlayerIdx
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
//
// **手札は出さない。** ストックは裏向きで、枚数だけが公開情報です。
func (p *SnapWebPresenter) buildPlayersOutput(s interfaces.SnapGame) []*controller.SnapWebPlayer {
	out := make([]*controller.SnapWebPlayer, 0)
	for i := 0; i < s.GetPlayerCnt(); i++ {
		player := s.GetPlayer(i)
		out = append(out, &controller.SnapWebPlayer{
			ID:        i,
			IsHuman:   player.GetIsHuman(),
			StockSize: player.GetStockSize(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *SnapWebPresenter) buildMessage(s interfaces.SnapGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if s.GetGameEndFlag() {
		switch {
		case s.GetWinnerIdx() == 0:
			return "", "snap.result.you", nil
		case s.GetWinnerIdx() > 0:
			return "", "snap.result.cpu", map[string]string{"idx": strconv.Itoa(s.GetWinnerIdx())}
		default:
			return "", "snap.result.none", nil
		}
	}
	if s.IsSnapAvailable() {
		return "", "snap.available", nil
	}
	return "", "snap.play", map[string]string{"n": strconv.Itoa(s.GetCenterPileSize())}
}

// HintOutput ヒント情報をJSON出力する
func (p *SnapWebPresenter) HintOutput(s interfaces.SnapGame) string {
	resObj := p.buildBase(s)
	resObj.Hint = snapHintOutput(s)
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *SnapWebPresenter) ActionLogOutput(s interfaces.SnapGame) string {
	return actionLogOutputJSON(s)
}
