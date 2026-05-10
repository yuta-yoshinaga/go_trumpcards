import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { dragontigerApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { DragonTigerResponse } from '../types/card';
import { DragonTigerBetType, DragonTigerHistoryResult, DragonTigerPhase } from '../types/phases';
import { DragonTigerPage } from './DragonTigerPage';

vi.mock('../api/gameApi', () => ({
  dragontigerApi: { exec: vi.fn() },
  actionLogApi: { dragontiger: vi.fn() },
}));

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

const mockApi = vi.mocked(dragontigerApi.exec);
const mockUseCliMode = vi.mocked(useCliMode);

const betState: DragonTigerResponse = {
  phase: DragonTigerPhase.BET,
  chips: 1000,
  betAmount: 0,
  betType: 0,
  result: 0,
  payout: 0,
  history: [],
  message: '',
};

const dragonWinState: DragonTigerResponse = {
  dragonCard: { design: 'SPADE', value: 13 },
  tigerCard: { design: 'HEART', value: 5 },
  phase: DragonTigerPhase.END,
  chips: 1100,
  betAmount: 100,
  betType: DragonTigerBetType.DRAGON,
  result: 1,
  payout: 200,
  history: [DragonTigerHistoryResult.DRAGON],
  message: 'Dragon wins!',
  messageCode: 'dragontiger.result.dragonWins',
};

const tigerWinOnTigerBetState: DragonTigerResponse = {
  dragonCard: { design: 'SPADE', value: 3 },
  tigerCard: { design: 'HEART', value: 13 },
  phase: DragonTigerPhase.END,
  chips: 1100,
  betAmount: 100,
  betType: DragonTigerBetType.TIGER,
  // result===2 (GameResultLose, tiger-side won) but the player took Tiger so they win.
  result: 2,
  payout: 200,
  history: [DragonTigerHistoryResult.TIGER],
  message: 'Tiger wins!',
  messageCode: 'dragontiger.result.tigerWins',
};

const tieWinOnTieBetState: DragonTigerResponse = {
  dragonCard: { design: 'SPADE', value: 7 },
  tigerCard: { design: 'HEART', value: 7 },
  phase: DragonTigerPhase.END,
  chips: 1800,
  betAmount: 100,
  betType: DragonTigerBetType.TIE,
  result: 3,
  payout: 900,
  history: [DragonTigerHistoryResult.TIE],
  message: 'Tie! You win the tie bet.',
  messageCode: 'dragontiger.result.tieWin',
};

const tieRefundOnDragonBetState: DragonTigerResponse = {
  dragonCard: { design: 'SPADE', value: 7 },
  tigerCard: { design: 'HEART', value: 7 },
  phase: DragonTigerPhase.END,
  chips: 950,
  betAmount: 100,
  betType: DragonTigerBetType.DRAGON,
  // Tie outcome: half-refund on a Dragon bet means payout < betAmount → no celebration.
  result: 3,
  payout: 50,
  history: [DragonTigerHistoryResult.TIE],
  message: 'Tie. Half of your bet is refunded.',
  messageCode: 'dragontiger.result.tieRefund',
};

describe('DragonTigerPage', () => {
  beforeEach(() => {
    mockApi.mockReset();
    mockApi.mockResolvedValue(betState);
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

  it('issues a reset on mount', async () => {
    renderWithProviders(<DragonTigerPage />);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  it('renders three bet buttons in the bet phase', async () => {
    renderWithProviders(<DragonTigerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ドラゴン' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'タイガー' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /タイ \(8:1\)/ })).toBeInTheDocument();
  });

  it('dispatches a Dragon bet on button click', async () => {
    renderWithProviders(<DragonTigerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ドラゴン' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'ドラゴン' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', 100, DragonTigerBetType.DRAGON));
  });

  it('dispatches a Tiger bet on button click', async () => {
    renderWithProviders(<DragonTigerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'タイガー' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'タイガー' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', 100, DragonTigerBetType.TIGER));
  });

  it('dispatches a Tie bet on button click', async () => {
    renderWithProviders(<DragonTigerPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /タイ \(8:1\)/ })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: /タイ \(8:1\)/ }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', 100, DragonTigerBetType.TIE));
  });

  it('renders dragon and tiger cards + Big Road in the end phase', async () => {
    mockApi.mockResolvedValueOnce(dragonWinState);
    renderWithProviders(<DragonTigerPage />);
    await waitFor(() => expect(screen.getByText('ドラゴン')).toBeInTheDocument());
    expect(screen.getByText('タイガー')).toBeInTheDocument();
    expect(screen.getByTestId('bigroad')).toBeInTheDocument();
    expect(screen.getByText('払戻し: 200')).toBeInTheDocument();
  });

  // Regression coverage for the gemini/Claude review: winShow must use
  // payout > betAmount, not result===Win-only, otherwise Tiger and Tie wins
  // are silently uncelebrated.
  it('renders payout for a Tiger-bet Tiger-win (player wins)', async () => {
    mockApi.mockResolvedValueOnce(tigerWinOnTigerBetState);
    renderWithProviders(<DragonTigerPage />);
    await waitFor(() => expect(screen.getByText('払戻し: 200')).toBeInTheDocument());
    expect(screen.getByTestId('bigroad')).toBeInTheDocument();
  });

  it('renders payout for a Tie-bet Tie outcome (8:1 win)', async () => {
    mockApi.mockResolvedValueOnce(tieWinOnTieBetState);
    renderWithProviders(<DragonTigerPage />);
    await waitFor(() => expect(screen.getByText('払戻し: 900')).toBeInTheDocument());
  });

  it('renders payout for a Dragon-bet half-refund tie (no celebration expected)', async () => {
    mockApi.mockResolvedValueOnce(tieRefundOnDragonBetState);
    renderWithProviders(<DragonTigerPage />);
    await waitFor(() => expect(screen.getByText('払戻し: 50')).toBeInTheDocument());
  });
});
