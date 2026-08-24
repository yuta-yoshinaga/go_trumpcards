//go:build !js || !wasm || solo

package domain

// QuodlibetContracts.go はコントラクトごとの罰点計算 (Strategy パターン)。
//
// **点は少ないほうが良い。** 12 ディールを終えて累計の罰点が最も少ない人が
// 勝つので、ここが返すのは常に 0 以上の「負う点」である。
//
// 数字の典拠は 2 つ。多くは英語版 Wikipedia の Quodlibet と catsatcards の
// ルール解説が一致しているが、**1238 の最後のトリックだけ食い違う** ──
// Wikipedia は「50 (または 80)」と自分でも揺れており、catsatcards は 100。
// この実装は 100 を採る。どのコントラクトも「破局は 100」という書き方で
// 揃っているのがこのゲームの書式で、自分でも揺れている数字より、
// 通しで一貫している側を信じるほうが盤面が読みやすい。

// 罰点の定数。
const (
	// QuodlibetTrickPenalty はトリック 1 つあたりの基本罰点。
	QuodlibetTrickPenalty = 10
	// QuodlibetSweepPenalty は「全部やってしまった」ときの罰点。
	QuodlibetSweepPenalty = 100
	// QuodlibetKingOfHeartsPenalty はアラリック (K♥) の罰点。
	QuodlibetKingOfHeartsPenalty = 50
	// QuodlibetRedRuffianPenalty は赤い破落戸 (Q♦) の罰点。
	QuodlibetRedRuffianPenalty = 30
	// QuodlibetLowHeartPenalty は「赤なし」での低いハート (7-10) の罰点。
	QuodlibetLowHeartPenalty = 20
	// QuodlibetHighHeartPenalty は「赤なし」での高いハート (J-A) の罰点。
	QuodlibetHighHeartPenalty = 10
	// QuodlibetQueenPenalty はオーバー / ウンターでの Q の罰点。
	QuodlibetQueenPenalty = 30
	// QuodlibetJackPenalty はオーバー / ウンターでの J の罰点。
	QuodlibetJackPenalty = 20
	// QuodlibetBribeTrickPenalty は賄賂でトリックを取った罰点。
	QuodlibetBribeTrickPenalty = 30
	// QuodlibetBribeLowCardPenalty は賄賂で最低の札を出した罰点。
	QuodlibetBribeLowCardPenalty = 20
	// QuodlibetShedStep はシェディング系で 1 枚残すごとの罰点の刻み。
	QuodlibetShedStep = 10
)

// quodlibetFirstThreeAndLastPenalties は 1238 のトリック別罰点。
// index は 0 始まりのトリック番号。
var quodlibetFirstThreeAndLastPenalties = map[int]int{
	0: 10,  // 第 1 トリック
	1: 20,  // 第 2 トリック
	2: 30,  // 第 3 トリック
	7: 100, // 最終トリック
}

// quodlibetContractDef は 1 コントラクトの罰点関数。
type quodlibetContractDef struct {
	id    int
	score func(q *Quodlibet) map[int]int
}

// quodlibetContractDefs は全 12 コントラクトの定義表 (id 昇順)。
var quodlibetContractDefs = []quodlibetContractDef{
	{id: QuodlibetPlus, score: quodlibetScorePlus},
	{id: QuodlibetMinus, score: quodlibetScoreMinus},
	{id: QuodlibetBadNeighbour, score: quodlibetScoreBadNeighbour},
	{id: QuodlibetAlarich, score: quodlibetScoreAlarich},
	{id: QuodlibetFirstThreeAndLast, score: quodlibetScoreFirstThreeAndLast},
	{id: QuodlibetNoReds, score: quodlibetScoreNoReds},
	{id: QuodlibetOberUnter, score: quodlibetScoreOberUnter},
	{id: QuodlibetBribe, score: quodlibetScoreBribe},
	{id: QuodlibetOpen, score: quodlibetScoreMinus},
	{id: QuodlibetHunt, score: quodlibetScoreMinus},
	{id: QuodlibetQuadrature, score: quodlibetScoreShedding},
	{id: QuodlibetSnack, score: quodlibetScoreShedding},
}

