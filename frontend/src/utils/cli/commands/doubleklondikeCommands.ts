import type { doubleklondikeApi } from '../../../api/gameApi';
import type { DoubleKlondikeResponse } from '../../../types/card';
import { DoubleKlondikePhase } from '../../../types/phases';
import { splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

export type DoubleKlondikeCliArgs = Parameters<typeof doubleklondikeApi.exec>;
const VALID_COMMANDS = ['p', 'play', 'd', 'draw', 'ac', 'auto', 'u', 'undo', 'h', 'hint', 'g', 'giveup', 'r', 'reset'];

/** Parse Double Klondike commands into the exact arguments used by page actions. */
export function parseDoubleKlondikeCommand(
  input: string,
  state?: DoubleKlondikeResponse | null,
): CliParseResult<DoubleKlondikeCliArgs> {
  const { cmd, args } = splitCommand(input);
  switch (cmd) {
    case 'p':
    case 'play': {
      if (state && state.phase !== DoubleKlondikePhase.PLAYING)
        return { error: 'Moves are available only while playing' };
      const [from, to] = args;
      if ((from === 'waste' && args.length !== 2) || (from !== 'waste' && args.length !== 3))
        return { error: 'Usage: p waste <col|f> | p <col> <idx> <to>' };
      if (from === 'waste' && to === 'f') {
        if (state && (state.phase !== DoubleKlondikePhase.PLAYING || !state.legalTargets.waste?.foundation.length))
          return { error: 'Waste cannot move to a foundation' };
        return { args: ['mwf'] };
      }
      if (from === 'waste' && /^\d+$/.test(to ?? '')) {
        const col = Number(to);
        if (!state || col < 0 || col >= state.tableau.length) return { error: 'Tableau column is out of range' };
        if (!state.legalTargets.waste?.tableau.includes(col))
          return { error: 'Waste cannot move to that tableau column' };
        return { args: ['mwt', { col }] };
      }
      if (/^\d+$/.test(from ?? '') && (args[2] === 'f' || /^\d+$/.test(args[2] ?? ''))) {
        const col = Number(from);
        const cardIndex = Number(to);
        if (
          !state ||
          col < 0 ||
          col >= state.tableau.length ||
          cardIndex < 0 ||
          cardIndex >= (state.tableau[col]?.length ?? 0)
        )
          return { error: 'Tableau card index is out of range' };
        const key = `tableau-${col}-${cardIndex}`;
        if (args[2] === 'f') {
          if (cardIndex !== (state.tableau[col]?.length ?? 0) - 1 || !state.legalTargets[key]?.foundation.length)
            return { error: 'That card cannot move to a foundation' };
          return { args: ['mtf', { col }] };
        }
        const toCol = Number(args[2]);
        if (toCol < 0 || toCol >= state.tableau.length) return { error: 'Tableau column is out of range' };
        if (!state.legalTargets[key]?.tableau.includes(toCol))
          return { error: 'That card cannot move to that tableau column' };
        return { args: ['mtt', { fromCol: col, cardIndex, toCol }] };
      }
      return { error: 'Usage: p <from> <to> (waste or tableau index)' };
    }
    case 'd':
    case 'draw':
      if (
        state &&
        (state.phase !== DoubleKlondikePhase.PLAYING || (state.stockCount === 0 && state.waste.length === 0))
      )
        return { error: 'Cannot draw in the current state' };
      return { args: ['d'] };
    case 'ac':
    case 'auto':
      if (state?.phase !== undefined && state.phase !== DoubleKlondikePhase.PLAYING)
        return { error: 'Auto move is only available while playing' };
      return { args: ['ac'] };
    case 'u':
    case 'undo':
      if (state && (state.phase !== DoubleKlondikePhase.PLAYING || !state.canUndo))
        return { error: 'There is no move to undo' };
      return { args: ['u'] };
    case 'h':
    case 'hint':
      if (state?.phase !== undefined && state.phase !== DoubleKlondikePhase.PLAYING)
        return { error: 'Hints are available only while playing' };
      return { args: ['hint'] };
    case 'g':
    case 'giveup':
      if (state?.phase !== undefined && state.phase !== DoubleKlondikePhase.PLAYING)
        return { error: 'Give up is only available while playing' };
      return { args: ['g'] };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    default:
      return unknownCommand(cmd, VALID_COMMANDS);
  }
}

export const DOUBLEKLONDIKE_HELP = [
  'p waste <col|f> - Move the waste card to a tableau column or foundation',
  'p <col> <idx> <to> - Move tableau card [idx] in column <col> to <to> column or f',
  'd / draw - Draw from stock or recycle waste',
  'ac / auto - Move available cards to foundations',
  'u / undo - Undo the last move',
  'h / hint - Request a hint',
  'g / giveup - Give up the game',
  'r / reset - Reset the game',
];
