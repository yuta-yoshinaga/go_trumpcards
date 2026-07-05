import type { machiavelliApi } from '../../../api/gameApi';
import { parseIntArg, parseIntSlice, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type MachiavelliArgs = Parameters<typeof machiavelliApi.exec>;

const VALID_COMMANDS = [
  'dr',
  'draw',
  'nm',
  'newmeld',
  'lo',
  'layoff',
  'nr',
  'nextround',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Machiavelli CLI command into API exec arguments. */
export function parseMachiavelliCommand(input: string): CliParseResult<MachiavelliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'dr':
    case 'draw':
      return { args: ['draw'] };
    case 'nm':
    case 'newmeld': {
      const parsed = parseIntSlice(args);
      if ('error' in parsed) return { error: 'Usage: nm <i j k...> (3+ hand indices)' };
      if (parsed.values.length < 3) return { error: 'Usage: nm <i j k...> (a meld needs 3+ cards)' };
      return { args: ['newmeld', { handIndices: parsed.values }] };
    }
    case 'lo':
    case 'layoff': {
      const meld = parseIntArg(args, 0);
      const hand = parseIntArg(args, 1);
      if ('error' in meld || 'error' in hand) return { error: 'Usage: lo <meldIdx handIdx>' };
      return { args: ['layoff', { meldIdx: meld.value, handIndex: hand.value }] };
    }
    case 'nr':
    case 'nextround':
      return { args: ['nextround'] };
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

/** Help text for Machiavelli CLI mode. */
export const MACHIAVELLI_HELP: string[] = [
  'dr/draw          - Draw one card from stock (ends turn)',
  'nm <i j k...>    - Form a new meld from your hand',
  'lo <meld hand>   - Lay off a hand card onto a table meld',
  'nr/nextround     - Next round',
  'log              - Show action log',
  'r/reset          - Reset game',
];
