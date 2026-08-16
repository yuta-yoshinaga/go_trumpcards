package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// sevensPlayerStr returns the display string for a single Sevens player.
//
// playable は人間の手札のうち今出せるインデックス。nil は「判定していない」
// (人間の手番でない) で、空スライスは「1枚も出せない」— 意味が違うので
// 呼び出し側で分けて扱う。
func sevensPlayerStr(player *domain.SevensPlayer, i int, playable []int) string {
	var b strings.Builder
	b.WriteString(cuiPlayerName(player, i))
	if player.GetIsFinished() {
		b.WriteString(i18n.Tf("sevens.playerOut", "rank", strconv.Itoa(player.GetRank())))
		b.WriteString("\n")
	} else {
		count := strconv.Itoa(player.GetCardsSize())
		used := strconv.Itoa(player.GetPassesUsed())
		if player.GetMaxPasses() == 0 {
			b.WriteString(i18n.Tf("sevens.playerActiveUnlimited", "count", count, "used", used))
		} else {
			b.WriteString(i18n.Tf("sevens.playerActiveLimited",
				"count", count,
				"used", used,
				"max", strconv.Itoa(player.GetMaxPasses())))
		}
		b.WriteString("\n")
		if player.GetIsHuman() {
			b.WriteString(cuiPlayableMarkedCardListStr(player, playable))
			b.WriteString("\n")
			// **空スライスを無印のまま出すと「判定していない」と区別が付かない。**
			// 7並べでは出せる札が1枚も無い局面が普通に起きるので、明示的に言う。
			if playable != nil && len(playable) == 0 && player.GetCardsSize() > 0 {
				b.WriteString(i18n.T("sevens.noPlayable") + "\n")
			}
		}
	}
	return b.String()
}

// sevensActionStr returns the display string for a Sevens action.
func sevensActionStr(playerName string, action *domain.SevensCpuAction) string {
	if action.PlayedCard == nil {
		if action.ForcedPass {
			return i18n.Tf("sevens.actionPassedNoCard", "name", playerName) + "\n"
		}
		return i18n.Tf("sevens.actionPassed", "name", playerName) + "\n"
	}
	if action.PlayedCard.GetDesign() == domain.CardDesignJoker && action.TargetSuit > 0 {
		return i18n.Tf("sevens.actionPlayedJoker",
			"name", playerName,
			"joker", cuiCardStr(action.PlayedCard),
			"suit", cuiSuitName(action.TargetSuit),
			"value", strconv.Itoa(action.TargetValue)) + "\n"
	}
	return i18n.Tf("sevens.actionPlayed",
		"name", playerName,
		"card", cuiCardStr(action.PlayedCard)) + "\n"
}

// sevensRuleLabels appends the active rule badges to b. The conditional
// list mirrors the original guard expression so the "Rules:" header is
// only emitted when at least one non-default rule is in play.
func sevensRuleLabels(b *strings.Builder, config *domain.SevensConfig) {
	hasNonDefault := config.TunnelEnabled ||
		config.TunnelSkipWidth >= 2 ||
		config.JokerCount > 0 ||
		config.CpuStrategy != domain.SevensCpuSimple ||
		config.MaxPasses != domain.SevensMaxPasses ||
		config.NoJokerFinish ||
		config.JokerReclaimEnabled ||
		config.EndStopEnabled ||
		config.JokerConsecutiveBanned
	if !hasNonDefault {
		return
	}
	b.WriteString(i18n.T("sevens.ruleLabel"))
	if config.TunnelEnabled {
		b.WriteString(" " + color.Yellow(i18n.T("sevens.ruleTunnel")))
	}
	if config.TunnelSkipWidth >= 2 {
		b.WriteString(" " + color.Yellow(i18n.Tf("sevens.ruleTunnelSkip", "width", strconv.Itoa(config.TunnelSkipWidth))))
	}
	if config.JokerCount > 0 {
		b.WriteString(" " + color.Yellow(i18n.Tf("sevens.ruleJokers", "count", strconv.Itoa(config.JokerCount))))
	}
	switch config.CpuStrategy {
	case domain.SevensCpuStrategic:
		b.WriteString(" " + color.Yellow(i18n.T("sevens.ruleCpuStrategic")))
	case domain.SevensCpuHarassment:
		b.WriteString(" " + color.Yellow(i18n.T("sevens.ruleCpuHarassment")))
	}
	if config.MaxPasses == 0 {
		b.WriteString(" " + color.Yellow(i18n.T("sevens.ruleUnlimitedPasses")))
	} else if config.MaxPasses != domain.SevensMaxPasses {
		b.WriteString(" " + color.Yellow(i18n.Tf("sevens.rulePassLimit", "count", strconv.Itoa(config.MaxPasses))))
	}
	if config.NoJokerFinish {
		b.WriteString(" " + color.Yellow(i18n.T("sevens.ruleNoJokerFinish")))
	}
	if config.JokerReclaimEnabled {
		b.WriteString(" " + color.Yellow(i18n.T("sevens.ruleJokerReclaim")))
	}
	if config.EndStopEnabled {
		b.WriteString(" " + color.Yellow(i18n.T("sevens.ruleEndStop")))
	}
	if config.JokerConsecutiveBanned {
		b.WriteString(" " + color.Yellow(i18n.T("sevens.ruleNoConsecutiveJokers")))
	}
	b.WriteString("\n")
}

