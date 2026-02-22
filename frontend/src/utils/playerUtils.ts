/** Return display name for a player. */
export function playerName(id: number, isHuman: boolean): string {
  return isHuman ? 'あなた' : `CPU ${id}`;
}

/** Look up a player's display name by array index. */
export function findPlayerName(players: { id: number; isHuman: boolean }[], idx: number): string {
  const p = players[idx];
  return p ? playerName(p.id, p.isHuman) : `Player ${idx}`;
}
