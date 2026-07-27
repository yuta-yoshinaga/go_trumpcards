import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { spanish21Api } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BlackJackResponse } from '../types/card';
import { Spanish21Page } from './Spanish21Page';

vi.mock('../api/gameApi', () => ({
  spanish21Api: { exec: vi.fn() },
  actionLogApi: { spanish21: vi.fn() },
}));

const mockExec = vi.mocked(spanish21Api.exec);

const betPhaseState: BlackJackResponse = {
  dealer: { chips: 1000 },
  player: { chips: 1000 },
  phase: 1,
  currentHandIdx: 0,
  insuranceBet: 0,
  insuranceAvailable: false,
  message: '',
  hintEnabled: false,
  suggestedAction: 0,
  deckCount: 1,
  dealerHitsSoft17: false,
  countingEnabled: false,
  cpuPlayerCount: 0,
  runningCount: 0,
  trueCount: 0,
  perfectPairsBet: 0,
  twentyOnePlus3Bet: 0,
  doubleAfterSplit: true,
  countingSystem: 0,
  deckPenetration: 75,
  multiHandCount: 0,
  surrenderRule: 0,
};

const endPhaseWithBonus: BlackJackResponse = {
  ...betPhaseState,
  dealer: { score: 19, cards: [{ design: 'CLOVER', value: 9 }], chips: 1000 },
  player: { chips: 1150 },
  hands: [
    {
      score: 21,
      cards: [
        { design: 'SPADE', value: 7 },
        { design: 'SPADE', value: 7 },
        { design: 'SPADE', value: 7 },
      ],
      bet: 100,
      stood: true,
      doubled: false,
      busted: false,
      isBlackJack: false,
      canSplit: false,
      surrendered: false,
      canSurrender: false,
    },
  ],
  phase: 5,
  message: 'You are the winner.',
  bonuses: ['spanish21.bonus.777.spade'],
};

beforeEach(() => {
  mockExec.mockResolvedValue(betPhaseState);
});

describe('Spanish21Page', () => {
  it('resets via the Spanish 21 API on mount', async () => {
    renderWithProviders(<Spanish21Page />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
  });

  it('renders a bonus badge with the translated label, not the raw key', async () => {
    mockExec.mockResolvedValue(endPhaseWithBonus);
    renderWithProviders(<Spanish21Page />);
    const badge = await screen.findByTestId('bj-bonus-badge');
    // The fully-qualified backend key `spanish21.bonus.777.spade` resolves to
    // its ja translation; a broken namespace strip would leave the raw key.
    expect(badge).toHaveTextContent('7-7-7 (全スペード)');
    expect(badge).not.toHaveTextContent('spanish21.bonus');
    expect(badge).not.toHaveTextContent('bonus.777.spade');
  });

  it('shows the reset confirmation dialog when the next game button is clicked', async () => {
    mockExec.mockResolvedValue(endPhaseWithBonus);
    renderWithProviders(<Spanish21Page />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    expect(screen.getByText('本当にゲームをリセットしますか？')).toBeInTheDocument();
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('resets after confirming the dialog', async () => {
    mockExec.mockResolvedValue(endPhaseWithBonus);
    renderWithProviders(<Spanish21Page />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(betPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, expect.any(Object)));
  });

  it('does not reset when the confirmation dialog is cancelled', async () => {
    mockExec.mockResolvedValue(endPhaseWithBonus);
    renderWithProviders(<Spanish21Page />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    await waitFor(() => expect(screen.queryByText('本当にゲームをリセットしますか？')).not.toBeInTheDocument());
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('opens a Spanish 21 specific tutorial step describing the 48-card deck and bonuses', async () => {
    renderWithProviders(<Spanish21Page />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());

    // Steps: betControls → betButton → (Spanish 21) payout reference.
    fireEvent.click(screen.getByRole('button', { name: '次へ' }));
    fireEvent.click(screen.getByRole('button', { name: '次へ' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toHaveTextContent('48枚デッキ'));
  });
});
