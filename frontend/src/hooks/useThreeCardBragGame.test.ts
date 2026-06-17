import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { threeCardBragApi } from '../api/gameApi';
import { makeThreeCardBragState } from '../test/stateFactories';
import { DEFAULT_THREE_CARD_BRAG_CONFIG, useThreeCardBragGame } from './useThreeCardBragGame';

vi.mock('../api/gameApi', () => ({
  threeCardBragApi: { exec: vi.fn() },
  actionLogApi: { threecardbrag: vi.fn() },
}));

const mockExec = vi.mocked(threeCardBragApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeThreeCardBragState());
});

describe('useThreeCardBragGame', () => {
  it('reset dispatches reset with the default config', async () => {
    const { result } = renderHook(() => useThreeCardBragGame(), { wrapper: createWrapper() });
    act(() => result.current.reset());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: DEFAULT_THREE_CARD_BRAG_CONFIG }));
  });

  it('handleSee dispatches see', async () => {
    const { result } = renderHook(() => useThreeCardBragGame(), { wrapper: createWrapper() });
    act(() => result.current.handleSee());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('see'));
  });

  it('handleBet dispatches bet', async () => {
    const { result } = renderHook(() => useThreeCardBragGame(), { wrapper: createWrapper() });
    act(() => result.current.handleBet());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet'));
  });

  it('handleRaise dispatches raise with the stake', async () => {
    const { result } = renderHook(() => useThreeCardBragGame(), { wrapper: createWrapper() });
    act(() => result.current.handleRaise(6));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('raise', { raiseStake: 6 }));
  });

  it('handleFold dispatches fold', async () => {
    const { result } = renderHook(() => useThreeCardBragGame(), { wrapper: createWrapper() });
    act(() => result.current.handleFold());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('fold'));
  });

  it('handleShow dispatches show', async () => {
    const { result } = renderHook(() => useThreeCardBragGame(), { wrapper: createWrapper() });
    act(() => result.current.handleShow());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('show'));
  });

  it('handleNextRound dispatches next', async () => {
    const { result } = renderHook(() => useThreeCardBragGame(), { wrapper: createWrapper() });
    act(() => result.current.handleNextRound());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('handleConfigChange updates the config used by reset', async () => {
    const { result } = renderHook(() => useThreeCardBragGame(), { wrapper: createWrapper() });
    act(() => result.current.handleConfigChange('ante', '5'));
    act(() => result.current.reset());
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { ...DEFAULT_THREE_CARD_BRAG_CONFIG, ante: 5 } }),
    );
  });
});
