//go:build !js || !wasm || extra3

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func trexCardListStr(cards []*domain.Card, indexed bool) string {
	if len(cards) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(cards))
	for i, c := range cards {
		if indexed {
			parts = append(parts, "["+strconv.Itoa(i)+"]"+cuiCardStr(c))
			continue
		}
		parts = append(parts, cuiCardStr(c))
	}
	return strings.Join(parts, " ")
}

// trexContractKeys は契約番号から i18n キーへの対応。
var trexContractKeys = map[domain.TrexContract]string{
	domain.TrexContractKingOfHearts: "trex.contractKingOfHearts",
	domain.TrexContractDiamonds:     "trex.contractDiamonds",
	domain.TrexContractQueens:       "trex.contractQueens",
	domain.TrexContractTricks:       "trex.contractTricks",
	domain.TrexContractTrix:         "trex.contractTrix",
}

// trexContractName は契約名を返す。未選択は専用の文言。
func trexContractName(c domain.TrexContract) string {
	if key, ok := trexContractKeys[c]; ok {
		return i18n.T(key)
	}
	return i18n.T("trex.contractNone")
}

// TrexCuiPresenter renders the Trex CUI view.
type TrexCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *TrexCuiPresenter) Output(c interfaces.TrexGame, lastErr error) string {
	return buildCuiOutput(i18n.T("trex.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("trex.header",
			"deal", strconv.Itoa(c.GetDealNumber()),
			"total", strconv.Itoa(domain.TrexTotalDeals),
			"king", strconv.Itoa(c.GetKingIdx()),
			"contract", trexContractName(c.GetContract())) + "\n")
		sb.WriteString(i18n.T("trex.ruleLine") + "\n")

		if c.IsTrix() {
			sb.WriteString(p.runsBlock(c))
		} else if c.GetPhase() == domain.TrexPhasePlay {
			trick := make([]*domain.Card, 0, len(c.GetTrick()))
			for _, tc := range c.GetTrick() {
				trick = append(trick, tc.Card)
			}
			sb.WriteString(i18n.Tf("trex.trickLine", "cards", trexCardListStr(trick, false)) + "\n")
		}

		for i, player := range c.GetPlayers() {
			sb.WriteString(i18n.Tf("trex.playerLine",
				"name", cuiPlayerName(player, i),
				"score", strconv.Itoa(c.GetScore(i)),
				"deal", strconv.Itoa(c.GetDealScore(i)),
				"cards", strconv.Itoa(player.GetCardsSize())) + "\n")
			if player.GetIsHuman() && player.GetCardsSize() > 0 {
				hand := make([]*domain.Card, 0, player.GetCardsSize())
				for j := range player.GetCardsSize() {
					hand = append(hand, player.GetCard(j))
				}
				sb.WriteString("  " + trexCardListStr(hand, true) + "\n")
			}
		}

		sb.WriteString("----------\n")
		cuiErrorBlock(sb, lastErr)
		sb.WriteString(p.promptBlock(c))
	})
}

// runsBlock はドミノの 4 列を描く。
func (p *TrexCuiPresenter) runsBlock(c interfaces.TrexGame) string {
	var sb strings.Builder
	for _, suit := range trexSuits {
		started, low, high := c.GetSuitRun(suit)
		span := i18n.T("trex.runEmpty")
		if started {
			span = strconv.Itoa(low) + "-" + strconv.Itoa(high)
		}
		sb.WriteString(i18n.Tf("trex.runLine", "suit", cuiSuitName(suit), "span", span) + "\n")
	}
	return sb.String()
}

// promptBlock はフェーズごとの案内を返す。
func (p *TrexCuiPresenter) promptBlock(c interfaces.TrexGame) string {
	if c.GetGameEndFlag() {
		key := "trex.gameEndLose"
		if c.GetWinnerIdx() == 0 {
			key = "trex.gameEndWin"
		}
		return color.Green(i18n.T(key)) + "\n"
	}

	var sb strings.Builder
	switch c.GetPhase() {
	case domain.TrexPhaseChoose:
		if c.GetKingIdx() != 0 {
			sb.WriteString(i18n.Tf("trex.waitingForKing", "king", strconv.Itoa(c.GetKingIdx())) + "\n")
			return sb.String()
		}
		// 残りの契約を番号つきで出す。1 王国に 1 度ずつしか選べないので、
		// 何が残っているかが見えていないと選べない。
		parts := make([]string, 0, 5)
		for _, ct := range c.AvailableContracts() {
			parts = append(parts, strconv.Itoa(int(ct))+"="+trexContractName(ct))
		}
		sb.WriteString(i18n.Tf("trex.promptChoose", "list", strings.Join(parts, " / ")) + "\n")
	case domain.TrexPhasePlay:
		if c.IsTrix() && len(c.GetValidPlayIndices(0)) == 0 && c.GetCurrentPlayerIdx() == 0 {
			sb.WriteString(i18n.T("trex.promptPass") + "\n")
			return sb.String()
		}
		sb.WriteString(i18n.T("trex.promptPlay") + "\n")
	case domain.TrexPhaseDealEnd:
		sb.WriteString(i18n.T("trex.promptNext") + "\n")
	case domain.TrexPhaseGameEnd:
		return ""
	}
	return sb.String()
}

// HintOutput emits the current Trex hint.
func (p *TrexCuiPresenter) HintOutput(c interfaces.TrexGame) string {
	hint := trexHint(c)
	key := trexHintReasonKeys[hint.Reason]
	if key == "" {
		key = "trex.hintNone"
	}
	switch {
	case hint.Contract != nil:
		return color.Yellow(i18n.Tf("trex.hintChoose",
			"n", strconv.Itoa(*hint.Contract),
			"name", trexContractName(domain.TrexContract(*hint.Contract)),
			"reason", i18n.T(key))) + "\n"
	case hint.Pass:
		return color.Yellow(i18n.Tf("trex.hintPass", "reason", i18n.T(key))) + "\n"
	case hint.CardIndex != nil:
		return color.Yellow(i18n.Tf("trex.hintPlay",
			"idx", strconv.Itoa(*hint.CardIndex), "reason", i18n.T(key))) + "\n"
	default:
		return color.Yellow(i18n.T(key)) + "\n"
	}
}

// trexHintReasonKeys maps the reason identifiers trexHint returns to i18n keys.
// The Web presenter ships the identifier and the frontend resolves it; the CUI
// must resolve it here or it prints the raw key.
var trexHintReasonKeys = map[string]string{
	"trex.hint.game_end":      "trex.hintReasonGameEnd",
	"trex.hint.not_your_turn": "trex.hintReasonNotYourTurn",
	"trex.hint.choose":        "trex.hintReasonChoose",
	"trex.hint.play":          "trex.hintReasonPlay",
	"trex.hint.pass":          "trex.hintReasonPass",
	"trex.hint.none":          "trex.hintNone",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *TrexCuiPresenter) ActionLogOutput(c interfaces.TrexGame) string {
	return actionLogOutputText(c)
}
