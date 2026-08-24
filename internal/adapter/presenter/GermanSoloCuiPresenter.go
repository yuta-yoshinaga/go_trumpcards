//go:build !js || !wasm || classic

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// germanSoloBidLabel maps a bid value to its i18n label key.
func germanSoloBidLabel(bid domain.GermanSoloBid) string {
	switch bid {
	case domain.GermanSoloBidMussfrage:
		return i18n.T("germansolo.bidMussfrage")
	case domain.GermanSoloBidFrage:
		return i18n.T("germansolo.bidFrage")
	case domain.GermanSoloBidSolo:
		return i18n.T("germansolo.bidSolo")
	case domain.GermanSoloBidTout:
		return i18n.T("germansolo.bidTout")
	default:
		return i18n.T("germansolo.bidNone")
	}
}

// germanSoloTrumpLabel maps a trump suit value to its i18n label key.
func germanSoloTrumpLabel(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return i18n.T("germansolo.suitSpade")
	case domain.CardDesignClover:
		return i18n.T("germansolo.suitClub")
	case domain.CardDesignHeart:
		return i18n.T("germansolo.suitHeart")
	case domain.CardDesignDiamond:
		return i18n.T("germansolo.suitDiamond")
	default:
		return i18n.T("germansolo.suitNone")
	}
}

func germanSoloPlayerStr(g interfaces.GermanSoloGame, idx int) string {
	player := g.GetPlayer(idx)
	if player == nil {
		return ""
	}
	scores := g.GetPlayerScores()
	role := i18n.T("germansolo.roleDefender")
	switch {
	case idx == g.GetDeclarerIdx():
		role = i18n.T("germansolo.roleDeclarer")
	// **味方の役柄は伏せている間は出さない。** GetPartnerIdx は呼ばれたエースが
	// 場に出るまで -1 を返すので、その間は全員が「連合」に見える。
	case idx == g.GetPartnerIdx():
		role = i18n.T("germansolo.rolePartner")
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("germansolo.playerLine",
		"name", cuiPlayerName(player, idx),
		"role", role,
		"cards", strconv.Itoa(player.GetCardsSize()),
		"score", strconv.Itoa(scores[idx]),
		"tricks", strconv.Itoa(player.GetTrickCount()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(germanSoloIndexedHandStr(player, g.GetTrumpSuit()) + "\n")
	}
	return b.String()
}

// germanSoloMatadorKeys はマタドールの序列と表示名キーの対応。
var germanSoloMatadorKeys = map[int]string{
	1: "germansolo.matadorSpadille",
	2: "germansolo.matadorManille",
	3: "germansolo.matadorBasta",
}

// germanSoloIndexedHandStr は手札をインデックス付きで並べ、マタドールに注記を付ける。
//
// **マタドールは常にトリックに勝つ。**Web はバッジで示すのに、CUI は素の
// 一覧しか出しておらず、序列を覚えていないと気づけなかった (#4919)。
// 切り札が未確定なら注記は付かない。
//
// 索引と区切りは他ゲームと同じ formatCardList に任せる。自前で並べると
// 区切り幅が揃わない。
func germanSoloIndexedHandStr(player cuiCardList, trump int) string {
	return formatCardList(player, func(c *domain.Card) string {
		s := cuiCardStr(c)
		if key := germanSoloMatadorKeys[domain.GermanSoloMatadorRank(c, trump)]; key != "" {
			s += "(" + i18n.T(key) + ")"
		}
		return s
	}, "  ", true)
}

// GermanSoloCuiPresenter renders the GermanSolo CUI view.
type GermanSoloCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *GermanSoloCuiPresenter) Output(g interfaces.GermanSoloGame, lastErr error) string {
	return buildCuiOutput(i18n.T("germansolo.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("germansolo.round",
			"round", strconv.Itoa(g.GetRoundNumber()),
			"trick", strconv.Itoa(g.GetTrickNumber()),
			"trump", germanSoloTrumpLabel(g.GetTrumpSuit())) + "\n")

		b.WriteString(germanSoloContractLine(g))
		b.WriteString(germanSoloAceLine(g))

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(germanSoloPlayerStr(g, i))
		}

		b.WriteString("----------\n")

		cuiTrickBlock(b, g.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(g.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if g.GetGameEndFlag() {
			winner := g.GetWinnerPlayer()
			var winnerStr string
			if winner >= 0 {
				winnerStr = cuiPlayerName(g.GetPlayer(winner), winner)
			}
			banner := i18n.Tf("germansolo.gameEnd", "name", winnerStr)
			b.WriteString(color.Green(banner) + "\n")
			return
		}
		switch g.GetPhase() {
		case domain.GermanSoloPhaseBid:
			bidderIdx := g.GetCurrentBidderIdx()
			b.WriteString(i18n.Tf("germansolo.promptBid",
				"bid", germanSoloBidLabel(g.GetHighestBid()),
				"name", cuiPlayerName(g.GetPlayer(bidderIdx), bidderIdx)) + "\n")
			// **上回れる宣言だけを案内する。** 既に Solo が出ている卓で
			// frage を勧めると、その通り打った人間だけが弾かれる。
			b.WriteString(i18n.Tf("germansolo.promptBiddable",
				"bids", germanSoloBiddableStr(g)) + "\n")
			b.WriteString(i18n.T("germansolo.promptBidHelp") + "\n")
		case domain.GermanSoloPhaseAceCall:
			b.WriteString(i18n.Tf("germansolo.promptAce",
				"name", cuiPlayerName(g.GetPlayer(g.GetDeclarerIdx()), g.GetDeclarerIdx())) + "\n")
			b.WriteString(i18n.T("germansolo.promptAceHelp") + "\n")
		case domain.GermanSoloPhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			b.WriteString(i18n.Tf("germansolo.promptPlay",
				"name", cuiPlayerName(g.GetPlayer(currentIdx), currentIdx),
				"trump", germanSoloTrumpLabel(g.GetTrumpSuit())) + "\n")
			b.WriteString(i18n.T("germansolo.promptPlayHelp") + "\n")
		case domain.GermanSoloPhaseTrickEnd:
			b.WriteString(i18n.T("germansolo.promptTrickEnd") + "\n")
			b.WriteString(i18n.T("germansolo.promptTrickEndHelp") + "\n")
		case domain.GermanSoloPhaseRoundEnd:
			declarerTricks, defenderTricks := g.GetSideTrickCounts()
			b.WriteString(i18n.Tf("germansolo.promptRoundEnd",
				"declarer", cuiPlayerName(g.GetPlayer(g.GetDeclarerIdx()), g.GetDeclarerIdx()),
				"outcome", germanSoloOutcomeLabel(g.GetOutcome()),
				"tricks", strconv.Itoa(declarerTricks),
				"opp", strconv.Itoa(defenderTricks),
				"need", strconv.Itoa(g.RequiredTricks())) + "\n")
			b.WriteString(i18n.T("germansolo.promptRoundEndHelp") + "\n")
		}
	})
}

