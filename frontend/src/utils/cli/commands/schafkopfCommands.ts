import type { schafkopfApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type SchafkopfArgs = Parameters<typeof schafkopfApi.exec>;

const VALID_COMMANDS = [
  'pick',
  'wenz',
  'w',
  'solo',
  'so',
  'pass',
  'call',
  'p',
  'play',
  'n',
  'next',
  'nr',
  'nextround',
  'h',
  'hint',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Schafkopf CLI command into API exec arguments. */
export function parseSchafkopfCommand(input: string): CliParseResult<SchafkopfArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'pick':
      return { args: ['pick', { pick: true, contract: 0 }] };
    case 'wenz':
    case 'w':
      return { args: ['pick', { pick: true, contract: 1 }] };
    case 'solo':
    case 'so': {
      // Solo は切り札スートを引数に取る。既定値を置くと、指定を忘れた
      // 宣言が黙って別のスートの Solo になる。
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed || parsed.value < 1 || parsed.value > 4) {
        return { error: 'Usage: solo <suit> (1=spade 2=club 3=heart 4=diamond)' };
      }
      return { args: ['pick', { pick: true, contract: 2, soloSuit: parsed.value }] };
    }
    case 'pass':
      return { args: ['pick', { pick: false }] };
    case 'call': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: call <suit> (1=spade 2=club 3=heart)' };
      return { args: ['call', { callSuit: parsed.value }] };
    }
    case 'p':
    case 'play': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: p <idx>' };
      return { args: ['play', { cardIndex: parsed.value }] };
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

/** Help text for Schafkopf CLI mode. */
export const SCHAFKOPF_HELP: string[] = [
  'pick             - Declare Rufspiel, calling an ace (Pick phase)',
  'w/wenz           - Declare Wenz, only Unters are trump (Pick phase)',
  'so/solo <suit>   - Declare Solo, 1=spade 2=club 3=heart 4=diamond (Pick phase)',
  'pass             - Pass, declaring nothing (Pick phase)',
  'call <suit>      - Call partner suit 1=spade 2=club 3=heart (Call phase)',
  'p <idx>          - Play a card (Play phase)',
  'n/next           - Next trick',
  'nr/nextround     - Next round',
  'h/hint           - Show hint',
  'r/reset          - Reset game',
];
