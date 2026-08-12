//go:build test

package domain

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newRikkenAllCpu は全席 CPU の卓を返す (自動進行の観測用)。
//
// **全席 CPU だと Reset() の中でラウンドが終わり切ります**（人間の入力待ちが
// 無いため）。途中を観測したいテストは newRikkenWithHuman を使ってください。
func newRikkenAllCpu(t *testing.T, seed int64) *Rikken {
	t.Helper()
	seats := make([]*RikkenPlayer, RikkenPlayerCnt)
	for i := range seats {
		seats[i] = NewRikkenPlayer(false)
	}
	g := NewRikken(NewTrumpCards(0), seats, DefaultRikkenConfig())
	g.SetRand(rand.New(rand.NewSource(seed)))
	return g
}

// newRikkenWithHuman は席 0 が人間の卓を返す。
func newRikkenWithHuman(t *testing.T, seed int64) *Rikken {
	t.Helper()
	g := NewDefaultRikken()
	g.SetRand(rand.New(rand.NewSource(seed)))
	g.Reset()
	return g
}

// rikkenDriveHuman は人間の手番を合法な手で埋めながら進める。
func rikkenDriveHuman(t *testing.T, g *Rikken, steps int, observe func(*Rikken)) int {
	t.Helper()
	seen := 0
	for range steps {
		if g.GetGameEndFlag() {
			return seen
		}
		switch {
		case g.GetPhase() == RikkenPhaseRoundEnd:
			require.NoError(t, g.NextRound())
		case !g.IsHumanTurn():
			g.CpuPlay()
		case g.GetPhase() == RikkenPhaseBid:
			require.NoError(t, g.Bid(RikkenContractNone))
		case g.GetPhase() == RikkenPhaseCall:
			require.NoError(t, g.Call(CardDesignSpade))
		case g.GetPhase() == RikkenPhasePlay:
			if observe != nil {
				observe(g)
				seen++
			}
			valid := g.GetValidPlayIndices(0)
			require.NotEmpty(t, valid, "人間に出せる札が無い")
			require.NoError(t, g.PlayCard(valid[0]))
		default:
			return seen
		}
	}
	return seen
}

// **52 枚を 13 枚ずつ、過不足なく配る。**
func TestRikkenDealsThirteenEach(t *testing.T) {
	t.Parallel()

	g := newRikkenWithHuman(t, 1)
	require.Equal(t, RikkenPhaseBid, g.GetPhase(), "人間の競り待ちで止まる")
	for i := range RikkenPlayerCnt {
		assert.Equal(t, RikkenHandSize, g.GetPlayer(i).GetCardsSize(), "席 %d", i)
	}
	assert.Equal(t, 1, g.GetRoundNumber())
}

// **契約の梯子は上へしか積めない。**
func TestRikkenBidsMustRise(t *testing.T) {
	t.Parallel()

	g := newRikkenWithHuman(t, 2)
	if !g.IsHumanTurn() || g.GetPhase() != RikkenPhaseBid {
		t.Skip("この配りでは人間の競り番が来ない")
	}
	// **固定の契約は使えません。** CPU が先に競り上げていると、その契約は
	// 「上へしか積めない」規則で弾かれます。いまの契約の 1 つ上を積みます。
	below := g.GetContract()
	if below >= RikkenContractMax {
		t.Skip("すでに最強の契約まで競り上がっている")
	}
	require.NoError(t, g.Bid(below+1))
	assert.Equal(t, below+1, g.GetContract())
	assert.Equal(t, 0, g.GetDeclarerIdx(), "競り上げた席が落札者になる")

	// 同じか下は弾く（自分の番がまだ回っているあいだに確かめる）。
	if g.IsHumanTurn() && g.GetPhase() == RikkenPhaseBid {
		assert.Error(t, g.Bid(g.GetContract()), "同じ契約は積めない")
		if g.GetContract() > RikkenContractRik {
			assert.Error(t, g.Bid(g.GetContract()-1), "下の契約は積めない")
		}
	}
}

func TestRikkenBidRejectsOutOfRange(t *testing.T) {
	t.Parallel()

	g := newRikkenWithHuman(t, 3)
	if !g.IsHumanTurn() {
		t.Skip("この配りでは人間の競り番が来ない")
	}
	assert.Error(t, g.Bid(99))
	assert.Error(t, g.Bid(-1))
	assert.Error(t, g.PlayCard(0), "まだプレイの場面ではない")
	assert.Error(t, g.Call(CardDesignSpade), "まだ指名の場面ではない")
	assert.Error(t, g.NextRound(), "まだラウンドの区切りではない")
}

