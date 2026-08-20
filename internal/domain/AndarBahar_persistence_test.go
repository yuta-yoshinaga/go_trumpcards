//go:build test

package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// andarBaharTampered は本物の盤面を JSON にしてから 1 か所だけ壊し、復元を試みる。
//
// **土台は手書きせず、実際に回した局面から作ります。** 手で組んだ JSON は書き込み側の
// 形を外していても気付けず、検証したいガードの手前で落ちてしまいます。
func andarBaharTampered(t *testing.T, base *AndarBahar, mutate func(m map[string]any)) error {
	t.Helper()
	data, err := json.Marshal(base)
	require.NoError(t, err)

	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	mutate(m)

	broken, err := json.Marshal(m)
	require.NoError(t, err)

	var back AndarBahar
	return json.Unmarshal(broken, &back)
}

// andarBaharEnded は決着済みの盤面を返す。
func andarBaharEnded(t *testing.T) *AndarBahar {
	t.Helper()
	ab := NewDefaultAndarBahar()
	require.NoError(t, ab.Bet(100, AndarBaharBetAndar, 50, AndarBaharSide2To5))
	require.True(t, ab.GetGameEndFlag())
	return ab
}

// **ベット前・決着後のどちらも往復する。** これが負のコントロールです。
func TestAndarBahar_ReachableStatesSurviveARoundTrip(t *testing.T) {
	for round := range 60 {
		ab := NewDefaultAndarBahar()

		data, err := json.Marshal(ab)
		require.NoError(t, err)
		var fresh AndarBahar
		require.NoError(t, json.Unmarshal(data, &fresh),
			"ラウンド %d: ベット前の盤面で書き込み側が不変条件を破った", round)
		assert.Equal(t, ab.GetFirstColumn(), fresh.GetFirstColumn())

		require.NoError(t, ab.Bet(100, AndarBaharBetAndar, 50, AndarBaharSide6To10))
		data, err = json.Marshal(ab)
		require.NoError(t, err)
		var done AndarBahar
		require.NoError(t, json.Unmarshal(data, &done),
			"ラウンド %d: 決着後の盤面で書き込み側が不変条件を破った", round)

		assert.Equal(t, ab.GetWinner(), done.GetWinner())
		assert.Equal(t, ab.DealtCount(), done.DealtCount())
		assert.Equal(t, ab.GetPayout(), done.GetPayout())
		assert.Equal(t, ab.GetChips(), done.GetChips())
	}
}

