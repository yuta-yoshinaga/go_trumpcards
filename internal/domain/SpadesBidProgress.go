package domain

// SpadesBidProgressKind は入札契約の進捗の種類。
type SpadesBidProgressKind int

const (
	// SpadesBidNilOk ニル宣言でまだ1トリックも取っていない (継続中)。
	SpadesBidNilOk SpadesBidProgressKind = iota
	// SpadesBidNilFail ニル宣言なのにトリックを取ってしまった (確定で失敗)。
	SpadesBidNilFail
	// SpadesBidRemaining 通常入札で、契約にあと Remaining トリック足りない。
	SpadesBidRemaining
	// SpadesBidMade 契約を満たした。超過分は Bags。
	SpadesBidMade
)

// SpadesBidProgress は入札に対する現在地。
type SpadesBidProgress struct {
	Kind      SpadesBidProgressKind
	Remaining int // Kind == SpadesBidRemaining のときだけ意味を持つ
	Bags      int // Kind == SpadesBidMade のときだけ意味を持つ
}

// SpadesBidProgressOf は入札とトリック数から進捗を導く。
//
// Web 側 (frontend/src/utils/spadesBid.ts の spadesBidProgress) と同じ規則。
// **ニルは1トリック取った時点で失敗が確定する**ので、その瞬間から nilFail に
// なる ── 残りトリックで挽回できる契約ではない。
func SpadesBidProgressOf(bid, trickCount int) SpadesBidProgress {
	if bid <= 0 {
		if trickCount > 0 {
			return SpadesBidProgress{Kind: SpadesBidNilFail}
		}
		return SpadesBidProgress{Kind: SpadesBidNilOk}
	}
	if trickCount < bid {
		return SpadesBidProgress{Kind: SpadesBidRemaining, Remaining: bid - trickCount}
	}
	return SpadesBidProgress{Kind: SpadesBidMade, Bags: trickCount - bid}
}
