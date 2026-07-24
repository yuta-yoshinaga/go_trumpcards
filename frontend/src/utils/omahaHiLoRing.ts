/** Which showdown highlight a card qualifies for under the Hi-Lo split. */
export type HiLoHighlight = 'both' | 'hi' | 'lo' | 'none';

/** A card's Hi-Lo highlight category plus the Tailwind ring classes to render it. */
export interface HiLoRingStyle {
  /** Whether the card is used by the Hi hand, the Lo hand, both, or neither. */
  category: HiLoHighlight;
  /** Tailwind ring/utility classes conveying the highlight (empty for `none`). */
  ring: string;
}

/** Hi-only ring: raised green outline. */
const HI_RING = '-translate-y-1 ring-2 ring-ds-success motion-safe:animate-pulse';
/** Lo-only ring: blue outline. */
const LO_RING = 'ring-2 ring-ds-info motion-safe:animate-pulse';
/** Dual ring: raised outer green + inner blue offset, so a card used by BOTH
 * the Hi and the Lo hand shows both attributes at once instead of Lo winning. */
const BOTH_RING = '-translate-y-1 ring-2 ring-ds-success ring-offset-2 ring-offset-ds-info motion-safe:animate-pulse';

/**
 * Resolve the showdown highlight for a single card given whether it belongs to
 * the player's best Hi five and/or their qualifying Lo five. A card used by both
 * halves gets a distinct dual ring (outer green + inner blue) rather than
 * collapsing to the Lo (blue) style, so scoop-level hands are verifiable.
 *
 * @param inHi - The card is part of the best Hi five-card hand.
 * @param inLo - The card is part of the qualifying Lo five-card hand.
 * @returns The highlight category and the ring classes to apply.
 */
export function hiLoRingStyle(inHi: boolean, inLo: boolean): HiLoRingStyle {
  if (inHi && inLo) return { category: 'both', ring: BOTH_RING };
  if (inHi) return { category: 'hi', ring: HI_RING };
  if (inLo) return { category: 'lo', ring: LO_RING };
  return { category: 'none', ring: '' };
}
