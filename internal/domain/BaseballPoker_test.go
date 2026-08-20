//go:build test

package domain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newBaseballForTest は配り終えた卓を返す。
func newBaseballForTest(t *testing.T) *BaseballPoker {
	t.Helper()
	g := NewDefaultBaseballPoker()
	g.Reset()
	return g
}

// baseballTotalChips は卓とポットのチップ総量を返す。
func baseballTotalChips(g *BaseballPoker) int {
	total := g.GetPot()
	for _, p := range g.GetPlayers() {
		total += p.GetChips()
	}
	return total
}

// baseballDriveToShowdown は人間の席を機械的に打たせてハンドを閉じる。
//
// **人間の席は残したまま駆動する。** 全席を CPU にすると `HumanSeat()` が
// 席 0 に落ちて、その席を誰も動かさないまま盤面が止まる。
func baseballDriveToShowdown(t *testing.T, g *BaseballPoker) {
	t.Helper()
	for steps := 0; g.GetPhase() != BaseballPhaseShowdown && !g.GetGameEndFlag(); steps++ {
		require.Less(t, steps, 400, "ハンドが終わらない (フェーズ %d)", g.GetPhase())
		switch {
		case g.IsHumanBuying():
			require.NoError(t, g.AnswerBuyIn(BaseballBuyPay))
		case g.IsHumanTurn():
			if err := g.PlayerAction(BaseballActionCheck, 0); err != nil {
				require.NoError(t, g.PlayerAction(BaseballActionCall, 0))
			}
		default:
			g.CpuPlay()
		}
	}
}

// --- 配札 ---

// **3rd ストリートは 2 伏せ + 1 表。**
// **配った直後に既にボーナスが出ていることがある。** 3rd ストリートの表札が
// 4 なら、その場で伏せ札が 1 枚増える ── 「配り終えたら必ず 3 枚」と書くと、
// 表の 4 が出た配りでだけ落ちる (実測でそうなった)。
func TestBaseballPoker_DealsTwoDownAndOneUp(t *testing.T) {
	sawBonus := false
	for range 40 {
		g := newBaseballForTest(t)
		for i, p := range g.GetPlayers() {
			require.Len(t, p.GetFaceUp(), len(p.GetCards()), "席 %d の向きの列", i)
			// 最初の 3 枚は必ず 伏せ・伏せ・表。
			require.GreaterOrEqual(t, len(p.GetCards()), BaseballDownCards+1, "席 %d の枚数", i)
			assert.Equal(t, []bool{false, false, true}, p.GetFaceUp()[:BaseballDownCards+1],
				"席 %d の最初の 3 枚の向き", i)
			// 余分な札はボーナスで、必ず伏せて配られている。
			assert.Equal(t, BaseballDownCards+1+p.GetBonusCards(), len(p.GetCards()),
				"席 %d の枚数がボーナスの枚数と合わない", i)
			for k := BaseballDownCards + 1; k < len(p.GetCards()); k++ {
				sawBonus = true
				assert.False(t, p.GetFaceUp()[k], "席 %d のボーナスが表向き", i)
			}
		}
		assert.Equal(t, 1, g.GetStreet())
		assert.Positive(t, g.GetPot(), "アンティがポットに入っていない")
	}
	assert.True(t, sawBonus, "40 回の配りで 3rd ストリートの 4 が一度も出なかった")
}

// **手札の枚数と向きの列は常に同じ長さ。** ずれると伏せ札が公開される。
func TestBaseballPoker_FaceUpFlagsTrackTheHand(t *testing.T) {
	for range 20 {
		g := newBaseballForTest(t)
		baseballDriveToShowdown(t, g)
		for i, p := range g.GetPlayers() {
			require.Len(t, p.GetFaceUp(), len(p.GetCards()),
				"席 %d で向きの列が手札とずれた", i)
		}
	}
}

