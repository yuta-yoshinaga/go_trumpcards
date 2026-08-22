import type { ramschApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type RamschArgs = Parameters<typeof ramschApi.exec>;

/**
 * **入札系のコマンドは無い。** Ramsch には競りもスカット取りもゲーム宣言も
 * 無いので、Skat から持ってきた b/pa/ps/sk/d/g は受け付けない ── 残しておくと
 * 「打てるのに何も起きないコマンド」になる。
 */
const VALID_COMMANDS = ['p', 'play', 'n', 'next', 'nr', 'nextround', 'h', 'hint', 'log', 'r', 'reset', 'help', '?'];

/** Parse a Ramsch CLI command into API exec arguments. */
export function parseRamschCommand(input: string): CliParseResult<RamschArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      const ci = parseIntArg(args, 0);
      if ('error' in ci) return { error: 'Usage: p <cardIdx>' };
      return { args: ['play', { cardIndex: ci.value }] };
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
    case 'log':
      return { args: ['log'] };
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

/** Help text for Ramsch CLI mode. */
export const RAMSCH_HELP: string[] = [
  'p <idx>       - Play card from hand by index',
  'n/next        - Next trick',
  'nr/nextround  - Next round',
  'h/hint        - Show suggested move',
  'log           - Show action log',
  'r/reset       - Reset match',
];
