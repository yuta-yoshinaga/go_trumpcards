import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { blackholeApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BlackHoleResponse, Card } from '../types/card';
import { BlackHolePage } from './BlackHolePage';

vi.mock('../api/gameApi', () => ({
  blackholeApi: { exec: vi.fn() },
  actionLogApi: { blackhole: vi.fn() },
}));

const mockExec = vi.mocked(blackholeApi.exec);

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<BlackHoleResponse> = {}): BlackHoleResponse {
  const fans: Card[][] = Array.from({ length: 17 }, () => []);
  fans[0] = [card('HEART', 9), card('CLUB', 6)];
  fans[1] = [card('DIAMOND', 10)];
  return {
    fans,
    blackHole: [card('SPADE', 1)],
    phase: 0,
    moveCount: 0,
    canUndo: false,
    isStalemate: false,
    message: '',
    ...overrides,
  };
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeState());
});

describe('BlackHolePage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<BlackHolePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<BlackHolePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders the 17 fans', async () => {
    renderWithProviders(<BlackHolePage />);
    await waitFor(() => expect(screen.getByTestId('fan-0')).toBeInTheDocument());
    expect(screen.getByTestId('fan-16')).toBeInTheDocument();
  });

  it('plays a fan top into the black hole', async () => {
    renderWithProviders(<BlackHolePage />);
    const top = await screen.findByTestId('card-0-1');
    mockExec.mockClear();
    fireEvent.click(top);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('mb', { fan: 0 }));
  });

  it('does not play a non-top card', async () => {
    renderWithProviders(<BlackHolePage />);
    const buried = await screen.findByTestId('card-0-0');
    mockExec.mockClear();
    fireEvent.click(buried);
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('undoes, hints and gives up', async () => {
    mockExec.mockResolvedValue(makeState({ canUndo: true }));
    renderWithProviders(<BlackHolePage />);
    await screen.findByTestId('hint-button');
    for (const [testid, cmd] of [
      ['undo-button', 'u'],
      ['hint-button', 'hint'],
      ['giveup-button', 'g'],
    ] as const) {
      mockExec.mockClear();
      fireEvent.click(screen.getByTestId(testid));
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith(cmd));
    }
  });

  it('hides controls at game over', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 2 }));
    renderWithProviders(<BlackHolePage />);
    await waitFor(() => expect(screen.getByTestId('fan-0')).toBeInTheDocument());
    expect(screen.queryByTestId('hint-button')).not.toBeInTheDocument();
  });
});
