import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { faroApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, FaroResponse } from '../types/card';
import { FaroPage } from './FaroPage';

vi.mock('../api/gameApi', () => ({
  faroApi: { exec: vi.fn() },
  actionLogApi: { faro: vi.fn() },
}));

const mockExec = vi.mocked(faroApi.exec);

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<FaroResponse> = {}): FaroResponse {
  return {
    phase: 1,
    chips: 1000,
    bets: [],
    soda: null,
    losingCard: null,
    winningCard: null,
    split: false,
    turnsPlayed: 0,
    turnsTotal: 25,
    remaining: 52,
    callCards: [],
    callOrder: [],
    callWon: false,
    totalPayout: 0,
    gameEndFlag: false,
    message: '',
    ...overrides,
  };
}

const bettingState = makeState();
const turnState = makeState({
  phase: 2,
  bets: [{ rank: 7, amount: 100, copper: false }],
  losingCard: card('SPADE', 3),
  winningCard: card('HEART', 7),
  turnsPlayed: 1,
  remaining: 49,
});
const splitState = makeState({
  phase: 2,
  losingCard: card('SPADE', 5),
  winningCard: card('HEART', 5),
  split: true,
});
const callState = makeState({
  phase: 3,
  callCards: [card('SPADE', 3), card('HEART', 9), card('DIAMOND', 12)],
});
const roundEndWinState = makeState({ phase: 4, totalPayout: 200, chips: 1200 });
const gameEndState = makeState({ phase: 5, gameEndFlag: true, chips: 0 });

beforeEach(() => {
  localStorage.clear();
  mockExec.mockReset();
  mockExec.mockResolvedValue(bettingState);
});

