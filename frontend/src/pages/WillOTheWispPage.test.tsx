import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { willothewispApi } from '../api/games/willothewisp';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign } from '../types/card';
import type { WillOTheWispResponse, WillOTheWispTableauCard } from '../types/games/willothewisp';
import { WillOTheWispPage } from './WillOTheWispPage';

/**
 * This page's own hint region.
 *
 * **`GameMessageBox` is also `role="status"`**, and it now renders on every
 * phase because this game's messageCodes are translated (#5291). Querying the
 * role alone therefore matches two elements; the message box is the one built
 * from `glass-panel`, so the hint region is the other one.
 */
const hintLiveRegion = () =>
  screen.queryAllByRole('status').find((el) => !el.classList.contains('glass-panel')) ?? null;

vi.mock('../api/gameApi', () => ({
  actionLogApi: { willothewisp: vi.fn() },
}));

vi.mock('../api/games/willothewisp', () => ({
  willothewispApi: { exec: vi.fn() },
}));

vi.mock('../styles/gameTheme', () => ({
  gameTheme: {
    willothewisp: { bg: 'bg-table', footer: 'bg-footer' },
  },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const playSoundMock = vi.fn();
vi.mock('../providers/SoundProvider', async () => {
  const actual = await vi.importActual<typeof import('../providers/SoundProvider')>('../providers/SoundProvider');
  return {
    ...actual,
    useSound: () => ({ playSound: playSoundMock, muted: false, toggleMute: vi.fn() }),
  };
});

const mockSend = vi.mocked(willothewispApi.exec);

function makeTableau(cols: WillOTheWispTableauCard[][]): WillOTheWispTableauCard[][] {
  const result: WillOTheWispTableauCard[][] = [];
  for (let i = 0; i < 7; i++) {
    result.push(cols[i] ?? []);
  }
  return result;
}

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: WillOTheWispResponse = {
  tableau: makeTableau([
    [13, 12, 11].map((value) => ({ card: card('SPADE', value), faceUp: true })),
    [7, 5, 3].map((value) => ({ card: card('HEART', value), faceUp: true })),
    [10, 8, 6].map((value) => ({ card: card('DIAMOND', value), faceUp: true })),
    [9, 7, 5].map((value) => ({ card: card('CLOVER', value), faceUp: true })),
    [6, 4, 2].map((value) => ({ card: card('SPADE', value), faceUp: true })),
    [12, 10, 8].map((value) => ({ card: card('HEART', value), faceUp: true })),
    [11, 9, 7].map((value) => ({ card: card('DIAMOND', value), faceUp: true })),
  ]),
  stockCount: 31,
  completedSuits: 0,
  score: 500,
  scoring: { start: 500, movePenalty: 1, suitBonus: 100 },
  phase: 0,
  moveCount: 5,
  canUndo: false,
  isStalemate: false,
  message: '',
};

const gameClearState: WillOTheWispResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'willothewisp.gameClear',
  messageParams: { moveCount: '42', score: '500' },
};

beforeEach(() => {
  mockSend.mockResolvedValue(playingState);
  vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  playSoundMock.mockClear();
});

describe('WillOTheWispPage', () => {
  it('renders skeleton when no state', () => {
    mockSend.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<WillOTheWispPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('highlights the hint source card and target column after requesting a hint', async () => {
    renderWithProviders(<WillOTheWispPage />);
    await screen.findByTestId('willothewisp-card-1-1');
    // The Hint button fetches a hint: move HEART 5 (col 1, idx 1) onto column 0.
    mockSend.mockResolvedValue({ ...playingState, hint: { fromCol: 1, cardIndex: 1, toCol: 0 } });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    // Hint-suggested source uses ring-ds-info (distinct from user-selected ring-ds-warning).
    await waitFor(() => expect(screen.getByTestId('willothewisp-card-1-1').className).toContain('ring-ds-info'));
    expect(screen.getByTestId('willothewisp-card-1-1').className).not.toContain('ring-ds-warning');
    expect(screen.getByTestId('willothewisp-col-0').className).toContain('ring-ds-success');
    // A non-target column is not highlighted.
    expect(screen.getByTestId('willothewisp-col-2').className).not.toContain('ring-ds-success');
    // The hint text is exposed to screen readers via an aria-live status region.
    const status = hintLiveRegion();
    expect(status).toHaveAttribute('aria-live', 'polite');
    expect(status?.textContent).toContain('場札');
  });

  it('hides the frontend hint tooltip when hints are disabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'move', reason: 'frontendHint.buildSameSuit', confidence: 'strong' },
      hintEnabled: false,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<WillOTheWispPage />);
    await screen.findByTestId('willothewisp-card-1-1');
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();
  });

  it('shows the frontend hint tooltip when hints are enabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'move', reason: 'frontendHint.buildSameSuit', confidence: 'strong' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<WillOTheWispPage />);
    const tooltip = await screen.findByTestId('hint-tooltip');
    expect(tooltip).toHaveTextContent('同スートで積み重ねられるカードがあります');
  });

  it('announces the empty-column deal guard to screen readers', async () => {
    mockSend.mockResolvedValue({
      ...playingState,
      tableau: makeTableau([playingState.tableau[0], [], [], [], [], [], []]),
    });
    renderWithProviders(<WillOTheWispPage />);
    // An empty tableau column blocks dealing while stock remains.
    const dealBtn = (await screen.findAllByRole('button', { name: '配る' }))[0];
    fireEvent.click(dealBtn);
    await waitFor(() => {
      const warn = screen.getByText('空の列をすべて埋めないと配れません');
      expect(warn).toHaveAttribute('role', 'status');
      expect(warn).toHaveAttribute('aria-live', 'assertive');
    });
  });

  it('renders stock count', async () => {
    renderWithProviders(<WillOTheWispPage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());
    expect(screen.getByText(/\(31\)/)).toBeInTheDocument();
  });

  it('renders move count', async () => {
    renderWithProviders(<WillOTheWispPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent(/手数: 5/));
  });

  it('renders completed suits 0/4', async () => {
    renderWithProviders(<WillOTheWispPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent(/完成: 0\/4/));
  });

  it('deals from the stock when clicked in the normal state', async () => {
    renderWithProviders(<WillOTheWispPage />);
    const stockButton = (await screen.findAllByRole('button', { name: '配る' }))[0];
    mockSend.mockClear();

    fireEvent.click(stockButton);

    await waitFor(() => expect(mockSend).toHaveBeenCalledWith('deal'));
  });

  it('does not deal from the stock again while loading', async () => {
    renderWithProviders(<WillOTheWispPage />);
    const stockButton = (await screen.findAllByRole('button', { name: '配る' }))[0];
    mockSend.mockClear();
    mockSend.mockReturnValue(new Promise(() => undefined));

    fireEvent.click(stockButton);
    await waitFor(() => expect(mockSend).toHaveBeenCalledWith('deal'));
    const disabledStockButton = (await screen.findAllByRole('button', { name: '配る' }))[0];
    expect(disabledStockButton).toBeDisabled();
    fireEvent.click(disabledStockButton);

    expect(mockSend).toHaveBeenCalledTimes(1);
    expect(mockSend).toHaveBeenCalledWith('deal');
  });

  it('does not deal from the stock while auto-completing', async () => {
    const autoCompleteReadyState = { ...playingState, stockCount: 0 };
    mockSend.mockReset();
    mockSend.mockResolvedValue(playingState);
    mockSend.mockResolvedValueOnce(autoCompleteReadyState).mockResolvedValueOnce(playingState);
    renderWithProviders(<WillOTheWispPage />);
    const autoCompleteButton = await screen.findByTestId('autocomplete-button');
    fireEvent.click(autoCompleteButton);

    const stockButton = (await screen.findAllByRole('button', { name: '配る' }))[0];
    expect(stockButton).toBeDisabled();
    mockSend.mockClear();
    fireEvent.click(stockButton);

    await flushPendingDispatch();
    expect(mockSend).not.toHaveBeenCalledWith('deal');
  });

  it('shows game clear phase label', async () => {
    mockSend.mockResolvedValue(gameClearState);
    renderWithProviders(<WillOTheWispPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
  });

  it('give up button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    renderWithProviders(<WillOTheWispPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());

    mockSend.mockClear();
    // Clicking give-up must NOT dispatch immediately — it opens a confirm dialog (#2099).
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    await flushPendingDispatch();
    expect(mockSend).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();

    // Confirming dispatches giveup.
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockSend).toHaveBeenCalledWith('giveup'));
  });
});

