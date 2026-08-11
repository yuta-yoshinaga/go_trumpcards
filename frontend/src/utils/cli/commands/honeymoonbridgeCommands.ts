import type { honeymoonbridgeApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type HoneymoonBridgeArgs = Parameters<typeof honeymoonbridgeApi.exec>;

const VALID_COMMANDS = [
  'b',
  'bid',
  'pass',
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

/** Contract levels accepted by `bid`. Level n needs 6 + n tricks. */
const LEVEL_MIN = 1;
const LEVEL_MAX = 7;

/** Suit codes accepted by `bid`. **0 is no-trump, not "unset".** */
const SUIT_MIN = 0;
const SUIT_MAX = 4;

/** Parse a Honeymoon Bridge CLI command into API exec arguments (indices are 0-based). */
export function parseHoneymoonBridgeCommand(input: string): CliParseResult<HoneymoonBridgeArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bid': {
      const level = parseIntArg(args, 0);
      if ('error' in level) return { error: 'Usage: b <level 1-7> <suit 0-4>' };
      if (level.value < LEVEL_MIN || level.value > LEVEL_MAX) return { error: 'Usage: b <level 1-7> <suit 0-4>' };
      // **スートを既定値で埋めない。** 0 はノートランプという明示の選択肢なので、
      // 省略を 0 に丸めると宣言していない契約を落札してしまう。
      const suit = parseIntArg(args, 1);
      if ('error' in suit) return { error: 'Usage: b <level 1-7> <suit 0-4>' };
      if (suit.value < SUIT_MIN || suit.value > SUIT_MAX) return { error: 'Usage: b <level 1-7> <suit 0-4>' };
      return { args: ['bid', undefined, undefined, level.value, suit.value] as HoneymoonBridgeArgs };
    }
    case 'pass':
      return { args: ['pass'] as HoneymoonBridgeArgs };
    case 'p':
    case 'play': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: p <cardIdx>' };
      return { args: ['play', idx.value] as HoneymoonBridgeArgs };
    }
    case 'n':
    case 'next':
      return { args: ['next'] as HoneymoonBridgeArgs };
    case 'h':
    case 'hint':
      return { args: ['hint'] as HoneymoonBridgeArgs };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] as HoneymoonBridgeArgs };
    case 'log':
    case 'l':
      return { args: ['log'] as HoneymoonBridgeArgs };
    case 'r':
    case 'reset':
      return { args: ['reset'] as HoneymoonBridgeArgs };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Honeymoon Bridge CLI mode. */
export const HONEYMOONBRIDGE_HELP: string[] = [
  'b <lvl> <s>  - Bid a contract (level 1-7; suit 0=NT 1=S 2=C 3=H 4=D)',
  'pass         - Drop out of the auction',
  'p <cardIdx>  - Play a card (marked * in your hand)',
  'n/next       - Start the next deal',
  'h/hint       - Show a hint',
  'g/giveup     - Give up',
  'log          - Show the action log',
  'r/reset      - Reset game',
];
