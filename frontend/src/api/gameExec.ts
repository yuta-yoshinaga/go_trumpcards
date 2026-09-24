// Shared plumbing for the per-game API modules, split out of the 5,409-line
// gameApi.ts (issue #4434). gameApi.ts stays a barrel, so no existing import
// anywhere has to change.
//
// sessionId lives here and ONLY here on purpose: it is a module-scope
// crypto.randomUUID() evaluated once at import time. Duplicating it into each
// game module would give every game a different session id and silently break
// server-side session continuity.

import type { VideoPokerResponse } from '../types/card';

/** Unique session identifier for correlating API requests. */
export const sessionId: string = crypto.randomUUID();

/** Worker base URLs for Cloudflare deployment. Empty strings for Docker (relative URLs). */
export const WORKER_CASINO = import.meta.env.VITE_WORKER_CASINO_URL || '';

export const WORKER_CLASSIC = import.meta.env.VITE_WORKER_CLASSIC_URL || '';

export const WORKER_SOLO = import.meta.env.VITE_WORKER_SOLO_URL || '';

export const WORKER_EXTRA = import.meta.env.VITE_WORKER_EXTRA_URL || '';
/** Fifth size bucket (ADR-0036). */
export const WORKER_EXTRA2 = import.meta.env.VITE_WORKER_EXTRA2_URL || '';
/** Sixth size bucket (ADR-0036). */
export const WORKER_EXTRA3 = import.meta.env.VITE_WORKER_EXTRA3_URL || '';
/** Seventh size bucket (ADR-0037). */
export const WORKER_EXTRA4 = import.meta.env.VITE_WORKER_EXTRA4_URL || '';
/** Eighth size bucket (ADR-0038). */
export const WORKER_EXTRA5 = import.meta.env.VITE_WORKER_EXTRA5_URL || '';

/** Ninth size bucket (ADR-0041). */
export const WORKER_EXTRA6 = import.meta.env.VITE_WORKER_EXTRA6_URL || '';
/** Tenth size bucket (ADR-0041). */
export const WORKER_EXTRA7 = import.meta.env.VITE_WORKER_EXTRA7_URL || '';
/** Eleventh size bucket (ADR-0043). */
export const WORKER_EXTRA8 = import.meta.env.VITE_WORKER_EXTRA8_URL || '';
/** Twelfth size bucket (ADR-0043). */
export const WORKER_EXTRA9 = import.meta.env.VITE_WORKER_EXTRA9_URL || '';
/** Thirteenth size bucket (ADR-0043). */
export const WORKER_EXTRA10 = import.meta.env.VITE_WORKER_EXTRA10_URL || '';
/** Fourteenth size bucket (ADR-0043). */
export const WORKER_EXTRA11 = import.meta.env.VITE_WORKER_EXTRA11_URL || '';

