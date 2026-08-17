import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { sessionId, threecardApi } from './gameApi';

describe('threecardApi', () => {
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
    playerHand: [],
    dealerHand: [],
    phase: 0,
    chips: 0,
    anteBet: 0,
    pairPlusBet: 0,
    playBet: 0,
    result: 0,
    antePayout: 0,
    playPayout: 0,
    anteBonusPayout: 0,
    pairPlusPayout: 0,
    totalPayout: 0,
    dealerQualified: false,
    playerHandRank: 0,
    dealerHandRank: 0,
    message: '',
  };

  it('bet sends both the ante and the pair plus amount', async () => {
    mockFetch.mockReturnValue(ok(payload));
    const result = await threecardApi.exec('bet', 100, 50);
    expect(mockFetch).toHaveBeenCalledWith('/threecard/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'bet', amount: 100, pairPlusBet: 50, sessionId }),
    });
    expect(result).toEqual(payload);
  });

  // #5513: **rebet は額を送らない。** 直前の額を覚えているのはサーバ側なので、
  // ここで額を送れるようにすると「賭け直し」が普通の bet と区別できなくなる。
  it('rebet sends no amount at all', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await threecardApi.exec('rebet');
    expect(mockFetch).toHaveBeenCalledWith(
      '/threecard/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'rebet', sessionId }) }),
    );
  });

  it('throws on HTTP error', async () => {
    mockFetch.mockReturnValue(err());
    await expect(threecardApi.exec('rebet')).rejects.toThrow('HTTP error: 500');
  });
});
