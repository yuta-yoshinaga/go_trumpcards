//go:build !js || !wasm || casino

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// FiveCardStudCuiPresenter renders the Five Card Stud CUI view.
type FiveCardStudCuiPresenter struct{}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *FiveCardStudCuiPresenter) ActionLogOutput(s interfaces.FiveCardStudGame) string {
	return actionLogOutputTextForSeats[*domain.FiveCardStudPlayer](s)
}

// Output renders the current game state for the active locale.
func (p *FiveCardStudCuiPresenter) Output(s interfaces.FiveCardStudGame, lastErr error) string {
	return buildCuiOutput(i18n.T("fivecardstud.outputTitle"), func(b *strings.Builder) {
		cfg := s.GetConfig()
		if cfg.TournamentMode {
			b.WriteString(i18n.Tf("fivecardstud.tournamentLine",
				"hand", strconv.Itoa(s.GetHandCount()),
				"ante", strconv.Itoa(cfg.Ante),
				"bringIn", strconv.Itoa(cfg.BringIn),
				"levelup", strconv.Itoa(cfg.AnteLevelHands)) + "\n")
			if cfg.RebuyEnabled {
				b.WriteString(i18n.Tf("fivecardstud.rebuyLine",
					"chips", strconv.Itoa(cfg.RebuyChips),
					"max", strconv.Itoa(cfg.RebuyMaxCount),
					"period", strconv.Itoa(cfg.RebuyPeriodHands)) + "\n")
			}
			if cfg.AddonEnabled {
				b.WriteString(i18n.Tf("fivecardstud.addonLine",
					"chips", strconv.Itoa(cfg.AddonChips),
					"after", strconv.Itoa(cfg.AddonAfterHand)) + "\n")
			}
		}

		b.WriteString(i18n.Tf("fivecardstud.tableMax", "n", strconv.Itoa(s.GetPlayerCnt())) + "\n")
		b.WriteString(i18n.Tf("fivecardstud.dealerLine", "idx", strconv.Itoa(s.GetDealerIdx())) + "\n")

		b.WriteString(i18n.Tf("fivecardstud.anteLine",
			"ante", strconv.Itoa(cfg.Ante),
			"bringIn", strconv.Itoa(cfg.BringIn),
			"small", strconv.Itoa(cfg.SmallBet),
			"big", strconv.Itoa(cfg.BigBet)) + "\n")

		b.WriteString(i18n.Tf("fivecardstud.potLine", "pot", strconv.Itoa(s.GetPot())) + "\n")

		if int(cfg.BettingLimit) < len(domain.BettingLimitNames) {
			b.WriteString(i18n.Tf("fivecardstud.limitLine", "name", domain.BettingLimitNames[cfg.BettingLimit]) + "\n")
		}

		// **Soko の役順位は標準ポーカーと違う。**チュートリアルを閉じると
		// 確認する手段が無くなるので、対局中いつでも読めるよう常設で出す (#5737)。
		if s.GetIsSoko() {
			b.WriteString(sokoRankingLine() + "\n")
		}

		b.WriteString("----------\n")
		for i := 0; i < s.GetPlayerCnt(); i++ {
			player := s.GetPlayer(i)
			b.WriteString(cuiPlayerNameWithStyle(player, i))
			b.WriteString(i18n.Tf("fivecardstud.playerChips", "chips", strconv.Itoa(player.GetChips())))

			if player.GetTotalHands() > 0 {
				b.WriteString(i18n.Tf("fivecardstud.playerStats",
					"vpip", strconv.Itoa(player.GetVPIP()),
					"pfr", strconv.Itoa(player.GetPFR()),
					"tb", strconv.Itoa(player.GetThreeBet()),
					"af", player.GetAFDisplay()))
			}

			if player.GetFolded() {
				b.WriteString(color.BoldYellow(i18n.T("fivecardstud.playerFolded")))
			} else if player.GetAllIn() {
				b.WriteString(color.BoldYellow(i18n.T("fivecardstud.playerAllIn")))
			}

			if player.GetCurrentBet() > 0 {
				b.WriteString(i18n.Tf("fivecardstud.playerBet", "bet", strconv.Itoa(player.GetCurrentBet())))
			}
			b.WriteString("\n")

			// Door cards (face up — visible to all players)
			if doorCards := player.GetDoorCards(); len(doorCards) > 0 {
				// **ストリートごとに 1 枚ずつ増える。**末尾が今回公開された札
				// なので、そこだけ印を付ける (#5674)。Web は同じ 1 枚にリングを
				// 付けている。CUI は全部を平らに並べていたので、前回の出力と
				// 見比べないと何が増えたのか分からなかった。
				b.WriteString(i18n.Tf("fivecardstud.doorCards",
					"cards", cuiCardSliceStrEmojiNewest(doorCards)) + "\n")
			}

			// Hole cards (face down — only the human sees them)
			if player.GetIsHuman() && !player.GetFolded() {
				if holeCards := player.GetHoleCards(); len(holeCards) > 0 {
					b.WriteString(i18n.Tf("fivecardstud.holeCards", "cards", cuiCardSliceStrEmoji(holeCards)) + "\n")
				}
			}
		}

		cpuActions := s.GetCpuActions()
		if len(cpuActions) > 0 {
			b.WriteString("----------\n")
			b.WriteString(color.Bold(i18n.T("fivecardstud.cpuActionsHeader")) + "\n")
			for _, action := range cpuActions {
				b.WriteString(i18n.Tf("fivecardstud.cpuActionLine",
					"idx", strconv.Itoa(action.PlayerIdx),
					"action", cuiBettingActionName(action.Action)))
				if action.Amount > 0 {
					b.WriteString(i18n.Tf("fivecardstud.cpuActionAmount", "amount", strconv.Itoa(action.Amount)))
				}
				b.WriteString("\n")
			}
		}

		results := s.GetRoundResults()
		if len(results) > 0 && (s.GetPhase() == domain.FiveCardStudPhaseEnd || s.GetPhase() == domain.FiveCardStudPhaseShowdown) {
			b.WriteString("==========\n")
			b.WriteString(color.Bold(i18n.T("fivecardstud.resultsHeader")) + "\n")
			for _, r := range results {
				name := cuiPlayerName(s.GetPlayer(r.PlayerIdx), r.PlayerIdx)
				kickers := ""
				if ks := domain.FormatKickers(r.Kickers); ks != "" {
					kickers = i18n.Tf("fivecardstud.resultKickers", "kickers", ks)
				}
				switch {
				case r.Mucked:
					b.WriteString(i18n.Tf("fivecardstud.resultMucked", "name", name))
				case r.HandName != "":
					b.WriteString(i18n.Tf("fivecardstud.resultHand",
						"name", name,
						"hand", cuiHandNameForStud(s, r.HandRank),
						"kickers", kickers))
				default:
					b.WriteString(i18n.Tf("fivecardstud.resultName", "name", name))
				}
				if r.WonAmount > 0 {
					b.WriteString(i18n.Tf("fivecardstud.wonAmount", "total", strconv.Itoa(r.WonAmount)))
				}
				b.WriteString("\n")
			}
		}

		if s.IsMuckAvailable() {
			b.WriteString("----------\n")
			b.WriteString(i18n.T("fivecardstud.muckPrompt") + "\n")
		}

		if s.GetPhase() == domain.FiveCardStudPhaseRebuy {
			b.WriteString("----------\n")
			switch s.GetRebuyPhaseType() {
			case domain.FiveCardStudRebuyPhaseRebuy:
				rebuyCounts := s.GetRebuyCounts()
				humanIdx := -1
				for i := 0; i < s.GetPlayerCnt(); i++ {
					if s.GetPlayer(i).GetIsHuman() {
						humanIdx = i
						break
					}
				}
				if humanIdx >= 0 {
					b.WriteString(i18n.Tf("fivecardstud.rebuyPrompt",
						"chips", strconv.Itoa(cfg.RebuyChips),
						"used", strconv.Itoa(rebuyCounts[humanIdx]),
						"max", strconv.Itoa(cfg.RebuyMaxCount)) + "\n")
				}
			case domain.FiveCardStudRebuyPhaseAddon:
				b.WriteString(i18n.Tf("fivecardstud.addonPrompt", "chips", strconv.Itoa(cfg.AddonChips)) + "\n")
			}
		}

		cuiErrorBlock(b, lastErr)

		if s.GetGameEndFlag() {
			b.WriteString(i18n.T("fivecardstud.gameEnd") + "\n")
		}
	})
}

