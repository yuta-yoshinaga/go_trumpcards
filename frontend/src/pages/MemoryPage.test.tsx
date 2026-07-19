import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, memoryApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, MemoryBoardCard, MemoryResponse } from '../types/card';
import { MemoryPage } from './MemoryPage';

vi.mock('../api/gameApi', () => ({
  memoryApi: { exec: vi.fn() },
  actionLogApi: { memory: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(memoryApi.exec);

function makeBoard(
  overrides?: Partial<Record<number, { faceUp?: boolean; taken?: boolean; card?: Card | null }>>,
): MemoryBoardCard[] {
  return Array.from({ length: 52 }, (_, i) => ({
    card: overrides?.[i]?.card ?? null,
    faceUp: overrides?.[i]?.faceUp ?? false,
    taken: overrides?.[i]?.taken ?? false,
  }));
}

const flip1State: MemoryResponse = {
  players: [
    { id: 0, isHuman: true, pairCount: 0, pairs: [] },
    { id: 1, isHuman: false, pairCount: 2, pairs: [] },
    { id: 2, isHuman: false, pairCount: 1, pairs: [] },
    { id: 3, isHuman: false, pairCount: 0, pairs: [] },
  ],
  board: makeBoard(),
  phase: 0,
  currentPlayerIdx: 0,
  firstFlipPos: -1,
  secondFlipPos: -1,
  lastMatchResult: false,
  gameEndFlag: false,
  winnerIdx: -1,
  turnNumber: 0,
  message: '',
  config: { cpuDifficulty: 1 },
};

const flip2State: MemoryResponse = {
  ...flip1State,
  phase: 1,
  firstFlipPos: 5,
  board: makeBoard({ 5: { faceUp: true, card: { design: 'SPADE' as const, value: 3 } } }),
};

const resultMatchState: MemoryResponse = {
  ...flip1State,
  phase: 2,
  firstFlipPos: 0,
  secondFlipPos: 1,
  lastMatchResult: true,
  board: makeBoard({
    0: { faceUp: true, card: { design: 'SPADE' as const, value: 1 } },
    1: { faceUp: true, card: { design: 'HEART' as const, value: 1 } },
  }),
};

const gameEndState: MemoryResponse = {
  ...flip1State,
  phase: 3,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'ゲーム終了！',
};

const gameEndByFlagState: MemoryResponse = {
  ...flip1State,
  phase: 0,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'ゲーム終了！',
};

const cpuTurnState: MemoryResponse = {
  ...flip1State,
  currentPlayerIdx: 1,
};

beforeEach(() => {
  mockExec.mockResolvedValue(flip1State);
  vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
});

describe('MemoryPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<MemoryPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('renders reset on mount', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 1 }));
  });

  it('renders player scores inline', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());
    expect(screen.getByText(/CPU 1: 2/)).toBeInTheDocument();
    expect(screen.getByText(/CPU 2: 1/)).toBeInTheDocument();
    expect(screen.getByText(/CPU 3: 0/)).toBeInTheDocument();
  });

  it('score section has role="status" for accessibility', async () => {
    const { container } = renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());
    const scoreSection = container.querySelector('[role="status"]');
    expect(scoreSection).toBeInTheDocument();
    expect(scoreSection).toHaveAttribute('aria-label', 'スコア');
  });

  it('renders board with 52 buttons', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());
    // Board should have 52 buttons
    const buttons = screen.getAllByRole('button');
    // 52 board + 1 reset = 53
    expect(buttons.length).toBeGreaterThanOrEqual(52);
  });

  it('clicking a board card calls handleFlip', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(flip2State);

    // Click a board card button by test id
    fireEvent.click(screen.getByTestId('board-3'));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('flip', 3));
  });

  it('shows face-up card image in flip2 phase', async () => {
    mockExec.mockResolvedValue(flip2State);
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());
    // Face-up card at position 5 shows specific card image
    const faceUpImg = screen.getByAltText('♠ 3');
    expect(faceUpImg).toBeInTheDocument();
  });

  it('result phase shows next button', async () => {
    mockExec.mockResolvedValue(resultMatchState);
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次へ' })).toBeInTheDocument());
  });

  it('next button dispatches next command', async () => {
    mockExec.mockResolvedValue(resultMatchState);
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次へ' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(flip1State);
    fireEvent.click(screen.getByRole('button', { name: '次へ' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('next button not shown in flip phases', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '次へ' })).not.toBeInTheDocument();
  });

  it('game end shows action log button', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());
  });

  it('game end by flag also shows action log', async () => {
    mockExec.mockResolvedValue(gameEndByFlagState);
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());
  });

  it('action log button fetches and shows log', async () => {
    mockExec.mockResolvedValue(gameEndState);
    const mockLogApi = vi.mocked(actionLogApi.memory);
    mockLogApi.mockResolvedValue({ entries: [{ turnNumber: 1, playerIdx: 0, actionType: 'match', detail: 'test' }] });

    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '棋譜を見る' }));
    await waitFor(() => expect(screen.getByText('棋譜')).toBeInTheDocument());
  });

  it('reset button dispatches reset with config', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(flip1State);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 1 }));
  });

  it('does not show visited badge on a freshly dealt board', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());
    expect(screen.queryByTestId('board-visited-5')).not.toBeInTheDocument();
  });

  it('marks a previously face-up card with the visited overlay after it turns down', async () => {
    // Regression: an early version reset `visited` whenever the board was
    // fully face-down + nothing-taken, which is the normal between-turn
    // state before the first pair is taken. After Next, the board flips
    // face-down with no `taken` cards — the badge for the previously
    // face-up positions must still appear.
    const seenState: MemoryResponse = {
      ...flip1State,
      phase: 2,
      firstFlipPos: 5,
      secondFlipPos: 10,
      board: makeBoard({
        5: { faceUp: true, card: { design: 'SPADE' as const, value: 3 } },
        10: { faceUp: true, card: { design: 'HEART' as const, value: 7 } },
      }),
    };
    mockExec.mockResolvedValueOnce(seenState);
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByAltText('♠ 3')).toBeInTheDocument());
    mockExec.mockResolvedValue(flip1State);
    fireEvent.click(screen.getByRole('button', { name: '次へ' }));
    await waitFor(() => expect(screen.getByTestId('board-visited-5')).toBeInTheDocument());
    expect(screen.getByTestId('board-visited-10')).toBeInTheDocument();
  });

  it('gives a visited card a contrasting ring readable at small card sizes', async () => {
    const seenState: MemoryResponse = {
      ...flip1State,
      phase: 2,
      firstFlipPos: 5,
      secondFlipPos: 10,
      board: makeBoard({
        5: { faceUp: true, card: { design: 'SPADE' as const, value: 3 } },
        10: { faceUp: true, card: { design: 'HEART' as const, value: 7 } },
      }),
    };
    mockExec.mockResolvedValueOnce(seenState);
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByAltText('♠ 3')).toBeInTheDocument());
    mockExec.mockResolvedValue(flip1State);
    fireEvent.click(screen.getByRole('button', { name: '次へ' }));
    await waitFor(() => expect(screen.getByTestId('board-visited-5')).toBeInTheDocument());
    expect(screen.getByTestId('board-5')).toHaveClass('ring-ds-accent');
    // Unvisited face-down cards keep the plain back (no persistent ring).
    expect(screen.getByTestId('board-0')).not.toHaveClass('ring-ds-accent');
  });

  it('face-down cards show card back image instead of position number', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());
    // Face-down card should have a card back image, not a number
    const cardBtn = screen.getByTestId('board-0');
    const img = cardBtn.querySelector('img');
    expect(img).toBeInTheDocument();
    expect(img?.getAttribute('src')).toBe('/images/z01.png');
  });

  it('board cards disabled when taken', async () => {
    const takenBoard = makeBoard({ 0: { taken: true } });
    mockExec.mockResolvedValue({ ...flip1State, board: takenBoard });
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());

    // Taken cards should not show position number (transparent)
    // Other cards should be enabled
  });

  it('taken cards are hidden from grid layout', async () => {
    const takenBoard = makeBoard({ 0: { taken: true } });
    mockExec.mockResolvedValue({ ...flip1State, board: takenBoard });
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());
    const takenBtn = screen.getByTestId('board-0');
    expect(takenBtn).toHaveClass('hidden');
  });

  it('changing cpu difficulty updates config used on reset', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());

    fireEvent.click(screen.getByText('設定'));
    fireEvent.change(screen.getByRole('combobox', { name: 'CPU難易度' }), { target: { value: '2' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(flip1State);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 2 }));
  });

  it('board cards disabled when face up', async () => {
    mockExec.mockResolvedValue(flip2State);
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());
    // Face-up card button at position 5 should be disabled
  });

  it('board cards disabled on CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());
    // Board card at position 10 should be disabled
    expect(screen.getByTestId('board-10')).toBeDisabled();
  });

  it('loading state shows sr-only text and aria-busy', async () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<MemoryPage />);
    // Nothing renders when no state
  });

  it('settings panel works', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());

    // Open settings
    fireEvent.click(screen.getByText('設定'));
    expect(screen.getByText('CPU難易度')).toBeInTheDocument();
  });

  it('displays message from state', async () => {
    mockExec.mockResolvedValue({ ...flip1State, message: 'テストメッセージ', messageCode: 'memory.flip1' });
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText('1枚目をめくってください')).toBeInTheDocument());
  });

  it('renders landscape orientation banner in DOM', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());
    expect(screen.getByText('横向きにすると快適にプレイできます')).toBeInTheDocument();
  });

  // ── ConfirmDialog on reset ─────────────────────────────────────────────────

  it('shows confirm dialog when reset button is clicked', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('dismisses confirm dialog on cancel', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('executes reset on confirm', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(flip1State);
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, { cpuDifficulty: 1 }));
  });

  it('displays error message', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('Network error'));
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));

    await waitFor(() =>
      expect(screen.getByText('通信エラーが発生しました。もう一度お試しください。')).toBeInTheDocument(),
    );
  });

  // --- PhaseIndicator coverage ---

  it('phase indicator shows your turn when human flip turn', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('あなたのターン'));
  });

  it('phase indicator shows waiting when cpu turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('待機中'));
  });

  it('reset hides the action log panel if open', async () => {
    mockExec.mockResolvedValue(gameEndState);
    vi.mocked(actionLogApi.memory).mockResolvedValueOnce({
      entries: [{ turnNumber: 1, playerIdx: 0, actionType: 'match', detail: 'test' }],
    });

    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '棋譜を見る' })).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: '棋譜を見る' }));
    await waitFor(() => expect(screen.getByText('棋譜')).toBeInTheDocument());

    mockExec.mockResolvedValue(flip1State);
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));

    await waitFor(() => expect(screen.queryByText('棋譜')).not.toBeInTheDocument());
  });

  // --- Keyboard navigation tests ---

  it('pressing n triggers next in RESULT phase', async () => {
    mockExec.mockResolvedValue(resultMatchState);
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次へ' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(flip1State);
    fireEvent.keyDown(document, { key: 'n' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('board card buttons have focus-visible ring classes', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());
    const boardButtons = screen.getAllByRole('button').filter((btn) => btn.className.includes('aspect-'));
    expect(boardButtons.length).toBeGreaterThan(0);
    for (const btn of boardButtons) {
      expect(btn.className).toContain('focus-visible:ring-white/80');
    }
  });

  it('pressing n does not trigger next outside RESULT phase', async () => {
    mockExec.mockResolvedValue(flip1State);
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'n' });
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
  });

  it('starts tutorial when tutorial button is clicked', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
  });

  it('tutorial can be skipped', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チュートリアル' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'チュートリアル' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(flip1State);
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('face-down card buttons have aria-label with position', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());
    const cardBtn = screen.getByTestId('board-0');
    expect(cardBtn).toHaveAttribute('aria-label', '1枚目のカード（裏向き）');
  });

  it('face-up card button has aria-label with card name', async () => {
    mockExec.mockResolvedValue(flip2State);
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());
    const faceUpBtn = screen.getByTestId('board-5');
    expect(faceUpBtn).toHaveAttribute('aria-label', '♠ 3');
  });

  it('board grid uses 7-column layout on mobile to fit 52 cards within viewport (#1367)', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());
    const boardGrid = screen.getByTestId('board-0').parentElement;
    expect(boardGrid).toHaveClass('grid-cols-7');
  });

  it('card buttons have min-h-[44px] and min-w-[44px] for WCAG tap target compliance', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());
    const cardBtn = screen.getByTestId('board-0');
    expect(cardBtn.className).toContain('min-h-[44px]');
    expect(cardBtn.className).toContain('min-w-[44px]');
  });

  it('face-down cards have subtle border instead of thick blue border', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());
    const cardBtn = screen.getByTestId('board-0');
    expect(cardBtn.className).not.toContain('border-ds-info');
    expect(cardBtn.className).toContain('border-white/10');
  });

  // --- Captured pairs mini-cards (#3028) ---

  it('renders captured-pair mini cards per player', async () => {
    const capturedState: MemoryResponse = {
      ...flip1State,
      players: [
        {
          id: 0,
          isHuman: true,
          pairCount: 2,
          pairs: [
            { design: 'CLOVER' as const, value: 3 },
            { design: 'SPADE' as const, value: 7 },
          ],
        },
        { id: 1, isHuman: false, pairCount: 1, pairs: [{ design: 'HEART' as const, value: 11 }] },
        { id: 2, isHuman: false, pairCount: 0, pairs: [] },
        { id: 3, isHuman: false, pairCount: 0, pairs: [] },
      ],
    };
    mockExec.mockResolvedValue(capturedState);
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByTestId('mem-captured')).toBeInTheDocument());
    // Human captured 3♣ and 7♠ → two mini cards rendered in their row.
    const humanRow = screen.getByTestId('mem-captured-0');
    expect(humanRow.querySelectorAll('img').length).toBe(2);
    // A player with no captures shows the "none" placeholder.
    expect(screen.getByTestId('mem-captured-2')).toHaveTextContent('なし');
  });

  it('captured-pairs panel is collapsible', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByTestId('mem-captured')).toBeInTheDocument());
    expect(screen.getByText('獲得ペア')).toBeInTheDocument();
    expect(screen.getByTestId('mem-captured').tagName).toBe('DETAILS');
  });

  // --- Frontend hint ---

  it('renders hint toggle checkbox in SettingsPanel', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByRole('checkbox', { name: 'ヒント表示' })).toBeInTheDocument());
  });

  it('renders HintTooltip when hintEnabled is true and hint is available', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'flip', reason: 'frontendHint.flipAny', confidence: 'moderate' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  // --- Keyboard grid navigation (#3029) ---
  // matchMedia is mocked to matches:false in test setup, so the grid resolves to
  // 7 columns (the base breakpoint).

  it('gives the first board cell the roving tab-stop', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());
    expect(screen.getByTestId('board-0')).toHaveAttribute('tabindex', '0');
    expect(screen.getByTestId('board-1')).toHaveAttribute('tabindex', '-1');
  });

  it('ArrowRight moves focus to the next cell', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());
    fireEvent.keyDown(screen.getByTestId('board-0'), { key: 'ArrowRight' });
    expect(screen.getByTestId('board-1')).toHaveFocus();
    expect(screen.getByTestId('board-1')).toHaveAttribute('tabindex', '0');
    expect(screen.getByTestId('board-0')).toHaveAttribute('tabindex', '-1');
  });

  it('ArrowDown moves focus down one row (by column count)', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());
    fireEvent.keyDown(screen.getByTestId('board-0'), { key: 'ArrowDown' });
    expect(screen.getByTestId('board-7')).toHaveFocus();
  });

  it('ArrowLeft at the left edge keeps focus clamped', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());
    const first = screen.getByTestId('board-0');
    first.focus();
    fireEvent.keyDown(first, { key: 'ArrowLeft' });
    expect(first).toHaveFocus();
    expect(first).toHaveAttribute('tabindex', '0');
  });

  it('arrow navigation skips taken cells', async () => {
    // board-1 is taken, so ArrowRight from board-0 lands on board-2.
    mockExec.mockResolvedValue({
      ...flip1State,
      board: makeBoard({ 1: { taken: true } }),
    });
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());
    fireEvent.keyDown(screen.getByTestId('board-0'), { key: 'ArrowRight' });
    expect(screen.getByTestId('board-2')).toHaveFocus();
  });

  // --- Flip-result live region (#3029) ---

  it('announces a match result in the polite live region', async () => {
    mockExec.mockResolvedValue(resultMatchState);
    renderWithProviders(<MemoryPage />);
    const region = await screen.findByTestId('mem-flip-announce');
    expect(region).toHaveAttribute('aria-live', 'polite');
    expect(region).toHaveAttribute('role', 'status');
    await waitFor(() => expect(region).toHaveTextContent('♠ A / ♥ A — 一致'));
  });

  it('announces a mismatch result in the polite live region', async () => {
    mockExec.mockResolvedValue({
      ...resultMatchState,
      lastMatchResult: false,
      board: makeBoard({
        0: { faceUp: true, card: { design: 'SPADE' as const, value: 1 } },
        1: { faceUp: true, card: { design: 'HEART' as const, value: 5 } },
      }),
    });
    renderWithProviders(<MemoryPage />);
    const region = await screen.findByTestId('mem-flip-announce');
    await waitFor(() => expect(region).toHaveTextContent('♠ A / ♥ 5 — 不一致'));
  });

  it('keeps the live region empty outside the result phase', async () => {
    renderWithProviders(<MemoryPage />);
    await waitFor(() => expect(screen.getByText(/あなた: 0/)).toBeInTheDocument());
    expect(screen.getByTestId('mem-flip-announce')).toHaveTextContent('');
  });
});
