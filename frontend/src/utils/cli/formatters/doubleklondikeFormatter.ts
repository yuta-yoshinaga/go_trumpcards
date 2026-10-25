import type { DoubleKlondikeResponse } from '../../../types/card';
import { DoubleKlondikePhase } from '../../../types/phases';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

const PHASE_NAMES: Readonly<Record<number, string>> = {
  [DoubleKlondikePhase.PLAYING]: 'Playing',
  [DoubleKlondikePhase.GAME_CLEAR]: 'Game Clear',
  [DoubleKlondikePhase.GAME_OVER]: 'Game Over',
};

/** Format a Double Klondike board for terminal play. */
export function formatDoubleKlondikeState(state: DoubleKlondikeResponse): string {
  const lines = [
    formatHeader('Double Klondike'),
    `phase: ${PHASE_NAMES[state.phase] ?? state.phase}`,
    `moves: ${state.moveCount} | stock: ${state.stockCount}`,
  ];
  const wasteTop = state.waste[state.waste.length - 1];
  lines.push(`waste: ${wasteTop ? `[${state.waste.length - 1}]${formatCard(wasteTop)}` : '[empty]'}`);
  lines.push('Foundations:');
  state.foundation.forEach((pile, i) => {
    const top = pile[pile.length - 1];
    lines.push(`  [${i}] ${top ? formatCard(top) : '[empty]'}`);
  });
  lines.push('Tableau (columns top at right; use p <col> <idx> <to>):');
  state.tableau.forEach((column, col) => {
    const cards = column
      .map((cell, idx) => `${cell.faceUp && cell.card ? `[${idx}]${formatCard(cell.card)}` : `[${idx}]XX`}`)
      .join(' ');
    const movable = Object.entries(state.legalTargets).some(
      ([key, targets]) =>
        key.startsWith(`tableau-${col}-`) && (targets.tableau.length > 0 || targets.foundation.length > 0),
    );
    lines.push(`${movable ? '*' : ' '}[${col}] ${cards || '[empty]'}`);
  });
  if (state.phase === DoubleKlondikePhase.PLAYING) {
    const commands = [
      'd (draw)',
      'ac (auto)',
      ...(state.canUndo ? ['u (undo)'] : []),
      'h (hint)',
      'g (give up)',
      'r (reset)',
    ];
    lines.push(`Available commands: ${commands.join(', ')}`);
    lines.push('Moves: p waste <col|f>, p <col> <idx> <to|f>');
  } else {
    lines.push('Available command: r (reset)');
    if (state.phase === DoubleKlondikePhase.GAME_CLEAR) lines.push('All cards moved to foundations. Game clear!');
    if (state.phase === DoubleKlondikePhase.GAME_OVER) lines.push('No more moves. Game over.');
  }
  if (state.messageCode === 'doubleklondike.noHint') lines.push('No hint available.');
  if (state.hint && isRequestedHint(state)) {
    const from =
      state.hint.fromZone === 'waste' ? 'waste' : `column ${state.hint.fromCol} card ${state.hint.cardIndex}`;
    const to = state.hint.toZone === 'foundation' ? 'a foundation' : `column ${state.hint.toCol}`;
    lines.push(`HINT: Move ${from} to ${to}.`);
  }
  if (state.phase === DoubleKlondikePhase.PLAYING && state.message && !state.messageCode?.startsWith('doubleklondike.'))
    lines.push(state.message);
  lines.push(formatSeparator());
  return lines.join('\n');
}
