import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { gutsApi, sessionId } from './gameApi';

describe('gutsApi', () => {
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
    pot: 40,
    carryPot: 0,
    ante: 10,
    chips: 200,
    winnerIdx: -1,
    matchWinnerIdx: -1,
    result: 0,
    matchers: [],
    gameEndFlag: false,
    config: { playerCount: 4, ante: 10, startingChips: 200, targetRounds: 10 },
    message: '',
  };

  it('reset hits /guts/exec with command=reset', async () => {
    mockFetch.mockReturnValue(ok(payload));
    const result = await gutsApi.exec('reset');
    expect(mockFetch).toHaveBeenCalledWith('/guts/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'reset', sessionId }),
    });
    expect(result).toEqual(payload);
  });

  it('reset with config includes the config object', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await gutsApi.exec('reset', undefined, { playerCount: 6, ante: 20, startingChips: 500, targetRounds: 20 });
    expect(mockFetch).toHaveBeenCalledWith(
      '/guts/exec',
      expect.objectContaining({
        body: JSON.stringify({
          command: 'reset',
          config: { playerCount: 6, ante: 20, startingChips: 500, targetRounds: 20 },
          sessionId,
        }),
      }),
    );
  });

  it('declare sends command=declare with the declaration', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await gutsApi.exec('declare', 1);
    expect(mockFetch).toHaveBeenCalledWith(
      '/guts/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'declare', declaration: 1, sessionId }) }),
    );
  });

  it('declare 0 sends the out declaration', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await gutsApi.exec('declare', 0);
    expect(mockFetch).toHaveBeenCalledWith(
      '/guts/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'declare', declaration: 0, sessionId }) }),
    );
  });

  it('nextround sends command=nextround', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await gutsApi.exec('nextround');
    expect(mockFetch).toHaveBeenCalledWith(
      '/guts/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'nextround', sessionId }) }),
    );
  });

  it('throws on HTTP error', async () => {
    mockFetch.mockReturnValue(err());
    await expect(gutsApi.exec('reset')).rejects.toThrow('HTTP error: 500');
  });
});
