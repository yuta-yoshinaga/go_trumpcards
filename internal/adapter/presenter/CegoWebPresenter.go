//go:build !js || !wasm || solo

package presenter

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// cegoFace は Cego の 54 枚タロックデッキ全札を手続き描画するための自己記述子を返す。スート札は
// 隅ラベルにランク (1-4, J/C/Q/K)、グリフにスート記号 (♠♥♦♣)、黒スート (スペード/クラブ) は
// black、赤スート (ハート/ダイヤ) は red。切り札は番号 1-21 をラベルに、グリフ "✦"、色 purple。
// スキュースはグリフ "★"、ラベル "Sküs"、色 gold。deck:"tarot" を付与しフロントエンドの手続き
// 描画パスへ切り替える。詳細は ADR-0033。
func cegoFace(card *domain.Card) *CardFace {
	if card == nil {
		return nil
	}
	switch card.GetDesign() {
	case domain.CegoSkusDesign:
		return &CardFace{Glyph: "★", Label: "Sküs", Color: "gold", Deck: "tarot"}
	case domain.CegoTrumpDesign:
		label := strconv.Itoa(card.GetValue())
		return &CardFace{Glyph: "✦", Label: label, Color: "purple", Deck: "tarot"}
	default:
		glyphs := map[int]string{
			domain.CardDesignSpade:   "♠",
			domain.CardDesignClover:  "♣",
			domain.CardDesignHeart:   "♥",
			domain.CardDesignDiamond: "♦",
		}
		colors := map[int]string{
			domain.CardDesignSpade:   "black",
			domain.CardDesignClover:  "black",
			domain.CardDesignHeart:   "red",
			domain.CardDesignDiamond: "red",
		}
		return &CardFace{
			Glyph: glyphs[card.GetDesign()],
			Label: cegoRankLabel(card.GetValue()),
			Color: colors[card.GetDesign()],
			Deck:  "tarot",
		}
	}
}

// cegoRankLabel スート札のランクラベルを返す (1-4 は数字、コート札は J/C/Q/K)。
func cegoRankLabel(value int) string {
	switch value {
	case 5:
		return "J" // Jack
	case 6:
		return "C" // Cavalier
	case 7:
		return "Q" // Queen
	case 8:
		return "K" // King
	default:
		return strconv.Itoa(value)
	}
}

