import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { sessionId, unsunKarutaApi } from './gameApi';

describe('unsunKarutaApi', () => {
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

  it('reset hits /unsunkaruta/exec', async () => {
    mockFetch.mockReturnValue(ok(payload));
    const result = await unsunKarutaApi.exec('reset');
    expect(mockFetch).toHaveBeenCalledWith('/unsunkaruta/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'reset', sessionId }),
    });
    expect(result).toEqual(payload);
  });

  it('reset carries the match length', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await unsunKarutaApi.exec('reset', { config: { cpuDifficulty: 2, targetDeals: 8 } });
    expect(mockFetch).toHaveBeenCalledWith(
      '/unsunkaruta/exec',
      expect.objectContaining({
        body: JSON.stringify({ command: 'reset', config: { cpuDifficulty: 2, targetDeals: 8 }, sessionId }),
      }),
    );
  });

  // **宣言は札と同じリクエストに乗る。** 別便で送れると、宣言だけ済んで札が
  // 出ていない盤面が作れてしまう。
  it('play sends the card index and the declaration together', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await unsunKarutaApi.exec('play', { cardIndex: 3, declare: true });
    expect(mockFetch).toHaveBeenCalledWith(
      '/unsunkaruta/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'play', cardIndex: 3, declare: true, sessionId }) }),
    );
  });

  // 0 は正当な札番号。undefined と同じ扱いにすると先頭の札が出せない。
  it('keeps card index 0 in the body', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await unsunKarutaApi.exec('play', { cardIndex: 0, declare: false });
    expect(mockFetch).toHaveBeenCalledWith(
      '/unsunkaruta/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'play', cardIndex: 0, declare: false, sessionId }) }),
    );
  });

  it.each(['next', 'nextround', 'hint', 'log'] as const)('%s carries no extra fields', async (command) => {
    mockFetch.mockReturnValue(ok(payload));
    await unsunKarutaApi.exec(command);
    expect(mockFetch).toHaveBeenCalledWith(
      '/unsunkaruta/exec',
      expect.objectContaining({ body: JSON.stringify({ command, sessionId }) }),
    );
  });

  it('throws when the server fails', async () => {
    mockFetch.mockReturnValue(err());
    await expect(unsunKarutaApi.exec('next')).rejects.toThrow();
  });
});
