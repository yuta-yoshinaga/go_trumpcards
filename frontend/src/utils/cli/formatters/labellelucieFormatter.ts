import type { LaBelleLucieResponse } from '../../../types/card';
import { LABELLELUCIE_MAX_REDEALS } from '../../../types/labellelucie';
import { LaBelleLuciePhase } from '../../../types/phases';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [LaBelleLuciePhase.PLAYING]: 'Playing',
  [LaBelleLuciePhase.GAME_CLEAR]: 'Game Clear',
  [LaBelleLuciePhase.GAME_OVER]: 'Game Over',
};

/** Format a La Belle Lucie game state as terminal text. */
export function formatLaBelleLucieState(state: LaBelleLucieResponse): string {
  const lines = [
    formatHeader('La Belle Lucie'),
    `phase: ${PHASE_NAMES[state.phase] ?? state.phase} | moves: ${state.moveCount} | redeals: ${LABELLELUCIE_MAX_REDEALS - state.redealsLeft}/${LABELLELUCIE_MAX_REDEALS}`,
    'Foundation:',
  ];
  lines.push(
    state.foundation.map((pile, i) => `  ${i}: ${pile.length ? formatCard(pile[pile.length - 1]) : '[  ]'}`).join('\n'),
  );
  lines.push('Fans:');
  state.fans.forEach((fan, i) => {
    const cards = fan
      .map((card, cardIndex) => `${formatCard(card)}${cardIndex === fan.length - 1 && state.movableFans[i] ? '*' : ''}`)
      .join('  ');
    lines.push(`  ${i}: ${cards || '[empty]'}`);
  });
  if (state.phase === LaBelleLuciePhase.PLAYING) {
    lines.push('Moves: use p <from> <to> or f <idx> for a fan marked *');
    const actions = [state.redealsLeft > 0 && 'rd', 'ac', state.canUndo && 'u', 'h', 'g'].filter(Boolean);
    lines.push(`Other actions: ${actions.join(', ')}`);
  } else {
    lines.push('Game finished. Use r to start a new game.');
  }
  if (state.hint && isRequestedHint(state)) {
    lines.push(
      state.hint.toFoundation
        ? `HINT: move fan ${state.hint.fromFan} to foundation`
        : `HINT: move fan ${state.hint.fromFan} to fan ${state.hint.toFan}`,
    );
  }
  if (state.messageCode === 'labellelucie.noHint') lines.push('No hint available.');
  if (state.message) lines.push(state.message);
  if (state.phase === LaBelleLuciePhase.GAME_CLEAR) lines.push('Congratulations! You win!');
  if (state.phase === LaBelleLuciePhase.GAME_OVER) lines.push('Game over.');
  lines.push(formatSeparator());
  return lines.join('\n');
}
