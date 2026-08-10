//go:build !js || !wasm || extra3

package presenter

import (
	"sort"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// kaiserReveal は手札を公開する局面かを返す。
func kaiserReveal(g interfaces.KaiserGame) bool {
	phase := g.GetPhase()
	return phase == domain.KaiserPhaseHandEnd || phase == domain.KaiserPhaseGameEnd
}

// kaiserContractName は契約種別の表示名キーを返す。
func kaiserContractName(c domain.KaiserContract) string {
	switch c {
	case domain.KaiserContractNoTrump:
		return i18n.T("kaiser.noTrump")
	case domain.KaiserContractLowNoTrump:
		return i18n.T("kaiser.lowNoTrump")
	}
	return i18n.T("kaiser.withTrump")
}

// kaiserSuitName は切札スートの表示名を返す。
func kaiserSuitName(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return "♠"
	case domain.CardDesignClover:
		return "♣"
	case domain.CardDesignHeart:
		return "♥"
	case domain.CardDesignDiamond:
		return "♦"
	}
	return "-"
}

// kaiserPlayerStr returns the display string for a single Kaiser player.
func kaiserPlayerStr(g interfaces.KaiserGame, i int) string {
	player := g.GetPlayer(i)
	if player == nil {
		return ""
	}
	hand := i18n.Tf("kaiser.hiddenHand", "count", strconv.Itoa(player.GetCardsSize()))
	if player.GetIsHuman() || kaiserReveal(g) {
		var b strings.Builder
		for j := range player.GetCardsSize() {
			b.WriteString("[" + strconv.Itoa(j) + "]" + cuiCardStr(player.GetCard(j)) + " ")
		}
		hand = strings.TrimSpace(b.String())
	}
	role := ""
	if i == g.GetDealerIdx() {
		role += " " + i18n.T("kaiser.dealerTag")
	}
	if i == g.GetDeclarerIdx() {
		role += " " + i18n.T("kaiser.declarerTag")
	}
	return i18n.Tf("kaiser.playerLine",
		"name", cuiPlayerName(player, i),
		"team", strconv.Itoa(domain.KaiserTeamOf(i)),
		"role", role,
		"hand", hand) + "\n"
}

