import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { braidApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BraidResponse, Card, CardDesign } from '../types/card';
import { BraidPage } from './BraidPage';

vi.mock('../api/gameApi', () => ({
  braidApi: { exec: vi.fn() },
  actionLogApi: { braid: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(braidApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function makeSlots(n: number, filled: Record<number, Card>): (Card | null)[] {
  return Array.from({ length: n }, (_, i) => filled[i] ?? null);
}

const playingState: BraidResponse = {
  braid: [card('CLOVER', 3), card('SPADE', 9)],
  fields: makeSlots(4, { 0: card('HEART', 8), 2: card('SPADE', 7) }),
  helpers: makeSlots(8, { 0: card('DIAMOND', 4) }),
  foundation: Array.from({ length: 8 }, () => []),
  stockCount: 71,
  waste: [],
  baseRank: 5,
  direction: 1,
  awaitingDirection: false,
  redealsLeft: 2,
  canRedeal: false,
  phase: 0,
  moveCount: 3,
  canUndo: false,
  isStalemate: false,
  message: '',
};

const awaitingDirectionState: BraidResponse = {
  ...playingState,
  direction: 0,
  awaitingDirection: true,
  moveCount: 0,
};

const gameClearState: BraidResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'braid.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: BraidResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'braid.gameOver',
};

describe('BraidPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  });

  it('calls reset on initial render', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BraidPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
  });

  it('renders heading, base rank, direction and move count', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BraidPage />);
    // "ブレイド" is both the page title and the zone label, so target the heading.
    await waitFor(() => expect(screen.getByRole('heading', { name: 'ブレイド' })).toBeInTheDocument());
    expect(screen.getByText(/開始ランク: 5/)).toBeInTheDocument();
    expect(screen.getByText(/向き: 昇順/)).toBeInTheDocument();
    expect(screen.getByText(/手数: 3/)).toBeInTheDocument();
  });

  it('renders eight foundations, four fields and eight helpers', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BraidPage />);
    await waitFor(() => expect(screen.getAllByLabelText(/空の組札\d+/).length).toBe(8));
    expect(screen.getAllByLabelText(/空のブレイド札\d+/).length).toBe(2);
    expect(screen.getAllByLabelText(/空のヘルパー\d+/).length).toBe(7);
  });

  // Nothing reaches a foundation until the direction is fixed, so the board
  // leads with the two buttons rather than a passive notice.
  it('offers the direction buttons while it is unset', async () => {
    mockExec.mockResolvedValue(awaitingDirectionState);
    renderWithProviders(<BraidPage />);
    await waitFor(() => expect(screen.getByText(/向き: 未選択/)).toBeInTheDocument());
    expect(screen.getByTestId('direction-up')).toBeInTheDocument();
    expect(screen.getByTestId('direction-down')).toBeInTheDocument();
  });

  it.each([
    ['direction-up', true],
    ['direction-down', false],
  ])('%s dispatches dir with the flag', async (testId, ascending) => {
    mockExec.mockResolvedValue(awaitingDirectionState);
    renderWithProviders(<BraidPage />);
    const btn = await screen.findByTestId(testId);
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('dir', undefined, undefined, undefined, ascending));
  });

  it('hides the direction buttons once it is fixed', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BraidPage />);
    await waitFor(() => expect(screen.getByText(/向き: 昇順/)).toBeInTheDocument());
    expect(screen.queryByTestId('direction-up')).not.toBeInTheDocument();
  });

  it('draws from the stock', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BraidPage />);
    const stock = await screen.findByRole('button', { name: /山札 残り71枚/ });
    mockExec.mockClear();
    fireEvent.click(stock);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  // The braid's only destination is a foundation, and its depth matters because
  // it only ever shrinks.
  it('sends the braid tail to a foundation and shows its depth', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BraidPage />);
    const braid = await screen.findByRole('button', {
      name: 'ブレイド 残り2枚（末尾のみ・組札にのみ出せます）',
    });
    fireEvent.click(braid);
    await waitFor(() => expect(braid).toHaveAttribute('aria-pressed', 'true'));
    mockExec.mockClear();

    fireEvent.click(screen.getByRole('button', { name: '空の組札0' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'braid' }, { zone: 'foundation', col: 0 }),
    );
  });

  it('shows an empty braid slot once it runs out', async () => {
    mockExec.mockResolvedValue({ ...playingState, braid: [] });
    renderWithProviders(<BraidPage />);
    await waitFor(() => expect(screen.getByLabelText('ブレイドは空です')).toBeInTheDocument());
  });

  // A braid field refills itself from the braid, so an empty one is not a target.
  it('renders an empty braid field as a non-interactive slot', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BraidPage />);
    await waitFor(() => expect(screen.getByLabelText(/空のブレイド札1/)).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: /空のブレイド札1/ })).not.toBeInTheDocument();
  });

  // An empty helper *is* a target -- but only the waste can fill it.
  it('renders an empty helper as a target', async () => {
    mockExec.mockResolvedValue({ ...playingState, waste: [card('HEART', 2)] });
    renderWithProviders(<BraidPage />);
    const waste = await screen.findByRole('button', { name: '♥ 2' });
    fireEvent.click(waste);
    await waitFor(() => expect(waste).toHaveAttribute('aria-pressed', 'true'));
    mockExec.mockClear();

    fireEvent.click(screen.getByRole('button', { name: /空のヘルパー3/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'waste' }, { zone: 'helper', col: 3 }));
  });

  it('sends a braid field to a foundation', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BraidPage />);
    const fieldCard = await screen.findByRole('button', { name: '♥ 8' });
    fireEvent.click(fieldCard);
    await waitFor(() => expect(fieldCard).toHaveAttribute('aria-pressed', 'true'));
    mockExec.mockClear();

    fireEvent.click(screen.getByRole('button', { name: '空の組札1' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'field', col: 0 }, { zone: 'foundation', col: 1 }),
    );
  });

  it('shows an empty waste slot when nothing has been turned', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BraidPage />);
    await waitFor(() => expect(screen.getByLabelText('捨て札は空です')).toBeInTheDocument());
  });

  it('shows the remaining redeals', async () => {
    mockExec.mockResolvedValue({ ...playingState, redealsLeft: 1 });
    renderWithProviders(<BraidPage />);
    await waitFor(() => expect(screen.getByText(/めくり直し残り1回/)).toBeInTheDocument());
  });

  // An empty stock is still drawable while a redeal is left -- that is what
  // recycles the waste.
  it('keeps the stock drawable while a redeal remains', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0, canRedeal: true });
    renderWithProviders(<BraidPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'めくる' })).toBeEnabled());
  });

  it('disables the stock once the redeals are gone', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0, canRedeal: false, redealsLeft: 0 });
    renderWithProviders(<BraidPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /山札は空です/ })).toBeDisabled());
    expect(screen.getByRole('button', { name: 'めくる' })).toBeDisabled();
  });

  it('renders giveup button when playing and hides it once cleared', async () => {
    mockExec.mockResolvedValue(playingState);
    const { unmount } = renderWithProviders(<BraidPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
    unmount();

    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<BraidPage />);
    await waitFor(() => expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument());
  });

  it('giveup opens a confirm dialog and only dispatches after confirm', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BraidPage />);
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

  it('counts the summary against 104 cards, not 52', async () => {
    mockExec.mockResolvedValue({
      ...gameOverState,
      foundation: [[card('SPADE', 5)], [], [], [], [], [], [], []],
    });
    renderWithProviders(<BraidPage />);
    const summary = await screen.findByTestId('br-gameover-summary');
    expect(summary).toHaveTextContent('1/104');
  });

  it('does not show the progress summary on game clear', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<BraidPage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームクリア/).length).toBeGreaterThan(0));
    expect(screen.queryByTestId('br-gameover-summary')).not.toBeInTheDocument();
  });

  // Auto-complete cannot move anything before the direction is fixed, even
  // though the starter card already sits on a foundation.
  it('keeps auto-complete disabled until the direction is fixed', async () => {
    mockExec.mockResolvedValue({
      ...awaitingDirectionState,
      foundation: [[card('SPADE', 5)], [], [], [], [], [], [], []],
    });
    const { unmount } = renderWithProviders(<BraidPage />);
    const btn = await screen.findByTestId('autocomplete-button');
    expect(btn).toBeDisabled();
    expect(btn).toHaveAttribute('title');
    unmount();

    mockExec.mockResolvedValue({
      ...playingState,
      foundation: [[card('SPADE', 5)], [], [], [], [], [], [], []],
    });
    renderWithProviders(<BraidPage />);
    await waitFor(() => expect(screen.getByTestId('autocomplete-button')).toBeEnabled());
  });

  it('shows StalemateEscapeButton when the stalemate flag is set', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: 2, canUndo: true });
    renderWithProviders(<BraidPage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
  });

  it.each([
    ['braid field', { fromZone: 'field', fromIdx: 2, toZone: 'foundation', toIdx: 3 }, '組札3'],
    ['helper', { fromZone: 'helper', fromIdx: 1, toZone: 'foundation', toIdx: 0 }, 'ヘルパー1'],
    ['draw', { fromZone: 'stock', fromIdx: -1, toZone: 'waste', toIdx: -1 }, '山札'],
  ])('renders a %s hint after the hint button is pressed', async (_name, hint, expected) => {
    mockExec.mockResolvedValueOnce(playingState).mockResolvedValueOnce({ ...playingState, hint });
    renderWithProviders(<BraidPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getAllByText(new RegExp(expected)).length).toBeGreaterThan(0));
  });

  it('swaps the board for a terminal when CLI mode is toggled', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BraidPage />);
    await waitFor(() => expect(screen.getByLabelText(/空の組札0/)).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /CLI/i }));
    await waitFor(() => expect(screen.queryByLabelText(/空の組札0/)).not.toBeInTheDocument());
  });

  it('rings the hinted cards, not just their names', async () => {
    localStorage.clear();
    mockExec.mockReset();
    mockExec.mockResolvedValueOnce(playingState).mockResolvedValueOnce({
      ...playingState,
      hint: { fromZone: 'field', fromIdx: 2, toZone: 'foundation', toIdx: 3 },
    });
    renderWithProviders(<BraidPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    expect(document.querySelectorAll('[data-hint-slot]')).toHaveLength(0);

    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    // Four fields, eight helpers and eight foundations: prose alone left the
    // player hunting for the two cards named.
    await waitFor(() => expect(screen.getAllByText(/組札3/).length).toBeGreaterThan(0));
    expect(document.querySelectorAll('[data-hint-slot="from"]')).toHaveLength(1);
    // The destination may be an empty foundation, which must ring too.
    expect(document.querySelectorAll('[data-hint-slot="to"]')).toHaveLength(1);
  });
});

// Keyboard shortcuts are bound by useActionKeyboardNav and advertised by
// ActionShortcutsPanel; assert the keys actually run their action (#4429).
describe('BraidPage keyboard shortcuts', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  });

  it.each([
    ['d', 'draw'],
    ['h', 'hint'],
    ['a', 'autocomplete'],
    ['z', 'undo'],
  ])('pressing %s dispatches %s', async (key, command) => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BraidPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.keyDown(document, { key });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
  });

  it('pressing g asks for give-up confirmation rather than firing it', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BraidPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'g' });
    expect(await screen.findByText('投了確認')).toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('ignores shortcuts once the game has ended', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<BraidPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    for (const key of ['d', 'h', 'a', 'z']) {
      fireEvent.keyDown(document, { key });
    }
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });
});
