package presenter_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// newTrogguGame は開始直後の局面を返す。
func newTrogguGame() *domain.Troggu {
	g := domain.NewDefaultTroggu()
	g.Reset()
	return g
}

// trogguTable は 4 席 (席 0 が人間) を返す。
func trogguTable() []*domain.TrogguPlayer {
	ps := make([]*domain.TrogguPlayer, domain.TrogguPlayerCnt)
	ps[0] = domain.NewTrogguPlayer(true)
	for i := 1; i < domain.TrogguPlayerCnt; i++ {
		ps[i] = domain.NewTrogguPlayer(false)
	}
	return ps
}

// trogguStepGame は 1 歩進める。
func trogguStepGame(t *testing.T, g *domain.Troggu) {
	t.Helper()
	if g.IsHumanTurn() {
		switch g.GetPhase() {
		case domain.TrogguPhaseBid:
			require.NoError(t, g.PlayerPass())
		case domain.TrogguPhasePlay:
			h := g.GetHint()
			require.NotNil(t, h)
			require.NotNil(t, h.CardIndex)
			require.NoError(t, g.PlayerPlayCard(*h.CardIndex))
		}
		return
	}
	switch g.GetPhase() {
	case domain.TrogguPhaseBid:
		g.CpuBid()
	case domain.TrogguPhasePlay:
		g.CpuPlayCard()
	case domain.TrogguPhaseTrickEnd:
		g.NextTrick()
	case domain.TrogguPhaseRoundEnd:
		g.NextRound()
	}
}

// trogguPlayOutGame は終局まで進める。
func trogguPlayOutGame(t *testing.T, g *domain.Troggu) {
	t.Helper()
	for range 5000 {
		if g.GetGameEndFlag() {
			return
		}
		trogguStepGame(t, g)
	}
	t.Fatal("終局しなかった")
}

// --- Web ---

func TestTrogguWebPresenter_Output(t *testing.T) {
	g := newTrogguGame()
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(
		[]byte(new(presenter.TrogguWebPresenter).Output(g, nil)), &decoded))

	assert.Equal(t, float64(domain.TrogguPhaseBid), decoded["phase"])
	assert.Equal(t, float64(1), decoded["roundNumber"])
	assert.Len(t, decoded["players"].([]any), domain.TrogguPlayerCnt)
	assert.Equal(t, float64(domain.TrogguTalonSize), decoded["talonCount"])
	assert.Equal(t, float64(-1), decoded["declarerIdx"])
	assert.Equal(t, "pass", decoded["contractName"])
	assert.Contains(t, decoded, "playableIndices")
}

// **相手の手札はワイヤに乗せない。** 枚数だけを出す。
func TestTrogguWebPresenter_HidesOpponentHands(t *testing.T) {
	g := newTrogguGame()
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
		[]byte(new(presenter.TrogguWebPresenter).Output(g, nil)), &parsed))

	human, cpu := false, false
	for _, p := range parsed.Players {
		assert.Equal(t, domain.TrogguHandSize, p.CardCount)
		if p.IsHuman {
			human = true
			require.Len(t, p.Cards, domain.TrogguHandSize)
			for _, c := range p.Cards {
				assert.NotEmpty(t, c.Deck, "タローは手続き描画 (ADR-0033)")
				assert.NotEmpty(t, c.Glyph)
			}
			continue
		}
		cpu = true
		assert.Empty(t, p.Cards, "相手の手札が出力に載っている")
	}
	assert.True(t, human && cpu, "両方の席を確かめていない")
}

func TestTrogguWebPresenter_Error(t *testing.T) {
	g := newTrogguGame()
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(
		[]byte(new(presenter.TrogguWebPresenter).Output(g, errors.New("boom"))), &decoded))
	assert.Equal(t, "boom", decoded["message"])
}

