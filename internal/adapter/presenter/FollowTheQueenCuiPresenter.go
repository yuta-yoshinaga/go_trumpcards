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

// FollowTheQueenCuiPresenter renders the Follow the Queen CUI view.
type FollowTheQueenCuiPresenter struct{}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *FollowTheQueenCuiPresenter) ActionLogOutput(s interfaces.FollowTheQueenGame) string {
	return actionLogOutputTextForSeats[*domain.FollowTheQueenPlayer](s)
}

// Output renders the current game state for the active locale (#1699).
func (p *FollowTheQueenCuiPresenter) Output(s interfaces.FollowTheQueenGame, lastErr error) string {
	return buildCuiOutput(i18n.T("followthequeen.outputTitle"), func(b *strings.Builder) {
		cfg := s.GetConfig()
		if cfg.TournamentMode {
			b.WriteString(i18n.Tf("followthequeen.tournamentLine",
				"hand", strconv.Itoa(s.GetHandCount()),
				"ante", strconv.Itoa(cfg.Ante),
				"bringIn", strconv.Itoa(cfg.BringIn),
				"levelup", strconv.Itoa(cfg.AnteLevelHands)) + "\n")
			if cfg.RebuyEnabled {
				b.WriteString(i18n.Tf("followthequeen.rebuyLine",
					"chips", strconv.Itoa(cfg.RebuyChips),
					"max", strconv.Itoa(cfg.RebuyMaxCount),
					"period", strconv.Itoa(cfg.RebuyPeriodHands)) + "\n")
			}
			if cfg.AddonEnabled {
				b.WriteString(i18n.Tf("followthequeen.addonLine",
					"chips", strconv.Itoa(cfg.AddonChips),
					"after", strconv.Itoa(cfg.AddonAfterHand)) + "\n")
			}
		}

		b.WriteString(i18n.Tf("followthequeen.tableMax", "n", strconv.Itoa(s.GetPlayerCnt())) + "\n")
		b.WriteString(i18n.Tf("followthequeen.dealerLine", "idx", strconv.Itoa(s.GetDealerIdx())) + "\n")

		// ブリングイン (強制ベットを払い、3rd street で最初に動く席)。Web は
		// バッジで示しているのに CUI には手掛かりが無かった (#5542)。
		if bi := s.GetBringInPlayerIdx(); bi >= 0 && s.GetPhase() == domain.FollowTheQueenPhaseThirdStreet {
			b.WriteString(i18n.Tf("followthequeen.bringInLine",
				"name", cuiPlayerName(s.GetPlayer(bi), bi)) + "\n")
		}

		b.WriteString(i18n.Tf("followthequeen.anteLine",
			"ante", strconv.Itoa(cfg.Ante),
			"bringIn", strconv.Itoa(cfg.BringIn),
			"small", strconv.Itoa(cfg.SmallBet),
			"big", strconv.Itoa(cfg.BigBet)) + "\n")

		// **いま何がワイルドかを必ず出す。** 表向きのクイーンが出た次のカードの
		// ランクが全員のワイルドになる、というのがこのゲームそのもので、
		// まだ出ていない間も「無い」と言い切らないと自分の役が読めない。
		if wr := s.GetWildRank(); wr != 0 {
			b.WriteString(i18n.Tf("followthequeen.wildLine", "rank", cuiRankName(wr)) + "\n")
		} else {
			b.WriteString(i18n.T("followthequeen.wildNone") + "\n")
		}

		b.WriteString(i18n.Tf("followthequeen.potLine", "pot", strconv.Itoa(s.GetPot())) + "\n")

		if int(cfg.BettingLimit) < len(domain.BettingLimitNames) {
			b.WriteString(i18n.Tf("followthequeen.limitLine", "name", domain.BettingLimitNames[cfg.BettingLimit]) + "\n")
		}

		b.WriteString("----------\n")
		for i := 0; i < s.GetPlayerCnt(); i++ {
			player := s.GetPlayer(i)
			b.WriteString(cuiPlayerNameWithStyle(player, i))
			b.WriteString(i18n.Tf("followthequeen.playerChips", "chips", strconv.Itoa(player.GetChips())))

			if player.GetTotalHands() > 0 {
				b.WriteString(i18n.Tf("followthequeen.playerStats",
					"vpip", strconv.Itoa(player.GetVPIP()),
					"pfr", strconv.Itoa(player.GetPFR()),
					"tb", strconv.Itoa(player.GetThreeBet()),
					"af", player.GetAFDisplay()))
			}

			if player.GetFolded() {
				b.WriteString(color.BoldYellow(i18n.T("followthequeen.playerFolded")))
			} else if player.GetAllIn() {
				b.WriteString(color.BoldYellow(i18n.T("followthequeen.playerAllIn")))
			}

			if player.GetCurrentBet() > 0 {
				b.WriteString(i18n.Tf("followthequeen.playerBet", "bet", strconv.Itoa(player.GetCurrentBet())))
			}
			b.WriteString("\n")

			// Door cards (face up — visible to all players)
			if doorCards := player.GetDoorCards(); len(doorCards) > 0 {
				b.WriteString(i18n.Tf("followthequeen.doorCards", "cards", cuiCardSliceStrEmoji(doorCards)) + "\n")
			}

			// Hole cards (face down — only the human sees them)
			if player.GetIsHuman() && !player.GetFolded() {
				if holeCards := player.GetHoleCards(); len(holeCards) > 0 {
					b.WriteString(i18n.Tf("followthequeen.holeCards", "cards", cuiCardSliceStrEmoji(holeCards)) + "\n")
				}
				if rank, best := player.PeekBestHand(); len(best) > 0 {
					// **Web は常時「いまの最善役」を出しているのに、ハイ戦の CUI は
					// ショーダウンまで役名を一切出していなかった (#4695)。**3rd〜7th
					// street の間ずっと自分の手が何に達しているか分からない。
					// PeekBestHand は状態を変えないので、描画から呼んで安全。
					b.WriteString(i18n.Tf("followthequeen.currentBestHand",
						"hand", cuiPokerHandName(rank),
						"cards", cuiCardSliceStrEmoji(best)) + "\n")
				}
			}
		}

		cpuActions := s.GetCpuActions()
		if len(cpuActions) > 0 {
			b.WriteString("----------\n")
			b.WriteString(color.Bold(i18n.T("followthequeen.cpuActionsHeader")) + "\n")
			for _, action := range cpuActions {
				b.WriteString(i18n.Tf("followthequeen.cpuActionLine",
					"idx", strconv.Itoa(action.PlayerIdx),
					"action", cuiBettingActionName(action.Action)))
				if action.Amount > 0 {
					b.WriteString(i18n.Tf("followthequeen.cpuActionAmount", "amount", strconv.Itoa(action.Amount)))
				}
				b.WriteString("\n")
			}
		}

		phase := s.GetPhase()
		isBettingStreet := phase >= domain.FollowTheQueenPhaseThirdStreet && phase <= domain.FollowTheQueenPhaseSeventhStreet
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
					b.WriteString(i18n.Tf("followthequeen.actionPromptCheck",
						"raise", strconv.Itoa(s.GetMinRaise())) + "\n")
				} else {
					b.WriteString(i18n.Tf("followthequeen.actionPrompt",
						"call", strconv.Itoa(toCall),
						"raise", strconv.Itoa(s.GetMinRaise())) + "\n")
				}
			}
		}

		results := s.GetRoundResults()
		if len(results) > 0 && (s.GetPhase() == domain.FollowTheQueenPhaseEnd || s.GetPhase() == domain.FollowTheQueenPhaseShowdown) {
			b.WriteString("==========\n")
			b.WriteString(color.Bold(i18n.T("followthequeen.resultsHeader")) + "\n")
			for _, r := range results {
				name := cuiPlayerName(s.GetPlayer(r.PlayerIdx), r.PlayerIdx)
				kickers := ""
				if ks := domain.FormatKickers(r.Kickers); ks != "" {
					kickers = i18n.Tf("followthequeen.resultKickers", "kickers", ks)
				}
				switch {
				case r.Mucked:
					b.WriteString(i18n.Tf("followthequeen.resultMucked", "name", name))
				case r.HandName != "":
					b.WriteString(i18n.Tf("followthequeen.resultHand",
						"name", name,
						"hand", cuiPokerHandName(r.HandRank),
						"kickers", kickers))
				default:
					b.WriteString(i18n.Tf("followthequeen.resultName", "name", name))
				}
				if r.WonAmount > 0 {
					b.WriteString(i18n.Tf("followthequeen.wonAmount", "total", strconv.Itoa(r.WonAmount)))
				}
				b.WriteString("\n")
			}
		}

		if s.IsMuckAvailable() {
			b.WriteString("----------\n")
			b.WriteString(i18n.T("followthequeen.muckPrompt") + "\n")
		}

		if s.GetPhase() == domain.FollowTheQueenPhaseRebuy {
			b.WriteString("----------\n")
			switch s.GetRebuyPhaseType() {
			case domain.FollowTheQueenRebuyPhaseRebuy:
				rebuyCounts := s.GetRebuyCounts()
				humanIdx := -1
				for i := 0; i < s.GetPlayerCnt(); i++ {
					if s.GetPlayer(i).GetIsHuman() {
						humanIdx = i
						break
					}
				}
				if humanIdx >= 0 {
					b.WriteString(i18n.Tf("followthequeen.rebuyPrompt",
						"chips", strconv.Itoa(cfg.RebuyChips),
						"used", strconv.Itoa(rebuyCounts[humanIdx]),
						"max", strconv.Itoa(cfg.RebuyMaxCount)) + "\n")
				}
			case domain.FollowTheQueenRebuyPhaseAddon:
				b.WriteString(i18n.Tf("followthequeen.addonPrompt", "chips", strconv.Itoa(cfg.AddonChips)) + "\n")
			}
		}

		cuiErrorBlock(b, lastErr)

		if s.GetGameEndFlag() {
			b.WriteString(i18n.T("followthequeen.gameEnd") + "\n")
		}
	})
}

