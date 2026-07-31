import type { sjavsApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type SjavsArgs = Parameters<typeof sjavsApi.exec>;

const VALID_COMMANDS = ['b', 'bid', 'p', 'play', 'n', 'next', 'log', 'r', 'reset', 'help', '?'];

/** Parse a Sjavs CLI command into API exec arguments. */
export function parseSjavsCommand(input: string): CliParseResult<SjavsArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bid': {
      if (args.length === 0) return { error: `Usage: ${cmd} <length>` };
      const n = Number(args[0]);
      // 0 is a pass, so it must be accepted rather than treated as missing.
      if (!Number.isInteger(n) || n < 0) return { error: `Invalid bid length: ${args[0]}` };
      return { args: ['bid', n] };
    }
    case 'p':
    case 'play': {
      if (args.length === 0) return { error: `Usage: ${cmd} <index>` };
      const n = Number(args[0]);
      if (!Number.isInteger(n) || n < 0) return { error: `Invalid card index: ${args[0]}` };
      // The hand index goes in the THIRD position so it can never be read as a
      // bid length.
      return { args: ['play', undefined, n] };
    }
    case 'n':
    case 'next':
      return { args: ['next'] };
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

/** Help text for Sjavs CLI mode. */
export const SJAVS_HELP: string[] = [
  'b <n>           - Bid a trump-suit length of n (0 passes)',
  'p <i>           - Play the hand card at index i',
  'n/next          - Deal the next hand',
  'log             - Show action log',
  'r/reset         - New rubber',
];
