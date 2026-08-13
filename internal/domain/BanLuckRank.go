//go:build !js || !wasm || casino

package domain

// BanLuckRank は 1 つの手札の役。**大きいほど強い。**
type BanLuckRank int

// 役の一覧。順序そのものが規則なので、値の並びを入れ替えないこと。
const (
	// BanLuckRankBust はバスト。21 を超えた手。**役の比較以前に負ける。**
	BanLuckRankBust BanLuckRank = iota
	// BanLuckRankPoint は普通の手。強さは合計値で決まる。
	BanLuckRankPoint
	// BanLuckRankFiveDragon は 5 枚で 21 以下。
	BanLuckRankFiveDragon
	// BanLuckRankBanLuck は配られた 2 枚で 21 (チャイニーズ・ブラックジャック)。
	BanLuckRankBanLuck
	// BanLuckRankBanBan は A + A。**最強。**
	BanLuckRankBanBan
)

// BanLuckRankName は役の識別子を返す (i18n キーの一部に使う)。
func BanLuckRankName(r BanLuckRank) string {
	switch r {
	case BanLuckRankPoint:
		return "point"
	case BanLuckRankFiveDragon:
		return "fiveDragon"
	case BanLuckRankBanLuck:
		return "banLuck"
	case BanLuckRankBanBan:
		return "banBan"
	default:
		return "bust"
	}
}

// BanLuckRankMax は最大の役 (復元時の範囲検査に使う)。
const BanLuckRankMax = BanLuckRankBanBan

// EvalBanLuckHand は手札の役を判定する。
//
// **枚数が役を決めるので、合計値だけでは足りない。** 2 枚の 21 は Ban Luck、
// 引いて到達した 21 は普通の手で、同じ 21 でも強さが違う。5 枚で 21 以下なら
// 合計に関係なく Five Dragon になる ── 16 の Five Dragon が 20 の普通の手に
// 勝つ、というのがこのゲームの見どころで、合計値だけで比べると消える。
func EvalBanLuckHand(h *BlackJackHand) BanLuckRank {
	if h == nil {
		return BanLuckRankBust
	}
	cards := h.GetCards()
	score := h.GetScore()
	if score > BanLuckTarget {
		return BanLuckRankBust
	}
	if len(cards) == 2 {
		if bjValue(cards[0]) == 1 && bjValue(cards[1]) == 1 {
			// **A+A は 21 ではない (12 か 2)。** 合計を見ていると取りこぼす。
			return BanLuckRankBanBan
		}
		if score == BanLuckTarget {
			return BanLuckRankBanLuck
		}
	}
	if len(cards) >= BanLuckFiveDragonCards {
		return BanLuckRankFiveDragon
	}
	return BanLuckRankPoint
}

// BanLuckPayoutFor は役に対応する配当倍率を返す。
func BanLuckPayoutFor(r BanLuckRank) int {
	switch r {
	case BanLuckRankBanBan:
		return BanLuckPayoutBanBan
	case BanLuckRankBanLuck:
		return BanLuckPayoutBanLuck
	case BanLuckRankFiveDragon:
		return BanLuckPayoutFiveDragon
	default:
		return BanLuckPayoutNormal
	}
}

// BanLuckOutcome は子から見た 1 席の決着。
type BanLuckOutcome int

const (
	// BanLuckOutcomeLose は子の負け。
	BanLuckOutcomeLose BanLuckOutcome = iota
	// BanLuckOutcomePush は引き分け。
	BanLuckOutcomePush
	// BanLuckOutcomeWin は子の勝ち。
	BanLuckOutcomeWin
)

// CompareBanLuck は子と親を比べ、子から見た決着と**適用される倍率**を返す。
//
// **倍率は勝った側の役で決まる。** 子が Ban Ban で勝てば 3 倍を受け取り、親が
// Ban Ban で勝てば子は 3 倍を払う。負けた側の役は倍率に関わらない ── ここを
// 「常に子の役で決める」と書くと、親の Ban Ban が普通の 1 倍しか取れなくなり、
// 親の義務ヒットに見合わなくなる。
//
// 両方バストは引き分けにしない。**親のバストが子の勝ちになるのは、子が
// 生きているときだけ**で、子も飛んでいれば子の負けである (先に飛んでいる)。
func CompareBanLuck(player, banker BanLuckRank, playerScore, bankerScore int) (BanLuckOutcome, int) {
	switch {
	case player == BanLuckRankBust:
		// **子のバストが最優先。** 親が後で飛んでも子は救われない。
		return BanLuckOutcomeLose, BanLuckPayoutFor(banker)
	case banker == BanLuckRankBust:
		return BanLuckOutcomeWin, BanLuckPayoutFor(player)
	case player > banker:
		return BanLuckOutcomeWin, BanLuckPayoutFor(player)
	case player < banker:
		return BanLuckOutcomeLose, BanLuckPayoutFor(banker)
	}
	// ここから先は同じ役同士。
	if player != BanLuckRankPoint {
		// **特別役同士は合計を比べない。** Five Dragon 同士は 16 でも 21 でも
		// 引き分け、Ban Luck 同士も Ban Ban 同士も同じ。枚数で決まる役なので、
		// 合計で優劣を付けると「5 枚 21 が 5 枚 16 に勝つ」規則を足すことになる。
		return BanLuckOutcomePush, BanLuckPayoutFor(player)
	}
	switch {
	case playerScore > bankerScore:
		return BanLuckOutcomeWin, BanLuckPayoutNormal
	case playerScore < bankerScore:
		return BanLuckOutcomeLose, BanLuckPayoutNormal
	default:
		return BanLuckOutcomePush, BanLuckPayoutNormal
	}
}
