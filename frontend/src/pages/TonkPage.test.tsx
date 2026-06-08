import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { tonkApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, TonkResponse } from '../types/card';
import { TonkPhase } from '../types/phases';
import { TonkPage } from './TonkPage';

vi.mock('../api/gameApi', () => ({
  tonkApi: { exec: vi.fn() },
  actionLogApi: { tonk: vi.fn() },
}));

const mockExec = vi.mocked(tonkApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function makeState(overrides: Partial<TonkResponse> = {}): TonkResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 5,
        cards: [card('SPADE', 3), card('HEART', 5), card('DIAMOND', 7), card('CLOVER', 9), card('SPADE', 11)],
        roundScore: 0,
        cumulativeScore: 0,
      },
      { id: 1, isHuman: false, cardCount: 5, cards: [], roundScore: 0, cumulativeScore: 0 },
    ],
    phase: TonkPhase.DISCARD,
    roundNumber: 1,
    currentPlayerIdx: 0,
    discardTop: card('HEART', 2),
    drawPileCount: 30,
    gameEndFlag: false,
    winnerIdx: -1,
    knockerIdx: -1,
    knockerMelds: [],
    knockerDeadwood: [],
    opponentMelds: [],
    opponentDeadwood: [],
    isTonk: false,
    isUndercut: false,
    message: '',
    config: { cpuDifficulty: 1, pointLimit: 250 },
    ...overrides,
  };
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeState());
});

describe('TonkPage', () => {
  it('renders the Knock button without an undercut warning when opponents hold > 2 cards', async () => {
    renderWithProviders(<TonkPage />);
    const knock = await screen.findByRole('button', { name: /ノック/ });
    expect(knock).not.toHaveAttribute('data-undercut-risk');
    expect(knock).not.toHaveAttribute('title');
    expect(knock.className).not.toContain('ring-ds-warning');
    expect(knock.className).not.toContain('animate-pulse');
    expect(knock.textContent).not.toContain('⚠️');
  });

  it('flags the Knock button as undercut-risk when any opponent has ≤ 2 cards', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 5,
            cards: [card('SPADE', 3), card('HEART', 5), card('DIAMOND', 7), card('CLOVER', 9), card('SPADE', 11)],
            roundScore: 0,
            cumulativeScore: 0,
          },
          { id: 1, isHuman: false, cardCount: 2, cards: [], roundScore: 0, cumulativeScore: 0 },
        ],
      }),
    );
    renderWithProviders(<TonkPage />);
    const knock = await screen.findByRole('button', { name: /ノック/ });
    expect(knock).toHaveAttribute('data-undercut-risk', 'true');
    expect(knock).toHaveAttribute('title');
    expect(knock.className).toContain('ring-ds-warning');
    expect(knock.className).toContain('motion-safe:animate-pulse');
    expect(knock.textContent).toContain('⚠️');
  });

  const meldHandState = () =>
    makeState({
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 5,
          // Three 7s (a set) plus two unrelated cards.
          cards: [card('SPADE', 7), card('HEART', 7), card('CLOVER', 7), card('DIAMOND', 4), card('HEART', 9)],
          roundScore: 0,
          cumulativeScore: 0,
        },
        { id: 1, isHuman: false, cardCount: 5, cards: [], roundScore: 0, cumulativeScore: 0 },
      ],
    });

  it('highlights cards that form a meld once a discard candidate is selected', async () => {
    mockExec.mockResolvedValue(meldHandState());
    renderWithProviders(<TonkPage />);
    // No highlight before a discard is chosen.
    await waitFor(() => expect(screen.getByTestId('tonk-hand-0')).toBeInTheDocument());
    expect(screen.getByTestId('tonk-hand-0')).not.toHaveAttribute('data-meld');

    // Select the non-meld card (index 4) as the discard → the three 7s light up.
    fireEvent.click(screen.getByTestId('tonk-hand-4'));
    await waitFor(() => expect(screen.getByTestId('tonk-hand-0')).toHaveAttribute('data-meld', 'true'));
    expect(screen.getByTestId('tonk-hand-1')).toHaveAttribute('data-meld', 'true');
    expect(screen.getByTestId('tonk-hand-2')).toHaveAttribute('data-meld', 'true');
    expect(screen.getByTestId('tonk-hand-3')).not.toHaveAttribute('data-meld');
  });

  it('shows no meld highlight when discarding a meld card breaks the set', async () => {
    mockExec.mockResolvedValue(meldHandState());
    renderWithProviders(<TonkPage />);
    // Discard one of the 7s → only two 7s remain among the other four → no meld.
    fireEvent.click(await screen.findByTestId('tonk-hand-0'));
    await waitFor(() => expect(screen.getByTestId('tonk-hand-0')).toHaveAttribute('aria-pressed', 'true'));
    expect(screen.getByTestId('tonk-hand-1')).not.toHaveAttribute('data-meld');
    expect(screen.getByTestId('tonk-hand-2')).not.toHaveAttribute('data-meld');
  });

  it('clears the warning when opponents drop back above the threshold', async () => {
    // First render: opponent at 2 → warning on.
    mockExec.mockResolvedValueOnce(
      makeState({
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 5,
            cards: [card('SPADE', 3), card('HEART', 5), card('DIAMOND', 7), card('CLOVER', 9), card('SPADE', 11)],
            roundScore: 0,
            cumulativeScore: 0,
          },
          { id: 1, isHuman: false, cardCount: 2, cards: [], roundScore: 0, cumulativeScore: 0 },
        ],
      }),
    );
    renderWithProviders(<TonkPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /ノック/ })).toHaveAttribute('data-undercut-risk', 'true'),
    );

    // Reset via the page's reset flow returns a state with opponent at 5 → warning gone.
    mockExec.mockResolvedValueOnce(makeState());
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /ノック/ })).not.toHaveAttribute('data-undercut-risk'),
    );
  });
});
