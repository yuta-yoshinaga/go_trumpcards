import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { curdsandwheyApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CurdsAndWheyResponse } from '../types/card';
import { CurdsAndWheyPage } from './CurdsAndWheyPage';

vi.mock('../api/gameApi', () => ({
  curdsandwheyApi: { exec: vi.fn() },
  actionLogApi: { curdsandwhey: vi.fn() },
}));

const mockExec = vi.mocked(curdsandwheyApi.exec);

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<CurdsAndWheyResponse> = {}): CurdsAndWheyResponse {
  // 13 columns: what the backend deals. The Curds and Whey fixtures this was
  // cloned from built 10, which gave no coverage of the real board width.
  const columns: Card[][] = Array.from({ length: 13 }, () => []);
  columns[0] = [card('SPADE', 9)];
  columns[1] = [card('SPADE', 8)];
  return {
    columns,
    completedSuits: 0,
    phase: 0,
    moveCount: 0,
    canUndo: false,
    message: '',
    ...overrides,
  };
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeState());
});

describe('CurdsAndWheyPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<CurdsAndWheyPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<CurdsAndWheyPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders all 13 columns', async () => {
    renderWithProviders(<CurdsAndWheyPage />);
    await waitFor(() => expect(screen.getByTestId('column-0')).toBeInTheDocument());
    expect(screen.getByTestId('column-9')).toBeInTheDocument();
  });

  it('selects a card then moves the run to another column', async () => {
    renderWithProviders(<CurdsAndWheyPage />);
    const srcCard = await screen.findByTestId('card-1-0');
    fireEvent.click(srcCard);
    mockExec.mockClear();
    // Click a card in column 0 → move col1 onto col0.
    fireEvent.click(screen.getByTestId('card-0-0'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('m', { fromCol: 1, cardIndex: 0, toCol: 0 }));
  });

  it('labels cards with name+position and reflects selection with aria-pressed', async () => {
    renderWithProviders(<CurdsAndWheyPage />);
    // column[1] is ♠8 at the top → "♠ 8（列2・上から1枚目）".
    const src = await screen.findByTestId('card-1-0');
    expect(src).toHaveAttribute('aria-label', '♠ 8（列2・上から1枚目）');
    expect(src).toHaveAttribute('aria-pressed', 'false');
    fireEvent.click(src);
    expect(screen.getByTestId('card-1-0')).toHaveAttribute('aria-pressed', 'true');
    // Empty column drop targets are named too.
    expect(screen.getByTestId('column-2-drop')).toHaveAttribute('aria-label', '列3（空）');
    // The selection guidance is a live region.
    expect(screen.getByTestId('cw-guidance')).toHaveAttribute('role', 'status');
  });

  it('moves onto an empty column', async () => {
    renderWithProviders(<CurdsAndWheyPage />);
    const srcCard = await screen.findByTestId('card-0-0');
    fireEvent.click(srcCard);
    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('column-5-drop'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('m', { fromCol: 0, cardIndex: 0, toCol: 5 }));
  });

  it('undoes, hints', async () => {
    mockExec.mockResolvedValue(makeState({ canUndo: true }));
    renderWithProviders(<CurdsAndWheyPage />);
    await screen.findByTestId('hint-button');
    for (const [testid, cmd] of [
      ['undo-button', 'u'],
      ['hint-button', 'hint'],
    ] as const) {
      mockExec.mockClear();
      fireEvent.click(screen.getByTestId(testid));
      await waitFor(() => expect(mockExec).toHaveBeenCalledWith(cmd));
    }
  });

  it('marks only the movable-run cards as grabbable and blocks invalid source selection', async () => {
    // Column 3: ♠9 (not a run head), then ♥6 ♥5 ♥4 (a same-suit descending run).
    mockExec.mockResolvedValue(
      makeState({
        columns: (() => {
          const cols: Card[][] = Array.from({ length: 13 }, () => []);
          cols[0] = [card('SPADE', 5)];
          cols[3] = [card('SPADE', 9), card('HEART', 6), card('HEART', 5), card('HEART', 4)];
          return cols;
        })(),
      }),
    );
    renderWithProviders(<CurdsAndWheyPage />);
    const top = await screen.findByTestId('card-3-0');
    // The out-of-run top card is not grabbable and is disabled.
    expect(top).toHaveAttribute('data-grabbable', 'false');
    expect(top).toBeDisabled();
    // The run head and its tail are grabbable.
    expect(screen.getByTestId('card-3-1')).toHaveAttribute('data-grabbable', 'true');
    expect(screen.getByTestId('card-3-3')).toHaveAttribute('data-grabbable', 'true');

    // Clicking the invalid top card selects nothing (no aria-pressed run).
    fireEvent.click(top);
    expect(screen.getByTestId('card-3-1')).toHaveAttribute('aria-pressed', 'false');

    // Clicking the run head selects it and its descendants.
    fireEvent.click(screen.getByTestId('card-3-1'));
    expect(screen.getByTestId('card-3-1')).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByTestId('card-3-3')).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByTestId('card-3-0')).toHaveAttribute('aria-pressed', 'false');
  });

  it('double-clicks a grabbable card to auto-move it to the best column', async () => {
    // Default state: col0 = ♠9, col1 = ♠8. Double-clicking ♠8 links same-suit onto ♠9.
    renderWithProviders(<CurdsAndWheyPage />);
    const srcCard = await screen.findByTestId('card-1-0');
    mockExec.mockClear();
    fireEvent.doubleClick(srcCard);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('m', { fromCol: 1, cardIndex: 0, toCol: 0 }));
  });

  it('shows a notice and issues no move when a double-click has no destination', async () => {
    // A lone ♠2 has no valid destination (no rank-3 top; empty columns are not
    // offered for a whole-column move).
    mockExec.mockResolvedValue(
      makeState({
        columns: (() => {
          const cols: Card[][] = Array.from({ length: 13 }, () => []);
          cols[0] = [card('SPADE', 2)];
          return cols;
        })(),
      }),
    );
    renderWithProviders(<CurdsAndWheyPage />);
    const srcCard = await screen.findByTestId('card-0-0');
    mockExec.mockClear();
    fireEvent.doubleClick(srcCard);
    await waitFor(() => expect(screen.getByTestId('cw-automove-notice')).toBeInTheDocument());
    expect(mockExec).not.toHaveBeenCalledWith('m', expect.anything());
  });

  it('hides controls at game over', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 2 }));
    renderWithProviders(<CurdsAndWheyPage />);
    await waitFor(() => expect(screen.getByTestId('column-0')).toBeInTheDocument());
    expect(screen.queryByTestId('hint-button')).not.toBeInTheDocument();
  });

  // **ギブアップは取り消せない** (#6475)。リセットには確認が挟まるのに、
  // ここは即座に対局を打ち切っていた。
  it('asks before giving up, and only then dispatches', async () => {
    renderWithProviders(<CurdsAndWheyPage />);
    await screen.findByTestId('giveup-button');

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('giveup-button'));
    await waitFor(() => expect(screen.getByText('投了確認')).toBeInTheDocument());
    expect(mockExec).not.toHaveBeenCalledWith('g');

    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('g'));
  });

  it('leaves the game untouched when the give-up dialog is cancelled', async () => {
    renderWithProviders(<CurdsAndWheyPage />);
    await screen.findByTestId('giveup-button');

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('giveup-button'));
    await waitFor(() => expect(screen.getByText('投了確認')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'キャンセル' }));

    await waitFor(() => expect(screen.queryByText('投了確認')).not.toBeInTheDocument());
    expect(mockExec).not.toHaveBeenCalled();
  });
});

