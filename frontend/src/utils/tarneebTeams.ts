import type { TarneebPlayerData } from '../types/card';

/** A Tarneeb team with its member players and their aggregate round trick count. */
export interface TarneebTeamBreakdown {
  /** Team index (0 or 1). */
  team: number;
  /** Players belonging to this team, preserving their order in the source array. */
  members: TarneebPlayerData[];
  /** Total tricks the team took this round (sum of member `trickCount`s). */
  roundTricks: number;
}

/**
 * Group Tarneeb players by team, summing each team's round trick counts.
 *
 * Returns one entry per team index `0..teamCount-1` (teams with no members
 * yield an empty `members` array and a `roundTricks` of 0), so the result can
 * be indexed directly by team.
 */
export function groupTarneebPlayersByTeam(players: TarneebPlayerData[], teamCount: number): TarneebTeamBreakdown[] {
  return Array.from({ length: teamCount }, (_, team) => {
    const members = players.filter((p) => p.team === team);
    const roundTricks = members.reduce((sum, p) => sum + p.trickCount, 0);
    return { team, members, roundTricks };
  });
}