/** Maps each game to its Worker base URL. */
export const workerUrl: Record<string, string> = {
  blackjack: WORKER_EXTRA8,
  spanish21: WORKER_EXTRA8,
  doubleexposure: WORKER_EXTRA8,
  baccarat: WORKER_EXTRA10,
  poker: WORKER_CASINO,
  holdem: WORKER_CASINO,
  omaha: WORKER_CASINO,
  omahahilo: WORKER_CASINO,
  dramaha: WORKER_CASINO,
  bigo: WORKER_CASINO,
  courchevel: WORKER_CASINO,
  bigohilo: WORKER_CASINO,
  courchevelhilo: WORKER_CASINO,
  shortdeck: WORKER_CASINO,
  indianpoker: WORKER_CASINO,
  videopoker: WORKER_EXTRA8,
  deuceswild: WORKER_EXTRA8,
  jokerpoker: WORKER_EXTRA8,
  threecard: WORKER_CASINO,
  threecardrummy: WORKER_EXTRA8,
  caribbeanstud: WORKER_EXTRA10,
  caribbeandraw: WORKER_EXTRA6,
  texasholdembonus: WORKER_CASINO,
  casinoholdem: WORKER_CASINO,
  paigow: WORKER_EXTRA6,
  pineapple: WORKER_CASINO,
  crazypineapple: WORKER_CASINO,
  irishpoker: WORKER_CASINO,
  followthequeen: WORKER_CASINO,
  sevencardstud: WORKER_CASINO,
  fivecardstud: WORKER_CASINO,
  razz: WORKER_CASINO,
  sevencardstudhilo: WORKER_CASINO,
  chicago: WORKER_CASINO,
  eightgame: WORKER_CASINO,
  badugi: WORKER_CASINO,
  deucetoseven: WORKER_CASINO,
  ecarte: WORKER_EXTRA6,
  threecardbrag: WORKER_CASINO,
  teenpatti: WORKER_CASINO,
  spoons: WORKER_EXTRA6,
  kemps: WORKER_EXTRA2,
  cuckoo: WORKER_EXTRA11,
  pishti: WORKER_EXTRA2,
  ristikontra: WORKER_EXTRA2,
  cuarenta: WORKER_EXTRA2,
  faro: WORKER_EXTRA6,
  basset: WORKER_EXTRA4,
  openfacechinese: WORKER_CASINO,
  calculation: WORKER_SOLO,
  sirtommy: WORKER_EXTRA2,
  auldlangsyne: WORKER_EXTRA2,
  soko: WORKER_CASINO,
  fourseasons: WORKER_SOLO,
  colorado: WORKER_CLASSIC,
  slyfox: WORKER_EXTRA2,
  cribbagesquares: WORKER_EXTRA2,
  diplomat: WORKER_EXTRA,
  royalcotillion: WORKER_CLASSIC,
  matrimony: WORKER_EXTRA,
  crazyquilt: WORKER_SOLO,
  germanwhist: WORKER_EXTRA7,
  slobberhannes: WORKER_EXTRA7,
  polignac: WORKER_EXTRA2,
  reversis: WORKER_EXTRA8,
  rams: WORKER_EXTRA2,
  tarabish: WORKER_EXTRA6,
  baloot: WORKER_EXTRA9,
  estimation: WORKER_EXTRA4,
  israeliwhist: WORKER_EXTRA4,
  hokm: WORKER_EXTRA7,
  shelem: WORKER_EXTRA5,
  mendikot: WORKER_EXTRA4,
  bhabhi: WORKER_EXTRA4,
  teendopaanch: WORKER_EXTRA5,
  hasenpfeffer: WORKER_EXTRA5,
  sergeantmajor: WORKER_EXTRA4,
  honeymoonbridge: WORKER_EXTRA4,
  minibridge: WORKER_EXTRA5,
  pasur: WORKER_EXTRA7,
  snap: WORKER_EXTRA9,
  rollingstone: WORKER_EXTRA3,
  lingerlonger: WORKER_EXTRA4,
  pig: WORKER_EXTRA6,
  stealingbundles: WORKER_EXTRA3,
  cucumber: WORKER_CLASSIC,
  goofspiel: WORKER_EXTRA5,
  bisley: WORKER_EXTRA2,
  napoleonssquare: WORKER_EXTRA2,
  grandfathersclock: WORKER_EXTRA2,
  bigben: WORKER_EXTRA3,
  duchess: WORKER_EXTRA2,
  windmill: WORKER_EXTRA2,
  americantoad: WORKER_EXTRA2,
  congress: WORKER_EXTRA3,
  saliclaw: WORKER_EXTRA3,
  terrace: WORKER_EXTRA3,
  braid: WORKER_EXTRA2,
  pontoon: WORKER_EXTRA9,
  settemezzo: WORKER_EXTRA9,
  quinze: WORKER_EXTRA9,
  niuniu: WORKER_EXTRA3,
  bura: WORKER_EXTRA3,
  mushi: WORKER_EXTRA11,
  toepen: WORKER_EXTRA3,
  chineseten: WORKER_EXTRA11,
  laughandliedown: WORKER_EXTRA2,
  loba: WORKER_EXTRA2,
  desmoche: WORKER_EXTRA3,
  zwicker: WORKER_EXTRA9,
  poch: WORKER_EXTRA10,
  popejoan: WORKER_EXTRA3,
  nainjaune: WORKER_EXTRA3,
  kille: WORKER_EXTRA10,
  klaberjass: WORKER_EXTRA6,
  kaiser: WORKER_EXTRA10,
  boston: WORKER_EXTRA3,
  vint: WORKER_EXTRA3,
  bideuchre: WORKER_EXTRA9,
  sixbidsolo: WORKER_EXTRA8,
  karnoffel: WORKER_CLASSIC,
  unsunkaruta: WORKER_CLASSIC,
  quodlibet: WORKER_EXTRA10,
  dehlapakad: WORKER_EXTRA5,
  sutda: WORKER_EXTRA6,
  cirulla: WORKER_EXTRA10,
  diloti: WORKER_EXTRA5,
  comet: WORKER_EXTRA5,
  costlycolours: WORKER_EXTRA6,
  baccaratbanque: WORKER_EXTRA2,
  continentalrummy: WORKER_EXTRA5,
  literature: WORKER_EXTRA4,
  guandan: WORKER_EXTRA11,
  aluette: WORKER_EXTRA9,
  shengji: WORKER_EXTRA8,
  sjavs: WORKER_EXTRA9,
  skitgubbe: WORKER_EXTRA3,
  trex: WORKER_EXTRA3,
  missmilligan: WORKER_EXTRA2,
  hearts: WORKER_EXTRA7,
  spades: WORKER_CLASSIC,
  pitch: WORKER_EXTRA5,
  euchre: WORKER_EXTRA11,
  bridge: WORKER_EXTRA3,
  napoleon: WORKER_EXTRA10,
  ninetynine: WORKER_CLASSIC,
  ohhell: WORKER_CLASSIC,
  wizard: WORKER_EXTRA5,
  oldmaid: WORKER_CLASSIC,
  doubt: WORKER_EXTRA5,
  durak: WORKER_CLASSIC,
  daifugo: WORKER_EXTRA11,
  bigtwo: WORKER_EXTRA2,
  tienlen: WORKER_EXTRA9,
  zheng: WORKER_EXTRA9,
  sevens: WORKER_CLASSIC,
  crazyeights: WORKER_CLASSIC,
  prsi: WORKER_CLASSIC,
  pageone: WORKER_CLASSIC,
  speed: WORKER_EXTRA2,
  war: WORKER_EXTRA2,
  fiftyone: WORKER_EXTRA2,
  gofish: WORKER_EXTRA2,
  pinochle: WORKER_EXTRA5,
  pigtail: WORKER_EXTRA2,
  twotenjack: WORKER_CLASSIC,
  whitehead: WORKER_SOLO,
  klondike: WORKER_SOLO,
  freecell: WORKER_SOLO,
  bakersgame: WORKER_SOLO,
  seahaventowers: WORKER_EXTRA9,
  cruel: WORKER_SOLO,
  spider: WORKER_SOLO,
  pyramid: WORKER_SOLO,
  pokersquares: WORKER_EXTRA9,
  tripeaks: WORKER_SOLO,
  memory: WORKER_EXTRA9,
  ginrummy: WORKER_CLASSIC,
  indianrummy: WORKER_EXTRA,
  marriage: WORKER_EXTRA11,
  machiavelli: WORKER_EXTRA,
  conquian: WORKER_EXTRA7,
  chinchon: WORKER_EXTRA7,
  threethirteen: WORKER_EXTRA,
  canasta: WORKER_EXTRA7,
  samba: WORKER_EXTRA7,
  bolivia: WORKER_EXTRA7,
  handandfoot: WORKER_EXTRA7,
  burraco: WORKER_EXTRA7,
  biriba: WORKER_EXTRA7,
  cribbage: WORKER_EXTRA6,
  golf: WORKER_SOLO,
  acesup: WORKER_SOLO,
  clocksolitaire: WORKER_EXTRA9,
  fortythieves: WORKER_SOLO,
  canfield: WORKER_SOLO,
  osmosis: WORKER_SOLO,
  fivehundred: WORKER_EXTRA11,
  yukon: WORKER_SOLO,
  alaska: WORKER_EXTRA4,
  speculation: WORKER_EXTRA5,
  russiansolitaire: WORKER_SOLO,
  scorpion: WORKER_SOLO,
  wasp: WORKER_SOLO,
  accordion: WORKER_SOLO,
  sevenbridge: WORKER_EXTRA3,
  trash: WORKER_EXTRA2,
  whist: WORKER_CLASSIC,
  catchten: WORKER_CLASSIC,
  letitride: WORKER_EXTRA4,
  reddog: WORKER_EXTRA11,
  casinowar: WORKER_EXTRA11,
  president: WORKER_CLASSIC,
  cassino: WORKER_EXTRA8,
  spiteandmalice: WORKER_EXTRA9,
  ramsch: WORKER_EXTRA6,
  skat: WORKER_EXTRA6,
  shithead: WORKER_CLASSIC,
  nertz: WORKER_EXTRA5,
  slapjack: WORKER_CLASSIC,
  egyptianratscrew: WORKER_CLASSIC,
  bakersdozen: WORKER_SOLO,
  thirtyone: WORKER_EXTRA7,
  yaniv: WORKER_EXTRA9,
  trappola: WORKER_EXTRA2,
  julepe: WORKER_EXTRA9,
  schafkopf: WORKER_EXTRA7,
  coinche: WORKER_EXTRA6,
  germansolo: WORKER_EXTRA11,
  gleek: WORKER_EXTRA5,
  madrasso: WORKER_EXTRA3,
  tressette: WORKER_EXTRA6,
  tonk: WORKER_CLASSIC,
  andarbahar: WORKER_EXTRA7,
  botifarra: WORKER_EXTRA8,
  rikken: WORKER_EXTRA9,
  colourwhist: WORKER_EXTRA4,
  chemindefer: WORKER_EXTRA8,
  crazyfourpoker: WORKER_EXTRA7,
  doubleattack: WORKER_EXTRA8,
  freebet: WORKER_EXTRA8,
  banluck: WORKER_EXTRA8,
  montebank: WORKER_EXTRA11,
  tehonbiki: WORKER_EXTRA2,
  cincinnati: WORKER_CASINO,
  ironcross: WORKER_CASINO,
  baseballpoker: WORKER_CASINO,
  dragontiger: WORKER_EXTRA4,
  blackjackswitch: WORKER_EXTRA8,
  montecarlo: WORKER_EXTRA9,
  contractrummy: WORKER_EXTRA,
  carioca: WORKER_EXTRA,
  kalooki: WORKER_EXTRA,
  ultimatetexasholdem: WORKER_CASINO,
  crescent: WORKER_SOLO,
  sthelena: WORKER_EXTRA,
  mississippistud: WORKER_EXTRA6,
  belote: WORKER_EXTRA6,
  spiderette: WORKER_SOLO,
  willothewisp: WORKER_SOLO,
  mighty: WORKER_EXTRA6,
  oasispoker: WORKER_SOLO,
  russianpoker: WORKER_EXTRA8,
  stalactites: WORKER_EXTRA9,
  somerset: WORKER_SOLO,
  fortress: WORKER_SOLO,
  beleagueredcastle: WORKER_SOLO,
  piquet: WORKER_EXTRA8,
  callbreak: WORKER_CLASSIC,
  tarneeb: WORKER_EXTRA6,
  highcardflush: WORKER_EXTRA4,
  briscola: WORKER_CLASSIC,
  brusquembille: WORKER_EXTRA7,
  schnapsen: WORKER_EXTRA9,
  gaps: WORKER_SOLO,
  fourcardpoker: WORKER_EXTRA8,
  rummy500: WORKER_EXTRA,
  streetsandalleys: WORKER_EXTRA,
  kingalbert: WORKER_EXTRA10,
  flowergarden: WORKER_EXTRA4,
  fortyandeight: WORKER_EXTRA3,
  sultan: WORKER_EXTRA,
  agnes: WORKER_EXTRA5,
  jass: WORKER_EXTRA6,
  bauernschnapsen: WORKER_EXTRA9,
  gaigel: WORKER_EXTRA9,
  king: WORKER_EXTRA,
  tysiac: WORKER_EXTRA9,
  calabresella: WORKER_EXTRA9,
  ombre: WORKER_EXTRA8,
  quadrille: WORKER_EXTRA10,
  ulti: WORKER_EXTRA5,
  piedmontesetarot: WORKER_EXTRA4,
  scarto: WORKER_EXTRA4,
  cego: WORKER_EXTRA10,
  frenchtarot: WORKER_EXTRA,
  koenigrufen: WORKER_EXTRA,
  rook: WORKER_EXTRA10,
  cinch: WORKER_EXTRA,
  loo: WORKER_EXTRA5,
  basra: WORKER_EXTRA3,
  hachihachi: WORKER_EXTRA,
  koikoi: WORKER_EXTRA3,
  gostop: WORKER_EXTRA10,
  tablanet: WORKER_EXTRA3,
  trenteetquarante: WORKER_EXTRA4,
  guts: WORKER_EXTRA11,
  seventwentyseven: WORKER_EXTRA11,
  anaconda: WORKER_EXTRA8,
  bouillotte: WORKER_EXTRA10,
  primero: WORKER_EXTRA10,
  michigan: WORKER_EXTRA4,
  watten: WORKER_EXTRA,
  pan: WORKER_EXTRA,
  oichokabu: WORKER_EXTRA10,
  kingo: WORKER_EXTRA10,
  tusac: WORKER_EXTRA10,
  sakura: WORKER_EXTRA3,
  zwanzigerrufen: WORKER_EXTRA,
  tapptarock: WORKER_EXTRA,
  troggu: WORKER_EXTRA,
  horse: WORKER_CASINO,
  eightoff: WORKER_SOLO,
  penguin: WORKER_EXTRA9,
  chinesepoker: WORKER_CASINO,
  sixcardgolf: WORKER_EXTRA2,
  doudizhu: WORKER_EXTRA7,
  truco: WORKER_CLASSIC,
  put: WORKER_EXTRA4,
  scopa: WORKER_EXTRA7,
  scopone: WORKER_EXTRA7,
  escoba: WORKER_EXTRA7,
  barbu: WORKER_EXTRA8,
  macau: WORKER_EXTRA5,
  mao: WORKER_EXTRA10,
  russianbank: WORKER_EXTRA4,
  shamrocks: WORKER_CLASSIC,
  perseverance: WORKER_EXTRA4,
  fourteenout: WORKER_EXTRA4,
  narcotic: WORKER_EXTRA4,
  mrsmop: WORKER_EXTRA4,
  rankandfile: WORKER_EXTRA4,
  labellelucie: WORKER_EXTRA8,
  curdsandwhey: WORKER_EXTRA8,
  simplesimon: WORKER_EXTRA8,
  doubleklondike: WORKER_EXTRA2,
  blackhole: WORKER_EXTRA9,
  gongzhu: WORKER_EXTRA4,
  bristol: WORKER_SOLO,
  bidwhist: WORKER_EXTRA11,
  easthaven: WORKER_SOLO,
  tichu: WORKER_EXTRA6,
  bourre: WORKER_EXTRA6,
  sheepshead: WORKER_EXTRA11,
  doppelkopf: WORKER_EXTRA6,
  mus: WORKER_EXTRA6,
  tute: WORKER_EXTRA11,
  sueca: WORKER_EXTRA6,
  klaverjas: WORKER_EXTRA7,
  manille: WORKER_EXTRA7,
  marias: WORKER_CLASSIC,
  sedma: WORKER_CLASSIC,
  knockoutwhist: WORKER_EXTRA8,
  spoilfive: WORKER_EXTRA8,
  solowhist: WORKER_EXTRA8,
  fortyfives: WORKER_EXTRA11,
  nap: WORKER_EXTRA8,
  preference: WORKER_EXTRA4,
  ganjifa: WORKER_EXTRA,
  minchiate: WORKER_EXTRA7,
  tarocchini: WORKER_EXTRA7,
  vira: WORKER_EXTRA9,
  twentynine: WORKER_EXTRA11,
  courtpiece: WORKER_EXTRA6,
  bezique: WORKER_EXTRA4,
  beggarmyneighbour: WORKER_EXTRA2,
  allfours: WORKER_EXTRA8,
  citadel: WORKER_SOLO,
  batak: WORKER_EXTRA5,
  binokel: WORKER_EXTRA5,
  marjapussi: WORKER_EXTRA11,
  omi: WORKER_EXTRA5,
  tongits: WORKER_EXTRA5,
};

