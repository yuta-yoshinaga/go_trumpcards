import type { montecarloApi } from '../../../api/games/montecarlo';
import { parseIntArg, splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type MonteCarloArgs = Parameters<typeof montecarloApi.exec>;

const VALID_COMMANDS = [
  'p',
  'play',
  'remove',
  'd',
  'deal',
  'u',
  'undo',
  'g',
  'giveup',
  'h',
  'hint',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Monte Carlo Solitaire command into API exec arguments. */
export function parseMonteCarloCommand(input: string): CliParseResult<MonteCarloArgs> {
  const { cmd, args } = splitCommand(input);
  switch (cmd) {
    case 'p':
    case 'play':
    case 'remove': {
      if (args.length < 2) return { error: 'Usage: p <idx1> <idx2> (0-24)' };
      const first = parseIntArg(args, 0);
      const second = parseIntArg(args, 1);
      if ('error' in first || 'error' in second) return { error: 'Usage: p <idx1> <idx2> (0-24)' };
      if ([first.value, second.value].some((value) => value < 0 || value > 24)) {
        return { error: 'Cell indices must be between 0 and 24' };
      }
      if (first.value === second.value) return { error: 'Choose two different cells' };
      return {
        args: ['remove', Math.floor(first.value / 5), first.value % 5, Math.floor(second.value / 5), second.value % 5],
      };
    }
    case 'd':
    case 'deal':
      return { args: ['deal'] };
    case 'u':
    case 'undo':
      return { args: ['undo'] };
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

/** Help text shown in Monte Carlo CLI mode. */
export const MONTECARLO_HELP = [
  'p <idx1> <idx2> - Remove an adjacent same-rank pair (board indices 0-24)',
  'd / deal - Deal five cards from the stock',
  'u / undo - Undo the last action',
  'g / giveup - Give up the game',
  'h / hint - Show a hint',
  'r / reset - Reset the game',
];
