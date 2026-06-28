import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { beggarmyneighbourApi, sessionId } from './gameApi';

describe('beggarmyneighbourApi', () => {
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
    currentPlayerIdx: 0,
    penaltyOwnerIdx: -1,
    penaltyRemaining: 0,
    centralPileSize: 0,
    lastCardPlayed: null,
    roundsPlayed: 0,
    config: { maxRounds: 2000 },
    message: '',
  };

  it('reset hits /beggarmyneighbour/exec with command=reset', async () => {
    mockFetch.mockReturnValue(ok(payload));
    const result = await beggarmyneighbourApi.exec('reset');
    expect(mockFetch).toHaveBeenCalledWith('/beggarmyneighbour/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'reset', sessionId }),
    });
    expect(result).toEqual(payload);
  });

  it('step sends command=step', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await beggarmyneighbourApi.exec('step');
    expect(mockFetch).toHaveBeenCalledWith(
      '/beggarmyneighbour/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'step', sessionId }) }),
    );
  });

  it('autoplay sends command=autoplay', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await beggarmyneighbourApi.exec('autoplay');
    expect(mockFetch).toHaveBeenCalledWith(
      '/beggarmyneighbour/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'autoplay', sessionId }) }),
    );
  });

  it('reset with config includes maxRounds', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await beggarmyneighbourApi.exec('reset', { maxRounds: 1000 });
    expect(mockFetch).toHaveBeenCalledWith(
      '/beggarmyneighbour/exec',
      expect.objectContaining({
        body: JSON.stringify({ command: 'reset', maxRounds: 1000, sessionId }),
      }),
    );
  });

  it('throws on HTTP error', async () => {
    mockFetch.mockReturnValue(err());
    await expect(beggarmyneighbourApi.exec('reset')).rejects.toThrow('HTTP error: 500');
  });
});
