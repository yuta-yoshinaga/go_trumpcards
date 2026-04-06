package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SevenCardStudCuiPresenter セブンカードスタッドCUIプレゼンタークラス
type SevenCardStudCuiPresenter struct{}

// ActionLogOutput 棋譜をテキスト出力
func (p *SevenCardStudCuiPresenter) ActionLogOutput(s interfaces.SevenCardStudGame) string {
	return actionLogOutputText(s)
}

// Output ゲーム状態を文字列出力
func (p *SevenCardStudCuiPresenter) Output(s interfaces.SevenCardStudGame, lastErr error) string {
	var b strings.Builder

	b.WriteString("==========\n")
	b.WriteString("Seven Card Stud\n")
	b.WriteString("==========\n")

	// トーナメントモードヘッダー
	cfg := s.GetConfig()
	if cfg.TournamentMode {
		fmt.Fprintf(&b, "トーナメント ハンド#%d Ante:%d BringIn:%d (レベルアップ:%dハンド毎)\n",
			s.GetHandCount(), cfg.Ante, cfg.BringIn, cfg.AnteLevelHands)
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
	fmt.Fprintf(&b, "テーブル: %d-max\n", s.GetPlayerCnt())

	// ディーラー位置
	fmt.Fprintf(&b, "ディーラー: Player %d\n", s.GetDealerIdx())

	// アンティ・ブリングイン情報
	fmt.Fprintf(&b, "Ante:%d BringIn:%d SmallBet:%d BigBet:%d\n", cfg.Ante, cfg.BringIn, cfg.SmallBet, cfg.BigBet)

	// ポット
	fmt.Fprintf(&b, "ポット: %d\n", s.GetPot())

	// ベッティングリミット
	if int(cfg.BettingLimit) < len(domain.BettingLimitNames) {
		fmt.Fprintf(&b, "リミット: %s\n", domain.BettingLimitNames[cfg.BettingLimit])
	}

	// プレイヤー情報
	b.WriteString("----------\n")
	for i := 0; i < s.GetPlayerCnt(); i++ {
		player := s.GetPlayer(i)
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

		// ドアカード (表向き — 全プレイヤー表示)
		doorCards := player.GetDoorCards()
		if len(doorCards) > 0 {
			fmt.Fprintf(&b, "  表札: %s\n", cuiCardSliceStrEmoji(doorCards))
		}

		// ホールカード (伏せ札 — 人間のみ表示)
		if player.GetIsHuman() && !player.GetFolded() {
			holeCards := player.GetHoleCards()
			if len(holeCards) > 0 {
				fmt.Fprintf(&b, "  伏せ札: %s\n", cuiCardSliceStrEmoji(holeCards))
			}
		}
	}

	// CPU行動記録
	cpuActions := s.GetCpuActions()
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
	results := s.GetRoundResults()
	if len(results) > 0 && (s.GetPhase() == domain.SevenCardStudPhaseEnd || s.GetPhase() == domain.SevenCardStudPhaseShowdown) {
		b.WriteString("==========\n")
		b.WriteString(color.Bold("[結果]") + "\n")
		for _, r := range results {
			name := cuiPlayerName(s.GetPlayer(r.PlayerIdx), r.PlayerIdx)
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
	if s.IsMuckAvailable() {
		b.WriteString("----------\n")
		b.WriteString("マックしますか? (m=マック / sh=ショー)\n")
	}

	// リバイ/アドオンプロンプト
	if s.GetPhase() == domain.SevenCardStudPhaseRebuy {
		b.WriteString("----------\n")
		rebuyPhaseType := s.GetRebuyPhaseType()
		if rebuyPhaseType == domain.SevenCardStudRebuyPhaseRebuy {
			rebuyCounts := s.GetRebuyCounts()
			humanIdx := -1
			for i := 0; i < s.GetPlayerCnt(); i++ {
				if s.GetPlayer(i).GetIsHuman() {
					humanIdx = i
					break
				}
			}
			if humanIdx >= 0 {
				fmt.Fprintf(&b, "リバイしますか? (%dチップ, %d/%d回使用済) (rb=リバイ / sr=スキップ)\n",
					cfg.RebuyChips, rebuyCounts[humanIdx], cfg.RebuyMaxCount)
			}
		} else if rebuyPhaseType == domain.SevenCardStudRebuyPhaseAddon {
			fmt.Fprintf(&b, "アドオンしますか? (%dチップ) (ad=アドオン / sa=スキップ)\n",
				cfg.AddonChips)
		}
	}

	// エラーメッセージ
	if lastErr != nil {
		fmt.Fprintf(&b, "%s\n", color.Red(lastErr.Error()))
	}

	// ゲーム終了メッセージ
	if s.GetGameEndFlag() {
		b.WriteString("ゲーム終了\n")
	}

	return b.String()
}