export async function postJson<T>(url: string, body: unknown): Promise<T> {
  const res = await fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  if (!res.ok) throw new Error(`HTTP error: ${res.status}`);
  return res.json() as Promise<T>;
}

export function gameExec<T>(game: string, body: Record<string, unknown>): Promise<T> {
  const base = workerUrl[game] || '';
  return postJson<T>(`${base}/${game}/exec`, { ...body, sessionId });
}

/** Factory for bid-play trick-taking APIs that share the same exec pattern. */
export function createBidPlayApi<T, C>(game: string) {
  return {
    exec: (
      command: 'reset' | 'bid' | 'play' | 'next' | 'nextround' | 'hint' | 'log',
      bid?: number,
      cardIndex?: number,
      config?: C,
    ) => gameExec<T>(game, { command, bid, cardIndex, config }),
  };
}

/**
 * Factory for solitaire-style move APIs whose request body is `{ command, from, to, n }`.
 *
 * Used by Canfield, FreeCell, Yukon, Scorpion, Accordion, FortyThieves, and
 * Calculation — every solitaire variant whose move endpoint takes only
 * source/target zones and an optional batch-undo count.
 *
 * `Cmd` is intentionally not defaulted: each call site declares the exact
 * command union its game accepts so invalid commands are rejected at compile
 * time instead of being silently widened to a broader shared union.
 */
