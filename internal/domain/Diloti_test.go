//go:build test

package domain

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newDilotiGame(t *testing.T) *Diloti {
	t.Helper()
	d := NewDefaultDiloti()
	d.Reset()
	return d
}

// **開幕は人間の手番。** 非親が先に打つ規則なので、親を席 1 にしてある ──
// 親を 0 にすると人間は最初の 4 枚に一度も手を出せない。
func TestDiloti_ResetDealsAndStartsWithTheHuman(t *testing.T) {
	d := newDilotiGame(t)
	assert.Equal(t, DilotiPhasePlay, d.GetPhase())
	assert.Equal(t, 1, d.GetRoundNumber())
	assert.Equal(t, 0, d.GetCurrentPlayerIdx(), "人間が先に打てない")
	assert.True(t, d.IsHumanTurn())
	assert.Equal(t, DilotiPlayerCnt-1, d.GetDealerIdx())
	for i := 0; i < DilotiPlayerCnt; i++ {
		assert.Equal(t, DilotiHandSize, d.GetPlayer(i).GetCardsSize(), "席 %d の手札", i)
	}
	assert.Len(t, d.GetTable(), DilotiTableSize)
	// 52 - 6*2 - 4 = 36
	assert.Equal(t, DilotiDeckSize-DilotiHandSize*DilotiPlayerCnt-DilotiTableSize, d.GetDeckRemaining())
}

// **配り直しでは場札を足さない。** 足すと 52 枚が場に溢れ、山が尽きる前に
// 取りきれない札が積み上がる。
func TestDiloti_RedealAddsNoTableCards(t *testing.T) {
	d := newDilotiGame(t)
	d.table = dcards([2]int{CardDesignSpade, 2})
	before := len(d.table)
	deckBefore := d.GetDeckRemaining()
	// 両者の手札を 1 枚ずつにして、2 手で配り直しに入らせる。
	// **絵札にしない。** 先に置いた絵札と同ランクだと 2 手目が塞がれる。
	for i := 0; i < DilotiPlayerCnt; i++ {
		p := d.GetPlayer(i)
		p.Reset()
		p.AddCard(dc(CardDesignHeart, 8+i))
	}
	require.NoError(t, d.applyPlay(0, 0, DilotiActionTrail, nil, nil, 0))
	require.NoError(t, d.applyPlay(1, 0, DilotiActionTrail, nil, nil, 0))

	assert.Equal(t, DilotiHandSize, d.GetPlayer(0).GetCardsSize(), "配り直しが起きていない")
	assert.Equal(t, deckBefore-DilotiHandSize*DilotiPlayerCnt, d.GetDeckRemaining())
	assert.Len(t, d.GetTable(), before+DilotiPlayerCnt, "配り直しで場札が足されている")
}

// **クセリは 1 枚で場を払ったとき。** 局の初手は数えない。
func TestDiloti_XeriRequiresClearingTheTable(t *testing.T) {
	d := newDilotiGame(t)
	d.table = dcards([2]int{CardDesignSpade, 5})
	d.decls = nil
	d.GetPlayer(0).Reset()
	d.GetPlayer(0).AddCard(dc(CardDesignHeart, 5))

	// 1 手目はクセリにならない。
	d.firstPlayDone = false
	require.NoError(t, d.applyPlay(0, 0, DilotiActionCapture, []int{0}, nil, 0))
	assert.Equal(t, 0, d.GetPlayer(0).GetXeri(), "局の初手がクセリに数えられている")
	assert.Empty(t, d.GetTable())

	// 2 手目以降は数える。
	d.table = dcards([2]int{CardDesignClover, 5})
	d.GetPlayer(0).Reset()
	d.GetPlayer(0).AddCard(dc(CardDesignDiamond, 5))
	require.NoError(t, d.applyPlay(0, 0, DilotiActionCapture, []int{0}, nil, 0))
	assert.Equal(t, 1, d.GetPlayer(0).GetXeri(), "場を払ってもクセリにならない")
}

