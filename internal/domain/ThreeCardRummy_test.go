//go:build test && (!js || !wasm || casino)

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTcrAtAction は「配り終わってアクション待ち」の卓を、手札を指定して作る。
// Bet を通さないので配りの乱数に依存しない。
func newTcrAtAction(t *testing.T, player, dealer []*Card, ante, lowBonus int) *ThreeCardRummy {
	t.Helper()
	tc := NewDefaultThreeCardRummy()
	tc.SetPhase(ThreeCardRummyPhaseAction)
	tc.SetPlayerHand(player)
	tc.SetDealerHand(dealer)
	tc.SetAnteBet(ante)
	tc.SetLowBonusBet(lowBonus)
	return tc
}

func TestThreeCardRummy_TheLowerTotalWins(t *testing.T) {
	// **このゲームの肝。** 6 点 vs 15 点でプレイヤーの勝ち。素直に大小を取ると
	// 全部逆になるので、両向きと引き分けを固定する。
	tests := []struct {
		name           string
		player, dealer []*Card
		want           GameResult
	}{
		{"lower player total wins", tcrHand(0, 2, 1, 3, 2, 5 /*=10*/), tcrHand(0, 5, 1, 6, 2, 9 /*=20*/), GameResultWin},
		{"higher player total loses", tcrHand(0, 5, 1, 6, 2, 9), tcrHand(0, 2, 1, 3, 2, 5), GameResultLose},
		{"equal totals push", tcrHand(0, 4, 1, 5, 2, 6 /*=15*/), tcrHand(1, 6, 2, 4, 0, 5 /*=15*/), GameResultDraw},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := newTcrAtAction(t, tt.player, tt.dealer, 10, 0)
			require.NoError(t, tc.Play())
			assert.Equal(t, tt.want, tc.GetResult())
		})
	}
}

func TestThreeCardRummy_AMeldBeatsEveryUnmeldedHand(t *testing.T) {
	// K-K-K は素点 30 だが役なので 0 点 -- ディーラーの 3 点 (A-A-A ではない
	// ばらけた低い手) にも勝つ。役を 0 点に落とす扱いが勝敗まで届いているか。
	// ディーラーも同ランク3枚だと同じ 0 点になるので、役にならない低い手を渡す。
	tc := newTcrAtAction(t, tcrHand(0, 13, 1, 13, 2, 13), tcrHand(0, 1, 1, 1, 2, 2 /*=1+1+2=4*/), 10, 0)
	require.NoError(t, tc.Play())
	assert.Equal(t, 0, tc.GetPlayerScore())
	assert.Equal(t, 4, tc.GetDealerScore())
	assert.Equal(t, GameResultWin, tc.GetResult())
}

func TestThreeCardRummy_DealerQualifiesOnTwentyOrLess(t *testing.T) {
	// 境界を両側から押さえる。20 でクオリファイし、21 でしない。
	overLimit := tcrHand(0, 13, 1, 13, 2, 1 /*=10+10+1=21*/)
	require.Equal(t, 21, ThreeCardRummyScore(overLimit), "境界のすぐ外側であること")

	tc := newTcrAtAction(t, tcrHand(0, 2, 1, 3, 2, 5), tcrHand(0, 13, 1, 13, 2, 1), 10, 0)
	require.NoError(t, tc.Play())
	assert.False(t, tc.GetDealerQualified(), "21 点はクオリファイ上限を超える")

	tc = newTcrAtAction(t, tcrHand(0, 2, 1, 3, 2, 5), tcrHand(0, 13, 1, 9, 2, 1 /*=10+9+1=20*/), 10, 0)
	require.NoError(t, tc.Play())
	assert.True(t, tc.GetDealerQualified(), "20 点ちょうどはクオリファイする")
}

