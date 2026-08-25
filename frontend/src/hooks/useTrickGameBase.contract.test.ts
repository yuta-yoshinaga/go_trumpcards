import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import {
  bauernschnapsenApi,
  beloteApi,
  callBreakApi,
  catchtenApi,
  coincheApi,
  gaigelApi,
  gongzhuApi,
  heartsApi,
  jassApi,
  madrassoApi,
  spadesApi,
  tarneebApi,
  trappolaApi,
  tressetteApi,
  twoTenJackApi,
  whistApi,
} from '../api/gameApi';

/**
 * Contract between `useTrickGameBase` and every client it dispatches to.
 *
 * The hook sends `('play', undefined, cardIndex)` and
 * `('reset', undefined, undefined, config)` — **the card index is the second
 * positional slot, the config is the fourth.** Belote, Catch the Ten and Whist
 * read the index from the *first* slot, so `cardIndex` never reached the server
 * and all three shipped unplayable from the Web GUI (#6227).
 *
 * Per-game API tests could not catch it: each one calls its own client the way
 * that client happens to be written, so a client and its test agree with each
 * other while disagreeing with the hook. This test calls **every** client the
 * way the hook actually calls it, which is the only place the two meet.
 */
const CLIENTS = [
  { game: 'bauernschnapsen', api: bauernschnapsenApi },
  { game: 'belote', api: beloteApi },
  { game: 'callbreak', api: callBreakApi },
  { game: 'catchten', api: catchtenApi },
  { game: 'coinche', api: coincheApi },
  { game: 'gaigel', api: gaigelApi },
  { game: 'gongzhu', api: gongzhuApi },
  { game: 'hearts', api: heartsApi },
  { game: 'jass', api: jassApi },
  { game: 'madrasso', api: madrassoApi },
  { game: 'spades', api: spadesApi },
  { game: 'tarneeb', api: tarneebApi },
  { game: 'trappola', api: trappolaApi },
  { game: 'tressette', api: tressetteApi },
  { game: 'twotenjack', api: twoTenJackApi },
  { game: 'whist', api: whistApi },
] as const;

// biome-ignore lint/suspicious/noExplicitAny: the clients differ in arity by design; the point is to call them all the one way the hook does.
type AnyExec = (...args: any[]) => Promise<unknown>;

describe('useTrickGameBase dispatch contract', () => {
  const mockFetch = vi.fn();

  beforeEach(() => {
    vi.stubGlobal('fetch', mockFetch);
    mockFetch.mockReturnValue(Promise.resolve({ ok: true, status: 200, json: () => Promise.resolve({}) }));
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  const sentBody = () => JSON.parse(mockFetch.mock.calls[0][1].body);

  it.each(CLIENTS)('$game: play puts the card index in the second slot', async ({ game, api }) => {
    // useTrickGameBase.ts: (exec)('play', undefined, selectedCardIndices[0])
    await (api.exec as AnyExec)('play', undefined, 3);
    const body = sentBody();
    expect(mockFetch.mock.calls[0][0]).toBe(`/${game}/exec`);
    expect(body.command).toBe('play');
    expect(body.cardIndex, `${game} dropped the card index the hook sent`).toBe(3);
  });

  it.each(CLIENTS)('$game: reset carries the config in the fourth slot', async ({ game, api }) => {
    // useTrickGameBase.ts: (exec)('reset', undefined, undefined, defaultConfigRef.current)
    const config = { cpuDifficulty: 2 };
    await (api.exec as AnyExec)('reset', undefined, undefined, config);
    const body = sentBody();
    expect(mockFetch.mock.calls[0][0]).toBe(`/${game}/exec`);
    expect(body.command).toBe('reset');
    expect(body.config, `${game} dropped the reset config the hook sent`).toMatchObject(config);
  });

  // 負のコントロール: 1 番目のスロットは札のインデックスではない。
  // ここが cardIndex として読まれていたのが #6227 の中身。
  it.each(CLIENTS)('$game: the first slot is not read as the card index', async ({ api }) => {
    await (api.exec as AnyExec)('play', 7, 3);
    expect(sentBody().cardIndex).toBe(3);
  });
});
