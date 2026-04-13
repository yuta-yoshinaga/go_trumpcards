package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// whistPlayerStr returns the display string for a single Whist player.
func whistPlayerStr(player *domain.WhistPlayer, i int) string {
	var b strings.Builder
	name := cuiPlayerName(player, i)
	fmt.Fprintf(&b, "%s (チーム%d): 獲得%dトリック 累積%d点 ラウンド%d点 %d枚\n",
		name,
		player.GetTeam(),
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

// WhistCuiPresenter ホイストCUIプレゼンタークラス
type WhistCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *WhistCuiPresenter) Output(w interfaces.WhistGame, lastErr error) string {
	return buildCuiOutput("Whist (ホイスト)", func(b *strings.Builder) {
		fmt.Fprintf(b, "ラウンド: %d  トリック: %d\n", w.GetRoundNumber(), w.GetTrickNumber())
		fmt.Fprintf(b, "トランプ: %s\n", suitDisplayName(w.GetTrumpSuit()))
		fmt.Fprintf(b, "チームスコア: チーム0=%d  チーム1=%d\n", w.GetTeamScore(0), w.GetTeamScore(1))

		// プレイヤー情報
		for i := 0; i < w.GetPlayerCnt(); i++ {
			b.WriteString(whistPlayerStr(w.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		// 現在のトリック
		trick := w.GetCurrentTrick()
		cuiTrickBlock(b, trick,
			func(tc *domain.WhistTrickCard) int { return tc.PlayerIdx },
			func(tc *domain.WhistTrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(w.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		// ゲーム状態
		if w.GetGameEndFlag() {
			winnerTeam := w.GetWinnerTeam()
			fmt.Fprintf(b, "ゲーム終了！ %s\n", color.Green(fmt.Sprintf("チーム%dの勝利です！", winnerTeam)))
		} else {
			phase := w.GetPhase()
			switch phase {
			case domain.WhistPhasePlay:
				currentIdx := w.GetCurrentPlayerIdx()
				player := w.GetPlayer(currentIdx)
				fmt.Fprintf(b, "手番: %s\n", cuiPlayerName(player, currentIdx))
				b.WriteString("play <idx>・・・カードを出す\n")
			case domain.WhistPhaseTrickEnd:
				b.WriteString("トリック終了\n")
				b.WriteString("next・・・次のトリックへ\n")
			case domain.WhistPhaseRoundEnd:
				b.WriteString("ラウンド終了\n")
				b.WriteString("nr / nextround・・・次のラウンドへ\n")
			}
		}
	})
}

// HintOutput ヒント情報を出力する
func (p *WhistCuiPresenter) HintOutput(w interfaces.WhistGame) string {
	hint := w.GetHint()
	if hint == nil {
		return "ヒントはありません。\n"
	}
	if hint.CardIndex == nil {
		return "ヒントはありません。\n"
	}
	player := w.GetPlayer(0)
	card := player.GetCard(*hint.CardIndex)
	return fmt.Sprintf("%s\n", color.Yellow(fmt.Sprintf("[HINT: [%d]%s (%s)]", *hint.CardIndex, cuiCardStr(card), whistHintReasonStr(hint.Reason))))
}

// whistHintReasons はWhist固有のヒント理由翻訳
var whistHintReasons = map[string]string{
	"trump_cut": "トランプでカット",
}

// whistHintReasonStr ヒント理由を日本語に変換する
func whistHintReasonStr(reason string) string {
	return lookupHintReason(reason, whistHintReasons)
}

// ActionLogOutput 棋譜をテキスト出力
func (p *WhistCuiPresenter) ActionLogOutput(w interfaces.WhistGame) string {
	return actionLogOutputText(w)
}
