import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { scoponeApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { makeScoponeState } from '../test/stateFactories';
import { useScoponeGame } from './useScoponeGame';

vi.mock('../api/gameApi', () => ({
  scoponeApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(scoponeApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeScoponeState());
});

describe('useScoponeGame', () => {
  it('calls reset on mount with the short "r" command', async () => {
    renderHook(() => useScoponeGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('r'));
  });

  it('returns state after mount', async () => {
    const { result } = renderHook(() => useScoponeGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
  });

  it('toggleTable adds and removes a table index', async () => {
    const { result } = renderHook(() => useScoponeGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => result.current.toggleTable(1));
    expect(result.current.tableIndices).toEqual([1]);
    act(() => result.current.toggleTable(1));
    expect(result.current.tableIndices).toEqual([]);
  });

  it('play dispatches "p" with sorted table indices', async () => {
    const { result } = renderHook(() => useScoponeGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => result.current.setHandIndex(0));
    act(() => result.current.toggleTable(2));
    act(() => result.current.toggleTable(1));

    mockExec.mockClear();
    act(() => result.current.play());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('p', { handIndex: 0, tableIndices: [1, 2] }));
  });

  it('play is a no-op when no hand card is selected', async () => {
    const { result } = renderHook(() => useScoponeGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    act(() => result.current.play());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handleNextRound dispatches "n"', async () => {
    const { result } = renderHook(() => useScoponeGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    act(() => result.current.handleNextRound());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('n'));
  });

  it('handleConfigChange and reset send the config', async () => {
    const { result } = renderHook(() => useScoponeGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => result.current.handleConfigChange('cpuDifficulty', 2));
    mockExec.mockClear();
    act(() => result.current.handleResetWithConfig());
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith(
        'r',
        expect.objectContaining({ config: expect.objectContaining({ cpuDifficulty: 2 }) }),
      ),
    );
  });

  it('clearSelection resets hand and table selections', async () => {
    const { result } = renderHook(() => useScoponeGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => result.current.setHandIndex(1));
    act(() => result.current.toggleTable(0));
    act(() => result.current.clearSelection());
    expect(result.current.handIndex).toBeNull();
    expect(result.current.tableIndices).toEqual([]);
  });
});
