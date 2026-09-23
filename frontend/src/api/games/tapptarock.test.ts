import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { sessionId } from '../gameApi';
import { tapptarockApi } from './tapptarock';

describe('tapptarockApi', () => {
  const mockFetch = vi.fn();
  beforeEach(() => {
    vi.stubGlobal('fetch', mockFetch);
    mockFetch.mockResolvedValue({ ok: true, status: 200, json: () => Promise.resolve({ phase: 0 }) });
  });
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.clearAllMocks();
  });

  it.each([
    ['reset', undefined, { command: 'reset', sessionId }],
    ['pass', undefined, { command: 'pass', sessionId }],
    ['next', undefined, { command: 'next', sessionId }],
    ['nextround', undefined, { command: 'nextround', sessionId }],
    ['hint', undefined, { command: 'hint', sessionId }],
    ['log', undefined, { command: 'log', sessionId }],
    ['bid', { bid: 'solo' }, { command: 'bid', bid: 'solo', sessionId }],
    [
      'discard',
      { cardIndices: [0, 2, 4, 6, 8, 10] },
      { command: 'discard', cardIndices: [0, 2, 4, 6, 8, 10], sessionId },
    ],
    ['play', { cardIndex: 7 }, { command: 'play', cardIndex: 7, sessionId }],
    [
      'reset',
      { config: { cpuDifficulty: 2, targetDeals: 3 } },
      { command: 'reset', config: { cpuDifficulty: 2, targetDeals: 3 }, sessionId },
    ],
  ] as const)('%s sends its command body', async (command, params, body) => {
    const normalized = params && 'cardIndices' in params ? { ...params, cardIndices: [...params.cardIndices] } : params;
    await tapptarockApi.exec(command, normalized);
    expect(mockFetch).toHaveBeenCalledWith('/tapptarock/exec', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
  });
});
