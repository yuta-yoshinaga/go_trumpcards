package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// sevensPlayerStr returns the display string for a single Sevens player.
func sevensPlayerStr(player *domain.SevensPlayer, i int) string {
	var b strings.Builder
	if player.GetIsHuman() {
		b.WriteString("[You]")
	} else {
		fmt.Fprintf(&b, "CPU %d", i)
	}
	if player.GetIsFinished() {
		fmt.Fprintf(&b, ": 上がり/失格 (ランク: %d位)\n", player.GetRank())
	} else {
		if player.GetMaxPasses() == 0 {
			fmt.Fprintf(&b, ": %d枚 (パス: %d/∞)\n", player.GetCardsSize(), player.GetPassesUsed())
		} else {
			fmt.Fprintf(&b, ": %d枚 (パス: %d/%d)\n", player.GetCardsSize(), player.GetPassesUsed(), player.GetMaxPasses())
		}
		if player.GetIsHuman() {
			b.WriteString(cuiIndexedCardListStr(player))
			b.WriteString("\n")
		}
	}
	return b.String()
}

// sevensActionStr returns the display string for a Sevens action.
func sevensActionStr(playerName string, action *domain.SevensCpuAction) string {
	if action.PlayedCard == nil {
		if action.ForcedPass {
			return fmt.Sprintf("%sがパスしました (出せるカードなし)\n", playerName)
		}
		return fmt.Sprintf("%sがパスしました\n", playerName)
	}
	if action.PlayedCard.GetDesign() == domain.CardDesignJoker && action.TargetSuit > 0 {
		return fmt.Sprintf("%sが %s → %s %d を出しました\n",
			playerName, cuiCardStr(action.PlayedCard),
			cuiSuitName(action.TargetSuit), action.TargetValue)
	}
	return fmt.Sprintf("%sが %s を出しました\n", playerName, cuiCardStr(action.PlayedCard))
}

// SevensCuiPresenter 7並べCUIプレゼンタークラス
type SevensCuiPresenter struct{}

// ActionLogOutput 棋譜をテキスト出力
func (p *SevensCuiPresenter) ActionLogOutput(s interfaces.SevensGame) string {
	return actionLogOutputText(s)
}

// Output ゲーム状態を文字列出力
func (p *SevensCuiPresenter) Output(s interfaces.SevensGame, lastErr error) string {
	var b strings.Builder

	b.WriteString("==========\n")
	b.WriteString("Sevens (7並べ)\n")
	config := s.GetConfig()
	if config.TunnelEnabled || config.TunnelSkipWidth >= 2 || config.JokerCount > 0 || config.CpuStrategy != domain.SevensCpuSimple || config.MaxPasses != domain.SevensMaxPasses || config.NoJokerFinish || config.JokerReclaimEnabled || config.EndStopEnabled || config.JokerConsecutiveBanned {
		b.WriteString("ルール:")
		if config.TunnelEnabled {
			b.WriteString(" " + color.Yellow("[トンネル]"))
		}
		if config.TunnelSkipWidth >= 2 {
			b.WriteString(" " + color.Yellow(fmt.Sprintf("[トンネルスキップ%d]", config.TunnelSkipWidth)))
		}
		if config.JokerCount > 0 {
			b.WriteString(" " + color.Yellow(fmt.Sprintf("[ジョーカー×%d]", config.JokerCount)))
		}
		switch config.CpuStrategy {
		case domain.SevensCpuStrategic:
			b.WriteString(" " + color.Yellow("[CPU戦略]"))
		case domain.SevensCpuHarassment:
			b.WriteString(" " + color.Yellow("[嫌がらせ特化]"))
		}
		if config.MaxPasses == 0 {
			b.WriteString(" " + color.Yellow("[パス無制限]"))
		} else if config.MaxPasses != domain.SevensMaxPasses {
			b.WriteString(" " + color.Yellow(fmt.Sprintf("[パス%d回]", config.MaxPasses)))
		}
		if config.NoJokerFinish {
			b.WriteString(" " + color.Yellow("[ジョーカー上がり禁止]"))
		}
		if config.JokerReclaimEnabled {
			b.WriteString(" " + color.Yellow("[ジョーカー回収]"))
		}
		if config.EndStopEnabled {
			b.WriteString(" " + color.Yellow("[片側ストップ]"))
		}
		if config.JokerConsecutiveBanned {
			b.WriteString(" " + color.Yellow("[ジョーカー連続禁止]"))
		}
		b.WriteString("\n")
	}
	b.WriteString("==========\n")

	for i := 0; i < s.GetPlayerCnt(); i++ {
		b.WriteString(sevensPlayerStr(s.GetPlayer(i), i))
	}

	b.WriteString("----------\n")

	// ボード状態
	b.WriteString("ボード:\n")
	suits := []int{domain.CardDesignSpade, domain.CardDesignClover, domain.CardDesignHeart, domain.CardDesignDiamond}
	suitNames := []string{"SPADE", "CLOVER", "HEART", "DIAMOND"}
	mins := s.GetTableMinVals()
	maxs := s.GetTableMaxVals()
	for i, suit := range suits {
		fmt.Fprintf(&b, "  %s: %d〜%d\n", suitNames[i], mins[suit], maxs[suit])
	}

	// 人間の前の行動
	humanAction := s.GetHumanAction()
	if humanAction != nil {
		b.WriteString(sevensActionStr(cuiPlayerName(s.GetPlayer(humanAction.PlayerIdx), humanAction.PlayerIdx), humanAction))
	}

	// CPUの行動履歴を表示
	cpuActions := s.GetCpuActions()
	if len(cpuActions) > 0 {
		b.WriteString(color.Bold("[CPUの行動]") + "\n")
		for _, action := range cpuActions {
			actPlayerName := cuiPlayerName(s.GetPlayer(action.PlayerIdx), action.PlayerIdx)
			b.WriteString(sevensActionStr(actPlayerName, action))
		}
	}

	// エラーメッセージ
	if lastErr != nil {
		fmt.Fprintf(&b, "%s\n", color.Red(lastErr.Error()))
	}

	if s.GetGameEndFlag() {
		b.WriteString("ゲーム終了！\n")
		for i := 0; i < s.GetPlayerCnt(); i++ {
			player := s.GetPlayer(i)
			fmt.Fprintf(&b, "  %s: %d位\n", cuiPlayerName(s.GetPlayer(i), i), player.GetRank())
		}
	} else {
		currentTurn := s.GetCurrentTurn()
		currentName := cuiPlayerName(s.GetPlayer(currentTurn), currentTurn)
		fmt.Fprintf(&b, "手番: %s\n", currentName)
		b.WriteString("p [インデックス] でカードを出す / p でパス\n")
		if config.JokerCount > 0 {
			b.WriteString("j [カードインデックス] [スート] [値] でジョーカーを配置\n")
		}
	}

	b.WriteString("==========\n")
	return b.String()
}