func TestThreeCardRummy_AnUnqualifiedDealerPaysTheAnteAndPushesThePlay(t *testing.T) {
	// **負けの手でも払い戻される。** ディーラーが降りているので勝敗は関係ない。
	tc := newTcrAtAction(t, tcrHand(0, 13, 1, 12, 2, 5 /*=25*/), tcrHand(0, 13, 1, 12, 2, 1 /*=21 -> 未クオリファイ*/), 10, 0)
	tc.SetChips(10) // プレイベット分ちょうど。解決後の残高が配当そのものになる。
	require.NoError(t, tc.Play())

	assert.False(t, tc.GetDealerQualified())
	assert.Equal(t, GameResultWin, tc.GetResult(),
		"点数では負けているが、クオリファイ不成立なら取り分は増えるので負けとは呼ばない")
	assert.Equal(t, 20, tc.GetAntePayout(), "アンテは 1:1 で払う")
	assert.Equal(t, 10, tc.GetPlayPayout(), "プレイはプッシュ (返却のみ)")
	assert.Equal(t, 30, tc.GetChips(), "負けた手でもアンテ配当とプレイの返却が乗る")
}

func TestThreeCardRummy_AnteBonusPaysByHowLowTheHandIs(t *testing.T) {
	// ディーラーの手に関係なく自分の点だけで決まる段。境界の両側を取る。
	tests := []struct {
		name       string
		hand       []*Card
		wantScore  int
		wantPayout int
	}{
		{"meld pays the top rate", tcrHand(0, 13, 1, 13, 2, 13), 0, 10 * ThreeCardRummyAnteBonusPerfect},
		{"five or less pays the middle rate", tcrHand(0, 1, 1, 1, 2, 3 /*=5*/), 5, 10 * ThreeCardRummyAnteBonusVeryLow},
		{"six pays the low rate", tcrHand(0, 1, 1, 2, 2, 3 /*=6*/), 6, 10 * ThreeCardRummyAnteBonusLow},
		{"ten pays the low rate", tcrHand(0, 1, 1, 4, 2, 5 /*=10*/), 10, 10 * ThreeCardRummyAnteBonusLow},
		{"eleven pays nothing", tcrHand(0, 2, 1, 4, 2, 5 /*=11*/), 11, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ディーラーは常に未クオリファイの高い手にして、ボーナス以外を固定する。
			tc := newTcrAtAction(t, tt.hand, tcrHand(0, 13, 1, 12, 2, 6), 10, 0)
			require.NoError(t, tc.Play())
			assert.Equal(t, tt.wantScore, tc.GetPlayerScore())
			assert.Equal(t, tt.wantPayout, tc.GetAnteBonusPayout())
		})
	}
}

func TestThreeCardRummy_LowBonusIsIndependentOfTheDealer(t *testing.T) {
	// **降りても評価される。** 賭けたのは勝負ではなく自分の点の低さ。
	tc := newTcrAtAction(t, tcrHand(0, 13, 1, 13, 2, 13), tcrHand(0, 1, 1, 2, 2, 4 /*=7*/), 10, 20)
	require.NoError(t, tc.Fold())
	assert.Equal(t, GameResultLose, tc.GetResult(), "勝負そのものは降りたので負け")
	assert.Equal(t, 0, tc.GetAnteBonusPayout(), "アンテボーナスは降りたら付かない")
	assert.Equal(t, 20+20*ThreeCardRummyLowBonusPerfect, tc.GetLowBonusPayout())
}

func TestThreeCardRummy_LowBonusRatesStepWithTheScore(t *testing.T) {
	tests := []struct {
		name string
		hand []*Card
		want int
	}{
		{"meld", tcrHand(0, 13, 1, 13, 2, 13), 20 + 20*ThreeCardRummyLowBonusPerfect},
		{"five", tcrHand(0, 1, 1, 1, 2, 3), 20 + 20*ThreeCardRummyLowBonusVeryLow},
		{"ten", tcrHand(0, 1, 1, 4, 2, 5), 20 + 20*ThreeCardRummyLowBonusLow},
		{"eleven forfeits the side bet", tcrHand(0, 2, 1, 4, 2, 5), 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := newTcrAtAction(t, tt.hand, tcrHand(0, 13, 1, 12, 2, 6), 10, 20)
			require.NoError(t, tc.Fold())
			assert.Equal(t, tt.want, tc.GetLowBonusPayout())
		})
	}
}

