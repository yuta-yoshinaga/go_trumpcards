package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// daifugoPlayerStr returns the display string for a single Daifugo player.
func daifugoPlayerStr(player *domain.DaifugoPlayer, i int) string {
	var b strings.Builder
	if player.GetIsHuman() {
		b.WriteString("[You]")
	} else {
		fmt.Fprintf(&b, "CPU %d", i)
	}
	if player.GetIsFinished() {
		fmt.Fprintf(&b, ": 上がり (ランク: %s)\n", daifugoRankName(player.GetRank()))
	} else {
		fmt.Fprintf(&b, ": %d枚\n", player.GetCardsSize())
		if player.GetIsHuman() {
			b.WriteString(cuiIndexedCardListStr(player))
			b.WriteString("\n")
		}
	}
	return b.String()
}

// DaifugoCuiPresenter 大富豪CUIプレゼンタークラス
type DaifugoCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *DaifugoCuiPresenter) Output(dg interfaces.DaifugoGame, lastErr error) string {
	return buildCuiOutput("Daifugo (大富豪)", func(b *strings.Builder) {
		for i := 0; i < dg.GetPlayerCnt(); i++ {
			b.WriteString(daifugoPlayerStr(dg.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		// ローカルルール状態
		if dg.GetRevolutionActive() {
			b.WriteString("【革命中】2が最弱、3が最強\n")
		}
		if dg.GetElevenBackActive() {
			b.WriteString("【11バック】強さが逆転中\n")
		}
		if dg.GetSuitLocked() {
			fmt.Fprintf(b, "【スート縛り】%s\n", cuiSuitName(dg.GetLockedSuit()))
		}
		if dg.GetTableIsSequence() {
			b.WriteString("【階段】\n")
		}
		if dg.GetReverseDirection() {
			b.WriteString("【9リバース】\n")
		}
		if dg.GetNumberLocked() {
			b.WriteString("【連番縛り】\n")
		}

		// カード交換記録
		exchangeActions := dg.GetExchangeActions()
		if len(exchangeActions) > 0 {
			b.WriteString("[カード交換]\n")
			for _, ex := range exchangeActions {
				fmt.Fprintf(b, "%s → %s: %s\n",
					cuiPlayerName(dg.GetPlayer(ex.FromPlayerIdx), ex.FromPlayerIdx),
					cuiPlayerName(dg.GetPlayer(ex.ToPlayerIdx), ex.ToPlayerIdx),
					cuiCardSliceStr(ex.Cards))
			}
		}

		// 場のカード
		tableCards := dg.GetTableCards()
		if len(tableCards) > 0 {
			lastPlayIdx := dg.GetLastPlayPlayerIdx()
			fmt.Fprintf(b, "場: %s (出したプレイヤー: %s)\n",
				cuiCardSliceStr(tableCards),
				cuiPlayerName(dg.GetPlayer(lastPlayIdx), lastPlayIdx))
		} else {
			b.WriteString("場: なし (誰でも出せます)\n")
		}

		// 人間の前の行動
		humanAction := dg.GetHumanAction()
		if humanAction != nil {
			if len(humanAction.PlayedCards) == 0 {
				fmt.Fprintf(b, "%sがパスしました\n", cuiPlayerName(dg.GetPlayer(humanAction.PlayerIdx), humanAction.PlayerIdx))
			} else {
				fmt.Fprintf(b, "%sが %s を出しました\n",
					cuiPlayerName(dg.GetPlayer(humanAction.PlayerIdx), humanAction.PlayerIdx),
					cuiCardSliceStr(humanAction.PlayedCards))
			}
		}

		// CPUの行動履歴を表示
		cpuActions := dg.GetCpuActions()
		if len(cpuActions) > 0 {
			b.WriteString("[CPUの行動]\n")
			for _, action := range cpuActions {
				actPlayerName := cuiPlayerName(dg.GetPlayer(action.PlayerIdx), action.PlayerIdx)
				if len(action.PlayedCards) == 0 {
					fmt.Fprintf(b, "%sがパスしました\n", actPlayerName)
				} else {
					fmt.Fprintf(b, "%sが %s を出しました\n", actPlayerName, cuiCardSliceStr(action.PlayedCards))
				}
			}
		}

		// エラーメッセージ
		if lastErr != nil {
			fmt.Fprintf(b, "%s\n", lastErr.Error())
		}

		if dg.GetGameEndFlag() {
			b.WriteString("ゲーム終了！\n")
			for i := 0; i < dg.GetPlayerCnt(); i++ {
				player := dg.GetPlayer(i)
				penalty := ""
				if player.GetIllegalFinishPenalty() {
					penalty = " [反則上がり]"
				}
				fmt.Fprintf(b, "  %s: %s%s\n", cuiPlayerName(dg.GetPlayer(i), i), daifugoRankName(player.GetRank()), penalty)
			}
		} else {
			currentTurn := dg.GetCurrentTurn()
			currentName := cuiPlayerName(dg.GetPlayer(currentTurn), currentTurn)
			fmt.Fprintf(b, "手番: %s\n", currentName)
			switch dg.GetPendingActionType() {
			case domain.DaifugoPendingSevenPass:
				b.WriteString("【7渡し】渡すカードを選択してください (p [インデックス])\n")
			case domain.DaifugoPendingTenDiscard:
				b.WriteString("【10捨て】捨てるカードを選択してください (p [インデックス])\n")
			case domain.DaifugoPendingQueenBomber:
				b.WriteString("【12ボンバー】除去するカードの数字を入力してください (p [1-13])\n")
			default:
				b.WriteString("p [インデックス...] でカードを出す / p でパス\n")
			}
		}
	})
}

// ActionLogOutput 棋譜をテキスト出力
func (p *DaifugoCuiPresenter) ActionLogOutput(dg interfaces.DaifugoGame) string {
	return actionLogOutputText(dg)
}
