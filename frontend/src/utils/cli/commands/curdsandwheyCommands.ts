import type { curdsandwheyApi } from '../../../api/gameApi';
import type { CurdsAndWheyResponse } from '../../../types/card';
import { CurdsAndWheyPhase } from '../../../types/phases';
import { parseIntArg, splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

export type CurdsAndWheyCliArgs = Parameters<typeof curdsandwheyApi.exec>;

const VALID_COMMANDS = ['p', 'play', 'u', 'undo', 'un', 'undo_n', 'g', 'giveup', 'h', 'hint', 'r', 'reset'];

/** Parse a Curds and Whey command into arguments accepted by the page API. */
export function parseCurdsAndWheyCommand(
  input: string,
  state?: CurdsAndWheyResponse | null,
): CliParseResult<CurdsAndWheyCliArgs> {
  const { cmd, args } = splitCommand(input);
  switch (cmd) {
    case 'p':
    case 'play': {
      if (args.length !== 3) return { error: 'Usage: p <fromCol> <cardIndex> <toCol>' };
      const parsed = [0, 1, 2].map((i) => parseIntArg(args, i));
      if (parsed.some((item) => 'error' in item)) return { error: 'Usage: p <fromCol> <cardIndex> <toCol>' };
      const [fromCol, cardIndex, toCol] = parsed.map((item) => ('value' in item ? item.value : -1));
      if (!state || state.phase !== CurdsAndWheyPhase.PLAYING)
        return { error: 'Moves are available only while playing' };
      if (fromCol < 0 || fromCol >= state.columns.length || toCol < 0 || toCol >= state.columns.length) {
        return { error: `Column indices must be between 0 and ${state.columns.length - 1}` };
      }
      if (fromCol === toCol) return { error: 'Choose a different destination column' };
      if (cardIndex < 0 || cardIndex >= (state.columns[fromCol]?.length ?? 0))
        return { error: 'Card index is out of range' };
      if (!state.runStarts[fromCol]?.includes(cardIndex)) return { error: 'Choose the start of a movable run' };
      return { args: ['m', { fromCol, cardIndex, toCol }] };
    }
    case 'u':
    case 'undo':
      return state?.canUndo ? { args: ['u'] } : { error: 'There is no move to undo' };
    case 'un':
    case 'undo_n': {
      const count = parseIntArg(args, 0);
      if ('error' in count || count.value < 1) return { error: 'Usage: undo_n <count>' };
      return { args: ['undo_n', { n: count.value }] };
    }
    case 'g':
    case 'giveup':
      return state?.phase === CurdsAndWheyPhase.PLAYING
        ? { args: ['g'] }
        : { error: 'You can only give up while playing' };
    case 'h':
    case 'hint':
      return state?.phase === CurdsAndWheyPhase.PLAYING
        ? { args: ['hint'] }
        : { error: 'Hints are available only while playing' };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    default:
      return unknownCommand(cmd, VALID_COMMANDS);
  }
}

/** Help text shown in the Curds and Whey CLI terminal. */
export const CURDSANDWHEY_HELP = [
  'p <fromCol> <cardIndex> <toCol> - Move a run to another column (indices start at 0)',
  'u / undo                         - Undo the last move',
  'un <count>                       - Undo multiple moves',
  'g / giveup                       - Give up the game',
  'h / hint                         - Request a hint',
  'r / reset                        - Reset the game',
];
