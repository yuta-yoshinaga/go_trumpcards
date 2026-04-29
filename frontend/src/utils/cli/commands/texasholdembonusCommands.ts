import type { texasholdembonusApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type TexasHoldemBonusArgs = Parameters<typeof texasholdembonusApi.exec>;

const VALID_COMMANDS = [
  'b',
  'bet',
  'p',
  'play',
  'f',
  'fold',
  'c',
  'check',
  'ra',
  'raise',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Texas Hold'em Bonus Poker CLI command into API exec arguments. */
export function parseTexasholdembonusCommand(input: string): CliParseResult<TexasHoldemBonusArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bet': {
      const amount = parseIntArg(args, 0);
      if ('error' in amount) return { error: 'Usage: b <amount> [bonusBet]' };
      if (args.length >= 2) {
        const bonus = parseIntArg(args, 1);
        if ('error' in bonus) return { error: 'Usage: b <amount> [bonusBet]' };
        return { args: ['bet', amount.value, bonus.value] };
      }
      return { args: ['bet', amount.value] };
    }
    case 'p':
    case 'play':
      return { args: ['play'] };
    case 'f':
    case 'fold':
      return { args: ['fold'] };
    case 'c':
    case 'check':
      return { args: ['check'] };
    case 'ra':
    case 'raise':
      return { args: ['raise'] };
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

/** Help text for Texas Hold'em Bonus Poker CLI mode. */
export const TEXASHOLDEMBONUS_HELP: string[] = [
  'b <amt> [bn] - Ante bet (optional bonus side bet)',
  'p/play       - Pre-flop play (2x ante)',
  'f/fold       - Pre-flop fold',
  'c/check      - Check (flop / turn)',
  'ra/raise     - Raise 1x ante (flop / turn)',
  'log          - Show action log',
  'r/reset      - Reset game',
];
