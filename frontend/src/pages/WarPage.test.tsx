import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { warApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { WarResponse } from '../types/card';
import { WarPhase } from '../types/phases';
import { WarPage } from './WarPage';

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
  warApi: { exec: vi.fn() },
  actionLogApi: { war: vi.fn() },
}));

const mockExec = vi.mocked(warApi.exec);

const baseState: WarResponse = {
  players: [
    { id: 0, isHuman: true, drawPileSize: 26, discardPileSize: 0, totalCards: 26 },
    { id: 1, isHuman: false, drawPileSize: 26, discardPileSize: 0, totalCards: 26 },
  ],
  phase: WarPhase.REVEAL,
  gameEndFlag: false,
  winnerIdx: -1,
  playerRevealed: null,
  cpuRevealed: null,
  warPotSize: 0,
  lastWinnerIdx: -1,
  lastBurialCount: 0,
  roundsPlayed: 0,
  config: { maxRounds: 500 },
  message: '',
  messageCode: 'reveal',
};

const warPhaseState: WarResponse = {
  ...baseState,
  phase: WarPhase.WAR_BURY,
  playerRevealed: { design: 'SPADE', value: 7 },
  cpuRevealed: { design: 'HEART', value: 7 },
  warPotSize: 2,
  messageCode: 'war',
};

const gameEndState: WarResponse = {
  ...baseState,
  phase: WarPhase.GAME_END,
  gameEndFlag: true,
  winnerIdx: 0,
  players: [
    { id: 0, isHuman: true, drawPileSize: 0, discardPileSize: 52, totalCards: 52 },
    { id: 1, isHuman: false, drawPileSize: 0, discardPileSize: 0, totalCards: 0 },
  ],
  messageCode: 'gameEnd',
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

describe('WarPage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<WarPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders pile info after state loads', async () => {
    renderWithProviders(<WarPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeInTheDocument());
    // Both players show draw pile = 26 (rendered as "山札: 26")
    const pileLines = screen.getAllByText(/26/);
    expect(pileLines.length).toBeGreaterThan(0);
  });

  it('step button calls exec with step', async () => {
    renderWithProviders(<WarPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('step-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('step'));
  });

  it('shows war phase pot count when tie occurred', async () => {
    mockExec.mockResolvedValueOnce(warPhaseState);
    renderWithProviders(<WarPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeInTheDocument());
    // Pot size 2 should appear in the label
    expect(screen.getAllByText(/2/).length).toBeGreaterThan(0);
  });

  it('disables step button on game end', async () => {
    mockExec.mockResolvedValueOnce(gameEndState);
    renderWithProviders(<WarPage />);
    await waitFor(() => expect(screen.getByTestId('step-button')).toBeDisabled());
  });

  it('autoplay button calls exec with autoplay', async () => {
    renderWithProviders(<WarPage />);
    await waitFor(() => expect(screen.getByTestId('autoplay-button')).toBeInTheDocument());
    fireEvent.click(screen.getByTestId('autoplay-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autoplay'));
  });

  it('disables autoplay button on game end', async () => {
    mockExec.mockResolvedValueOnce(gameEndState);
    renderWithProviders(<WarPage />);
    await waitFor(() => expect(screen.getByTestId('autoplay-button')).toBeDisabled());
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
    renderWithProviders(<WarPage />);
    await waitFor(() => expect(screen.queryByTestId('step-button')).not.toBeInTheDocument());
  });
});
