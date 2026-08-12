//go:build test

package domain

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// chemindeFerCard は値 v のスペードを返す。スートは合計値に関係しない。
func chemindeFerCard(v int) *Card { return NewCard(CardDesignSpade, v, false) }

// chemindeFerHandWorth は合計が total になる 2 枚を返す。
//
// 10/J/Q/K が 0 なので、**1 枚を K (=0) に固定すれば残り 1 枚で合計を作れる**。
func chemindeFerHandWorth(total int) []*Card {
	return []*Card{chemindeFerCard(13), chemindeFerCard(total)}
}

// newChemindeFerAllCpu は人間の居ない卓を返す。CPU 同士で最後まで進められる。
func newChemindeFerAllCpu(t *testing.T, seed int64) *ChemindeFer {
	t.Helper()
	cfg := DefaultChemindeFerConfig()
	players := make([]*ChemindeFerPlayer, ChemindeFerSeatCnt)
	for i := range players {
		players[i] = NewChemindeFerPlayer("cpu", cfg.InitialChips, false)
	}
	g := NewChemindeFer(newChemindeFerShoe(), players, cfg)
	g.SetRand(rand.New(rand.NewSource(seed)))
	// **Reset は人間の番まで自動で進む。** 人間の居ない卓では 1 ラウンド走り切って
	// しまうので、張り待ちの状態を観察したいテストのために巻き戻す。
	g.reset()
	return g
}

// --- 合計値 ---

func TestChemindeFerHandTotal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		cards []*Card
		want  int
	}{
		{"空の手札は 0", nil, 0},
		{"A は 1", []*Card{chemindeFerCard(1)}, 1},
		{"絵札は 0", []*Card{chemindeFerCard(11), chemindeFerCard(12), chemindeFerCard(13)}, 0},
		{"10 も 0", []*Card{chemindeFerCard(10)}, 0},
		{"9+7 は 16 なので 6", []*Card{chemindeFerCard(9), chemindeFerCard(7)}, 6},
		{"9+9+9 は 27 なので 7", []*Card{chemindeFerCard(9), chemindeFerCard(9), chemindeFerCard(9)}, 7},
		{"5+5 はちょうど 10 なので 0", []*Card{chemindeFerCard(5), chemindeFerCard(5)}, 0},
		{"nil 札は 0 として数える", []*Card{nil, chemindeFerCard(4)}, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ChemindeFerHandTotal(tt.cards))
		})
	}
}

// **子の引き方の規則は 0-9 の全域で定義されている。**
func TestChemindeFer_PunterRuleCoversEveryTotal(t *testing.T) {
	t.Parallel()

	for total := range 10 {
		draw := ChemindeFerPunterMustDraw(total)
		stand := ChemindeFerPunterMustStand(total)
		choose := ChemindeFerPunterMayChoose(total)

		assert.False(t, draw && stand, "合計 %d が引きと立ちの両方を強制している", total)
		assert.Equal(t, 1,
			boolToIntChemindeFer(draw)+boolToIntChemindeFer(stand)+boolToIntChemindeFer(choose),
			"合計 %d でちょうど 1 つの扱いになっていない", total)

		switch {
		case total <= 4:
			assert.True(t, draw, "合計 %d は引かされる", total)
		case total == 5:
			assert.True(t, choose, "合計 5 だけが子の自由")
		default:
			assert.True(t, stand, "合計 %d は立たされる", total)
		}
	}
}

func boolToIntChemindeFer(b bool) int {
	if b {
		return 1
	}
	return 0
}

func TestChemindeFerIsNatural(t *testing.T) {
	t.Parallel()

	for total := range 10 {
		assert.Equal(t, total >= 8, ChemindeFerIsNatural(total), "合計 %d", total)
	}
}

// --- 張りと賭け ---

// chemindeFerAtBet は席 0 が親で張り済み、席 1 の賭け待ちの卓を返す。
func chemindeFerAtBet(t *testing.T, stake int) *ChemindeFer {
	t.Helper()
	g := newChemindeFerAllCpu(t, 7)
	require.Equal(t, ChemindeFerPhaseStake, g.GetPhase())
	// **CPU を進めない内部版を使う。** 公開版は次の人間の番まで走り切るので、
	// 人間の居ない卓では賭けの途中を観察できない。
	require.NoError(t, g.setStake(stake))
	require.Equal(t, ChemindeFerPhaseBet, g.GetPhase())
	return g
}

