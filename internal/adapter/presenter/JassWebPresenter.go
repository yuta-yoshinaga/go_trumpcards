//go:build !js || !wasm || extra3

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// JassWebPresenter ヤス(シーバー)Webプレゼンタークラス
type JassWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *JassWebPresenter) Output(g interfaces.JassGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, g.GetCurrentTrick(), lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**Jass.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.JassWebOutputHint{
			CardIndex: hint.CardIndex,
			Schieben:  hint.Schieben,
			Suit:      hint.Suit,
			Reason:    hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

func (p *JassWebPresenter) buildBase(g interfaces.JassGame) *controller.JassWebOutput {
	resObj := new(controller.JassWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.BidPlayerIdx = g.GetBidPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.ForehandIdx = g.GetForehandIdx()
	resObj.TrumpSuit = g.GetTrumpSuit()
	resObj.Schieben = g.GetSchieben()
	resObj.MakerTeam = g.GetMakerTeam()
	resObj.MakerPlayerIdx = g.GetMakerPlayerIdx()
	resObj.TeamScores = [2]int{g.GetTeamScore(0), g.GetTeamScore(1)}
	resObj.RoundPoints = [2]int{g.GetRoundPoints(0), g.GetRoundPoints(1)}
	resObj.RoundWeisPoints = [2]int{g.GetRoundWeisPoints(0), g.GetRoundWeisPoints(1)}
	resObj.RoundStockPoints = [2]int{g.GetRoundStockPoints(0), g.GetRoundStockPoints(1)}
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerTeam = g.GetWinnerTeam()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()

	cfg := g.GetConfig()
	resObj.Config = controller.JassWebOutputConfig{
		CpuDifficulty:  int(cfg.CpuDifficulty),
		TargetScore:    cfg.TargetScore,
		LastTrickBonus: cfg.LastTrickBonus,
		EnableWeis:     cfg.EnableWeis,
	}

	resObj.CurrentTrick = trickCardsToOutput(g.GetCurrentTrick())
	resObj.LastTrick, resObj.LastTrickWinner = p.buildLastTrickOutput(g)
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// buildLastTrickOutput は直前トリックを Web 出力の形に変換する。復元そのものは
// jassLastTrick に持たせている（CUI も同じ振り返りを出すため）。
func (p *JassWebPresenter) buildLastTrickOutput(g interfaces.JassGame) ([]*controller.WebOutputTrickCard, int) {
	plays, winner := jassLastTrick(g)
	out := make([]*controller.WebOutputTrickCard, 0, len(plays))
	for _, pl := range plays {
		out = append(out, &controller.WebOutputTrickCard{PlayerIdx: pl.PlayerIdx, Card: cardToOutput(pl.Card)})
	}
	return out, winner
}

func (p *JassWebPresenter) buildPlayersOutput(g interfaces.JassGame) []*controller.JassWebOutputPlayer {
	out := make([]*controller.JassWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		pObj := &controller.JassWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			Team:       player.GetTeam(),
			TrickCount: player.GetTrickCount(),
		}
		out = append(out, pObj)
	}
	return out
}

func (p *JassWebPresenter) buildMessage(g interfaces.JassGame, trick []*domain.TrickCard, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerTeam := g.GetWinnerTeam()
		msg := fmt.Sprintf("ゲーム終了！ チーム%dの勝ち！", winnerTeam)
		code := fmt.Sprintf("jass.result.team%dWin", winnerTeam)
		params := map[string]string{"team": fmt.Sprintf("%d", winnerTeam)}
		return msg, code, params
	}
	switch g.GetPhase() {
	case domain.JassPhaseBidTrump:
		return "", "jass.bidTrumpPhase", nil
	case domain.JassPhaseBidPartner:
		return "", "jass.bidPartnerPhase", nil
	case domain.JassPhasePlay:
		if len(trick) == 0 {
			return "", "jass.playPhase.lead", nil
		}
		return "", "jass.playPhase.follow", nil
	case domain.JassPhaseTrickEnd:
		return "", "jass.trickEnd", nil
	case domain.JassPhaseRoundEnd:
		return "", "jass.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *JassWebPresenter) HintOutput(g interfaces.JassGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.JassWebOutputHint{
			CardIndex: hint.CardIndex,
			Schieben:  hint.Schieben,
			Suit:      hint.Suit,
			Reason:    hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *JassWebPresenter) ActionLogOutput(g interfaces.JassGame) string {
	return actionLogOutputJSON(g)
}
