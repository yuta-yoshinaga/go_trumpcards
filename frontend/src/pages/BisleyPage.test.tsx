import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { bisleyApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BisleyResponse, BisleyTableauCard, Card, CardDesign } from '../types/card';
import { BisleyPage } from './BisleyPage';

vi.mock('../api/gameApi', () => ({
  bisleyApi: { exec: vi.fn() },
  actionLogApi: { bisley: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(bisleyApi.exec);

function makeTableau(cols: BisleyTableauCard[][]): BisleyTableauCard[][] {
  return Array.from({ length: 13 }, (_, i) => cols[i] ?? []);
}

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: BisleyResponse = {
  tableau: makeTableau([
    [
      { card: card('SPADE', 9), faceUp: true },
      { card: card('SPADE', 5), faceUp: true },
    ],
    [{ card: card('HEART', 6), faceUp: true }],
  ]),
  aceFoundations: [[card('SPADE', 1)], [card('CLOVER', 1)], [card('HEART', 1)], [card('DIAMOND', 1)]],
  kingFoundations: [[], [], [], []],
  phase: 0,
  moveCount: 3,
  canUndo: false,
  isStalemate: false,
  message: '',
};

const gameClearState: BisleyResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'bisley.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: BisleyResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'bisley.gameOver',
};

describe('BisleyPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  });

  it('calls reset on initial render', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BisleyPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
  });

  it('renders heading and move count', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BisleyPage />);
    await waitFor(() => expect(screen.getByText(/ビズリー/)).toBeInTheDocument());
    expect(screen.getByText(/手数: 3/)).toBeInTheDocument();
  });

  it('renders four ascending piles and four empty descending piles', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BisleyPage />);
    await waitFor(() => expect(screen.getAllByLabelText(/昇順組札 1枚/).length).toBe(4));
    expect(screen.getAllByLabelText(/空の降順組札/).length).toBe(4);
  });

  it('labels all thirteen tableau columns with their 0-based index (matching hint text)', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BisleyPage />);
    await waitFor(() => expect(screen.getByText('#0')).toBeInTheDocument());
    // Columns are numbered #0..#12 to match formatHintZone's raw fromCol/toIdx.
    for (let i = 0; i < 13; i++) {
      expect(screen.getByText(`#${i}`)).toBeInTheDocument();
    }
  });

  it('renders empty columns as non-interactive placeholders', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BisleyPage />);
    // Bisley forbids moving onto an empty column, so unlike the other open
    // tableaus these must not be drop/click targets at all.
    await waitFor(() => expect(screen.getAllByLabelText(/空のタブロー列/).length).toBe(11));
    expect(screen.queryByRole('button', { name: /空のタブロー列/ })).not.toBeInTheDocument();
  });

  it('renders giveup button when playing', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BisleyPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
  });

  it('hides giveup button when game cleared', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<BisleyPage />);
    await waitFor(() => expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument());
  });

  it('giveup button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BisleyPage />);
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

  it('shows phase name in header for game over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<BisleyPage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームオーバー/).length).toBeGreaterThan(0));
  });

  it('counts both foundation sets in the game-over summary', async () => {
    mockExec.mockResolvedValue(gameOverState); // 4 aces, no kings → 4/52 (8%)
    renderWithProviders(<BisleyPage />);
    const summary = await screen.findByTestId('bisley-gameover-summary');
    expect(summary).toHaveTextContent('4/52');
    expect(summary).toHaveTextContent('8%');
  });

  it('includes the descending piles in the summary count', async () => {
    mockExec.mockResolvedValue({
      ...gameOverState,
      kingFoundations: [[card('SPADE', 13), card('SPADE', 12)], [card('CLOVER', 13)], [], []],
    });
    renderWithProviders(<BisleyPage />);
    const summary = await screen.findByTestId('bisley-gameover-summary');
    expect(summary).toHaveTextContent('7/52');
  });

  it('does not show the progress summary on game clear', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<BisleyPage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームクリア/).length).toBeGreaterThan(0));
    expect(screen.queryByTestId('bisley-gameover-summary')).not.toBeInTheDocument();
  });

  it('disables the auto-complete button and shows a reason when only the dealt aces are up', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BisleyPage />);
    const btn = await screen.findByTestId('autocomplete-button');
    expect(btn).toBeDisabled();
    expect(btn.className).not.toContain('animate-pulse');
    expect(btn).toHaveAttribute('title');
  });

  it('enables and pulses auto-complete once an ascending pile builds past its ace', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      aceFoundations: [
        [card('SPADE', 1), card('SPADE', 2)],
        [card('CLOVER', 1)],
        [card('HEART', 1)],
        [card('DIAMOND', 1)],
      ],
    });
    renderWithProviders(<BisleyPage />);
    const btn = await screen.findByTestId('autocomplete-button');
    expect(btn).toBeEnabled();
    expect(btn.className).toContain('animate-pulse');
  });

  it('enables auto-complete as soon as a descending pile opens', async () => {
    // The King foundations start empty, so a single card there already counts as
    // progress — unlike the aces, which are seeded by the deal.
    mockExec.mockResolvedValue({
      ...playingState,
      kingFoundations: [[card('SPADE', 13)], [], [], []],
    });
    renderWithProviders(<BisleyPage />);
    const btn = await screen.findByTestId('autocomplete-button');
    expect(btn).toBeEnabled();
  });

  it('shows StalemateEscapeButton when stalemate flag is set', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      isStalemate: true,
      undoToEscape: 2,
      canUndo: true,
    });
    renderWithProviders(<BisleyPage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
  });

  it('selecting a tableau card marks it as selected', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BisleyPage />);
    const sourceBtn = await screen.findByRole('button', { name: /^♠ 5（/ });
    fireEvent.click(sourceBtn);
    await waitFor(() => expect(sourceBtn).toHaveAttribute('aria-pressed', 'true'));
  });

  it('sends the descending zone when a King foundation is chosen as target', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BisleyPage />);
    const sourceBtn = await screen.findByRole('button', { name: /^♠ 5（/ });
    fireEvent.click(sourceBtn);
    await waitFor(() => expect(sourceBtn).toHaveAttribute('aria-pressed', 'true'));
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '空の降順組札 (♠)' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'tableau', col: 0 }, { zone: 'king', col: 0 }),
    );
  });

  it('names each tableau card with its position for screen readers', async () => {
    // Earlier tests in this file queue one-shot resolutions and can leave CLI
    // mode persisted in localStorage; reset both so the board actually renders.
    localStorage.clear();
    mockExec.mockReset();
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BisleyPage />);
    await waitFor(() => expect(screen.getAllByLabelText(/列\d+・上から\d+枚目/).length).toBeGreaterThan(0));
  });
});

