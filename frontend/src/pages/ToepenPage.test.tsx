import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, toepenApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, ToepenPlayer, ToepenResponse } from '../types/card';
import { ToepenPhase } from '../types/phases';
import { cardAlt } from '../utils/cardAlt';
import { ToepenPage } from './ToepenPage';

vi.mock('../api/gameApi', () => ({
  toepenApi: { exec: vi.fn() },
  actionLogApi: { toepen: vi.fn() },
}));

const mockExec = vi.mocked(toepenApi.exec);
const mockActionLog = vi.mocked(actionLogApi.toepen);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function human(overrides?: Partial<ToepenPlayer>): ToepenPlayer {
  return {
    id: 0,
    isHuman: true,
    cardCount: 3,
    cards: [card('SPADE', 10), card('HEART', 11), card('CLOVER', 7)],
    lives: 0,
    folded: false,
    eliminated: false,
    hidden: false,
    ...overrides,
  };
}

function cpu(id: number, overrides?: Partial<ToepenPlayer>): ToepenPlayer {
  // A hidden seat arrives with a count and NO cards.
  return {
    id,
    isHuman: false,
    cardCount: 3,
    cards: [],
    lives: 0,
    folded: false,
    eliminated: false,
    hidden: true,
    ...overrides,
  };
}

function makeState(overrides?: Partial<ToepenResponse>): ToepenResponse {
  return {
    players: [human(), cpu(1), cpu(2), cpu(3)],
    phase: 0,
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    dealerIdx: 0,
    currentTrick: [],
    currentTrickWinnerIdx: -1,
    leadSuit: -1,
    trickNumber: 0,
    handNumber: 1,
    stake: 1,
    knockerIdx: -1,
    pendingRespondent: -1,
    lastTrickWinner: -1,
    maxLives: 10,
    validPlayIndices: [0, 1, 2],
    canRedeal: false,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    ...overrides,
  };
}

