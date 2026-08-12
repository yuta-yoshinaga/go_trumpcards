import type { botifarraApi } from '../../../api/gameApi';
import { BOTIFARRA_NO_TRUMP } from '../../../types/games/botifarra';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type BotifarraArgs = Parameters<typeof botifarraApi.exec>;

const VALID_COMMANDS = [
  'p',
  'play',
  'declare',
  'delegate',
  'double',
  'pass',
  'next',
  'giveup',
  'hint',
  'h',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

const DECLARE_USAGE = 'Usage: declare <spade|clover|heart|diamond|none>';

/**
 * Maps a trump token to its suit value, or null if unknown.
 *
 * **Suits are 1..4 and no-trump is -1**, so 0 is not a legal value here — a
 * missing argument must not fall through as "spades".
 */
function parseTrump(token: string | undefined): number | null {
  switch ((token ?? '').toLowerCase()) {
    case 's':
    case 'spade':
      return 1;
    case 'c':
    case 'club':
    case 'clover':
      return 2;
    case 'h':
    case 'heart':
      return 3;
    case 'd':
    case 'diamond':
      return 4;
    case 'n':
    case 'none':
    case 'notrump':
      return BOTIFARRA_NO_TRUMP;
    default:
      return null;
  }
}

/** Parse a Botifarra CLI command into API exec arguments. */
export function parseBotifarraCommand(input: string): CliParseResult<BotifarraArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: play <cardIndex>' };
      return { args: ['play', idx.value] };
    }
    case 'declare': {
      const suit = parseTrump(args[0]);
      if (suit === null) return { error: DECLARE_USAGE };
      return { args: ['declare', undefined, suit] };
    }
    case 'delegate':
      return { args: ['delegate'] };
    case 'double':
      return { args: ['double'] };
    case 'pass':
      return { args: ['passdouble'] };
    case 'next':
      return { args: ['next'] };
    case 'giveup':
      return { args: ['giveup'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
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

/** Help text for Botifarra CLI mode. */
export const BOTIFARRA_CLI_HELP: string[] = [
  'play <idx>                   - Play the card at <idx>',
  'declare <s|c|h|d|none>       - Declare trump (none = botifarra)',
  'delegate                     - Hand the declaration to your partner',
  'double / pass                - Double the stake / let it stand',
  'next                         - Deal the next round',
  'giveup                       - Resign',
  'hint                         - Show a hint',
  'log                          - Show action log',
  'r/reset                      - Reset game',
];
