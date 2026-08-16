import { describe, expect, it } from 'vitest';
import { BADUGI_HELP, parseBadugiCommand } from './badugiCommands';
import { BARBU_HELP, parseBarbuCommand } from './barbuCommands';
import { BID_WHIST_HELP, parseBidWhistCommand } from './bidwhistCommands';
import { BURA_HELP, parseBuraCommand } from './buraCommands';
import { BURRACO_HELP, parseBurracoCommand } from './burracoCommands';
import { CARIBBEANSTUD_HELP, parseCaribbeanstudCommand } from './caribbeanstudCommands';
import { CASINOHOLDEM_HELP, parseCasinoholdemCommand } from './casinoholdemCommands';
import { CASSINO_HELP, parseCassinoCommand } from './cassinoCommands';
import { CHINESEPOKER_HELP, parseChinesepokerCommand } from './chinesepokerCommands';
import { CHINESETEN_HELP, parseChineseTenCommand } from './chinesetenCommands';
import { CRAZYEIGHTS_HELP, parseCrazyeightsCommand } from './crazyeightsCommands';
import { CRIBBAGE_HELP, parseCribbageCommand } from './cribbageCommands';
import { DESMOCHE_HELP, parseDesmocheCommand } from './desmocheCommands';
import { DEUCE_TO_SEVEN_HELP, parseDeuceToSevenCommand } from './deuceToSevenCommands';
import { DURAK_HELP, parseDurakCommand } from './durakCommands';
import { ESCOBA_HELP, parseEscobaCommand } from './escobaCommands';
import { FIVE_HUNDRED_HELP, parseFiveHundredCommand } from './fivehundredCommands';
import { HANDANDFOOT_HELP, parseHandAndFootCommand } from './handandfootCommands';
import { HIGHCARDFLUSH_HELP, parseHighcardflushCommand } from './highcardflushCommands';
import { KAISER_HELP, parseKaiserCommand } from './kaiserCommands';
import { LAUGHANDLIEDOWN_HELP, parseLaughAndLieDownCommand } from './laughandliedownCommands';
import { LOBA_HELP, parseLobaCommand } from './lobaCommands';
import { MACAU_HELP, parseMacauCommand } from './macauCommands';
import { MUSHI_HELP, parseMushiCommand } from './mushiCommands';
import { NAINJAUNE_HELP, parseNainJauneCommand } from './nainjauneCommands';
import { OASISPOKER_HELP, parseOasispokerCommand } from './oasispokerCommands';
import { OPENFACECHINESE_HELP, parseOpenfacechineseCommand } from './openfacechineseCommands';
import { PAGEONE_HELP, parsePageoneCommand } from './pageoneCommands';
import { PAIGOW_HELP, parsePaigowCommand } from './paigowCommands';
import { POCH_HELP, parsePochCommand } from './pochCommands';
import { POKERSQUARES_HELP, parsePokerSquaresCommand } from './pokersquaresCommands';
import { POPEJOAN_HELP, parsePopeJoanCommand } from './popejoanCommands';
import { PRESIDENT_HELP, parsePresidentCommand } from './presidentCommands';
import { parseReddogCommand, REDDOG_HELP } from './reddogCommands';
import { parseRookCommand, ROOK_HELP } from './rookCommands';
import { parseScopaCommand, SCOPA_HELP } from './scopaCommands';
import { parseScoponeCommand, SCOPONE_HELP } from './scoponeCommands';
import { parseSevenBridgeCommand, SEVENBRIDGE_HELP } from './sevenBridgeCommands';
import { parseSixCardGolfCommand } from './sixcardgolfCommands';
import { parseSjavsCommand, SJAVS_HELP } from './sjavsCommands';
import { parseSkitgubbeCommand, SKITGUBBE_HELP } from './skitgubbeCommands';
import { parseThreecardCommand, THREECARD_HELP } from './threecardCommands';
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
  // #5792: these also needed their Web controller and API command union widened.
  ['badugi', parseBadugiCommand, BADUGI_HELP],
  ['barbu', parseBarbuCommand, BARBU_HELP],
  ['burraco', parseBurracoCommand, BURRACO_HELP],
  ['caribbeanstud', parseCaribbeanstudCommand, CARIBBEANSTUD_HELP],
  ['casinoholdem', parseCasinoholdemCommand, CASINOHOLDEM_HELP],
  ['cassino', parseCassinoCommand, CASSINO_HELP],
  ['chinesepoker', parseChinesepokerCommand, CHINESEPOKER_HELP],
  ['cribbage', parseCribbageCommand, CRIBBAGE_HELP],
  ['deuceToSeven', parseDeuceToSevenCommand, DEUCE_TO_SEVEN_HELP],
  ['escoba', parseEscobaCommand, ESCOBA_HELP],
  ['handandfoot', parseHandAndFootCommand, HANDANDFOOT_HELP],
  ['highcardflush', parseHighcardflushCommand, HIGHCARDFLUSH_HELP],
  ['kaiser', parseKaiserCommand, KAISER_HELP],
  ['macau', parseMacauCommand, MACAU_HELP],
  ['oasispoker', parseOasispokerCommand, OASISPOKER_HELP],
  ['pageone', parsePageoneCommand, PAGEONE_HELP],
  ['paigow', parsePaigowCommand, PAIGOW_HELP],
  ['president', parsePresidentCommand, PRESIDENT_HELP],
  ['reddog', parseReddogCommand, REDDOG_HELP],
  ['scopa', parseScopaCommand, SCOPA_HELP],
  ['scopone', parseScoponeCommand, SCOPONE_HELP],
  ['sevenBridge', parseSevenBridgeCommand, SEVENBRIDGE_HELP],
  ['threecard', parseThreecardCommand, THREECARD_HELP],
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

  // sixcardgolf builds its CLI help from i18n rather than exporting a *_HELP
  // array, so only its parser can be asserted here.
  it('sixcardgolf accepts "hint"', () => {
    for (const input of ['hint', 'h']) {
      expect(parseSixCardGolfCommand(input)).not.toHaveProperty('error');
    }
  });

  it('covers every module this change wired', () => {
    expect(WIRED).toHaveLength(44);
  });
});
