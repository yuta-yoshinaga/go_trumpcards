import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { teenPattiApi } from '../api/gameApi';
import { makeTeenPattiState } from '../test/stateFactories';
import { DEFAULT_TEEN_PATTI_CONFIG, useTeenPattiGame } from './useTeenPattiGame';

vi.mock('../api/gameApi', () => ({
  teenPattiApi: { exec: vi.fn() },
  actionLogApi: { teenpatti: vi.fn() },
}));

const mockExec = vi.mocked(teenPattiApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeTeenPattiState());
});

describe('useTeenPattiGame', () => {
  it('reset dispatches reset with the default config', async () => {
    const { result } = renderHook(() => useTeenPattiGame(), { wrapper: createWrapper() });
    act(() => result.current.reset());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: DEFAULT_TEEN_PATTI_CONFIG }));
  });

  it('handleSee dispatches see', async () => {
    const { result } = renderHook(() => useTeenPattiGame(), { wrapper: createWrapper() });
    act(() => result.current.handleSee());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('see'));
  });

  it('handleBet dispatches bet', async () => {
    const { result } = renderHook(() => useTeenPattiGame(), { wrapper: createWrapper() });
    act(() => result.current.handleBet());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet'));
  });

  it('handleRaise dispatches raise with the stake', async () => {
    const { result } = renderHook(() => useTeenPattiGame(), { wrapper: createWrapper() });
    act(() => result.current.handleRaise(6));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('raise', { raiseStake: 6 }));
  });

  it('handleFold dispatches fold', async () => {
    const { result } = renderHook(() => useTeenPattiGame(), { wrapper: createWrapper() });
    act(() => result.current.handleFold());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('fold'));
  });

  it('handleShow dispatches show', async () => {
    const { result } = renderHook(() => useTeenPattiGame(), { wrapper: createWrapper() });
    act(() => result.current.handleShow());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('show'));
  });

  it('handleSideShow dispatches sideshow', async () => {
    const { result } = renderHook(() => useTeenPattiGame(), { wrapper: createWrapper() });
    act(() => result.current.handleSideShow());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('sideshow'));
  });

  it('handleRespondSideShow dispatches respond with the accept flag', async () => {
    const { result } = renderHook(() => useTeenPattiGame(), { wrapper: createWrapper() });
    act(() => result.current.handleRespondSideShow(true));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('respond', { accept: true }));
    act(() => result.current.handleRespondSideShow(false));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('respond', { accept: false }));
  });

  it('handleNextRound dispatches next', async () => {
    const { result } = renderHook(() => useTeenPattiGame(), { wrapper: createWrapper() });
    act(() => result.current.handleNextRound());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('handleConfigChange updates the config used by reset', async () => {
    const { result } = renderHook(() => useTeenPattiGame(), { wrapper: createWrapper() });
    act(() => result.current.handleConfigChange('ante', '5'));
    act(() => result.current.reset());
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { ...DEFAULT_TEEN_PATTI_CONFIG, ante: 5 } }),
    );
  });
});
