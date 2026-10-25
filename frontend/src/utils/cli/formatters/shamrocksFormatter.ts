import type { ShamrocksResponse } from '../../../types/card';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  0: 'Playing',
  1: 'GameClear',
  2: 'GameOver',
};

/** Format a Shamrocks game state as terminal text. */
export function formatShamrocksState(state: ShamrocksResponse): string {
  const lines: string[] = [];

  lines.push(formatHeader('Shamrocks'));
  const phaseName = PHASE_NAMES[state.phase] ?? `Phase ${state.phase}`;
  lines.push(`phase: ${phaseName}  moves: ${state.moveCount}  undo: ${state.canUndo ? 'yes' : 'no'}`);

  // Foundations (4 piles)
  const fnd = state.foundation.map((pile, i) => {
    const top = pile.length > 0 ? formatCard(pile[pile.length - 1]) : '[  ]';
    return `f${i}:${top}`;
  });
  lines.push(`foundation: ${fnd.join('  ')}`);
  lines.push('----------');

  // Fans
  const movableFans = new Set(Object.keys(state.legalTargets).map((key) => Number(key.slice('fan-'.length))));
  for (let i = 0; i < state.fans.length; i++) {
    const fan = state.fans[i];
    const isMovable = movableFans.has(i);
    const marker = isMovable ? '*' : ' ';
    const prefix = `fan ${String(i).padStart(2, ' ')}${marker}:`;
    if (fan.length === 0) {
      lines.push(`${prefix} [empty]`);
    } else {
      const cardStrs = fan.map((c, idx) => `[${idx}]${formatCard(c)}`);
      lines.push(`${prefix} ${cardStrs.join(' ')}`);
    }
  }
  lines.push('----------');

  if (state.hint && isRequestedHint(state)) {
    if (state.hint.toFoundation) {
      lines.push(`HINT: fan ${state.hint.fromFan} → foundation`);
    } else {
      lines.push(`HINT: fan ${state.hint.fromFan} → fan ${state.hint.toFan}`);
    }
  } else if (state.messageCode === 'shamrocks.noHint') {
    lines.push('No hint available.');
  }

  if (state.phase === 0 && Object.keys(state.legalTargets).length === 0) {
    lines.push('Deadlocked - no legal moves remain');
  }

  if (state.message) {
    lines.push(state.message);
  }

  if (state.phase === 1) {
    lines.push('Congratulations! You win!');
  } else if (state.phase === 2) {
    lines.push('Game Over!');
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
