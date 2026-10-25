import type { fourteenoutApi } from '../../../api/games/fourteenout';
import type { FourteenOutResponse } from '../../../types/card';
import { FourteenOutPhase } from '../../../types/phases';
import { fourteenOutCanUndo, fourteenOutPartners, fourteenOutTail } from '../../fourteenOutRules';
import { parseIntArg, splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by fourteenoutApi.exec. */
export type FourteenOutCliArgs = Parameters<typeof fourteenoutApi.exec>;

const VALID_COMMANDS = ['p', 'play', 'u', 'undo', 'g', 'giveup', 'h', 'hint', 'r', 'reset'];

/** Parse a Fourteen Out command using the same pair legality as the page. */
export function parseFourteenOutCommand(
  input: string,
  state?: FourteenOutResponse | null,
): CliParseResult<FourteenOutCliArgs> {
  const { cmd, args } = splitCommand(input);
  switch (cmd) {
    case 'p':
    case 'play': {
      if (args.length !== 2) return { error: 'Usage: p <from> <to>' };
      const from = parseIntArg(args, 0);
      const to = parseIntArg(args, 1);
      if ('error' in from || 'error' in to) return { error: 'Usage: p <from> <to>' };
      if (!state || state.phase !== FourteenOutPhase.PLAYING)
        return { error: 'You can only remove cards while playing' };
      if (from.value < 0 || from.value >= state.columns.length || to.value < 0 || to.value >= state.columns.length)
        return { error: `Column indices must be between 0 and ${state.columns.length - 1}` };
      if (from.value === to.value) return { error: 'Choose two different columns' };
      if (!fourteenOutTail(state.columns[from.value]) || !fourteenOutTail(state.columns[to.value]))
        return { error: 'Both columns must contain a card' };
      if (!fourteenOutPartners(state, from.value).includes(to.value))
        return { error: 'The selected cards do not make 14' };
      return { args: ['remove', from.value, to.value] };
    }
    case 'u':
    case 'undo':
      if (state?.phase !== FourteenOutPhase.PLAYING) return { error: 'You can only undo while playing' };
      if (!fourteenOutCanUndo(state)) return { error: 'There is no action to undo' };
      return { args: ['undo'] };
    case 'g':
    case 'giveup':
      if (state?.phase !== FourteenOutPhase.PLAYING) return { error: 'You can only give up while playing' };
      return { args: ['giveup'] };
    case 'h':
    case 'hint':
      if (state?.phase !== FourteenOutPhase.PLAYING) return { error: 'You can only request a hint while playing' };
      return { args: ['hint'] };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    default:
      return unknownCommand(cmd, VALID_COMMANDS);
  }
}

/** Help text shown in the Fourteen Out CLI terminal. */
export const FOURTEENOUT_HELP = [
  'p <from> <to> - Remove two exposed cards that make 14 (column indices start at 0)',
  'u / undo - Undo the last move',
  'g / giveup - Give up the game',
  'h / hint - Show a hint',
  'r / reset - Reset the game',
];
