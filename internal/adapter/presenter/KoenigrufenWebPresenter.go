//go:build !js || !wasm || extra

package presenter

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// koenigrufenFace は Königrufen の 54 枚タロックデッキ全札を手続き描画するための自己記述子を
// 返す。スート札は隅ラベルにランク (1-4, J/C/Q/K)、グリフにスート記号 (♠♥♦♣)、黒スート
// (スペード/クラブ) は black、赤スート (ハート/ダイヤ) は red。切り札は番号 1-21 をラベルに、
// グリフ "✦"、色 purple。スキュースはグリフ "★"、ラベル "Sküs"、色 gold。deck:"tarot" を
// 付与しフロントエンドの手続き描画パスへ切り替える。詳細は ADR-0033。
func koenigrufenFace(card *domain.Card) *CardFace {
	if card == nil {
		return nil
	}
	switch card.GetDesign() {
	case domain.KoenigrufenSkusDesign:
		return &CardFace{Glyph: "★", Label: "Sküs", Color: "gold", Deck: "tarot"}
	case domain.KoenigrufenTrumpDesign:
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
			Label: koenigrufenRankLabel(card.GetValue()),
			Color: colors[card.GetDesign()],
			Deck:  "tarot",
		}
	}
}

// koenigrufenRankLabel スート札のランクラベルを返す (1-4 は数字、コート札は J/C/Q/K)。
func koenigrufenRankLabel(value int) string {
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

// KoenigrufenWebPresenter ケーニッヒルーフェン (Königrufen) のWebプレゼンタークラス
type KoenigrufenWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *KoenigrufenWebPresenter) Output(g interfaces.KoenigrufenGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **Koenigrufen.GetHint() の各フェーズを読んで、席を確かめていることを確認した。**
	// 他ゲームがそうだから、で済ませない —— Pinochle は見ていなかった (#4585)。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.KoenigrufenWebOutputHint{
			Bid:         hint.Bid,
			CallSuit:    hint.CallSuit,
			CardIndices: hint.CardIndices,
			Reason:      hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *KoenigrufenWebPresenter) buildBase(g interfaces.KoenigrufenGame) *controller.KoenigrufenWebOutput {
	resObj := new(controller.KoenigrufenWebOutput)
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
	resObj.CalledKing = g.GetCalledKing()
	resObj.PartnerRevealed = g.GetPartnerRevealed()
	// 秘密のパートナーは公開されるまで漏らさない: partnerIdx は -1 を出力する。
	resObj.PartnerIdx = koenigrufenVisiblePartner(g)
	resObj.TalonCount = g.GetTalonCount()
	resObj.StashOwner = g.GetStashOwner()
	resObj.Outcome = int(g.GetOutcome())
	resObj.Result = int(g.GetResult())
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerPlayer = g.GetWinnerPlayer()
	resObj.PlayerScores = g.GetPlayerScores()
	resObj.LastTrickWinner = -1
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.IsHumanBidTurn = g.IsHumanBidTurn()
	resObj.IsHumanCall = g.IsHumanCallTurn()
	resObj.IsHumanDiscard = g.IsHumanDiscardTurn()

	resObj.PlayableIndices = p.playableIndices(g)

	cfg := g.GetConfig()
	resObj.Config = controller.KoenigrufenWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetDeals:   cfg.TargetDeals,
	}

	resObj.Talon = p.buildTalonOutput(g)
	resObj.CurrentTrick = trickCardsToOutputWithFace(g.GetCurrentTrick(), koenigrufenFace)
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// koenigrufenVisiblePartner 公開してよいパートナーインデックスを返す。partnerRevealed=false の
// 間は常に -1 (秘密のパートナーを漏らさない)。
func koenigrufenVisiblePartner(g interfaces.KoenigrufenGame) int {
	if !g.GetPartnerRevealed() {
		return -1
	}
	return g.GetPartnerIdx()
}

