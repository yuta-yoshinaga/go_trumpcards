import type { blackjackApi } from '../../../api/gameApi';
import type { BlackJackBetOptions } from '../../../api/games/blackjack';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type BjArgs = Parameters<typeof blackjackApi.exec>;

/** Usage line for the bet command, shared by every rejection path. */
const BET_USAGE = 'Usage: b <amount> [ppBet] [t3Bet] [handCount]';

const VALID_COMMANDS = [
  'h',
  'hit',
  's',
  'stand',
  'd',
  'doubledown',
  'sp',
  'split',
  'i',
  'insurance',
  'di',
  'declineinsurance',
  'sur',
  'surrender',
  'es',
  'earlysurrender',
  'des',
  'declineearlysurrender',
  'b',
  'bet',
  'hint',
  'soft17',
  'counting',
  'das',
  'sd',
  'setdeckcount',
  'scc',
  'setcpucount',
  'scs',
  'setcountingsystem',
  'pen',
  'setpenetration',
  'ssr',
  'setsurrenderrule',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a BlackJack CLI command into API exec arguments. */
export function parseBlackjackCommand(input: string): CliParseResult<BjArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'h':
    case 'hit':
      return { args: ['hit'] };
    case 's':
    case 'stand':
      return { args: ['stand'] };
    case 'd':
    case 'doubledown':
      return { args: ['doubledown'] };
    case 'sp':
    case 'split':
      return { args: ['split'] };
    case 'i':
    case 'insurance':
      return { args: ['insurance'] };
    case 'di':
    case 'declineinsurance':
      return { args: ['declineinsurance'] };
    case 'sur':
    case 'surrender':
      return { args: ['surrender'] };
    case 'es':
    case 'earlysurrender':
      return { args: ['earlysurrender'] };
    case 'des':
    case 'declineearlysurrender':
      return { args: ['declineearlysurrender'] };
    case 'hint':
    case 'togglehint':
      return { args: ['togglehint'] };
    case 'soft17':
    case 'togglesoft17':
      return { args: ['togglesoft17'] };
    case 'counting':
    case 'togglecounting':
      return { args: ['togglecounting'] };
    case 'das':
    case 'toggledas':
      return { args: ['toggledas'] };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    case 'b':
    case 'bet': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: BET_USAGE };
      // サーバ (BlackJackCuiController) は amount / ppBet / t3Bet / handCount の
      // 4引数を受ける。金額しか読まないと、同じページのベットフォームと独立 CUI で
      // 使えるサイドベットと複数ハンドが CLI だけ使えない (#5474)。
      // **余った引数を黙って捨てない。** `b 100 20 30 2 oops` が通ると、
      // 打ち間違いに気づかないまま意図と違うベットが成立する。
      if (args.length > 4) return { error: BET_USAGE };
      const options: BlackJackBetOptions = {};
      const extras: [keyof BlackJackBetOptions, number][] = [
        ['perfectPairsBet', 0],
        ['twentyOnePlus3Bet', 0],
        ['handCount', 1],
      ];
      for (let i = 0; i < extras.length; i++) {
        const [key, min] = extras[i] as [keyof BlackJackBetOptions, number];
        if (args.length <= i + 1) break;
        const extra = parseIntArg(args, i + 1);
        // **数字でない引数を黙って捨てない。** 捨てると `b 100 xx 30` が
        // 「21+3 に 30」ではない別の意味で通ってしまう。
        if ('error' in extra) return { error: BET_USAGE };
        if (extra.value < min) return { error: BET_USAGE };
        options[key] = extra.value;
      }
      // **省略した引数は送らない。** 0 を送ると「0 を賭ける」と「賭けない」が
      // サーバ側で区別できない。
      if (Object.keys(options).length === 0) return { args: ['bet', parsed.value] };
      return { args: ['bet', parsed.value, undefined, options] };
    }
    case 'sd':
    case 'setdeckcount': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: sd <count>' };
      return { args: ['setdeckcount', parsed.value] };
    }
    case 'scc':
    case 'setcpucount': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: scc <0-3>' };
      return { args: ['setcpucount', parsed.value] };
    }
    case 'scs':
    case 'setcountingsystem': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: scs <0-3>' };
      return { args: ['setcountingsystem', parsed.value] };
    }
    case 'pen':
    case 'setpenetration': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: pen <percent>' };
      return { args: ['setpenetration', parsed.value] };
    }
    case 'ssr':
    case 'setsurrenderrule': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: ssr <0-2>' };
      return { args: ['setsurrenderrule', parsed.value] };
    }
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for BlackJack CLI mode. */
export const BLACKJACK_HELP: string[] = [
  'h/hit       - Draw a card',
  's/stand     - End turn',
  'd/doubledown- Double down',
  'sp/split    - Split pair',
  'i/insurance - Take insurance',
  'di          - Decline insurance',
  'sur/surrender- Surrender',
  'b <amount> [ppBet] [t3Bet] [handCount] - Place bet (side bets / multi-hand)',
  'hint        - Toggle strategy hint',
  'sd <n>      - Set deck count',
  'scc <0-3>   - Set CPU count',
  'r/reset     - Reset game',
];
