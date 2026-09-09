import type { quinzeApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type QuinzeArgs = Parameters<typeof quinzeApi.exec>;

const VALID_COMMANDS = ['b', 'bet', 'deal', 'h', 'hit', 's', 'stand', 'bh', 'bs', 'log', 'r', 'reset', 'help', '?'];

/** Parse a positive integer amount. */
function parseAmount(args: string[], usage: string): number | { error: string } {
  if (args.length === 0) return { error: usage };
  const n = Number(args[0]);
  if (!Number.isFinite(n) || n <= 0) return { error: usage };
  return n;
}

/** Parse a Quinze CLI command into API exec arguments. */
export function parseQuinzeCommand(input: string): CliParseResult<QuinzeArgs> {
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

/** Help text for Quinze CLI mode. */
export const QUINZE_HELP: string[] = [
  'b <amount>      - Bet and deal (the banker uses deal instead)',
  'deal            - Deal the round you are banking',
  'h/hit           - Draw one card',
  's/stand         - Stand',
  'bh              - Draw a card as the banker',
  'bs              - Stop drawing and settle as the banker',
  'log             - Show action log',
  'r/reset         - Next round',
];
