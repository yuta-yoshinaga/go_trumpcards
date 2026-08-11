import type { mendikotApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type MendikotArgs = Parameters<typeof mendikotApi.exec>;

// **切り札を選ぶコマンドは無い。** フォローできずに出した札のスートがそのまま
// 切り札になるので、`t`/`trump` を受け付けると規則が二重になる。
const VALID_COMMANDS = ['p', 'play', 'n', 'next', 'h', 'hint', 'g', 'giveup', 'r', 'reset', 'log', 'l'];

/** Parse a Mendikot CLI command into API exec arguments (indices are 0-based). */
export function parseMendikotCommand(input: string): CliParseResult<MendikotArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: p <cardIdx>' };
      return { args: ['play', idx.value] as MendikotArgs };
    }
    case 'n':
    case 'next':
      return { args: ['next'] as MendikotArgs };
    case 'h':
    case 'hint':
      return { args: ['hint'] as MendikotArgs };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] as MendikotArgs };
    case 'log':
    case 'l':
      return { args: ['log'] as MendikotArgs };
    case 'r':
    case 'reset':
      return { args: ['reset'] as MendikotArgs };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Mendikot CLI mode. */
export const MENDIKOT_HELP: string[] = [
  'p <cardIdx>  - Play a card (marked * in your hand)',
  'n/next       - Start the next hand',
  'h/hint       - Show a hint',
  'g/giveup     - Give up',
  'log          - Show the action log',
  'r/reset      - Reset game',
];
