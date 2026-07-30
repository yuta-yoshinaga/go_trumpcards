import type { laughandliedownApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type LaughAndLieDownArgs = Parameters<typeof laughandliedownApi.exec>;

const VALID_COMMANDS = ['p', 'play', 'log', 'r', 'reset', 'help', '?'];

/** Parse a Laugh and Lie Down CLI command into API exec arguments. */
export function parseLaughAndLieDownCommand(input: string): CliParseResult<LaughAndLieDownArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      if (args.length === 0) return { error: `Usage: ${cmd} <index> [take]` };
      const idx = Number(args[0]);
      if (!Number.isInteger(idx) || idx < 0) return { error: `Invalid card index: ${args[0]}` };
      if (args.length === 1) return { args: ['play', idx, 1] };
      const take = Number(args[1]);
      // 1 と 3 だけが合法。ここで弾かず素通しするとサーバーのエラーになるので、
      // 打ち間違いはその場で返す。
      if (take !== 1 && take !== 3) return { error: `Take count must be 1 or 3: ${args[1]}` };
      return { args: ['play', idx, take] };
    }
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

/** Help text for Laugh and Lie Down CLI mode. */
export const LAUGHANDLIEDOWN_HELP: string[] = [
  'p <i>           - Capture one table card with the hand card at i',
  'p <i> 3         - Capture three of that rank (only when three are on the table)',
  'log             - Show action log',
  'r/reset         - New game',
];
