//go:build !js || !wasm || casino

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

// SevenCardStudCuiPresenter renders the Seven Card Stud CUI view.
type SevenCardStudCuiPresenter struct{}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *SevenCardStudCuiPresenter) ActionLogOutput(s interfaces.SevenCardStudGame) string {
	return actionLogOutputText(s)
}

// Output renders the current game state for the active locale (#1699).
func (p *SevenCardStudCuiPresenter) Output(s interfaces.SevenCardStudGame, lastErr error) string {
	return buildCuiOutput(i18n.T("sevencardstud.outputTitle"), func(b *strings.Builder) {
		cfg := s.GetConfig()
		if cfg.TournamentMode {
			b.WriteString(i18n.Tf("sevencardstud.tournamentLine",
				"hand", strconv.Itoa(s.GetHandCount()),
				"ante", strconv.Itoa(cfg.Ante),
				"bringIn", strconv.Itoa(cfg.BringIn),
				"levelup", strconv.Itoa(cfg.AnteLevelHands)) + "\n")
			if cfg.RebuyEnabled {
				b.WriteString(i18n.Tf("sevencardstud.rebuyLine",
					"chips", strconv.Itoa(cfg.RebuyChips),
					"max", strconv.Itoa(cfg.RebuyMaxCount),
					"period", strconv.Itoa(cfg.RebuyPeriodHands)) + "\n")
			}
			if cfg.AddonEnabled {
				b.WriteString(i18n.Tf("sevencardstud.addonLine",
					"chips", strconv.Itoa(cfg.AddonChips),
					"after", strconv.Itoa(cfg.AddonAfterHand)) + "\n")
			}
		}

		b.WriteString(i18n.Tf("sevencardstud.tableMax", "n", strconv.Itoa(s.GetPlayerCnt())) + "\n")
		b.WriteString(i18n.Tf("sevencardstud.dealerLine", "idx", strconv.Itoa(s.GetDealerIdx())) + "\n")

		b.WriteString(i18n.Tf("sevencardstud.anteLine",
			"ante", strconv.Itoa(cfg.Ante),
			"bringIn", strconv.Itoa(cfg.BringIn),
			"small", strconv.Itoa(cfg.SmallBet),
			"big", strconv.Itoa(cfg.BigBet)) + "\n")

		b.WriteString(i18n.Tf("sevencardstud.potLine", "pot", strconv.Itoa(s.GetPot())) + "\n")

		if int(cfg.BettingLimit) < len(domain.BettingLimitNames) {
			b.WriteString(i18n.Tf("sevencardstud.limitLine", "name", domain.BettingLimitNames[cfg.BettingLimit]) + "\n")
		}

		b.WriteString("----------\n")
		for i := 0; i < s.GetPlayerCnt(); i++ {
			player := s.GetPlayer(i)
			b.WriteString(cuiPlayerNameWithStyle(player, i))
			b.WriteString(i18n.Tf("sevencardstud.playerChips", "chips", strconv.Itoa(player.GetChips())))

			if player.GetTotalHands() > 0 {
				b.WriteString(i18n.Tf("sevencardstud.playerStats",
					"vpip", strconv.Itoa(player.GetVPIP()),
					"pfr", strconv.Itoa(player.GetPFR()),
					"tb", strconv.Itoa(player.GetThreeBet()),
					"af", player.GetAFDisplay()))
			}

			if player.GetFolded() {
				b.WriteString(color.BoldYellow(i18n.T("sevencardstud.playerFolded")))
			} else if player.GetAllIn() {
				b.WriteString(color.BoldYellow(i18n.T("sevencardstud.playerAllIn")))
			}

			if player.GetCurrentBet() > 0 {
				b.WriteString(i18n.Tf("sevencardstud.playerBet", "bet", strconv.Itoa(player.GetCurrentBet())))
			}
			b.WriteString("\n")

			// Door cards (face up — visible to all players)
			if doorCards := player.GetDoorCards(); len(doorCards) > 0 {
				b.WriteString(i18n.Tf("sevencardstud.doorCards", "cards", cuiCardSliceStrEmoji(doorCards)) + "\n")
			}

			// Hole cards (face down — only the human sees them)
			if player.GetIsHuman() && !player.GetFolded() {
				if holeCards := player.GetHoleCards(); len(holeCards) > 0 {
					b.WriteString(i18n.Tf("sevencardstud.holeCards", "cards", cuiCardSliceStrEmoji(holeCards)) + "\n")
				}
				// In Razz (lowball) the goal is the lowest hand; surface the human's
				// current best low, since the shared high-hand view never shows it.
				if s.GetIsLowball() {
					if low, complete := domain.SevenCardStudRazzBestLow(player.GetAllCards()); len(low) > 0 {
						key := "sevencardstud.razzLowIncomplete"
						if complete {
							key = "sevencardstud.razzLowComplete"
						}
						b.WriteString(i18n.Tf(key, "cards", cuiCardSliceStrEmoji(low)) + "\n")
					}
				} else if rank, best := player.PeekBestHand(); len(best) > 0 {
					// **Web は常時「いまの最善役」を出しているのに、ハイ戦の CUI は
					// ショーダウンまで役名を一切出していなかった (#4695)。**3rd〜7th
					// street の間ずっと自分の手が何に達しているか分からない。
					// PeekBestHand は状態を変えないので、描画から呼んで安全。
					b.WriteString(i18n.Tf("sevencardstud.currentBestHand",
						"hand", cuiPokerHandName(rank),
						"cards", cuiCardSliceStrEmoji(best)) + "\n")
				}
			}
		}

		cpuActions := s.GetCpuActions()
		if len(cpuActions) > 0 {
			b.WriteString("----------\n")
			b.WriteString(color.Bold(i18n.T("sevencardstud.cpuActionsHeader")) + "\n")
			for _, action := range cpuActions {
				b.WriteString(i18n.Tf("sevencardstud.cpuActionLine",
					"idx", strconv.Itoa(action.PlayerIdx),
					"action", cuiBettingActionName(action.Action)))
				if action.Amount > 0 {
					b.WriteString(i18n.Tf("sevencardstud.cpuActionAmount", "amount", strconv.Itoa(action.Amount)))
				}
				b.WriteString("\n")
			}
		}

		phase := s.GetPhase()
		isBettingStreet := phase >= domain.SevenCardStudPhaseThirdStreet && phase <= domain.SevenCardStudPhaseSeventhStreet
		if !s.GetGameEndFlag() && isBettingStreet {
			turnIdx := s.GetCurrentTurn()
			cur := s.GetPlayer(turnIdx)
			if cur != nil && cur.GetIsHuman() && !cur.GetFolded() {
				toCall := s.GetLastBet() - cur.GetCurrentBet()
				if toCall < 0 {
					toCall = 0
				}
				b.WriteString("----------\n")
				if toCall == 0 {
					b.WriteString(i18n.Tf("sevencardstud.actionPromptCheck",
						"raise", strconv.Itoa(s.GetMinRaise())) + "\n")
				} else {
					b.WriteString(i18n.Tf("sevencardstud.actionPrompt",
						"call", strconv.Itoa(toCall),
						"raise", strconv.Itoa(s.GetMinRaise())) + "\n")
				}
			}
		}

		results := s.GetRoundResults()
		if len(results) > 0 && (s.GetPhase() == domain.SevenCardStudPhaseEnd || s.GetPhase() == domain.SevenCardStudPhaseShowdown) {
			b.WriteString("==========\n")
			b.WriteString(color.Bold(i18n.T("sevencardstud.resultsHeader")) + "\n")
			for _, r := range results {
				name := cuiPlayerName(s.GetPlayer(r.PlayerIdx), r.PlayerIdx)
				kickers := ""
				if ks := domain.FormatKickers(r.Kickers); ks != "" {
					kickers = i18n.Tf("sevencardstud.resultKickers", "kickers", ks)
				}
				switch {
				case r.Mucked:
					b.WriteString(i18n.Tf("sevencardstud.resultMucked", "name", name))
				case r.HandName != "":
					b.WriteString(i18n.Tf("sevencardstud.resultHand",
						"name", name,
						"hand", cuiPokerHandName(r.HandRank),
						"kickers", kickers))
				default:
					b.WriteString(i18n.Tf("sevencardstud.resultName", "name", name))
				}
				if r.WonAmount > 0 {
					b.WriteString(i18n.Tf("sevencardstud.wonAmount", "total", strconv.Itoa(r.WonAmount)))
				}
				b.WriteString("\n")
			}
		}

		if s.IsMuckAvailable() {
			b.WriteString("----------\n")
			b.WriteString(i18n.T("sevencardstud.muckPrompt") + "\n")
		}

		if s.GetPhase() == domain.SevenCardStudPhaseRebuy {
			b.WriteString("----------\n")
			switch s.GetRebuyPhaseType() {
			case domain.SevenCardStudRebuyPhaseRebuy:
				rebuyCounts := s.GetRebuyCounts()
				humanIdx := -1
				for i := 0; i < s.GetPlayerCnt(); i++ {
					if s.GetPlayer(i).GetIsHuman() {
						humanIdx = i
						break
					}
				}
				if humanIdx >= 0 {
					b.WriteString(i18n.Tf("sevencardstud.rebuyPrompt",
						"chips", strconv.Itoa(cfg.RebuyChips),
						"used", strconv.Itoa(rebuyCounts[humanIdx]),
						"max", strconv.Itoa(cfg.RebuyMaxCount)) + "\n")
				}
			case domain.SevenCardStudRebuyPhaseAddon:
				b.WriteString(i18n.Tf("sevencardstud.addonPrompt", "chips", strconv.Itoa(cfg.AddonChips)) + "\n")
			}
		}

		cuiErrorBlock(b, lastErr)

		if s.GetGameEndFlag() {
			b.WriteString(i18n.T("sevencardstud.gameEnd") + "\n")
		}
	})
}