// followthequeenThreeStraight reports whether the three distinct card values form
// a run (including the Q-K-A wheel where the Ace is high).
func followTheQueenThreeStraight(cards []*domain.Card) bool {
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

// followTheQueenThirdStreetAdvice returns whether to continue on third street and
// the i18n reason key, using classic starting-hand rules (any pair, three to a
// flush, three to a straight, or three high cards).
func followTheQueenThirdStreetAdvice(cards []*domain.Card, isWild func(*domain.Card) bool) (cont bool, reasonKey string) {
	if len(cards) < 3 {
		return false, "followthequeen.hintReasonFold"
	}

	// **ワイルドを持っているかが他のどの条件より大きい。** ワイルド 1 枚は
	// 常にペア以上を意味し、2 枚あればスリーカード以上が確定する。ペア判定を
	// 先に置くと「ワンペア」と言ってしまい、実際の手の強さを 2 段階見誤る。
	// 判定そのものはドメインの IsWild に委ねる (クイーン常時ワイルドを
	// ここで書き直さない)。
	wilds := 0
	for _, c := range cards {
		if isWild(c) {
			wilds++
		}
	}
	switch {
	case wilds >= 2:
		return true, "followthequeen.hintReasonWildTwo"
	case wilds == 1:
		return true, "followthequeen.hintReasonWildOne"
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
			return true, "followthequeen.hintReasonPair"
		}
	}
	for _, n := range suits {
		if n >= 3 {
			return true, "followthequeen.hintReasonFlush"
		}
	}
	if followTheQueenThreeStraight(cards) {
		return true, "followthequeen.hintReasonStraight"
	}
	if high >= 3 {
		return true, "followthequeen.hintReasonHigh"
	}
	return false, "followthequeen.hintReasonFold"
}

// HintOutput advises on the human's turn, using basic starting-hand strategy on
// third street.
func (p *FollowTheQueenCuiPresenter) HintOutput(s interfaces.FollowTheQueenGame) string {
	if !s.IsHumanTurn() {
		return i18n.T("followthequeen.hintNone") + "\n"
	}
	if s.GetPhase() != domain.FollowTheQueenPhaseThirdStreet {
		return i18n.T("followthequeen.hintNone") + "\n"
	}
	player := s.GetPlayer(s.GetCurrentTurn())
	if player == nil {
		return i18n.T("followthequeen.hintNone") + "\n"
	}
	cont, reasonKey := followTheQueenThirdStreetAdvice(player.GetAllCards(), s.IsWild)
	action := i18n.T("followthequeen.hintFold")
	if cont {
		action = i18n.T("followthequeen.hintContinue")
	}
	return color.Yellow(i18n.Tf("followthequeen.hint",
		"action", action, "reason", i18n.T(reasonKey))) + "\n"
}
