import type { GoStopResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Go-Stop (ゴーストップ), or null when
 * no suggestion is available.
 *
 * Go-Stop's hint is computed entirely by the Go backend and surfaced on the
 * response's `hint` field (with a `reason` i18n suffix such as `capture`,
 * `discard_low`, `go_lowscore`, or `stop_secure`). This adapter re-maps that
 * server hint into the frontend HintResult shape so the shared
 * {@link hooks/useGameHint.useGameHint | useGameHint} tooltip can render it. During the GoDecision phase the
 * hint points at the go/stop choice; during Play it points at which card to
 * play.
 */
export function getGoStopHint(state: GoStopResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: state.phase === 1 ? 'decide' : 'play',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
