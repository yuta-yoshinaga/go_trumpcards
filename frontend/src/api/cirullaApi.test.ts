import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { cirullaApi, sessionId } from './gameApi';

describe('cirullaApi', () => {
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

  it('reset hits /cirulla/exec', async () => {
    mockFetch.mockReturnValue(ok(payload));
    const result = await cirullaApi.exec('reset');
    expect(mockFetch).toHaveBeenCalledWith('/cirulla/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'reset', sessionId }),
    });
    expect(result).toEqual(payload);
  });

  it('reset carries the target score', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await cirullaApi.exec('reset', { config: { cpuDifficulty: 2, targetScore: 21 } });
    expect(mockFetch).toHaveBeenCalledWith(
      '/cirulla/exec',
      expect.objectContaining({
        body: JSON.stringify({ command: 'reset', config: { cpuDifficulty: 2, targetScore: 21 }, sessionId }),
      }),
    );
  });

  // **取る札は出す札と同じ要求に乗る。** 別便にすると「出したが取っていない」
  // 盤面が生まれる。
  it('play sends the hand index and the captured group together', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await cirullaApi.exec('play', { handIndex: 0, captureIndices: [1, 2] });
    expect(mockFetch).toHaveBeenCalledWith(
      '/cirulla/exec',
      expect.objectContaining({
        body: JSON.stringify({ command: 'play', handIndex: 0, captureIndices: [1, 2], sessionId }),
      }),
    );
  });

  it('omits captureIndices when laying off, and keeps hand index 0', async () => {
    mockFetch.mockReturnValue(ok(payload));
    await cirullaApi.exec('play', { handIndex: 0 });
    expect(mockFetch).toHaveBeenCalledWith(
      '/cirulla/exec',
      expect.objectContaining({ body: JSON.stringify({ command: 'play', handIndex: 0, sessionId }) }),
    );
  });

  it.each(['nextround', 'hint', 'log'] as const)('%s carries no extra fields', async (command) => {
    mockFetch.mockReturnValue(ok(payload));
    await cirullaApi.exec(command);
    expect(mockFetch).toHaveBeenCalledWith(
      '/cirulla/exec',
      expect.objectContaining({ body: JSON.stringify({ command, sessionId }) }),
    );
  });

  it('throws when the server fails', async () => {
    mockFetch.mockReturnValue(err());
    await expect(cirullaApi.exec('nextround')).rejects.toThrow();
  });
});
