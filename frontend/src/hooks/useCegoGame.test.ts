import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { cegoApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { useCegoGame } from './useCegoGame';

vi.mock('../api/gameApi', () => ({ cegoApi: { exec: vi.fn() }, actionLogApi: { cego: vi.fn() } }));
const mockExec = vi.mocked(cegoApi.exec);
const wrapper = ({ children }: { children: ReactNode }) =>
  createElement(
    QueryClientProvider,
    { client: new QueryClient({ defaultOptions: { mutations: { retry: false } } }) },
    children,
  );

describe('useCegoGame config validation', () => {
  beforeEach(() => mockExec.mockReset());
  it('does not send a target deal count below one or a fraction', async () => {
    const { result } = renderHook(() => useCegoGame(), { wrapper });
    act(() => result.current.handleConfigChange('targetDeals', '0'));
    act(() => result.current.reset());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
    act(() => result.current.handleConfigChange('targetDeals', '1.5'));
    act(() => result.current.reset());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });
});
