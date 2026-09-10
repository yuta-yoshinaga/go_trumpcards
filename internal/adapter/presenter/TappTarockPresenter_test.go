package presenter_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// newTappTarockGame は開始直後の局面を返す。
func newTappTarockGame() *domain.TappTarock {
	g := domain.NewDefaultTappTarock()
	g.Reset()
	return g
}

// tapptarockStep はフェーズに応じて 1 歩進める (CPU 側 / 自動進行)。
func tapptarockStep(g *domain.TappTarock) {
	switch g.GetPhase() {
	case domain.TappTarockPhaseBid:
		g.CpuBid()
	case domain.TappTarockPhaseTalon:
		g.CpuDiscard()
	case domain.TappTarockPhasePlay:
		g.CpuPlayCard()
	case domain.TappTarockPhaseTrickEnd:
		g.NextTrick()
	case domain.TappTarockPhaseRoundEnd:
		g.NextRound()
	}
}

// tapptarockHumanStep は人間の手番を推奨手で 1 歩進める。
func tapptarockHumanStep(t *testing.T, g *domain.TappTarock) {
	t.Helper()
	h := g.GetHint()
	require.NotNil(t, h)
	switch g.GetPhase() {
	case domain.TappTarockPhaseBid:
		require.NoError(t, g.PlayerPass())
	case domain.TappTarockPhaseTalon:
		require.NoError(t, g.PlayerDiscard(h.DiscardIndices))
	case domain.TappTarockPhasePlay:
		require.NotNil(t, h.CardIndex)
		require.NoError(t, g.PlayerPlayCard(*h.CardIndex))
	}
}

// tapptarockTablePlayers は 4 席 (席 0 が人間) を返す。
func tapptarockTablePlayers() []*domain.TappTarockPlayer {
	ps := make([]*domain.TappTarockPlayer, domain.TappTarockPlayerCnt)
	ps[0] = domain.NewTappTarockPlayer(true)
	for i := 1; i < domain.TappTarockPlayerCnt; i++ {
		ps[i] = domain.NewTappTarockPlayer(false)
	}
	return ps
}

// tapptarockPlayOut は終局まで進める。
func tapptarockPlayOut(t *testing.T, g *domain.TappTarock) {
	t.Helper()
	for range 3000 {
		if g.GetGameEndFlag() {
			return
		}
		if g.IsHumanTurn() {
			tapptarockHumanStep(t, g)
			continue
		}
		tapptarockStep(g)
	}
	t.Fatal("終局しなかった")
}

// --- Web ---

func TestTappTarockWebPresenter_Output(t *testing.T) {
	g := newTappTarockGame()
	out := new(presenter.TappTarockWebPresenter).Output(g, nil)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, float64(domain.TappTarockPhaseBid), decoded["phase"])
	assert.Equal(t, float64(1), decoded["roundNumber"])
	assert.Len(t, decoded["players"].([]any), domain.TappTarockPlayerCnt)
	assert.Equal(t, float64(domain.TappTarockTalonSize), decoded["talonCount"])
	assert.Equal(t, float64(-1), decoded["declarerIdx"])
	assert.Contains(t, decoded, "playableIndices")
}

// **相手の手札はワイヤに乗せない。** 枚数だけを出す。
func TestTappTarockWebPresenter_HidesOpponentHands(t *testing.T) {
	g := newTappTarockGame()
	var parsed struct {
		Players []struct {
			IsHuman   bool `json:"isHuman"`
			CardCount int  `json:"cardCount"`
			Cards     []struct {
				Glyph string `json:"glyph"`
				Deck  string `json:"deck"`
			} `json:"cards"`
		} `json:"players"`
	}
	require.NoError(t, json.Unmarshal(
		[]byte(new(presenter.TappTarockWebPresenter).Output(g, nil)), &parsed))

	human, cpu := false, false
	for _, p := range parsed.Players {
		assert.Equal(t, domain.TappTarockHandSize, p.CardCount)
		if p.IsHuman {
			human = true
			require.Len(t, p.Cards, domain.TappTarockHandSize)
			for _, c := range p.Cards {
				assert.Equal(t, "tarot", c.Deck, "タロックは手続き描画 (ADR-0033)")
				assert.NotEmpty(t, c.Glyph)
			}
			continue
		}
		cpu = true
		assert.Empty(t, p.Cards, "相手の手札が出力に載っている")
	}
	assert.True(t, human && cpu, "両方の席を確かめていない")
}

