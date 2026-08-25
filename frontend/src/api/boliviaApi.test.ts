import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { boliviaApi, sessionId } from './gameApi';

describe('boliviaApi', () => {
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

  const payload = { players: [], teamScores: [0, 0], phase: 0, roundNumber: 1, winnerIdx: -1, message: '' };

  it('reset hits /bolivia/exec with command=reset', async () => {
    mockFetch.mockReturnValue(ok(payload));
    const result = await boliviaApi.exec('reset');
    expect(mockFetch).toHaveBeenCalledWith('/bolivia/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'reset', sessionId }),
    });
    expect(result).toEqual(payload);
  });

  it('reset with config includes the config object', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await boliviaApi.exec('reset', undefined, { cpuDifficulty: 2, pointLimit: 15000 });
    expect(mockFetch).toHaveBeenCalledWith(
      '/bolivia/exec',
      expect.objectContaining({
        body: JSON.stringify({ command: 'reset', config: { cpuDifficulty: 2, pointLimit: 15000 }, sessionId }),
      }),
    );
  });

  it('drawstock sends command=drawstock', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await boliviaApi.exec('drawstock');
    expect(mockFetch).toHaveBeenCalledWith(
      '/bolivia/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'drawstock', sessionId }) }),
    );
  });

  it('drawdiscard forwards the natural pair indices', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await boliviaApi.exec('drawdiscard', undefined, undefined, [0, 1]);
    expect(mockFetch).toHaveBeenCalledWith(
      '/bolivia/exec',
      expect.objectContaining({
        body: JSON.stringify({ command: 'drawdiscard', naturalPairIndices: [0, 1], sessionId }),
      }),
    );
  });

  it('meld forwards the meld groups', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await boliviaApi.exec('meld', undefined, undefined, undefined, [[0, 1, 2]]);
    expect(mockFetch).toHaveBeenCalledWith(
      '/bolivia/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'meld', meldGroups: [[0, 1, 2]], sessionId }) }),
    );
  });

  it('discard sends the card index', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await boliviaApi.exec('discard', 3);
    expect(mockFetch).toHaveBeenCalledWith(
      '/bolivia/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'discard', cardIndex: 3, sessionId }) }),
    );
  });

  it('goout sends command=goout', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await boliviaApi.exec('goout');
    expect(mockFetch).toHaveBeenCalledWith(
      '/bolivia/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'goout', sessionId }) }),
    );
  });

  it('throws on HTTP error', async () => {
    mockFetch.mockReturnValue(err());
    await expect(boliviaApi.exec('reset')).rejects.toThrow('HTTP error: 500');
  });
});