// playableIndices 人間プレイヤーがプレイできるカードのインデックスを返す
func (p *KoenigrufenWebPresenter) playableIndices(g interfaces.KoenigrufenGame) []int {
	if g.GetPhase() != domain.KoenigrufenPhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// buildTalonOutput 場札を構築。人間デクレアラーの場札交換フェーズでのみ内容を公開する。
func (p *KoenigrufenWebPresenter) buildTalonOutput(g interfaces.KoenigrufenGame) []*controller.WebOutputCard {
	if !g.IsHumanDiscardTurn() {
		return make([]*controller.WebOutputCard, 0)
	}
	talon := g.GetTalon()
	out := make([]*controller.WebOutputCard, 0, len(talon))
	for _, c := range talon {
		out = append(out, cardToOutputWithFace(c, koenigrufenFace))
	}
	return out
}

// buildPlayersOutput プレイヤー情報を構築 (人間のみ手札を公開)。IsPartner は partnerRevealed=true
// のときのみ真になり得る (秘密のパートナーを漏らさない)。
func (p *KoenigrufenWebPresenter) buildPlayersOutput(g interfaces.KoenigrufenGame) []*controller.KoenigrufenWebOutputPlayer {
	scores := g.GetPlayerScores()
	declarer := g.GetDeclarerIdx()
	partner := koenigrufenVisiblePartner(g)
	out := make([]*controller.KoenigrufenWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		out = append(out, &controller.KoenigrufenWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutputWithFace(player, player.GetIsHuman(), koenigrufenFace),
			TrickCount: player.GetTrickCount(),
			CardPoints: g.GetCardPoints(i),
			Score:      scores[i],
			IsDeclarer: i == declarer,
			IsPartner:  partner >= 0 && i == partner,
		})
	}
	return out
}

// buildMessage ゲーム結果/フェーズメッセージを構築
func (p *KoenigrufenWebPresenter) buildMessage(g interfaces.KoenigrufenGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.KoenigrufenPhaseBid:
		return "", "koenigrufen.bidPhase", nil
	case domain.KoenigrufenPhaseCall:
		return "", "koenigrufen.callPhase", nil
	case domain.KoenigrufenPhaseTalon:
		return "", "koenigrufen.talonPhase", nil
	case domain.KoenigrufenPhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "koenigrufen.playPhase.lead", nil
		}
		return "", "koenigrufen.playPhase.follow", nil
	case domain.KoenigrufenPhaseTrickEnd:
		return "", "koenigrufen.trickEnd", nil
	case domain.KoenigrufenPhaseRoundEnd:
		return "", koenigrufenOutcomeMessageCode(g.GetOutcome()), nil
	}
	return "", "", nil
}

// koenigrufenOutcomeMessageCode ディール結果に対応するメッセージコードを返す。
func koenigrufenOutcomeMessageCode(o domain.KoenigrufenOutcome) string {
	switch o {
	case domain.KoenigrufenOutcomeWin:
		return "koenigrufen.roundEnd.win"
	case domain.KoenigrufenOutcomeLoss:
		return "koenigrufen.roundEnd.loss"
	default:
		return "koenigrufen.roundEnd"
	}
}

// winnerMessage 勝者プレイヤーメッセージを構築する
func (p *KoenigrufenWebPresenter) winnerMessage(g interfaces.KoenigrufenGame) (string, string, map[string]string) {
	winner := g.GetWinnerPlayer()
	if winner < 0 {
		return "ゲーム終了！ 引き分け！", "koenigrufen.result.draw", nil
	}
	humanIdx := -1
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if player := g.GetPlayer(i); player != nil && player.GetIsHuman() {
			humanIdx = i
			break
		}
	}
	if humanIdx >= 0 && winner == humanIdx {
		return "ゲーム終了！ あなたの勝ち！", "koenigrufen.result.humanWin", nil
	}
	params := map[string]string{"player": fmt.Sprintf("%d", winner)}
	return fmt.Sprintf("ゲーム終了！ プレイヤー%dの勝ち！", winner), "koenigrufen.result.cpuWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *KoenigrufenWebPresenter) HintOutput(g interfaces.KoenigrufenGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.KoenigrufenWebOutputHint{
			Bid:         hint.Bid,
			CallSuit:    hint.CallSuit,
			CardIndices: hint.CardIndices,
			Reason:      hint.Reason,
		}
	}
	// **「頼んだヒントか」を CLI が見分けられるようにする。**このゲーム群の
	// `hintAvailable` は画面のラベルとして既に使われているので、別キーを出す (#4483)。
	if g.GetHint() != nil {
		resObj.MessageCode = "koenigrufen.hintRequested"
	} else {
		resObj.MessageCode = "koenigrufen.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *KoenigrufenWebPresenter) ActionLogOutput(g interfaces.KoenigrufenGame) string {
	return actionLogOutputJSON(g)
}
