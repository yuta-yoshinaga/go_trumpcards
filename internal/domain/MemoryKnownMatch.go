//go:build !js || !wasm || solo

package domain

// MemoryKnownMatchIdx は、表向き 1 枚に一致する「once めくって見た」伏せ札の
// 位置を返す。該当が無ければ ok=false。
//
// Web 側 (frontend/src/utils/memoryKnownMatch.ts) と同じ規則。あちらは
// ブラウザが貯めた seen を使うが、CUI はドメインを直接読めるので Visited
// (一度でも表になった) を同じ意味で使う。どちらも**プレイヤーが実際に見た札**
// しか対象にしないので、記憶の再生であって情報の追加ではない。
//
// **表向きがちょうど 1 枚のときだけ答える。** 0 枚や 2 枚のときに答えると、
// 「一致がある」という表示が実際には打てない手を指すことになる。
func MemoryKnownMatchIdx(board []*MemoryBoardCard) (int, bool) {
	faceUpIdx := -1
	for i, bc := range board {
		if bc == nil || !bc.FaceUp || bc.Taken {
			continue
		}
		if faceUpIdx >= 0 {
			return 0, false
		}
		faceUpIdx = i
	}
	if faceUpIdx < 0 {
		return 0, false
	}
	target := board[faceUpIdx].Card
	if target == nil {
		return 0, false
	}
	for i, bc := range board {
		if i == faceUpIdx || bc == nil || bc.FaceUp || bc.Taken || !bc.Visited || bc.Card == nil {
			continue
		}
		if bc.Card.GetDesign() == target.GetDesign() && bc.Card.GetValue() == target.GetValue() {
			return i, true
		}
	}
	return 0, false
}
