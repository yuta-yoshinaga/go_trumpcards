//go:build !js || !wasm || extra

package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// LingerLongerWebPresenter リンガーロンガーWebプレゼンタークラス
type LingerLongerWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *LingerLongerWebPresenter) Output(s interfaces.LingerLongerGame, lastErr error) string {
	resObj := p.buildBase(s)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(s, lastErr)
	// 受動ヒントは Output() でも埋める (#4483)。
	resObj.Hint = lingerLongerHintOutput(s)
	return marshalOrError(resObj)
}

// lingerLongerHintOutput はヒントを出力形に変換する。
func lingerLongerHintOutput(s interfaces.LingerLongerGame) *controller.LingerLongerWebOutputHint {
	hint := s.GetHint()
	if hint == nil {
		return nil
	}
	return &controller.LingerLongerWebOutputHint{CardIndex: hint.CardIndex, Reason: hint.Reason}
}

// buildBase 共通フィールドを構築
func (p *LingerLongerWebPresenter) buildBase(s interfaces.LingerLongerGame) *controller.LingerLongerWebOutput {
	resObj := new(controller.LingerLongerWebOutput)
	resObj.Phase = int(s.GetPhase())
	resObj.ValidPlays = intSliceOrEmpty(s.GetValidPlayIndices(0))
	resObj.StockSize = s.GetStockSize()
	resObj.CurrentTrick = trickCardsToOutput(s.GetCurrentTrick())
	resObj.CurrentPlayerIdx = s.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = s.GetLeadPlayerIdx()
	resObj.TrickNumber = s.GetTrickNumber()
	resObj.LastDrawIdx = s.GetLastDrawIdx()
	resObj.EliminatedCnt = s.GetEliminatedCnt()
	resObj.Discarded = s.GetDiscarded()
	resObj.GameEndFlag = s.GetGameEndFlag()
	resObj.WinnerIdx = s.GetWinnerIdx()
	resObj.Players = p.buildPlayersOutput(s)
	resObj.Config = controller.LingerLongerWebOutputConfig{PlayerCnt: s.GetConfig().PlayerCnt}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *LingerLongerWebPresenter) buildPlayersOutput(s interfaces.LingerLongerGame) []*controller.LingerLongerWebOutputPlayer {
	out := make([]*controller.LingerLongerWebOutputPlayer, 0)
	for i := 0; i < s.GetPlayerCnt(); i++ {
		player := s.GetPlayer(i)
		out = append(out, &controller.LingerLongerWebOutputPlayer{
			ID:           i,
			IsHuman:      player.GetIsHuman(),
			CardCount:    player.GetCardsSize(),
			Cards:        playerCardsToOutput(player, player.GetIsHuman()),
			TricksWon:    player.GetTricksWon(),
			EliminatedAt: player.GetEliminatedAt(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *LingerLongerWebPresenter) buildMessage(s interfaces.LingerLongerGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if s.GetGameEndFlag() {
		if s.GetWinnerIdx() == 0 {
			return "", "lingerlonger.result.you", nil
		}
		return "", "lingerlonger.result.cpu", map[string]string{
			"idx": strconv.Itoa(s.GetWinnerIdx()),
		}
	}
	// **人間が脱落しても盤面は続く。** そうと言わないと、なぜ打てないのか分からない。
	if human := s.GetPlayer(0); human != nil && human.IsEliminated() {
		return "", "lingerlonger.eliminated", nil
	}
	// **山札が尽きると誰も補充できない。** そこから脱落が一気に進みます。
	if s.GetStockSize() == 0 {
		return "", "lingerlonger.noStock", map[string]string{
			"n": strconv.Itoa(s.GetPlayer(0).GetCardsSize()),
		}
	}
	return "", "lingerlonger.play", map[string]string{
		"stock": strconv.Itoa(s.GetStockSize()),
		"n":     strconv.Itoa(s.GetPlayer(0).GetCardsSize()),
	}
}

// HintOutput ヒント情報をJSON出力する
func (p *LingerLongerWebPresenter) HintOutput(s interfaces.LingerLongerGame) string {
	resObj := p.buildBase(s)
	resObj.Hint = lingerLongerHintOutput(s)
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *LingerLongerWebPresenter) ActionLogOutput(s interfaces.LingerLongerGame) string {
	return actionLogOutputJSON(s)
}