// **改竄した保存データは、壊した場所を名指しして弾く。**
//
// エラー本文まで検査するのは、**手前の別のガードが落としたのを「検出できた」と
// 数えないため**です。
func TestAndarBahar_UnmarshalRejectsTamperedState(t *testing.T) {
	tests := []struct {
		name   string
		ended  bool
		mutate func(m map[string]any)
		want   string
	}{
		{
			name:   "フェーズが範囲外",
			mutate: func(m map[string]any) { m["ps"] = 9 },
			want:   "phase out of range",
		},
		{
			name:   "終了フラグとフェーズが矛盾する",
			mutate: func(m map[string]any) { m["ge"] = true },
			want:   "the game-end flag and the phase disagree",
		},
		{
			name:   "先に配る列が範囲外",
			mutate: func(m map[string]any) { m["fc"] = 7 },
			want:   "first column out of range",
		},
		{
			// **先に配る列は基準札の色で決まる。** 食い違いは改竄。
			name:   "先に配る列が基準札の色と食い違う",
			mutate: func(m map[string]any) { m["fc"] = 1 - andarBaharToInt(m["fc"]) },
			want:   "does not match the joker's colour",
		},
		{
			name:   "ベット先が範囲外",
			mutate: func(m map[string]any) { m["bt"] = 5 },
			want:   "bet target out of range",
		},
		{
			name:   "ベット額が 10 の倍数でない",
			ended:  true,
			mutate: func(m map[string]any) { m["ba"] = 15 },
			want:   "bet amount out of range",
		},
		{
			name:   "ベット額が上限超え",
			ended:  true,
			mutate: func(m map[string]any) { m["ba"] = AndarBaharMaxBet + 10 },
			want:   "bet amount out of range",
		},
		{
			// **内訳と合計は同時に決まる** (#5770)。片方だけ書き換わった保存を
			// 通すと、画面が「メインは当たったのに合計は減っている」と出す。
			name:   "払い戻しの内訳が合計と食い違う",
			ended:  true,
			mutate: func(m map[string]any) { m["pm"] = andarBaharToInt(m["pm"]) + 10 },
			want:   "does not add up to",
		},
		{
			name:   "内訳が負",
			ended:  true,
			mutate: func(m map[string]any) { m["pm"] = -10; m["pd"] = andarBaharToInt(m["po"]) + 10 },
			want:   "payout breakdown cannot be negative",
		},
		{
			// **張っていないサイドベットには払い戻せない。**
			name:  "帯なしなのにサイドの払い戻しがある",
			ended: true,
			mutate: func(m map[string]any) {
				m["sb"] = AndarBaharSideNone
				m["sa"] = 0
				// 合計も一緒に持ち上げる。**手前の「和が合わない」ガードに
				// 落とされると、このガードを検査したことにならない。**
				m["pd"] = 20
				m["po"] = andarBaharToInt(m["pm"]) + 20
			},
			want: "paid on a side bet that was never placed",
		},
		{
			name:   "サイドベット額が範囲外",
			ended:  true,
			mutate: func(m map[string]any) { m["sa"] = -10 },
			want:   "side bet amount out of range",
		},
		{
			name:   "サイドベットの帯が範囲外",
			ended:  true,
			mutate: func(m map[string]any) { m["sb"] = 99 },
			want:   "side bet band out of range",
		},
		{
			// **賭けていない帯に金額は載らない。**
			name:   "帯なしなのに金額が載っている",
			ended:  true,
			mutate: func(m map[string]any) { m["sb"] = AndarBaharSideNone; m["sa"] = 50 },
			want:   "staked on no side bet band",
		},
		{
			// **逆向きも同じく作れない。** 帯 0 は有効な値なので、番号だけを見て
			// 「賭けていない」とは判定できない——金額が 0 なら帯は載りません。
			name:   "帯があるのに金額が 0",
			ended:  true,
			mutate: func(m map[string]any) { m["sa"] = 0 },
			want:   "carries no stake",
		},
		{
			name:   "払戻額が負",
			ended:  true,
			mutate: func(m map[string]any) { m["po"] = -1 },
			want:   "payout cannot be negative",
		},
		{
			name:   "ベット前なのに払い戻されている",
			mutate: func(m map[string]any) { m["po"] = 100 },
			want:   "paid out before the round was dealt",
		},
		{
			name:   "基準札が伏せられている",
			mutate: func(m map[string]any) { m["jk"] = nil },
			want:   "the joker must be face up",
		},
		{
			name:   "ベット前なのに札が配られている",
			mutate: func(m map[string]any) { m["an"] = []any{andarBaharCardJSON(CardDesignSpade, 3)} },
			want:   "cards are already dealt in the bet phase",
		},
		{
			// **交互配布なので、先の列は後の列と同数か 1 枚多いだけ。**
			name:  "列の枚数が交互配布と食い違う",
			ended: true,
			mutate: func(m map[string]any) {
				// **空の列は JSON では `null`** なので comma-ok で受ける。素で
				// キャストすると、バハールが先に配られて 1 枚目で決着した回に落ちる。
				andar, _ := m["an"].([]any)
				m["an"] = append(andar, andarBaharCardJSON(CardDesignSpade, 2), andarBaharCardJSON(CardDesignClover, 3))
			},
			want: "the columns do not alternate",
		},
		{
			name:   "決着したのに勝った列が無い",
			ended:  true,
			mutate: func(m map[string]any) { m["wn"] = -1 },
			want:   "winner out of range",
		},
		{
			// **勝った列の末尾が同ランクでなければ、そこで止まった説明がつかない。**
			name:  "勝った列が基準札と同ランクで終わっていない",
			ended: true,
			mutate: func(m map[string]any) {
				j := m["jk"].(map[string]any)
				j["v"] = andarBaharToInt(j["v"])%13 + 1
			},
			want: "did not end on a card matching the joker",
		},
		{
			name:   "決着後なのに 1 枚も配られていない",
			ended:  true,
			mutate: func(m map[string]any) { m["an"] = []any{}; m["bh"] = []any{} },
			want:   "ended without dealing a card",
		},
		{
			name:  "棋譜が長すぎる",
			ended: true,
			mutate: func(m map[string]any) {
				log := make([]any, andarBaharMaxSliceLen+1)
				for i := range log {
					log[i] = map[string]any{}
				}
				m["al"] = log
			},
			want: "exceeds maximum allowed size",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := NewDefaultAndarBahar()
			if tt.ended {
				base = andarBaharEnded(t)
			}
			err := andarBaharTampered(t, base, tt.mutate)
			require.Error(t, err, "改竄した保存データが素通しした")
			assert.ErrorContains(t, err, tt.want, "別のガードが先に落としている可能性がある")
		})
	}
}

