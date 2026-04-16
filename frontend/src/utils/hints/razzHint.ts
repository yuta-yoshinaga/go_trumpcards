import type { HintResult } from '../../types/hint';

/** Razz hint evaluator -- returns null (no decisional hint). */
export function razzHint(_state: unknown): HintResult | null {
  return null;
}