func TestChemindeFer_SetStakeRejectsOutOfRange(t *testing.T) {
	g := newChemindeFerAllCpu(t, 3)
	chips := g.GetPlayer(g.GetBankerIdx()).GetChips()

	assert.ErrorIs(t, g.SetStake(ChemindeFerStakeMin-1), errChemindeFerStakeRange)
	assert.ErrorIs(t, g.SetStake(chips+1), errChemindeFerStakeRange)
	assert.ErrorIs(t, g.SetStake(-1), errChemindeFerStakeRange)
	require.NoError(t, g.SetStake(chips), "手持ち全額はちょうど張れる")
}

func TestChemindeFer_SetStakeIsOnlyLegalInTheStakePhase(t *testing.T) {
	g := chemindeFerAtBet(t, 100)
	assert.ErrorIs(t, g.SetStake(50), errChemindeFerWrongPhase)
}

// **賭けの総額はバンク額を超えない。** 超えたら親が払えない額を晒すことになる。
func TestChemindeFer_BetsCannotExceedTheBank(t *testing.T) {
	g := chemindeFerAtBet(t, 100)

	first := g.GetBetTurn()
	require.GreaterOrEqual(t, first, 0)
	_, hi := g.BetRangeFor(first)
	assert.Equal(t, 100, hi, "最初の子はバンク額まで賭けられる")
	assert.ErrorIs(t, g.PlaceBet(first, 101), errChemindeFerBetRange)

	require.NoError(t, g.placeBet(first, 60))
	assert.Equal(t, 40, g.GetRemainingStake())

	second := g.GetBetTurn()
	require.GreaterOrEqual(t, second, 0)
	_, hi2 := g.BetRangeFor(second)
	assert.Equal(t, 40, hi2, "2 人目は残り分しか賭けられない")
	assert.ErrorIs(t, g.PlaceBet(second, 41), errChemindeFerBetRange)
}

// **バンク額が覆い尽くされたら、順番が残っていても賭けは締め切る。**
func TestChemindeFer_BettingClosesOnceTheBankIsCovered(t *testing.T) {
	g := chemindeFerAtBet(t, 100)

	first := g.GetBetTurn()
	require.NoError(t, g.placeBet(first, 100))

	assert.NotEqual(t, ChemindeFerPhaseBet, g.GetPhase(), "まだ賭けを受け付けている")
	assert.Equal(t, -1, g.GetBetTurn(), "順番が残っている")
	assert.Equal(t, first, g.GetRepresentativeIdx(), "全額を覆った子が代表になる")
}

func TestChemindeFer_PlaceBetRejectsOtherSeats(t *testing.T) {
	g := chemindeFerAtBet(t, 100)
	turn := g.GetBetTurn()
	other := (turn + 1) % ChemindeFerSeatCnt
	if other == g.GetBankerIdx() {
		other = (other + 1) % ChemindeFerSeatCnt
	}
	require.NotEqual(t, turn, other)
	assert.ErrorIs(t, g.PlaceBet(other, 10), errChemindeFerNotYourSeat)
}

// **同額なら親に近い側が代表。**
func TestChemindeFer_HighestBettorRepresentsTheseWithTiesToTheBankersRight(t *testing.T) {
	g := chemindeFerAtBet(t, 100)

	first := g.GetBetTurn()
	require.NoError(t, g.placeBet(first, 30))
	second := g.GetBetTurn()
	require.NoError(t, g.placeBet(second, 30))

	// 賭けはまだ締まっていないので、代表は配りの直前に決まる。同額なら先の席。
	for g.GetPhase() == ChemindeFerPhaseBet {
		require.NoError(t, g.placeBet(g.GetBetTurn(), 0))
	}
	assert.Equal(t, first, g.GetRepresentativeIdx())
}

// 誰も乗らなかったラウンドは流れ、**バンクが隣へ渡る**。
func TestChemindeFer_RoundIsVoidWhenNobodyCovers(t *testing.T) {
	g := chemindeFerAtBet(t, 100)
	banker := g.GetBankerIdx()

	for g.GetPhase() == ChemindeFerPhaseBet {
		require.NoError(t, g.placeBet(g.GetBetTurn(), 0))
	}

	assert.Equal(t, ChemindeFerPhaseRoundEnd, g.GetPhase())
	assert.Equal(t, ChemindeFerResultNone, g.GetResult())
	assert.NotEqual(t, banker, g.GetBankerIdx(), "乗り手が居なければバンクは隣へ渡る")
	assert.Empty(t, g.GetBankerHand(), "配られていない")
}