// **最後の 1 枚は伏せて配る。** ここで表にすると 7th でイベントが起きる。
func TestBaseballPoker_LastCardIsDealtFaceDown(t *testing.T) {
	seen := 0
	for range 20 {
		g := newBaseballForTest(t)
		baseballDriveToShowdown(t, g)
		for _, p := range g.GetPlayers() {
			if p.GetFolded() || len(p.GetCards()) < BaseballBaseCards {
				continue
			}
			seen++
			faceUp := p.GetFaceUp()
			assert.False(t, faceUp[len(faceUp)-1], "最後の札が表向きで配られている")
			// 表向きは 4 枚まで (3rd〜6th)。ボーナスは伏せ札。
			up := 0
			for _, f := range faceUp {
				if f {
					up++
				}
			}
			assert.LessOrEqual(t, up, BaseballUpCards, "表札が多すぎる")
		}
	}
	require.Positive(t, seen, "最後まで残った席が 1 つも無かった")
}

// --- ワイルドとイベント ---

// **3 と 9 はワイルド。** それ以外はワイルドでない。
func TestBaseballIsWild(t *testing.T) {
	for v := 1; v <= 13; v++ {
		want := v == BaseballWildThree || v == BaseballWildNine
		assert.Equal(t, want, BaseballIsWild(NewCard(CardDesignSpade, v, true)), "値 %d", v)
	}
	assert.False(t, BaseballIsWild(nil), "nil をワイルドと判定している")
}

// **ワイルドは役を押し上げる。** ここが効いていないとただのスタッドになる。
func TestBaseballPokerPlayer_WildsRaiseTheHand(t *testing.T) {
	// A A K Q J — ワンペア。
	plain := NewBaseballPokerPlayer("YOU", 1000, true)
	for _, c := range []*Card{
		NewCard(CardDesignSpade, 1, true), NewCard(CardDesignHeart, 1, true),
		NewCard(CardDesignClover, 13, true), NewCard(CardDesignDiamond, 12, true),
		NewCard(CardDesignSpade, 11, true),
	} {
		plain.AddDealtCard(c, true)
	}
	plainRank := plain.EvaluateBest()
	assert.False(t, plain.GetUsedWild(), "ワイルドを使っていないのに使ったことになっている")

	// 同じ手の J を 3 (ワイルド) に差し替えるとスリーカード以上になる。
	wild := NewBaseballPokerPlayer("YOU", 1000, true)
	for _, c := range []*Card{
		NewCard(CardDesignSpade, 1, true), NewCard(CardDesignHeart, 1, true),
		NewCard(CardDesignClover, 13, true), NewCard(CardDesignDiamond, 12, true),
		NewCard(CardDesignSpade, BaseballWildThree, true),
	} {
		wild.AddDealtCard(c, true)
	}
	wildRank := wild.EvaluateBest()
	assert.Greater(t, wildRank, plainRank, "ワイルドが役を押し上げていない")
	assert.True(t, wild.GetUsedWild(), "ワイルドを使ったのに記録されていない")
}

// **表の 4 で伏せ札が 1 枚増える。** 表で配るとイベントが連鎖する。
func TestBaseballPokerPlayer_BonusCardIsDealtFaceDown(t *testing.T) {
	p := NewBaseballPokerPlayer("YOU", 1000, true)
	p.AddDealtCard(NewCard(CardDesignSpade, BaseballBonusFour, true), true)
	p.AddBonusCard(NewCard(CardDesignHeart, 7, true))

	require.Len(t, p.GetCards(), 2)
	assert.Equal(t, []bool{true, false}, p.GetFaceUp(), "ボーナスが表向きで配られている")
	assert.Equal(t, 1, p.GetBonusCards())
}

// **表の 4 は必ずボーナスを生む。** 実戦の盤面で数える。
func TestBaseballPoker_FaceUpFourGrantsABonus(t *testing.T) {
	saw := 0
	for range 30 {
		g := newBaseballForTest(t)
		baseballDriveToShowdown(t, g)
		for i, p := range g.GetPlayers() {
			fours := 0
			for k, c := range p.GetCards() {
				if k < len(p.GetFaceUp()) && p.GetFaceUp()[k] && c.GetValue() == BaseballBonusFour {
					fours++
				}
			}
			if fours == 0 {
				continue
			}
			saw++
			// 降りた席はそのストリート以降を受け取らないので、下限で見る。
			assert.Positive(t, p.GetBonusCards(),
				"席 %d は表の 4 を %d 枚もらったのにボーナスが 0 枚", i, fours)
		}
	}
	require.Positive(t, saw, "30 局で表の 4 が一度も出なかった")
}

