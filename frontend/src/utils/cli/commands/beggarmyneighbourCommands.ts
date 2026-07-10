import type { beggarmyneighbourApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type BeggarMyNeighbourArgs = Parameters<typeof beggarmyneighbourApi.exec>;

const VALID_COMMANDS = ['s', 'step', 'a', 'autoplay', 'l', 'log', 'r', 'reset', 'help', '?'];

/** Parse a Beggar-My-Neighbour CLI command into API exec arguments. */
export function parseBeggarMyNeighbourCommand(input: string): CliParseResult<BeggarMyNeighbourArgs> {
  const { cmd } = splitCommand(input);

  switch (cmd) {
    case 's':
    case 'step':
      return { args: ['step'] };
    case 'a':
    case 'autoplay':
      return { args: ['autoplay'] };
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

/** Help text for Beggar-My-Neighbour CLI mode. */
export const BEGGARMYNEIGHBOUR_HELP: string[] = [
  's/step     - Play next card',
  'a/autoplay - Auto play to end',
  'l/log      - Show action log',
  'r/reset    - Reset game',
];