func TestThreeCardRummy_NoLowBonusBetPaysNothingEvenOnAMeld(t *testing.T) {
	tc := newTcrAtAction(t, tcrHand(0, 13, 1, 13, 2, 13), tcrHand(0, 1, 1, 1, 2, 2), 10, 0)
	require.NoError(t, tc.Play())
	assert.Equal(t, 0, tc.GetLowBonusPayout(), "賭けていない側注に配当は出ない")
}

func TestThreeCardRummy_TheDealSetsThePlayerScoreButNotTheDealerScore(t *testing.T) {
	// **配った時点で自分の点は確定する。** ここが 0 のままだと CUI も Web も
	// 「0 点 = 役 = 最強」を全ハンドで表示してしまう。相手の点はまだ伏せる。
	// 役 (0 点) が配られると「設定し忘れ」と区別が付かないので、役でない手が
	// 出るまで配り直す。役は稀なので数回で抜ける。
	tc := NewDefaultThreeCardRummy()
	var score int
	for range 100 {
		tc.Reset()
		require.NoError(t, tc.Bet(10, 0))
		if score = ThreeCardRummyScore(tc.GetPlayerHand()); score > 0 {
			break
		}
	}
	require.Greater(t, score, 0, "役でない手が 100 回配って一度も出なかった")
	assert.Equal(t, score, tc.GetPlayerScore())
	assert.Equal(t, 0, tc.GetDealerScore(), "ディーラーの点は勝負するまで出さない")
}

func TestThreeCardRummy_BetRejectsBadAmountsAndDealsSixCards(t *testing.T) {
	tc := NewDefaultThreeCardRummy()
	assert.Error(t, tc.Bet(ThreeCardRummyMinBet-1, 0), "最低額未満")
	assert.Error(t, tc.Bet(15, 0), "最低額の倍数でない")
	assert.Error(t, tc.Bet(ThreeCardRummyMaxBet+ThreeCardRummyMinBet, 0), "上限超え")
	assert.Error(t, tc.Bet(10, -10), "側注が負")
	assert.Error(t, tc.Bet(10, 15), "側注が最低額の倍数でない")
	assert.Equal(t, ThreeCardRummyPhaseBet, tc.GetPhase(), "弾いたベットで進まない")

	require.NoError(t, tc.Bet(10, 20))
	assert.Len(t, tc.GetPlayerHand(), ThreeCardRummyHandSize)
	assert.Len(t, tc.GetDealerHand(), ThreeCardRummyHandSize)
	assert.Equal(t, ThreeCardRummyPhaseAction, tc.GetPhase())
	assert.Equal(t, ThreeCardRummyDefaultChips-30, tc.GetChips(), "アンテと側注の両方が引かれる")
}

func TestThreeCardRummy_InsufficientChipsRejectsTheBet(t *testing.T) {
	tc := NewDefaultThreeCardRummy()
	tc.SetChips(15)
	assert.Error(t, tc.Bet(10, 10), "合計 20 は 15 チップでは払えない")
	assert.Equal(t, 15, tc.GetChips(), "弾いたベットでチップは減らない")
	assert.Equal(t, ThreeCardRummyPhaseBet, tc.GetPhase())
}

func TestThreeCardRummy_ActionsAreRejectedOutsideTheirPhase(t *testing.T) {
	tc := NewDefaultThreeCardRummy()
	assert.Error(t, tc.Play(), "ベット前にプレイできない")
	assert.Error(t, tc.Fold(), "ベット前に降りられない")
	require.NoError(t, tc.Bet(10, 0))
	assert.Error(t, tc.Bet(10, 0), "アクションフェーズでベットし直せない")
	require.NoError(t, tc.Play())
	assert.Error(t, tc.Play(), "解決後にもう一度プレイできない")
	assert.Error(t, tc.Fold(), "解決後に降りられない")
}

