//go:build !js || !wasm || extra

package domain

import "encoding/json"

// KingPlayer はキングのプレイヤー。基底の GamePlayer (手札) に加えて、現ディールで
// 獲得したトリック (TrickHolder)、各トリックのトリック番号 (No Last Two 判定用)、
// 全ディール通算の累計得点を持つ。
type KingPlayer struct {
	*GamePlayer
	*TrickHolder
	trickRanks []int // 獲得した各トリックのトリック番号 (1..13)
	totalScore int   // 全ディール通算の累計得点
}

// NewKingPlayer constructs a KingPlayer.
func NewKingPlayer(isHuman bool) *KingPlayer {
	return &KingPlayer{
		GamePlayer:  NewGamePlayer(isHuman),
		TrickHolder: &TrickHolder{},
		trickRanks:  make([]int, 0),
		totalScore:  0,
	}
}

// AddTrickWithRank は獲得トリックをトリック番号付きで追加する。
func (p *KingPlayer) AddTrickWithRank(cards []*Card, trickNumber int) {
	p.AddTrick(cards)
	p.trickRanks = append(p.trickRanks, trickNumber)
}

// GetTrickRanks は獲得した各トリックのトリック番号を返す。
func (p *KingPlayer) GetTrickRanks() []int { return p.trickRanks }

// GetTotalScore は累計得点を返す (全ディール通算)。
func (p *KingPlayer) GetTotalScore() int { return p.totalScore }

// AddScore は得点を加算する (負の値で減点)。
func (p *KingPlayer) AddScore(n int) { p.totalScore += n }

// ResetTotalScore は累計得点を 0 に戻す (新規ゲーム開始時)。
func (p *KingPlayer) ResetTotalScore() { p.totalScore = 0 }

// ResetDeal は 1 ディール分の状態 (手札・トリック・トリック番号) をクリアする。
// 累計得点は維持する。
func (p *KingPlayer) ResetDeal() {
	p.Reset()
	p.ResetTricks()
	p.SetIsFinished(false)
	p.trickRanks = make([]int, 0)
}

// CapturedHearts は獲得トリック中のハート枚数を返す。
func (p *KingPlayer) CapturedHearts() int {
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
func (p *KingPlayer) CapturedQueens() int {
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

// CapturedMen は獲得トリック中の J (11) / K (13) 枚数を返す。
func (p *KingPlayer) CapturedMen() int {
	cnt := 0
	for _, trick := range p.GetTricksTaken() {
		for _, c := range trick {
			if c.GetValue() == 11 || c.GetValue() == 13 {
				cnt++
			}
		}
	}
	return cnt
}

// HasKingOfHearts は獲得トリック中に K♥ が含まれるかを返す。
func (p *KingPlayer) HasKingOfHearts() bool {
	for _, trick := range p.GetTricksTaken() {
		for _, c := range trick {
			if c.GetDesign() == CardDesignHeart && c.GetValue() == 13 {
				return true
			}
		}
	}
	return false
}

// kingPlayerJSON is the JSON wire format for KingPlayer.
type kingPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	TrickRanks  []int        `json:"tr"`
	TotalScore  int          `json:"ts"`
}

// MarshalJSON implements json.Marshaler.
func (p *KingPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(kingPlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: p.TrickHolder,
		TrickRanks:  p.trickRanks,
		TotalScore:  p.totalScore,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *KingPlayer) UnmarshalJSON(data []byte) error {
	var j kingPlayerJSON
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
	if j.TrickRanks != nil {
		p.trickRanks = j.TrickRanks
	} else {
		p.trickRanks = make([]int, 0)
	}
	p.totalScore = j.TotalScore
	return nil
}
