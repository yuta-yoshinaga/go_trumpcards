//go:build !js || !wasm || extra2

package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PigWebPresenter ピッグWebプレゼンタークラス
type PigWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *PigWebPresenter) Output(s interfaces.PigGame, lastErr error) string {
	resObj := p.buildBase(s)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(s, lastErr)
	// 受動ヒントは Output() でも埋める (#4483)。
	resObj.Hint = pigHintOutput(s)
	return marshalOrError(resObj)
}

// pigHintOutput はヒントを出力形に変換する。
func pigHintOutput(s interfaces.PigGame) *controller.PigWebOutputHint {
	hint := s.GetHint()
	if hint == nil {
		return nil
	}
	return &controller.PigWebOutputHint{CardIndex: hint.CardIndex, Reason: hint.Reason}
}

// buildBase 共通フィールドを構築
func (p *PigWebPresenter) buildBase(s interfaces.PigGame) *controller.PigWebOutput {
	resObj := new(controller.PigWebOutput)
	resObj.Phase = int(s.GetPhase())
	resObj.ValidPlays = intSliceOrEmpty(s.GetValidPassIndices(0))
	resObj.SignallerIdx = s.GetSignallerIdx()
	resObj.NoticedCnt = s.GetNoticedCnt()
	resObj.RoundLoserIdx = s.GetRoundLoserIdx()
	resObj.RoundNumber = s.GetRoundNumber()
	resObj.LetterTarget = domain.PigLetterTargetWord
	resObj.PassCount = s.GetPassCount()
	resObj.DeckSize = s.GetDeckSize()
	resObj.CurrentPlayerIdx = s.GetCurrentPlayerIdx()
	resObj.GameEndFlag = s.GetGameEndFlag()
	resObj.WinnerIdx = s.GetWinnerIdx()
	resObj.Players = p.buildPlayersOutput(s)
	cfg := s.GetConfig()
	resObj.Config = controller.PigWebOutputConfig{
		PlayerCnt:     cfg.PlayerCnt,
		CpuDifficulty: int(cfg.CpuDifficulty),
	}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
func (p *PigWebPresenter) buildPlayersOutput(s interfaces.PigGame) []*controller.PigWebOutputPlayer {
	out := make([]*controller.PigWebOutputPlayer, 0)
	for i := 0; i < s.GetPlayerCnt(); i++ {
		player := s.GetPlayer(i)
		out = append(out, &controller.PigWebOutputPlayer{
			ID:            i,
			IsHuman:       player.GetIsHuman(),
			CardCount:     player.GetCardsSize(),
			Cards:         playerCardsToOutput(player, player.GetIsHuman()),
			Letters:       player.GetLetters(),
			LetterWord:    player.GetLetterWord(),
			Eliminated:    player.GetEliminated(),
			HasSignalled:  player.GetHasSignalled(),
			NoticedOrder:  player.GetNoticedOrder(),
			HasChosenPass: s.HasChosenPass(i),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *PigWebPresenter) buildMessage(s interfaces.PigGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if s.GetGameEndFlag() {
		if s.GetWinnerIdx() == 0 {
			return "", "pig.result.you", nil
		}
		return "", "pig.result.cpu", map[string]string{"idx": strconv.Itoa(s.GetWinnerIdx())}
	}

	switch s.GetPhase() {
	case domain.PigPhaseSignal:
		// **合図が出ています。遅れた 1 人だけが文字をもらう。**
		if s.GetPlayer(0).GetHasSignalled() {
			return "", "pig.signalDone", nil
		}
		return "", "pig.signal", nil
	case domain.PigPhaseRoundEnd:
		loser := s.GetRoundLoserIdx()
		if loser == 0 {
			return "", "pig.round.you", map[string]string{
				"word": s.GetPlayer(0).GetLetterWord(),
			}
		}
		return "", "pig.round.cpu", map[string]string{
			"idx":  strconv.Itoa(loser),
			"word": s.GetPlayer(loser).GetLetterWord(),
		}
	default:
		if s.GetPlayer(0).GetEliminated() {
			return "", "pig.eliminated", nil
		}
		if s.HasChosenPass(0) {
			// **同時に渡すので、全員が選ぶまで待ちます。**
			return "", "pig.waiting", nil
		}
		return "", "pig.pass", nil
	}
}

// HintOutput ヒント情報をJSON出力する
func (p *PigWebPresenter) HintOutput(s interfaces.PigGame) string {
	resObj := p.buildBase(s)
	resObj.Hint = pigHintOutput(s)
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *PigWebPresenter) ActionLogOutput(s interfaces.PigGame) string {
	return actionLogOutputJSON(s)
}
