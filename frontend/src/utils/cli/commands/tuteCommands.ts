import type { tuteApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type TuteArgs = Parameters<typeof tuteApi.exec>;

const VALID_COMMANDS = [
  'p',
  'play',
  'm',
  'marriage',
  'tute',
  'n',
  'next',
  'nr',
  'nextround',
  'h',
  'hint',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Tute CLI command into API exec arguments. */
export function parseTuteCommand(input: string): CliParseResult<TuteArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: p <idx>' };
      return { args: ['play', { cardIndex: parsed.value }] };
    }
    case 'm':
    case 'marriage': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: m <suit> (1=♠ 2=♣ 3=♥ 4=♦)' };
      return { args: ['marriage', { suit: parsed.value }] };
    }
    case 'tute':
      return { args: ['tute'] };
    case 'n':
    case 'next':
      return { args: ['next'] };
    case 'nr':
    case 'nextround':
      return { args: ['nextround'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
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

/** Help text for Tute CLI mode. */
export const TUTE_HELP: string[] = [
  'p <idx>          - Play a card (Play phase, must follow suit)',
  'm <suit>         - Declare a King+Queen marriage (1=♠ 2=♣ 3=♥ 4=♦)',
  'tute             - Declare Tute (four Kings or four Queens) for an instant win',
  'n/next           - Next trick',
  'nr/nextround     - Next round',
  'h/hint           - Show hint',
  'r/reset          - Reset game',
];
