import type { TehonbikiResponse } from '../../../types/games/tehonbiki';
/** Formats Tehonbiki state as terminal text. */
export function formatTehonbikiState(s: TehonbikiResponse): string {
  return [
    `Tehonbiki`,
    `Phase: ${s.phase}`,
    `Round: ${s.roundNumber} (chips: ${s.chips})`,
    `Wager: ${s.betType} ${s.numbers.join(',')}`,
    s.parentCard === undefined ? '' : `Parent card: ${s.parentCard}`,
    s.result === 0 ? '' : `Result: ${s.result} (payout ${s.payout})`,
  ]
    .filter(Boolean)
    .join('\n');
}
