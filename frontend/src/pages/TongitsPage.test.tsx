import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { tongitsApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, TongitsResponse } from '../types/card';
import { TongitsPhase } from '../types/phases';
import { TongitsPage } from './TongitsPage';

vi.mock('../api/gameApi', () => ({ tongitsApi: { exec: vi.fn() }, actionLogApi: { tongits: vi.fn() } }));
const mockExec = vi.mocked(tongitsApi.exec);
const card = (design: Card['design'], value: number): Card => ({ design, value });
const meld = (cards: Card[]) => ({ cards });

function state(overrides: Partial<TongitsResponse> = {}): TongitsResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 5,
        cards: [card('SPADE', 3), card('HEART', 3), card('CLOVER', 3), card('DIAMOND', 9), card('SPADE', 11)],
        melds: [],
        roundScore: 0,
        cumulativeScore: 0,
      },
      {
        id: 1,
        isHuman: false,
        cardCount: 5,
        cards: [],
        melds: [meld([card('HEART', 5), card('HEART', 6), card('HEART', 7)])],
        roundScore: 0,
        cumulativeScore: 0,
      },
      {
        id: 2,
        isHuman: false,
        cardCount: 4,
        cards: [],
        melds: [meld([card('DIAMOND', 8), card('DIAMOND', 9), card('DIAMOND', 10)])],
        roundScore: 0,
        cumulativeScore: 0,
      },
    ],
    phase: TongitsPhase.DISCARD,
    roundNumber: 1,
    currentPlayerIdx: 0,
    discardTop: card('HEART', 2),
    drawPileCount: 30,
    gameEndFlag: false,
    winnerIdx: -1,
    isTongits: false,
    remainingPoints: 3,
    message: '',
    config: { cpuDifficulty: 1, pointLimit: 50 },
    ...overrides,
  };
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(state());
});

describe('TongitsPage', () => {
  it('renders three player rows and every public meld', async () => {
    renderWithProviders(<TongitsPage />);
    expect(await screen.findByTestId('tongits-melds')).toBeInTheDocument();
    expect(screen.getByText('あなた')).toBeInTheDocument();
    expect(screen.getAllByText('CPU 1').length).toBeGreaterThanOrEqual(2);
    expect(screen.getAllByText('CPU 2').length).toBeGreaterThanOrEqual(2);
    const publicMelds = screen.getByTestId('tongits-melds');
    expect(within(publicMelds).getByText('CPU 1')).toBeInTheDocument();
    expect(within(publicMelds).getByText('CPU 2')).toBeInTheDocument();
  });

  it('melds the selected cards through the PlayerMeld API action', async () => {
    renderWithProviders(<TongitsPage />);
    for (const index of [0, 1, 2]) fireEvent.click(await screen.findByTestId(`tongits-hand-${index}`));
    fireEvent.click(screen.getByRole('button', { name: 'メルド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('meld', undefined, undefined, [0, 1, 2]));
  });

  it('passes the target player, meld number, and selected card to sapaw', async () => {
    renderWithProviders(<TongitsPage />);
    fireEvent.click(await screen.findByTestId('tongits-hand-3'));
    fireEvent.click(screen.getByRole('button', { name: 'CPU 1のメルド1にサパウ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('sapaw', 3, undefined, undefined, 1, 0));
  });

  it('declares a challenge with both agreement flags', async () => {
    renderWithProviders(<TongitsPage />);
    fireEvent.click(await screen.findByRole('button', { name: 'チャレンジ' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('challenge', undefined, undefined, undefined, undefined, undefined, [
        true,
        true,
      ]),
    );
  });

  it('shows the win celebration after a challenge ends the game', async () => {
    mockExec
      .mockResolvedValueOnce(state())
      .mockResolvedValueOnce(state({ phase: TongitsPhase.GAME_END, gameEndFlag: true, winnerIdx: 0 }));
    renderWithProviders(<TongitsPage />);
    fireEvent.click(await screen.findByRole('button', { name: 'チャレンジ' }));
    await waitFor(() => expect(screen.getByTestId('win-celebration')).toBeInTheDocument());
  });

  it('shows immediate victory when the human hand is Tongits', async () => {
    const winnerCards = [
      card('SPADE', 10),
      card('HEART', 10),
      card('CLOVER', 10),
      card('DIAMOND', 9),
      card('SPADE', 10),
    ];
    const initial = state();
    mockExec.mockResolvedValue(
      state({
        phase: TongitsPhase.ROUND_END,
        isTongits: true,
        winnerIdx: 0,
        players: [{ ...initial.players[0], cards: winnerCards }, ...initial.players.slice(1)],
      }),
    );
    renderWithProviders(<TongitsPage />);
    expect(await screen.findByTestId('tongits-on-deal-celebration')).toHaveAttribute('data-visible', 'true');
    expect(screen.getByText('TONGITS!')).toBeInTheDocument();
  });

  it('allows drawing the discard only during the draw phase', async () => {
    mockExec.mockResolvedValue(state({ phase: TongitsPhase.DRAW }));
    renderWithProviders(<TongitsPage />);
    const pile = await screen.findByTestId('tongits-discard-pile');
    expect(pile).toBeEnabled();
    fireEvent.click(pile);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawdiscard'));
  });

  it('disables discard controls during the draw phase', async () => {
    mockExec.mockResolvedValue(state({ phase: TongitsPhase.DRAW }));
    renderWithProviders(<TongitsPage />);
    expect(await screen.findByRole('button', { name: '山札から引く' })).toBeEnabled();
    expect(screen.queryByRole('button', { name: '捨てる' })).not.toBeInTheDocument();
  });

  it('does not draw from the discard pile during the discard phase', async () => {
    renderWithProviders(<TongitsPage />);
    const pile = await screen.findByTestId('tongits-discard-pile');
    expect(pile).toBeDisabled();
    fireEvent.click(pile);
    await waitFor(() => expect(mockExec).not.toHaveBeenCalledWith('drawdiscard'));
  });

  it('requires exactly one card before enabling discard', async () => {
    renderWithProviders(<TongitsPage />);
    const discard = await screen.findByRole('button', { name: '捨てる' });
    expect(discard).toBeDisabled();
    fireEvent.click(screen.getByTestId('tongits-hand-3'));
    expect(discard).toBeEnabled();
  });

  it('requires three selected cards before enabling meld', async () => {
    renderWithProviders(<TongitsPage />);
    const meldButton = await screen.findByRole('button', { name: 'メルド' });
    expect(meldButton).toBeDisabled();
    fireEvent.click(screen.getByTestId('tongits-hand-0'));
    fireEvent.click(screen.getByTestId('tongits-hand-1'));
    fireEvent.click(screen.getByTestId('tongits-hand-2'));
    expect(meldButton).toBeEnabled();
  });

  it('shows the remaining-point challenge status in discard phase', async () => {
    renderWithProviders(<TongitsPage />);
    expect(await screen.findByTestId('tongits-remaining-points')).toHaveTextContent('残り点: 3');
    expect(screen.getByTestId('tongits-remaining-points')).toHaveTextContent('チャレンジ可能');
  });

  it('offers next round after the round ends', async () => {
    mockExec.mockResolvedValue(state({ phase: TongitsPhase.ROUND_END }));
    renderWithProviders(<TongitsPage />);
    fireEvent.click(await screen.findByRole('button', { name: '次のラウンド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });
});