// **場を払い切らなければクセリではない。**
func TestDiloti_PartialCaptureIsNotXeri(t *testing.T) {
	d := newDilotiGame(t)
	d.firstPlayDone = true
	d.table = dcards([2]int{CardDesignSpade, 5}, [2]int{CardDesignHeart, 9})
	d.decls = nil
	d.GetPlayer(0).Reset()
	d.GetPlayer(0).AddCard(dc(CardDesignClover, 5))
	require.NoError(t, d.applyPlay(0, 0, DilotiActionCapture, []int{0}, nil, 0))
	assert.Equal(t, 0, d.GetPlayer(0).GetXeri())
	assert.Len(t, d.GetTable(), 1)
}

// **宣言は場に残り、値ちょうどの札で取れる。**
func TestDiloti_DeclareThenCapture(t *testing.T) {
	d := newDilotiGame(t)
	d.firstPlayDone = true
	d.table = dcards([2]int{CardDesignSpade, 4})
	d.decls = nil
	d.GetPlayer(0).Reset()
	d.GetPlayer(0).AddCard(dc(CardDesignHeart, 3)) // 3 + 4 = 7
	d.GetPlayer(0).AddCard(dc(CardDesignClover, 7))

	require.NoError(t, d.applyPlay(0, 0, DilotiActionDeclare, []int{0}, nil, 7))
	require.Len(t, d.GetDeclarations(), 1)
	assert.Equal(t, 7, d.GetDeclarations()[0].Value)
	assert.Equal(t, 0, d.GetDeclarations()[0].OwnerIdx)
	assert.Empty(t, d.GetTable(), "宣言に使った場札が残っている")

	// **宣言を抱えている間は場に置けない。**
	assert.False(t, d.CanTrail(0, 0), "宣言を抱えたまま置けてしまう")
	err := d.applyPlay(0, 0, DilotiActionTrail, nil, nil, 0)
	require.Error(t, err)

	require.NoError(t, d.applyPlay(0, 0, DilotiActionCapture, nil, []int{0}, 0))
	assert.Empty(t, d.GetDeclarations(), "宣言が取れていない")
	assert.Len(t, d.GetPlayer(0).GetCaptured(), 3, "宣言の 2 枚と出した 1 枚")
}

// **裏付けの札を持たない宣言はできない。**
func TestDiloti_DeclarationNeedsABackingCard(t *testing.T) {
	d := newDilotiGame(t)
	d.table = dcards([2]int{CardDesignSpade, 4})
	d.decls = nil
	d.GetPlayer(0).Reset()
	d.GetPlayer(0).AddCard(dc(CardDesignHeart, 3))
	d.GetPlayer(0).AddCard(dc(CardDesignClover, 2)) // 7 は持っていない
	assert.Error(t, d.applyPlay(0, 0, DilotiActionDeclare, []int{0}, nil, 7))
	assert.Empty(t, d.GetDeclarations())
}

// **単一宣言は上げられ、グループ宣言は上げられない。**
func TestDiloti_RaiseAndGroup(t *testing.T) {
	d := newDilotiGame(t)
	d.decls = []*DilotiDeclaration{NewDilotiDeclaration(1, 6, dcards([2]int{CardDesignSpade, 6}))}
	d.table = dcards([2]int{CardDesignHeart, 2})
	d.GetPlayer(0).Reset()
	d.GetPlayer(0).AddCard(dc(CardDesignClover, 2)) // 2 + 場の 2 + 宣言 6 = 10
	d.GetPlayer(0).AddCard(dc(CardDesignDiamond, 10))

	require.NoError(t, d.applyPlay(0, 0, DilotiActionDeclare, []int{0}, []int{0}, 10))
	require.Len(t, d.GetDeclarations(), 1)
	assert.Equal(t, 10, d.GetDeclarations()[0].Value, "上げた値が入っていない")
	assert.Equal(t, 0, d.GetDeclarations()[0].OwnerIdx, "上げた側がオーナーになっていない")
	assert.False(t, d.GetDeclarations()[0].IsGroup)

	// グループ宣言にすると上げられなくなる。
	d.decls[0].AddGroup(dcards([2]int{CardDesignSpade, 10}))
	require.True(t, d.decls[0].IsGroup)
	d.GetPlayer(1).Reset()
	d.GetPlayer(1).AddCard(dc(CardDesignHeart, 1))
	d.GetPlayer(1).AddCard(dc(CardDesignClover, 11))
	assert.Error(t, d.applyPlay(1, 0, DilotiActionDeclare, nil, []int{0}, 11),
		"グループ宣言が上げられてしまっている")
}

