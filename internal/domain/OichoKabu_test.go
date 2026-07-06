//go:build !js || !wasm || extra

package domain_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// oichokabuCard は指定した数字のカブ札を作るテストヘルパー（design はコピー番号1固定）。
func oichokabuCard(value int) *domain.Card {
	return domain.NewCard(1, value, true)
}

func TestNewDefaultOichoKabu(t *testing.T) {
	o := domain.NewDefaultOichoKabu()
	assert.Equal(t, domain.OichoKabuPhaseBet, o.GetPhase())
	assert.Equal(t, domain.OichoKabuDefaultChips, o.GetChips())
	assert.False(t, o.GetGameEndFlag())
	assert.Empty(t, o.GetPlayerHand())
	assert.Empty(t, o.GetBankerHand())
	assert.Equal(t, 0, o.GetBet())
}

func TestOichoKabu_Reset(t *testing.T) {
	o := domain.NewDefaultOichoKabu()
	require.NoError(t, o.Bet(100))
	o.SetPhase(domain.OichoKabuPhaseEnd)
	o.Reset()
	assert.Equal(t, domain.OichoKabuPhaseBet, o.GetPhase())
	assert.False(t, o.GetGameEndFlag())
	assert.Empty(t, o.GetPlayerHand())
	assert.Empty(t, o.GetBankerHand())
	assert.Equal(t, 0, o.GetBet())
	assert.Equal(t, 0, o.GetTotalPayout())
}

func TestOichoKabu_Reset_RefillChips(t *testing.T) {
	o := domain.NewDefaultOichoKabu()
	o.SetChips(5)
	o.Reset()
	assert.Equal(t, domain.OichoKabuDefaultChips, o.GetChips())
}

func TestOichoKabu_Reset_NoRefillAboveThreshold(t *testing.T) {
	o := domain.NewDefaultOichoKabu()
	o.SetChips(500)
	o.Reset()
	assert.Equal(t, 500, o.GetChips())
}

func TestOichoKabu_Bet_WrongPhase(t *testing.T) {
	o := domain.NewDefaultOichoKabu()
	o.SetPhase(domain.OichoKabuPhaseDraw)
	err := o.Bet(100)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestOichoKabu_Bet_InvalidAmount(t *testing.T) {
	for _, tt := range []struct {
		name   string
		amount int
	}{
		{"TooLow", 5},
		{"NotMultiple", 15},
		{"TooHigh", 20000},
		{"Zero", 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			o := domain.NewDefaultOichoKabu()
			err := o.Bet(tt.amount)
			assert.Error(t, err)
			assert.True(t, errors.Is(err, domain.ErrInvalidAmount))
		})
	}
}

func TestOichoKabu_Bet_InsufficientChips(t *testing.T) {
	o := domain.NewDefaultOichoKabu()
	o.SetChips(50)
	err := o.Bet(100)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInsufficientChips))
}

func TestOichoKabu_Bet_Success(t *testing.T) {
	o := domain.NewDefaultOichoKabu()
	require.NoError(t, o.Bet(100))
	assert.Equal(t, 100, o.GetBet())
	assert.Len(t, o.GetPlayerHand(), 2)
	assert.Len(t, o.GetBankerHand(), 2)
	assert.Equal(t, domain.OichoKabuPhaseDraw, o.GetPhase())
	assert.Equal(t, domain.OichoKabuDefaultChips-100, o.GetChips())
	assert.False(t, o.GetGameEndFlag())
}

func TestOichoKabu_Draw_WrongPhase(t *testing.T) {
	o := domain.NewDefaultOichoKabu()
	err := o.Draw()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

func TestOichoKabu_Draw_HandFull(t *testing.T) {
	o := domain.NewDefaultOichoKabu()
	o.SetPhase(domain.OichoKabuPhaseDraw)
	o.SetPlayerHand([]*domain.Card{oichokabuCard(1), oichokabuCard(2), oichokabuCard(3)})
	err := o.Draw()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrInvalidPlay))
}

func TestOichoKabu_Draw_NaturalFlow(t *testing.T) {
	o := domain.NewDefaultOichoKabu()
	require.NoError(t, o.Bet(100))
	require.NoError(t, o.Draw())
	assert.Len(t, o.GetPlayerHand(), 3)
	assert.Equal(t, domain.OichoKabuPhaseEnd, o.GetPhase())
	assert.True(t, o.GetGameEndFlag())
}

func TestOichoKabu_Stand_WrongPhase(t *testing.T) {
	o := domain.NewDefaultOichoKabu()
	err := o.Stand()
	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrWrongPhase))
}

// setupResolve は Draw フェーズにして子・親の手を固定する。
func setupResolve(playerVals, bankerVals []int) *domain.OichoKabu {
	o := domain.NewDefaultOichoKabu()
	o.SetPhase(domain.OichoKabuPhaseDraw)
	o.SetBet(100)
	ph := make([]*domain.Card, len(playerVals))
	for i, v := range playerVals {
		ph[i] = oichokabuCard(v)
	}
	bh := make([]*domain.Card, len(bankerVals))
	for i, v := range bankerVals {
		bh[i] = oichokabuCard(v)
	}
	o.SetPlayerHand(ph)
	o.SetBankerHand(bh)
	return o
}

func TestOichoKabu_Stand_PlayerWins(t *testing.T) {
	o := setupResolve([]int{9}, []int{8}) // banker rank 8 > 6 → stands
	o.SetChips(900)
	require.NoError(t, o.Stand())
	assert.Equal(t, domain.OichoKabuResultWin, o.GetResult())
	assert.Equal(t, domain.OichoKabuPhaseEnd, o.GetPhase())
	assert.True(t, o.GetGameEndFlag())
	assert.Len(t, o.GetBankerHand(), 1) // no draw
	assert.Equal(t, 200, o.GetTotalPayout())
	assert.Equal(t, 900+200, o.GetChips())
}

