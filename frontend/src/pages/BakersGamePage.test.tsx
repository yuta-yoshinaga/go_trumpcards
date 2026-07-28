import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, bakersgameApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, FreeCellResponse } from '../types/card';
import { BakersGamePage } from './BakersGamePage';

vi.mock('../api/gameApi', () => ({
  bakersgameApi: { exec: vi.fn() },
  actionLogApi: { bakersgame: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(bakersgameApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: FreeCellResponse = {
  tableau: [[card('SPADE', 13)], [card('HEART', 12)], [], [], [], [], [], []],
  freeCells: [null, null, null, null],
  foundation: [[], [], [], []],
  phase: 0,
  moveCount: 5,
  canUndo: true,
  isStalemate: false,
  message: '',
  messageCode: 'bakersgame.playing',
};

const gameClearState: FreeCellResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'bakersgame.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: FreeCellResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'bakersgame.gameOver',
};

const withFoundationState: FreeCellResponse = {
  ...playingState,
  foundation: [[card('SPADE', 1)], [], [card('HEART', 1), card('HEART', 2)], []],
};

const withHintState: FreeCellResponse = {
  ...playingState,
  hint: { fromZone: 'freecell', fromCol: -1, cardIndex: -1, toZone: 'tableau', toCol: 3 },
};

const withHintFromColState: FreeCellResponse = {
  ...playingState,
  hint: { fromZone: 'tableau', fromCol: 2, cardIndex: 0, toZone: 'foundation', toCol: -1 },
};

const withFreeCellCardState: FreeCellResponse = {
  ...playingState,
  freeCells: [card('DIAMOND', 7), null, null, null],
};

beforeEach(() => {
  mockExec.mockResolvedValue(playingState);
  vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
});

describe('BakersGamePage', () => {
  // --- Skeleton ---

  it('renders skeleton when state is null', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<BakersGamePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  // --- Tableau ---

  it('renders tableau without index headers', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
  });

  it('renders empty tableau columns with a neutral any-card placeholder, not a King-only "K"', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    // Baker's Game empty columns accept ANY card, so the placeholder must not
    // suggest a King-only rule (the old "K" label borrowed from Klondike).
    const anyElements = screen.getAllByText('任意');
    expect(anyElements.length).toBeGreaterThanOrEqual(1);
    const placeholderButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === 'K');
    expect(placeholderButtons).toHaveLength(0);
  });

  it('labels empty tableau columns for screen readers', async () => {
    renderWithProviders(<BakersGamePage />);
    // playingState columns 2-7 are empty → each empty-column button is labelled
    // to convey that any card may be placed.
    await waitFor(() =>
      expect(screen.getByLabelText('空のタブロー列 2（任意のカードを置けます）')).toBeInTheDocument(),
    );
  });

  // --- Foundation ---

  it('renders foundation piles with suit symbols', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());
    expect(screen.getByText('♣')).toBeInTheDocument();
    expect(screen.getByText('♥')).toBeInTheDocument();
    expect(screen.getByText('♦')).toBeInTheDocument();
  });

  it('renders empty foundation with A placeholder', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());
    const aElements = screen.getAllByText('A');
    expect(aElements.length).toBeGreaterThanOrEqual(1);
  });

  it('renders foundation with cards', async () => {
    mockExec.mockResolvedValue(withFoundationState);
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());
    const imgs = screen.getAllByRole('img');
    expect(imgs.length).toBeGreaterThanOrEqual(1);
  });

  // --- Free cells ---

  it('renders free cells (empty)', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    const emptyButtons = screen.getAllByText('空');
    expect(emptyButtons.length).toBe(4);
  });

  it('shows the movable-count badge during play', async () => {
    // playingState: 4 empty free cells, 6 empty tableau columns → (1+4)*2^6 = 320.
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByTestId('bg-movable-count')).toHaveTextContent('320'));
  });

  it.each([
    ['game clear', gameClearState],
    ['game over', gameOverState],
  ])('hides the movable-count badge once the game ends (%s)', async (_label, endState) => {
    mockExec.mockResolvedValue(endState);
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    expect(screen.queryByTestId('bg-movable-count')).not.toBeInTheDocument();
  });

  it('renders freecell with card occupied', async () => {
    mockExec.mockResolvedValue(withFreeCellCardState);
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    // The occupied freecell should show a card image
    expect(screen.getByAltText('♦ 7')).toBeInTheDocument();
    // 3 empty freecells remain
    const emptyButtons = screen.getAllByText('空');
    expect(emptyButtons.length).toBe(3);
  });

  // --- Playing phase buttons ---

  it('renders playing phase buttons', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'オートコンプリート' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument();
  });

  it('undo button disabled when canUndo is false', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: false });
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeDisabled());
  });

  // --- Button interactions ---

  it('handleHint called on hint button click', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...playingState, hint: undefined });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('handleAutoComplete called on autocomplete button click', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'オートコンプリート' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: 'オートコンプリート' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('handleGiveUp called on giveup button click after confirm', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(gameOverState);
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('handleUndo called on undo button click', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: '元に戻す' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  // --- Card selection ---

  it('card selection via handleSelectSource on tableau card click', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    const cardImg = screen.getByAltText('♠ K');
    const cardButton = cardImg.closest('button') as HTMLButtonElement;
    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton.className).toContain('ring-2'));
  });

  it('card selection via handleSelectSource on freecell card click', async () => {
    mockExec.mockResolvedValue(withFreeCellCardState);
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByAltText('♦ 7')).toBeInTheDocument());

    const cardImg = screen.getByAltText('♦ 7');
    const cardButton = cardImg.closest('button') as HTMLButtonElement;
    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton.className).toContain('ring-2'));
  });

  it('target selection via handleSelectTarget on foundation click when source selected', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    // Select tableau card as source
    const cardImg = screen.getByAltText('♠ K');
    const cardButton = cardImg.closest('button') as HTMLButtonElement;
    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton.className).toContain('ring-2'));

    // Click empty foundation (A placeholder)
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    const aButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === 'A');
    if (aButtons.length > 0) {
      fireEvent.click(aButtons[0]);
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', expect.any(Object), expect.any(Object)));
    }
  });

  it('target selection via handleSelectTarget on empty freecell click when source selected', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    // Select tableau card as source
    const cardImg = screen.getByAltText('♠ K');
    const cardButton = cardImg.closest('button') as HTMLButtonElement;
    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton.className).toContain('ring-2'));

    // Click empty freecell
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    const emptyButtons = screen.getAllByText('空');
    if (emptyButtons.length > 0) {
      const emptyFcButton = emptyButtons[0].closest('button') as HTMLButtonElement;
      fireEvent.click(emptyFcButton);
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', expect.any(Object), expect.any(Object)));
    }
  });

  it('target selection via handleSelectTarget on empty tableau click when source selected', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    // Select tableau card as source
    const cardImg = screen.getByAltText('♠ K');
    const cardButton = cardImg.closest('button') as HTMLButtonElement;
    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton.className).toContain('ring-2'));

    // Click empty tableau column (any-card placeholder)
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    const kButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === '任意');
    if (kButtons.length > 0) {
      fireEvent.click(kButtons[0]);
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', expect.any(Object), expect.any(Object)));
    }
  });

  // --- End phases ---

  it('game clear phase shows action log section', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());
  });

  it('game over phase shows action log section', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());
  });

  it('playing buttons not shown when game is over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    expect(screen.queryByRole('button', { name: 'ヒント' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'オートコンプリート' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument();
  });

  // --- Hint display ---

  it('hint display when hint is set', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockResolvedValue(withHintState);
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(screen.getAllByText(/ヒント/).length).toBeGreaterThanOrEqual(1));
    // Zone identifiers are localized (ja), not shown as raw English.
    await waitFor(() => expect(screen.getByText(/フリーセル.*→.*タブロー 3/)).toBeInTheDocument());
  });

  it('hint display shows fromCol when fromCol is non-negative', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockResolvedValue(withHintFromColState);
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(screen.getByText(/タブロー 2/)).toBeInTheDocument());
  });

  // --- Keyboard shortcuts ---

  it('pressing h triggers hint in PLAYING phase', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(withHintState);
    fireEvent.keyDown(document, { key: 'h' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('pressing a triggers autocomplete in PLAYING phase', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'オートコンプリート' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.keyDown(document, { key: 'a' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('pressing g opens the give up confirm dialog in PLAYING phase', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(gameOverState);
    fireEvent.keyDown(document, { key: 'g' });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('pressing z triggers undo in PLAYING phase', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.keyDown(document, { key: 'z' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('keyboard shortcuts are disabled when game is over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('ゲームオーバー'));
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'h' });
    fireEvent.keyDown(document, { key: 'a' });
    fireEvent.keyDown(document, { key: 'g' });
    fireEvent.keyDown(document, { key: 'z' });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  // --- Reset confirmation dialog ---

  it('shows confirm dialog when reset button is clicked', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('dismisses confirm dialog on cancel', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('executes reset on confirm', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // --- Error display ---

  it('displays error message', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('Network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(screen.getByText('通信エラーが発生しました。もう一度お試しください。')).toBeInTheDocument(),
    );
  });

  it('displays hint error when hint fetch fails', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockRejectedValue(new Error('Network error'));
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() =>
      expect(screen.getByText('通信エラーが発生しました。もう一度お試しください。')).toBeInTheDocument(),
    );
  });

  // --- Move count display ---

  it('renders move count', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent(/手数: 5/));
  });

  // --- Action log ---

  it('action log button fetches and shows log', async () => {
    mockExec.mockResolvedValue(gameClearState);
    const mockLogApi = vi.mocked(actionLogApi.bakersgame);
    mockLogApi.mockResolvedValue({
      entries: [{ turnNumber: 1, playerIdx: 0, actionType: 'move', detail: 'tableau→foundation' }],
    });

    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '棋譜を見る' }));
    await waitFor(() => expect(screen.getByText('棋譜')).toBeInTheDocument());
  });

  it('reset hides the action log panel if open', async () => {
    mockExec.mockResolvedValue(gameClearState);
    vi.mocked(actionLogApi.bakersgame).mockResolvedValueOnce({
      entries: [{ turnNumber: 1, playerIdx: 0, actionType: 'move', detail: 'tableau→foundation' }],
    });

    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '棋譜を見る' }));
    await waitFor(() => expect(screen.getByText('棋譜')).toBeInTheDocument());

    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));

    await waitFor(() => expect(screen.queryByText('棋譜')).not.toBeInTheDocument());
  });

  // --- Foundation aria labels ---

  it('empty foundation buttons have aria-label', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());

    for (const suit of ['♠', '♣', '♥', '♦']) {
      expect(screen.getByRole('button', { name: `${suit} ファンデーション (空)` })).toBeInTheDocument();
    }
  });

  it('foundation with cards has aria-label with card count', async () => {
    mockExec.mockResolvedValue(withFoundationState);
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());

    expect(screen.getByRole('button', { name: '♠ ファンデーション (1枚)' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '♥ ファンデーション (2枚)' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '♣ ファンデーション (空)' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '♦ ファンデーション (空)' })).toBeInTheDocument();
  });

  // --- Freecell aria labels ---

  it('empty freecell buttons have aria-label', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    for (let i = 0; i < 4; i++) {
      expect(screen.getByRole('button', { name: `フリーセル ${i} (空)` })).toBeInTheDocument();
    }
  });

  // --- Tableau card aria ---

  it('tableau face-up card button has aria-label with card name', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    const cardButton = screen.getByRole('button', { name: '♠ K' });
    expect(cardButton).toHaveAttribute('aria-label', '♠ K');
  });

  it('tableau face-up card button has aria-pressed false initially and true when selected', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    const cardImg = screen.getByAltText('♠ K');
    const cardButton = cardImg.closest('button') as HTMLButtonElement;
    expect(cardButton).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton).toHaveAttribute('aria-pressed', 'true'));
  });

  // --- Empty targets disabled without source ---

  it('foundation disabled when no source selected', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());

    const aButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === 'A');
    for (const btn of aButtons) {
      expect(btn).toBeDisabled();
    }
  });

  it('empty tableau column disabled when no source selected', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    const kButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === '任意');
    for (const btn of kButtons) {
      expect(btn).toBeDisabled();
    }
  });

  it('empty freecell disabled when no source selected', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    const emptyButtons = screen.getAllByText('空');
    for (const btn of emptyButtons) {
      const button = btn.closest('button') as HTMLButtonElement;
      expect(button).toBeDisabled();
    }
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });

  it('starts tutorial when tutorial button is clicked', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
  });

  it('tutorial can be skipped', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('renders mobile viewport with flex-1 min-w-0 tableau columns', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    try {
      mockExec.mockResolvedValue(playingState);
      renderWithProviders(<BakersGamePage />);
      await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
      const tableau = document.querySelector('[data-tutorial="fc-tableau"]');
      const firstCol = tableau?.firstElementChild;
      expect(firstCol?.className).toContain('flex-1');
      expect(firstCol?.className).toContain('min-w-0');
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  it('renders desktop viewport with responsive tableau columns', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 800 });
    try {
      mockExec.mockResolvedValue(playingState);
      renderWithProviders(<BakersGamePage />);
      await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
      const tableau = document.querySelector('[data-tutorial="fc-tableau"]');
      const firstCol = tableau?.firstElementChild;
      expect(firstCol?.className).toContain('flex-1');
      expect(firstCol?.className).toContain('min-w-0');
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  // --- Frontend hint ---

  it('renders hint toggle checkbox in SettingsPanel', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByRole('checkbox', { name: 'ヒント表示' })).toBeInTheDocument());
  });

  it('renders HintTooltip when hintEnabled is true and hint is available', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'move', reason: 'frontendHint.useFreeCells', confidence: 'moderate' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  it('does not show stalemate escape button when not stalemate', async () => {
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('stalemate-escape-button')).not.toBeInTheDocument();
  });

  it('shows stalemate escape button when isStalemate is true', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: 5 });
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
    expect(screen.getByTestId('stalemate-escape-button')).toHaveTextContent('5');
  });

  it('clicking stalemate escape button dispatches undo_n', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: 3 });
    renderWithProviders(<BakersGamePage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByTestId('stalemate-escape-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo_n', undefined, undefined, 3));
  });

  describe('drag and drop', () => {
    function buildDataTransfer() {
      const store: Record<string, string> = {};
      return {
        setData: (type: string, val: string) => {
          store[type] = val;
        },
        getData: (type: string) => store[type] ?? '',
        effectAllowed: '',
        dropEffect: '',
      };
    }

    it('tableau face-up card is draggable when playing', async () => {
      renderWithProviders(<BakersGamePage />);
      await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
      const cardImg = screen.getByAltText('♠ K');
      const cardButton = cardImg.closest('button') as HTMLButtonElement;
      expect(cardButton).toHaveAttribute('draggable', 'true');
    });

    it('marks only the deepest movable cards as draggable when free cells + empty cols are exhausted', async () => {
      // Construct a stack of 3 cards in column 0 with all free cells full and no empty tableau cols.
      // Supermove limit becomes (1+0)*2^0 = 1, so only the bottom card (cardIndex=2) can move.
      const tightState: FreeCellResponse = {
        ...playingState,
        tableau: [
          [card('SPADE', 13), card('HEART', 12), card('CLOVER', 11)],
          [card('DIAMOND', 1)],
          [card('SPADE', 2)],
          [card('HEART', 3)],
          [card('DIAMOND', 4)],
          [card('CLOVER', 5)],
          [card('SPADE', 6)],
          [card('HEART', 7)],
        ],
        freeCells: [card('DIAMOND', 8), card('CLOVER', 9), card('SPADE', 10), card('HEART', 11)],
      };
      mockExec.mockResolvedValue(tightState);
      renderWithProviders(<BakersGamePage />);
      await waitFor(() => expect(screen.getByAltText('♠ K')).toBeInTheDocument());

      // Bottom card of the 3-card stack is movable; the two above are not.
      const topButton = screen.getByAltText('♠ K').closest('button') as HTMLButtonElement;
      const middleButton = screen.getByAltText('♥ Q').closest('button') as HTMLButtonElement;
      const bottomButton = screen.getByAltText('♣ J').closest('button') as HTMLButtonElement;
      expect(bottomButton).toHaveAttribute('draggable', 'true');
      expect(bottomButton).not.toHaveAttribute('data-supermove-blocked');
      expect(middleButton).toHaveAttribute('draggable', 'false');
      expect(middleButton).toHaveAttribute('data-supermove-blocked', 'true');
      expect(topButton).toHaveAttribute('draggable', 'false');
      expect(topButton).toHaveAttribute('data-supermove-blocked', 'true');
    });

    it('highlights the in-limit block under the cursor', async () => {
      const looseState: FreeCellResponse = {
        ...playingState,
        tableau: [
          [card('SPADE', 13), card('HEART', 12), card('CLOVER', 11)],
          [card('DIAMOND', 1)],
          [card('SPADE', 2)],
          [card('HEART', 3)],
          [card('DIAMOND', 4)],
          [card('CLOVER', 5)],
          [card('SPADE', 6)],
          [card('HEART', 7)],
        ],
        freeCells: [card('DIAMOND', 8), card('CLOVER', 9), null, null],
      };
      mockExec.mockResolvedValue(looseState);
      renderWithProviders(<BakersGamePage />);
      await waitFor(() => expect(screen.getByAltText('♥ Q')).toBeInTheDocument());

      const middleButton = screen.getByAltText('♥ Q').closest('button') as HTMLButtonElement;
      const bottomButton = screen.getByAltText('♣ J').closest('button') as HTMLButtonElement;
      fireEvent.mouseEnter(middleButton);
      expect(middleButton).toHaveAttribute('data-supermove-block', 'true');
      expect(bottomButton).toHaveAttribute('data-supermove-block', 'true');
      fireEvent.mouseLeave(middleButton);
      expect(middleButton).not.toHaveAttribute('data-supermove-block');
    });

    it('lets the entire stack drag when free cells + empty cols allow it', async () => {
      // 3-card stack, 2 empty free cells, 0 empty cols → limit = (1+2)*2^0 = 3 → all 3 movable.
      const looseState: FreeCellResponse = {
        ...playingState,
        tableau: [
          [card('SPADE', 13), card('HEART', 12), card('CLOVER', 11)],
          [card('DIAMOND', 1)],
          [card('SPADE', 2)],
          [card('HEART', 3)],
          [card('DIAMOND', 4)],
          [card('CLOVER', 5)],
          [card('SPADE', 6)],
          [card('HEART', 7)],
        ],
        freeCells: [card('DIAMOND', 8), card('CLOVER', 9), null, null],
      };
      mockExec.mockResolvedValue(looseState);
      renderWithProviders(<BakersGamePage />);
      await waitFor(() => expect(screen.getByAltText('♠ K')).toBeInTheDocument());

      const topButton = screen.getByAltText('♠ K').closest('button') as HTMLButtonElement;
      const middleButton = screen.getByAltText('♥ Q').closest('button') as HTMLButtonElement;
      const bottomButton = screen.getByAltText('♣ J').closest('button') as HTMLButtonElement;
      expect(topButton).toHaveAttribute('draggable', 'true');
      expect(middleButton).toHaveAttribute('draggable', 'true');
      expect(bottomButton).toHaveAttribute('draggable', 'true');
      expect(topButton).not.toHaveAttribute('data-supermove-blocked');
      expect(middleButton).not.toHaveAttribute('data-supermove-blocked');
      expect(bottomButton).not.toHaveAttribute('data-supermove-blocked');
    });

    it('free cell card is draggable when playing', async () => {
      mockExec.mockResolvedValue(withFreeCellCardState);
      renderWithProviders(<BakersGamePage />);
      await waitFor(() => expect(screen.getByAltText('♦ 7')).toBeInTheDocument());
      const cardButton = screen.getByAltText('♦ 7').closest('button') as HTMLButtonElement;
      expect(cardButton).toHaveAttribute('draggable', 'true');
    });

    it('dragging tableau card to empty tableau column dispatches move', async () => {
      renderWithProviders(<BakersGamePage />);
      await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

      const sourceCard = screen.getByAltText('♠ K');
      const sourceButton = sourceCard.closest('button') as HTMLButtonElement;
      const dataTransfer = buildDataTransfer();
      fireEvent.dragStart(sourceButton, { dataTransfer });

      // Find an empty tableau column (any-card placeholder)
      const kButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === '任意');
      expect(kButtons.length).toBeGreaterThan(0);
      const dropZone = kButtons[0].closest('div');
      mockExec.mockClear();
      mockExec.mockResolvedValue(playingState);
      fireEvent.dragOver(dropZone as HTMLElement, { dataTransfer });
      fireEvent.drop(dropZone as HTMLElement, { dataTransfer });

      await waitFor(() =>
        expect(mockExec).toHaveBeenCalledWith(
          'move',
          expect.objectContaining({ zone: 'tableau' }),
          expect.objectContaining({ zone: 'tableau' }),
        ),
      );
    });
  });

  // --- Double-click / double-tap auto-move ---

  describe('double-click auto-move', () => {
    it('double-clicking a tableau top card that can reach a foundation issues the foundation move', async () => {
      mockExec.mockResolvedValue({
        ...playingState,
        tableau: [[card('SPADE', 2)], [], [], [], [], [], [], []],
        foundation: [[card('SPADE', 1)], [], [], []],
      });
      renderWithProviders(<BakersGamePage />);
      const cardButton = (await screen.findByAltText('♠ 2')).closest('button') as HTMLButtonElement;

      mockExec.mockClear();
      mockExec.mockResolvedValue(playingState);
      fireEvent.doubleClick(cardButton);
      await waitFor(() =>
        expect(mockExec).toHaveBeenCalledWith('move', expect.objectContaining({ zone: 'tableau', col: 0 }), {
          zone: 'foundation',
          col: 0,
        }),
      );
    });

    it('double-clicking a free-cell card that can reach a foundation issues the foundation move', async () => {
      mockExec.mockResolvedValue({
        ...playingState,
        freeCells: [card('SPADE', 1), null, null, null],
        foundation: [[], [], [], []],
      });
      renderWithProviders(<BakersGamePage />);
      const cardButton = (await screen.findByAltText('♠ A')).closest('button') as HTMLButtonElement;

      mockExec.mockClear();
      mockExec.mockResolvedValue(playingState);
      fireEvent.doubleClick(cardButton);
      await waitFor(() =>
        expect(mockExec).toHaveBeenCalledWith('move', { zone: 'freecell', cell: 0 }, { zone: 'foundation', col: 0 }),
      );
    });

    it('double-clicking a card with no foundation move falls back to an empty free cell', async () => {
      mockExec.mockResolvedValue({
        ...playingState,
        tableau: [[card('SPADE', 13)], [], [], [], [], [], [], []],
        freeCells: [null, null, null, null],
        foundation: [[], [], [], []],
      });
      renderWithProviders(<BakersGamePage />);
      const cardButton = (await screen.findByAltText('♠ K')).closest('button') as HTMLButtonElement;

      mockExec.mockClear();
      mockExec.mockResolvedValue(playingState);
      fireEvent.doubleClick(cardButton);
      await waitFor(() =>
        expect(mockExec).toHaveBeenCalledWith('move', expect.objectContaining({ zone: 'tableau', col: 0 }), {
          zone: 'freecell',
          cell: 0,
        }),
      );
    });

    it('double-clicking a card with no legal auto-move issues no move', async () => {
      mockExec.mockResolvedValue({
        ...playingState,
        tableau: [[card('SPADE', 13)], [], [], [], [], [], [], []],
        freeCells: [card('HEART', 5), card('CLOVER', 6), card('DIAMOND', 7), card('SPADE', 8)],
        foundation: [[], [], [], []],
      });
      renderWithProviders(<BakersGamePage />);
      const cardButton = (await screen.findByAltText('♠ K')).closest('button') as HTMLButtonElement;

      mockExec.mockClear();
      fireEvent.doubleClick(cardButton);
      await Promise.resolve();
      expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());
    });

    it('single-clicking a card still selects it without issuing a move (no regression)', async () => {
      mockExec.mockResolvedValue({
        ...playingState,
        tableau: [[card('SPADE', 2)], [], [], [], [], [], [], []],
        foundation: [[card('SPADE', 1)], [], [], []],
      });
      renderWithProviders(<BakersGamePage />);
      const cardButton = (await screen.findByAltText('♠ 2')).closest('button') as HTMLButtonElement;

      mockExec.mockClear();
      fireEvent.click(cardButton);
      await Promise.resolve();
      expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());
      expect(cardButton).toHaveAttribute('aria-pressed', 'true');
    });
  });
});