// newTappTarockPresenterMock はパートナー以外を無害な既定値で埋めたモックを返す。
func newTappTarockPresenterMock() *interfaces.MockTappTarockGame {
	gm := new(interfaces.MockTappTarockGame)
	gm.On("GetConfig").Return(domain.DefaultTappTarockConfig())
	gm.On("GetPhase").Return(domain.TappTarockPhasePlay)
	gm.On("GetRoundNumber").Return(1)
	gm.On("GetTrickNumber").Return(1)
	gm.On("GetCurrentPlayerIdx").Return(0)
	gm.On("GetDealerIdx").Return(0)
	gm.On("GetBidPlayerIdx").Return(0)
	gm.On("GetHighestBid").Return(domain.TappTarockBidDreier)
	gm.On("GetDeclarerIdx").Return(0)
	gm.On("GetContract").Return(domain.TappTarockBidDreier)
	gm.On("GetTalonSize").Return(0)
	gm.On("GetLastTrickWinner").Return(-1)
	gm.On("GetLastTrickCards").Return([]*domain.Card(nil))
	gm.On("GetOutcome").Return(domain.TappTarockOutcomeNone)
	gm.On("GetBreakdown").Return((*domain.TappTarockBreakdown)(nil))
	gm.On("GetGameEndFlag").Return(false)
	gm.On("GetWinnerPlayer").Return(-1)
	gm.On("IsHumanTurn").Return(false)
	gm.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	gm.On("GetValidPlayIndices", mock.Anything).Return([]int(nil))
	gm.On("GetDiscardableIndices").Return([]int(nil))
	gm.On("GetPlayerCnt").Return(domain.TappTarockPlayerCnt)
	gm.On("GetPlayerScore", mock.Anything).Return(0)
	gm.On("GetCardPoints", mock.Anything).Return(0)
	gm.On("GetHint").Return((*domain.TappTarockHint)(nil))
	for i := range domain.TappTarockPlayerCnt {
		gm.On("GetPlayer", i).Return(domain.NewTappTarockPlayer(i == 0))
	}
	return gm
}

