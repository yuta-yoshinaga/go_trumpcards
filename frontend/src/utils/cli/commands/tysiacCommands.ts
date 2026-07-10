import type { tysiacApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type TysiacArgs = Parameters<typeof tysiacApi.exec>;

const VALID_COMMANDS = [
  'b',
  'bid',
  'd',
  'discard',
  'p',
  'play',
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

/** Parse a Tysiąc (Thousand) CLI command into API exec arguments. */
export function parseTysiacCommand(input: string): CliParseResult<TysiacArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bid': {
      const arg = args[0]?.toLowerCase();
      if (arg === 'raise' || arg === 'r' || arg === '+10') return { args: ['bid', { raise: true }] };
      if (arg === 'pass' || arg === 'p') return { args: ['bid', { raise: false }] };
      return { error: 'Usage: bid raise|pass' };
    }
    case 'd':
    case 'discard': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: d <idx>' };
      return { args: ['discard', { cardIndex: parsed.value }] };
    }
    case 'p':
    case 'play': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: p <idx>' };
      return { args: ['play', { cardIndex: parsed.value }] };
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

/** Help text for Tysiąc (Thousand) CLI mode. */
export const TYSIAC_HELP: string[] = [
  'bid raise|pass   - Raise the bid (+10) or pass (Bid phase)',
  'd <idx>          - Give a card to an opponent (Talon exchange, twice as Declarer)',
  'p <idx>          - Play a card (Play phase, must follow suit / trump when void)',
  'n/next           - Next trick',
  'nr/nextround     - Next round',
  'h/hint           - Show hint',
  'r/reset          - Reset game',
];
