import type { calabresellaApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type CalabresellaArgs = Parameters<typeof calabresellaApi.exec>;

const VALID_COMMANDS = [
  'b',
  'bid',
  'd',
  'discard',
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

/** Parse a Calabresella (Terziglio) CLI command into API exec arguments. */
export function parseCalabresellaCommand(input: string): CliParseResult<CalabresellaArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bid': {
      const arg = args[0]?.toLowerCase();
      if (arg === 'pass' || arg === 'p') return { args: ['bid', { bid: 0 }] };
      if (arg === 'chiamo' || arg === 'c') return { args: ['bid', { bid: 1 }] };
      if (arg === 'solo' || arg === 's') return { args: ['bid', { bid: 2 }] };
      return { error: 'Usage: bid pass|chiamo|solo' };
    }
    case 'd':
    case 'discard': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: d <idx>' };
      return { args: ['discard', { cardIndex: parsed.value }] };
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

/** Help text for Calabresella (Terziglio) CLI mode. */
export const CALABRESELLA_HELP: string[] = [
  'bid pass|chiamo|solo - Pass, or declare chiamo (stake 1) / solo (stake 2) (Bid phase)',
  'd <idx>              - Discard a card (monte exchange, four times as Soloist)',
  'p <idx>              - Play a card (Play phase, must follow the led suit)',
  'n/next               - Next trick',
  'nr/nextround         - Next round',
  'h/hint               - Show hint',
  'r/reset              - Reset game',
];
