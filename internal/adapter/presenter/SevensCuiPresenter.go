package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SevensCuiPresenter 7並べCUIプレゼンタークラス
type SevensCuiPresenter struct{}

// NewSevensCuiPresenter コンストラクタ
func NewSevensCuiPresenter() *SevensCuiPresenter {
	return &SevensCuiPresenter{}
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
			b.WriteString(" [トンネル]")
		}
		if config.TunnelSkipWidth >= 2 {
			fmt.Fprintf(&b, " [トンネルスキップ%d]", config.TunnelSkipWidth)
		}
		if config.JokerCount > 0 {
			fmt.Fprintf(&b, " [ジョーカー×%d]", config.JokerCount)
		}
		switch config.CpuStrategy {
		case domain.SevensCpuStrategic:
			b.WriteString(" [CPU戦略]")
		case domain.SevensCpuHarassment:
			b.WriteString(" [嫌がらせ特化]")
		}
		if config.MaxPasses == 0 {
			b.WriteString(" [パス無制限]")
		} else if config.MaxPasses != domain.SevensMaxPasses {
			fmt.Fprintf(&b, " [パス%d回]", config.MaxPasses)
		}
		if config.NoJokerFinish {
			b.WriteString(" [ジョーカー上がり禁止]")
		}
		if config.JokerReclaimEnabled {
			b.WriteString(" [ジョーカー回収]")
		}
		if config.EndStopEnabled {
			b.WriteString(" [片側ストップ]")
		}
		if config.JokerConsecutiveBanned {
			b.WriteString(" [ジョーカー連続禁止]")
		}
		b.WriteString("\n")
	}
	b.WriteString("==========\n")

	for i := 0; i < s.GetPlayerCnt(); i++ {
		player := s.GetPlayer(i)
		if player.GetIsHuman() {
			b.WriteString("[You]")
		} else {
			fmt.Fprintf(&b, "CPU %d", i)
		}
		if player.GetIsFinished() {
			fmt.Fprintf(&b, ": 上がり/失格 (ランク: %d位)\n", player.GetRank())
		} else {
			if player.GetMaxPasses() == 0 {
				fmt.Fprintf(&b, ": %d枚 (パス: %d/∞)\n",
					player.GetCardsSize(), player.GetPassesUsed())
			} else {
				fmt.Fprintf(&b, ": %d枚 (パス: %d/%d)\n",
					player.GetCardsSize(), player.GetPassesUsed(), player.GetMaxPasses())
			}
			if player.GetIsHuman() {
				for j := 0; j < player.GetCardsSize(); j++ {
					if j != 0 {
						b.WriteString("  ")
					}
					fmt.Fprintf(&b, "[%d]%s", j, cuiCardStr(player.GetCard(j)))
				}
				b.WriteString("\n")
			}
		}
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
		if humanAction.PlayedCard == nil {
			if humanAction.ForcedPass {
				fmt.Fprintf(&b, "%sがパスしました (出せるカードなし)\n", cuiPlayerName(s.GetPlayer(humanAction.PlayerIdx), humanAction.PlayerIdx))
			} else {
				fmt.Fprintf(&b, "%sがパスしました\n", cuiPlayerName(s.GetPlayer(humanAction.PlayerIdx), humanAction.PlayerIdx))
			}
		} else {
			if humanAction.PlayedCard.GetDesign() == domain.CardDesignJoker && humanAction.TargetSuit > 0 {
				fmt.Fprintf(&b, "%sが %s → %s %d を出しました\n",
					cuiPlayerName(s.GetPlayer(humanAction.PlayerIdx), humanAction.PlayerIdx), cuiCardStr(humanAction.PlayedCard),
					cuiSuitName(humanAction.TargetSuit), humanAction.TargetValue)
			} else {
				fmt.Fprintf(&b, "%sが %s を出しました\n",
					cuiPlayerName(s.GetPlayer(humanAction.PlayerIdx), humanAction.PlayerIdx), cuiCardStr(humanAction.PlayedCard))
			}
		}
	}

	// CPUの行動履歴を表示
	cpuActions := s.GetCpuActions()
	if len(cpuActions) > 0 {
		b.WriteString("[CPUの行動]\n")
		for _, action := range cpuActions {
			actPlayerName := cuiPlayerName(s.GetPlayer(action.PlayerIdx), action.PlayerIdx)
			if action.PlayedCard == nil {
				if action.ForcedPass {
					fmt.Fprintf(&b, "%sがパスしました (出せるカードなし)\n", actPlayerName)
				} else {
					fmt.Fprintf(&b, "%sがパスしました\n", actPlayerName)
				}
			} else {
				if action.PlayedCard.GetDesign() == domain.CardDesignJoker && action.TargetSuit > 0 {
					fmt.Fprintf(&b, "%sが %s → %s %d を出しました\n",
						actPlayerName, cuiCardStr(action.PlayedCard),
						cuiSuitName(action.TargetSuit), action.TargetValue)
				} else {
					fmt.Fprintf(&b, "%sが %s を出しました\n", actPlayerName, cuiCardStr(action.PlayedCard))
				}
			}
		}
	}

	// エラーメッセージ
	if lastErr != nil {
		fmt.Fprintf(&b, "%s\n", lastErr.Error())
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
