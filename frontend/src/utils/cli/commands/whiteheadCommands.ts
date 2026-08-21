import type { whiteheadApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type WhiteheadArgs = Parameters<typeof whiteheadApi.exec>;

const VALID_COMMANDS = [
  'd',
  'draw',
  'm',
  'move',
  'g',
  'giveup',
  'ac',
  'autocomplete',
  'u',
  'undo',
  'h',
  'hint',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Whitehead CLI command into API exec arguments. */
export function parseWhiteheadCommand(input: string): CliParseResult<WhiteheadArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'd':
    case 'draw':
      return { args: ['draw'] };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] };
    case 'ac':
    case 'autocomplete':
      return { args: ['autocomplete'] };
    case 'u':
    case 'undo':
      return { args: ['undo'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    case 'r':
    case 'reset': {
      if (args.length > 0) {
        const parsed = parseIntArg(args, 0);
        if ('error' in parsed) return { error: 'Usage: r [1|3]' };
        return { args: ['reset', undefined, undefined, { drawCount: parsed.value }] };
      }
      return { args: ['reset'] };
    }
    case 'm':
    case 'move':
      return parseMoveCommand(args);
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

function parseMoveCommand(args: string[]): CliParseResult<WhiteheadArgs> {
  if (args.length < 2) return { error: 'Usage: m <from> <to> (e.g., m w t 3, m t 0 f, m t 0 2 t 3)' };
  const fromZone = args[0].toLowerCase();

  if (fromZone === 'w') {
    const toZone = args[1].toLowerCase();
    if (toZone === 'f') return { args: ['move', { zone: 'waste' }, { zone: 'foundation' }] };
    if (toZone === 't') {
      const parsed = parseIntArg(args, 2);
      if ('error' in parsed) return { error: 'Usage: m w t <col>' };
      return { args: ['move', { zone: 'waste' }, { zone: 'tableau', col: parsed.value }] };
    }
    return { error: 'Invalid target: use f (foundation) or t <col> (tableau)' };
  }

  if (fromZone === 't') {
    const fromCol = parseIntArg(args, 1);
    if ('error' in fromCol) return { error: 'Usage: m t <col> f | m t <col> <cardIdx> t <toCol>' };
    if (args.length >= 3 && args[2].toLowerCase() === 'f') {
      return { args: ['move', { zone: 'tableau', col: fromCol.value }, { zone: 'foundation' }] };
    }
    if (args.length >= 5 && args[3].toLowerCase() === 't') {
      const cardIdx = parseIntArg(args, 2);
      if ('error' in cardIdx) return { error: 'Usage: m t <col> <cardIdx> t <toCol>' };
      const toCol = parseIntArg(args, 4);
      if ('error' in toCol) return { error: 'Usage: m t <col> <cardIdx> t <toCol>' };
      return {
        args: [
          'move',
          { zone: 'tableau', col: fromCol.value, cardIndex: cardIdx.value },
          { zone: 'tableau', col: toCol.value },
        ],
      };
    }
    return { error: 'Usage: m t <col> f | m t <col> <cardIdx> t <toCol>' };
  }

  return { error: 'Invalid source: use w (waste) or t <col> (tableau)' };
}

/** Help text for Whitehead CLI mode. */
export const WHITEHEAD_HELP: string[] = [
  'd/draw      - Draw from stock',
  'm w t <col> - Move waste to tableau column',
  'm w f       - Move waste to foundation',
  'm t <col> f - Move tableau top to foundation',
  'm t <c> <i> t <c> - Move tableau card to column',
  'ac          - Auto-complete',
  'u/undo      - Undo last move',
  'h/hint      - Get a hint',
  'g/giveup    - Give up',
  'r [1|3]     - Reset (draw count)',
];
