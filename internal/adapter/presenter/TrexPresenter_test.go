//go:build test

package presenter

import (
	"encoding/json"
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func txTestGame(t *testing.T) *domain.Trex {
	t.Helper()
	tr := domain.NewDefaultTrex()
	tr.Reset()
	return tr
}

func txDecode(t *testing.T, raw string) map[string]any {
	t.Helper()
	var out map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &out))
	return out
}

// txStub wires a MockTrexGame with every accessor the presenters touch, so a
// test can pin an exact phase/contract rather than shuffling until one appears.
func txStub(phase domain.TrexPhase, contract domain.TrexContract, king int, gameEnd bool, winner int) *interfaces.MockTrexGame {
	g := new(interfaces.MockTrexGame)
	g.On("GetPhase").Return(phase)
	g.On("GetContract").Return(contract)
	g.On("IsTrix").Return(contract == domain.TrexContractTrix)
	g.On("GetGameEndFlag").Return(gameEnd)
	g.On("GetWinnerIdx").Return(winner)
	g.On("GetCurrentPlayerIdx").Return(0)
	g.On("GetKingIdx").Return(king)
	g.On("GetDealNumber").Return(3)
	g.On("GetTrickNumber").Return(0)
	g.On("AvailableContracts").Return([]domain.TrexContract{domain.TrexContractQueens, domain.TrexContractTrix})
	g.On("GetTrick").Return([]domain.TrexTrickCard{})
	g.On("GetSuitRun", mock.Anything).Return(true, 11, 13)
	g.On("GetFinishOrder").Return([]int{})
	g.On("GetValidPlayIndices", mock.Anything).Return([]int{0})
	g.On("GetScore", mock.Anything).Return(-40)
	g.On("GetDealScore", mock.Anything).Return(-10)
	g.On("GetTricksWon", mock.Anything).Return(1)
	g.On("GetConfig").Return(domain.DefaultTrexConfig())
	players := make([]*domain.TrexPlayer, 0, domain.TrexPlayerCnt)
	players = append(players, domain.NewTrexPlayer(true))
	for range domain.TrexPlayerCnt - 1 {
		players = append(players, domain.NewTrexPlayer(false))
	}
	g.On("GetPlayers").Return(players)
	g.On("GetPlayer", mock.Anything).Return(domain.NewTrexPlayer(false))
	g.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	g.On("TrexCpuDecide", mock.Anything).Return(domain.TrexCpuAction{Contract: domain.TrexContractQueens, HandIdx: 0})
	return g
}

func TestTrexWebPresenter_HidesTheCpuHandButNeverTheScores(t *testing.T) {
	// 個人戦で 20 ディール戦うので、誰がどれだけ沈んでいるかは契約選択の
	// 判断材料そのもの。
	out := txDecode(t, new(TrexWebPresenter).Output(txTestGame(t), nil))
	players, ok := out["players"].([]any)
	require.True(t, ok)
	require.Len(t, players, domain.TrexPlayerCnt)

	human, _ := players[0].(map[string]any)
	assert.False(t, human["hidden"].(bool))
	assert.NotEmpty(t, human["cards"])

	cpu, _ := players[1].(map[string]any)
	assert.True(t, cpu["hidden"].(bool))
	assert.Empty(t, cpu["cards"], "the opponent's hand must not reach the browser")
	assert.Positive(t, cpu["cardCount"], "but its size is public")
	assert.NotNil(t, cpu["score"], "and so is its score")
}

func TestTrexWebPresenter_ShipsWhatIsLeftToChoose(t *testing.T) {
	// 1 王国に 1 度ずつしか選べないので、消化済みを引いた残りを送らないと、
	// クライアントが選べない契約を出してしまう。
	g := txStub(domain.TrexPhaseChoose, domain.TrexContractNone, 0, false, -1)
	out := txDecode(t, new(TrexWebPresenter).Output(g, nil))
	assert.Equal(t, []any{float64(domain.TrexContractQueens), float64(domain.TrexContractTrix)}, out["availableContracts"])
	assert.Equal(t, float64(domain.TrexTotalDeals), out["totalDeals"])
}

