import type { speedApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type SpeedArgs = Parameters<typeof speedApi.exec>;

const VALID_COMMANDS = ['p', 'play', 'fl', 'flip', 'h', 'hint', 'log', 'r', 'reset', 'help', '?'];

/** Parse a Speed CLI command into API exec arguments. */
export function parseSpeedCommand(input: string): CliParseResult<SpeedArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      if (args.length < 2) return { error: 'Usage: p <cardIdx> <pileIdx>' };
      const cardIdx = parseIntArg(args, 0);
      if ('error' in cardIdx) return { error: 'Usage: p <cardIdx> <pileIdx>' };
      const pileIdx = parseIntArg(args, 1);
      if ('error' in pileIdx) return { error: 'Usage: p <cardIdx> <pileIdx>' };
      return { args: ['play', cardIdx.value, pileIdx.value] };
    }
    case 'fl':
    case 'flip':
      return { args: ['flip'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    case 'log':
      return { args: ['log'] };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    default:
      return unknownCommand(cmd, VALID_COMMANDS);
  }
}

/** Help text for Speed CLI mode. */
export const SPEED_HELP: string[] = [
  'p <card> <pile> - Play card to pile',
  'fl/flip     - Flip new center cards',
  'h/hint      - Show hint',
  'log         - Show action log',
  'r/reset     - Reset game',
];