// --- 引きの規則 ---

// chemindeFerDealt は指定の合計の手札を積んだ、引きの判断直前の卓を返す。
//
// **配りは乱数のままでは固定できない** (TrumpCards.Shuffle はグローバルな rand を
// 使う) ので、賭けまで進めてから手札を直接差し替える。
func chemindeFerDealt(t *testing.T, punterTotal, bankerTotal int, humanSeats ...int) *ChemindeFer {
	t.Helper()
	cfg := DefaultChemindeFerConfig()
	players := make([]*ChemindeFerPlayer, ChemindeFerSeatCnt)
	human := make(map[int]bool, len(humanSeats))
	for _, s := range humanSeats {
		human[s] = true
	}
	for i := range players {
		players[i] = NewChemindeFerPlayer("seat", cfg.InitialChips, human[i])
	}
	g := NewChemindeFer(newChemindeFerShoe(), players, cfg)
	g.SetRand(rand.New(rand.NewSource(11)))
	// **CPU を進めない reset を使う。** 公開版は人間の番まで走るので、席が全部
	// CPU だと 1 ラウンド遊び切ってしまい、チップが初期値から動いてしまう。
	g.reset()

	// **PlaceBet で配らせてはいけない。** PlaceBet はそのまま決着まで走り切るので、
	// 後から手札を差し替えて resolve を呼ぶと、同じ 1 回のクーが 2 度精算される。
	// 局面は手で組む。
	const bet = 100
	g.stake = bet
	g.betOrder = g.buildBetOrder()
	g.betPos = -1
	require.True(t, g.players[1].SubtractChips(bet))
	g.players[1].SetBet(bet)
	g.represIdx = 1
	g.punterHand = chemindeFerHandWorth(punterTotal)
	g.bankerHand = chemindeFerHandWorth(bankerTotal)
	g.phase = ChemindeFerPhasePunterDraw
	return g
}

// **ナチュラルが出たら 3 枚目は無い。**
func TestChemindeFer_NaturalEndsTheCoupAtOnce(t *testing.T) {
	tests := []struct {
		name           string
		punter, banker int
		want           ChemindeFerResult
	}{
		{"子が 9 で親が 7", 9, 7, ChemindeFerResultPunter},
		{"親が 8 で子が 6", 6, 8, ChemindeFerResultBanker},
		{"両者 8 は引き分け", 8, 8, ChemindeFerResultTie},
		{"子 9 対 親 8", 9, 8, ChemindeFerResultPunter},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := chemindeFerDealt(t, tt.punter, tt.banker)
			g.afterDeal()

			assert.Equal(t, ChemindeFerPhaseRoundEnd, g.GetPhase())
			assert.Equal(t, tt.want, g.GetResult())
			assert.Len(t, g.GetPunterHand(), ChemindeFerHandSize, "3 枚目を引いている")
			assert.Len(t, g.GetBankerHand(), ChemindeFerHandSize, "3 枚目を引いている")
		})
	}
}

// **子は 0-4 と 6-7 では選べない。** 選ぼうとしたら弾く。
func TestChemindeFer_PunterCannotChooseOutsideFive(t *testing.T) {
	for total := range 8 {
		if total == ChemindeFerPunterFreeTotal {
			continue
		}
		t.Run(chemindeFerTotalName(total), func(t *testing.T) {
			// 親を人間にして、子の自動処理の後で止まるようにする。
			g := chemindeFerDealt(t, total, 0, 0)
			g.bankerIdx = 0
			g.phase = ChemindeFerPhasePunterDraw
			g.punterHand = chemindeFerHandWorth(total)

			assert.False(t, g.PunterMayChoose(), "合計 %d に選択の余地は無い", total)
			assert.ErrorIs(t, g.PunterDraw(), errChemindeFerNoChoice)
			assert.ErrorIs(t, g.PunterStand(), errChemindeFerNoChoice)
		})
	}
}

func chemindeFerTotalName(total int) string {
	return "合計" + string(rune('0'+total))
}

