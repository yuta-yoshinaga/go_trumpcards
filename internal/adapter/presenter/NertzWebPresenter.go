package presenter

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// NertzWebPresenter Nertz / Pounce Web プレゼンター
type NertzWebPresenter struct{}

// Output ゲーム状態を JSON 出力
func (p *NertzWebPresenter) Output(g interfaces.NertzGame, lastErr error) string {
	resObj := p.buildBase(g)
	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	// このゲームは手詰まり判定を持たないので、ゲートは進行中かどうかだけ。
	if g.GetPhase() == domain.NertzPhasePlaying {
		if hint := g.GetHint(); hint != nil {
			resObj.Hint = &controller.NertzWebHint{
				FromZone:  hint.FromZone,
				FromCol:   hint.FromCol,
				CardIndex: hint.CardIndex,
				ToZone:    hint.ToZone,
				ToCol:     hint.ToCol,
			}
		}
	}

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		switch g.GetPhase() {
		case domain.NertzPhasePlaying:
			resObj.MessageCode = "nertz.playing"
		case domain.NertzPhaseRoundEnd:
			resObj.Message = fmt.Sprintf("ラウンド終了 (勝者: P%d)", g.GetWinnerIdx())
			resObj.MessageCode = "nertz.roundEnd"
			resObj.MessageParams = map[string]string{"winner": fmt.Sprintf("%d", g.GetWinnerIdx())}
		case domain.NertzPhaseGameEnd:
			if g.GetMatchWinner() == 0 {
				resObj.Message = "あなたの勝ち！"
				resObj.MessageCode = "nertz.win"
			} else {
				resObj.Message = fmt.Sprintf("プレイヤー%dの勝ち", g.GetMatchWinner())
				resObj.MessageCode = "nertz.lose"
			}
			resObj.MessageParams = map[string]string{"winner": fmt.Sprintf("%d", g.GetMatchWinner())}
		}
	}
	return marshalOrError(resObj)
}

// HintOutput ヒントを JSON 出力
func (p *NertzWebPresenter) HintOutput(g interfaces.NertzGame) string {
	resObj := p.buildBase(g)
	if hint := g.GetHint(); hint != nil {
		resObj.Hint = &controller.NertzWebHint{
			FromZone:  hint.FromZone,
			FromCol:   hint.FromCol,
			CardIndex: hint.CardIndex,
			ToZone:    hint.ToZone,
			ToCol:     hint.ToCol,
		}
		resObj.MessageCode = "nertz.hintAvailable"
	} else {
		resObj.MessageCode = "nertz.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput 棋譜を JSON 出力。
// プレイ中はログを意図的に空にする — リアルタイム進行中にクライアントへ
// CPU の手数や狙いを露出させると不公平になるため、ラウンド/マッチ終了後にの
// み完全ログを返す (他のリアルタイム系ゲームと同じ運用 / PR #1528 レビュー指
// 摘)。
func (p *NertzWebPresenter) ActionLogOutput(g interfaces.NertzGame) string {
	return actionLogOutputJSON(g)
}

// buildBase は共通フィールドを詰めたレスポンスオブジェクトを返す
func (p *NertzWebPresenter) buildBase(g interfaces.NertzGame) *controller.NertzWebOutput {
	cfg := g.GetConfig()
	resObj := &controller.NertzWebOutput{
		Phase:         int(g.GetPhase()),
		RoundNumber:   g.GetRoundNo(),
		WinnerIdx:     g.GetWinnerIdx(),
		MatchWinner:   g.GetMatchWinner(),
		MoveCount:     g.GetMoveCount(),
		CanUndo:       g.CanUndo(),
		PlayerCount:   cfg.PlayerCount,
		DrawCount:     cfg.DrawCount,
		TargetScore:   cfg.TargetScore,
		CpuDifficulty: int(cfg.CpuDifficulty),
		CpuTickMoves:  cfg.ResolvedCpuTickMoves(),
		Players:       make([]*controller.NertzWebPlayer, 0),
		Foundations:   make([]*controller.NertzWebFoundation, 0),
	}

	for _, pl := range g.GetPlayers() {
		if pl == nil {
			continue
		}
		out := &controller.NertzWebPlayer{
			Name:      pl.GetName(),
			IsHuman:   !pl.GetIsCpu(),
			DeckIdx:   pl.GetDeckIdx(),
			Score:     pl.GetScore(),
			NertzSize: pl.NertzSize(),
			NertzTop:  cardToOutput(pl.NertzTop()),
			Tableau:   make([][]*controller.NertzWebTableauCard, domain.NertzTableauCnt),
			WasteTop:  cardToOutput(pl.WasteTop()),
			WasteSize: pl.WasteSize(),
			StockSize: pl.StockSize(),
		}
		for c := range domain.NertzTableauCnt {
			col := pl.GetTableauColumn(c)
			cards := make([]*controller.NertzWebTableauCard, 0, len(col))
			for _, tc := range col {
				cards = append(cards, &controller.NertzWebTableauCard{
					Card:   cardToOutput(tc.Card),
					FaceUp: tc.FaceUp,
				})
			}
			out.Tableau[c] = cards
		}
		resObj.Players = append(resObj.Players, out)
	}

	for _, f := range g.GetFoundations() {
		if f == nil {
			resObj.Foundations = append(resObj.Foundations, &controller.NertzWebFoundation{Suit: -1})
			continue
		}
		resObj.Foundations = append(resObj.Foundations, &controller.NertzWebFoundation{
			Top:  cardToOutput(f.Top()),
			Suit: f.Suit(),
			Size: f.Size(),
		})
	}

	return resObj
}
