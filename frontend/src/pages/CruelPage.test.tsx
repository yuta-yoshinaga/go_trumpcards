import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { cruelApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, CruelResponse } from '../types/card';
import { CruelPage } from './CruelPage';

vi.mock('../api/gameApi', () => ({
  cruelApi: { exec: vi.fn() },
  actionLogApi: { cruel: vi.fn() },
}));

const mockExec = vi.mocked(cruelApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

const fourColumn = (suit: CardDesign, base: number) => [
  { card: card(suit, base), faceUp: true },
  { card: card(suit, base + 1), faceUp: true },
  { card: card(suit, base + 2), faceUp: true },
  { card: card(suit, base + 3), faceUp: true },
];

const playingState: CruelResponse = {
  tableau: [
    fourColumn('SPADE', 2),
    fourColumn('HEART', 2),
    fourColumn('CLOVER', 2),
    fourColumn('DIAMOND', 2),
    fourColumn('SPADE', 6),
    fourColumn('HEART', 6),
    fourColumn('CLOVER', 6),
    fourColumn('DIAMOND', 6),
    fourColumn('SPADE', 10),
    fourColumn('HEART', 10),
    fourColumn('CLOVER', 10),
    fourColumn('DIAMOND', 10),
  ],
  foundation: [[card('SPADE', 1)], [card('CLOVER', 1)], [card('HEART', 1)], [card('DIAMOND', 1)]],
  phase: 0,
  moveCount: 0,
  canUndo: false,
  isStalemate: false,
  message: '',
  messageCode: 'cruel.playing',
};

const gameClearState: CruelResponse = {
  ...playingState,
  phase: 1,
  moveCount: 42,
  messageCode: 'cruel.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: CruelResponse = {
  ...playingState,
  phase: 2,
  messageCode: 'cruel.gameOver',
};

beforeEach(() => {
  vi.clearAllMocks();
  mockExec.mockResolvedValue(playingState);
});

describe('CruelPage', () => {
  it('renders heading', async () => {
    renderWithProviders(<CruelPage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<CruelPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows move count', async () => {
    renderWithProviders(<CruelPage />);
    await waitFor(() => expect(screen.getByText(/手数/)).toBeInTheDocument());
  });

  it('shows game clear phase', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<CruelPage />);
    await waitFor(() => expect(screen.getAllByText('ゲームクリア').length).toBeGreaterThan(0));
  });

  it('shows game over phase', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<CruelPage />);
    await waitFor(() => expect(screen.getAllByText('ゲームオーバー').length).toBeGreaterThan(0));
  });

  it('shift button fires shift command', async () => {
    renderWithProviders(<CruelPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    screen.getByTestId('shift-button').click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('shift'));
  });

  it('pulses the shift button and shows a banner on stalemate', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true });
    renderWithProviders(<CruelPage />);
    const shiftBtn = await screen.findByTestId('shift-button');
    expect(shiftBtn.className).toContain('animate-pulse');
    expect(shiftBtn).toHaveAttribute('aria-label', '手詰まりです。シフトで再構築できます');
    expect(screen.getByTestId('cruel-stalemate-banner')).toBeInTheDocument();
  });

  it('does not pulse the shift button when not stalemate', async () => {
    renderWithProviders(<CruelPage />);
    const shiftBtn = await screen.findByTestId('shift-button');
    expect(shiftBtn.className).not.toContain('animate-pulse');
    expect(screen.queryByTestId('cruel-stalemate-banner')).not.toBeInTheDocument();
  });

  it('hint button fires hint command', async () => {
    renderWithProviders(<CruelPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const btn = screen.getByRole('button', { name: 'ヒント' });
    btn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('shows the hint source and destination columns (tableau → foundation)', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      hint: { fromCol: 3, cardIndex: 0, toZone: 'foundation', toCol: -1 },
      messageCode: 'cruel.hintAvailable',
    });
    renderWithProviders(<CruelPage />);
    const hint = await screen.findByTestId('cruel-hint');
    expect(hint).toHaveTextContent('場札 3 → 組札');
  });

  it('shows the hint source and destination columns (tableau → tableau)', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      hint: { fromCol: 3, cardIndex: 0, toZone: 'tableau', toCol: 5 },
      messageCode: 'cruel.hintAvailable',
    });
    renderWithProviders(<CruelPage />);
    const hint = await screen.findByTestId('cruel-hint');
    expect(hint).toHaveTextContent('場札 3 → 場札 5');
  });

  it('autocomplete button fires autocomplete command', async () => {
    renderWithProviders(<CruelPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    screen.getByTestId('autocomplete-button').click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('autocomplete'));
  });

  it('giveup button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    renderWithProviders(<CruelPage />);
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

  it('shows error alert when API fails on mount', async () => {
    mockExec.mockRejectedValue(new Error('network error'));
    renderWithProviders(<CruelPage />);
    await waitFor(() => expect(screen.getByRole('alert')).toBeInTheDocument());
  });

  it('hides game action buttons after game over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<CruelPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'ヒント' })).not.toBeInTheDocument();
    expect(screen.queryByTestId('shift-button')).not.toBeInTheDocument();
  });

  it('undo button fires undo command when canUndo is true', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: true });
    renderWithProviders(<CruelPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const undoBtn = screen.getByRole('button', { name: '元に戻す' });
    undoBtn.click();
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('undo button is disabled when canUndo is false', async () => {
    renderWithProviders(<CruelPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    const undoBtn = screen.getByRole('button', { name: '元に戻す' });
    expect(undoBtn).toBeDisabled();
  });

  it('renders empty tableau column placeholder', async () => {
    const stateWithEmpty = {
      ...playingState,
      tableau: [[], ...playingState.tableau.slice(1)],
    };
    mockExec.mockResolvedValue(stateWithEmpty);
    renderWithProviders(<CruelPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByText('空')).toBeInTheDocument();
  });

  it('highlights only the suit-matching foundation when a card is click-selected (#3040)', async () => {
    renderWithProviders(<CruelPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    // Column 0 is fourColumn('SPADE', 2) → its top card is ♠5. Selecting it should
    // mark only foundation pile 0 (♠) as the suit target, leaving ♣/♥/♦ unmarked.
    fireEvent.click(screen.getByRole('button', { name: '♠ 5' }));

    await waitFor(() => expect(screen.getByTestId('cruel-foundation-0')).toHaveAttribute('data-suit-target', 'true'));
    expect(screen.getByTestId('cruel-foundation-1')).not.toHaveAttribute('data-suit-target');
    expect(screen.getByTestId('cruel-foundation-2')).not.toHaveAttribute('data-suit-target');
    expect(screen.getByTestId('cruel-foundation-3')).not.toHaveAttribute('data-suit-target');
    // The matching pile gets the design-token ring; non-matching piles do not.
    expect(screen.getByTestId('cruel-foundation-0').className).toContain('ring-ds-info');
    expect(screen.getByTestId('cruel-foundation-3').className).not.toContain('ring-ds-info');
  });

  it('highlights the ♦ foundation when a diamond card is selected (#3040)', async () => {
    renderWithProviders(<CruelPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    // Column 3 is fourColumn('DIAMOND', 2) → top card ♦5 maps to foundation pile 3 (♦).
    fireEvent.click(screen.getByRole('button', { name: '♦ 5' }));

    await waitFor(() => expect(screen.getByTestId('cruel-foundation-3')).toHaveAttribute('data-suit-target', 'true'));
    expect(screen.getByTestId('cruel-foundation-0')).not.toHaveAttribute('data-suit-target');
  });

  it('marks no foundation as a suit target when nothing is selected (#3040)', async () => {
    renderWithProviders(<CruelPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    for (let i = 0; i < 4; i += 1) {
      expect(screen.getByTestId(`cruel-foundation-${i}`)).not.toHaveAttribute('data-suit-target');
    }
  });

  it('renders foundation suit labels above placed Aces', async () => {
    const stateWithEmptyFoundation: CruelResponse = {
      ...playingState,
      foundation: [[], [], [], []],
    };
    mockExec.mockResolvedValue(stateWithEmptyFoundation);
    renderWithProviders(<CruelPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
    expect(screen.getByText('♠')).toBeInTheDocument();
    expect(screen.getByText('♣')).toBeInTheDocument();
    expect(screen.getByText('♥')).toBeInTheDocument();
    expect(screen.getByText('♦')).toBeInTheDocument();
  });
});

// Keyboard shortcuts are bound by useActionKeyboardNav and advertised by
// ActionShortcutsPanel, but nothing asserted that pressing a key actually runs
// its action — a wrong `key` or a wrong `enabled` condition would have failed no
// test. See issue #4429.
describe('CruelPage keyboard shortcuts', () => {
  it.each([
    ['h', 'hint'],
    ['a', 'autocomplete'],
    ['z', 'undo'],
    ['s', 'shift'],
  ])('pressing %s dispatches %s', async (key, command) => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<CruelPage />);
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
    renderWithProviders(<CruelPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'g' });
    expect(await screen.findByText('投了確認')).toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('ignores shortcuts once the game has ended', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<CruelPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    for (const key of ['h', 'a', 'z', 's']) {
      fireEvent.keyDown(document, { key });
    }
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });
});

// **52枚すべてを組札に収めることが唯一の勝利条件 (#4779)。**CUI は
// foundationProgress で合計を常に出しているのに、Web は各組札を個別に描くだけで、
// あと何枚かを知るには目で数えるしかなかった。
describe('CruelPage foundation progress', () => {
  it('shows the total across all four foundations', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      foundation: [
        [
          { design: 'SPADE', value: 1 },
          { design: 'SPADE', value: 2 },
        ],
        [{ design: 'HEART', value: 1 }],
        [],
        [],
      ],
    });
    renderWithProviders(<CruelPage />);
    expect(await screen.findByTestId('cruel-foundation-progress')).toHaveTextContent('3/52');
  });

  // **1山だけ数えない。**合計でないと「あと何枚か」の見通しにならない。
  it('sums every pile rather than only the first', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      foundation: [[], [], [], [{ design: 'DIAMOND', value: 1 }]],
    });
    renderWithProviders(<CruelPage />);
    expect(await screen.findByTestId('cruel-foundation-progress')).toHaveTextContent('1/52');
  });

  it('starts at zero on an empty foundation', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      foundation: [[], [], [], []],
    });
    renderWithProviders(<CruelPage />);
    expect(await screen.findByTestId('cruel-foundation-progress')).toHaveTextContent('0/52');
  });
});
