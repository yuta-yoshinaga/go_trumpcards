//go:build !js || !wasm || solo

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

// quodlibetCardStr は 32 枚デッキの札を「♠K」の形で返す。
//
// **共通の cuiCardStrEmoji は数値をそのまま出す。** それだと A・J・Q・K が
// 1・11・12・13 と並び、7〜10 と地続きに見えてしまう ── このゲームは Q と J
// にだけ点が付く種目 (オーバー / ウンター) を持つので、絵札が読めないと
// 何を避ければよいのか画面から分からない。
func quodlibetCardStr(card *domain.Card) string {
	if card == nil {
		return "??"
	}
	designs := []string{"🃏", "♠", "♣", "♥", "♦"}
	d := card.GetDesign()
	if d < 0 || d >= len(designs) {
		d = 0
	}
	s := designs[d] + cuiRankLabel(card.GetValue())
	if card.GetDesign() == domain.CardDesignHeart || card.GetDesign() == domain.CardDesignDiamond {
		return color.Red(s)
	}
	return s
}

// QuodlibetCuiPresenter renders the Quodlibet CUI view.
type QuodlibetCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *QuodlibetCuiPresenter) Output(g interfaces.QuodlibetGame, lastErr error) string {
	return buildCuiOutput(i18n.T("quodlibet.helpTitle"), func(b *strings.Builder) {
		p.writeHeader(b, g)
		p.writeSeats(b, g)
		b.WriteString("----------\n")
		p.writeTable(b, g)
		cuiErrorBlock(b, lastErr)
		if g.GetGameEndFlag() {
			p.writeGameEnd(b, g)
			return
		}
		p.writePrompt(b, g)
	})
}

// writeHeader はディール番号・輪・コントラクトを書く。
func (p *QuodlibetCuiPresenter) writeHeader(b *strings.Builder, g interfaces.QuodlibetGame) {
	b.WriteString(i18n.Tf("quodlibet.deal",
		"n", strconv.Itoa(g.GetDealNumber()+1),
		"total", strconv.Itoa(domain.QuodlibetTotalDeals),
		"round", strconv.Itoa(g.GetRoundNumber())) + "\n")
	contract := g.GetCurrentContract()
	if contract >= 0 {
		b.WriteString(i18n.Tf("quodlibet.contract",
			"name", quodlibetContractLabel(contract)) + "\n")
		if !domain.QuodlibetIsSheddingContract(contract) {
			b.WriteString(i18n.Tf("quodlibet.trick",
				"n", strconv.Itoa(g.GetTrickNumber()),
				"total", strconv.Itoa(domain.QuodlibetHandSize)) + "\n")
		}
	}
}

// quodlibetContractLabel はコントラクトの訳名を返す。
func quodlibetContractLabel(contract int) string {
	return i18n.T("quodlibet.contractName." + domain.QuodlibetContractName(contract))
}

// writeSeats は席ごとの罰点と手札を書く。
func (p *QuodlibetCuiPresenter) writeSeats(b *strings.Builder, g interfaces.QuodlibetGame) {
	contract := g.GetCurrentContract()
	human := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		if pl := g.GetPlayer(i); pl != nil && pl.GetIsHuman() {
			human = i
			break
		}
	}
	for i := 0; i < g.GetPlayerCnt(); i++ {
		player := g.GetPlayer(i)
		if player == nil {
			continue
		}
		role := i18n.T("quodlibet.rolePlayer")
		if i == g.GetDealerIdx() {
			role = i18n.T("quodlibet.roleDealer")
		}
		b.WriteString(i18n.Tf("quodlibet.playerLine",
			"name", cuiPlayerName(player, i),
			"role", role,
			"cards", strconv.Itoa(player.GetCardsSize()),
			"tricks", strconv.Itoa(player.GetTrickCount()),
			"penalty", strconv.Itoa(player.GetPenalty())) + "\n")
		// **第 3 の輪では手札の見え方そのものが規則。** 「開いたズボン」では
		// 自分の手札が伏せられ、「狩猟」では全員のぶんが開く。
		if player.GetCardsSize() > 0 && domain.QuodlibetHandVisibility(contract, human, i) {
			b.WriteString(quodlibetIndexedHand(player) + "\n")
		}
	}
}

