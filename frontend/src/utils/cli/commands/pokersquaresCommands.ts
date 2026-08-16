import type { pokersquaresApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type PokerSquaresArgs = Parameters<typeof pokersquaresApi.exec>;

const VALID_COMMANDS = ['p', 'place', 'u', 'undo', 'g', 'giveup', 'log', 'r', 'reset', 'help', '?'];

/** Parse a Poker Squares CLI command into API exec arguments. */
export function parsePokerSquaresCommand(input: string): CliParseResult<PokerSquaresArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'place': {
      const row = parseIntArg(args, 0);
      if ('error' in row) return { error: 'Usage: p <row 0-4> <col 0-4>' };
      const col = parseIntArg(args, 1);
      if ('error' in col) return { error: 'Usage: p <row 0-4> <col 0-4>' };
      if (row.value < 0 || row.value > 4 || col.value < 0 || col.value > 4) {
        return { error: 'Row and col must be between 0 and 4' };
      }
      return { args: ['place', row.value, col.value] };
    }
    case 'u':
    case 'undo':
      return { args: ['undo'] };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] };
    case 'log':
      return { args: ['log'] };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Poker Squares CLI mode. */
export const POKERSQUARES_HELP: string[] = [
  'p <r> <c>   - Place current card at row r, col c (0-4)',
  'u/undo      - Undo last placement',
  'g/giveup    - Give up current game',
  'log         - Show action log',
  'r/reset     - Reset game',
  'h/hint      - Get a hint',
];
