import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { doubleexposureApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BlackJackResponse } from '../types/card';
import { DoubleExposurePage } from './DoubleExposurePage';

vi.mock('../api/gameApi', () => ({
  doubleexposureApi: { exec: vi.fn() },
  actionLogApi: { doubleexposure: vi.fn() },
}));

const mockExec = vi.mocked(doubleexposureApi.exec);

const betPhaseState: BlackJackResponse = {
  dealer: { chips: 1000 },
  player: { chips: 1000 },
  phase: 1,
  currentHandIdx: 0,
  insuranceBet: 0,
  insuranceAvailable: false,
  message: '',
  hintEnabled: false,
  suggestedAction: 0,
  deckCount: 1,
  dealerHitsSoft17: false,
  countingEnabled: false,
  cpuPlayerCount: 0,
  runningCount: 0,
  trueCount: 0,
  perfectPairsBet: 0,
  twentyOnePlus3Bet: 0,
  doubleAfterSplit: true,
  countingSystem: 0,
  deckPenetration: 75,
  multiHandCount: 0,
  surrenderRule: 0,
};

beforeEach(() => {
  mockExec.mockResolvedValue(betPhaseState);
});

describe('DoubleExposurePage', () => {
  it('resets via the Double Exposure API on mount', async () => {
    renderWithProviders(<DoubleExposurePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
  });

  it('shows the reset confirmation dialog before starting the next game', async () => {
    const endState: BlackJackResponse = { ...betPhaseState, phase: 5, message: 'You are the winner.' };
    mockExec.mockResolvedValue(endState);
    renderWithProviders(<DoubleExposurePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    expect(screen.getByText('本当にゲームをリセットしますか？')).toBeInTheDocument();
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('cancels the reset confirmation without calling the API', async () => {
    const endState: BlackJackResponse = { ...betPhaseState, phase: 5, message: 'You are the winner.' };
    mockExec.mockResolvedValue(endState);
    renderWithProviders(<DoubleExposurePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    await waitFor(() => expect(screen.queryByText('本当にゲームをリセットしますか？')).not.toBeInTheDocument());
    expect(mockExec).not.toHaveBeenCalled();
  });
});
