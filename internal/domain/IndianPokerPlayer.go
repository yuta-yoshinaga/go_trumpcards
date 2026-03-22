package domain

// IndianPokerPlayer インディアンポーカープレイヤークラス
type IndianPokerPlayer struct {
	Player                            // 親クラス
	ChipHolder                        // チップ管理
	bettingPlayerBase                 // ベッティング共通状態
	isHuman           bool            // 人間フラグ
	playStyle         HoldemPlayStyle // CPUプレイスタイル
}

// NewIndianPokerPlayer コンストラクタ
func NewIndianPokerPlayer(isHuman bool, style HoldemPlayStyle) *IndianPokerPlayer {
	return &IndianPokerPlayer{
		Player:    Player{cards: make([]*Card, 0)},
		isHuman:   isHuman,
		playStyle: style,
	}
}

// GetIsHuman 人間フラグ取得
func (p *IndianPokerPlayer) GetIsHuman() bool { return p.isHuman }

// GetPlayStyle プレイスタイル取得
func (p *IndianPokerPlayer) GetPlayStyle() HoldemPlayStyle { return p.playStyle }

// GetPlayStyleName プレイスタイル名取得
func (p *IndianPokerPlayer) GetPlayStyleName() string {
	return playStyleName(int(p.playStyle), HoldemPlayStyleNames)
}

// GetComparisonCards ハンド比較用カード取得 (BettingPlayerインターフェース)
func (p *IndianPokerPlayer) GetComparisonCards() []*Card {
	if p.GetCardsSize() == 0 {
		return nil
	}
	return []*Card{p.GetCard(0)}
}