// **合計 5 の子だけが選べる。** 引いても立ってもよい。
func TestChemindeFer_PunterChoosesAtFive(t *testing.T) {
	for _, draw := range []bool{true, false} {
		name := "立つ"
		if draw {
			name = "引く"
		}
		t.Run(name, func(t *testing.T) {
			g := chemindeFerDealt(t, ChemindeFerPunterFreeTotal, 0, 0)
			g.bankerIdx = 0 // 親を人間にして親の判断で止める
			g.represIdx = 1
			g.phase = ChemindeFerPhasePunterDraw
			g.punterHand = chemindeFerHandWorth(ChemindeFerPunterFreeTotal)

			require.True(t, g.PunterMayChoose())
			if draw {
				require.NoError(t, g.PunterDraw())
				assert.Len(t, g.GetPunterHand(), ChemindeFerMaxHandSize)
				assert.True(t, g.GetPunterDrew())
			} else {
				require.NoError(t, g.PunterStand())
				assert.Len(t, g.GetPunterHand(), ChemindeFerHandSize)
				assert.False(t, g.GetPunterDrew())
			}
			assert.Equal(t, ChemindeFerPhaseBankerDraw, g.GetPhase())
		})
	}
}

// **これがこのゲームの核心。親はどの合計でも引くことも立つこともできる。**
//
// プント・バンコならここが表で固定されている。固定されていないことを、
// 0-7 のすべてで両方の手が通ることとして押さえる。
func TestChemindeFer_BankerMayAlwaysChoose(t *testing.T) {
	for total := range 8 {
		for _, draw := range []bool{true, false} {
			t.Run(chemindeFerTotalName(total), func(t *testing.T) {
				g := chemindeFerDealt(t, 0, total, 0)
				g.bankerIdx = 0
				g.represIdx = 1
				g.phase = ChemindeFerPhaseBankerDraw
				g.bankerHand = chemindeFerHandWorth(total)

				var err error
				if draw {
					err = g.BankerDraw()
				} else {
					err = g.BankerStand()
				}
				require.NoError(t, err, "合計 %d で親の選択が拒否された", total)
				assert.Equal(t, ChemindeFerPhaseRoundEnd, g.GetPhase())
			})
		}
	}
}

// 引きの操作はフェーズ外では通らない。
func TestChemindeFer_DrawActionsRejectWrongPhase(t *testing.T) {
	g := newChemindeFerAllCpu(t, 5)
	assert.ErrorIs(t, g.PunterDraw(), errChemindeFerWrongPhase)
	assert.ErrorIs(t, g.PunterStand(), errChemindeFerWrongPhase)
	assert.ErrorIs(t, g.BankerDraw(), errChemindeFerWrongPhase)
	assert.ErrorIs(t, g.BankerStand(), errChemindeFerWrongPhase)
}

// --- 決着とバンクの行方 ---

func TestChemindeFer_SettlementMovesTheRightChips(t *testing.T) {
	tests := []struct {
		name           string
		punter, banker int
		wantResult     ChemindeFerResult
		wantBankerDiff int
		wantPunterDiff int
	}{
		{"親の勝ち", 3, 7, ChemindeFerResultBanker, +100, -100},
		{"子の勝ち", 7, 3, ChemindeFerResultPunter, -100, +100},
		{"引き分けは動かない", 5, 5, ChemindeFerResultTie, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := chemindeFerDealt(t, tt.punter, tt.banker)
			banker := g.GetBankerIdx()
			repres := g.GetRepresentativeIdx()
			require.GreaterOrEqual(t, repres, 0)

			before := DefaultChemindeFerConfig().InitialChips
			g.resolve()

			assert.Equal(t, tt.wantResult, g.GetResult())
			assert.Equal(t, before+tt.wantBankerDiff, g.GetPlayer(banker).GetChips(), "親のチップ")
			assert.Equal(t, before+tt.wantPunterDiff, g.GetPlayer(repres).GetChips(), "子のチップ")
		})
	}
}

// **バンクは負けたときだけ渡る。**
func TestChemindeFer_BankPassesOnlyWhenTheBankerLoses(t *testing.T) {
	tests := []struct {
		name           string
		punter, banker int
		wantPass       bool
	}{
		{"親が勝てば親のまま", 3, 7, false},
		{"引き分けでも親のまま", 5, 5, false},
		{"親が負けたら隣へ", 7, 3, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := chemindeFerDealt(t, tt.punter, tt.banker)
			before := g.GetBankerIdx()
			g.resolve()

			if tt.wantPass {
				assert.NotEqual(t, before, g.GetBankerIdx())
			} else {
				assert.Equal(t, before, g.GetBankerIdx())
			}
		})
	}
}

