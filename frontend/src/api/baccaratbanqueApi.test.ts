import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { baccaratbanqueApi, sessionId } from './gameApi';

describe('baccaratbanqueApi', () => {
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
  const payload = { players: [], phase: 'banker', message: '' };

  it('reset hits /baccaratbanque/exec', async () => {
    mockFetch.mockReturnValue(ok(payload));
    const result = await baccaratbanqueApi.exec('reset');
    expect(mockFetch).toHaveBeenCalledWith('/baccaratbanque/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'reset', sessionId }),
    });
    expect(result).toEqual(payload);
  });

  // **引くと止まるは別のコマンドとして届く。** 真偽値ひとつに畳むと、
  // 付け忘れた要求が黙ってどちらかに倒れる。
  it.each([
    ['draw', 'draw'],
    ['stand', 'stand'],
    ['nextcoup', 'nextcoup'],
    ['retire', 'retire'],
  ] as const)('%s reaches the server as its own command', async (cmd, wire) => {
    mockFetch.mockReturnValue(ok(payload));
    await baccaratbanqueApi.exec(cmd);
    expect(JSON.parse(mockFetch.mock.calls[0][1].body)).toEqual({ command: wire, sessionId });
  });

  it('sends the config on reset', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await baccaratbanqueApi.exec('reset', { cpuDifficulty: 2, startChips: 5000, betAmount: 100 });
    expect(JSON.parse(mockFetch.mock.calls[0][1].body).config).toEqual({
      cpuDifficulty: 2,
      startChips: 5000,
      betAmount: 100,
    });
  });

  // 負のコントロール: 設定を渡さなければ config を積まない (サーバ側の既定を使う)。
  it('omits config entirely when none is given', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await baccaratbanqueApi.exec('reset');
    expect(JSON.parse(mockFetch.mock.calls[0][1].body)).not.toHaveProperty('config');
  });

  it('rejects on a server error', async () => {
    mockFetch.mockReturnValue(err());
    await expect(baccaratbanqueApi.exec('reset')).rejects.toThrow();
  });
});
