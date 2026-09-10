import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { sessionId } from '../gameApi';
import { tehonbikiApi } from './tehonbiki';

describe('tehonbikiApi', () => {
  const mockFetch = vi.fn();
  const payload = { phase: 0, numbers: [], betType: '', message: '' };

  beforeEach(() => {
    vi.stubGlobal('fetch', mockFetch);
    mockFetch.mockResolvedValue({ ok: true, status: 200, json: () => Promise.resolve(payload) });
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it.each([
    ['reset', undefined, { command: 'reset', sessionId }],
    ['hint', undefined, { command: 'hint', sessionId }],
    ['log', undefined, { command: 'log', sessionId }],
    ['next', undefined, { command: 'next', sessionId }],
    [
      'bet',
      { numbers: [2, 4], betType: 'double', bet: 50 },
      { command: 'bet', numbers: [2, 4], betType: 'double', bet: 50, sessionId },
    ],
  ] as [
    'reset' | 'bet' | 'next' | 'hint' | 'log',
    { numbers?: number[]; betType?: string; bet?: number } | undefined,
    Record<string, unknown>,
  ][])('%s sends its command body', async (command, params, body) => {
    await tehonbikiApi.exec(command, params);
    expect(mockFetch).toHaveBeenCalledWith('/tehonbiki/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
  });

  it('returns the response and reports HTTP errors', async () => {
    await expect(tehonbikiApi.exec('reset')).resolves.toEqual(payload);
    mockFetch.mockResolvedValue({ ok: false, status: 500, json: () => Promise.resolve(null) });
    await expect(tehonbikiApi.exec('reset')).rejects.toThrow('HTTP error: 500');
  });
});
