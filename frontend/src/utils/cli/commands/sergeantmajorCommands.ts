import type { sergeantmajorApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type SergeantMajorArgs = Parameters<typeof sergeantmajorApi.exec>;

// **ノルマを宣言するコマンドは無い。** 8・5・3 は席順で決まる。
const VALID_COMMANDS = [
  't',
  'trump',
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

/** Suit codes accepted by `trump` (1=♠ 2=♣ 3=♥ 4=♦). */
const SUIT_MIN = 1;
const SUIT_MAX = 4;

/** How many cards the dealer discards for the kitty. */
const DISCARD_COUNT = 4;

/** Parse a Sergeant Major CLI command into API exec arguments (indices are 0-based). */
export function parseSergeantMajorCommand(input: string): CliParseResult<SergeantMajorArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 't':
    case 'trump': {
      const suit = parseIntArg(args, 0);
      if ('error' in suit) return { error: 'Usage: t <suit 1-4>' };
      // **スート無しの宣言を通さない。** 既定値で埋めると選んでいないスートが
      // そのラウンドの切り札になる。
      if (suit.value < SUIT_MIN || suit.value > SUIT_MAX) return { error: 'Usage: t <suit 1-4>' };
      return { args: ['trump', undefined, undefined, suit.value] as SergeantMajorArgs };
    }
    case 'd':
    case 'discard': {
      const indices: number[] = [];
      for (let i = 0; i < DISCARD_COUNT; i++) {
        const v = parseIntArg(args, i);
        if ('error' in v) return { error: 'Usage: d <i> <i> <i> <i>' };
        indices.push(v.value);
      }
      return { args: ['discard', undefined, undefined, undefined, indices] as SergeantMajorArgs };
    }
    case 'p':
    case 'play': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: p <cardIdx>' };
      return { args: ['play', idx.value] as SergeantMajorArgs };
    }
    case 'n':
    case 'next':
      return { args: ['next'] as SergeantMajorArgs };
    case 'h':
    case 'hint':
      return { args: ['hint'] as SergeantMajorArgs };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] as SergeantMajorArgs };
    case 'log':
    case 'l':
      return { args: ['log'] as SergeantMajorArgs };
    case 'r':
    case 'reset':
      return { args: ['reset'] as SergeantMajorArgs };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Sergeant Major CLI mode. */
export const SERGEANTMAJOR_HELP: string[] = [
  't <suit>     - Declare trump (1=S 2=C 3=H 4=D; the dealer only)',
  'd <i>x4      - Discard four cards for the kitty (the dealer only)',
  'p <cardIdx>  - Play a card (marked * in your hand)',
  'n/next       - Start the next round',
  'h/hint       - Show a hint',
  'g/giveup     - Give up',
  'log          - Show the action log',
  'r/reset      - Reset game',
];
