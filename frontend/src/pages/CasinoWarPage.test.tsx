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

  it('advertises the bet keyboard shortcut on the button', async () => {
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<CasinoWarPage />);
    const betBtn = await screen.findByRole('button', { name: /ベット/ });
    expect(betBtn).toHaveAttribute('aria-keyshortcuts', 'b');
    // The KbdBadge chip is present (its text is aria-hidden, so the accessible name is unaffected).
    expect(betBtn.querySelector('kbd')?.textContent).toBe('B');
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

  it('shows the War cost badge with the ante amount', async () => {
    mockApi.mockResolvedValue(tieState); // ante 100
    renderWithProviders(<CasinoWarPage />);
    const badge = await screen.findByTestId('war-cost-badge');
    expect(badge).toHaveTextContent('100');
    expect(screen.queryByTestId('war-insufficient')).not.toBeInTheDocument();
  });

  it('disables War and shows an insufficient-chips alert when chips < ante', async () => {
    mockApi.mockResolvedValue({ ...tieState, chips: 50, ante: 100 });
    renderWithProviders(<CasinoWarPage />);
    await waitFor(() => expect(screen.getByTestId('war-insufficient')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: /ウォー/ })).toBeDisabled();
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

  it('renders initial cards in INITIAL_DEALT phase', async () => {
    const initialDealtState: CasinoWarResponse = {
      playerCard: card('SPADE', 10),
      dealerCard: card('HEART', 5),
      burnCards: [],
      phase: CasinoWarPhase.INITIAL_DEALT,
      chips: 900,
      ante: 100,
      warBet: 0,
      result: 0,
      totalPayout: 0,
      message: '',
    };
    mockApi.mockResolvedValue(initialDealtState);
    const { container } = renderWithProviders(<CasinoWarPage />);
    await waitFor(() => expect(container.querySelector('[data-tutorial="cw-results"]')).toBeInTheDocument());
  });

  it('renders burn cards and war cards in WAR_DEALT phase', async () => {
    const warDealtState: CasinoWarResponse = {
      playerCard: card('SPADE', 7),
      dealerCard: card('HEART', 7),
      burnCards: [card('CLOVER', 2), card('CLOVER', 3), card('CLOVER', 4)],
      playerWarCard: card('DIAMOND', 13),
      dealerWarCard: card('SPADE', 5),
      phase: CasinoWarPhase.WAR_DEALT,
      chips: 800,
      ante: 100,
      warBet: 100,
      result: 0,
      totalPayout: 0,
      message: '',
    };
    mockApi.mockResolvedValue(warDealtState);
    renderWithProviders(<CasinoWarPage />);
    await waitFor(() => expect(screen.getByText(/burn|焼き札|焼/i)).toBeInTheDocument());
    // Burn cards are shown face-up (as AnimatedCard), matching the CUI — no backs.
    expect(screen.queryAllByTestId('animated-card-back')).toHaveLength(0);
    // 2 main + 3 burn + 2 war cards are all rendered face-up.
    expect(screen.getAllByTestId('animated-card')).toHaveLength(7);
  });

  it('renders the CLI terminal when useCliMode reports cliEnabled', async () => {
    mockUseCliMode.mockReturnValueOnce({
      cliEnabled: true,
      toggleCli: vi.fn(),
      logEntries: [],
      addInput: vi.fn(),
      addOutput: vi.fn(),
      addError: vi.fn(),
      clearLog: vi.fn(),
    });
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<CasinoWarPage />);
    await waitFor(() => expect(screen.queryAllByRole('button', { name: /ベット/ })).toHaveLength(0));
  });

  it('shows a Rebet button at end-phase after a bet has been placed', async () => {
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<CasinoWarPage />);
    const betBtn = await screen.findByRole('button', { name: /ベット/ });
    mockApi.mockClear();
    mockApi.mockResolvedValue(winState);
    fireEvent.click(betBtn);
    await waitFor(() => expect(screen.getByTestId('cw-rebet-button')).toBeInTheDocument());

    mockApi.mockClear();
    mockApi.mockResolvedValueOnce(betState);
    mockApi.mockResolvedValueOnce(winState);
    fireEvent.click(screen.getByTestId('cw-rebet-button'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', 100));
  });

  it('disables War button when chips < ante', async () => {
    const broke: CasinoWarResponse = { ...tieState, chips: 50 };
    mockApi.mockResolvedValue(broke);
    renderWithProviders(<CasinoWarPage />);
    const warBtn = await screen.findByRole('button', { name: /ウォー/ });
    expect(warBtn).toBeDisabled();
  });

  it('does not show the Rebet button when chips are insufficient to replay', async () => {
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<CasinoWarPage />);
    const betBtn = await screen.findByRole('button', { name: /ベット/ });
    mockApi.mockResolvedValue({ ...winState, chips: 50 });
    fireEvent.click(betBtn);
    await waitFor(() => expect(screen.getByRole('button', { name: /次のゲーム/ })).toBeInTheDocument());
    expect(screen.queryByTestId('cw-rebet-button')).not.toBeInTheDocument();
  });

  it("snapshots the bet when the 'b' keyboard shortcut is used so Rebet is available at end phase", async () => {
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<CasinoWarPage />);
    await screen.findByRole('button', { name: /ベット/ });
    mockApi.mockClear();
    mockApi.mockResolvedValue(winState);
    fireEvent.keyDown(document, { key: 'b' });
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', 100));
    await waitFor(() => expect(screen.getByTestId('cw-rebet-button')).toBeInTheDocument());
  });

  it("the 'e' keyboard shortcut fires Rebet at end phase", async () => {
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<CasinoWarPage />);
    const betBtn = await screen.findByRole('button', { name: /ベット/ });
    mockApi.mockResolvedValue(winState);
    fireEvent.click(betBtn);
    await waitFor(() => expect(screen.getByTestId('cw-rebet-button')).toBeInTheDocument());

    mockApi.mockClear();
    mockApi.mockResolvedValueOnce(betState);
    mockApi.mockResolvedValueOnce(winState);
    fireEvent.keyDown(document, { key: 'e' });
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', 100));
  });

  it('suggests the previous bet in the bet phase and refills the input when clicked', async () => {
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<CasinoWarPage />);
    const input = (await screen.findByLabelText(/アンテ/)) as HTMLInputElement;

    // Place a non-default bet of 300, then return to the bet phase.
    fireEvent.change(input, { target: { value: '300' } });
    fireEvent.click(screen.getByRole('button', { name: /^ベット$/ }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', 300));

    // Lower the current input below the last bet so the suggestion appears.
    fireEvent.change(input, { target: { value: '150' } });
    const suggest = await screen.findByTestId('cw-previous-bet');
    expect(suggest).toHaveTextContent('300');

    fireEvent.click(suggest);
    await waitFor(() => expect(input.value).toBe('300'));
    // Once the input matches the previous bet, the suggestion is redundant and hides.
    await waitFor(() => expect(screen.queryByTestId('cw-previous-bet')).not.toBeInTheDocument());
  });

  it('does not suggest the previous bet before any bet is placed', async () => {
    mockApi.mockResolvedValue(betState);
    renderWithProviders(<CasinoWarPage />);
    await screen.findByRole('button', { name: /^ベット$/ });
    expect(screen.queryByTestId('cw-previous-bet')).not.toBeInTheDocument();
  });
});
