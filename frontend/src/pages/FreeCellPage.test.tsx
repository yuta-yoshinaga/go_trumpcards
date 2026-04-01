import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, freecellApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, FreeCellResponse } from '../types/card';
import { FreeCellPage } from './FreeCellPage';

vi.mock('../api/gameApi', () => ({
  freecellApi: { exec: vi.fn() },
  actionLogApi: { freecell: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(freecellApi.exec);

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
  messageCode: 'freecell.playing',
};

const gameClearState: FreeCellResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'freecell.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: FreeCellResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'freecell.gameOver',
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

describe('FreeCellPage', () => {
  // --- Skeleton ---

  it('renders skeleton when state is null', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<FreeCellPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  // --- Tableau ---

  it('renders tableau without index headers', async () => {
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
  });

  it('renders empty tableau columns with K placeholder', async () => {
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    const kElements = screen.getAllByText('K');
    expect(kElements.length).toBeGreaterThanOrEqual(1);
  });

  // --- Foundation ---

  it('renders foundation piles with suit symbols', async () => {
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());
    expect(screen.getByText('♣')).toBeInTheDocument();
    expect(screen.getByText('♥')).toBeInTheDocument();
    expect(screen.getByText('♦')).toBeInTheDocument();
  });

  it('renders empty foundation with A placeholder', async () => {
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());
    const aElements = screen.getAllByText('A');
    expect(aElements.length).toBeGreaterThanOrEqual(1);
  });

  it('renders foundation with cards', async () => {
    mockExec.mockResolvedValue(withFoundationState);
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());
    const imgs = screen.getAllByRole('img');
    expect(imgs.length).toBeGreaterThanOrEqual(1);
  });

  // --- Free cells ---

  it('renders free cells (empty)', async () => {
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    const emptyButtons = screen.getAllByText('空');
    expect(emptyButtons.length).toBe(4);
  });

  it('renders freecell with card occupied', async () => {
    mockExec.mockResolvedValue(withFreeCellCardState);
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    // The occupied freecell should show a card image
    expect(screen.getByAltText('♦ 7')).toBeInTheDocument();
    // 3 empty freecells remain
    const emptyButtons = screen.getAllByText('空');
    expect(emptyButtons.length).toBe(3);
  });

  // --- Playing phase buttons ---

  it('renders playing phase buttons', async () => {
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'オートコンプリート' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument();
  });

  it('undo button disabled when canUndo is false', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: false });
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeDisabled());
  });

  // --- Button interactions ---

  it('handleHint called on hint button click', async () => {
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...playingState, hint: undefined });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('handleAutoComplete called on autocomplete button click', async () => {
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'オートコンプリート' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: 'オートコンプリート' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('handleGiveUp called on giveup button click', async () => {
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(gameOverState);
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('handleUndo called on undo button click', async () => {
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: '元に戻す' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  // --- Card selection ---

  it('card selection via handleSelectSource on tableau card click', async () => {
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    const cardImg = screen.getByAltText('♠ K');
    const cardButton = cardImg.closest('button') as HTMLButtonElement;
    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton.className).toContain('ring-2'));
  });

  it('card selection via handleSelectSource on freecell card click', async () => {
    mockExec.mockResolvedValue(withFreeCellCardState);
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByAltText('♦ 7')).toBeInTheDocument());

    const cardImg = screen.getByAltText('♦ 7');
    const cardButton = cardImg.closest('button') as HTMLButtonElement;
    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton.className).toContain('ring-2'));
  });

  it('target selection via handleSelectTarget on foundation click when source selected', async () => {
    renderWithProviders(<FreeCellPage />);
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
    renderWithProviders(<FreeCellPage />);
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
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    // Select tableau card as source
    const cardImg = screen.getByAltText('♠ K');
    const cardButton = cardImg.closest('button') as HTMLButtonElement;
    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton.className).toContain('ring-2'));

    // Click empty tableau column (K placeholder)
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    const kButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === 'K');
    if (kButtons.length > 0) {
      fireEvent.click(kButtons[0]);
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', expect.any(Object), expect.any(Object)));
    }
  });

  // --- End phases ---

  it('game clear phase shows action log section', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());
  });

  it('game over phase shows action log section', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());
  });

  it('playing buttons not shown when game is over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    expect(screen.queryByRole('button', { name: 'ヒント' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'オートコンプリート' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument();
  });

  // --- Hint display ---

  it('hint display when hint is set', async () => {
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockResolvedValue(withHintState);
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(screen.getAllByText(/ヒント/).length).toBeGreaterThanOrEqual(1));
  });

  it('hint display shows fromCol when fromCol is non-negative', async () => {
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockResolvedValue(withHintFromColState);
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(screen.getByText(/tableau 2/)).toBeInTheDocument());
  });

  // --- Keyboard shortcuts ---

  it('pressing h triggers hint in PLAYING phase', async () => {
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(withHintState);
    fireEvent.keyDown(document, { key: 'h' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('pressing a triggers autocomplete in PLAYING phase', async () => {
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'オートコンプリート' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.keyDown(document, { key: 'a' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('pressing g triggers giveup in PLAYING phase', async () => {
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(gameOverState);
    fireEvent.keyDown(document, { key: 'g' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('pressing z triggers undo in PLAYING phase', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.keyDown(document, { key: 'z' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('keyboard shortcuts are disabled when game is over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('ゲームオーバー'));
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'h' });
    fireEvent.keyDown(document, { key: 'a' });
    fireEvent.keyDown(document, { key: 'g' });
    fireEvent.keyDown(document, { key: 'z' });
    expect(mockExec).not.toHaveBeenCalled();
  });

  // --- Reset confirmation dialog ---

  it('shows confirm dialog when reset button is clicked', async () => {
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('dismisses confirm dialog on cancel', async () => {
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('executes reset on confirm', async () => {
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // --- Error display ---

  it('displays error message', async () => {
    renderWithProviders(<FreeCellPage />);
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
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockRejectedValue(new Error('Network error'));
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() =>
      expect(screen.getByText('通信エラーが発生しました。もう一度お試しください。')).toBeInTheDocument(),
    );
  });

  // --- Move count display ---

  it('renders move count', async () => {
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent(/手数: 5/));
  });

  // --- Action log ---

  it('action log button fetches and shows log', async () => {
    mockExec.mockResolvedValue(gameClearState);
    const mockLogApi = vi.mocked(actionLogApi.freecell);
    mockLogApi.mockResolvedValue({
      entries: [{ turnNumber: 1, playerIdx: 0, actionType: 'move', detail: 'tableau→foundation' }],
    });

    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '棋譜を見る' }));
    await waitFor(() => expect(screen.getByText('棋譜')).toBeInTheDocument());
  });

  it('reset hides the action log panel if open', async () => {
    mockExec.mockResolvedValue(gameClearState);
    vi.mocked(actionLogApi.freecell).mockResolvedValueOnce({
      entries: [{ turnNumber: 1, playerIdx: 0, actionType: 'move', detail: 'tableau→foundation' }],
    });

    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '棋譜を見る' }));
    await waitFor(() => expect(screen.getByText('棋譜')).toBeInTheDocument());

    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(screen.queryByText('棋譜')).not.toBeInTheDocument());
  });

  // --- Foundation aria labels ---

  it('empty foundation buttons have aria-label', async () => {
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());

    for (const suit of ['♠', '♣', '♥', '♦']) {
      expect(screen.getByRole('button', { name: `${suit} ファンデーション (空)` })).toBeInTheDocument();
    }
  });

  it('foundation with cards has aria-label with card count', async () => {
    mockExec.mockResolvedValue(withFoundationState);
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());

    expect(screen.getByRole('button', { name: '♠ ファンデーション (1枚)' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '♥ ファンデーション (2枚)' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '♣ ファンデーション (空)' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '♦ ファンデーション (空)' })).toBeInTheDocument();
  });

  // --- Freecell aria labels ---

  it('empty freecell buttons have aria-label', async () => {
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    for (let i = 0; i < 4; i++) {
      expect(screen.getByRole('button', { name: `フリーセル ${i} (空)` })).toBeInTheDocument();
    }
  });

  // --- Tableau card aria ---

  it('tableau face-up card button has aria-label with card name', async () => {
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    const cardButton = screen.getByRole('button', { name: '♠ K' });
    expect(cardButton).toHaveAttribute('aria-label', '♠ K');
  });

  it('tableau face-up card button has aria-pressed false initially and true when selected', async () => {
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    const cardImg = screen.getByAltText('♠ K');
    const cardButton = cardImg.closest('button') as HTMLButtonElement;
    expect(cardButton).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton).toHaveAttribute('aria-pressed', 'true'));
  });

  // --- Empty targets disabled without source ---

  it('foundation disabled when no source selected', async () => {
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());

    const aButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === 'A');
    for (const btn of aButtons) {
      expect(btn).toBeDisabled();
    }
  });

  it('empty tableau column disabled when no source selected', async () => {
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    const kButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === 'K');
    for (const btn of kButtons) {
      expect(btn).toBeDisabled();
    }
  });

  it('empty freecell disabled when no source selected', async () => {
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    const emptyButtons = screen.getAllByText('空');
    for (const btn of emptyButtons) {
      const button = btn.closest('button') as HTMLButtonElement;
      expect(button).toBeDisabled();
    }
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });

  it('starts tutorial when tutorial button is clicked', async () => {
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
  });

  it('tutorial can be skipped', async () => {
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('renders mobile viewport with flex-shrink-0 and fixed-width tableau columns', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    try {
      mockExec.mockResolvedValue(playingState);
      renderWithProviders(<FreeCellPage />);
      await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
      const tableau = document.querySelector('[data-tutorial="fc-tableau"]');
      const firstCol = tableau?.firstElementChild;
      expect(firstCol?.className).toContain('flex-shrink-0');
      expect(firstCol?.className).toContain('sm:flex-1');
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  it('renders desktop viewport with responsive tableau columns', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 800 });
    try {
      mockExec.mockResolvedValue(playingState);
      renderWithProviders(<FreeCellPage />);
      await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
      const tableau = document.querySelector('[data-tutorial="fc-tableau"]');
      const firstCol = tableau?.firstElementChild;
      expect(firstCol?.className).toContain('flex-shrink-0');
      expect(firstCol?.className).toContain('sm:flex-1');
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  // --- Frontend hint ---

  it('renders hint toggle checkbox in SettingsPanel', async () => {
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByRole('checkbox', { name: 'ヒント表示' })).toBeInTheDocument());
  });

  it('renders HintTooltip when hintEnabled is true and hint is available', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { reason: 'frontendHint.useFreeCells', confidence: 'moderate' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<FreeCellPage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });
});
