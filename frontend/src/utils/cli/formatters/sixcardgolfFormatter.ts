import type { SixCardGolfResponse } from '../../../types/card';

/** Format Six Card Golf state for CLI display. */
export function formatSixCardGolfState(s: SixCardGolfResponse): string {
  return `Round ${s.roundNumber}/${s.totalRounds} | Phase ${s.phase}`;
}