// **全員が降りたら親が Rik を引き受ける。** 流局にすると終わりません。
func TestRikkenAllPassForcesTheDealer(t *testing.T) {
	t.Parallel()

	g := newRikkenAllCpu(t, 5)
	g.Reset()
	// 手で全員パスの状況を作る。
	g.phase = RikkenPhaseBid
	g.contract = RikkenContractNone
	g.declarerIdx = -1
	g.passed = []bool{true, true, true, true}
	g.finishBidding()

	assert.Equal(t, RikkenContractRik, g.GetContract())
	assert.Equal(t, g.GetDealerIdx(), g.GetDeclarerIdx(), "親が引き受ける")
}

// **得点はゼロサム。** 卓の合計は常に 0 です。
func TestRikkenScoresAreZeroSum(t *testing.T) {
	for seed := range 40 {
		g := newRikkenAllCpu(t, int64(seed)+1)
		g.Reset()

		for round := 0; round < 30; round++ {
			total := 0
			for i := range RikkenPlayerCnt {
				total += g.GetPlayer(i).GetScore()
			}
			require.Zero(t, total, "seed %d ラウンド %d: 合計が %d", seed, round, total)

			if g.GetGameEndFlag() {
				break
			}
			require.Equal(t, RikkenPhaseRoundEnd, g.GetPhase(),
				"seed %d: ラウンドが終わっていない (phase=%d)", seed, g.GetPhase())
			require.NoError(t, g.NextRound())
		}
		require.True(t, g.GetGameEndFlag(), "seed %d: 規定ラウンドで終わらなかった", seed)
	}
}

// **人数比で割り切れる。** 宣言側は 1 人か 2 人、守備側は 3 人か 2 人なので、
// どちらの組み合わせでも整数になります（丸めが出ません）。
func TestRikkenScoreSplitHasNoRounding(t *testing.T) {
	t.Parallel()

	for _, contract := range []int{
		RikkenContractRik, RikkenContractMisere, RikkenContractSolo, RikkenContractOpenMisere,
	} {
		g := newRikkenAllCpu(t, 7)
		g.Reset()
		g.contract = contract
		g.declarerIdx = 0
		g.partnerIdx = -1
		if RikkenHasPartner(contract) {
			g.partnerIdx = 2
		}
		declarers, defenders := g.sideSizes()
		require.Positive(t, declarers)
		require.Positive(t, defenders)
		stake := g.perDefenderStake()
		assert.Zero(t, (stake*defenders)%declarers,
			"%s: %d * %d / %d に端数が出る", RikkenContractName(contract), stake, defenders, declarers)
	}
}

// **どの契約でもゼロサムが崩れない。** 成立/不成立の両方で確かめます。
func TestRikkenEveryContractIsZeroSum(t *testing.T) {
	t.Parallel()

	for _, contract := range []int{
		RikkenContractRik, RikkenContractMisere, RikkenContractSolo, RikkenContractOpenMisere,
	} {
		for _, made := range []bool{true, false} {
			g := newRikkenAllCpu(t, 9)
			g.Reset()
			for i := range RikkenPlayerCnt {
				g.players[i].SetScore(0)
			}
			g.contract = contract
			g.declarerIdx = 1
			g.partnerIdx = -1
			if RikkenHasPartner(contract) {
				g.partnerIdx = 3
			}
			g.declarerTricks = 9

			total := 0
			for i := range RikkenPlayerCnt {
				total += g.roundScoreFor(i, made)
			}
			assert.Zero(t, total, "%s made=%v: 合計が %d",
				RikkenContractName(contract), made, total)
		}
	}
}

// **CPU 同士で規定ラウンドまで進み切る。** 打ち切りに落ちた回数を数えます。
func TestRikkenAlwaysTerminates(t *testing.T) {
	const games = 40
	stalled := 0

	for seed := range games {
		g := newRikkenAllCpu(t, int64(seed)+1)
		g.Reset()

		steps := 0
		for !g.GetGameEndFlag() {
			steps++
			if steps > 500 {
				stalled++
				break
			}
			if g.GetPhase() == RikkenPhaseRoundEnd {
				require.NoError(t, g.NextRound())
				continue
			}
			before := g.GetTrickCount() + g.GetRoundNumber()
			g.CpuPlay()
			if before == g.GetTrickCount()+g.GetRoundNumber() && g.GetPhase() == RikkenPhasePlay {
				stalled++
				break
			}
		}
	}
	assert.Zero(t, stalled, "%d 局中 %d 局が進まなくなった", games, stalled)
}