func TestThreeCardRummy_RebetRepeatsTheLastAmountsAcrossReset(t *testing.T) {
	tc := NewDefaultThreeCardRummy()
	err := tc.Rebet()
	require.Error(t, err, "まだ賭けていない")
	assert.Contains(t, err.Error(), "再ベット", "額 0 のベットエラーではなく、専用のメッセージを返す")
	require.NoError(t, tc.Bet(20, 10))
	require.NoError(t, tc.Fold())
	tc.Reset()
	assert.Equal(t, 0, tc.GetAnteBet(), "Reset で今回のベットは消える")
	require.NoError(t, tc.Rebet())
	assert.Equal(t, 20, tc.GetAnteBet())
	assert.Equal(t, 10, tc.GetLowBonusBet())
}

func TestThreeCardRummy_ResetClearsTheRoundButRefillsABrokePlayer(t *testing.T) {
	tc := NewDefaultThreeCardRummy()
	require.NoError(t, tc.Bet(10, 10))
	require.NoError(t, tc.Fold())
	tc.SetChips(ThreeCardRummyMinBet - 1)
	tc.Reset()

	assert.Equal(t, ThreeCardRummyPhaseBet, tc.GetPhase())
	assert.False(t, tc.GetGameEndFlag())
	assert.Empty(t, tc.GetPlayerHand())
	assert.Empty(t, tc.GetDealerHand())
	assert.Equal(t, 0, tc.GetPlayerScore())
	assert.Equal(t, 0, tc.GetDealerScore())
	assert.Equal(t, 0, tc.GetLowBonusPayout())
	assert.False(t, tc.GetDealerQualified())
	assert.Equal(t, ThreeCardRummyDefaultChips, tc.GetChips(), "最低ベット額を割ったら補充する")
}

func TestThreeCardRummy_ResetLeavesAFundedPlayerAlone(t *testing.T) {
	tc := NewDefaultThreeCardRummy()
	tc.SetChips(ThreeCardRummyMinBet)
	tc.Reset()
	assert.Equal(t, ThreeCardRummyMinBet, tc.GetChips(), "最低ベット額ちょうどは補充しない")
}

func TestThreeCardRummy_WinningCreditsEveryPayoutToTheStack(t *testing.T) {
	tc := newTcrAtAction(t, tcrHand(0, 13, 1, 13, 2, 13 /*役 0*/), tcrHand(0, 13, 1, 9, 2, 1 /*=20, クオリファイ*/), 10, 20)
	tc.SetChips(10) // プレイベット分ちょうど。
	require.NoError(t, tc.Play())

	assert.Equal(t, GameResultWin, tc.GetResult())
	want := tc.GetAntePayout() + tc.GetPlayPayout() + tc.GetAnteBonusPayout() + tc.GetLowBonusPayout()
	assert.Equal(t, want, tc.GetTotalPayout())
	assert.Equal(t, want, tc.GetChips(), "プレイベットを引いた残り 0 に合計配当がそのまま乗る")
	assert.Greater(t, want, 0)
}

func TestThreeCardRummy_LosingForfeitsTheAnteAndPlayBets(t *testing.T) {
	tc := newTcrAtAction(t, tcrHand(0, 13, 1, 12, 2, 5 /*=25*/), tcrHand(0, 2, 1, 3, 2, 5 /*=10*/), 10, 0)
	tc.SetChips(10)
	require.NoError(t, tc.Play())
	assert.Equal(t, GameResultLose, tc.GetResult())
	assert.Equal(t, 0, tc.GetAntePayout())
	assert.Equal(t, 0, tc.GetPlayPayout())
	assert.Equal(t, 0, tc.GetChips(), "アンテもプレイも没収され残高 0")
}

func TestThreeCardRummy_APushReturnsBothBets(t *testing.T) {
	tc := newTcrAtAction(t, tcrHand(0, 4, 1, 5, 2, 6), tcrHand(1, 6, 2, 4, 0, 5), 10, 0)
	tc.SetChips(10)
	require.NoError(t, tc.Play())
	assert.Equal(t, GameResultDraw, tc.GetResult())
	assert.Equal(t, 10, tc.GetAntePayout(), "アンテは返るだけ")
	assert.Equal(t, 10, tc.GetPlayPayout(), "プレイも返るだけ")
	assert.Equal(t, 20, tc.GetChips(), "アンテとプレイがそのまま戻る")
}

