import { screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { callBreakApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeCallBreakState } from '../test/stateFactories';
import { CallBreakPage } from './CallBreakPage';

vi.mock('../api/gameApi', () => ({
  callBreakApi: { exec: vi.fn() },
  actionLogApi: { callbreak: vi.fn() },
}));

const mockExec = vi.mocked(callBreakApi.exec);

const playPhaseState = makeCallBreakState();

const bidPhaseState = makeCallBreakState({
  phase: 0,
  bidPlayerIdx: 0,
  players: makeCallBreakState().players.map((p) => ({ ...p, bid: -1 })),
});

const gameEndState = makeCallBreakState({
  phase: 4,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'Game end!',
});

beforeEach(() => {
  mockExec.mockResolvedValue(playPhaseState);
});

describe('CallBreakPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<CallBreakPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with default config', async () => {
    renderWithProviders(<CallBreakPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        maxRounds: 5,
      }),
    );
  });

  it('renders bid phase elements', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<CallBreakPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /ビッド/ })).toBeInTheDocument();
    });
  });

  it('shows decimal score in the score table (cumulativeScore 41 → 4.1)', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<CallBreakPage />);
    await waitFor(() => {
      // CPU 1 has cumulativeScore=41 in the factory; should render as 4.1
      expect(screen.getAllByText(/4\.1/).length).toBeGreaterThan(0);
    });
  });

  it('renders game end winner banner', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<CallBreakPage />);
    await waitFor(() => {
      expect(screen.getByText(/Game end!/i)).toBeInTheDocument();
    });
  });
});