// sevenCardStudThreeStraight reports whether the three distinct card values form
// a run (including the Q-K-A wheel where the Ace is high).
func sevenCardStudThreeStraight(cards []*domain.Card) bool {
	vs := make([]int, 0, len(cards))
	seen := map[int]bool{}
	for _, c := range cards {
		v := c.GetValue()
		if seen[v] {
			return false
		}
		seen[v] = true
		vs = append(vs, v)
	}
	if len(vs) != 3 {
		return false
	}
	sort.Ints(vs)
	if vs[2]-vs[0] == 2 {
		return true
	}
	return vs[0] == 1 && vs[1] == 12 && vs[2] == 13 // Q-K-A
}

// sevenCardStudThirdStreetAdvice returns whether to continue on third street and
// the i18n reason key, using classic starting-hand rules (any pair, three to a
// flush, three to a straight, or three high cards).
func sevenCardStudThirdStreetAdvice(cards []*domain.Card) (cont bool, reasonKey string) {
	if len(cards) < 3 {
		return false, "sevencardstud.hintReasonFold"
	}
	vals := map[int]int{}
	suits := map[int]int{}
	high := 0
	for _, c := range cards {
		vals[c.GetValue()]++
		suits[c.GetDesign()]++
		if v := c.GetValue(); v == 1 || v >= 10 {
			high++
		}
	}
	for _, n := range vals {
		if n >= 2 {
			return true, "sevencardstud.hintReasonPair"
		}
	}
	for _, n := range suits {
		if n >= 3 {
			return true, "sevencardstud.hintReasonFlush"
		}
	}
	if sevenCardStudThreeStraight(cards) {
		return true, "sevencardstud.hintReasonStraight"
	}
	if high >= 3 {
		return true, "sevencardstud.hintReasonHigh"
	}
	return false, "sevencardstud.hintReasonFold"
}

