import type { easthavenApi } from '../../../api/gameApi';
import i18n from '../../../i18n';
import type { CliParseResult } from '../types';

type EasthavenArgs = Parameters<typeof easthavenApi.exec>;

/** Localized CLI help lines for Easthaven (resolved via the i18n instance). */
export function easthavenHelp(): string[] {
  return i18n.t('easthaven:cli.help', { returnObjects: true }) as string[];
}

/** Parse an Easthaven CLI command into API exec arguments. Error strings are localized. */
export function parseEasthavenCommand(input: string): CliParseResult<EasthavenArgs> {
  const parts = input.trim().split(/\s+/);
  const cmd = parts[0]?.toLowerCase();
  switch (cmd) {
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    case 'd':
    case 'deal':
      return { args: ['deal'] };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    case 'ac':
    case 'autocomplete':
      return { args: ['autocomplete'] };
    case 'u':
    case 'undo':
      return { args: ['undo'] };
    case 'm':
    case 'move': {
      // m t <col> f  → tableau to foundation
      if (parts.length === 4 && parts[1] === 't' && parts[3] === 'f') {
        const col = Number.parseInt(parts[2], 10);
        if (Number.isNaN(col)) return { error: i18n.t('easthaven:cli.error.invalidColumn') };
        return { args: ['move', { zone: 'tableau', col, cardIndex: -1 }, { zone: 'foundation' }] };
      }
      // m t <col> <idx> t <col>  → move a run by card index
      if (parts.length === 6 && parts[1] === 't' && parts[4] === 't') {
        const from = Number.parseInt(parts[2], 10);
        const idx = Number.parseInt(parts[3], 10);
        const to = Number.parseInt(parts[5], 10);
        if (Number.isNaN(from) || Number.isNaN(idx) || Number.isNaN(to)) {
          return { error: i18n.t('easthaven:cli.error.invalidArg') };
        }
        return {
          args: ['move', { zone: 'tableau', col: from, cardIndex: idx }, { zone: 'tableau', col: to }],
        };
      }
      // m <from> <to>  → tableau top card
      if (parts.length === 3) {
        const from = Number.parseInt(parts[1], 10);
        const to = Number.parseInt(parts[2], 10);
        if (Number.isNaN(from) || Number.isNaN(to)) return { error: i18n.t('easthaven:cli.error.invalidColumn') };
        return {
          args: ['move', { zone: 'tableau', col: from, cardIndex: -1 }, { zone: 'tableau', col: to }],
        };
      }
      return { error: i18n.t('easthaven:cli.error.usageMove') };
    }
    default:
      return { error: i18n.t('easthaven:cli.error.unknownCommand', { cmd }) };
  }
}