// --- 保存則 ---

// **卓のチップ総額は 1 手ごとに変わらない。**
//
// 賭け・配当・引き分けの返却・バンクの移動のどこかで 1 枚でも湧いたり消えたりすれば、
// ここで落ちる。ゼロサムのゲームで最初に書くべき検査。
func TestChemindeFer_ChipsAreConservedThroughAWholeSession(t *testing.T) {
	for seed := range 20 {
		g := newChemindeFerAllCpu(t, int64(seed)+1)
		want := g.GetTotalChips()
		require.Equal(t, ChemindeFerSeatCnt*DefaultChemindeFerConfig().InitialChips, want)

		for step := 0; step < 5000 && !g.GetGameEndFlag(); step++ {
			if g.GetPhase() == ChemindeFerPhaseRoundEnd {
				require.NoError(t, g.NextRound())
			} else {
				g.CpuPlay()
			}
			require.Equal(t, want, g.GetTotalChips(),
				"seed %d の %d 手目でチップ総額が %d に変わった", seed, step, g.GetTotalChips())
		}
	}
}

// **CPU だけの卓は必ず終わる。**
//
// 引き取り手の居ないラウンドを「同じ親が張り直す」形にすると Stake と Bet を往復して
// 永久に終わらない。規則が停止性を持つことは保存則とは別に測らないと分からないので、
// 手数を数えて上限と比べる。
func TestChemindeFer_EverySessionTerminates(t *testing.T) {
	const maxSteps = 2000

	for seed := range 50 {
		g := newChemindeFerAllCpu(t, int64(seed)+100)
		steps := 0
		for ; steps < maxSteps && !g.GetGameEndFlag(); steps++ {
			if g.GetPhase() == ChemindeFerPhaseRoundEnd {
				require.NoError(t, g.NextRound())
				continue
			}
			g.CpuPlay()
		}
		require.True(t, g.GetGameEndFlag(),
			"seed %d が %d 手でも終わらなかった (round=%d, phase=%d)",
			seed, maxSteps, g.GetRoundNumber(), g.GetPhase())
		assert.LessOrEqual(t, g.GetRoundNumber(), DefaultChemindeFerConfig().Rounds)
	}
}

// ラウンド数を使い切ったらゲームが終わる。
func TestChemindeFer_SessionEndsAfterTheConfiguredRounds(t *testing.T) {
	g := newChemindeFerAllCpu(t, 42)
	g.SetConfig(ChemindeFerConfig{Rounds: ChemindeFerRoundsMin, InitialChips: ChemindeFerDefaultChips})

	for step := 0; step < 2000 && !g.GetGameEndFlag(); step++ {
		if g.GetPhase() == ChemindeFerPhaseRoundEnd {
			require.NoError(t, g.NextRound())
			continue
		}
		g.CpuPlay()
	}
	require.True(t, g.GetGameEndFlag())
	assert.Equal(t, ChemindeFerRoundsMin, g.GetRoundNumber())
	assert.ErrorIs(t, g.NextRound(), errChemindeFerGameFinished)
	assert.ErrorIs(t, g.SetStake(50), errChemindeFerGameFinished)
}

func TestChemindeFer_NextRoundRejectsMidCoup(t *testing.T) {
	g := chemindeFerAtBet(t, 100)
	assert.ErrorIs(t, g.NextRound(), errChemindeFerWrongPhase)
}

// --- 親の持ち回りと手番 ---

func TestChemindeFer_PassBankIsOnlyLegalAtRoundEnd(t *testing.T) {
	g := chemindeFerAtBet(t, 100)
	assert.ErrorIs(t, g.PassBank(), errChemindeFerWrongPhase)
}

func TestChemindeFer_PassBankMovesTheBankOnDemand(t *testing.T) {
	g := chemindeFerDealt(t, 3, 7) // 親の勝ち = 本来は親のまま
	g.resolve()
	before := g.GetBankerIdx()
	require.NoError(t, g.PassBank())
	assert.NotEqual(t, before, g.GetBankerIdx(), "自分から降りればバンクは渡る")
}