// **表の 3 は買い増しを迫る。** ワイルドで嬉しい札が、表では請求書になる。
func TestBaseballPoker_FaceUpThreeAsksToBuyThePot(t *testing.T) {
	saw := 0
	for range 30 {
		g := newBaseballForTest(t)
		for steps := 0; g.GetPhase() != BaseballPhaseShowdown && !g.GetGameEndFlag(); steps++ {
			require.Less(t, steps, 400)
			if g.GetPhase() == BaseballPhaseBuyIn {
				saw++
				buyer := g.GetBuyerSeat()
				require.GreaterOrEqual(t, buyer, 0, "買い増しフェーズなのに買い手がいない")
				// 迫られた席には表の 3 がある。
				p := g.GetPlayers()[buyer]
				found := false
				for k, c := range p.GetCards() {
					if k < len(p.GetFaceUp()) && p.GetFaceUp()[k] && c.GetValue() == BaseballWildThree {
						found = true
					}
				}
				assert.True(t, found, "表の 3 が無い席が買い増しを迫られている")
				assert.LessOrEqual(t, g.GetBuyCost(), p.GetChips(), "払えない額を請求している")
			}
			switch {
			case g.IsHumanBuying():
				require.NoError(t, g.AnswerBuyIn(BaseballBuyPay))
			case g.IsHumanTurn():
				if err := g.PlayerAction(BaseballActionCheck, 0); err != nil {
					require.NoError(t, g.PlayerAction(BaseballActionCall, 0))
				}
			default:
				g.CpuPlay()
			}
		}
	}
	require.Positive(t, saw, "30 局で表の 3 が一度も出なかった")
}

// baseballAdvanceToBuyIn は買い増しの場面まで進める。
//
// **表の 3 が出るかどうかは配り次第。** 1 卓を 200 手回すだけだと、たまたま
// 出ない配りを引いた回に「到達しなかった」で落ちる (CI で実際に 1 度落ちた)。
// 卓ごと引き直す外側のループを持つことで、確率の低い配りに当たっても
// テストの結論が変わらないようにする。
func baseballAdvanceToBuyIn(t *testing.T) *BaseballPoker {
	t.Helper()
	for attempt := range 50 {
		g := newBaseballForTest(t)
		for range 200 {
			if g.GetPhase() == BaseballPhaseBuyIn {
				return g
			}
			if g.GetPhase() == BaseballPhaseShowdown || g.GetGameEndFlag() {
				if err := g.NextHand(); err != nil {
					break // この卓はもう続けられない。引き直す。
				}
				continue
			}
			if g.IsHumanTurn() {
				if err := g.PlayerAction(BaseballActionCheck, 0); err != nil {
					require.NoError(t, g.PlayerAction(BaseballActionCall, 0))
				}
				continue
			}
			g.CpuPlay()
		}
		_ = attempt
	}
	t.Fatal("買い増しの場面まで到達しなかった (50 卓 × 200 手)")
	return nil
}

// **買い増しの返事はポットとチップに正しく効く。**
func TestBaseballPoker_BuyInMovesChips(t *testing.T) {
	g := baseballAdvanceToBuyIn(t)

	buyer := g.GetBuyerSeat()
	p := g.GetPlayers()[buyer]
	beforeChips, beforePot, cost := p.GetChips(), g.GetPot(), g.GetBuyCost()

	require.NoError(t, g.AnswerBuyIn(BaseballBuyPay))
	assert.Equal(t, beforeChips-cost, p.GetChips(), "買い増しでチップが減っていない")
	assert.GreaterOrEqual(t, g.GetPot(), beforePot+cost, "買い増しがポットに入っていない")
	assert.False(t, p.GetFolded(), "払ったのに降ろされている")
}

// **降りる返事はその席を降ろす。** 両方向を踏む。
func TestBaseballPoker_BuyInFoldDropsTheSeat(t *testing.T) {
	g := baseballAdvanceToBuyIn(t)

	buyer := g.GetBuyerSeat()
	p := g.GetPlayers()[buyer]
	beforeChips := p.GetChips()

	require.NoError(t, g.AnswerBuyIn(BaseballBuyFold))
	assert.True(t, p.GetFolded(), "降りたのに残っている")
	assert.Equal(t, beforeChips, p.GetChips(), "降りたのにチップが減っている")
}