func TestThreeCardRummy_PlayMatchesTheAnte(t *testing.T) {
	tc := NewDefaultThreeCardRummy()
	require.NoError(t, tc.Bet(30, 0))
	before := tc.GetChips()
	require.NoError(t, tc.Play())
	assert.Equal(t, 30, tc.GetPlayBet(), "プレイベットはアンテと同額")
	assert.Equal(t, before-30+tc.GetTotalPayout(), tc.GetChips())
}

func TestThreeCardRummy_FoldingLeavesNoPlayBet(t *testing.T) {
	tc := NewDefaultThreeCardRummy()
	require.NoError(t, tc.Bet(10, 0))
	before := tc.GetChips()
	require.NoError(t, tc.Fold())
	assert.Equal(t, 0, tc.GetPlayBet(), "降りたらプレイベットは置かない")
	assert.Equal(t, before, tc.GetChips(), "降りてもそれ以上は引かれない")
	assert.True(t, tc.GetGameEndFlag())
	assert.Equal(t, ThreeCardRummyPhaseEnd, tc.GetPhase())
}

func TestThreeCardRummy_JSONRoundTripKeepsEveryField(t *testing.T) {
	// **フィールドが全て非公開**なので、marshaller を書き忘れると `{}` になり、
	// Worker はリクエストのたびに卓を作り直す。往復で全部戻ることを固定する。
	// lastAnteBet / lastLowBonusBet は Bet を通らないと立たないので、実際の
	// 配りの経路で 1 ラウンド回してから往復させる。
	tc := NewDefaultThreeCardRummy()
	require.NoError(t, tc.Bet(20, 30))
	tc.SetPlayerHand(tcrHand(0, 13, 1, 13, 2, 13))
	tc.SetDealerHand(tcrHand(0, 1, 1, 1, 2, 2))
	require.NoError(t, tc.Play())

	data, err := json.Marshal(tc)
	require.NoError(t, err)
	assert.Greater(t, len(data), 100, "非公開フィールドだけの構造体は黙って `{}` になる")

	var got ThreeCardRummy
	require.NoError(t, json.Unmarshal(data, &got))

	assert.Equal(t, tc.GetChips(), got.GetChips())
	assert.Equal(t, tc.GetAnteBet(), got.GetAnteBet())
	assert.Equal(t, tc.GetLowBonusBet(), got.GetLowBonusBet())
	assert.Equal(t, tc.GetPlayBet(), got.GetPlayBet())
	assert.Equal(t, tc.GetPhase(), got.GetPhase())
	assert.Equal(t, tc.GetGameEndFlag(), got.GetGameEndFlag())
	assert.Equal(t, tc.GetResult(), got.GetResult())
	assert.Equal(t, tc.GetAntePayout(), got.GetAntePayout())
	assert.Equal(t, tc.GetPlayPayout(), got.GetPlayPayout())
	assert.Equal(t, tc.GetAnteBonusPayout(), got.GetAnteBonusPayout())
	assert.Equal(t, tc.GetLowBonusPayout(), got.GetLowBonusPayout())
	assert.Equal(t, tc.GetDealerQualified(), got.GetDealerQualified())
	assert.Equal(t, tc.GetPlayerScore(), got.GetPlayerScore())
	assert.Equal(t, tc.GetDealerScore(), got.GetDealerScore())
	assert.Len(t, got.GetPlayerHand(), ThreeCardRummyHandSize)
	assert.Len(t, got.GetDealerHand(), ThreeCardRummyHandSize)
	assert.Equal(t, tc.GetActionLog(), got.GetActionLog())

	// lastAnteBet / lastLowBonusBet は Rebet の唯一の手掛かり -- 往復で消えると
	// KV から戻した卓で再ベットできなくなる。
	got.Reset()
	require.NoError(t, got.Rebet())
	assert.Equal(t, 20, got.GetAnteBet())
	assert.Equal(t, 30, got.GetLowBonusBet())
}

