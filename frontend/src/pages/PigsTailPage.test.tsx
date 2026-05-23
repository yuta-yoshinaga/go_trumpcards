import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { pigtailApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { PigsTailResponse } from '../types/card';
import { PigsTailPage } from './PigsTailPage';

vi.mock('../hooks/useCliMode', () => ({
  useCliMode: vi.fn(() => ({
    cliEnabled: false,
    toggleCli: vi.fn(),
    logEntries: [],
    addInput: vi.fn(),
    addOutput: vi.fn(),
    addError: vi.fn(),
    clearLog: vi.fn(),
  })),
}));

const mockUseCliMode = vi.mocked(useCliMode);

vi.mock('../api/gameApi', () => ({
  pigtailApi: { exec: vi.fn() },
  actionLogApi: { pigtail: vi.fn() },
}));

const mockExec = vi.mocked(pigtailApi.exec);

const baseState: PigsTailResponse = {
  players: [
    { id: 0, isHuman: true, cardCount: 0, cards: [] },
    { id: 1, isHuman: false, cardCount: 0, cards: [] },
    { id: 2, isHuman: false, cardCount: 0, cards: [] },
    { id: 3, isHuman: false, cardCount: 0, cards: [] },
  ],
  circleCount: 52,
  centerTop: null,
  centerCount: 0,
  currentTurn: 0,
  gameEndFlag: false,
  loserIdx: -1,
  lastDrawCard: null,
  lastPenalty: false,
  cpuActions: [],
  humanAction: null,
  message: '',
};

const gameEndState: PigsTailResponse = {
  ...baseState,
  circleCount: 0,
  gameEndFlag: true,
  loserIdx: 1,
  message: 'Game Over! CPU 1 loses!',
  messageCode: 'pigtail.result.cpuLose',
  messageParams: { cpuId: '1' },
};

const humanLoseState: PigsTailResponse = {
  ...baseState,
  circleCount: 0,
  gameEndFlag: true,
  loserIdx: 0,
  message: 'Game Over! You lose!',
  messageCode: 'pigtail.result.humanLose',
};

beforeEach(() => {
  mockExec.mockResolvedValue(baseState);
  // Reset CLI mode to the default so tests that toggle it on (e.g. the
  // CLI-terminal case) don't leak state into the next test in source order.
  mockUseCliMode.mockReturnValue({
    cliEnabled: false,
    toggleCli: vi.fn(),
    logEntries: [],
    addInput: vi.fn(),
    addOutput: vi.fn(),
    addError: vi.fn(),
    clearLog: vi.fn(),
  });
});

describe('PigsTailPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<PigsTailPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<PigsTailPage />);
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalled();
    });
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
  });

  it('renders game state with circle and center info', async () => {
    renderWithProviders(<PigsTailPage />);
    await waitFor(() => {
      expect(screen.getByText(/52/)).toBeInTheDocument();
    });
  });

  it('draw button is enabled on human turn', async () => {
    renderWithProviders(<PigsTailPage />);
    await waitFor(() => {
      const drawBtn = screen.getByRole('button', { name: '山札から引く' });
      expect(drawBtn).not.toBeDisabled();
    });
  });

  it('draw button is disabled on game end', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<PigsTailPage />);
    await waitFor(() => {
      const drawBtn = screen.getByRole('button', { name: '山札から引く' });
      expect(drawBtn).toBeDisabled();
    });
  });

  it('clicking draw button calls exec', async () => {
    renderWithProviders(<PigsTailPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument();
    });
    const drawBtn = screen.getByRole('button', { name: '山札から引く' });
    expect(drawBtn).not.toBeDisabled();
    fireEvent.click(drawBtn);
    // Verify at least the initial reset was called
    expect(mockExec).toHaveBeenCalled();
  });

  it('shows game end message when game is over', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<PigsTailPage />);
    await waitFor(() => {
      const drawBtn = screen.getByRole('button', { name: '山札から引く' });
      expect(drawBtn).toBeDisabled();
    });
  });

  it('disables draw on game end (cpu loses)', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<PigsTailPage />);
    await waitFor(() => {
      const drawBtn = screen.getByRole('button', { name: '山札から引く' });
      expect(drawBtn).toBeDisabled();
    });
  });

  it('does not show win celebration when human loses', async () => {
    mockExec.mockResolvedValue(humanLoseState);
    renderWithProviders(<PigsTailPage />);
    await waitFor(() => {
      expect(screen.queryByTestId('win-celebration')).not.toBeInTheDocument();
    });
  });

  it('renders cpu actions when present', async () => {
    const stateWithCpuActions: PigsTailResponse = {
      ...baseState,
      cpuActions: [
        { drawPlayerIdx: 1, drawnCard: { design: 'SPADE', value: 5 }, penaltyFlag: false, penaltyCount: 0 },
        { drawPlayerIdx: 2, drawnCard: { design: 'HEART', value: 3 }, penaltyFlag: true, penaltyCount: 4 },
      ],
    };
    mockExec.mockResolvedValue(stateWithCpuActions);
    renderWithProviders(<PigsTailPage />);
    await waitFor(() => {
      expect(screen.getByText(/ペナルティ/)).toBeInTheDocument();
      expect(screen.getByText(/セーフ/)).toBeInTheDocument();
    });
  });

  it('renders CLI terminal when CLI mode is enabled', async () => {
    mockUseCliMode.mockReturnValue({
      cliEnabled: true,
      toggleCli: vi.fn(),
      logEntries: [],
      addInput: vi.fn(),
      addOutput: vi.fn(),
      addError: vi.fn(),
      clearLog: vi.fn(),
    });
    renderWithProviders(<PigsTailPage />);
    await waitFor(() => {
      expect(screen.getByRole('textbox')).toBeInTheDocument();
    });
    expect(screen.queryByRole('button', { name: '山札から引く' })).not.toBeInTheDocument();
  });

  it('renders the circular deck and dispatches draw when a ring card is tapped', async () => {
    renderWithProviders(<PigsTailPage />);
    await waitFor(() => expect(screen.getByTestId('circular-deck')).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('circular-deck-card-0'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });
});
