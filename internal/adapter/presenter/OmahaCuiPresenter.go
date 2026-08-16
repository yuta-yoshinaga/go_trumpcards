//go:build !js || !wasm || casino

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

// OmahaCuiPresenter renders the Omaha Hold'em CUI view.
type OmahaCuiPresenter struct{}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *OmahaCuiPresenter) ActionLogOutput(o interfaces.OmahaGame) string {
	return actionLogOutputText(o)
}

// omahaTitleKey は、ホールカード枚数 (4=オマハ, 5=Big O) と Hi-Lo フラグから
// CUI ヘッダーに使う i18n タイトルキーを選択する。
func omahaTitleKey(holeCards int, hiLo bool) string {
	if holeCards >= 5 {
		if hiLo {
			return "omaha.helpTitleBigOHiLo"
		}
		return "omaha.helpTitleBigO"
	}
	if hiLo {
		return "omaha.helpTitleHiLo"
	}
	return "omaha.helpTitle"
}

// Output renders the current game state for the active locale (#1699).
func (p *OmahaCuiPresenter) Output(o interfaces.OmahaGame, lastErr error) string {
	titleKey := omahaTitleKey(o.GetHoleCardCount(), o.GetIsHiLo())
	return buildCuiOutput(i18n.T(titleKey), func(b *strings.Builder) {
		// Omaha's defining pitfall: exactly two hole cards must be used. Surface
		// it every render (the count adapts for Big O's five hole cards).
		b.WriteString(i18n.Tf("omaha.mandatoryRuleLine",
			"hole", strconv.Itoa(o.GetHoleCardCount())) + "\n")

		// In Hi-Lo the split's low qualifier (eight-or-better) is easy to miss from
		// the title alone, so spell it out during play — not just in the result.
		if o.GetIsHiLo() {
			b.WriteString(i18n.T("omaha.hiLoRuleLine") + "\n")
		}

		cfg := o.GetConfig()
		if cfg.TournamentMode {
			b.WriteString(i18n.Tf("omaha.tournamentLine",
				"hand", strconv.Itoa(o.GetHandCount()),
				"sb", strconv.Itoa(cfg.SmallBlind),
				"bb", strconv.Itoa(cfg.BigBlind),
				"levelup", strconv.Itoa(cfg.BlindLevelHands)) + "\n")
			if cfg.RebuyEnabled {
				b.WriteString(i18n.Tf("omaha.rebuyLine",
					"chips", strconv.Itoa(cfg.RebuyChips),
					"max", strconv.Itoa(cfg.RebuyMaxCount),
					"period", strconv.Itoa(cfg.RebuyPeriodHands)) + "\n")
			}
			if cfg.AddonEnabled {
				b.WriteString(i18n.Tf("omaha.addonLine",
					"chips", strconv.Itoa(cfg.AddonChips),
					"after", strconv.Itoa(cfg.AddonAfterHand)) + "\n")
			}
		}

		b.WriteString(i18n.Tf("omaha.tableMax", "n", strconv.Itoa(o.GetPlayerCnt())) + "\n")
		b.WriteString(i18n.Tf("omaha.dealerLine", "idx", strconv.Itoa(o.GetDealerIdx())) + "\n")

		cc := o.GetCommunityCards()
		if len(cc) == 0 {
			b.WriteString(i18n.T("omaha.communityNone") + "\n")
		} else {
			b.WriteString(i18n.Tf("omaha.communityCards", "cards", cuiCardSliceStrEmoji(cc)) + "\n")
		}

		// **ロー成立の見通しはボードだけで決まる。** Web は BoardLowBadge で
		// ベッティング中ずっと出しているのに、CUI はショーダウンの resultLow まで
		// 何も言わず、プレイヤーが毎回自分で低ランクを数えていた (#5483)。
		//
		// フロップ以降だけ。ボードが空では「まだ可能」以外に言うことがなく、
		// 毎ハンド必ず出る行は情報でなく雑音になる。ショーダウンでは resultLow が
		// 実際の結果を出すので見通しは要らない。
		if o.GetIsHiLo() && len(cc) > 0 && o.GetPhase() != domain.OmahaPhaseShowdown &&
			o.GetPhase() != domain.OmahaPhaseEnd {
			outlook := o.GetBoardLowOutlook()
			switch outlook.Status {
			case domain.OmahaBoardLowLive:
				b.WriteString(i18n.T("omaha.boardLowLive") + "\n")
			case domain.OmahaBoardLowPossible:
				b.WriteString(i18n.Tf("omaha.boardLowPossible",
					"needed", strconv.Itoa(outlook.Needed)) + "\n")
			case domain.OmahaBoardLowImpossible:
				b.WriteString(i18n.T("omaha.boardLowImpossible") + "\n")
			}
		}

		b.WriteString(i18n.Tf("omaha.potLine", "pot", strconv.Itoa(o.GetPot())) + "\n")

		if int(cfg.BettingLimit) < len(domain.BettingLimitNames) {
			b.WriteString(i18n.Tf("omaha.limitLine", "name", domain.BettingLimitNames[cfg.BettingLimit]) + "\n")
		}

		b.WriteString("----------\n")
		for i := 0; i < o.GetPlayerCnt(); i++ {
			player := o.GetPlayer(i)
			b.WriteString(cuiPlayerNameWithStyle(player, i))
			b.WriteString(i18n.Tf("omaha.playerChips", "chips", strconv.Itoa(player.GetChips())))

			if player.GetTotalHands() > 0 {
				b.WriteString(i18n.Tf("omaha.playerStats",
					"vpip", strconv.Itoa(player.GetVPIP()),
					"pfr", strconv.Itoa(player.GetPFR()),
					"tb", strconv.Itoa(player.GetThreeBet()),
					"af", player.GetAFDisplay()))
			}

			if player.GetFolded() {
				b.WriteString(" " + color.BoldYellow(i18n.T("omaha.playerFolded")))
			} else if player.GetAllIn() {
				b.WriteString(" " + color.BoldYellow(i18n.T("omaha.playerAllIn")))
			}

			if player.GetCurrentBet() > 0 {
				b.WriteString(i18n.Tf("omaha.playerBet", "bet", strconv.Itoa(player.GetCurrentBet())))
			}
			b.WriteString("\n")

			if player.GetIsHuman() && !player.GetFolded() {
				b.WriteString(i18n.Tf("omaha.humanHand", "cards", cuiCardListStrEmoji(player)) + "\n")
				// **Web は omaha-live-besthand で暫定ベストを常時出しているのに、
				// CUI は「2枚使用」の注意書きだけで実際の役を出していなかった
				// (#4680)。**手札4枚から必ず2枚という特殊ルールがあるぶん、
				// 暫定表示はミスを防ぐ補助になる。PeekBestHand は状態を変えない。
				if rank, best := player.PeekBestHand(cc); len(best) > 0 {
					b.WriteString(i18n.Tf("omaha.currentBestHand",
						"hand", cuiPokerHandName(rank),
						"cards", cuiCardSliceStrEmoji(best)) + "\n")
				}
			}
		}

		// **学習モードの値は Web にしか出ていなかった (#5482)。** GetEquity /
		// GetPotOdds は共有ヘルパ経由で Web へ送られ、Holdem 系の CUI も出して
		// いるのに、Omaha の CUI だけが取り残されていた。GetEquity は人間が
		// 降りている局面などで nil を返し、そのとき表示は消える。
		if o.IsHumanTurn() {
			if eq := o.GetEquity(); eq != nil {
				potOdds := o.GetPotOdds()
				b.WriteString("----------\n")
				b.WriteString(color.Bold(i18n.T("omaha.learningHeader")) + "\n")
				b.WriteString(i18n.Tf("omaha.learningLine",
					"equity", fmt.Sprintf("%.1f", eq.Equity*100),
					"potodds", fmt.Sprintf("%.1f", potOdds)) + "\n")
				if potOdds > 0 {
					if eq.Equity*100 > potOdds {
						b.WriteString(i18n.T("omaha.learningEvPlus") + "\n")
					} else {
						b.WriteString(i18n.T("omaha.learningEvMinus") + "\n")
					}
				}
			}
		}

		cpuActions := o.GetCpuActions()
		if len(cpuActions) > 0 {
			b.WriteString("----------\n")
			b.WriteString(color.Bold(i18n.T("omaha.cpuActionsHeader")) + "\n")
			for _, action := range cpuActions {
				b.WriteString(i18n.Tf("omaha.cpuActionLine", "name", cuiPlayerName(o.GetPlayer(action.PlayerIdx), action.PlayerIdx), "action", cuiBettingActionName(action.Action)))
				if action.Amount > 0 {
					b.WriteString(i18n.Tf("omaha.cpuActionAmount", "amount", strconv.Itoa(action.Amount)))
				}
				b.WriteString("\n")
			}
		}

		results := o.GetRoundResults()
		if len(results) > 0 && (o.GetPhase() == domain.OmahaPhaseEnd || o.GetPhase() == domain.OmahaPhaseShowdown) {
			b.WriteString("==========\n")
			b.WriteString(color.Bold(i18n.T("omaha.resultsHeader")) + "\n")
			// 印の凡例は1度だけ。プレイヤーごとに繰り返すと、4人ショーダウンでは
			// 同じ注記が4回並ぶ。
			for _, r := range results {
				if !r.Mucked && len(r.BestHand) > 0 {
					b.WriteString(i18n.T("omaha.resultBestLegend") + "\n")
					break
				}
			}
			for _, r := range results {
				name := cuiPlayerName(o.GetPlayer(r.PlayerIdx), r.PlayerIdx)
				kickers := ""
				if ks := domain.FormatKickers(r.Kickers); ks != "" {
					kickers = i18n.Tf("omaha.resultKickers", "kickers", ks)
				}
				switch {
				case r.Mucked:
					b.WriteString(i18n.Tf("omaha.resultMucked", "name", name))
				case r.HandName != "":
					b.WriteString(i18n.Tf("omaha.resultHand", "name", name, "hand", cuiPokerHandName(r.HandRank), "kickers", kickers))
				default:
					b.WriteString(i18n.Tf("omaha.resultPlayerOnly", "name", name))
				}
				// **どの2枚を手札から使ったのかは印が無いと追えない。** オマハ系は
				// 手札から使う枚数が固定で (通常4枚のうち2枚、Big O は5枚のうち2枚)、
				// Web は cardUsed/cardUnused ラベルでそれを見せている。CUI は役名と
				// キッカーだけで、RoundResult.BestHand を使っていなかった (#5484)。
				//
				// マックした手は見せない。BestHand が空の結果 (フォールド勝ちなど) も
				// 出さない -- 空行を出すと「役が無かった」ように読める。
				if !r.Mucked && len(r.BestHand) > 0 {
					b.WriteString(i18n.Tf("omaha.resultBest",
						"cards", omahaBestHandStr(r.BestHand, o.GetPlayer(r.PlayerIdx))))
				}
				if o.GetIsHiLo() && r.LowQualifies {
					b.WriteString(i18n.Tf("omaha.resultLow", "cards", cuiCardSliceStrEmoji(r.LowBestHand)))
				}
				if r.WonAmount > 0 {
					total := strconv.Itoa(r.WonAmount)
					switch {
					case o.GetIsHiLo() && r.HiWonAmount > 0 && r.LowWonAmount > 0:
						b.WriteString(i18n.Tf("omaha.wonHiLoBoth",
							"total", total,
							"hi", strconv.Itoa(r.HiWonAmount),
							"lo", strconv.Itoa(r.LowWonAmount)))
					case o.GetIsHiLo() && r.LowWonAmount > 0:
						b.WriteString(i18n.Tf("omaha.wonLoOnly", "total", total))
					case o.GetIsHiLo() && r.HiWonAmount > 0:
						b.WriteString(i18n.Tf("omaha.wonHiOnly", "total", total))
					default:
						b.WriteString(i18n.Tf("omaha.wonAmount", "total", total))
					}
				}
				b.WriteString("\n")
			}
		}

		if o.IsMuckAvailable() {
			b.WriteString("----------\n")
			b.WriteString(i18n.T("omaha.muckPrompt") + "\n")
		}

		if o.GetPhase() == domain.OmahaPhaseRebuy {
			b.WriteString("----------\n")
			switch o.GetRebuyPhaseType() {
			case domain.OmahaRebuyPhaseRebuy:
				rebuyCounts := o.GetRebuyCounts()
				humanIdx := -1
				for i := 0; i < o.GetPlayerCnt(); i++ {
					if o.GetPlayer(i).GetIsHuman() {
						humanIdx = i
						break
					}
				}
				if humanIdx >= 0 {
					b.WriteString(i18n.Tf("omaha.rebuyPrompt",
						"chips", strconv.Itoa(cfg.RebuyChips),
						"used", strconv.Itoa(rebuyCounts[humanIdx]),
						"max", strconv.Itoa(cfg.RebuyMaxCount)) + "\n")
				}
			case domain.OmahaRebuyPhaseAddon:
				b.WriteString(i18n.Tf("omaha.addonPrompt", "chips", strconv.Itoa(cfg.AddonChips)) + "\n")
			}
		}

		cuiErrorBlock(b, lastErr)

		if o.GetGameEndFlag() {
			b.WriteString(i18n.T("omaha.gameEnd") + "\n")
		}
	})
}

// omahaBestHandStr renders the winning five cards, suffixing the ones that came
// from the player's own hole with CuiHoleMark.
//
// player が nil のときは印だけ落として5枚を並べる -- 印が付かないほうが、
// 全部が手札由来に見えるより正直。
func omahaBestHandStr(best []*domain.Card, player *domain.OmahaPlayer) string {
	hole := make(map[string]bool)
	if player != nil {
		for i := 0; i < player.GetCardsSize(); i++ {
			if c := player.GetCard(i); c != nil {
				hole[omahaCardKey(c)] = true
			}
		}
	}
	parts := make([]string, 0, len(best))
	for _, c := range best {
		s := cuiCardStrEmoji(c)
		if c != nil && hole[omahaCardKey(c)] {
			s += CuiHoleMark
		}
		parts = append(parts, s)
	}
	return strings.Join(parts, "  ")
}

// omahaCardKey identifies a card within a single deck (suit+rank is unique).
func omahaCardKey(c *domain.Card) string {
	return strconv.Itoa(c.GetDesign()) + ":" + strconv.Itoa(c.GetValue())
}
