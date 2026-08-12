// This file is now a barrel. Its 165 api objects and 213 game-specific types
// moved to ./games/<game>.ts, and the shared plumbing to ./gameExec.ts
// (issue #4434). 810 files import from here, so every one still resolves.
//
// Prefer importing from ./games/<game> in new code; this stays as the
// compatibility layer for the existing imports.

import type { ActionLogResponse } from '../types/card';
import { gameExec } from './gameExec';

export * from './gameExec';
export * from './games/accordion';
export * from './games/acesup';
export * from './games/agnes';
export * from './games/allfours';
export * from './games/aluette';
export * from './games/americantoad';
export * from './games/anaconda';
export * from './games/andarbahar';
export * from './games/auldlangsyne';
export * from './games/baccarat';
export * from './games/badugi';
export * from './games/bakersdozen';
export * from './games/bakersgame';
export * from './games/baloot';
export * from './games/barbu';
export * from './games/basra';
export * from './games/beggarmyneighbour';
export * from './games/beleagueredcastle';
export * from './games/belote';
export * from './games/bezique';
export * from './games/bhabhi';
export * from './games/bideuchre';
export * from './games/bidwhist';
export * from './games/bigo';
export * from './games/bigohilo';
export * from './games/bigtwo';
export * from './games/bisley';
export * from './games/blackhole';
export * from './games/blackjack';
export * from './games/blackjackswitch';
export * from './games/boston';
export * from './games/botifarra';
export * from './games/bouillotte';
export * from './games/bourre';
export * from './games/braid';
export * from './games/bridge';
export * from './games/briscola';
export * from './games/bristol';
export * from './games/bura';
export * from './games/burraco';
export * from './games/calabresella';
export * from './games/calculation';
export * from './games/callbreak';
export * from './games/canasta';
export * from './games/canfield';
export * from './games/caribbeanstud';
export * from './games/carioca';
export * from './games/casinoholdem';
export * from './games/casinowar';
export * from './games/cassino';
export * from './games/catchten';
export * from './games/cego';
export * from './games/chemindefer';
export * from './games/chinchon';
export * from './games/chinesepoker';
export * from './games/chineseten';
export * from './games/cinch';
export * from './games/clocksolitaire';
export * from './games/colorado';
export * from './games/colourwhist';
export * from './games/congress';
export * from './games/conquian';
export * from './games/contractrummy';
export * from './games/courtpiece';
export * from './games/crazyeights';
export * from './games/crazypineapple';
export * from './games/crazyquilt';
export * from './games/crescent';
export * from './games/cribbage';
export * from './games/cribbagesquares';
export * from './games/cruel';
export * from './games/cuarenta';
export * from './games/cuckoo';
export * from './games/cucumber';
export * from './games/daifugo';
export * from './games/desmoche';
export * from './games/deuceswild';
export * from './games/deucetoseven';
export * from './games/diplomat';
export * from './games/doppelkopf';
export * from './games/doubleklondike';
export * from './games/doubt';
export * from './games/doudizhu';
export * from './games/dragontiger';
export * from './games/duchess';
export * from './games/durak';
export * from './games/easthaven';
export * from './games/ecarte';
export * from './games/egyptianratscrew';
export * from './games/eightoff';
export * from './games/escoba';
export * from './games/estimation';
export * from './games/euchre';
export * from './games/faro';
export * from './games/fiftyone';
export * from './games/fivecardstud';
export * from './games/fivehundred';
export * from './games/flowergarden';
export * from './games/fortyandeight';
export * from './games/fortyfives';
export * from './games/fortythieves';
export * from './games/fourcardpoker';
export * from './games/fourseasons';
export * from './games/freecell';
export * from './games/frenchtarot';
export * from './games/gaigel';
export * from './games/ganjifa';
export * from './games/gaps';
export * from './games/germanwhist';
export * from './games/ginrummy';
export * from './games/gofish';
export * from './games/golf';
export * from './games/gongzhu';
export * from './games/goofspiel';
export * from './games/gostop';
export * from './games/grandfathersclock';
export * from './games/guandan';
export * from './games/guts';
export * from './games/hachihachi';
export * from './games/handandfoot';
export * from './games/hasenpfeffer';
export * from './games/hearts';
export * from './games/highcardflush';
export * from './games/hokm';
export * from './games/holdem';
export * from './games/honeymoonbridge';
export * from './games/indianpoker';
export * from './games/indianrummy';
export * from './games/irishpoker';
export * from './games/israeliwhist';
export * from './games/jass';
export * from './games/jokerpoker';
export * from './games/kaiser';
export * from './games/kalooki';
export * from './games/karnoffel';
export * from './games/kemps';
export * from './games/kille';
export * from './games/king';
export * from './games/kingalbert';
export * from './games/klaberjass';
export * from './games/klaverjas';
export * from './games/klondike';
export * from './games/knockoutwhist';
export * from './games/koenigrufen';
export * from './games/koikoi';
export * from './games/labellelucie';
export * from './games/laughandliedown';
export * from './games/letitride';
export * from './games/lingerlonger';
export * from './games/literature';
export * from './games/loba';
export * from './games/loo';
export * from './games/macau';
export * from './games/machiavelli';
export * from './games/manille';
export * from './games/mao';
export * from './games/marias';
export * from './games/memory';
export * from './games/mendikot';
export * from './games/michigan';
export * from './games/mighty';
export * from './games/minchiate';
export * from './games/minibridge';
export * from './games/mississippistud';
export * from './games/missmilligan';
export * from './games/montecarlo';
export * from './games/mus';
export * from './games/mushi';
export * from './games/nainjaune';
export * from './games/nap';
export * from './games/napoleon';
export * from './games/napoleonssquare';
export * from './games/nertz';
export * from './games/ninetynine';
export * from './games/niuniu';
export * from './games/oasispoker';
export * from './games/ohhell';
export * from './games/oichokabu';
export * from './games/oldmaid';
export * from './games/omaha';
export * from './games/omahahilo';
export * from './games/ombre';
export * from './games/openfacechinese';
export * from './games/osmosis';
export * from './games/pageone';
export * from './games/paigow';
export * from './games/pan';
export * from './games/pasur';
export * from './games/penguin';
export * from './games/pig';
export * from './games/pigtail';
export * from './games/pineapple';
export * from './games/pinochle';
export * from './games/piquet';
export * from './games/pishti';
export * from './games/pitch';
export * from './games/poch';
export * from './games/poker';
export * from './games/pokersquares';
export * from './games/polignac';
export * from './games/pontoon';
export * from './games/popejoan';
export * from './games/preference';
export * from './games/president';
export * from './games/primero';
export * from './games/prsi';
export * from './games/pyramid';
export * from './games/rams';
export * from './games/razz';
export * from './games/reddog';
export * from './games/reversis';
export * from './games/rikken';
export * from './games/rollingstone';
export * from './games/rook';
export * from './games/royalcotillion';
export * from './games/rummy500';
export * from './games/russianbank';
export * from './games/russianpoker';
export * from './games/russiansolitaire';
export * from './games/samba';
export * from './games/scarto';
export * from './games/schnapsen';
export * from './games/scopa';
export * from './games/scopone';
export * from './games/scorpion';
export * from './games/seahaventowers';
export * from './games/sedma';
export * from './games/sergeantmajor';
export * from './games/settemezzo';
export * from './games/sevenbridge';
export * from './games/sevencardstud';
export * from './games/sevencardstudhilo';
export * from './games/sevens';
export * from './games/sheepshead';
export * from './games/shelem';
export * from './games/shengji';
export * from './games/shithead';
export * from './games/shortdeck';
export * from './games/simplesimon';
export * from './games/sirtommy';
export * from './games/sixbidsolo';
export * from './games/sixcardgolf';
export * from './games/sjavs';
export * from './games/skat';
export * from './games/skitgubbe';
export * from './games/slapjack';
export * from './games/slobberhannes';
export * from './games/snap';
export * from './games/soko';
export * from './games/solowhist';
export * from './games/spades';
export * from './games/spanish21';
export * from './games/speed';
export * from './games/spider';
export * from './games/spiderette';
export * from './games/spiteandmalice';
export * from './games/spoilfive';
export * from './games/spoons';
export * from './games/stealingbundles';
export * from './games/streetsandalleys';
export * from './games/sueca';
export * from './games/sultan';
export * from './games/tablanet';
export * from './games/tarabish';
export * from './games/tarneeb';
export * from './games/tarocchini';
export * from './games/teendopaanch';
export * from './games/teenpatti';
export * from './games/terrace';
export * from './games/texasholdembonus';
export * from './games/thirtyone';
export * from './games/threecard';
export * from './games/threecardbrag';
export * from './games/threethirteen';
export * from './games/tichu';
export * from './games/tienlen';
export * from './games/toepen';
export * from './games/tonk';
export * from './games/trash';
export * from './games/trenteetquarante';
export * from './games/tressette';
export * from './games/trex';
export * from './games/tripeaks';
export * from './games/truco';
export * from './games/tute';
export * from './games/twentynine';
export * from './games/twotenjack';
export * from './games/tysiac';
export * from './games/ulti';
export * from './games/ultimatetexasholdem';
export * from './games/videopoker';
export * from './games/vint';
export * from './games/vira';
export * from './games/war';
export * from './games/wasp';
export * from './games/watten';
export * from './games/whist';
export * from './games/windmill';
export * from './games/wizard';
export * from './games/yaniv';
export * from './games/yukon';
export * from './games/zheng';
export * from './games/zwicker';

