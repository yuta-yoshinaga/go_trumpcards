package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// napoleonPlayerStr returns the display string for a single Napoleon player.
func napoleonPlayerStr(player *domain.NapoleonPlayer, i int, adjRevealed bool) string {
	var b strings.Builder
	name := cuiPlayerName(player, i)
	bidStr := "未ビッド"
	if player.GetBid() == 0 {
		bidStr = "パス"
	} else if player.GetBid() > 0 {
		bidStr = fmt.Sprintf("%d", player.GetBid())
	}
	role := ""
	if player.GetIsNapoleon() {
		role = " [ナポレオン]"
	}
	if adjRevealed && player.GetIsAdjutant() {
		role = " [副官]"
	}
	fmt.Fprintf(&b, "%s%s: ビッド=%s 獲得%dトリック 絵札%d枚 累積%d点 ラウンド%d点 %d枚\n",
		name, role, bidStr,
		player.GetTrickCount(),
		player.GetPictureCards(),
		player.GetCumulativeScore(),
		player.GetRoundScore(),
		player.GetCardsSize(),
	)
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player))
		b.WriteString("\n")
	}
	return b.String()
}

// NapoleonCuiPresenter ナポレオンCUIプレゼンタークラス
type NapoleonCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *NapoleonCuiPresenter) Output(n interfaces.NapoleonGame, lastErr error) string {
	return buildCuiOutput("Napoleon (ナポレオン)", func(b *strings.Builder) {
		fmt.Fprintf(b, "ラウンド: %d  トリック: %d\n", n.GetRoundNumber(), n.GetTrickNumber())

		suitNames := map[int]string{
			domain.CardDesignSpade: "♠", domain.CardDesignClover: "♣",
			domain.CardDesignHeart: "♥", domain.CardDesignDiamond: "♦",
		}
		if n.GetTrumpSuit() > 0 {
			fmt.Fprintf(b, "切り札: %s\n", suitNames[n.GetTrumpSuit()])
		}

		adjCard := n.GetAdjutantCard()
		if adjCard != nil {
			fmt.Fprintf(b, "副官カード: %s", napoleonCuiCardStr(adjCard))
			if n.GetAdjutantRevealed() {
				b.WriteString(" (公開済み)")
			} else {
				b.WriteString(" (非公開)")
			}
			b.WriteString("\n")
		}

		if n.GetHighestBid() > 0 {
			fmt.Fprintf(b, "最高ビッド: %d\n", n.GetHighestBid())
		}

		// プレイヤー情報
		for i := 0; i < n.GetPlayerCnt(); i++ {
			b.WriteString(napoleonPlayerStr(n.GetPlayer(i), i, n.GetAdjutantRevealed()))
		}

		b.WriteString("----------\n")

		// 現在のトリック
		trick := n.GetCurrentTrick()
		cuiTrickBlock(b, trick,
			func(tc *domain.NapoleonTrickCard) int { return tc.PlayerIdx },
			func(tc *domain.NapoleonTrickCard) string { return napoleonCuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(n.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		// ゲーム状態
		if n.GetGameEndFlag() {
			winnerTeam := n.GetWinnerTeam()
			if winnerTeam == domain.NapoleonWinnerNapoleon {
				b.WriteString(color.Green("ゲーム終了！ ナポレオン軍の勝利です！") + "\n")
			} else {
				b.WriteString(color.Green("ゲーム終了！ 連合軍の勝利です！") + "\n")
			}
		} else {
			phase := n.GetPhase()
			switch phase {
			case domain.NapoleonPhaseBid:
				bidIdx := n.GetBidPlayerIdx()
				player := n.GetPlayer(bidIdx)
				fmt.Fprintf(b, "ビッドフェーズ: %sの番\n", cuiPlayerName(player, bidIdx))
				b.WriteString("b <n>・・・ビッドを宣言 (0=パス)\n")
			case domain.NapoleonPhaseTrumpDeclaration:
				b.WriteString("切り札宣言フェーズ\n")
				b.WriteString("t <suit> <adjSuit> <adjVal>・・・切り札と副官を宣言\n")
				b.WriteString("  suit: 1=♠ 2=♣ 3=♥ 4=♦\n")
			case domain.NapoleonPhaseKittyExchange:
				b.WriteString("場札交換フェーズ\n")
				b.WriteString("e <idx>・・・捨てるカードを選択\n")
			case domain.NapoleonPhasePlay:
				currentIdx := n.GetCurrentPlayerIdx()
				player := n.GetPlayer(currentIdx)
				fmt.Fprintf(b, "手番: %s\n", cuiPlayerName(player, currentIdx))
				b.WriteString("p <idx>・・・カードを出す\n")
			case domain.NapoleonPhaseTrickEnd:
				b.WriteString("トリック終了\n")
				b.WriteString("n / next・・・次のトリックへ\n")
			case domain.NapoleonPhaseRoundEnd:
				b.WriteString("ラウンド終了\n")
				b.WriteString("nr / nextround・・・次のラウンドへ\n")
			}
		}
	})
}

// HintOutput ヒント情報を出力する
func (p *NapoleonCuiPresenter) HintOutput(n interfaces.NapoleonGame) string {
	hint := n.GetHint()
	if hint == nil {
		return "ヒントはありません。\n"
	}
	if hint.Bid != nil {
		return fmt.Sprintf("%s\n", color.Yellow(fmt.Sprintf("[HINT: ビッド %d を推奨 (%s)]", *hint.Bid, napoleonHintReasonStr(hint.Reason))))
	}
	if hint.TrumpSuit != nil {
		suitNames := map[int]string{1: "♠", 2: "♣", 3: "♥", 4: "♦"}
		return fmt.Sprintf("%s\n", color.Yellow(fmt.Sprintf("[HINT: 切り札 %s を推奨 (%s)]", suitNames[*hint.TrumpSuit], napoleonHintReasonStr(hint.Reason))))
	}
	if hint.DiscardIndex != nil {
		player := n.GetPlayer(0)
		card := player.GetCard(*hint.DiscardIndex)
		return fmt.Sprintf("%s\n", color.Yellow(fmt.Sprintf("[HINT: [%d]%s を捨てる (%s)]", *hint.DiscardIndex, napoleonCuiCardStr(card), napoleonHintReasonStr(hint.Reason))))
	}
	if hint.CardIndex != nil {
		player := n.GetPlayer(0)
		card := player.GetCard(*hint.CardIndex)
		return fmt.Sprintf("%s\n", color.Yellow(fmt.Sprintf("[HINT: [%d]%s (%s)]", *hint.CardIndex, napoleonCuiCardStr(card), napoleonHintReasonStr(hint.Reason))))
	}
	return "ヒントはありません。\n"
}

// napoleonCuiCardStr ナポレオン用カード文字列表現 (ジョーカー対応)
func napoleonCuiCardStr(card *domain.Card) string {
	if card.GetDesign() == domain.CardDesignJoker {
		return "Joker"
	}
	return cuiCardStr(card)
}

// napoleonHintReasons はNapoleon固有のヒント理由翻訳
var napoleonHintReasons = map[string]string{
	"strategic_declare": "戦略的な宣言",
	"strategic_discard": "戦略的な捨て",
	"play_joker":        "ジョーカーをプレイ",
	"discard_low":       "低いカードを捨てる",
}

// napoleonHintReasonStr ヒント理由を日本語に変換する
func napoleonHintReasonStr(reason string) string {
	return lookupHintReason(reason, napoleonHintReasons)
}

// ActionLogOutput 棋譜をテキスト出力
func (p *NapoleonCuiPresenter) ActionLogOutput(n interfaces.NapoleonGame) string {
	return actionLogOutputText(n)
}
