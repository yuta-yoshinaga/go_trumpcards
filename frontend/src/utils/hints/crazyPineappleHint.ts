import type { PineappleResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { getPineappleHint } from './pineappleHint';

/** Returns a frontend HintResult for Crazy Pineapple Poker, or null if no
 * suggestion available. The hint logic is identical to standard Pineapple —
 * only the discard timing differs, and that is enforced server-side. */
export function getCrazyPineappleHint(state: PineappleResponse): HintResult | null {
  return getPineappleHint(state);
}