// germanSoloOutcomeLabel maps a deal outcome to its i18n label key.
func germanSoloOutcomeLabel(o domain.GermanSoloOutcome) string {
	switch o {
	case domain.GermanSoloOutcomeMade:
		return i18n.T("germansolo.outcomeMade")
	case domain.GermanSoloOutcomeFailed:
		return i18n.T("germansolo.outcomeFailed")
	default:
		return i18n.T("germansolo.outcomeNone")
	}
}

// HintOutput emits the current GermanSolo hint.
func (p *GermanSoloCuiPresenter) HintOutput(g interfaces.GermanSoloGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("germansolo.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, germanSoloHintReasonKeys)
	if len(hint.CardIndices) > 0 {
		playerIdx := g.GetCurrentPlayerIdx()
		player := g.GetPlayer(playerIdx)
		cards := make([]string, len(hint.CardIndices))
		for i, idx := range hint.CardIndices {
			if player != nil && idx >= 0 && idx < player.GetCardsSize() {
				cards[i] = "[" + strconv.Itoa(idx) + "]" + cuiCardStr(player.GetCard(idx))
			} else {
				cards[i] = strconv.Itoa(idx)
			}
		}
		return color.Yellow(i18n.Tf("germansolo.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	}
	// Bid-phase decisions carry no cards; render them as an action recommendation
	// instead of the meaningless "recommended cards: -" line.
	if actionKey, ok := germanSoloBidActionKeys[hint.Reason]; ok {
		return color.Yellow(i18n.Tf("germansolo.hintDecision",
			"action", i18n.T(actionKey),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("germansolo.hintCard",
		"cards", "-",
		"reason", reason)) + "\n"
}

// germanSoloBidActionKeys maps bid-phase hint-reason identifiers to the i18n key for
// the recommended action name (Frage / Solo / Tout / Pass).
var germanSoloBidActionKeys = map[string]string{
	"bid_frage": "germansolo.hintActionFrage",
	"bid_solo":  "germansolo.hintActionSolo",
	"bid_tout":  "germansolo.hintActionTout",
	"bid_pass":  "germansolo.hintActionPass",
}

// germanSoloHintReasonKeys maps GermanSolo-specific hint-reason identifiers to i18n keys.
var germanSoloHintReasonKeys = map[string]string{
	"lead_high":    "germansolo.hintReasonLeadHigh",
	"lead_low":     "germansolo.hintReasonLeadLow",
	"follow_win":   "germansolo.hintReasonFollowWin",
	"follow_duck":  "germansolo.hintReasonFollowDuck",
	"give_partner": "germansolo.hintReasonGivePartner",
	"discard_low":  "germansolo.hintReasonDiscardLow",
	"call_ace":     "germansolo.hintReasonCallAce",
	"bid_frage":    "germansolo.hintReasonBidFrage",
	"bid_solo":     "germansolo.hintReasonBidSolo",
	"bid_tout":     "germansolo.hintReasonBidTout",
	"bid_pass":     "germansolo.hintReasonBidPass",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *GermanSoloCuiPresenter) ActionLogOutput(g interfaces.GermanSoloGame) string {
	return actionLogOutputTextForSeats[*domain.GermanSoloPlayer](g)
}

// germanSoloContractLine は確定した契約と必要トリック数の行を組み立てる。
//
// **必要トリック数は契約ごとに違う。** Tout だけ 8 で、それ以外は 5。
// 画面に出さないと、5 取って喜んだ Tout がその場で失敗になる理由が読めない。
func germanSoloContractLine(g interfaces.GermanSoloGame) string {
	if g.GetDeclarerIdx() < 0 || g.GetWinningBid() == domain.GermanSoloBidNone {
		return ""
	}
	declarerTricks, defenderTricks := g.GetSideTrickCounts()
	return i18n.Tf("germansolo.contractLine",
		"bid", germanSoloBidLabel(g.GetWinningBid()),
		"need", strconv.Itoa(g.RequiredTricks()),
		"tricks", strconv.Itoa(declarerTricks),
		"opp", strconv.Itoa(defenderTricks)) + "\n"
}

// germanSoloBiddableStr は今の卓で宣言できるビッドを並べた文字列を返す。
func germanSoloBiddableStr(g interfaces.GermanSoloGame) string {
	bids := g.GetBiddableBids()
	if len(bids) == 0 {
		return i18n.T("germansolo.biddableNone")
	}
	labels := make([]string, 0, len(bids))
	for _, b := range bids {
		labels = append(labels, germanSoloBidLabel(domain.GermanSoloBid(b)))
	}
	return strings.Join(labels, " / ")
}

// germanSoloAceLine はエース呼びの行を組み立てる。
//
// **呼ばれたエースは公開情報、持ち主は伏せる。** 呼び声は卓で聞こえるのでエースの
// スートは全員が知っているが、誰が持っているかはそのエースが場に出るまで
// 分からない —— そこがこのゲームの緊張感そのもの。
func germanSoloAceLine(g interfaces.GermanSoloGame) string {
	if g.IsPlayingAlone() {
		return i18n.T("germansolo.playsAlone") + "\n"
	}
	suit := g.GetCalledAceSuit()
	if !germanSoloSuitInRange(suit) {
		return i18n.T("germansolo.aceUncalled") + "\n"
	}
	partner := g.GetPartnerIdx()
	if partner < 0 || partner >= g.GetPlayerCnt() {
		return i18n.Tf("germansolo.aceCalledHidden",
			"suit", germanSoloTrumpLabel(suit)) + "\n"
	}
	return i18n.Tf("germansolo.aceCalledRevealed",
		"suit", germanSoloTrumpLabel(suit),
		"name", cuiPlayerName(g.GetPlayer(partner), partner)) + "\n"
}

// germanSoloSuitInRange は suit が本物のスート (1..4) かを返す。
func germanSoloSuitInRange(suit int) bool {
	return suit >= domain.CardDesignSpade && suit <= domain.CardDesignMax
}
