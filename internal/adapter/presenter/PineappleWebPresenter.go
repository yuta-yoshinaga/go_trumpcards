package presenter

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PineappleWebPresenter パイナップルポーカーWebプレゼンタークラス
type PineappleWebPresenter struct{}

// Output ゲーム状態をJSON出力
func (pp *PineappleWebPresenter) Output(p interfaces.PineappleGame, lastErr error) string {
	resObj := pp.buildOutput(p, lastErr)
	return marshalOrError(resObj)
}

// buildOutput ゲーム状態をPineappleWebOutputに変換
func (pp *PineappleWebPresenter) buildOutput(p interfaces.PineappleGame, lastErr error) *controller.PineappleWebOutput {
	// 共通フィールドを HoldemWebOutput として構築し、Pineapple固有フィールドを追加
	base := buildCommunityCardBaseOutput(p)
	resObj := &controller.PineappleWebOutput{
		Players:          buildPokerPlayersOutput(p.GetPhase(), p.GetPlayerCnt(), func(i int) communityCardPresenterPlayer { return p.GetPlayer(i) }, domain.PineapplePhaseShowdown, domain.PineapplePhaseEnd, pokerHandName),
		CommunityCards:   base.CommunityCards,
		Pot:              base.Pot,
		SidePots:         base.SidePots,
		DealerIdx:        base.DealerIdx,
		CurrentTurn:      base.CurrentTurn,
		Phase:            base.Phase,
		GameEndFlag:      base.GameEndFlag,
		LastBet:          base.LastBet,
		MinRaise:         base.MinRaise,
		BettingLimit:     base.BettingLimit,
		RaiseCount:       base.RaiseCount,
		MaxBetAmount:     base.MaxBetAmount,
		RoundResults:     base.RoundResults,
		CpuActions:       base.CpuActions,
		HandCount:        base.HandCount,
		SmallBlind:       base.SmallBlind,
		BigBlind:         base.BigBlind,
		TournamentMode:   base.TournamentMode,
		BlindLevelHands:  base.BlindLevelHands,
		BlindMultiplier:  base.BlindMultiplier,
		TableSize:        base.TableSize,
		RebuyAvailable:   base.RebuyAvailable,
		AddonAvailable:   base.AddonAvailable,
		RebuyCounts:      base.RebuyCounts,
		AddonUsed:        base.AddonUsed,
		RebuyEnabled:     base.RebuyEnabled,
		AddonEnabled:     base.AddonEnabled,
		RebuyMaxCount:    base.RebuyMaxCount,
		RebuyChips:       base.RebuyChips,
		AddonChips:       base.AddonChips,
		RebuyPeriodHands: base.RebuyPeriodHands,
		AddonAfterHand:   base.AddonAfterHand,
		RebuyPhaseType:   base.RebuyPhaseType,
		MuckAvailable:    base.MuckAvailable,
		Equity:           base.Equity,
		PotOdds:          base.PotOdds,
		MetaAI:           base.MetaAI,
		Profile:          base.Profile,
		// Pineapple 固有フィールド
		IsDiscardPhase: p.IsDiscardPhase(),
		DiscardDone:    p.GetDiscardDone(),
	}

	resObj.Message, resObj.MessageCode, resObj.MessageParams = pp.buildMessage(p, lastErr)
	return resObj
}

// buildMessage ゲーム結果メッセージを構築
func (pp *PineappleWebPresenter) buildMessage(p interfaces.PineappleGame, lastErr error) (string, string, map[string]string) {
	if lastErr != nil {
		return lastErr.Error(), "", nil
	}
	if p.IsDiscardPhase() {
		return "Select a card to discard.", "pineapple.discard.prompt", nil
	}
	if p.IsMuckAvailable() {
		return "Muck or show your hand.", "pineapple.muck.prompt", nil
	}
	if p.GetGameEndFlag() {
		msg, code := pp.buildResultMessage(p)
		return msg, code, nil
	}
	return "", "", nil
}

// buildResultMessage builds the end-of-round message and its i18n code
func (pp *PineappleWebPresenter) buildResultMessage(p interfaces.PineappleGame) (string, string) {
	results := p.GetRoundResults()
	if len(results) == 0 {
		return "Game over.", "pineapple.result.gameOver"
	}

	for _, r := range results {
		if p.GetPlayer(r.PlayerIdx).GetIsHuman() {
			if r.WonAmount > 0 {
				return "You are the winner.", "pineapple.result.win"
			}
		}
	}

	for i := 0; i < p.GetPlayerCnt(); i++ {
		if p.GetPlayer(i).GetIsHuman() && p.GetPlayer(i).GetFolded() {
			return "You folded.", "pineapple.result.folded"
		}
	}

	for _, r := range results {
		if p.GetPlayer(r.PlayerIdx).GetIsHuman() && r.Mucked {
			return "You mucked.", "pineapple.result.mucked"
		}
	}

	return "You lose.", "pineapple.result.lose"
}

// ActionLogOutput 棋譜をJSON出力
func (pp *PineappleWebPresenter) ActionLogOutput(p interfaces.PineappleGame) string {
	return actionLogOutputJSON(p)
}