func TestTrexWebPresenter_ShipsTheFourDominoRuns(t *testing.T) {
	// J=11 起点で上下に伸びる範囲をそのまま送る。クライアントが端を数え直さない。
	g := txStub(domain.TrexPhasePlay, domain.TrexContractTrix, 0, false, -1)
	out := txDecode(t, new(TrexWebPresenter).Output(g, nil))
	runs, ok := out["runs"].([]any)
	require.True(t, ok)
	assert.Len(t, runs, 4, "one run per suit")
	first, _ := runs[0].(map[string]any)
	assert.Equal(t, float64(domain.CardDesignSpade), first["suit"], "suits are 1-based, not 0-based")
	assert.Equal(t, float64(11), first["low"])
	assert.True(t, out["isTrix"].(bool))
}

func TestTrexWebPresenter_PassOnlyExistsInTheDominoes(t *testing.T) {
	// トリック契約にパスは存在しない。そこで押せると規則が壊れて見える。
	trix := new(interfaces.MockTrexGame)
	for _, m := range []string{"GetGameEndFlag"} {
		trix.On(m).Return(false)
	}
	trix.On("GetPhase").Return(domain.TrexPhasePlay)
	trix.On("GetContract").Return(domain.TrexContractTrix)
	trix.On("IsTrix").Return(true)
	trix.On("GetCurrentPlayerIdx").Return(0)
	trix.On("GetKingIdx").Return(0)
	trix.On("GetDealNumber").Return(0)
	trix.On("GetTrickNumber").Return(0)
	trix.On("GetWinnerIdx").Return(-1)
	trix.On("AvailableContracts").Return([]domain.TrexContract{})
	trix.On("GetTrick").Return([]domain.TrexTrickCard{})
	trix.On("GetSuitRun", mock.Anything).Return(false, 0, 0)
	trix.On("GetFinishOrder").Return([]int{})
	trix.On("GetValidPlayIndices", mock.Anything).Return([]int(nil))
	trix.On("GetScore", mock.Anything).Return(0)
	trix.On("GetDealScore", mock.Anything).Return(0)
	trix.On("GetTricksWon", mock.Anything).Return(0)
	trix.On("GetConfig").Return(domain.DefaultTrexConfig())
	trix.On("GetPlayers").Return([]*domain.TrexPlayer{domain.NewTrexPlayer(true)})
	trix.On("GetActionLog").Return([]*domain.ActionLogEntry{})
	trix.On("TrexCpuDecide", mock.Anything).Return(domain.TrexCpuAction{HandIdx: -1, Pass: true})
	assert.True(t, txDecode(t, new(TrexWebPresenter).Output(trix, nil))["canPass"].(bool))

	// 同じ「出せる札が無い」状態でも、トリック契約ならパスできない。
	tricks := txStub(domain.TrexPhasePlay, domain.TrexContractTricks, 0, false, -1)
	tricks.ExpectedCalls = nil
	tricks = txStub(domain.TrexPhasePlay, domain.TrexContractTricks, 0, false, -1)
	assert.False(t, txDecode(t, new(TrexWebPresenter).Output(tricks, nil))["canPass"].(bool))
}

