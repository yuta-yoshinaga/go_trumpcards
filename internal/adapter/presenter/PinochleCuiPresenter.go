package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// pinochleMeldTypeName メルド種類名
var pinochleMeldTypeName = map[domain.PinochleMeldType]string{
	domain.PinochleMeldDix:                "ディクス",
	domain.PinochleMeldCommonMarriage:     "コモンマリッジ",
	domain.PinochleMeldRoyalMarriage:      "ロイヤルマリッジ",
	domain.PinochleMeldPinochle:           "ピノクル",
	domain.PinochleMeldJacksAround:        "ジャックアラウンド",
	domain.PinochleMeldQueensAround:       "クイーンアラウンド",
	domain.PinochleMeldKingsAround:        "キングアラウンド",
	domain.PinochleMeldAcesAround:         "エースアラウンド",
	domain.PinochleMeldRun:                "ラン",
	domain.PinochleMeldDoublePinochle:     "ダブルピノクル",
	domain.PinochleMeldDoubleJacksAround:  "ダブルジャックアラウンド",
	domain.PinochleMeldDoubleQueensAround: "ダブルクイーンアラウンド",
	domain.PinochleMeldDoubleKingsAround:  "ダブルキングアラウンド",
	domain.PinochleMeldDoubleAcesAround:   "ダブルエースアラウンド",
	domain.PinochleMeldDoubleRun:          "ダブルラン",
}

// pinochlePlayerStr returns the display string for a single Pinochle player.
func pinochlePlayerStr(player *domain.PinochlePlayer, i int) string {
	var b strings.Builder
	name := cuiPlayerName(player, i)
	fmt.Fprintf(&b, "%s: チーム%d ビッド:%d メルド:%d点 トリック:%dT/%d点 %d枚\n",
		name,
		player.GetTeam(),
		player.GetBid(),
		player.GetMeldScore(),
		player.GetTrickCount(),
		player.GetTrickPoints(),
		player.GetCardsSize(),
	)
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player))
		b.WriteString("\n")
	}
	return b.String()
}

// PinochleCuiPresenter ピノクルCUIプレゼンタークラス
type PinochleCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *PinochleCuiPresenter) Output(g interfaces.PinochleGame, lastErr error) string {
	return buildCuiOutput("Pinochle (ピノクル)", func(b *strings.Builder) {
		fmt.Fprintf(b, "ラウンド: %d  トリック: %d\n", g.GetRoundNumber(), g.GetTrickNumber())
		fmt.Fprintf(b, "ディーラー: %s\n", cuiPlayerName(g.GetPlayer(g.GetDealerIdx()), g.GetDealerIdx()))

		trumpSuit := g.GetTrumpSuit()
		if trumpSuit > 0 {
			fmt.Fprintf(b, "切り札: %s (ビッドチーム: チーム%d)\n", cuiSuitName(trumpSuit), g.GetPlayer(g.GetHighestBidder()).GetTeam())
		} else {
			b.WriteString("切り札: 未決定\n")
		}

		fmt.Fprintf(b, "最高ビッド: %d", g.GetHighestBid())
		if g.GetHighestBidder() >= 0 {
			fmt.Fprintf(b, " (%s)", cuiPlayerName(g.GetPlayer(g.GetHighestBidder()), g.GetHighestBidder()))
		}
		b.WriteString("\n")

		// チームスコア
		fmt.Fprintf(b, "チーム0: %d点  チーム1: %d点\n", g.GetTeamScore(0), g.GetTeamScore(1))

		// プレイヤー情報
		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(pinochlePlayerStr(g.GetPlayer(i), i))
		}

		// メルド情報
		phase := g.GetPhase()
		if phase == domain.PinochlePhaseMeld || phase == domain.PinochlePhaseRoundEnd {
			melds := g.GetPlayerMelds()
			for i := range domain.PinochlePlayerCnt {
				if len(melds[i]) > 0 {
					fmt.Fprintf(b, "%sのメルド:\n", cuiPlayerName(g.GetPlayer(i), i))
					for _, m := range melds[i] {
						fmt.Fprintf(b, "  %s: %d点\n", pinochleMeldTypeName[m.Type], m.Points)
					}
				}
			}
		}

		// 現在のトリック
		trick := g.GetCurrentTrick()
		if len(trick) > 0 {
			b.WriteString("テーブル: ")
			for j, tc := range trick {
				if j > 0 {
					b.WriteString(", ")
				}
				fmt.Fprintf(b, "%s=%s", cuiPlayerName(g.GetPlayer(tc.PlayerIdx), tc.PlayerIdx), cuiCardStr(tc.Card))
			}
			b.WriteString("\n")
		}

		p.buildCuiMessage(b, g, lastErr)
	})
}

// HintOutput ヒント情報を出力
func (p *PinochleCuiPresenter) HintOutput(g interfaces.PinochleGame) string {
	hint := g.GetHint()
	if hint == nil {
		return "ヒントなし"
	}
	var b strings.Builder
	b.WriteString("ヒント: ")
	if hint.BidAmount != nil {
		fmt.Fprintf(&b, "ビッド %d", *hint.BidAmount)
	}
	if hint.Pass != nil && *hint.Pass {
		b.WriteString("パス推奨")
	}
	if hint.Suit != nil {
		fmt.Fprintf(&b, "スート %s", cuiSuitName(*hint.Suit))
	}
	if hint.CardIndex != nil {
		fmt.Fprintf(&b, "カード %d", *hint.CardIndex)
	}
	return b.String()
}

// ActionLogOutput 棋譜を出力
func (p *PinochleCuiPresenter) ActionLogOutput(g interfaces.PinochleGame) string {
	return actionLogOutputText(g)
}

// buildCuiMessage CUI用メッセージを構築
func (p *PinochleCuiPresenter) buildCuiMessage(b *strings.Builder, g interfaces.PinochleGame, lastErr error) {
	if lastErr != nil {
		fmt.Fprintf(b, "エラー: %s\n", lastErr.Error())
		return
	}
	if g.GetGameEndFlag() {
		fmt.Fprintf(b, "ゲーム終了！ チーム%dの勝ち！\n", g.GetWinnerTeam())
		return
	}
	switch g.GetPhase() {
	case domain.PinochlePhaseBid:
		fmt.Fprintf(b, "%sのビッド番です。(bid <n> / pass)\n",
			cuiPlayerName(g.GetPlayer(g.GetBidPlayerIdx()), g.GetBidPlayerIdx()))
	case domain.PinochlePhaseTrump:
		fmt.Fprintf(b, "%sがトランプスートを選んでください。(trump <1-4>)\n",
			cuiPlayerName(g.GetPlayer(g.GetCurrentPlayerIdx()), g.GetCurrentPlayerIdx()))
	case domain.PinochlePhaseMeld:
		b.WriteString("メルドフェーズです。(meld で確認して次に進む)\n")
	case domain.PinochlePhasePlay:
		fmt.Fprintf(b, "%sの番です。(play <i>)\n",
			cuiPlayerName(g.GetPlayer(g.GetCurrentPlayerIdx()), g.GetCurrentPlayerIdx()))
	case domain.PinochlePhaseTrickEnd:
		b.WriteString("トリック終了。(next で次のトリックへ)\n")
	case domain.PinochlePhaseRoundEnd:
		b.WriteString("ラウンド終了。(nextround で次のラウンドへ)\n")
	}
}