describe('ToepenPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(makeState());
  });

  it('resets on mount', async () => {
    renderWithProviders(<ToepenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('opens and closes the action log at hand end and restores focus to its trigger', async () => {
    mockExec.mockResolvedValue(makeState({ phase: ToepenPhase.HAND_END, lastTrickWinner: 2 }));
    mockActionLog.mockResolvedValue({ entries: [] });
    renderWithProviders(<ToepenPage />);

    const trigger = await screen.findByRole('button', { name: '棋譜を見る' });
    trigger.focus();
    fireEvent.click(trigger);

    expect(await screen.findByText('棋譜はありません。')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '閉じる' }));
    expect(screen.getByRole('button', { name: '棋譜を見る' })).toHaveFocus();
  });

  it('keeps the action log available after the game ends', async () => {
    mockExec.mockResolvedValue(makeState({ phase: ToepenPhase.GAME_END, gameEndFlag: true }));
    renderWithProviders(<ToepenPage />);

    expect(await screen.findByRole('button', { name: '棋譜を見る' })).toBeInTheDocument();
  });

  it('shows the inverted ranking permanently', async () => {
    // It is the one thing about this game that is easy to get backwards, so it
    // is on screen at all times rather than only in the tutorial.
    renderWithProviders(<ToepenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.getByText(/10 > 9 > 8 > 7 > A > K > Q > J/)).toBeInTheDocument();
  });

  it('never renders the opponent hands as cards', async () => {
    renderWithProviders(<ToepenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    const handButtons = screen.getAllByRole('button').filter((b) => b.dataset.hintAction === 'play');
    expect(handButtons).toHaveLength(3);
    const cpuHand = screen.getByRole('img', { name: /CPU1 .*手札/ });
    expect(cpuHand.querySelectorAll('img[src="/images/z01.png"]')).toHaveLength(3);
    expect(within(cpuHand).queryByRole('img', { name: cardAlt(card('HEART', 11)) })).not.toBeInTheDocument();
  });

  it('reveals the CPU hands only after the game ends', async () => {
    const cpuCards = [card('HEART', 12), card('DIAMOND', 13), card('SPADE', 1)];
    mockExec.mockResolvedValueOnce(
      makeState({
        phase: ToepenPhase.GAME_END,
        gameEndFlag: true,
        players: [human(), cpu(1, { cards: cpuCards, hidden: false }), cpu(2), cpu(3)],
      }),
    );
    renderWithProviders(<ToepenPage />);

    const cpuHand = await screen.findByRole('group', { name: 'CPU1 の手札（公開）' });
    for (const revealedCard of cpuCards) {
      expect(within(cpuHand).getByRole('img', { name: cardAlt(revealedCard) })).toBeInTheDocument();
    }
    expect(cpuHand.querySelectorAll('img[src^="/images/"]')).toHaveLength(3);
    expect(cpuHand.querySelector('img[src="/images/z01.png"]')).not.toBeInTheDocument();
  });

  it('labels each trick card with its player and marks folded players in text', async () => {
    mockExec.mockResolvedValue(
      makeState({
        currentTrick: [
          { playerIdx: 0, card: card('SPADE', 10) },
          { playerIdx: 2, card: card('HEART', 11) },
        ],
        players: [human(), cpu(1), cpu(2, { folded: true }), cpu(3)],
      }),
    );
    renderWithProviders(<ToepenPage />);

    const trick = await screen.findByTestId('toepen-current-trick');
    expect(trick).toHaveTextContent('あなた');
    expect(trick).toHaveTextContent('CPU2');
    expect(trick).toHaveTextContent('[降参]');
    expect(trick.querySelectorAll('[data-testid="toepen-trick-card"]')).toHaveLength(2);
  });

  it('visually marks and announces the server-reported current trick winner', async () => {
    mockExec.mockResolvedValue(
      makeState({
        currentTrick: [
          { playerIdx: 0, card: card('SPADE', 10) },
          { playerIdx: 2, card: card('HEART', 11) },
        ],
        currentTrickWinnerIdx: 2,
      }),
    );
    renderWithProviders(<ToepenPage />);
    const cards = await screen.findAllByTestId('toepen-trick-card');
    expect(cards[0]).toHaveAttribute('data-winning', 'false');
    expect(cards[1]).toHaveAttribute('data-winning', 'true');
    expect(cards[1].firstElementChild).toHaveAttribute('aria-label', expect.stringContaining('暫定勝ち札'));
  });

  it('only plays the cards the server marked legal', async () => {
    // The follow-suit obligation lives on the server; the page must not accept
    // a click on a card it did not offer.
    mockExec.mockResolvedValue(makeState({ validPlayIndices: [1] }));
    renderWithProviders(<ToepenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    const handButtons = screen.getAllByRole('button').filter((b) => b.dataset.hintAction === 'play');
    mockExec.mockClear();

    fireEvent.click(handButtons[0]);
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    fireEvent.click(handButtons[1]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 1));
  });

  it('raises the stake with toep', async () => {
    renderWithProviders(<ToepenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: /toep/i }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('toep'));
  });

  // **誰の判断に応答しているのか。** knockerIdx はサーバから届いているのに読んで
  // おらず、賭け点しか出していなかった。相手が複数いると誰に応答するのか分からない
  // (#5570)。CUI の respondLine は最初から名前を出している。
  it('names the player who declared the toep', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, pendingRespondent: 0, knockerIdx: 2, stake: 3 }));
    renderWithProviders(<ToepenPage />);

    const line = await screen.findByTestId('toepen-toeped-by');
    expect(line).toHaveTextContent('CPU2');
    // 賭け点は引き続き出ること。名前で置き換えては情報が減る。
    expect(line).toHaveTextContent('3');
  });

  // **負のコントロール: まだ誰も toep していなければ名前を出さない。**
  // knockerIdx は -1 で来る。ここで players[-1] を引くと undefined になる。
  it('falls back to the stake-only wording when nobody has toeped', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, pendingRespondent: 0, knockerIdx: -1, stake: 2 }));
    renderWithProviders(<ToepenPage />);

    const line = await screen.findByTestId('toepen-toeped-by');
    expect(line).toHaveTextContent('toep されました');
    expect(line).not.toHaveTextContent('CPU');
  });

  it('offers stay and fold only while a toep is on the human, and prices the fold below the stake', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1, pendingRespondent: 0, knockerIdx: 2, stake: 3 }));
    renderWithProviders(<ToepenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    // Folding costs the stake BEFORE the raise: 2, not 3.
    expect(screen.getByRole('button', { name: '降りる（2点）' })).toBeInTheDocument();

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '追随する' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('answer', undefined, true));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '降りる（2点）' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('answer', undefined, false));
  });

  it('offers the next hand only once the hand has settled', async () => {
    renderWithProviders(<ToepenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: '次のハンド' })).not.toBeInTheDocument();

    mockExec.mockResolvedValue(makeState({ phase: 2 }));
    renderWithProviders(<ToepenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のハンド' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のハンド' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('reports the outcome and marks an eliminated seat', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 3,
        gameEndFlag: true,
        winnerIdx: 0,
        messageCode: 'toepen.win',
        players: [human(), cpu(1, { lives: 10, eliminated: true }), cpu(2), cpu(3)],
      }),
    );
    renderWithProviders(<ToepenPage />);
    await waitFor(() => expect(screen.getByText('あなたの勝ちです')).toBeInTheDocument());
    expect(screen.getByText('10/10')).toBeInTheDocument();
  });

  it('explains why a card cannot be played', async () => {
    mockExec.mockResolvedValue(makeState({ validPlayIndices: [1] }));
    renderWithProviders(<ToepenPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    const buttons = screen.getAllByTestId(/^toepen-hand-/);
    const blocked = buttons[0] as HTMLElement;
    const allowed = buttons[1] as HTMLElement;
    expect(blocked).toHaveAttribute('title', expect.stringContaining('スート'));
    expect(blocked.getAttribute('aria-label')).toMatch(/スート/);
    expect(allowed).not.toHaveAttribute('title');
  });
});

