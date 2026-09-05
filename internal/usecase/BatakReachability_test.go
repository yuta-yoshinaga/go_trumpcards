//go:build test

package usecase_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// batakOut は Web プレゼンタが返す JSON のうち、この試験で見る分。
type batakOut struct {
	Phase            int    `json:"phase"`
	RoundNumber      int    `json:"roundNumber"`
	TrickNumber      int    `json:"trickNumber"`
	CurrentPlayerIdx int    `json:"currentPlayerIdx"`
	BidPlayerIdx     int    `json:"bidPlayerIdx"`
	DeclarerIdx      int    `json:"declarerIdx"`
	HighBid          int    `json:"highBid"`
	MinLegalBid      int    `json:"minLegalBid"`
	ValidPlayIndices []int  `json:"validPlayIndices"`
	GameEndFlag      bool   `json:"gameEndFlag"`
	Message          string `json:"message"`
}

func parseBatakOut(t *testing.T, s string) batakOut {
	t.Helper()
	var o batakOut
	require.NoError(t, json.Unmarshal([]byte(s), &o))
	return o
}

// TestBatakInteractor_PlayableFromReset は Reset() から人間のビッド・プレイ到達を検証する。
func TestBatakInteractor_PlayableFromReset(t *testing.T) {
	g := domain.NewDefaultBatak()
	gi := usecase.NewBatakInteractor(g, &presenter.BatakWebPresenter{})

	out := parseBatakOut(t, gi.Reset())

	// Reset 直後はビッドフェーズで、seat 0 (人間) のビッド待ちで止まること。
	require.Equal(t, int(domain.BatakPhaseBid), out.Phase, "Reset 後はビッドフェーズにいること")
	assert.Equal(t, 0, out.BidPlayerIdx, "初期ラウンドは人間 (seat 0) からビッド開始")
	assert.Equal(t, domain.BatakMinBid, out.MinLegalBid, "初手なら MinLegalBid は 5")

	// 不正なビッド (1〜4) はドメインエラーがそのまま返ること。
	invalidOut := parseBatakOut(t, gi.Bid(3))
	assert.NotEmpty(t, invalidOut.Message, "3 は MinLegalBid(5) 未満なのでエラーメッセージが返ること")
	assert.Equal(t, int(domain.BatakPhaseBid), invalidOut.Phase, "エラー時はビッドフェーズのまま")

	// 人間が適正なビッド (5) を宣言する。
	out = parseBatakOut(t, gi.Bid(domain.BatakMinBid))

	// 全員のビッドが完了し、プレイフェーズに入っていること。
	require.Equal(t, int(domain.BatakPhasePlay), out.Phase, "ビッド後はプレイフェーズへ進むこと")
	assert.GreaterOrEqual(t, out.DeclarerIdx, 0, "親が確定していること")
	assert.GreaterOrEqual(t, out.HighBid, domain.BatakMinBid, "最高ビッドが確定していること")

	// プレイフェーズで人間がカードを 1 枚出すまで到達できること。
	// 親が CPU の場合はリードを CPU が済ませて人間の手番で止まり、
	// 親が人間の場合は人間のリード待ちで止まる。
	require.Equal(t, 0, out.CurrentPlayerIdx, "止まるのは人間の手番であること")
	require.NotEmpty(t, out.ValidPlayIndices, "出せる札があること")

	// 人間が 1 枚出す。
	playedOut := parseBatakOut(t, gi.Play(out.ValidPlayIndices[0]))
	assert.Empty(t, playedOut.Message, "カードプレイでエラーが出ないこと")
}

// TestBatakInteractor_BidPass はパス (Bid(0)) を通してプレイフェーズへ到達できることを検証する。
func TestBatakInteractor_BidPass(t *testing.T) {
	g := domain.NewDefaultBatak()
	gi := usecase.NewBatakInteractor(g, &presenter.BatakWebPresenter{})

	out := parseBatakOut(t, gi.Reset())
	require.Equal(t, int(domain.BatakPhaseBid), out.Phase)

	// 人間がパス (0) を宣言する。
	out = parseBatakOut(t, gi.Bid(domain.BatakPassBid))
	assert.Empty(t, out.Message, "パスでエラーが出ないこと")

	// CPU のビッドを経てプレイフェーズに到達すること。
	require.Equal(t, int(domain.BatakPhasePlay), out.Phase, "パス後も全員ビッド完了でプレイフェーズへ進む")
	assert.GreaterOrEqual(t, out.DeclarerIdx, 0, "親が確定していること")

	// 人間の手番で札が出せること。
	require.Equal(t, 0, out.CurrentPlayerIdx)
	require.NotEmpty(t, out.ValidPlayIndices)
	out = parseBatakOut(t, gi.Play(out.ValidPlayIndices[0]))
	assert.Empty(t, out.Message)
}

// TestBatakInteractor_RoundCompletes は 1 ラウンド (13 トリック) を最後まで回す。
// 働いた量 (出した枚数 13 枚、完了したトリック数 13) を assert する。
func TestBatakInteractor_RoundCompletes(t *testing.T) {
	g := domain.NewDefaultBatak()
	gi := usecase.NewBatakInteractor(g, &presenter.BatakWebPresenter{})

	out := parseBatakOut(t, gi.Reset())

	plays := 0
	tricksCompleted := 0

	// 1 ラウンド (13 トリック) を回すループ。上限 200 回。
	for i := 0; i < 200 && out.Phase != int(domain.BatakPhaseRoundEnd) && !out.GameEndFlag; i++ {
		switch domain.BatakPhase(out.Phase) {
		case domain.BatakPhaseBid:
			out = parseBatakOut(t, gi.Bid(domain.BatakMinBid))
		case domain.BatakPhasePlay:
			require.NotEmpty(t, out.ValidPlayIndices, "手番なら出せるカードがあること")
			out = parseBatakOut(t, gi.Play(out.ValidPlayIndices[0]))
			plays++
		case domain.BatakPhaseTrickEnd:
			out = parseBatakOut(t, gi.NextTrick())
			tricksCompleted++
		default:
			t.Fatalf("予期せぬフェーズ %d で止まった", out.Phase)
		}
	}

	// 働いた量 (アサーション)
	assert.Equal(t, 13, plays, "人間が 1 ラウンドで 13 枚すべて出せていること")
	totalTricks := 0
	for i := 0; i < g.GetPlayerCnt(); i++ {
		totalTricks += g.GetPlayer(i).GetTrickCount()
	}
	assert.Equal(t, 13, totalTricks, "全プレイヤーの獲得トリック合計が 13 であること")
	assert.Equal(t, int(domain.BatakPhaseRoundEnd), out.Phase, "ラウンド終了フェーズに到達していること")
}
