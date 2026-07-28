import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, penguinApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, PenguinResponse } from '../types/card';
import { PenguinPage } from './PenguinPage';

vi.mock('../api/gameApi', () => ({
  penguinApi: { exec: vi.fn() },
  actionLogApi: { penguin: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(penguinApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: PenguinResponse = {
  tableau: [[card('SPADE', 13)], [card('HEART', 12)], [], [], [], [], []],
  freeCells: [card('DIAMOND', 3), card('CLOVER', 5), card('HEART', 7), null, null, null, null],
  foundation: [[], [], [], []],
  baseRank: 4,
  phase: 0,
  moveCount: 5,
  canUndo: true,
  isStalemate: false,
  message: '',
  messageCode: 'penguin.playing',
};

// All free cells occupied and every column non-empty → supermoveLimit = (1+0)*2^0 = 1,
// so the 2-card stack in column 0 exceeds the limit and shows the tooltip.
const supermoveExceedState: PenguinResponse = {
  ...playingState,
  freeCells: [
    card('DIAMOND', 3),
    card('CLOVER', 5),
    card('HEART', 7),
    card('SPADE', 9),
    card('DIAMOND', 10),
    card('CLOVER', 11),
    card('HEART', 2),
  ],
  tableau: [
    [card('SPADE', 13), card('SPADE', 12)],
    [card('HEART', 6)],
    [card('CLOVER', 8)],
    [card('DIAMOND', 4)],
    [card('SPADE', 6)],
    [card('HEART', 9)],
    [card('CLOVER', 2)],
  ],
};

const gameClearState: PenguinResponse = {
  ...playingState,
  phase: 1,
  message: 'Game Clear!',
  messageCode: 'penguin.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: PenguinResponse = {
  ...playingState,
  phase: 2,
  message: 'Game Over',
  messageCode: 'penguin.gameOver',
};

const withFoundationState: PenguinResponse = {
  ...playingState,
  foundation: [[card('SPADE', 4)], [], [card('HEART', 4), card('HEART', 5)], []],
};

const withHintState: PenguinResponse = {
  ...playingState,
  hint: { fromZone: 'freecell', fromCol: -1, cardIndex: -1, toZone: 'tableau', toCol: 3 },
};

const withHintFromColState: PenguinResponse = {
  ...playingState,
  hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: -1 },
};

beforeEach(() => {
  mockExec.mockResolvedValue(playingState);
  vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
});

describe('PenguinPage', () => {
  // --- Skeleton ---

  it('renders skeleton when state is null', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<PenguinPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  // --- Tableau ---

  it('renders tableau without index headers', async () => {
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
  });

  it('renders empty tableau columns with prevRank placeholder (3 for baseRank=4)', async () => {
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    const threeElements = screen.getAllByText('3');
    expect(threeElements.length).toBeGreaterThanOrEqual(1);
  });

  it('gives empty tableau columns a descriptive aria-label whose rank follows baseRank', async () => {
    renderWithProviders(<PenguinPage />);
    // baseRank 4 → only the rank one below (3) may be placed on an empty column.
    await waitFor(() =>
      expect(screen.getAllByRole('button', { name: '空き列（3 のみ置けます）' }).length).toBeGreaterThanOrEqual(1),
    );

    // Changing the base rank moves the placeable rank (baseRank 1 → K).
    mockExec.mockResolvedValue({ ...playingState, baseRank: 1 });
    renderWithProviders(<PenguinPage />);
    await waitFor(() =>
      expect(screen.getAllByRole('button', { name: '空き列（K のみ置けます）' }).length).toBeGreaterThanOrEqual(1),
    );
  });

  // --- Supermove limit ---

  it('supermove-limit tooltip includes the limit, free-cell count, and empty-column count', async () => {
    mockExec.mockResolvedValue(supermoveExceedState);
    renderWithProviders(<PenguinPage />);
    // Top card of column 0 exceeds the limit (stack of 2 > limit of 1).
    const topCard = await screen.findByTestId('pg-tableau-0-0');
    expect(topCard).toHaveAttribute('data-supermove-blocked', 'true');
    // limit=1, cells=0, cols=0
    expect(topCard).toHaveAttribute('title', '一度に動かせるのは1枚まで（空きセル0・空き列0）');
  });

  it('shows a supermove-limit badge reflecting the free-cell/column counts', async () => {
    // playingState: 4 empty free cells, 5 empty columns → (1+4)*2^5 = 160.
    renderWithProviders(<PenguinPage />);
    const badge = await screen.findByTestId('pg-supermove-badge');
    expect(badge).toHaveTextContent('最大移動: 160枚');
  });

  it('supermove-limit badge updates when free space shrinks', async () => {
    mockExec.mockResolvedValue(supermoveExceedState);
    renderWithProviders(<PenguinPage />);
    const badge = await screen.findByTestId('pg-supermove-badge');
    // 0 empty free cells, 0 empty columns → (1+0)*2^0 = 1.
    expect(badge).toHaveTextContent('最大移動: 1枚');
  });

  // --- Foundation ---

  it('renders foundation piles with suit symbols', async () => {
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());
    expect(screen.getByText('♣')).toBeInTheDocument();
    expect(screen.getByText('♥')).toBeInTheDocument();
    expect(screen.getByText('♦')).toBeInTheDocument();
  });

  it('renders empty foundation with baseRank placeholder (4)', async () => {
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());
    const fourElements = screen.getAllByText('4');
    expect(fourElements.length).toBeGreaterThanOrEqual(1);
  });

  it('shows a base-rank legend near the foundation with the wraparound rule', async () => {
    renderWithProviders(<PenguinPage />);
    const legend = await screen.findByTestId('pg-base-rank-legend');
    // baseRank 4 → starts at 4, ends at 3 (rank below).
    expect(legend).toHaveTextContent('4');
    expect(legend).toHaveTextContent('3');
    expect(legend).toHaveTextContent(/折り返し/);
  });

  it('renders foundation with cards', async () => {
    mockExec.mockResolvedValue(withFoundationState);
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());
    const imgs = screen.getAllByRole('img');
    expect(imgs.length).toBeGreaterThanOrEqual(1);
  });

  // --- Free cells ---

  it('renders free cells (3 occupied, 4 empty)', async () => {
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    const emptyButtons = screen.getAllByText('空');
    expect(emptyButtons.length).toBe(4);
  });

  // --- Playing phase buttons ---

  it('renders playing phase buttons', async () => {
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'オートコンプリート' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument();
  });

  it('undo button disabled when canUndo is false', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: false });
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeDisabled());
  });

  // --- Button interactions ---

  it('handleHint called on hint button click', async () => {
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...playingState, hint: undefined });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('handleAutoComplete called on autocomplete button click', async () => {
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'オートコンプリート' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: 'オートコンプリート' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('handleGiveUp called on giveup button click via confirm dialog', async () => {
    renderWithProviders(<PenguinPage />);
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
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: '元に戻す' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  // --- Card selection ---

  it('card selection via handleSelectSource on tableau card click', async () => {
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    const cardImg = screen.getByAltText('♠ K');
    const cardButton = cardImg.closest('button') as HTMLButtonElement;
    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton.className).toContain('ring-2'));
  });

  it('card selection via handleSelectSource on freecell card click', async () => {
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByAltText('♦ 3')).toBeInTheDocument());

    const cardImg = screen.getByAltText('♦ 3');
    const cardButton = cardImg.closest('button') as HTMLButtonElement;
    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton.className).toContain('ring-2'));
  });

  it('target selection via handleSelectTarget on foundation click when source selected', async () => {
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    // Select tableau card as source
    const cardImg = screen.getByAltText('♠ K');
    const cardButton = cardImg.closest('button') as HTMLButtonElement;
    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton.className).toContain('ring-2'));

    // Click empty foundation (baseRank placeholder = 4)
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    const fourButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === '4');
    if (fourButtons.length > 0) {
      fireEvent.click(fourButtons[0]);
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', expect.any(Object), expect.any(Object)));
    }
  });

  // --- End phases ---

  it('game clear phase shows action log section', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());
  });

  it('game over phase shows action log section', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());
  });

  it('playing buttons not shown when game is over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    expect(screen.queryByRole('button', { name: 'ヒント' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'オートコンプリート' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument();
  });

  // --- Hint display ---

  it('highlights the hint target column (success ring) when a hint is set', async () => {
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockResolvedValue(withHintState); // toZone tableau, toCol 3
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(screen.getByTestId('pg-col-3').className).toContain('ring-ds-success'));
    expect(screen.getByTestId('pg-col-0').className).not.toContain('ring-ds-success');
  });

  it('highlights the hint source card with an info ring when fromZone is tableau', async () => {
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockResolvedValue(withHintFromColState); // fromZone tableau, fromCol 0, cardIndex 0
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(screen.getByTestId('pg-tableau-0-0').className).toContain('ring-ds-info'));
  });

  it('highlights a free-cell hint source with an info ring', async () => {
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockResolvedValue({
      ...playingState, // freeCells[0] has DIAMOND 3
      hint: { fromZone: 'freecell', fromCol: 0, cardIndex: -1, toZone: 'tableau', toCol: -1 },
    });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(screen.getByTestId('pg-freecell-0').className).toContain('ring-ds-info'));
  });

  it('highlights an empty free-cell hint target with a success ring', async () => {
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockResolvedValue({
      ...playingState, // freeCells[3] is null
      hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 0, toZone: 'freecell', toCol: 3 },
    });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(screen.getByTestId('pg-freecell-empty-3').className).toContain('ring-ds-success'));
  });

  it('highlights an empty foundation hint target with a success ring', async () => {
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockResolvedValue({
      ...playingState,
      hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: 0 },
    });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(screen.getByTestId('pg-foundation-empty-0').className).toContain('ring-ds-success'));
  });

  it('highlights a filled foundation hint target with a success ring', async () => {
    mockExec.mockResolvedValue({
      ...withFoundationState, // foundation[0] has SPADE 4
      hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: 0 },
    });
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(screen.getByTestId('pg-foundation-0').className).toContain('ring-ds-success'));
  });

  // --- Keyboard shortcuts ---

  it('pressing h triggers hint in PLAYING phase', async () => {
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(withHintState);
    fireEvent.keyDown(document, { key: 'h' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('pressing a triggers autocomplete in PLAYING phase', async () => {
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'オートコンプリート' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.keyDown(document, { key: 'a' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('pressing g opens the give up confirm dialog in PLAYING phase', async () => {
    renderWithProviders(<PenguinPage />);
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

  it('keyboard shortcuts are disabled when game is over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<PenguinPage />);
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
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('executes reset on confirm', async () => {
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // --- Error display ---

  it('displays error message', async () => {
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('Network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(screen.getByText('通信エラーが発生しました。もう一度お試しください。')).toBeInTheDocument(),
    );
  });

  // --- Move count display ---

  it('renders move count and base rank', async () => {
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent(/手数: 5/));
    expect(screen.getByTestId('phase-indicator')).toHaveTextContent(/基準ランク: 4/);
  });

  // --- Action log ---

  it('action log button fetches and shows log', async () => {
    mockExec.mockResolvedValue(gameClearState);
    const mockLogApi = vi.mocked(actionLogApi.penguin);
    mockLogApi.mockResolvedValue({
      entries: [{ turnNumber: 1, playerIdx: 0, actionType: 'move', detail: 'tableau→foundation' }],
    });

    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '棋譜を見る' }));
    await waitFor(() => expect(screen.getByText('棋譜')).toBeInTheDocument());
  });

  // --- Frontend hint ---

  it('renders hint toggle checkbox in SettingsPanel', async () => {
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByRole('checkbox', { name: 'ヒント表示' })).toBeInTheDocument());
  });

  it('renders HintTooltip when hintEnabled is true and hint is available', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'move', reason: 'frontendHint.useFreeCells', confidence: 'moderate' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  it('does not show stalemate escape button when not stalemate', async () => {
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('stalemate-escape-button')).not.toBeInTheDocument();
  });

  it('shows stalemate escape button when isStalemate is true', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: 5 });
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
    expect(screen.getByTestId('stalemate-escape-button')).toHaveTextContent('5');
  });

  it('clicking stalemate escape button dispatches undo_n', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: 3 });
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByTestId('stalemate-escape-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo_n', undefined, undefined, 3));
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<PenguinPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });
});
