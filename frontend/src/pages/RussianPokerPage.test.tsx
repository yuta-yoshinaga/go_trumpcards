import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { russianpokerApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, RussianPokerResponse } from '../types/card';
import { RussianPokerPhase } from '../types/phases';
import { RussianPokerPage } from './RussianPokerPage';

vi.mock('../api/gameApi', () => ({
  russianpokerApi: { exec: vi.fn() },
  actionLogApi: { russianpoker: vi.fn() },
}));

const mockExec = vi.mocked(russianpokerApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function makeState(overrides: Partial<RussianPokerResponse> = {}): RussianPokerResponse {
  return {
    playerHand: [card('SPADE', 10), card('HEART', 11), card('DIAMOND', 12), card('CLOVER', 3), card('SPADE', 5)],
    dealerHand: [{ design: '', value: 0 }] as RussianPokerResponse['dealerHand'],
    phase: RussianPokerPhase.ACTION,
    chips: 900,
    anteBet: 100,
    exchangeCount: 0,
    exchangeFee: 0,
    bought6th: false,
    buy6thFee: 0,
    forceExchanged: false,
    forceExchangeFee: 0,
    playBet: 0,
    result: 0,
    antePayout: 0,
    playPayout: 0,
    totalPayout: 0,
    dealerQualified: false,
    playerHandRank: 0,
    dealerHandRank: 0,
    message: '',
    ...overrides,
  };
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeState());
});

describe('RussianPokerPage', () => {
  it('resets on mount', async () => {
    renderWithProviders(<RussianPokerPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('labels player cards with localized names via cardAlt', async () => {
    renderWithProviders(<RussianPokerPage />);
    // playerHand[0] is ♠10 → cardAlt label, not the old "Card 1".
    const firstCard = await screen.findByTestId('player-card-0');
    expect(firstCard).toHaveAttribute('aria-label', '♠ 10');
    expect(firstCard.getAttribute('aria-label')).not.toMatch(/Card 1/);
  });

  it('shows the exchange-cost line from the start of the action phase (0 cards = 0 fee)', async () => {
    renderWithProviders(<RussianPokerPage />);
    const line = await screen.findByTestId('russian-exchange-fee-line');
    // Visible with zero selection: count 0 and fee 0 (ante 100).
    expect(line).toHaveTextContent('選択中: 0枚');
    expect(line).toHaveTextContent('× 100 = 0');
    // High-risk warning is absent at 0 selected.
    expect(within(line).queryByText(/⚠/)).not.toBeInTheDocument();
  });

  it('updates the exchange fee as cards are selected', async () => {
    renderWithProviders(<RussianPokerPage />);
    await screen.findByTestId('russian-exchange-fee-line');
    // Select a card from the hand → fee becomes ante*1 = 100.
    fireEvent.click(screen.getByAltText('♠ 10'));
    const line = screen.getByTestId('russian-exchange-fee-line');
    expect(line).toHaveTextContent('選択中: 1枚');
  });

  it('shows the high-risk warning and error styling at 4+ selected cards', async () => {
    renderWithProviders(<RussianPokerPage />);
    await screen.findByTestId('russian-exchange-fee-line');
    // Select 4 of the 5 hand cards (11=J, 12=Q via cardAlt).
    fireEvent.click(screen.getByAltText('♠ 10'));
    fireEvent.click(screen.getByAltText('♥ J'));
    fireEvent.click(screen.getByAltText('♦ Q'));
    fireEvent.click(screen.getByAltText('♣ 3'));
    const line = screen.getByTestId('russian-exchange-fee-line');
    expect(line).toHaveTextContent('選択中: 4枚');
    expect(within(line).getByText(/⚠/)).toBeInTheDocument();
    expect(line.querySelector('.text-ds-error')).not.toBeNull();
  });

  const betState = makeState({ phase: RussianPokerPhase.BET, playerHand: [], dealerHand: [], chips: 900 });

  it('renders the ante as a ChipBetInput with steppers', async () => {
    mockExec.mockResolvedValue(betState);
    renderWithProviders(<RussianPokerPage />);
    expect(await screen.findByRole('button', { name: 'アンテ +10' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'アンテ −10' })).toBeInTheDocument();
  });

  it('sends the bet command with the ante entered via ChipBetInput', async () => {
    mockExec.mockResolvedValue(betState);
    renderWithProviders(<RussianPokerPage />);
    const anteInput = (await screen.findByLabelText('アンテ')) as HTMLInputElement;
    fireEvent.change(anteInput, { target: { value: '200' } });
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 200));
  });

  it('disables Bet and shows an alert for an out-of-range ante', async () => {
    mockExec.mockResolvedValue(betState);
    renderWithProviders(<RussianPokerPage />);
    const anteInput = (await screen.findByLabelText('アンテ')) as HTMLInputElement;
    // 15 is not a multiple of 10 → invalid (below-min/over-balance are auto-clamped away).
    fireEvent.change(anteInput, { target: { value: '15' } });
    expect(screen.getByRole('alert')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ベット' })).toBeDisabled();
  });
});
