import type { simplesimonApi } from '../../../api/gameApi';
import type { SimpleSimonResponse } from '../../../types/card';
import { SimpleSimonPhase } from '../../../types/phases';
import { splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by simplesimonApi.exec. */
export type SimpleSimonCliArgs = Parameters<typeof simplesimonApi.exec>;
const VALID_COMMANDS = ['m', 'g', 'u', 'undo_n', 'h', 'hint', 'r', 'reset'];

/** Parses Simple Simon commands; state enables the same legal move checks as the page. */
export function parseSimpleSimonCommand(
  input: string,
  state?: SimpleSimonResponse,
): CliParseResult<SimpleSimonCliArgs> {
  const { cmd, args } = splitCommand(input);
  switch (cmd) {
    case 'm': {
      const fromCol = Number(args[0]);
      const cardIndex = Number(args[1]);
      const toCol = Number(args[2]);
      if (args.length < 3 || ![fromCol, cardIndex, toCol].every(Number.isInteger))
        return { error: 'Usage: m <fromCol> <cardIndex> <toCol>' };
      if (
        state &&
        (fromCol < 0 ||
          fromCol >= state.columns.length ||
          cardIndex < 0 ||
          cardIndex >= state.columns[fromCol].length ||
          toCol < 0 ||
          toCol >= state.columns.length)
      )
        return { error: 'Move index out of range' };
      if (
        state &&
        (state.phase !== SimpleSimonPhase.PLAYING ||
          !state.runStarts[fromCol]?.includes(cardIndex) ||
          !(state.legalTargets[`tableau-${fromCol}-${cardIndex}`]?.tableau ?? []).includes(toCol))
      )
        return { error: 'Illegal move' };
      return { args: ['m', { fromCol, cardIndex, toCol }] };
    }
    case 'g':
      return { args: ['g'] };
    case 'u':
      return { args: ['u'] };
    case 'undo_n': {
      const n = Number(args[0]);
      if (!args.length || !Number.isInteger(n) || n < 1 || (state && n > state.moveCount))
        return { error: 'Usage: undo_n <n>' };
      return { args: ['undo_n', { n }] };
    }
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    default:
      return unknownCommand(cmd, VALID_COMMANDS);
  }
}

/** Help text shown in the Simple Simon terminal. */
export const SIMPLESIMON_HELP = [
  'm <fromCol> <cardIndex> <toCol> - Move a run',
  'g - Give up',
  'u - Undo',
  'undo_n <n> - Undo n moves',
  'h / hint - Show hint',
  'r / reset - Reset game',
];
