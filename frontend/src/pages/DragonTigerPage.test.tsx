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
});
