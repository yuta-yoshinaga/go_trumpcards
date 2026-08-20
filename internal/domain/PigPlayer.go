//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// PigPlayer はピッグのプレイヤー。
//
// 手札に加えて、溜まった文字数 (letters)、脱落フラグ (eliminated)、
// 当該ラウンドで合図に気づいたか (hasSignalled) を持つ。
type PigPlayer struct {
	*GamePlayer
	letters      int
	eliminated   bool
	hasSignalled bool
	// noticedOrder は合図に気づいた順 (1 始まり、0 = まだ)。
	//
	// **順番そのものが罰の根拠。** 最後の 1 人だけが文字を受け取るので、
	// 「気づいたかどうか」だけでは誰が最後だったか復元できません。
	noticedOrder int
}

// NewPigPlayer はコンストラクタ。
func NewPigPlayer(isHuman bool) *PigPlayer {
	return &PigPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// GetLetters は溜まった文字数を返す。
func (p *PigPlayer) GetLetters() int { return p.letters }

// SetLetters は文字数を設定する (主にテスト/復元用)。
func (p *PigPlayer) SetLetters(n int) { p.letters = n }

// AddLetter は文字を 1 つ加える。PigMaxLetters に達したら脱落させ true を返す。
func (p *PigPlayer) AddLetter() bool {
	p.letters++
	if p.letters >= PigMaxLetters {
		p.letters = PigMaxLetters
		p.SetEliminated(true)
		return true
	}
	return false
}

// GetEliminated は脱落済みかどうかを返す。
func (p *PigPlayer) GetEliminated() bool { return p.eliminated }

// SetEliminated は脱落状態を設定する (主にテスト/復元用)。
func (p *PigPlayer) SetEliminated(v bool) {
	p.eliminated = v
	p.SetIsFinished(v)
}

// GetHasSignalled は当該ラウンドで合図に気づいたかを返す。
func (p *PigPlayer) GetHasSignalled() bool { return p.hasSignalled }

// SetHasSignalled は合図の状態を設定する。
func (p *PigPlayer) SetHasSignalled(v bool) { p.hasSignalled = v }

// GetNoticedOrder は合図に気づいた順 (1 始まり、0 = まだ) を返す。
func (p *PigPlayer) GetNoticedOrder() int { return p.noticedOrder }

// SetNoticedOrder は気づいた順を設定する。
func (p *PigPlayer) SetNoticedOrder(n int) { p.noticedOrder = n }

// PigLetterTargetWord は脱落までに溜まる語。**3 文字で脱落**という規則そのもの
// なので、画面・CUI・レスポンスはこの 1 箇所を参照します (#5766)。
const PigLetterTargetWord = "PIG"

// GetLetterWord は溜まった文字を "PIG" から切り出して返す。
func (p *PigPlayer) GetLetterWord() string {
	const word = PigLetterTargetWord
	n := p.letters
	if n < 0 {
		n = 0
	}
	if n > len(word) {
		n = len(word)
	}
	return word[:n]
}

// HasFourOfAKind は手札 4 枚がすべて同じランクかどうかを返す。
func (p *PigPlayer) HasFourOfAKind() bool {
	if p.GetCardsSize() < PigHandSize {
		return false
	}
	first := p.GetCard(0)
	if first == nil {
		return false
	}
	v := first.GetValue()
	for i := 1; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c == nil || c.GetValue() != v {
			return false
		}
	}
	return true
}

// pigPlayerJSON is the JSON wire format for PigPlayer.
type pigPlayerJSON struct {
	GamePlayer   *GamePlayer `json:"gp"`
	Letters      int         `json:"lt"`
	Eliminated   bool        `json:"el"`
	HasSignalled bool        `json:"hs"`
	NoticedOrder int         `json:"no"`
}

// MarshalJSON implements json.Marshaler.
func (p *PigPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(pigPlayerJSON{
		GamePlayer:   p.GamePlayer,
		Letters:      p.letters,
		Eliminated:   p.eliminated,
		HasSignalled: p.hasSignalled,
		NoticedOrder: p.noticedOrder,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **クランプせずに拒否します。** 文字数を黙って丸めると、「脱落しているのに
// 文字が 2 つ」のような盤面がそのまま復元され、以降ずっと辻褄が合いません。
func (p *PigPlayer) UnmarshalJSON(data []byte) error {
	var j pigPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer == nil {
		return errors.New("pig player is missing its base player")
	}
	if j.Letters < 0 || j.Letters > PigMaxLetters {
		return fmt.Errorf("letters must be between 0 and %d, got %d", PigMaxLetters, j.Letters)
	}
	// **脱落と文字数は同じ事実の裏表。** 片方だけ書き換わった状態を受けない。
	if j.Eliminated != (j.Letters == PigMaxLetters) {
		return fmt.Errorf("a seat is eliminated exactly when it holds %d letters (eliminated=%v, letters=%d)",
			PigMaxLetters, j.Eliminated, j.Letters)
	}
	if j.NoticedOrder < 0 || j.NoticedOrder > PigPlayerCntMax {
		return fmt.Errorf("noticed order must be between 0 and %d, got %d", PigPlayerCntMax, j.NoticedOrder)
	}
	// **気づいた順は「気づいた」ことの内訳。** 順位だけあって未通知はありえない。
	if (j.NoticedOrder > 0) != j.HasSignalled {
		return fmt.Errorf("a seat has a notice order exactly when it has signalled (order=%d, signalled=%v)",
			j.NoticedOrder, j.HasSignalled)
	}
	p.GamePlayer = j.GamePlayer
	p.letters = j.Letters
	p.eliminated = j.Eliminated
	p.SetIsFinished(j.Eliminated)
	p.hasSignalled = j.HasSignalled
	p.noticedOrder = j.NoticedOrder
	return nil
}
