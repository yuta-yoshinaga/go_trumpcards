//go:build !js || !wasm || extra2

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SjavsWebPresenter シャウスWebプレゼンタークラス
type SjavsWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *SjavsWebPresenter) Output(c interfaces.SjavsGame, lastErr error) string {
	resObj := p.buildBase(c)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(c, lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *SjavsWebPresenter) buildBase(c interfaces.SjavsGame) *controller.SjavsWebOutput {
	resObj := new(controller.SjavsWebOutput)
	resObj.Phase = int(c.GetPhase())
	resObj.CurrentPlayerIdx = c.GetCurrentPlayerIdx()
	resObj.DealerIdx = c.GetDealerIdx()
	resObj.TrumpSuit = c.GetTrumpSuit()
	// 切札の総枚数はサーバーで数える。常時切札 6 枚を含むので、クライアントが
	// 「切札スートの枚数」だけ数えると必ず足りない。
	if resObj.TrumpSuit >= 0 {
		resObj.TrumpCount = domain.SjavsTrumpCount(resObj.TrumpSuit)
	}
	resObj.BidderIdx = c.GetBidderIdx()
	resObj.BidLength = c.GetBidLength()
	resObj.MinBid = domain.SjavsMinBid
	resObj.MyLongest = c.LongestTrumpLength(0)
	resObj.TrickNo = c.GetTrickNumber()
	resObj.CarryOver = c.GetCarryOver()
	resObj.GameEndFlag = c.GetGameEndFlag()
	resObj.WinnerTeam = c.GetWinnerTeam()
	resObj.DoubleVictory = c.IsDoubleVictory()

	resObj.TeamPoints = []int{c.GetTeamPoints(0), c.GetTeamPoints(1)}
	resObj.Remaining = []int{c.GetRemaining(0), c.GetRemaining(1)}
	resObj.Crosses = []int{c.GetCrosses(0), c.GetCrosses(1)}

	trick := c.GetTrick()
	resObj.Trick = make([]*controller.SjavsWebOutputTrickCard, 0, len(trick))
	for _, tc := range trick {
		if tc.Card == nil {
			continue
		}
		resObj.Trick = append(resObj.Trick, &controller.SjavsWebOutputTrickCard{
			PlayerIdx: tc.PlayerIdx,
			Card:      cardToOutput(tc.Card),
		})
	}

	valid := c.GetValidPlayIndices(0)
	resObj.ValidIndices = make([]int, 0, len(valid))
	if !c.GetGameEndFlag() && c.GetCurrentPlayerIdx() == 0 && c.GetPhase() == domain.SjavsPhasePlay {
		resObj.ValidIndices = append(resObj.ValidIndices, valid...)
	}

	// 切札の添字。**常時切札の 6 枚はスートを見ても分からない**ので、6 枚の
	// 一覧をクライアントに書き写させず、強さを決めている判定を通す (#5575)。
	resObj.TrumpIndices = make([]int, 0, domain.SjavsHandSize)
	if trump := c.GetTrumpSuit(); trump >= 0 {
		if human := c.GetPlayer(0); human != nil {
			for j := range human.GetCardsSize() {
				if domain.SjavsIsTrump(human.GetCard(j), trump) {
					resObj.TrumpIndices = append(resObj.TrumpIndices, j)
				}
			}
		}
	}

	if hr := c.GetHandResult(); hr != nil {
		resObj.HandResult = &controller.SjavsWebOutputHandResult{
			DeclarerTeam:   hr.DeclarerTeam,
			DeclarerPoints: hr.DeclarerPoints,
			ScoringTeam:    hr.ScoringTeam,
			Amount:         hr.Amount,
			Vol:            hr.Vol,
			TrumpWasClubs:  hr.TrumpWasClubs,
		}
	}

	cfg := c.GetConfig()
	resObj.Config = controller.SjavsWebOutputConfig{CpuDifficulty: int(cfg.CpuDifficulty)}
	resObj.Players = p.buildPlayersOutput(c)

	// ヒントは通常のレスポンスにも載せる。HintOutput にしか設定しないと、
	// フロントは通常の state を読むので何も表示されない。
	if !c.GetGameEndFlag() && c.GetCurrentPlayerIdx() == 0 {
		resObj.Hint = sjavsHint(c)
	}
	return resObj
}

// buildPlayersOutput プレイヤー情報を構築する。
//
// CPU の手札は伏せる。**申告枚数は公開**する -- 卓上で聞こえる宣言であり、
// 味方が何枚持っているかはパートナーシップの読みそのもの。
func (p *SjavsWebPresenter) buildPlayersOutput(c interfaces.SjavsGame) []*controller.SjavsWebOutputPlayer {
	players := c.GetPlayers()
	bids := c.GetBids()
	out := make([]*controller.SjavsWebOutputPlayer, 0, len(players))
	for i, player := range players {
		if player == nil {
			continue
		}
		reveal := player.GetIsHuman() || c.GetGameEndFlag()
		cards := make([]*controller.WebOutputCard, 0, player.GetCardsSize())
		if reveal {
			for j := range player.GetCardsSize() {
				if card := player.GetCard(j); card != nil {
					cards = append(cards, cardToOutput(card))
				}
			}
		}
		bid := 0
		if i < len(bids) {
			bid = bids[i]
		}
		out = append(out, &controller.SjavsWebOutputPlayer{
			ID:        i,
			IsHuman:   player.GetIsHuman(),
			Team:      domain.SjavsTeamOf(i),
			CardCount: player.GetCardsSize(),
			Cards:     cards,
			Bid:       bid,
			Hidden:    !reveal,
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *SjavsWebPresenter) buildMessage(c interfaces.SjavsGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if !c.GetGameEndFlag() {
		return "", "", nil
	}
	if c.GetWinnerTeam() == domain.SjavsTeamOf(0) {
		if c.IsDoubleVictory() {
			return "double victory", "sjavs.winDouble", nil
		}
		return "you win the rubber", "sjavs.win", nil
	}
	if c.IsDoubleVictory() {
		return "double defeat", "sjavs.loseDouble", nil
	}
	return "you lose the rubber", "sjavs.lose", nil
}

// HintOutput ヒント情報を出力する
func (p *SjavsWebPresenter) HintOutput(c interfaces.SjavsGame) string {
	resObj := p.buildBase(c)
	resObj.Hint = sjavsHint(c)
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜を出力する
func (p *SjavsWebPresenter) ActionLogOutput(c interfaces.SjavsGame) string {
	return actionLogOutputJSON(c)
}

// sjavsHint 人間プレイヤーへの推奨手を返す。CPU と同じ意思決定を通す。
func sjavsHint(c interfaces.SjavsGame) *controller.SjavsWebOutputHint {
	if c.GetGameEndFlag() {
		return &controller.SjavsWebOutputHint{Reason: "sjavs.hint.game_end"}
	}
	if c.GetCurrentPlayerIdx() != 0 {
		return &controller.SjavsWebOutputHint{Reason: "sjavs.hint.not_your_turn"}
	}
	action := c.SjavsCpuDecide(0)
	if c.GetPhase() == domain.SjavsPhaseBid {
		n := action.BidLength
		if n == 0 {
			return &controller.SjavsWebOutputHint{BidLength: &n, Reason: "sjavs.hint.pass"}
		}
		return &controller.SjavsWebOutputHint{BidLength: &n, Reason: "sjavs.hint.bid"}
	}
	if action.HandIdx < 0 {
		return &controller.SjavsWebOutputHint{Reason: "sjavs.hint.none"}
	}
	idx := action.HandIdx
	return &controller.SjavsWebOutputHint{CardIndex: &idx, Reason: "sjavs.hint.play"}
}
