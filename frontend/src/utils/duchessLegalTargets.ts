import type { Card, DuchessTableauCard } from '../types/card';

/** Where the selected card may legally go in Duchess. */
export interface DuchessLegalTargets {
  /** Tableau column indices that accept the card. */
  tableau: Set<number>;
  /** Foundation indices that accept the card. */
  foundation: Set<number>;
}

/** Foundation index → suit, fixed so the layout does not move between deals. */
const FOUNDATION_SUITS = ['SPADE', 'CLOVER', 'HEART', 'DIAMOND'] as const;

/** A foundation is finished at thirteen cards. */
const FOUNDATION_TARGET = 13;

/** King follows Ace going down. Sync: `duchessPrevRank`. */
function prevRank(value: number): number {
  return value <= 1 ? 13 : value - 1;
}

/** Ace follows King going up. Sync: `duchessNextRank`. */
function nextRank(value: number): number {
  return value >= 13 ? 1 : value + 1;
}

/** Whether a card is red in Duchess. Sync: `duchessIsRed`. */
function isRed(design: Card['design']): boolean {
  return design === 'HEART' || design === 'DIAMOND';
}

/**
 * Returns the legal destinations for a selected Duchess card.
 *
 * Sync: `Duchess.canPlaceOnTableau` / `Duchess.canPlaceOnFoundation` /
 * `Duchess.emptyColumnIsReservedForReserve`.
 *
 * **タブローは色違いの 1 つ下。** A と K は地続き (A の下は K) なので、ここを
 * 普通の 1..13 で書くと折り返しの手が置けない列として表示される。
 *
 * **空き列はリザーブが残っている間、文字どおりリザーブ専用。**これは札そのもの
 * ではなく出どころで決まるので `fromZone` が要る。ドメインは
 * `MoveWasteToTableau` (299 行) と `MoveTableauToTableau` (363 行) の**両方**で
 * `duchess.errEmptyColumnReserveOnly` を返し、`MoveReserveToTableau` にだけ
 * この検査が無い。つまり除外すべきは「タブロー以外」ではなく「リザーブ以外」で、
 * ここを緩めるとウェイストの札に、押しても弾かれるだけのリングが出る。
 *
 * @param fromZone - The source zone of the selected card.
 */
export function duchessLegalTargets(
  tableau: readonly DuchessTableauCard[][],
  foundation: readonly Card[][],
  reserve: readonly Card[][],
  baseRank: number,
  awaitingBaseRank: boolean,
  card: Card | null | undefined,
  fromZone?: string,
): DuchessLegalTargets {
  const result: DuchessLegalTargets = { tableau: new Set(), foundation: new Set() };
  if (!card) return result;

  // **開始ランクが決まるまでは、組札どころかどの手も打てない。**
  // `Duchess.requireBaseChosen` は `MoveTableauToTableau` (Duchess.go:336) や
  // `MoveWasteToTableau` (286) を含む**移動 7 箇所すべて**の入口にあり、
  // `baseRank == 0` なら例外なく弾く。組札だけを黙らせてタブローにリングを出すと、
  // 配りによっては置ける先があるように見えて、押すとサーバに拒まれる。
  if (awaitingBaseRank || baseRank === 0) return result;

  const reserveRemaining = reserve.reduce((sum, fan) => sum + fan.length, 0);
  tableau.forEach((column, index) => {
    const top = column[column.length - 1]?.card;
    if (!top) {
      // An empty column is the reserve's exit while reserve cards remain: only
      // a reserve card may consume that slot, waste and tableau alike are shut out.
      if (reserveRemaining === 0 || fromZone === 'reserve') result.tableau.add(index);
      return;
    }
    if (card.value === prevRank(top.value) && isRed(card.design) !== isRed(top.design)) {
      result.tableau.add(index);
    }
  });

  foundation.forEach((pile, index) => {
    if (FOUNDATION_SUITS[index] !== card.design) return;
    if (pile.length === 0) {
      if (card.value === baseRank) result.foundation.add(index);
      return;
    }
    if (pile.length >= FOUNDATION_TARGET) return;
    if (card.value === nextRank(pile[pile.length - 1].value)) result.foundation.add(index);
  });

  return result;
}
