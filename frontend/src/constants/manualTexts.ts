/**
 * Build-time imports of game manual Markdown files (Japanese, web version).
 * Keyed by game route path for direct lookup from the current URL.
 */
import baccarat from '../../../docs/manual/web/baccarat.md?raw';
import blackjack from '../../../docs/manual/web/blackjack.md?raw';
import bridge from '../../../docs/manual/web/bridge.md?raw';
import canasta from '../../../docs/manual/web/canasta.md?raw';
import crazyeights from '../../../docs/manual/web/crazyeights.md?raw';
import cribbage from '../../../docs/manual/web/cribbage.md?raw';
import daifugo from '../../../docs/manual/web/daifugo.md?raw';
import deuceswild from '../../../docs/manual/web/deuceswild.md?raw';
import doubt from '../../../docs/manual/web/doubt.md?raw';
import euchre from '../../../docs/manual/web/euchre.md?raw';
import freecell from '../../../docs/manual/web/freecell.md?raw';
import ginrummy from '../../../docs/manual/web/ginrummy.md?raw';
import gofish from '../../../docs/manual/web/gofish.md?raw';
import golf from '../../../docs/manual/web/golf.md?raw';
import hearts from '../../../docs/manual/web/hearts.md?raw';
import holdem from '../../../docs/manual/web/holdem.md?raw';
import indianpoker from '../../../docs/manual/web/indianpoker.md?raw';
import jokerpoker from '../../../docs/manual/web/jokerpoker.md?raw';
import klondike from '../../../docs/manual/web/klondike.md?raw';
import memory from '../../../docs/manual/web/memory.md?raw';
import napoleon from '../../../docs/manual/web/napoleon.md?raw';
import ohhell from '../../../docs/manual/web/ohhell.md?raw';
import oldmaid from '../../../docs/manual/web/oldmaid.md?raw';
import omaha from '../../../docs/manual/web/omaha.md?raw';
import pineapple from '../../../docs/manual/web/pineapple.md?raw';
import pinochle from '../../../docs/manual/web/pinochle.md?raw';
import poker from '../../../docs/manual/web/poker.md?raw';
import pyramid from '../../../docs/manual/web/pyramid.md?raw';
import sevens from '../../../docs/manual/web/sevens.md?raw';
import shortdeck from '../../../docs/manual/web/shortdeck.md?raw';
import spades from '../../../docs/manual/web/spades.md?raw';
import speed from '../../../docs/manual/web/speed.md?raw';
import spider from '../../../docs/manual/web/spider.md?raw';
import threecard from '../../../docs/manual/web/threecard.md?raw';
import tripeaks from '../../../docs/manual/web/tripeaks.md?raw';
import videopoker from '../../../docs/manual/web/videopoker.md?raw';

/** Map from game route path to raw Markdown manual text. */
export const manualTexts: Readonly<Record<string, string>> = {
  '/': blackjack,
  '/baccarat': baccarat,
  '/bridge': bridge,
  '/canasta': canasta,
  '/crazyeights': crazyeights,
  '/cribbage': cribbage,
  '/daifugo': daifugo,
  '/deuceswild': deuceswild,
  '/doubt': doubt,
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
  '/memory': memory,
  '/napoleon': napoleon,
  '/ohhell': ohhell,
  '/oldmaid': oldmaid,
  '/omaha': omaha,
  '/pineapple': pineapple,
  '/pinochle': pinochle,
  '/poker': poker,
  '/pyramid': pyramid,
  '/sevens': sevens,
  '/shortdeck': shortdeck,
  '/spades': spades,
  '/speed': speed,
  '/spider': spider,
  '/threecard': threecard,
  '/tripeaks': tripeaks,
  '/videopoker': videopoker,
};
