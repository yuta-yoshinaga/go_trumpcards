import type { EgyptianRatscrewResponse } from '../../../types/card';
import { EgyptianRatscrewPhase } from '../../../types/phases';

/** Format an Egyptian Ratscrew game state as terminal text. */
export function formatEgyptianRatscrewState(s: EgyptianRatscrewResponse): string {
  const lines: string[] = [];
  const phase = s.phase === EgyptianRatscrewPhase.GAME_END ? 'End' : 'Play';
  const top = s.topCard ? `${s.topCard.value}` : '--';
  const slap = s.isSlappable ? ' [SLAP!]' : '';
  const chance = s.chanceRemaining > 0 ? ` | Chance: ${s.chanceRemaining}` : '';
  lines.push(`Phase: ${phase} | Pile: ${s.centerPileSize} | Top: ${top}${slap}${chance} | Turn: P${s.currentTurnIdx}`);
  for (const p of s.players) {
    const tag = p.isHuman ? 'You' : 'CPU';
    lines.push(`${tag}: stock=${p.stockSize}`);
  }
  if (s.message) lines.push(s.message);
  return lines.join('\n');
}
