import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { catchtenApi, sessionId } from './gameApi';

describe('catchtenApi', () => {
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
    trickNumber: 1,
    currentPlayerIdx: 0,
    currentTrick: [],
    trumpSuit: 0,
    dealerIdx: 0,
    teamScores: [0, 0],
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: 0,
    config: { cpuDifficulty: 1, pointLimit: 41 },
    message: '',
  };

  it('reset hits /catchten/exec with command=reset', async () => {
    mockFetch.mockReturnValue(ok(payload));
    const result = await catchtenApi.exec('reset', undefined);
    expect(mockFetch).toHaveBeenCalledWith('/catchten/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'reset', sessionId }),
    });
    expect(result).toEqual(payload);
  });

  // **札のインデックスは 2 番目のスロット。** ここを 1 番目で書いていたので、
  // クライアントが arg1 を読む実装のままでも試験が通り、Web から札が出せない
  // 状態が出荷された (#6227)。
  it('play sends the card index from the second slot', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await catchtenApi.exec('play', undefined, 3);
    expect(mockFetch).toHaveBeenCalledWith(
      '/catchten/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'play', cardIndex: 3, sessionId }) }),
    );
  });

  it('next sends command=next', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await catchtenApi.exec('next', undefined);
    expect(mockFetch).toHaveBeenCalledWith(
      '/catchten/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'next', sessionId }) }),
    );
  });

  it('nextround sends command=nextround', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await catchtenApi.exec('nextround', undefined);
    expect(mockFetch).toHaveBeenCalledWith(
      '/catchten/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'nextround', sessionId }) }),
    );
  });

  it('reset with config includes cpuDifficulty and pointLimit', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await catchtenApi.exec('reset', undefined, undefined, { cpuDifficulty: 2, pointLimit: 51 });
    expect(mockFetch).toHaveBeenCalledWith(
      '/catchten/exec',
      expect.objectContaining({
        body: JSON.stringify({ command: 'reset', config: { cpuDifficulty: 2, pointLimit: 51 }, sessionId }),
      }),
    );
  });

  it('throws on HTTP error', async () => {
    mockFetch.mockReturnValue(err());
    await expect(catchtenApi.exec('reset', undefined)).rejects.toThrow('HTTP error: 500');
  });
});
