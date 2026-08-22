import type { threecardrummyApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type ThreeCardRummyArgs = Parameters<typeof threecardrummyApi.exec>;

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
  'l',
  'r',
  'reset',
  'h',
  'hint',
  'help',
  '?',
];

/** Parse a Three Card Rummy CLI command into API exec arguments. */
export function parseThreecardrummyCommand(input: string): CliParseResult<ThreeCardRummyArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bet': {
      const amount = parseIntArg(args, 0);
      if ('error' in amount) return { error: 'Usage: b <amount> [lowBonusBet]' };
      if (args.length >= 2) {
        const lowBonus = parseIntArg(args, 1);
        if ('error' in lowBonus) return { error: 'Usage: b <amount> [lowBonusBet]' };
        return { args: ['bet', amount.value, lowBonus.value] };
      }
      return { args: ['bet', amount.value] };
    }
    // **毎ラウンド同じ額を打ち直させない。** ボタン (tc-rebet-button) は
    // ワンクリックなのに、CLI は bet <ante> <lowBonus> を手打ちしていた (#5513)。
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
    case 'l':
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

/** Help text for Three Card Rummy CLI mode. */
export const THREECARDRUMMY_HELP: string[] = [
  'b <amt> [lb] - Ante bet (optional low bonus)',
  'rb/rebet     - Bet the same as last round',
  'p/play       - Play (match ante)',
  'f/fold       - Fold hand',
  'log          - Show action log',
  'r/reset      - Reset game',
  'h/hint       - Get a hint',
];