// **張れる席が 1 つも無くなればゲームは終わる。**
func TestChemindeFer_GameEndsWhenNobodyCanBank(t *testing.T) {
	g := chemindeFerDealt(t, 3, 7)
	g.resolve()
	for i := range ChemindeFerSeatCnt {
		g.GetPlayer(i).SetChips(0)
	}
	require.NoError(t, g.PassBank())
	assert.True(t, g.GetGameEndFlag())
}

func TestChemindeFer_IsHumanTurnFollowsTheRole(t *testing.T) {
	// 席 0 が人間で親。張りは人間の番。
	g := chemindeFerDealt(t, 0, 0, 0)
	g.bankerIdx = 0
	g.phase = ChemindeFerPhaseStake
	assert.True(t, g.IsHumanTurn(), "人間が親なら張りは人間の番")

	// 親が CPU なら張りは人間の番ではない。
	g.bankerIdx = 1
	assert.False(t, g.IsHumanTurn())

	// 決着後は誰の番でもない。
	g.phase = ChemindeFerPhaseRoundEnd
	assert.False(t, g.IsHumanTurn())

	// 終了後も同じ。
	g.gameEndFlag = true
	assert.False(t, g.IsHumanTurn())
}

// 人間の番では CpuPlay が何もしない。
func TestChemindeFer_CpuPlayDoesNothingOnAHumanTurn(t *testing.T) {
	g := chemindeFerDealt(t, 0, 0, 0)
	g.bankerIdx = 0
	g.phase = ChemindeFerPhaseStake
	g.stake = 0
	require.True(t, g.IsHumanTurn())

	g.CpuPlay()
	assert.Equal(t, 0, g.GetStake(), "人間の番なのに CPU が張った")
	assert.Equal(t, ChemindeFerPhaseStake, g.GetPhase())
}

// --- CPU 親の戦略 ---

// **戦略であって規則ではない。** ここが表と一致していても、親は従う義務が無い。
func TestChemindeFerCpuBankerDraws(t *testing.T) {
	t.Parallel()

	t.Run("子が立ったら 5 以下で引く", func(t *testing.T) {
		for total := range 10 {
			assert.Equal(t, total <= 5, chemindeFerCpuBankerDraws(total, nil), "合計 %d", total)
		}
	})

	tests := []struct {
		banker, third int
		want          bool
	}{
		{0, 9, true}, {1, 9, true}, {2, 9, true},
		{3, 8, false}, {3, 7, true},
		{4, 1, false}, {4, 2, true}, {4, 7, true}, {4, 8, false},
		{5, 3, false}, {5, 4, true}, {5, 7, true}, {5, 8, false},
		{6, 5, false}, {6, 6, true}, {6, 7, true}, {6, 8, false},
		{7, 6, false}, {8, 6, false}, {9, 6, false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want,
			chemindeFerCpuBankerDraws(tt.banker, chemindeFerCard(tt.third)),
			"親 %d に対し子の 3 枚目が %d", tt.banker, tt.third)
	}
}

// --- ヒント ---

func TestChemindeFer_GetHint(t *testing.T) {
	t.Run("判断どころでなければ nil", func(t *testing.T) {
		g := newChemindeFerAllCpu(t, 9)
		assert.Nil(t, g.GetHint(), "人間の居ない卓に助言は要らない")
	})

	t.Run("合計 5 の子には引きを薦める", func(t *testing.T) {
		g := chemindeFerDealt(t, ChemindeFerPunterFreeTotal, 0, 1)
		g.bankerIdx = 0
		g.represIdx = 1
		g.phase = ChemindeFerPhasePunterDraw
		g.punterHand = chemindeFerHandWorth(ChemindeFerPunterFreeTotal)

		hint := g.GetHint()
		require.NotNil(t, hint)
		assert.True(t, hint.Draw)
		assert.Equal(t, "punterFive", hint.Reason)
	})

	t.Run("親には合計に応じた助言", func(t *testing.T) {
		g := chemindeFerDealt(t, 0, 2, 0)
		g.bankerIdx = 0
		g.represIdx = 1
		g.phase = ChemindeFerPhaseBankerDraw
		g.bankerHand = chemindeFerHandWorth(2)

		hint := g.GetHint()
		require.NotNil(t, hint)
		assert.True(t, hint.Draw, "合計 2 なら引く")
		assert.Equal(t, "bankerDraw", hint.Reason)

		g.bankerHand = chemindeFerHandWorth(7)
		hint = g.GetHint()
		require.NotNil(t, hint)
		assert.False(t, hint.Draw, "合計 7 なら立つ")
		assert.Equal(t, "bankerStand", hint.Reason)
	})

	t.Run("終了後は nil", func(t *testing.T) {
		g := chemindeFerDealt(t, 0, 2, 0)
		g.gameEndFlag = true
		assert.Nil(t, g.GetHint())
	})
}

