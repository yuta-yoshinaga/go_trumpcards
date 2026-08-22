//go:build test && (!js || !wasm || casino)

package usecase

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// stubThreeCardRummyPresenter is a minimal presenter for the snapshot tests.
// It records the error it was handed so a restored interactor can be driven
// for real and the outcome inspected — the mock presenter's canned strings
// would hide whether the domain accepted the call.
type stubThreeCardRummyPresenter struct {
	lastErr error
}

func (s *stubThreeCardRummyPresenter) Output(_ interfaces.ThreeCardRummyGame, lastErr error) string {
	s.lastErr = lastErr
	return `{"ok":true}`
}

func (s *stubThreeCardRummyPresenter) ActionLogOutput(_ interfaces.ThreeCardRummyGame) string {
	return `{"log":[]}`
}

func (s *stubThreeCardRummyPresenter) HintOutput(_ interfaces.ThreeCardRummyGame) string {
	return `{"hint":true}`
}

// tcrCard builds one dealt card.
func tcrCard(design, value int) *domain.Card { return domain.NewCard(design, value, true) }

// tcrPlayedGame returns a finished round with **every** persisted field at a
// distinctive non-zero value, so a dropped field shows up as a mismatch rather
// than as zero-equals-zero.
//
// 配りには依存しない: Bet でデッキから配られた手は捨てて、点数の決まった手を
// 置き直してから Play する。
//
//	player 2♠ 3♥ 5♦ = 10 点 (混合スート・非連番なので役ではない)
//	dealer 10♣ 8♠ 2♥ = 20 点 → ちょうどクオリファイ
//
// 10 < 20 なので **低いほうが勝つ**このゲームではプレイヤーの勝ち。10 点は
// アンテボーナス (1 倍) とローボーナス (4 倍) の両方に掛かるので、4 つの配当
// フィールドが全部埋まる。
func tcrPlayedGame(t *testing.T) (*domain.ThreeCardRummy, *ThreeCardRummyInteractor, *stubThreeCardRummyPresenter) {
	t.Helper()

	tc := domain.NewDefaultThreeCardRummy()
	sp := &stubThreeCardRummyPresenter{}
	ti := NewThreeCardRummyInteractor(tc, sp)

	ti.Bet(100, 50)
	require.NoError(t, sp.lastErr, "bet should be accepted from the bet phase")

	tc.SetPlayerHand([]*domain.Card{
		tcrCard(domain.CardDesignSpade, 2),
		tcrCard(domain.CardDesignHeart, 3),
		tcrCard(domain.CardDesignDiamond, 5),
	})
	tc.SetDealerHand([]*domain.Card{
		tcrCard(domain.CardDesignClover, 10),
		tcrCard(domain.CardDesignSpade, 8),
		tcrCard(domain.CardDesignHeart, 2),
	})

	ti.Play()
	require.NoError(t, sp.lastErr, "play should be accepted from the action phase")

	// 前提が崩れたらここで気づく。以下の比較は「両方 0」で通ってしまうので、
	// 元の状態が非退化であることを先に固定しておく。
	require.Equal(t, domain.ThreeCardRummyPhaseEnd, tc.GetPhase())
	require.Equal(t, 10, tc.GetPlayerScore())
	require.Equal(t, 20, tc.GetDealerScore())
	require.True(t, tc.GetDealerQualified())
	require.Equal(t, domain.GameResultWin, tc.GetResult())
	require.NotZero(t, tc.GetAntePayout())
	require.NotZero(t, tc.GetPlayPayout())
	require.NotZero(t, tc.GetAnteBonusPayout())
	require.NotZero(t, tc.GetLowBonusPayout())
	require.NotEqual(t, domain.ThreeCardRummyDefaultChips, tc.GetChips())
	require.NotEmpty(t, tc.GetActionLog())

	return tc, ti, sp
}

// assertSameHand compares two hands card for card, including the dealt flag.
func assertSameHand(t *testing.T, want, got []*domain.Card, what string) {
	t.Helper()
	require.Len(t, got, len(want), "%s: card count", what)
	for i := range want {
		require.NotNil(t, got[i], "%s[%d]", what, i)
		assert.Equal(t, want[i].GetDesign(), got[i].GetDesign(), "%s[%d] suit", what, i)
		assert.Equal(t, want[i].GetValue(), got[i].GetValue(), "%s[%d] value", what, i)
		assert.Equal(t, want[i].GetDraw(), got[i].GetDraw(), "%s[%d] dealt flag", what, i)
	}
}

