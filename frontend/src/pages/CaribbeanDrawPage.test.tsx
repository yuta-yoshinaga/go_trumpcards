import { fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { caribbeandrawApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, CaribbeanDrawResponse } from '../types/card';
import { CaribbeanDrawPhase } from '../types/phases';
import { CaribbeanDrawPage } from './CaribbeanDrawPage';

vi.mock('../api/gameApi', () => ({
  caribbeandrawApi: { exec: vi.fn() },
  actionLogApi: { caribbeandraw: vi.fn() },
}));

const mockApi = vi.mocked(caribbeandrawApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const betPhaseState: CaribbeanDrawResponse = {
  playerHand: [],
  dealerHand: [],
  phase: CaribbeanDrawPhase.BET,
  chips: 1000,
  anteBet: 0,
  jackpotBet: 0,
  playBet: 0,
  result: 0,
  antePayout: 0,
  playPayout: 0,
  jackpotPayout: 0,
  totalPayout: 0,
  dealerQualified: false,
  drawCost: 0,
  playerHandRank: 0,
  dealerHandRank: 0,
  message: '',
};

const maskedCard = { design: '' as CardDesign, value: 0 };

// The draw phase sits between BET and ACTION, so ACTION is 3 and END is 4 here
// — one higher than in Caribbean Stud, which this game was cloned from.
const drawPhaseState: CaribbeanDrawResponse = {
  ...betPhaseState,
  phase: CaribbeanDrawPhase.DRAW,
  playerHand: [card('SPADE', 10), card('HEART', 11), card('DIAMOND', 13), card('CLOVER', 5), card('SPADE', 7)],
  // 1 face-up dealer card + 4 masked cards (security: dealer hand not fully revealed)
  dealerHand: [card('HEART', 13), maskedCard, maskedCard, maskedCard, maskedCard],
  anteBet: 100,
  chips: 900,
};

const actionPhaseState: CaribbeanDrawResponse = {
  ...drawPhaseState,
  phase: CaribbeanDrawPhase.ACTION,
};

const endPhasePlayerWins: CaribbeanDrawResponse = {
  playerHand: [card('SPADE', 7), card('CLOVER', 7), card('HEART', 7), card('DIAMOND', 4), card('SPADE', 2)],
  dealerHand: [card('CLOVER', 5), card('DIAMOND', 5), card('HEART', 8), card('SPADE', 11), card('DIAMOND', 1)],
  phase: CaribbeanDrawPhase.END,
  chips: 1500,
  anteBet: 100,
  jackpotBet: 0,
  playBet: 200,
  result: 1,
  antePayout: 200,
  playPayout: 600,
  jackpotPayout: 0,
  totalPayout: 800,
  dealerQualified: true,
  drawCost: 0,
  playerHandRank: 3,
  dealerHandRank: 1,
  message: '勝利！',
  messageCode: 'caribbeandraw.result.playerWins',
};

const endPhaseDealerWins: CaribbeanDrawResponse = {
  ...endPhasePlayerWins,
  result: -1,
  antePayout: 0,
  playPayout: 0,
  totalPayout: 0,
  message: 'ディーラー勝利！',
  messageCode: 'caribbeandraw.result.dealerWins',
};

const endPhaseFold: CaribbeanDrawResponse = {
  ...endPhasePlayerWins,
  result: -1,
  playBet: 0,
  antePayout: 0,
  playPayout: 0,
  totalPayout: 0,
  dealerHand: [],
  dealerQualified: false,
  dealerHandRank: 0,
  message: 'フォールド',
  messageCode: 'caribbeandraw.result.fold',
};

const endPhasePush: CaribbeanDrawResponse = {
  ...endPhasePlayerWins,
  result: 0,
  antePayout: 100,
  playPayout: 200,
  totalPayout: 300,
  message: '引き分け！',
  messageCode: 'caribbeandraw.result.push',
};

const endPhaseDealerNotQualified: CaribbeanDrawResponse = {
  ...endPhasePlayerWins,
  dealerQualified: false,
  antePayout: 200,
  playPayout: 200,
  totalPayout: 400,
  message: 'ディーラー未クオリファイ！',
  messageCode: 'caribbeandraw.result.dealerNotQualified',
};

const endPhaseWithJackpot: CaribbeanDrawResponse = {
  ...endPhasePlayerWins,
  jackpotBet: 10,
  jackpotPayout: 1000,
  totalPayout: 1800,
};

/** A round the player paid to draw in; the fee is a cost, so it is absent from totalPayout. */
const endPhaseAfterDraw: CaribbeanDrawResponse = {
  ...endPhasePlayerWins,
  drawCost: 100,
};

beforeEach(() => {
  vi.clearAllMocks();
});

afterEach(() => {
  localStorage.clear();
});

describe('CaribbeanDrawPage', () => {
  it('renders bet phase on mount', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<CaribbeanDrawPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument();
  });

  it('shows an accessible jackpot help in the bet phase', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<CaribbeanDrawPage />);
    const help = await screen.findByTestId('jackpot-help');
    expect(help).toHaveTextContent('ジャックポットとは？');
    // Native <details>/<summary> is keyboard-focusable.
    expect(help.querySelector('summary')).toBeInTheDocument();
  });

  it('renders skeleton before state loads', () => {
    mockApi.mockReturnValue(new Promise(() => {}));
    renderWithProviders(<CaribbeanDrawPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('lands in the draw phase after betting, not straight in call/fold', async () => {
    mockApi.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(drawPhaseState);
    renderWithProviders(<CaribbeanDrawPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByTestId('cd-stand-pat-button')).toBeInTheDocument());
    // Call/fold belong to the *next* phase; offering them here would let the
    // player skip the draw the game exists for.
    expect(screen.queryByRole('button', { name: 'コール' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'フォールド' })).not.toBeInTheDocument();
  });

  it('shows action phase with call and fold buttons', async () => {
    mockApi.mockResolvedValueOnce(drawPhaseState).mockResolvedValueOnce(actionPhaseState);
    renderWithProviders(<CaribbeanDrawPage />);
    await waitFor(() => expect(screen.getByTestId('cd-stand-pat-button')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('cd-stand-pat-button'));
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
  });

  describe('draw phase', () => {
    it('states the fee before the player commits to the exchange', async () => {
      mockApi.mockResolvedValue(drawPhaseState);
      renderWithProviders(<CaribbeanDrawPage />);
      // The fee is charged on confirm, so it has to be legible while the player
      // is still deciding — not only in the payout panel afterwards.
      await waitFor(() => expect(screen.getByTestId('cd-draw-fee')).toHaveTextContent('交換手数料 100'));
      expect(screen.getByTestId('cd-draw-button')).toHaveTextContent('100');
    });

    it('sends the selected 0-based indices when the draw is confirmed', async () => {
      mockApi.mockResolvedValue(drawPhaseState);
      renderWithProviders(<CaribbeanDrawPage />);
      await waitFor(() => expect(screen.getByTestId('cd-player-card-0')).toBeInTheDocument());

      fireEvent.click(screen.getByTestId('cd-player-card-0'));
      fireEvent.click(screen.getByTestId('cd-player-card-2'));
      expect(screen.getByTestId('cd-player-card-0')).toHaveAttribute('aria-pressed', 'true');
      expect(screen.getByTestId('cd-player-card-2')).toHaveAttribute('aria-pressed', 'true');
      expect(screen.getByTestId('cd-draw-selected')).toHaveTextContent('2 / 2');

      mockApi.mockClear();
      fireEvent.click(screen.getByTestId('cd-draw-button'));
      await waitFor(() => expect(mockApi).toHaveBeenCalledWith('draw', undefined, undefined, [0, 2]));
    });

    it('caps the selection at 2 cards', async () => {
      mockApi.mockResolvedValue(drawPhaseState);
      renderWithProviders(<CaribbeanDrawPage />);
      await waitFor(() => expect(screen.getByTestId('cd-player-card-0')).toBeInTheDocument());

      fireEvent.click(screen.getByTestId('cd-player-card-0'));
      fireEvent.click(screen.getByTestId('cd-player-card-1'));
      fireEvent.click(screen.getByTestId('cd-player-card-4'));
      // The third card must not join the selection: the backend rejects a third
      // index outright, so accepting it here would only buy a failed round-trip.
      expect(screen.getByTestId('cd-player-card-4')).toHaveAttribute('aria-pressed', 'false');
      expect(screen.getByTestId('cd-draw-selected')).toHaveTextContent('2 / 2');

      mockApi.mockClear();
      fireEvent.click(screen.getByTestId('cd-draw-button'));
      await waitFor(() => expect(mockApi).toHaveBeenCalledWith('draw', undefined, undefined, [0, 1]));
    });

    it('deselects a card that is clicked twice', async () => {
      mockApi.mockResolvedValue(drawPhaseState);
      renderWithProviders(<CaribbeanDrawPage />);
      await waitFor(() => expect(screen.getByTestId('cd-player-card-1')).toBeInTheDocument());

      fireEvent.click(screen.getByTestId('cd-player-card-1'));
      fireEvent.click(screen.getByTestId('cd-player-card-1'));
      expect(screen.getByTestId('cd-player-card-1')).toHaveAttribute('aria-pressed', 'false');
      expect(screen.getByTestId('cd-draw-selected')).toHaveTextContent('0 / 2');
      expect(screen.getByTestId('cd-draw-button')).toBeDisabled();
    });

    it('stands pat with an empty index list, discarding any picks made first', async () => {
      mockApi.mockResolvedValue(drawPhaseState);
      renderWithProviders(<CaribbeanDrawPage />);
      await waitFor(() => expect(screen.getByTestId('cd-player-card-0')).toBeInTheDocument());

      // A selection has to be on the board for this to mean anything: standing
      // pat while nothing is picked sends [] no matter what the handler does.
      fireEvent.click(screen.getByTestId('cd-player-card-0'));
      expect(screen.getByTestId('cd-draw-selected')).toHaveTextContent('1 / 2');

      mockApi.mockClear();
      fireEvent.click(screen.getByTestId('cd-stand-pat-button'));
      // Standing pat is free; sending the picks would exchange cards and charge
      // the fee the player just declined.
      await waitFor(() => expect(mockApi).toHaveBeenCalledWith('draw', undefined, undefined, []));
    });

    it('will not confirm an exchange of nothing', async () => {
      mockApi.mockResolvedValue(drawPhaseState);
      renderWithProviders(<CaribbeanDrawPage />);
      await waitFor(() => expect(screen.getByTestId('cd-draw-button')).toBeInTheDocument());
      // Paying the fee to exchange zero cards is strictly worse than standing pat.
      expect(screen.getByTestId('cd-draw-button')).toBeDisabled();
    });

    it('shows the hand rank while there is still time to act on it', async () => {
      mockApi.mockResolvedValue({ ...drawPhaseState, playerHandRank: 1 });
      renderWithProviders(<CaribbeanDrawPage />);
      await waitFor(() => expect(screen.getByText(/ワンペア/)).toBeInTheDocument());
    });

    it('shows the stand-pat hint on a made hand when hints are enabled', async () => {
      localStorage.setItem('hint_enabled_caribbeandraw', 'true');
      mockApi.mockResolvedValue({ ...drawPhaseState, playerHandRank: 1 });
      renderWithProviders(<CaribbeanDrawPage />);
      await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toHaveTextContent('そのまま勝負'));
    });

    it('keeps the cards inert once the draw is over', async () => {
      mockApi.mockResolvedValue(actionPhaseState);
      renderWithProviders(<CaribbeanDrawPage />);
      await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());
      expect(screen.queryByTestId('cd-player-card-0')).not.toBeInTheDocument();
      expect(screen.queryByTestId('cd-draw-fee')).not.toBeInTheDocument();
    });
  });

  it('shows end phase with player wins', async () => {
    mockApi.mockResolvedValueOnce(actionPhaseState).mockResolvedValueOnce(endPhasePlayerWins);
    renderWithProviders(<CaribbeanDrawPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'コール' }));
    await waitFor(() => expect(screen.getByText('勝利！')).toBeInTheDocument());
    expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument();
  });

  it('shows end phase with dealer wins', async () => {
    mockApi.mockResolvedValueOnce(actionPhaseState).mockResolvedValueOnce(endPhaseDealerWins);
    renderWithProviders(<CaribbeanDrawPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'コール' }));
    await waitFor(() => expect(screen.getByText('ディーラー勝利！')).toBeInTheDocument());
  });

  it('shows end phase with fold', async () => {
    mockApi.mockResolvedValueOnce(actionPhaseState).mockResolvedValueOnce(endPhaseFold);
    renderWithProviders(<CaribbeanDrawPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'フォールド' }));
    await waitFor(() => expect(screen.getByText('フォールド')).toBeInTheDocument());
  });

  it('shows end phase with push', async () => {
    mockApi.mockResolvedValueOnce(actionPhaseState).mockResolvedValueOnce(endPhasePush);
    renderWithProviders(<CaribbeanDrawPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'コール' }));
    await waitFor(() => expect(screen.getByText('引き分け！')).toBeInTheDocument());
  });

  it('shows end phase with dealer not qualified', async () => {
    mockApi.mockResolvedValueOnce(actionPhaseState).mockResolvedValueOnce(endPhaseDealerNotQualified);
    renderWithProviders(<CaribbeanDrawPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'コール' }));
    await waitFor(() => expect(screen.getAllByText(/未クオリファイ/).length).toBeGreaterThanOrEqual(1));
  });

  it('shows payout breakdown with jackpot', async () => {
    mockApi.mockResolvedValue(endPhaseWithJackpot);
    renderWithProviders(<CaribbeanDrawPage />);
    await waitFor(() => expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument());
    expect(screen.getByText(/ジャックポット: 1000/)).toBeInTheDocument();
    expect(screen.getByText(/合計: 1800/)).toBeInTheDocument();
  });

  it('shows the draw fee in the result breakdown', async () => {
    mockApi.mockResolvedValue(endPhaseAfterDraw);
    renderWithProviders(<CaribbeanDrawPage />);
    await waitFor(() => expect(screen.getByTestId('cd-draw-cost')).toBeInTheDocument());
    // The fee is a cost and is deliberately not part of totalPayout, so without
    // this line the chip count moves by an amount the panel cannot explain.
    expect(screen.getByTestId('cd-draw-cost')).toHaveTextContent('交換手数料: -100');
    expect(screen.getByText(/合計: 800/)).toBeInTheDocument();
  });

  it('omits the draw fee line when the player stood pat', async () => {
    mockApi.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<CaribbeanDrawPage />);
    await waitFor(() => expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument());
    expect(screen.queryByTestId('cd-draw-cost')).not.toBeInTheDocument();
  });

  it('can change ante and jackpot amounts', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<CaribbeanDrawPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    const anteInput = screen.getByLabelText('アンテ');
    fireEvent.change(anteInput, { target: { value: '200' } });

    const jpInput = screen.getByLabelText('ジャックポット');
    fireEvent.change(jpInput, { target: { value: '10' } });

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', 200, 10));
  });

  it('steps the ante and jackpot amounts with the chip steppers', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<CaribbeanDrawPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    // Default ante 100 → +10 → 110; jackpot 0 → +10 → 10.
    fireEvent.click(screen.getByRole('button', { name: 'アンテ +10' }));
    fireEvent.click(screen.getByRole('button', { name: 'ジャックポット +10' }));
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', 110, 10));
  });

  it('disables the jackpot minus stepper at zero', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<CaribbeanDrawPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ジャックポット −10' })).toBeDisabled();
  });

  it('next game button executes reset without confirm dialog', async () => {
    mockApi.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<CaribbeanDrawPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  it('shows network error', async () => {
    mockApi.mockResolvedValueOnce(betPhaseState).mockRejectedValueOnce(new Error('Network'));
    renderWithProviders(<CaribbeanDrawPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('renders player and dealer cards in end phase', async () => {
    mockApi.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<CaribbeanDrawPage />);
    await waitFor(() => expect(screen.getAllByRole('img').length).toBe(10));
    expect(screen.getByText('🟡')).toBeInTheDocument();
    expect(screen.getByText('🔴')).toBeInTheDocument();
  });

  it('does not label unevaluated hand ranks as High Card', async () => {
    mockApi.mockResolvedValue({ ...endPhasePlayerWins, playerHandRank: -1, dealerHandRank: -1 });
    renderWithProviders(<CaribbeanDrawPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    // Ranks of -1 (unevaluated) must not fall back to the "High Card" label.
    expect(screen.queryByText(/ハイカード/)).not.toBeInTheDocument();
  });

  it('renders 1 face-up and 4 face-down dealer cards in action phase', async () => {
    mockApi.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(actionPhaseState);
    renderWithProviders(<CaribbeanDrawPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());

    // The 4 masked dealer cards are each announced as "hidden card".
    expect(screen.getAllByRole('img', { name: '非公開のカード' })).toHaveLength(4);
    // Dealer section is shown
    expect(screen.getByText('🔴')).toBeInTheDocument();
  });

  it('shows dealer qualification status', async () => {
    mockApi.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<CaribbeanDrawPage />);
    await waitFor(() => expect(screen.getByText(/クオリファイ/)).toBeInTheDocument());
  });

  it('renders hint toggle checkbox', async () => {
    mockApi.mockResolvedValue(actionPhaseState);
    renderWithProviders(<CaribbeanDrawPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());
    expect(screen.getByRole('checkbox')).toBeInTheDocument();
  });

  it('shows HintTooltip when hint is enabled in action phase', async () => {
    localStorage.setItem('hint_enabled_caribbeandraw', 'true');
    mockApi.mockResolvedValue(actionPhaseState);
    renderWithProviders(<CaribbeanDrawPage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  describe('session stats', () => {
    it('has no session summary before any round completes', async () => {
      mockApi.mockResolvedValue(betPhaseState);
      renderWithProviders(<CaribbeanDrawPage />);
      await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
      expect(screen.queryByTestId('cd-session-stats')).not.toBeInTheDocument();
    });

    it('records a win at END and shows the tally + positive net', async () => {
      // net = totalPayout(800) - (ante 100 + jackpot 0 + play 200) = +500
      mockApi.mockResolvedValue(endPhasePlayerWins);
      renderWithProviders(<CaribbeanDrawPage />);
      await waitFor(() => expect(screen.getByTestId('cd-session-stats')).toBeInTheDocument());
      expect(screen.getByTestId('cd-session-tally')).toHaveTextContent('1勝 0敗 0分');
      const netEl = screen.getByTestId('cd-session-net');
      expect(netEl).toHaveTextContent('収支: +500');
      expect(netEl).toHaveClass('text-ds-success');
    });

    it('records a loss at END and shows a negative net', async () => {
      // net = totalPayout(0) - (ante 100 + jackpot 0 + play 200) = -300
      mockApi.mockResolvedValue(endPhaseDealerWins);
      renderWithProviders(<CaribbeanDrawPage />);
      await waitFor(() => expect(screen.getByTestId('cd-session-stats')).toBeInTheDocument());
      expect(screen.getByTestId('cd-session-tally')).toHaveTextContent('0勝 1敗 0分');
      const netEl = screen.getByTestId('cd-session-net');
      expect(netEl).toHaveTextContent('収支: -300');
      expect(netEl).toHaveClass('text-ds-error');
    });

    it('does not double-count the same END round on re-render', async () => {
      mockApi.mockResolvedValue(endPhasePlayerWins);
      renderWithProviders(<CaribbeanDrawPage />);
      await waitFor(() => expect(screen.getByTestId('cd-session-stats')).toBeInTheDocument());
      expect(JSON.parse(localStorage.getItem('trumpcards-caribbeandraw-history') ?? '[]')).toHaveLength(1);
      // Re-render the page while it stays at END (toggling the hint checkbox
      // updates local state) — the record-once guard must not append again.
      fireEvent.click(screen.getByRole('checkbox'));
      await waitFor(() => expect(screen.getByTestId('cd-session-tally')).toHaveTextContent('1勝 0敗 0分'));
      expect(JSON.parse(localStorage.getItem('trumpcards-caribbeandraw-history') ?? '[]')).toHaveLength(1);
    });

    it('rehydrates prior session stats from localStorage on mount', async () => {
      localStorage.setItem(
        'trumpcards-caribbeandraw-history',
        JSON.stringify([
          { outcome: 1, net: 500 },
          { outcome: 2, net: -200 },
        ]),
      );
      mockApi.mockResolvedValue(betPhaseState);
      renderWithProviders(<CaribbeanDrawPage />);
      await waitFor(() => expect(screen.getByTestId('cd-session-stats')).toBeInTheDocument());
      expect(screen.getByTestId('cd-session-tally')).toHaveTextContent('1勝 1敗 0分');
      expect(screen.getByTestId('cd-session-net')).toHaveTextContent('収支: +300');
    });

    it('clears the session summary when the clear button is pressed', async () => {
      mockApi.mockResolvedValue(endPhasePlayerWins);
      renderWithProviders(<CaribbeanDrawPage />);
      await waitFor(() => expect(screen.getByTestId('cd-session-stats')).toBeInTheDocument());
      fireEvent.click(screen.getByTestId('cd-session-clear'));
      await waitFor(() => expect(screen.queryByTestId('cd-session-stats')).not.toBeInTheDocument());
      expect(localStorage.getItem('trumpcards-caribbeandraw-history')).toBeNull();
    });
  });
});

// --- keyboard shortcut execution (#4429) ---
// Each key is gated on its own phase, so the case pairs it with the fixture that
// enables it; the assertion pins the command and its arguments so a swapped
// binding fails rather than passing on "exec was called at all".
const kbdCases: [string, unknown[], CaribbeanDrawResponse][] = [
  ['b', ['bet', 100, 0], betPhaseState],
  ['s', ['draw', undefined, undefined, []], drawPhaseState],
  ['p', ['play'], actionPhaseState],
  ['f', ['fold'], actionPhaseState],
  ['r', ['reset'], endPhasePlayerWins],
];

describe('CaribbeanDrawPage keyboard shortcuts', () => {
  it.each(kbdCases)('pressing %s dispatches %j', async (key, expected, state) => {
    mockApi.mockResolvedValue(state);
    renderWithProviders(<CaribbeanDrawPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockApi.mockClear();
    mockApi.mockResolvedValue(state);
    fireEvent.keyDown(document, { key });
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith(...expected));
  });

  it('gates the d shortcut on there being something to exchange', async () => {
    mockApi.mockResolvedValue(drawPhaseState);
    renderWithProviders(<CaribbeanDrawPage />);
    await waitFor(() => expect(screen.getByTestId('cd-player-card-0')).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.keyDown(document, { key: 'd' });
    await flushPendingDispatch();
    // Nothing is selected, so 'd' would otherwise pay the fee to exchange nothing.
    expect(mockApi).not.toHaveBeenCalled();

    fireEvent.click(screen.getByTestId('cd-player-card-3'));
    fireEvent.keyDown(document, { key: 'd' });
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('draw', undefined, undefined, [3]));
  });

  it('ignores a key whose phase gate is closed', async () => {
    // 'b' places the ante and is enabled only in the BET phase.
    mockApi.mockResolvedValue(actionPhaseState);
    renderWithProviders(<CaribbeanDrawPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockApi.mockClear();
    fireEvent.keyDown(document, { key: 'b' });
    await flushPendingDispatch();
    expect(mockApi).not.toHaveBeenCalled();
  });
});
