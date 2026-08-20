//go:build !js || !wasm || extra4

package domain

import "encoding/json"

// MightyPlayer マイティプレイヤークラス
type MightyPlayer struct {
	*GamePlayer
	RoundScoreHolder
	TrickHolder
	bid             int  // 宣言した得点札数 (-1 = 未ビッド, 0 = パス)
	bidNoTrump      bool // ノートランプ宣言かどうか
	isDeclarer      bool // 宣言者 (与党リーダー) かどうか
	isPartner       bool // 隠匿パートナー (フレンド) かどうか
	partnerRevealed bool // パートナーが全体に公開されたかどうか
	pointCards      int  // 獲得した得点札数 (今ラウンド: 10/J/Q/K/A)
}

// mightyPlayerJSON is the JSON wire format for MightyPlayer.
type mightyPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	TrickHolder      *TrickHolder      `json:"th"`
	Bid              int               `json:"bd"`
	BidNoTrump       bool              `json:"bn"`
	IsDeclarer       bool              `json:"id"`
	IsPartner        bool              `json:"ip"`
	PartnerRevealed  bool              `json:"pr"`
	PointCards       int               `json:"pc"`
}

// MarshalJSON implements json.Marshaler.
func (p *MightyPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(mightyPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		TrickHolder:      &p.TrickHolder,
		Bid:              p.bid,
		BidNoTrump:       p.bidNoTrump,
		IsDeclarer:       p.isDeclarer,
		IsPartner:        p.isPartner,
		PartnerRevealed:  p.partnerRevealed,
		PointCards:       p.pointCards,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *MightyPlayer) UnmarshalJSON(data []byte) error {
	var j mightyPlayerJSON
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
	p.bidNoTrump = j.BidNoTrump
	p.isDeclarer = j.IsDeclarer
	p.isPartner = j.IsPartner
	p.partnerRevealed = j.PartnerRevealed
	p.pointCards = j.PointCards
	return nil
}

// NewMightyPlayer コンストラクタ
func NewMightyPlayer(isHuman bool) *MightyPlayer {
	return &MightyPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		bid:        -1,
	}
}

// GetBid ビッド取得 (-1 = 未ビッド, 0 = パス)
func (p *MightyPlayer) GetBid() int { return p.bid }

// SetBid ビッド設定
func (p *MightyPlayer) SetBid(bid int) { p.bid = bid }

// GetBidNoTrump ノートランプ宣言かどうか
func (p *MightyPlayer) GetBidNoTrump() bool { return p.bidNoTrump }

// SetBidNoTrump ノートランプ宣言設定
func (p *MightyPlayer) SetBidNoTrump(v bool) { p.bidNoTrump = v }

// GetIsDeclarer 宣言者かどうか
func (p *MightyPlayer) GetIsDeclarer() bool { return p.isDeclarer }

// SetIsDeclarer 宣言者設定
func (p *MightyPlayer) SetIsDeclarer(v bool) { p.isDeclarer = v }

// GetIsPartner 隠匿パートナーかどうか
func (p *MightyPlayer) GetIsPartner() bool { return p.isPartner }

// SetIsPartner 隠匿パートナー設定
func (p *MightyPlayer) SetIsPartner(v bool) { p.isPartner = v }

// GetPartnerRevealed パートナー公開状態
func (p *MightyPlayer) GetPartnerRevealed() bool { return p.partnerRevealed }

// SetPartnerRevealed パートナー公開状態設定
func (p *MightyPlayer) SetPartnerRevealed(v bool) { p.partnerRevealed = v }

// GetPointCards 獲得した得点札数
func (p *MightyPlayer) GetPointCards() int { return p.pointCards }

// SetPointCards 得点札数設定
func (p *MightyPlayer) SetPointCards(n int) { p.pointCards = n }

// ResetRound ラウンドをリセット（ビッド・トリック・手札・チーム状態を初期化）
func (p *MightyPlayer) ResetRound() {
	p.bid = -1
	p.bidNoTrump = false
	p.isDeclarer = false
	p.isPartner = false
	p.partnerRevealed = false
	p.pointCards = 0
	p.SetRoundScore(0)
	p.ResetTricks()
	p.Reset()
	p.SetIsFinished(false)
}
