import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { sessionId, warApi } from './gameApi';

describe('warApi', () => {
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
    gameEndFlag: false,
    winnerIdx: -1,
    playerRevealed: null,
    cpuRevealed: null,
    warPotSize: 0,
    lastWinnerIdx: -1,
    lastBurialCount: 0,
    roundsPlayed: 0,
    config: { maxRounds: 500 },
    message: '',
  };

  it('reset hits /war/exec with command=reset', async () => {
    mockFetch.mockReturnValue(ok(payload));
    const result = await warApi['exec']('reset');
    expect(mockFetch).toHaveBeenCalledWith('/war/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'reset', sessionId }),
    });
    expect(result).toEqual(payload);
  });

  it('step sends command=step', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await warApi['exec']('step');
    expect(mockFetch).toHaveBeenCalledWith(
      '/war/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'step', sessionId }) }),
    );
  });

  it('reset with config includes maxRounds', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await warApi['exec']('reset', { maxRounds: 200 });
    expect(mockFetch).toHaveBeenCalledWith(
      '/war/exec',
      expect.objectContaining({
        body: JSON.stringify({ command: 'reset', maxRounds: 200, sessionId }),
      }),
    );
  });

  it('throws on HTTP error', async () => {
    mockFetch.mockReturnValue(err());
    await expect(warApi['exec']('reset')).rejects.toThrow('HTTP error: 500');
  });
});
