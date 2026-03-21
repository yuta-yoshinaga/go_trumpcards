package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// OmahaCuiPresenter オマハホールデムCUIプレゼンタークラス
type OmahaCuiPresenter struct{}

// ActionLogOutput 棋譜をテキスト出力
func (p *OmahaCuiPresenter) ActionLogOutput(o interfaces.OmahaGame) string {
	return actionLogOutputText(o)
}

// Output ゲーム状態を文字列出力
func (p *OmahaCuiPresenter) Output(o interfaces.OmahaGame, lastErr error) string {
	var b strings.Builder

	b.WriteString("==========\n")
	b.WriteString("Omaha Hold'em\n")
	b.WriteString("==========\n")

	cfg := o.GetConfig()
	if cfg.TournamentMode {
		fmt.Fprintf(&b, "トーナメント ハンド#%d SB:%d BB:%d (レベルアップ:%dハンド毎)\n",
			o.GetHandCount(), cfg.SmallBlind, cfg.BigBlind, cfg.BlindLevelHands)
		if cfg.RebuyEnabled {
			fmt.Fprintf(&b, "リバイ: %dチップ (最大%d回, %dハンド目まで)\n",
				cfg.RebuyChips, cfg.RebuyMaxCount, cfg.RebuyPeriodHands)
		}
		if cfg.AddonEnabled {
			fmt.Fprintf(&b, "アドオン: %dチップ (%dハンド目に提供)\n",
				cfg.AddonChips, cfg.AddonAfterHand)
		}
	}

	fmt.Fprintf(&b, "テーブル: %d-max\n", o.GetPlayerCnt())
	fmt.Fprintf(&b, "ディーラー: Player %d\n", o.GetDealerIdx())

	cc := o.GetCommunityCards()
	if len(cc) == 0 {
		b.WriteString("コミュニティ: (なし)\n")
	} else {
		fmt.Fprintf(&b, "コミュニティ: %s\n", cuiCardSliceStrEmoji(cc))
	}

	fmt.Fprintf(&b, "ポット: %d\n", o.GetPot())

	if int(cfg.BettingLimit) < len(domain.BettingLimitNames) {
		fmt.Fprintf(&b, "リミット: %s\n", domain.BettingLimitNames[cfg.BettingLimit])
	}

	b.WriteString("----------\n")
	for i := 0; i < o.GetPlayerCnt(); i++ {
		player := o.GetPlayer(i)
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

		if player.GetIsHuman() && !player.GetFolded() {
			fmt.Fprintf(&b, "  手札: %s\n", cuiCardListStrEmoji(player))
		}
	}

	cpuActions := o.GetCpuActions()
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

	results := o.GetRoundResults()
	if len(results) > 0 && (o.GetPhase() == domain.OmahaPhaseEnd || o.GetPhase() == domain.OmahaPhaseShowdown) {
		b.WriteString("==========\n")
		b.WriteString(color.Bold("[結果]") + "\n")
		for _, r := range results {
			name := cuiPlayerName(o.GetPlayer(r.PlayerIdx), r.PlayerIdx)
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

	if o.IsMuckAvailable() {
		b.WriteString("----------\n")
		b.WriteString("マックしますか? (m=マック / sh=ショー)\n")
	}

	if o.GetPhase() == domain.OmahaPhaseRebuy {
		b.WriteString("----------\n")
		rebuyPhaseType := o.GetRebuyPhaseType()
		if rebuyPhaseType == domain.OmahaRebuyPhaseRebuy {
			rebuyCounts := o.GetRebuyCounts()
			humanIdx := -1
			for i := 0; i < o.GetPlayerCnt(); i++ {
				if o.GetPlayer(i).GetIsHuman() {
					humanIdx = i
					break
				}
			}
			if humanIdx >= 0 {
				fmt.Fprintf(&b, "リバイしますか? (%dチップ, %d/%d回使用済) (rb=リバイ / sr=スキップ)\n",
					cfg.RebuyChips, rebuyCounts[humanIdx], cfg.RebuyMaxCount)
			}
		} else if rebuyPhaseType == domain.OmahaRebuyPhaseAddon {
			fmt.Fprintf(&b, "アドオンしますか? (%dチップ) (ad=アドオン / sa=スキップ)\n",
				cfg.AddonChips)
		}
	}

	if lastErr != nil {
		fmt.Fprintf(&b, "%s\n", color.Red(lastErr.Error()))
	}

	if o.GetGameEndFlag() {
		b.WriteString("ゲーム終了\n")
	}

	return b.String()
}