func TestTappTarockWebPresenter_DiscardableIndices(t *testing.T) {
	decode := func(g interfaces.TappTarockGame) controller.TappTarockWebOutput {
		p := new(presenter.TappTarockWebPresenter)
		var parsed controller.TappTarockWebOutput
		require.NoError(t, json.Unmarshal([]byte(p.Output(g, nil)), &parsed))
		return parsed
	}

	t.Run("人間がデクレアラーのタロンフェーズではインデックスを載せる", func(t *testing.T) {
		gm := newTappTarockPresenterMock()
		gm.ExpectedCalls = nil
		gm.On("GetConfig").Return(domain.DefaultTappTarockConfig())
		gm.On("GetPhase").Return(domain.TappTarockPhaseTalon)
		gm.On("GetRoundNumber").Return(1)
		gm.On("GetTrickNumber").Return(0)
		gm.On("GetCurrentPlayerIdx").Return(0)
		gm.On("GetDealerIdx").Return(0)
		gm.On("GetBidPlayerIdx").Return(0)
		gm.On("GetHighestBid").Return(domain.TappTarockBidDreier)
		gm.On("GetDeclarerIdx").Return(0)
		gm.On("GetContract").Return(domain.TappTarockBidDreier)
		gm.On("GetTalonSize").Return(6)
		gm.On("GetLastTrickWinner").Return(-1)
		gm.On("GetLastTrickCards").Return([]*domain.Card(nil))
		gm.On("GetOutcome").Return(domain.TappTarockOutcomeNone)
		gm.On("GetBreakdown").Return((*domain.TappTarockBreakdown)(nil))
		gm.On("GetGameEndFlag").Return(false)
		gm.On("GetWinnerPlayer").Return(-1)
		gm.On("IsHumanTurn").Return(true)
		gm.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
		gm.On("GetValidPlayIndices", mock.Anything).Return([]int(nil))
		gm.On("GetDiscardableIndices").Return([]int{0, 1, 2})
		gm.On("GetPlayerCnt").Return(domain.TappTarockPlayerCnt)
		gm.On("GetPlayerScore", mock.Anything).Return(0)
		gm.On("GetCardPoints", mock.Anything).Return(0)
		gm.On("GetHint").Return((*domain.TappTarockHint)(nil))
		for i := range domain.TappTarockPlayerCnt {
			gm.On("GetPlayer", i).Return(domain.NewTappTarockPlayer(i == 0))
		}

		got := decode(gm)
		assert.Equal(t, []int{0, 1, 2}, got.DiscardableIndices)
	})

	t.Run("プレイフェーズでは空", func(t *testing.T) {
		gm := newTappTarockPresenterMock()
		gm.On("GetPhase").Return(domain.TappTarockPhasePlay)
		gm.On("GetDeclarerIdx").Return(0)

		got := decode(gm)
		assert.Empty(t, got.DiscardableIndices)
	})

	t.Run("デクレアラーが人間でないときは空", func(t *testing.T) {
		gm := newTappTarockPresenterMock()
		gm.On("GetPhase").Return(domain.TappTarockPhaseTalon)
		gm.On("GetDeclarerIdx").Return(1)

		got := decode(gm)
		assert.Empty(t, got.DiscardableIndices)
	})
}

func TestTappTarockWebPresenter_Error(t *testing.T) {
	g := newTappTarockGame()
	out := new(presenter.TappTarockWebPresenter).Output(g, errors.New("boom"))
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "boom", decoded["message"])
}

func TestTappTarockWebPresenter_GameEndCarriesScores(t *testing.T) {
	g := domain.NewTappTarock(tapptarockTablePlayers(), domain.TappTarockConfig{TargetDeals: 1})
	g.Reset()
	tapptarockPlayOut(t, g)

	var decoded struct {
		GameEndFlag   bool              `json:"gameEndFlag"`
		MessageCode   string            `json:"messageCode"`
		MessageParams map[string]string `json:"messageParams"`
		Breakdown     *struct {
			Contract int    `json:"contract"`
			Name     string `json:"name"`
			Seats    []int  `json:"seats"`
			Loser    int    `json:"loser"`
		} `json:"breakdown"`
	}
	require.NoError(t, json.Unmarshal(
		[]byte(new(presenter.TappTarockWebPresenter).Output(g, nil)), &decoded))
	assert.True(t, decoded.GameEndFlag)
	assert.Equal(t, "tapptarock.result.scores", decoded.MessageCode)
	assert.Contains(t, decoded.MessageParams["scores"], "0:")
	require.NotNil(t, decoded.Breakdown)
	assert.Len(t, decoded.Breakdown.Seats, domain.TappTarockPlayerCnt)
	assert.NotEmpty(t, decoded.Breakdown.Name, "契約名が出ていない")
	assert.Len(t, decoded.Breakdown.Seats, domain.TappTarockPlayerCnt)
}

func TestTappTarockWebPresenter_HintAndLog(t *testing.T) {
	g := newTappTarockGame()
	var decoded struct {
		Hint *struct {
			Bid    *int   `json:"bid"`
			Reason string `json:"reason"`
		} `json:"hint"`
	}
	require.NoError(t, json.Unmarshal(
		[]byte(new(presenter.TappTarockWebPresenter).HintOutput(g)), &decoded))
	if g.IsHumanTurn() {
		require.NotNil(t, decoded.Hint, "人間の手番なのにヒントが出ていない")
		assert.NotEmpty(t, decoded.Hint.Reason)
	}

	var logDecoded map[string]any
	require.NoError(t, json.Unmarshal(
		[]byte(new(presenter.TappTarockWebPresenter).ActionLogOutput(g)), &logDecoded))
}