// quodlibetIndexedHand は手札を番号付きで返す。
func quodlibetIndexedHand(p *domain.QuodlibetPlayer) string {
	parts := make([]string, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		parts[i] = fmt.Sprintf("[%d]%s", i, quodlibetCardStr(p.GetCard(i)))
	}
	return strings.Join(parts, "  ")
}

// writeTable はトリック / 場 / 重ねを書く。
func (p *QuodlibetCuiPresenter) writeTable(b *strings.Builder, g interfaces.QuodlibetGame) {
	switch g.GetCurrentContract() {
	case domain.QuodlibetSnack:
		p.writeSnackTable(b, g)
	case domain.QuodlibetQuadrature:
		p.writeStack(b, g)
	default:
		cuiTrickBlock(b, g.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return quodlibetCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)
	}
}

// writeSnackTable は小食いの場を書く。
func (p *QuodlibetCuiPresenter) writeSnackTable(b *strings.Builder, g interfaces.QuodlibetGame) {
	placed := g.GetTablePlaced()
	for suit := domain.CardDesignSpade; suit <= domain.CardDesignDiamond; suit++ {
		parts := make([]string, 0, domain.QuodlibetHandSize)
		for i := 0; i < domain.QuodlibetHandSize; i++ {
			if placed[suit]&(uint16(1)<<uint(i)) != 0 {
				parts = append(parts, quodlibetCardStr(domain.NewCard(suit, quodlibetValueAt(i), false)))
			}
		}
		line := i18n.T("quodlibet.tableEmpty")
		if len(parts) > 0 {
			line = strings.Join(parts, " ")
		}
		b.WriteString(i18n.Tf("quodlibet.tableRow",
			"suit", cuiSuitName(suit), "cards", line) + "\n")
	}
}

// quodlibetValueAt は弱い順の位置から札の値を返す。
func quodlibetValueAt(rankIdx int) int {
	order := []int{7, 8, 9, 10, 11, 12, 13, 1}
	if rankIdx < 0 || rankIdx >= len(order) {
		return 0
	}
	return order[rankIdx]
}

// writeStack は四分の重ねを書く。
func (p *QuodlibetCuiPresenter) writeStack(b *strings.Builder, g interfaces.QuodlibetGame) {
	cards := g.GetStack()
	if len(cards) == 0 {
		b.WriteString(i18n.T("quodlibet.stackEmpty") + "\n")
		return
	}
	parts := make([]string, 0, len(cards))
	for _, c := range cards {
		parts = append(parts, quodlibetCardStr(c))
	}
	b.WriteString(i18n.Tf("quodlibet.stack", "cards", strings.Join(parts, " → ")) + "\n")
}

// writeGameEnd は終局の行を書く。
func (p *QuodlibetCuiPresenter) writeGameEnd(b *strings.Builder, g interfaces.QuodlibetGame) {
	winners := g.GetWinners()
	names := make([]string, 0, len(winners))
	for _, w := range winners {
		names = append(names, cuiPlayerName(g.GetPlayer(w), w))
	}
	// **勝つのは罰点が一番少ない人。** 名前だけ出すと、どちらの向きの勝負か
	// 画面から読めない。
	b.WriteString(color.Green(i18n.Tf("quodlibet.gameEnd",
		"names", strings.Join(names, ", "))) + "\n")
}

// writePrompt はフェーズに応じた案内を書く。
func (p *QuodlibetCuiPresenter) writePrompt(b *strings.Builder, g interfaces.QuodlibetGame) {
	switch g.GetPhase() {
	case domain.QuodlibetPhaseSelectContract:
		p.writeContractPrompt(b, g)
	case domain.QuodlibetPhasePlay:
		p.writePlayPrompt(b, g)
	case domain.QuodlibetPhaseDealEnd:
		p.writeDealEndPrompt(b, g)
	}
}

// writeContractPrompt は選べるコントラクトを並べる。
func (p *QuodlibetCuiPresenter) writeContractPrompt(b *strings.Builder, g interfaces.QuodlibetGame) {
	dealer := g.GetDealerIdx()
	b.WriteString(i18n.Tf("quodlibet.promptContract",
		"name", cuiPlayerName(g.GetPlayer(dealer), dealer)) + "\n")
	// **選べるのはこの輪の残りだけ。** 番号だけ出しても何の種目か分からない
	// ので、訳名を添える。
	parts := make([]string, 0, domain.QuodlibetContractsPerRound)
	for _, c := range g.GetAvailableContracts() {
		parts = append(parts, fmt.Sprintf("[%d]%s", c, quodlibetContractLabel(c)))
	}
	if len(parts) > 0 {
		b.WriteString(i18n.Tf("quodlibet.contractList", "list", strings.Join(parts, "  ")) + "\n")
	}
	b.WriteString(i18n.T("quodlibet.promptContractHelp") + "\n")
}

