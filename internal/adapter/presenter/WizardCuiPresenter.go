package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// wizardCuiCardStr は非52枚デッキのウィザード/ジェスター札を含むカードを
// テキスト描画する。design 5 → "Wizard"、design 6 → "Jester"、それ以外は
// 標準の "SPADE 5" 形式 (cuiCardStr)。
func wizardCuiCardStr(card *domain.Card) string {
	if card == nil {
		return "??"
	}
	switch card.GetDesign() {
	case domain.WizardDesignWizard:
		return color.BoldYellow("Wizard")
	case domain.WizardDesignJester:
		return color.Green("Jester")
	}
	return cuiCardStr(card)
}

// wizardLeadSuit returns the lead suit (design 1..4) of the in-progress trick,
// or -1 when none is established yet — a Wizard led (suit is irrelevant) or only
// Jesters have been played. Mirrors Wizard.leadSuitOfTrick so the CUI can name
// the suit players must follow without a dedicated domain accessor.
func wizardLeadSuit(trick []*domain.TrickCard) int {
	for _, tc := range trick {
		switch tc.Card.GetDesign() {
		case domain.WizardDesignWizard:
			return -1
		case domain.WizardDesignJester:
			continue
		default:
			return tc.Card.GetDesign()
		}
	}
	return -1
}

// WizardLegalMark は「この札は今のトリックに出せる」ことを示す印。
const WizardLegalMark = "*"

// wizardIndexedCardListStr は手札をインデックス付き (ウィザード/ジェスター対応) で描画する。
// legal が非 nil なら、出せる札に WizardLegalMark を付ける。
func wizardIndexedCardListStr(hand cuiCardList, legal []bool) string {
	parts := make([]string, hand.GetCardsSize())
	for i := range parts {
		parts[i] = fmt.Sprintf("[%d]%s", i, wizardCuiCardStr(hand.GetCard(i)))
		if i < len(legal) && legal[i] {
			parts[i] += WizardLegalMark
		}
	}
	return strings.Join(parts, "  ")
}

// wizardLegalIndices は手札の各札が今のトリックに出せるかを返す。
// **全部出せるなら nil を返す。**リード時のように全部に印が付くだけの場面では、
// 印も凡例もノイズにしかならない。
func wizardLegalIndices(player *domain.WizardPlayer, trick []*domain.TrickCard) []bool {
	legal := make([]bool, player.GetCardsSize())
	all := true
	for i := range legal {
		legal[i] = domain.WizardIsLegalPlay(player.GetCard(i), trick, player)
		if !legal[i] {
			all = false
		}
	}
	if all {
		return nil
	}
	return legal
}