// **山が尽きたら取り残しは最後に取った側へ。** ただしクセリではない。
func TestDiloti_LeftoverGoesToTheLastCapturerAndIsNotXeri(t *testing.T) {
	d := newDilotiGame(t)
	d.firstPlayDone = true
	d.drawIdx = len(d.deck)
	d.table = dcards([2]int{CardDesignSpade, 9}, [2]int{CardDesignHeart, 4})
	d.decls = nil
	d.lastCapturer = 1
	for i := 0; i < DilotiPlayerCnt; i++ {
		d.GetPlayer(i).Reset()
	}
	before := len(d.GetPlayer(1).GetCaptured())
	xeriBefore := d.GetPlayer(1).GetXeri()

	d.finishRound()
	assert.Equal(t, before+2, len(d.GetPlayer(1).GetCaptured()), "取り残しが渡っていない")
	assert.Equal(t, xeriBefore, d.GetPlayer(1).GetXeri(), "取り残しの回収がクセリに数えられている")
	assert.Empty(t, d.GetTable())
	assert.Equal(t, DilotiPhaseRoundEnd, d.GetPhase())
}

// **集計は 5 項目。** 最多枚数・A・10♦・2♣・クセリ。
func TestDiloti_ScoreRound(t *testing.T) {
	d := newDilotiGame(t)
	for i := 0; i < DilotiPlayerCnt; i++ {
		d.GetPlayer(i).ResetRound()
	}
	// 席 0: 30 枚 (A×3、10♦、2♣ を含む)、クセリ 1 回。
	p0 := []*Card{dc(CardDesignSpade, 1), dc(CardDesignHeart, 1), dc(CardDesignClover, 1),
		dc(CardDesignDiamond, 10), dc(CardDesignClover, 2)}
	for len(p0) < 30 {
		p0 = append(p0, dc(CardDesignSpade, 8))
	}
	d.GetPlayer(0).AddCaptured(p0)
	d.GetPlayer(0).AddXeri()
	// 席 1: 22 枚 (A×1)。
	p1 := []*Card{dc(CardDesignDiamond, 1)}
	for len(p1) < 22 {
		p1 = append(p1, dc(CardDesignHeart, 8))
	}
	d.GetPlayer(1).AddCaptured(p1)

	res := d.scoreRound()
	require.Len(t, res.Lines, 5)
	byKey := map[string][]int{}
	for _, l := range res.Lines {
		byKey[l.Key] = l.Points
	}
	assert.Equal(t, []int{4, 0}, byKey[DilotiScoreCards], "最多枚数")
	assert.Equal(t, []int{3, 1}, byKey[DilotiScoreAces], "A は 1 枚 1 点")
	assert.Equal(t, []int{2, 0}, byKey[DilotiScoreTenOfDiamonds], "10♦ は 2 点")
	assert.Equal(t, []int{1, 0}, byKey[DilotiScoreTwoOfClubs], "2♣ は 1 点")
	assert.Equal(t, []int{10, 0}, byKey[DilotiScoreXeri], "クセリは 1 回 10 点")
	assert.Equal(t, 4+3+2+1+10, res.Totals[0])
	assert.Equal(t, 1, res.Totals[1])
	// 固定点は 11 点で、残りはすべてクセリ。
	assert.Equal(t, 11, 4+4*DilotiAcePoints+DilotiTenOfDiamondsPoints+DilotiTwoOfClubsPoints-4+4)
}

