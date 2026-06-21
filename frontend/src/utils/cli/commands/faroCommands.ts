import type { faroApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by faroApi.exec. */
export type FaroCliArgs = Parameters<typeof faroApi.exec>;

const VALID_COMMANDS = [
  'b',
  'bet',
  'cb',
  'clearbet',
  'ca',
  'clearall',
  'd',
  'deal',
  'call',
  'n',
  'next',
  'l',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/**
 * Parses a single CLI command line for the Faro game into {@link faroApi}.exec
 * arguments.
 *
 * `b <rank> <amount> [c]` places a chip bet on a rank (1=A .. 13=K); the optional
 * trailing `c` makes it a copper (betting the rank to lose). `cb <rank>` clears a
 * single rank's bet, `ca` clears all bets, `d` deals the next turn of two cards,
 * `call <r1> <r2> <r3>` predicts the order of the final three cards, and `n`
 * starts the next deal.
 */
export function parseFaroCommand(input: string): CliParseResult<FaroCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bet': {
      const rank = Number.parseInt(args[0] ?? '', 10);
      const amount = Number.parseInt(args[1] ?? '', 10);
      if (Number.isNaN(rank) || rank < 1 || rank > 13) return { error: 'Usage: b <rank 1-13> <amount> [c]' };
      if (Number.isNaN(amount) || amount <= 0) return { error: 'Usage: b <rank 1-13> <amount> [c]' };
      const copper = (args[2] ?? '').toLowerCase() === 'c';
      return { args: ['bet', { rank, amount, copper }] };
    }
    case 'cb':
    case 'clearbet': {
      const rank = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(rank) || rank < 1 || rank > 13) return { error: 'Usage: cb <rank 1-13>' };
      return { args: ['clearBet', { rank }] };
    }
    case 'ca':
    case 'clearall':
      return { args: ['clearAll'] };
    case 'd':
    case 'deal':
      return { args: ['deal'] };
    case 'call': {
      if (args.length < 3) return { args: ['call', { order: [] }] };
      const order = args.slice(0, 3).map((a) => Number.parseInt(a, 10));
      if (order.some((n) => Number.isNaN(n) || n < 1 || n > 13)) return { error: 'Usage: call <r1> <r2> <r3>' };
      return { args: ['call', { order }] };
    }
    case 'n':
    case 'next':
      return { args: ['next'] };
    case 'l':
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

/** Help text shown in the CLI terminal for Faro. */
export const FARO_HELP: string[] = [
  'b <rank> <amount> [c] - Bet chips on a rank (1=A..13=K); c = copper (bet to lose)',
  'cb <rank>             - Clear the bet on a rank',
  'ca                    - Clear all bets',
  'd / deal              - Deal one turn (two cards)',
  'call <r1> <r2> <r3>   - Predict the order of the last 3 cards (no args to skip)',
  'n / next              - Start the next deal',
  'l / log               - Show action log',
  'r / reset             - Reset game',
];
