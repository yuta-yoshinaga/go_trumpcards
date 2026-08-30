import type { TarneebPlayerData } from '../types/card';

/** A Tarneeb team with its member players and their aggregate round trick count. */
export interface TarneebTeamBreakdown {
  /** Team index (0 or 1). */
  team: number;
  /** Players belonging to this team, preserving their order in the source array. */
  members: TarneebPlayerData[];
  /** Total tricks the team took this round (sum of member `trickCount`s). */
  roundTricks: number;
  /**
   * Team's round score (delta for this round).
   *
   * As implemented in `internal/domain/Tarneeb.go:388-406`, the domain duplicates
   * the team's score delta to each member's `roundScore` (via `SetRoundScore`),
   * rather than dividing it. Therefore, this value is taken from a single member's
   * `roundScore` (or 0 if the team has no members) instead of summing member scores,
   * to avoid doubling the team score.
   * (チームのデルタを各メンバーに複製した値なので合計しない)
   */
  roundScore: number;
}

/**
 * Group Tarneeb players by team, summing each team's round trick counts
 * and extracting each team's round score.
 *
 * Returns one entry per team index `0..teamCount-1` (teams with no members
 * yield an empty `members` array, a `roundTricks` of 0, and a `roundScore` of 0),
 * so the result can be indexed directly by team.
 */
export function groupTarneebPlayersByTeam(players: TarneebPlayerData[], teamCount: number): TarneebTeamBreakdown[] {
  return Array.from({ length: teamCount }, (_, team) => {
    const members = players.filter((p) => p.team === team);
    const roundTricks = members.reduce((sum, p) => sum + p.trickCount, 0);
    const roundScore = members[0]?.roundScore ?? 0;
    return { team, members, roundTricks, roundScore };
  });
}