// **26 対 26 では最多枚数はどちらにも入らない。**
func TestDiloti_TiedCardCountScoresForNobody(t *testing.T) {
	d := newDilotiGame(t)
	for i := 0; i < DilotiPlayerCnt; i++ {
		d.GetPlayer(i).ResetRound()
		cards := make([]*Card, 0, DilotiHalfDeck)
		for len(cards) < DilotiHalfDeck {
			cards = append(cards, dc(CardDesignSpade, 8))
		}
		d.GetPlayer(i).AddCaptured(cards)
	}
	res := d.scoreRound()
	for _, l := range res.Lines {
		if l.Key == DilotiScoreCards {
			assert.Equal(t, []int{0, 0}, l.Points, "引き分けなのに最多枚数が入っている")
		}
	}
}

// **同点では終局しない。**
func TestDiloti_ATieDoesNotEndTheMatch(t *testing.T) {
	d := newDilotiGame(t)
	for i := 0; i < DilotiPlayerCnt; i++ {
		d.GetPlayer(i).ResetScore()
		d.GetPlayer(i).AddScore(d.GetConfig().TargetScore)
	}
	d.checkGameEnd()
	assert.False(t, d.GetGameEndFlag(), "同点で終局している")

	d.GetPlayer(0).AddScore(1)
	d.checkGameEnd()
	assert.True(t, d.GetGameEndFlag())
	assert.Equal(t, 0, d.GetWinnerIdx())
}

// **1 局を最後まで打てて、札は 52 枚のまま。**
func TestDiloti_PlaysARoundThrough(t *testing.T) {
	d := newDilotiGame(t)
	total := func() int {
		n := d.GetDeckRemaining() + len(d.GetTable())
		for _, x := range d.GetDeclarations() {
			n += len(x.AllCards())
		}
		for i := 0; i < d.GetPlayerCnt(); i++ {
			n += d.GetPlayer(i).GetCardsSize() + len(d.GetPlayer(i).GetCaptured())
		}
		return n
	}
	require.Equal(t, DilotiDeckSize, total())
	for step := 0; step < 400 && d.GetPhase() == DilotiPhasePlay; step++ {
		require.Equal(t, DilotiDeckSize, total(), "手 %d で札が合わない", step)
		if d.IsHumanTurn() {
			h := d.GetHint()
			require.GreaterOrEqual(t, h.Move.HandIdx, 0, "手番なのに打てる手が無い")
			require.NoError(t, d.applyPlay(0, h.Move.HandIdx, h.Move.Action,
				h.Move.TableIdxs, h.Move.DeclIdxs, h.Move.Value))
			continue
		}
		d.CpuPlay()
	}
	// **局の終わりは 2 通りある。** 得点が目標に届けば `finishRound` は
	// そのまま終局へ進むので、局が正しく閉じても phase は GameEnd になる。
	// 1 局の得点は最大 101 点、目標の上限も 101 点なので、**目標を上げても
	// 起き得る** —— 「必ず RoundEnd」という前提はそもそも立たない。
	// このテストが見たいのは「1 局を打ち切れること」と「札が保存されること」
	// なので、終わり方の区別はしない (#6249)。
	require.Contains(t, []string{DilotiPhaseRoundEnd, DilotiPhaseGameEnd}, d.GetPhase(),
		"局が終わらない (phase が play のまま)")
	assert.Equal(t, DilotiDeckSize, total(), "局の終わりで札が合わない")

	res := d.GetLastResult()
	require.NotNil(t, res)
	assert.Equal(t, DilotiDeckSize, res.CardCounts[0]+res.CardCounts[1],
		"取り札の合計が 52 枚でない")
}

