import type { shelemApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type ShelemArgs = Parameters<typeof shelemApi.exec>;

const VALID_COMMANDS = [
  'b',
  'bid',
  'shelem',
  'pass',
  'd',
  'discard',
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

/**
 * Bidding runs from 55 to 100 in steps of five (sync:
 * `ShelemMinBid`/`ShelemMaxBid` in `internal/domain/Shelem.go`).
 *
 * **The ceiling is the whole round's card points.** A contract above 100 could
 * never be made, so it is not offered.
 */
const BID_MIN = 55;
const BID_MAX = 100;

/** Suit codes accepted by `discard` (1=♠ 2=♣ 3=♥ 4=♦). */
const SUIT_MIN = 1;
const SUIT_MAX = 4;

/** How many cards the declarer discards after taking the widow. */
const DISCARD_COUNT = 4;

/** Parse a Shelem CLI command into API exec arguments (indices are 0-based). */
export function parseShelemCommand(input: string): CliParseResult<ShelemArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bid': {
      const bid = parseIntArg(args, 0);
      if ('error' in bid) return { error: 'Usage: b <55-100>' };
      if (bid.value < BID_MIN || bid.value > BID_MAX) return { error: 'Usage: b <55-100>' };
      return { args: ['bid', undefined, undefined, undefined, bid.value] as ShelemArgs };
    }
    case 'shelem':
      return { args: ['shelem'] as ShelemArgs };
    case 'pass':
      return { args: ['pass'] as ShelemArgs };
    case 'd':
    case 'discard': {
      // **4つのインデックスとスートで1つの意思決定。** どれが欠けても成立しない。
      const discards: number[] = [];
      for (let i = 0; i < DISCARD_COUNT; i++) {
        const idx = parseIntArg(args, i);
        if ('error' in idx) return { error: 'Usage: d <i> <i> <i> <i> <suit 1-4>' };
        discards.push(idx.value);
      }
      const suit = parseIntArg(args, DISCARD_COUNT);
      if ('error' in suit) return { error: 'Usage: d <i> <i> <i> <i> <suit 1-4>' };
      if (suit.value < SUIT_MIN || suit.value > SUIT_MAX) {
        return { error: 'Usage: d <i> <i> <i> <i> <suit 1-4>' };
      }
      return { args: ['discard', undefined, undefined, suit.value, undefined, discards] as ShelemArgs };
    }
    case 'p':
    case 'play': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: p <cardIdx>' };
      return { args: ['play', idx.value] as ShelemArgs };
    }
    case 'n':
    case 'next':
      return { args: ['next'] as ShelemArgs };
    case 'h':
    case 'hint':
      return { args: ['hint'] as ShelemArgs };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] as ShelemArgs };
    case 'log':
    case 'l':
      return { args: ['log'] as ShelemArgs };
    case 'r':
    case 'reset':
      return { args: ['reset'] as ShelemArgs };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Shelem CLI mode. */
export const SHELEM_HELP: string[] = [
  'b <55-100>   - Bid that many points (in steps of 5; 100 is every card point)',
  'shelem       - Declare Shelem (take every trick)',
  'pass         - Drop out of the bidding',
  'd <i>x4 <s>  - Discard four cards and name trump (1=S 2=C 3=H 4=D)',
  'p <cardIdx>  - Play a card (marked * in your hand)',
  'n/next       - Start the next round',
  'h/hint       - Show a hint',
  'g/giveup     - Give up',
  'log          - Show the action log',
  'r/reset      - Reset game',
];
