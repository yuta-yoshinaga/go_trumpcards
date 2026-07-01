import type { cinchApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type CinchArgs = Parameters<typeof cinchApi.exec>;

const VALID_COMMANDS = [
  'b',
  'bid',
  't',
  'trump',
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

/** Parse a Cinch (Double Pedro) CLI command into API exec arguments. */
export function parseCinchCommand(input: string): CliParseResult<CinchArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bid': {
      const arg = args[0]?.toLowerCase();
      if (arg === 'pass' || arg === 'p') return { args: ['bid', { bid: 0 }] };
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed || parsed.value < 0 || parsed.value > 14) return { error: 'Usage: bid <0-14> (0=pass)' };
      return { args: ['bid', { bid: parsed.value }] };
    }
    case 't':
    case 'trump': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed || parsed.value < 1 || parsed.value > 4) {
        return { error: 'Usage: t <1-4> (1=♠ 2=♣ 3=♥ 4=♦)' };
      }
      return { args: ['trump', { trumpSuit: parsed.value }] };
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

/** Help text for Cinch (Double Pedro) CLI mode. */
export const CINCH_HELP: string[] = [
  'bid <0-14>   - Bid 0 (pass) or 1-14 (Bid phase)',
  't <1-4>      - Name trump: 1=♠ 2=♣ 3=♥ 4=♦ (after winning the bid)',
  'p <idx>      - Play a card (Play phase, must follow the led suit)',
  'nr/nextround - Next deal (score the deal)',
  'h/hint       - Show hint',
  'r/reset      - Reset game',
];
