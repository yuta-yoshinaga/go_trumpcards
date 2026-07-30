import type { toepenApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type ToepenArgs = Parameters<typeof toepenApi.exec>;

const VALID_COMMANDS = [
  'p',
  'play',
  't',
  'toep',
  's',
  'stay',
  'f',
  'fold',
  'n',
  'next',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Toepen CLI command into API exec arguments. */
export function parseToepenCommand(input: string): CliParseResult<ToepenArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      if (args.length === 0) return { error: 'Usage: p <index>' };
      const n = Number(args[0]);
      if (!Number.isInteger(n) || n < 0) return { error: `Invalid card index: ${args[0]}` };
      return { args: ['play', n] };
    }
    case 't':
    case 'toep':
      return { args: ['toep'] };
    // stay and fold are the same command with opposite answers; keeping them
    // as separate words means a mistyped boolean cannot silently invert.
    case 's':
    case 'stay':
      return { args: ['answer', undefined, true] };
    case 'f':
    case 'fold':
      return { args: ['answer', undefined, false] };
    case 'd':
    case 'redeal':
      return { args: ['redeal'] };
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

/** Help text for Toepen CLI mode. */
export const TOEPEN_HELP: string[] = [
  'p <i>           - Play the hand card at index i',
  't/toep          - Raise the stake by one',
  's/stay          - Stay in after a toep',
  'f/fold          - Fold to a toep',
  'd/redeal       - Throw in a poverty hand (A/K/Q/J only)',
  'n/next          - Start the next hand',
  'log             - Show action log',
  'r/reset         - New game',
];
