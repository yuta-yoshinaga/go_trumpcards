import type { prsiApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type PrsiArgs = Parameters<typeof prsiApi.exec>;

const VALID_COMMANDS = ['p', 'play', 'd', 'draw', 'r', 'reset', 'l', 'log', 'help', '?'];

/** Parse a Prší CLI command into API exec arguments. */
export function parsePrsiCommand(input: string): CliParseResult<PrsiArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: p <idx>' };
      return { args: ['play', parsed.value] };
    }
    case 'd':
    case 'draw':
      return { args: ['draw'] };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    case 'l':
    case 'log':
      return { args: ['log'] };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Prší CLI mode. */
export const PRSI_HELP: string[] = [
  'p <idx>     - Play a card',
  'd/draw      - Draw from pile',
  'r/reset     - Reset game',
  'l/log       - Show action log',
];
