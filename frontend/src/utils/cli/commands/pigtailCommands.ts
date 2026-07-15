import type { pigtailApi } from '../../../api/gameApi';
import i18n from '../../../i18n';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type PigtailArgs = Parameters<typeof pigtailApi.exec>;

const VALID_COMMANDS = ['draw', 'reset'];

/** Localized CLI help lines for Pig's Tail (resolved via the i18n instance). */
export function pigtailHelp(): string[] {
  return i18n.t('pigtail:cli.help', { returnObjects: true }) as string[];
}

/** Parse a Pig's Tail CLI command into API exec arguments. Error strings are localized. */
export function parsePigtailCommand(input: string): CliParseResult<PigtailArgs> {
  const { cmd } = splitCommand(input);
  switch (cmd) {
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    case 'd':
    case 'draw':
      return { args: ['draw'] };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: i18n.t('pigtail:cli.error.unknownSuggest', { cmd, suggestion }) };
      return { error: i18n.t('pigtail:cli.error.unknown', { cmd }) };
    }
  }
}
