import type { ShamrocksResponse } from '../../../types/card';
import { ShamrocksPhase } from '../../../types/phases';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

const PHASE_NAMES: Record<number, string> = {
  [ShamrocksPhase.PLAYING]: 'Playing',
  [ShamrocksPhase.GAME_CLEAR]: 'GameClear',
  [ShamrocksPhase.GAME_OVER]: 'GameOver',
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
    const prefix = `fan ${String(i).padStart(2, ' ')}:`;
    if (fan.length === 0) {
      lines.push(`${prefix} [empty]`);
    } else {
      const cardStrs = fan.map((c, idx) => `${formatCard(c)}${idx === fan.length - 1 && isMovable ? '*' : ''}`);
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

  if (state.phase === ShamrocksPhase.PLAYING && Object.keys(state.legalTargets).length === 0) {
    lines.push('Deadlocked - no legal moves remain');
  }

  if (state.message && state.phase !== ShamrocksPhase.GAME_CLEAR && state.phase !== ShamrocksPhase.GAME_OVER) {
    lines.push(state.message);
  }

  if (state.phase === ShamrocksPhase.GAME_CLEAR) {
    lines.push('Congratulations! You win!');
  } else if (state.phase === ShamrocksPhase.GAME_OVER) {
    lines.push('Game Over!');
  }

  // Available commands guide
  if (state.phase === ShamrocksPhase.PLAYING) {
    const hasFanMoves = Object.values(state.legalTargets).some((t) => (t.fan?.length ?? 0) > 0);
    const hasFoundationMoves = Object.values(state.legalTargets).some((t) => (t.foundation?.length ?? 0) > 0);
    if (hasFanMoves && hasFoundationMoves) {
      lines.push('Moves: use p <from> <to> or f <idx> for a fan marked *');
    } else if (hasFanMoves) {
      lines.push('Moves: use p <from> <to> for a fan marked *');
    } else if (hasFoundationMoves) {
      lines.push('Moves: use f <idx> for a fan marked *');
    }

    const actions = [hasFoundationMoves && 'ac', state.canUndo && 'u', 'h', 'g', 'r'].filter(Boolean);
    lines.push(`Other actions: ${actions.join(', ')}`);
  } else {
    lines.push('Game finished. Use r to start a new game.');
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
