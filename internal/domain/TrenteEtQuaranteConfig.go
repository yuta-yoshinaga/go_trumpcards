//go:build !js || !wasm || extra4

package domain

import "encoding/json"

// TrenteEtQuaranteBet はトラント・エ・カラント (Trente et Quarante / Rouge et Noir) の
// ベット種別。プレイヤーは 1 ラウンドにつきいずれか 1 種にステークを賭ける。
type TrenteEtQuaranteBet int

// Trente et Quarante のベット種別定数
const (
	// TrenteEtQuaranteBetNoir 黒 (Noir) 列が勝つ (合計が小さい) 方に賭ける
	TrenteEtQuaranteBetNoir TrenteEtQuaranteBet = iota
	// TrenteEtQuaranteBetRouge 赤 (Rouge) 列が勝つ方に賭ける
	TrenteEtQuaranteBetRouge
	// TrenteEtQuaranteBetCouleur 最初に配られた札の色が勝ち列の色と一致する方に賭ける
	TrenteEtQuaranteBetCouleur
	// TrenteEtQuaranteBetInverse 最初に配られた札の色が勝ち列の色と異なる方に賭ける
	TrenteEtQuaranteBetInverse
)

// TrenteEtQuaranteBetNames はベット種別名マップ。
var TrenteEtQuaranteBetNames = map[TrenteEtQuaranteBet]string{
	TrenteEtQuaranteBetNoir:    "Noir",
	TrenteEtQuaranteBetRouge:   "Rouge",
	TrenteEtQuaranteBetCouleur: "Couleur",
	TrenteEtQuaranteBetInverse: "Inverse",
}

// TrenteEtQuaranteBetValid はベット種別が有効な列挙値かどうかを返す。
func TrenteEtQuaranteBetValid(b TrenteEtQuaranteBet) bool {
	return b >= TrenteEtQuaranteBetNoir && b <= TrenteEtQuaranteBetInverse
}

// TrenteEtQuaranteConfig はトラント・エ・カラントのローカルルール設定。
type TrenteEtQuaranteConfig struct {
	// DefaultBet はラウンド開始時にあらかじめ選択しておくベット種別。
	DefaultBet TrenteEtQuaranteBet `json:"db"`
}

// DefaultTrenteEtQuaranteConfig はデフォルトのローカルルール設定を返す。
//   - デッキ: 6 デッキ 312 枚のシュー
//   - デフォルトベット: Noir
func DefaultTrenteEtQuaranteConfig() TrenteEtQuaranteConfig {
	return TrenteEtQuaranteConfig{
		DefaultBet: TrenteEtQuaranteBetNoir,
	}
}

// Validate は設定値のドメインバリデーションを行う。
func (c TrenteEtQuaranteConfig) Validate() error {
	return ValidateRange(
		"default bet",
		int(c.DefaultBet),
		int(TrenteEtQuaranteBetNoir),
		int(TrenteEtQuaranteBetInverse),
	)
}

// trenteEtQuaranteConfigJSON is the JSON wire format for TrenteEtQuaranteConfig.
type trenteEtQuaranteConfigJSON struct {
	DefaultBet TrenteEtQuaranteBet `json:"db"`
}

// MarshalJSON implements json.Marshaler.
func (c TrenteEtQuaranteConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(trenteEtQuaranteConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *TrenteEtQuaranteConfig) UnmarshalJSON(data []byte) error {
	var j trenteEtQuaranteConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	*c = TrenteEtQuaranteConfig(j)
	return nil
}
