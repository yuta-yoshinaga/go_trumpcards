import type { Card, PokerSquaresResponse } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = { 0: 'PLAYING', 1: 'COMPLETE' };

/** Render a single cell: card abbreviation or '..' when empty. */
function formatCell(card: Card | null): string {
  return card ? formatCard(card).padStart(3, ' ') : ' ..';
}

/** Format a Poker Squares game state as terminal text. */
export function formatPokerSquaresState(state: PokerSquaresResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Poker Squares'));
  lines.push(`phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}  placed: ${state.placedCount}/25`);
  lines.push(`current card: ${state.currentCard ? formatCard(state.currentCard) : '(none)'}`);
  lines.push('');

  for (let r = 0; r < state.board.length; r++) {
    const row = state.board[r];
    const cells = row.map((c) => formatCell(c.card)).join(' ');
    lines.push(`${cells}  | row${r}=${state.rowScores[r]}`);
  }
  lines.push('----------');
  const colParts = state.colScores.map((s, i) => `col${i}=${s}`).join(' ');
  lines.push(colParts);
  lines.push(`total: ${state.totalScore}`);

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
