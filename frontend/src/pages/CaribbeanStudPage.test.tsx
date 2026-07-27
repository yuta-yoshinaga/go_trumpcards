import { fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { caribbeanstudApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, CaribbeanStudResponse } from '../types/card';
import { CaribbeanStudPage } from './CaribbeanStudPage';

vi.mock('../api/gameApi', () => ({
  caribbeanstudApi: { exec: vi.fn() },
  actionLogApi: { caribbeanstud: vi.fn() },
}));

const mockApi = vi.mocked(caribbeanstudApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const betPhaseState: CaribbeanStudResponse = {
  playerHand: [],
  dealerHand: [],
  phase: 1,
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
  playerHandRank: 0,
  dealerHandRank: 0,
  message: '',
};

const maskedCard = { design: '' as CardDesign, value: 0 };

const actionPhaseState: CaribbeanStudResponse = {
  ...betPhaseState,
  phase: 2,
  playerHand: [card('SPADE', 10), card('HEART', 11), card('DIAMOND', 13), card('CLOVER', 5), card('SPADE', 7)],
  // 1 face-up dealer card + 4 masked cards (security: dealer hand not fully revealed)
  dealerHand: [card('HEART', 13), maskedCard, maskedCard, maskedCard, maskedCard],
  anteBet: 100,
  chips: 900,
};

const endPhasePlayerWins: CaribbeanStudResponse = {
  playerHand: [card('SPADE', 7), card('CLOVER', 7), card('HEART', 7), card('DIAMOND', 4), card('SPADE', 2)],
  dealerHand: [card('CLOVER', 5), card('DIAMOND', 5), card('HEART', 8), card('SPADE', 11), card('DIAMOND', 1)],
  phase: 3,
  chips: 1500,
  anteBet: 100,
  jackpotBet: 0,
  playBet: 200,
  result: 1,
  antePayout: 200,
  playPayout: 800,
  jackpotPayout: 0,
  totalPayout: 1000,
  dealerQualified: true,
  playerHandRank: 3,
  dealerHandRank: 1,
  message: '勝利！',
  messageCode: 'caribbeanstud.result.playerWins',
};

const endPhaseDealerWins: CaribbeanStudResponse = {
  ...endPhasePlayerWins,
  result: -1,
  antePayout: 0,
  playPayout: 0,
  totalPayout: 0,
  message: 'ディーラー勝利！',
  messageCode: 'caribbeanstud.result.dealerWins',
};

const endPhaseFold: CaribbeanStudResponse = {
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
  messageCode: 'caribbeanstud.result.fold',
};

const endPhasePush: CaribbeanStudResponse = {
  ...endPhasePlayerWins,
  result: 0,
  antePayout: 100,
  playPayout: 200,
  totalPayout: 300,
  message: '引き分け！',
  messageCode: 'caribbeanstud.result.push',
};

const endPhaseDealerNotQualified: CaribbeanStudResponse = {
  ...endPhasePlayerWins,
  dealerQualified: false,
  antePayout: 200,
  playPayout: 200,
  totalPayout: 400,
  message: 'ディーラー未クオリファイ！',
  messageCode: 'caribbeanstud.result.dealerNotQualified',
};

const endPhaseWithJackpot: CaribbeanStudResponse = {
  ...endPhasePlayerWins,
  jackpotBet: 10,
  jackpotPayout: 1000,
  totalPayout: 2000,
};

beforeEach(() => {
  vi.clearAllMocks();
});

afterEach(() => {
  localStorage.clear();
});

describe('CaribbeanStudPage', () => {
  it('renders bet phase on mount', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument();
  });

  it('shows an accessible jackpot help in the bet phase', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<CaribbeanStudPage />);
    const help = await screen.findByTestId('jackpot-help');
    expect(help).toHaveTextContent('ジャックポットとは？');
    // Native <details>/<summary> is keyboard-focusable.
    expect(help.querySelector('summary')).toBeInTheDocument();
  });

  it('renders skeleton before state loads', () => {
    mockApi.mockReturnValue(new Promise(() => {}));
    renderWithProviders(<CaribbeanStudPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('shows action phase with call and fold buttons', async () => {
    mockApi.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(actionPhaseState);
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument();
  });

  it('shows end phase with player wins', async () => {
    mockApi.mockResolvedValueOnce(actionPhaseState).mockResolvedValueOnce(endPhasePlayerWins);
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'コール' }));
    await waitFor(() => expect(screen.getByText('勝利！')).toBeInTheDocument());
    expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument();
  });

  it('shows end phase with dealer wins', async () => {
    mockApi.mockResolvedValueOnce(actionPhaseState).mockResolvedValueOnce(endPhaseDealerWins);
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'コール' }));
    await waitFor(() => expect(screen.getByText('ディーラー勝利！')).toBeInTheDocument());
  });

  it('shows end phase with fold', async () => {
    mockApi.mockResolvedValueOnce(actionPhaseState).mockResolvedValueOnce(endPhaseFold);
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'フォールド' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'フォールド' }));
    await waitFor(() => expect(screen.getByText('フォールド')).toBeInTheDocument());
  });

  it('shows end phase with push', async () => {
    mockApi.mockResolvedValueOnce(actionPhaseState).mockResolvedValueOnce(endPhasePush);
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'コール' }));
    await waitFor(() => expect(screen.getByText('引き分け！')).toBeInTheDocument());
  });

  it('shows end phase with dealer not qualified', async () => {
    mockApi.mockResolvedValueOnce(actionPhaseState).mockResolvedValueOnce(endPhaseDealerNotQualified);
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'コール' }));
    await waitFor(() => expect(screen.getAllByText(/未クオリファイ/).length).toBeGreaterThanOrEqual(1));
  });

  it('shows payout breakdown with jackpot', async () => {
    mockApi.mockResolvedValue(endPhaseWithJackpot);
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getByTestId('payout-breakdown')).toBeInTheDocument());
    expect(screen.getByText(/ジャックポット: 1000/)).toBeInTheDocument();
    expect(screen.getByText(/合計: 2000/)).toBeInTheDocument();
  });

  it('can change ante and jackpot amounts', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<CaribbeanStudPage />);
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
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    // Default ante 100 → +10 → 110; jackpot 0 → +10 → 10.
    fireEvent.click(screen.getByRole('button', { name: 'アンテ +10' }));
    fireEvent.click(screen.getByRole('button', { name: 'ジャックポット +10' }));
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('bet', 110, 10));
  });

  it('disables the jackpot minus stepper at zero', async () => {
    mockApi.mockResolvedValue(betPhaseState);
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ジャックポット −10' })).toBeDisabled();
  });

  it('next game button executes reset without confirm dialog', async () => {
    mockApi.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    mockApi.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith('reset'));
  });

  it('shows network error', async () => {
    mockApi.mockResolvedValueOnce(betPhaseState).mockRejectedValueOnce(new Error('Network'));
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('renders player and dealer cards in end phase', async () => {
    mockApi.mockResolvedValue(endPhasePlayerWins);
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getAllByRole('img').length).toBe(10));
    expect(screen.getByText('🟡')).toBeInTheDocument();
    expect(screen.getByText('🔴')).toBeInTheDocument();
  });

  it('does not label unevaluated hand ranks as High Card', async () => {
    mockApi.mockResolvedValue({ ...endPhasePlayerWins, playerHandRank: -1, dealerHandRank: -1 });
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    // Ranks of -1 (unevaluated) must not fall back to the "High Card" label.
    expect(screen.queryByText(/ハイカード/)).not.toBeInTheDocument();
  });

  it('renders 1 face-up and 4 face-down dealer cards in action phase', async () => {
    mockApi.mockResolvedValueOnce(betPhaseState).mockResolvedValueOnce(actionPhaseState);
    renderWithProviders(<CaribbeanStudPage />);
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
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getByText(/クオリファイ/)).toBeInTheDocument());
  });

  it('renders hint toggle checkbox', async () => {
    mockApi.mockResolvedValue(actionPhaseState);
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'コール' })).toBeInTheDocument());
    expect(screen.getByRole('checkbox')).toBeInTheDocument();
  });

  it('shows HintTooltip when hint is enabled in action phase', async () => {
    localStorage.setItem('hint_enabled_caribbeanstud', 'true');
    mockApi.mockResolvedValue(actionPhaseState);
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  describe('session stats', () => {
    it('has no session summary before any round completes', async () => {
      mockApi.mockResolvedValue(betPhaseState);
      renderWithProviders(<CaribbeanStudPage />);
      await waitFor(() => expect(screen.getByText('チップ: 1000')).toBeInTheDocument());
      expect(screen.queryByTestId('csp-session-stats')).not.toBeInTheDocument();
    });

    it('records a win at END and shows the tally + positive net', async () => {
      // net = totalPayout(1000) - (ante 100 + jackpot 0 + play 200) = +700
      mockApi.mockResolvedValue(endPhasePlayerWins);
      renderWithProviders(<CaribbeanStudPage />);
      await waitFor(() => expect(screen.getByTestId('csp-session-stats')).toBeInTheDocument());
      expect(screen.getByTestId('csp-session-tally')).toHaveTextContent('1勝 0敗 0分');
      const netEl = screen.getByTestId('csp-session-net');
      expect(netEl).toHaveTextContent('収支: +700');
      expect(netEl).toHaveClass('text-ds-success');
    });

    it('records a loss at END and shows a negative net', async () => {
      // net = totalPayout(0) - (ante 100 + jackpot 0 + play 200) = -300
      mockApi.mockResolvedValue(endPhaseDealerWins);
      renderWithProviders(<CaribbeanStudPage />);
      await waitFor(() => expect(screen.getByTestId('csp-session-stats')).toBeInTheDocument());
      expect(screen.getByTestId('csp-session-tally')).toHaveTextContent('0勝 1敗 0分');
      const netEl = screen.getByTestId('csp-session-net');
      expect(netEl).toHaveTextContent('収支: -300');
      expect(netEl).toHaveClass('text-ds-error');
    });

    it('does not double-count the same END round on re-render', async () => {
      mockApi.mockResolvedValue(endPhasePlayerWins);
      renderWithProviders(<CaribbeanStudPage />);
      await waitFor(() => expect(screen.getByTestId('csp-session-stats')).toBeInTheDocument());
      expect(JSON.parse(localStorage.getItem('trumpcards-caribbeanstud-history') ?? '[]')).toHaveLength(1);
      // Re-render the page while it stays at END (toggling the hint checkbox
      // updates local state) — the record-once guard must not append again.
      fireEvent.click(screen.getByRole('checkbox'));
      await waitFor(() => expect(screen.getByTestId('csp-session-tally')).toHaveTextContent('1勝 0敗 0分'));
      expect(JSON.parse(localStorage.getItem('trumpcards-caribbeanstud-history') ?? '[]')).toHaveLength(1);
    });

    it('rehydrates prior session stats from localStorage on mount', async () => {
      localStorage.setItem(
        'trumpcards-caribbeanstud-history',
        JSON.stringify([
          { outcome: 1, net: 500 },
          { outcome: 2, net: -200 },
        ]),
      );
      mockApi.mockResolvedValue(betPhaseState);
      renderWithProviders(<CaribbeanStudPage />);
      await waitFor(() => expect(screen.getByTestId('csp-session-stats')).toBeInTheDocument());
      expect(screen.getByTestId('csp-session-tally')).toHaveTextContent('1勝 1敗 0分');
      expect(screen.getByTestId('csp-session-net')).toHaveTextContent('収支: +300');
    });

    it('clears the session summary when the clear button is pressed', async () => {
      mockApi.mockResolvedValue(endPhasePlayerWins);
      renderWithProviders(<CaribbeanStudPage />);
      await waitFor(() => expect(screen.getByTestId('csp-session-stats')).toBeInTheDocument());
      fireEvent.click(screen.getByTestId('csp-session-clear'));
      await waitFor(() => expect(screen.queryByTestId('csp-session-stats')).not.toBeInTheDocument());
      expect(localStorage.getItem('trumpcards-caribbeanstud-history')).toBeNull();
    });
  });
});

