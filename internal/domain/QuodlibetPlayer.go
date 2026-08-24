//go:build !js || !wasm || solo

package domain

import "encoding/json"

// QuodlibetPlayer はクオドリベットのプレイヤー。
//
// 基底の GamePlayer (手札) に加えて、現ディールで獲得したトリック
// (TrickHolder)、シェディング系コントラクトでの上がり順位、12 ディール越しの
// **罰点** 累計を持つ。**点は少ないほうが良い** ── 最後に一番少ない人が勝つ。
type QuodlibetPlayer struct {
	*GamePlayer
	*TrickHolder
	outRank    int // シェディング系での上がり順位 (1..4, 0 = 未上がり)
	penalty    int // 全 12 ディール通算の罰点
	dealPoints int // 直近ディールで負った罰点 (表示用)
}

// NewQuodlibetPlayer constructs a QuodlibetPlayer.
func NewQuodlibetPlayer(isHuman bool) *QuodlibetPlayer {
	return &QuodlibetPlayer{
		GamePlayer:  NewGamePlayer(isHuman),
		TrickHolder: &TrickHolder{},
	}
}

// GetOutRank はシェディング系コントラクトでの上がり順位を返す (0 = 未上がり)。
func (p *QuodlibetPlayer) GetOutRank() int { return p.outRank }

// SetOutRank は上がり順位を設定する。
func (p *QuodlibetPlayer) SetOutRank(r int) { p.outRank = r }

// GetPenalty は累計罰点を返す (12 ディール通算)。
func (p *QuodlibetPlayer) GetPenalty() int { return p.penalty }

// AddPenalty は罰点を加算する。
func (p *QuodlibetPlayer) AddPenalty(n int) { p.penalty += n }

// ResetPenalty は累計罰点を 0 に戻す (新規ゲーム開始時)。
func (p *QuodlibetPlayer) ResetPenalty() { p.penalty = 0 }

// GetDealPoints は直近ディールで負った罰点を返す。
func (p *QuodlibetPlayer) GetDealPoints() int { return p.dealPoints }

// SetDealPoints は直近ディールで負った罰点を設定する。
func (p *QuodlibetPlayer) SetDealPoints(n int) { p.dealPoints = n }

// ResetDeal は 1 ディール分の状態をクリアする。累計罰点は維持する。
func (p *QuodlibetPlayer) ResetDeal() {
	p.Reset()
	p.ResetTricks()
	p.SetIsFinished(false)
	p.outRank = 0
	p.dealPoints = 0
}

// quodlibetPlayerJSON is the JSON wire format for QuodlibetPlayer.
type quodlibetPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	OutRank     int          `json:"or"`
	Penalty     int          `json:"pn"`
	DealPoints  int          `json:"dp"`
}

// MarshalJSON implements json.Marshaler.
func (p *QuodlibetPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(quodlibetPlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: p.TrickHolder,
		OutRank:     p.outRank,
		Penalty:     p.penalty,
		DealPoints:  p.dealPoints,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *QuodlibetPlayer) UnmarshalJSON(data []byte) error {
	var j quodlibetPlayerJSON
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
	p.outRank = j.OutRank
	p.penalty = j.Penalty
	p.dealPoints = j.DealPoints
	return nil
}
