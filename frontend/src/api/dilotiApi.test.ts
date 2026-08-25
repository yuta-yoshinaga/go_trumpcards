import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { dilotiApi, sessionId } from './gameApi';

describe('dilotiApi', () => {
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
  const payload = { players: [], phase: 'play', message: '' };

  it('reset hits /diloti/exec', async () => {
    mockFetch.mockReturnValue(ok(payload));
    const result = await dilotiApi.exec('reset');
    expect(mockFetch).toHaveBeenCalledWith('/diloti/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'reset', sessionId }),
    });
    expect(result).toEqual(payload);
  });

  // **取る対象は同じ 1 回の要求に乗る。** 別便にすると「出したが取っていない」
  // 盤面が生まれる。
  it('sends the capture targets with the card played', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await dilotiApi.exec('play', { handIndex: 1, action: 'capture', tableIndices: [0, 2], declIndices: [1] });
    const body = JSON.parse(mockFetch.mock.calls[0][1].body);
    expect(body).toMatchObject({
      command: 'play',
      handIndex: 1,
      action: 'capture',
      tableIndices: [0, 2],
      declIndices: [1],
    });
  });

  it('sends the declared value', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await dilotiApi.exec('play', { handIndex: 0, action: 'declare', tableIndices: [1], declValue: 8 });
    const body = JSON.parse(mockFetch.mock.calls[0][1].body);
    expect(body).toMatchObject({ command: 'play', handIndex: 0, action: 'declare', declValue: 8 });
  });

  it('sends the config on reset', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await dilotiApi.exec('reset', { config: { cpuDifficulty: 2, targetScore: 101 } });
    const body = JSON.parse(mockFetch.mock.calls[0][1].body);
    expect(body.config).toEqual({ cpuDifficulty: 2, targetScore: 101 });
  });

  it('rejects on a server error', async () => {
    mockFetch.mockReturnValue(err());
    await expect(dilotiApi.exec('reset')).rejects.toThrow();
  });
});
