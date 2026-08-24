import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { germansoloApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { makeGermanSoloState } from '../test/stateFactories';
import { DEFAULT_GERMAN_SOLO_CONFIG, useGermanSoloGame } from './useGermanSoloGame';

vi.mock('../api/gameApi', () => ({
  germansoloApi: { exec: vi.fn() },
  actionLogApi: { germansolo: vi.fn() },
}));

const mockExec = vi.mocked(germansoloApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeGermanSoloState());
});

describe('useGermanSoloGame', () => {
  it('reset dispatches reset with the default config', async () => {
    const { result } = renderHook(() => useGermanSoloGame(), { wrapper: createWrapper() });
    act(() => result.current.reset());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: DEFAULT_GERMAN_SOLO_CONFIG }));
  });

  it('handleBid dispatches a pass with no trump', async () => {
    const { result } = renderHook(() => useGermanSoloGame(), { wrapper: createWrapper() });
    act(() => result.current.handleBid(0));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 0, trumpSuit: undefined }));
  });

  it('handleBid dispatches a Frage with the chosen trump suit', async () => {
    const { result } = renderHook(() => useGermanSoloGame(), { wrapper: createWrapper() });
    act(() => result.current.handleBid(2, 3));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 2, trumpSuit: 3 }));
  });

  // **エース呼びを抜ける唯一の操作。** これが無いと Frage 落札の直後に
  // play が「フェーズが違う」で弾かれ続ける。
  it('handleCallAce dispatches the ace call with the chosen suit', async () => {
    const { result } = renderHook(() => useGermanSoloGame(), { wrapper: createWrapper() });
    act(() => result.current.handleCallAce(2));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('ace', { aceSuit: 2 }));
  });

  it('handlePlay does nothing without exactly one selected card', async () => {
    const { result } = renderHook(() => useGermanSoloGame(), { wrapper: createWrapper() });
    act(() => result.current.handlePlay());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handlePlay dispatches play for the single selected card', async () => {
    const { result } = renderHook(() => useGermanSoloGame(), { wrapper: createWrapper() });
    act(() => result.current.toggleCard(2));
    act(() => result.current.handlePlay());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 2 }));
  });

  it('handleNextTrick and handleNextRound dispatch their commands', async () => {
    const { result } = renderHook(() => useGermanSoloGame(), { wrapper: createWrapper() });
    act(() => result.current.handleNextTrick());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
    act(() => result.current.handleNextRound());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('handleConfigChange updates the config used by reset', async () => {
    const { result } = renderHook(() => useGermanSoloGame(), { wrapper: createWrapper() });
    act(() => result.current.handleConfigChange('targetRounds', '7'));
    act(() => result.current.reset());
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { ...DEFAULT_GERMAN_SOLO_CONFIG, targetRounds: 7 } }),
    );
  });
});
