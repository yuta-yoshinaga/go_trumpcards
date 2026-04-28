import type { ShitheadResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Shithead frontend hint stub.
 *
 * Hint logic is non-trivial because of the multi-layer pile and magic-card
 * effects; for now no client-side hint is computed. The function exists so
 * `useGameHint` can register `shithead` in its hintFactories map.
 */
export function getShitheadHint(_state: ShitheadResponse): HintResult | null {
  return null;
}