// razzLowCardMax はラズで「ロー札」と数える上限 (A-8)。フロントの
// razzHint.ts の LOW_CARD_MAX と同じ。
const razzLowCardMax = 8

// razzBettingPhases はラズのヒントを出すストリート。
//
// **ハイの助言と違って全ストリートで出す。**フロントの getRazzHint が
// そうなっている。ラズは引くたびにロー札の枚数が変わるので、3rd だけでは
// 足りない。
var razzBettingPhases = map[int]bool{
	domain.SevenCardStudPhaseThirdStreet:   true,
	domain.SevenCardStudPhaseFourthStreet:  true,
	domain.SevenCardStudPhaseFifthStreet:   true,
	domain.SevenCardStudPhaseSixthStreet:   true,
	domain.SevenCardStudPhaseSeventhStreet: true,
}

// razzAdvice はラズの助言を返す。判定はフロントの getRazzHint と同じ規則:
// ペアがあれば降りる、ロー札が十分あればレイズ、そこそこならコール。
//
// **ラズは役の強弱が逆。**ハイの基本戦略 (ペアがあれば続行) をそのまま
// 当てると、最悪の手で押すことになる。
func razzAdvice(cards []*domain.Card) (actionKey, reasonKey string) {
	seen := map[int]bool{}
	low := 0
	for _, c := range cards {
		v := c.GetValue()
		if seen[v] {
			return "sevencardstud.hintFold", "sevencardstud.hintReasonRazzPair"
		}
		seen[v] = true
		if v <= razzLowCardMax {
			low++
		}
	}
	total := len(cards)
	if low >= min(5, total) {
		return "sevencardstud.hintRaise", "sevencardstud.hintReasonRazzStrong"
	}
	if low >= min(4, total-1) {
		return "sevencardstud.hintCall", "sevencardstud.hintReasonRazzDecent"
	}
	return "sevencardstud.hintFold", "sevencardstud.hintReasonRazzWeak"
}

