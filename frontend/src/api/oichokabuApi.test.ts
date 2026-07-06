import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { oichokabuApi, sessionId } from './gameApi';

describe('oichokabuApi', () => {
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
    bankerHand: [],
    playerRank: 0,
    bankerRank: 0,
    phase: 1,
    chips: 1000,
    bet: 0,
    result: 0,
    totalPayout: 0,
    message: '',
  };

  it('reset hits /oichokabu/exec with command=reset', async () => {
    mockFetch.mockReturnValue(ok(payload));
    const result = await oichokabuApi.exec('reset');
    expect(mockFetch).toHaveBeenCalledWith('/oichokabu/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'reset', sessionId }),
    });
    expect(result).toEqual(payload);
  });

  it('bet sends command=bet with the amount', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await oichokabuApi.exec('bet', 100);
    expect(mockFetch).toHaveBeenCalledWith(
      '/oichokabu/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'bet', amount: 100, sessionId }) }),
    );
  });

  it('draw sends command=draw', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await oichokabuApi.exec('draw');
    expect(mockFetch).toHaveBeenCalledWith(
      '/oichokabu/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'draw', sessionId }) }),
    );
  });

  it('stand sends command=stand', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await oichokabuApi.exec('stand');
    expect(mockFetch).toHaveBeenCalledWith(
      '/oichokabu/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'stand', sessionId }) }),
    );
  });

  it('throws on HTTP error', async () => {
    mockFetch.mockReturnValue(err());
    await expect(oichokabuApi.exec('reset')).rejects.toThrow('HTTP error: 500');
  });
});
