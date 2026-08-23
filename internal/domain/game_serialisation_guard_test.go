//go:build test

package domain

import (
	"encoding/json"
	"testing"
)

// TestGameStateIsActuallySerialised は、盤面が **`{}` にならない** ことを見る。
//
// # なぜこのガードが要るか
//
// ゲーム型が非公開フィールドしか持たず `MarshalJSON` も無いと、
// `encoding/json` が出すのは `{}` の 2 バイトだけになる —— **エラーは出ない**。
// Cloudflare Worker はリクエストごとに KV から盤面を復元するので、保存が空だと
// 毎リクエスト初期状態の卓が作り直され、ゲームが進行しない。
//
// Skat がこの状態で出荷されていた (#6215)。静かに壊れるので、ここで一括して
// 網を張る。**新しいゲームを足したらこの表にも足すこと。**
func TestGameStateIsActuallySerialised(t *testing.T) {
	// MarshalJSON を持たない型を抜き出して並べてある。全 344 ゲームを列挙する
	// 代わりに、危険な形をしているものだけを見る。
	games := map[string]any{
		"Skat":       NewDefaultSkat(),
		"BigO":       NewDefaultBigO(),
		"Courchevel": NewDefaultCourchevel(),
		"Burraco":    NewDefaultBurraco(),
	}

	for name, g := range games {
		t.Run(name, func(t *testing.T) {
			blob, err := json.Marshal(g)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if len(blob) <= 2 {
				t.Fatalf("%s marshals to %q — the board is not being saved, so a Worker "+
					"rebuilds a fresh table on every request", name, blob)
			}
		})
	}
}
