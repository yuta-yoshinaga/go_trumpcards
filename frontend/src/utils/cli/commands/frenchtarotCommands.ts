import type { frenchtarotApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type FrenchTarotArgs = Parameters<typeof frenchtarotApi.exec>;

const VALID_COMMANDS = [
  'b',
  'bid',
  'pass',
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

/** Maps a bid shorthand to the backend contract string. */
const BID_ALIASES: Readonly<Record<string, string>> = {
  petite: 'petite',
  p: 'petite',
  garde: 'garde',
  g: 'garde',
  gardesans: 'gardesans',
  gs: 'gardesans',
  'garde-sans': 'gardesans',
  gardecontre: 'gardecontre',
  gc: 'gardecontre',
  'garde-contre': 'gardecontre',
};

/** Parse a French Tarot (フレンチタロット) CLI command into API exec arguments. */
export function parseFrenchTarotCommand(input: string): CliParseResult<FrenchTarotArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bid': {
      const arg = args[0]?.toLowerCase() ?? '';
      const bid = BID_ALIASES[arg];
      if (!bid) return { error: 'Usage: bid <petite|garde|gardesans|gardecontre>' };
      return { args: ['bid', { bid }] };
    }
    case 'pass':
      return { args: ['pass'] };
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

/** Help text for French Tarot (フレンチタロット) CLI mode. */
export const FRENCH_TAROT_HELP: string[] = [
  'bid petite                       - Declare Petite (chien revealed, exchange)',
  'bid garde                        - Declare Garde (chien revealed, exchange, ×2)',
  'bid gardesans                    - Declare Garde Sans (chien to declarer, ×4)',
  'bid gardecontre                  - Declare Garde Contre (chien to defenders, ×6)',
  'pass                             - Pass the auction',
  'discard <i1..i6>                 - Bury 6 écart cards (Petite/Garde declarer)',
  'p <idx>                          - Play a card (Play phase, must follow suit / trump)',
  'n/next                           - Next trick',
  'nr/nextround                     - Next deal',
  'h/hint                           - Show hint',
  'r/reset                          - Reset game',
];
