//go:build !js || !wasm || extra

package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// KingWebPresenter はキング Web プレゼンタークラス。
type KingWebPresenter struct{}

// Output はゲーム状態を JSON 出力する。
func (kwp *KingWebPresenter) Output(kg interfaces.KingGame, lastErr error) string {
	resObj := kwp.buildBase(kg)

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if kg.GetGameEndFlag() {
		resObj.Message = kwp.buildResultMessage(kg)
		resObj.MessageCode = "king.result.scores"
		resObj.MessageParams = map[string]string{
			"scores": kwp.encodeScoresParam(kg),
		}
	}

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**King.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := kg.GetHint(); hint != nil {
		resObj.Hint = &controller.WebOutputCardHint{
			CardIndices: hint.CardIndices,
			Reason:      hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// buildBase は基本フィールドを埋めた出力オブジェクトを生成する。
func (kwp *KingWebPresenter) buildBase(kg interfaces.KingGame) *controller.KingWebOutput {
	resObj := new(controller.KingWebOutput)
	resObj.Players = make([]*controller.KingWebOutputPlayer, 0)
	resObj.CurrentTrick = make([]*controller.WebOutputTrickCard, 0)
	resObj.LastTrick = make([]*controller.WebOutputTrickCard, 0)
	resObj.UsedContracts = make([]bool, 0)
	resObj.PlayableIndices = make([]int, 0)
	resObj.RoundWinners = make([]int, 0)

	resObj.Phase = kg.GetPhase()
	resObj.DealNumber = kg.GetDealNumber()
	resObj.TotalDeals = domain.KingTotalDeals
	resObj.DealerIdx = kg.GetDealerIdx()
	resObj.CurrentTurn = kg.GetCurrentTurn()
	resObj.CurrentContract = kg.GetCurrentContract()
	resObj.TrumpSuit = kg.GetTrumpSuit()
	resObj.TrickNumber = kg.GetTrickNumber()
	resObj.LastTrickWinner = kg.GetLastTrickWinner()
	resObj.GameEndFlag = kg.GetGameEndFlag()
	resObj.IsHumanTurn = kg.IsHumanTurn()
	resObj.RoundWinners = append(resObj.RoundWinners, kg.GetRoundWinners()...)

	config := kg.GetConfig()
	resObj.Config = controller.KingWebConfig{CpuDifficulty: int(config.CpuDifficulty)}

	resObj.CurrentTrick = kingTrickToOutput(kg.GetCurrentTrick())
	resObj.LastTrick = kingTrickToOutput(kg.GetLastTrick())
	resObj.UsedContracts = kingUsedToOutput(kg.GetUsedContracts())
	resObj.PlayableIndices = kwp.playableIndices(kg)

	for i := 0; i < kg.GetPlayerCnt(); i++ {
		player := kg.GetPlayer(i)
		if player == nil {
			continue
		}
		resObj.Players = append(resObj.Players, &controller.KingWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
			TrickCount: player.GetTrickCount(),
			TotalScore: player.GetTotalScore(),
		})
	}

	if det := kg.GetLastDealDetail(); det != nil {
		resObj.LastDealDetail = &controller.KingWebOutputDealDetail{
			Contract:  det.Contract,
			TrumpSuit: det.TrumpSuit,
			DealerIdx: det.DealerIdx,
			Gained:    det.Gained,
		}
	}
	return resObj
}

// playableIndices は人間プレイヤーがプレイできるカードのインデックスを返す。
func (kwp *KingWebPresenter) playableIndices(kg interfaces.KingGame) []int {
	if kg.GetPhase() != domain.KingPhasePlay || !kg.IsHumanTurn() {
		return make([]int, 0)
	}
	idx := kg.GetPlayableIndices(kg.GetCurrentTurn())
	if idx == nil {
		return make([]int, 0)
	}
	return idx
}

// kingTrickToOutput はトリックを WebOutput 表現に変換する。
func kingTrickToOutput(trick []*domain.TrickCard) []*controller.WebOutputTrickCard {
	return buildTrickCards(trick, func(tc *domain.TrickCard) *controller.WebOutputTrickCard {
		if tc == nil {
			return nil
		}
		return &controller.WebOutputTrickCard{PlayerIdx: tc.PlayerIdx, Card: cardToOutput(tc.Card)}
	})
}

// kingUsedToOutput は使用済みコントラクト配列を []bool に変換する。
func kingUsedToOutput(used [domain.KingContractCnt]bool) []bool {
	out := make([]bool, len(used))
	copy(out, used[:])
	return out
}

// encodeScoresParam は最終スコアを "0:12,1:-3" 形式の文字列に詰める。
func (kwp *KingWebPresenter) encodeScoresParam(kg interfaces.KingGame) string {
	parts := make([]string, 0, kg.GetPlayerCnt())
	for i := 0; i < kg.GetPlayerCnt(); i++ {
		p := kg.GetPlayer(i)
		if p == nil {
			continue
		}
		parts = append(parts, fmt.Sprintf("%d:%d", i, p.GetTotalScore()))
	}
	return strings.Join(parts, ",")
}

// buildResultMessage はゲーム終了時のフォールバック (英語) メッセージ。
func (kwp *KingWebPresenter) buildResultMessage(kg interfaces.KingGame) string {
	msg := "Game over. "
	for i := 0; i < kg.GetPlayerCnt(); i++ {
		p := kg.GetPlayer(i)
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

// HintOutput はヒント情報を JSON 出力する。
func (kwp *KingWebPresenter) HintOutput(kg interfaces.KingGame) string {
	resObj := kwp.buildBase(kg)
	if hint := kg.GetHint(); hint != nil {
		resObj.Hint = &controller.WebOutputCardHint{
			CardIndices: hint.CardIndices,
			Reason:      hint.Reason,
		}
	}
	// **「頼んだヒントか」を CLI が見分けられるようにする。**このゲーム群の
	// `hintAvailable` は画面のラベルとして既に使われているので、別キーを出す (#4483)。
	if kg.GetHint() != nil {
		resObj.MessageCode = "king.hintRequested"
	} else {
		resObj.MessageCode = "king.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput は棋譜を JSON 出力する。
func (kwp *KingWebPresenter) ActionLogOutput(kg interfaces.KingGame) string {
	return actionLogOutputJSON(kg)
}
