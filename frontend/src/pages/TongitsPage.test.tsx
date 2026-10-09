import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { tongitsApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, TongitsResponse } from '../types/card';
import { TongitsPhase } from '../types/phases';
import { publicMeldRowKey, TongitsPage } from './TongitsPage';

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
    roundEndReason: 0,
    roundWinner: -1,
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
  it('labels the score table with a translated caption and keeps its column headers', async () => {
    renderWithProviders(<TongitsPage />);

    const scoreTable = await screen.findByRole('table', { name: '現在のスコア表' });
    expect(within(scoreTable).getByRole('columnheader', { name: 'プレイヤー' })).toBeInTheDocument();
    expect(within(scoreTable).getByRole('columnheader', { name: 'ラウンド' })).toBeInTheDocument();
    expect(within(scoreTable).getByRole('columnheader', { name: '累計' })).toBeInTheDocument();
  });

  it('uses the player ID to make public meld row keys unique', () => {
    expect(publicMeldRowKey(1, 0)).not.toBe(publicMeldRowKey(2, 0));
  });

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

  it('keeps public meld rows unique when players have the same meld number', async () => {
    const initial = state();
    mockExec.mockResolvedValue(
      state({
        players: [
          initial.players[0],
          { ...initial.players[1], melds: [meld([card('HEART', 5), card('HEART', 6), card('HEART', 7)])] },
          { ...initial.players[2], melds: [meld([card('DIAMOND', 8), card('DIAMOND', 9), card('DIAMOND', 10)])] },
        ],
      }),
    );
    renderWithProviders(<TongitsPage />);
    const publicMelds = await screen.findByTestId('tongits-melds');

    expect(within(publicMelds).getByText('CPU 1')).toBeInTheDocument();
    expect(within(publicMelds).getByText('CPU 2')).toBeInTheDocument();
    expect(within(publicMelds).getAllByTestId('animated-card')).toHaveLength(6);
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

  it('announces the updated stock count only after drawing from the stock', async () => {
    mockExec
      .mockResolvedValueOnce(state({ phase: TongitsPhase.DRAW }))
      .mockResolvedValueOnce(state({ phase: TongitsPhase.DISCARD, drawPileCount: 29 }));
    renderWithProviders(<TongitsPage />);
    const announcement = await screen.findByTestId('tongits-draw-pile-announcement');
    expect(announcement).toBeEmptyDOMElement();
    fireEvent.click(await screen.findByRole('button', { name: '山札から引く' }));
    await waitFor(() => expect(announcement).toHaveTextContent('山札の残りは29枚です'));
  });

  it('does not announce stock count changes caused by another action', async () => {
    mockExec
      .mockResolvedValueOnce(state())
      .mockResolvedValueOnce(state({ phase: TongitsPhase.DISCARD, drawPileCount: 29 }));
    renderWithProviders(<TongitsPage />);
    const announcement = await screen.findByTestId('tongits-draw-pile-announcement');
    fireEvent.click(await screen.findByRole('button', { name: 'チャレンジ' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('challenge', undefined, undefined, undefined, undefined, undefined, [
        true,
        true,
      ]),
    );
    expect(announcement).toBeEmptyDOMElement();
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

  it.each([
    [1, 'トンギッツ'],
    [2, 'チャレンジ'],
    [3, '山札切れ'],
  ])('shows the round end reason %s and keeps next round available', async (reason, label) => {
    mockExec.mockResolvedValue(state({ phase: TongitsPhase.ROUND_END, roundEndReason: reason, roundWinner: 1 }));
    renderWithProviders(<TongitsPage />);
    const result = await screen.findByTestId('tongits-round-result');
    expect(result).toHaveTextContent(`終了理由: ${label}`);
    expect(result).toHaveTextContent('勝者: CPU 1');
    expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument();
  });

  it('shows a draw for a stock out round', async () => {
    mockExec.mockResolvedValue(state({ phase: TongitsPhase.ROUND_END, roundEndReason: 3, roundWinner: -1 }));
    renderWithProviders(<TongitsPage />);
    expect(await screen.findByTestId('tongits-round-result')).toHaveTextContent('勝者: 引き分け');
  });

  it('announces round scores in the persistent live region at round end', async () => {
    const roundEndState = state({
      phase: TongitsPhase.ROUND_END,
      players: state().players.map((player, index) => ({
        ...player,
        roundScore: index + 1,
        cumulativeScore: 10 + index,
      })),
    });
    mockExec.mockResolvedValueOnce(state()).mockResolvedValueOnce(roundEndState);
    renderWithProviders(<TongitsPage />);

    await screen.findByTestId('tongits-melds');
    const announcement = screen.getByTestId('tongits-round-score-announcement');
    const liveRegion = announcement.querySelector('[aria-live="polite"]');
    expect(liveRegion).toBeInTheDocument();
    expect(liveRegion).toBeEmptyDOMElement();
    fireEvent.click(await screen.findByRole('button', { name: 'チャレンジ' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('challenge', undefined, undefined, undefined, undefined, undefined, [
        true,
        true,
      ]),
    );
    await waitFor(() => {
      expect(liveRegion).toHaveTextContent('+1');
      expect(liveRegion).toHaveTextContent('+2');
      expect(liveRegion).toHaveTextContent('+3');
    });
    expect(liveRegion).toHaveTextContent('+1');
    expect(liveRegion).toHaveTextContent('+2');
    expect(announcement.querySelector('[aria-live="polite"]')).toBe(liveRegion);
  });
});
