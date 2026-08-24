import type { horseApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type HorseArgs = Parameters<typeof horseApi.exec>;

const VALID_COMMANDS = [
  'f',
  'fold',
  'd',
  'draw',
  'sp',
  'stand',
  'x',
  'check',
  'c',
  'call',
  'b',
  'bet',
  'raise',
  'allin',
  'n',
  'next',
  'h',
  'hint',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a H.O.R.S.E. CLI command into API exec arguments. */
export function parseHorseCommand(input: string): CliParseResult<HorseArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'f':
    case 'fold':
      return { args: ['action', { action: 'fold' }] };
    case 'x':
    case 'check':
      return { args: ['action', { action: 'check' }] };
    case 'c':
    case 'call':
      return { args: ['action', { action: 'call' }] };
    case 'b':
    case 'bet': {
      // **額の無いベットは送らない。** 0 として送るとサーバに断られるだけで、
      // 何を書き忘れたのか画面から分からない。
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: b <amount>' };
      return { args: ['action', { action: 'bet', amount: parsed.value }] };
    }
    case 'raise': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: raise <amount>' };
      return { args: ['action', { action: 'raise', amount: parsed.value }] };
    }
    case 'allin':
      return { args: ['action', { action: 'allin' }] };
    case 'd':
    case 'draw': {
      // **番号は 0 始まり。** 引き直しのある他のゲーム (2-7 単体・ムス) と
      // 同じ数え方にしないと、同じ操作が 1 枚ずれる。
      const indices: number[] = [];
      for (const raw of args) {
        const n = Number.parseInt(raw, 10);
        if (Number.isNaN(n) || String(n) !== raw.trim() || n < 0) return { error: 'Usage: d <idx>...' };
        indices.push(n);
      }
      // **引数無しは打ち間違い。** 引かないと決めたなら sp と打つ ── ここで
      // スタンドパットに読み替えると、番号を書き忘れた手が黙って通る。
      if (indices.length === 0) return { error: 'Usage: d <idx>...' };
      return { args: ['draw', { cardIndices: indices }] };
    }
    case 'sp':
    case 'stand':
      return { args: ['draw', { cardIndices: [] }] };
    case 'n':
    case 'next':
      return { args: ['next'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
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

/** Help text for Eight-Game Mix CLI mode. */
export const EIGHT_GAME_HELP: string[] = [
  'f / fold                         - Fold this hand',
  'x / check                        - Check',
  'c / call                         - Call the outstanding bet',
  'b <amount>                       - Bet the given amount',
  'raise <amount>                   - Raise to the given amount',
  'allin                            - Push every chip in',
  'd <idx>...                       - Exchange those cards (2-7 Triple Draw, 0-based)',
  'sp / stand                       - Stand pat (2-7 Triple Draw)',
  'n / next                         - Deal the next hand (the discipline may change)',
  'h / hint                         - Show a hint',
  'r / reset                        - Start a new match',
];

/** Help text for H.O.R.S.E. CLI mode. */
export const HORSE_HELP: string[] = [
  'f / fold                         - Fold this hand',
  'x / check                        - Check',
  'c / call                         - Call the outstanding bet',
  'b <amount>                       - Bet the given amount',
  'raise <amount>                   - Raise to the given amount',
  'allin                            - Push every chip in',
  'n / next                         - Deal the next hand (the discipline may change)',
  'h / hint                         - Show a hint',
  'r / reset                        - Start a new match',
];