// scoreDeal は現在のコントラクトの罰点関数を実行し、内訳を返す。
func (q *Quodlibet) scoreDeal() *QuodlibetDealDetail {
	points := quodlibetZeroPoints()
	if def := quodlibetContractDefFor(q.currentContract); def != nil {
		points = def.score(q)
	}
	return &QuodlibetDealDetail{
		Contract:  q.currentContract,
		Round:     QuodlibetRoundOf(q.currentContract),
		DealerIdx: q.dealerIdx,
		Points:    points,
	}
}

// quodlibetContractDefFor は id に対応する定義を返す (なければ nil)。
func quodlibetContractDefFor(id int) *quodlibetContractDef {
	for i := range quodlibetContractDefs {
		if quodlibetContractDefs[i].id == id {
			return &quodlibetContractDefs[i]
		}
	}
	return nil
}

// quodlibetZeroPoints は全席 0 の内訳を返す。
func quodlibetZeroPoints() map[int]int {
	points := make(map[int]int, QuodlibetPlayerCnt)
	for i := 0; i < QuodlibetPlayerCnt; i++ {
		points[i] = 0
	}
	return points
}

// quodlibetScorePlus: **取れなかった** トリック 1 つにつき罰点。1 つも取れな
// ければ 80 ではなく 100。
func quodlibetScorePlus(q *Quodlibet) map[int]int {
	points := quodlibetZeroPoints()
	for i, p := range q.players {
		won := p.GetTrickCount()
		if won == 0 {
			points[i] = QuodlibetSweepPenalty
			continue
		}
		points[i] = (QuodlibetHandSize - won) * QuodlibetTrickPenalty
	}
	return points
}

// quodlibetScoreMinus: 取ったトリック 1 つにつき罰点。全部取ると 80 ではなく 100。
func quodlibetScoreMinus(q *Quodlibet) map[int]int {
	points := quodlibetZeroPoints()
	for i, p := range q.players {
		won := p.GetTrickCount()
		if won == QuodlibetHandSize {
			points[i] = QuodlibetSweepPenalty
			continue
		}
		points[i] = won * QuodlibetTrickPenalty
	}
	return points
}

// quodlibetScoreBadNeighbour: マイナスと同じ点が **右隣** に付く。
func quodlibetScoreBadNeighbour(q *Quodlibet) map[int]int {
	base := quodlibetScoreMinus(q)
	points := quodlibetZeroPoints()
	for i := 0; i < QuodlibetPlayerCnt; i++ {
		points[QuodlibetRightOf(i)] += base[i]
	}
	return points
}

// quodlibetScoreAlarich: K♥ 50、Q♦ 30。**両方を同じトリックで取ると 100** で、
// 80 にはならない ── 合計より重い破局として置き換える。
func quodlibetScoreAlarich(q *Quodlibet) map[int]int {
	points := quodlibetZeroPoints()
	for i, p := range q.players {
		for _, trick := range p.GetTricksTaken() {
			king, ruffian := false, false
			for _, c := range trick {
				if c == nil {
					continue
				}
				if c.GetDesign() == CardDesignHeart && c.GetValue() == 13 {
					king = true
				}
				if c.GetDesign() == CardDesignDiamond && c.GetValue() == 12 {
					ruffian = true
				}
			}
			switch {
			case king && ruffian:
				points[i] += QuodlibetSweepPenalty
			case king:
				points[i] += QuodlibetKingOfHeartsPenalty
			case ruffian:
				points[i] += QuodlibetRedRuffianPenalty
			}
		}
	}
	return points
}

// quodlibetScoreFirstThreeAndLast: 第 1・2・3 と最終トリックにだけ罰点が付く。
// 途中の 4 トリックは無料。
func quodlibetScoreFirstThreeAndLast(q *Quodlibet) map[int]int {
	points := quodlibetZeroPoints()
	for trickIdx, winner := range q.trickWinners {
		if penalty, ok := quodlibetFirstThreeAndLastPenalties[trickIdx]; ok {
			points[winner] += penalty
		}
	}
	return points
}