func TestThreeCardRummy_UnmarshalRejectsOversizedArrays(t *testing.T) {
	assert.Error(t, json.Unmarshal([]byte(`{"ph":`+tcrHugeCardArray()+`}`), new(ThreeCardRummy)))
	assert.Error(t, json.Unmarshal([]byte(`not json`), new(ThreeCardRummy)))
}

// tcrHugeCardArray は上限を超える長さのカード配列 JSON を返す。
func tcrHugeCardArray() string {
	s := "["
	for i := range threeCardRummyMaxSliceLen + 1 {
		if i > 0 {
			s += ","
		}
		s += "null"
	}
	return s + "]"
}

func TestThreeCardRummy_FoldingStillRecordsWhetherTheDealerQualified(t *testing.T) {
	// 降りても結果画面はディーラーの手を開いて資格を書く。計算しないと
	// dealerQualified が false のまま残り、4 点の手に「クオリファイせず」と出る。
	tc := newTcrAtAction(t, tcrHand(0, 13, 1, 12, 2, 5), tcrHand(0, 1, 1, 1, 2, 2 /*=4*/), 10, 0)
	require.NoError(t, tc.Fold())
	assert.Equal(t, 4, tc.GetDealerScore())
	assert.True(t, tc.GetDealerQualified(), "4 点はクオリファイ上限の 20 を大きく下回る")

	tc = newTcrAtAction(t, tcrHand(0, 13, 1, 12, 2, 5), tcrHand(0, 13, 1, 12, 2, 1 /*=21*/), 10, 0)
	require.NoError(t, tc.Fold())
	assert.False(t, tc.GetDealerQualified(), "21 点は上限を超える")
}

func TestThreeCardRummy_AnUnqualifiedDealerIsNeverAnnouncedAsALoss(t *testing.T) {
	// **儲かった局面を負けと呼ばない。** ディーラーがクオリファイしなければ
	// アンテが 1:1、プレイは返却 —— 2×ante 賭けて 3×ante 戻るので、点数が
	// どれだけ悪くても取り分は必ず増える。素の大小で result を決めていた頃は
	// 30 点対 21 点で「ディーラーの勝ちです」と赤字で出しながら残高が
	// +10 されていた (全配り 8.7%)。
	tc := newTcrAtAction(t,
		tcrHand(0, 13, 1, 13, 2, 12 /*=30、最悪級の手*/),
		tcrHand(0, 12, 1, 12, 2, 1 /*=21、クオリファイ上限のすぐ上*/), 10, 0)
	tc.SetChips(10) // プレイベット分ちょうど
	require.NoError(t, tc.Play())

	require.False(t, tc.GetDealerQualified())
	require.Greater(t, tc.GetPlayerScore(), tc.GetDealerScore(), "点数では負けている局面であること")
	assert.Equal(t, GameResultWin, tc.GetResult(), "取り分が増えたのだから負けではない")
	// アンテ 10 + プレイ 10 = 20 賭けて、アンテ配当 20 + プレイ返却 10 = 30 戻る。
	assert.Equal(t, 30, tc.GetChips(), "差し引きちょうどアンテぶん増える")
}

func TestThreeCardRummy_AQualifiedDealerStillDecidesOnTheScores(t *testing.T) {
	// クオリファイした卓では従来どおり点数の大小で決める。上のショートカットが
	// 通常の勝敗まで飲み込んでいないことを押さえる。
	tc := newTcrAtAction(t, tcrHand(0, 13, 1, 12, 2, 5 /*=25*/), tcrHand(0, 2, 1, 3, 2, 5 /*=10*/), 10, 0)
	tc.SetChips(10)
	require.NoError(t, tc.Play())
	require.True(t, tc.GetDealerQualified())
	assert.Equal(t, GameResultLose, tc.GetResult())
	assert.Equal(t, 0, tc.GetChips(), "負けたので没収")
}
