import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { anacondaApi, sessionId } from './gameApi';

describe('anacondaApi', () => {
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
    dealerIdx: 0,
    currentPlayer: 0,
    passCount: 3,
    rollIndex: 0,
    pot: 40,
    currentBet: 0,
    raiseCount: 0,
    maxRaises: 3,
    ante: 10,
    chips: 200,
    winnerIdx: -1,
    matchWinnerIdx: -1,
    result: 0,
    gameEndFlag: false,
    isHumanTurn: true,
    canRaise: false,
    config: { playerCount: 4, ante: 10, startingChips: 200, targetRounds: 10 },
    message: '',
  };

  it('reset hits /anaconda/exec with command=reset', async () => {
    mockFetch.mockReturnValue(ok(payload));
    const result = await anacondaApi.exec('reset');
    expect(mockFetch).toHaveBeenCalledWith('/anaconda/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'reset', sessionId }),
    });
    expect(result).toEqual(payload);
  });

  it('reset with config includes the config object', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await anacondaApi.exec('reset', undefined, undefined, {
      playerCount: 6,
      ante: 20,
      startingChips: 500,
      targetRounds: 20,
    });
    expect(mockFetch).toHaveBeenCalledWith(
      '/anaconda/exec',
      expect.objectContaining({
        body: JSON.stringify({
          command: 'reset',
          config: { playerCount: 6, ante: 20, startingChips: 500, targetRounds: 20 },
          sessionId,
        }),
      }),
    );
  });

  it('pass sends command=pass with the card indices', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await anacondaApi.exec('pass', [0, 1, 2]);
    expect(mockFetch).toHaveBeenCalledWith(
      '/anaconda/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'pass', cardIndices: [0, 1, 2], sessionId }) }),
    );
  });

  it('keep sends command=keep with the 5 card indices', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await anacondaApi.exec('keep', [0, 1, 2, 3, 4]);
    expect(mockFetch).toHaveBeenCalledWith(
      '/anaconda/exec',
      expect.objectContaining({
        body: JSON.stringify({ command: 'keep', cardIndices: [0, 1, 2, 3, 4], sessionId }),
      }),
    );
  });

  it('bet sends command=bet with the action', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await anacondaApi.exec('bet', undefined, 'raise');
    expect(mockFetch).toHaveBeenCalledWith(
      '/anaconda/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'bet', action: 'raise', sessionId }) }),
    );
  });

  it('nextround sends command=nextround', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await anacondaApi.exec('nextround');
    expect(mockFetch).toHaveBeenCalledWith(
      '/anaconda/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'nextround', sessionId }) }),
    );
  });

  it('throws on HTTP error', async () => {
    mockFetch.mockReturnValue(err());
    await expect(anacondaApi.exec('reset')).rejects.toThrow('HTTP error: 500');
  });
});
