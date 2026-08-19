//go:build !js || !wasm || extra

package controller

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
)

// machiavelliSuitLetters maps the CUI suit letters to card designs.
var machiavelliSuitLetters = map[byte]int{
	's': domain.CardDesignSpade,
	'c': domain.CardDesignClover,
	'h': domain.CardDesignHeart,
	'd': domain.CardDesignDiamond,
}

// machiavelliRankLetters maps the face-card letters to values.
var machiavelliRankLetters = map[string]int{"a": 1, "j": 11, "q": 12, "k": 13}

// parseMachiavelliCard parses one "<suit><rank>" token (e.g. "s5", "hK").
func parseMachiavelliCard(token string) (domain.MachiavelliCardRef, bool) {
	token = strings.ToLower(strings.TrimSpace(token))
	if len(token) < 2 {
		return domain.MachiavelliCardRef{}, false
	}
	design, ok := machiavelliSuitLetters[token[0]]
	if !ok {
		return domain.MachiavelliCardRef{}, false
	}
	rank := token[1:]
	value, ok := machiavelliRankLetters[rank]
	if !ok {
		n, err := strconv.Atoi(rank)
		if err != nil || n < 1 || n > domain.CardValueMax {
			return domain.MachiavelliCardRef{}, false
		}
		value = n
	}
	return domain.MachiavelliCardRef{Design: design, Value: value}, true
}

// parseMachiavelliRearrange parses the `ra` arguments: the whole new table as
// ";"-separated groups of cards, then "/", then the hand indices to play.
//
// 場を組み替える手は「新しい場の全体」を渡す必要がある (applyPlay が保存則を
// 見るため)。手札から最低 1 枚出すのもルールなので、どちらか欠けたら弾く。
func parseMachiavelliRearrange(args []string) ([][]domain.MachiavelliCardRef, []int, bool) {
	joined := strings.Join(args, " ")
	left, right, found := strings.Cut(joined, "/")
	if !found {
		return nil, nil, false
	}
	var refs [][]domain.MachiavelliCardRef
	for _, group := range strings.Split(left, ";") {
		group = strings.TrimSpace(group)
		if group == "" {
			continue
		}
		var meld []domain.MachiavelliCardRef
		for _, token := range strings.Split(group, ",") {
			if strings.TrimSpace(token) == "" {
				continue
			}
			ref, ok := parseMachiavelliCard(token)
			if !ok {
				return nil, nil, false
			}
			meld = append(meld, ref)
		}
		if len(meld) == 0 {
			return nil, nil, false
		}
		refs = append(refs, meld)
	}
	if len(refs) == 0 {
		return nil, nil, false
	}
	var handIndices []int
	for _, token := range strings.Split(right, ",") {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		n, err := strconv.Atoi(token)
		if err != nil || n < 0 {
			return nil, nil, false
		}
		handIndices = append(handIndices, n)
	}
	if len(handIndices) == 0 {
		return nil, nil, false
	}
	return refs, handIndices, true
}
