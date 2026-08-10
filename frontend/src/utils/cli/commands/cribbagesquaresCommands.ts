import type { cribbagesquaresApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type CribbageSquaresArgs = Parameters<typeof cribbagesquaresApi.exec>;

const VALID_COMMANDS = ['p', 'place', 'h', 'hint', 'u', 'undo', 'g', 'giveup', 'log', 'r', 'reset', 'help', '?'];

/** Highest valid row/column index on the 4x4 grid. */
const MAX_INDEX = 3;

/** Parse a Cribbage Squares CLI command into API exec arguments. */
export function parseCribbageSquaresCommand(input: string): CliParseResult<CribbageSquaresArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'place': {
      const usage = `Usage: p <row 0-${MAX_INDEX}> <col 0-${MAX_INDEX}>`;
      const row = parseIntArg(args, 0);
      if ('error' in row) return { error: usage };
      const col = parseIntArg(args, 1);
      if ('error' in col) return { error: usage };
      if (row.value < 0 || row.value > MAX_INDEX || col.value < 0 || col.value > MAX_INDEX) {
        return { error: `Row and col must be between 0 and ${MAX_INDEX}` };
      }
      return { args: ['place', row.value, col.value] };
    }
    case 'h':
    case 'hint':
      return { args: ['hint'] };
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
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Cribbage Squares CLI mode. */
export const CRIBBAGESQUARES_HELP: string[] = [
  'p <r> <c>   - Place current card at row r, col c (0-3)',
  'h/hint      - Suggest a cell for the card in hand',
  'u/undo      - Undo last placement',
  'g/giveup    - Give up current game',
  'log         - Show action log',
  'r/reset     - Reset game',
];