// **途中に同ランクが混ざった列は、そこで止まっていたはず。**
//
// 枚数も勝者も範囲内のまま「あり得ない盤面」を作れるので、範囲チェックだけでは通ります。
func TestAndarBahar_UnmarshalRejectsAMatchBeforeTheEnd(t *testing.T) {
	// **1 枚目で決着した回は使えません。** 負けた列が空で、差し替える「途中の札」が
	// そもそも存在しないからです (5.88% で起きます)。負けた列に札が乗るまで引き直します。
	var base *AndarBahar
	for range 200 {
		b := andarBaharEnded(t)
		if b.DealtCount() >= 2 {
			base = b
			break
		}
	}
	require.NotNil(t, base, "2 枚以上配られる回を引けなかった")

	loser := "bh"
	if base.GetWinner() == AndarBaharBetBahar {
		loser = "an"
	}
	rank := andarBaharRank(base.GetJoker())

	err := andarBaharTampered(t, base, func(m map[string]any) {
		// 負けた列の先頭を基準札と同ランクに差し替える。枚数も勝者も変えない。
		col, ok := m[loser].([]any)
		require.True(t, ok, "負けた列が空だった")
		require.NotEmpty(t, col)
		c := col[0].(map[string]any)
		c["v"] = rank
		c["d"] = (andarBaharToInt(c["d"]) + 1) % 4
	})
	require.Error(t, err, "途中の同ランクが素通しした")
	assert.ErrorContains(t, err, "match the joker")
}

func TestAndarBahar_UnmarshalRejectsMalformedJSON(t *testing.T) {
	t.Parallel()

	var ab AndarBahar
	assert.Error(t, json.Unmarshal([]byte(`{"ps":`), &ab))
}

// **山札や棋譜が欠けていても落ちない。** 空で補います。
func TestAndarBahar_UnmarshalFillsMissingSlices(t *testing.T) {
	base := NewDefaultAndarBahar()
	data, err := json.Marshal(base)
	require.NoError(t, err)

	var m map[string]any
	require.NoError(t, json.Unmarshal(data, &m))
	delete(m, "tc")
	delete(m, "al")
	delete(m, "hi")
	broken, err := json.Marshal(m)
	require.NoError(t, err)

	var back AndarBahar
	require.NoError(t, json.Unmarshal(broken, &back))
	assert.NotNil(t, back.GetHistory(), "罫線は空スライスで補う")
	assert.Empty(t, back.GetHistory())
}

// andarBaharCardJSON は改竄用のカード JSON を作る。
func andarBaharCardJSON(design, value int) map[string]any {
	return map[string]any{"d": design, "v": value, "w": false}
}

// andarBaharToInt は JSON 由来の数値を int にする。
func andarBaharToInt(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	default:
		return 0
	}
}
