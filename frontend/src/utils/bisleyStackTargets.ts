import type { Card } from '../types/common';

/**
 * Returns the card values (ranks) that can be stacked on top of the given card
 * according to the Bisley tableau rules.
 *
 * domain_rule_quoted:
 * 	if len(dst) == 0 {
 * 		// 空き列は自由置き場ではない。ここを許すと事実上どの札でも掘り出せる。
 * 		return NewDomainErrorCode(ErrInvalidPlay, "bisley.errEmptyColumn", nil)
 * 	}
 * 	top := dst[len(dst)-1].Card
 * 	if top.GetDesign() != card.GetDesign() || abs(top.GetValue()-card.GetValue()) != 1 {
 * 		return NewDomainErrorCode(ErrInvalidPlay, "bisley.errNotAdjacentSameSuit", nil)
 * 	}
 *
 * edge_cases:
 * - 空列 (topCard === null) の場合は対象外 (errEmptyColumn) のため、空配列を返す。
 * - topCard.value が 1 (A) の場合は下方向 (0) は存在せず、2 のみを返す。
 * - topCard.value が 13 (K) の場合は上方向 (14) は存在せず、12 のみを返す。
 */
export function getBisleyStackTargets(topCard: Card | null): number[] {
  if (!topCard) {
    return [];
  }

  const val = topCard.value;
  const targets: number[] = [];

  if (val > 1) {
    targets.push(val - 1);
  }
  if (val < 13) {
    targets.push(val + 1);
  }

  return targets;
}
