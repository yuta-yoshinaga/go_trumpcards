//go:build !js || !wasm || extra4

package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// BarbuWebPresenter はバルブ Web プレゼンタークラス。
type BarbuWebPresenter struct{}

// Output はゲーム状態を JSON 出力する。
func (bwp *BarbuWebPresenter) Output(bg interfaces.BarbuGame, lastErr error) string {
	resObj := newBarbuOutput(bg)

	resObj.CurrentTrick = barbuTrickToOutput(bg.GetCurrentTrick())
	resObj.LastTrick = barbuTrickToOutput(bg.GetLastTrick())
	resObj.TablePlaced = barbuTableToOutput(bg.GetTablePlaced())
	resObj.UsedContracts = barbuUsedToOutput(bg.GetUsedContracts(bg.GetDealerIdx()))
	if bg.GetCurrentContract() == domain.BarbuContractDominoes && bg.IsHumanTurn() {
		resObj.DominoPlayable = append(resObj.DominoPlayable, bg.GetDominoPlayableIndices(bg.GetCurrentTurn())...)
	} else if bg.IsHumanTurn() {
		// **フォロー義務の可視化。**ドミノ以外の 6 契約では、リード色を持っていても
		// 全カードが同じように押せて、弾かれて初めて分かる状態だった (#4804)。
		resObj.PlayableIndices = append(resObj.PlayableIndices, bg.GetPlayableIndices(bg.GetCurrentTurn())...)
	}

	for i := 0; i < bg.GetPlayerCnt(); i++ {
		player := bg.GetPlayer(i)
		if player == nil {
			continue
		}
		resObj.Players = append(resObj.Players, &controller.BarbuWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			TrickCount: player.GetTrickCount(),
			DominoRank: player.GetDominoRank(),
			TotalScore: player.GetTotalScore(),
		})
	}

	if det := bg.GetLastDealDetail(); det != nil {
		resObj.LastDealDetail = barbuDealDetailToOutput(det)
	}
	for _, det := range bg.GetDealHistory() {
		if det == nil {
			continue
		}
		resObj.DealHistory = append(resObj.DealHistory, barbuDealDetailToOutput(det))
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if bg.GetGameEndFlag() {
		resObj.Message = bwp.buildResultMessage(bg)
		resObj.MessageCode = "barbu.result.scores"
		resObj.MessageParams = map[string]string{
			"scores": bwp.encodeScoresParam(bg),
		}
	}

	return marshalOrError(resObj)
}

// newBarbuOutput は基本フィールドを埋めた出力オブジェクトを生成する。
func newBarbuOutput(bg interfaces.BarbuGame) *controller.BarbuWebOutput {
	resObj := new(controller.BarbuWebOutput)
	resObj.Players = make([]*controller.BarbuWebOutputPlayer, 0)
	resObj.CurrentTrick = make([]*controller.WebOutputTrickCard, 0)
	resObj.LastTrick = make([]*controller.WebOutputTrickCard, 0)
	resObj.TablePlaced = make([]int, 0)
	resObj.DominoPlayable = make([]int, 0)
	resObj.UsedContracts = make([]bool, 0)
	resObj.RoundWinners = make([]int, 0)
	resObj.DealHistory = make([]*controller.BarbuWebOutputDealDetail, 0)

	resObj.Phase = bg.GetPhase()
	resObj.DealNumber = bg.GetDealNumber()
	resObj.TotalDeals = domain.BarbuTotalDeals
	resObj.DealerIdx = bg.GetDealerIdx()
	resObj.CurrentTurn = bg.GetCurrentTurn()
	resObj.CurrentContract = bg.GetCurrentContract()
	resObj.TrumpSuit = bg.GetTrumpSuit()
	resObj.TrickNumber = bg.GetTrickNumber()
	resObj.LastTrickWinner = bg.GetLastTrickWinner()
	resObj.GameEndFlag = bg.GetGameEndFlag()
	resObj.RoundWinners = append(resObj.RoundWinners, bg.GetRoundWinners()...)

	config := bg.GetConfig()
	resObj.Config = controller.BarbuWebConfig{CpuDifficulty: int(config.CpuDifficulty)}
	return resObj
}

// barbuTrickToOutput はトリックを WebOutput 表現に変換する。
func barbuTrickToOutput(trick []*domain.TrickCard) []*controller.WebOutputTrickCard {
	out := make([]*controller.WebOutputTrickCard, 0, len(trick))
	for _, tc := range trick {
		if tc == nil {
			continue
		}
		out = append(out, &controller.WebOutputTrickCard{
			PlayerIdx: tc.PlayerIdx,
			Card:      cardToOutput(tc.Card),
		})
	}
	return out
}

// barbuTableToOutput は Dominoes の場 (bitmask 配列) を []int に変換する。
func barbuTableToOutput(table [5]uint16) []int {
	out := make([]int, len(table))
	for i, v := range table {
		out[i] = int(v)
	}
	return out
}

// barbuDealDetailToOutput は 1 ディールの得点内訳を WebOutput 表現に変換する。
func barbuDealDetailToOutput(det *domain.BarbuDealDetail) *controller.BarbuWebOutputDealDetail {
	return &controller.BarbuWebOutputDealDetail{
		Contract:  det.Contract,
		TrumpSuit: det.TrumpSuit,
		DealerIdx: det.DealerIdx,
		Gained:    det.Gained,
	}
}

// barbuUsedToOutput は使用済みコントラクト配列を []bool に変換する。
func barbuUsedToOutput(used [domain.BarbuContractCnt]bool) []bool {
	out := make([]bool, len(used))
	copy(out, used[:])
	return out
}

// encodeScoresParam は最終スコアを "0:12,1:-3" 形式の文字列に詰める。
func (bwp *BarbuWebPresenter) encodeScoresParam(bg interfaces.BarbuGame) string {
	parts := make([]string, 0, bg.GetPlayerCnt())
	for i := 0; i < bg.GetPlayerCnt(); i++ {
		p := bg.GetPlayer(i)
		if p == nil {
			continue
		}
		parts = append(parts, fmt.Sprintf("%d:%d", i, p.GetTotalScore()))
	}
	return strings.Join(parts, ",")
}

// buildResultMessage はゲーム終了時のフォールバック (英語) メッセージ。
func (bwp *BarbuWebPresenter) buildResultMessage(bg interfaces.BarbuGame) string {
	msg := "Game over. "
	for i := 0; i < bg.GetPlayerCnt(); i++ {
		p := bg.GetPlayer(i)
		if p == nil {
			continue
		}
		name := fmt.Sprintf("CPU %d", i)
		if p.GetIsHuman() {
			name = "You"
		}
		msg += fmt.Sprintf("%s:%dpt ", name, p.GetTotalScore())
	}
	return msg
}

// ActionLogOutput は棋譜を JSON 出力する。
// HintOutput returns the current state as JSON. The Web GUI computes its own
// hint client-side, so this mirrors Output to satisfy the BarbuPresenter
// interface shared with the CUI.
func (bwp *BarbuWebPresenter) HintOutput(bg interfaces.BarbuGame) string {
	return bwp.Output(bg, nil)
}

func (bwp *BarbuWebPresenter) ActionLogOutput(bg interfaces.BarbuGame) string {
	return actionLogOutputJSON(bg)
}