// sokoRankingLine は Soko の役順位を強い順に 1 行で返す。
//
// **順序も名前も表を写さない。**役名は showdown と同じ sokoHandRank<N> を引き、
// 並びは domain の定数そのものを強い順にたどる。4 枚役だけは標準ポーカーとの
// 違いなので目印を付ける。
func sokoRankingLine() string {
	parts := make([]string, 0, domain.SokoHandRoyalFlush+1)
	for rank := domain.SokoHandRoyalFlush; rank >= domain.SokoHandHighCard; rank-- {
		name := i18n.T("sokoHandRank" + strconv.Itoa(rank))
		if rank == domain.SokoHandFourStraight || rank == domain.SokoHandFourFlush {
			name = i18n.Tf("fivecardstud.sokoRankingInserted", "name", name)
		}
		parts = append(parts, name)
	}
	return i18n.Tf("fivecardstud.sokoRanking", "list", strings.Join(parts, " > "))
}

// cuiHandNameForStud はファイブカード・スタッド系の役名をローカライズして返す。
//
// **ランクのスケールがゲームによって違うので、表も切り替える。** Soko
// (Canadian Stud) は4枚ストレート/4枚フラッシュをワンペアとツーペアの間に
// 挿入した独自スケールなので、標準の `pokerHandRank<N>` を引くと別の役名が出る
// （Soko の 4 はツーペアだが、標準の 4 はストレート）。
//
// **この関数が casino タグ側にあるのは意図的。** 無タグの cui_card_helper.go に
// 置くと、そこは6 worker すべてにコンパイルされるのに interfaces.FiveCardStudGame
// と domain.SokoHand* は casino 限定なので、casino 以外の5 worker が
// `undefined:` で落ちる（実際に落とした）。
func cuiHandNameForStud(g interfaces.FiveCardStudGame, rank int) string {
	if g.GetIsSoko() {
		if rank < 0 || rank > domain.SokoHandRoyalFlush {
			return ""
		}
		return i18n.T("sokoHandRank" + strconv.Itoa(rank))
	}
	return cuiPokerHandName(rank)
}
