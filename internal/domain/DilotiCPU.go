//go:build !js || !wasm || classic

package domain

import "math/rand"

// DilotiMove は 1 手。
type DilotiMove struct {
	// HandIdx は出す手札。
	HandIdx int
	// Action は手の種類 (capture / declare / trail)。
	Action string
	// TableIdxs は巻き込む場札。
	TableIdxs []int
	// DeclIdxs は巻き込む宣言。
	DeclIdxs []int
	// Value は宣言値 (declare のときだけ)。
	Value int
	// Reason は勧める理由の識別子 (ヒント用)。
	Reason string
}

// DilotiHint は人間への推奨手。
type DilotiHint struct {
	// Move は勧める手 (打てる手が無ければ HandIdx が -1)。
	Move DilotiMove
}

// enumerateDilotiMoves は席 seat が打てる手をすべて挙げる。
func (d *Diloti) enumerateDilotiMoves(seat int) []DilotiMove {
	hand := d.players[seat].GetHand()
	out := make([]DilotiMove, 0, len(hand)*2)
	for i, c := range hand {
		for _, t := range EnumerateDilotiTakes(c, d.table, d.decls) {
			out = append(out, DilotiMove{
				HandIdx: i, Action: DilotiActionCapture,
				TableIdxs: t.TableIdxs, DeclIdxs: t.DeclIdxs, Reason: "capture",
			})
		}
	}
	// **宣言を抱えている間は、宣言を取る手しか選べない。** 取れる手が無ければ
	// 場に置くしかないが、それも規則で塞がれているので、抱えた宣言を取る手が
	// 常に残るよう手札の裏付けを守っている。
	if d.hasOutstandingDeclaration(seat) {
		return out
	}
	for i, c := range hand {
		for _, cand := range EnumerateDilotiDeclarations(c, i, hand, d.table) {
			out = append(out, DilotiMove{
				HandIdx: i, Action: DilotiActionDeclare,
				TableIdxs: cand.TableIdxs, Value: cand.Value, Reason: "declare",
			})
		}
	}
	for i, c := range hand {
		if CanTrailDiloti(c, d.table) {
			out = append(out, DilotiMove{HandIdx: i, Action: DilotiActionTrail, Reason: "trail"})
		}
	}
	return out
}

// dilotiLegalMoves は打てる手を返し、1 つも無ければ最後の逃げ道を作る。
//
// **手番が回ってきて打てる手が無い盤面を作らない。** 規則どおりなら、宣言を
// 抱えた席は裏付けの札を持ち続けるので必ず取れるし、絵札も同ランクが場に
// あれば取れる ── それでも 0 件になったら、盤面のほうが壊れているので
// 「置く」だけは通す。ここが無いと CUI も Web も入力待ちで固まる。
func (d *Diloti) dilotiLegalMoves(seat int) []DilotiMove {
	moves := d.enumerateDilotiMoves(seat)
	if len(moves) > 0 || d.players[seat].GetCardsSize() == 0 {
		return moves
	}
	for i := range d.players[seat].GetHand() {
		moves = append(moves, DilotiMove{HandIdx: i, Action: DilotiActionTrail, Reason: "trail"})
	}
	return moves
}

// dilotiMoveScore は 1 手の値打ちを見積もる。
//
// **クセリ 1 回で 10 点。** 固定点は全部で 11 点しかないので、場を 1 枚で
// 払える手はほぼ常に最善になる。次に効くのは点になる札 (A・10♦・2♣) と枚数。
func (d *Diloti) dilotiMoveScore(seat int, m DilotiMove) int {
	if m.Action == DilotiActionTrail {
		// 場に置くのは最後の手段。取られやすい大きな札ほど不利。
		return -DilotiCardValue(d.players[seat].GetHand()[m.HandIdx])
	}
	if m.Action == DilotiActionDeclare {
		// **宣言は約束であって得点ではない。** 値の大小で加点すると、目の前の
		// 捕獲を見送って大きい値を宣言し続ける ── 実測で ♦K の同ランク取りを
		// 捨てて 6 を宣言した。囲い込んだ場札の枚数ぶんだけ数える。
		return 2 + len(m.TableIdxs)
	}

	taken := make([]*Card, 0, len(m.TableIdxs)+1)
	for _, i := range m.TableIdxs {
		taken = append(taken, d.table[i])
	}
	for _, i := range m.DeclIdxs {
		taken = append(taken, d.decls[i].AllCards()...)
	}
	// 出した札も自分の取り札になるので 1 枚数える。枚数は最多枚数 4 点に
	// つながるので、宣言より重く見る。
	score := 2 * (len(taken) + 1)
	for _, c := range taken {
		switch {
		case c.GetValue() == 1:
			score += 4
		case c.GetValue() == 10 && c.GetDesign() == CardDesignDiamond:
			score += 6
		case c.GetValue() == 2 && c.GetDesign() == CardDesignClover:
			score += 3
		}
	}
	// 場を払い切ればクセリ。
	if len(m.TableIdxs) == len(d.table) && len(m.DeclIdxs) == len(d.decls) &&
		len(d.table)+len(d.decls) > 0 && d.firstPlayDone {
		score += 30
	}
	return score
}

// smartDilotiMove は最も値打ちのある手を選ぶ。
//
// **ヒントは難易度で鈍らせない。** CPU の難易度は CPU の腕であって、人間への
// 助言の質ではない ── ここを共有すると Easy の卓では助言が出鱈目になる。
func (d *Diloti) smartDilotiMove(seat int) *DilotiMove {
	moves := d.dilotiLegalMoves(seat)
	if len(moves) == 0 {
		return nil
	}
	best, bestScore := 0, d.dilotiMoveScore(seat, moves[0])
	for i := 1; i < len(moves); i++ {
		if s := d.dilotiMoveScore(seat, moves[i]); s > bestScore {
			best, bestScore = i, s
		}
	}
	m := moves[best]
	return &m
}

// chooseCpuMove は CPU の手を選ぶ。
func (d *Diloti) chooseCpuMove(seat int) *DilotiMove {
	if d.config.CpuDifficulty == DilotiCpuDifficultyEasy {
		moves := d.dilotiLegalMoves(seat)
		if len(moves) == 0 {
			return nil
		}
		m := moves[rand.Intn(len(moves))] //nolint:gosec // ゲームの手選びに暗号強度は要らない
		return &m
	}
	return d.smartDilotiMove(seat)
}

// GetHint は人間への推奨手を返す。
func (d *Diloti) GetHint() *DilotiHint {
	human := findHumanIdx(d.players)
	if human < 0 || d.gameEndFlag || d.phase != DilotiPhasePlay || d.currentIdx != human {
		return &DilotiHint{Move: DilotiMove{HandIdx: -1, Reason: "none"}}
	}
	m := d.smartDilotiMove(human)
	if m == nil {
		return &DilotiHint{Move: DilotiMove{HandIdx: -1, Reason: "none"}}
	}
	return &DilotiHint{Move: *m}
}
