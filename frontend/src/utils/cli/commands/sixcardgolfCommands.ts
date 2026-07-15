import type { sixcardgolfApi } from '../../../api/gameApi';
import i18n from '../../../i18n';
import type { CliParseResult } from '../types';

type SixCardGolfArgs = Parameters<typeof sixcardgolfApi.exec>;

/** Parse a Six Card Golf CLI command into API exec arguments. Error strings are localized. */
export function parseSixCardGolfCommand(input: string): CliParseResult<SixCardGolfArgs> {
  // Module-level parser → use the i18n instance directly (frontend/CLAUDE.md convention).
  const usage = (cmd: string) => ({ error: i18n.t('sixcardgolf:cliUsagePos', { cmd }) });
  const parts = input.trim().split(/\s+/);
  const cmd = parts[0]?.toLowerCase();
  const posArg = parts[1] ? Number.parseInt(parts[1], 10) : undefined;
  switch (cmd) {
    case 'r':
    case 'reset':
      return { args: [{ command: 'reset' }] };
    case 'fi':
    case 'flipinitial':
      if (posArg === undefined || Number.isNaN(posArg)) return usage('fi');
      return { args: [{ command: 'flipinitial', position: posArg }] };
    case 'ds':
    case 'drawstock':
      return { args: [{ command: 'drawstock' }] };
    case 'dd':
    case 'drawdiscard':
      return { args: [{ command: 'drawdiscard' }] };
    case 'sw':
    case 'swap':
      if (posArg === undefined || Number.isNaN(posArg)) return usage('sw');
      return { args: [{ command: 'swap', position: posArg }] };
    case 'di':
    case 'discard':
      return { args: [{ command: 'discard' }] };
    case 'fl':
    case 'flip':
      if (posArg === undefined || Number.isNaN(posArg)) return usage('fl');
      return { args: [{ command: 'flip', position: posArg }] };
    case 'sf':
    case 'skipflip':
      return { args: [{ command: 'skipflip' }] };
    case 'nr':
    case 'nextround':
      return { args: [{ command: 'nextround' }] };
    case 'l':
    case 'log':
      return { args: [{ command: 'log' }] };
    default:
      return { error: i18n.t('sixcardgolf:cliUnknown', { cmd: cmd ?? '' }) };
  }
}
