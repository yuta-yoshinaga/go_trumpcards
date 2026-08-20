//go:build !js || !wasm || extra

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

// KingCuiPresenter renders the King CUI view.
type KingCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *KingCuiPresenter) Output(kg interfaces.KingGame, lastErr error) string {
	return buildCuiOutput(i18n.T("king.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("king.dealLine",
			"deal", strconv.Itoa(kg.GetDealNumber()+1),
			"total", strconv.Itoa(domain.KingTotalDeals),
			"dealer", cuiPlayerName(kg.GetPlayer(kg.GetDealerIdx()), kg.GetDealerIdx())) + "\n")

		for i := 0; i < kg.GetPlayerCnt(); i++ {
			b.WriteString(kingPlayerStr(kg.GetPlayer(i), i))
		}
		b.WriteString("----------\n")

		if kg.GetCurrentContract() >= 0 {
			b.WriteString(i18n.Tf("king.contractLine",
				"contract", kingContractLabel(kg.GetCurrentContract()),
				"trump", kingTrumpLabel(kg.GetTrumpSuit())) + "\n")
		}

		if trick := kg.GetCurrentTrick(); len(trick) > 0 {
			cards := make([]*domain.Card, 0, len(trick))
			for _, tc := range trick {
				if tc != nil && tc.Card != nil {
					cards = append(cards, tc.Card)
				}
			}
			b.WriteString(i18n.Tf("king.trickLine", "cards", cuiCardSliceStr(cards)) + "\n")
		} else {
			b.WriteString(i18n.T("king.tableEmpty") + "\n")
		}

		cuiErrorBlock(b, lastErr)

		if kg.GetGameEndFlag() {
			// 最後のディールは finishDeal が直接 GameEnd へ移すので DealEnd を通らない。
			// ここで出さないと、決着を決めた 1 ディールの内訳だけ見えない。
			kingDealGainedLine(b, kg)
			b.WriteString(i18n.T("king.gameEnd") + "\n")
			for i := 0; i < kg.GetPlayerCnt(); i++ {
				pl := kg.GetPlayer(i)
				if pl == nil {
					continue
				}
				b.WriteString(i18n.Tf("king.scoreEntry",
					"name", cuiPlayerName(pl, i),
					"score", strconv.Itoa(pl.GetTotalScore())) + "\n")
			}
			return
		}

		if kg.GetPhase() == domain.KingPhaseSelectContract {
			b.WriteString(i18n.Tf("king.selectPrompt",
				"name", cuiPlayerName(kg.GetPlayer(kg.GetDealerIdx()), kg.GetDealerIdx())) + "\n")
			// List the still-available contracts with their selection indices, so the
			// dealer need not recall which of the seven deals are already spent.
			used := kg.GetUsedContracts()
			var remaining []string
			for i := 0; i < domain.KingContractCnt; i++ {
				if !used[i] {
					remaining = append(remaining, "["+strconv.Itoa(i)+"]"+kingContractLabel(i))
				}
			}
			if len(remaining) > 0 {
				b.WriteString(i18n.Tf("king.remainingContracts",
					"contracts", strings.Join(remaining, ", ")) + "\n")
			}
		} else if kg.GetPhase() == domain.KingPhaseDealEnd {
			kingDealGainedLine(b, kg)
			b.WriteString(i18n.T("king.dealEndPrompt") + "\n")
		} else {
			b.WriteString(i18n.Tf("king.promptCurrentTurn",
				"name", cuiPlayerName(kg.GetPlayer(kg.GetCurrentTurn()), kg.GetCurrentTurn())) + "\n")
		}
		b.WriteString(i18n.T("king.promptHelp") + "\n")
	})
}

// kingPlayerStr returns the display string for a single King player.
func kingPlayerStr(player *domain.KingPlayer, i int) string {
	if player == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString(i18n.Tf("king.playerLine",
		"name", cuiPlayerName(player, i),
		"hand", strconv.Itoa(player.GetCardsSize()),
		"tricks", strconv.Itoa(player.GetTrickCount()),
		"total", strconv.Itoa(player.GetTotalScore())) + "\n")
	if player.GetIsHuman() {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// kingDealGainedLine writes what each player gained on the settled deal.
// プレイヤー行に出ているのは**累計**なので、このディールで動いた分は内訳を
// 読まないと分からない (Web の king-deal-breakdown と同じ情報)。
func kingDealGainedLine(b *strings.Builder, kg interfaces.KingGame) {
	detail := kg.GetLastDealDetail()
	if detail == nil {
		return
	}
	gains := make([]string, 0, kg.GetPlayerCnt())
	for i := 0; i < kg.GetPlayerCnt(); i++ {
		gains = append(gains, fmt.Sprintf("%s %d",
			cuiPlayerName(kg.GetPlayer(i), i), detail.Gained[i]))
	}
	b.WriteString(i18n.Tf("king.dealResultGained", "gains", strings.Join(gains, " / ")) + "\n")
}

// kingContractLabel はコントラクトの表示名を i18n で返す。
func kingContractLabel(contract int) string {
	keys := []string{
		"king.cNoTricks", "king.cNoHearts", "king.cNoQueens", "king.cNoKingHeart",
		"king.cNoLastTwo", "king.cNoMen", "king.cKingTrump",
	}
	if contract < 0 || contract >= len(keys) {
		return "-"
	}
	return i18n.T(keys[contract])
}

// kingTrumpLabel は切り札スートの表示名を返す (-1 = なし)。
func kingTrumpLabel(suit int) string {
	if suit < domain.CardDesignSpade || suit > domain.CardDesignDiamond {
		return "-"
	}
	return suitNames[suit]
}

// HintOutput emits a play recommendation for the human's turn.
func (p *KingCuiPresenter) HintOutput(kg interfaces.KingGame) string {
	hint := kg.GetHint()
	if hint == nil {
		return i18n.T("king.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, kingHintReasonKeys)
	if len(hint.CardIndices) > 0 {
		turn := kg.GetCurrentTurn()
		player := kg.GetPlayer(turn)
		cards := make([]string, len(hint.CardIndices))
		for i, idx := range hint.CardIndices {
			if player != nil && idx >= 0 && idx < player.GetCardsSize() {
				cards[i] = "[" + strconv.Itoa(idx) + "]" + cuiCardStr(player.GetCard(idx))
			} else {
				cards[i] = strconv.Itoa(idx)
			}
		}
		return color.Yellow(i18n.Tf("king.hintCard",
			"cards", strings.Join(cards, ", "),
			"reason", reason)) + "\n"
	}
	return color.Yellow(i18n.Tf("king.hintCard", "cards", "-", "reason", reason)) + "\n"
}

// kingHintReasonKeys maps King-specific hint-reason identifiers to i18n keys.
var kingHintReasonKeys = map[string]string{
	"avoid_low": "king.hintReasonAvoidLow",
	"win_high":  "king.hintReasonWinHigh",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *KingCuiPresenter) ActionLogOutput(kg interfaces.KingGame) string {
	return actionLogOutputTextForSeats[*domain.KingPlayer](kg)
}
