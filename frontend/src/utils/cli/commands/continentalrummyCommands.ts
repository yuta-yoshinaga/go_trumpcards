import type { continentalrummyApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type ContinentalRummyArgs = Parameters<typeof continentalrummyApi.exec>;

const VALID_COMMANDS = [
  'stock',
  'ds',
  'take',
  'dd',
  'discard',
  'd',
  'goout',
  'g',
  'gooutdeal',
  'gd',
  'next',
  'n',
  'hint',
  'h',
  'log',
  'l',
  'r',
  'reset',
  'help',
  '?',
];

const DISCARD_USAGE = 'Usage: discard <index>';
const GOOUT_USAGE = 'Usage: goout <index> (the card you throw)';

/** CLI help text for Continental Rummy. */
export const CONTINENTALRUMMY_CLI_HELP = [
  'stock (ds)       draw from the stock',
  'take (dd)        lift the top discard',
  'discard <idx>    throw the card at idx',
  'goout <idx>      lay all fifteen down, throwing the card at idx',
  'gooutdeal (gd)  lay the dealt fifteen down without drawing (worth more)',
  'next (n)         go to the next round',
  'hint (h)         show a hint',
  'log (l)          show the action log',
  'reset (r)        restart',
];

/** Parse a Continental Rummy CLI command into API exec arguments. */
export function parseContinentalRummyCommand(input: string): CliParseResult<ContinentalRummyArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    // **山と捨て札は別の命令。** 引数ひとつに畳むと、付け忘れた要求が
    // 黙ってどちらかに倒れる。
    case 'ds':
    case 'stock':
      return { args: ['stock'] };
    case 'dd':
    case 'take':
      return { args: ['take'] };
    case 'd':
    case 'discard': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: DISCARD_USAGE };
      return { args: ['discard', { handIndex: idx.value }] };
    }
    // **上がるときも捨てる 1 枚を名指す。** 15 枚を並べて 1 枚を手放すので、
    // どれを手放すかを決めずには上がれない。
    case 'g':
    case 'goout': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: GOOUT_USAGE };
      return { args: ['goout', { handIndex: idx.value }] };
    }
    // **引かずに上がるほうは札を捨てない。** 加点が違うので別の命令。
    case 'gd':
    case 'gooutdeal':
      return { args: ['gooutdeal'] };
    case 'n':
    case 'next':
      return { args: ['next'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    case 'l':
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
