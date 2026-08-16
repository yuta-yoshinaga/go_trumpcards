import type { caribbeanstudApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type CaribbeanStudArgs = Parameters<typeof caribbeanstudApi.exec>;

const VALID_COMMANDS = ['b', 'bet', 'p', 'play', 'f', 'fold', 'log', 'r', 'reset', 'h', 'hint', 'help', '?'];

/** Parse a Caribbean Stud Poker CLI command into API exec arguments. */
export function parseCaribbeanstudCommand(input: string): CliParseResult<CaribbeanStudArgs> {
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

/** Help text for Caribbean Stud Poker CLI mode. */
export const CARIBBEANSTUD_HELP: string[] = [
  'b <amt> [jp] - Ante bet (optional jackpot side bet)',
  'p/play       - Call (match 2x ante)',
  'f/fold       - Fold hand',
  'log          - Show action log',
  'r/reset      - Reset game',
  'h/hint       - Get a hint',
];
