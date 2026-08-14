import type { doubleattackApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type DoubleAttackArgs = Parameters<typeof doubleattackApi.exec>;

const VALID_COMMANDS = [
  'bet',
  'b',
  'attack',
  'a',
  'hit',
  'h',
  'stand',
  's',
  'double',
  'd',
  'split',
  'sp',
  'next',
  'hint',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

const BET_USAGE = 'Usage: bet <ante> [bustIt]';
const ATTACK_USAGE = 'Usage: attack <amount> (0 declines)';

/** CLI help text for Extra Bet Blackjack. */
export const DOUBLEATTACK_CLI_HELP = [
  'bet <ante> [bi]  place the ante (bi = Bust It side bet, optional)',
  'attack <amount>  raise after seeing the up-card; 0 declines',
  'hit (h)          take a card',
  'stand (s)        stand pat',
  'double (d)       double and take exactly one card',
  'split (sp)       split a pair',
  'next             deal the next hand',
  'hint             show a hint',
  'log              show the action log',
  'reset (r)        restart',
];

/** Parse an Extra Bet Blackjack CLI command into API exec arguments. */
export function parseDoubleAttackCommand(input: string): CliParseResult<DoubleAttackArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bet': {
      const ante = parseIntArg(args, 0);
      if ('error' in ante) return { error: BET_USAGE };
      if (args.length < 2) return { args: ['bet', { ante: ante.value, bustIt: 0 }] };
      const bi = parseIntArg(args.slice(1), 0);
      if ('error' in bi) return { error: BET_USAGE };
      return { args: ['bet', { ante: ante.value, bustIt: bi.value }] };
    }
    case 'a':
    case 'attack': {
      // **0 は「見送り」で、正当な入力。** 上限はサーバが持つのでここでは書かない。
      const amount = parseIntArg(args, 0);
      if ('error' in amount) return { error: ATTACK_USAGE };
      return { args: ['attack', { amount: amount.value }] };
    }
    case 'h':
    case 'hit':
      return { args: ['hit'] };
    case 's':
    case 'stand':
      return { args: ['stand'] };
    case 'd':
    case 'double':
      return { args: ['double'] };
    case 'sp':
    case 'split':
      return { args: ['split'] };
    case 'next':
      return { args: ['next'] };
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