// --- 名前と設定 ---

func TestChemindeFerNames(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "stake", ChemindeFerPhaseName(ChemindeFerPhaseStake))
	assert.Equal(t, "bet", ChemindeFerPhaseName(ChemindeFerPhaseBet))
	assert.Equal(t, "punterDraw", ChemindeFerPhaseName(ChemindeFerPhasePunterDraw))
	assert.Equal(t, "bankerDraw", ChemindeFerPhaseName(ChemindeFerPhaseBankerDraw))
	assert.Equal(t, "roundEnd", ChemindeFerPhaseName(ChemindeFerPhaseRoundEnd))

	assert.Equal(t, "banker", ChemindeFerResultName(ChemindeFerResultBanker))
	assert.Equal(t, "punter", ChemindeFerResultName(ChemindeFerResultPunter))
	assert.Equal(t, "tie", ChemindeFerResultName(ChemindeFerResultTie))
	assert.Equal(t, "none", ChemindeFerResultName(ChemindeFerResultNone))
}

func TestChemindeFerConfig_Validate(t *testing.T) {
	t.Parallel()

	assert.NoError(t, DefaultChemindeFerConfig().Validate())

	tests := []struct {
		name string
		cfg  ChemindeFerConfig
	}{
		{"ラウンドが少なすぎる", ChemindeFerConfig{Rounds: ChemindeFerRoundsMin - 1, InitialChips: 1000}},
		{"ラウンドが多すぎる", ChemindeFerConfig{Rounds: ChemindeFerRoundsMax + 1, InitialChips: 1000}},
		{"チップが少なすぎる", ChemindeFerConfig{Rounds: 8, InitialChips: ChemindeFerChipsMin - 1}},
		{"チップが多すぎる", ChemindeFerConfig{Rounds: 8, InitialChips: ChemindeFerChipsMax + 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Error(t, tt.cfg.Validate())
		})
	}
}

func TestChemindeFer_Accessors(t *testing.T) {
	g := newChemindeFerAllCpu(t, 13)

	assert.Len(t, g.GetPlayers(), ChemindeFerSeatCnt)
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(ChemindeFerSeatCnt))
	assert.NotNil(t, g.GetPlayer(0))
	assert.Equal(t, ChemindeFerDeckCnt*52, g.GetRemainingCards())
	assert.NotEmpty(t, g.GetActionLog())
	assert.Equal(t, 1, g.GetRoundNumber())
	assert.Equal(t, ChemindeFerDefaultRounds, g.GetConfig().Rounds)
	assert.Equal(t, -1, g.GetRepresentativeIdx())
	assert.Equal(t, 0, g.GetTotalBet())

	// StakeRangeFor は範囲外の席と、張れない席で (0, 0)。
	lo, hi := g.StakeRangeFor(-1)
	assert.Zero(t, lo+hi)
	g.GetPlayer(0).SetChips(ChemindeFerStakeMin - 1)
	lo, hi = g.StakeRangeFor(0)
	assert.Zero(t, lo+hi)

	// BetRangeFor は親自身と範囲外で (0, 0)。
	lo, hi = g.BetRangeFor(g.GetBankerIdx())
	assert.Zero(t, lo+hi)
	lo, hi = g.BetRangeFor(ChemindeFerSeatCnt)
	assert.Zero(t, lo+hi)
}

// 行動ログは上限を超えたら古い方から捨てる。
func TestChemindeFer_ActionLogIsBounded(t *testing.T) {
	g := newChemindeFerAllCpu(t, 17)
	for range chemindeFerMaxSliceLen + 50 {
		g.appendLog(0, "noise", "x", nil)
	}
	assert.Len(t, g.GetActionLog(), chemindeFerMaxSliceLen)
}

