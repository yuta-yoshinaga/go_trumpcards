import { cleanup, fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, eightoffApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, EightOffResponse } from '../types/card';
import { EightOffPage } from './EightOffPage';

vi.mock('../api/gameApi', () => ({
  eightoffApi: { exec: vi.fn() },
  actionLogApi: { eightoff: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(eightoffApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: EightOffResponse = {
  tableau: [[card('SPADE', 13)], [card('HEART', 12)], [], [], [], [], [], []],
  freeCells: [null, null, null, null, null, null, null, null],
  foundation: [[], [], [], []],
  phase: 0,
  moveCount: 5,
  canUndo: true,
  isStalemate: false,
  message: '',
  messageCode: 'eightoff.playing',
};

const gameClearState: EightOffResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'eightoff.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: EightOffResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'eightoff.gameOver',
};

const withFoundationState: EightOffResponse = {
  ...playingState,
  foundation: [[card('SPADE', 1)], [], [card('HEART', 1), card('HEART', 2)], []],
};

const withHintState: EightOffResponse = {
  ...playingState,
  hint: { fromZone: 'freecell', fromCol: -1, cardIndex: -1, toZone: 'tableau', toCol: 3 },
};

const withHintFromColState: EightOffResponse = {
  ...playingState,
  hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: -1 },
};

const withFreeCellCardState: EightOffResponse = {
  ...playingState,
  freeCells: [card('DIAMOND', 7), null, null, null, null, null, null, null],
};

// All free cells occupied and every tableau column non-empty → supermoveLimit = 1,
// so the deeper card of a 2-card column exceeds the limit and gets the blocked tooltip.
const supermoveBlockedState: EightOffResponse = {
  ...playingState,
  tableau: [
    [card('SPADE', 13), card('HEART', 12)],
    [card('CLOVER', 13)],
    [card('DIAMOND', 13)],
    [card('SPADE', 11)],
    [card('HEART', 11)],
    [card('CLOVER', 11)],
    [card('DIAMOND', 11)],
    [card('SPADE', 10)],
  ],
  freeCells: [
    card('SPADE', 2),
    card('SPADE', 3),
    card('SPADE', 4),
    card('SPADE', 5),
    card('HEART', 6),
    card('HEART', 7),
    card('HEART', 8),
    card('HEART', 9),
  ],
};

beforeEach(() => {
  mockExec.mockResolvedValue(playingState);
  vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
});

describe('EightOffPage', () => {
  // --- Skeleton ---

  it('renders skeleton when state is null', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<EightOffPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  // --- Tableau ---

  it('renders tableau without index headers', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
  });

  it('shows supermove limit tooltip with empty free-cell and column counts', async () => {
    mockExec.mockResolvedValue(supermoveBlockedState);
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    const blocked = screen.getByTestId('eo-tableau-0-0');
    expect(blocked).toHaveAttribute('data-supermove-blocked', 'true');
    const title = blocked.getAttribute('title') ?? '';
    expect(title).toMatch(/空きフリーセル0/);
    expect(title).toMatch(/空き列0/);

    // **同じ内容が読み上げにも出る。** title はホバー時にしか読まれない (#5820)。
    const label = blocked.getAttribute('aria-label') ?? '';
    expect(label).toContain('一度に動かせるのは');
    expect(label).toMatch(/空きフリーセル0/);
    // 負のコントロール: 上限内の札には付かない。
    const movable = screen.getByTestId('eo-tableau-0-1');
    expect(movable).not.toHaveAttribute('data-supermove-blocked');
    expect(movable.getAttribute('aria-label') ?? '').not.toContain('一度に動かせるのは');
  });

  it('renders empty tableau columns with K placeholder', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    const kElements = screen.getAllByText('K');
    expect(kElements.length).toBeGreaterThanOrEqual(1);
  });

  // --- Foundation ---

  it('renders foundation piles with suit symbols', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());
    expect(screen.getByText('♣')).toBeInTheDocument();
    expect(screen.getByText('♥')).toBeInTheDocument();
    expect(screen.getByText('♦')).toBeInTheDocument();
  });

  it('renders empty foundation with A placeholder', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());
    const aElements = screen.getAllByText('A');
    expect(aElements.length).toBeGreaterThanOrEqual(1);
  });

  it('renders foundation with cards', async () => {
    mockExec.mockResolvedValue(withFoundationState);
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());
    const imgs = screen.getAllByRole('img');
    expect(imgs.length).toBeGreaterThanOrEqual(1);
  });

  // --- Free cells ---

  it('renders free cells (empty)', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    const emptyButtons = screen.getAllByText('空');
    expect(emptyButtons.length).toBe(8);
  });

  it('renders freecell with card occupied', async () => {
    mockExec.mockResolvedValue(withFreeCellCardState);
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    // The occupied freecell should show a card image
    expect(screen.getByAltText('♦ 7')).toBeInTheDocument();
    // 7 empty freecells remain
    const emptyButtons = screen.getAllByText('空');
    expect(emptyButtons.length).toBe(7);
  });

  // --- Playing phase buttons ---

  it('renders playing phase buttons', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'オートコンプリート' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument();
  });

  it('undo button disabled when canUndo is false', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: false });
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeDisabled());
  });

  // --- Button interactions ---

  it('handleHint called on hint button click', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue({ ...playingState, hint: undefined });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('announces the hinted move (tableau source) in a polite live region', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    // playingState.tableau[0][0] is ♠ K; hint moves it to a foundation.
    mockExec.mockResolvedValue(withHintFromColState);
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    const region = await screen.findByTestId('eo-hint-announce');
    expect(region).toHaveAttribute('role', 'status');
    expect(region).toHaveAttribute('aria-live', 'polite');
    await waitFor(() => expect(region).toHaveTextContent('ヒント: ♠ K を タブロー 1 から ファンデーション へ移動'));
  });

  it('announces a free-cell source hint by resolving the free-cell card', async () => {
    // Mount with a free cell occupied so state.freeCells[0] stays ♦ 7 (handleHint
    // updates only the hint, not state).
    mockExec.mockResolvedValue(withFreeCellCardState);
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockResolvedValue({
      ...withFreeCellCardState,
      hint: { fromZone: 'freecell', fromCol: 0, cardIndex: -1, toZone: 'tableau', toCol: 3 },
    });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    const region = await screen.findByTestId('eo-hint-announce');
    await waitFor(() => expect(region).toHaveTextContent('ヒント: ♦ 7 を フリーセル 1 から タブロー 4 へ移動'));
  });

  it('announces that no moves are available when the hint request yields none', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockResolvedValue({ ...playingState, hint: undefined });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    const region = await screen.findByTestId('eo-hint-announce');
    await waitFor(() => expect(region).toHaveTextContent('移動可能な手がありません'));
  });

  it('keeps the hint live region empty before any hint is requested', async () => {
    renderWithProviders(<EightOffPage />);
    const region = await screen.findByTestId('eo-hint-announce');
    expect(region).toHaveTextContent('');
  });

  it('does not announce "no moves" when the hint request fails', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    // hint fetch rejects → hintError is set, hint stays null; the region must stay
    // empty rather than falsely announcing "移動可能な手がありません".
    mockExec.mockRejectedValueOnce(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
    const region = screen.getByTestId('eo-hint-announce');
    expect(region).toHaveTextContent('');
  });

  it('announces an empty card name when a free-cell hint points at an empty cell', async () => {
    renderWithProviders(<EightOffPage />); // playingState: all free cells null
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockResolvedValue({
      ...playingState,
      hint: { fromZone: 'freecell', fromCol: 0, cardIndex: -1, toZone: 'tableau', toCol: 3 },
    });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    const region = await screen.findByTestId('eo-hint-announce');
    await waitFor(() => expect(region).toHaveTextContent('ヒント: を フリーセル 1 から タブロー 4 へ移動'));
  });

  it('announces an empty card name when a tableau hint points at a missing card', async () => {
    renderWithProviders(<EightOffPage />); // playingState.tableau[2] is empty
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockResolvedValue({
      ...playingState,
      hint: { fromZone: 'tableau', fromCol: 2, cardIndex: 0, toZone: 'foundation', toCol: -1 },
    });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    const region = await screen.findByTestId('eo-hint-announce');
    await waitFor(() => expect(region).toHaveTextContent('ヒント: を タブロー 3 から ファンデーション へ移動'));
  });

  it('handleAutoComplete called on autocomplete button click', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'オートコンプリート' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: 'オートコンプリート' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('giveup button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(gameOverState);
    // Clicking give-up must NOT dispatch immediately — it opens a confirm dialog (#2099).
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();
    // Confirming dispatches giveup.
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('handleUndo called on undo button click', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: '元に戻す' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  // --- Card selection ---

  it('card selection via handleSelectSource on tableau card click', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    const cardImg = screen.getByAltText('♠ K');
    const cardButton = cardImg.closest('button') as HTMLButtonElement;
    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton.className).toContain('ring-2'));
  });

  it('card selection via handleSelectSource on freecell card click', async () => {
    mockExec.mockResolvedValue(withFreeCellCardState);
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByAltText('♦ 7')).toBeInTheDocument());

    const cardImg = screen.getByAltText('♦ 7');
    const cardButton = cardImg.closest('button') as HTMLButtonElement;
    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton.className).toContain('ring-2'));
  });

  it('target selection via handleSelectTarget on foundation click when source selected', async () => {
    renderWithProviders(<EightOffPage />);
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
    renderWithProviders(<EightOffPage />);
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
    renderWithProviders(<EightOffPage />);
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
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());
  });

  it('game over phase shows action log section', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());
  });

  it('playing buttons not shown when game is over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    expect(screen.queryByRole('button', { name: 'ヒント' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'オートコンプリート' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument();
  });

  // --- Hint display ---

  it('highlights the hint target column (success ring) when a hint is set', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockResolvedValue(withHintState); // toZone tableau, toCol 3
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(screen.getByTestId('eo-col-3').className).toContain('ring-ds-success'));
    expect(screen.getByTestId('eo-col-0').className).not.toContain('ring-ds-success');
  });

  it('highlights the hint source card with an info ring when fromZone is tableau', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockResolvedValue(withHintFromColState); // fromZone tableau, fromCol 0, cardIndex 0
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(screen.getByTestId('eo-tableau-0-0').className).toContain('ring-ds-info'));
  });

  it('highlights a free-cell hint source with an info ring', async () => {
    mockExec.mockResolvedValue({
      ...withFreeCellCardState,
      hint: { fromZone: 'freecell', fromCol: 0, cardIndex: -1, toZone: 'tableau', toCol: -1 },
    });
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(screen.getByTestId('eo-freecell-0').className).toContain('ring-ds-info'));
  });

  it('highlights an empty free-cell hint target with a success ring', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockResolvedValue({
      ...playingState,
      hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 0, toZone: 'freecell', toCol: 1 },
    });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(screen.getByTestId('eo-freecell-empty-1').className).toContain('ring-ds-success'));
  });

  it('highlights an empty foundation hint target with a success ring', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockResolvedValue({
      ...playingState,
      hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: 0 },
    });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(screen.getByTestId('eo-foundation-empty-0').className).toContain('ring-ds-success'));
  });

  it('highlights a filled foundation hint target with a success ring', async () => {
    mockExec.mockResolvedValue({
      ...withFoundationState, // foundation[0] has SPADE A
      hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 0, toZone: 'foundation', toCol: 0 },
    });
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(screen.getByTestId('eo-foundation-0').className).toContain('ring-ds-success'));
  });

  // --- Keyboard shortcuts ---

  it('pressing h triggers hint in PLAYING phase', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(withHintState);
    fireEvent.keyDown(document, { key: 'h' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('pressing a triggers autocomplete in PLAYING phase', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'オートコンプリート' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.keyDown(document, { key: 'a' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('pressing g opens the give up confirm dialog in PLAYING phase', async () => {
    renderWithProviders(<EightOffPage />);
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
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.keyDown(document, { key: 'z' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('keyboard shortcuts are disabled when game is over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<EightOffPage />);
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
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('dismisses confirm dialog on cancel', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('executes reset on confirm', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // --- Error display ---

  it('displays error message', async () => {
    renderWithProviders(<EightOffPage />);
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
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());

    mockExec.mockRejectedValue(new Error('Network error'));
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() =>
      expect(screen.getByText('通信エラーが発生しました。もう一度お試しください。')).toBeInTheDocument(),
    );
  });

  // --- Move count display ---

  it('renders move count', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent(/手数: 5/));
  });

  // --- Action log ---

  it('action log button fetches and shows log', async () => {
    mockExec.mockResolvedValue(gameClearState);
    const mockLogApi = vi.mocked(actionLogApi.eightoff);
    mockLogApi.mockResolvedValue({
      entries: [{ turnNumber: 1, playerIdx: 0, actionType: 'move', detail: 'tableau→foundation' }],
    });

    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '棋譜を見る' }));
    await waitFor(() => expect(screen.getByText('棋譜')).toBeInTheDocument());
  });

  it('reset hides the action log panel if open', async () => {
    mockExec.mockResolvedValue(gameClearState);
    vi.mocked(actionLogApi.eightoff).mockResolvedValueOnce({
      entries: [{ turnNumber: 1, playerIdx: 0, actionType: 'move', detail: 'tableau→foundation' }],
    });

    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '棋譜を見る' }));
    await waitFor(() => expect(screen.getByText('棋譜')).toBeInTheDocument());

    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));

    await waitFor(() => expect(screen.queryByText('棋譜')).not.toBeInTheDocument());
  });

  // --- Foundation aria labels ---

  it('empty foundation buttons have aria-label', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());

    for (const suit of ['♠', '♣', '♥', '♦']) {
      expect(screen.getByRole('button', { name: `${suit} ファンデーション (空)` })).toBeInTheDocument();
    }
  });

  it('foundation with cards has aria-label with card count', async () => {
    mockExec.mockResolvedValue(withFoundationState);
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());

    expect(screen.getByRole('button', { name: '♠ ファンデーション (1枚)' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '♥ ファンデーション (2枚)' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '♣ ファンデーション (空)' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '♦ ファンデーション (空)' })).toBeInTheDocument();
  });

  // --- Freecell aria labels ---

  it('empty freecell buttons have aria-label', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    for (let i = 0; i < 8; i++) {
      expect(screen.getByRole('button', { name: `フリーセル ${i} (空)` })).toBeInTheDocument();
    }
  });

  // --- Tableau card aria ---

  it('tableau face-up card button has aria-label with card name', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    const cardButton = screen.getByRole('button', { name: '♠ K' });
    expect(cardButton).toHaveAttribute('aria-label', '♠ K');
  });

  it('tableau face-up card button has aria-pressed false initially and true when selected', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    const cardImg = screen.getByAltText('♠ K');
    const cardButton = cardImg.closest('button') as HTMLButtonElement;
    expect(cardButton).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(cardButton);
    await waitFor(() => expect(cardButton).toHaveAttribute('aria-pressed', 'true'));
  });

  // --- Empty targets disabled without source ---

  it('foundation disabled when no source selected', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());

    const aButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === 'A');
    for (const btn of aButtons) {
      expect(btn).toBeDisabled();
    }
  });

  it('empty tableau column disabled when no source selected', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    const kButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === 'K');
    for (const btn of kButtons) {
      expect(btn).toBeDisabled();
    }
  });

  it('empty freecell disabled when no source selected', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    const emptyButtons = screen.getAllByText('空');
    for (const btn of emptyButtons) {
      const button = btn.closest('button') as HTMLButtonElement;
      expect(button).toBeDisabled();
    }
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });

  it('starts tutorial when tutorial button is clicked', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
  });

  it('tutorial can be skipped', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('renders mobile viewport with flex-1 min-w-0 tableau columns', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    try {
      mockExec.mockResolvedValue(playingState);
      renderWithProviders(<EightOffPage />);
      await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
      const tableau = document.querySelector('[data-tutorial="eo-tableau"]');
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
      renderWithProviders(<EightOffPage />);
      await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
      const tableau = document.querySelector('[data-tutorial="eo-tableau"]');
      const firstCol = tableau?.firstElementChild;
      expect(firstCol?.className).toContain('flex-1');
      expect(firstCol?.className).toContain('min-w-0');
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  // --- Frontend hint ---

  it('renders hint toggle checkbox in SettingsPanel', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByRole('checkbox', { name: 'ヒント表示' })).toBeInTheDocument());
  });

  it('renders HintTooltip when hintEnabled is true and hint is available', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'move', reason: 'frontendHint.useFreeCells', confidence: 'moderate' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  it('does not show stalemate escape button when not stalemate', async () => {
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('stalemate-escape-button')).not.toBeInTheDocument();
  });

  it('shows stalemate escape button when isStalemate is true', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: 5 });
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
    expect(screen.getByTestId('stalemate-escape-button')).toHaveTextContent('5');
  });

  it('clicking stalemate escape button dispatches undo_n', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: 3 });
    renderWithProviders(<EightOffPage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByTestId('stalemate-escape-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo_n', undefined, undefined, 3));
  });

  // --- Double-click foundation shortcut ---

  describe('double-click foundation shortcut', () => {
    const tableauAceTop: EightOffResponse = {
      ...playingState,
      tableau: [[card('SPADE', 1)], [card('HEART', 12)], [], [], [], [], [], []],
      foundation: [[], [], [], []],
    };

    const freeCellPlayable: EightOffResponse = {
      ...playingState,
      freeCells: [card('DIAMOND', 7), null, null, null, null, null, null, null],
      foundation: [[], [], [], [card('DIAMOND', 6)]],
    };

    it('double-clicking a foundation-playable tableau top card auto-sends it to the foundation', async () => {
      mockExec.mockResolvedValue(tableauAceTop);
      renderWithProviders(<EightOffPage />);
      await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

      mockExec.mockClear();
      mockExec.mockResolvedValue(playingState);
      const cardButton = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
      fireEvent.doubleClick(cardButton);

      await waitFor(() =>
        expect(mockExec).toHaveBeenCalledWith(
          'move',
          expect.objectContaining({ zone: 'tableau', col: 0, cardIndex: 0 }),
          { zone: 'foundation', col: 0 },
        ),
      );
    });

    it('double-clicking a foundation-playable free-cell card auto-sends it to the foundation', async () => {
      mockExec.mockResolvedValue(freeCellPlayable);
      renderWithProviders(<EightOffPage />);
      await waitFor(() => expect(screen.getByAltText('♦ 7')).toBeInTheDocument());

      mockExec.mockClear();
      mockExec.mockResolvedValue(playingState);
      const cardButton = screen.getByAltText('♦ 7').closest('button') as HTMLButtonElement;
      fireEvent.doubleClick(cardButton);

      await waitFor(() =>
        expect(mockExec).toHaveBeenCalledWith('move', expect.objectContaining({ zone: 'freecell', cell: 0 }), {
          zone: 'foundation',
          col: 3,
        }),
      );
    });

    it('double-clicking a non-foundation-playable tableau top card does nothing', async () => {
      // playingState: tableau[0] top is ♠ K with an empty foundation → no legal target.
      renderWithProviders(<EightOffPage />);
      await waitFor(() => expect(screen.getByAltText('♠ K')).toBeInTheDocument());

      mockExec.mockClear();
      const cardButton = screen.getByAltText('♠ K').closest('button') as HTMLButtonElement;
      fireEvent.doubleClick(cardButton);

      await flushPendingDispatch();
      expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());
    });

    it('ignores the trailing single-click of a double-click (detail >= 2) so no stray selection occurs', async () => {
      mockExec.mockResolvedValue(tableauAceTop);
      renderWithProviders(<EightOffPage />);
      await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());

      const cardButton = screen.getByAltText('♠ A').closest('button') as HTMLButtonElement;
      fireEvent.click(cardButton, { detail: 2 });
      expect(cardButton).toHaveAttribute('aria-pressed', 'false');
    });
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
      renderWithProviders(<EightOffPage />);
      await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
      const cardImg = screen.getByAltText('♠ K');
      const cardButton = cardImg.closest('button') as HTMLButtonElement;
      expect(cardButton).toHaveAttribute('draggable', 'true');
    });

    it('marks only the deepest movable cards as draggable when free cells + empty cols are exhausted', async () => {
      // Construct a stack of 3 cards in column 0 with all free cells full and no empty tableau cols.
      // Supermove limit becomes (1+0)*2^0 = 1, so only the bottom card (cardIndex=2) can move.
      const tightState: EightOffResponse = {
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
        freeCells: [
          card('DIAMOND', 8),
          card('CLOVER', 9),
          card('SPADE', 10),
          card('HEART', 11),
          card('DIAMOND', 5),
          card('CLOVER', 6),
          card('SPADE', 7),
          card('HEART', 8),
        ],
      };
      mockExec.mockResolvedValue(tightState);
      renderWithProviders(<EightOffPage />);
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
      const looseState: EightOffResponse = {
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
        freeCells: [card('DIAMOND', 8), card('CLOVER', 9), null, null, null, null, null, null],
      };
      mockExec.mockResolvedValue(looseState);
      renderWithProviders(<EightOffPage />);
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
      const looseState: EightOffResponse = {
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
        freeCells: [card('DIAMOND', 8), card('CLOVER', 9), null, null, null, null, null, null],
      };
      mockExec.mockResolvedValue(looseState);
      renderWithProviders(<EightOffPage />);
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
      renderWithProviders(<EightOffPage />);
      await waitFor(() => expect(screen.getByAltText('♦ 7')).toBeInTheDocument());
      const cardButton = screen.getByAltText('♦ 7').closest('button') as HTMLButtonElement;
      expect(cardButton).toHaveAttribute('draggable', 'true');
    });

    it('dragging tableau card to empty tableau column dispatches move', async () => {
      renderWithProviders(<EightOffPage />);
      await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

      const sourceCard = screen.getByAltText('♠ K');
      const sourceButton = sourceCard.closest('button') as HTMLButtonElement;
      const dataTransfer = buildDataTransfer();
      fireEvent.dragStart(sourceButton, { dataTransfer });

      // Find an empty tableau column (K placeholder)
      const kButtons = screen.getAllByRole('button').filter((btn) => btn.textContent === 'K');
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
});

