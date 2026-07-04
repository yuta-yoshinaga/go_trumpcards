import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { primeroApi, sessionId } from './gameApi';

describe('primeroApi', () => {
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
    ante: 10,
    chips: 190,
    currentBet: 10,
    raiseCount: 0,
    maxRaises: 4,
    currentPlayerIdx: 0,
    dealerIdx: 0,
    isHumanTurn: true,
    canRaise: true,
    winnerIdx: -1,
    matchWinnerIdx: -1,
    result: 0,
    gameEndFlag: false,
    config: { playerCount: 4, ante: 10, startingChips: 200, targetRounds: 10 },
    message: '',
  };

  it('reset hits /primero/exec with command=reset', async () => {
    mockFetch.mockReturnValue(ok(payload));
    const result = await primeroApi.exec('reset');
    expect(mockFetch).toHaveBeenCalledWith('/primero/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'reset', sessionId }),
    });
    expect(result).toEqual(payload);
  });

  it('reset with config includes the config object', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await primeroApi.exec('reset', undefined, { playerCount: 3, ante: 20, startingChips: 500, targetRounds: 20 });
    expect(mockFetch).toHaveBeenCalledWith(
      '/primero/exec',
      expect.objectContaining({
        body: JSON.stringify({
          command: 'reset',
          config: { playerCount: 3, ante: 20, startingChips: 500, targetRounds: 20 },
          sessionId,
        }),
      }),
    );
  });

  it('bet call sends command=bet with action=call', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await primeroApi.exec('bet', 'call');
    expect(mockFetch).toHaveBeenCalledWith(
      '/primero/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'bet', action: 'call', sessionId }) }),
    );
  });

  it('bet raise sends command=bet with action=raise', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await primeroApi.exec('bet', 'raise');
    expect(mockFetch).toHaveBeenCalledWith(
      '/primero/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'bet', action: 'raise', sessionId }) }),
    );
  });

  it('bet fold sends command=bet with action=fold', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await primeroApi.exec('bet', 'fold');
    expect(mockFetch).toHaveBeenCalledWith(
      '/primero/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'bet', action: 'fold', sessionId }) }),
    );
  });

  it('nextround sends command=nextround', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await primeroApi.exec('nextround');
    expect(mockFetch).toHaveBeenCalledWith(
      '/primero/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'nextround', sessionId }) }),
    );
  });

  it('throws on HTTP error', async () => {
    mockFetch.mockReturnValue(err());
    await expect(primeroApi.exec('reset')).rejects.toThrow('HTTP error: 500');
  });
});
