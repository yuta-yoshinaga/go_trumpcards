import type { DragonTigerResponse as _DragonTigerResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a Dragon Tiger hint. Dragon Tiger is a pure-luck game (no decisions
 * after the bet is placed), so there is no actionable hint. Returning null
 * keeps the hint API uniform across games (#1684 — checklist item 14 stub).
 */
export function getDragontigerHint(_state: _DragonTigerResponse): HintResult | null {
  return null;
}
