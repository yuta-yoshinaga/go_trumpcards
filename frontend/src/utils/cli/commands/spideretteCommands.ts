import type { spideretteApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type SpideretteArgs = Parameters<typeof spideretteApi.exec>;

const VALID_COMMANDS = [
  'd',
  'deal',
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
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Spiderette CLI command into API exec arguments. */
export function parseSpideretteCommand(input: string): CliParseResult<SpideretteArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'd':
    case 'deal':
      return { args: ['deal'] };
    case 'm':
    case 'move': {
      if (args.length < 2) return { error: 'Usage: m <fromCol> <toCol> or m <fromCol> <cardIdx> <toCol>' };
      const fromCol = parseIntArg(args, 0);
      if ('error' in fromCol) return { error: 'Usage: m <fromCol> <toCol>' };
      if (args.length >= 3) {
        const cardIdx = parseIntArg(args, 1);
        if ('error' in cardIdx) return { error: 'Usage: m <fromCol> <cardIdx> <toCol>' };
        const toCol = parseIntArg(args, 2);
        if ('error' in toCol) return { error: 'Usage: m <fromCol> <cardIdx> <toCol>' };
        return {
          args: [
            'move',
            { zone: 'tableau', col: fromCol.value, cardIndex: cardIdx.value },
            { zone: 'tableau', col: toCol.value },
          ],
        };
      }
      const toCol = parseIntArg(args, 1);
      if ('error' in toCol) return { error: 'Usage: m <fromCol> <toCol>' };
      return {
        args: ['move', { zone: 'tableau', col: fromCol.value }, { zone: 'tableau', col: toCol.value }],
      };
    }
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
    case 'log':
      return { args: ['log'] };
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

/** Help text for Spiderette CLI mode. */
export const SPIDERETTE_HELP: string[] = [
  'd/deal      - Deal new row',
  'm <c1> <c2> - Move top card from col to col',
  'm <c1> <i> <c2> - Move card at index to col',
  'ac          - Auto-complete',
  'u/undo      - Undo last move',
  'h/hint      - Get a hint',
  'g/giveup    - Give up',
  'log         - Show action log',
  'r/reset     - Reset the game',
];