// --- keyboard shortcut execution (#4429) ---
// Each key is gated on its own phase, so the case pairs it with the fixture that
// enables it; the assertion pins the command and its arguments so a swapped
// binding fails rather than passing on "exec was called at all".
const kbdCases: [string, unknown[], CaribbeanStudResponse][] = [
  ['b', ['bet', 100, 0], betPhaseState],
  ['p', ['play'], actionPhaseState],
  ['f', ['fold'], actionPhaseState],
  ['r', ['reset'], endPhasePlayerWins],
];

describe('CaribbeanStudPage keyboard shortcuts', () => {
  it.each(kbdCases)('pressing %s dispatches %j', async (key, expected, state) => {
    mockApi.mockResolvedValue(state);
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockApi.mockClear();
    mockApi.mockResolvedValue(state);
    fireEvent.keyDown(document, { key });
    await waitFor(() => expect(mockApi).toHaveBeenCalledWith(...expected));
  });

  it('ignores a key whose phase gate is closed', async () => {
    // 'b' places the ante and is enabled only in the BET phase.
    mockApi.mockResolvedValue(actionPhaseState);
    renderWithProviders(<CaribbeanStudPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockApi.mockClear();
    fireEvent.keyDown(document, { key: 'b' });
    await flushPendingDispatch();
    expect(mockApi).not.toHaveBeenCalled();
  });
});