// **試合は終局まで届く。**
func TestDiloti_ReachesTheTarget(t *testing.T) {
	d := NewDefaultDiloti()
	cfg := DefaultDilotiConfig()
	cfg.TargetScore = DilotiMinTarget
	cfg.CpuDifficulty = DilotiCpuDifficultyEasy
	d.SetConfig(cfg)
	d.Reset()
	for round := 0; round < 100 && !d.GetGameEndFlag(); round++ {
		for step := 0; step < 400 && d.GetPhase() == DilotiPhasePlay; step++ {
			if d.IsHumanTurn() {
				h := d.GetHint()
				require.GreaterOrEqual(t, h.Move.HandIdx, 0)
				require.NoError(t, d.applyPlay(0, h.Move.HandIdx, h.Move.Action,
					h.Move.TableIdxs, h.Move.DeclIdxs, h.Move.Value))
				continue
			}
			d.CpuPlay()
		}
		require.NotEqual(t, DilotiPhasePlay, d.GetPhase(), "局が終わらない")
		d.NextRound()
	}
	require.True(t, d.GetGameEndFlag(), "21 点勝負でも終局に届かない")
	assert.GreaterOrEqual(t, d.GetWinnerIdx(), 0)
}

// **ヒントは CPU の難易度で鈍らない。**
func TestDiloti_HintIgnoresCpuDifficulty(t *testing.T) {
	want := ""
	for _, diff := range []DilotiCpuDifficulty{
		DilotiCpuDifficultyEasy, DilotiCpuDifficultyNormal, DilotiCpuDifficultyHard,
	} {
		d := NewDefaultDiloti()
		cfg := DefaultDilotiConfig()
		cfg.CpuDifficulty = diff
		d.SetConfig(cfg)
		d.Reset()
		d.firstPlayDone = true
		d.table = dcards([2]int{CardDesignSpade, 5}, [2]int{CardDesignHeart, 9})
		d.decls = nil
		d.GetPlayer(0).Reset()
		d.GetPlayer(0).AddCard(dc(CardDesignClover, 13))
		d.GetPlayer(0).AddCard(dc(CardDesignDiamond, 5))

		// 20 回引いても同じ手を勧める (乱数が混ざっていないこと)。
		for i := 0; i < 20; i++ {
			h := d.GetHint()
			got := h.Move.Action + fmt.Sprint(h.Move.HandIdx, h.Move.TableIdxs)
			if want == "" {
				want = got
			}
			assert.Equal(t, want, got, "難易度 %d でヒントが変わった", diff)
		}
	}
}

func TestDiloti_RejectsBadInput(t *testing.T) {
	d := newDilotiGame(t)
	assert.Error(t, d.PlayerPlay(-1, DilotiActionTrail, nil, nil, 0))
	assert.Error(t, d.PlayerPlay(99, DilotiActionTrail, nil, nil, 0))
	assert.Error(t, d.PlayerPlay(0, "zzz", nil, nil, 0))
	assert.Error(t, d.PlayerPlay(0, DilotiActionCapture, []int{99}, nil, 0))

	d.phase = DilotiPhaseRoundEnd
	assert.Error(t, d.PlayerPlay(0, DilotiActionTrail, nil, nil, 0))
	d.phase = DilotiPhasePlay
	d.gameEndFlag = true
	assert.Error(t, d.PlayerPlay(0, DilotiActionTrail, nil, nil, 0))
	d.gameEndFlag = false
	d.currentIdx = 1
	assert.Error(t, d.PlayerPlay(0, DilotiActionTrail, nil, nil, 0))
}

