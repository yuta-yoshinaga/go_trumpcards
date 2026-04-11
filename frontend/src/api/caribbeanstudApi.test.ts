import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { caribbeanstudApi, sessionId } from './gameApi';

const csp = caribbeanstudApi;

describe('caribbeanstudApi', () => {
  const mockFetch = vi.fn();

  beforeEach(() => {
    vi.stubGlobal('fetch', mockFetch);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  function makeResponse(data: unknown, ok = true, status = 200) {
    return Promise.resolve({
      ok,
      status,
      json: () => Promise.resolve(data),
    });
  }

  const payload = {
    playerHand: [],
    dealerHand: [],
    phase: 1,
    chips: 1000,
    anteBet: 0,
    jackpotBet: 0,
    playBet: 0,
    result: 0,
    antePayout: 0,
    playPayout: 0,
    jackpotPayout: 0,
    totalPayout: 0,
    dealerQualified: false,
    playerHandRank: 0,
    dealerHandRank: 0,
    message: '',
  };

  it('calls the correct URL with reset command', async () => {
    mockFetch.mockReturnValue(makeResponse(payload));
    const result = await csp['exec']('reset');
    expect(mockFetch).toHaveBeenCalledWith('/caribbeanstud/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        command: 'reset',
        amount: undefined,
        jackpotBet: undefined,
        sessionId,
      }),
    });
    expect(result).toEqual(payload);
  });

  it('calls with bet command and ante only', async () => {
    mockFetch.mockReturnValue(makeResponse(payload));
    await csp['exec']('bet', 100);
    expect(mockFetch).toHaveBeenCalledWith(
      '/caribbeanstud/exec',
      expect.objectContaining({
        body: JSON.stringify({
          command: 'bet',
          amount: 100,
          jackpotBet: undefined,
          sessionId,
        }),
      }),
    );
  });

  it('calls with bet command and jackpot side bet', async () => {
    mockFetch.mockReturnValue(makeResponse(payload));
    await csp['exec']('bet', 100, 10);
    expect(mockFetch).toHaveBeenCalledWith(
      '/caribbeanstud/exec',
      expect.objectContaining({
        body: JSON.stringify({
          command: 'bet',
          amount: 100,
          jackpotBet: 10,
          sessionId,
        }),
      }),
    );
  });

  it('calls with play command', async () => {
    mockFetch.mockReturnValue(makeResponse(payload));
    await csp['exec']('play');
    expect(mockFetch).toHaveBeenCalledWith(
      '/caribbeanstud/exec',
      expect.objectContaining({
        body: JSON.stringify({
          command: 'play',
          amount: undefined,
          jackpotBet: undefined,
          sessionId,
        }),
      }),
    );
  });

  it('calls with fold command', async () => {
    mockFetch.mockReturnValue(makeResponse(payload));
    await csp['exec']('fold');
    expect(mockFetch).toHaveBeenCalledWith(
      '/caribbeanstud/exec',
      expect.objectContaining({
        body: JSON.stringify({
          command: 'fold',
          amount: undefined,
          jackpotBet: undefined,
          sessionId,
        }),
      }),
    );
  });

  it('throws on HTTP error', async () => {
    mockFetch.mockReturnValue(makeResponse(null, false, 500));
    await expect(csp['exec']('reset')).rejects.toThrow('HTTP error: 500');
  });
});
