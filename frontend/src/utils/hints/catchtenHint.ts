import type { CatchTenResponse } from '../../types/card';
import type { HintConfidence, HintResult } from '../../types/hint';

/** Map backend hint reason keys to frontend i18n keys + confidence. */
const REASON_MAP: Record<string, { reason: string; confidence: HintConfidence }> = {
  lead_strong: { reason: 'hint.leadStrategic', confidence: 'strong' },
  follow_suit: { reason: 'hint.followSuit', confidence: 'strong' },
  trump_cut: { reason: 'hint.trumpCut', confidence: 'moderate' },
  discard_high: { reason: 'hint.discardLowest', confidence: 'moderate' },
};

const FALLBACK = { reason: 'hint.leadStrategic', confidence: 'moderate' as HintConfidence };

/** Returns a Catch the Ten frontend hint derived from the backend hint, or null. */
export function getCatchTenHint(state: CatchTenResponse): HintResult | null {
  if (state.gameEndFlag) return null;
  if (!state.hint || state.hint.cardIndex == null) return null;

  const mapped = REASON_MAP[state.hint.reason] ?? FALLBACK;
  return { targetAction: 'play', reason: mapped.reason, confidence: mapped.confidence };
}
