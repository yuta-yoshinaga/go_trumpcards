import type { knockoutWhistApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type KnockoutWhistArgs = Parameters<typeof knockoutWhistApi.exec>;

const VALID_COMMANDS = [
  'p',
  'play',
  'st',
  'selecttrump',
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

/** Parse a Knockout Whist CLI command into API exec arguments. */
export function parseKnockoutWhistCommand(input: string): CliParseResult<KnockoutWhistArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: p <idx>' };
      return { args: ['play', { cardIndex: parsed.value }] };
    }
    case 'st':
    case 'selecttrump': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed || parsed.value < 1 || parsed.value > 4) {
        return { error: 'Usage: st <1-4> (1=S 2=C 3=H 4=D)' };
      }
      return { args: ['selecttrump', { trumpSuit: parsed.value }] };
    }
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

/** Help text for Knockout Whist CLI mode. */
export const KNOCKOUT_WHIST_HELP: string[] = [
  'p <idx>          - Play a card (must follow the lead suit; trump beats non-trump; Ace high)',
  'st <1-4>         - Select trump when you won the round (1=S 2=C 3=H 4=D)',
  'n/next           - Next trick',
  'nr/nextround     - Next round (shrinking hand; zero-trick players spend a Dogbone or are eliminated)',
  'h/hint           - Show hint',
  'r/reset          - Reset game',
];
