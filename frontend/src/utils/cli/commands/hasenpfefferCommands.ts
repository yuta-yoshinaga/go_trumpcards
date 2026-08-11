import type { hasenpfefferApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type HasenpfefferArgs = Parameters<typeof hasenpfefferApi.exec>;

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
  'h',
  'hint',
  'g',
  'giveup',
  'r',
  'reset',
  'log',
  'l',
];

/** Bid range accepted by the server (3..6). */
const BID_MIN = 3;
const BID_MAX = 6;

/** Suit codes accepted by `discard` (1=♠ 2=♣ 3=♥ 4=♦). */
const SUIT_MIN = 1;
const SUIT_MAX = 4;

/** Parse a Hasenpfeffer CLI command into API exec arguments (indices are 0-based). */
export function parseHasenpfefferCommand(input: string): CliParseResult<HasenpfefferArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bid': {
      const bid = parseIntArg(args, 0);
      if ('error' in bid) return { error: `Usage: b <${BID_MIN}-${BID_MAX}>` };
      // **0 は pass 専用。** ここで通すと下限の検査がすり抜ける。
      if (bid.value < BID_MIN || bid.value > BID_MAX) return { error: `Usage: b <${BID_MIN}-${BID_MAX}>` };
      return { args: ['bid', undefined, undefined, undefined, bid.value] as HasenpfefferArgs };
    }
    case 'pass':
      return { args: ['bid', undefined, undefined, undefined, 0] as HasenpfefferArgs };
    case 'd':
    case 'discard': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: d <cardIdx> <suit 1-4>' };
      const suit = parseIntArg(args, 1);
      if ('error' in suit) return { error: 'Usage: d <cardIdx> <suit 1-4>' };
      // **スート無しの宣言を通さない。** 既定値で埋めると選んでいないスートが
      // そのハンドの切り札になる。
      if (suit.value < SUIT_MIN || suit.value > SUIT_MAX) return { error: 'Usage: d <cardIdx> <suit 1-4>' };
      return { args: ['discard', idx.value, undefined, suit.value] as HasenpfefferArgs };
    }
    case 'p':
    case 'play': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: p <cardIdx>' };
      return { args: ['play', idx.value] as HasenpfefferArgs };
    }
    case 'n':
    case 'next':
      return { args: ['next'] as HasenpfefferArgs };
    case 'h':
    case 'hint':
      return { args: ['hint'] as HasenpfefferArgs };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] as HasenpfefferArgs };
    case 'log':
    case 'l':
      return { args: ['log'] as HasenpfefferArgs };
    case 'r':
    case 'reset':
      return { args: ['reset'] as HasenpfefferArgs };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Hasenpfeffer CLI mode. */
export const HASENPFEFFER_HELP: string[] = [
  'b <n>        - Bid n tricks (3-6)',
  'pass         - Pass (the dealer cannot, once the other three have)',
  'd <i> <suit> - Discard card i and name trump (1=S 2=C 3=H 4=D; declarer only)',
  'p <cardIdx>  - Play a card (marked * in your hand)',
  'n/next       - Start the next hand',
  'h/hint       - Show a hint',
  'g/giveup     - Give up',
  'log          - Show the action log',
  'r/reset      - Reset game',
];
