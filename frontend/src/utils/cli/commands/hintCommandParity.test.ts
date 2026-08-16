import { describe, expect, it } from 'vitest';
import { BID_WHIST_HELP, parseBidWhistCommand } from './bidwhistCommands';
import { BURA_HELP, parseBuraCommand } from './buraCommands';
import { CHINESETEN_HELP, parseChineseTenCommand } from './chinesetenCommands';
import { CRAZYEIGHTS_HELP, parseCrazyeightsCommand } from './crazyeightsCommands';
import { DESMOCHE_HELP, parseDesmocheCommand } from './desmocheCommands';
import { DURAK_HELP, parseDurakCommand } from './durakCommands';
import { FIVE_HUNDRED_HELP, parseFiveHundredCommand } from './fivehundredCommands';
import { LAUGHANDLIEDOWN_HELP, parseLaughAndLieDownCommand } from './laughandliedownCommands';
import { LOBA_HELP, parseLobaCommand } from './lobaCommands';
import { MUSHI_HELP, parseMushiCommand } from './mushiCommands';
import { NAINJAUNE_HELP, parseNainJauneCommand } from './nainjauneCommands';
import { OPENFACECHINESE_HELP, parseOpenfacechineseCommand } from './openfacechineseCommands';
import { POCH_HELP, parsePochCommand } from './pochCommands';
import { POKERSQUARES_HELP, parsePokerSquaresCommand } from './pokersquaresCommands';
import { POPEJOAN_HELP, parsePopeJoanCommand } from './popejoanCommands';
import { parseRookCommand, ROOK_HELP } from './rookCommands';
import { parseSjavsCommand, SJAVS_HELP } from './sjavsCommands';
import { parseSkitgubbeCommand, SKITGUBBE_HELP } from './skitgubbeCommands';
import { parseToepenCommand, TOEPEN_HELP } from './toepenCommands';
import { parseTrexCommand, TREX_HELP } from './trexCommands';
import { parseZwickerCommand, ZWICKER_HELP } from './zwickerCommands';

/**
 * Games whose backend already answers a `hint` action and whose API command union
 * already permits it, but whose Web CLI never exposed the command -- so switching to
 * CLI mode silently lost a feature the GUI had.
 *
 * Each entry pins both halves: the parser accepts the command, and `help` advertises
 * it. Wiring one without the other leaves the feature undiscoverable.
 */
// The parsers return game-specific argument tuples; widening the return to
// `unknown[]` keeps them all assignable without erasing the error branch.
type AnyParser = (input: string) => { args: unknown[] } | { error: string };

const WIRED: readonly [string, AnyParser, readonly string[]][] = [
  ['bidwhist', parseBidWhistCommand, BID_WHIST_HELP],
  ['bura', parseBuraCommand, BURA_HELP],
  ['chineseten', parseChineseTenCommand, CHINESETEN_HELP],
  ['crazyeights', parseCrazyeightsCommand, CRAZYEIGHTS_HELP],
  ['desmoche', parseDesmocheCommand, DESMOCHE_HELP],
  ['durak', parseDurakCommand, DURAK_HELP],
  ['fivehundred', parseFiveHundredCommand, FIVE_HUNDRED_HELP],
  ['laughandliedown', parseLaughAndLieDownCommand, LAUGHANDLIEDOWN_HELP],
  ['loba', parseLobaCommand, LOBA_HELP],
  ['mushi', parseMushiCommand, MUSHI_HELP],
  ['nainjaune', parseNainJauneCommand, NAINJAUNE_HELP],
  ['openfacechinese', parseOpenfacechineseCommand, OPENFACECHINESE_HELP],
  ['poch', parsePochCommand, POCH_HELP],
  ['pokersquares', parsePokerSquaresCommand, POKERSQUARES_HELP],
  ['popejoan', parsePopeJoanCommand, POPEJOAN_HELP],
  ['rook', parseRookCommand, ROOK_HELP],
  ['sjavs', parseSjavsCommand, SJAVS_HELP],
  ['skitgubbe', parseSkitgubbeCommand, SKITGUBBE_HELP],
  ['toepen', parseToepenCommand, TOEPEN_HELP],
  ['trex', parseTrexCommand, TREX_HELP],
  ['zwicker', parseZwickerCommand, ZWICKER_HELP],
];

describe('hint command parity between the GUI and the Web CLI', () => {
  it.each(WIRED)('%s accepts "hint" and lists it in help', (_name, parse, help) => {
    for (const input of ['hint', 'h']) {
      const result = parse(input);
      expect(result, `"${input}" was rejected`).not.toHaveProperty('error');
      expect(result).toEqual({ args: ['hint'] });
    }
    expect(
      help.some((line) => line.includes('hint')),
      'help text never mentions hint, so the command is undiscoverable',
    ).toBe(true);
  });

  it('covers every module this change wired', () => {
    expect(WIRED).toHaveLength(21);
  });
});
