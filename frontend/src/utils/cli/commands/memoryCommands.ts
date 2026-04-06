import type { memoryApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type MemoryArgs = Parameters<typeof memoryApi.exec>;

const VALID_COMMANDS = ['fl', 'flip', 'n', 'next', 'log', 'r', 'reset', 'help', '?'];

/** Parse a Memory CLI command into API exec arguments. */
export function parseMemoryCommand(input: string): CliParseResult<MemoryArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'fl':
    case 'flip': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: fl <position>' };
      return { args: ['flip', parsed.value] };
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

/** Help text for Memory CLI mode. */
export const MEMORY_HELP: string[] = [
  'fl <pos>    - Flip a card at position',
  'n/next      - Next turn (after match check)',
  'log         - Show action log',
  'r/reset     - Reset game',
];
