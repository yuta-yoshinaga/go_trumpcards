package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// bridgePlayerStr returns the display string for a single Bridge player.
func bridgePlayerStr(player *domain.BridgePlayer, i int) string {
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

// BridgeCuiPresenter ブリッジCUIプレゼンタークラス
type BridgeCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *BridgeCuiPresenter) Output(b interfaces.BridgeGame, lastErr error) string {
	return buildCuiOutput("Contract Bridge (コントラクトブリッジ)", func(sb *strings.Builder) {
		fmt.Fprintf(sb, "ラウンド: %d  トリック: %d\n", b.GetRoundNumber(), b.GetTrickNumber())
		fmt.Fprintf(sb, "ディーラー: %s\n", cuiPlayerName(b.GetPlayer(b.GetDealerIdx()), b.GetDealerIdx()))

		trumpSuit := b.GetTrumpSuit()
		if trumpSuit == -1 {
			sb.WriteString("切り札: ノートランプ\n")
		} else if trumpSuit > 0 {
			fmt.Fprintf(sb, "切り札: %s\n", cuiSuitName(trumpSuit))
		} else {
			sb.WriteString("切り札: 未決定\n")
		}

		contractLevel := b.GetContractLevel()
		if contractLevel > 0 {
			fmt.Fprintf(sb, "コントラクト: %dレベル スート%d", contractLevel, b.GetContractSuit())
			switch b.GetDoubled() {
			case 1:
				sb.WriteString(" ダブル")
			case 2:
				sb.WriteString(" リダブル")
			}
			sb.WriteString("\n")

			declarerIdx := b.GetDeclarerIdx()
			if declarerIdx >= 0 {
				fmt.Fprintf(sb, "デクレアラー: %s  ダミー: %s\n",
					cuiPlayerName(b.GetPlayer(declarerIdx), declarerIdx),
					cuiPlayerName(b.GetPlayer(b.GetDummyIdx()), b.GetDummyIdx()))
			}
		}

		// バルネラビリティ
		fmt.Fprintf(sb, "バルネラビリティ: チーム0=%v チーム1=%v\n", b.GetVulnerability(0), b.GetVulnerability(1))

		// チームスコア
		fmt.Fprintf(sb, "チーム0: %d点 (ゲーム%d勝 ライン下%d)  チーム1: %d点 (ゲーム%d勝 ライン下%d)\n",
			b.GetTeamScore(0), b.GetGamesWon(0), b.GetBelowLine(0),
			b.GetTeamScore(1), b.GetGamesWon(1), b.GetBelowLine(1))

		// プレイヤー情報
		for i := 0; i < b.GetPlayerCnt(); i++ {
			sb.WriteString(bridgePlayerStr(b.GetPlayer(i), i))
		}

		// ダミーの手札 (公開後)
		if b.IsOpeningLeadDone() {
			dummyHand := b.GetDummyHand()
			if len(dummyHand) > 0 {
				parts := make([]string, len(dummyHand))
				for i, c := range dummyHand {
					parts[i] = cuiCardStr(c)
				}
				fmt.Fprintf(sb, "ダミー手札: %s\n", strings.Join(parts, ", "))
			}
		}

		sb.WriteString("----------\n")

		// 現在のトリック
		trick := b.GetCurrentTrick()
		cuiTrickBlock(sb, trick,
			func(tc *domain.BridgeTrickCard) int { return tc.PlayerIdx },
			func(tc *domain.BridgeTrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(b.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(sb, lastErr)

		// ゲーム状態
		if b.GetGameEndFlag() {
			winnerTeam := b.GetWinnerTeam()
			fmt.Fprintf(sb, "ゲーム終了！ %s\n", color.Green(fmt.Sprintf("チーム%dの勝利です！", winnerTeam)))
		} else {
			phase := b.GetPhase()
			switch phase {
			case domain.BridgePhaseBid:
				bidIdx := b.GetBidPlayerIdx()
				player := b.GetPlayer(bidIdx)
				fmt.Fprintf(sb, "ビッドフェーズ: %sの番\n", cuiPlayerName(player, bidIdx))
				sb.WriteString("b <type> [level] [suit] (bid: 0=Pass, 1=Normal, 2=Double, 3=Redouble)\n")
			case domain.BridgePhasePlay:
				currentIdx := b.GetCurrentPlayerIdx()
				player := b.GetPlayer(currentIdx)
				fmt.Fprintf(sb, "手番: %s\n", cuiPlayerName(player, currentIdx))
				sb.WriteString("p <i> (play)\n")
			case domain.BridgePhaseTrickEnd:
				sb.WriteString("トリック終了\n")
				sb.WriteString("n (next trick)\n")
			case domain.BridgePhaseRoundEnd:
				sb.WriteString("ラウンド終了\n")
				sb.WriteString("nr (next round)\n")
			}
		}
	})
}

// HintOutput ヒント情報を出力する
func (p *BridgeCuiPresenter) HintOutput(b interfaces.BridgeGame) string {
	hint := b.GetHint()
	if hint == nil {
		return "ヒントはありません。\n"
	}
	if hint.BidType != nil {
		bidTypeStr := bridgeBidTypeStr(*hint.BidType)
		if *hint.BidType == int(domain.BridgeBidNormal) && hint.BidLevel != nil && hint.BidSuit != nil {
			return fmt.Sprintf("%s\n", color.Yellow(fmt.Sprintf("[HINT: %s %dレベル スート%d (%s)]", bidTypeStr, *hint.BidLevel, *hint.BidSuit, bridgeHintReasonStr(hint.Reason))))
		}
		return fmt.Sprintf("%s\n", color.Yellow(fmt.Sprintf("[HINT: %s (%s)]", bidTypeStr, bridgeHintReasonStr(hint.Reason))))
	}
	if hint.CardIndex == nil {
		return "ヒントはありません。\n"
	}
	player := b.GetPlayer(0)
	card := player.GetCard(*hint.CardIndex)
	return fmt.Sprintf("%s\n", color.Yellow(fmt.Sprintf("[HINT: [%d]%s (%s)]", *hint.CardIndex, cuiCardStr(card), bridgeHintReasonStr(hint.Reason))))
}

// bridgeBidTypeStr ビッドタイプを日本語に変換する
func bridgeBidTypeStr(bidType int) string {
	switch domain.BridgeBidType(bidType) {
	case domain.BridgeBidPass:
		return "パス"
	case domain.BridgeBidNormal:
		return "ビッド"
	case domain.BridgeBidDouble:
		return "ダブル"
	case domain.BridgeBidRedouble:
		return "リダブル"
	default:
		return fmt.Sprintf("不明(%d)", bidType)
	}
}

// bridgeHintReasons はBridge固有のヒント理由翻訳
var bridgeHintReasons = map[string]string{
	"support_partner": "パートナーをサポート",
	"competitive_bid": "競り合い",
}

// bridgeHintReasonStr ヒント理由を日本語に変換する
func bridgeHintReasonStr(reason string) string {
	return lookupHintReason(reason, bridgeHintReasons)
}

// ActionLogOutput 棋譜をテキスト出力
func (p *BridgeCuiPresenter) ActionLogOutput(b interfaces.BridgeGame) string {
	return actionLogOutputText(b)
}
