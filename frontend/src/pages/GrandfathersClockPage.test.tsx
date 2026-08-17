import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { grandfathersClockApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, GrandfathersClockResponse, GrandfathersClockTableauCard } from '../types/card';
import { GrandfathersClockPage } from './GrandfathersClockPage';

vi.mock('../api/gameApi', () => ({
  grandfathersClockApi: { exec: vi.fn() },
  actionLogApi: { grandfathersclock: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(grandfathersClockApi.exec);

const card = (design: CardDesign, value: number): Card => ({ design, value });

function makeTableau(cols: GrandfathersClockTableauCard[][]): GrandfathersClockTableauCard[][] {
  return Array.from({ length: 8 }, (_, i) => cols[i] ?? []);
}

// The real deal seeds every face; index i wants rank i+1.
const faces = Array.from({ length: 12 }, (_, i) => ({
  cards: [card('HEART', i + 2)],
  targetRank: i + 1,
  complete: false,
}));

const playingState: GrandfathersClockResponse = {
  tableau: makeTableau([
    [
      { card: card('SPADE', 9), faceUp: true },
      { card: card('SPADE', 6), faceUp: true },
    ],
    [{ card: card('HEART', 4), faceUp: true }],
  ]),
  foundation: faces,
  phase: 0,
  moveCount: 3,
  canUndo: false,
  isStalemate: false,
  message: '',
};

const gameClearState: GrandfathersClockResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'grandfathersclock.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: GrandfathersClockResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'grandfathersclock.gameOver',
};

describe('GrandfathersClockPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  });

  it('calls reset on initial render', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<GrandfathersClockPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
  });

  it('renders heading and move count', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<GrandfathersClockPage />);
    await waitFor(() => expect(screen.getByText(/グランドファーザーズ・クロック/)).toBeInTheDocument());
    expect(screen.getByText(/手数: 3/)).toBeInTheDocument();
  });

  it('renders all twelve clock faces', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<GrandfathersClockPage />);
    await waitFor(() => expect(screen.getAllByLabelText(/文字盤\d+ \(\d+時\)/).length).toBe(12));
  });

  // The target rank is what the player plans against, so it has to be on screen
  // rather than implied by the clock position.
  it('shows each face target: face 0 wants an Ace, face 11 a Queen', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<GrandfathersClockPage />);
    await waitFor(() => expect(screen.getByLabelText(/文字盤0 \(1時\) 目標1/)).toBeInTheDocument());
    expect(screen.getByLabelText(/文字盤11 \(12時\) 目標12/)).toBeInTheDocument();
  });

  it('labels all eight tableau columns with their 0-based index', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<GrandfathersClockPage />);
    await waitFor(() => expect(screen.getByText('#0')).toBeInTheDocument());
    for (let i = 0; i < 8; i++) {
      expect(screen.getByText(`#${i}`)).toBeInTheDocument();
    }
  });

  it('shows how many faces are done', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      foundation: faces.map((f, i) => ({ ...f, complete: i < 3 })),
    });
    renderWithProviders(<GrandfathersClockPage />);
    await waitFor(() => expect(screen.getByTestId('gc-face-progress')).toHaveTextContent('3/12'));
  });

  // A finished face accepts nothing more, so it must not be a click target.
  //
  // #5555: ただし native `disabled` はアクセシビリティツリーからボタンごと
  // 外すので、目標ランクと枚数を含む faceAriaLabel が読み上げられなくなる。
  // フォーカスは残したまま `aria-disabled` で拒否する。
  it('marks a completed face aria-disabled but keeps it focusable', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      foundation: faces.map((f, i) => ({ ...f, complete: i === 0 })),
    });
    renderWithProviders(<GrandfathersClockPage />);
    // 移動元を選ぶまでは「選択待ち」で全部の文字盤が native disabled。
    // 完成した文字盤だけが別の理由で拒否されることを見たいので、まず選ぶ。
    const source = await screen.findByRole('button', { name: /^♠ 6（/ });
    fireEvent.click(source);
    await waitFor(() => expect(source).toHaveAttribute('aria-pressed', 'true'));

    const face = screen.getByLabelText(/文字盤0 \(1時\)/);
    expect(face).toHaveAttribute('aria-disabled', 'true');
    expect(face).not.toBeDisabled();
    // ラベルは読み上げ可能なまま (目標ランクと枚数を含む)。
    expect(face.getAttribute('aria-label')).toMatch(/1時/);
  });

  // **拒否は維持する。**フォーカスできることと、動かせることは別。
  it('does not move a card onto a completed face', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      foundation: faces.map((f, i) => ({ ...f, complete: i === 0 })),
    });
    renderWithProviders(<GrandfathersClockPage />);
    const source = await screen.findByRole('button', { name: /^♠ 6（/ });
    fireEvent.click(source);
    await waitFor(() => expect(source).toHaveAttribute('aria-pressed', 'true'));
    mockExec.mockClear();

    fireEvent.click(screen.getByLabelText(/文字盤0 \(1時\)/));
    // ヒントなど別の呼び出しが挟まっても、move だけは飛ばない。
    fireEvent.click(screen.getByLabelText(/文字盤4 \(5時\)/));
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), { zone: 'foundation', col: 0 });
  });

  it('sends the face index with the move', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<GrandfathersClockPage />);
    const source = await screen.findByRole('button', { name: /^♠ 6（/ });
    fireEvent.click(source);
    await waitFor(() => expect(source).toHaveAttribute('aria-pressed', 'true'));
    mockExec.mockClear();

    fireEvent.click(screen.getByLabelText(/文字盤4 \(5時\)/));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('move', { zone: 'tableau', col: 0 }, { zone: 'foundation', col: 4 }),
    );
  });

  it('renders giveup button when playing and hides it once cleared', async () => {
    mockExec.mockResolvedValue(playingState);
    const { unmount } = renderWithProviders(<GrandfathersClockPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
    unmount();

    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<GrandfathersClockPage />);
    await waitFor(() => expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument());
  });

  it('giveup opens a confirm dialog and only dispatches after confirm', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<GrandfathersClockPage />);
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

  it('counts completed faces in the game-over summary', async () => {
    mockExec.mockResolvedValue({
      ...gameOverState,
      foundation: faces.map((f, i) => ({ ...f, complete: i < 5 })),
    });
    renderWithProviders(<GrandfathersClockPage />);
    const summary = await screen.findByTestId('gc-gameover-summary');
    expect(summary).toHaveTextContent('5/12');
  });

  it('does not show the progress summary on game clear', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<GrandfathersClockPage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームクリア/).length).toBeGreaterThan(0));
    expect(screen.queryByTestId('gc-gameover-summary')).not.toBeInTheDocument();
  });

  it('disables auto-complete while every face still holds only its starter', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<GrandfathersClockPage />);
    const btn = await screen.findByTestId('autocomplete-button');
    expect(btn).toBeDisabled();
    expect(btn).toHaveAttribute('title');
  });

  it('enables and pulses auto-complete once a face builds past its starter', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      foundation: faces.map((f, i) => (i === 0 ? { ...f, cards: [...f.cards, card('HEART', 3)] } : f)),
    });
    renderWithProviders(<GrandfathersClockPage />);
    const btn = await screen.findByTestId('autocomplete-button');
    expect(btn).toBeEnabled();
    expect(btn.className).toContain('animate-pulse');
  });

  it('shows StalemateEscapeButton when the stalemate flag is set', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: 2, canUndo: true });
    renderWithProviders(<GrandfathersClockPage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
  });

  it('labels empty columns 0-based, matching the visible #n headers', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<GrandfathersClockPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '空のタブロー列 2' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '空のタブロー列 7' })).toBeInTheDocument();
  });

  it.each([
    ['foundation', { fromCol: 1, toZone: 'foundation', toIdx: 4 }, '文字盤4'],
    ['tableau', { fromCol: 1, toZone: 'tableau', toIdx: 5 }, 'タブロー列5'],
  ])('renders a %s hint after the hint button is pressed', async (_name, hint, expected) => {
    mockExec.mockResolvedValueOnce(playingState).mockResolvedValueOnce({ ...playingState, hint });
    renderWithProviders(<GrandfathersClockPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getByText(new RegExp(expected))).toBeInTheDocument());
  });

  it('swaps the board for a terminal when CLI mode is toggled', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<GrandfathersClockPage />);
    await waitFor(() => expect(screen.getByText('#0')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /CLI/i }));
    await waitFor(() => expect(screen.queryByText('#0')).not.toBeInTheDocument());
  });

  it('names each tableau card with its position for screen readers', async () => {
    // Earlier tests in this file queue one-shot resolutions and can leave CLI
    // mode persisted in localStorage; reset both so the board actually renders.
    localStorage.clear();
    mockExec.mockReset();
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<GrandfathersClockPage />);
    await waitFor(() => expect(screen.getAllByLabelText(/列\d+・上から\d+枚目/).length).toBeGreaterThan(0));
  });
});

// Keyboard shortcuts are bound by useActionKeyboardNav and advertised by
// ActionShortcutsPanel; assert the keys actually run their action (#4429).
describe('GrandfathersClockPage keyboard shortcuts', () => {
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
    renderWithProviders(<GrandfathersClockPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(playingState);
    fireEvent.keyDown(document, { key });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
  });

  it('pressing g asks for give-up confirmation rather than firing it', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<GrandfathersClockPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.keyDown(document, { key: 'g' });
    expect(await screen.findByText('投了確認')).toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('ignores shortcuts once the game has ended', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<GrandfathersClockPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockExec.mockClear();
    for (const key of ['h', 'a', 'z']) {
      fireEvent.keyDown(document, { key });
    }
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });
});
