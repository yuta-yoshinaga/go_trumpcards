import type { pyramidApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type PyramidArgs = Parameters<typeof pyramidApi.exec>;

const VALID_COMMANDS = [
  'd',
  'draw',
  'rm',
  'remove',
  'k',
  'king',
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

/** Parse a Pyramid Solitaire CLI command into API exec arguments. */
export function parsePyramidCommand(input: string): CliParseResult<PyramidArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'd':
    case 'draw':
      return { args: ['draw'] };
    case 'rm':
    case 'remove': {
      // remove <r1> <c1> <r2> <c2> — pair two pyramid cards
      // remove <r1> <c1> w — pair pyramid card with waste
      if (args.length < 2) return { error: 'Usage: rm <row> <col> <row2> <col2> | rm <row> <col> w' };
      const r1 = parseIntArg(args, 0);
      if ('error' in r1) return { error: 'Usage: rm <row> <col> ...' };
      const c1 = parseIntArg(args, 1);
      if ('error' in c1) return { error: 'Usage: rm <row> <col> ...' };
      if (args.length >= 3 && args[2].toLowerCase() === 'w') {
        return { args: ['remove', { zone: 'pyramid', row: r1.value, col: c1.value }, { zone: 'waste' }] };
      }
      if (args.length >= 4) {
        const r2 = parseIntArg(args, 2);
        if ('error' in r2) return { error: 'Usage: rm <r1> <c1> <r2> <c2>' };
        const c2 = parseIntArg(args, 3);
        if ('error' in c2) return { error: 'Usage: rm <r1> <c1> <r2> <c2>' };
        return {
          args: [
            'remove',
            { zone: 'pyramid', row: r1.value, col: c1.value },
            { zone: 'pyramid', row: r2.value, col: c2.value },
          ],
        };
      }
      return { error: 'Usage: rm <r1> <c1> <r2> <c2> | rm <r1> <c1> w' };
    }
    case 'k':
    case 'king': {
      // Remove a lone king
      if (args.length < 2) return { error: 'Usage: k <row> <col> | k w' };
      if (args[0].toLowerCase() === 'w') {
        return { args: ['remove', { zone: 'waste' }] };
      }
      const r = parseIntArg(args, 0);
      if ('error' in r) return { error: 'Usage: k <row> <col>' };
      const c = parseIntArg(args, 1);
      if ('error' in c) return { error: 'Usage: k <row> <col>' };
      return { args: ['remove', { zone: 'pyramid', row: r.value, col: c.value }] };
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

/** Help text for Pyramid Solitaire CLI mode. */
export const PYRAMID_HELP: string[] = [
  'd/draw      - Draw from stock',
  'rm <r> <c> <r> <c> - Remove pair (pyramid)',
  'rm <r> <c> w      - Remove pair (pyramid+waste)',
  'k <r> <c>  - Remove king',
  'k w         - Remove king from waste',
  'u/undo      - Undo last move',
  'h/hint      - Get a hint',
  'g/giveup    - Give up',
  'log         - Show action log',
  'r/reset     - Reset game',
];
