//go:build !js || !wasm || extra3

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// minibridgePlayerStr returns the display string for a single player.
func minibridgePlayerStr(s interfaces.MinibridgeGame, idx int, current bool) string {
	player := s.GetPlayer(idx)
	var b strings.Builder
	role := ""
	switch idx {
	case s.GetDeclarerIdx():
		role = i18n.T("minibridge.roleDeclarer")
	case s.GetDummyIdx():
		role = i18n.T("minibridge.roleDummy")
	}
	marker := " "
	if current {
		marker = ">"
	}
	// **HCP は公開情報。** 競りが無いこのゲームでは、これが判断の唯一の材料。
	b.WriteString(marker + i18n.Tf("minibridge.playerLine",
		"name", cuiPlayerName(player, idx),
		"role", role,
		"team", strconv.Itoa(player.GetTeam()),
		"hcp", strconv.Itoa(player.GetHcp()),
		"took", strconv.Itoa(player.GetTrickCount()),
		"cards", strconv.Itoa(player.GetCardsSize()),
	))
	b.WriteString("\n")
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player) + "\n")
	}
	return b.String()
}

// MinibridgeCuiPresenter renders the Minibridge CUI view.
type MinibridgeCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *MinibridgeCuiPresenter) Output(s interfaces.MinibridgeGame, lastErr error) string {
	return buildCuiOutput(i18n.T("minibridge.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("minibridge.header",
			"round", strconv.Itoa(s.GetRoundNumber()),
			"total", strconv.Itoa(s.GetConfig().Rounds),
			"trick", strconv.Itoa(s.GetTrickNumber()+1),
			"tricks", strconv.Itoa(domain.MinibridgeTotalTricks)) + "\n")
		// **競りが無いこと自体が規則。** 毎回書く。
		sb.WriteString(i18n.T("minibridge.rule") + "\n")

		if s.GetContractLevel() > 0 {
			sb.WriteString(i18n.Tf("minibridge.contractLine",
				"level", strconv.Itoa(s.GetContractLevel()),
				"suit", minibridgeSuitName(s.GetContractSuit()),
				"name", cuiPlayerName(s.GetPlayer(s.GetDeclarerIdx()), s.GetDeclarerIdx()),
				"need", strconv.Itoa(s.RequiredTricks())) + "\n")
		} else {
			sb.WriteString(i18n.T("minibridge.contractUndecided") + "\n")
		}

		sb.WriteString(i18n.Tf("minibridge.scoreLine",
			"ns", strconv.Itoa(s.GetTeamScore(0)),
			"ew", strconv.Itoa(s.GetTeamScore(1))) + "\n")

		for i := 0; i < s.GetPlayerCnt(); i++ {
			sb.WriteString(minibridgePlayerStr(s, i,
				i == s.GetCurrentPlayerIdx() && !s.GetGameEndFlag()))
		}

		// **ダミーは契約が決まると公開される。** 人間デクレアラーはここから選ぶ。
		if dummy := s.GetDummyHand(); len(dummy) > 0 {
			parts := make([]string, 0, len(dummy))
			for i, c := range dummy {
				parts = append(parts, "["+strconv.Itoa(i)+"]"+cuiCardStr(c))
			}
			sb.WriteString(i18n.Tf("minibridge.dummyHand",
				"cards", strings.Join(parts, " ")) + "\n")
		}

		sb.WriteString("----------\n")

		cuiTrickBlock(sb, s.GetCurrentTrick(),
			func(tc *domain.TrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(s.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(sb, lastErr)

		if s.GetGameEndFlag() {
			var banner string
			switch s.GetWinnerTeam() {
			case 0:
				banner = i18n.T("minibridge.gameEndYou")
			case -1:
				banner = i18n.T("minibridge.gameEndTie")
			default:
				banner = i18n.T("minibridge.gameEndCpu")
			}
			sb.WriteString(color.Green(banner) + "\n")
			return
		}

		switch s.GetPhase() {
		case domain.MinibridgePhaseContract:
			if s.IsHumanContractTurn() {
				sb.WriteString(i18n.T("minibridge.promptContract") + "\n")
			} else {
				sb.WriteString(i18n.T("minibridge.promptContractWait") + "\n")
			}
		case domain.MinibridgePhaseRoundEnd:
			sb.WriteString(i18n.T("minibridge.promptNext") + "\n")
		default:
			currentIdx := s.GetCurrentPlayerIdx()
			sb.WriteString(i18n.Tf("minibridge.promptCurrentPlayer",
				"name", cuiPlayerName(s.GetPlayer(currentIdx), currentIdx)) + "\n")
			// **ダミーの手番も自分で出す。** 何を動かしているのか言い分ける。
			if currentIdx == s.GetDummyIdx() && s.IsHumanTurn() {
				sb.WriteString(i18n.T("minibridge.promptPlayDummy") + "\n")
			} else {
				sb.WriteString(i18n.T("minibridge.promptPlay") + "\n")
			}
		}
	})
}

// minibridgeSuitName 契約の種別を i18n の名前に変換する。
//
// **0 はノートランプ。** 「未決定」ではなく、選べる 5 つ目の選択肢です。
func minibridgeSuitName(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return i18n.T("minibridge.suitSpade")
	case domain.CardDesignClover:
		return i18n.T("minibridge.suitClover")
	case domain.CardDesignHeart:
		return i18n.T("minibridge.suitHeart")
	case domain.CardDesignDiamond:
		return i18n.T("minibridge.suitDiamond")
	default:
		return i18n.T("minibridge.suitNoTrump")
	}
}

// HintOutput emits the current hint.
func (p *MinibridgeCuiPresenter) HintOutput(s interfaces.MinibridgeGame) string {
	hint := s.GetHint()
	if hint == nil {
		return i18n.T("minibridge.hintNone") + "\n"
	}
	reason := hintReasonStr(hint.Reason, minibridgeHintReasonKeys)
	if hint.CardIndex == nil {
		// **契約の助言は札ではなくレベルと種別を指す。**
		return color.Yellow(i18n.Tf("minibridge.hintContract",
			"level", strconv.Itoa(hint.Level),
			"suit", minibridgeSuitName(hint.Suit),
			"reason", reason)) + "\n"
	}
	// **ダミーの手番なら、指しているのはダミーの手札。**
	seat := s.GetCurrentPlayerIdx()
	card := s.GetPlayer(seat).GetCard(*hint.CardIndex)
	return color.Yellow(i18n.Tf("minibridge.hintCard",
		"idx", strconv.Itoa(*hint.CardIndex),
		"card", cuiCardStr(card),
		"reason", reason)) + "\n"
}

// minibridgeHintReasonKeys maps hint-reason identifiers to their i18n keys.
var minibridgeHintReasonKeys = map[string]string{
	"minibridgeContract": "minibridge.hintReasonContract",
	"minibridgeWinTrick": "minibridge.hintReasonWinTrick",
	"minibridgeDummy":    "minibridge.hintReasonDummy",
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *MinibridgeCuiPresenter) ActionLogOutput(s interfaces.MinibridgeGame) string {
	return actionLogOutputTextForSeats[*domain.MinibridgePlayer](s)
}
