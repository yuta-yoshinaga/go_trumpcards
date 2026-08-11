import type { israeliwhistApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type IsraeliWhistArgs = Parameters<typeof israeliwhistApi.exec>;

const VALID_COMMANDS = [
  'a',
  'auction',
  'pass',
  'b',
  'bid',
  'p',
  'play',
  'n',
  'next',
  'h',
  'hint',
  'g',
  'giveup',
  'r',
  'reset',
  'log',
  'l',
];

/** Suit codes accepted by `auction` (1=♠ 2=♣ 3=♥ 4=♦). */
const SUIT_MIN = 1;
const SUIT_MAX = 4;

/** The auction opens at five; calls run up to the whole hand. */
const AUCTION_MIN = 5;
const HAND_SIZE = 13;

/** Parse an Israeli Whist CLI command into API exec arguments (indices are 0-based). */
export function parseIsraeliWhistCommand(input: string): CliParseResult<IsraeliWhistArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'a':
    case 'auction': {
      // **入札は数とスートの両方で 1 つの意思決定。** 片方だけでは成立しない。
      const bid = parseIntArg(args, 0);
      if ('error' in bid) return { error: 'Usage: a <n 5-13> <suit 1-4>' };
      if (bid.value < AUCTION_MIN || bid.value > HAND_SIZE) return { error: 'Usage: a <n 5-13> <suit 1-4>' };
      const suit = parseIntArg(args, 1);
      if ('error' in suit) return { error: 'Usage: a <n 5-13> <suit 1-4>' };
      if (suit.value < SUIT_MIN || suit.value > SUIT_MAX) return { error: 'Usage: a <n 5-13> <suit 1-4>' };
      return { args: ['auction', undefined, undefined, suit.value, bid.value] as IsraeliWhistArgs };
    }
    case 'pass':
      return { args: ['pass'] as IsraeliWhistArgs };
    case 'b':
    case 'bid': {
      const bid = parseIntArg(args, 0);
      if ('error' in bid) return { error: 'Usage: b <0-13>' };
      if (bid.value < 0 || bid.value > HAND_SIZE) return { error: 'Usage: b <0-13>' };
      return { args: ['bid', undefined, undefined, undefined, bid.value] as IsraeliWhistArgs };
    }
    case 'p':
    case 'play': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: p <cardIdx>' };
      return { args: ['play', idx.value] as IsraeliWhistArgs };
    }
    case 'n':
    case 'next':
      return { args: ['next'] as IsraeliWhistArgs };
    case 'h':
    case 'hint':
      return { args: ['hint'] as IsraeliWhistArgs };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] as IsraeliWhistArgs };
    case 'log':
    case 'l':
      return { args: ['log'] as IsraeliWhistArgs };
    case 'r':
    case 'reset':
      return { args: ['reset'] as IsraeliWhistArgs };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Israeli Whist CLI mode. */
export const ISRAELIWHIST_HELP: string[] = [
  'a <n> <suit> - Bid in the auction (n 5-13; suit 1=S 2=C 3=H 4=D)',
  'pass         - Drop out of the auction',
  'b <0-13>     - Call your tricks',
  'p <cardIdx>  - Play a card (marked * in your hand)',
  'n/next       - Start the next round',
  'h/hint       - Show a hint',
  'g/giveup     - Give up',
  'log          - Show the action log',
  'r/reset      - Reset game',
];
