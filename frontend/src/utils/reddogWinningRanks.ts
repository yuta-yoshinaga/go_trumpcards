import type { Card } from '../types/card';

/** Returns the ranks (Red Dog rank space — A=14) that the third card must
 * hit for the player to win, given the two initial cards. Returns an empty
 * array on pair or consecutive (no in-between ranks). */
export function reddogWinningRanks(initial: Card[]): number[] {
  if (initial.length !== 2) return [];
  const r1 = redDogRank(initial[0]);
  const r2 = redDogRank(initial[1]);
  const lo = Math.min(r1, r2);
  const hi = Math.max(r1, r2);
  return Array.from({ length: Math.max(0, hi - lo - 1) }, (_, i) => lo + 1 + i);
}

/** Maps a card value to Red Dog rank space (A=14, K=13, … 2=2). */
export function redDogRank(c: Card): number {
  return c.value === 1 ? 14 : c.value;
}

/** Short rank label suitable for chip display. */
export function rankLabel(rank: number): string {
  switch (rank) {
    case 14:
      return 'A';
    case 13:
      return 'K';
    case 12:
      return 'Q';
    case 11:
      return 'J';
    default:
      return String(rank);
  }
}
