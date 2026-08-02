import type { PrimeroResponse } from '../../../types/card';
import { formatHeader, formatIndexedCards, formatPlayerName, formatSeparator, isRequestedHint } from '../formatterBase';

const PHASE_NAMES = ['Betting', 'Result'];

/** Format a Primero game state as terminal text. */
export function formatPrimeroState(state: PrimeroResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Primero'));
  lines.push(`round: ${state.roundNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`);
  lines.push(`pot: ${state.pot}  ante: ${state.ante}  bet: ${state.currentBet}`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const status = p.out ? 'OUT' : p.isWinner ? 'WINNER' : p.folded ? 'folded' : 'active';
    const handText = p.handName ? `  hand=${p.handName}` : '';
    lines.push(`${name}: chips=${p.chips} bet=${p.roundBet} [${status}]${handText}`);
    if (p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.hint && isRequestedHint(state)) {
    lines.push(`HINT: ${state.hint.action} (${state.hint.reason})`);
  }

  if (state.message) lines.push(state.message);
  if (state.gameEndFlag && state.matchWinnerIdx >= 0) {
    lines.push(
      `Game Over! Winner: ${formatPlayerName(state.matchWinnerIdx, state.players[state.matchWinnerIdx]?.isHuman ?? false)}`,
    );
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
