package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// shitheadMagicFormatter は特殊ランクの札に印を付ける整形関数を返す。
//
// **どのランクが特殊かは設定次第で変わる。**Web は絵文字バッジと title で
// 出しているのに、CUI にはその注記も、Go 側ロケールの文言すら無かった (#5577)。
// 判定は効果を決めている domain.ShitheadIsMagicValue をそのまま呼ぶ ── 一覧を
// 別に持つと、設定を増やしたとき注記だけが古くなる。
func shitheadMagicFormatter(cfg domain.ShitheadConfig) func(*domain.Card) string {
	return func(c *domain.Card) string {
		str := cuiCardStr(c)
		if c != nil && domain.ShitheadIsMagicValue(cfg, c.GetValue()) {
			str += color.Yellow(i18n.T("shithead.magicMark"))
		}
		return str
	}
}

// shitheadMagicLegend は有効な特殊ランクとその効果を 1 行に並べる。何も有効で
// なければ空。印だけ出しても、何が起きるかは分からない。
func shitheadMagicLegend(cfg domain.ShitheadConfig) string {
	ranks := domain.ShitheadMagicRanks(cfg)
	if len(ranks) == 0 {
		return ""
	}
	keys := map[int]string{
		2:  "shithead.magicEffectTwo",
		7:  "shithead.magicEffectSeven",
		8:  "shithead.magicEffectEight",
		10: "shithead.magicEffectTen",
	}
	parts := make([]string, 0, len(ranks))
	for _, v := range ranks {
		// ランクは番号を添えて出す。効果の文そのものは Web と**同じ文字列**を
		// 使うので、突き合わせテストが等値で見られる。
		parts = append(parts, i18n.Tf("shithead.magicLegendItem",
			"rank", strconv.Itoa(v), "effect", i18n.T(keys[v])))
	}
	return i18n.Tf("shithead.magicLegend", "effects", strings.Join(parts, " / ")) + "\n"
}

// shitheadPlayerStr returns the display string for a single Shithead player.
func shitheadPlayerStr(player *domain.ShitheadPlayer, idx int, currentTurn int, cfg domain.ShitheadConfig) string {
	var b strings.Builder
	name := cuiPlayerName(player, idx)
	turnSuffix := ""
	if idx == currentTurn {
		turnSuffix = i18n.T("shithead.playerTurnSuffix")
	}
	if player.GetIsFinished() {
		b.WriteString(name + turnSuffix)
		b.WriteString(i18n.Tf("shithead.playerFinished",
			"rank", strconv.Itoa(player.GetRank())) + "\n")
		return b.String()
	}
	b.WriteString(i18n.Tf("shithead.playerLine",
		"name", name+turnSuffix,
		"hand", strconv.Itoa(player.GetCardsSize()),
		"up", strconv.Itoa(player.GetFaceUpSize()),
		"down", strconv.Itoa(player.GetFaceDownSize())) + "\n")
	if player.GetIsHuman() {
		if player.GetCardsSize() > 0 {
			b.WriteString(i18n.T("shithead.handLabel"))
			b.WriteString(formatCardList(player, shitheadMagicFormatter(cfg), "  ", true) + "\n")
		}
		if player.GetFaceUpSize() > 0 {
			b.WriteString(i18n.T("shithead.faceupLabel"))
			b.WriteString(formatCardSlice(player.GetFaceUpCards(), shitheadMagicFormatter(cfg), ", ") + "\n")
		}
	} else if player.GetFaceUpSize() > 0 {
		b.WriteString(i18n.T("shithead.faceupLabel"))
		b.WriteString(formatCardSlice(player.GetFaceUpCards(), shitheadMagicFormatter(cfg), ", ") + "\n")
	}
	return b.String()
}

