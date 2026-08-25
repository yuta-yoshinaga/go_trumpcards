import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { dehlaPakadApi, sessionId } from './gameApi';

describe('dehlaPakadApi', () => {
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

  it('reset hits /dehlapakad/exec', async () => {
    mockFetch.mockReturnValue(ok(payload));
    const result = await dehlaPakadApi.exec('reset');
    expect(mockFetch).toHaveBeenCalledWith('/dehlapakad/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'reset', sessionId }),
    });
    expect(result).toEqual(payload);
  });

  it('reset carries the match length', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await dehlaPakadApi.exec('reset', { config: { cpuDifficulty: 2, targetKots: 5 } });
    expect(mockFetch).toHaveBeenCalledWith(
      '/dehlapakad/exec',
      expect.objectContaining({
        body: JSON.stringify({ command: 'reset', config: { cpuDifficulty: 2, targetKots: 5 }, sessionId }),
      }),
    );
  });

  it('trump sends the called suit', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await dehlaPakadApi.exec('trump', { trumpSuit: 3 });
    expect(mockFetch).toHaveBeenCalledWith(
      '/dehlapakad/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'trump', trumpSuit: 3, sessionId }) }),
    );
  });

  // 0 は正当な札番号。undefined と同じ扱いにすると先頭の札が出せない。
  it('keeps card index 0 in the body', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await dehlaPakadApi.exec('play', { cardIndex: 0 });
    expect(mockFetch).toHaveBeenCalledWith(
      '/dehlapakad/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'play', cardIndex: 0, sessionId }) }),
    );
  });

  it.each(['nexthand', 'hint', 'log'] as const)('%s carries no extra fields', async (command) => {
    mockFetch.mockReturnValue(ok(payload));
    await dehlaPakadApi.exec(command);
    expect(mockFetch).toHaveBeenCalledWith(
      '/dehlapakad/exec',
      expect.objectContaining({ body: JSON.stringify({ command, sessionId }) }),
    );
  });

  it('throws when the server fails', async () => {
    mockFetch.mockReturnValue(err());
    await expect(dehlaPakadApi.exec('nexthand')).rejects.toThrow();
  });
});