func TestTrogguWebPresenter_GameEndCarriesScoresAndBreakdown(t *testing.T) {
	g := domain.NewTroggu(trogguTable(), domain.TrogguConfig{TargetDeals: 1})
	g.Reset()
	trogguPlayOutGame(t, g)

	var decoded struct {
		GameEndFlag   bool              `json:"gameEndFlag"`
		MessageCode   string            `json:"messageCode"`
		MessageParams map[string]string `json:"messageParams"`
		Breakdown     *struct {
			ContractName   string `json:"contractName"`
			TargetIsTricks bool   `json:"targetIsTricks"`
			Seats          []int  `json:"seats"`
		} `json:"breakdown"`
	}
	require.NoError(t, json.Unmarshal(
		[]byte(new(presenter.TrogguWebPresenter).Output(g, nil)), &decoded))
	assert.True(t, decoded.GameEndFlag)
	assert.Equal(t, "troggu.result.scores", decoded.MessageCode)
	assert.Contains(t, decoded.MessageParams["scores"], "0:")

	// 流局したディールには精算が無い。
	if g.GetBreakdown() == nil {
		assert.Nil(t, decoded.Breakdown)
		return
	}
	require.NotNil(t, decoded.Breakdown)
	assert.NotEmpty(t, decoded.Breakdown.ContractName)
	assert.Len(t, decoded.Breakdown.Seats, domain.TrogguPlayerCnt)
	sum := 0
	for _, v := range decoded.Breakdown.Seats {
		sum += v
	}
	assert.Equal(t, 0, sum, "精算がゼロサムでない")
}

func TestTrogguWebPresenter_HintAndLog(t *testing.T) {
	g := newTrogguGame()
	var decoded struct {
		Hint *struct {
			Reason string `json:"reason"`
		} `json:"hint"`
	}
	require.NoError(t, json.Unmarshal(
		[]byte(new(presenter.TrogguWebPresenter).HintOutput(g)), &decoded))
	if g.IsHumanTurn() {
		require.NotNil(t, decoded.Hint)
		assert.NotEmpty(t, decoded.Hint.Reason)
	}

	var logDecoded map[string]any
	require.NoError(t, json.Unmarshal(
		[]byte(new(presenter.TrogguWebPresenter).ActionLogOutput(g)), &logDecoded))
}

// --- CUI ---

func TestTrogguCuiPresenter_Output(t *testing.T) {
	g := newTrogguGame()
	out := new(presenter.TrogguCuiPresenter).Output(g, nil)

	assert.Contains(t, out, i18n.T("troggu.helpTitle"))
	assert.Contains(t, out, "[0]", "人間の手札に番号が付いていない")
	assert.NotContains(t, out, "troggu.", "i18n キーが生のまま出ている")
}

func TestTrogguCuiPresenter_Output_Error(t *testing.T) {
	g := newTrogguGame()
	assert.Contains(t, new(presenter.TrogguCuiPresenter).Output(g, errors.New("boom")), "boom")
}

