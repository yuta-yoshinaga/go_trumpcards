import type { AccordionPile } from '../types/card';

/** The two legal merge offsets in Accordion: pile-1 (adjacent) and pile-3 (three-back). */
export const ACCORDION_OFFSETS: readonly number[] = [1, 3] as const;

/**
 * Returns the indices of piles to the left of `fromIdx` (at offsets 1 and 3)
 * whose top card matches `fromIdx`'s top by suit OR rank — i.e. legal merge
 * targets in Accordion. Used by the hover affordance (#1887).
 */
export function accordionLegalTargets(piles: readonly AccordionPile[], fromIdx: number): number[] {
  const from = piles[fromIdx]?.cards[0];
  if (!from) return [];
  const targets: number[] = [];
  for (const offset of ACCORDION_OFFSETS) {
    const toIdx = fromIdx - offset;
    if (toIdx < 0) continue;
    const to = piles[toIdx]?.cards[0];
    if (!to) continue;
    if (to.design === from.design || to.value === from.value) {
      targets.push(toIdx);
    }
  }
  return targets;
}

/**
 * Returns the legal merge *offsets* (a subset of {@link ACCORDION_OFFSETS}, i.e.
 * `1` and/or `3`) available from `fromIdx`, derived from {@link accordionLegalTargets}.
 * Used to describe a selected pile's moves to assistive tech (#2596).
 */
export function accordionLegalOffsets(piles: readonly AccordionPile[], fromIdx: number): number[] {
  return accordionLegalTargets(piles, fromIdx).map((toIdx) => fromIdx - toIdx);
}
