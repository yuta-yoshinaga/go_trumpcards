package domain

import (
	"math/rand"
	"sort"
)

// findBestPlay CPUが出すべきカードのインデックス配列を返す (空配列 = パス)
// 難易度により戦略を変える:
//   - Easy:   常に最弱の合法手を選ぶ
//   - Normal: 場がクリアなら最も多い枚数を出す、場があれば最弱で合法な手
//   - Hard:   残り手札が少ないほど強いカードを温存する
func (p *President) findBestPlay(player *PresidentPlayer) []int {
	switch p.config.CpuDifficulty {
	case PresidentDifficultyHard:
		return p.findBestPlayHard(player)
	case PresidentDifficultyEasy:
		return p.findWeakestLegalPlay(player)
	default:
		return p.findBestPlayNormal(player)
	}
}

// SuggestWeakestPlay は playerIdx の最弱の合法手 (手札インデックス) を返す。
// 合法手が無い (=パスすべき) 場合は nil を返す。CUI ヒント用に findWeakestLegalPlay
// を公開する薄いラッパー。
func (p *President) SuggestWeakestPlay(playerIdx int) []int {
	if playerIdx < 0 || playerIdx >= len(p.players) {
		return nil
	}
	return p.findWeakestLegalPlay(p.players[playerIdx])
}

// findWeakestLegalPlay 最弱の合法手を返す (Easy)
func (p *President) findWeakestLegalPlay(player *PresidentPlayer) []int {
	candidates := p.enumeratePlays(player)
	if len(candidates) == 0 {
		return nil
	}
	// 強さ順にソート、最弱を選ぶ
	sort.Slice(candidates, func(i, j int) bool {
		return p.cardStrength(player.GetCard(candidates[i][0]).GetValue()) <
			p.cardStrength(player.GetCard(candidates[j][0]).GetValue())
	})
	return candidates[0]
}

// findBestPlayNormal Normal難易度の戦略
// 場がクリアなら最も枚数の多い手を出す (手札を減らしやすい)
// 場があれば最弱の合法手を出す
func (p *President) findBestPlayNormal(player *PresidentPlayer) []int {
	if p.round.tableCards == nil {
		return p.findLargestGroupPlay(player)
	}
	return p.findWeakestLegalPlay(player)
}

// findBestPlayHard Hard難易度: 上がり直前なら強いカードを使う, 序盤なら温存
func (p *President) findBestPlayHard(player *PresidentPlayer) []int {
	candidates := p.enumeratePlays(player)
	if len(candidates) == 0 {
		return nil
	}
	handSize := player.GetCardsSize()
	// 残り手札3枚以下なら勝ちを急ぐ - 最強カードを使う
	if handSize <= 3 {
		sort.Slice(candidates, func(i, j int) bool {
			return p.cardStrength(player.GetCard(candidates[i][0]).GetValue()) >
				p.cardStrength(player.GetCard(candidates[j][0]).GetValue())
		})
		return candidates[0]
	}
	// 序盤は最弱を出す
	sort.Slice(candidates, func(i, j int) bool {
		return p.cardStrength(player.GetCard(candidates[i][0]).GetValue()) <
			p.cardStrength(player.GetCard(candidates[j][0]).GetValue())
	})
	return candidates[0]
}

// findLargestGroupPlay 場がクリアの時に最も大きなグループを選ぶ
func (p *President) findLargestGroupPlay(player *PresidentPlayer) []int {
	groups := p.groupByValue(player)
	if len(groups) == 0 {
		return nil
	}
	// 枚数の多い順、同枚数なら弱い順
	sort.Slice(groups, func(i, j int) bool {
		if len(groups[i]) != len(groups[j]) {
			return len(groups[i]) > len(groups[j])
		}
		return p.cardStrength(player.GetCard(groups[i][0]).GetValue()) <
			p.cardStrength(player.GetCard(groups[j][0]).GetValue())
	})
	// 革命発動の誘惑を少し入れる (4-of-a-kindが序盤にあるなら使わない確率)
	best := groups[0]
	if len(best) == 4 && player.GetCardsSize() > 6 {
		// 序盤は4-of-a-kindは温存 (革命を使わない)
		for _, g := range groups[1:] {
			if len(g) >= 1 {
				return g
			}
		}
	}
	return best
}

// enumeratePlays 現在の場に対して合法な手 (カードインデックスの組み合わせ) を列挙する
func (p *President) enumeratePlays(player *PresidentPlayer) [][]int {
	groups := p.groupByValue(player)
	plays := make([][]int, 0)

	for _, g := range groups {
		// 各グループから出せる枚数のサブセットを生成
		for size := 1; size <= len(g); size++ {
			// グループは同じ値なので、どの組み合わせでも強さは同じ
			// よって先頭 size 枚を代表として使う
			indices := make([]int, size)
			copy(indices, g[:size])
			cards := make([]*Card, size)
			for i, idx := range indices {
				cards[i] = player.GetCard(idx)
			}
			if p.isPlayable(cards) {
				plays = append(plays, indices)
			}
		}
	}

	// ランダム性を少し加える (同じ強さなら順序はシャッフル)
	rand.Shuffle(len(plays), func(i, j int) {
		plays[i], plays[j] = plays[j], plays[i]
	})
	return plays
}

// groupByValue 手札を値ごとにグループ化し、各グループのインデックス配列を返す
func (p *President) groupByValue(player *PresidentPlayer) [][]int {
	byValue := make(map[int][]int)
	for i := 0; i < player.GetCardsSize(); i++ {
		c := player.GetCard(i)
		byValue[c.GetValue()] = append(byValue[c.GetValue()], i)
	}
	groups := make([][]int, 0, len(byValue))
	for _, idxs := range byValue {
		groups = append(groups, idxs)
	}
	return groups
}