// **これが KV 永続化の本道。** Cloudflare Worker はリクエストごとに状態を
// 持たないので、Snapshot が 1 フィールド落とすと卓が黙って組み直される
// (エラーは出ない)。全フィールドを往復で突き合わせる。
func TestThreeCardRummyInteractor_SnapshotRoundTripsEveryField(t *testing.T) {
	tc, ti, _ := tcrPlayedGame(t)

	data, err := ti.Snapshot()
	require.NoError(t, err)

	// **`{}` は「成功したがすべて落とした」印。** 非公開フィールドだけの構造体は
	// MarshalJSON が無いとエラーも出さずに 2 バイトになる。
	require.NotEqual(t, "{}", strings.TrimSpace(string(data)),
		"snapshot serialised to an empty object: MarshalJSON is missing or ignored")

	restored, err := RestoreThreeCardRummyInteractor(data, &stubThreeCardRummyPresenter{})
	require.NoError(t, err)
	require.NotNil(t, restored)
	got := restored.Game

	assert.Equal(t, tc.GetChips(), got.GetChips(), "chips")
	assert.Equal(t, tc.GetAnteBet(), got.GetAnteBet(), "ante bet")
	assert.Equal(t, tc.GetLowBonusBet(), got.GetLowBonusBet(), "low bonus bet")
	assert.Equal(t, tc.GetPlayBet(), got.GetPlayBet(), "play bet")
	assert.Equal(t, tc.GetPhase(), got.GetPhase(), "phase")
	assert.Equal(t, tc.GetGameEndFlag(), got.GetGameEndFlag(), "game end flag")
	assert.Equal(t, tc.GetResult(), got.GetResult(), "result")
	assert.Equal(t, tc.GetAntePayout(), got.GetAntePayout(), "ante payout")
	assert.Equal(t, tc.GetPlayPayout(), got.GetPlayPayout(), "play payout")
	assert.Equal(t, tc.GetAnteBonusPayout(), got.GetAnteBonusPayout(), "ante bonus payout")
	assert.Equal(t, tc.GetLowBonusPayout(), got.GetLowBonusPayout(), "low bonus payout")
	assert.Equal(t, tc.GetTotalPayout(), got.GetTotalPayout(), "total payout")
	assert.Equal(t, tc.GetDealerQualified(), got.GetDealerQualified(), "dealer qualified")
	assert.Equal(t, tc.GetPlayerScore(), got.GetPlayerScore(), "player score")
	assert.Equal(t, tc.GetDealerScore(), got.GetDealerScore(), "dealer score")

	assertSameHand(t, tc.GetPlayerHand(), got.GetPlayerHand(), "player hand")
	assertSameHand(t, tc.GetDealerHand(), got.GetDealerHand(), "dealer hand")

	wantLog := tc.GetActionLog()
	gotLog := got.GetActionLog()
	require.Len(t, gotLog, len(wantLog), "action log length")
	for i := range wantLog {
		require.NotNil(t, gotLog[i], "action log entry %d", i)
		assert.Equal(t, wantLog[i].TurnNumber, gotLog[i].TurnNumber, "log[%d] turn", i)
		assert.Equal(t, wantLog[i].PlayerIdx, gotLog[i].PlayerIdx, "log[%d] player", i)
		assert.Equal(t, wantLog[i].ActionType, gotLog[i].ActionType, "log[%d] type", i)
		assert.Equal(t, wantLog[i].Detail, gotLog[i].Detail, "log[%d] detail", i)
	}
}

// **Rebet の手掛かりは getter がない。** lastAnteBet / lastLowBonusBet が
// 落ちても上のフィールド比較は全部通ってしまうので、復元した卓で実際に
// 賭け直して確かめる。
func TestThreeCardRummyInteractor_SnapshotKeepsTheRebetStake(t *testing.T) {
	_, ti, _ := tcrPlayedGame(t)

	data, err := ti.Snapshot()
	require.NoError(t, err)

	sp := &stubThreeCardRummyPresenter{}
	restored, err := RestoreThreeCardRummyInteractor(data, sp)
	require.NoError(t, err)

	restored.Reset()
	require.NoError(t, sp.lastErr)

	restored.Rebet()
	require.NoError(t, sp.lastErr, "rebet lost the previous stake across the snapshot")
	assert.Equal(t, 100, restored.Game.GetAnteBet(), "rebet ante")
	assert.Equal(t, 50, restored.Game.GetLowBonusBet(), "rebet low bonus")
	assert.Equal(t, domain.ThreeCardRummyPhaseAction, restored.Game.GetPhase())
}

