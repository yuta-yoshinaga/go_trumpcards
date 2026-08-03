//go:build !js || !wasm || solo

package presenter

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// tarocchiniFace は 62 枚デッキの札を手続き描画するための自己記述子を返す。
//
// **パパは色で区別する。**同格の 4 枚が普通の切り札と同じ見た目だと、後出しが
// 勝つ札かどうかが手札から読めない。ラベルの番号だけでは「2 が 3 より弱い」と
// 誤読されるので、色で群を示す (Ganjifa #4661 と同じ考え方)。
func tarocchiniFace(card *domain.Card) *CardFace {
	if card == nil {
		return nil
	}
	switch card.GetDesign() {
	case domain.TarocchiniMattoDesign:
		return &CardFace{Glyph: "★", Label: "Matto", Color: "gold", Deck: "tarot"}
	case domain.TarocchiniTrumpDesign:
		if domain.TarocchiniIsPapa(card) {
			return &CardFace{Glyph: "✦", Label: "Papa", Color: "green", Deck: "tarot"}
		}
		return &CardFace{Glyph: "✦", Label: strconv.Itoa(card.GetValue()), Color: "purple", Deck: "tarot"}
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
			Label: tarocchiniRankLabel(card.GetValue()),
			Color: colors[card.GetDesign()],
			Deck:  "tarot",
		}
	}
}

// tarocchiniRankLabel スート札のランクラベルを返す。
func tarocchiniRankLabel(value int) string {
	switch value {
	case 11:
		return "F" // Fante
	case 12:
		return "C" // Cavallo
	case 13:
		return "R" // Regina
	case domain.TarocchiniKingValue:
		return "Re"
	default:
		return strconv.Itoa(value)
	}
}

// TarocchiniWebPresenter タロッキーニのWebプレゼンタークラス
type TarocchiniWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *TarocchiniWebPresenter) Output(g interfaces.TarocchiniGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"` 専用の
	// レスポンスで、ページの state にはマージされない (#4483)。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.WebOutputCardHint{
			CardIndices: hint.CardIndices,
			Reason:      hint.Reason,
		}
	}
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *TarocchiniWebPresenter) buildBase(g interfaces.TarocchiniGame) *controller.TarocchiniWebOutput {
	resObj := new(controller.TarocchiniWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.RoundNumber = g.GetRoundNumber()
	resObj.TrickNumber = g.GetTrickNumber()
	resObj.CurrentPlayerIdx = g.GetCurrentPlayerIdx()
	resObj.LeadPlayerIdx = g.GetLeadPlayerIdx()
	resObj.DealerIdx = g.GetDealerIdx()
	resObj.ScartoCount = g.GetScartoSize()
	resObj.TeamScores = g.GetTeamScores()
	resObj.RoundTricks = g.GetRoundTricks()
	resObj.LastTrickWinner = g.GetLastTrickWinner()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerTeam = g.GetWinnerTeam()
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.IsHumanScarto = g.IsHumanScartoTurn()
	resObj.PlayableIndices = p.playableIndices(g)

	cfg := g.GetConfig()
	resObj.Config = controller.TarocchiniWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetRounds:  cfg.TargetRounds,
	}

	resObj.CurrentTrick = trickCardsToOutputWithFace(g.GetCurrentTrick(), tarocchiniFace)
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// playableIndices 人間プレイヤーがプレイできるカードのインデックスを返す
func (p *TarocchiniWebPresenter) playableIndices(g interfaces.TarocchiniGame) []int {
	if g.GetPhase() != domain.TarocchiniPhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// buildPlayersOutput プレイヤー情報を構築
func (p *TarocchiniWebPresenter) buildPlayersOutput(g interfaces.TarocchiniGame) []*controller.TarocchiniWebOutputPlayer {
	dealer := g.GetDealerIdx()
	out := make([]*controller.TarocchiniWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		out = append(out, &controller.TarocchiniWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutputWithFace(player, player.GetIsHuman(), tarocchiniFace),
			TrickCount: player.GetTrickCount(),
			Team:       domain.TarocchiniTeamOf(i),
			IsDealer:   i == dealer,
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *TarocchiniWebPresenter) buildMessage(g interfaces.TarocchiniGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.TarocchiniPhaseScarto:
		return "", "tarocchini.scartoPhase", nil
	case domain.TarocchiniPhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "tarocchini.playPhase.lead", nil
		}
		return "", "tarocchini.playPhase.follow", nil
	case domain.TarocchiniPhaseTrickEnd:
		return "", "tarocchini.trickEnd", nil
	case domain.TarocchiniPhaseRoundEnd:
		return "", "tarocchini.roundEnd", nil
	}
	return "", "", nil
}

// winnerMessage 勝利チームのメッセージを構築する。
//
// **勝敗はチーム単位。**人間の席そのものではなく、人間が属するチームが勝ったかを見る。
func (p *TarocchiniWebPresenter) winnerMessage(g interfaces.TarocchiniGame) (string, string, map[string]string) {
	winner := g.GetWinnerTeam()
	if winner < 0 {
		return "ゲーム終了！ 引き分け。", "tarocchini.result.draw", nil
	}
	humanTeam := -1
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if player := g.GetPlayer(i); player != nil && player.GetIsHuman() {
			humanTeam = domain.TarocchiniTeamOf(i)
			break
		}
	}
	if humanTeam == winner {
		return "ゲーム終了！ あなたのチームの勝ち！", "tarocchini.result.humanWin", nil
	}
	params := map[string]string{"team": fmt.Sprintf("%d", winner)}
	return fmt.Sprintf("ゲーム終了！ チーム%dの勝ち！", winner), "tarocchini.result.cpuWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *TarocchiniWebPresenter) HintOutput(g interfaces.TarocchiniGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = &controller.WebOutputCardHint{
			CardIndices: hint.CardIndices,
			Reason:      hint.Reason,
		}
		resObj.MessageCode = "tarocchini.hintRequested"
	} else {
		resObj.MessageCode = "tarocchini.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *TarocchiniWebPresenter) ActionLogOutput(g interfaces.TarocchiniGame) string {
	return actionLogOutputJSON(g)
}
