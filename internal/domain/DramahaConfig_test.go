package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultDramahaConfig(t *testing.T) {
	cfg := DefaultDramahaConfig()
	assert.Equal(t, 5, cfg.SmallBlind)
	assert.Equal(t, 10, cfg.BigBlind)
	assert.Equal(t, 1000, cfg.InitChips)
	assert.Equal(t, HoldemTableSize4, cfg.TableSize)
}

func TestNewDramahaPlayersForTable(t *testing.T) {
	t.Run("4-max", func(t *testing.T) {
		players := NewDramahaPlayersForTable(HoldemTableSize4)
		assert.Equal(t, 4, len(players))
		assert.True(t, players[0].GetIsHuman())
		for i := 1; i < 4; i++ {
			assert.False(t, players[i].GetIsHuman())
		}
	})
	// 6-max / 9-max はクローン元 (Omaha) では通る。ドラマハでは山が足りない:
	// 1 席が最悪 10 枚 (ホール 5 + 交換 5) 使い、ボードに 5 枚要るので 10N+5 枚
	// 必要で、6-max は 65 枚、9-max は 95 枚 —— どちらも 52 枚を超える。
	// 足りないまま配ると 9-max の 85% が 3 枚ボードでショーダウンし、15% が
	// ボードの添字で panic する (実測)。よって 4-max に丸める。
	t.Run("6-max falls back to 4-max: the deck cannot deal 65 cards", func(t *testing.T) {
		players := NewDramahaPlayersForTable(HoldemTableSize6)
		assert.Equal(t, 4, len(players))
		assert.True(t, players[0].GetIsHuman())
	})
	t.Run("9-max falls back to 4-max: the deck cannot deal 95 cards", func(t *testing.T) {
		players := NewDramahaPlayersForTable(HoldemTableSize9)
		assert.Equal(t, 4, len(players))
		assert.True(t, players[0].GetIsHuman())
	})
	// 上の丸めが効いていることを、山の枚数そのもので押さえる。席数×10+5 が
	// 52 を超えないこと —— これが破れた瞬間に上のバグが戻る。
	t.Run("every reachable table fits in one deck", func(t *testing.T) {
		for _, size := range []int{HoldemTableSize4, HoldemTableSize6, HoldemTableSize9, 5, 0, -1} {
			n := len(NewDramahaPlayersForTable(size))
			worst := n*(DramahaHoleCards*2) + 5
			assert.LessOrEqual(t, worst, 52,
				"table of %d needs %d cards worst-case (requested %d)", n, worst, size)
		}
	})
	t.Run("invalid falls back to 4-max", func(t *testing.T) {
		players := NewDramahaPlayersForTable(5)
		assert.Equal(t, 4, len(players))
		assert.True(t, players[0].GetIsHuman())
	})
}
