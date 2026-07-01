import type { looApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type LooArgs = Parameters<typeof looApi.exec>;

const VALID_COMMANDS = [
  'd',
  'decide',
  'play',
  'pass',
  'p',
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

/** Parse a Loo (Lanterloo) CLI command into API exec arguments. */
export function parseLooCommand(input: string): CliParseResult<LooArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'd':
    case 'decide': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed || parsed.value < 0 || parsed.value > 1) {
        return { error: 'Usage: decide <0-1> (0=pass 1=play)' };
      }
      return { args: ['decide', { play: parsed.value === 1 }] };
    }
    case 'pass':
      return { args: ['decide', { play: false }] };
    case 'play': {
      // `play` with an index plays a card; bare `play` is the play decision.
      if (args.length === 0) return { args: ['decide', { play: true }] };
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: play <idx>' };
      return { args: ['play', { cardIndex: parsed.value }] };
    }
    case 'p': {
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

/** Help text for Loo (Lanterloo) CLI mode. */
export const LOO_HELP: string[] = [
  'decide <0-1> - Decide participation: 0=pass 1=play (Decide phase)',
  'play / pass  - Play or pass in the Decide phase',
  'p <idx>      - Play a card (Play phase, must follow & head)',
  'nr/nextround - Next deal (settle the deal)',
  'h/hint       - Show hint',
  'r/reset      - Reset game',
];