// **買い増しの場面ではベットの手を受け付けない。** 逆もまた然り。
// **フェーズごとに受け付ける手が違う。** 配りによって開始フェーズが変わる
// (3rd の表札が 3 なら買い増しから始まる) ので、両方の枝を明示的に踏む。
func TestBaseballPoker_RejectsTheWrongCommandForThePhase(t *testing.T) {
	sawBetting, sawBuying := false, false
	for range 40 {
		g := newBaseballForTest(t)
		switch g.GetPhase() {
		case BaseballPhaseBetting:
			sawBetting = true
			// ベット中に買い増しの返事は受け付けない。
			assert.ErrorIs(t, g.AnswerBuyIn(BaseballBuyPay), errBaseballNotBuying)
			assert.ErrorIs(t, g.AnswerBuyIn(99), errBaseballNotBuying)
		case BaseballPhaseBuyIn:
			sawBuying = true
			// 買い増し中はベットの手を受け付けず、返事は pay / fold だけ。
			assert.ErrorIs(t, g.PlayerAction(BaseballActionCheck, 0), errBaseballNotBetting)
			assert.ErrorIs(t, g.AnswerBuyIn(99), errBaseballBadBuyAnswer)
		}
	}
	assert.True(t, sawBetting, "40 回の配りでベット開始の局面が出なかった")
	assert.True(t, sawBuying, "40 回の配りで買い増し開始の局面が出なかった")
}

// --- ベット ---

func TestBaseballPoker_ActionRules(t *testing.T) {
	g := newBaseballForTest(t)
	if !g.IsHumanTurn() {
		t.Skip("配り直後に人間の手番でない盤面")
	}
	p := g.GetPlayers()[g.HumanSeat()]

	assert.ErrorIs(t, g.PlayerAction(999, 0), errBaseballUnknownAction)
	assert.ErrorIs(t, g.PlayerAction(BaseballActionBet, 0), errBaseballBetRange)
	assert.ErrorIs(t, g.PlayerAction(BaseballActionBet, p.GetChips()+1), errBaseballBetRange)
}

// **レイズには上限がある。** 上限が無いと 2 席が撃ち尽くすまで終わらない。
// **上限はラウンドごと。** ストリートが変わればまた 3 回打てるので、
// 通算で数えると上限を超えたように見える ── 数えるのは 1 ラウンド内だけ。
func TestBaseballPoker_RaiseIsCapped(t *testing.T) {
	g := newBaseballForTest(t)
	street, raises, sawCap := g.GetStreet(), 0, false
	for range 400 {
		if g.GetPhase() != BaseballPhaseBetting {
			if g.GetPhase() == BaseballPhaseShowdown || g.GetGameEndFlag() {
				break
			}
			if g.IsHumanBuying() {
				require.NoError(t, g.AnswerBuyIn(BaseballBuyPay))
			} else {
				g.CpuPlay()
			}
			continue
		}
		if g.GetStreet() != street {
			street, raises = g.GetStreet(), 0
		}
		require.LessOrEqual(t, g.GetRaiseCount(), baseballMaxRaisesPerRound,
			"ドメインのレイズ回数が上限を超えている")
		if !g.IsHumanTurn() {
			g.CpuPlay()
			continue
		}
		if err := g.PlayerAction(BaseballActionRaise, 10); err != nil {
			if errors.Is(err, errBaseballRaiseCapped) {
				sawCap = true
				assert.Equal(t, baseballMaxRaisesPerRound, g.GetRaiseCount(),
					"上限で弾いたのに回数が上限に達していない")
				assert.False(t, g.CanRaise(), "上限に達したのに CanRaise が真")
				break
			}
			// 額が足りないなど別の理由なら、この配りでは上限まで押せない。
			break
		}
		raises++
		require.LessOrEqual(t, raises, baseballMaxRaisesPerRound,
			"1 ラウンドで上限を超えてレイズが通っている")
	}
	_ = sawCap
}

// --- 終局と保存則 ---

