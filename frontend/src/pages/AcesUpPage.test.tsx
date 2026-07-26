import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { acesupApi, actionLogApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { AcesUpCard, AcesUpResponse, Card, CardDesign } from '../types/card';
import { AcesUpPage } from './AcesUpPage';

vi.mock('../api/gameApi', () => ({
  acesupApi: { exec: vi.fn() },
  actionLogApi: { acesup: vi.fn() },
}));

const mockExec = vi.mocked(acesupApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function makeCard(c: Card, opts: Partial<Omit<AcesUpCard, 'card'>> = {}): AcesUpCard {
  return { card: c, top: opts.top ?? false, removable: opts.removable ?? false, movable: opts.movable ?? false };
}

/** col0: removable 5♠, col1: movable 9♠, col2: empty, col3: 6♦ */
function makeColumns(): AcesUpCard[][] {
  return [
    [makeCard(card('SPADE', 5), { top: true, removable: true })],
    [makeCard(card('SPADE', 9), { top: true, movable: true })],
    [],
    [makeCard(card('DIAMOND', 6), { top: true })],
  ];
}

const playingState: AcesUpResponse = {
  columns: makeColumns(),
  stockCount: 44,
  discardCount: 4,
  discardTop: card('CLOVER', 7),
  phase: 0,
  moveCount: 3,
  canUndo: true,
  isStalemate: false,
  message: '',
};

const gameClearState: AcesUpResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'acesup.gameClear',
  messageParams: { moveCount: '20' },
};

const gameOverState: AcesUpResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'acesup.gameOver',
};

beforeEach(() => {
  mockExec.mockResolvedValue(playingState);
});

