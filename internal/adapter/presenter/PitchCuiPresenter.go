package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// pitchPlayerStr returns the display string for a single Pitch player.
func pitchPlayerStr(player *domain.PitchPlayer, i int) string {
	var b strings.Builder
	name := cuiPlayerName(player, i)
	bidStr := "未ビッド"
	if player.GetBid() == 0 {
		bidStr = "pass"
	} else if player.GetBid() > 0 {
		bidStr = fmt.Sprintf("%d", player.GetBid())
	}
	fmt.Fprintf(&b, "%s: ビッド=%s 獲得%dトリック 累積%d点 ラウンド%d点 %d枚\n",
		name,
		bidStr,
		player.GetTrickCount(),
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

// pitchSuitName Pitch CUI 用のスート名 (絵文字)
func pitchSuitName(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return "♠"
	case domain.CardDesignClover:
		return "♣"
	case domain.CardDesignHeart:
		return "♥"
	case domain.CardDesignDiamond:
		return "♦"
	}
	return "(未確定)"
}

// PitchCuiPresenter ピッチCUIプレゼンタークラス
type PitchCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *PitchCuiPresenter) Output(s interfaces.PitchGame, lastErr error) string {
	return buildCuiOutput("Pitch (ピッチ)", func(b *strings.Builder) {
		fmt.Fprintf(b, "ラウンド: %d  トリック: %d  親: %s\n",
			s.GetRoundNumber(), s.GetTrickNumber(), cuiPlayerName(s.GetPlayer(s.GetDealerIdx()), s.GetDealerIdx()))
		fmt.Fprintf(b, "ビッド: %d  トランプ: %s\n",
			s.GetCurrentBid(), pitchSuitName(s.GetTrumpSuit()))
		if s.GetBidWinnerIdx() >= 0 {
			fmt.Fprintf(b, "ビッド勝者: %s\n",
				cuiPlayerName(s.GetPlayer(s.GetBidWinnerIdx()), s.GetBidWinnerIdx()))
		}

		// プレイヤー情報
		for i := 0; i < s.GetPlayerCnt(); i++ {
			b.WriteString(pitchPlayerStr(s.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		// 現在のトリック
		trick := s.GetCurrentTrick()
		cuiTrickBlock(b, trick,
			func(tc *domain.PitchTrickCard) int { return tc.PlayerIdx },
			func(tc *domain.PitchTrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(s.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		// ゲーム状態
		if s.GetGameEndFlag() {
			winnerIdx := s.GetWinnerIdx()
			player := s.GetPlayer(winnerIdx)
			fmt.Fprintf(b, "ゲーム終了！ %s\n", color.Green(cuiPlayerName(player, winnerIdx)+"の勝利です！"))
		} else {
			phase := s.GetPhase()
			switch phase {
			case domain.PitchPhaseBid:
				bidIdx := s.GetBidPlayerIdx()
				player := s.GetPlayer(bidIdx)
				fmt.Fprintf(b, "ビッドフェーズ: %sの番\n", cuiPlayerName(player, bidIdx))
				b.WriteString("b <n>・・・ビッドを宣言 (0=pass, 2-4)\n")
			case domain.PitchPhasePlay:
				currentIdx := s.GetCurrentPlayerIdx()
				player := s.GetPlayer(currentIdx)
				fmt.Fprintf(b, "手番: %s\n", cuiPlayerName(player, currentIdx))
				b.WriteString("play <idx>・・・カードを出す\n")
			case domain.PitchPhaseTrickEnd:
				b.WriteString("トリック終了\n")
				b.WriteString("next・・・次のトリックへ\n")
			case domain.PitchPhaseRoundEnd:
				b.WriteString("ラウンド終了\n")
				b.WriteString("nr / nextround・・・次のラウンドへ\n")
			}
		}
	})
}

// HintOutput ヒント情報を出力する
func (p *PitchCuiPresenter) HintOutput(s interfaces.PitchGame) string {
	hint := s.GetHint()
	if hint == nil {
		return "ヒントはありません。\n"
	}
	if hint.Bid != nil {
		bidStr := "pass"
		if *hint.Bid > 0 {
			bidStr = fmt.Sprintf("%d", *hint.Bid)
		}
		return fmt.Sprintf("%s\n", color.Yellow(fmt.Sprintf("[HINT: ビッド %s を推奨 (%s)]", bidStr, pitchHintReasonStr(hint.Reason))))
	}
	if hint.CardIndex == nil {
		return "ヒントはありません。\n"
	}
	humanIdx := -1
	for i := 0; i < s.GetPlayerCnt(); i++ {
		if pl := s.GetPlayer(i); pl != nil && pl.GetIsHuman() {
			humanIdx = i
			break
		}
	}
	if humanIdx < 0 {
		return "ヒントはありません。\n"
	}
	player := s.GetPlayer(humanIdx)
	card := player.GetCard(*hint.CardIndex)
	return fmt.Sprintf("%s\n", color.Yellow(fmt.Sprintf("[HINT: [%d]%s (%s)]", *hint.CardIndex, cuiCardStr(card), pitchHintReasonStr(hint.Reason))))
}

// pitchHintReasons はPitch固有のヒント理由翻訳
var pitchHintReasons = map[string]string{
	"set_trump_lead": "リードでトランプを宣言",
	"trump_cut":      "トランプでカット",
	"bid_strong":     "強い手札なので入札",
	"bid_pass":       "弱い手札なのでパス",
}

// pitchHintReasonStr ヒント理由を日本語に変換する
func pitchHintReasonStr(reason string) string {
	return lookupHintReason(reason, pitchHintReasons)
}

// ActionLogOutput 棋譜をテキスト出力
func (p *PitchCuiPresenter) ActionLogOutput(s interfaces.PitchGame) string {
	return actionLogOutputText(s)
}