// CegoWebPresenter チェゴ (Cego) のWebプレゼンタークラス
type CegoWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *CegoWebPresenter) Output(g interfaces.CegoGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *CegoWebPresenter) buildBase(g interfaces.CegoGame) *controller.CegoWebOutput {
	resObj := new(controller.CegoWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.BidPlayerIdx = g.GetBidPlayerIdx()
	resObj.HighestBid = int(g.GetHighestBid())
	resObj.HighestBidder = g.GetHighestBidder()
	resObj.DeclarerIdx = g.GetDeclarerIdx()
	resObj.Contract = int(g.GetContract())
	resObj.ContractType = int(g.GetContractType())
	resObj.BlindCount = g.GetBlindCount()
	resObj.StashOwner = g.GetStashOwner()
	resObj.Outcome = int(g.GetOutcome())
	resObj.Result = int(g.GetResult())
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerPlayer = g.GetWinnerPlayer()
	resObj.PlayerScores = g.GetPlayerScores()
	resObj.LastTrickWinner = -1
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.IsHumanBidTurn = g.IsHumanBidTurn()
	resObj.IsHumanContract = g.IsHumanContractTurn()
	resObj.IsHumanExchange = g.IsHumanExchangeTurn()

	resObj.PlayableIndices = p.playableIndices(g)

	cfg := g.GetConfig()
	resObj.Config = controller.CegoWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetDeals:   cfg.TargetDeals,
	}

	// 場札 (Cego / blind) の中身は決して公開しない (伏せ札)。BlindCount のみを出力する。
	resObj.Blind = make([]*controller.WebOutputCard, 0)
	resObj.CurrentTrick = p.buildTrickOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// playableIndices 人間プレイヤーがプレイできるカードのインデックスを返す
func (p *CegoWebPresenter) playableIndices(g interfaces.CegoGame) []int {
	if g.GetPhase() != domain.CegoPhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// buildTrickOutput 現在のトリック情報を構築
func (p *CegoWebPresenter) buildTrickOutput(trick []*domain.TrickCard) []*controller.CegoWebOutputTrickCard {
	return buildTrickCards(trick, func(tc *domain.TrickCard) *controller.CegoWebOutputTrickCard {
		return &controller.CegoWebOutputTrickCard{PlayerIdx: tc.PlayerIdx, Card: cardToOutputWithFace(tc.Card, cegoFace)}
	})
}

// buildPlayersOutput プレイヤー情報を構築 (人間のみ手札を公開)。
func (p *CegoWebPresenter) buildPlayersOutput(g interfaces.CegoGame) []*controller.CegoWebOutputPlayer {
	scores := g.GetPlayerScores()
	declarer := g.GetDeclarerIdx()
	out := make([]*controller.CegoWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		out = append(out, &controller.CegoWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutputWithFace(player, player.GetIsHuman(), cegoFace),
			TrickCount: player.GetTrickCount(),
			CardPoints: g.GetCardPoints(i),
			Score:      scores[i],
			IsDeclarer: i == declarer,
		})
	}
	return out
}

// buildMessage ゲーム結果/フェーズメッセージを構築
func (p *CegoWebPresenter) buildMessage(g interfaces.CegoGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.CegoPhaseBid:
		return "", "cego.bidPhase", nil
	case domain.CegoPhaseContract:
		return "", "cego.contractPhase", nil
	case domain.CegoPhaseExchange:
		return "", "cego.exchangePhase", nil
	case domain.CegoPhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "cego.playPhase.lead", nil
		}
		return "", "cego.playPhase.follow", nil
	case domain.CegoPhaseTrickEnd:
		return "", "cego.trickEnd", nil
	case domain.CegoPhaseRoundEnd:
		return "", cegoOutcomeMessageCode(g.GetOutcome()), nil
	}
	return "", "", nil
}

// cegoOutcomeMessageCode ディール結果に対応するメッセージコードを返す。
func cegoOutcomeMessageCode(o domain.CegoOutcome) string {
	switch o {
	case domain.CegoOutcomeWin:
		return "cego.roundEnd.win"
	case domain.CegoOutcomeLoss:
		return "cego.roundEnd.loss"
	default:
		return "cego.roundEnd"
	}
}

// winnerMessage 勝者プレイヤーメッセージを構築する
func (p *CegoWebPresenter) winnerMessage(g interfaces.CegoGame) (string, string, map[string]string) {
	winner := g.GetWinnerPlayer()
	if winner < 0 {
		return "ゲーム終了！ 引き分け！", "cego.result.draw", nil
	}
	humanIdx := -1
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if player := g.GetPlayer(i); player != nil && player.GetIsHuman() {
			humanIdx = i
			break
		}
	}
	if humanIdx >= 0 && winner == humanIdx {
		return "ゲーム終了！ あなたの勝ち！", "cego.result.humanWin", nil
	}
	params := map[string]string{"player": fmt.Sprintf("%d", winner)}
	return fmt.Sprintf("ゲーム終了！ プレイヤー%dの勝ち！", winner), "cego.result.cpuWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *CegoWebPresenter) HintOutput(g interfaces.CegoGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.CegoWebOutputHint{
			Bid:         hint.Bid,
			Contract:    hint.Contract,
			CardIndices: hint.CardIndices,
			Reason:      hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *CegoWebPresenter) ActionLogOutput(g interfaces.CegoGame) string {
	return actionLogOutputJSON(g)
}
