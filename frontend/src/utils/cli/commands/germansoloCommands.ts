import type { germansoloApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type GermanSoloArgs = Parameters<typeof germansoloApi.exec>;

const VALID_COMMANDS = [
  'b',
  'bid',
  'a',
  'ace',
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

/** Maps a suit letter to its numeric code (1=♠ 2=♣ 3=♥ 4=♦). */
const SUIT_CODES: Readonly<Record<string, number>> = { s: 1, c: 2, h: 3, d: 4 };

/**
 * Maps a bid word (and its one-letter shorthand) to the wire value.
 *
 * **Mussfrage is absent on purpose.** It is what the table forces on the holder
 * of Spadille when everyone passes, not something a seat may declare; sending
 * it would just come back rejected.
 */
const BID_CODES: Readonly<Record<string, number>> = {
  frage: 2,
  f: 2,
  solo: 3,
  s: 3,
  tout: 4,
  t: 4,
};

/** Parse a German Solo CLI command into API exec arguments. */
export function parseGermanSoloCommand(input: string): CliParseResult<GermanSoloArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bid': {
      const arg = args[0]?.toLowerCase();
      if (arg === 'pass' || arg === 'p') return { args: ['bid', { bid: 0 }] };
      const bid = BID_CODES[arg ?? ''];
      if (bid !== undefined) {
        const suit = SUIT_CODES[args[1]?.toLowerCase() ?? ''];
        if (!suit) return { error: 'Usage: bid frage|solo|tout <s|c|h|d>' };
        return { args: ['bid', { bid, trumpSuit: suit }] };
      }
      return { error: 'Usage: bid pass | bid frage|solo|tout <s|c|h|d>' };
    }
    // **The only way out of the ace-call phase.** Without it the board freezes
    // right after a Frage is won: `play` is rejected until an ace is named.
    case 'a':
    case 'ace': {
      const suit = SUIT_CODES[args[0]?.toLowerCase() ?? ''];
      if (!suit) return { error: 'Usage: ace <s|c|h|d>' };
      return { args: ['ace', { aceSuit: suit }] };
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

/** Help text for German Solo CLI mode. */
export const GERMAN_SOLO_HELP: string[] = [
  'bid pass                     - Pass in the auction (Bid phase)',
  'bid frage <s|c|h|d>          - Declare Frage: call an ace for a hidden partner, 5 tricks',
  'bid solo <s|c|h|d>           - Declare Solo: five tricks alone',
  'bid tout <s|c|h|d>           - Declare Tout: all eight tricks alone',
  'a/ace <s|c|h|d>              - Call the ace that names your partner (AceCall phase)',
  'p <idx>                      - Play a card (Play phase, must follow the led suit)',
  'n/next                       - Next trick',
  'nr/nextround                 - Next round',
  'h/hint                       - Show hint',
  'r/reset                      - Reset game',
];
