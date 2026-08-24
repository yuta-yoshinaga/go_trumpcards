import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { germansoloApi, sessionId } from './gameApi';

describe('germansoloApi', () => {
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

  const payload = { players: [], phase: 0, message: '' };

  it('reset hits /germansolo/exec with command=reset', async () => {
    mockFetch.mockReturnValue(ok(payload));
    const result = await germansoloApi.exec('reset');
    expect(mockFetch).toHaveBeenCalledWith('/germansolo/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'reset', sessionId }),
    });
    expect(result).toEqual(payload);
  });

  it('reset with config includes the config object', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await germansoloApi.exec('reset', { config: { cpuDifficulty: 2, targetRounds: 7 } });
    expect(mockFetch).toHaveBeenCalledWith(
      '/germansolo/exec',
      expect.objectContaining({
        body: JSON.stringify({ command: 'reset', config: { cpuDifficulty: 2, targetRounds: 7 }, sessionId }),
      }),
    );
  });

  it('pass bid sends command=bid with bid=0', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await germansoloApi.exec('bid', { bid: 0 });
    expect(mockFetch).toHaveBeenCalledWith(
      '/germansolo/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'bid', bid: 0, sessionId }) }),
    );
  });

  it('a Frage bid sends bid=2 with the chosen trump suit', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await germansoloApi.exec('bid', { bid: 2, trumpSuit: 3 });
    expect(mockFetch).toHaveBeenCalledWith(
      '/germansolo/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'bid', bid: 2, trumpSuit: 3, sessionId }) }),
    );
  });

  it('a Tout bid sends bid=4 with the chosen trump suit', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await germansoloApi.exec('bid', { bid: 4, trumpSuit: 1 });
    expect(mockFetch).toHaveBeenCalledWith(
      '/germansolo/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'bid', bid: 4, trumpSuit: 1, sessionId }) }),
    );
  });

  // **エース呼びを抜ける唯一の呼び出し。** ここが落ちると Frage 落札の直後で
  // 盤面が固まる。
  it('ace sends command=ace with the called suit', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await germansoloApi.exec('ace', { aceSuit: 2 });
    expect(mockFetch).toHaveBeenCalledWith(
      '/germansolo/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'ace', aceSuit: 2, sessionId }) }),
    );
  });

  it('play sends command=play with the card index', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await germansoloApi.exec('play', { cardIndex: 2 });
    expect(mockFetch).toHaveBeenCalledWith(
      '/germansolo/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'play', cardIndex: 2, sessionId }) }),
    );
  });

  it('next and nextround send their bare commands', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await germansoloApi.exec('next');
    expect(mockFetch).toHaveBeenCalledWith(
      '/germansolo/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'next', sessionId }) }),
    );
    await germansoloApi.exec('nextround');
    expect(mockFetch).toHaveBeenCalledWith(
      '/germansolo/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'nextround', sessionId }) }),
    );
  });

  it('throws on HTTP error', async () => {
    mockFetch.mockReturnValue(err());
    await expect(germansoloApi.exec('reset')).rejects.toThrow('HTTP error: 500');
  });
});
