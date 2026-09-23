//go:build test

package domain

// ForcePassForTest 指定席を強制的にパスさせる (テスト用)。
//
// CPU の入札は手札の強さから決まるため、テストから「全員パス」を確実に作る
// 手段が無い。入札の適用経路そのものは applyBid を通るので、検証したい
// 「全パスなら流局しポットは持ち越す」の筋道は本番と同じものを通る。
func (g *Vira) ForcePassForTest(idx int) error {
	if idx < 0 || idx >= ViraPlayerCnt {
		return NewDomainErrorCode(ErrInvalidPlay, "vira.errPlayerIndexOutOfRange", nil)
	}
	return g.applyBid(idx, ViraBidPass)
}
