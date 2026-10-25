import type { TehonbikiResponse } from '../../../types/games/tehonbiki';
import { isRequestedHint } from '../formatterBase';
/** Formats Tehonbiki state as terminal text. */
export function formatTehonbikiState(s: TehonbikiResponse): string {
  return [
    `Tehonbiki`,
    `Phase: ${s.phase}`,
    `Round: ${s.roundNumber} (chips: ${s.chips})`,
    `Wager: ${s.betType} ${s.numbers.join(',')}`,
    s.parentCard === undefined ? '' : `Parent card: ${s.parentCard}`,
    s.result === 0 ? '' : `Result: ${s.result} (payout ${s.payout})`,
    s.hint && isRequestedHint(s) ? `HINT: ${s.hint.reason}` : '',
    s.message ?? '',
  ]
    .filter(Boolean)
    .join('\n');
}
