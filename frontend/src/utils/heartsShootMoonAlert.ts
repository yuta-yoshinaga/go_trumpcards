import type { HeartsPlayerData } from '../types/card';

/** Round-points threshold above which a single-player monopoly counts as
 * "shooting the moon in progress" (♥×13 + ♠Q=13 → up to 26 total). */
const SHOOT_THE_MOON_ALERT_THRESHOLD = 13;

/** Returns the player index that appears to be shooting the moon, or `null`
 * if no one is dominating yet. The criterion is: the total round points
 * distributed so far is at least the threshold AND a single player holds
 * all of them.
 *
 * Limitation: `roundScore` collapses penalty count + Omnibus J♦ bonus
 * (-10) into a single number, so we can't reliably tell whether a
 * negative score hides penalty cards. When any player has a negative
 * round score (only happens with the Omnibus variant) we conservatively
 * skip the alert rather than risk a false positive. */
export function shootTheMoonAlertIdx(players: readonly HeartsPlayerData[]): number | null {
  let total = 0;
  let leaderIdx: number | null = null;
  let leaderScore = 0;
  for (const p of players) {
    if (p.roundScore < 0) return null;
    if (p.roundScore === 0) continue;
    total += p.roundScore;
    if (p.roundScore > leaderScore) {
      leaderScore = p.roundScore;
      leaderIdx = p.id;
    }
  }
  if (leaderIdx == null) return null;
  if (total < SHOOT_THE_MOON_ALERT_THRESHOLD) return null;
  if (leaderScore < total) return null;
  return leaderIdx;
}
