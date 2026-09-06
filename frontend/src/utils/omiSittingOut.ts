import type { OmiResponse } from '../types/card';

/** Returns the index of the player who is sitting out the current round,
 * or `null` if no one is. In Omi, when a player declares "Going Alone",
 * their teammate (same `team`, different `id`) skips the round. */
export function omiSittingOutIdx(
  state: Pick<OmiResponse, 'players' | 'goingAlone' | 'goingAlonePlayerIdx'>,
): number | null {
  if (!state.goingAlone) return null;
  const goer = state.players.find((p) => p.id === state.goingAlonePlayerIdx);
  if (!goer) return null;
  const partner = state.players.find((p) => p.team === goer.team && p.id !== goer.id);
  return partner ? partner.id : null;
}
