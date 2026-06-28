import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { allfoursApi, sessionId } from './gameApi';

describe('allfoursApi', () => {
  const mockFetch = vi.fn();

  beforeEach(() => {
    vi.stubGlobal('fetch', mockFetch);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  const ok = (data: unknown) => Promise.resolve({ ok: true, status: 200, json: () => Promise.resolve(data) });
  const err = () => Promise.resolve({ ok: false, status: 500, json: () => Promise.resolve(null) });

  const payload = {
    players: [],
    phase: 0,
    roundNumber: 1,
    trickNumber: 0,
    dealerIdx: 1,
    nonDealerIdx: 0,
    currentPlayerIdx: 0,
    trumpSuit: 0,
    turnUp: null,
    runCount: 0,
    currentTrick: [],
    gameEndFlag: false,
    winnerIdx: -1,
    leadPlayerIdx: -1,
    validPlayIndices: [],
    config: { cpuDifficulty: 1, pointLimit: 7 },
    message: '',
  };

  it('reset hits /allfours/exec with command=reset', async () => {
    mockFetch.mockReturnValue(ok(payload));
    const result = await allfoursApi.exec('reset');
    expect(mockFetch).toHaveBeenCalledWith('/allfours/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'reset', sessionId }),
    });
    expect(result).toEqual(payload);
  });

  it('beg sends command=beg with beg flag', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await allfoursApi.exec('beg', true);
    expect(mockFetch).toHaveBeenCalledWith(
      '/allfours/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'beg', beg: true, sessionId }) }),
    );
  });

  it('stand sends command=beg with beg=false', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await allfoursApi.exec('beg', false);
    expect(mockFetch).toHaveBeenCalledWith(
      '/allfours/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'beg', beg: false, sessionId }) }),
    );
  });

  it('respond sends command=respond with run flag', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await allfoursApi.exec('respond', undefined, true);
    expect(mockFetch).toHaveBeenCalledWith(
      '/allfours/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'respond', run: true, sessionId }) }),
    );
  });

  it('play sends command=play with cardIndex', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await allfoursApi.exec('play', undefined, undefined, 3);
    expect(mockFetch).toHaveBeenCalledWith(
      '/allfours/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'play', cardIndex: 3, sessionId }) }),
    );
  });

  it('reset with config includes cpuDifficulty and pointLimit', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await allfoursApi.exec('reset', undefined, undefined, undefined, { cpuDifficulty: 2, pointLimit: 11 });
    expect(mockFetch).toHaveBeenCalledWith(
      '/allfours/exec',
      expect.objectContaining({
        body: JSON.stringify({ command: 'reset', config: { cpuDifficulty: 2, pointLimit: 11 }, sessionId }),
      }),
    );
  });

  it('throws on HTTP error', async () => {
    mockFetch.mockReturnValue(err());
    await expect(allfoursApi.exec('reset')).rejects.toThrow('HTTP error: 500');
  });
});