describe('ToepenPage redeal', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('offers the redeal only when the server says it is available', async () => {
    // Whether a hand qualifies, and whether the window is still open, are both
    // the server's call -- the page must not count ranks itself.
    mockExec.mockResolvedValue(makeState({ canRedeal: false }));
    renderWithProviders(<ToepenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: '配り直し（貧民）' })).not.toBeInTheDocument();

    mockExec.mockResolvedValue(makeState({ canRedeal: true }));
    renderWithProviders(<ToepenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '配り直し（貧民）' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '配り直し（貧民）' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('redeal'));
  });

  it('displays redeal condition hint and title when canRedeal is true, and hides when false', async () => {
    mockExec.mockResolvedValue(makeState({ canRedeal: false }));
    renderWithProviders(<ToepenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    expect(screen.queryByTestId('toepen-redeal-hint')).not.toBeInTheDocument();

    mockExec.mockResolvedValue(makeState({ canRedeal: true }));
    renderWithProviders(<ToepenPage />);
    await waitFor(() => expect(screen.getByTestId('toepen-redeal-hint')).toBeInTheDocument());

    const hintEl = screen.getByTestId('toepen-redeal-hint');
    expect(hintEl).toHaveTextContent('手札が A/K/Q/J だけなので配り直しを要求できます');
    expect(hintEl).not.toHaveTextContent('redealHint');
    expect(hintEl).not.toHaveTextContent('{{');

    const button = screen.getByRole('button', { name: '配り直し（貧民）' });
    expect(button).toHaveAttribute('title', '手札が A/K/Q/J だけなので配り直しを要求できます');
  });

  describe('last trick winner display', () => {
    beforeEach(() => {
      vi.clearAllMocks();
      localStorage.clear();
    });

    it('displays the last trick winner when hand ends with a valid winner', async () => {
      mockExec.mockResolvedValue(makeState({ phase: 2 /* HAND_END */, lastTrickWinner: 2 }));
      renderWithProviders(<ToepenPage />);
      await waitFor(() => expect(screen.getByTestId('toepen-last-trick-winner')).toBeInTheDocument());
      expect(screen.getByTestId('toepen-last-trick-winner')).toHaveTextContent('最終トリック勝者: CPU2');
      expect(screen.getByTestId('toepen-last-trick-winner')).not.toHaveTextContent('{{');
    });

    it('does not display the last trick winner when the value is the sentinel (-1)', async () => {
      mockExec.mockResolvedValue(makeState({ phase: 2 /* HAND_END */, lastTrickWinner: -1 }));
      renderWithProviders(<ToepenPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalled());
      expect(screen.queryByTestId('toepen-last-trick-winner')).not.toBeInTheDocument();
    });
  });
});
