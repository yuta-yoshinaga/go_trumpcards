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
	}

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
			b.WriteString(p.getCardStr(card))
		}
	}
	b.WriteString("\n")

	// ポット
	fmt.Fprintf(&b, "ポット: %d\n", h.GetPot())

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
			fmt.Fprintf(&b, " VPIP:%d%% PFR:%d%%", player.GetVPIP(), player.GetPFR())
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
				b.WriteString(p.getCardStr(player.GetCard(j)))
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
			fmt.Fprintf(&b, "  Player %d: %s", action.PlayerIdx, p.getActionName(action.Action))
			if action.Amount > 0 {
				fmt.Fprintf(&b, " (%d)", action.Amount)
			}
			b.WriteString("\n")
		}
	}

	// ショーダウン結果
	results := h.GetRoundResults()
	if len(results) > 0 && h.GetPhase() == domain.HoldemPhaseEnd {
		b.WriteString("==========\n")
		b.WriteString("[結果]\n")
		for _, r := range results {
			name := "You"
			if !h.GetPlayer(r.PlayerIdx).GetIsHuman() {
				name = fmt.Sprintf("CPU %d", r.PlayerIdx)
			}
			if r.HandName != "" {
				fmt.Fprintf(&b, "  %s: %s", name, r.HandName)
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
	if h.GetGameEndFlag() {
		b.WriteString("ゲーム終了\n")
	}

	return b.String()
}

// getCardStr カード文字列取得
func (p *HoldemCuiPresenter) getCardStr(card *domain.Card) string {
	designs := []string{"🃏", "♠", "♣", "♥", "♦"}
	d := card.GetDesign()
	if d < 0 || d >= len(designs) {
		d = 0
	}
	return fmt.Sprintf("%s%d", designs[d], card.GetValue())
}

// getActionName アクション名取得
func (p *HoldemCuiPresenter) getActionName(action int) string {
	switch action {
	case domain.HoldemActionFold:
		return "フォールド"
	case domain.HoldemActionCheck:
		return "チェック"
	case domain.HoldemActionCall:
		return "コール"
	case domain.HoldemActionBet:
		return "ベット"
	case domain.HoldemActionRaise:
		return "レイズ"
	case domain.HoldemActionAllIn:
		return "オールイン"
	default:
		return "不明"
	}
}
