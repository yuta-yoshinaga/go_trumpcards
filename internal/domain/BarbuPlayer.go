//go:build !js || !wasm || extra4

package domain

import "encoding/json"

// BarbuPlayer はバルブのプレイヤー。
// 基底の GamePlayer (手札) に加えて、現ディールで獲得したトリック (TrickHolder)、
// Dominoes コントラクトでの上がり順位、28 ディール越しの累計得点を持つ。
type BarbuPlayer struct {
	*GamePlayer
	*TrickHolder
	dominoRank int // Dominoes での上がり順位 (1..4, 0 = 未上がり)
	totalScore int // 全 28 ディール通算の累計得点
}

// NewBarbuPlayer constructs a BarbuPlayer.
func NewBarbuPlayer(isHuman bool) *BarbuPlayer {
	return &BarbuPlayer{
		GamePlayer:  NewGamePlayer(isHuman),
		TrickHolder: &TrickHolder{},
		dominoRank:  0,
		totalScore:  0,
	}
}

// GetDominoRank は Dominoes での上がり順位を返す (0 = 未上がり)。
func (p *BarbuPlayer) GetDominoRank() int { return p.dominoRank }

// SetDominoRank は Dominoes での上がり順位を設定する。
func (p *BarbuPlayer) SetDominoRank(r int) { p.dominoRank = r }

// GetTotalScore は累計得点を返す (28 ディール通算)。
func (p *BarbuPlayer) GetTotalScore() int { return p.totalScore }

// AddScore は得点を加算する (負の値で減点)。
func (p *BarbuPlayer) AddScore(n int) { p.totalScore += n }

// ResetTotalScore は累計得点を 0 に戻す (新規ゲーム開始時)。
func (p *BarbuPlayer) ResetTotalScore() { p.totalScore = 0 }

// ResetDeal は 1 ディール分の状態 (手札・トリック・上がり順位・上がりフラグ) を
// クリアする。累計得点は維持する。
func (p *BarbuPlayer) ResetDeal() {
	p.Reset()
	p.ResetTricks()
	p.SetIsFinished(false)
	p.dominoRank = 0
}

// CapturedHearts は獲得トリック中のハート枚数を返す。
func (p *BarbuPlayer) CapturedHearts() int {
	cnt := 0
	for _, trick := range p.GetTricksTaken() {
		for _, c := range trick {
			if c.GetDesign() == CardDesignHeart {
				cnt++
			}
		}
	}
	return cnt
}

// CapturedQueens は獲得トリック中の Q (value 12) 枚数を返す。
func (p *BarbuPlayer) CapturedQueens() int {
	cnt := 0
	for _, trick := range p.GetTricksTaken() {
		for _, c := range trick {
			if c.GetValue() == 12 {
				cnt++
			}
		}
	}
	return cnt
}

// HasKingOfHearts は獲得トリック中に K♥ が含まれるかを返す。
func (p *BarbuPlayer) HasKingOfHearts() bool {
	for _, trick := range p.GetTricksTaken() {
		for _, c := range trick {
			if c.GetDesign() == CardDesignHeart && c.GetValue() == 13 {
				return true
			}
		}
	}
	return false
}

// barbuPlayerJSON is the JSON wire format for BarbuPlayer.
type barbuPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	DominoRank  int          `json:"dr"`
	TotalScore  int          `json:"ts"`
}

// MarshalJSON implements json.Marshaler.
func (p *BarbuPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(barbuPlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: p.TrickHolder,
		DominoRank:  p.dominoRank,
		TotalScore:  p.totalScore,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *BarbuPlayer) UnmarshalJSON(data []byte) error {
	var j barbuPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	if j.TrickHolder != nil {
		p.TrickHolder = j.TrickHolder
	} else {
		p.TrickHolder = &TrickHolder{}
	}
	p.dominoRank = j.DominoRank
	p.totalScore = j.TotalScore
	return nil
}
