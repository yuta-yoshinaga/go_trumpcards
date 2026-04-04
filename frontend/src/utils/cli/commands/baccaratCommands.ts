import type { baccaratApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type BaccaratArgs = Parameters<typeof baccaratApi.exec>;

const VALID_COMMANDS = [
  'b',
  'bet',
  'bp',
  'bankerpair',
  'pp',
  'playerpair',
  'log',
  'cl',
  'clearhistory',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Baccarat CLI command into API exec arguments. */
export function parseBaccaratCommand(input: string): CliParseResult<BaccaratArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bet': {
      if (args.length < 2) return { error: 'Usage: b <amount> <type> (type: 0=player, 1=banker, 2=tie)' };
      const amount = parseIntArg(args, 0);
      if ('error' in amount) return { error: 'Usage: b <amount> <type>' };
      const betType = parseIntArg(args, 1);
      if ('error' in betType) return { error: 'Usage: b <amount> <type> (type: 0=player, 1=banker, 2=tie)' };
      return { args: ['bet', amount.value, betType.value] };
    }
    case 'pp':
    case 'playerpair': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: pp <amount>' };
      return { args: ['bet', undefined, undefined, parsed.value] };
    }
    case 'bp':
    case 'bankerpair': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: bp <amount>' };
      return { args: ['bet', undefined, undefined, undefined, parsed.value] };
    }
    case 'log':
      return { args: ['log'] };
    case 'cl':
    case 'clearhistory':
      return { args: ['clearhistory'] };
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

/** Help text for Baccarat CLI mode. */
export const BACCARAT_HELP: string[] = [
  'b <amt> <type> - Bet (0=player, 1=banker, 2=tie)',
  'pp <amount>    - Player pair side bet',
  'bp <amount>    - Banker pair side bet',
  'log            - Show action log',
  'cl/clearhistory- Clear history',
  'r/reset        - Reset game',
];
