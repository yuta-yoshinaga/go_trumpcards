import type { letitrideApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type LetItRideArgs = Parameters<typeof letitrideApi.exec>;

const VALID_COMMANDS = ['b', 'bet', 'p', 'pull', 'l', 'letitride', 'log', 'r', 'reset', 'help', '?'];

/** Parse a Let It Ride CLI command into API exec arguments. */
export function parseLetitrideCommand(input: string): CliParseResult<LetItRideArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bet': {
      const amount = parseIntArg(args, 0);
      if ('error' in amount) return { error: 'Usage: b <amount>' };
      return { args: ['bet', amount.value] };
    }
    case 'p':
    case 'pull':
      return { args: ['pull'] };
    case 'l':
    case 'letitride':
      return { args: ['letitride'] };
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

/** Help text for Let It Ride CLI mode. */
export const LETITRIDE_HELP: string[] = [
  'b <amt>     - Place bet (3 equal bets)',
  'p/pull      - Pull (withdraw) a bet',
  'l/letitride - Let It Ride (keep bet)',
  'log         - Show action log',
  'r/reset     - Reset game',
];
