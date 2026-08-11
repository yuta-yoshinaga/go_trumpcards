import type { bhabhiApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type BhabhiArgs = Parameters<typeof bhabhiApi.exec>;

// **次のハンドへ進むコマンドは無い。** 配り切りの 1 ゲームで最後の 1 人まで
// 続くので、`n`/`next` を受け付けるとありもしない区切りを案内することになる。
const VALID_COMMANDS = ['p', 'play', 'h', 'hint', 'g', 'giveup', 'r', 'reset', 'log', 'l'];

/** Parse a Bhabhi CLI command into API exec arguments (indices are 0-based). */
export function parseBhabhiCommand(input: string): CliParseResult<BhabhiArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: p <cardIdx>' };
      return { args: ['play', idx.value] as BhabhiArgs };
    }
    case 'h':
    case 'hint':
      return { args: ['hint'] as BhabhiArgs };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] as BhabhiArgs };
    case 'log':
    case 'l':
      return { args: ['log'] as BhabhiArgs };
    case 'r':
    case 'reset':
      return { args: ['reset'] as BhabhiArgs };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Bhabhi CLI mode. */
export const BHABHI_HELP: string[] = [
  'p <cardIdx>  - Play a card (marked * in your hand)',
  'h/hint       - Show a hint',
  'g/giveup     - Give up',
  'log          - Show the action log',
  'r/reset      - Reset game',
];
