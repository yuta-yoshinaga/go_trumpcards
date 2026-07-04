import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { michiganApi, sessionId } from './gameApi';

describe('michiganApi', () => {
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
    boodles: [],
    phase: 0,
    roundNumber: 1,
    ante: 8,
    chips: 192,
    betBudget: 8,
    humanBetPlaced: false,
    currentPlayerIdx: 0,
    dealerIdx: 0,
    leadPlayerIdx: 0,
    seqSuit: 0,
    seqSuitName: '',
    seqHighValue: 0,
    needNewSequence: true,
    deadHandCount: 3,
    isHumanTurn: true,
    playableIndices: [],
    winnerIdx: -1,
    matchWinnerIdx: -1,
    result: 0,
    gameEndFlag: false,
    config: { playerCount: 4, ante: 8, startingChips: 200, targetRounds: 10 },
    message: '',
  };

  it('reset hits /michigan/exec with command=reset', async () => {
    mockFetch.mockReturnValue(ok(payload));
    const result = await michiganApi.exec('reset');
    expect(mockFetch).toHaveBeenCalledWith('/michigan/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'reset', sessionId }),
    });
    expect(result).toEqual(payload);
  });

  it('reset with config includes the config object', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await michiganApi.exec('reset', undefined, undefined, {
      playerCount: 6,
      ante: 12,
      startingChips: 500,
      targetRounds: 20,
    });
    expect(mockFetch).toHaveBeenCalledWith(
      '/michigan/exec',
      expect.objectContaining({
        body: JSON.stringify({
          command: 'reset',
          config: { playerCount: 6, ante: 12, startingChips: 500, targetRounds: 20 },
          sessionId,
        }),
      }),
    );
  });

  it('bet sends command=bet with the boodleBets array', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await michiganApi.exec('bet', [2, 2, 2, 2]);
    expect(mockFetch).toHaveBeenCalledWith(
      '/michigan/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'bet', boodleBets: [2, 2, 2, 2], sessionId }) }),
    );
  });

  it('play sends command=play with the cardIndex', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await michiganApi.exec('play', undefined, 3);
    expect(mockFetch).toHaveBeenCalledWith(
      '/michigan/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'play', cardIndex: 3, sessionId }) }),
    );
  });

  it('nextround sends command=nextround', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await michiganApi.exec('nextround');
    expect(mockFetch).toHaveBeenCalledWith(
      '/michigan/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'nextround', sessionId }) }),
    );
  });

  it('throws on HTTP error', async () => {
    mockFetch.mockReturnValue(err());
    await expect(michiganApi.exec('reset')).rejects.toThrow('HTTP error: 500');
  });
});
