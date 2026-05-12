import type { SpideretteResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Frontend strategic hint for Spiderette is intentionally a `null` stub.
 *
 * The backend already returns the next-mechanical-move hint via the API; a
 * frontend strategic hint would duplicate that without adding meaningful
 * judgement, since Spiderette has no scoring or memory dimension beyond what
 * the backend already evaluates.
 */
export function getSpideretteHint(_state: SpideretteResponse | null | undefined): HintResult | null {
  return null;
}