// **保存した盤で打ち続けられる。**
func TestDiloti_SaveRestoreKeepsPlaying(t *testing.T) {
	d := newDilotiGame(t)
	h := d.GetHint()
	require.GreaterOrEqual(t, h.Move.HandIdx, 0)
	require.NoError(t, d.applyPlay(0, h.Move.HandIdx, h.Move.Action,
		h.Move.TableIdxs, h.Move.DeclIdxs, h.Move.Value))

	data, err := json.Marshal(d)
	require.NoError(t, err)
	require.Greater(t, len(data), 2, "空の JSON になっている")

	var r Diloti
	require.NoError(t, json.Unmarshal(data, &r))
	assert.Equal(t, d.GetPhase(), r.GetPhase())
	assert.Equal(t, d.GetDeckRemaining(), r.GetDeckRemaining(), "山の位置が消えている")
	assert.Equal(t, d.GetLastCapturer(), r.GetLastCapturer())
	assert.Equal(t, len(d.GetTable()), len(r.GetTable()))
	assert.Equal(t, len(d.GetDeclarations()), len(r.GetDeclarations()))
	assert.Equal(t, d.firstPlayDone, r.firstPlayDone, "初手の済み印が消えている")
	for i := 0; i < DilotiPlayerCnt; i++ {
		assert.Equal(t, d.GetPlayer(i).GetCardsSize(), r.GetPlayer(i).GetCardsSize())
		assert.Equal(t, len(d.GetPlayer(i).GetCaptured()), len(r.GetPlayer(i).GetCaptured()),
			"席 %d の取り札が消えている", i)
		assert.Equal(t, d.GetPlayer(i).GetXeri(), r.GetPlayer(i).GetXeri())
	}

	for step := 0; step < 400 && r.GetPhase() == DilotiPhasePlay; step++ {
		if r.IsHumanTurn() {
			hh := r.GetHint()
			require.GreaterOrEqual(t, hh.Move.HandIdx, 0)
			require.NoError(t, r.applyPlay(0, hh.Move.HandIdx, hh.Move.Action,
				hh.Move.TableIdxs, hh.Move.DeclIdxs, hh.Move.Value))
			continue
		}
		r.CpuPlay()
	}
	// 同上: 得点が目標に届けば復元した盤でもそのまま終局する (#6249)。
	assert.Contains(t, []string{DilotiPhaseRoundEnd, DilotiPhaseGameEnd}, r.GetPhase(),
		"復元した盤で局が終わらない (phase が play のまま)")
}

func TestDilotiConfig_Validate(t *testing.T) {
	assert.NoError(t, DefaultDilotiConfig().Validate())
	bad := DefaultDilotiConfig()
	bad.TargetScore = 1
	assert.Error(t, bad.Validate())
	bad.TargetScore = 999
	assert.Error(t, bad.Validate())
	bad = DefaultDilotiConfig()
	bad.CpuDifficulty = 9
	assert.Error(t, bad.Validate())
	for _, v := range DilotiTargetOptions {
		cfg := DefaultDilotiConfig()
		cfg.TargetScore = v
		assert.NoError(t, cfg.Validate(), "選べる目標点 %d が弾かれる", v)
	}
}

// **宣言の裏付け札は取り切るまで手放せない。** 手放せてしまうと、宣言を
// 抱えたまま取る手も置く手も無い ── 手番が回ってきても打てない盤面になる。
func TestDiloti_BackingCardCannotBeSpentElsewhere(t *testing.T) {
	d := newDilotiGame(t)
	d.firstPlayDone = true
	d.decls = []*DilotiDeclaration{NewDilotiDeclaration(0, 7, dcards([2]int{CardDesignSpade, 7}))}
	d.table = dcards([2]int{CardDesignHeart, 7})
	d.GetPlayer(0).Reset()
	d.GetPlayer(0).AddCard(dc(CardDesignClover, 7)) // 唯一の裏付け

	// 場の 7 を取る手には使えない (宣言を取っていないので)。
	assert.Error(t, d.applyPlay(0, 0, DilotiActionCapture, []int{0}, nil, 0),
		"裏付けの札を宣言以外に使えてしまっている")
	// 宣言ごと取るなら通る。
	require.NoError(t, d.applyPlay(0, 0, DilotiActionCapture, []int{0}, []int{0}, 0))
	assert.Empty(t, d.GetDeclarations())

	// 同じ値がもう 1 枚あれば手放してよい。
	d.decls = []*DilotiDeclaration{NewDilotiDeclaration(0, 7, dcards([2]int{CardDesignSpade, 7}))}
	d.table = dcards([2]int{CardDesignHeart, 7})
	d.GetPlayer(0).Reset()
	d.GetPlayer(0).AddCard(dc(CardDesignClover, 7))
	d.GetPlayer(0).AddCard(dc(CardDesignDiamond, 7))
	assert.NoError(t, d.applyPlay(0, 0, DilotiActionCapture, []int{0}, nil, 0),
		"予備の 7 があるのに弾かれている")
}