func TestTrexWebPresenter_HintReasonsCoverEveryBranch(t *testing.T) {
	for name, tc := range map[string]struct {
		gameEnd bool
		phase   domain.TrexPhase
		king    int
		current int
		action  domain.TrexCpuAction
		want    string
	}{
		"game over":         {true, domain.TrexPhasePlay, 0, 0, domain.TrexCpuAction{}, "trex.hint.game_end"},
		"someone else king": {false, domain.TrexPhaseChoose, 2, 2, domain.TrexCpuAction{}, "trex.hint.not_your_turn"},
		"choose":            {false, domain.TrexPhaseChoose, 0, 0, domain.TrexCpuAction{Contract: domain.TrexContractQueens}, "trex.hint.choose"},
		"not your turn":     {false, domain.TrexPhasePlay, 0, 1, domain.TrexCpuAction{}, "trex.hint.not_your_turn"},
		"pass":              {false, domain.TrexPhasePlay, 0, 0, domain.TrexCpuAction{HandIdx: -1, Pass: true}, "trex.hint.pass"},
		"no card":           {false, domain.TrexPhasePlay, 0, 0, domain.TrexCpuAction{HandIdx: -1}, "trex.hint.none"},
		"play":              {false, domain.TrexPhasePlay, 0, 0, domain.TrexCpuAction{HandIdx: 2}, "trex.hint.play"},
	} {
		t.Run(name, func(t *testing.T) {
			g := new(interfaces.MockTrexGame)
			g.On("GetGameEndFlag").Return(tc.gameEnd)
			g.On("GetPhase").Return(tc.phase)
			g.On("GetKingIdx").Return(tc.king)
			g.On("GetCurrentPlayerIdx").Return(tc.current)
			g.On("TrexCpuDecide", 0).Return(tc.action)

			assert.Equal(t, tc.want, trexHint(g).Reason)
		})
	}
}

func TestTrexWebPresenter_MessageCodes(t *testing.T) {
	win := txStub(domain.TrexPhaseGameEnd, domain.TrexContractQueens, 0, true, 0)
	assert.Equal(t, "trex.win", txDecode(t, new(TrexWebPresenter).Output(win, nil))["messageCode"])

	lose := txStub(domain.TrexPhaseGameEnd, domain.TrexContractQueens, 0, true, 2)
	assert.Equal(t, "trex.lose", txDecode(t, new(TrexWebPresenter).Output(lose, nil))["messageCode"])

	running := txStub(domain.TrexPhasePlay, domain.TrexContractQueens, 0, false, -1)
	assert.Empty(t, txDecode(t, new(TrexWebPresenter).Output(running, nil))["message"])
	assert.Equal(t, "boom", txDecode(t, new(TrexWebPresenter).Output(running, errors.New("boom")))["message"])
}

func TestTrexWebPresenter_HintOutputAndActionLog(t *testing.T) {
	tr := txTestGame(t)
	assert.NotNil(t, txDecode(t, new(TrexWebPresenter).HintOutput(tr))["hint"])
	assert.NotEmpty(t, new(TrexWebPresenter).ActionLogOutput(tr))
}

func TestTrexCuiPresenter_ListsWhatIsLeftToChoose(t *testing.T) {
	// 何が残っているかが見えていないと選べない。
	g := txStub(domain.TrexPhaseChoose, domain.TrexContractNone, 0, false, -1)
	out := new(TrexCuiPresenter).Output(g, nil)
	assert.Contains(t, out, i18n.T("trex.contractQueens"))
	assert.Contains(t, out, i18n.T("trex.contractTrix"))
	assert.Contains(t, out, i18n.T("trex.ruleLine"))
}

func TestTrexCuiPresenter_WaitsSilentlyWhenSomeoneElseIsKing(t *testing.T) {
	// 人間が王でないときに契約一覧を出すと、押せない選択肢が並ぶ。
	g := txStub(domain.TrexPhaseChoose, domain.TrexContractNone, 2, false, -1)
	out := new(TrexCuiPresenter).Output(g, nil)
	assert.Contains(t, out, i18n.Tf("trex.waitingForKing", "king", "2"))
	assert.NotContains(t, out, i18n.Tf("trex.promptChoose", "list", ""))
}

func TestTrexCuiPresenter_ShowsTheRunsOnlyInTheDominoes(t *testing.T) {
	trix := txStub(domain.TrexPhasePlay, domain.TrexContractTrix, 0, false, -1)
	assert.Contains(t, new(TrexCuiPresenter).Output(trix, nil), i18n.Tf("trex.runLine", "suit", cuiSuitName(domain.CardDesignSpade), "span", "11-13"))

	tricks := txStub(domain.TrexPhasePlay, domain.TrexContractTricks, 0, false, -1)
	assert.NotContains(t, new(TrexCuiPresenter).Output(tricks, nil), i18n.Tf("trex.runLine", "suit", cuiSuitName(domain.CardDesignSpade), "span", "11-13"))
}