// **13 トリックちょうどで 1 ラウンドが終わる。**
func TestRikkenPlaysThirteenTricks(t *testing.T) {
	for seed := range 20 {
		g := newRikkenAllCpu(t, int64(seed)+1)
		g.Reset()
		assert.Equal(t, RikkenTrickCnt, g.GetTrickCount(), "seed %d", seed)
		assert.LessOrEqual(t, g.GetDeclarerTricks(), RikkenTrickCnt)
		for i := range RikkenPlayerCnt {
			assert.Zero(t, g.GetPlayer(i).GetCardsSize(), "seed %d 席 %d: 手札が残った", seed, i)
		}
	}
}

// **Misere 系に切り札は無い。**
func TestRikkenMisereHasNoTrump(t *testing.T) {
	t.Parallel()

	for _, contract := range []int{RikkenContractMisere, RikkenContractOpenMisere} {
		assert.False(t, RikkenNeedsTrump(contract), "%s", RikkenContractName(contract))
		assert.True(t, RikkenIsMisere(contract))
		assert.False(t, RikkenHasPartner(contract))
		assert.Zero(t, RikkenContractTarget(contract), "目標は 0 トリック")
	}
	assert.True(t, RikkenNeedsTrump(RikkenContractRik))
	assert.True(t, RikkenNeedsTrump(RikkenContractSolo))
	assert.True(t, RikkenHasPartner(RikkenContractRik))
	assert.False(t, RikkenHasPartner(RikkenContractSolo))
	assert.Equal(t, RikkenRikTarget, RikkenContractTarget(RikkenContractRik))
	assert.Equal(t, RikkenSoloTarget, RikkenContractTarget(RikkenContractSolo))
}

// **相方は指名した札が出るまで伏せる。**
func TestRikkenPartnerStaysHiddenUntilTheCalledCardIsPlayed(t *testing.T) {
	t.Parallel()

	// **全席 CPU だと Reset() でラウンドが終わり切って手札が空になります。**
	// 指名する札は手札から探すので、配り立ての盤面（人間席で止まる）が要ります。
	g := newRikkenWithHuman(t, 11)
	require.Equal(t, RikkenHandSize, g.GetPlayer(0).GetCardsSize(), "配り立て")
	g.phase = RikkenPhaseCall
	g.contract = RikkenContractRik
	g.declarerIdx = 0
	g.currentTurn = 0
	require.NoError(t, g.call(0, CardDesignSpade))

	require.NotNil(t, g.GetCalledCard(), "Rik は札を指名する")
	assert.NotEqual(t, -1, g.partnerIdx, "内部では分かっている")
	assert.NotEqual(t, 0, g.partnerIdx, "自分は相方にならない")

	// **公開の可否がそのままアクセサに出る。** call() の直後に CPU が第1トリックを
	// 進めるので、指名札がその場で出て公開済みになることもあります。だから
	// 「常に -1」ではなく「公開フラグと一致する」ことを検査します。
	if g.partnerRevealed {
		assert.Equal(t, g.partnerIdx, g.GetPartnerIdx(), "公開後は席を返す")
	} else {
		assert.Equal(t, -1, g.GetPartnerIdx(), "公開前は伏せる")
	}

	// 伏せた状態を明示的に作って確かめる。
	g.partnerRevealed = false
	assert.Equal(t, -1, g.GetPartnerIdx(), "伏せているあいだは -1")
	g.partnerRevealed = true
	assert.Equal(t, g.partnerIdx, g.GetPartnerIdx(), "公開したら席を返す")
}

// **指名する札は自分が持っていないもの。**
func TestRikkenCalledCardIsNotInTheDeclarerHand(t *testing.T) {
	for seed := range 20 {
		g := newRikkenWithHuman(t, int64(seed)+1)
		require.Equal(t, RikkenHandSize, g.GetPlayer(1).GetCardsSize(), "seed %d: 配り立て", seed)
		g.phase = RikkenPhaseCall
		g.contract = RikkenContractRik
		g.declarerIdx = 1
		g.currentTurn = 1
		require.NoError(t, g.call(1, CardDesignHeart))

		card := g.GetCalledCard()
		require.NotNil(t, card, "seed %d", seed)
		assert.NotEqual(t, 1, g.partnerIdx, "seed %d: 自分を相方にしている", seed)
		assert.GreaterOrEqual(t, g.partnerIdx, 0, "seed %d: 誰も持っていない札を指名した", seed)
	}
}

