import type { labellelucieApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type LaBelleLucieArgs = Parameters<typeof labellelucieApi.exec>;
const VALID_COMMANDS = [
  'p',
  'play',
  'f',
  'foundation',
  'rd',
  'redeal',
  'ac',
  'autocomplete',
  'u',
  'undo',
  'undo_n',
  'g',
  'giveup',
  'h',
  'hint',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a La Belle Lucie CLI command into API exec arguments. */
export function parseLaBelleLucieCommand(input: string): CliParseResult<LaBelleLucieArgs> {
  const { cmd, args } = splitCommand(input);
  switch (cmd) {
    case 'p':
    case 'play': {
      if (args.length < 2) return { error: 'Usage: p <from> <to>' };
      const from = parseIntArg(args, 0);
      const to = parseIntArg(args, 1);
      if ('error' in from || 'error' in to || from.value < 0 || from.value > 8 || to.value < 0 || to.value > 8) {
        return { error: 'Usage: p <from> <to> (fan indices 0-8)' };
      }
      return { args: ['mf', from.value, to.value] };
    }
    case 'f':
    case 'foundation': {
      const from = parseIntArg(args, 0);
      if ('error' in from || from.value < 0 || from.value > 8) return { error: 'Usage: f <idx> (fan index 0-8)' };
      return { args: ['ff', from.value] };
    }
    case 'rd':
    case 'redeal':
      return { args: ['rd'] };
    case 'ac':
    case 'autocomplete':
      return { args: ['ac'] };
    case 'u':
    case 'undo':
      return { args: ['u'] };
    case 'undo_n': {
      const n = parseIntArg(args, 0);
      if ('error' in n || n.value < 1) return { error: 'Usage: undo_n <count>' };
      return { args: ['undo_n', undefined, undefined, n.value] };
    }
    case 'g':
    case 'giveup':
      return { args: ['giveup'] };
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

/** Help text shown in La Belle Lucie's CLI terminal. */
export const LABELLELUCIE_HELP = [
  'p <from> <to> - Move a card between fan indices (0-8)',
  'f <idx>       - Move a fan top to foundation (0-8)',
  'rd / redeal   - Gather and redeal cards',
  'ac / autocomplete - Move all available cards to foundation',
  'u / undo      - Undo the last move',
  'undo_n <count> - Undo multiple moves',
  'g / giveup    - Give up',
  'h / hint      - Get a hint',
  'r / reset     - Reset game',
];
