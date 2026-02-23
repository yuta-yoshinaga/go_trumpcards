import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { blackjackApi } from '../api/gameApi';
import type { BlackJackResponse } from '../types/card';
import { BlackJackPage } from './BlackJackPage';

vi.mock('../api/gameApi', () => ({
  blackjackApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(blackjackApi.exec);

const betPhaseState: BlackJackResponse = {
  dealer: { chips: 1000 },
  player: { chips: 1000 },
  phase: 1,
  currentHandIdx: 0,
  insuranceBet: 0,
  insuranceAvailable: false,
  message: '',
};

const actionPhaseState: BlackJackResponse = {
  dealer: { cards: [{ design: 'SPADE', value: 1 }], chips: 1000 },
  player: { chips: 900 },
  hands: [
    {
      score: 15,
      cards: [
        { design: 'HEART', value: 5 },
        { design: 'DIAMOND', value: 10 },
      ],
      bet: 100,
      stood: false,
      doubled: false,
      busted: false,
      isBlackJack: false,
      canSplit: false,
    },
  ],
  phase: 4,
  currentHandIdx: 0,
  insuranceBet: 0,
  insuranceAvailable: false,
  message: '',
};

const insurancePhaseState: BlackJackResponse = {
  dealer: { cards: [{ design: 'SPADE', value: 1 }], chips: 1000 },
  player: { chips: 900 },
  hands: [
    {
      score: 15,
      cards: [
        { design: 'HEART', value: 5 },
        { design: 'DIAMOND', value: 10 },
      ],
      bet: 100,
      stood: false,
      doubled: false,
      busted: false,
      isBlackJack: false,
      canSplit: false,
    },
  ],
  phase: 3,
  currentHandIdx: 0,
  insuranceBet: 0,
  insuranceAvailable: true,
  message: '',
};

const endPhaseState: BlackJackResponse = {
  dealer: {
    score: 19,
    cards: [
      { design: 'CLOVER', value: 9 },
      { design: 'DIAMOND', value: 10 },
    ],
    chips: 1000,
  },
  player: { chips: 1150 },
  hands: [
    {
      score: 21,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 10 },
      ],
      bet: 100,
      stood: true,
      doubled: false,
      busted: false,
      isBlackJack: true,
      canSplit: false,
    },
  ],
  phase: 5,
  currentHandIdx: 0,
  insuranceBet: 0,
  insuranceAvailable: false,
  message: 'You are the winner.',
};

beforeEach(() => {
  mockExec.mockResolvedValue(betPhaseState);
});

describe('BlackJackPage', () => {
  it('calls reset command on mount', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined));
  });

  it('shows chip info bar', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText(/プレイヤー: 1000 chips/)).toBeInTheDocument());
    expect(screen.getByText(/ディーラー: 1000 chips/)).toBeInTheDocument();
  });

  it('shows bet button in bet phase', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
  });

  it('shows bet amount input in bet phase', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByLabelText('ベット額:')).toBeInTheDocument());
  });

  it('calls bet command when bet button is clicked', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined));
    mockExec.mockClear();
    mockExec.mockResolvedValue(actionPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 10));
  });

  it('shows hit and stand buttons in action phase', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒット' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'スタンド' })).toBeInTheDocument();
  });

  it('shows double down button when 2 cards and sufficient chips', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ダブルダウン' })).toBeInTheDocument());
  });

  it('shows split button when canSplit and sufficient chips', async () => {
    const splitState = {
      ...actionPhaseState,
      hands: [
        {
          ...actionPhaseState.hands?.[0],
          canSplit: true,
        },
      ],
    };
    mockExec.mockResolvedValue(splitState);
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'スプリット' })).toBeInTheDocument());
  });

  it('does not show split button when canSplit is false', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒット' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'スプリット' })).not.toBeInTheDocument();
  });

  it('shows insurance buttons in insurance phase', async () => {
    mockExec.mockResolvedValue(insurancePhaseState);
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'インシュランス' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '辞退' })).toBeInTheDocument();
  });

  it('calls insurance command when insurance button is clicked', async () => {
    mockExec.mockResolvedValue(insurancePhaseState);
    render(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined));
    mockExec.mockClear();
    mockExec.mockResolvedValue(actionPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'インシュランス' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('insurance', undefined));
  });

  it('calls declineinsurance command when decline button is clicked', async () => {
    mockExec.mockResolvedValue(insurancePhaseState);
    render(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined));
    mockExec.mockClear();
    mockExec.mockResolvedValue(actionPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '辞退' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('declineinsurance', undefined));
  });

  it('shows reset button in end phase', async () => {
    mockExec.mockResolvedValue(endPhaseState);
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());
  });

  it('shows message overlay when message is non-empty', async () => {
    mockExec.mockResolvedValue(endPhaseState);
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText('You are the winner.')).toBeInTheDocument());
  });

  it('does not show message overlay when message is empty', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText('ヒット')).toBeInTheDocument());
    expect(screen.queryByText('You are the winner.')).not.toBeInTheDocument();
  });

  it('calls hit command when Hit button is clicked', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    render(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined));
    mockExec.mockClear();
    mockExec.mockResolvedValue(actionPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'ヒット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hit', undefined));
  });

  it('calls stand command when Stand button is clicked', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    render(<BlackJackPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined));
    mockExec.mockClear();
    mockExec.mockResolvedValue(endPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'スタンド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('stand', undefined));
  });

  it('displays player score and bet', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText(/スコア 15/)).toBeInTheDocument());
    expect(screen.getByText(/ベット 100/)).toBeInTheDocument();
  });

  it('displays dealer score when non-zero in end phase', async () => {
    mockExec.mockResolvedValue(endPhaseState);
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText(/スコア 19/)).toBeInTheDocument());
  });

  it('shows card back when dealer score is zero', async () => {
    mockExec.mockResolvedValue(actionPhaseState);
    render(<BlackJackPage />);
    await waitFor(() => {
      const imgs = screen.getAllByRole('img');
      const cardBackImg = imgs.find((img) => img.getAttribute('src') === '/images/z01.png');
      expect(cardBackImg).toBeInTheDocument();
    });
  });

  it('shows BJ flag for blackjack hand', async () => {
    mockExec.mockResolvedValue(endPhaseState);
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByText(/\[BJ\]/)).toBeInTheDocument());
  });

  it('does not show dealer area in bet phase', async () => {
    render(<BlackJackPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ベット' })).toBeInTheDocument());
    expect(screen.queryByText('ディーラー手札')).not.toBeInTheDocument();
  });
});