// **ハンドは必ず終わる。** 買い増しは 3 が出るたびに割り込むので、素直に
// 書くと解決の途中で同じ席に戻って止まらなくなる。
func TestBaseballPoker_HandsAlwaysTerminate(t *testing.T) {
	for hand := range 30 {
		g := newBaseballForTest(t)
		baseballDriveToShowdown(t, g)
		require.True(t, g.GetPhase() == BaseballPhaseShowdown || g.GetGameEndFlag(),
			"%d 局目が終わらなかった", hand)
	}
}

// **チップは増えも減りもしない。** 決着後にポットが空であることも見る ──
// 総量だけでは、取り残されたチップが検出できない。
func TestBaseballPoker_ChipsAreConservedAndThePotEmpties(t *testing.T) {
	for hand := range 30 {
		g := newBaseballForTest(t)
		want := baseballTotalChips(g)
		baseballDriveToShowdown(t, g)
		assert.Equal(t, want, baseballTotalChips(g), "%d 局目でチップ総量が動いた", hand)
		assert.Zero(t, g.GetPot(), "%d 局目の決着後にポットが残っている", hand)
	}
}

// **山は尽きない。** 7 席だと 4 のボーナスで 53 枚要るので、設定で弾いている。
func TestBaseballPoker_DeckNeverRunsOut(t *testing.T) {
	for range 30 {
		g := newBaseballForTest(t)
		baseballDriveToShowdown(t, g)
		assert.GreaterOrEqual(t, g.GetRemainingCards(), 0)
		for i, p := range g.GetPlayers() {
			for k, c := range p.GetCards() {
				require.NotNil(t, c, "席 %d の %d 枚目が nil (山が尽きた)", i, k)
			}
		}
	}
}

func TestBaseballPokerConfig_RejectsATableTheDeckCannotServe(t *testing.T) {
	assert.ErrorIs(t, BaseballPokerConfig{Seats: 7, InitialChips: 1000, Ante: 10}.Validate(),
		errBaseballSeatsRange)
	// 範囲内でも山が足りなければ弾く (上限を広げても守られることを見る)。
	assert.ErrorIs(t, BaseballPokerConfig{Seats: 1, InitialChips: 1000, Ante: 10}.Validate(),
		errBaseballSeatsRange)
	assert.ErrorIs(t, BaseballPokerConfig{Seats: 4, InitialChips: 10, Ante: 10}.Validate(),
		errBaseballChipsRange)
	assert.ErrorIs(t, BaseballPokerConfig{Seats: 4, InitialChips: 1000, Ante: 1}.Validate(),
		errBaseballAnteRange)
	assert.NoError(t, DefaultBaseballPokerConfig().Validate())

	// **最大席数は山の枚数に収まる。** 6 席 × 7 枚 + ボーナス 4 枚 = 46 枚。
	assert.LessOrEqual(t, BaseballMaxSeats*BaseballBaseCards+BaseballMaxBonusCards, 52)
}

// **人間が破産したら終わる。** チップ 0 の席にアンティを払わせ続けない。
func TestBaseballPoker_EndsWhenTheHumanIsBroke(t *testing.T) {
	g := newBaseballForTest(t)
	human := g.GetPlayers()[g.HumanSeat()]
	human.SetChips(0)
	// **勝ってしまうと破産しない。** 降ろしてからハンドを閉じる。
	human.SetFolded(true)
	for steps := 0; g.GetPhase() != BaseballPhaseShowdown && !g.GetGameEndFlag(); steps++ {
		require.Less(t, steps, 400)
		if g.IsHumanBuying() {
			require.NoError(t, g.AnswerBuyIn(BaseballBuyFold))
			continue
		}
		g.CpuPlay()
	}
	require.Zero(t, human.GetChips(), "降りた席にチップが入っている")
	assert.True(t, g.GetGameEndFlag(), "人間が破産しても終局していない")
	assert.ErrorIs(t, g.NextHand(), errBaseballGameOver)
	assert.ErrorIs(t, g.PlayerAction(BaseballActionCheck, 0), errBaseballGameOver)
	assert.ErrorIs(t, g.AnswerBuyIn(BaseballBuyPay), errBaseballGameOver)
	assert.Nil(t, g.GetHint(), "終局後に助言が出ている")
}

