import type { CuckooResponse } from '../../../types/card';
import { formatCard, formatHeader, formatPlayerName, formatSeparator } from '../formatterBase';

const PHASE_NAMES = ['Turn', 'Refuse', 'RoundEnd', 'GameEnd'];

/** Format a Cuckoo game state as terminal text. */
export function formatCuckooState(state: CuckooResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Cuckoo'));
  lines.push(`round: ${state.roundNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`);
  lines.push(`dealer: ${formatPlayerName(state.dealerIdx, state.dealerIdx === 0)}  stock: ${state.stockCount}`);
  lines.push('----------');

  state.players.forEach((p, i) => {
    const name = formatPlayerName(i, p.isHuman);
    const hearts = '♥'.repeat(Math.max(0, p.lives));
    const status = p.isEliminated ? ' (out)' : p.isCurrentTurn ? ' <- turn' : '';
    const king = p.kingRevealed ? ' [K revealed]' : '';
    const card = p.card ? ` ${formatCard(p.card)}` : '';
    lines.push(`${name}: ${hearts || '-'}${card}${status}${king}`);
  });
  lines.push('----------');

  if (state.phase === 0 && state.currentPlayerIdx === 0 && !state.gameEndFlag) {
    lines.push('(your turn — keep with "k" or swap with "s")');
  } else if (state.phase === 1 && state.pendingSwapTo === 0) {
    lines.push('(someone wants to swap with you — refuse with "rf" or accept with "ac")');
  } else if (state.phase === 2) {
    lines.push('(round over — advance with "nr")');
  }

  if (state.roundLosers.length > 0) {
    const losers = state.roundLosers.map((idx) => formatPlayerName(idx, idx === 0)).join(', ');
    lines.push(`lowest card lost a life: ${losers}`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag && state.winnerIdx >= 0) {
    lines.push(`Game Over! Winner: ${formatPlayerName(state.winnerIdx, state.winnerIdx === 0)}`);
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