func TestTrexCuiPresenter_ContractNamesAreAllMapped(t *testing.T) {
	// 未マッピングの契約は生キーを表示してしまう。
	for c := domain.TrexContractKingOfHearts; c <= domain.TrexContractTrix; c++ {
		assert.NotEmpty(t, trexContractKeys[c], "unmapped contract %d", c)
		assert.NotEqual(t, i18n.T("trex.contractNone"), trexContractName(c), "contract %d", c)
	}
	assert.Equal(t, i18n.T("trex.contractNone"), trexContractName(domain.TrexContractNone))
}

func TestTrexCuiPresenter_BannersAndErrors(t *testing.T) {
	p := new(TrexCuiPresenter)
	win := txStub(domain.TrexPhaseGameEnd, domain.TrexContractQueens, 0, true, 0)
	lose := txStub(domain.TrexPhaseGameEnd, domain.TrexContractQueens, 0, true, 2)
	assert.NotEqual(t, p.Output(win, nil), p.Output(lose, nil), "the two endings must not read the same")

	running := txStub(domain.TrexPhasePlay, domain.TrexContractQueens, 0, false, -1)
	assert.Contains(t, p.Output(running, errors.New("boom")), "boom")

	dealEnd := txStub(domain.TrexPhaseDealEnd, domain.TrexContractQueens, 0, false, -1)
	assert.Contains(t, p.Output(dealEnd, nil), i18n.T("trex.promptNext"))
}

func TestTrexCuiPresenter_HintRendersEveryShape(t *testing.T) {
	p := new(TrexCuiPresenter)

	choose := new(interfaces.MockTrexGame)
	choose.On("GetGameEndFlag").Return(false)
	choose.On("GetPhase").Return(domain.TrexPhaseChoose)
	choose.On("GetKingIdx").Return(0)
	choose.On("TrexCpuDecide", 0).Return(domain.TrexCpuAction{Contract: domain.TrexContractTrix})
	assert.Contains(t, p.HintOutput(choose), i18n.T("trex.contractTrix"))

	pass := new(interfaces.MockTrexGame)
	pass.On("GetGameEndFlag").Return(false)
	pass.On("GetPhase").Return(domain.TrexPhasePlay)
	pass.On("GetKingIdx").Return(0)
	pass.On("GetCurrentPlayerIdx").Return(0)
	pass.On("TrexCpuDecide", 0).Return(domain.TrexCpuAction{HandIdx: -1, Pass: true})
	assert.NotEmpty(t, p.HintOutput(pass))

	play := new(interfaces.MockTrexGame)
	play.On("GetGameEndFlag").Return(false)
	play.On("GetPhase").Return(domain.TrexPhasePlay)
	play.On("GetKingIdx").Return(0)
	play.On("GetCurrentPlayerIdx").Return(0)
	play.On("TrexCpuDecide", 0).Return(domain.TrexCpuAction{HandIdx: 2})
	assert.Contains(t, p.HintOutput(play), "2")

	over := new(interfaces.MockTrexGame)
	over.On("GetGameEndFlag").Return(true)
	assert.NotEmpty(t, p.HintOutput(over))
}

func TestTrexCuiPresenter_HintReasonKeysAreAllMapped(t *testing.T) {
	for _, reason := range []string{
		"trex.hint.game_end", "trex.hint.not_your_turn", "trex.hint.choose",
		"trex.hint.play", "trex.hint.pass", "trex.hint.none",
	} {
		assert.NotEmpty(t, trexHintReasonKeys[reason], "unmapped reason %s", reason)
	}
}

func TestTrexCuiPresenter_ActionLog(t *testing.T) {
	assert.NotEmpty(t, new(TrexCuiPresenter).ActionLogOutput(txTestGame(t)))
}