// **上限を知るには赤くなった札にマウスを乗せるしかなかった (#4801)。**後追いの
// 体験になる。ほぼ同型の Penguin は同じ計算式のバッジをヘッダーに常設している。
describe('EightOffPage supermove badge', () => {
  it('shows the current limit without any interaction', async () => {
    renderWithProviders(<EightOffPage />);
    const badge = await screen.findByTestId('eo-supermove-badge');
    expect(badge).toBeInTheDocument();
  });

  // **空きセルが増えると上限も増える。**固定値を出すと、空きを作った意味が
  // 見えない。
  it('tracks the free cells rather than showing a fixed number', async () => {
    mockExec.mockResolvedValue({ ...playingState, freeCells: [null, null, null, null, null, null, null, null] });
    renderWithProviders(<EightOffPage />);
    const withAllFree = (await screen.findByTestId('eo-supermove-badge')).textContent;

    cleanup();
    mockExec.mockResolvedValue({
      ...playingState,
      freeCells: [{ design: 'SPADE', value: 5 }, { design: 'HEART', value: 6 }, null, null, null, null, null, null],
    });
    renderWithProviders(<EightOffPage />);
    const withTwoUsed = (await screen.findByTestId('eo-supermove-badge')).textContent;

    expect(withAllFree).not.toEqual(withTwoUsed);
  });
});