// Keyboard shortcuts are bound by useActionKeyboardNav and advertised by
// ActionShortcutsPanel; assert the keys actually run their action (#4429).
describe('BisleyPage keyboard shortcuts', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  });

  it.each([
    ['h', 'hint'],
    ['a', 'autocomplete'],
    ['z', 'undo'],
  ])('pressing %s dispatches %s', async (key, command) => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BisleyPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.keyDown(document, { key });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
  });

  it('pressing g asks for give-up confirmation rather than firing it', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BisleyPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'g' });
    expect(await screen.findByText('投了確認')).toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('ignores shortcuts once the game has ended', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<BisleyPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    for (const key of ['h', 'a', 'z']) {
      fireEvent.keyDown(document, { key });
    }
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });
});
describe('BisleyPage hover targets', () => {
  beforeEach(() => {
    localStorage.clear();
    vi.clearAllMocks();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  });

  it('renders hover targets on the top card of a tableau column', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      tableau: makeTableau([
        // Normal card (5) -> [4, 6] -> 4, 6
        [{ card: card('SPADE', 5), faceUp: true }],
        // Ace (1) -> [2] -> 2
        [{ card: card('HEART', 1), faceUp: true }],
        // King (13) -> [12] -> Q
        [{ card: card('CLOVER', 13), faceUp: true }],
        // Empty column
        [],
      ]),
    });
    renderWithProviders(<BisleyPage />);

    // **同スート限定なのでスートまで出す。** ランクだけだと「どの 5 でも置ける」
    // と読めてしまう (規則は同スートで1つ違い)。
    //
    // 文言は2箇所に出る: 目で見るツールチップ (aria-hidden) と、読み上げ用の
    // 常設 sr-only。視覚のツールチップは display:none なので a11y ツリーに
    // 載らず、aria-describedby が届くのは sr-only の方 (#6349)。
    expect(await screen.findAllByText('♠ 4 / ♠ 6 が置けます')).toHaveLength(2);
    expect(await screen.findAllByText('♥ 2 が置けます')).toHaveLength(2);
    expect(await screen.findAllByText('♣ Q が置けます')).toHaveLength(2);

    // 読み上げに届く形で紐づいていること。説明が在っても紐づいていなければ読まれない。
    const spadeFive = screen.getByRole('button', { name: /♠ 5/ });
    const describedBy = spadeFive.getAttribute('aria-describedby');
    expect(describedBy).toBeTruthy();
    expect(document.getElementById(describedBy ?? '')?.textContent).toBe('♠ 4 / ♠ 6 が置けます');

    // 否定コントロール: Ace や King の逆方向がないこと
    expect(screen.queryByText(/0.*が置けます/)).not.toBeInTheDocument();
    expect(screen.queryByText(/14.*が置けます/)).not.toBeInTheDocument();
    expect(screen.queryByText(/A が置けます/)).not.toBeInTheDocument();
    expect(screen.queryByText(/K が置けます/)).not.toBeInTheDocument();

    // 空列には出ないこと。3列 x（sr-only + 目で見るツールチップ）= 6 で、
    // 4列目（空）は何も出さない。
    expect(screen.getAllByText(/が置けます/)).toHaveLength(6);
  });
});
