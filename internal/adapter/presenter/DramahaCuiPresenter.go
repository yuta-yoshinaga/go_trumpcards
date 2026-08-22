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

// DramahaCuiPresenter renders the Dramaha Hold'em CUI view.
type DramahaCuiPresenter struct{}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *DramahaCuiPresenter) ActionLogOutput(o interfaces.DramahaGame) string {
	return actionLogOutputTextForSeats[*domain.DramahaPlayer](o)
}

// dramahaTitleKey は、ホールカード枚数 (4=ドラマハ, 5=Big O) と Hi-Lo フラグから
// CUI ヘッダーに使う i18n タイトルキーを選択する。

// Output renders the current game state for the active locale (#1699).
func (p *DramahaCuiPresenter) Output(o interfaces.DramahaGame, lastErr error) string {
	// **変種は無い。** ドラマハは常に 5 枚配りで常に二分するので、
	// クローン元のような「Big O か」「Hi-Lo か」の題名の出し分けは要らない。
	return buildCuiOutput(i18n.T("dramaha.helpTitle"), func(b *strings.Builder) {
		// Dramaha's defining pitfall: exactly two hole cards must be used. Surface
		// it every render (the count adapts for Big O's five hole cards).
		b.WriteString(i18n.Tf("dramaha.mandatoryRuleLine",
			"hole", strconv.Itoa(o.GetHoleCardCount())) + "\n")

		// **ポットが二分されることと、その二つの評価軸**は毎回書く。同じ 5 枚を
		// 「ボードと組み合わせたオマハ役」と「そのままのドロー役」の両方に使う
		// のがこのゲームの発想で、題名だけでは伝わらない。
		b.WriteString(i18n.T("dramaha.splitRuleLine") + "\n")

		// **ドローできることを画面に出す。** フェーズ名を出さないと、
		// 交換できる一瞬に何も促されないまま通り過ぎる。
		if o.GetPhase() == domain.DramahaPhaseDraw {
			b.WriteString(i18n.T("dramaha.drawPhaseLine") + "\n")
		}

		cfg := o.GetConfig()
		if cfg.TournamentMode {
			b.WriteString(i18n.Tf("dramaha.tournamentLine",
				"hand", strconv.Itoa(o.GetHandCount()),
				"sb", strconv.Itoa(cfg.SmallBlind),
				"bb", strconv.Itoa(cfg.BigBlind),
				"levelup", strconv.Itoa(cfg.BlindLevelHands)) + "\n")
			if cfg.RebuyEnabled {
				b.WriteString(i18n.Tf("dramaha.rebuyLine",
					"chips", strconv.Itoa(cfg.RebuyChips),
					"max", strconv.Itoa(cfg.RebuyMaxCount),
					"period", strconv.Itoa(cfg.RebuyPeriodHands)) + "\n")
			}
			if cfg.AddonEnabled {
				b.WriteString(i18n.Tf("dramaha.addonLine",
					"chips", strconv.Itoa(cfg.AddonChips),
					"after", strconv.Itoa(cfg.AddonAfterHand)) + "\n")
			}
		}

		b.WriteString(i18n.Tf("dramaha.tableMax", "n", strconv.Itoa(o.GetPlayerCnt())) + "\n")
		b.WriteString(i18n.Tf("dramaha.dealerLine", "idx", strconv.Itoa(o.GetDealerIdx())) + "\n")

		cc := o.GetCommunityCards()
		if len(cc) == 0 {
			b.WriteString(i18n.T("dramaha.communityNone") + "\n")
		} else {
			b.WriteString(i18n.Tf("dramaha.communityCards", "cards", cuiCardSliceStrEmoji(cc)) + "\n")
		}

		// **ロー成立の見通しはボードだけで決まる。** Web は BoardLowBadge で
		// ベッティング中ずっと出しているのに、CUI はショーダウンの resultLow まで
		// **ボードの見通し行は落とした。** クローン元 (Hi-Lo) はボードに低い札が
		// 残っているかを助言していたが、ドロー側はボードを一切使わないので、
		// 同じ助言はここでは何の意味も持たない。

		b.WriteString(i18n.Tf("dramaha.potLine", "pot", strconv.Itoa(o.GetPot())) + "\n")

		if int(cfg.BettingLimit) < len(domain.BettingLimitNames) {
			b.WriteString(i18n.Tf("dramaha.limitLine", "name", domain.BettingLimitNames[cfg.BettingLimit]) + "\n")
		}

		b.WriteString("----------\n")
		for i := 0; i < o.GetPlayerCnt(); i++ {
			player := o.GetPlayer(i)
			b.WriteString(cuiPlayerNameWithStyle(player, i))
			b.WriteString(i18n.Tf("dramaha.playerChips", "chips", strconv.Itoa(player.GetChips())))

			if player.GetTotalHands() > 0 {
				b.WriteString(i18n.Tf("dramaha.playerStats",
					"vpip", strconv.Itoa(player.GetVPIP()),
					"pfr", strconv.Itoa(player.GetPFR()),
					"tb", strconv.Itoa(player.GetThreeBet()),
					"af", player.GetAFDisplay()))
			}

			if player.GetFolded() {
				b.WriteString(" " + color.BoldYellow(i18n.T("dramaha.playerFolded")))
			} else if player.GetAllIn() {
				b.WriteString(" " + color.BoldYellow(i18n.T("dramaha.playerAllIn")))
			}

			if player.GetCurrentBet() > 0 {
				b.WriteString(i18n.Tf("dramaha.playerBet", "bet", strconv.Itoa(player.GetCurrentBet())))
			}
			b.WriteString("\n")

			if player.GetIsHuman() && !player.GetFolded() {
				b.WriteString(i18n.Tf("dramaha.humanHand", "cards", cuiCardListStrEmoji(player)) + "\n")
				// **Web は dramaha-live-besthand で暫定ベストを常時出しているのに、
				// CUI は「2枚使用」の注意書きだけで実際の役を出していなかった
				// (#4680)。**手札4枚から必ず2枚という特殊ルールがあるぶん、
				// 暫定表示はミスを防ぐ補助になる。PeekBestHand は状態を変えない。
				if rank, best := player.PeekBestHand(cc); len(best) > 0 {
					b.WriteString(i18n.Tf("dramaha.currentBestHand",
						"hand", cuiPokerHandName(rank),
						"cards", cuiCardSliceStrEmoji(best)) + "\n")
				}
			}
		}

		// **学習モードの値は Web にしか出ていなかった (#5482)。** GetEquity /
		// GetPotOdds は共有ヘルパ経由で Web へ送られ、Holdem 系の CUI も出して
		// いるのに、Dramaha の CUI だけが取り残されていた。GetEquity は人間が
		// 降りている局面などで nil を返し、そのとき表示は消える。
		if o.IsHumanTurn() {
			if eq := o.GetEquity(); eq != nil {
				potOdds := o.GetPotOdds()
				b.WriteString("----------\n")
				b.WriteString(color.Bold(i18n.T("dramaha.learningHeader")) + "\n")
				b.WriteString(i18n.Tf("dramaha.learningLine",
					"equity", fmt.Sprintf("%.1f", eq.Equity*100),
					"potodds", fmt.Sprintf("%.1f", potOdds)) + "\n")
				if potOdds > 0 {
					if eq.Equity*100 > potOdds {
						b.WriteString(i18n.T("dramaha.learningEvPlus") + "\n")
					} else {
						b.WriteString(i18n.T("dramaha.learningEvMinus") + "\n")
					}
				}
			}
		}

		cpuActions := o.GetCpuActions()
		if len(cpuActions) > 0 {
			b.WriteString("----------\n")
			b.WriteString(color.Bold(i18n.T("dramaha.cpuActionsHeader")) + "\n")
			for _, action := range cpuActions {
				b.WriteString(i18n.Tf("dramaha.cpuActionLine", "name", cuiPlayerName(o.GetPlayer(action.PlayerIdx), action.PlayerIdx), "action", cuiBettingActionName(action.Action)))
				if action.Amount > 0 {
					b.WriteString(i18n.Tf("dramaha.cpuActionAmount", "amount", strconv.Itoa(action.Amount)))
				}
				b.WriteString("\n")
			}
		}

		results := o.GetRoundResults()
		if len(results) > 0 && (o.GetPhase() == domain.DramahaPhaseEnd || o.GetPhase() == domain.DramahaPhaseShowdown) {
			b.WriteString("==========\n")
			b.WriteString(color.Bold(i18n.T("dramaha.resultsHeader")) + "\n")
			// 印の凡例は1度だけ。プレイヤーごとに繰り返すと、4人ショーダウンでは
			// 同じ注記が4回並ぶ。
			for _, r := range results {
				if !r.Mucked && len(r.BestHand) > 0 {
					b.WriteString(i18n.T("dramaha.resultBestLegend") + "\n")
					break
				}
			}
			for _, r := range results {
				name := cuiPlayerName(o.GetPlayer(r.PlayerIdx), r.PlayerIdx)
				kickers := ""
				if ks := domain.FormatKickers(r.Kickers); ks != "" {
					kickers = i18n.Tf("dramaha.resultKickers", "kickers", ks)
				}
				switch {
				case r.Mucked:
					b.WriteString(i18n.Tf("dramaha.resultMucked", "name", name))
				case r.HandName != "":
					b.WriteString(i18n.Tf("dramaha.resultHand", "name", name, "hand", cuiPokerHandName(r.HandRank), "kickers", kickers))
				default:
					b.WriteString(i18n.Tf("dramaha.resultPlayerOnly", "name", name))
				}
				// **どの2枚を手札から使ったのかは印が無いと追えない。** ドラマハ系は
				// 手札から使う枚数が固定で (通常4枚のうち2枚、Big O は5枚のうち2枚)、
				// Web は cardUsed/cardUnused ラベルでそれを見せている。CUI は役名と
				// キッカーだけで、RoundResult.BestHand を使っていなかった (#5484)。
				//
				// マックした手は見せない。BestHand が空の結果 (フォールド勝ちなど) も
				// 出さない -- 空行を出すと「役が無かった」ように読める。
				if !r.Mucked && len(r.BestHand) > 0 {
					b.WriteString(i18n.Tf("dramaha.resultBest",
						"cards", dramahaBestHandStr(r.BestHand, o.GetPlayer(r.PlayerIdx))))
				}
				// **ドロー側は必ず成立する。** クローン元は「ローが成立した場合
				// のみ」だったが、5 枚あればどんな手でも役が付く。
				if len(r.LowBestHand) > 0 {
					b.WriteString(i18n.Tf("dramaha.resultDraw", "cards", cuiCardSliceStrEmoji(r.LowBestHand)))
				}
				if r.WonAmount > 0 {
					total := strconv.Itoa(r.WonAmount)
					switch {
					case true /* ドラマハは常に二分する */ && r.HiWonAmount > 0 && r.LowWonAmount > 0:
						// **両取りは特別な結果。** Web は専用バッジで強調している
						// のに、CUI は金額の内訳を出すだけでそうとは言っていなかった
						// (#5485)。内訳は残す -- 置き換えると情報が減る。
						b.WriteString(i18n.Tf("dramaha.wonHiLoBoth",
							"total", total,
							"hi", strconv.Itoa(r.HiWonAmount),
							"lo", strconv.Itoa(r.LowWonAmount)))
						b.WriteString(" " + color.BoldYellow(i18n.T("dramaha.scoop")))
					case true /* ドラマハは常に二分する */ && r.LowWonAmount > 0:
						b.WriteString(i18n.Tf("dramaha.wonLoOnly", "total", total))
					case true /* ドラマハは常に二分する */ && r.HiWonAmount > 0:
						b.WriteString(i18n.Tf("dramaha.wonHiOnly", "total", total))
					default:
						b.WriteString(i18n.Tf("dramaha.wonAmount", "total", total))
					}
				}
				b.WriteString("\n")
			}
		}

		if o.IsMuckAvailable() {
			b.WriteString("----------\n")
			b.WriteString(i18n.T("dramaha.muckPrompt") + "\n")
		}

		if o.GetPhase() == domain.DramahaPhaseRebuy {
			b.WriteString("----------\n")
			switch o.GetRebuyPhaseType() {
			case domain.DramahaRebuyPhaseRebuy:
				rebuyCounts := o.GetRebuyCounts()
				humanIdx := -1
				for i := 0; i < o.GetPlayerCnt(); i++ {
					if o.GetPlayer(i).GetIsHuman() {
						humanIdx = i
						break
					}
				}
				if humanIdx >= 0 {
					b.WriteString(i18n.Tf("dramaha.rebuyPrompt",
						"chips", strconv.Itoa(cfg.RebuyChips),
						"used", strconv.Itoa(rebuyCounts[humanIdx]),
						"max", strconv.Itoa(cfg.RebuyMaxCount)) + "\n")
				}
			case domain.DramahaRebuyPhaseAddon:
				b.WriteString(i18n.Tf("dramaha.addonPrompt", "chips", strconv.Itoa(cfg.AddonChips)) + "\n")
			}
		}

		cuiErrorBlock(b, lastErr)

		if o.GetGameEndFlag() {
			b.WriteString(i18n.T("dramaha.gameEnd") + "\n")
		}
	})
}

// dramahaBestHandStr renders the winning five cards, suffixing the ones that came
// from the player's own hole with CuiHoleMark.
//
// player が nil のときは印だけ落として5枚を並べる -- 印が付かないほうが、
// 全部が手札由来に見えるより正直。
func dramahaBestHandStr(best []*domain.Card, player *domain.DramahaPlayer) string {
	hole := make(map[string]bool)
	if player != nil {
		for i := 0; i < player.GetCardsSize(); i++ {
			if c := player.GetCard(i); c != nil {
				hole[dramahaCardKey(c)] = true
			}
		}
	}
	parts := make([]string, 0, len(best))
	for _, c := range best {
		s := cuiCardStrEmoji(c)
		if c != nil && hole[dramahaCardKey(c)] {
			s += CuiHoleMark
		}
		parts = append(parts, s)
	}
	return strings.Join(parts, "  ")
}

// dramahaCardKey identifies a card within a single deck (suit+rank is unique).
func dramahaCardKey(c *domain.Card) string {
	return strconv.Itoa(c.GetDesign()) + ":" + strconv.Itoa(c.GetValue())
}