// **契約ごとに結果の文型が違う。** ソロは点数、他はトリック数で語る。
func TestTrogguCuiPresenter_RoundResultSpeaksTheContractsUnits(t *testing.T) {
	p := new(presenter.TrogguCuiPresenter)
	g := domain.NewTroggu(trogguTable(), domain.TrogguConfig{TargetDeals: 1})
	g.Reset()

	seen := map[domain.TrogguPhase]bool{}
	for range 5000 {
		if g.GetGameEndFlag() {
			break
		}
		phase := g.GetPhase()
		if !seen[phase] {
			seen[phase] = true
			switch phase {
			case domain.TrogguPhaseBid:
				// 案内は今の最高入札で変わるので、テンプレートではなく展開後と比べる (#5808)。
				assert.Contains(t, p.Output(g, nil),
					i18n.Tf("troggu.promptBidHelp", "bids", "trois|solo|piccolo|misere"))
			case domain.TrogguPhasePlay:
				assert.Contains(t, p.Output(g, nil), i18n.T("troggu.promptPlayHelp"))
			case domain.TrogguPhaseTrickEnd:
				assert.Contains(t, p.Output(g, nil), i18n.T("troggu.promptTrickEndHelp"))
			case domain.TrogguPhaseRoundEnd:
				out := p.Output(g, nil)
				assert.Contains(t, out, i18n.T("troggu.promptRoundEndHelp"))
				bd := g.GetBreakdown()
				if bd == nil {
					assert.Contains(t, out, i18n.T("troggu.roundThrownIn"))
					break
				}
				// トリック目標の契約なら「トリック」、ソロなら「点」で語る。
				if bd.TargetIsTricks {
					assert.Contains(t, out, "トリック")
					break
				}
				assert.Contains(t, out, "点")
			}
		}
		trogguStepGame(t, g)
	}
	require.True(t, g.GetGameEndFlag())
	out := p.Output(g, nil)
	assert.True(t,
		strings.Contains(out, "の勝ち") || strings.Contains(out, "引き分け"),
		"終局の行が出ていない: %s", out)
}

func TestTrogguCuiPresenter_HintOutput(t *testing.T) {
	g := newTrogguGame()
	out := new(presenter.TrogguCuiPresenter).HintOutput(g)
	if g.IsHumanTurn() {
		assert.Contains(t, out, "HINT")
		assert.NotContains(t, out, "troggu.hintReason", "理由が訳されていない")
		return
	}
	assert.Contains(t, out, i18n.T("troggu.hintNone"))
}

func TestTrogguCuiPresenter_ActionLogOutput(t *testing.T) {
	g := newTrogguGame()
	assert.NotEmpty(t, new(presenter.TrogguCuiPresenter).ActionLogOutput(g))
}

// #5808: ミゼール(4) はソロ(2) より上なので、CPU がミゼールを宣言した配りでは
// 人間のソロは却下される。それでも案内は 4 契約を並べたままで、**打てるのに
// 弾かれる**ように見えていた。
func TestTrogguCuiPresenter_OffersOnlyTheBidsThatCanWin(t *testing.T) {
	p := new(presenter.TrogguCuiPresenter)

	t.Run("no bid yet offers all four", func(t *testing.T) {
		out := p.Output(newTrogguGame(), nil)
		for _, name := range []string{"trois", "solo", "piccolo", "misere"} {
			assert.Contains(t, out, name, "入札前なのに %s が案内されていない", name)
		}
	})

	// **CPU が先に宣言した局面。**issue #5808 の再現形そのもの: ミゼールが出て
	// いれば人間のソロは却下されるので、案内に残してはいけない。
	t.Run("a bid on the table hides everything it beats", func(t *testing.T) {
		names := map[domain.TrogguBid]string{
			domain.TrogguBidTrois: "trois", domain.TrogguBidSolo: "solo",
			domain.TrogguBidPiccolo: "piccolo", domain.TrogguBidMisere: "misere",
		}
		// 配りを引き直す。CPU が全員パスする配りでは検証にならない。
		var g *domain.Troggu
		for attempt := 0; attempt < 50 && g == nil; attempt++ {
			c := newTrogguGame()
			for i := 0; i < 20 && !c.IsHumanTurn() && c.GetPhase() == domain.TrogguPhaseBid; i++ {
				c.CpuBid()
			}
			if c.GetPhase() == domain.TrogguPhaseBid && c.GetHighestBid() >= domain.TrogguBidTrois {
				g = c
			}
		}
		require.NotNil(t, g, "CPU が入札する配りが 50 回で引けなかった")

		out := p.Output(g, nil)
		for bid, name := range names {
			if bid <= g.GetHighestBid() {
				assert.NotContains(t, out, name+"|", "%s は却下されるのに案内に残っている", name)
			} else {
				assert.Contains(t, out, name, "%s は打てるのに案内から消えている", name)
			}
		}
	})
}
