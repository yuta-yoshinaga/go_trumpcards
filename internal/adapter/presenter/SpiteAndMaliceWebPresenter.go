package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SpiteAndMaliceWebPresenter Spite & Malice Web プレゼンター
type SpiteAndMaliceWebPresenter struct{}

// Output ゲーム状態を JSON 出力
func (p *SpiteAndMaliceWebPresenter) Output(g interfaces.SpiteAndMaliceGame, lastErr error) string {
	resObj := p.buildBase(g)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	// このゲームは手詰まり判定を持たないので、ゲートは進行中かどうかだけ。
	if g.GetPhase() == domain.SpiteAndMalicePhasePlaying {
		if hint := g.GetHint(); hint != nil {
			resObj.Hint = &controller.SpiteAndMaliceWebHint{
				Source:        sourceToString(hint.Source),
				Index:         hint.Index,
				FoundationIdx: hint.FoundationIdx,
				Discard:       hint.Discard,
			}
		}
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch g.GetPhase() {
		case domain.SpiteAndMalicePhasePlaying:
			resObj.MessageCode = "spiteandmalice.playing"
		case domain.SpiteAndMalicePhaseGameOver:
			if g.GetWinner() == domain.SpiteAndMaliceHumanIdx {
				resObj.Message = fmt.Sprintf("あなたの勝ち！ 手数: %d", g.GetMoveCount())
				resObj.MessageCode = "spiteandmalice.win"
			} else {
				resObj.Message = fmt.Sprintf("CPU の勝ち 手数: %d", g.GetMoveCount())
				resObj.MessageCode = "spiteandmalice.lose"
			}
			resObj.MessageParams = map[string]string{"moveCount": fmt.Sprintf("%d", g.GetMoveCount())}
		}
	}
	return marshalOrError(resObj)
}

// HintOutput ヒントを JSON 出力
func (p *SpiteAndMaliceWebPresenter) HintOutput(g interfaces.SpiteAndMaliceGame) string {
	resObj := p.buildBase(g)
	hint := g.GetHint()
	if hint != nil {
		resObj.Hint = &controller.SpiteAndMaliceWebHint{
			Source:        sourceToString(hint.Source),
			Index:         hint.Index,
			FoundationIdx: hint.FoundationIdx,
			Discard:       hint.Discard,
		}
		resObj.MessageCode = "spiteandmalice.hintAvailable"
	} else {
		resObj.MessageCode = "spiteandmalice.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜を JSON 出力
func (p *SpiteAndMaliceWebPresenter) ActionLogOutput(g interfaces.SpiteAndMaliceGame) string {
	return actionLogOutputJSON(g)
}

// buildBase は共通フィールドを詰めたレスポンスオブジェクトを返す
func (p *SpiteAndMaliceWebPresenter) buildBase(g interfaces.SpiteAndMaliceGame) *controller.SpiteAndMaliceWebOutput {
	resObj := new(controller.SpiteAndMaliceWebOutput)
	resObj.Phase = int(g.GetPhase())
	resObj.Current = g.GetCurrent()
	resObj.MoveCount = g.GetMoveCount()
	resObj.Winner = g.GetWinner()
	resObj.StockSize = g.GetStockSize()
	resObj.CompletedSize = g.GetCompletedSize()
	cfg := g.GetConfig()
	resObj.GoalSize = cfg.GoalSize
	resObj.CpuDifficulty = int(cfg.CpuDifficulty)
	resObj.CanAutoComplete = g.CanAutoComplete()

	foundations := g.GetFoundations()
	for i := range domain.SpiteAndMaliceFoundationCnt {
		pile := foundations[i]
		out := make([]*controller.WebOutputCard, len(pile))
		for j, card := range pile {
			out[j] = cardToOutput(card)
		}
		resObj.Foundations[i] = out
		resObj.FoundationTops[i] = g.GetFoundationTopValue(i)
	}

	for i := range domain.SpiteAndMalicePlayerCnt {
		pl := g.GetPlayer(i)
		var player controller.SpiteAndMaliceWebPlayer
		if pl != nil {
			player.IsCpu = pl.GetIsCpu()
			hand := pl.GetHand()
			// 人間プレイヤーの手札のみ公開。CPU の手札は枚数だけ返す。
			player.Hand = make([]*controller.WebOutputCard, len(hand))
			if !pl.GetIsCpu() {
				for k, c := range hand {
					player.Hand[k] = cardToOutput(c)
				}
			}
			if top := pl.GoalTop(); top != nil {
				player.GoalTop = cardToOutput(top)
			}
			player.GoalSize = pl.GoalSize()
			for s := range domain.SpiteAndMaliceSideCnt {
				side := pl.GetSide(s)
				out := make([]*controller.WebOutputCard, len(side))
				for k, c := range side {
					out[k] = cardToOutput(c)
				}
				player.Sides[s] = out
			}
		} else {
			for s := range domain.SpiteAndMaliceSideCnt {
				player.Sides[s] = make([]*controller.WebOutputCard, 0)
			}
		}
		resObj.Players[i] = player
	}

	return resObj
}

// sourceToString ヒントソースを JSON 文字列へ
func sourceToString(src domain.SpiteAndMaliceSource) string {
	switch src {
	case domain.SpiteAndMaliceSourceGoal:
		return "goal"
	case domain.SpiteAndMaliceSourceHand:
		return "hand"
	case domain.SpiteAndMaliceSourceSide:
		return "side"
	}
	return ""
}
