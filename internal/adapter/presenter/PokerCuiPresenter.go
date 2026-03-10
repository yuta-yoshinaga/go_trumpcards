package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PokerCuiPresenter ポーカーCUIプレゼンタークラス
type PokerCuiPresenter struct{}

// NewPokerCuiPresenter コンストラクタ
func NewPokerCuiPresenter() *PokerCuiPresenter {
	return &PokerCuiPresenter{}
}

// Output ゲーム状態を出力
func (pcp *PokerCuiPresenter) Output(p interfaces.PokerGame, lastErr error) string {
	var b strings.Builder
	players := p.GetPlayers()

	b.WriteString("==========\n")
	if p.GetConfig().IsLowball {
		b.WriteString("5-Card Draw Poker [2-7 Lowball]\n")
	} else {
		b.WriteString("5-Card Draw Poker\n")
	}
	b.WriteString("==========\n")

	// ディーラー位置
	fmt.Fprintf(&b, "ディーラー: Player %d\n", p.GetDealerIdx())

	// ポット
	fmt.Fprintf(&b, "ポット: %d\n", p.GetPot())

	// ジョーカー枚数
	if p.GetConfig().JokerCount > 0 {
		fmt.Fprintf(&b, "ジョーカー: %d枚\n", p.GetConfig().JokerCount)
	}

	// ベッティングリミット
	cfg := p.GetConfig()
	if int(cfg.BettingLimit) < len(domain.BettingLimitNames) {
		fmt.Fprintf(&b, "リミット: %s\n", domain.BettingLimitNames[cfg.BettingLimit])
	}

	// プレイヤー情報
	b.WriteString("----------\n")
	isEnd := p.GetPhase() == domain.PokerPhaseEnd
	for i, player := range players {
		if player.GetIsHuman() {
			b.WriteString("[You]")
		} else {
			fmt.Fprintf(&b, "CPU %d (%s)", i, player.GetPlayStyleName())
		}

		fmt.Fprintf(&b, " チップ:%d", player.GetChips())

		if player.GetFolded() {
			b.WriteString(" [フォールド]")
		} else if player.GetAllIn() {
			b.WriteString(" [オールイン]")
		}

		if player.GetCurrentBet() > 0 {
			fmt.Fprintf(&b, " ベット:%d", player.GetCurrentBet())
		}

		if player.GetExchangeCount() > 0 && (p.GetPhase() == domain.PokerPhaseSecondBet || isEnd) {
			fmt.Fprintf(&b, " 交換:%d枚", player.GetExchangeCount())
		}
		b.WriteString("\n")

		// 人間の手札は常に表示
		if player.GetIsHuman() && !player.GetFolded() {
			b.WriteString("  手札: ")
			for j := 0; j < player.GetCardsSize(); j++ {
				if j > 0 {
					b.WriteString("  ")
				}
				fmt.Fprintf(&b, "[%d]%s", j, cuiCardStrEmoji(player.GetCard(j)))
			}
			if isEnd {
				fmt.Fprintf(&b, "  [%s]", player.GetHandName())
			}
			b.WriteString("\n")
		}

		// CPUの手札は終了時のみ表示
		if !player.GetIsHuman() && isEnd && !player.GetFolded() {
			b.WriteString("  手札: ")
			for j := 0; j < player.GetCardsSize(); j++ {
				if j > 0 {
					b.WriteString("  ")
				}
				b.WriteString(cuiCardStrEmoji(player.GetCard(j)))
			}
			fmt.Fprintf(&b, "  [%s]", player.GetHandName())
			b.WriteString("\n")
		}
	}

	// CPU行動記録
	cpuActions := p.GetCpuActions()
	if len(cpuActions) > 0 {
		b.WriteString("----------\n")
		b.WriteString("[CPU行動]\n")
		for _, action := range cpuActions {
			fmt.Fprintf(&b, "  Player %d: %s", action.PlayerIdx, cuiBettingActionName(action.Action))
			if action.Amount > 0 {
				fmt.Fprintf(&b, " (%d)", action.Amount)
			}
			b.WriteString("\n")
		}
	}

	// CPU交換記録
	cpuExchanges := p.GetCpuExchanges()
	if len(cpuExchanges) > 0 {
		b.WriteString("----------\n")
		b.WriteString("[CPU交換]\n")
		for _, ex := range cpuExchanges {
			fmt.Fprintf(&b, "  Player %d: %d枚交換\n", ex.PlayerIdx, ex.ExchangeCount)
		}
	}

	// ショーダウン結果
	results := p.GetRoundResults()
	if len(results) > 0 && isEnd {
		b.WriteString("==========\n")
		b.WriteString("[結果]\n")
		for _, r := range results {
			name := "You"
			if !players[r.PlayerIdx].GetIsHuman() {
				name = fmt.Sprintf("CPU %d", r.PlayerIdx)
			}
			if r.HandName != "" {
				fmt.Fprintf(&b, "  %s: %s", name, r.HandName)
				if ks := domain.FormatKickers(r.Kickers); ks != "" {
					fmt.Fprintf(&b, " (キッカー: %s)", ks)
				}
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
		fmt.Fprintf(&b, "[エラー] %s\n", lastErr.Error())
	}

	// ゲーム終了メッセージ
	if p.GetGameEndFlag() {
		b.WriteString("ゲーム終了\n")
	}

	return b.String()
}

// ActionLogOutput 棋譜をテキスト出力
func (pcp *PokerCuiPresenter) ActionLogOutput(p interfaces.PokerGame) string {
	return actionLogOutputText(p)
}

// OutputWithOdds ゲーム状態 + オッズ出力
func (pcp *PokerCuiPresenter) OutputWithOdds(p interfaces.PokerGame, lastErr error, odds []domain.PokerDrawOdds) string {
	base := pcp.Output(p, lastErr)
	if len(odds) == 0 {
		return base
	}
	var oddsBuilder strings.Builder
	oddsBuilder.WriteString("==========\n")
	oddsBuilder.WriteString("[ドローオッズ]\n")
	for _, o := range odds {
		fmt.Fprintf(&oddsBuilder, "  %s: %.2f%% (%d/%d)\n", o.HandName, o.Probability*100, o.Count, o.Total)
	}
	return base + oddsBuilder.String()
}
