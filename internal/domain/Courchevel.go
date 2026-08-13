//go:build !js || !wasm || casino

package domain

// Courchevel (クールシュヴェル) はフランス発祥の 5 枚オマハのバリアントで、
// **フロップの 1 枚目を賭ける前に見せる**ところだけが Big O と違う。
//
// 配る枚数も、役を「手札からちょうど 2 枚 + コミュニティからちょうど 3 枚」で
// 作る規則も、Big O とまったく同じ。変わるのは**公開の時刻**だけで、最初の
// ベットラウンドが 1 枚多い情報の上で行われる ── プリフロップが「見えない
// 5 枚をどう評価するか」から「見えている 1 枚と噛み合うか」に変わり、通常の
// オマハより早い段階で降りる根拠が手に入る。
//
// 実装は既存の Omaha エンジンをそのまま流用し、ホールカード枚数 (5) に加えて
// プリフロップ公開枚数 (1) をパラメータ化しているだけ。フロップは「3 枚まで
// 足す」ので、先に見せた 1 枚と合わせて場は必ず 5 枚で終わる。

// NewDefaultCourchevel は Courchevel をデフォルトのテーブルサイズと
// DefaultOmahaConfig で生成する。CUI / Web / Worker の構築サイトで共有する
// 単一の生成元。
func NewDefaultCourchevel() *Omaha {
	o := NewDefaultBigO()
	o.preflopCommunity = courchevelPreflopCommunity
	return o
}

// NewDefaultCourchevelHiLo は Courchevel を 8 or Better (Hi-Lo スプリット)
// として生成する。
func NewDefaultCourchevelHiLo() *Omaha {
	o := NewDefaultCourchevel()
	o.hiLo = true
	return o
}

// courchevelPreflopCommunity はプリフロップ前に見せるコミュニティの枚数。
const courchevelPreflopCommunity = 1
