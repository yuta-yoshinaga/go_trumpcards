import type { pitchApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type PitchArgs = Parameters<typeof pitchApi.exec>;

const VALID_COMMANDS = [
  'b',
  'bid',
  'pass',
  'p',
  'play',
  'n',
  'next',
  'nr',
  'nextround',
  'h',
  'hint',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Pitch CLI command into API exec arguments. */
export function parsePitchCommand(input: string): CliParseResult<PitchArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bid': {
      const v = parseIntArg(args, 0);
      if ('error' in v) return { error: 'Usage: bid <0|2|3|4>' };
      return { args: ['bid', v.value] };
    }
    case 'pass':
      return { args: ['bid', 0] };
    case 'p':
    case 'play': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: play <index>' };
      return { args: ['play', undefined, idx.value] };
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
    case 'log':
      return { args: ['hint'] }; // log share-displays via hint endpoint when no log endpoint exists
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

/** Help text for Pitch CLI mode. */
export const PITCH_HELP: string[] = [
  'b <n>        - Bid (0=pass, 2-4)',
  'pass         - Pass',
  'p <i>        - Play card by index',
  'n / next     - Next trick',
  'nr           - Next round (score current round first)',
  'h / hint     - Show hint',
  'r / reset    - Reset game',
];
