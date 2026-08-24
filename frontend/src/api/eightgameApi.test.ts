import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { eightGameApi, sessionId } from './gameApi';

describe('eightGameApi', () => {
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

  const payload = { seats: [], phase: 0, message: '' };

  // **叩く先が /horse ではない。** 同じレスポンス型を共有しているので、
  // ルートを間違えても型では気付けない。
  it('reset hits /eightgame/exec with command=reset', async () => {
    mockFetch.mockReturnValue(ok(payload));
    const result = await eightGameApi.exec('reset');
    expect(mockFetch).toHaveBeenCalledWith('/eightgame/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'reset', sessionId }),
    });
    expect(result).toEqual(payload);
  });

  it('reset with config includes the config object', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await eightGameApi.exec('reset', { config: { seats: 4, handsPerDiscipline: 3 } });
    expect(mockFetch).toHaveBeenCalledWith(
      '/eightgame/exec',
      expect.objectContaining({
        body: JSON.stringify({ command: 'reset', config: { seats: 4, handsPerDiscipline: 3 }, sessionId }),
      }),
    );
  });

  it('a raise sends the action and the amount', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await eightGameApi.exec('action', { action: 'raise', amount: 40 });
    expect(mockFetch).toHaveBeenCalledWith(
      '/eightgame/exec',
      expect.objectContaining({
        body: JSON.stringify({ command: 'action', action: 'raise', amount: 40, sessionId }),
      }),
    );
  });

  // **引き直しはこのゲームにしかない命令。** 0 始まりの位置をそのまま送る。
  it('draw sends the card indices', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await eightGameApi.exec('draw', { cardIndices: [0, 2] });
    expect(mockFetch).toHaveBeenCalledWith(
      '/eightgame/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'draw', cardIndices: [0, 2], sessionId }) }),
    );
  });

  // 空のまま送るのがスタンドパット。省略すると「引かない」が打てない。
  it('draw with an empty list stands pat', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await eightGameApi.exec('draw', { cardIndices: [] });
    expect(mockFetch).toHaveBeenCalledWith(
      '/eightgame/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'draw', cardIndices: [], sessionId }) }),
    );
  });

  it('next and hint send their bare commands', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await eightGameApi.exec('next');
    expect(mockFetch).toHaveBeenCalledWith(
      '/eightgame/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'next', sessionId }) }),
    );
    await eightGameApi.exec('hint');
    expect(mockFetch).toHaveBeenCalledWith(
      '/eightgame/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'hint', sessionId }) }),
    );
  });

  it('throws on HTTP error', async () => {
    mockFetch.mockReturnValue(err());
    await expect(eightGameApi.exec('reset')).rejects.toThrow('HTTP error: 500');
  });
});