/** Every registered game name — the SSoT the {@link Game} union is derived from. */
export const games = [
  'blackjack',
  'poker',
  'oldmaid',
  'daifugo',
  'bigtwo',
  'tienlen',
  'sevens',
  'doubt',
  'durak',
  'holdem',
  'omaha',
  'omahahilo',
  'bigo',
  'bigohilo',
  'shortdeck',
  'pineapple',
  'crazypineapple',
  'irishpoker',
  'sevencardstud',
  'fivecardstud',
  'soko',
  'fourseasons',
  'colorado',
  'cribbagesquares',
  'diplomat',
  'royalcotillion',
  'crazyquilt',
  'germanwhist',
  'slobberhannes',
  'polignac',
  'reversis',
  'rams',
  'tarabish',
  'baloot',
  'estimation',
  'israeliwhist',
  'hokm',
  'shelem',
  'mendikot',
  'bhabhi',
  'teendopaanch',
  'hasenpfeffer',
  'sergeantmajor',
  'honeymoonbridge',
  'minibridge',
  'pasur',
  'snap',
  'rollingstone',
  'lingerlonger',
  'pig',
  'stealingbundles',
  'cucumber',
  'goofspiel',
  'andarbahar',
  'botifarra',
  'rikken',
  'colourwhist',
  'chemindefer',
  'razz',
  'sevencardstudhilo',
  'badugi',
  'deucetoseven',
  'hearts',
  'spades',
  'twotenjack',
  'napoleon',
  'ohhell',
  'wizard',
  'ninetynine',
  'memory',
  'klondike',
  'freecell',
  'bakersgame',
  'seahaventowers',
  'cruel',
  'baccarat',
  'crazyeights',
  'prsi',
  'ginrummy',
  'indianrummy',
  'machiavelli',
  'conquian',
  'chinchon',
  'threethirteen',
  'canasta',
  'samba',
  'handandfoot',
  'burraco',
  'spider',
  'indianpoker',
  'videopoker',
  'deuceswild',
  'jokerpoker',
  'euchre',
  'bridge',
  'pyramid',
  'tripeaks',
  'cribbage',
  'threecard',
  'caribbeanstud',
  'texasholdembonus',
  'paigow',
  'speed',
  'war',
  'fiftyone',
  'gofish',
  'pinochle',
  'golf',
  'pigtail',
  'clocksolitaire',
  'fortythieves',
  'calculation',
  'sirtommy',
  'auldlangsyne',
  'bisley',
  'napoleonssquare',
  'grandfathersclock',
  'duchess',
  'windmill',
  'americantoad',
  'congress',
  'terrace',
  'braid',
  'pontoon',
  'settemezzo',
  'niuniu',
  'bura',
  'mushi',
  'toepen',
  'chineseten',
  'skitgubbe',
  'laughandliedown',
  'sjavs',
  'trex',
  'loba',
  'desmoche',
  'zwicker',
  'poch',
  'popejoan',
  'nainjaune',
  'kille',
  'klaberjass',
  'kaiser',
  'boston',
  'vint',
  'bideuchre',
  'sixbidsolo',
  'karnoffel',
  'literature',
  'guandan',
  'shengji',
  'missmilligan',
  'canfield',
  'osmosis',
  'fivehundred',
  'yukon',
  'russiansolitaire',
  'scorpion',
  'wasp',
  'accordion',
  'sevenbridge',
  'trash',
  'whist',
  'catchten',
  'letitride',
  'pokersquares',
  'pageone',
  'reddog',
  'president',
  'cassino',
  'scopa',
  'scopone',
  'escoba',
  'barbu',
  'macau',
  'mao',
  'bristol',
  'bidwhist',
  'spanish21',
  'spiteandmalice',
  'skat',
  'shithead',
  'nertz',
  'slapjack',
  'egyptianratscrew',
  'bakersdozen',
  'thirtyone',
  'yaniv',
  'gongzhu',
  'tonk',
  'casinowar',
  'pitch',
  'dragontiger',
  'blackjackswitch',
  'montecarlo',
  'contractrummy',
  'carioca',
  'kalooki',
  'ultimatetexasholdem',
  'crescent',
  'mississippistud',
  'belote',
  'jass',
  'watten',
  'spiderette',
  'mighty',
  'oasispoker',
  'beleagueredcastle',
  'streetsandalleys',
  'kingalbert',
  'flowergarden',
  'fortyandeight',
  'sultan',
  'agnes',
  'piquet',
  'casinoholdem',
  'callbreak',
  'tarneeb',
  'highcardflush',
  'briscola',
  'gaps',
  'fourcardpoker',
  'rummy500',
  'eightoff',
  'penguin',
  'russianpoker',
  'chinesepoker',
  'sixcardgolf',
  'doudizhu',
  'truco',
  'acesup',
  'schnapsen',
  'tressette',
  'easthaven',
  'tichu',
  'bourre',
  'sheepshead',
  'doppelkopf',
  'mus',
  'tute',
  'sueca',
  'klaverjas',
  'manille',
  'marias',
  'sedma',
  'knockoutwhist',
  'spoilfive',
  'solowhist',
  'fortyfives',
  'nap',
  'ganjifa',
  'preference',
  'vira',
  'twentynine',
  'courtpiece',
  'bezique',
  'ecarte',
  'threecardbrag',
  'teenpatti',
  'spoons',
  'kemps',
  'cuckoo',
  'pishti',
  'cuarenta',
  'faro',
  'openfacechinese',
  'russianbank',
  'labellelucie',
  'simplesimon',
  'doubleklondike',
  'blackhole',
  'beggarmyneighbour',
  'allfours',
  'gaigel',
  'king',
  'tysiac',
  'calabresella',
  'ombre',
  'ulti',
  'rook',
  'cinch',
  'loo',
  'basra',
  'hachihachi',
  'koikoi',
  'gostop',
  'tablanet',
  'trenteetquarante',
  'guts',
  'anaconda',
  'bouillotte',
  'primero',
  'michigan',
  'pan',
  'oichokabu',
  'aluette',
  'minchiate',
  'tarocchini',
  'scarto',
  'cego',
  'frenchtarot',
  'koenigrufen',
  'zheng',
] as const;

/** Union of every registered game name — the keys of the `games` list. */
export type Game = (typeof games)[number];

/** API clients for fetching action logs from each game's /log endpoint. */
export const actionLogApi: { [K in Game]: () => Promise<ActionLogResponse> } = games.reduce(
  (acc, game) => {
    acc[game] = () => gameExec<ActionLogResponse>(game, { command: 'log' });
    return acc;
  },
  {} as { [K in Game]: () => Promise<ActionLogResponse> },
);
