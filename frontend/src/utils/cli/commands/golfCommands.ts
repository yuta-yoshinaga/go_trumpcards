import type { golfApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type GolfArgs = Parameters<typeof golfApi.exec>;

const VALID_COMMANDS = [
  'd',
  'draw',
  'rm',
  'remove',
  'g',
  'giveup',
  'h',
  'hint',
  'log',
  'u',
  'undo',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Golf Solitaire CLI command into API exec arguments. */
export function parseGolfCommand(input: string): CliParseResult<GolfArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'd':
    case 'draw':
      return { args: ['draw'] };
    case 'rm':
    case 'remove': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: rm <col>' };
      return { args: ['remove', parsed.value] };
    }
    case 'g':
    case 'giveup':
      return { args: ['giveup'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    case 'log':
      return { args: ['log'] };
    case 'u':
    case 'undo':
      return { args: ['undo'] };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Golf Solitaire CLI mode. */
export const GOLF_HELP: string[] = [
  'd/draw      - Draw from stock',
  'rm <col>    - Remove bottom card from column',
  'u/undo      - Undo last move',
  'h/hint      - Get a hint',
  'g/giveup    - Give up',
  'log         - Show action log',
  'r/reset     - Reset game',
];
