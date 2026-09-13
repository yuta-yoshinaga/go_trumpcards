import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { sessionId } from '../gameExec';
import { trenteetquaranteApi } from './trenteetquarante';

describe('trenteetquaranteApi', () => {
  const mockFetch = vi.fn();

  beforeEach(() => {
    vi.stubGlobal('fetch', mockFetch);
  });

  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  const ok = () => Promise.resolve({ ok: true, status: 200, json: () => Promise.resolve({}) });

  it('includes config in the request body', async () => {
    mockFetch.mockReturnValue(ok());

    await trenteetquaranteApi.exec('reset', undefined, undefined, { defaultBet: 3 });

    expect(mockFetch).toHaveBeenCalledWith('/trenteetquarante/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'reset', config: { defaultBet: 3 }, sessionId }),
    });
  });

  it('does not add config when it is omitted', async () => {
    mockFetch.mockReturnValue(ok());

    await trenteetquaranteApi.exec('bet', 1, 100);

    expect(mockFetch).toHaveBeenCalledWith('/trenteetquarante/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ command: 'bet', bet: 1, stake: 100, sessionId }),
    });
  });
});
