package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// euchrePlayerStr returns the display string for a single Euchre player.
func euchrePlayerStr(player *domain.EuchrePlayer, i int) string {
	var b strings.Builder
	name := cuiPlayerName(player, i)
	fmt.Fprintf(&b, "%s: チーム%d 獲得%dトリック %d枚\n",
		name,
		player.GetTeam(),
		player.GetTrickCount(),
		player.GetCardsSize(),
	)
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player))
		b.WriteString("\n")
	}
	return b.String()
}

// EuchreCuiPresenter ユーカーCUIプレゼンタークラス
type EuchreCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *EuchreCuiPresenter) Output(e interfaces.EuchreGame, lastErr error) string {
	return buildCuiOutput("Euchre (ユーカー)", func(b *strings.Builder) {
		fmt.Fprintf(b, "ラウンド: %d  トリック: %d\n", e.GetRoundNumber(), e.GetTrickNumber())
		fmt.Fprintf(b, "ディーラー: %s\n", cuiPlayerName(e.GetPlayer(e.GetDealerIdx()), e.GetDealerIdx()))

		trumpSuit := e.GetTrumpSuit()
		if trumpSuit > 0 {
			fmt.Fprintf(b, "切り札: %s (メイカー: チーム%d)\n", cuiSuitName(trumpSuit), e.GetMakerTeam())
		} else {
			b.WriteString("切り札: 未決定\n")
		}

		faceUpCard := e.GetFaceUpCard()
		if faceUpCard != nil {
			fmt.Fprintf(b, "表向きカード: %s\n", cuiCardStr(faceUpCard))
		}

		if e.GetGoingAlone() {
			aloneIdx := e.GetGoingAlonePlayerIdx()
			fmt.Fprintf(b, "ゴーアローン: %s\n", cuiPlayerName(e.GetPlayer(aloneIdx), aloneIdx))
		}

		// チームスコア
		fmt.Fprintf(b, "チーム0: %d点  チーム1: %d点\n", e.GetTeamScore(0), e.GetTeamScore(1))

		// プレイヤー情報
		for i := 0; i < e.GetPlayerCnt(); i++ {
			b.WriteString(euchrePlayerStr(e.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		// 現在のトリック
		trick := e.GetCurrentTrick()
		if len(trick) > 0 {
			parts := make([]string, len(trick))
			for i, tc := range trick {
				player := e.GetPlayer(tc.PlayerIdx)
				parts[i] = fmt.Sprintf("%s=%s", cuiPlayerName(player, tc.PlayerIdx), cuiCardStr(tc.Card))
			}
			fmt.Fprintf(b, "トリック: %s\n", strings.Join(parts, ", "))
		}

		// エラーメッセージ
		if lastErr != nil {
			fmt.Fprintf(b, "%s\n", color.Red(lastErr.Error()))
		}

		// ゲーム状態
		if e.GetGameEndFlag() {
			winnerTeam := e.GetWinnerTeam()
			fmt.Fprintf(b, "ゲーム終了！ %s\n", color.Green(fmt.Sprintf("チーム%dの勝利です！", winnerTeam)))
		} else {
			phase := e.GetPhase()
			switch phase {
			case domain.EuchrePhasePickUp:
				bidIdx := e.GetBidPlayerIdx()
				player := e.GetPlayer(bidIdx)
				fmt.Fprintf(b, "ピックアップフェーズ: %sの番\n", cuiPlayerName(player, bidIdx))
				b.WriteString("o (order up) / oa (order up alone) / pa (pass)\n")
			case domain.EuchrePhaseCallTrump:
				bidIdx := e.GetBidPlayerIdx()
				player := e.GetPlayer(bidIdx)
				fmt.Fprintf(b, "コールトランプフェーズ: %sの番\n", cuiPlayerName(player, bidIdx))
				b.WriteString("c <suit> (call) / ca <suit> (call alone) / pa (pass)\n")
			case domain.EuchrePhaseDiscard:
				b.WriteString("ディスカードフェーズ\n")
				b.WriteString("d <i> (discard)\n")
			case domain.EuchrePhasePlay:
				currentIdx := e.GetCurrentPlayerIdx()
				player := e.GetPlayer(currentIdx)
				fmt.Fprintf(b, "手番: %s\n", cuiPlayerName(player, currentIdx))
				b.WriteString("p <i> (play)\n")
			case domain.EuchrePhaseTrickEnd:
				b.WriteString("トリック終了\n")
				b.WriteString("n (next trick)\n")
			case domain.EuchrePhaseRoundEnd:
				b.WriteString("ラウンド終了\n")
				b.WriteString("nr (next round)\n")
			}
		}
	})
}

// HintOutput ヒント情報を出力する
func (p *EuchreCuiPresenter) HintOutput(e interfaces.EuchreGame) string {
	hint := e.GetHint()
	if hint == nil {
		return "ヒントはありません。\n"
	}
	if hint.OrderUp != nil {
		if *hint.OrderUp {
			alone := ""
			if hint.GoAlone != nil && *hint.GoAlone {
				alone = " (ゴーアローン)"
			}
			return fmt.Sprintf("%s\n", color.Yellow(fmt.Sprintf("[HINT: オーダーアップ%s (%s)]", alone, euchreHintReasonStr(hint.Reason))))
		}
		return fmt.Sprintf("%s\n", color.Yellow(fmt.Sprintf("[HINT: パス (%s)]", euchreHintReasonStr(hint.Reason))))
	}
	if hint.Suit != nil {
		alone := ""
		if hint.GoAlone != nil && *hint.GoAlone {
			alone = " (ゴーアローン)"
		}
		return fmt.Sprintf("%s\n", color.Yellow(fmt.Sprintf("[HINT: %sをコール%s (%s)]", cuiSuitName(*hint.Suit), alone, euchreHintReasonStr(hint.Reason))))
	}
	if hint.CardIndex == nil {
		return "ヒントはありません。\n"
	}
	player := e.GetPlayer(0)
	card := player.GetCard(*hint.CardIndex)
	return fmt.Sprintf("%s\n", color.Yellow(fmt.Sprintf("[HINT: [%d]%s (%s)]", *hint.CardIndex, cuiCardStr(card), euchreHintReasonStr(hint.Reason))))
}

// euchreHintReasonStr ヒント理由を日本語に変換する
func euchreHintReasonStr(reason string) string {
	switch reason {
	case "strong_hand":
		return "強い手札"
	case "weak_hand":
		return "弱い手札"
	case "follow_suit":
		return "リードスートに追随"
	case "trump_cut":
		return "切り札でカット"
	case "discard_weakest":
		return "最弱カードを捨てる"
	case "lead_strong":
		return "強いカードでリード"
	case "lead_low":
		return "低いカードでリード"
	default:
		return reason
	}
}

// ActionLogOutput 棋譜をテキスト出力
func (p *EuchreCuiPresenter) ActionLogOutput(e interfaces.EuchreGame) string {
	return actionLogOutputText(e)
}
