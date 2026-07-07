import type { koenigrufenApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type KoenigrufenArgs = Parameters<typeof koenigrufenApi.exec>;

const VALID_COMMANDS = [
  'b',
  'bid',
  'pass',
  'ck',
  'callking',
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

/** Maps a King-suit shorthand or name to the backend callSuit index (1=♠ 2=♣ 3=♥ 4=♦). */
const SUIT_ALIASES: Readonly<Record<string, number>> = {
  '1': 1,
  s: 1,
  spade: 1,
  spades: 1,
  '2': 2,
  c: 2,
  club: 2,
  clubs: 2,
  clover: 2,
  '3': 3,
  h: 3,
  heart: 3,
  hearts: 3,
  '4': 4,
  d: 4,
  diamond: 4,
  diamonds: 4,
};

/** Parse a Königrufen (ケーニッヒルーフェン) CLI command into API exec arguments. */
export function parseKoenigrufenCommand(input: string): CliParseResult<KoenigrufenArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bid': {
      const arg = args[0]?.toLowerCase() ?? '';
      if (arg !== 'rufer' && arg !== 'r') return { error: 'Usage: bid rufer' };
      return { args: ['bid', { bid: 'rufer' }] };
    }
    case 'pass':
      return { args: ['pass'] };
    case 'ck':
    case 'callking': {
      const arg = args[0]?.toLowerCase() ?? '';
      const suit = SUIT_ALIASES[arg];
      if (!suit) return { error: 'Usage: callking <1-4|spade|club|heart|diamond>' };
      return { args: ['callking', { callSuit: suit }] };
    }
    case 'd':
    case 'discard': {
      const indices: number[] = [];
      for (let i = 0; i < 6; i++) {
        const parsed = parseIntArg(args, i);
        if ('error' in parsed) return { error: 'Usage: discard <i1> <i2> <i3> <i4> <i5> <i6>' };
        indices.push(parsed.value);
      }
      return { args: ['discard', { cardIndices: indices }] };
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

/** Help text for Königrufen (ケーニッヒルーフェン) CLI mode. */
export const KOENIGRUFEN_HELP: string[] = [
  'bid rufer                        - Declare the Rufer (King-calling) contract',
  'pass                             - Pass the auction',
  'callking <1-4>                   - Call a King suit (1=♠ 2=♣ 3=♥ 4=♦) to name your partner',
  'discard <i1..i6>                 - Bury 6 talon cards (declarer)',
  'p <idx>                          - Play a card (Play phase, must follow suit / trump)',
  'n/next                           - Next trick',
  'nr/nextround                     - Next deal',
  'h/hint                           - Show hint',
  'r/reset                          - Reset game',
];
