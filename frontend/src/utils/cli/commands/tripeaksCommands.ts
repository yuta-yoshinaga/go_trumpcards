import type { tripeaksApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type TriPeaksArgs = Parameters<typeof tripeaksApi.exec>;

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

/** Parse a TriPeaks CLI command into API exec arguments. */
export function parseTripeaksCommand(input: string): CliParseResult<TriPeaksArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'd':
    case 'draw':
      return { args: ['draw'] };
    case 'rm':
    case 'remove': {
      if (args.length < 2) return { error: 'Usage: rm <row> <col>' };
      const row = parseIntArg(args, 0);
      if ('error' in row) return { error: 'Usage: rm <row> <col>' };
      const col = parseIntArg(args, 1);
      if ('error' in col) return { error: 'Usage: rm <row> <col>' };
      return { args: ['remove', row.value, col.value] };
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

/** Help text for TriPeaks CLI mode. */
export const TRIPEAKS_HELP: string[] = [
  'd/draw      - Draw from stock',
  'rm <row> <col> - Remove a card',
  'u/undo      - Undo last move',
  'h/hint      - Get a hint',
  'g/giveup    - Give up',
  'log         - Show action log',
  'r/reset     - Reset game',
];
