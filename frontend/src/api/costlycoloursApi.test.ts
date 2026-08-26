import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { costlycoloursApi, sessionId } from './gameApi';

describe('costlycoloursApi', () => {
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
  const payload = { players: [], phase: 'mog', message: '' };

  it('reset hits /costlycolours/exec', async () => {
    mockFetch.mockReturnValue(ok(payload));
    const result = await costlycoloursApi.exec('reset');
    expect(mockFetch).toHaveBeenCalledWith('/costlycolours/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'reset', sessionId }),
    });
    expect(result).toEqual(payload);
  });

  // **交換の可否は必ず明示する。** 断ると相手に 1 点入るので、既定に落とさない。
  it('sends the exchange decision explicitly, both ways', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await costlycoloursApi.exec('mog', { accept: true });
    expect(JSON.parse(mockFetch.mock.calls[0][1].body)).toMatchObject({ command: 'mog', accept: true });

    mockFetch.mockClear();
    await costlycoloursApi.exec('mog', { accept: false });
    expect(JSON.parse(mockFetch.mock.calls[0][1].body)).toMatchObject({ command: 'mog', accept: false });
  });

  it('sends the hand index with play', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await costlycoloursApi.exec('play', { handIndex: 2 });
    expect(JSON.parse(mockFetch.mock.calls[0][1].body)).toMatchObject({ command: 'play', handIndex: 2 });
  });

  it('sends the config on reset', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await costlycoloursApi.exec('reset', { config: { cpuDifficulty: 2, targetScore: 121 } });
    expect(JSON.parse(mockFetch.mock.calls[0][1].body).config).toEqual({ cpuDifficulty: 2, targetScore: 121 });
  });

  it('rejects on a server error', async () => {
    mockFetch.mockReturnValue(err());
    await expect(costlycoloursApi.exec('reset')).rejects.toThrow();
  });
});