// #5572: どの札が失点かは 5 つの契約で入れ替わるので覚えられない (#4911)。Web は
// 赤いリングで印を付けているのに、CUI は素で並べるだけで暗算を強いていた。
func TestTrexCuiPresenter_MarksThePenaltyCardsOfEachContract(t *testing.T) {
	i18n.SetLang("ja")
	trick := func(cards ...*domain.Card) []domain.TrexTrickCard {
		out := make([]domain.TrexTrickCard, 0, len(cards))
		for i, c := range cards {
			out = append(out, domain.TrexTrickCard{PlayerIdx: i, Card: c})
		}
		return out
	}
	kingHeart := domain.NewCard(domain.CardDesignHeart, 13, true)
	queenSpade := domain.NewCard(domain.CardDesignSpade, 12, true)
	twoDiamond := domain.NewCard(domain.CardDesignDiamond, 2, true)
	nineClover := domain.NewCard(domain.CardDesignClover, 9, true)

	cards := []*domain.Card{kingHeart, queenSpade, twoDiamond, nineClover}

	for _, tc := range []struct {
		name     string
		contract domain.TrexContract
		marked   []*domain.Card
	}{
		{"king of hearts", domain.TrexContractKingOfHearts, []*domain.Card{kingHeart}},
		{"diamonds", domain.TrexContractDiamonds, []*domain.Card{twoDiamond}},
		{"queens", domain.TrexContractQueens, []*domain.Card{queenSpade}},
		// 個別札の減点が無い契約は何も印を付けない。トリックそのものが点を持つ。
		{"tricks", domain.TrexContractTricks, nil},
		{"trix", domain.TrexContractTrix, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := txStub(domain.TrexPhasePlay, tc.contract, 0, false, -1)
			g.ExpectedCalls = filterCalls(g.ExpectedCalls, "GetTrick")
			g.On("GetTrick").Return(trick(cards...))

			out := new(TrexCuiPresenter).Output(g, nil)
			for _, c := range cards {
				want := false
				for _, m := range tc.marked {
					if m == c {
						want = true
					}
				}
				mark := cuiCardStr(c) + i18n.Tf("trex.penaltyMark",
					"points", strconv.Itoa(domain.TrexCardPenalty(tc.contract, c)))
				if want {
					assert.Contains(t, out, mark, "%s must be marked", cuiCardStr(c))
					continue
				}
				// **印の付いていない札に印を付けない。**全部に付ける実装でも
				// 「含む」検査だけなら通ってしまう。
				assert.NotContains(t, out, cuiCardStr(c)+"(", "%s must not be marked", cuiCardStr(c))
			}
		})
	}
}

// 印の数字は得点を決めている関数から来ること。書き写すと、印と実際の失点が
// ずれても誰も気づけない。
func TestTrexCardPenaltyIsTheOneUsedForScoring(t *testing.T) {
	assert.Equal(t, domain.TrexKingOfHeartsPenalty,
		domain.TrexCardPenalty(domain.TrexContractKingOfHearts, domain.NewCard(domain.CardDesignHeart, 13, true)))
	assert.Equal(t, domain.TrexDiamondPenalty,
		domain.TrexCardPenalty(domain.TrexContractDiamonds, domain.NewCard(domain.CardDesignDiamond, 5, true)))
	assert.Equal(t, domain.TrexQueenPenalty,
		domain.TrexCardPenalty(domain.TrexContractQueens, domain.NewCard(domain.CardDesignClover, 12, true)))
	// ♥K 契約の ♥Q は失点でない (クイーン契約と混ざらないこと)。
	assert.Zero(t, domain.TrexCardPenalty(domain.TrexContractKingOfHearts, domain.NewCard(domain.CardDesignHeart, 12, true)))
	assert.Zero(t, domain.TrexCardPenalty(domain.TrexContractTricks, domain.NewCard(domain.CardDesignHeart, 13, true)))
	assert.Zero(t, domain.TrexCardPenalty(domain.TrexContractQueens, nil))
}
