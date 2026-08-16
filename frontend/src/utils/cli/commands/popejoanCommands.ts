import type { popejoanApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type PopeJoanArgs = Parameters<typeof popejoanApi.exec>;

const VALID_COMMANDS = ['p', 'play', 'n', 'next', 'log', 'r', 'reset', 'help', '?'];

/** Parse a Pope Joan CLI command into API exec arguments. */
export function parsePopeJoanCommand(input: string): CliParseResult<PopeJoanArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      if (args.length === 0) return { error: `Usage: ${cmd} <index>` };
      const n = Number(args[0]);
      if (!Number.isInteger(n) || n < 0) return { error: `Invalid card index: ${args[0]}` };
      return { args: ['play', n] };
    }
    case 'n':
    case 'next':
      return { args: ['next'] };
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

/** Help text for Pope Joan CLI mode. */
export const POPEJOAN_HELP: string[] = [
  'p <i>           - Play hand card i',
  'n/next          - Deal the next hand',
  'log             - Show action log',
  'r/reset         - New game',
  'h/hint      - Get a hint',
];
