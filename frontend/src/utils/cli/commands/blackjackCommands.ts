import type { blackjackApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type BjArgs = Parameters<typeof blackjackApi.exec>;

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
      if ('error' in parsed) return { error: 'Usage: b <amount>' };
      return { args: ['bet', parsed.value] };
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
  'b <amount>  - Place bet',
  'hint        - Toggle strategy hint',
  'sd <n>      - Set deck count',
  'scc <0-3>   - Set CPU count',
  'r/reset     - Reset game',
];
