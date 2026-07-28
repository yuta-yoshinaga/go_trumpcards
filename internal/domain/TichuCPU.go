//go:build !js || !wasm || extra2

package domain

// cpuDeclare CPUの宣言を決める (保守的・決定的)。
// 龍を持ち、かつエース(ランク14)を2枚以上持つ強い手のときだけティチューを宣言する。
func (t *Tichu) cpuDeclare(idx int) int {
	player := t.players[idx]
	hasDragon := false
	aces := 0
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if tichuSpecialKind(c) == TichuDragon {
			hasDragon = true
		}
		if tichuSpecialKind(c) == 0 && tichuRank(c) == 14 {
			aces++
		}
	}
	if hasDragon && aces >= 2 && t.config.CpuDifficulty != TichuDifficultyEasy {
		return TichuDeclTichu
	}
	return TichuDeclNone
}

// cpuFindPlay CPUが出すカードのインデックスを返す (nil/空=パス)。
func (t *Tichu) cpuFindPlay(player *TichuPlayer) []int {
	idx := t.round.currentTurn
	table := t.round.tableCombo
	if table == nil {
		return t.cpuLead(player)
	}
	// パートナーが場を支配しているなら譲る
	if t.round.lastPlayIdx == (idx+2)%TichuPlayerCnt {
		return nil
	}
	switch table.Type {
	case TichuComboSingle:
		if r := t.cpuBeatSingle(player, table); r != nil {
			return r
		}
		return t.cpuTryBomb(player, table)
	case TichuComboPair:
		if r := t.cpuBeatNofKind(player, table, 2); r != nil {
			return r
		}
		return t.cpuTryBomb(player, table)
	case TichuComboTriple:
		if r := t.cpuBeatNofKind(player, table, 3); r != nil {
			return r
		}
		return t.cpuTryBomb(player, table)
	default:
		return t.cpuTryBomb(player, table)
	}
}

// cpuLead リード時は最弱の単体を出す (確実に手を進める)。
func (t *Tichu) cpuLead(player *TichuPlayer) []int {
	if player.GetCardsSize() == 0 {
		return nil
	}
	return []int{0}
}

// cpuBeatSingle 場の単体を上回る最弱の単体を探す。
func (t *Tichu) cpuBeatSingle(player *TichuPlayer, table *TichuCombo) []int {
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		if tichuSpecialKind(c) == TichuDog {
			continue
		}
		combo := ClassifyTichu([]*Card{c})
		if combo == nil {
			continue
		}
		if combo.PhoenixSingle {
			combo.Rank = table.Rank
		}
		if TichuCanBeat(combo, table) {
			return []int{i}
		}
	}
	return nil
}

// cpuHandByRank ランク別の手札インデックスと鳳凰位置を返す。
func (t *Tichu) cpuHandByRank(player *TichuPlayer) (map[int][]int, int) {
	m := make(map[int][]int)
	phoenixIdx := -1
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		switch tichuSpecialKind(c) {
		case TichuPhoenix:
			phoenixIdx = i
		case TichuDragon, TichuDog:
		default:
			r := tichuRank(c)
			if r >= 2 && r <= 14 {
				m[r] = append(m[r], i)
			}
		}
	}
	return m, phoenixIdx
}

// cpuBeatNofKind 場の k枚同ランク役を上回る最弱の役を探す (鳳凰補完あり)。
func (t *Tichu) cpuBeatNofKind(player *TichuPlayer, table *TichuCombo, k int) []int {
	byRank, phoenixIdx := t.cpuHandByRank(player)
	// 自然な k枚を優先
	for r := table.Rank + 1; r <= 14; r++ {
		if len(byRank[r]) >= k {
			return byRank[r][:k]
		}
	}
	// 鳳凰で補完
	if phoenixIdx >= 0 {
		for r := table.Rank + 1; r <= 14; r++ {
			if len(byRank[r]) >= k-1 {
				res := append([]int{}, byRank[r][:k-1]...)
				res = append(res, phoenixIdx)
				return res
			}
		}
	}
	return nil
}

// cpuTryBomb 状況次第で4枚ボムを出す。
func (t *Tichu) cpuTryBomb(player *TichuPlayer, table *TichuCombo) []int {
	trickPoints := TichuCardsPoints(t.round.trickCards)
	if !tichuIsBomb(table) && trickPoints < 10 {
		return nil // 価値の低いトリックにはボムを使わない
	}
	byRank, _ := t.cpuHandByRank(player)
	for r := 2; r <= 14; r++ {
		if len(byRank[r]) == 4 {
			cards := []*Card{player.GetCard(byRank[r][0]), player.GetCard(byRank[r][1]),
				player.GetCard(byRank[r][2]), player.GetCard(byRank[r][3])}
			combo := ClassifyTichu(cards)
			if combo != nil && TichuCanBeat(combo, table) {
				return byRank[r]
			}
		}
	}
	return nil
}
