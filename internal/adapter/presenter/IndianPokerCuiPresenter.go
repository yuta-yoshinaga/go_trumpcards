package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// IndianPokerCuiPresenter インディアンポーカーCUIプレゼンタークラス
type IndianPokerCuiPresenter struct{}

// ActionLogOutput 棋譜をテキスト出力
func (p *IndianPokerCuiPresenter) ActionLogOutput(ip interfaces.IndianPokerGame) string {
	return actionLogOutputText(ip)
}

// Output ゲーム状態を文字列出力
func (p *IndianPokerCuiPresenter) Output(ip interfaces.IndianPokerGame, lastErr error) string {
	var b strings.Builder

	b.WriteString("==========\n")
	b.WriteString("Indian Poker\n")
	b.WriteString("==========\n")

	// ディーラー位置
	fmt.Fprintf(&b, "ディーラー: Player %d\n", ip.GetDealerIdx())

	// ポット
	fmt.Fprintf(&b, "ポット: %d\n", ip.GetPot())

	// ベッティングリミット
	cfg := ip.GetConfig()
	if int(cfg.BettingLimit) < len(domain.BettingLimitNames) {
		fmt.Fprintf(&b, "リミット: %s\n", domain.BettingLimitNames[cfg.BettingLimit])
	}

	// アンティ
	fmt.Fprintf(&b, "アンティ: %d\n", cfg.Ante)

	// プレイヤー情報
	b.WriteString("----------\n")
	isShowdown := ip.GetPhase() == domain.IndianPokerPhaseShowdown || ip.GetPhase() == domain.IndianPokerPhaseEnd
	for i := 0; i < ip.GetPlayerCnt(); i++ {
		player := ip.GetPlayer(i)
		b.WriteString(cuiPlayerNameWithStyle(player, i))

		fmt.Fprintf(&b, " チップ:%d", player.GetChips())

		if player.GetFolded() {
			b.WriteString(" " + color.BoldYellow("[フォールド]"))
		} else if player.GetAllIn() {
			b.WriteString(" " + color.BoldYellow("[オールイン]"))
		}

		if player.GetCurrentBet() > 0 {
			fmt.Fprintf(&b, " ベット:%d", player.GetCurrentBet())
		}
		b.WriteString("\n")

		// カード表示
		if player.GetCardsSize() > 0 {
			if player.GetIsHuman() {
				// 人間は自分のカードが見えない
				if isShowdown {
					fmt.Fprintf(&b, "  カード: %s\n", cuiCardStrEmoji(player.GetCard(0)))
				} else {
					b.WriteString("  カード: ??\n")
				}
			} else {
				// CPUのカードは常に表示 (インディアンポーカーでは他人のカードが見える)
				fmt.Fprintf(&b, "  カード: %s\n", cuiCardStrEmoji(player.GetCard(0)))
			}
		}
	}

	// CPU行動記録
	cpuActions := ip.GetCpuActions()
	if len(cpuActions) > 0 {
		b.WriteString("----------\n")
		b.WriteString(color.Bold("[CPU行動]") + "\n")
		for _, action := range cpuActions {
			fmt.Fprintf(&b, "  Player %d: %s", action.PlayerIdx, cuiBettingActionName(action.Action))
			if action.Amount > 0 {
				fmt.Fprintf(&b, " (%d)", action.Amount)
			}
			b.WriteString("\n")
		}
	}

	// ショーダウン結果
	results := ip.GetRoundResults()
	if len(results) > 0 && isShowdown {
		b.WriteString("==========\n")
		b.WriteString(color.Bold("[結果]") + "\n")
		for _, r := range results {
			name := cuiPlayerName(ip.GetPlayer(r.PlayerIdx), r.PlayerIdx)
			if r.Card != nil {
				fmt.Fprintf(&b, "  %s: %s", name, cuiCardStrEmoji(r.Card))
			} else {
				fmt.Fprintf(&b, "  %s", name)
			}
			if r.WonAmount > 0 {
				fmt.Fprintf(&b, " → %dチップ獲得", r.WonAmount)
			}
			b.WriteString("\n")
		}
	}

	// エラーメッセージ
	if lastErr != nil {
		fmt.Fprintf(&b, "%s\n", color.Red(lastErr.Error()))
	}

	// ゲーム終了メッセージ
	if ip.GetGameEndFlag() {
		b.WriteString("ゲーム終了\n")
	}

	return b.String()
}
