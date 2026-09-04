import type { BeggarMyNeighbourResponse } from '../../../types/card';
import { BeggarMyNeighbourPhase } from '../../../types/phases';
import { formatHeader, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [BeggarMyNeighbourPhase.PLAY]: 'Play',
  [BeggarMyNeighbourPhase.PAY_PENALTY]: 'PayPenalty',
  [BeggarMyNeighbourPhase.COLLECT]: 'Collect',
  [BeggarMyNeighbourPhase.GAME_END]: 'End',
};

/** Format a Beggar-My-Neighbour game state as terminal text. */
export function formatBeggarMyNeighbourState(state: BeggarMyNeighbourResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Beggar-My-Neighbour'));
  const phaseName = PHASE_NAMES[state.phase] ?? 'Unknown';
  lines.push(
    `Phase: ${phaseName} | Pile: ${state.centralPileSize} | Penalty: ${state.penaltyRemaining} | Round: ${state.roundsPlayed}`,
  );
  // **誰が払っているかは CLI にも要る** ── Web GUI と同じで、残り枚数だけでは
  // 自分が払う側か相手が払う側か分からない (#6478)。
  if (state.phase === BeggarMyNeighbourPhase.PAY_PENALTY && state.penaltyOwnerIdx >= 0) {
    lines.push(`Paying: ${state.penaltyOwnerIdx === 0 ? 'You' : `CPU ${state.penaltyOwnerIdx}`}`);
  }
  for (const p of state.players) {
    const tag = p.isHuman ? 'You' : 'CPU';
    lines.push(`${tag}: draw=${p.drawPileSize} discard=${p.discardPileSize} total=${p.totalCards}`);
  }
  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
