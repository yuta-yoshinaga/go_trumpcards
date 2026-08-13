package presenter_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// newZwanzigerrufenGame は開始直後の局面を返す。
func newZwanzigerrufenGame() *domain.Zwanzigerrufen {
	g := domain.NewDefaultZwanzigerrufen()
	g.Reset()
	return g
}

// zwanzigerrufenStep はフェーズに応じて 1 歩進める (CPU 側 / 自動進行)。
func zwanzigerrufenStep(g *domain.Zwanzigerrufen) {
	switch g.GetPhase() {
	case domain.ZwanzigerrufenPhaseBid:
		g.CpuBid()
	case domain.ZwanzigerrufenPhaseTalon:
		g.CpuDiscard()
	case domain.ZwanzigerrufenPhasePlay:
		g.CpuPlayCard()
	case domain.ZwanzigerrufenPhaseTrickEnd:
		g.NextTrick()
	case domain.ZwanzigerrufenPhaseRoundEnd:
		g.NextRound()
	}
}

// zwanzigerrufenHumanStep は人間の手番を推奨手で 1 歩進める。
func zwanzigerrufenHumanStep(t *testing.T, g *domain.Zwanzigerrufen) {
	t.Helper()
	h := g.GetHint()
	require.NotNil(t, h)
	switch g.GetPhase() {
	case domain.ZwanzigerrufenPhaseBid:
		require.NoError(t, g.PlayerPass())
	case domain.ZwanzigerrufenPhaseTalon:
		require.NoError(t, g.PlayerDiscard(h.DiscardIndices))
	case domain.ZwanzigerrufenPhasePlay:
		require.NotNil(t, h.CardIndex)
		require.NoError(t, g.PlayerPlayCard(*h.CardIndex))
	}
}

// zwanzigerrufenTablePlayers は 4 席 (席 0 が人間) を返す。
func zwanzigerrufenTablePlayers() []*domain.ZwanzigerrufenPlayer {
	ps := make([]*domain.ZwanzigerrufenPlayer, domain.ZwanzigerrufenPlayerCnt)
	ps[0] = domain.NewZwanzigerrufenPlayer(true)
	for i := 1; i < domain.ZwanzigerrufenPlayerCnt; i++ {
		ps[i] = domain.NewZwanzigerrufenPlayer(false)
	}
	return ps
}

// zwanzigerrufenPlayOut は終局まで進める。
func zwanzigerrufenPlayOut(t *testing.T, g *domain.Zwanzigerrufen) {
	t.Helper()
	for range 3000 {
		if g.GetGameEndFlag() {
			return
		}
		if g.IsHumanTurn() {
			zwanzigerrufenHumanStep(t, g)
			continue
		}
		zwanzigerrufenStep(g)
	}
	t.Fatal("終局しなかった")
}

// --- Web ---

func TestZwanzigerrufenWebPresenter_Output(t *testing.T) {
	g := newZwanzigerrufenGame()
	out := new(presenter.ZwanzigerrufenWebPresenter).Output(g, nil)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, float64(domain.ZwanzigerrufenPhaseBid), decoded["phase"])
	assert.Equal(t, float64(1), decoded["roundNumber"])
	assert.Len(t, decoded["players"].([]any), domain.ZwanzigerrufenPlayerCnt)
	assert.Equal(t, float64(domain.ZwanzigerrufenTalonSize), decoded["talonCount"])
	assert.Equal(t, float64(-1), decoded["declarerIdx"])
	assert.Equal(t, float64(-1), decoded["calledTrump"])
	assert.Contains(t, decoded, "playableIndices")
}