func TestOichoKabu_Stand_BankerWins(t *testing.T) {
	o := setupResolve([]int{7}, []int{8}) // banker 8 stands, player 7 loses
	o.SetChips(900)
	require.NoError(t, o.Stand())
	assert.Equal(t, domain.OichoKabuResultLose, o.GetResult())
	assert.Equal(t, 0, o.GetTotalPayout())
	assert.Equal(t, 900, o.GetChips())
}

func TestOichoKabu_Stand_Push(t *testing.T) {
	o := setupResolve([]int{8}, []int{8}) // both rank 8, banker stands → push
	o.SetChips(900)
	require.NoError(t, o.Stand())
	assert.Equal(t, domain.OichoKabuResultDraw, o.GetResult())
	assert.Equal(t, 100, o.GetTotalPayout()) // bet returned
	assert.Equal(t, 900+100, o.GetChips())
}

func TestOichoKabu_Stand_BankerDrawsWhenRankLow(t *testing.T) {
	o := setupResolve([]int{5}, []int{6})       // banker rank 6 ≤ 6 → draws
	o.SetDeck([]*domain.Card{oichokabuCard(3)}) // banker draws a 3 → rank 9
	require.NoError(t, o.Stand())
	assert.Len(t, o.GetBankerHand(), 2)
	assert.Equal(t, 9, o.GetBankerRank())
	assert.Equal(t, domain.OichoKabuResultLose, o.GetResult()) // player 5 < banker 9
}

func TestOichoKabu_Stand_BankerStandsWhenRankHigh(t *testing.T) {
	o := setupResolve([]int{9}, []int{7}) // banker rank 7 > 6 → stands
	o.SetDeck([]*domain.Card{oichokabuCard(1)})
	require.NoError(t, o.Stand())
	assert.Len(t, o.GetBankerHand(), 1) // no draw despite deck available
	assert.Equal(t, 7, o.GetBankerRank())
	assert.Equal(t, domain.OichoKabuResultWin, o.GetResult())
}

func TestOichoKabu_TenScoresZero(t *testing.T) {
	// 10 + 9 → (0 + 9) % 10 = 9 (kabu), the best hand.
	o := setupResolve([]int{10, 9}, []int{8})
	require.NoError(t, o.Stand())
	assert.Equal(t, 9, o.GetPlayerRank())
	assert.Equal(t, domain.OichoKabuResultWin, o.GetResult())
}

func TestOichoKabu_RankWrapsAroundModulo(t *testing.T) {
	// 7 + 6 = 13 → 3.  Buta (0) example: 4 + 6 = 10 → 0.
	assert.Equal(t, 3, setupResolveRank([]int{7, 6}))
	assert.Equal(t, 0, setupResolveRank([]int{4, 6}))
}

func setupResolveRank(vals []int) int {
	o := setupResolve(vals, []int{8})
	return o.GetPlayerRank()
}

func TestOichoKabu_GetActionLog(t *testing.T) {
	o := domain.NewDefaultOichoKabu()
	require.NoError(t, o.Bet(100))
	assert.NotEmpty(t, o.GetActionLog())
}

func TestOichoKabu_JSONRoundTrip(t *testing.T) {
	o := setupResolve([]int{9}, []int{8})
	require.NoError(t, o.Stand())

	data, err := json.Marshal(o)
	require.NoError(t, err)
	var restored domain.OichoKabu
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, o.GetPhase(), restored.GetPhase())
	assert.Equal(t, o.GetChips(), restored.GetChips())
	assert.Equal(t, o.GetBet(), restored.GetBet())
	assert.Equal(t, o.GetResult(), restored.GetResult())
	assert.Equal(t, o.GetTotalPayout(), restored.GetTotalPayout())
	assert.Equal(t, o.GetGameEndFlag(), restored.GetGameEndFlag())
	assert.Equal(t, o.GetPlayerRank(), restored.GetPlayerRank())
	assert.Equal(t, o.GetBankerRank(), restored.GetBankerRank())
}

func TestOichoKabu_UnmarshalJSON_InvalidData(t *testing.T) {
	var o domain.OichoKabu
	err := o.UnmarshalJSON([]byte("not json"))
	assert.Error(t, err)
}

func TestOichoKabu_UnmarshalJSON_NilFields(t *testing.T) {
	data := []byte(`{"dk":null,"ph":null,"bh":null,"ch":null,"bt":0,"ps":1,"ge":false,"gr":0,"tp":0,"al":null}`)
	var o domain.OichoKabu
	require.NoError(t, json.Unmarshal(data, &o))
	assert.NotNil(t, o.GetActionLog())
	assert.Empty(t, o.GetPlayerHand())
}

func TestOichoKabu_UnmarshalJSON_HandTooLong(t *testing.T) {
	data := []byte(`{"ph":[{"d":1,"v":1},{"d":1,"v":2},{"d":1,"v":3},{"d":1,"v":4}],"ps":1}`)
	var o domain.OichoKabu
	err := json.Unmarshal(data, &o)
	assert.Error(t, err)
}

func TestOichoKabu_UnmarshalJSON_NilCardElement(t *testing.T) {
	data := []byte(`{"ph":[null],"ps":1}`)
	var o domain.OichoKabu
	err := json.Unmarshal(data, &o)
	assert.Error(t, err)
}