describe('AcesUpPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<AcesUpPage />);
    const pulseElements = document.querySelectorAll('.animate-pulse');
    expect(pulseElements.length).toBeGreaterThan(0);
  });

  it('renders stock count', async () => {
    renderWithProviders(<AcesUpPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());
    expect(screen.getByText(/\(44\)/)).toBeInTheDocument();
  });

  it('renders move count', async () => {
    renderWithProviders(<AcesUpPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent(/手数: 3/));
  });

  it('renders the four columns', async () => {
    renderWithProviders(<AcesUpPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());
    const colDivs = document.querySelectorAll('[data-tutorial="acesup-columns"] > div');
    expect(colDivs.length).toBe(4);
  });

  it('renders the discard pile with progress and the top card', async () => {
    renderWithProviders(<AcesUpPage />);
    await waitFor(() => expect(screen.getByTestId('acesup-discard-pile')).toBeInTheDocument());
    // Progress readout: discardCount out of the 48-card goal.
    expect(screen.getByText(/捨て札/)).toBeInTheDocument();
    expect(screen.getByText(/\(4\/48\)/)).toBeInTheDocument();
    // The most recently removed card is shown face-up.
    expect(screen.getByTestId('acesup-discard-top')).toBeInTheDocument();
    expect(screen.getByRole('img', { name: /♣ 7/ })).toBeInTheDocument();
  });

  it('renders an empty discard placeholder when nothing has been removed', async () => {
    mockExec.mockResolvedValue({ ...playingState, discardCount: 0, discardTop: null });
    renderWithProviders(<AcesUpPage />);
    await waitFor(() => expect(screen.getByTestId('acesup-discard-pile')).toBeInTheDocument());
    expect(screen.getByTestId('acesup-discard-empty')).toBeInTheDocument();
    expect(screen.getByText(/\(0\/48\)/)).toBeInTheDocument();
    expect(screen.queryByTestId('acesup-discard-top')).not.toBeInTheDocument();
  });

  it('renders empty column placeholder', async () => {
    renderWithProviders(<AcesUpPage />);
    await waitFor(() => expect(screen.getAllByText('空').length).toBeGreaterThanOrEqual(1));
  });

  it('renders empty stock placeholder', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0 });
    renderWithProviders(<AcesUpPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());
    expect(screen.getAllByText('空').length).toBeGreaterThanOrEqual(1);
  });

  it('clicking deal button dispatches draw', async () => {
    renderWithProviders(<AcesUpPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    // Both the stock card-back and the footer button expose the "配る" label.
    const dealButtons = screen.getAllByRole('button', { name: '配る' });
    fireEvent.click(dealButtons[dealButtons.length - 1]);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('clicking a removable top card dispatches remove', async () => {
    renderWithProviders(<AcesUpPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    const cardButtons = screen.getAllByRole('button', { name: /♠ 5/ });
    fireEvent.click(cardButtons[0]);

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('remove', 0));
  });

  it('clicking discard-all dispatches remove for every removable card', async () => {
    renderWithProviders(<AcesUpPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());

    mockExec.mockClear();
    // After the first removal nothing is removable, so the batch loop stops.
    const cleared: AcesUpResponse = { ...playingState, columns: [[], [], [], []] };
    mockExec.mockResolvedValueOnce(cleared);

    fireEvent.click(screen.getByTestId('acesup-remove-all'));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('remove', 0));
    expect(mockExec).toHaveBeenCalledTimes(1);
  });

  it('disables discard-all when no card is removable', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      columns: [
        [makeCard(card('SPADE', 5), { top: true })],
        [makeCard(card('SPADE', 9), { top: true, movable: true })],
        [],
        [makeCard(card('DIAMOND', 6), { top: true })],
      ],
    });
    renderWithProviders(<AcesUpPage />);
    await waitFor(() => expect(screen.getByTestId('acesup-remove-all')).toBeInTheDocument());
    expect(screen.getByTestId('acesup-remove-all')).toBeDisabled();
  });

  it('clicking a move button dispatches move', async () => {
    renderWithProviders(<AcesUpPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: '移動 [1]' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', 1));
  });

  it('dragging a movable top card onto an empty column dispatches move', async () => {
    const buildDataTransfer = () => {
      const store: Record<string, string> = {};
      return {
        setData: (type: string, val: string) => {
          store[type] = val;
        },
        getData: (type: string) => store[type] ?? '',
        effectAllowed: '',
        dropEffect: '',
      };
    };

    renderWithProviders(<AcesUpPage />);
    await waitFor(() => expect(screen.getByTestId('acesup-empty-2')).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);

    // Drag col1's movable 9♠ (the only movable top card) onto the empty col2.
    const dataTransfer = buildDataTransfer();
    fireEvent.dragStart(screen.getByRole('button', { name: '♠ 9' }), { dataTransfer });
    const dropZone = screen.getByTestId('acesup-empty-2');
    fireEvent.dragOver(dropZone, { dataTransfer });
    fireEvent.drop(dropZone, { dataTransfer });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('move', 1));
  });

  it('clicking giveup button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    renderWithProviders(<AcesUpPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(gameOverState);
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('clicking undo button dispatches undo', async () => {
    renderWithProviders(<AcesUpPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.click(screen.getByRole('button', { name: '元に戻す' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('undo'));
  });

  it('clicking hint button dispatches hint', async () => {
    renderWithProviders(<AcesUpPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());

    mockExec.mockResolvedValue({ ...playingState, hint: { type: 'remove', col: 0 } });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
    // The hint renders via HintTooltip with a 0-based column reason matching the move buttons.
    await waitFor(() => expect(screen.getByText(/列\[0\]の一番上の札を除去/)).toBeInTheDocument());
  });

  it('renders a move hint via HintTooltip with the column', async () => {
    renderWithProviders(<AcesUpPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());
    mockExec.mockResolvedValue({ ...playingState, hint: { type: 'move', col: 2 } });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getByText(/列\[2\]の札を空き列へ移動/)).toBeInTheDocument());
  });

  it('renders a draw hint via HintTooltip without a column', async () => {
    renderWithProviders(<AcesUpPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());
    mockExec.mockResolvedValue({ ...playingState, hint: { type: 'draw', col: -1 } });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getByText(/山札から各列へ1枚ずつ/)).toBeInTheDocument());
  });

  it('renders game clear state', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<AcesUpPage />);
    await waitFor(() => expect(screen.getByText('ゲームクリア')).toBeInTheDocument());
  });

  it('renders game over state and hides action buttons', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<AcesUpPage />);
    await waitFor(() => expect(screen.getAllByText('ゲームオーバー').length).toBeGreaterThanOrEqual(1));
    expect(screen.queryByRole('button', { name: '配る' })).not.toBeInTheDocument();
  });

  it('disables undo button when canUndo is false', async () => {
    mockExec.mockResolvedValue({ ...playingState, canUndo: false });
    renderWithProviders(<AcesUpPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '元に戻す' })).toBeDisabled();
  });

  it('disables deal button when stock empty', async () => {
    mockExec.mockResolvedValue({ ...playingState, stockCount: 0 });
    renderWithProviders(<AcesUpPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '配る' })).toBeDisabled();
  });

  it('renders stalemate message', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      isStalemate: true,
      undoToEscape: 1,
      messageCode: 'acesup.stalemate',
      message: '手詰まりです。',
    });
    renderWithProviders(<AcesUpPage />);
    await waitFor(() => expect(screen.getByText('手詰まりです。')).toBeInTheDocument());
  });

  it('suppresses unused import warning', () => {
    expect(actionLogApi).toBeDefined();
  });
});
