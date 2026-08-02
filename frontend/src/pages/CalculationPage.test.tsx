import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { calculationApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { CalculationResponse, Card, CardDesign } from '../types/card';
import { CalculationPage, calculationNextRank } from './CalculationPage';

vi.mock('../api/gameApi', () => ({
  calculationApi: { exec: vi.fn() },
  actionLogApi: { calculation: vi.fn() },
}));

const mockExec = vi.mocked(calculationApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: CalculationResponse = {
  foundations: [[card('SPADE', 1)], [card('HEART', 2)], [card('DIAMOND', 3)], [card('CLOVER', 4)]],
  wastes: [[], [], [], []],
  stockCount: 48,
  stockTop: card('SPADE', 7),
  phase: 0,
  moveCount: 0,
  canUndo: false,
  isStalemate: false,
  message: '',
  messageCode: 'calculation.playing',
};

const gameClearState: CalculationResponse = {
  ...playingState,
  phase: 1,
  stockCount: 0,
  stockTop: undefined,
  moveCount: 100,
  messageCode: 'calculation.gameClear',
  messageParams: { moveCount: '100' },
};

const gameOverState: CalculationResponse = {
  ...playingState,
  phase: 2,
  messageCode: 'calculation.gameOver',
};

beforeEach(() => {
  vi.clearAllMocks();
  localStorage.clear();
  mockExec.mockResolvedValue(playingState);
});

describe('CalculationPage', () => {
  it('renders heading', async () => {
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('labels the stock button with its top card and empty waste piles explicitly', async () => {
    renderWithProviders(<CalculationPage />);
    const stock = await screen.findByTestId('calc-stock-button');
    expect(stock).toHaveAttribute('aria-label', '山札のトップ: ♠ 7');
    // Empty waste piles now carry an explicit name (previously undefined).
    expect(screen.getByTestId('calc-waste-button-0')).toHaveAttribute('aria-label', 'ウェイスト0: 空');
    expect(screen.getByTestId('calc-waste-button-3')).toHaveAttribute('aria-label', 'ウェイスト3: 空');
  });

  it('labels a non-empty waste pile with its top ranks, not the empty text', async () => {
    mockExec.mockResolvedValue({ ...playingState, wastes: [[card('HEART', 9)], [], [], []] });
    renderWithProviders(<CalculationPage />);
    const waste0 = await screen.findByTestId('calc-waste-button-0');
    await waitFor(() => expect(waste0.getAttribute('aria-label')).toContain('ウェイスト0'));
    expect(waste0).not.toHaveAttribute('aria-label', 'ウェイスト0: 空');
  });

  it('shows move count', async () => {
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(screen.getByText(/手数/)).toBeInTheDocument());
  });

  it('shows game clear phase', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームクリア/).length).toBeGreaterThan(0));
  });

  it('shows game over phase', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームオーバー/).length).toBeGreaterThan(0));
  });

  it('hint button triggers hint command', async () => {
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    screen.getByRole('button', { name: 'ヒント' }).click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('giveup button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    // Clicking give-up must NOT dispatch immediately — it opens a confirm dialog (#2099).
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();
    // Confirming dispatches giveup.
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('undo button disabled when canUndo is false', async () => {
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByRole('button', { name: '元に戻す' })).toBeDisabled();
  });

  it('undo button fires undo command when canUndo is true', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    screen.getByRole('button', { name: '元に戻す' }).click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('shows error alert when API fails on mount', async () => {
    mockExec.mockRejectedValue(new Error('network error'));
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('autocomplete button is disabled while stock is non-empty', async () => {
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByTestId('autocomplete-button')).toBeDisabled();
  });

  it('autocomplete button enables when stock is empty and wastes are non-empty', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      stockCount: 0,
      stockTop: undefined,
      wastes: [[card('SPADE', 5)], [], [], []],
    });
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const btn = screen.getByTestId('autocomplete-button');
    expect(btn).not.toBeDisabled();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('shows the bottom-up rank preview as a tooltip/aria-label on a non-empty waste pile', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      // Array order is bottom→top; the last <=3 (5, K, A) render in order.
      wastes: [[card('SPADE', 2), card('HEART', 5), card('DIAMOND', 13), card('CLOVER', 1)], [], [], []],
    });
    renderWithProviders(<CalculationPage />);
    const btn = await screen.findByTestId('calc-waste-button-0');
    expect(btn).toHaveAttribute('title', 'ウェイスト0（上3枚）: 5・K・A');
    expect(btn).toHaveAttribute('aria-label', 'ウェイスト0（上3枚）: 5・K・A');
  });

  it('omits the rank tooltip on an empty waste pile but still names it for SR users', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<CalculationPage />);
    const btn = await screen.findByTestId('calc-waste-button-0');
    // No hover tooltip when empty…
    expect(btn).not.toHaveAttribute('title');
    // …but an explicit spoken name so the empty pile isn't anonymous.
    expect(btn).toHaveAttribute('aria-label', 'ウェイスト0: 空');
  });

  it('clicking foundation without a source has no effect', async () => {
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    const f0 = screen.getByLabelText(/ファンデーション 0 \+1/);
    fireEvent.click(f0);
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('selecting stock and clicking a foundation dispatches a stock→foundation move', async () => {
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();

    fireEvent.click(screen.getByTestId('calc-stock-button'));
    fireEvent.click(screen.getByLabelText(/ファンデーション 2 \+3/));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'stock' }, { zone: 'foundation', idx: 2 }),
    );
  });

  it('selecting stock and clicking an empty waste dispatches a stock→waste move', async () => {
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();

    fireEvent.click(screen.getByTestId('calc-stock-button'));
    fireEvent.click(screen.getByTestId('calc-waste-button-1'));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', { zone: 'stock' }, { zone: 'waste', idx: 1 }));
  });

  it('selecting a waste and clicking a foundation dispatches a waste→foundation move', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      wastes: [[], [card('HEART', 4)], [], []],
    });
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();

    fireEvent.click(screen.getByTestId('calc-waste-button-1'));
    fireEvent.click(screen.getByLabelText(/ファンデーション 1 \+2/));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'waste', idx: 1 }, { zone: 'foundation', idx: 1 }),
    );
  });

  it('clicking the same source twice toggles the selection off', async () => {
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const stockBtn = screen.getByTestId('calc-stock-button');
    fireEvent.click(stockBtn);
    expect(stockBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.click(stockBtn);
    expect(stockBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('cancel button clears the active selection', async () => {
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    fireEvent.click(screen.getByTestId('calc-stock-button'));
    const cancelBtn = await screen.findByRole('button', { name: 'キャンセル' });
    fireEvent.click(cancelBtn);
    await waitFor(() => expect(screen.queryByRole('button', { name: 'キャンセル' })).toBeNull());
  });

  it('clicking different waste tops switches the selection', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      wastes: [[card('SPADE', 9)], [card('HEART', 5)], [], []],
    });
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const w0 = screen.getByTestId('calc-waste-button-0');
    const w1 = screen.getByTestId('calc-waste-button-1');
    fireEvent.click(w0);
    expect(w0).toHaveAttribute('aria-pressed', 'true');
    fireEvent.click(w1);
    expect(w0).toHaveAttribute('aria-pressed', 'false');
    expect(w1).toHaveAttribute('aria-pressed', 'true');
  });

  it('renders the backend hint banner when state.hint is a stock hint', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      hint: { fromZone: 'stock', wasteIdx: -1, foundationIdx: 2 },
      messageCode: 'calculation.hintAvailable',
    });
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // messageCode を付けるとメッセージ枠にも同じ文が出るので、バナー側だけを見る。
    // Hint uses localized zone names + index, not raw F/W symbols.
    expect(screen.getByText(/ストック → ファンデーション 2/)).toBeInTheDocument();
  });

  it('renders the backend hint banner when state.hint is a waste hint', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      wastes: [[card('SPADE', 5)], [], [], []],
      hint: { fromZone: 'waste', wasteIdx: 0, foundationIdx: 1 },
      messageCode: 'calculation.hintAvailable',
    });
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByText(/ウェイスト 0 → ファンデーション 1/)).toBeInTheDocument();
  });

  it('renders a stalemate escape button when the game is stalled', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      isStalemate: true,
      canUndo: true,
      undoToEscape: 2,
    });
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByRole('button', { name: /脱出する/ })).toBeInTheDocument();
  });

  it('reset confirmation dialog dispatches a reset when confirmed', async () => {
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);

    const resetBtn = screen.getAllByRole('button', { name: /reset|リセット/i })[0];
    fireEvent.click(resetBtn);
    const confirmBtn = await screen.findByRole('button', { name: '確認' });
    fireEvent.click(confirmBtn);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('CLI toggle switches to the CLI terminal', async () => {
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    const cliToggle = screen.getByRole('button', { name: 'CLIモードに切り替え' });
    fireEvent.click(cliToggle);

    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
  });

  it('renders the next-rank badge on each foundation reflecting the +1/+2/+3/+4 progression', async () => {
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // Foundations seeded with A, 2, 3, 4 ⇒ next required ranks 2, 4, 6, 8.
    expect(screen.getByTestId('calc-foundation-next-0')).toHaveTextContent('次:2');
    expect(screen.getByTestId('calc-foundation-next-1')).toHaveTextContent('次:4');
    expect(screen.getByTestId('calc-foundation-next-2')).toHaveTextContent('次:6');
    expect(screen.getByTestId('calc-foundation-next-3')).toHaveTextContent('次:8');
  });

  it('hides the next-rank badge once a foundation reaches K (13 cards)', async () => {
    const fullPile = Array.from({ length: 13 }, (_, i) => card('SPADE', i + 1));
    mockExec.mockResolvedValue({
      ...playingState,
      foundations: [fullPile, [card('HEART', 2)], [card('DIAMOND', 3)], [card('CLOVER', 4)]],
    });
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByTestId('calc-foundation-next-0')).not.toBeInTheDocument();
    expect(screen.getByTestId('calc-foundation-next-1')).toHaveTextContent('次:4');
    // The count readout reflects the foundation-cap constant (13/13) for a full pile.
    expect(screen.getByText('13/13')).toBeInTheDocument();
  });

  it('exposes the full upcoming-rank sequence as visible, tappable text (not only a title attr)', async () => {
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // F0 seeded with A, step +1 ⇒ upcoming full sequence starts at 2.
    const details = screen.getByTestId('calc-foundation-upcoming-details-0');
    expect(details).toBeInTheDocument();
    // The full sequence is rendered as real DOM text, reachable on touch devices
    // by tapping the <summary> (native <details> disclosure — no hover needed).
    const summary = details.querySelector('summary');
    expect(summary).not.toBeNull();
    if (summary) {
      fireEvent.click(summary);
    }
    const full = screen.getByTestId('calc-foundation-upcoming-full-0');
    expect(full).toBeVisible();
    // Sequence must be the complete run to K (13), joined with arrows.
    expect(full.textContent).toContain('→');
    expect(full).toHaveAccessibleName(/2/);
  });

  it('shows the seed rank on an empty foundation (no top card yet)', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      foundations: [[], [card('HEART', 2)], [card('DIAMOND', 3)], [card('CLOVER', 4)]],
    });
    renderWithProviders(<CalculationPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    // F0 with step +1 and no cards ⇒ next required rank is A (1).
    expect(screen.getByTestId('calc-foundation-next-0')).toHaveTextContent('次:A');
  });
});

describe('calculationNextRank', () => {
  it('returns the step value when the pile is empty', () => {
    expect(calculationNextRank(0, undefined, 0)).toBe(1);
    expect(calculationNextRank(1, undefined, 0)).toBe(2);
    expect(calculationNextRank(2, undefined, 0)).toBe(3);
    expect(calculationNextRank(3, undefined, 0)).toBe(4);
  });

  it('wraps modulo 13 when the next value exceeds K', () => {
    // F3 (+4): K(13) → 4 → 8 → 12 → 3 → 7 → ...
    expect(calculationNextRank(3, 13, 4)).toBe(4);
    expect(calculationNextRank(3, 12, 7)).toBe(3);
    // F2 (+3): J(11) → A(1)
    expect(calculationNextRank(2, 11, 4)).toBe(1);
  });

  it('returns null once the pile already contains 13 cards', () => {
    expect(calculationNextRank(0, 13, 13)).toBeNull();
  });
});