func TestBaseballPoker_NextHandNeedsAFinishedHand(t *testing.T) {
	g := newBaseballForTest(t)
	if g.GetPhase() != BaseballPhaseShowdown {
		assert.ErrorIs(t, g.NextHand(), errBaseballHandInProgress)
	}
	baseballDriveToShowdown(t, g)
	if !g.GetGameEndFlag() {
		require.NoError(t, g.NextHand())
		assert.Equal(t, 2, g.GetHandNumber())
		assert.Equal(t, 1, g.GetStreet(), "次のハンドで配り直されていない")
	}
}

// --- 助言 ---

func TestBaseballPoker_HintCoversBothDecisions(t *testing.T) {
	g := newBaseballForTest(t)
	if g.IsHumanTurn() {
		h := g.GetHint()
		require.NotNil(t, h, "手番なのに助言が無い")
		assert.Contains(t, []string{"fold", "check", "call", "bet", "raise"}, h.Action)
		assert.NotEmpty(t, h.Reason)
	}

	// 買い増しの場面では pay / fold を薦める。
	for range 200 {
		if g.IsHumanBuying() {
			h := g.GetHint()
			require.NotNil(t, h, "買い増しを迫られているのに助言が無い")
			assert.Contains(t, []string{"pay", "fold"}, h.Action)
			return
		}
		if g.GetPhase() == BaseballPhaseShowdown || g.GetGameEndFlag() {
			if g.GetGameEndFlag() {
				return
			}
			require.NoError(t, g.NextHand())
			continue
		}
		if g.IsHumanTurn() {
			if err := g.PlayerAction(BaseballActionCheck, 0); err != nil {
				require.NoError(t, g.PlayerAction(BaseballActionCall, 0))
			}
			continue
		}
		g.CpuPlay()
	}
}

// --- 進行の駆動 ---

// **人間の 1 手のあと、盤面は人間の判断が要る場面まで戻る。**
func TestBaseballPoker_OneActionReturnsControl(t *testing.T) {
	for range 30 {
		g := newBaseballForTest(t)
		if !g.IsHumanTurn() {
			continue
		}
		require.NoError(t, g.PlayerAction(BaseballActionCheck, 0))
		g.CpuPlay()
		if g.GetPhase() == BaseballPhaseShowdown || g.GetGameEndFlag() {
			continue
		}
		require.True(t, g.IsHumanTurn() || g.IsHumanBuying(),
			"人間が 1 手指したのに席 %d (CPU) で盤面が止まっている", g.GetTurnSeat())
	}
}

func TestBaseballPoker_ActionLogRecordsTheEvents(t *testing.T) {
	g := newBaseballForTest(t)
	baseballDriveToShowdown(t, g)
	log := g.GetActionLog()
	require.NotEmpty(t, log)
	kinds := map[string]bool{}
	for _, e := range log {
		kinds[e.ActionType] = true
	}
	assert.True(t, kinds["deal"], "配札が棋譜に残っていない")
}

func TestBaseballPoker_Accessors(t *testing.T) {
	g := newBaseballForTest(t)
	assert.Len(t, g.GetPlayers(), BaseballDefaultSeats)
	assert.Equal(t, 0, g.HumanSeat())
	assert.Equal(t, 1, g.GetHandNumber())
	assert.Equal(t, DefaultBaseballPokerConfig(), g.GetConfig())
	assert.GreaterOrEqual(t, g.WinnerSeat(), 0)
	assert.GreaterOrEqual(t, g.GetToCall(), 0)
	assert.GreaterOrEqual(t, g.GetRaiseCount(), 0)
	assert.NotNil(t, g.GetResults())

	cfg := BaseballPokerConfig{Seats: 3, InitialChips: 500, Ante: 5}
	g.SetConfig(cfg)
	assert.Equal(t, cfg, g.GetConfig())
	g.Reset()
	assert.Len(t, g.GetPlayers(), 3)
}

