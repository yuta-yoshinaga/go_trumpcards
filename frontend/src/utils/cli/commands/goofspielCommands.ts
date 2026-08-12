import type { goofspielApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type GoofspielArgs = Parameters<typeof goofspielApi.exec>;

const VALID_COMMANDS = ['b', 'bid', 'n', 'next', 'h', 'hint', 'g', 'giveup', 'r', 'reset', 'log', 'l'];

/** Parse a Goofspiel CLI command into API exec arguments (indices are 0-based). */
export function parseGoofspielCommand(input: string): CliParseResult<GoofspielArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bid': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: b <cardIdx>' };
      return { args: ['bid', idx.value] as GoofspielArgs };
    }
    case 'n':
    case 'next':
      return { args: ['next'] as GoofspielArgs };
    case 'h':
    case 'hint':
      return { args: ['hint'] as GoofspielArgs };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] as GoofspielArgs };
    case 'log':
    case 'l':
      return { args: ['log'] as GoofspielArgs };
    case 'r':
    case 'reset':
      return { args: ['reset'] as GoofspielArgs };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Goofspiel CLI mode. */
export const GOOFSPIEL_HELP: string[] = [
  'b <cardIdx> - Bid that card face down (everyone bids at the same time)',
  'n/next      - Turn the next prize card',
  'h/hint      - Show a hint',
  'g/giveup    - Give up',
  'log         - Show the action log',
  'r/reset     - Reset game',
];