// KaiserCuiPresenter renders the Kaiser CUI view.
type KaiserCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *KaiserCuiPresenter) Output(g interfaces.KaiserGame, lastErr error) string {
	return buildCuiOutput(i18n.T("kaiser.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("kaiser.header",
			"hand", strconv.Itoa(g.GetHandNumber()),
			"target", strconv.Itoa(g.GetTargetScore()),
			"t0", strconv.Itoa(g.GetScore(0)),
			"t1", strconv.Itoa(g.GetScore(1))) + "\n")

		if hb := g.GetHighBid(); hb != nil {
			b.WriteString(i18n.Tf("kaiser.contractLine",
				"value", strconv.Itoa(hb.Value),
				"kind", kaiserContractName(g.GetContract()),
				"trump", kaiserSuitName(g.GetTrumpSuit())) + "\n")
		}

		for i := range g.GetPlayers() {
			b.WriteString(kaiserPlayerStr(g, i))
		}

		if trick := g.GetTrick(); len(trick) > 0 {
			var t strings.Builder
			for _, c := range trick {
				t.WriteString(cuiCardStr(c) + " ")
			}
			b.WriteString(i18n.Tf("kaiser.trick", "cards", strings.TrimSpace(t.String())) + "\n")
		}

		b.WriteString(i18n.Tf("kaiser.handPoints",
			"t0", strconv.Itoa(g.GetHandPoints(0)),
			"t1", strconv.Itoa(g.GetHandPoints(1))) + "\n")

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			banner := i18n.Tf("kaiser.gameEnd", "team", strconv.Itoa(g.GetWinnerTeam()))
			b.WriteString(color.Green(banner) + "\n")
			return
		}

		switch g.GetPhase() {
		case domain.KaiserPhaseBid:
			idx := g.GetBidPlayerIdx()
			b.WriteString(i18n.Tf("kaiser.promptBid", "name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
			b.WriteString(i18n.T("kaiser.promptBidHelp") + "\n")
		case domain.KaiserPhaseDiscard:
			if g.GetContract() == domain.KaiserContractTrump && g.GetTrumpSuit() == 0 {
				b.WriteString(i18n.T("kaiser.promptTrump") + "\n")
			} else {
				b.WriteString(i18n.T("kaiser.promptDiscard") + "\n")
			}
		case domain.KaiserPhasePlay:
			idx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("kaiser.promptPlay", "name", cuiPlayerName(g.GetPlayer(idx), idx)) + "\n")
			if g.IsHumanTurn() {
				var v strings.Builder
				for _, i := range g.KaiserValidPlays(idx) {
					v.WriteString(strconv.Itoa(i) + " ")
				}
				b.WriteString(i18n.Tf("kaiser.playable", "indexes", strings.TrimSpace(v.String())) + "\n")
			}
			b.WriteString(i18n.T("kaiser.promptPlayHelp") + "\n")
		case domain.KaiserPhaseHandEnd:
			if g.IsBidMade() {
				b.WriteString(i18n.T("kaiser.madeLine") + "\n")
			} else {
				b.WriteString(i18n.T("kaiser.setLine") + "\n")
			}
			b.WriteString(i18n.T("kaiser.promptHandEndHelp") + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *KaiserCuiPresenter) ActionLogOutput(g interfaces.KaiserGame) string {
	return actionLogOutputText(g)
}

// kaiserHintKey は今の局面で出すべき助言の i18n キーと、必要なら添える手札
// インデックスを返す。キーが空なら助言なし。
//
// **Web のヒントはプレイ局面しか扱わない** (`utils/hints/kaiserHint.ts` は
// `phase !== Play` で null を返す)。CUI はビッド・切札・捨て札まで面倒を見る。
// 判断点が多いのに CUI 側は今まで一切の補助が無かった (#4938)。
func kaiserHintKey(g interfaces.KaiserGame) (key string, idxs []int, suits []int) {
	if g.GetGameEndFlag() {
		return "kaiser.hintGameEnd", nil, nil
	}
	human := kaiserHumanIdx(g)
	if human < 0 {
		return "kaiser.hintNone", nil, nil
	}
	p := g.GetPlayer(human)

	switch g.GetPhase() {
	case domain.KaiserPhaseBid:
		if g.GetBidPlayerIdx() != human {
			return "kaiser.hintNotYourTurn", nil, nil
		}
		return kaiserBidHint(g, p, human), nil, nil
	case domain.KaiserPhaseDiscard:
		if g.GetDeclarerIdx() != human {
			return "kaiser.hintNotYourTurn", nil, nil
		}
		// **切札が先。**指定前に捨てるとドメインに弾かれる。
		if g.GetContract() == domain.KaiserContractTrump && g.GetTrumpSuit() == 0 {
			return "kaiser.hintTrump", nil, kaiserLongestSuits(p)
		}
		return "kaiser.hintDiscard", kaiserDiscardCandidates(g, p), nil
	case domain.KaiserPhasePlay:
		if g.GetCurrentPlayerIdx() != human {
			return "kaiser.hintNotYourTurn", nil, nil
		}
		k, i := kaiserPlayHint(g, p)
		return k, i, nil
	case domain.KaiserPhaseHandEnd:
		return "kaiser.hintHandEnd", nil, nil
	}
	return "kaiser.hintNone", nil, nil
}

// kaiserHumanIdx は人間の席を返す (居なければ -1)。
func kaiserHumanIdx(g interfaces.KaiserGame) int {
	for i, p := range g.GetPlayers() {
		if p != nil && p.GetIsHuman() {
			return i
		}
	}
	return -1
}

// kaiserBidHint はビッドすべきかを返す。
//
// **1 局で動くのは 10 点しかなく、最低ビッドが 7。**強い手でなければ降りる。
// 目安は「最長スートが 4 枚以上あり、その A か K を持っている」か「♥5 を
// 持っている」(単独で 5 点)。加えて 45 点以上ではビッドしないと加点できない
// ので、そのときは降りるより取りに行くほうがよい。
func kaiserBidHint(g interfaces.KaiserGame, p *domain.KaiserPlayer, human int) string {
	if g.GetScore(domain.KaiserTeamOf(human)) >= domain.KaiserMustBidThreshold {
		return "kaiser.hintBidMust"
	}
	strong := false
	for _, suit := range kaiserLongestSuits(p) {
		length, hasTop := 0, false
		for j := range p.GetCardsSize() {
			c := p.GetCard(j)
			if c.GetDesign() != suit {
				continue
			}
			length++
			if c.GetValue() == 1 || c.GetValue() == 13 {
				hasTop = true
			}
		}
		if length >= 4 && hasTop {
			strong = true
		}
	}
	for j := range p.GetCardsSize() {
		if domain.IsKaiserHeartFive(p.GetCard(j)) {
			strong = true
		}
	}
	if strong {
		return "kaiser.hintBid"
	}
	return "kaiser.hintPass"
}

// kaiserLongestSuits は手札で最も長いスートを返す (同数なら複数)。
func kaiserLongestSuits(p *domain.KaiserPlayer) []int {
	counts := map[int]int{}
	for j := range p.GetCardsSize() {
		counts[p.GetCard(j).GetDesign()]++
	}
	best := 0
	for _, n := range counts {
		if n > best {
			best = n
		}
	}
	out := []int{}
	// スート番号の昇順で返す。map の反復順に任せると出力が揺れる。
	for suit := domain.CardDesignSpade; suit <= domain.CardDesignDiamond; suit++ {
		if best > 0 && counts[suit] == best {
			out = append(out, suit)
		}
	}
	return out
}

// kaiserDiscardCandidates は捨てる 2 枚の候補インデックスを返す。
//
// **♥5 と ♠3 は捨てられない** (ドメインが拒否する)。切札も残す。残りから
// 低いものを 2 枚。
func kaiserDiscardCandidates(g interfaces.KaiserGame, p *domain.KaiserPlayer) []int {
	trump := g.GetTrumpSuit()
	type cand struct{ idx, value int }
	cands := []cand{}
	for j := range p.GetCardsSize() {
		c := p.GetCard(j)
		if domain.IsKaiserHeartFive(c) || domain.IsKaiserSpadeThree(c) {
			continue
		}
		if trump != 0 && c.GetDesign() == trump {
			continue
		}
		// A は最強なので最後に落とす。
		v := c.GetValue()
		if v == 1 {
			v = 14
		}
		cands = append(cands, cand{j, v})
	}
	sort.Slice(cands, func(a, b int) bool {
		if cands[a].value != cands[b].value {
			return cands[a].value < cands[b].value
		}
		return cands[a].idx < cands[b].idx
	})
	out := []int{}
	for i := 0; i < len(cands) && i < domain.KaiserKittySize; i++ {
		out = append(out, cands[i].idx)
	}
	sort.Ints(out)
	return out
}

// kaiserPlayHint は出す札の助言を返す。Web の getKaiserHint と同じ判断:
// 強制手 → ♠3 を手放す → ♥5 は自分から出さない。
func kaiserPlayHint(g interfaces.KaiserGame, p *domain.KaiserPlayer) (string, []int) {
	plays := g.KaiserValidPlays(g.GetCurrentPlayerIdx())
	if len(plays) == 0 {
		return "kaiser.hintNone", nil
	}
	if len(plays) == 1 {
		return "kaiser.hintForced", plays
	}
	for _, i := range plays {
		if domain.IsKaiserSpadeThree(p.GetCard(i)) {
			return "kaiser.hintDumpSpadeThree", []int{i}
		}
	}
	for _, i := range plays {
		if !domain.IsKaiserHeartFive(p.GetCard(i)) {
			return "kaiser.hintChoose", []int{i}
		}
	}
	return "kaiser.hintChoose", plays[:1]
}

// HintOutput emits the current Kaiser hint.
func (p *KaiserCuiPresenter) HintOutput(g interfaces.KaiserGame) string {
	key, idxs, suits := kaiserHintKey(g)
	if key == "" {
		key = "kaiser.hintNone"
	}
	msg := i18n.T(key)
	parts := make([]string, 0, len(idxs)+len(suits))
	for _, v := range idxs {
		parts = append(parts, "["+strconv.Itoa(v)+"]")
	}
	for _, s := range suits {
		parts = append(parts, kaiserSuitName(s))
	}
	if len(parts) > 0 {
		msg = i18n.Tf("kaiser.hintWith", "hint", msg, "list", strings.Join(parts, ", "))
	}
	return color.Yellow(msg) + "\n"
}
