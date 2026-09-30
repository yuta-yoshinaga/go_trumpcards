import type { goFishApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type GoFishArgs = Parameters<typeof goFishApi.exec>;

const VALID_COMMANDS = ['ask', 'log', 'r', 'reset', 'help', '?'];

/** Parse a Go Fish CLI command into API exec arguments. */
export function parseGofishCommand(input: string): CliParseResult<GoFishArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'ask': {
      if (args.length < 2) return { error: 'Usage: ask <targetIdx> <rank> (rank: 1-13)' };
      const target = parseIntArg(args, 0);
      if ('error' in target) return { error: 'Usage: ask <targetIdx> <rank>' };
      const rank = parseIntArg(args, 1);
      if ('error' in rank) return { error: 'Usage: ask <targetIdx> <rank>' };
      return { args: ['ask', target.value, rank.value] };
    }
    case 'log':
      return { args: ['log'] };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    default:
      return unknownCommand(cmd, VALID_COMMANDS);
  }
}

/** Help text for Go Fish CLI mode. */
export const GOFISH_HELP: string[] = [
  'ask <target> <rank> - Ask player for rank (1=A, 11=J, 12=Q, 13=K)',
  'log         - Show action log',
  'r/reset     - Reset game',
];
