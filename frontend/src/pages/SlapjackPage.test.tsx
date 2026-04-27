import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { slapjackApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { SlapjackResponse } from '../types/card';
import { SlapjackPhase } from '../types/phases';
import { SlapjackPage } from './SlapjackPage';

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
  slapjackApi: { exec: vi.fn() },
  actionLogApi: { slapjack: vi.fn() },
}));

const mockExec = vi.mocked(slapjackApi.exec);

const baseState: SlapjackResponse = {
  phase: SlapjackPhase.PLAY,
  gameEndFlag: false,
  winnerIdx: -1,
  currentTurnIdx: 0,
  isHumanTurn: true,
  isTopJack: false,
  centerPileSize: 0,
  topCard: null,
  players: [
    { name: 'You', isHuman: true, stockSize: 26 },
    { name: 'CPU', isHuman: false, stockSize: 26 },
  ],
  cpuDifficulty: 1,
  pendingKind: 0,
  pendingDeadlineMs: 0,
  lastEventKind: 0,
  lastEventPlayerIdx: 0,
  message: '',
};

const jackOnTopState: SlapjackResponse = {
  ...baseState,
  isTopJack: true,
  centerPileSize: 1,
  topCard: { design: 'SPADE', value: 11 },
};

const gameEndState: SlapjackResponse = {
  ...baseState,
  phase: SlapjackPhase.GAME_END,
  gameEndFlag: true,
  winnerIdx: 0,
  players: [
    { name: 'You', isHuman: true, stockSize: 52 },
    { name: 'CPU', isHuman: false, stockSize: 0 },
  ],
};

beforeEach(() => {
  mockExec.mockResolvedValue(baseState);
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

describe('SlapjackPage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<SlapjackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders stock counts after state loads', async () => {
    renderWithProviders(<SlapjackPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeInTheDocument());
    expect(screen.getAllByText(/26/).length).toBeGreaterThan(0);
  });

  it('step button calls exec with step', async () => {
    renderWithProviders(<SlapjackPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('step-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('step'));
  });

  it('slap button calls exec with slap and playerIdx 0', async () => {
    mockExec.mockResolvedValueOnce(jackOnTopState);
    renderWithProviders(<SlapjackPage />);
    await waitFor(() => expect(screen.getByTestId('slap-button')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('slap-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('slap', { playerIdx: 0 }));
  });

  it('disables slap button when pile is empty', async () => {
    renderWithProviders(<SlapjackPage />);
    await waitFor(() => expect(screen.getByTestId('slap-button')).toBeDisabled());
  });

  it('disables step and slap buttons on game end', async () => {
    mockExec.mockResolvedValueOnce(gameEndState);
    renderWithProviders(<SlapjackPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeDisabled());
    expect(screen.getByTestId('slap-button')).toBeDisabled();
  });

  it('disables step button when it is not the human turn', async () => {
    mockExec.mockResolvedValueOnce({ ...baseState, isHumanTurn: false, currentTurnIdx: 1 });
    renderWithProviders(<SlapjackPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeDisabled());
  });

  it('renders CLI terminal when enabled', async () => {
    mockUseCliMode.mockReturnValue({
      cliEnabled: true,
      toggleCli: vi.fn(),
      logEntries: [],
      addInput: vi.fn(),
      addOutput: vi.fn(),
      addError: vi.fn(),
      clearLog: vi.fn(),
    });
    renderWithProviders(<SlapjackPage />);
    await waitFor(() => expect(screen.queryByTestId('step-button')).not.toBeInTheDocument());
  });
});