export function createSolitaireMoveApi<T, Zone, Cmd extends string>(game: string) {
  return {
    exec: (command: Cmd, from?: Zone, to?: Zone, n?: number) => gameExec<T>(game, { command, from, to, n }),
  };
}

/**
 * Factory for solitaire-style move APIs that also carry an optional `config`
 * object (Klondike, Spider). Body shape: `{ command, from, to, config, n }`.
 *
 * Like {@link createSolitaireMoveApi}, the `Cmd` generic is not defaulted —
 * each call site declares its exact command union.
 */
export function createSolitaireMoveApiWithConfig<T, Zone, C, Cmd extends string>(game: string) {
  return {
    exec: (command: Cmd, from?: Zone, to?: Zone, config?: C, n?: number) =>
      gameExec<T>(game, { command, from, to, config, n }),
  };
}

/** Factory for video poker variant APIs that share the same exec pattern. */
export function createVideoPokerApi(game: string) {
  return {
    exec: (command: 'reset' | 'bet' | 'hold' | 'log', amount?: number, indices?: number[]) =>
      gameExec<VideoPokerResponse>(game, { command, amount, indices }),
  };
}

/**
 * Factory for casino bet APIs whose request body is `{ command, amount }`.
 * Used by Let It Ride and Red Dog — table games whose only per-action input
 * is the wager amount.
 */
export function createBetAmountApi<T, Cmd extends string>(game: string) {
  return {
    exec: (command: Cmd, amount?: number) => gameExec<T>(game, { command, amount }),
  };
}
