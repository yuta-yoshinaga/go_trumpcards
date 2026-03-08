package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// HoldemCuiPresenter テキサスホールデムCUIプレゼンタークラス
type HoldemCuiPresenter struct{}

// NewHoldemCuiPresenter コンストラクタ
func NewHoldemCuiPresenter() *HoldemCuiPresenter {
	return &HoldemCuiPresenter{}
}

// Output ゲーム状態を文字列出力
func (p *HoldemCuiPresenter) Output(h interfaces.HoldemGame, lastErr error) string {
	var b strings.Builder

	b.WriteString("==========\n")
	b.WriteString("Texas Hold'em\n")
	b.WriteString("==========\n")

	// トーナメントモードヘッダー
	cfg := h.GetConfig()
	if cfg.TournamentMode {
		fmt.Fprintf(&b, "トーナメント ハンド#%d SB:%d BB:%d (レベルアップ:%dハンド毎)\n",
			h.GetHandCount(), cfg.SmallBlind, cfg.BigBlind, cfg.BlindLevelHands)
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
	fmt.Fprintf(&b, "テーブル: %d-max\n", h.GetPlayerCnt())

	// ディーラー位置
	fmt.Fprintf(&b, "ディーラー: Player %d\n", h.GetDealerIdx())

	// コミュニティカード
	b.WriteString("コミュニティ: ")
	cc := h.GetCommunityCards()
	if len(cc) == 0 {
		b.WriteString("(なし)")
	} else {
		for i, card := range cc {
			if i > 0 {
				b.WriteString("  ")
			}
			b.WriteString(cuiCardStrEmoji(card))
		}
	}
	b.WriteString("\n")

	// ポット
	fmt.Fprintf(&b, "ポット: %d\n", h.GetPot())

	// ベッティングリミット
	if int(cfg.BettingLimit) < len(domain.BettingLimitNames) {
		fmt.Fprintf(&b, "リミット: %s\n", domain.BettingLimitNames[cfg.BettingLimit])
	}

	// プレイヤー情報
	b.WriteString("----------\n")
	for i := 0; i < h.GetPlayerCnt(); i++ {
		player := h.GetPlayer(i)
		if player.GetIsHuman() {
			b.WriteString("[You]")
		} else {
			fmt.Fprintf(&b, "CPU %d (%s)", i, player.GetPlayStyleName())
		}

		fmt.Fprintf(&b, " チップ:%d", player.GetChips())

		if player.GetTotalHands() > 0 {
			fmt.Fprintf(&b, " VPIP:%d%% PFR:%d%% 3Bet:%d%% AF:%s", player.GetVPIP(), player.GetPFR(), player.GetThreeBet(), player.GetAFDisplay())
		}

		if player.GetFolded() {
			b.WriteString(" [フォールド]")
		} else if player.GetAllIn() {
			b.WriteString(" [オールイン]")
		}

		if player.GetCurrentBet() > 0 {
			fmt.Fprintf(&b, " ベット:%d", player.GetCurrentBet())
		}
		b.WriteString("\n")

		// 人間のカードを表示
		if player.GetIsHuman() && !player.GetFolded() {
			b.WriteString("  手札: ")
			for j := 0; j < player.GetCardsSize(); j++ {
				if j > 0 {
					b.WriteString("  ")
				}
				b.WriteString(cuiCardStrEmoji(player.GetCard(j)))
			}
			b.WriteString("\n")
		}
	}

	// CPU行動記録
	cpuActions := h.GetCpuActions()
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

	// ショーダウン結果
	results := h.GetRoundResults()
	if len(results) > 0 && (h.GetPhase() == domain.HoldemPhaseEnd || h.GetPhase() == domain.HoldemPhaseShowdown) {
		b.WriteString("==========\n")
		b.WriteString("[結果]\n")
		for _, r := range results {
			name := "You"
			if !h.GetPlayer(r.PlayerIdx).GetIsHuman() {
				name = fmt.Sprintf("CPU %d", r.PlayerIdx)
			}
			if r.Mucked {
				fmt.Fprintf(&b, "  %s: マック", name)
			} else if r.HandName != "" {
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

	// マックプロンプト
	if h.IsMuckAvailable() {
		b.WriteString("----------\n")
		b.WriteString("マックしますか? (m=マック / sh=ショー)\n")
	}

	// リバイ/アドオンプロンプト
	if h.GetPhase() == domain.HoldemPhaseRebuy {
		b.WriteString("----------\n")
		rebuyPhaseType := h.GetRebuyPhaseType()
		if rebuyPhaseType == domain.HoldemRebuyPhaseRebuy {
			rebuyCounts := h.GetRebuyCounts()
			humanIdx := -1
			for i := 0; i < h.GetPlayerCnt(); i++ {
				if h.GetPlayer(i).GetIsHuman() {
					humanIdx = i
					break
				}
			}
			if humanIdx >= 0 {
				fmt.Fprintf(&b, "リバイしますか? (%dチップ, %d/%d回使用済) (rb=リバイ / sr=スキップ)\n",
					cfg.RebuyChips, rebuyCounts[humanIdx], cfg.RebuyMaxCount)
			}
		} else if rebuyPhaseType == domain.HoldemRebuyPhaseAddon {
			fmt.Fprintf(&b, "アドオンしますか? (%dチップ) (ad=アドオン / sa=スキップ)\n",
				cfg.AddonChips)
		}
	}

	// エラーメッセージ
	if lastErr != nil {
		fmt.Fprintf(&b, "[エラー] %s\n", lastErr.Error())
	}

	// ゲーム終了メッセージ
	if h.GetGameEndFlag() {
		b.WriteString("ゲーム終了\n")
	}

	return b.String()
}