describe('FaroPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<FaroPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<FaroPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows the chips and remaining cards', async () => {
    renderWithProviders(<FaroPage />);
    await waitFor(() => expect(screen.getByText(/チップ: 1000/)).toBeInTheDocument());
    expect(screen.getByText(/残り: 52枚/)).toBeInTheDocument();
  });

  it('renders the 13-rank betting layout', async () => {
    renderWithProviders(<FaroPage />);
    await waitFor(() => expect(screen.getByTestId('rank-1')).toBeInTheDocument());
    expect(screen.getByTestId('rank-13')).toBeInTheDocument();
  });

  it('places a bet when a rank is clicked during betting', async () => {
    renderWithProviders(<FaroPage />);
    const rankBtn = await screen.findByTestId('rank-7');
    mockExec.mockClear();
    fireEvent.click(rankBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', { rank: 7, amount: 10, copper: false }));
  });

  it('selects a chip amount and bets with it', async () => {
    renderWithProviders(<FaroPage />);
    await screen.findByTestId('rank-7');
    fireEvent.click(screen.getByTestId('chip-100'));
    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('rank-7'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', { rank: 7, amount: 100, copper: false }));
  });

  it('toggles copper and bets to lose', async () => {
    renderWithProviders(<FaroPage />);
    await screen.findByTestId('rank-7');
    fireEvent.click(screen.getByTestId('copper-toggle'));
    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('rank-7'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', { rank: 7, amount: 10, copper: true }));
  });

  it('shows the copper-mode hint on the layout when the copper toggle is on', async () => {
    renderWithProviders(<FaroPage />);
    await screen.findByTestId('rank-7');
    expect(screen.queryByTestId('faro-copper-mode')).not.toBeInTheDocument();
    fireEvent.click(screen.getByTestId('copper-toggle'));
    expect(screen.getByTestId('faro-copper-mode')).toBeInTheDocument();
  });

  it('distinguishes a copper bet with accent styling and a coin marker', async () => {
    mockExec.mockResolvedValue(makeState({ bets: [{ rank: 7, amount: 100, copper: true }] }));
    renderWithProviders(<FaroPage />);
    const rankBtn = await screen.findByTestId('rank-7');
    expect(rankBtn.className).toContain('border-ds-accent');
    expect(rankBtn.textContent).toContain('🪙');
    // A normal (non-copper) rank stays on the warning palette.
    expect(screen.getByTestId('rank-8').className).not.toContain('border-ds-accent');
  });

  it('shows the no-bets message when no chips are placed', async () => {
    renderWithProviders(<FaroPage />);
    await waitFor(() => expect(screen.getByText('まだベットがありません。')).toBeInTheDocument());
  });

  it('lists a placed bet and clears it', async () => {
    mockExec.mockResolvedValue(makeState({ bets: [{ rank: 7, amount: 100, copper: false }] }));
    renderWithProviders(<FaroPage />);
    const clearBtn = await screen.findByTestId('clear-bet-7');
    mockExec.mockClear();
    fireEvent.click(clearBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('clearBet', { rank: 7 }));
  });

  it('deals a turn from the deal button', async () => {
    renderWithProviders(<FaroPage />);
    const dealBtn = await screen.findByTestId('deal-button');
    mockExec.mockClear();
    fireEvent.click(dealBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('deal'));
  });

  it('reveals the losing and winning cards of the last turn', async () => {
    mockExec.mockResolvedValue(turnState);
    renderWithProviders(<FaroPage />);
    await waitFor(() => expect(screen.getByText('直前のターン')).toBeInTheDocument());
    expect(screen.getByText('負け札（バンク回収）')).toBeInTheDocument();
    expect(screen.getByText('勝ち札（1:1配当）')).toBeInTheDocument();
  });

  it('shows the split indicator on a split turn', async () => {
    mockExec.mockResolvedValue(splitState);
    renderWithProviders(<FaroPage />);
    await waitFor(() => expect(screen.getByText('スプリット（バンクが半分回収）')).toBeInTheDocument());
  });

  it('submits a call once all three ranks are ordered by tapping the card images', async () => {
    mockExec.mockResolvedValue(callState);
    renderWithProviders(<FaroPage />);
    await screen.findByTestId('call-card-3-0');
    fireEvent.click(screen.getByTestId('call-card-3-0'));
    fireEvent.click(screen.getByTestId('call-card-9-1'));
    fireEvent.click(screen.getByTestId('call-card-12-2'));
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'コールする' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('call', { order: [3, 9, 12] }));
  });

  it('shows an order badge and marks the tapped card as pressed', async () => {
    mockExec.mockResolvedValue(callState);
    renderWithProviders(<FaroPage />);
    const secondCard = await screen.findByTestId('call-card-9-1');
    // Nothing selected yet: no badges and no pressed cards.
    expect(screen.queryByTestId('call-order-badge-9-1')).not.toBeInTheDocument();
    expect(secondCard).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(screen.getByTestId('call-card-3-0'));
    fireEvent.click(secondCard);
    // The second tap becomes position 2.
    expect(screen.getByTestId('call-order-badge-9-1')).toHaveTextContent('2');
    expect(secondCard).toHaveAttribute('aria-pressed', 'true');

    // Tapping again clears the selection and its badge.
    fireEvent.click(secondCard);
    expect(screen.queryByTestId('call-order-badge-9-1')).not.toBeInTheDocument();
    expect(secondCard).toHaveAttribute('aria-pressed', 'false');
  });

  it('selects duplicate ranks independently by index', async () => {
    mockExec.mockResolvedValue(
      makeState({ phase: 3, callCards: [card('SPADE', 5), card('HEART', 5), card('DIAMOND', 9)] }),
    );
    renderWithProviders(<FaroPage />);
    // Tap the second 5 first, then the first 5, then the 9.
    await screen.findByTestId('call-card-5-1');
    fireEvent.click(screen.getByTestId('call-card-5-1'));
    fireEvent.click(screen.getByTestId('call-card-5-0'));
    fireEvent.click(screen.getByTestId('call-card-9-2'));
    expect(screen.getByTestId('call-order-badge-5-1')).toHaveTextContent('1');
    expect(screen.getByTestId('call-order-badge-5-0')).toHaveTextContent('2');
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'コールする' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('call', { order: [5, 5, 9] }));
  });

  it('skips the call without an order', async () => {
    mockExec.mockResolvedValue(callState);
    renderWithProviders(<FaroPage />);
    const skipBtn = await screen.findByRole('button', { name: 'コールせずに終える' });
    mockExec.mockClear();
    fireEvent.click(skipBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('call', { order: [] }));
  });

  it('shows a next button at round end and dispatches next', async () => {
    mockExec.mockResolvedValue(roundEndWinState);
    renderWithProviders(<FaroPage />);
    const nextBtn = await screen.findByTestId('next-button');
    mockExec.mockClear();
    fireEvent.click(nextBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('shows the game-over message when out of chips', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<FaroPage />);
    await waitFor(() => expect(screen.getByText('チップが尽きました。ゲーム終了です。')).toBeInTheDocument());
  });

  describe('case keeper', () => {
    it('starts every rank at four before any card is revealed', async () => {
      renderWithProviders(<FaroPage />);
      const grid = await screen.findByTestId('case-keeper');
      expect(grid).toBeInTheDocument();
      for (const rank of [1, 7, 13]) {
        expect(screen.getByTestId(`case-keeper-rank-${rank}`)).toHaveTextContent('4');
      }
    });

    it('decrements ranks by the cards revealed on a turn', async () => {
      // soda (SPADE A), losing (SPADE 3), winning (HEART 7) -> A, 3, 7 each drop to 3.
      mockExec.mockResolvedValue(
        makeState({
          phase: 2,
          soda: card('SPADE', 1),
          losingCard: card('SPADE', 3),
          winningCard: card('HEART', 7),
          turnsPlayed: 1,
        }),
      );
      renderWithProviders(<FaroPage />);
      await waitFor(() => expect(screen.getByTestId('case-keeper-rank-1')).toHaveTextContent('3'));
      expect(screen.getByTestId('case-keeper-rank-3')).toHaveTextContent('3');
      expect(screen.getByTestId('case-keeper-rank-7')).toHaveTextContent('3');
      // An untouched rank stays at four.
      expect(screen.getByTestId('case-keeper-rank-5')).toHaveTextContent('4');
    });

    it('resets all ranks to four when the game is reset', async () => {
      mockExec.mockResolvedValue(
        makeState({ phase: 2, losingCard: card('SPADE', 3), winningCard: card('HEART', 7), turnsPlayed: 1 }),
      );
      renderWithProviders(<FaroPage />);
      await waitFor(() => expect(screen.getByTestId('case-keeper-rank-3')).toHaveTextContent('3'));

      // Reset returns a fresh deck; the case keeper must clear back to four.
      mockExec.mockResolvedValue(bettingState);
      fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
      const confirm = await screen.findByRole('button', { name: '確認' });
      fireEvent.click(confirm);
      await waitFor(() => expect(screen.getByTestId('case-keeper-rank-3')).toHaveTextContent('4'));
    });
  });
});
