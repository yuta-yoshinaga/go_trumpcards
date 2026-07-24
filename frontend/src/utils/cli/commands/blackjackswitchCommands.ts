import type { blackjackswitchApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type BlackjackSwitchArgs = Parameters<typeof blackjackswitchApi.exec>;

const VALID_COMMANDS = [
  'b',
  'bet',
  'sw',
  'switch',
  'k',
  'keep',
  'hit',
  's',
  'stand',
  'dd',
  'double',
  'doubledown',
  'log',
  'l',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Blackjack Switch CLI command into API exec arguments. */
export function parseBlackjackSwitchCommand(input: string): CliParseResult<BlackjackSwitchArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bet': {
      const amount = parseIntArg(args, 0);
      if ('error' in amount) return { error: 'Usage: bet <amount>' };
      return { args: ['bet', amount.value] };
    }
    case 'sw':
    case 'switch':
      return { args: ['switch'] };
    case 'k':
    case 'keep':
      return { args: ['keep'] };
    case 'hit':
      return { args: ['hit'] };
    case 's':
    case 'stand':
      return { args: ['stand'] };
    case 'dd':
    case 'double':
    case 'doubledown':
      return { args: ['doubledown'] };
    case 'log':
    case 'l':
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

/** Help text for Blackjack Switch CLI mode. */
export const BLACKJACKSWITCH_HELP: string[] = [
  'bet <amt>   - Place the ante and deal',
  'sw/switch   - Switch the two hands’ second cards',
  'k/keep      - Keep the hands as dealt',
  'hit         - Draw a card for the current hand',
  's/stand     - Stand the current hand',
  'dd          - Double down the current hand',
  'log         - Show action log',
  'r/reset     - Reset the game',
];
