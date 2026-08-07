//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// MississippiStudWebPresenter ミシシッピ・スタッドWebプレゼンター
type MississippiStudWebPresenter struct{}

// Output ゲーム状態を JSON で出力する。
func (mp *MississippiStudWebPresenter) Output(g interfaces.MississippiStudGame, lastErr error) string {
	resObj := new(controller.MississippiStudWebOutput)
	resObj.PlayerHand = cardsToOutputOrEmpty(g.GetPlayerHand())
	resObj.CommunityCards = mississippiStudMaskCommunity(g)
	resObj.CommunityRevealed = mississippiStudCommunityRevealedSlice(g)
	resObj.Phase = g.GetPhase()
	resObj.Chips = g.GetChips()
	resObj.AnteAmount = g.GetAnteAmount()
	resObj.StreetMultipliers = mississippiStudStreetMultipliersSlice(g)
	resObj.Folded = g.GetFolded()
	resObj.TotalBet = g.GetTotalBet()
	resObj.Result = int(g.GetResult())
	resObj.HandRank = g.GetHandRank()
	resObj.PayoutMultiplier = g.GetPayoutMultiplier()
	resObj.AntePayout = g.GetAntePayout()
	resObj.StreetPayouts = mississippiStudStreetPayoutsSlice(g)
	resObj.TotalPayout = g.GetTotalPayout()

	switch {
	case lastErr != nil:
		resObj.Message = lastErr.Error()
	case g.GetGameEndFlag():
		switch g.GetResult() {
		case domain.GameResultWin:
			resObj.Message = "Player wins!"
			resObj.MessageCode = "mississippistud.result.playerWins"
		case domain.GameResultLose:
			resObj.Message = "Player loses."
			resObj.MessageCode = "mississippistud.result.playerLoses"
		case domain.GameResultDraw:
			resObj.Message = "Push."
			resObj.MessageCode = "mississippistud.result.push"
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントを出力する。Web ではヒントはクライアント側 (useGameHint) で
// 算出するため、通常の状態出力を返す。MississippiStudPresenter インタフェースを
// 満たすための実装。
func (mp *MississippiStudWebPresenter) HintOutput(g interfaces.MississippiStudGame) string {
	return mp.Output(g, nil)
}

// ActionLogOutput 棋譜を JSON 出力する。
func (mp *MississippiStudWebPresenter) ActionLogOutput(g interfaces.MississippiStudGame) string {
	return actionLogOutputJSON(g)
}

// mississippiStudMaskCommunity は公開状態に応じてコミュニティカードをマスクして返す。
// ゲーム終了時は全公開。
func mississippiStudMaskCommunity(g interfaces.MississippiStudGame) []*controller.WebOutputCard {
	cards := g.GetCommunityCards()
	if len(cards) == 0 {
		return make([]*controller.WebOutputCard, 0)
	}
	revealed := g.GetCommunityRevealed()
	result := make([]*controller.WebOutputCard, len(cards))
	for i, c := range cards {
		if i < len(revealed) && revealed[i] {
			result[i] = cardToOutput(c)
		} else {
			result[i] = &controller.WebOutputCard{Design: "", Value: 0}
		}
	}
	return result
}

func mississippiStudCommunityRevealedSlice(g interfaces.MississippiStudGame) []bool {
	r := g.GetCommunityRevealed()
	out := make([]bool, len(r))
	copy(out, r[:])
	return out
}

func mississippiStudStreetMultipliersSlice(g interfaces.MississippiStudGame) []int {
	m := g.GetStreetMultipliers()
	out := make([]int, len(m))
	copy(out, m[:])
	return out
}

func mississippiStudStreetPayoutsSlice(g interfaces.MississippiStudGame) []int {
	p := g.GetStreetPayouts()
	out := make([]int, len(p))
	copy(out, p[:])
	return out
}
