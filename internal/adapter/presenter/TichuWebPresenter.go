//go:build !js || !wasm || extra2

package presenter

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// TichuWebPresenter ティチューWebプレゼンタークラス
type TichuWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *TichuWebPresenter) Output(tg interfaces.TichuGame, lastErr error) string {
	resObj := new(controller.TichuWebOutput)

	resObj.Phase = tichuPhaseName(tg.GetPhase())
	resObj.CurrentTurn = tg.GetCurrentTurn()
	resObj.LastPlayIdx = tg.GetLastPlayIdx()
	resObj.StartLeader = tg.GetStartLeader()
	resObj.FinishOrder = tg.GetFinishOrder()
	if resObj.FinishOrder == nil {
		resObj.FinishOrder = make([]int, 0)
	}
	resObj.Scores = tg.GetScores()
	resObj.IsOneTwo = tg.GetIsOneTwo()
	resObj.BombCount = tg.GetBombCount()
	resObj.GameEndFlag = tg.GetGameEndFlag()

	config := tg.GetConfig()
	resObj.Config = controller.TichuWebConfig{
		CpuDifficulty: int(config.CpuDifficulty),
	}

	if combo := tg.GetTableCombo(); combo != nil {
		resObj.TableCards = cardsToOutputOrEmpty(combo.Cards)
		resObj.TableCombo = tichuComboName(combo.Type)
	} else {
		resObj.TableCards = make([]*controller.WebOutputCard, 0)
	}

	resObj.CpuActions = make([]*controller.TichuWebOutputAction, 0)
	for _, action := range tg.GetCpuActions() {
		resObj.CpuActions = append(resObj.CpuActions, tichuActionToOutput(action))
	}
	if humanAction := tg.GetHumanAction(); humanAction != nil {
		resObj.HumanAction = tichuActionToOutput(humanAction)
	}

	resObj.Players = make([]*controller.TichuWebOutputPlayer, 0)
	for i := 0; i < tg.GetPlayerCnt(); i++ {
		player := tg.GetPlayer(i)
		if player == nil {
			continue
		}
		resObj.Players = append(resObj.Players, &controller.TichuWebOutputPlayer{
			ID:         i,
			IsHuman:    player.GetIsHuman(),
			IsFinished: player.GetIsFinished(),
			Team:       domain.TichuTeamOf(i),
			Rank:       player.GetRank(),
			DeclType:   player.GetDeclType(),
			CardCount:  player.GetCardsSize(),
			Cards:      playerCardsToOutput(player, player.GetIsHuman()),
		})
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if tg.GetGameEndFlag() {
		resObj.Message = p.buildResultMessage(tg)
		resObj.MessageCode = "tichu.result.summary"
		resObj.MessageParams = map[string]string{"summary": resObj.Message}
	} else if tg.GetDogLeadPassed() {
		p.setDogLeadMessage(resObj, tg)
	}

	return marshalOrError(resObj)
}

// setDogLeadMessage fills in the "the dog moved the lead" notice.
//
// **席の名前は Go 側で組まない。**`MessageCode` + `MessageParams` はフロント側が
// 自分のロケールで組み直すためのもので、ここで「あなた」を埋めると英語の文の中に
// 日本語が残る。誰が人間かで文そのものが変わるので、席番号だけを渡して
// コードを 3 つに分ける (CallBreak の `cpuId` と同じ作り)。
func (p *TichuWebPresenter) setDogLeadMessage(resObj *controller.TichuWebOutput, tg interfaces.TichuGame) {
	fromIdx := tg.GetDogLeadFrom()
	toIdx := (fromIdx + 2) % tg.GetPlayerCnt()
	from, to := strconv.Itoa(fromIdx), strconv.Itoa(toIdx)

	var args []string
	switch {
	case isTichuHuman(tg, fromIdx):
		resObj.MessageCode = "tichu.dogLeadPassedByYou"
		args = []string{"to", to}
	case isTichuHuman(tg, toIdx):
		resObj.MessageCode = "tichu.dogLeadPassedToYou"
		args = []string{"from", from}
	default:
		resObj.MessageCode = "tichu.dogLeadPassedBetweenCpus"
		args = []string{"from", from, "to", to}
	}

	resObj.MessageParams = make(map[string]string, len(args)/2)
	for i := 0; i < len(args); i += 2 {
		resObj.MessageParams[args[i]] = args[i+1]
	}
	resObj.Message = i18n.Tf(resObj.MessageCode, args...)
}

// isTichuHuman reports whether the seat at idx is the human player.
func isTichuHuman(tg interfaces.TichuGame, idx int) bool {
	player := tg.GetPlayer(idx)
	return player != nil && player.GetIsHuman()
}

// buildResultMessage ディール終了メッセージを生成
func (p *TichuWebPresenter) buildResultMessage(tg interfaces.TichuGame) string {
	scores := tg.GetScores()
	var winner string
	switch {
	case scores[0] > scores[1]:
		winner = "Team A (P0/P2)"
	case scores[1] > scores[0]:
		winner = "Team B (P1/P3)"
	default:
		winner = "Draw"
	}
	return fmt.Sprintf("%s — A:%d B:%d", winner, scores[0], scores[1])
}

// ActionLogOutput 棋譜をJSON出力
func (p *TichuWebPresenter) ActionLogOutput(tg interfaces.TichuGame) string {
	return actionLogOutputJSON(tg)
}

func tichuActionToOutput(a *domain.TichuCpuAction) *controller.TichuWebOutputAction {
	return &controller.TichuWebOutputAction{
		PlayerIdx:   a.PlayerIdx,
		PlayedCards: cardsToOutput(a.PlayedCards),
		DeclType:    a.DeclType,
		IsPass:      a.IsPass,
	}
}

func tichuPhaseName(phase domain.TichuPhase) string {
	switch phase {
	case domain.TichuPhaseDeclare:
		return "declare"
	case domain.TichuPhasePlay:
		return "play"
	case domain.TichuPhaseEnd:
		return "end"
	default:
		return "unknown"
	}
}

func tichuComboName(t domain.TichuComboType) string {
	switch t {
	case domain.TichuComboSingle:
		return "single"
	case domain.TichuComboPair:
		return "pair"
	case domain.TichuComboTriple:
		return "triple"
	case domain.TichuComboFullHouse:
		return "fullHouse"
	case domain.TichuComboStraight:
		return "straight"
	case domain.TichuComboStairs:
		return "stairs"
	case domain.TichuComboBomb:
		return "bomb"
	case domain.TichuComboStraightFlush:
		return "straightFlush"
	case domain.TichuComboDog:
		return "dog"
	default:
		return "none"
	}
}
