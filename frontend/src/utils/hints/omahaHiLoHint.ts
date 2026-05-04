import type { OmahaResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { getHoldemBaseHint } from './holdemBaseHint';

/** Returns a frontend HintResult for Omaha Hi-Lo (8 or Better), or null
 * if no suggestion available. Reuses the Hold'em base hint for hi-side
 * decisioning; lo-side strategy is intentionally omitted because A-2-3
 * "low draw" reads require board texture analysis the base hint doesn't
 * model — a tighter heuristic can be layered in later. */
export function getOmahaHiLoHint(state: OmahaResponse): HintResult | null {
  return getHoldemBaseHint(state);
}