// **拠出を超えて勝てない。** レビュー指摘 (#5334)。
//
// 表の 3 の買い増しは払えない席をその場でオールインにするので、拠出額の
// 違う席が同じハンドに何段も並ぶ。ポット全体を最高役に渡すと、早くに
// オールインした短いスタックが、その後のラウンドで自分が付き合えなかった
// チップまで持っていく ── **卓の総量は変わらないので保存則のテストでは
// 絶対に見つからない**。
func TestBaseballPoker_ShortStackWinsOnlyTheMainPot(t *testing.T) {
	g := NewDefaultBaseballPoker()
	g.Reset()

	players := g.GetPlayers()
	require.GreaterOrEqual(t, len(players), 3, "3 席以上でないと段が作れない")

	// 席 0 が 50 だけ出してオールイン、席 1 と 2 が 200 ずつ出した局面を組む。
	// 開始時のチップを基準に拠出が決まるので、そこから逆算して積む。
	g.startingChips = []int{50, 250, 250}
	for i := range g.startingChips {
		if i >= len(players) {
			break
		}
		g.startingChips[i] = []int{50, 250, 250}[i]
	}
	for i := 3; i < len(players); i++ {
		// 残りの席は降りている扱いにして、段の計算から外す。
		g.startingChips = append(g.startingChips, players[i].GetChips())
		players[i].SetFolded(true)
	}

	players[0].SetChips(0)
	players[0].SetAllIn(true)
	players[0].SetFolded(false)
	players[1].SetChips(50)
	players[1].SetFolded(false)
	players[2].SetChips(50)
	players[2].SetFolded(false)
	g.pot = 50 + 200 + 200

	// 席 0 に最強、席 1 にその次の手を持たせる。
	for _, p := range players {
		p.cards = p.cards[:0]
		p.faceUp = p.faceUp[:0]
	}
	// 席 0: A のフォーカード相当 (ワイルド 2 枚 + A 2 枚)。
	for _, c := range []*Card{
		NewCard(CardDesignSpade, 1, true), NewCard(CardDesignHeart, 1, true),
		NewCard(CardDesignSpade, BaseballWildThree, true), NewCard(CardDesignHeart, BaseballWildNine, true),
		NewCard(CardDesignClover, 7, true),
	} {
		players[0].AddDealtCard(c, true)
	}
	// 席 1: ワンペア。
	for _, c := range []*Card{
		NewCard(CardDesignSpade, 5, true), NewCard(CardDesignHeart, 5, true),
		NewCard(CardDesignClover, 8, true), NewCard(CardDesignDiamond, 11, true),
		NewCard(CardDesignSpade, 12, true),
	} {
		players[1].AddDealtCard(c, true)
	}
	// 席 2: ハイカード。
	for _, c := range []*Card{
		NewCard(CardDesignSpade, 2, true), NewCard(CardDesignHeart, 4, true),
		NewCard(CardDesignClover, 6, true), NewCard(CardDesignDiamond, 8, true),
		NewCard(CardDesignSpade, 10, true),
	} {
		players[2].AddDealtCard(c, true)
	}

	totalBefore := baseballTotalChips(g)
	g.finishHand()

	assert.Equal(t, totalBefore, baseballTotalChips(g), "チップ総量が動いた")
	assert.Zero(t, g.GetPot(), "決着後にポットが残っている")

	// **メインポットは 50 × 3 = 150 まで。** 残り 300 は席 1 と 2 の勝負。
	assert.Equal(t, 150, g.GetResults()[0].WonAmount,
		"オールインの席が拠出を超えて受け取っている (サイドポットが効いていない)")
	assert.Equal(t, 300, g.GetResults()[1].WonAmount,
		"サイドポットが 2 番手に渡っていない")
	assert.Zero(t, g.GetResults()[2].WonAmount)
}

// **どのストリートも棋譜に残す。** レビュー指摘 (#5334)。3rd だけ記録して
// 4th〜7th を落とすと、棋譜が配札の半分を語らないものになる ── 「deal が
// 1 回でも出ていれば通る」検査では、落ちていても気づけない。
func TestBaseballPoker_ActionLogRecordsEveryStreet(t *testing.T) {
	g := newBaseballForTest(t)
	baseballDriveToShowdown(t, g)

	deals := 0
	for _, e := range g.GetActionLog() {
		if e.ActionType == "deal" {
			deals++
		}
	}
	// 3rd ストリートは席ごとに 1 行 + 4th〜7th の 4 行。
	want := len(g.GetPlayers()) + baseballStreets
	assert.GreaterOrEqual(t, deals, want,
		"配札の棋譜が %d 行しかない (4th〜7th が落ちている)", deals)
}
