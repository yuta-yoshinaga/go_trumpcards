//go:build !js || !wasm || extra3

package presenter

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// rookFace は Rook の札(非52枚デッキ)を手続き描画するための自己記述子を返す。
// 4色を CardFace の INK_COLORS に存在する読みやすいトークンへ対応付け
// (黄色→gold)、ルーク鳥は 🐦 グリフ + purple で描画する。deck:"rook" を
// 付与することでフロントエンドは手続き描画パスへ切り替える。詳細は ADR-0033。
func rookFace(card *domain.Card) *CardFace {
	if card == nil {
		return nil
	}
	if card.GetDesign() == domain.RookBirdDesign {
		return &CardFace{Glyph: "🐦", Label: "Rook", Color: "purple", Deck: "rook"}
	}
	colors := map[int]string{1: "red", 2: "gold", 3: "green", 4: "black"} // yellow→gold for readability
	label := strconv.Itoa(card.GetValue())
	return &CardFace{Glyph: label, Label: label, Color: colors[card.GetDesign()], Deck: "rook"}
}

// RookWebPresenter ルーク(Rook) Webプレゼンタークラス
type RookWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *RookWebPresenter) Output(g interfaces.RookGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**Rook.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.RookWebOutputHint{
			Bid:            hint.Bid,
			Pass:           hint.Pass,
			DiscardIndices: hint.DiscardIndices,
			TrumpColor:     hint.TrumpColor,
			CardIndex:      hint.CardIndex,
			Reason:         hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *RookWebPresenter) buildBase(g interfaces.RookGame) *controller.RookWebOutput {
	resObj := new(controller.RookWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.BidPlayerIdx = g.GetBidPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.TrumpColor = g.GetTrumpColor()
	resObj.ContractBid = g.GetContractBid()
	resObj.DeclarerIdx = g.GetDeclarerIdx()
	resObj.HighestBid = g.GetHighestBid()
	resObj.HighestBidder = g.GetHighestBidder()
	resObj.NestCount = len(g.GetNest())
	resObj.TeamScores = [2]int{g.GetTeamScore(0), g.GetTeamScore(1)}
	resObj.TeamPoints = [2]int{g.GetTeamPoints(0), g.GetTeamPoints(1)}
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerTeam = g.GetWinnerTeam()

	cfg := g.GetConfig()
	resObj.Config = controller.RookWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetScore:   cfg.TargetScore,
	}

	resObj.Nest = p.buildNestOutput(g)
	resObj.CurrentTrick = trickCardsToOutputWithFace(g.GetCurrentTrick(), rookFace)
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// buildNestOutput ネストを構築。落札者(人間)のネスト交換フェーズでのみ内容を公開する。
func (p *RookWebPresenter) buildNestOutput(g interfaces.RookGame) []*controller.WebOutputCard {
	reveal := g.GetPhase() == domain.RookPhaseNestExchange && g.GetDeclarerIdx() >= 0 &&
		g.GetPlayer(g.GetDeclarerIdx()) != nil && g.GetPlayer(g.GetDeclarerIdx()).GetIsHuman()
	if !reveal {
		return make([]*controller.WebOutputCard, 0)
	}
	nest := g.GetNest()
	out := make([]*controller.WebOutputCard, 0, len(nest))
	for _, c := range nest {
		out = append(out, cardToOutputWithFace(c, rookFace))
	}
	return out
}

// buildPlayersOutput プレイヤー情報を構築 (人間のみ手札を公開)
func (p *RookWebPresenter) buildPlayersOutput(g interfaces.RookGame) []*controller.RookWebOutputPlayer {
	out := make([]*controller.RookWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		out = append(out, &controller.RookWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutputWithFace(player, player.GetIsHuman(), rookFace),
			Team:       player.GetTeam(),
			TrickCount: player.GetTrickCount(),
			Points:     player.GetPoints(),
			Bid:        player.GetBid(),
			Passed:     player.GetPassed(),
			IsDeclarer: player.GetIsDeclarer(),
		})
	}
	return out
}

// buildMessage ゲーム結果/フェーズメッセージを構築
func (p *RookWebPresenter) buildMessage(g interfaces.RookGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		winnerTeam := g.GetWinnerTeam()
		msg := fmt.Sprintf("ゲーム終了！ チーム%dの勝ち！", winnerTeam)
		code := fmt.Sprintf("rook.result.team%dWin", winnerTeam)
		return msg, code, map[string]string{"team": fmt.Sprintf("%d", winnerTeam)}
	}
	switch g.GetPhase() {
	case domain.RookPhaseBid:
		return "", "rook.bidPhase", nil
	case domain.RookPhaseNestExchange:
		return "", "rook.nestExchangePhase", nil
	case domain.RookPhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "rook.playPhase.lead", nil
		}
		return "", "rook.playPhase.follow", nil
	case domain.RookPhaseTrickEnd:
		return "", "rook.trickEnd", nil
	case domain.RookPhaseRoundEnd:
		return "", "rook.roundEnd", nil
	}
	return "", "", nil
}

// HintOutput ヒント情報をJSON出力する
func (p *RookWebPresenter) HintOutput(g interfaces.RookGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.RookWebOutputHint{
			Bid:            hint.Bid,
			Pass:           hint.Pass,
			DiscardIndices: hint.DiscardIndices,
			TrumpColor:     hint.TrumpColor,
			CardIndex:      hint.CardIndex,
			Reason:         hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *RookWebPresenter) ActionLogOutput(g interfaces.RookGame) string {
	return actionLogOutputJSON(g)
}
