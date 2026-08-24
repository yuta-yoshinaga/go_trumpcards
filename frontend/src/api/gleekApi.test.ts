import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { gleekApi, sessionId } from './gameApi';

describe('gleekApi', () => {
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

  it('reset hits /gleek/exec with command=reset', async () => {
    mockFetch.mockReturnValue(ok(payload));
    const result = await gleekApi.exec('reset');
    expect(mockFetch).toHaveBeenCalledWith('/gleek/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'reset', sessionId }),
    });
    expect(result).toEqual(payload);
  });

  it('reset with config includes the config object', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await gleekApi.exec('reset', { config: { cpuDifficulty: 2, targetRounds: 7 } });
    expect(mockFetch).toHaveBeenCalledWith(
      '/gleek/exec',
      expect.objectContaining({
        body: JSON.stringify({ command: 'reset', config: { cpuDifficulty: 2, targetRounds: 7 }, sessionId }),
      }),
    );
  });

  it('a raise sends command=bid with the amount', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await gleekApi.exec('bid', { bid: 14 });
    expect(mockFetch).toHaveBeenCalledWith(
      '/gleek/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'bid', bid: 14, sessionId }) }),
    );
  });

  // **0 は「降りる」で、省略ではない。** omitempty のような扱いで落とすと、
  // サーバは「bid is required」を返す。
  it('dropping out sends bid=0 rather than omitting it', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await gleekApi.exec('bid', { bid: 0 });
    expect(mockFetch).toHaveBeenCalledWith(
      '/gleek/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'bid', bid: 0, sessionId }) }),
    );
  });

  // **捨て札フェーズを抜ける唯一の呼び出し。**
  it('discard sends the whole index list', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await gleekApi.exec('discard', { discardIndices: [0, 1, 2, 3, 4, 5, 6] });
    expect(mockFetch).toHaveBeenCalledWith(
      '/gleek/exec',
      expect.objectContaining({
        body: JSON.stringify({ command: 'discard', discardIndices: [0, 1, 2, 3, 4, 5, 6], sessionId }),
      }),
    );
  });

  it('play sends command=play with the card index', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await gleekApi.exec('play', { cardIndex: 2 });
    expect(mockFetch).toHaveBeenCalledWith(
      '/gleek/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'play', cardIndex: 2, sessionId }) }),
    );
  });

  it('next and nextround send their bare commands', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await gleekApi.exec('next');
    expect(mockFetch).toHaveBeenCalledWith(
      '/gleek/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'next', sessionId }) }),
    );
    await gleekApi.exec('nextround');
    expect(mockFetch).toHaveBeenCalledWith(
      '/gleek/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'nextround', sessionId }) }),
    );
  });

  it('throws on HTTP error', async () => {
    mockFetch.mockReturnValue(err());
    await expect(gleekApi.exec('reset')).rejects.toThrow('HTTP error: 500');
  });
});
