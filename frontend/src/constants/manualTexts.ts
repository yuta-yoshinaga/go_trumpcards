/**
 * Build-time imports of game manual Markdown files (Japanese, web version).
 * Keyed by game route path for direct lookup from the current URL.
 */
import accordion from '../../../docs/manual/web/accordion.md?raw';
import baccarat from '../../../docs/manual/web/baccarat.md?raw';
import badugi from '../../../docs/manual/web/badugi.md?raw';
import bakersdozen from '../../../docs/manual/web/bakersdozen.md?raw';
import beleagueredcastle from '../../../docs/manual/web/beleagueredcastle.md?raw';
import belote from '../../../docs/manual/web/belote.md?raw';
import blackjack from '../../../docs/manual/web/blackjack.md?raw';
import blackjackswitch from '../../../docs/manual/web/blackjackswitch.md?raw';
import bridge from '../../../docs/manual/web/bridge.md?raw';
import calculation from '../../../docs/manual/web/calculation.md?raw';
import canasta from '../../../docs/manual/web/canasta.md?raw';
import canfield from '../../../docs/manual/web/canfield.md?raw';
import caribbeanstud from '../../../docs/manual/web/caribbeanstud.md?raw';
import casinoholdem from '../../../docs/manual/web/casinoholdem.md?raw';
import casinowar from '../../../docs/manual/web/casinowar.md?raw';
import cassino from '../../../docs/manual/web/cassino.md?raw';
import clocksolitaire from '../../../docs/manual/web/clocksolitaire.md?raw';
import contractrummy from '../../../docs/manual/web/contractrummy.md?raw';
import crazyeights from '../../../docs/manual/web/crazyeights.md?raw';
import crazypineapple from '../../../docs/manual/web/crazypineapple.md?raw';
import crescent from '../../../docs/manual/web/crescent.md?raw';
import cribbage from '../../../docs/manual/web/cribbage.md?raw';
import daifugo from '../../../docs/manual/web/daifugo.md?raw';
import deuceswild from '../../../docs/manual/web/deuceswild.md?raw';
import doubt from '../../../docs/manual/web/doubt.md?raw';
import dragontiger from '../../../docs/manual/web/dragontiger.md?raw';
import durak from '../../../docs/manual/web/durak.md?raw';
import egyptianratscrew from '../../../docs/manual/web/egyptianratscrew.md?raw';
import euchre from '../../../docs/manual/web/euchre.md?raw';
import fiftyone from '../../../docs/manual/web/fiftyone.md?raw';
import fortythieves from '../../../docs/manual/web/fortythieves.md?raw';
import freecell from '../../../docs/manual/web/freecell.md?raw';
import ginrummy from '../../../docs/manual/web/ginrummy.md?raw';
import gofish from '../../../docs/manual/web/gofish.md?raw';
import golf from '../../../docs/manual/web/golf.md?raw';
import hearts from '../../../docs/manual/web/hearts.md?raw';
import holdem from '../../../docs/manual/web/holdem.md?raw';
import indianpoker from '../../../docs/manual/web/indianpoker.md?raw';
import jokerpoker from '../../../docs/manual/web/jokerpoker.md?raw';
import klondike from '../../../docs/manual/web/klondike.md?raw';
import letitride from '../../../docs/manual/web/letitride.md?raw';
import memory from '../../../docs/manual/web/memory.md?raw';
import mighty from '../../../docs/manual/web/mighty.md?raw';
import mississippistud from '../../../docs/manual/web/mississippistud.md?raw';
import montecarlo from '../../../docs/manual/web/montecarlo.md?raw';
import napoleon from '../../../docs/manual/web/napoleon.md?raw';
import nertz from '../../../docs/manual/web/nertz.md?raw';
import oasispoker from '../../../docs/manual/web/oasispoker.md?raw';
import ohhell from '../../../docs/manual/web/ohhell.md?raw';
import oldmaid from '../../../docs/manual/web/oldmaid.md?raw';
import omaha from '../../../docs/manual/web/omaha.md?raw';
import omahahilo from '../../../docs/manual/web/omahahilo.md?raw';
import pageone from '../../../docs/manual/web/pageone.md?raw';
import paigow from '../../../docs/manual/web/paigow.md?raw';
import pigtail from '../../../docs/manual/web/pigtail.md?raw';
import pineapple from '../../../docs/manual/web/pineapple.md?raw';
import pinochle from '../../../docs/manual/web/pinochle.md?raw';
import piquet from '../../../docs/manual/web/piquet.md?raw';
import pitch from '../../../docs/manual/web/pitch.md?raw';
import poker from '../../../docs/manual/web/poker.md?raw';
import pokersquares from '../../../docs/manual/web/pokersquares.md?raw';
import president from '../../../docs/manual/web/president.md?raw';
import pyramid from '../../../docs/manual/web/pyramid.md?raw';
import razz from '../../../docs/manual/web/razz.md?raw';
import reddog from '../../../docs/manual/web/reddog.md?raw';
import russiansolitaire from '../../../docs/manual/web/russiansolitaire.md?raw';
import scorpion from '../../../docs/manual/web/scorpion.md?raw';
import sevenbridge from '../../../docs/manual/web/sevenbridge.md?raw';
import sevencardstud from '../../../docs/manual/web/sevencardstud.md?raw';
import sevens from '../../../docs/manual/web/sevens.md?raw';
import shithead from '../../../docs/manual/web/shithead.md?raw';
import shortdeck from '../../../docs/manual/web/shortdeck.md?raw';
import skat from '../../../docs/manual/web/skat.md?raw';
import slapjack from '../../../docs/manual/web/slapjack.md?raw';
import spades from '../../../docs/manual/web/spades.md?raw';
import spanish21 from '../../../docs/manual/web/spanish21.md?raw';
import speed from '../../../docs/manual/web/speed.md?raw';
import spider from '../../../docs/manual/web/spider.md?raw';
import spiderette from '../../../docs/manual/web/spiderette.md?raw';
import spiteandmalice from '../../../docs/manual/web/spiteandmalice.md?raw';
import texasholdembonus from '../../../docs/manual/web/texasholdembonus.md?raw';
import threecard from '../../../docs/manual/web/threecard.md?raw';
import tonk from '../../../docs/manual/web/tonk.md?raw';
import trash from '../../../docs/manual/web/trash.md?raw';
import tripeaks from '../../../docs/manual/web/tripeaks.md?raw';
import twotenjack from '../../../docs/manual/web/twotenjack.md?raw';
import ultimatetexasholdem from '../../../docs/manual/web/ultimatetexasholdem.md?raw';
import videopoker from '../../../docs/manual/web/videopoker.md?raw';
import war from '../../../docs/manual/web/war.md?raw';
import whist from '../../../docs/manual/web/whist.md?raw';
import yukon from '../../../docs/manual/web/yukon.md?raw';