// --- CUI ---

func TestTappTarockCuiPresenter_Output(t *testing.T) {
	g := newTappTarockGame()
	out := new(presenter.TappTarockCuiPresenter).Output(g, nil)

	assert.Contains(t, out, i18n.T("tapptarock.helpTitle"))
	assert.Contains(t, out, "[0]", "人間の手札に番号が付いていない")
	assert.NotContains(t, out, "tapptarock.", "i18n キーが生のまま出ている")
}

func TestTappTarockCuiPresenter_Output_Error(t *testing.T) {
	g := newTappTarockGame()
	assert.Contains(t, new(presenter.TappTarockCuiPresenter).Output(g, errors.New("boom")), "boom")
}

// フェーズごとにプロンプトが変わり、終局まで進む。
func TestTappTarockCuiPresenter_PromptsByPhase(t *testing.T) {
	p := new(presenter.TappTarockCuiPresenter)
	g := newTappTarockGame()
	assert.Contains(t, p.Output(g, nil), i18n.T("tapptarock.promptBidHelp"))

	seen := map[domain.TappTarockPhase]bool{}
	for range 3000 {
		if g.GetGameEndFlag() {
			break
		}
		phase := g.GetPhase()
		if !seen[phase] {
			seen[phase] = true
			switch phase {
			case domain.TappTarockPhaseTalon:
				assert.Contains(t, p.Output(g, nil), i18n.T("tapptarock.promptTalonHelp"))
			case domain.TappTarockPhasePlay:
				assert.Contains(t, p.Output(g, nil), i18n.T("tapptarock.promptPlayHelp"))
			case domain.TappTarockPhaseTrickEnd:
				assert.Contains(t, p.Output(g, nil), i18n.T("tapptarock.promptTrickEndHelp"))
			case domain.TappTarockPhaseRoundEnd:
				assert.Contains(t, p.Output(g, nil), i18n.T("tapptarock.promptRoundEndHelp"))
			}
		}
		if g.IsHumanTurn() {
			tapptarockHumanStep(t, g)
			continue
		}
		tapptarockStep(g)
	}
	require.True(t, g.GetGameEndFlag())
	assert.True(t, seen[domain.TappTarockPhasePlay], "プレイフェーズを通っていない")

	out := p.Output(g, nil)
	assert.True(t,
		strings.Contains(out, "の勝ち") || strings.Contains(out, "引き分け"),
		"終局の行が出ていない: %s", out)
}

func TestTappTarockCuiPresenter_HintOutput(t *testing.T) {
	g := newTappTarockGame()
	out := new(presenter.TappTarockCuiPresenter).HintOutput(g)
	if g.IsHumanTurn() {
		assert.Contains(t, out, "HINT")
		assert.NotContains(t, out, "tapptarock.hintReason", "理由が訳されていない")
		return
	}
	assert.Contains(t, out, i18n.T("tapptarock.hintNone"))
}

func TestTappTarockCuiPresenter_HintNoneOutsideHumanTurns(t *testing.T) {
	g := newTappTarockGame()
	for range 200 {
		if !g.IsHumanTurn() {
			break
		}
		tapptarockHumanStep(t, g)
	}
	if g.IsHumanTurn() {
		t.Skip("この配りでは人間の手番から離れなかった")
	}
	assert.Contains(t, new(presenter.TappTarockCuiPresenter).HintOutput(g),
		i18n.T("tapptarock.hintNone"))
}

func TestTappTarockCuiPresenter_ActionLogOutput(t *testing.T) {
	g := newTappTarockGame()
	assert.NotEmpty(t, new(presenter.TappTarockCuiPresenter).ActionLogOutput(g))
}
