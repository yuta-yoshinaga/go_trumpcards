import type { OmahaResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { getHoldemBaseHint } from './holdemBaseHint';

/** Returns a frontend HintResult for Omaha Hold'em, or null if no suggestion available. */
export function getOmahaHint(state: OmahaResponse): HintResult | null {
  return getHoldemBaseHint(state);
}