// **手番が回ってきて打てる手が無い盤面を作らない。**
func TestDiloti_AlwaysHasALegalMove(t *testing.T) {
	d := newDilotiGame(t)
	d.firstPlayDone = true
	// 絵札しか無く、場に同ランクの絵札が並ぶ ── 置くことも取ることも塞がれた形。
	d.table = dcards([2]int{CardDesignSpade, 11}, [2]int{CardDesignHeart, 12},
		[2]int{CardDesignClover, 13})
	d.decls = []*DilotiDeclaration{NewDilotiDeclaration(1, 5, dcards([2]int{CardDesignSpade, 5}))}
	d.GetPlayer(0).Reset()
	d.GetPlayer(0).AddCard(dc(CardDesignDiamond, 11))
	require.NotEmpty(t, d.dilotiLegalMoves(0), "打てる手が 1 つも無い")

	// 逃げ道が空回りしていないこと: 取れる盤面では取る手が挙がる。
	moves := d.dilotiLegalMoves(0)
	found := false
	for _, m := range moves {
		if m.Action == DilotiActionCapture {
			found = true
		}
	}
	assert.True(t, found, "同ランクの J を取る手が挙がっていない")
}

// **宣言も保存されなければならない。** DilotiDeclaration の codec が抜けると
// 復元した盤から束が消え、値ちょうどで取るはずの札が宙に浮く ── しかも
// 枚数だけ見る検査は「宣言が 0 個で一致」で通ってしまう。
func TestDilotiDeclaration_JSONRoundTrip(t *testing.T) {
	decl := NewDilotiDeclaration(1, 8, dcards([2]int{CardDesignSpade, 3}, [2]int{CardDesignHeart, 5}))
	decl.AddGroup(dcards([2]int{CardDesignClover, 8}))
	require.True(t, decl.IsGroup)

	data, err := json.Marshal(decl)
	require.NoError(t, err)
	require.Greater(t, len(data), 2, "空の JSON になっている")

	var back DilotiDeclaration
	require.NoError(t, json.Unmarshal(data, &back))
	assert.Equal(t, 1, back.OwnerIdx, "オーナーが消えている")
	assert.Equal(t, 8, back.Value, "宣言値が消えている")
	assert.True(t, back.IsGroup, "グループ宣言が単一に戻っている")
	require.Len(t, back.Groups, 2, "束の区切りが失われている")
	assert.Len(t, back.Groups[0], 2)
	assert.Len(t, back.Groups[1], 1)
	require.Len(t, back.AllCards(), 3)
	assert.Equal(t, CardDesignSpade, back.AllCards()[0].GetDesign())
	assert.Equal(t, 3, back.AllCards()[0].GetValue())

	assert.Error(t, json.Unmarshal([]byte("{"), &back))
}

// **宣言を抱えた盤も保存できる。** 上の codec が正しくても、集約が宣言を
// 書き出していなければ意味がない。
func TestDiloti_SaveRestoreCarriesDeclarations(t *testing.T) {
	d := newDilotiGame(t)
	d.firstPlayDone = true
	decl := NewDilotiDeclaration(0, 7, dcards([2]int{CardDesignSpade, 7}))
	decl.AddGroup(dcards([2]int{CardDesignHeart, 3}, [2]int{CardDesignClover, 4}))
	d.SetTableForTest(dcards([2]int{CardDesignDiamond, 2}), []*DilotiDeclaration{decl})

	data, err := json.Marshal(d)
	require.NoError(t, err)
	var r Diloti
	require.NoError(t, json.Unmarshal(data, &r))

	require.Len(t, r.GetDeclarations(), 1, "宣言が消えている")
	got := r.GetDeclarations()[0]
	assert.Equal(t, 7, got.Value)
	assert.Equal(t, 0, got.OwnerIdx)
	assert.True(t, got.IsGroup, "グループ宣言が単一に戻っている")
	assert.Len(t, got.AllCards(), 3)
	assert.Len(t, r.GetTable(), 1)
}
