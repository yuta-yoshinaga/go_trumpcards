import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { seahaventowersApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, SeahavenTowersResponse } from '../types/card';
import { SeahavenTowersPage } from './SeahavenTowersPage';

vi.mock('../api/gameApi', () => ({
  seahaventowersApi: { exec: vi.fn() },
  actionLogApi: { seahaventowers: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(seahaventowersApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: SeahavenTowersResponse = {
  tableau: [[card('SPADE', 13)], [card('SPADE', 12)], [], [], [], [], [], [], [], []],
  reservedCells: [null, null],
  foundation: [[], [], [], []],
  phase: 0,
  moveCount: 5,
  canUndo: true,
  isStalemate: false,
  message: '',
  messageCode: 'seahaventowers.playing',
};

const gameClearState: SeahavenTowersResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'seahaventowers.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: SeahavenTowersResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'seahaventowers.gameOver',
};

const withFoundationState: SeahavenTowersResponse = {
  ...playingState,
  foundation: [[card('SPADE', 1)], [], [card('HEART', 1), card('HEART', 2)], []],
};

const withReservedCardState: SeahavenTowersResponse = {
  ...playingState,
  reservedCells: [card('DIAMOND', 7), null],
};

const withHintState: SeahavenTowersResponse = {
  ...playingState,
  hint: { fromZone: 'tableau', fromCol: 0, cardIndex: 0, toZone: 'tableau', toCol: 2 },
};

// A column that ascends bottom→top (5 sits below 9) is not a clean mop-up, so
// auto-complete cannot guarantee a win: readiness is false.
const notReadyState: SeahavenTowersResponse = {
  ...playingState,
  tableau: [[card('SPADE', 5), card('SPADE', 9)], [], [], [], [], [], [], [], [], []],
};

beforeEach(() => {
  mockExec.mockResolvedValue(playingState);
  vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
});

describe('SeahavenTowersPage', () => {
  it('renders skeleton when state is null', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<SeahavenTowersPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('renders empty tableau columns with K placeholder', async () => {
    renderWithProviders(<SeahavenTowersPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    const kElements = screen.getAllByText('K');
    expect(kElements.length).toBeGreaterThanOrEqual(1);
  });

  it('shows the bulk-move (supermove) limit from empty reserved cells', async () => {
    // 2 empty reserved cells → 1 + 2 = 3.
    renderWithProviders(<SeahavenTowersPage />);
    const limit = await screen.findByTestId('st-supermove-limit');
    expect(limit).toHaveTextContent('3');
  });

  it('renders foundation piles with all four suit symbols', async () => {
    renderWithProviders(<SeahavenTowersPage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());
    expect(screen.getByText('♣')).toBeInTheDocument();
    expect(screen.getByText('♥')).toBeInTheDocument();
    expect(screen.getByText('♦')).toBeInTheDocument();
  });

  it('renders empty foundation with A placeholder', async () => {
    renderWithProviders(<SeahavenTowersPage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());
    const aElements = screen.getAllByText('A');
    expect(aElements.length).toBeGreaterThanOrEqual(1);
  });

  it('renders foundation with cards', async () => {
    mockExec.mockResolvedValue(withFoundationState);
    renderWithProviders(<SeahavenTowersPage />);
    await waitFor(() => expect(screen.getByText('♠')).toBeInTheDocument());
    const imgs = screen.getAllByRole('img');
    expect(imgs.length).toBeGreaterThanOrEqual(1);
  });

  it('renders both reserved cells (empty)', async () => {
    renderWithProviders(<SeahavenTowersPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    const emptyButtons = screen.getAllByText('空');
    expect(emptyButtons.length).toBe(2);
  });

  it('renders reserved cell with card occupied', async () => {
    mockExec.mockResolvedValue(withReservedCardState);
    renderWithProviders(<SeahavenTowersPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    expect(screen.getByAltText('♦ 7')).toBeInTheDocument();
    const emptyButtons = screen.getAllByText('空');
    expect(emptyButtons.length).toBe(1); // one remaining empty reserved cell
  });

  it('renders playing phase buttons', async () => {
    renderWithProviders(<SeahavenTowersPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'オートコンプリート' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument();
  });

  it('undo button disabled when canUndo is false', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: false });
    renderWithProviders(<SeahavenTowersPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '元に戻す' })).toBeDisabled());
  });

  it('hint button triggers hint API call', async () => {
    renderWithProviders(<SeahavenTowersPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValueOnce(withHintState);
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('hint display localizes zone names instead of raw English', async () => {
    renderWithProviders(<SeahavenTowersPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    mockExec.mockResolvedValue({
      ...withHintState,
      hint: { fromZone: 'reserved', fromCol: 1, cardIndex: -1, toZone: 'foundation', toCol: -1 },
    });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    // Zone identifiers render as localized names (ja), matching the CUI terminology.
    await waitFor(() => expect(screen.getByText(/リザーブセル 1.*→.*ファンデーション/)).toBeInTheDocument());
  });

  it('hint display omits column when col is negative and localizes both zones', async () => {
    renderWithProviders(<SeahavenTowersPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    mockExec.mockResolvedValue({
      ...withHintState,
      hint: { fromZone: 'foundation', fromCol: -1, cardIndex: -1, toZone: 'tableau', toCol: 2 },
    });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    // fromCol < 0 -> no number after the source zone; toCol >= 0 -> "タブロー 2".
    await waitFor(() => expect(screen.getByText(/ファンデーション.*→.*タブロー 2/)).toBeInTheDocument());
  });

  it('autocomplete button triggers autocomplete API call', async () => {
    renderWithProviders(<SeahavenTowersPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'オートコンプリート' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValueOnce(playingState);
    fireEvent.click(screen.getByRole('button', { name: 'オートコンプリート' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('shows the auto-complete ready badge and enables the button when all columns descend', async () => {
    // playingState columns are single/empty cards → strictly descending → ready.
    renderWithProviders(<SeahavenTowersPage />);
    await waitFor(() => expect(screen.getByTestId('seahaventowers-autocomplete-ready-badge')).toBeInTheDocument());
    const autoCompleteBtn = screen.getByTestId('autocomplete-button');
    expect(autoCompleteBtn).toBeEnabled();
    expect(autoCompleteBtn).not.toHaveAttribute('title');
    expect(autoCompleteBtn.className).toContain('animate-pulse');
  });

  it('hides the badge, disables the button, and shows a tooltip when not ready', async () => {
    mockExec.mockResolvedValue(notReadyState);
    renderWithProviders(<SeahavenTowersPage />);
    await waitFor(() => expect(screen.getByTestId('autocomplete-button')).toBeInTheDocument());
    expect(screen.queryByTestId('seahaventowers-autocomplete-ready-badge')).not.toBeInTheDocument();
    const autoCompleteBtn = screen.getByTestId('autocomplete-button');
    expect(autoCompleteBtn).toBeDisabled();
    expect(autoCompleteBtn).toHaveAttribute('title');
    expect(autoCompleteBtn.className).not.toContain('animate-pulse');
  });

  it('give up button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    renderWithProviders(<SeahavenTowersPage />);
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

  it('renders game-clear state', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<SeahavenTowersPage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームクリア/).length).toBeGreaterThan(0));
  });

  it('renders game-over state', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<SeahavenTowersPage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームオーバー/).length).toBeGreaterThan(0));
  });

  it('highlights the in-limit supermove block under the cursor', async () => {
    // Both reserved cells empty → limit = 1 + 2 = 3, so the bottom 3 cards form the movable block.
    const looseState: SeahavenTowersResponse = {
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
        [card('CLOVER', 8)],
        [card('DIAMOND', 9)],
      ],
      reservedCells: [null, null],
    };
    mockExec.mockResolvedValue(looseState);
    renderWithProviders(<SeahavenTowersPage />);
    await waitFor(() => expect(screen.getByAltText('♥ Q')).toBeInTheDocument());

    const middleButton = screen.getByAltText('♥ Q').closest('button') as HTMLButtonElement;
    const bottomButton = screen.getByAltText('♣ J').closest('button') as HTMLButtonElement;
    fireEvent.mouseEnter(middleButton);
    expect(middleButton).toHaveAttribute('data-supermove-block', 'true');
    expect(bottomButton).toHaveAttribute('data-supermove-block', 'true');
    fireEvent.mouseLeave(middleButton);
    expect(middleButton).not.toHaveAttribute('data-supermove-block');
  });
});

// Keyboard shortcuts are bound by useActionKeyboardNav and advertised by
// ActionShortcutsPanel, but nothing asserted that pressing a key actually runs
// its action — a wrong `key` or a wrong `enabled` condition would have failed no
// test. See issue #4429.
describe('SeahavenTowersPage keyboard shortcuts', () => {
  it.each([
    ['h', 'hint'],
    ['a', 'autocomplete'],
    ['z', 'undo'],
  ])('pressing %s dispatches %s', async (key, command) => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<SeahavenTowersPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.keyDown(document, { key });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
  });

  it('pressing g asks for give-up confirmation rather than firing it', async () => {
    // give-up is irreversible, so the key must route through the dialog (#2099)
    // instead of dispatching straight away.
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<SeahavenTowersPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'g' });
    expect(await screen.findByText('投了確認')).toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('ignores shortcuts once the game has ended', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<SeahavenTowersPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    for (const key of ['h', 'a', 'z']) {
      fireEvent.keyDown(document, { key });
    }
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });
});
