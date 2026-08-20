//go:build !js || !wasm || extra4

package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PiquetWebPresenter Piquet Webプレゼンター
type PiquetWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (p *PiquetWebPresenter) Output(g interfaces.PiquetGame, lastErr error) string {
	resObj := newPiquetWebOutputBase(g)

	if lastErr != nil {
		resObj.Message = lastErr.Error()
	} else {
		resObj.MessageCode = piquetPhaseMessageCode(g.GetPhase())
	}

	// 合法プレイインデックス (人間ターンのみ)
	if g.GetPhase() == domain.PiquetPhasePlay && g.IsHumanTurn() {
		resObj.LegalPlayIndices = g.GetLegalPlayIndices(g.GetCurrentPlayerIdx())
	}

	// **受動ヒントは Output() でも埋める。**HintOutput() は `command: "hint"`
	// 専用のレスポンスで、ページの state にはマージされない。ここで埋めないと
	// フロントの `state.hint` は常に undefined で、それを読む分岐は全部死ぬ (#4483)。
	//
	// **フェーズと手番はここでは見ない。**Piquet.GetHint() が自分で
	// 「人間の手番で、かつ行動を選べる状態か」を確かめて nil を返す。
	if hint := g.GetHint(g.GetCurrentPlayerIdx()); hint != nil {
		resObj.Hint = &controller.PiquetWebOutputHint{
			CardIndex:      hint.CardIndex,
			DiscardIndices: hint.DiscardIndices,
			Reason:         hint.Reason,
		}
	}

	return marshalOrError(resObj)
}

// HintOutput ヒントをJSON出力
func (p *PiquetWebPresenter) HintOutput(g interfaces.PiquetGame) string {
	resObj := newPiquetWebOutputBase(g)
	hint := g.GetHint(g.GetCurrentPlayerIdx())
	if hint != nil {
		resObj.Hint = &controller.PiquetWebOutputHint{
			CardIndex:      hint.CardIndex,
			DiscardIndices: hint.DiscardIndices,
			Reason:         hint.Reason,
		}
		resObj.MessageCode = "piquet.hintAvailable"
	} else {
		resObj.MessageCode = "piquet.noHint"
	}
	return marshalOrError(resObj)
}

// ActionLogOutput アクションログをJSON出力
func (p *PiquetWebPresenter) ActionLogOutput(g interfaces.PiquetGame) string {
	return actionLogOutputJSON(g)
}

func newPiquetWebOutputBase(g interfaces.PiquetGame) *controller.PiquetWebOutput {
	cfg := g.GetConfig()
	res := &controller.PiquetWebOutput{
		Players:              piquetBuildPlayers(g),
		Phase:                int(g.GetPhase()),
		DealNumber:           g.GetDealNumber(),
		DealsPerPartie:       g.GetDealsPerPartie(),
		ElderIdx:             g.GetElderIdx(),
		YoungerIdx:           g.GetYoungerIdx(),
		CurrentPlayerIdx:     g.GetCurrentPlayerIdx(),
		LeadPlayerIdx:        g.GetLeadPlayerIdx(),
		TrickNumber:          g.GetTrickNumber(),
		TricksWon:            [2]int{g.GetTricksWon(0), g.GetTricksWon(1)},
		ExchangeTurn:         int(g.GetExchangeTurn()),
		ElderExchangedCnt:    g.GetElderExchangedCnt(),
		YoungerExchangedCnt:  g.GetYoungerExchangedCnt(),
		ElderTalon:           cardsToOutput(g.GetElderTalon()),
		YoungerTalon:         cardsToOutput(g.GetYoungerTalon()),
		ElderRevealedTalon:   cardsToOutput(g.GetElderRevealedTalon()),
		YoungerRevealedTalon: cardsToOutput(g.GetYoungerRevealedTalon()),
		CarteBlanche:         [2]bool{g.GetCarteBlanche(0), g.GetCarteBlanche(1)},
		DeclStage:            int(g.GetDeclStage()),
		DeclResults:          piquetBuildDeclResults(g.GetDeclResults()),
		CurrentTrick:         piquetBuildCurrentTrick(g.GetCurrentTrick()),
		GameEndFlag:          g.GetGameEndFlag(),
		WinnerIdx:            g.GetWinnerIdx(),
		Config: controller.PiquetWebOutputConfig{
			CpuDifficulty:  int(cfg.CpuDifficulty),
			DealsPerPartie: cfg.DealsPerPartie,
		},
	}
	return res
}

