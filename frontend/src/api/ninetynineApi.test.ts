import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ninetyNineApi, sessionId } from './gameApi';

describe('ninetyNineApi', () => {
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
    dealNumber: 1,
    targetScore: 100,
    handSize: 9,
    trickNumber: 0,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 0,
    trumpSuit: 1,
    currentTrick: [],
    gameEndFlag: false,
    winnerIdx: -1,
    leadPlayerIdx: 0,
    config: { cpuDifficulty: 1, targetScore: 100 },
    message: '',
  };

  it('reset hits /ninetynine/exec with command=reset', async () => {
    mockFetch.mockReturnValue(ok(payload));
    const result = await ninetyNineApi.exec('reset');
    expect(mockFetch).toHaveBeenCalledWith('/ninetynine/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'reset', sessionId }),
    });
    expect(result).toEqual(payload);
  });

  it('bid sends buryIndices array', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await ninetyNineApi.exec('bid', [0, 1, 2]);
    expect(mockFetch).toHaveBeenCalledWith(
      '/ninetynine/exec',
      expect.objectContaining({
        body: JSON.stringify({ command: 'bid', buryIndices: [0, 1, 2], sessionId }),
      }),
    );
  });

  it('play sends cardIndex', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await ninetyNineApi.exec('play', undefined, 3);
    expect(mockFetch).toHaveBeenCalledWith(
      '/ninetynine/exec',
      expect.objectContaining({
        body: JSON.stringify({ command: 'play', cardIndex: 3, sessionId }),
      }),
    );
  });

  it('reset with config includes cpuDifficulty and targetScore', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await ninetyNineApi.exec('reset', undefined, undefined, { cpuDifficulty: 2, targetScore: 100 });
    expect(mockFetch).toHaveBeenCalledWith(
      '/ninetynine/exec',
      expect.objectContaining({
        body: JSON.stringify({ command: 'reset', config: { cpuDifficulty: 2, targetScore: 100 }, sessionId }),
      }),
    );
  });

  it('next sends command=next', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await ninetyNineApi.exec('next');
    expect(mockFetch).toHaveBeenCalledWith(
      '/ninetynine/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'next', sessionId }) }),
    );
  });

  it('throws on HTTP error', async () => {
    mockFetch.mockReturnValue(err());
    await expect(ninetyNineApi.exec('reset')).rejects.toThrow('HTTP error: 500');
  });
});
