import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { rookApi, sessionId } from './gameApi';

describe('rookApi', () => {
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
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 3,
    leadPlayerIdx: 0,
    trumpColor: -1,
    contractBid: 0,
    declarerIdx: -1,
    highestBid: 0,
    highestBidder: -1,
    nestCount: 5,
    nest: [],
    currentTrick: [],
    teamScores: [0, 0],
    teamPoints: [0, 0],
    gameEndFlag: false,
    winnerTeam: -1,
    config: { cpuDifficulty: 1, targetScore: 500 },
    message: '',
  };

  it('reset hits /rook/exec with command=reset', async () => {
    mockFetch.mockReturnValue(ok(payload));
    const result = await rookApi.exec('reset');
    expect(mockFetch).toHaveBeenCalledWith('/rook/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'reset', sessionId }),
    });
    expect(result).toEqual(payload);
  });

  it('reset with config includes the config object', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await rookApi.exec('reset', { config: { cpuDifficulty: 2, targetScore: 700 } });
    expect(mockFetch).toHaveBeenCalledWith(
      '/rook/exec',
      expect.objectContaining({
        body: JSON.stringify({ command: 'reset', config: { cpuDifficulty: 2, targetScore: 700 }, sessionId }),
      }),
    );
  });

  it('bid sends command=bid with the point value', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await rookApi.exec('bid', { bid: 85 });
    expect(mockFetch).toHaveBeenCalledWith(
      '/rook/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'bid', bid: 85, sessionId }) }),
    );
  });

  it('pass sends command=pass', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await rookApi.exec('pass');
    expect(mockFetch).toHaveBeenCalledWith(
      '/rook/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'pass', sessionId }) }),
    );
  });

  it('exchange sends discardIndices and trumpColor', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await rookApi.exec('exchange', { discardIndices: [0, 1, 2, 3, 4], trumpColor: 2 });
    expect(mockFetch).toHaveBeenCalledWith(
      '/rook/exec',
      expect.objectContaining({
        body: JSON.stringify({ command: 'exchange', discardIndices: [0, 1, 2, 3, 4], trumpColor: 2, sessionId }),
      }),
    );
  });

  it('play sends command=play with the card index', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await rookApi.exec('play', { cardIndex: 3 });
    expect(mockFetch).toHaveBeenCalledWith(
      '/rook/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'play', cardIndex: 3, sessionId }) }),
    );
  });

  it('nextround sends command=nextround', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await rookApi.exec('nextround');
    expect(mockFetch).toHaveBeenCalledWith(
      '/rook/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'nextround', sessionId }) }),
    );
  });

  it('throws on HTTP error', async () => {
    mockFetch.mockReturnValue(err());
    await expect(rookApi.exec('reset')).rejects.toThrow('HTTP error: 500');
  });
});
