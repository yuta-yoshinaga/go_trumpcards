import type { settemezzoApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type SetteEMezzoArgs = Parameters<typeof settemezzoApi.exec>;

const VALID_COMMANDS = [
  'b',
  'bet',
  'deal',
  'h',
  'hit',
  's',
  'stand',
  'matta',
  'bh',
  'bs',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a positive integer amount. */
function parseAmount(args: string[], usage: string): number | { error: string } {
  if (args.length === 0) return { error: usage };
  const n = Number(args[0]);
  if (!Number.isFinite(n) || n <= 0) return { error: usage };
  return n;
}

/**
 * Parse the matta's value.
 *
 * The player types POINTS -- 0.5 or a whole number from 1 to 7 -- and this
 * converts to the halves the API takes. Asking for "6" to mean three points
 * would leak the internal representation into the interface.
 */
function parseMatta(args: string[]): number | { error: string } {
  const usage = 'Usage: matta <0.5 or 1-7>';
  if (args.length === 0) return { error: usage };
  if (args[0] === '0.5') return 1;
  const n = Number(args[0]);
  if (!Number.isInteger(n) || n < 1 || n > 7) return { error: usage };
  return n * 2;
}

/** Parse a Sette e Mezzo CLI command into API exec arguments. */
export function parseSetteEMezzoCommand(input: string): CliParseResult<SetteEMezzoArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bet': {
      const amount = parseAmount(args, 'Usage: b <amount>');
      if (typeof amount !== 'number') return amount;
      return { args: ['bet', amount] };
    }
    case 'deal':
      return { args: ['deal'] };
    case 'h':
    case 'hit':
      return { args: ['hit'] };
    case 's':
    case 'stand':
      return { args: ['stand'] };
    case 'matta': {
      const halves = parseMatta(args);
      if (typeof halves !== 'number') return halves;
      return { args: ['matta', halves] };
    }
    case 'bh':
      return { args: ['bankerhit'] };
    case 'bs':
      return { args: ['bankerstand'] };
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

/** Help text for Sette e Mezzo CLI mode. */
export const SETTEMEZZO_HELP: string[] = [
  'b <amount>      - Bet and deal (the banker uses deal instead)',
  'deal            - Deal the round you are banking',
  'h/hit           - Draw one card',
  's/stand         - Stand',
  'matta <v>       - Set the matta to 0.5 or a whole number from 1 to 7',
  'bh              - Draw a card as the banker',
  'bs              - Stop drawing and settle as the banker',
  'log             - Show action log',
  'r/reset         - Next round',
];
