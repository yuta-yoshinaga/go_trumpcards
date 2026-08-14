import type { teendopaanchApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type TeenDoPaanchArgs = Parameters<typeof teendopaanchApi.exec>;

// **ノルマを宣言するコマンドは無い。** 3・2・5 は割り当てで選ぶ余地が無い。
const VALID_COMMANDS = ['t', 'trump', 'p', 'play', 'n', 'next', 'h', 'hint', 'g', 'giveup', 'r', 'reset', 'log', 'l'];

/** Suit codes accepted by `trump` (1=♠ 2=♣ 3=♥ 4=♦). */
const SUIT_MIN = 1;
const SUIT_MAX = 4;

/** Parse a 3-2-5 CLI command into API exec arguments (indices are 0-based). */
export function parseTeenDoPaanchCommand(input: string): CliParseResult<TeenDoPaanchArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 't':
    case 'trump': {
      const suit = parseIntArg(args, 0);
      if ('error' in suit) return { error: 'Usage: t <suit 1-4>' };
      // **スート無しの宣言を通さない。** 既定値で埋めると、選んでいない
      // スートがそのラウンドの切り札になる。
      if (suit.value < SUIT_MIN || suit.value > SUIT_MAX) return { error: 'Usage: t <suit 1-4>' };
      return { args: ['trump', undefined, undefined, suit.value] as TeenDoPaanchArgs };
    }
    case 'p':
    case 'play': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: p <cardIdx>' };
      return { args: ['play', idx.value] as TeenDoPaanchArgs };
    }
    case 'n':
    case 'next':
      return { args: ['next'] as TeenDoPaanchArgs };
    case 'h':
    case 'hint':
      return { args: ['hint'] as TeenDoPaanchArgs };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] as TeenDoPaanchArgs };
    case 'log':
    case 'l':
      return { args: ['log'] as TeenDoPaanchArgs };
    case 'r':
    case 'reset':
      return { args: ['reset'] as TeenDoPaanchArgs };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for 3-2-5 CLI mode. */
export const TEENDOPAANCH_HELP: string[] = [
  't <suit>     - Declare trump from your first five cards (1=S 2=C 3=H 4=D; the 5-target seat only)',
  'p <cardIdx>  - Play a card (marked * in your hand)',
  'n/next       - Start the next round',
  'h/hint       - Show a hint',
  'g/giveup     - Give up',
  'log          - Show the action log',
  'r/reset      - Reset game',
];
