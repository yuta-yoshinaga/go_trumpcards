import type { karnoffelApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by karnoffelApi.exec. */
export type KarnoffelCliArgs = Parameters<typeof karnoffelApi.exec>;

const VALID_COMMANDS = ['p', 'play', 'n', 'next', 'l', 'log', 'r', 'reset', 'help', '?'];

/** Cards per hand (sync: `KarnoffelHandSize`). **Five, not twelve.** */
const HAND_MAX_INDEX = 4;

/**
 * Parses a single CLI command line for the Karnöffel game into
 * {@link karnoffelApi}.exec arguments.
 *
 * There is only one action: playing a card. **Following suit is not required**,
 * so the index is bounded by the hand alone; the one restriction — that the
 * devil cannot lead the first trick — is decided server-side, because it
 * depends on the chosen suit rather than on anything the parser can see.
 */
export function parseKarnoffelCommand(input: string): CliParseResult<KarnoffelCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      const cardIndex = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(cardIndex) || cardIndex < 0 || cardIndex > HAND_MAX_INDEX) {
        return { error: `Usage: p <0-${HAND_MAX_INDEX}>` };
      }
      return { args: ['play', { cardIndex }] };
    }
    case 'n':
    case 'next':
      return { args: ['next'] };
    case 'l':
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

/** Help text shown in the CLI terminal for Karnöffel. */
export const KARNOFFEL_HELP: string[] = [
  'p <0-4>             - Play a hand card (five each; NO obligation to follow suit)',
  'n / next            - Deal the next hand',
  'l / log             - Show action log',
  'r / reset           - Reset game',
];
