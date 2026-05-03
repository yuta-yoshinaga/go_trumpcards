import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { casinowarApi } from '../api/gameApi';
import { useCliMode } from '../hooks/useCliMode';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, CasinoWarResponse } from '../types/card';
import { CasinoWarPhase } from '../types/phases';
import { CasinoWarPage } from './CasinoWarPage';

vi.mock('../api/gameApi', () => ({
  casinowarApi: { exec: vi.fn() },
  actionLogApi: { casinowar: vi.fn() },
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

const mockApi = vi.mocked(casinowarApi.exec);
const mockUseCliMode = vi.mocked(useCliMode);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const betState: CasinoWarResponse = {
  burnCards: [],
  phase: CasinoWarPhase.BET,
  chips: 1000,
  ante: 0,
  warBet: 0,
  result: 0,
  totalPayout: 0,
  message: '',
};

const tieState: CasinoWarResponse = {
  playerCard: card('SPADE', 7),
  dealerCard: card('HEART', 7),
  burnCards: [],
  phase: CasinoWarPhase.TIE_DECISION,
  chips: 900,
  ante: 100,
  warBet: 0,
  result: 0,
  totalPayout: 0,
  message: '',
};

const winState: CasinoWarResponse = {
  playerCard: card('SPADE', 13),
  dealerCard: card('HEART', 7),
  burnCards: [],
  phase: CasinoWarPhase.END,
  chips: 1100,
  ante: 100,
  warBet: 0,
  result: 1,
  totalPayout: 200,
  message: 'You win!',
};

beforeEach(() => {
  vi.clearAllMocks();
  mockApi.mockResolvedValue(betState);
});

describe('CasinoWarPage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<CasinoWarPage />);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  it('renders bet phase with bet button', async () => {
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<CasinoWarPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /ベット/ })).toBeInTheDocument());
  });

  it('triggers bet action with current amount', async () => {
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<CasinoWarPage />);
    const betBtn = await screen.findByRole('button', { name: /ベット/ });
    mockApi.mockClear();
    fireEvent.click(betBtn);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', 100));
  });

  it('renders tie decision with surrender & war buttons', async () => {
    mockApi.mockResolvedValue(tieState);
    renderWithProviders(<CasinoWarPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /ウォー/ })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: /サレンダー/ })).toBeInTheDocument();
  });

  it('triggers surrender on surrender click', async () => {
    mockApi.mockResolvedValue(tieState);
    renderWithProviders(<CasinoWarPage />);
    const btn = await screen.findByRole('button', { name: /サレンダー/ });
    mockApi.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('surrender'));
  });

  it('triggers war on war click', async () => {
    mockApi.mockResolvedValue(tieState);
    renderWithProviders(<CasinoWarPage />);
    const btn = await screen.findByRole('button', { name: /ウォー/ });
    mockApi.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('war'));
  });

  it('renders end phase with reset and payout', async () => {
    mockApi.mockResolvedValue(winState);
    renderWithProviders(<CasinoWarPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /次のゲーム/ })).toBeInTheDocument());
    expect(screen.getByText(/200/)).toBeInTheDocument();
  });

  it('reads from useCliMode', async () => {
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<CasinoWarPage />);
    await waitFor(() => expect(mockUseCliMode).toHaveBeenCalledWith('casinowar'));
  });
});