// Keyboard shortcuts are bound by useActionKeyboardNav and advertised by
// ActionShortcutsPanel, but nothing asserted that pressing a key actually runs
// its action — a wrong `key` or a wrong `enabled` condition would have failed no
// test. See issue #4429.
describe('WillOTheWispPage keyboard shortcuts', () => {
  it.each([
    ['h', 'hint'],
    ['a', 'autocomplete'],
    ['z', 'undo'],
  ])('pressing %s dispatches %s', async (key, command) => {
    mockSend.mockResolvedValue(playingState);
    renderWithProviders(<WillOTheWispPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockSend.mockClear();
    mockSend.mockResolvedValue(playingState);
    fireEvent.keyDown(document, { key });
    await waitFor(() => expect(mockSend).toHaveBeenCalledWith(command));
  });

  it('pressing g asks for give-up confirmation rather than firing it', async () => {
    // give-up is irreversible, so the key must route through the dialog (#2099)
    // instead of dispatching straight away.
    mockSend.mockResolvedValue(playingState);
    renderWithProviders(<WillOTheWispPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockSend.mockClear();
    fireEvent.keyDown(document, { key: 'g' });
    expect(await screen.findByText('投了確認')).toBeInTheDocument();
    expect(mockSend).not.toHaveBeenCalled();
  });

  it('pressing d does not deal while an empty column blocks it', async () => {
    // handleDealGuarded refuses to deal when a tableau column is empty and stock remains.
    mockSend.mockResolvedValue({
      ...playingState,
      tableau: makeTableau([playingState.tableau[0], [], [], [], [], [], []]),
    });
    renderWithProviders(<WillOTheWispPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockSend.mockClear();
    fireEvent.keyDown(document, { key: 'd' });
    await flushPendingDispatch();
    expect(mockSend).not.toHaveBeenCalledWith('deal');
  });

  it('ignores shortcuts once the game has ended', async () => {
    mockSend.mockResolvedValue(gameClearState);
    renderWithProviders(<WillOTheWispPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockSend.mockClear();
    for (const key of ['d', 'h', 'a', 'z']) {
      fireEvent.keyDown(document, { key });
    }
    await flushPendingDispatch();
    expect(mockSend).not.toHaveBeenCalled();
  });

  // #5593, #6389: スコアが動く理由（手数の減点・スート完成の加点）はどこにも書かれて
  // いなかった。title ツールチップに加え、タッチ端末でも到達可能なポップオーバーを提供する。
  describe('score rule tooltip and popover', () => {
    it('explains the arithmetic with the figures the server sent in tooltip', async () => {
      mockSend.mockResolvedValue({ ...playingState, scoring: { start: 500, movePenalty: 1, suitBonus: 100 } });
      renderWithProviders(<WillOTheWispPage />);
      const score = await screen.findByTestId('willothewisp-score');
      const tip = score.getAttribute('title') ?? '';
      expect(tip).toContain('500');
      expect(tip).toContain('100');
    });

    // **数字を焼き込んでいない証拠。**別の決まりを返せばそのまま出る。
    it('renders whatever rule the server sends', async () => {
      mockSend.mockResolvedValue({ ...playingState, scoring: { start: 900, movePenalty: 5, suitBonus: 250 } });
      renderWithProviders(<WillOTheWispPage />);
      const tip = (await screen.findByTestId('willothewisp-score')).getAttribute('title') ?? '';
      expect(tip).toContain('900');
      expect(tip).toContain('250');
      expect(tip).not.toContain('500');
    });

    it('toggles score popover on click and shows breakdown with server figures', async () => {
      // 減点は 7。1 は ja の文言に「1 手ごとに」として必ず出るうえ 100 の部分文字列
      // でもあるので、部分一致では何も証明しない。符号ごと重ならない数字で見る。
      mockSend.mockResolvedValue({ ...playingState, scoring: { start: 500, movePenalty: 7, suitBonus: 100 } });
      renderWithProviders(<WillOTheWispPage />);
      const scoreBtn = await screen.findByTestId('willothewisp-score');

      // 初期状態: ポップオーバーは非表示、aria-expanded は false
      expect(screen.queryByTestId('willothewisp-score-popover')).not.toBeInTheDocument();
      expect(scoreBtn.getAttribute('aria-expanded')).toBe('false');

      // クリックで開く
      fireEvent.click(scoreBtn);
      const popover = screen.getByTestId('willothewisp-score-popover');
      expect(popover).toBeInTheDocument();
      expect(scoreBtn.getAttribute('aria-expanded')).toBe('true');
      // 領域名は本文の繰り返しであってはならない。
      expect(popover.getAttribute('aria-label')).not.toBe(popover.textContent);
      expect(popover.textContent).toContain('500');
      expect(popover.textContent).toContain('-7');
      expect(popover.textContent).toContain('+100');

      // もう一度クリックで閉じる
      fireEvent.click(scoreBtn);
      expect(screen.queryByTestId('willothewisp-score-popover')).not.toBeInTheDocument();
      expect(scoreBtn.getAttribute('aria-expanded')).toBe('false');
    });

    it('renders whatever rule the server sends in popover', async () => {
      mockSend.mockResolvedValue({ ...playingState, scoring: { start: 900, movePenalty: 13, suitBonus: 250 } });
      renderWithProviders(<WillOTheWispPage />);
      const scoreBtn = await screen.findByTestId('willothewisp-score');

      fireEvent.click(scoreBtn);
      const popover = screen.getByTestId('willothewisp-score-popover');
      expect(popover.textContent).toContain('900');
      expect(popover.textContent).toContain('-13');
      expect(popover.textContent).toContain('+250');
      expect(popover.textContent).not.toContain('500');
    });

    it('closes the score popover on Escape and on an outside click', async () => {
      mockSend.mockResolvedValue({ ...playingState, scoring: { start: 500, movePenalty: 1, suitBonus: 100 } });
      renderWithProviders(<WillOTheWispPage />);
      const scoreBtn = await screen.findByTestId('willothewisp-score');

      // Escape で閉じる
      fireEvent.click(scoreBtn);
      expect(screen.getByTestId('willothewisp-score-popover')).toBeInTheDocument();
      expect(scoreBtn.getAttribute('aria-expanded')).toBe('true');
      fireEvent.keyDown(document, { key: 'Escape' });
      expect(screen.queryByTestId('willothewisp-score-popover')).not.toBeInTheDocument();
      expect(scoreBtn.getAttribute('aria-expanded')).toBe('false');

      // 外側クリックで閉じる
      fireEvent.click(scoreBtn);
      expect(screen.getByTestId('willothewisp-score-popover')).toBeInTheDocument();
      expect(scoreBtn.getAttribute('aria-expanded')).toBe('true');
      fireEvent.mouseDown(document.body);
      expect(screen.queryByTestId('willothewisp-score-popover')).not.toBeInTheDocument();
      expect(scoreBtn.getAttribute('aria-expanded')).toBe('false');
    });

    it('score button meets the 44px tap-target minimum', async () => {
      mockSend.mockResolvedValue(playingState);
      renderWithProviders(<WillOTheWispPage />);
      const scoreBtn = await screen.findByTestId('willothewisp-score');
      expect(scoreBtn.className).toContain('min-h-[44px]');
      expect(scoreBtn.className).toContain('min-w-[44px]');
    });
  });
});
