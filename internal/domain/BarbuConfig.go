//go:build !js || !wasm || extra4

package domain

import "encoding/json"

// BarbuCpuDifficulty は CPU の難易度レベル。
type BarbuCpuDifficulty int

// BarbuCpuDifficulty 定数
const (
	// BarbuDifficultyEasy 簡単 (ランダムな合法手・コントラクトをランダム選択)
	BarbuDifficultyEasy BarbuCpuDifficulty = 0
	// BarbuDifficultyNormal 通常 (貪欲: 失点を避ける / 得点を狙う)
	BarbuDifficultyNormal BarbuCpuDifficulty = 1
	// BarbuDifficultyHard 難しい (手札評価でコントラクト・プレイを最適化)
	BarbuDifficultyHard BarbuCpuDifficulty = 2
)

// BarbuDifficultyNames 難易度名マップ
var BarbuDifficultyNames = map[BarbuCpuDifficulty]string{
	BarbuDifficultyEasy:   "Easy",
	BarbuDifficultyNormal: "Normal",
	BarbuDifficultyHard:   "Hard",
}

// Barbu の各コントラクトの得点定数 (1 ディールあたり)。
// 負のコントラクト (失点を避ける) は減点、正のコントラクト (Trumps / Dominoes)
// は加点を与える。バルブはコンペンディウム型で、28 ディール後の累計が最も
// 高いプレイヤーが勝者となる。
const (
	// BarbuNoTrickPenalty No Tricks: 取ったトリック 1 つにつき減点。
	BarbuNoTrickPenalty = 2
	// BarbuHeartPenalty No Hearts: 取ったハート 1 枚につき減点。
	BarbuHeartPenalty = 2
	// BarbuQueenPenalty No Queens: 取った Q 1 枚につき減点。
	BarbuQueenPenalty = 6
	// BarbuKingHeartPenalty Barbu: K♥ (ひげおじさん) を取ると大幅減点。
	BarbuKingHeartPenalty = 20
	// BarbuLastTrickPenalty No Last Trick: 最後の (13 番目) トリックを取ると減点。
	BarbuLastTrickPenalty = 20
	// BarbuTrumpReward Trumps: 取ったトリック 1 つにつき加点。
	BarbuTrumpReward = 5
)

// BarbuDominoScores は Dominoes コントラクトの順位別得点。
// インデックス 0 = 1 着, 3 = 4 着。早く上がるほど高得点。
var BarbuDominoScores = [BarbuPlayerCnt]int{45, 20, -5, -30}

// BarbuConfig はバルブのローカルルール設定。
type BarbuConfig struct {
	// CpuDifficulty CPU 難易度
	CpuDifficulty BarbuCpuDifficulty
}

// DefaultBarbuConfig はデフォルトのローカルルール設定を返す。
// 問題 #1991 の仕様により:
//   - プレイヤー数: 4
//   - デッキ: 52 枚
//   - 計 28 ディール (各プレイヤーが 7 つのコントラクトを 1 回ずつ選択)
//   - CPU 難易度: 3 段階
func DefaultBarbuConfig() BarbuConfig {
	return BarbuConfig{
		CpuDifficulty: BarbuDifficultyNormal,
	}
}

// Validate は設定値のドメインバリデーションを行う。
func (c BarbuConfig) Validate() error {
	return ValidateRange(
		"CPU difficulty",
		int(c.CpuDifficulty),
		int(BarbuDifficultyEasy),
		int(BarbuDifficultyHard),
	)
}

// barbuConfigJSON is the JSON wire format for BarbuConfig.
type barbuConfigJSON struct {
	CpuDifficulty BarbuCpuDifficulty `json:"di"`
}

// MarshalJSON implements json.Marshaler.
func (c BarbuConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(barbuConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *BarbuConfig) UnmarshalJSON(data []byte) error {
	var j barbuConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	*c = BarbuConfig(j)
	return nil
}
