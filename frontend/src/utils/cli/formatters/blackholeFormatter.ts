import type { BlackHoleResponse } from '../../../types/games/blackhole';
import { BlackHolePhase } from '../../../types/phases';
import { valueName } from '../../cardUtils';
import { getAcceptableRanks, getLegalFans } from '../commands/blackholeCommands';
import { formatCard, formatHeader, formatSeparator, isRequestedHint } from '../formatterBase';

const TOTAL_CARDS = 52;

const PHASE_NAMES: Record<number, string> = {
  [BlackHolePhase.PLAYING]: 'Playing',
  [BlackHolePhase.GAME_CLEAR]: 'Game Clear',
  [BlackHolePhase.GAME_OVER]: 'Game Over',
};

/**
 * Format a Black Hole game state as terminal text.
 */
export function formatBlackHoleState(state: BlackHoleResponse): string {
  const lines: string[] = [];
  lines.push(formatHeader('Black Hole'));

  const phaseName = PHASE_NAMES[state.phase] ?? `Phase ${state.phase}`;
  lines.push(`moves: ${state.moveCount}  progress: ${state.blackHole.length}/${TOTAL_CARDS}  phase: ${phaseName}`);

  const holeTop = state.blackHole.length > 0 ? state.blackHole[state.blackHole.length - 1] : null;
  const acceptableRanks = getAcceptableRanks(holeTop);
  const acceptsStr = acceptableRanks.length > 0 ? acceptableRanks.map(valueName).join(' / ') : 'none';
  const holeCardStr = holeTop ? formatCard(holeTop) : '[empty]';
  lines.push(`black hole: ${holeCardStr} (accepts: ${acceptsStr})`);
  lines.push('----------');

  const legalFans = getLegalFans(state);

  for (let i = 0; i < state.fans.length; i++) {
    const fan = state.fans[i];
    const prefix = `[${i.toString().padStart(2, ' ')}]`;
    if (fan.length === 0) {
      lines.push(`${prefix} [empty]`);
    } else {
      const isPlayable = state.phase === BlackHolePhase.PLAYING && legalFans.has(i);
      const mark = isPlayable ? '*' : 'x';
      const cardsStr = fan.map(formatCard).join(' ');
      lines.push(`${prefix} ${cardsStr} [${mark}]`);
    }
  }

  lines.push('----------');
  lines.push(`undo: ${state.canUndo ? 'available' : 'none'}`);

  // Hint line: surfaced only when explicitly requested by player
  if (state.messageCode?.endsWith('.noHint')) {
    lines.push('No hint available.');
  } else if (isRequestedHint(state) && state.hint) {
    const recFan = state.hint.fan;
    const recFanCards = state.fans[recFan];
    const recTop = recFanCards && recFanCards.length > 0 ? recFanCards[recFanCards.length - 1] : null;
    const cardStr = recTop ? ` (${formatCard(recTop)})` : '';
    const clearStr = state.hint.canClear
      ? `clear in ${state.hint.continuationMoves} moves`
      : `moves after: ${state.hint.movesAfter}`;
    lines.push(`HINT: Fan ${recFan}${cardStr} [${clearStr}]`);
  }

  if (state.message && state.phase !== BlackHolePhase.GAME_CLEAR && state.phase !== BlackHolePhase.GAME_OVER) {
    lines.push(state.message);
  }

  if (state.phase === BlackHolePhase.GAME_CLEAR) {
    lines.push(`Game Clear! Total moves: ${state.moveCount}`);
  } else if (state.phase === BlackHolePhase.GAME_OVER) {
    lines.push('Game Over!');
  } else if (state.isStalemate && state.phase === BlackHolePhase.PLAYING) {
    lines.push('Stalemate - no legal moves remaining');
  }

  // Available commands guide
  if (state.phase === BlackHolePhase.PLAYING) {
    const available: string[] = [];
    if (legalFans.size > 0) {
      available.push(`p <fan> (${[...legalFans].join(', ')})`);
    }
    if (state.canUndo) {
      available.push('u (undo)');
    }
    available.push('h (hint)');
    available.push('g (giveup)');
    available.push('r (reset)');
    lines.push(`Available commands: ${available.join(', ')}`);
  } else {
    lines.push('Available commands: r (reset)');
  }

  lines.push(formatSeparator());
  return lines.join('\n');
}