// **フォローできるならフォローする。**
func TestRikkenMustFollowSuit(t *testing.T) {
	checked := 0
	for seed := range 10 {
		g := newRikkenWithHuman(t, int64(seed)+1)
		checked += rikkenDriveHuman(t, g, 400, func(g *Rikken) {
			if len(g.GetTrick()) == 0 {
				return
			}
			p := g.GetPlayer(0)
			leadSuit := g.GetTrick()[0].Card.GetDesign()
			has := false
			for k := range p.GetCardsSize() {
				if p.GetCard(k).GetDesign() == leadSuit {
					has = true
					break
				}
			}
			if !has {
				return
			}
			for _, v := range g.GetValidPlayIndices(0) {
				assert.Equal(t, leadSuit, p.GetCard(v).GetDesign(), "フォローを外せてしまう")
			}
		})
	}
	assert.Positive(t, checked, "人間の手番を一度も踏めなかった")
}

func TestRikkenConfigValidate(t *testing.T) {
	t.Parallel()

	assert.NoError(t, DefaultRikkenConfig().Validate())
	assert.Error(t, RikkenConfig{Rounds: RikkenRoundsMin - 1}.Validate())
	assert.Error(t, RikkenConfig{Rounds: RikkenRoundsMax + 1}.Validate())

	g := NewRikken(nil, nil, RikkenConfig{Rounds: 9999})
	assert.Equal(t, DefaultRikkenConfig(), g.GetConfig(), "壊れた設定は既定に落とす")
	assert.Equal(t, RikkenPlayerCnt, g.GetPlayerCnt())
	assert.True(t, g.GetPlayer(0).GetIsHuman())

	g.SetConfig(RikkenConfig{Rounds: 12})
	assert.Equal(t, 12, g.GetConfig().Rounds)
	g.SetConfig(RikkenConfig{Rounds: 1})
	assert.Equal(t, 12, g.GetConfig().Rounds, "範囲外は無視する")
}

func TestRikkenGiveUp(t *testing.T) {
	g := newRikkenWithHuman(t, 1)
	g.GiveUp()
	assert.True(t, g.GetGameEndFlag())
	assert.NotEqual(t, 0, g.GetWinnerIdx(), "投了した人は勝たない")
	assert.False(t, g.IsHumanTurn())
	assert.Nil(t, g.GetHint())

	g.GiveUp()
	assert.True(t, g.GetGameEndFlag())
}

func TestRikkenAccessorsRejectOutOfRange(t *testing.T) {
	t.Parallel()

	g := NewDefaultRikken()
	assert.Nil(t, g.GetPlayer(-1))
	assert.Nil(t, g.GetPlayer(99))
	assert.Nil(t, g.GetValidPlayIndices(-1))
	assert.Nil(t, g.GetValidPlayIndices(99))
	assert.False(t, g.HasPassed(-1))
	assert.False(t, g.HasPassed(99))
	assert.Zero(t, rikkenRank(nil))
	assert.Equal(t, 14, rikkenRank(NewCard(CardDesignSpade, 1, false)), "エースが最強")

	assert.Equal(t, "rik", RikkenContractName(RikkenContractRik))
	assert.Equal(t, "misere", RikkenContractName(RikkenContractMisere))
	assert.Equal(t, "solo", RikkenContractName(RikkenContractSolo))
	assert.Equal(t, "openMisere", RikkenContractName(RikkenContractOpenMisere))
	assert.Equal(t, "none", RikkenContractName(99))

	assert.Equal(t, "spade", rikkenSuitName(CardDesignSpade))
	assert.Equal(t, "clover", rikkenSuitName(CardDesignClover))
	assert.Equal(t, "heart", rikkenSuitName(CardDesignHeart))
	assert.Equal(t, "diamond", rikkenSuitName(CardDesignDiamond))
	assert.Equal(t, "notrump", rikkenSuitName(RikkenNoTrump))
}

func TestRikkenCallRejectsBadSuit(t *testing.T) {
	t.Parallel()

	g := newRikkenWithHuman(t, 13)
	g.phase = RikkenPhaseCall
	g.contract = RikkenContractRik
	g.declarerIdx = 0
	g.currentTurn = 0
	assert.Error(t, g.call(0, 99))
	assert.Error(t, g.call(0, 0), "スートは 1..4")
}

func TestRikkenHint(t *testing.T) {
	g := newRikkenWithHuman(t, 4)
	if !g.IsHumanTurn() {
		t.Skip("この配りでは人間の番が来ない")
	}
	h := g.GetHint()
	require.NotNil(t, h)
	assert.Equal(t, "rikkenBidStrength", h.Reason)
	require.NotNil(t, h.Contract)
	assert.Nil(t, h.CardIndex)
}

