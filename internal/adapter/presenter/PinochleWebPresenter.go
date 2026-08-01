package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PinochleWebPresenter ピノクルWebプレゼンタークラス
type PinochleWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *PinochleWebPresenter) Output(g interfaces.PinochleGame, lastErr error) string {
	resObj := p.buildBase(g, lastErr)

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**Pinochle.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.PinochleWebOutputHint{
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
// **HintOutput もこれを使う。**以前はヒント構造体を裸で返していて `hint` キーが
// 無く、フロントの `state.hint` からは読めなかった (#4483)。
func (p *PinochleWebPresenter) buildBase(g interfaces.PinochleGame, lastErr error) *controller.PinochleWebOutput {
	resObj := new(controller.PinochleWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.BidPlayerIdx = g.GetBidPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.TrumpSuit = g.GetTrumpSuit()
	resObj.HighestBid = g.GetHighestBid()
	resObj.HighestBidder = g.GetHighestBidder()
	resObj.TeamScores = [2]int{g.GetTeamScore(0), g.GetTeamScore(1)}
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerTeam = g.GetWinnerTeam()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()

	// 設定
	cfg := g.GetConfig()
	resObj.Config = controller.PinochleWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		PointLimit:    cfg.PointLimit,
	}

	trick := g.GetCurrentTrick()
	resObj.CurrentTrick = trickCardsToOutput(trick)
	resObj.Players = p.buildPlayersOutput(g)
	resObj.PlayerMelds = p.buildMeldsOutput(g)

	// プレイフェーズでは有効なプレイインデックスを含める
	if g.GetPhase() == domain.PinochlePhasePlay && g.IsHumanTurn() {
		resObj.ValidPlayIndices = g.GetValidPlayIndices(g.GetCurrentPlayerIdx())
	}

	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)

	return resObj
}

// HintOutput ヒント情報を出力
func (p *PinochleWebPresenter) HintOutput(g interfaces.PinochleGame) string {
	// **状態レスポンスを返す。**以前はヒント構造体を裸で返していたので `hint`
	// キーが無く、フロントの `state.hint` からは読めなかった (#4483)。
	resObj := p.buildBase(g, nil)
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.PinochleWebOutputHint{
			CardIndex: hint.CardIndex,
			BidAmount: hint.BidAmount,
			Pass:      hint.Pass,
			Suit:      hint.Suit,
			Reason:    hint.Reason,
		}
		// **「頼んだヒントか」を CLI が見分けられるようにする。**このゲーム群の
		// `hintAvailable` は画面のラベルとして既に使われているので、別キーを出す。
		resObj.MessageCode = "pinochle.hintRequested"
	} else {
		resObj.MessageCode = "pinochle.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜を出力
func (p *PinochleWebPresenter) ActionLogOutput(g interfaces.PinochleGame) string {
	return actionLogToJSON(g.GetActionLog())
}

// buildPlayersOutput プレイヤー情報を構築
func (p *PinochleWebPresenter) buildPlayersOutput(g interfaces.PinochleGame) []*controller.PinochleWebOutputPlayer {
	out := make([]*controller.PinochleWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		showCards := player.GetIsHuman() || g.GetPhase() == domain.PinochlePhaseGameEnd
		pObj := &controller.PinochleWebOutputPlayer{
			ID:          i,
			IsHuman:     player.GetIsHuman(),
			CardCount:   player.GetCardsSize(),
			Cards:       playerCardsToOutput(player, showCards),
			Team:        player.GetTeam(),
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
func (p *PinochleWebPresenter) buildMeldsOutput(g interfaces.PinochleGame) [4][]*controller.PinochleWebOutputMeld {
	var out [4][]*controller.PinochleWebOutputMeld
	melds := g.GetPlayerMelds()
	for i := range 4 {
		playerMelds := make([]*controller.PinochleWebOutputMeld, 0)
		for _, m := range melds[i] {
			playerMelds = append(playerMelds, &controller.PinochleWebOutputMeld{
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
func (p *PinochleWebPresenter) buildMessage(g interfaces.PinochleGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerTeam := g.GetWinnerTeam()
		msg := fmt.Sprintf("ゲーム終了！ チーム%dの勝ち！", winnerTeam)
		code := fmt.Sprintf("pinochle.result.team%dWin", winnerTeam)
		params := map[string]string{"team": fmt.Sprintf("%d", winnerTeam)}
		return msg, code, params
	}
	switch g.GetPhase() {
	case domain.PinochlePhaseBid:
		return "", "pinochle.bidPhase", nil
	case domain.PinochlePhaseTrump:
		return "", "pinochle.trumpPhase", nil
	case domain.PinochlePhaseMeld:
		return "", "pinochle.meldPhase", nil
	case domain.PinochlePhasePlay:
		return "", "pinochle.playPhase", nil
	case domain.PinochlePhaseTrickEnd:
		return "", "pinochle.trickEndPhase", nil
	case domain.PinochlePhaseRoundEnd:
		return "", "pinochle.roundEndPhase", nil
	default:
		return "", "", nil
	}
}
