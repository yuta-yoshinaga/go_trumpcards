import type { cruelApi } from '../../../api/gameApi';
import i18n from '../../../i18n';
import type { CliParseResult } from '../types';

type CruelArgs = Parameters<typeof cruelApi.exec>;

/** Localized CLI help lines for Cruel (resolved via the i18n instance). */
export function cruelHelp(): string[] {
  return i18n.t('cruel:cli.help', { returnObjects: true }) as string[];
}

/** Parse a Cruel CLI command into API exec arguments. Error strings are localized. */
export function parseCruelCommand(input: string): CliParseResult<CruelArgs> {
  const parts = input.trim().split(/\s+/);
  const cmd = parts[0]?.toLowerCase();
  switch (cmd) {
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    case 's':
    case 'shift':
      return { args: ['shift'] };
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
      if (parts.length === 3) {
        const from = Number.parseInt(parts[1], 10);
        if (Number.isNaN(from)) return { error: i18n.t('cruel:cli.error.invalidSource') };
        if (parts[2] === 'f') {
          return { args: ['move', { zone: 'tableau', col: from }, { zone: 'foundation' }] };
        }
        const to = Number.parseInt(parts[2], 10);
        if (Number.isNaN(to)) return { error: i18n.t('cruel:cli.error.invalidDestination') };
        return { args: ['move', { zone: 'tableau', col: from }, { zone: 'tableau', col: to }] };
      }
      return { error: i18n.t('cruel:cli.error.usageMove') };
    }
    default:
      return { error: i18n.t('cruel:cli.error.unknownCommand', { cmd }) };
  }
}
