import type { threecardApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type ThreeCardArgs = Parameters<typeof threecardApi.exec>;

const VALID_COMMANDS = [
  'b',
  'bet',
  'rb',
  'rebet',
  'p',
  'play',
  'f',
  'fold',
  'log',
  'r',
  'reset',
  'h',
  'hint',
  'help',
  '?',
];

/** Parse a Three Card Poker CLI command into API exec arguments. */
export function parseThreecardCommand(input: string): CliParseResult<ThreeCardArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bet': {
      const amount = parseIntArg(args, 0);
      if ('error' in amount) return { error: 'Usage: b <amount> [pairPlusBet]' };
      if (args.length >= 2) {
        const ppBet = parseIntArg(args, 1);
        if ('error' in ppBet) return { error: 'Usage: b <amount> [pairPlusBet]' };
        return { args: ['bet', amount.value, ppBet.value] };
      }
      return { args: ['bet', amount.value] };
    }
    // **毎ラウンド同じ額を打ち直させない。** ボタン (tc-rebet-button) は
    // ワンクリックなのに、CLI は bet <ante> <pairPlus> を手打ちしていた (#5513)。
    // 金額はサーバが覚えているので送らない。
    case 'rb':
    case 'rebet':
      return { args: ['rebet'] };
    case 'p':
    case 'play':
      return { args: ['play'] };
    case 'f':
    case 'fold':
      return { args: ['fold'] };
    case 'log':
      return { args: ['log'] };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Three Card Poker CLI mode. */
export const THREECARD_HELP: string[] = [
  'b <amt> [pp] - Ante bet (optional pair plus)',
  'rb/rebet     - Bet the same as last round',
  'p/play       - Play (match ante)',
  'f/fold       - Fold hand',
  'log          - Show action log',
  'r/reset      - Reset game',
  'h/hint       - Get a hint',
];
