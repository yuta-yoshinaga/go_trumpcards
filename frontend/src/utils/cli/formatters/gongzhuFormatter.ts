import type { GongZhuResponse } from '../../../types/card';
import {
  formatCard,
  formatHeader,
  formatIndexedCards,
  formatPlayerName,
  formatSeparator,
  isRequestedHint,
} from '../formatterBase';

/** Build the exposure summary line (e.g. "♠Q, ♦J" or "none"). */
function exposureSummary(state: GongZhuResponse): string {
  const parts: string[] = [];
  if (state.exposed.pig) parts.push('♠Q');
  if (state.exposed.sheep) parts.push('♦J');
  if (state.exposed.ace) parts.push('♥A');
  if (state.exposed.doubler) parts.push('♣10');
  return parts.length > 0 ? parts.join(', ') : 'none';
}

/** Format a Gong Zhu game state as terminal text. */
export function formatGongZhuState(state: GongZhuResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Gong Zhu'));
  lines.push(`round: ${state.roundNumber}  trick: ${state.trickNumber}  exposed: ${exposureSummary(state)}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    lines.push(`${name}: total=${p.cumulativeScore} round=${p.roundScore} cards=${p.cardCount} tricks=${p.trickCount}`);
    if (p.isHuman && p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.currentTrick.length > 0) {
    const trickParts = state.currentTrick.map((tc) => {
      const name = formatPlayerName(tc.playerIdx, state.players[tc.playerIdx]?.isHuman ?? false);
      return `${name}=${formatCard(tc.card)}`;
    });
    lines.push(`trick: ${trickParts.join(', ')}`);
  }

  if (state.phase === 0) {
    lines.push('Exposure phase: expose [i...] to reveal point cards (doubles their value)');
  }

  if (state.hint && isRequestedHint(state)) {
    const indices = state.hint.cardIndices ?? [];
    lines.push(`HINT: card indices [${indices.join(', ')}] (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag) {
    const winner = formatPlayerName(state.winnerIdx, state.players[state.winnerIdx]?.isHuman ?? false);
    lines.push(`Game Over! Winner: ${winner}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