// **相手の手札はワイヤに乗せない。** 枚数だけを出す。
func TestZwanzigerrufenWebPresenter_HidesOpponentHands(t *testing.T) {
	g := newZwanzigerrufenGame()
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
		[]byte(new(presenter.ZwanzigerrufenWebPresenter).Output(g, nil)), &parsed))

	human, cpu := false, false
	for _, p := range parsed.Players {
		assert.Equal(t, domain.ZwanzigerrufenHandSize, p.CardCount)
		if p.IsHuman {
			human = true
			require.Len(t, p.Cards, domain.ZwanzigerrufenHandSize)
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

// **秘密のパートナーは判明するまで漏らさない。**
//
// ドメインに本物の局面を作らせると配り依存になるので、プレゼンターの防壁そのものを
// 見る ── ゲーム側が席 2 をパートナーと答えていても、未公開なら出力は -1 でなければ
// ならない。モックが -1 を返すのを確かめるだけでは、防壁を消しても気付けない。
func TestZwanzigerrufenWebPresenter_KeepsThePartnerSecret(t *testing.T) {
	tests := []struct {
		name     string
		revealed bool
		want     int
	}{
		{"未公開なら伏せる", false, -1},
		{"公開後は出す", true, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gm := newZwanzigerrufenPresenterMock()
			gm.On("GetPartnerRevealed").Return(tt.revealed)
			gm.On("GetPartnerIdx").Return(2)

			var parsed struct {
				PartnerIdx      int  `json:"partnerIdx"`
				PartnerRevealed bool `json:"partnerRevealed"`
				Players         []struct {
					IsPartner bool `json:"isPartner"`
				} `json:"players"`
			}
			require.NoError(t, json.Unmarshal(
				[]byte(new(presenter.ZwanzigerrufenWebPresenter).Output(gm, nil)), &parsed))

			assert.Equal(t, tt.want, parsed.PartnerIdx)
			assert.Equal(t, tt.revealed, parsed.PartnerRevealed)
			seen := false
			for _, p := range parsed.Players {
				seen = seen || p.IsPartner
			}
			assert.Equal(t, tt.revealed, seen, "席のパートナー印が食い違っている")
		})
	}
}

// newZwanzigerrufenPresenterMock はパートナー以外を無害な既定値で埋めたモックを返す。
func newZwanzigerrufenPresenterMock() *interfaces.MockZwanzigerrufenGame {
	gm := new(interfaces.MockZwanzigerrufenGame)
	gm.On("GetConfig").Return(domain.DefaultZwanzigerrufenConfig())
	gm.On("GetPhase").Return(domain.ZwanzigerrufenPhasePlay)
	gm.On("GetRoundNumber").Return(1)
	gm.On("GetTrickNumber").Return(1)
	gm.On("GetCurrentPlayerIdx").Return(0)
	gm.On("GetDealerIdx").Return(0)
	gm.On("GetBidPlayerIdx").Return(0)
	gm.On("GetHighestBid").Return(domain.ZwanzigerrufenBidRufer)
	gm.On("GetDeclarerIdx").Return(0)
	gm.On("GetContract").Return(domain.ZwanzigerrufenBidRufer)
	gm.On("GetCalledTrump").Return(domain.ZwanzigerrufenCallTrump)
	gm.On("GetTalonSize").Return(0)
	gm.On("GetLastTrickWinner").Return(-1)
	gm.On("GetLastTrickCards").Return([]*domain.Card(nil))
	gm.On("GetOutcome").Return(domain.ZwanzigerrufenOutcomeNone)
	gm.On("GetBreakdown").Return((*domain.ZwanzigerrufenBreakdown)(nil))
	gm.On("GetGameEndFlag").Return(false)
	gm.On("GetWinnerPlayer").Return(-1)
	gm.On("IsHumanTurn").Return(false)
	gm.On("GetCurrentTrick").Return([]*domain.TrickCard(nil))
	gm.On("GetValidPlayIndices", mock.Anything).Return([]int(nil))
	gm.On("GetPlayerCnt").Return(domain.ZwanzigerrufenPlayerCnt)
	gm.On("GetPlayerScore", mock.Anything).Return(0)
	gm.On("GetCardPoints", mock.Anything).Return(0)
	gm.On("GetHint").Return((*domain.ZwanzigerrufenHint)(nil))
	for i := range domain.ZwanzigerrufenPlayerCnt {
		gm.On("GetPlayer", i).Return(domain.NewZwanzigerrufenPlayer(i == 0))
	}
	return gm
}

func TestZwanzigerrufenWebPresenter_Error(t *testing.T) {
	g := newZwanzigerrufenGame()
	out := new(presenter.ZwanzigerrufenWebPresenter).Output(g, errors.New("boom"))
	var decoded map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &decoded))
	assert.Equal(t, "boom", decoded["message"])
}

func TestZwanzigerrufenWebPresenter_GameEndCarriesScores(t *testing.T) {
	g := domain.NewZwanzigerrufen(zwanzigerrufenTablePlayers(), domain.ZwanzigerrufenConfig{TargetDeals: 1})
	g.Reset()
	zwanzigerrufenPlayOut(t, g)

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
		[]byte(new(presenter.ZwanzigerrufenWebPresenter).Output(g, nil)), &decoded))
	assert.True(t, decoded.GameEndFlag)
	assert.Equal(t, "zwanzigerrufen.result.scores", decoded.MessageCode)
	assert.Contains(t, decoded.MessageParams["scores"], "0:")
	require.NotNil(t, decoded.Breakdown)
	assert.Len(t, decoded.Breakdown.Seats, domain.ZwanzigerrufenPlayerCnt)
	assert.NotEmpty(t, decoded.Breakdown.Name, "契約名が出ていない")
	// **精算はゼロサム。**
	sum := 0
	for _, v := range decoded.Breakdown.Seats {
		sum += v
	}
	assert.Equal(t, 0, sum)
}

