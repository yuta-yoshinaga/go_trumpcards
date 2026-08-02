import type { AnacondaResponse } from '../../../types/card';
import { formatHeader, formatIndexedCards, formatPlayerName, formatSeparator, isRequestedHint } from '../formatterBase';

const PHASE_NAMES = ['Pass', 'Set', 'Roll', 'Result'];

/** Format an Anaconda (Pass the Trash) game state as terminal text. */
export function formatAnacondaState(state: AnacondaResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Anaconda'));
  lines.push(`round: ${state.roundNumber}  phase: ${PHASE_NAMES[state.phase] ?? state.phase}`);
  lines.push(`pot: ${state.pot}  ante: ${state.ante}  bet: ${state.currentBet}`);
  if (state.phase === 0) lines.push(`pass: ${state.passCount} card(s) left`);
  if (state.phase === 2) lines.push(`revealed: ${state.rollIndex}/5`);
  lines.push('');

  for (const p of state.players) {
    const name = formatPlayerName(p.id, p.isHuman);
    const status = p.out
      ? 'OUT'
      : p.folded
        ? 'folded'
        : p.isWinner
          ? 'WINNER'
          : p.id === state.currentPlayer
            ? 'to act'
            : 'waiting';
    const handText = p.handName ? `  hand=${p.handName}` : '';
    lines.push(`${name}: chips=${p.chips} bet=${p.roundBet} [${status}]${handText}`);
    if (p.cards.length > 0) {
      lines.push(`  ${formatIndexedCards(p.cards)}`);
    }
  }
  lines.push('----------');

  if (state.hint && isRequestedHint(state)) {
    const idx =
      state.hint.cardIndices && state.hint.cardIndices.length > 0 ? ` [${state.hint.cardIndices.join(' ')}]` : '';
    lines.push(`HINT: ${state.hint.action}${idx} (${state.hint.reason})`);
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
