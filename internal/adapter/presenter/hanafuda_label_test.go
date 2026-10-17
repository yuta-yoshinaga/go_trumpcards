package presenter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

func TestHanafudaCuiLabelsResolveInBothLanguages(t *testing.T) {
	oldLang := i18n.Lang()
	t.Cleanup(func() { i18n.SetLang(oldLang) })

	cases := []struct {
		name   string
		card   *domain.Card
		label  func(*domain.Card) string
		ja, en string
	}{
		{"koi-koi", domain.NewCard(1, 1, false), koikoiCuiCardStr, "松·光", "Pine·Bright"},
		{"hachi-hachi", domain.NewCard(1, 1, false), hachihachiCuiCardStr, "松·光", "Pine·Bright"},
		{"go-stop", domain.NewCard(1, 1, false), gostopCuiCardStr, "松·光", "Pine·Gwang"},
		{"sakura", domain.NewCard(domain.SakuraCurtainMonth, domain.SakuraCurtainIndex, false), sakuraCuiCardStr, "桜·光(20)", "Cherry·Bright(20)"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			i18n.SetLang("ja")
			require.Contains(t, tc.label(tc.card), tc.ja)
			i18n.SetLang("en")
			require.Contains(t, tc.label(tc.card), tc.en)
		})
	}
}

func TestHanafudaWebFacesCarryMonthAndKind(t *testing.T) {
	cases := []struct {
		name  string
		face  func(*domain.Card) *CardFace
		kind  string
		month int
	}{
		{"koi-koi", koikoiFace, "bright", 1},
		{"hachi-hachi", hachihachiFace, "bright", 1},
		{"go-stop", gostopFace, "gwang", 1},
		{"mushi", mushiFace, "bright", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := cardToOutputWithFace(domain.NewCard(1, 1, false), tc.face)
			require.Equal(t, tc.month, out.Month)
			require.Equal(t, tc.kind, out.Kind)
		})
	}
}

func TestHanafudaCategoryCodes(t *testing.T) {
	for _, tc := range []struct {
		name string
		got  string
		want string
	}{
		{"koi bright", koiKoiCategoryCode(domain.KoiKoiBright), "bright"},
		{"koi animal", koiKoiCategoryCode(domain.KoiKoiAnimal), "animal"},
		{"koi ribbon", koiKoiCategoryCode(domain.KoiKoiRibbon), "ribbon"},
		{"koi chaff", koiKoiCategoryCode(domain.KoiKoiChaff), "chaff"},
		{"hachi bright", hachiHachiCategoryCode(domain.HachiHachiBright), "bright"},
		{"hachi animal", hachiHachiCategoryCode(domain.HachiHachiAnimal), "animal"},
		{"hachi ribbon", hachiHachiCategoryCode(domain.HachiHachiRibbon), "ribbon"},
		{"hachi chaff", hachiHachiCategoryCode(domain.HachiHachiChaff), "chaff"},
		{"go-stop gwang", goStopCategoryCode(domain.GoStopGwang), "gwang"},
		{"go-stop yeol", goStopCategoryCode(domain.GoStopYeol), "yeol"},
		{"go-stop tti", goStopCategoryCode(domain.GoStopTti), "tti"},
		{"go-stop pi", goStopCategoryCode(domain.GoStopPi), "pi"},
		{"mushi bright", mushiCategoryCode(domain.MushiBright), "bright"},
		{"mushi animal", mushiCategoryCode(domain.MushiAnimal), "animal"},
		{"mushi ribbon", mushiCategoryCode(domain.MushiRibbon), "ribbon"},
		{"mushi chaff", mushiCategoryCode(domain.MushiChaff), "chaff"},
	} {
		t.Run(tc.name, func(t *testing.T) { require.Equal(t, tc.want, tc.got) })
	}
}

func TestMushiCardOutputKeepsMonthAndKind(t *testing.T) {
	out := mushiCardOutput(domain.NewCard(1, 1, false))
	require.Equal(t, 1, out.Month)
	require.Equal(t, "bright", out.Kind)
	payload, err := json.Marshal(out)
	require.NoError(t, err)
	var fields map[string]any
	require.NoError(t, json.Unmarshal(payload, &fields))
	require.Equal(t, float64(1), fields["month"])
	require.Equal(t, "bright", fields["kind"])
	require.Equal(t, float64(domain.MushiBright), fields["category"])
}
