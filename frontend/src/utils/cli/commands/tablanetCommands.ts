import type { tablanetApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type TablanetArgs = Parameters<typeof tablanetApi.exec>;

const VALID_COMMANDS = ['p', 'play', 'n', 'next', 'nr', 'nextround', 'h', 'hint', 'r', 'reset', 'help', '?'];

/** Parse a Tablanet (Tablić) CLI command into API exec arguments. */
export function parseTablanetCommand(input: string): CliParseResult<TablanetArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: p <handIdx> [tableIdx...]' };
      const tableIndices: number[] = [];
      for (const raw of args.slice(1)) {
        const n = Number.parseInt(raw, 10);
        if (Number.isNaN(n)) return { error: `Invalid table index: ${raw}` };
        tableIndices.push(n);
      }
      return { args: ['play', { cardIndex: parsed.value, tableIndices }] };
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

/** Help text for Tablanet (Tablić) CLI mode. */
export const TABLANET_HELP: string[] = [
  'p <h> [t...] - Play hand card h, capturing table cards t... (omit t to trail; a Jack sweeps)',
  'nr/nextround - Start a new game',
  'h/hint       - Show hint',
  'r/reset      - Reset game',
];
