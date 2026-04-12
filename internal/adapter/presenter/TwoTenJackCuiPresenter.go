package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// twoTenJackPlayerStr returns the display string for a single TwoTenJack player.
func twoTenJackPlayerStr(player *domain.TwoTenJackPlayer, i int) string {
	var b strings.Builder
	name := cuiPlayerName(player, i)
	team := i % 2
	fmt.Fprintf(&b, "%s (チーム%d): 獲得%dトリック 点札%d点 累積%d点 ラウンド%d点 %d枚\n",
		name,
		team,
		player.GetTrickCount(),
		player.GetCapturedPointCards(),
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

// twoTenJackSuitLabel returns a human-readable suit label.
func twoTenJackSuitLabel(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return "SPADE"
	case domain.CardDesignClover:
		return "CLUB"
	case domain.CardDesignHeart:
		return "HEART"
	case domain.CardDesignDiamond:
		return "DIAMOND"
	}
	return "未宣言"
}

// TwoTenJackCuiPresenter ツーテンジャックCUIプレゼンタークラス
type TwoTenJackCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *TwoTenJackCuiPresenter) Output(s interfaces.TwoTenJackGame, lastErr error) string {
	return buildCuiOutput("Two Ten Jack (ツーテンジャック)", func(b *strings.Builder) {
		fmt.Fprintf(b, "ラウンド: %d  トリック: %d\n", s.GetRoundNumber(), s.GetTrickNumber())
		fmt.Fprintf(b, "トランプ: %s  宣言者: %d\n", twoTenJackSuitLabel(s.GetTrumpSuit()), s.GetDeclarerIdx())

		for i := 0; i < s.GetPlayerCnt(); i++ {
			b.WriteString(twoTenJackPlayerStr(s.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		trick := s.GetCurrentTrick()
		cuiTrickBlock(b, trick,
			func(tc *domain.TwoTenJackTrickCard) int { return tc.PlayerIdx },
			func(tc *domain.TwoTenJackTrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(s.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		if s.GetGameEndFlag() {
			team := s.GetWinnerTeam()
			fmt.Fprintf(b, "ゲーム終了！ %s\n", color.Green(fmt.Sprintf("チーム%dの勝利です！", team)))
		} else {
			phase := s.GetPhase()
			switch phase {
			case domain.TwoTenJackPhaseDeclare:
				declIdx := s.GetDeclarerIdx()
				player := s.GetPlayer(declIdx)
				fmt.Fprintf(b, "宣言フェーズ: %sの番\n", cuiPlayerName(player, declIdx))
				b.WriteString("d <s>・・・トランプ宣言 (1=S, 2=C, 3=H, 4=D)\n")
			case domain.TwoTenJackPhasePlay:
				currentIdx := s.GetCurrentPlayerIdx()
				player := s.GetPlayer(currentIdx)
				fmt.Fprintf(b, "手番: %s\n", cuiPlayerName(player, currentIdx))
				b.WriteString("play <idx>・・・カードを出す\n")
			case domain.TwoTenJackPhaseTrickEnd:
				b.WriteString("トリック終了\n")
				b.WriteString("next・・・次のトリックへ\n")
			case domain.TwoTenJackPhaseRoundEnd:
				b.WriteString("ラウンド終了\n")
				b.WriteString("nr / nextround・・・次のラウンドへ\n")
			}
		}
	})
}

// HintOutput ヒント情報を出力する
func (p *TwoTenJackCuiPresenter) HintOutput(s interfaces.TwoTenJackGame) string {
	hint := s.GetHint()
	if hint == nil {
		return "ヒントはありません。\n"
	}
	if hint.TrumpSuit != nil {
		return fmt.Sprintf("%s\n", color.Yellow(fmt.Sprintf("[HINT: トランプ %s を推奨 (%s)]", twoTenJackSuitLabel(*hint.TrumpSuit), twoTenJackHintReasonStr(hint.Reason))))
	}
	if hint.CardIndex == nil {
		return "ヒントはありません。\n"
	}
	player := s.GetPlayer(0)
	card := player.GetCard(*hint.CardIndex)
	return fmt.Sprintf("%s\n", color.Yellow(fmt.Sprintf("[HINT: [%d]%s (%s)]", *hint.CardIndex, cuiCardStr(card), twoTenJackHintReasonStr(hint.Reason))))
}

// twoTenJackHintReasons はTwoTenJack固有のヒント理由翻訳
var twoTenJackHintReasons = map[string]string{
	"strategic_trump": "長いスートを宣言",
	"lead":            "リード",
	"follow_suit":     "リードスートに追随",
	"trump_cut":       "トランプでカット",
	"discard":         "不要カードを捨てる",
}

// twoTenJackHintReasonStr ヒント理由を日本語に変換する
func twoTenJackHintReasonStr(reason string) string {
	return lookupHintReason(reason, twoTenJackHintReasons)
}

// ActionLogOutput 棋譜をテキスト出力
func (p *TwoTenJackCuiPresenter) ActionLogOutput(s interfaces.TwoTenJackGame) string {
	return actionLogOutputText(s)
}
