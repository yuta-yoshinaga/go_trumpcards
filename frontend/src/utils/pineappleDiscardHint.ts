import type { Card } from '../types/card';

/**
 * A qualitative feature of the two hole cards a Pineapple player would keep
 * after discarding the third one, judged without any community cards:
 *
 * - `pair` — the two cards share the same rank.
 * - `suited` — the two cards share the same suit.
 * - `connector` — the two ranks are adjacent (Ace counts as both low and high,
 *   so A-2 and A-K are connectors).
 * - `highcard` — none of the above.
 *
 * `pair` is exclusive (a pair is neither suited nor a connector); `suited` and
 * `connector` can co-occur (a suited connector), so callers receive an array.
 */
export type PineappleKeepFeature = 'pair' | 'suited' | 'connector' | 'highcard';

/** True when the two ranks are adjacent, treating an Ace (value 1) as both low
 * (adjacent to 2) and high (adjacent to King, value 13). */
function isConnector(v1: number, v2: number): boolean {
  const expand = (v: number) => (v === 1 ? [1, 14] : [v]);
  for (const a of expand(v1)) {
    for (const b of expand(v2)) {
      if (Math.abs(a - b) === 1) return true;
    }
  }
  return false;
}

/**
 * Classify the two Pineapple hole cards that would remain after a discard,
 * using only the cards themselves (no board). Returns every applicable
 * {@link PineappleKeepFeature}, in a stable order (pair, suited, connector,
 * highcard). A pair short-circuits to `['pair']`; otherwise `highcard` is the
 * fallback when neither suited nor connector applies.
 */
export function pineappleKeepFeatures(a: Card, b: Card): PineappleKeepFeature[] {
  if (a.value === b.value) return ['pair'];
  const features: PineappleKeepFeature[] = [];
  if (a.design === b.design) features.push('suited');
  if (isConnector(a.value, b.value)) features.push('connector');
  if (features.length === 0) features.push('highcard');
  return features;
}
