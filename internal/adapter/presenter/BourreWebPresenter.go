//go:build !js || !wasm || casino

package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BourreWebPresenter ブーレWebプレゼンタークラス
type BourreWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *BourreWebPresenter) Output(bg interfaces.BourreGame, lastErr error) string {
	resObj := new(controller.BourreWebOutput)

	resObj.Phase = bourrePhaseName(bg.GetPhase())
	resObj.CurrentPlayerIdx = bg.GetCurrentPlayerIdx()
	resObj.DealerIdx = bg.GetDealerIdx()
	resObj.Pot = bg.GetPot()
	resObj.CarryPot = bg.GetCarryPot()
	resObj.TrumpSuit = cardDesignToString(bg.GetTrumpSuit())
	resObj.TrumpCard = cardToOutput(bg.GetTrumpCard())
	resObj.TrickNumber = bg.GetTrickNumber()
	resObj.LastTrickWinner = bg.GetLastTrickWinner()
	resObj.LeadPlayerIdx = bg.GetLeadPlayerIdx()
	resObj.HandNumber = bg.GetHandNumber()
	resObj.GameEndFlag = bg.GetGameEndFlag()
	resObj.WinnerIdx = bg.GetWinnerIdx()

	config := bg.GetConfig()
	resObj.Config = controller.BourreWebConfig{CpuDifficulty: int(config.CpuDifficulty)}

	resObj.CurrentTrick = bourreTrickToOutput(bg.GetCurrentTrick())
	resObj.LastTrick = bourreTrickToOutput(bg.GetLastTrick())

	humanIdx := bourreHumanIdx(bg)
	resObj.ValidPlays = bg.GetValidPlayIndices(humanIdx)
	if resObj.ValidPlays == nil {
		resObj.ValidPlays = make([]int, 0)
	}

	resObj.Results = make([]*controller.BourreWebResult, 0)
	for _, r := range bg.GetLastResults() {
		resObj.Results = append(resObj.Results, &controller.BourreWebResult{
			PlayerIdx: r.PlayerIdx,
			Tricks:    r.Tricks,
			WonAmount: r.WonAmount,
			Bourreed:  r.Bourreed,
			Folded:    r.Folded,
		})
	}

	resObj.Players = make([]*controller.BourreWebOutputPlayer, 0)
	for i := 0; i < bg.GetPlayerCnt(); i++ {
		player := bg.GetPlayer(i)
		if player == nil {
			continue
		}
		resObj.Players = append(resObj.Players, &controller.BourreWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			IsFinished: player.GetIsFinished(),
			Folded:     player.GetFolded(),
			Decided:    player.GetDecided(),
			Drawn:      player.GetDrawn(),
			Bourreed:   player.GetBourreed(),
			Chips:      player.GetChips(),
			Tricks:     player.GetTrickCount(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
		})
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if bg.GetGameEndFlag() {
		resObj.Message = p.buildResultMessage(bg)
	}

	return marshalOrError(resObj)
}

// buildResultMessage ゲーム終了メッセージを生成
func (p *BourreWebPresenter) buildResultMessage(bg interfaces.BourreGame) string {
	idx := bg.GetWinnerIdx()
	winner := "You"
	if player := bg.GetPlayer(idx); player != nil && !player.GetIsHuman() {
		winner = fmt.Sprintf("CPU %d", idx)
	}
	chips := 0
	if player := bg.GetPlayer(idx); player != nil {
		chips = player.GetChips()
	}
	return fmt.Sprintf("%s wins with %d chips!", winner, chips)
}

// ActionLogOutput 棋譜をJSON出力
func (p *BourreWebPresenter) ActionLogOutput(bg interfaces.BourreGame) string {
	return actionLogOutputJSON(bg)
}

// bourreHumanIdx 人間プレイヤーのインデックスを返す (見つからなければ0)
func bourreHumanIdx(bg interfaces.BourreGame) int {
	for i := 0; i < bg.GetPlayerCnt(); i++ {
		if player := bg.GetPlayer(i); player != nil && player.GetIsHuman() {
			return i
		}
	}
	return 0
}

// bourreTrickToOutput トリックを WebOutput 形式に変換する
func bourreTrickToOutput(trick []*domain.TrickCard) []*controller.BourreWebTrickCard {
	out := make([]*controller.BourreWebTrickCard, 0, len(trick))
	for _, tc := range trick {
		if tc == nil {
			continue
		}
		out = append(out, &controller.BourreWebTrickCard{
			PlayerIdx: tc.PlayerIdx,
			Card:      cardToOutput(tc.Card),
		})
	}
	return out
}

func bourrePhaseName(phase domain.BourrePhase) string {
	switch phase {
	case domain.BourrePhaseDecide:
		return "decide"
	case domain.BourrePhaseDraw:
		return "draw"
	case domain.BourrePhasePlay:
		return "play"
	case domain.BourrePhaseRoundEnd:
		return "roundEnd"
	case domain.BourrePhaseGameEnd:
		return "gameEnd"
	default:
		return "unknown"
	}
}