// quodlibetScoreNoReds: ハートを取ると罰点。
//
// **低い札のほうが重い。** 7-10 が 20 点、J-A が 10 点で、普通のハート系とは
// 逆 ── 安く見える札を押しつけ合うのがこのコントラクトの肝。8 枚すべてを
// 同じトリックで取ると 120 ではなく 100。
func quodlibetScoreNoReds(q *Quodlibet) map[int]int {
	points := quodlibetZeroPoints()
	for i, p := range q.players {
		for _, trick := range p.GetTricksTaken() {
			hearts, sum := 0, 0
			for _, c := range trick {
				if c == nil || c.GetDesign() != CardDesignHeart {
					continue
				}
				hearts++
				sum += quodlibetHeartPenalty(c)
			}
			if hearts == QuodlibetHandSize {
				points[i] += QuodlibetSweepPenalty
				continue
			}
			points[i] += sum
		}
	}
	return points
}

// quodlibetHeartPenalty は「赤なし」でのハート 1 枚の罰点を返す。
func quodlibetHeartPenalty(c *Card) int {
	v := c.GetValue()
	if v >= 7 && v <= 10 {
		return QuodlibetLowHeartPenalty
	}
	return QuodlibetHighHeartPenalty
}

// quodlibetScoreOberUnter: Q 30、J 20。**同じトリックで Q と J を両方取ると
// 100** で、内訳の合計ではない。
func quodlibetScoreOberUnter(q *Quodlibet) map[int]int {
	points := quodlibetZeroPoints()
	for i, p := range q.players {
		for _, trick := range p.GetTricksTaken() {
			queens, jacks, sum := 0, 0, 0
			for _, c := range trick {
				if c == nil {
					continue
				}
				switch c.GetValue() {
				case 12:
					queens++
					sum += QuodlibetQueenPenalty
				case 11:
					jacks++
					sum += QuodlibetJackPenalty
				}
			}
			if queens > 0 && jacks > 0 {
				points[i] += QuodlibetSweepPenalty
				continue
			}
			points[i] += sum
		}
	}
	return points
}

// quodlibetScoreBribe: トリックを取ると 30、そのトリックに **最低の札** を
// 出しても 20。**最低の札で取ると 100** で、50 にはならない。
func quodlibetScoreBribe(q *Quodlibet) map[int]int {
	points := quodlibetZeroPoints()
	for _, trick := range q.dealTricks() {
		if len(trick.cards) == 0 {
			continue
		}
		lowest := trick.lowestPlayer()
		if lowest == trick.winner {
			points[trick.winner] += QuodlibetSweepPenalty
			continue
		}
		points[trick.winner] += QuodlibetBribeTrickPenalty
		points[lowest] += QuodlibetBribeLowCardPenalty
	}
	return points
}

// quodlibetScoreShedding: 上がりが早い順に、残した手札 1 枚あたりの罰点が
// 10 / 20 / 30 と重くなる。最初に上がった人は 0。
func quodlibetScoreShedding(q *Quodlibet) map[int]int {
	points := quodlibetZeroPoints()
	for i, p := range q.players {
		rank := p.GetOutRank()
		if rank <= 1 {
			continue
		}
		points[i] = (rank - 1) * QuodlibetShedStep * p.GetCardsSize()
	}
	return points
}

// quodlibetTrick は 1 トリックぶんの、誰が何を出したかの記録。
type quodlibetTrick struct {
	cards  []*TrickCard
	winner int
}

// lowestPlayer はそのトリックに最も弱い札を出した席を返す。
//
// **スートをまたいで比べる。** 賄賂は「一番安い札を出した人」に点が付く規則で、
// 台札スートに限った話ではない。
func (t quodlibetTrick) lowestPlayer() int {
	best, idx := -1, t.winner
	for _, tc := range t.cards {
		if tc == nil || tc.Card == nil {
			continue
		}
		s := QuodlibetCardStrength(tc.Card)
		if best < 0 || s < best {
			best, idx = s, tc.PlayerIdx
		}
	}
	return idx
}

// dealTricks はこのディールで完了した全トリックを、出た順に返す。
//
// **獲得トリックだけでは足りない。** 賄賂は「取らなかった人が出した札」にも
// 点を付けるので、勝者ごとにまとめた TrickHolder からは誰が何を出したのかを
// 復元できない。だからプレイ中に席つきで控えておく。
func (q *Quodlibet) dealTricks() []quodlibetTrick {
	out := make([]quodlibetTrick, 0, QuodlibetHandSize)
	for i, cards := range q.trickRecord {
		if i >= len(q.trickWinners) {
			break
		}
		out = append(out, quodlibetTrick{cards: cards, winner: q.trickWinners[i]})
	}
	return out
}