// 復元した卓はそのまま次のラウンドを配れる。デッキ (trumpCards) が
// nil のまま戻ると、ここで初めて落ちる。
func TestThreeCardRummyInteractor_RestoredTableCanDealAgain(t *testing.T) {
	_, ti, _ := tcrPlayedGame(t)

	data, err := ti.Snapshot()
	require.NoError(t, err)

	sp := &stubThreeCardRummyPresenter{}
	restored, err := RestoreThreeCardRummyInteractor(data, sp)
	require.NoError(t, err)

	restored.Reset()
	require.NoError(t, sp.lastErr)
	assert.Equal(t, domain.ThreeCardRummyPhaseBet, restored.Game.GetPhase())
	assert.Empty(t, restored.Game.GetPlayerHand(), "reset should clear the previous hand")

	restored.Bet(10, 0)
	require.NoError(t, sp.lastErr)
	assert.Len(t, restored.Game.GetPlayerHand(), domain.ThreeCardRummyHandSize)
	assert.Len(t, restored.Game.GetDealerHand(), domain.ThreeCardRummyHandSize)
}

// ベットフェーズの卓も往復できる。ここで落ちる実装は「一度も賭けていない
// 卓」を復元できないので、最初のリクエストから壊れる。
func TestThreeCardRummyInteractor_SnapshotRoundTripsAFreshTable(t *testing.T) {
	tc := domain.NewDefaultThreeCardRummy()
	ti := NewThreeCardRummyInteractor(tc, &stubThreeCardRummyPresenter{})

	data, err := ti.Snapshot()
	require.NoError(t, err)
	require.NotEqual(t, "{}", strings.TrimSpace(string(data)))

	sp := &stubThreeCardRummyPresenter{}
	restored, err := RestoreThreeCardRummyInteractor(data, sp)
	require.NoError(t, err)

	assert.Equal(t, domain.ThreeCardRummyDefaultChips, restored.Game.GetChips())
	assert.Equal(t, domain.ThreeCardRummyPhaseBet, restored.Game.GetPhase())

	restored.Bet(100, 0)
	require.NoError(t, sp.lastErr)
	assert.Equal(t, domain.ThreeCardRummyDefaultChips-100, restored.Game.GetChips())
}

// **壊れたバイト列はエラーで返る。** ここで nil エラーを返すと、KV に入った
// ゴミがそのまま「まっさらな卓」として通ってしまう。
func TestRestoreThreeCardRummyInteractor_RejectsMalformedBytes(t *testing.T) {
	cases := map[string]string{
		"not json":            "{not json",
		"truncated":           `{"ab":100`,
		"wrong type for hand": `{"ph":5}`,
		"wrong root type":     `[1,2,3]`,
		"empty":               ``,
	}
	for name, payload := range cases {
		t.Run(name, func(t *testing.T) {
			restored, err := RestoreThreeCardRummyInteractor([]byte(payload), &stubThreeCardRummyPresenter{})
			require.Error(t, err)
			assert.Nil(t, restored)
		})
	}
}

// 巨大な配列は上限で弾く。復元は信頼できない入力 (KV の中身) を読むので、
// 長さを見ずに確保すると Worker の割り当てを食い潰せる。
func TestRestoreThreeCardRummyInteractor_RejectsOversizedArrays(t *testing.T) {
	oversized := make([]*domain.Card, threeCardRummyTestOverCap)
	payload, err := json.Marshal(map[string]any{"ph": oversized})
	require.NoError(t, err)

	restored, err := RestoreThreeCardRummyInteractor(payload, &stubThreeCardRummyPresenter{})
	require.Error(t, err)
	assert.Nil(t, restored)
	assert.Contains(t, err.Error(), "exceeds maximum")
}

// threeCardRummyTestOverCap is one past the deserialisation cap in the domain.
const threeCardRummyTestOverCap = 1001
