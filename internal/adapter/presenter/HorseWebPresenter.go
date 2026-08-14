//go:build !js || !wasm || casino

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// HorseWebPresenter は H.O.R.S.E. の Web プレゼンター。
type HorseWebPresenter struct{}

// Output はゲーム状態を JSON 出力する。
func (p *HorseWebPresenter) Output(g interfaces.HorseGame, lastErr error) string {
	resObj := p.buildBase(g)
	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else if g.GetGameEndFlag() {
		resObj.MessageCode = "horse.result.winner"
		resObj.MessageParams = map[string]string{"name": g.GetSeatName(g.WinnerSeat())}
		resObj.Message = "Game over. Winner: " + g.GetSeatName(g.WinnerSeat())
	}
	return marshalOrError(resObj)
}

// buildBase は基本フィールドを埋めた出力オブジェクトを生成する。
func (p *HorseWebPresenter) buildBase(g interfaces.HorseGame) *controller.HorseWebOutput {
	resObj := new(controller.HorseWebOutput)
	cfg := g.GetConfig()
	resObj.Seats = make([]*controller.HorseWebOutputSeat, 0, g.GetSeatCount())
	for i := 0; i < g.GetSeatCount(); i++ {
		resObj.Seats = append(resObj.Seats, &controller.HorseWebOutputSeat{
			ID:      i,
			Name:    g.GetSeatName(i),
			IsHuman: g.GetSeatIsHuman(i),
			Chips:   g.GetSeatLiveChips(i),
			Cards:   cardsToOutputOrEmpty(g.GetSeatCards(i)),
		})
	}
	resObj.Phase = int(g.GetPhase())
	resObj.Discipline = int(g.GetDiscipline())
	resObj.DisciplineLetter = g.GetDisciplineLetter()
	resObj.DisciplineName = domain.HorseDisciplineName(g.GetDiscipline())
	resObj.HandInDiscipline = g.GetHandInDiscipline()
	resObj.HandNumber = g.GetHandNumber()
	resObj.CurrentTurn = g.GetCurrentTurn()
	resObj.HumanSeat = g.GetHumanSeat()
	resObj.IsHumanTurn = g.IsHumanTurn()
	resObj.CommunityCards = cardsToOutputOrEmpty(g.GetCommunityCards())
	resObj.Pot = g.GetPot()
	resObj.ToCall = g.GetToCall()
	resObj.MinRaise = g.GetMinRaise()
	resObj.TablePhase = g.GetTablePhase()
	resObj.GameEndFlag = g.GetGameEndFlag()
	resObj.WinnerSeat = -1
	if g.GetGameEndFlag() {
		resObj.WinnerSeat = g.WinnerSeat()
	}
	resObj.Config = controller.HorseWebOutputConfig{
		Seats:              cfg.Seats,
		InitialChips:       cfg.InitialChips,
		HandsPerDiscipline: cfg.HandsPerDiscipline,
	}
	return resObj
}

// HintOutput はヒント情報を JSON 出力する。
func (p *HorseWebPresenter) HintOutput(g interfaces.HorseGame) string {
	return p.Output(g, nil)
}

// ActionLogOutput は棋譜を JSON 出力する。
func (p *HorseWebPresenter) ActionLogOutput(g interfaces.HorseGame) string {
	return actionLogOutputJSON(g)
}
