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
/** Fifth size bucket (ADR-0036). Empty until Phase 2 moves games in. */
export const WORKER_EXTRA2 = import.meta.env.VITE_WORKER_EXTRA2_URL || '';
/** Sixth size bucket (ADR-0036). Empty until Phase 2 moves games in. */
export const WORKER_EXTRA3 = import.meta.env.VITE_WORKER_EXTRA3_URL || '';

/** Maps each game to its Worker base URL. */
export const workerUrl: Record<string, string> = {
  blackjack: WORKER_CASINO,
  spanish21: WORKER_CASINO,
  baccarat: WORKER_CASINO,
  poker: WORKER_CASINO,
  holdem: WORKER_CASINO,
  omaha: WORKER_CASINO,
  omahahilo: WORKER_CASINO,
  bigo: WORKER_CASINO,
  bigohilo: WORKER_CASINO,
  shortdeck: WORKER_CASINO,
  indianpoker: WORKER_CASINO,
  videopoker: WORKER_CASINO,
  deuceswild: WORKER_CASINO,
  jokerpoker: WORKER_CASINO,
  threecard: WORKER_CASINO,
  caribbeanstud: WORKER_CASINO,
  texasholdembonus: WORKER_CASINO,
  casinoholdem: WORKER_CASINO,
  paigow: WORKER_CASINO,
  pineapple: WORKER_CASINO,
  crazypineapple: WORKER_CASINO,
  irishpoker: WORKER_CASINO,
  sevencardstud: WORKER_CASINO,
  fivecardstud: WORKER_CASINO,
  razz: WORKER_CASINO,
  badugi: WORKER_CASINO,
  deucetoseven: WORKER_CASINO,
  ecarte: WORKER_CASINO,
  threecardbrag: WORKER_CASINO,
  teenpatti: WORKER_CASINO,
  spoons: WORKER_EXTRA2,
  kemps: WORKER_EXTRA2,
  cuckoo: WORKER_EXTRA2,
  pishti: WORKER_EXTRA2,
  cuarenta: WORKER_EXTRA2,
  faro: WORKER_EXTRA2,
  openfacechinese: WORKER_CASINO,
  calculation: WORKER_SOLO,
  sirtommy: WORKER_EXTRA2,
  bisley: WORKER_EXTRA2,
  napoleonssquare: WORKER_EXTRA2,
  grandfathersclock: WORKER_EXTRA2,
  duchess: WORKER_EXTRA2,
  windmill: WORKER_EXTRA2,
  americantoad: WORKER_EXTRA2,
  congress: WORKER_EXTRA3,
  terrace: WORKER_EXTRA3,
  braid: WORKER_EXTRA2,
  pontoon: WORKER_EXTRA2,
  settemezzo: WORKER_EXTRA2,
  niuniu: WORKER_EXTRA3,
  bura: WORKER_EXTRA3,
  mushi: WORKER_EXTRA2,
  toepen: WORKER_EXTRA3,
  chineseten: WORKER_EXTRA2,
  missmilligan: WORKER_EXTRA2,
  hearts: WORKER_CLASSIC,
  spades: WORKER_CLASSIC,
  pitch: WORKER_CLASSIC,
  euchre: WORKER_SOLO,
  bridge: WORKER_EXTRA3,
  napoleon: WORKER_CASINO,
  ninetynine: WORKER_CLASSIC,
  ohhell: WORKER_CLASSIC,
  wizard: WORKER_EXTRA3,
  oldmaid: WORKER_CLASSIC,
  doubt: WORKER_EXTRA2,
  durak: WORKER_CLASSIC,
  daifugo: WORKER_CLASSIC,
  bigtwo: WORKER_EXTRA2,
  tienlen: WORKER_SOLO,
  zheng: WORKER_SOLO,
  sevens: WORKER_CLASSIC,
  crazyeights: WORKER_CLASSIC,
  prsi: WORKER_CLASSIC,
  pageone: WORKER_CLASSIC,
  speed: WORKER_EXTRA2,
  war: WORKER_EXTRA2,
  fiftyone: WORKER_EXTRA2,
  gofish: WORKER_EXTRA2,
  pinochle: WORKER_EXTRA2,
  pigtail: WORKER_EXTRA2,
  twotenjack: WORKER_CLASSIC,
  klondike: WORKER_SOLO,
  freecell: WORKER_SOLO,
  bakersgame: WORKER_SOLO,
  seahaventowers: WORKER_SOLO,
  cruel: WORKER_SOLO,
  spider: WORKER_SOLO,
  pyramid: WORKER_SOLO,
  pokersquares: WORKER_SOLO,
  tripeaks: WORKER_SOLO,
  memory: WORKER_SOLO,
  ginrummy: WORKER_EXTRA,
  indianrummy: WORKER_EXTRA,
  machiavelli: WORKER_EXTRA,
  conquian: WORKER_EXTRA,
  chinchon: WORKER_EXTRA,
  threethirteen: WORKER_EXTRA,
  canasta: WORKER_EXTRA,
  samba: WORKER_EXTRA,
  handandfoot: WORKER_EXTRA,
  burraco: WORKER_EXTRA,
  cribbage: WORKER_EXTRA3,
  golf: WORKER_SOLO,
  acesup: WORKER_SOLO,
  clocksolitaire: WORKER_SOLO,
  fortythieves: WORKER_SOLO,
  canfield: WORKER_SOLO,
  osmosis: WORKER_SOLO,
  fivehundred: WORKER_SOLO,
  yukon: WORKER_SOLO,
  russiansolitaire: WORKER_SOLO,
  scorpion: WORKER_SOLO,
  wasp: WORKER_SOLO,
  accordion: WORKER_SOLO,
  sevenbridge: WORKER_EXTRA3,
  trash: WORKER_EXTRA2,
  whist: WORKER_CLASSIC,
  catchten: WORKER_CLASSIC,
  letitride: WORKER_CASINO,
  reddog: WORKER_CASINO,
  casinowar: WORKER_CASINO,
  president: WORKER_CLASSIC,
  cassino: WORKER_CLASSIC,
  spiteandmalice: WORKER_EXTRA2,
  skat: WORKER_EXTRA3,
  shithead: WORKER_CLASSIC,
  nertz: WORKER_EXTRA2,
  slapjack: WORKER_CLASSIC,
  egyptianratscrew: WORKER_CLASSIC,
  bakersdozen: WORKER_SOLO,
  thirtyone: WORKER_SOLO,
  yaniv: WORKER_SOLO,
  tressette: WORKER_CASINO,
  tonk: WORKER_CLASSIC,
  dragontiger: WORKER_CASINO,
  blackjackswitch: WORKER_CASINO,
  montecarlo: WORKER_SOLO,
  contractrummy: WORKER_EXTRA,
  carioca: WORKER_EXTRA,
  kalooki: WORKER_EXTRA,
  ultimatetexasholdem: WORKER_CASINO,
  crescent: WORKER_SOLO,
  mississippistud: WORKER_CASINO,
  belote: WORKER_EXTRA3,
  spiderette: WORKER_SOLO,
  mighty: WORKER_EXTRA2,
  oasispoker: WORKER_CASINO,
  russianpoker: WORKER_CASINO,
  beleagueredcastle: WORKER_SOLO,
  piquet: WORKER_EXTRA3,
  callbreak: WORKER_CLASSIC,
  tarneeb: WORKER_CASINO,
  highcardflush: WORKER_CASINO,
  briscola: WORKER_CLASSIC,
  schnapsen: WORKER_SOLO,
  gaps: WORKER_SOLO,
  fourcardpoker: WORKER_CASINO,
  rummy500: WORKER_EXTRA,
  streetsandalleys: WORKER_EXTRA,
  kingalbert: WORKER_EXTRA,
  flowergarden: WORKER_EXTRA,
  fortyandeight: WORKER_EXTRA3,
  sultan: WORKER_EXTRA,
  agnes: WORKER_EXTRA,
  jass: WORKER_EXTRA3,
  gaigel: WORKER_EXTRA,
  king: WORKER_EXTRA,
  tysiac: WORKER_EXTRA,
  calabresella: WORKER_EXTRA,
  ombre: WORKER_EXTRA3,
  ulti: WORKER_EXTRA3,
  scarto: WORKER_EXTRA3,
  cego: WORKER_EXTRA3,
  frenchtarot: WORKER_EXTRA,
  koenigrufen: WORKER_EXTRA,
  rook: WORKER_EXTRA3,
  cinch: WORKER_EXTRA,
  loo: WORKER_EXTRA3,
  basra: WORKER_EXTRA3,
  hachihachi: WORKER_EXTRA,
  koikoi: WORKER_EXTRA3,
  gostop: WORKER_EXTRA,
  tablanet: WORKER_EXTRA3,
  trenteetquarante: WORKER_EXTRA,
  guts: WORKER_EXTRA,
  anaconda: WORKER_EXTRA,
  bouillotte: WORKER_EXTRA3,
  primero: WORKER_EXTRA3,
  michigan: WORKER_EXTRA3,
  watten: WORKER_EXTRA,
  pan: WORKER_EXTRA,
  oichokabu: WORKER_EXTRA,
  eightoff: WORKER_SOLO,
  penguin: WORKER_SOLO,
  chinesepoker: WORKER_CASINO,
  sixcardgolf: WORKER_EXTRA2,
  doudizhu: WORKER_CLASSIC,
  truco: WORKER_CLASSIC,
  scopa: WORKER_CLASSIC,
  scopone: WORKER_CLASSIC,
  escoba: WORKER_CLASSIC,
  barbu: WORKER_SOLO,
  macau: WORKER_SOLO,
  mao: WORKER_EXTRA3,
  russianbank: WORKER_SOLO,
  labellelucie: WORKER_CLASSIC,
  simplesimon: WORKER_CLASSIC,
  doubleklondike: WORKER_EXTRA2,
  blackhole: WORKER_SOLO,
  gongzhu: WORKER_SOLO,
  bristol: WORKER_SOLO,
  bidwhist: WORKER_SOLO,
  easthaven: WORKER_SOLO,
  tichu: WORKER_EXTRA2,
  bourre: WORKER_CASINO,
  sheepshead: WORKER_EXTRA3,
  doppelkopf: WORKER_CASINO,
  mus: WORKER_CASINO,
  tute: WORKER_CASINO,
  sueca: WORKER_CASINO,
  klaverjas: WORKER_CLASSIC,
  manille: WORKER_CLASSIC,
  marias: WORKER_CLASSIC,
  sedma: WORKER_CLASSIC,
  knockoutwhist: WORKER_CLASSIC,
  spoilfive: WORKER_CLASSIC,
  solowhist: WORKER_CLASSIC,
  fortyfives: WORKER_CASINO,
  nap: WORKER_CLASSIC,
  preference: WORKER_CLASSIC,
  twentynine: WORKER_CASINO,
  courtpiece: WORKER_CASINO,
  bezique: WORKER_CLASSIC,
  beggarmyneighbour: WORKER_EXTRA2,
  allfours: WORKER_CLASSIC,
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