// シューが尽きかけたら組み直す。**nil 札を配らないための最後の砦。**
func TestChemindeFer_ShoeIsReplenishedBeforeItRunsOut(t *testing.T) {
	g := newChemindeFerAllCpu(t, 19)
	// ちょうど 6 枚残しでは足りている (1 ラウンドは最大 6 枚)。**下回るまで引く。**
	for g.GetRemainingCards() >= ChemindeFerMaxHandSize*2 {
		g.shoe.DrawCard()
	}
	require.Less(t, g.GetRemainingCards(), ChemindeFerMaxHandSize*2)
	g.ensureShoe()
	assert.Greater(t, g.GetRemainingCards(), ChemindeFerMaxHandSize*2)
}

// newChemindeFerHumanBanker は席 0 だけが人間で、その席が親の卓を返す。
func newChemindeFerHumanBanker(t *testing.T, seed int64) *ChemindeFer {
	t.Helper()
	cfg := DefaultChemindeFerConfig()
	players := make([]*ChemindeFerPlayer, ChemindeFerSeatCnt)
	for i := range players {
		players[i] = NewChemindeFerPlayer("seat", cfg.InitialChips, i == 0)
	}
	g := NewChemindeFer(newChemindeFerShoe(), players, cfg)
	g.SetRand(rand.New(rand.NewSource(seed)))
	g.Reset()
	return g
}

// **人間が操作したら、卓は次の判断どころまで自分で進む。**
//
// これを入れ忘れると、親が張った瞬間に卓が止まる。子は全員 CPU なので誰も賭けず、
// フェーズは Bet のまま、画面には押せるボタンが 1 つも無い。ドメインのテストは
// (CpuPlay を自分で呼んでいたので) 全部緑のまま、**E2E だけが落ちた**。
func TestChemindeFer_AdvancesToTheNextHumanDecision(t *testing.T) {
	for seed := range 20 {
		g := newChemindeFerHumanBanker(t, int64(seed)+1)
		require.Equal(t, ChemindeFerPhaseStake, g.GetPhase(), "seed %d", seed)
		require.True(t, g.IsHumanTurn(), "seed %d: 張りは人間の番のはず", seed)

		require.NoError(t, g.SetStake(200))

		assert.NotEqual(t, ChemindeFerPhaseBet, g.GetPhase(),
			"seed %d: 賭けの途中で止まっている (CPU が進んでいない)", seed)
		assert.True(t, g.IsHumanTurn() || g.GetPhase() == ChemindeFerPhaseRoundEnd,
			"seed %d: 人間の番でもラウンド終了でもない場面で止まった (phase=%d, betTurn=%d)",
			seed, g.GetPhase(), g.GetBetTurn())
		// **精算で賭け金は 0 に戻る**ので、GetTotalBet では「賭けたか」を見られない。
		// 棋譜に賭けの記録が残っていることで確かめる。
		bet := 0
		for _, e := range g.GetActionLog() {
			if e.ActionType == "bet" {
				bet++
			}
		}
		assert.Positive(t, bet, "seed %d: 子が 1 人も賭けていない", seed)
	}
}

// 次のラウンドへ進めたときも同じく自分で進む。
func TestChemindeFer_NextRoundAdvancesToo(t *testing.T) {
	g := newChemindeFerHumanBanker(t, 5)
	require.NoError(t, g.SetStake(200))

	for step := 0; step < 200 && g.GetPhase() != ChemindeFerPhaseRoundEnd; step++ {
		require.True(t, g.IsHumanTurn(), "人間の番でないのに止まっている (phase=%d)", g.GetPhase())
		switch g.GetPhase() {
		case ChemindeFerPhaseBankerDraw:
			require.NoError(t, g.BankerStand())
		case ChemindeFerPhasePunterDraw:
			require.NoError(t, g.PunterStand())
		case ChemindeFerPhaseBet:
			require.NoError(t, g.PlaceBet(g.GetBetTurn(), 0))
		default:
			require.NoError(t, g.SetStake(200))
		}
	}
	require.Equal(t, ChemindeFerPhaseRoundEnd, g.GetPhase())

	require.NoError(t, g.NextRound())
	assert.True(t, g.IsHumanTurn() || g.GetPhase() == ChemindeFerPhaseRoundEnd || g.GetGameEndFlag(),
		"NextRound のあとに誰の番でもない場面で止まった (phase=%d)", g.GetPhase())
}
