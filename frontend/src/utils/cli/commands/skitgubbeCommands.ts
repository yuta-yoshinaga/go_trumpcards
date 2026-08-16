import type { skitgubbeApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type SkitgubbeArgs = Parameters<typeof skitgubbeApi.exec>;

const VALID_COMMANDS = ['p', 'play', 'u', 'pickup', 'log', 'r', 'reset', 'help', '?'];

/** Parse a Skitgubbe CLI command into API exec arguments. */
export function parseSkitgubbeCommand(input: string): CliParseResult<SkitgubbeArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      if (args.length === 0) return { error: `Usage: ${cmd} <index>` };
      const n = Number(args[0]);
      if (!Number.isInteger(n) || n < 0) return { error: `Invalid card index: ${args[0]}` };
      return { args: ['play', n] };
    }
    case 'u':
    case 'pickup':
      // No index: the server refuses the pick-up whenever anything still beats
      // the pile, so there is nothing to choose.
      return { args: ['pickup'] };
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

/** Help text for Skitgubbe CLI mode. */
export const SKITGUBBE_HELP: string[] = [
  'p <i>           - Play the hand card at index i',
  'u/pickup        - Take the pile (only when nothing beats it)',
  'log             - Show action log',
  'r/reset         - New game',
  'h/hint      - Get a hint',
];
