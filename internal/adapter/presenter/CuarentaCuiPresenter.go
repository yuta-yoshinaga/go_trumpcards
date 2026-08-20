//go:build !js || !wasm || extra2

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// CuarentaCuiPresenter renders the Cuarenta CUI view.
type CuarentaCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *CuarentaCuiPresenter) Output(cg interfaces.CuarentaGame, lastErr error) string {
	return buildCuiOutput(i18n.T("cuarenta.helpTitle"), func(b *strings.Builder) {
		// チームスコア表示。
		for t := 0; t < domain.CuarentaTeamCnt; t++ {
			b.WriteString(i18n.Tf("cuarenta.teamScoreLine",
				"team", strconv.Itoa(t),
				"score", strconv.Itoa(cg.GetTeamScore(t))) + "\n")
		}
		// **ボーナスまでの残りをチーム単位で出す。**プレイヤー単位の捕獲数しか
		// 出しておらず、2 人分を毎回自分で合計させていた (#4893)。
		// **閾値は「超えた」チーム** — 20 枚ちょうどでは付かない。
		for t := 0; t < domain.CuarentaTeamCnt; t++ {
			captured := cg.GetTeamCapturedCount(t)
			line := i18n.Tf("cuarenta.teamCaptured",
				"team", strconv.Itoa(t),
				"count", strconv.Itoa(captured),
				"need", strconv.Itoa(domain.CuarentaMostCardsThreshold+1),
				"bonus", strconv.Itoa(domain.CuarentaScoreMostCards))
			// Web と同じく、閾値の 1 つ手前から強調する (あと 2 枚)。
			if captured >= domain.CuarentaMostCardsThreshold-1 {
				line = color.Yellow(line)
			}
			b.WriteString(line + "\n")
		}
		b.WriteString("----------\n")

		for i := 0; i < cg.GetPlayerCnt(); i++ {
			b.WriteString(cuarentaPlayerStr(cg, cg.GetPlayer(i), i))
		}
		b.WriteString("----------\n")

		if tableCards := cg.GetTableCards(); len(tableCards) > 0 {
			b.WriteString(i18n.Tf("cuarenta.tableLine", "cards", cuiCardSliceStr(tableCards)) + "\n")
		} else {
			b.WriteString(i18n.T("cuarenta.tableEmpty") + "\n")
		}

		if ha := cg.GetHumanAction(); ha != nil {
			b.WriteString(i18n.Tf("cuarenta.humanActionLine", "text", cuarentaActionStr(ha)) + "\n")
		}
		if cpu := cg.GetCpuActions(); len(cpu) > 0 {
			b.WriteString(color.Bold(i18n.T("cuarenta.cpuActionsHeader")) + "\n")
			for _, a := range cpu {
				b.WriteString(i18n.Tf("cuarenta.cpuActionLine",
					"name", cuiPlayerName(cg.GetPlayer(a.PlayerIdx), a.PlayerIdx),
					"text", cuarentaActionStr(a)) + "\n")
			}
		}

		cuiErrorBlock(b, lastErr)

		if cg.GetGameEndFlag() {
			b.WriteString(i18n.T("cuarenta.gameEnd") + "\n")
			for t := 0; t < domain.CuarentaTeamCnt; t++ {
				b.WriteString(i18n.Tf("cuarenta.scoreEntry",
					"team", strconv.Itoa(t),
					"score", strconv.Itoa(cg.GetTeamScore(t))) + "\n")
			}
			return
		}
		currentTurn := cg.GetCurrentTurn()
		b.WriteString(i18n.Tf("cuarenta.promptCurrentTurn",
			"name", cuiPlayerName(cg.GetPlayer(currentTurn), currentTurn)) + "\n")
		b.WriteString(i18n.T("cuarenta.promptHelp") + "\n")
	})
}

// cuarentaPlayerStr returns the display string for a single Cuarenta player.
func cuarentaPlayerStr(cg interfaces.CuarentaGame, player *domain.CuarentaPlayer, i int) string {
	if player == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("cuarenta.playerLine",
		"name", cuiPlayerName(player, i),
		"team", strconv.Itoa(domain.CuarentaTeamOf(i)),
		"hand", strconv.Itoa(player.GetCardsSize()),
		"captured", strconv.Itoa(player.CapturedCount())) + "\n")
	if player.GetIsHuman() {
		// **どの札で何枚取れるかが手の選択そのもの** (#5673)。Web は手札に
		// フォーカスすると捕獲対象の場札にリングを出すのに、CUI は素の一覧で
		// 手探りだった。枚数はドメインの捕獲判定から数える。
		if !cg.IsHumanTurn() {
			b.WriteString(cuiIndexedCardListStr(player) + "\n")
			return b.String()
		}
		table := cg.GetTableCards()
		parts := make([]string, player.GetCardsSize())
		for idx := 0; idx < player.GetCardsSize(); idx++ {
			parts[idx] = "[" + strconv.Itoa(idx) + "]" + cuiCardStr(player.GetCard(idx))
			// 0 枚のときは何も付けない -- 「取れる手がある」と読めてしまう。
			if n := domain.CuarentaCaptureCount(player.GetCard(idx), table); n > 0 {
				parts[idx] += i18n.Tf("cuarenta.capturePreview", "count", strconv.Itoa(n))
			}
		}
		b.WriteString(strings.Join(parts, "  ") + "\n")
	}
	return b.String()
}

// cuarentaActionStr renders an action as a short readable line.
func cuarentaActionStr(a *domain.CuarentaAction) string {
	if a == nil {
		return ""
	}
	if len(a.CapturedCards) > 0 {
		suffix := ""
		if a.IsCaida {
			suffix += i18n.T("cuarenta.actionCaidaSuffix")
		}
		if a.RondaBonus > 0 {
			suffix += i18n.Tf("cuarenta.actionRondaSuffix", "bonus", strconv.Itoa(a.RondaBonus))
		}
		if a.IsLimpia {
			suffix += i18n.T("cuarenta.actionLimpiaSuffix")
		}
		return i18n.Tf("cuarenta.actionCapture",
			"played", cuarentaCardShort(a.PlayedCard),
			"count", strconv.Itoa(len(a.CapturedCards)),
			"suffix", suffix)
	}
	return i18n.Tf("cuarenta.actionLay", "played", cuarentaCardShort(a.PlayedCard))
}

// cuarentaCardShort renders a single card as a short text representation.
func cuarentaCardShort(c *domain.Card) string {
	if c == nil {
		return "-"
	}
	return cuiCardSliceStr([]*domain.Card{c})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *CuarentaCuiPresenter) ActionLogOutput(cg interfaces.CuarentaGame) string {
	return actionLogOutputTextForSeats[*domain.CuarentaPlayer](cg)
}
