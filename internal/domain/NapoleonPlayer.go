//go:build !js || !wasm || extra4

package domain

import "encoding/json"

// NapoleonPlayer ナポレオンプレイヤークラス
type NapoleonPlayer struct {
	*GamePlayer
	RoundScoreHolder
	TrickHolder
	bid              int  // 宣言した絵札数 (-1 = 未ビッド, 0 = パス)
	isNapoleon       bool // ナポレオンかどうか
	isAdjutant       bool // 副官かどうか
	adjutantRevealed bool // 副官が公開されたかどうか
	pictureCards     int  // 獲得した絵札数 (今ラウンド)
}

// napoleonPlayerJSON is the JSON wire format for NapoleonPlayer.
type napoleonPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	TrickHolder      *TrickHolder      `json:"th"`
	Bid              int               `json:"bd"`
	IsNapoleon       bool              `json:"in"`
	IsAdjutant       bool              `json:"ia"`
	AdjutantRevealed bool              `json:"ar"`
	PictureCards     int               `json:"pc"`
}

// MarshalJSON implements json.Marshaler.
func (p *NapoleonPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(napoleonPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		TrickHolder:      &p.TrickHolder,
		Bid:              p.bid,
		IsNapoleon:       p.isNapoleon,
		IsAdjutant:       p.isAdjutant,
		AdjutantRevealed: p.adjutantRevealed,
		PictureCards:     p.pictureCards,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *NapoleonPlayer) UnmarshalJSON(data []byte) error {
	var j napoleonPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	if j.RoundScoreHolder != nil {
		p.RoundScoreHolder = *j.RoundScoreHolder
	}
	if j.TrickHolder != nil {
		p.TrickHolder = *j.TrickHolder
	}
	p.bid = j.Bid
	p.isNapoleon = j.IsNapoleon
	p.isAdjutant = j.IsAdjutant
	p.adjutantRevealed = j.AdjutantRevealed
	p.pictureCards = j.PictureCards
	return nil
}

// NewNapoleonPlayer コンストラクタ
func NewNapoleonPlayer(isHuman bool) *NapoleonPlayer {
	return &NapoleonPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		bid:        -1,
	}
}

// GetBid ビッド取得 (-1 = 未ビッド, 0 = パス)
func (p *NapoleonPlayer) GetBid() int { return p.bid }

// SetBid ビッド設定
func (p *NapoleonPlayer) SetBid(bid int) { p.bid = bid }

// GetIsNapoleon ナポレオンかどうか
func (p *NapoleonPlayer) GetIsNapoleon() bool { return p.isNapoleon }

// SetIsNapoleon ナポレオン設定
func (p *NapoleonPlayer) SetIsNapoleon(v bool) { p.isNapoleon = v }

// GetIsAdjutant 副官かどうか
func (p *NapoleonPlayer) GetIsAdjutant() bool { return p.isAdjutant }

// SetIsAdjutant 副官設定
func (p *NapoleonPlayer) SetIsAdjutant(v bool) { p.isAdjutant = v }

// GetAdjutantRevealed 副官公開状態
func (p *NapoleonPlayer) GetAdjutantRevealed() bool { return p.adjutantRevealed }

// SetAdjutantRevealed 副官公開状態設定
func (p *NapoleonPlayer) SetAdjutantRevealed(v bool) { p.adjutantRevealed = v }

// GetPictureCards 獲得した絵札数
func (p *NapoleonPlayer) GetPictureCards() int { return p.pictureCards }

// SetPictureCards 絵札数設定
func (p *NapoleonPlayer) SetPictureCards(n int) { p.pictureCards = n }

// ResetRound ラウンドをリセット（ビッド・トリック・手札・チーム状態を初期化）
func (p *NapoleonPlayer) ResetRound() {
	p.bid = -1
	p.isNapoleon = false
	p.isAdjutant = false
	p.adjutantRevealed = false
	p.pictureCards = 0
	p.SetRoundScore(0)
	p.ResetTricks()
	p.Reset()
	p.SetIsFinished(false)
}
