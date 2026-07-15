import type { slapjackApi } from '../../../api/gameApi';
import i18n from '../../../i18n';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type SlapjackArgs = Parameters<typeof slapjackApi.exec>;

const VALID_COMMANDS = ['step', 'slap', 'tick', 'reset', 'log'];

/** Localized CLI help lines for Slap Jack (resolved via the i18n instance). */
export function slapjackHelp(): string[] {
  return i18n.t('slapjack:cli.help', { returnObjects: true }) as string[];
}

/** Parse a Slap Jack CLI command into API exec arguments. Error strings are localized. */
export function parseSlapjackCommand(input: string): CliParseResult<SlapjackArgs> {
  const { cmd } = splitCommand(input);
  switch (cmd) {
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    case 's':
    case 'step':
      return { args: ['step'] };
    case 'j':
    case 'slap':
      return { args: ['slap'] };
    case 'tick':
      return { args: ['tick'] };
    case 'l':
    case 'log':
      return { args: ['log'] };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: i18n.t('slapjack:cli.error.unknownSuggest', { cmd, suggestion }) };
      return { error: i18n.t('slapjack:cli.error.unknown', { cmd }) };
    }
  }
}