func TestZwanzigerrufenWebPresenter_HintAndLog(t *testing.T) {
	g := newZwanzigerrufenGame()
	var decoded struct {
		Hint *struct {
			Bid    *int   `json:"bid"`
			Reason string `json:"reason"`
		} `json:"hint"`
	}
	require.NoError(t, json.Unmarshal(
		[]byte(new(presenter.ZwanzigerrufenWebPresenter).HintOutput(g)), &decoded))
	if g.IsHumanTurn() {
		require.NotNil(t, decoded.Hint, "人間の手番なのにヒントが出ていない")
		assert.NotEmpty(t, decoded.Hint.Reason)
	}

	var logDecoded map[string]any
	require.NoError(t, json.Unmarshal(
		[]byte(new(presenter.ZwanzigerrufenWebPresenter).ActionLogOutput(g)), &logDecoded))
}

// --- CUI ---

func TestZwanzigerrufenCuiPresenter_Output(t *testing.T) {
	g := newZwanzigerrufenGame()
	out := new(presenter.ZwanzigerrufenCuiPresenter).Output(g, nil)

	assert.Contains(t, out, i18n.T("zwanzigerrufen.helpTitle"))
	assert.Contains(t, out, "[0]", "人間の手札に番号が付いていない")
	assert.NotContains(t, out, "zwanzigerrufen.", "i18n キーが生のまま出ている")
}

func TestZwanzigerrufenCuiPresenter_Output_Error(t *testing.T) {
	g := newZwanzigerrufenGame()
	assert.Contains(t, new(presenter.ZwanzigerrufenCuiPresenter).Output(g, errors.New("boom")), "boom")
}

// フェーズごとにプロンプトが変わり、終局まで進む。
func TestZwanzigerrufenCuiPresenter_PromptsByPhase(t *testing.T) {
	p := new(presenter.ZwanzigerrufenCuiPresenter)
	g := newZwanzigerrufenGame()
	assert.Contains(t, p.Output(g, nil), i18n.T("zwanzigerrufen.promptBidHelp"))

	seen := map[domain.ZwanzigerrufenPhase]bool{}
	for range 3000 {
		if g.GetGameEndFlag() {
			break
		}
		phase := g.GetPhase()
		if !seen[phase] {
			seen[phase] = true
			switch phase {
			case domain.ZwanzigerrufenPhaseTalon:
				assert.Contains(t, p.Output(g, nil), i18n.T("zwanzigerrufen.promptTalonHelp"))
			case domain.ZwanzigerrufenPhasePlay:
				assert.Contains(t, p.Output(g, nil), i18n.T("zwanzigerrufen.promptPlayHelp"))
			case domain.ZwanzigerrufenPhaseTrickEnd:
				assert.Contains(t, p.Output(g, nil), i18n.T("zwanzigerrufen.promptTrickEndHelp"))
			case domain.ZwanzigerrufenPhaseRoundEnd:
				assert.Contains(t, p.Output(g, nil), i18n.T("zwanzigerrufen.promptRoundEndHelp"))
			}
		}
		if g.IsHumanTurn() {
			zwanzigerrufenHumanStep(t, g)
			continue
		}
		zwanzigerrufenStep(g)
	}
	require.True(t, g.GetGameEndFlag())
	assert.True(t, seen[domain.ZwanzigerrufenPhasePlay], "プレイフェーズを通っていない")

	out := p.Output(g, nil)
	assert.True(t,
		strings.Contains(out, "の勝ち") || strings.Contains(out, "引き分け"),
		"終局の行が出ていない: %s", out)
}

func TestZwanzigerrufenCuiPresenter_HintOutput(t *testing.T) {
	g := newZwanzigerrufenGame()
	out := new(presenter.ZwanzigerrufenCuiPresenter).HintOutput(g)
	if g.IsHumanTurn() {
		assert.Contains(t, out, "HINT")
		assert.NotContains(t, out, "zwanzigerrufen.hintReason", "理由が訳されていない")
		return
	}
	assert.Contains(t, out, i18n.T("zwanzigerrufen.hintNone"))
}

func TestZwanzigerrufenCuiPresenter_HintNoneOutsideHumanTurns(t *testing.T) {
	g := newZwanzigerrufenGame()
	for range 200 {
		if !g.IsHumanTurn() {
			break
		}
		zwanzigerrufenHumanStep(t, g)
	}
	if g.IsHumanTurn() {
		t.Skip("この配りでは人間の手番から離れなかった")
	}
	assert.Contains(t, new(presenter.ZwanzigerrufenCuiPresenter).HintOutput(g),
		i18n.T("zwanzigerrufen.hintNone"))
}

func TestZwanzigerrufenCuiPresenter_ActionLogOutput(t *testing.T) {
	g := newZwanzigerrufenGame()
	assert.NotEmpty(t, new(presenter.ZwanzigerrufenCuiPresenter).ActionLogOutput(g))
}
