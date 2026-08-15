//go:build !js || !wasm || solo

package presenter

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// minchiateFace は 97 枚デッキの札を手続き描画するための自己記述子を返す。
//
// **切札はラベルに呼び名を出す。**Minchiate の切札は 40 枚あり、上位は星座・
// 四大元素・美徳といった固有の札。番号だけを出すと「35 と 36 のどちらが強いか」
// 以外の情報が画面から消える。同格札は無いので色は一律で構わない。
func minchiateFace(card *domain.Card) *CardFace {
	if card == nil {
		return nil
	}
	switch card.GetDesign() {
	case domain.MinchiateMattoDesign:
		return &CardFace{Glyph: "★", Label: "Matto", Color: "gold", Deck: "tarot"}
	case domain.MinchiateTrumpDesign:
		return &CardFace{
			Glyph: "✦",
			Label: domain.MinchiateTrumpName(card.GetValue()),
			Color: "purple",
			Deck:  "tarot",
		}
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
			Label: minchiateRankLabel(card.GetValue()),
			Color: colors[card.GetDesign()],
			Deck:  "tarot",
		}
	}
}

// minchiateRankLabel スート札のランクラベルを返す。
func minchiateRankLabel(value int) string {
	switch value {
	case 11:
		return "F" // Fante
	case 12:
		return "C" // Cavallo
	case 13:
		return "R" // Regina
	case domain.MinchiateSuitMax:
		return "Re"
	default:
		return strconv.Itoa(value)
	}
}

// MinchiateWebPresenter ミンキアーテのWebプレゼンタークラス
type MinchiateWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *MinchiateWebPresenter) Output(g interfaces.MinchiateGame, lastErr error) string {
	resObj := p.buildBase(g)
	resObj.Message, resObj.MessageCode, resObj.MessageParams = p.buildMessage(g, lastErr)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"` 専用の
	// レスポンスで、ページの state にはマージされない (#4483)。
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = cardHint(hint.CardIndices, hint.Reason)
	}
	return marshalOrError(resObj)
}

// buildBase 共通フィールドを構築
func (p *MinchiateWebPresenter) buildBase(g interfaces.MinchiateGame) *controller.MinchiateWebOutput {
	resObj := new(controller.MinchiateWebOutput)
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
	resObj.Config = controller.MinchiateWebOutputConfig{
		CpuDifficulty: int(cfg.CpuDifficulty),
		TargetRounds:  cfg.TargetRounds,
	}

	resObj.CurrentTrick = trickCardsToOutputWithFace(g.GetCurrentTrick(), minchiateFace)
	resObj.Players = p.buildPlayersOutput(g)
	return resObj
}

// playableIndices 人間プレイヤーがプレイできるカードのインデックスを返す
func (p *MinchiateWebPresenter) playableIndices(g interfaces.MinchiateGame) []int {
	if g.GetPhase() != domain.MinchiatePhasePlay || !g.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := g.GetPlayableIndices(g.GetCurrentPlayerIdx())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// buildPlayersOutput プレイヤー情報を構築
func (p *MinchiateWebPresenter) buildPlayersOutput(g interfaces.MinchiateGame) []*controller.MinchiateWebOutputPlayer {
	dealer := g.GetDealerIdx()
	out := make([]*controller.MinchiateWebOutputPlayer, 0)
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		out = append(out, &controller.MinchiateWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutputWithFace(player, player.GetIsHuman(), minchiateFace),
			TrickCount: player.GetTrickCount(),
			Team:       domain.MinchiateTeamOf(i),
			IsDealer:   i == dealer,
		})
	}
	return out
}

// buildMessage ゲーム結果メッセージを構築
func (p *MinchiateWebPresenter) buildMessage(g interfaces.MinchiateGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if g.GetGameEndFlag() {
		return p.winnerMessage(g)
	}
	switch g.GetPhase() {
	case domain.MinchiatePhaseScarto:
		return "", "minchiate.scartoPhase", nil
	case domain.MinchiatePhasePlay:
		if len(g.GetCurrentTrick()) == 0 {
			return "", "minchiate.playPhase.lead", nil
		}
		return "", "minchiate.playPhase.follow", nil
	case domain.MinchiatePhaseTrickEnd:
		return "", "minchiate.trickEnd", nil
	case domain.MinchiatePhaseRoundEnd:
		return "", "minchiate.roundEnd", nil
	}
	return "", "", nil
}

// winnerMessage 勝利チームのメッセージを構築する。
//
// **勝敗はチーム単位。**人間の席そのものではなく、人間が属するチームが勝ったかを見る。
func (p *MinchiateWebPresenter) winnerMessage(g interfaces.MinchiateGame) (string, string, map[string]string) {
	winner := g.GetWinnerTeam()
	if winner < 0 {
		return "ゲーム終了！ 引き分け。", "minchiate.result.draw", nil
	}
	humanTeam := -1
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if player := g.GetPlayer(i); player != nil && player.GetIsHuman() {
			humanTeam = domain.MinchiateTeamOf(i)
			break
		}
	}
	if humanTeam == winner {
		return "ゲーム終了！ あなたのチームの勝ち！", "minchiate.result.humanWin", nil
	}
	params := map[string]string{"team": fmt.Sprintf("%d", winner)}
	return fmt.Sprintf("ゲーム終了！ チーム%dの勝ち！", winner), "minchiate.result.cpuWin", params
}

// HintOutput ヒント情報をJSON出力する
func (p *MinchiateWebPresenter) HintOutput(g interfaces.MinchiateGame) string {
	hint := g.GetHint()
	resObj := p.buildBase(g)
	if hint != nil {
		resObj.Hint = cardHint(hint.CardIndices, hint.Reason)
		resObj.MessageCode = "minchiate.hintRequested"
	} else {
		resObj.MessageCode = "minchiate.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜をJSON出力
func (p *MinchiateWebPresenter) ActionLogOutput(g interfaces.MinchiateGame) string {
	return actionLogOutputJSON(g)
}
