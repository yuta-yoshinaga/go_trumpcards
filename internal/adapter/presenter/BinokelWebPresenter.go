package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BinokelWebPresenter ビノクルWebプレゼンタークラス
type BinokelWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *BinokelWebPresenter) Output(g interfaces.BinokelGame, lastErr error) string {
	resObj := p.buildBase(g, lastErr)

	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.BinokelWebOutputHint{
			CardIndex: hint.CardIndex,
			BidAmount: hint.BidAmount,
			Pass:      hint.Pass,
			Suit:      hint.Suit,
			Reason:    hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// buildBase は状態レスポンスを組み立てる。Output と HintOutput で共有する。
func (p *BinokelWebPresenter) buildBase(g interfaces.BinokelGame, lastErr error) *controller.BinokelWebOutput {
	resObj := new(controller.BinokelWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.BidPlayerIdx = g.GetBidPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.TrumpSuit = g.GetTrumpSuit()
	resObj.HighestBid = g.GetHighestBid()
	resObj.HighestBidder = g.GetHighestBidder()
	resObj.Scores = g.GetScores()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerPlayer = g.GetWinnerPlayer()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()

	table := domain.BinokelMeldTable()
	resObj.MeldTable = make([]*controller.BinokelWebOutputMeldTableEntry, 0, len(table))
	for _, e := range table {
		resObj.MeldTable = append(resObj.MeldTable, &controller.BinokelWebOutputMeldTableEntry{
			Type:   int(e.Type),
			Points: e.Points,
		})
	}

	cfg := g.GetConfig()
	resObj.Config = controller.BinokelWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
	}

	trick := g.GetCurrentTrick()
	resObj.CurrentTrick = trickCardsToOutput(trick)
	resObj.Dabb = cardsToOutput(g.GetDabb())
	resObj.DabbDiscarded = cardsToOutput(g.GetDabbDiscarded())
	resObj.Players = p.buildPlayersOutput(g)
	resObj.PlayerMelds = p.buildMeldsOutput(g)

	if g.GetPhase() == domain.BinokelPhasePlay && g.IsHumanTurn() {
		resObj.ValidPlayIndices = g.GetValidPlayIndices(g.GetCurrentPlayerIdx())
	}

	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return resObj
}

// HintOutput ヒント情報を出力
func (p *BinokelWebPresenter) HintOutput(g interfaces.BinokelGame) string {
	resObj := p.buildBase(g, nil)
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.BinokelWebOutputHint{
			CardIndex: hint.CardIndex,
			BidAmount: hint.BidAmount,
			Pass:      hint.Pass,
			Suit:      hint.Suit,
			Reason:    hint.Reason,
		}
		resObj.MessageCode = "binokel.hintRequested"
	} else {
		resObj.MessageCode = "binokel.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜を出力
func (p *BinokelWebPresenter) ActionLogOutput(g interfaces.BinokelGame) string {
	return actionLogToJSON(g.GetActionLog())
}

// buildPlayersOutput プレイヤー情報を構築
func (p *BinokelWebPresenter) buildPlayersOutput(g interfaces.BinokelGame) []*controller.BinokelWebOutputPlayer {
	out := make([]*controller.BinokelWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		showCards := player.GetIsHuman() || g.GetPhase() == domain.BinokelPhaseGameEnd
		pObj := &controller.BinokelWebOutputPlayer{
			ID:          i,
			IsHuman:     player.GetIsHuman(),
			CardCount:   player.GetCardsSize(),
			Cards:       playerCardsToOutput(player, showCards),
			Score:       g.GetScore(i),
			TrickCount:  player.GetTrickCount(),
			Bid:         player.GetBid(),
			HasPassed:   player.GetHasPassed(),
			MeldScore:   player.GetMeldScore(),
			TrickPoints: player.GetTrickPoints(),
		}
		out = append(out, pObj)
	}
	return out
}

// buildMeldsOutput メルド情報を構築
func (p *BinokelWebPresenter) buildMeldsOutput(g interfaces.BinokelGame) [domain.BinokelPlayerCnt][]*controller.BinokelWebOutputMeld {
	var out [domain.BinokelPlayerCnt][]*controller.BinokelWebOutputMeld
	melds := g.GetPlayerMelds()
	for i := range domain.BinokelPlayerCnt {
		playerMelds := make([]*controller.BinokelWebOutputMeld, 0)
		for _, m := range melds[i] {
			playerMelds = append(playerMelds, &controller.BinokelWebOutputMeld{
				Type:   int(m.Type),
				Points: m.Points,
				Cards:  cardsToOutput(m.Cards),
			})
		}
		out[i] = playerMelds
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *BinokelWebPresenter) buildMessage(g interfaces.BinokelGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winner := g.GetWinnerPlayer()
		msg := fmt.Sprintf("ゲーム終了！ プレイヤー%dの勝ち！", winner)
		code := fmt.Sprintf("binokel.result.player%dWin", winner)
		params := map[string]string{"player": fmt.Sprintf("%d", winner)}
		return msg, code, params
	}
	switch g.GetPhase() {
	case domain.BinokelPhaseBid:
		return "", "binokel.bidPhase", nil
	case domain.BinokelPhaseDabb:
		return "", "binokel.dabbPhase", nil
	case domain.BinokelPhaseTrump:
		return "", "binokel.trumpPhase", nil
	case domain.BinokelPhaseMeld:
		return "", "binokel.meldPhase", nil
	case domain.BinokelPhasePlay:
		return "", "binokel.playPhase", nil
	case domain.BinokelPhaseTrickEnd:
		return "", "binokel.trickEndPhase", nil
	case domain.BinokelPhaseRoundEnd:
		return "", "binokel.roundEndPhase", nil
	default:
		return "", "", nil
	}
}