// writePlayPrompt は出せる札を並べる。
func (p *QuodlibetCuiPresenter) writePlayPrompt(b *strings.Builder, g interfaces.QuodlibetGame) {
	idx := g.GetCurrentTurn()
	b.WriteString(i18n.Tf("quodlibet.promptPlay",
		"name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
	player := g.GetPlayer(idx)
	if player == nil || !player.GetIsHuman() {
		return
	}
	valid := g.GetPlayableIndices(idx)
	parts := make([]string, 0, len(valid))
	for _, i := range valid {
		if i >= 0 && i < player.GetCardsSize() {
			parts = append(parts, "["+strconv.Itoa(i)+"]"+quodlibetCardStr(player.GetCard(i)))
		}
	}
	if len(parts) > 0 {
		b.WriteString(i18n.Tf("quodlibet.playableList", "cards", strings.Join(parts, "  ")) + "\n")
		b.WriteString(i18n.T("quodlibet.promptPlayHelp") + "\n")
		return
	}
	// **出せる札が無いときだけパスできる。** 四分は重ねが続かないと全員が
	// 詰まるので、この案内が無いと盤面が止まったように見える。
	if domain.QuodlibetIsSheddingContract(g.GetCurrentContract()) {
		b.WriteString(i18n.T("quodlibet.promptPass") + "\n")
	}
}

// writeDealEndPrompt はこのディールの罰点内訳を書く。
func (p *QuodlibetCuiPresenter) writeDealEndPrompt(b *strings.Builder, g interfaces.QuodlibetGame) {
	if detail := g.GetLastDealDetail(); detail != nil {
		parts := make([]string, 0, domain.QuodlibetPlayerCnt)
		for i := 0; i < g.GetPlayerCnt(); i++ {
			parts = append(parts, fmt.Sprintf("%s %+d",
				cuiPlayerName(g.GetPlayer(i), i), detail.Points[i]))
		}
		b.WriteString(i18n.Tf("quodlibet.promptDealEnd",
			"name", quodlibetContractLabel(detail.Contract),
			"points", strings.Join(parts, " / ")) + "\n")
	}
	b.WriteString(i18n.T("quodlibet.promptDealEndHelp") + "\n")
}

// HintOutput emits the current hint.
func (p *QuodlibetCuiPresenter) HintOutput(g interfaces.QuodlibetGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("quodlibet.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, quodlibetHintReasonKeys)
	if hint.Contract >= 0 {
		return color.Yellow(i18n.Tf("quodlibet.hintContract",
			"name", quodlibetContractLabel(hint.Contract), "reason", reason)) + "\n"
	}
	if len(hint.CardIndices) == 0 {
		return color.Yellow(i18n.Tf("quodlibet.hintCard", "cards", "-", "reason", reason)) + "\n"
	}
	player := g.GetPlayer(g.GetCurrentTurn())
	cards := make([]string, len(hint.CardIndices))
	for i, idx := range hint.CardIndices {
		if player != nil && idx >= 0 && idx < player.GetCardsSize() {
			cards[i] = "[" + strconv.Itoa(idx) + "]" + quodlibetCardStr(player.GetCard(idx))
			continue
		}
		cards[i] = strconv.Itoa(idx)
	}
	return color.Yellow(i18n.Tf("quodlibet.hintCard",
		"cards", strings.Join(cards, ", "), "reason", reason)) + "\n"
}

// quodlibetHintReasonKeys はヒント理由と i18n キーの対応。
var quodlibetHintReasonKeys = map[string]string{
	"pick_contract": "quodlibet.hintReasonContract",
	"avoid_penalty": "quodlibet.hintReasonAvoid",
	"shed_low":      "quodlibet.hintReasonShed",
	"pass":          "quodlibet.hintReasonPass",
	"next_deal":     "quodlibet.hintReasonNextDeal",
	"none":          "quodlibet.hintNone",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *QuodlibetCuiPresenter) ActionLogOutput(g interfaces.QuodlibetGame) string {
	return actionLogOutputTextForSeats[*domain.QuodlibetPlayer](g)
}
