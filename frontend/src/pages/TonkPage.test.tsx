import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { tonkApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
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
    bestDeadwood: -1,
    knockThreshold: 5,
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
    // A visible, screen-reader-announced warning accompanies the button (not just the hover title).
    const warning = screen.getByTestId('tonk-undercut-warning');
    expect(warning).toHaveAttribute('role', 'status');
    const title = knock.getAttribute('title');
    expect(title).toBeTruthy();
    expect(warning.textContent).toContain(title as string);
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

  it('highlights a run and loses it when a middle card is discarded', async () => {
    const runHand = () =>
      makeState({
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 5,
            // A 4-card spade run plus one off card.
            cards: [card('SPADE', 5), card('SPADE', 6), card('SPADE', 7), card('SPADE', 8), card('HEART', 2)],
            roundScore: 0,
            cumulativeScore: 0,
          },
          { id: 1, isHuman: false, cardCount: 5, cards: [], roundScore: 0, cumulativeScore: 0 },
        ],
      });

    // Discard the off card (index 4) → the 5-6-7-8 run stays intact and lights up.
    mockExec.mockResolvedValue(runHand());
    const { unmount } = renderWithProviders(<TonkPage />);
    fireEvent.click(await screen.findByTestId('tonk-hand-4'));
    await waitFor(() => expect(screen.getByTestId('tonk-hand-0')).toHaveAttribute('data-meld', 'true'));
    expect(screen.getByTestId('tonk-hand-3')).toHaveAttribute('data-meld', 'true');
    unmount();

    // Discard a middle card (index 1, the 6) → remaining 5,7,8 is not a run → no highlight.
    mockExec.mockResolvedValue(runHand());
    renderWithProviders(<TonkPage />);
    fireEvent.click(await screen.findByTestId('tonk-hand-1'));
    await waitFor(() => expect(screen.getByTestId('tonk-hand-1')).toHaveAttribute('aria-pressed', 'true'));
    expect(screen.getByTestId('tonk-hand-0')).not.toHaveAttribute('data-meld');
    expect(screen.getByTestId('tonk-hand-2')).not.toHaveAttribute('data-meld');
  });

  const drawPhaseState = () => makeState({ phase: TonkPhase.DRAW });

  it('draws from the discard pile when its top card is clicked on the human draw turn', async () => {
    mockExec.mockResolvedValue(drawPhaseState());
    renderWithProviders(<TonkPage />);
    const pile = await screen.findByTestId('tonk-discard-pile');
    expect(pile).not.toBeDisabled();
    expect(pile.className).toContain('ring-ds-info');
    mockExec.mockClear();
    fireEvent.click(pile);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawdiscard'));
  });

  it('does not draw from the discard pile when it is not the draw phase', async () => {
    // Default makeState() is the DISCARD phase → the pile is inert/disabled.
    renderWithProviders(<TonkPage />);
    const pile = await screen.findByTestId('tonk-discard-pile');
    expect(pile).toBeDisabled();
    expect(pile.className).not.toContain('ring-ds-info');
    mockExec.mockClear();
    fireEvent.click(pile);
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('drawdiscard');
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
    expect(screen.queryByTestId('tonk-undercut-warning')).not.toBeInTheDocument();
  });

  // **CUI は毎ターン「ノック可能/不可」を出しているのに、Web はプレイヤーの
  // 手計算に任せていた (#4750)。**ノックボタンは1枚選択されていれば常に活性化
  // するので、実際に押すまで合法かどうか分からなかった。
  it('says whether the hand can knock right now', async () => {
    mockExec.mockResolvedValue(makeState({ bestDeadwood: 3, knockThreshold: 5 }));
    renderWithProviders(<TonkPage />);

    const badge = await screen.findByTestId('tonk-deadwood');
    expect(badge).toHaveTextContent('3');
    expect(badge).toHaveAttribute('data-knockable', 'true');
  });

  it('says when the hand cannot knock yet', async () => {
    mockExec.mockResolvedValue(makeState({ bestDeadwood: 9, knockThreshold: 5 }));
    renderWithProviders(<TonkPage />);

    const badge = await screen.findByTestId('tonk-deadwood');
    expect(badge).toHaveTextContent('9');
    expect(badge).not.toHaveAttribute('data-knockable');
  });

  // **-1 は「まだ聞くべき場面でない」印。**0 と混同すると、ドローフェーズで
  // 「デッドウッド0 = ノック可能」と誤って案内してしまう。
  it('shows nothing outside the human discard turn', async () => {
    mockExec.mockResolvedValue(makeState({ bestDeadwood: -1, knockThreshold: 5 }));
    renderWithProviders(<TonkPage />);

    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('tonk-deadwood')).not.toBeInTheDocument();
  });
});
