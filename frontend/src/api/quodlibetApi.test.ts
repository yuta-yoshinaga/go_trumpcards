import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { quodlibetApi, sessionId } from './gameApi';

describe('quodlibetApi', () => {
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

  it('reset hits /quodlibet/exec', async () => {
    mockFetch.mockReturnValue(ok(payload));
    const result = await quodlibetApi.exec('reset');
    expect(mockFetch).toHaveBeenCalledWith('/quodlibet/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'reset', sessionId }),
    });
    expect(result).toEqual(payload);
  });

  it('reset carries the config', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await quodlibetApi.exec('reset', { config: { cpuDifficulty: 2, autoSelectContract: true } });
    expect(mockFetch).toHaveBeenCalledWith(
      '/quodlibet/exec',
      expect.objectContaining({
        body: JSON.stringify({
          command: 'reset',
          config: { cpuDifficulty: 2, autoSelectContract: true },
          sessionId,
        }),
      }),
    );
  });

  it('contract sends the chosen contract', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await quodlibetApi.exec('contract', { contract: 3 });
    expect(mockFetch).toHaveBeenCalledWith(
      '/quodlibet/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'contract', contract: 3, sessionId }) }),
    );
  });

  // 0 は正当な種目番号 (プラス)。undefined と同じ扱いにすると選べない。
  it('keeps contract 0 in the body', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await quodlibetApi.exec('contract', { contract: 0 });
    expect(mockFetch).toHaveBeenCalledWith(
      '/quodlibet/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'contract', contract: 0, sessionId }) }),
    );
  });

  it('play sends the card index, including 0', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await quodlibetApi.exec('play', { cardIndex: 0 });
    expect(mockFetch).toHaveBeenCalledWith(
      '/quodlibet/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'play', cardIndex: 0, sessionId }) }),
    );
  });

  it.each(['pass', 'nextdeal', 'hint', 'log'] as const)('%s carries no extra fields', async (command) => {
    mockFetch.mockReturnValue(ok(payload));
    await quodlibetApi.exec(command);
    expect(mockFetch).toHaveBeenCalledWith(
      '/quodlibet/exec',
      expect.objectContaining({ body: JSON.stringify({ command, sessionId }) }),
    );
  });

  it('throws when the server fails', async () => {
    mockFetch.mockReturnValue(err());
    await expect(quodlibetApi.exec('nextdeal')).rejects.toThrow();
  });
});
