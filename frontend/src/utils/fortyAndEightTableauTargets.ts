import type { Card } from '../types/card';
import type { FortyAndEightTableauCard } from '../types/games/fortyandeight';

/**
 * Computes the set of tableau column indices that `card` can legally move to,
 * given the current `tableau` columns.
 *
 * Sync: `FortyAndEight.canPlaceOnTableau` (`internal/domain/FortyAndEight.go`).
 *
 * - ファンデーションと向きが逆: ファンデーションは同スート昇順 (+1) だが、
 *   タブローは同スート降順 (-1)。(`card.design === topCard.design && card.value === topCard.value - 1`)
 * - 空列は任意: 空の列にはどのカードでも置ける。
 * - 選択元の列は自動的に除外: 選択中カードが置かれている列の最上段カードは
 *   自分自身であるため、`card.value === topCard.value - 1` は `value === value - 1`
 *   となり必ず偽になる。そのため選択元の列を特別扱いして除外する必要はない。
 *
 * @param card - The selected source card (waste top card or a tableau card), or `null`/`undefined`.
 * @param tableau - The eight tableau columns.
 * @returns A set of 0-based tableau column indices the card can be placed on (empty when none).
 */
export function fortyAndEightTableauTargets(
  card: Card | null | undefined,
  tableau: FortyAndEightTableauCard[][],
): Set<number> {
  const targets = new Set<number>();
  if (!card) return targets;

  tableau.forEach((col, idx) => {
    if (col.length === 0) {
      targets.add(idx);
      return;
    }
    const topCard = col[col.length - 1]?.card;
    if (topCard && card.design === topCard.design && card.value === topCard.value - 1) {
      targets.add(idx);
    }
  });

  return targets;
}
