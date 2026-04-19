/**
 * Build-time imports of game manual Markdown files (Japanese, CUI version).
 * Keyed by game route path for direct lookup from the current URL.
 * Used when the page has CLI mode enabled so the manual matches the terminal UI.
 */
import accordion from '../../../docs/manual/cui/accordion.md?raw';
import baccarat from '../../../docs/manual/cui/baccarat.md?raw';
import badugi from '../../../docs/manual/cui/badugi.md?raw';
import blackjack from '../../../docs/manual/cui/blackjack.md?raw';
import bridge from '../../../docs/manual/cui/bridge.md?raw';
import canasta from '../../../docs/manual/cui/canasta.md?raw';
import canfield from '../../../docs/manual/cui/canfield.md?raw';
import caribbeanstud from '../../../docs/manual/cui/caribbeanstud.md?raw';
import clocksolitaire from '../../../docs/manual/cui/clocksolitaire.md?raw';
import crazyeights from '../../../docs/manual/cui/crazyeights.md?raw';
import cribbage from '../../../docs/manual/cui/cribbage.md?raw';
import daifugo from '../../../docs/manual/cui/daifugo.md?raw';
import deuceswild from '../../../docs/manual/cui/deuceswild.md?raw';
import doubt from '../../../docs/manual/cui/doubt.md?raw';
import durak from '../../../docs/manual/cui/durak.md?raw';
import euchre from '../../../docs/manual/cui/euchre.md?raw';
import fiftyone from '../../../docs/manual/cui/fiftyone.md?raw';
import fortythieves from '../../../docs/manual/cui/fortythieves.md?raw';
import freecell from '../../../docs/manual/cui/freecell.md?raw';
import ginrummy from '../../../docs/manual/cui/ginrummy.md?raw';
import gofish from '../../../docs/manual/cui/gofish.md?raw';
import golf from '../../../docs/manual/cui/golf.md?raw';
import hearts from '../../../docs/manual/cui/hearts.md?raw';
import holdem from '../../../docs/manual/cui/holdem.md?raw';
import indianpoker from '../../../docs/manual/cui/indianpoker.md?raw';
import jokerpoker from '../../../docs/manual/cui/jokerpoker.md?raw';
import klondike from '../../../docs/manual/cui/klondike.md?raw';
import letitride from '../../../docs/manual/cui/letitride.md?raw';
import memory from '../../../docs/manual/cui/memory.md?raw';
import napoleon from '../../../docs/manual/cui/napoleon.md?raw';
import ohhell from '../../../docs/manual/cui/ohhell.md?raw';
import oldmaid from '../../../docs/manual/cui/oldmaid.md?raw';
import omaha from '../../../docs/manual/cui/omaha.md?raw';
import pageone from '../../../docs/manual/cui/pageone.md?raw';
import paigow from '../../../docs/manual/cui/paigow.md?raw';
import pigtail from '../../../docs/manual/cui/pigtail.md?raw';
import pineapple from '../../../docs/manual/cui/pineapple.md?raw';
import pinochle from '../../../docs/manual/cui/pinochle.md?raw';
import poker from '../../../docs/manual/cui/poker.md?raw';
import pokersquares from '../../../docs/manual/cui/pokersquares.md?raw';
import pyramid from '../../../docs/manual/cui/pyramid.md?raw';
import razz from '../../../docs/manual/cui/razz.md?raw';
import reddog from '../../../docs/manual/cui/reddog.md?raw';
import scorpion from '../../../docs/manual/cui/scorpion.md?raw';
import sevencardstud from '../../../docs/manual/cui/sevencardstud.md?raw';
import sevens from '../../../docs/manual/cui/sevens.md?raw';
import shortdeck from '../../../docs/manual/cui/shortdeck.md?raw';
import spades from '../../../docs/manual/cui/spades.md?raw';
import speed from '../../../docs/manual/cui/speed.md?raw';
import spider from '../../../docs/manual/cui/spider.md?raw';
import threecard from '../../../docs/manual/cui/threecard.md?raw';
import tripeaks from '../../../docs/manual/cui/tripeaks.md?raw';
import twotenjack from '../../../docs/manual/cui/twotenjack.md?raw';
import videopoker from '../../../docs/manual/cui/videopoker.md?raw';
import war from '../../../docs/manual/cui/war.md?raw';
import whist from '../../../docs/manual/cui/whist.md?raw';
import yukon from '../../../docs/manual/cui/yukon.md?raw';

/** Map from game route path to raw CUI Markdown manual text. */
export const cuiManualTexts: Readonly<Record<string, string>> = {
  '/': blackjack,
  '/baccarat': baccarat,
  '/bridge': bridge,
  '/canasta': canasta,
  '/canfield': canfield,
  '/caribbeanstud': caribbeanstud,
  '/clocksolitaire': clocksolitaire,
  '/crazyeights': crazyeights,
  '/cribbage': cribbage,
  '/daifugo': daifugo,
  '/deuceswild': deuceswild,
  '/doubt': doubt,
  '/durak': durak,
  '/fiftyone': fiftyone,
  '/fortythieves': fortythieves,
  '/euchre': euchre,
  '/freecell': freecell,
  '/ginrummy': ginrummy,
  '/gofish': gofish,
  '/golf': golf,
  '/hearts': hearts,
  '/holdem': holdem,
  '/indianpoker': indianpoker,
  '/jokerpoker': jokerpoker,
  '/klondike': klondike,
  '/letitride': letitride,
  '/memory': memory,
  '/napoleon': napoleon,
  '/ohhell': ohhell,
  '/oldmaid': oldmaid,
  '/omaha': omaha,
  '/pageone': pageone,
  '/paigow': paigow,
  '/pineapple': pineapple,
  '/pigtail': pigtail,
  '/pinochle': pinochle,
  '/poker': poker,
  '/poker-squares': pokersquares,
  '/pyramid': pyramid,
  '/razz': razz,
  '/badugi': badugi,
  '/reddog': reddog,
  '/sevencardstud': sevencardstud,
  '/sevens': sevens,
  '/shortdeck': shortdeck,
  '/spades': spades,
  '/speed': speed,
  '/spider': spider,
  '/threecard': threecard,
  '/tripeaks': tripeaks,
  '/twotenjack': twotenjack,
  '/videopoker': videopoker,
  '/war': war,
  '/whist': whist,
  '/yukon': yukon,
  '/scorpion': scorpion,
  '/accordion': accordion,
};

/** Returns true when CLI mode is enabled for the game at the given path. */
export function isCliModeEnabled(gamePath: string): boolean {
  const gameName = gamePath === '/' ? 'blackjack' : gamePath.slice(1);
  try {
    return localStorage.getItem(`cli-mode-${gameName}`) === 'true';
  } catch {
    return false;
  }
}
