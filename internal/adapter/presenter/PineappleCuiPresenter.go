package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// PineappleCuiPresenter パイナップルポーカーCUIプレゼンタークラス
type PineappleCuiPresenter struct{}

// ActionLogOutput 棋譜をテキスト出力
func (pp *PineappleCuiPresenter) ActionLogOutput(p interfaces.PineappleGame) string {
	return actionLogOutputText(p)
}

// Output ゲーム状態を文字列出力
func (pp *PineappleCuiPresenter) Output(p interfaces.PineappleGame, lastErr error) string {
	var b strings.Builder

	b.WriteString("==========\n")
	b.WriteString("Pineapple Poker\n")
	b.WriteString("==========\n")

	// トーナメントモードヘッダー
	cfg := p.GetConfig()
	if cfg.TournamentMode {
		fmt.Fprintf(&b, "トーナメント ハンド#%d SB:%d BB:%d (レベルアップ:%dハンド毎)\n",
			p.GetHandCount(), cfg.SmallBlind, cfg.BigBlind, cfg.BlindLevelHands)
		if cfg.RebuyEnabled {
			fmt.Fprintf(&b, "リバイ: %dチップ (最大%d回, %dハンド目まで)\n",
				cfg.RebuyChips, cfg.RebuyMaxCount, cfg.RebuyPeriodHands)
		}
		if cfg.AddonEnabled {
			fmt.Fprintf(&b, "アドオン: %dチップ (%dハンド目に提供)\n",
				cfg.AddonChips, cfg.AddonAfterHand)
		}
	}

	// テーブルサイズ
	fmt.Fprintf(&b, "テーブル: %d-max\n", p.GetPlayerCnt())

	// ディーラー位置
	fmt.Fprintf(&b, "ディーラー: Player %d\n", p.GetDealerIdx())

	// コミュニティカード
	cc := p.GetCommunityCards()
	if len(cc) == 0 {
		b.WriteString("コミュニティ: (なし)\n")
	} else {
		fmt.Fprintf(&b, "コミュニティ: %s\n", cuiCardSliceStrEmoji(cc))
	}

	// ポット
	fmt.Fprintf(&b, "ポット: %d\n", p.GetPot())

	// ベッティングリミット
	if int(cfg.BettingLimit) < len(domain.BettingLimitNames) {
		fmt.Fprintf(&b, "リミット: %s\n", domain.BettingLimitNames[cfg.BettingLimit])
	}

	// プレイヤー情報
	b.WriteString("----------\n")
	for i := 0; i < p.GetPlayerCnt(); i++ {
		player := p.GetPlayer(i)
		b.WriteString(cuiPlayerNameWithStyle(player, i))

		fmt.Fprintf(&b, " チップ:%d", player.GetChips())

		if player.GetTotalHands() > 0 {
			fmt.Fprintf(&b, " VPIP:%d%% PFR:%d%% 3Bet:%d%% AF:%s", player.GetVPIP(), player.GetPFR(), player.GetThreeBet(), player.GetAFDisplay())
		}

		if player.GetFolded() {
			b.WriteString(" " + color.BoldYellow("[フォールド]"))
		} else if player.GetAllIn() {
			b.WriteString(" " + color.BoldYellow("[オールイン]"))
		}

		if player.GetCurrentBet() > 0 {
			fmt.Fprintf(&b, " ベット:%d", player.GetCurrentBet())
		}
		b.WriteString("\n")

		// 人間のカードを表示
		if player.GetIsHuman() && !player.GetFolded() {
			fmt.Fprintf(&b, "  手札: %s\n", cuiCardListStrEmoji(player))
		}
	}

	// CPU行動記録
	cpuActions := p.GetCpuActions()
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

	// ディスカードフェーズ
	if p.GetPhase() == domain.PineapplePhaseDiscard {
		b.WriteString("----------\n")
		b.WriteString("--- ディスカードフェーズ ---\n")
		b.WriteString("手札から1枚選んでディスカードしてください\n")
	}

	// ショーダウン結果
	results := p.GetRoundResults()
	if len(results) > 0 && (p.GetPhase() == domain.PineapplePhaseEnd || p.GetPhase() == domain.PineapplePhaseShowdown) {
		b.WriteString("==========\n")
		b.WriteString(color.Bold("[結果]") + "\n")
		for _, r := range results {
			name := cuiPlayerName(p.GetPlayer(r.PlayerIdx), r.PlayerIdx)
			kickers := ""
			if ks := domain.FormatKickers(r.Kickers); ks != "" {
				kickers = " (キッカー: " + ks + ")"
			}
			if r.Mucked {
				fmt.Fprintf(&b, "  %s: マック", name)
			} else if r.HandName != "" {
				fmt.Fprintf(&b, "  %s: %s%s", name, r.HandName, kickers)
			} else {
				fmt.Fprintf(&b, "  %s", name)
			}
			if r.WonAmount > 0 {
				fmt.Fprintf(&b, " → %dチップ獲得", r.WonAmount)
			}
			b.WriteString("\n")
		}
	}

	// マックプロンプト
	if p.IsMuckAvailable() {
		b.WriteString("----------\n")
		b.WriteString("マックしますか? (m=マック / sh=ショー)\n")
	}

	// リバイ/アドオンプロンプト
	if p.GetPhase() == domain.PineapplePhaseRebuy {
		b.WriteString("----------\n")
		rebuyPhaseType := p.GetRebuyPhaseType()
		if rebuyPhaseType == domain.PineappleRebuyPhaseRebuy {
			rebuyCounts := p.GetRebuyCounts()
			humanIdx := -1
			for i := 0; i < p.GetPlayerCnt(); i++ {
				if p.GetPlayer(i).GetIsHuman() {
					humanIdx = i
					break
				}
			}
			if humanIdx >= 0 {
				fmt.Fprintf(&b, "リバイしますか? (%dチップ, %d/%d回使用済) (rb=リバイ / sr=スキップ)\n",
					cfg.RebuyChips, rebuyCounts[humanIdx], cfg.RebuyMaxCount)
			}
		} else if rebuyPhaseType == domain.PineappleRebuyPhaseAddon {
			fmt.Fprintf(&b, "アドオンしますか? (%dチップ) (ad=アドオン / sa=スキップ)\n",
				cfg.AddonChips)
		}
	}

	// エラーメッセージ
	if lastErr != nil {
		fmt.Fprintf(&b, "%s\n", color.Red(lastErr.Error()))
	}

	// ゲーム終了メッセージ
	if p.GetGameEndFlag() {
		b.WriteString("ゲーム終了\n")
	}

	return b.String()
}
