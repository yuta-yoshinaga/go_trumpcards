package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// DoubtCuiPresenter ダウトCUIプレゼンタークラス
type DoubtCuiPresenter struct{}

// NewDoubtCuiPresenter コンストラクタ
func NewDoubtCuiPresenter() *DoubtCuiPresenter {
	return &DoubtCuiPresenter{}
}

// Output ゲーム状態を文字列出力
func (p *DoubtCuiPresenter) Output(d interfaces.DoubtGame, lastErr error) string {
	var b strings.Builder

	b.WriteString("==========\n")
	b.WriteString("Doubt (ダウト)\n")
	b.WriteString("==========\n")

	// プレイヤー情報
	for i := 0; i < d.GetPlayerCnt(); i++ {
		player := d.GetPlayer(i)
		if player.GetIsHuman() {
			b.WriteString("[You]")
		} else {
			fmt.Fprintf(&b, "CPU %d", i)
		}
		if player.GetIsFinished() {
			b.WriteString(": 上がり\n")
		} else {
			fmt.Fprintf(&b, ": %d枚\n", player.GetCardsSize())
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
	fmt.Fprintf(&b, "テーブル: %d枚\n", d.GetTableCardCount())

	// 最後のプレイ情報
	lastAction := d.GetLastAction()
	if lastAction != nil {
		fmt.Fprintf(&b, "[最後のプレイ] %sが「%d」を%d枚出しました\n",
			cuiPlayerName(d.GetPlayer(lastAction.PlayerIdx), lastAction.PlayerIdx),
			lastAction.ClaimedValue,
			lastAction.CardCount,
		)
	}

	// ダウト結果
	lastResult := d.GetLastDoubtResult()
	if lastResult != nil {
		doubterName := cuiPlayerName(d.GetPlayer(lastResult.DoubterIdx), lastResult.DoubterIdx)
		cardPlayerName := cuiPlayerName(d.GetPlayer(lastResult.CardPlayerIdx), lastResult.CardPlayerIdx)
		loserName := cuiPlayerName(d.GetPlayer(lastResult.LoserIdx), lastResult.LoserIdx)
		if lastResult.WasLying {
			fmt.Fprintf(&b, "[ダウト] %sが%sをダウト → 嘘つき！ %sが%d枚引き取りました\n",
				doubterName, cardPlayerName, loserName, lastResult.CardCount)
		} else {
			fmt.Fprintf(&b, "[ダウト] %sが%sをダウト → 正直者！ %sが%d枚引き取りました\n",
				doubterName, cardPlayerName, loserName, lastResult.CardCount)
		}
		// 公開されたカード
		if len(lastResult.RevealedCards) > 0 {
			b.WriteString("  公開カード: ")
			for i, card := range lastResult.RevealedCards {
				if i != 0 {
					b.WriteString(", ")
				}
				b.WriteString(cuiCardStr(card))
			}
			b.WriteString("\n")
		}
	}

	// 人間の行動履歴
	humanAction := d.GetHumanAction()
	if humanAction != nil {
		fmt.Fprintf(&b, "[あなたの行動] 「%d」を%d枚出しました\n",
			humanAction.ClaimedValue, humanAction.CardCount)
	}

	// CPUの行動履歴
	cpuActions := d.GetCpuActions()
	if len(cpuActions) > 0 {
		b.WriteString("[CPUの行動]\n")
		for _, action := range cpuActions {
			fmt.Fprintf(&b, "%sが「%d」を%d枚出しました\n",
				cuiPlayerName(d.GetPlayer(action.PlayerIdx), action.PlayerIdx),
				action.ClaimedValue,
				action.CardCount,
			)
		}
	}

	// エラーメッセージ
	if lastErr != nil {
		fmt.Fprintf(&b, "%s\n", lastErr.Error())
	}

	// ゲーム状態
	if d.GetGameEndFlag() {
		winnerIdx := d.GetWinnerIdx()
		fmt.Fprintf(&b, "ゲーム終了！ %sの勝利です！\n", cuiPlayerName(d.GetPlayer(winnerIdx), winnerIdx))
	} else {
		currentTurn := d.GetCurrentTurn()
		phase := d.GetPhase()
		if phase == domain.DoubtPhaseDoubt {
			lastAct := d.GetLastAction()
			if lastAct != nil {
				fmt.Fprintf(&b, "ダウトフェーズ: %sのプレイにダウトしますか？\n",
					cuiPlayerName(d.GetPlayer(lastAct.PlayerIdx), lastAct.PlayerIdx))
			} else {
				b.WriteString("ダウトフェーズ\n")
			}
			b.WriteString("d <idx...>・・・ダウト / s・・・スキップ\n")
		} else {
			fmt.Fprintf(&b, "手番: %s\n", cuiPlayerName(d.GetPlayer(currentTurn), currentTurn))
			b.WriteString("p <値> <idx...>・・・カードを出す\n")
		}
	}

	b.WriteString("==========\n")
	return b.String()
}