// #5612: 一括移動される塊のプレビューが `hoveredStack` (onMouseEnter/onFocus) にしか
// 追従していなかった。タッチ端末にホバーは無いので、タップで選んだ直後にフォーカスが
// 外れると、どこまでが一緒に動くのか手がかりが消える。Easthaven は #4815 で
// 選択状態にも追従させている。
describe('EightOffPage supermove block preview on touch', () => {
  // 2 枚の列を持ち、空きセル 8 + 空き列 6 → 上限は十分大きいので塊は動かせる。
  const twoCardCol: EightOffResponse = {
    ...playingState,
    tableau: [[card('SPADE', 13), card('HEART', 12)], [card('CLOVER', 13)], [], [], [], [], [], []],
  };

  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
    mockExec.mockResolvedValue(twoCardCol);
  });

  it('keeps the block marked after a tap selects the stack', async () => {
    renderWithProviders(<EightOffPage />);
    const deep = await screen.findByTestId('eo-tableau-0-0');

    // タップ = click。ホバーもフォーカスも無い状態にする。
    fireEvent.click(deep);
    await waitFor(() => expect(deep).toHaveAttribute('aria-pressed', 'true'));
    fireEvent.blur(deep);

    expect(deep).toHaveAttribute('data-supermove-block', 'true');
    // 選んだ札より下 (後ろ) の札も同じ塊。
    expect(screen.getByTestId('eo-tableau-0-1')).toHaveAttribute('data-supermove-block', 'true');
    // 別の列は巻き込まない。
    expect(screen.getByTestId('eo-tableau-1-0')).not.toHaveAttribute('data-supermove-block');
  });

  it('still marks the block on hover, as before', async () => {
    renderWithProviders(<EightOffPage />);
    const deep = await screen.findByTestId('eo-tableau-0-0');

    fireEvent.mouseEnter(deep);
    expect(deep).toHaveAttribute('data-supermove-block', 'true');
    fireEvent.mouseLeave(deep);
    expect(deep).not.toHaveAttribute('data-supermove-block');
  });

  // 上限を超える塊は、選んでもハイライトしない ── 動かない塊を「動く」と見せない。
  it('does not mark a selected stack that exceeds the supermove limit', async () => {
    mockExec.mockResolvedValue(supermoveBlockedState);
    renderWithProviders(<EightOffPage />);
    const deep = await screen.findByTestId('eo-tableau-0-0');
    expect(deep).toHaveAttribute('data-supermove-blocked', 'true');

    fireEvent.click(deep);
    fireEvent.blur(deep);
    expect(deep).not.toHaveAttribute('data-supermove-block');
  });
});
