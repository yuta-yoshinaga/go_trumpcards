import type { Card, CribbageSquaresResponse, CribbageSquaresScore } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = { 0: 'PLAYING', 1: 'COMPLETE' };

/** Total cells on the 4x4 grid. */
const TOTAL_CELLS = 16;

/** Render a single cell: card abbreviation or '..' when empty. */
function formatCell(card: Card | null): string {
  return card ? formatCard(card).padStart(3, ' ') : ' ..';
}

/**
 * Render one hand's breakdown, dropping the components that scored nothing.
 * "15:0 pairs:0 runs:0" hides the two points that actually landed.
 */
function formatBreakdown(d: CribbageSquaresScore | undefined): string {
  if (!d) return '';
  const parts: string[] = [];
  if (d.fifteens > 0) parts.push(`15s ${d.fifteens}`);
  if (d.pairs > 0) parts.push(`pairs ${d.pairs}`);
  if (d.runs > 0) parts.push(`runs ${d.runs}`);
  if (d.flush > 0) parts.push(`flush ${d.flush}`);
  if (d.nobs > 0) parts.push(`nobs ${d.nobs}`);
  return parts.length > 0 ? ` (${parts.join(', ')})` : '';
}

/** Format a Cribbage Squares game state as terminal text. */
export function formatCribbageSquaresState(state: CribbageSquaresResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Cribbage Squares'));
  lines.push(`phase: ${PHASE_NAMES[state.phase] ?? 'UNKNOWN'}  placed: ${state.placedCount}/${TOTAL_CELLS}`);
  lines.push(`current card: ${state.currentCard ? formatCard(state.currentCard) : '(none)'}`);
  // Saying the starter is face down beats omitting the line, which reads as a
  // rendering bug rather than a rule.
  lines.push(`starter: ${state.starter ? formatCard(state.starter) : '(face down)'}`);
  lines.push('');

  for (let r = 0; r < state.board.length; r++) {
    const cells = state.board[r].map((c) => formatCell(c.card)).join(' ');
    lines.push(`${cells}  | row${r}=${state.rowScores[r]}${formatBreakdown(state.rowDetails?.[r])}`);
  }
  lines.push('----------');
  for (let c = 0; c < state.colScores.length; c++) {
    lines.push(`col${c}=${state.colScores[c]}${formatBreakdown(state.colDetails?.[c])}`);
  }
  lines.push(`total: ${state.totalScore} / ${state.winScore}${state.isWin ? '  WIN' : ''}`);

  if (state.message) lines.push(state.message);

  lines.push(formatSeparator());
  return lines.join('\n');
}