// Keyboard shortcuts are bound by useActionKeyboardNav and advertised by
// ActionShortcutsPanel; assert the keys actually run their action (#4429, #6591).
describe('CurdsAndWheyPage keyboard shortcuts', () => {
  it('pressing h dispatches hint', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<CurdsAndWheyPage />);
    await waitFor(() => expect(screen.getByTestId('column-0')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(makeState());
    fireEvent.keyDown(document, { key: 'h' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('pressing z dispatches u', async () => {
    mockExec.mockResolvedValue(makeState({ canUndo: true }));
    renderWithProviders(<CurdsAndWheyPage />);
    await waitFor(() => expect(screen.getByTestId('column-0')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(makeState());
    fireEvent.keyDown(document, { key: 'z' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('u'));
  });

  it('pressing g asks for give-up confirmation rather than firing it', async () => {
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<CurdsAndWheyPage />);
    await waitFor(() => expect(screen.getByTestId('column-0')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'g' });
    expect(await screen.findByText('投了確認')).toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('ignores shortcuts when loading', async () => {
    let resolveAction!: (val: CurdsAndWheyResponse) => void;
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<CurdsAndWheyPage />);
    await waitFor(() => expect(screen.getByTestId('column-0')).toBeInTheDocument());

    mockExec.mockImplementation(
      () =>
        new Promise((resolve) => {
          resolveAction = resolve;
        }),
    );
    fireEvent.click(screen.getByTestId('hint-button'));
    await waitFor(() => expect(screen.getByTestId('hint-button')).toBeDisabled());

    const callCount = mockExec.mock.calls.length;
    fireEvent.keyDown(document, { key: 'h' });
    await flushPendingDispatch();
    expect(mockExec).toHaveBeenCalledTimes(callCount);

    resolveAction(makeState());
  });

  it('ignores shortcuts once the game has ended', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 2 }));
    renderWithProviders(<CurdsAndWheyPage />);
    await waitFor(() => expect(screen.getByTestId('column-0')).toBeInTheDocument());
    mockExec.mockClear();
    for (const key of ['h', 'z', 'g']) {
      fireEvent.keyDown(document, { key });
    }
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
    expect(screen.queryByText('投了確認')).not.toBeInTheDocument();
  });

  it('renders ActionShortcutsPanel', async () => {
    renderWithProviders(<CurdsAndWheyPage />);
    await waitFor(() => expect(screen.getByTestId('cw-kbd-shortcuts')).toBeInTheDocument());
  });
});
