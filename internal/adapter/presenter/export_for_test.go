//go:build test

package presenter

// CuiPlayerNameAtForTest exposes cuiPlayerNameAt to the external test package.
//
// **範囲外の枝は 3 つの呼び出し元からは踏めない** (`dealerIdx` は常に範囲内)
// ので、出力を組み立てる経路からは確かめられない。踏まれない枝を残すなら
// 直接踏む、というのがこの穴の理由 (#6470 のレビュー指摘)。
func CuiPlayerNameAtForTest[P cuiPlayer](players []P, idx int) string {
	return cuiPlayerNameAt(players, idx)
}
