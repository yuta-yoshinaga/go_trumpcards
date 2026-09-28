//go:build !js || !wasm || extra7

package domain

import "sort"

// samba_bolivia_shared.go — Samba と Bolivia で完全に同一な純粋ロジック (#8056)。
//
// 型 (SambaMeld / BoliviaMeld / SambaPlayer / BoliviaPlayer) には依存しない。
// 引数はすべて []*Card, int, bool, func(*Card)bool 等のプリミティブ。
// エラーコードはゲーム名接頭辞を呼び出し側 (Samba.go / Bolivia.go) で埋める。
//
// ビルドタグは両ゲームが属する extra7 Worker と共通の !js || !wasm || extra7。

// sbIsWildCard はジョーカーまたは2であればワイルドカードと判定する。
// SambaIsWild / BoliviaIsWild は実装が完全に同一なため、
// 共有コードの内部ヘルパーとしてまとめる。
// 公開 API (SambaIsWild / BoliviaIsWild) はそれぞれのファイルに残す。
func sbIsWildCard(card *Card) bool {
	return card.GetDesign() == CardDesignJoker || card.GetValue() == 2
}

// sbSequenceCardValue はシーケンス判定用のカード値を返す。
// エースは高位 (14) 扱いで、ラップアラウンドはしない。
// Samba の sambaSequenceValue / Bolivia の boliviaEscaleraValue と完全一致。
func sbSequenceCardValue(card *Card) int {
	if card.GetValue() == 1 {
		return 14
	}
	return card.GetValue()
}

// sbSequenceCardValues はカード列のシーケンス値スライスを返す。
// ワイルドカードはスキップする。
// Samba の sambaSequenceValues / Bolivia の boliviaEscaleraValues と完全一致。
func sbSequenceCardValues(cards []*Card) []int {
	out := make([]int, 0, len(cards))
	for _, c := range cards {
		if sbIsWildCard(c) {
			continue
		}
		out = append(out, sbSequenceCardValue(c))
	}
	return out
}

// sbSequenceSuit はシーケンス形状グループのスートを返す (ナチュラルカードがなければ -1)。
// Samba の sambaSequenceSuit / Bolivia の boliviaEscaleraSuit と完全一致。
func sbSequenceSuit(cards []*Card) int {
	for _, c := range cards {
		if !sbIsWildCard(c) {
			return c.GetDesign()
		}
	}
	return -1
}

// sbGroupIsSetShaped はグループの全ナチュラルカードが同ランクかどうかを返す。
// Samba の sambaGroupIsSetShaped / Bolivia の boliviaGroupIsSetShaped と完全一致。
func sbGroupIsSetShaped(cards []*Card) bool {
	rank := 0
	for _, c := range cards {
		if sbIsWildCard(c) {
			continue
		}
		if rank == 0 {
			rank = c.GetValue()
		} else if c.GetValue() != rank {
			return false
		}
	}
	return true
}

// sbNaturalRank はグループ内の最初のナチュラルカードのランクを返す (なければ 0)。
// Samba の sambaNaturalRank / Bolivia の boliviaNaturalRank と完全一致。
func sbNaturalRank(cards []*Card) int {
	for _, c := range cards {
		if !sbIsWildCard(c) {
			return c.GetValue()
		}
	}
	return 0
}

// sbFilterUnused はまだ使われていないカードのみを返す。
// Samba の filterUnused / Bolivia の boliviaFilterUnused と完全一致。
func sbFilterUnused(cards []*Card, used map[*Card]bool) []*Card {
	out := cards[:0:0]
	for _, c := range cards {
		if !used[c] {
			out = append(out, c)
		}
	}
	return out
}

// sbValidateSequenceCards は一連のカードが有効なシーケンス（同スート連番、
// ワイルド・3を含まない、重複なし）かどうかを検証する。スパンを測る前に
// 必ず値をソートする。
//
// errPrefix はエラーコードの接頭辞 ("samba" または "bolivia")。
// 呼び出し側がゲーム名に応じて渡す。エラーコード・メッセージ引数は一切変えない。
//
// Samba の sambaValidateSequenceCards / Bolivia の boliviaValidateEscaleraCards の
// 純粋ロジックを統合したもの（差分は errPrefix のみ）。
func sbValidateSequenceCards(cards []*Card, errPrefix string) error {
	if len(cards) < 3 {
		return NewDomainErrorCode(ErrInvalidPlay, errPrefix+".errSequenceNeedsAtLeastThreeCards", nil)
	}
	design := -1
	vals := make([]int, 0, len(cards))
	for _, c := range cards {
		if sbIsWildCard(c) {
			return NewDomainErrorCode(ErrInvalidPlay, errPrefix+".errSequenceCannotUseWildCards", nil)
		}
		if c.GetValue() == 3 {
			return NewDomainErrorCode(ErrInvalidPlay, errPrefix+".errThreeCannotBeUsedInSequence", nil)
		}
		if design == -1 {
			design = c.GetDesign()
		} else if c.GetDesign() != design {
			return NewDomainErrorCode(ErrInvalidPlay, errPrefix+".errSequenceMeldMustUseSameSuit", nil)
		}
		vals = append(vals, sbSequenceCardValue(c))
	}
	sort.Ints(vals)
	for i := 1; i < len(vals); i++ {
		if vals[i] == vals[i-1] {
			return NewDomainErrorCode(ErrInvalidPlay, errPrefix+".errSequenceMeldCannotDuplicateCard", nil)
		}
		if vals[i] != vals[i-1]+1 {
			return NewDomainErrorCode(ErrInvalidPlay, errPrefix+".errSequenceMeldRanksMustBeConsecutive", nil)
		}
	}
	return nil
}
