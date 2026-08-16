import type { oasispokerApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type OasisPokerArgs = Parameters<typeof oasispokerApi.exec>;

const VALID_COMMANDS = [
  'b',
  'bet',
  'e',
  'exchange',
  's',
  'stand',
  'p',
  'play',
  'f',
  'fold',
  'log',
  'r',
  'reset',
  'h',
  'hint',
  'help',
  '?',
];

/** Parse an Oasis Poker CLI command into API exec arguments. */
export function parseOasispokerCommand(input: string): CliParseResult<OasisPokerArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bet': {
      const amount = parseIntArg(args, 0);
      if ('error' in amount) return { error: 'Usage: b <amount> [jackpotBet]' };
      if (args.length >= 2) {
        const jpBet = parseIntArg(args, 1);
        if ('error' in jpBet) return { error: 'Usage: b <amount> [jackpotBet]' };
        return { args: ['bet', amount.value, jpBet.value] };
      }
      return { args: ['bet', amount.value] };
    }
    case 'e':
    case 'exchange': {
      const indices: number[] = [];
      for (let i = 0; i < args.length; i++) {
        const v = parseIntArg(args, i);
        if ('error' in v) return { error: 'Usage: e [idx ...]  (each idx in 0..4)' };
        if (v.value < 0 || v.value > 4) return { error: 'Indices must be 0..4' };
        indices.push(v.value);
      }
      return { args: ['exchange', undefined, undefined, indices] };
    }
    case 's':
    case 'stand':
      return { args: ['stand'] };
    case 'p':
    case 'play':
      return { args: ['play'] };
    case 'f':
    case 'fold':
      return { args: ['fold'] };
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

/** Help text for Oasis Poker CLI mode. */
export const OASISPOKER_HELP: string[] = [
  'b <amt> [jp]  - Ante bet (optional jackpot side bet)',
  'e [idx ...]   - Exchange cards at indices (fee = ante x count)',
  's/stand       - Stand (no exchange)',
  'p/play        - Call (match 2x ante)',
  'f/fold        - Fold hand',
  'log           - Show action log',
  'r/reset       - Reset game',
  'h/hint        - Get a hint',
];
