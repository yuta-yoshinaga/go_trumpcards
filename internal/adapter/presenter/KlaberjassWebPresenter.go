//go:build !js || !wasm || extra3

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// KlaberjassWebPresenter クラバーヤス Webプレゼンタークラス
type KlaberjassWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *KlaberjassWebPresenter) Output(g interfaces.KlaberjassGame, lastErr error) string {
	resObj := new(controller.KlaberjassWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.DealNumber = g.GetDealNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.BidPlayerIdx = g.GetBidPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.TrumpSuit = g.GetTrumpSuit()
	resObj.TurnUpCard = cardToOutput(g.GetTurnUpCard())
	resObj.MakerIdx = g.GetMakerIdx()
	resObj.TrickLeaderIdx = g.GetTrickLeaderIdx()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.SequenceWinner = g.GetSequenceWinner()
	resObj.LastTrickWinner = g.GetLastTrickWinner()
	resObj.LastTrickBonus = domain.KlaberjassLastTrickBonus
	resObj.BelaHolder = g.GetBelaHolder()
	resObj.BelaScored = g.IsBelaScored()
	resObj.DixUsed = g.IsDixUsed()
	resObj.Bete = g.IsBete()
	resObj.SchmeissBy = g.GetSchmeissBy()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerIdx = g.GetWinnerIdx()

	trick := g.GetTrick()
	resObj.Trick = make([]*controller.WebOutputCard, 0, len(trick))
	for _, c := range trick {
		if out := cardToOutput(c); out != nil {
			resObj.Trick = append(resObj.Trick, out)
		}
	}

	cfg := g.GetConfig()
	resObj.TargetScore = cfg.TargetScore
	resObj.Config = controller.KlaberjassWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetScore:   cfg.TargetScore,
		AllowSchmeiss: cfg.AllowSchmeiss,
	}

	// **出せる札はサーバーが決める。**追随・切札・上乗せがすべて強制なので、
	// フロントで再現すると必ずずれる。
	resObj.ValidPlays = make([]int, 0)
	if g.GetPhase() == domain.KlaberjassPhasePlay && g.IsHumanTurn() {
		resObj.ValidPlays = append(resObj.ValidPlays, g.KlaberjassValidPlays(g.GetCurrentPlayerIdx())...)
	}

	resObj.Players = p.buildPlayersOutput(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return marshalOrError(resObj)
}

// klaberjassWebReveal は手札と役を公開する局面かを返す。
func klaberjassWebReveal(g interfaces.KlaberjassGame) bool {
	phase := g.GetPhase()
	return phase == domain.KlaberjassPhaseHandEnd || phase == domain.KlaberjassPhaseGameEnd
}

// buildPlayersOutput プレイヤー情報を構築
func (p *KlaberjassWebPresenter) buildPlayersOutput(g interfaces.KlaberjassGame) []*controller.KlaberjassWebOutputPlayer {
	players := g.GetPlayers()
	out := make([]*controller.KlaberjassWebOutputPlayer, 0, len(players))
	reveal := klaberjassWebReveal(g)
	for i := range players {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		cards := make([]*controller.WebOutputCard, 0, player.GetCardsSize())
		// **相手の手札は伏せる。**枚数だけ送る。
		if player.GetIsHuman() || reveal {
			for j := range player.GetCardsSize() {
				if c := cardToOutput(player.GetCard(j)); c != nil {
					cards = append(cards, c)
				}
			}
		}
		// **役は精算まで伏せる。**申告し合う前に相手の役が見えると勝負にならない。
		seqs := make([]*controller.KlaberjassWebOutputSequence, 0)
		if reveal {
			for _, s := range g.GetSequences(i) {
				if s == nil {
					continue
				}
				seqs = append(seqs, &controller.KlaberjassWebOutputSequence{
					Suit: s.Suit, TopValue: s.TopValue, Length: s.Length, Points: s.Points,
				})
			}
		}
		out = append(out, &controller.KlaberjassWebOutputPlayer{
			ID:            i,
			IsHuman:       player.GetIsHuman(),
			CardCount:     player.GetCardsSize(),
			Cards:         cards,
			Sequences:     seqs,
			HandPoints:    g.GetHandPoints(i),
			Score:         g.GetScore(i),
			IsMaker:       i == g.GetMakerIdx(),
			IsDealer:      i == g.GetDealerIdx(),
			IsCurrentTurn: g.GetPhase() == domain.KlaberjassPhasePlay && i == g.GetCurrentPlayerIdx(),
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *KlaberjassWebPresenter) buildMessage(g interfaces.KlaberjassGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerIdx := g.GetWinnerIdx()
		player := g.GetPlayer(winnerIdx)
		isHuman := player != nil && player.GetIsHuman()
		return buildWinnerWebMessage("klaberjass", winnerIdx, isHuman)
	}
	switch g.GetPhase() {
	case domain.KlaberjassPhaseBidTurnUp:
		return "", "klaberjass.bidTurnUp", nil
	case domain.KlaberjassPhaseBidFree:
		return "", "klaberjass.bidFree", nil
	case domain.KlaberjassPhaseSchmeiss:
		return "", "klaberjass.schmeissPending", nil
	case domain.KlaberjassPhasePlay:
		return "", "klaberjass.playPhase", nil
	case domain.KlaberjassPhaseHandEnd:
		if g.IsBete() {
			return "", "klaberjass.bete", nil
		}
		return "", "klaberjass.handEnd", nil
	}
	return "", "", nil
}

// ActionLogOutput 棋譜をJSON出力
func (p *KlaberjassWebPresenter) ActionLogOutput(g interfaces.KlaberjassGame) string {
	return actionLogOutputJSON(g)
}
