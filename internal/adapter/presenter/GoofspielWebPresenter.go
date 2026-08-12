//go:build !js || !wasm || extra

package presenter

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// GoofspielWebPresenter ゴフスピールWebプレゼンタークラス
type GoofspielWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *GoofspielWebPresenter) Output(s interfaces.GoofspielGame, lastErr error) string {
	resObj := p.buildBase(s)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(s, lastErr)
	// 受動ヒントは Output() でも埋める (#4483)。
	resObj.Hint = goofspielHintOutput(s)
	return marshalOrError(resObj)
}

// goofspielHintOutput はヒントを出力形に変換する。
func goofspielHintOutput(s interfaces.GoofspielGame) *controller.GoofspielWebOutputHint {
	hint := s.GetHint()
	if hint == nil {
		return nil
	}
	return &controller.GoofspielWebOutputHint{CardIndex: hint.CardIndex, Reason: hint.Reason}
}

// buildBase 共通フィールドを構築
func (p *GoofspielWebPresenter) buildBase(s interfaces.GoofspielGame) *controller.GoofspielWebOutput {
	resObj := new(controller.GoofspielWebOutput)
	resObj.Phase = int(s.GetPhase())
	resObj.ValidPlays = intSliceOrEmpty(s.GetValidBidIndices(0))
	if prize := s.GetCurrentPrize(); prize != nil {
		resObj.CurrentPrize = cardToOutput(prize)
	}
	resObj.CarriedPrizes = cardsToOutput(s.GetCarriedPrizes())
	resObj.PrizeValue = s.PrizeValue()
	resObj.PrizeRemaining = s.GetPrizeRemaining()
	resObj.LastWinnerIdx = s.GetLastWinnerIdx()
	resObj.LastGained = s.GetLastGained()
	resObj.RoundNumber = s.GetRoundNumber()
	resObj.GameEndFlag = s.GetGameEndFlag()
	resObj.WinnerIdx = s.GetWinnerIdx()
	resObj.Players = p.buildPlayersOutput(s)
	cfg := s.GetConfig()
	resObj.Config = controller.GoofspielWebOutputConfig{
		PlayerCnt: cfg.PlayerCnt,
		TieRule:   int(cfg.TieRule),
	}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築
//
// **CPU の残り手札も公開します。** 使った札は場に出るので誰にでも数えられます——
// 隠しても手間が増えるだけで、隠せていません。伏せるのは今この瞬間の入札だけです。
func (p *GoofspielWebPresenter) buildPlayersOutput(s interfaces.GoofspielGame) []*controller.GoofspielWebOutputPlayer {
	revealed := s.GetRevealedBids()
	out := make([]*controller.GoofspielWebOutputPlayer, 0)
	for i := 0; i < s.GetPlayerCnt(); i++ {
		player := s.GetPlayer(i)
		entry := &controller.GoofspielWebOutputPlayer{
			ID:        i,
			IsHuman:   player.GetIsHuman(),
			CardCount: player.GetCardsSize(),
			Cards:     playerCardsToOutput(player, true),
			Score:     player.GetScore(),
			HasBid:    s.HasBid(i),
		}
		// **入札は公開された後にだけ載せます。**
		if s.GetPhase() == domain.GoofspielPhaseReveal && i < len(revealed) && revealed[i] != nil {
			entry.RevealedBid = cardToOutput(revealed[i])
		}
		out = append(out, entry)
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *GoofspielWebPresenter) buildMessage(s interfaces.GoofspielGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if s.GetGameEndFlag() {
		params := map[string]string{"n": strconv.Itoa(s.GetPlayer(s.GetWinnerIdx()).GetScore())}
		if s.GetWinnerIdx() == 0 {
			return "", "goofspiel.result.you", params
		}
		params["idx"] = strconv.Itoa(s.GetWinnerIdx())
		return "", "goofspiel.result.cpu", params
	}

	if s.GetPhase() == domain.GoofspielPhaseReveal {
		// **同点は誰も取りません。** 勝者が居ない結果を言い分けます。
		if s.GetLastWinnerIdx() < 0 {
			return "", "goofspiel.round.tie", nil
		}
		params := map[string]string{"n": strconv.Itoa(s.GetLastGained())}
		if s.GetLastWinnerIdx() == 0 {
			return "", "goofspiel.round.you", params
		}
		params["idx"] = strconv.Itoa(s.GetLastWinnerIdx())
		return "", "goofspiel.round.cpu", params
	}

	if s.HasBid(0) {
		return "", "goofspiel.waiting", nil
	}
	return "", "goofspiel.bid", map[string]string{"n": strconv.Itoa(s.PrizeValue())}
}

// HintOutput ヒント情報をJSON出力する
func (p *GoofspielWebPresenter) HintOutput(s interfaces.GoofspielGame) string {
	resObj := p.buildBase(s)
	resObj.Hint = goofspielHintOutput(s)
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *GoofspielWebPresenter) ActionLogOutput(s interfaces.GoofspielGame) string {
	return actionLogOutputJSON(s)
}