// ShitheadCuiPresenter renders the Shithead CUI view.
type ShitheadCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *ShitheadCuiPresenter) Output(sg interfaces.ShitheadGame, lastErr error) string {
	return buildCuiOutput(i18n.T("shithead.outputTitle"), func(b *strings.Builder) {
		cfg := sg.GetConfig()
		// 印だけ出しても何が起きるかは分からないので、効果も並べる。
		b.WriteString(shitheadMagicLegend(cfg))

		currentTurn := sg.GetCurrentTurn()
		for i := 0; i < sg.GetPlayerCnt(); i++ {
			b.WriteString(shitheadPlayerStr(sg.GetPlayer(i), i, currentTurn, cfg))
		}

		b.WriteString("----------\n")

		// Discard pile + stock
		discard := sg.GetDiscardPile()
		if len(discard) > 0 {
			b.WriteString(i18n.Tf("shithead.discardLine",
				"cards", formatCardSlice(discard, shitheadMagicFormatter(cfg), ", ")) + "\n")
		} else {
			b.WriteString(i18n.T("shithead.discardEmpty") + "\n")
		}
		b.WriteString(i18n.Tf("shithead.stockLine",
			"count", strconv.Itoa(sg.GetStockSize())) + "\n")
		if sg.GetSevenActive() {
			b.WriteString(color.BoldYellow(i18n.T("shithead.noticeSeven")) + "\n")
		}
		if sg.GetSkipNext() {
			b.WriteString(color.BoldYellow(i18n.T("shithead.noticeSkip")) + "\n")
		}

		// Last human action
		if humanAction := sg.GetHumanAction(); humanAction != nil {
			b.WriteString(formatShitheadAction(sg, humanAction))
		}

		// CPU action history
		cpuActions := sg.GetCpuActions()
		if len(cpuActions) > 0 {
			b.WriteString(color.Bold(i18n.T("shithead.cpuActionsHeader")) + "\n")
			for _, action := range cpuActions {
				b.WriteString(formatShitheadAction(sg, action))
			}
		}

		cuiErrorBlock(b, lastErr)

		if sg.GetGameEndFlag() {
			b.WriteString(i18n.T("shithead.gameEnd") + "\n")
			for i := 0; i < sg.GetPlayerCnt(); i++ {
				player := sg.GetPlayer(i)
				rank := player.GetRank()
				suffix := ""
				if rank == sg.GetPlayerCnt() {
					suffix = i18n.T("shithead.rankShithead")
				}
				b.WriteString(i18n.Tf("shithead.rankLine",
					"name", cuiPlayerName(player, i),
					"rank", strconv.Itoa(rank),
					"suffix", suffix) + "\n")
			}
			return
		}
		source := sg.CurrentSource()
		currentName := cuiPlayerName(sg.GetPlayer(currentTurn), currentTurn)
		b.WriteString(i18n.Tf("shithead.promptCurrentTurn",
			"name", currentName,
			"source", source) + "\n")
		// On a blind (face-down) human turn, list the selectable slot indices so
		// the player knows which number to play; the card faces stay hidden.
		if source == domain.ShitheadSourceFaceDown {
			if human := sg.GetPlayer(currentTurn); human != nil && human.GetIsHuman() && human.GetFaceDownSize() > 0 {
				slots := make([]string, human.GetFaceDownSize())
				for i := range slots {
					slots[i] = "[" + strconv.Itoa(i) + "]??"
				}
				b.WriteString(i18n.Tf("shithead.facedownSlots",
					"slots", strings.Join(slots, " ")) + "\n")
			}
		}
		b.WriteString(i18n.T("shithead.promptPlayHelp") + "\n")
	})
}

// formatShitheadAction returns one line describing a player action.
func formatShitheadAction(sg interfaces.ShitheadGame, action *domain.ShitheadCpuAction) string {
	name := cuiPlayerName(sg.GetPlayer(action.PlayerIdx), action.PlayerIdx)
	if action.Pickup {
		return i18n.Tf("shithead.actionPickup", "name", name) + "\n"
	}
	suffix := ""
	if action.Burned {
		suffix += i18n.T("shithead.actionSuffixBurned")
	}
	if action.Skipped {
		suffix += i18n.T("shithead.actionSuffixSkipped")
	}
	return i18n.Tf("shithead.actionPlay",
		"name", name,
		"cards", cuiCardSliceStr(action.PlayedCards),
		"source", action.Source,
		"suffix", suffix) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *ShitheadCuiPresenter) ActionLogOutput(sg interfaces.ShitheadGame) string {
	return actionLogOutputTextForSeats[*domain.ShitheadPlayer](sg)
}
