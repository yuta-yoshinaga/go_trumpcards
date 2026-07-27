//go:build !js || !wasm || extra

package presenter

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// frenchTarotFace は French Tarot の 78 枚デッキ全札を手続き描画するための自己記述子を返す。
// スート札は隅ラベルにランク (1-10, V/C/D/R)、グリフにスート記号 (♠♥♦♣) を用い、黒スート
// (スペード/クラブ) は black、赤スート (ハート/ダイヤ) は red。切り札は番号 1-21 をラベルに、
// グリフ "✦"、色 purple。エクスキューズはグリフ "★"、ラベル "Excuse"、色 gold。deck:"tarot"
// を付与しフロントエンドの手続き描画パスへ切り替える。詳細は ADR-0033。
func frenchTarotFace(card *domain.Card) *CardFace {
	if card == nil {
		return nil
	}
	switch card.GetDesign() {
	case domain.FrenchTarotExcuseDesign:
		return &CardFace{Glyph: "★", Label: "Excuse", Color: "gold", Deck: "tarot"}
	case domain.FrenchTarotTrumpDesign:
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
			Label: frenchTarotRankLabel(card.GetValue()),
			Color: colors[card.GetDesign()],
			Deck:  "tarot",
		}
	}
}

// frenchTarotRankLabel スート札のランクラベルを返す (1-10 は数字、コート札は V/C/D/R)。
func frenchTarotRankLabel(value int) string {
	switch value {
	case 11:
		return "V" // Valet
	case 12:
		return "C" // Cavalier
	case 13:
		return "D" // Dame
	case 14:
		return "R" // Roi
	default:
		return strconv.Itoa(value)
	}
}

// FrenchTarotWebPresenter フレンチタロット (French Tarot) のWebプレゼンタークラス
type FrenchTarotWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *FrenchTarotWebPresenter) Output(g interfaces.FrenchTarotGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *FrenchTarotWebPresenter) buildBase(g interfaces.FrenchTarotGame) *controller.FrenchTarotWebOutput {
	resObj := new(controller.FrenchTarotWebOutput)
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
	resObj.ChienCount = g.GetChienCount()
	resObj.ChienRevealed = g.GetChienRevealed()
	resObj.StashOwner = g.GetStashOwner()
	resObj.Outcome = int(g.GetOutcome())
	resObj.Result = int(g.GetResult())
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerPlayer = g.GetWinnerPlayer()
	resObj.PlayerScores = g.GetPlayerScores()
	resObj.LastTrickWinner = -1
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.IsHumanBidTurn = g.IsHumanBidTurn()
	resObj.IsHumanDiscard = g.IsHumanDiscardTurn()

	resObj.PlayableIndices = p.playableIndices(g)

	cfg := g.GetConfig()
	resObj.Config = controller.FrenchTarotWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetDeals:   cfg.TargetDeals,
	}

	resObj.Chien = p.buildChienOutput(g)
	resObj.CurrentTrick = p.buildTrickOutput(g.GetCurrentTrick())
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// playableIndices 人間プレイヤーがプレイできるカードのインデックスを返す
func (p *FrenchTarotWebPresenter) playableIndices(g interfaces.FrenchTarotGame) []int {
	if g.GetPhase() != domain.FrenchTarotPhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// buildChienOutput シアンを構築。人間デクレアラーのシアン交換フェーズでのみ内容を公開する。
func (p *FrenchTarotWebPresenter) buildChienOutput(g interfaces.FrenchTarotGame) []*controller.WebOutputCard {
	if !g.IsHumanDiscardTurn() {
		return make([]*controller.WebOutputCard, 0)
	}
	chien := g.GetChien()
	out := make([]*controller.WebOutputCard, 0, len(chien))
	for _, c := range chien {
		out = append(out, cardToOutputWithFace(c, frenchTarotFace))
	}
	return out
}

// buildTrickOutput 現在のトリック情報を構築
func (p *FrenchTarotWebPresenter) buildTrickOutput(trick []*domain.TrickCard) []*controller.FrenchTarotWebOutputTrickCard {
	return buildTrickCards(trick, func(tc *domain.TrickCard) *controller.FrenchTarotWebOutputTrickCard {
		return &controller.FrenchTarotWebOutputTrickCard{PlayerIdx: tc.PlayerIdx, Card: cardToOutputWithFace(tc.Card, frenchTarotFace)}
	})
}

// buildPlayersOutput プレイヤー情報を構築 (人間のみ手札を公開)
func (p *FrenchTarotWebPresenter) buildPlayersOutput(g interfaces.FrenchTarotGame) []*controller.FrenchTarotWebOutputPlayer {
	scores := g.GetPlayerScores()
	declarer := g.GetDeclarerIdx()
	out := make([]*controller.FrenchTarotWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		out = append(out, &controller.FrenchTarotWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutputWithFace(player, player.GetIsHuman(), frenchTarotFace),
			TrickCount: player.GetTrickCount(),
			CardPoints: g.GetCardPoints(i),
			Score:      scores[i],
			IsDeclarer: i == declarer,
		})
	}
	return out
}

// buildMessage ゲーム結果/フェーズメッセージを構築
func (p *FrenchTarotWebPresenter) buildMessage(g interfaces.FrenchTarotGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.FrenchTarotPhaseBid:
		return "", "frenchtarot.bidPhase", nil
	case domain.FrenchTarotPhaseChien:
		return "", "frenchtarot.chienPhase", nil
	case domain.FrenchTarotPhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "frenchtarot.playPhase.lead", nil
		}
		return "", "frenchtarot.playPhase.follow", nil
	case domain.FrenchTarotPhaseTrickEnd:
		return "", "frenchtarot.trickEnd", nil
	case domain.FrenchTarotPhaseRoundEnd:
		return "", frenchTarotOutcomeMessageCode(g.GetOutcome()), nil
	}
	return "", "", nil
}

// frenchTarotOutcomeMessageCode ディール結果に対応するメッセージコードを返す。
func frenchTarotOutcomeMessageCode(o domain.FrenchTarotOutcome) string {
	switch o {
	case domain.FrenchTarotOutcomeWin:
		return "frenchtarot.roundEnd.win"
	case domain.FrenchTarotOutcomeLoss:
		return "frenchtarot.roundEnd.loss"
	default:
		return "frenchtarot.roundEnd"
	}
}

// winnerMessage 勝者プレイヤーメッセージを構築する
func (p *FrenchTarotWebPresenter) winnerMessage(g interfaces.FrenchTarotGame) (string, string, map[string]string) {
	winner := g.GetWinnerPlayer()
	if winner < 0 {
		return "ゲーム終了！ 引き分け！", "frenchtarot.result.draw", nil
	}
	humanIdx := -1
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if player := g.GetPlayer(i); player != nil && player.GetIsHuman() {
			humanIdx = i
			break
		}
	}
	if humanIdx >= 0 && winner == humanIdx {
		return "ゲーム終了！ あなたの勝ち！", "frenchtarot.result.humanWin", nil
	}
	params := map[string]string{"player": fmt.Sprintf("%d", winner)}
	return fmt.Sprintf("ゲーム終了！ プレイヤー%dの勝ち！", winner), "frenchtarot.result.cpuWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *FrenchTarotWebPresenter) HintOutput(g interfaces.FrenchTarotGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.FrenchTarotWebOutputHint{
			Bid:         hint.Bid,
			CardIndices: hint.CardIndices,
			Reason:      hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *FrenchTarotWebPresenter) ActionLogOutput(g interfaces.FrenchTarotGame) string {
	return actionLogOutputJSON(g)
}