/** Map from game route path to raw Markdown manual text. */
export const manualTexts: Readonly<Record<string, string>> = {
  '/': blackjack,
  '/baccarat': baccarat,
  '/belote': belote,
  '/bridge': bridge,
  '/calculation': calculation,
  '/canasta': canasta,
  '/canfield': canfield,
  '/caribbeanstud': caribbeanstud,
  '/casinoholdem': casinoholdem,
  '/casinowar': casinowar,
  '/texasholdembonus': texasholdembonus,
  '/clocksolitaire': clocksolitaire,
  '/crazyeights': crazyeights,
  '/cribbage': cribbage,
  '/daifugo': daifugo,
  '/deuceswild': deuceswild,
  '/doubt': doubt,
  '/dragontiger': dragontiger,
  '/blackjackswitch': blackjackswitch,
  '/durak': durak,
  '/bakersdozen': bakersdozen,
  '/beleagueredcastle': beleagueredcastle,
  '/egyptianratscrew': egyptianratscrew,
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
  '/mighty': mighty,
  '/napoleon': napoleon,
  '/oasispoker': oasispoker,
  '/ohhell': ohhell,
  '/oldmaid': oldmaid,
  '/omaha': omaha,
  '/omahahilo': omahahilo,
  '/pageone': pageone,
  '/paigow': paigow,
  '/pineapple': pineapple,
  '/crazypineapple': crazypineapple,
  '/pigtail': pigtail,
  '/pinochle': pinochle,
  '/piquet': piquet,
  '/poker': poker,
  '/pokersquares': pokersquares,
  '/montecarlo': montecarlo,
  '/contractrummy': contractrummy,
  '/ultimatetexasholdem': ultimatetexasholdem,
  '/crescent': crescent,
  '/mississippistud': mississippistud,
  '/pyramid': pyramid,
  '/razz': razz,
  '/badugi': badugi,
  '/reddog': reddog,
  '/sevencardstud': sevencardstud,
  '/sevens': sevens,
  '/shortdeck': shortdeck,
  '/spades': spades,
  '/pitch': pitch,
  '/spanish21': spanish21,
  '/speed': speed,
  '/spider': spider,
  '/spiderette': spiderette,
  '/threecard': threecard,
  '/tonk': tonk,
  '/tripeaks': tripeaks,
  '/twotenjack': twotenjack,
  '/videopoker': videopoker,
  '/war': war,
  '/fiftyone': fiftyone,
  '/whist': whist,
  '/yukon': yukon,
  '/russiansolitaire': russiansolitaire,
  '/scorpion': scorpion,
  '/sevenbridge': sevenbridge,
  '/accordion': accordion,
  '/trash': trash,
  '/president': president,
  '/cassino': cassino,
  '/spiteandmalice': spiteandmalice,
  '/skat': skat,
  '/shithead': shithead,
  '/nertz': nertz,
  '/slapjack': slapjack,
};