// HintOutput advises on the human's turn. The high game uses basic
// starting-hand strategy on third street; Razz inverts hand strength, so it
// gets its own fold/call/raise advice on every betting street (#4703).
func (p *SevenCardStudCuiPresenter) HintOutput(s interfaces.SevenCardStudGame) string {
	if !s.IsHumanTurn() {
		return i18n.T("sevencardstud.hintNone") + "\n"
	}
	if s.GetIsLowball() {
		return razzHintOutput(s)
	}
	if s.GetIsHiLo() {
		return sevenCardStudHiLoHintOutput(s)
	}
	if s.GetPhase() != domain.SevenCardStudPhaseThirdStreet {
		return i18n.T("sevencardstud.hintNone") + "\n"
	}
	player := s.GetPlayer(s.GetCurrentTurn())
	if player == nil {
		return i18n.T("sevencardstud.hintNone") + "\n"
	}
	cont, reasonKey := sevenCardStudThirdStreetAdvice(player.GetAllCards())
	action := i18n.T("sevencardstud.hintFold")
	if cont {
		action = i18n.T("sevencardstud.hintContinue")
	}
	return color.Yellow(i18n.Tf("sevencardstud.hint",
		"action", action, "reason", i18n.T(reasonKey))) + "\n"
}

// sevenCardStudLowQualifier は Hi-Lo でローに数える上限 (8 or Better)。
const sevenCardStudLowQualifier = 8

// sevenCardStudLowCardsNeeded はロー成立に必要な枚数。
const sevenCardStudLowCardsNeeded = 5

// sevenCardStudHiLoHintOutput は Hi-Lo (8 or Better) 向けのヒント行を返す。
//
// **ハイ専用の基本戦略をそのまま当てない (#4704)。**Hi-Lo はポットの半分が
// ローに行くので、ハイとしては弱くてもロー札がそろっていれば続ける価値がある。
// ハイ用の判定 (ペア/3フラッシュ/3ストレート/3ハイカード) だけで見ると、
// 勝ち目のある手を降ろすことになる。
//
// **ハイと違って全ストリートで出す。**ロー札は5枚必要なので、3枚しか無い
// 3rd street ではロー分岐に永遠に届かない。フロントの getSevenCardStudHint も
// 全ベッティングストリートで判定している。
func sevenCardStudHiLoHintOutput(s interfaces.SevenCardStudGame) string {
	if !razzBettingPhases[s.GetPhase()] {
		return i18n.T("sevencardstud.hintNone") + "\n"
	}
	player := s.GetPlayer(s.GetCurrentTurn())
	if player == nil || player.GetFolded() || len(player.GetAllCards()) == 0 {
		return i18n.T("sevencardstud.hintNone") + "\n"
	}
	cards := player.GetAllCards()
	action, reasonKey := i18n.T("sevencardstud.hintFold"), "sevencardstud.hintReasonFold"
	switch {
	case sevenCardStudHasPair(cards):
		action, reasonKey = i18n.T("sevencardstud.hintContinue"), "sevencardstud.hintReasonPair"
	case sevenCardStudLowCards(cards) >= sevenCardStudLowCardsNeeded:
		action, reasonKey = i18n.T("sevencardstud.hintContinue"), "sevencardstud.hintReasonHiLoLow"
	case sevenCardStudHasHighCard(cards):
		action, reasonKey = i18n.T("sevencardstud.hintContinue"), "sevencardstud.hintReasonHigh"
	}
	return color.Yellow(i18n.Tf("sevencardstud.hint",
		"action", action, "reason", i18n.T(reasonKey))) + "\n"
}

// sevenCardStudHasPair は同じランクが2枚以上あるかを返す。
func sevenCardStudHasPair(cards []*domain.Card) bool {
	seen := map[int]bool{}
	for _, c := range cards {
		if seen[c.GetValue()] {
			return true
		}
		seen[c.GetValue()] = true
	}
	return false
}

// sevenCardStudLowCards は8以下の札の枚数を返す。**エースはローの最強札**
// なので 1 のまま数えて問題ない。
func sevenCardStudLowCards(cards []*domain.Card) int {
	n := 0
	for _, c := range cards {
		if c.GetValue() <= sevenCardStudLowQualifier {
			n++
		}
	}
	return n
}

// sevenCardStudHasHighCard は 10 以上またはエースを持っているかを返す。
func sevenCardStudHasHighCard(cards []*domain.Card) bool {
	for _, c := range cards {
		if v := c.GetValue(); v == 1 || v >= 10 {
			return true
		}
	}
	return false
}

// razzHintOutput はラズ (Lowball) 向けのヒント行を返す。
func razzHintOutput(s interfaces.SevenCardStudGame) string {
	if !razzBettingPhases[s.GetPhase()] {
		return i18n.T("sevencardstud.hintNone") + "\n"
	}
	player := s.GetPlayer(s.GetCurrentTurn())
	if player == nil || player.GetFolded() || len(player.GetAllCards()) == 0 {
		return i18n.T("sevencardstud.hintNone") + "\n"
	}
	actionKey, reasonKey := razzAdvice(player.GetAllCards())
	return color.Yellow(i18n.Tf("sevencardstud.hint",
		"action", i18n.T(actionKey), "reason", i18n.T(reasonKey))) + "\n"
}