// **切り札のエースは後回しにするだけで、捨てない。**
//
// 非切り札のエース 3 枚を自分で持っていて切り札のエースだけ持っていない手は
// よくある形です。ここで切り札のエースを飛ばすと、他家が持っている札を無視して
// キングを呼んでしまいます——doc は「4 枚とも持っているならキング」と書いている
// ので、コードとコメントが食い違っていました。
func TestRikkenCallsTheTrumpAceWhenTheOthersAreInHand(t *testing.T) {
	t.Parallel()

	g := newRikkenWithHuman(t, 31)
	trump := CardDesignSpade
	g.trumpSuit = trump
	g.phase = RikkenPhaseCall
	g.contract = RikkenContractRik
	g.declarerIdx = 0
	g.currentTurn = 0

	// 宣言者に「切り札以外のエース 3 枚」を持たせ、切り札のエースは他家に置く。
	for i := range RikkenPlayerCnt {
		g.players[i].Reset()
	}
	declarer := g.players[0]
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		if suit == trump {
			continue
		}
		declarer.AddCard(NewCard(suit, 1, false))
	}
	declarer.AddCard(NewCard(CardDesignHeart, 5, false))
	g.players[2].AddCard(NewCard(trump, 1, false)) // 切り札のエースは席 2
	// **キングも他家に置きます。** これが無いと、切り札のエースを飛ばす壊れた実装でも
	// 最後の総当たりフォールバックが同じ札を返してしまい、テストが理由を取り違えて
	// 通ります（変異テストで実際に素通りしました）。
	g.players[1].AddCard(NewCard(CardDesignHeart, 13, false))

	card := g.chooseCalledCard(0)
	require.NotNil(t, card)
	assert.Equal(t, trump, card.GetDesign(), "切り札のエースを呼ぶべき")
	assert.Equal(t, 1, card.GetValue(), "キングに落ちてはいけない")
	assert.Equal(t, 2, g.holderOf(card), "他家が持っている札を呼ぶ")
}

// **4 枚とも持っているときだけキングに落ちる。**
func TestRikkenFallsBackToAKingOnlyWithAllFourAces(t *testing.T) {
	t.Parallel()

	g := newRikkenWithHuman(t, 33)
	g.trumpSuit = CardDesignSpade
	g.phase = RikkenPhaseCall
	g.contract = RikkenContractRik
	g.declarerIdx = 0

	for i := range RikkenPlayerCnt {
		g.players[i].Reset()
	}
	declarer := g.players[0]
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		declarer.AddCard(NewCard(suit, 1, false))
	}
	g.players[1].AddCard(NewCard(CardDesignHeart, 13, false))

	card := g.chooseCalledCard(0)
	require.NotNil(t, card)
	assert.Equal(t, 13, card.GetValue(), "エースを全部持っていればキング")
	assert.Equal(t, 1, g.holderOf(card))
}

// **Rik の相方も「宣言側」の印で漏れてはいけない。**
//
// ColourWhist と同じ形のバグで、こちらは先にマージされていました。
func TestRikkenPartnerIsNotLeakedByDeclarerSide(t *testing.T) {
	t.Parallel()

	g := newRikkenWithHuman(t, 51)
	g.phase = RikkenPhaseCall
	g.contract = RikkenContractRik
	g.declarerIdx = 0
	g.currentTurn = 0
	g.partnerRevealed = false
	require.NoError(t, g.call(0, CardDesignSpade))

	partner := g.partnerIdx
	require.GreaterOrEqual(t, partner, 0)
	require.NotEqual(t, 0, partner)

	// **必ず伏せた状態にしてから検査します。** 条件付きだとテストが何も検査せずに
	// 通ることがあります。
	g.partnerRevealed = false
	assert.Equal(t, -1, g.GetPartnerIdx(), "相方の席が漏れている")
	assert.False(t, g.IsDeclarerSideVisible(partner), "宣言側の印で相方が漏れている")
	// **内部の真値は伏せない。**
	assert.True(t, g.IsDeclarerSide(partner), "内部では宣言側のまま")
	assert.True(t, g.IsDeclarerSideVisible(0))

	g.partnerRevealed = true
	assert.True(t, g.IsDeclarerSideVisible(partner))
}

// **単独契約では宣言者だけが宣言側。**
func TestRikkenSoloContractHasNoVisiblePartner(t *testing.T) {
	t.Parallel()

	g := newRikkenAllCpu(t, 53)
	g.contract = RikkenContractSolo
	g.declarerIdx = 1
	g.partnerIdx = -1
	g.partnerRevealed = false

	assert.True(t, g.IsDeclarerSideVisible(1))
	for _, i := range []int{0, 2, 3} {
		assert.False(t, g.IsDeclarerSideVisible(i), "席 %d", i)
	}
}
