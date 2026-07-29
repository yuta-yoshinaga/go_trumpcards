import type { pontoonApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type PontoonArgs = Parameters<typeof pontoonApi.exec>;

const VALID_COMMANDS = [
  'b',
  'bet',
  'deal',
  's',
  'stick',
  't',
  'twist',
  'buy',
  'sp',
  'split',
  'bt',
  'bs',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse an amount argument shared by bet and buy. */
function parseAmount(args: string[], usage: string): number | { error: string } {
  if (args.length === 0) return { error: usage };
  const n = Number(args[0]);
  if (!Number.isFinite(n) || n <= 0) return { error: usage };
  return n;
}

/** Parse a Pontoon CLI command into API exec arguments. */
export function parsePontoonCommand(input: string): CliParseResult<PontoonArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bet': {
      const amount = parseAmount(args, 'Usage: b <amount>');
      if (typeof amount !== 'number') return amount;
      return { args: ['bet', amount] };
    }
    case 'deal':
      return { args: ['deal'] };
    case 's':
    case 'stick':
      return { args: ['stick'] };
    case 't':
    case 'twist':
      return { args: ['twist'] };
    case 'buy': {
      const extra = parseAmount(args, 'Usage: buy <amount>');
      if (typeof extra !== 'number') return extra;
      return { args: ['buy', extra] };
    }
    case 'sp':
    case 'split':
      return { args: ['split'] };
    case 'bt':
      return { args: ['bankertwist'] };
    case 'bs':
      return { args: ['bankerstay'] };
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

/** Help text for Pontoon CLI mode. */
export const PONTOON_HELP: string[] = [
  'b <amount>      - Bet and deal (the banker uses deal instead)',
  'deal            - Deal the round you are banking',
  's/stick         - Stick (only on 15 or more)',
  't/twist         - Twist (draw one, stake unchanged)',
  'buy <amount>    - Buy (raise and draw; closed after a twist)',
  'sp/split        - Split two cards of equal rank',
  'bt              - Draw a card as the banker',
  'bs              - Stop drawing and settle as the banker',
  'log             - Show action log',
  'r/reset         - Next round',
];