func piquetBuildPlayers(g interfaces.PiquetGame) []*controller.PiquetWebOutputPlayer {
	all := g.GetPlayers()
	out := make([]*controller.PiquetWebOutputPlayer, len(all))
	for i, pl := range all {
		var cards []*controller.WebOutputCard
		if pl.GetIsHuman() {
			cards = make([]*controller.WebOutputCard, pl.GetCardsSize())
			for j := 0; j < pl.GetCardsSize(); j++ {
				cards[j] = cardToOutput(pl.GetCard(j))
			}
		} else {
			cards = []*controller.WebOutputCard{}
		}
		out[i] = &controller.PiquetWebOutputPlayer{
			ID:         i,
			IsHuman:    pl.GetIsHuman(),
			CardCount:  pl.GetCardsSize(),
			Cards:      cards,
			TrickCount: g.GetTricksWon(i),
			DeclScore:  pl.GetDeclScore(),
			TrickScore: pl.GetTrickScore(),
			BonusScore: pl.GetBonusScore(),
			RoundScore: pl.GetRoundScore(),
			MatchScore: pl.GetMatchScore(),
		}
	}
	return out
}

func piquetBuildCurrentTrick(trick []*domain.TrickCard) []*controller.WebOutputTrickCard {
	out := make([]*controller.WebOutputTrickCard, len(trick))
	for i, tc := range trick {
		out[i] = &controller.WebOutputTrickCard{
			PlayerIdx: tc.PlayerIdx,
			Card:      cardToOutput(tc.Card),
		}
	}
	return out
}

func piquetBuildDeclResults(results []*domain.PiquetDeclarationResult) []*controller.PiquetWebOutputDeclaration {
	out := make([]*controller.PiquetWebOutputDeclaration, len(results))
	for i, r := range results {
		out[i] = &controller.PiquetWebOutputDeclaration{
			Kind:         int(r.Kind),
			ElderClaim:   piquetBuildClaim(r.ElderClaim),
			YoungerClaim: piquetBuildClaim(r.YoungerClaim),
			Winner:       r.Winner,
			Score:        r.Score,
			ScoredBy:     r.ScoredBy,
			Sets:         piquetBuildClaims(r.Sets),
		}
	}
	return out
}

func piquetBuildClaim(c *domain.PiquetClaim) *controller.PiquetWebOutputClaim {
	if c == nil {
		return nil
	}
	return &controller.PiquetWebOutputClaim{
		Length:   c.Length,
		TopRank:  c.TopRank,
		PipTotal: c.PipTotal,
		Suit:     c.Suit,
		Cards:    cardsToOutput(c.Cards),
	}
}

func piquetBuildClaims(claims []*domain.PiquetClaim) []*controller.PiquetWebOutputClaim {
	if len(claims) == 0 {
		return nil
	}
	out := make([]*controller.PiquetWebOutputClaim, len(claims))
	for i, c := range claims {
		out[i] = piquetBuildClaim(c)
	}
	return out
}

func piquetPhaseMessageCode(phase domain.PiquetPhase) string {
	switch phase {
	case domain.PiquetPhaseExchange:
		return "piquet.exchange"
	case domain.PiquetPhaseDeclaration:
		return "piquet.declaration"
	case domain.PiquetPhasePlay:
		return "piquet.play"
	case domain.PiquetPhaseScore:
		return "piquet.score"
	case domain.PiquetPhaseGameEnd:
		return "piquet.gameEnd"
	}
	return ""
}