// SevensCuiPresenter renders the Sevens CUI view.
type SevensCuiPresenter struct{}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *SevensCuiPresenter) ActionLogOutput(s interfaces.SevensGame) string {
	return actionLogOutputText(s)
}

// Output renders the current game state for the active locale (#1699).
func (p *SevensCuiPresenter) Output(s interfaces.SevensGame, lastErr error) string {
	config := s.GetConfig()

	// The original output put rule badges between the title and the mid-divider
	// (inside the header block). buildCuiOutput appends "\n" after the title, so
	// we splice the rules in right after the title and trim the trailing "\n"
	// that sevensRuleLabels adds, to avoid a blank line before the mid-divider.
	title := i18n.T("sevens.helpTitle")
	var rb strings.Builder
	sevensRuleLabels(&rb, &config)
	if rules := strings.TrimRight(rb.String(), "\n"); rules != "" {
		title = title + "\n" + rules
	}

	return buildCuiOutput(title, func(b *strings.Builder) {
		playable := s.GetPlayableCardIndices()
		for i := 0; i < s.GetPlayerCnt(); i++ {
			b.WriteString(sevensPlayerStr(s.GetPlayer(i), i, playable))
		}

		b.WriteString("----------\n")

		// Board state: render each suit's 1..13 positions individually (from the
		// placed bitmask) so joker substitutions and skip placements show
		// exactly which cards are down, unlike the old min~max range.
		b.WriteString(i18n.T("sevens.boardLabel") + "\n")
		suits := []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond}
		placed := s.GetTablePlaced()
		for _, suit := range suits {
			cells := make([]string, 13)
			for v := 1; v <= 13; v++ {
				if placed[suit]&(1<<uint(v)) != 0 {
					cells[v-1] = cuiRankLabel(v)
				} else {
					cells[v-1] = "_"
				}
			}
			b.WriteString(i18n.Tf("sevens.boardSuitLine",
				"suit", cuiSuitName(suit),
				"cells", strings.Join(cells, " ")) + "\n")
		}

		// Human's previous action
		humanAction := s.GetHumanAction()
		if humanAction != nil {
			b.WriteString(sevensActionStr(cuiPlayerName(s.GetPlayer(humanAction.PlayerIdx), humanAction.PlayerIdx), humanAction))
		}

		// CPU action history
		cpuActions := s.GetCpuActions()
		if len(cpuActions) > 0 {
			b.WriteString(color.Bold(i18n.T("sevens.cpuActionsLabel")) + "\n")
			for _, action := range cpuActions {
				actPlayerName := cuiPlayerName(s.GetPlayer(action.PlayerIdx), action.PlayerIdx)
				b.WriteString(sevensActionStr(actPlayerName, action))
			}
		}

		cuiErrorBlock(b, lastErr)

		if s.GetGameEndFlag() {
			b.WriteString(i18n.T("sevens.gameEnd") + "\n")
			for i := 0; i < s.GetPlayerCnt(); i++ {
				player := s.GetPlayer(i)
				if player == nil {
					continue
				}
				b.WriteString(i18n.Tf("sevens.rankLine",
					"name", cuiPlayerName(player, i),
					"rank", strconv.Itoa(player.GetRank())) + "\n")
			}
		} else {
			currentTurn := s.GetCurrentTurn()
			currentName := cuiPlayerName(s.GetPlayer(currentTurn), currentTurn)
			b.WriteString(i18n.Tf("sevens.turnLine", "name", currentName) + "\n")
			b.WriteString(i18n.T("sevens.playHint") + "\n")
			if config.JokerCount > 0 {
				b.WriteString(i18n.T("sevens.jokerHint") + "\n")
			}
		}
	})
}
