//go:build !js || !wasm || extra4

package domain

import "encoding/json"

// PiquetPlayer Piquetプレイヤークラス
//
// 2人用のディール式トリックテイキング (Elder vs Younger) で使う。
//   - declScore: 宣言フェーズで獲得した得点 (Point/Sequence/Set/CarteBlanche)
//   - trickScore: プレイフェーズで獲得した得点 (リード/トリック獲得/最終トリック)
//   - bonusScore: ラウンド末ボーナス (cards/capot/repique/pique)
//   - matchScore: パルティ通算スコア
type PiquetPlayer struct {
	*GamePlayer
	TrickHolder
	declScore  int
	trickScore int
	bonusScore int
	matchScore int
}

// NewPiquetPlayer コンストラクタ
func NewPiquetPlayer(isHuman bool) *PiquetPlayer {
	return &PiquetPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// GetDeclScore 宣言フェーズで獲得した得点を取得
func (p *PiquetPlayer) GetDeclScore() int { return p.declScore }

// SetDeclScore 宣言得点を設定
func (p *PiquetPlayer) SetDeclScore(n int) { p.declScore = n }

// AddDeclScore 宣言得点を加算
func (p *PiquetPlayer) AddDeclScore(n int) { p.declScore += n }

// GetTrickScore トリック得点を取得
func (p *PiquetPlayer) GetTrickScore() int { return p.trickScore }

// SetTrickScore トリック得点を設定
func (p *PiquetPlayer) SetTrickScore(n int) { p.trickScore = n }

// AddTrickScore トリック得点を加算
func (p *PiquetPlayer) AddTrickScore(n int) { p.trickScore += n }

// GetBonusScore ラウンド末ボーナスを取得
func (p *PiquetPlayer) GetBonusScore() int { return p.bonusScore }

// SetBonusScore ラウンド末ボーナスを設定
func (p *PiquetPlayer) SetBonusScore(n int) { p.bonusScore = n }

// AddBonusScore ラウンド末ボーナスを加算
func (p *PiquetPlayer) AddBonusScore(n int) { p.bonusScore += n }

// GetRoundScore このラウンドの合計得点 (decl+trick+bonus) を取得
func (p *PiquetPlayer) GetRoundScore() int {
	return p.declScore + p.trickScore + p.bonusScore
}

// GetMatchScore パルティ通算スコアを取得
func (p *PiquetPlayer) GetMatchScore() int { return p.matchScore }

// SetMatchScore パルティ通算スコアを設定
func (p *PiquetPlayer) SetMatchScore(n int) { p.matchScore = n }

// AddMatchScore パルティ通算スコアを加算
func (p *PiquetPlayer) AddMatchScore(n int) { p.matchScore += n }

// ResetRound ラウンドリセット (手札・トリック・ラウンドスコアを初期化、matchScore は保持)
func (p *PiquetPlayer) ResetRound() {
	p.ResetTricks()
	p.Reset()
	p.SetIsFinished(false)
	p.declScore = 0
	p.trickScore = 0
	p.bonusScore = 0
}

// ResetMatch パルティ全体をリセット (matchScore も初期化)
func (p *PiquetPlayer) ResetMatch() {
	p.ResetRound()
	p.matchScore = 0
}

// piquetPlayerJSON JSON wire format
type piquetPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	DeclScore   int          `json:"ds"`
	TrickScore  int          `json:"ts"`
	BonusScore  int          `json:"bs"`
	MatchScore  int          `json:"ms"`
}

// MarshalJSON implements json.Marshaler.
func (p *PiquetPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(piquetPlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
		DeclScore:   p.declScore,
		TrickScore:  p.trickScore,
		BonusScore:  p.bonusScore,
		MatchScore:  p.matchScore,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *PiquetPlayer) UnmarshalJSON(data []byte) error {
	var j piquetPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	if j.TrickHolder != nil {
		p.TrickHolder = *j.TrickHolder
	}
	p.declScore = j.DeclScore
	p.trickScore = j.TrickScore
	p.bonusScore = j.BonusScore
	p.matchScore = j.MatchScore
	return nil
}