// wizardPlayerStr returns the display string for a single Wizard player.
func wizardPlayerStr(player *domain.WizardPlayer, i int, legal []bool) string {
	var b strings.Builder
	bidStr := i18n.T("wizard.bidPending")
	if player.GetBid() >= 0 {
		bidStr = strconv.Itoa(player.GetBid())
	}
	b.WriteString(i18n.Tf("wizard.playerLine",
		"name", cuiPlayerName(player, i),
		"bid", bidStr,
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"cum", strconv.Itoa(player.GetCumulativeScore()),
		"round", strconv.Itoa(player.GetRoundScore()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(wizardIndexedCardListStr(player, legal) + "\n")
	}
	return b.String()
}

// WizardCuiPresenter renders the Wizard CUI view.
type WizardCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *WizardCuiPresenter) Output(o interfaces.WizardGame, lastErr error) string {
	return buildCuiOutput(i18n.T("wizard.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("wizard.header",
			"round", strconv.Itoa(o.GetRoundNumber()),
			"total", strconv.Itoa(o.GetTotalRounds()),
			"hand", strconv.Itoa(o.GetHandSize()),
			"trick", strconv.Itoa(o.GetTrickNumber())) + "\n")

		if trumpCard := o.GetTrumpCard(); trumpCard != nil {
			b.WriteString(i18n.Tf("wizard.trumpCard", "card", wizardCuiCardStr(trumpCard)) + "\n")
		} else {
			b.WriteString(i18n.T("wizard.trumpNone") + "\n")
		}

		dealerIdx := o.GetDealerIdx()
		b.WriteString(i18n.Tf("wizard.dealer",
			"name", cuiPlayerName(o.GetPlayer(dealerIdx), dealerIdx)) + "\n")

		legalShown := false
		for i := 0; i < o.GetPlayerCnt(); i++ {
			// 出せる札の印は、人間の手番のプレイフェーズだけ付ける。
			var legal []bool
			pl := o.GetPlayer(i)
			// 手番の判定は席番号で行う。IsHumanTurn() は「人間は 1 人」という
			// 暗黙の前提に寄りかかる。
			if pl != nil && pl.GetIsHuman() && o.GetPhase() == domain.WizardPhasePlay && i == o.GetCurrentPlayerIdx() {
				legal = wizardLegalIndices(pl, o.GetCurrentTrick())
				legalShown = legalShown || legal != nil
			}
			b.WriteString(wizardPlayerStr(pl, i, legal))
		}

		b.WriteString("----------\n")

		// Current trick
		trick := o.GetCurrentTrick()
		cuiTrickBlock(b, trick,
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return wizardCuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(o.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		// Game state
		if o.GetGameEndFlag() {
			winnerIdx := o.GetWinnerIdx()
			player := o.GetPlayer(winnerIdx)
			banner := i18n.Tf("wizard.gameEnd", "name", cuiPlayerName(player, winnerIdx))
			b.WriteString(color.Green(banner) + "\n")
			// 最終ラウンドの的中もここで見たい (ゲーム終了は RoundEnd を経由せずに
			// 出ることがある)。
			wizardBidAccuracyLine(b, o)
			return
		}
		switch o.GetPhase() {
		case domain.WizardPhaseBid:
			bidIdx := o.GetBidPlayerIdx()
			name := cuiPlayerName(o.GetPlayer(bidIdx), bidIdx)
			if restricted := o.GetRestrictedBid(); restricted >= 0 {
				b.WriteString(i18n.Tf("wizard.promptBidRestricted",
					"name", name,
					"restricted", strconv.Itoa(restricted)) + "\n")
			} else {
				b.WriteString(i18n.Tf("wizard.promptBid", "name", name) + "\n")
			}
			b.WriteString(i18n.T("wizard.promptBidHelp") + "\n")
		case domain.WizardPhasePlay:
			currentIdx := o.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("wizard.promptPlay",
				"name", cuiPlayerName(o.GetPlayer(currentIdx), currentIdx)) + "\n")
			// Spell out the trick number and the lead suit players must follow, so
			// the human need not infer it from the raw trick block.
			leadName := i18n.T("wizard.leadNone")
			if leadSuit := wizardLeadSuit(o.GetCurrentTrick()); leadSuit >= 1 && leadSuit <= 4 {
				leadName = cuiSuitName(leadSuit)
			}
			b.WriteString(i18n.Tf("wizard.promptLead",
				"trick", strconv.Itoa(o.GetTrickNumber()),
				"lead", leadName) + "\n")
			if legalShown {
				b.WriteString(i18n.T("wizard.legalMark") + "\n")
			}
			b.WriteString(i18n.T("wizard.promptPlayHelp") + "\n")
		case domain.WizardPhaseTrickEnd:
			b.WriteString(i18n.T("wizard.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("wizard.promptTrickEndHelp") + "\n")
		case domain.WizardPhaseRoundEnd:
			b.WriteString(i18n.T("wizard.promptRoundEnd") + "\n")
			wizardBidAccuracyLine(b, o)
			b.WriteString(i18n.T("wizard.promptRoundEndHelp") + "\n")
		}
	})
}

// wizardBidAccuracyLine writes the bid-vs-actual summary for the finished round.
// 得点はビッドとの一致で決まるので、各行を見比べさせずにまとめて出す
// (Web の wizardBidAccuracy と同じ判定: delta = トリック - ビッド、
// 未ビッド (bid < 0) の席は契約が無いので対象外)。
func wizardBidAccuracyLine(b *strings.Builder, o interfaces.WizardGame) {
	entries := make([]string, 0, o.GetPlayerCnt())
	for i := 0; i < o.GetPlayerCnt(); i++ {
		player := o.GetPlayer(i)
		if player == nil || player.GetBid() < 0 {
			continue
		}
		name := cuiPlayerName(player, i)
		bid, tricks := player.GetBid(), player.GetTrickCount()
		switch {
		case tricks == bid:
			entries = append(entries, i18n.Tf("wizard.bidAccuracyMade",
				"name", name, "bid", strconv.Itoa(bid)))
		case tricks > bid:
			entries = append(entries, i18n.Tf("wizard.bidAccuracyOver",
				"name", name, "bid", strconv.Itoa(bid), "tricks", strconv.Itoa(tricks)))
		default:
			entries = append(entries, i18n.Tf("wizard.bidAccuracyUnder",
				"name", name, "bid", strconv.Itoa(bid), "tricks", strconv.Itoa(tricks)))
		}
	}
	if len(entries) == 0 {
		return
	}
	b.WriteString(i18n.Tf("wizard.bidAccuracyLine", "entries", strings.Join(entries, " / ")) + "\n")
}

// HintOutput emits the current Wizard hint.
func (p *WizardCuiPresenter) HintOutput(o interfaces.WizardGame) string {
	hint := o.GetHint()
	if hint == nil {
		return i18n.T("wizard.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, nil)
	if hint.Bid != nil {
		return color.Yellow(i18n.Tf("wizard.hintBid",
			"bid", strconv.Itoa(*hint.Bid),
			"reason", reason)) + "\n"
	}
	if hint.CardIndex == nil {
		return i18n.T("wizard.hintNone") + "\n"
	}
	humanIdx := -1
	for i := 0; i < o.GetPlayerCnt(); i++ {
		if o.GetPlayer(i).GetIsHuman() {
			humanIdx = i
			break
		}
	}
	if humanIdx < 0 {
		return i18n.T("wizard.hintNone") + "\n"
	}
	card := o.GetPlayer(humanIdx).GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("wizard.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", wizardCuiCardStr(card),
		"reason", reason)) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *WizardCuiPresenter) ActionLogOutput(o interfaces.WizardGame) string {
	return actionLogOutputTextWithNames(o, func(idx int) string { return cuiPlayerName(o.GetPlayer(idx), idx) })
}
