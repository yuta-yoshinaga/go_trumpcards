import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { spideretteApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, SpideretteResponse, SpideretteTableauCard } from '../types/card';
import { SpiderettePage } from './SpiderettePage';

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
  spideretteApi: { exec: vi.fn() },
  actionLogApi: { spiderette: vi.fn() },
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

const mockSend = vi.mocked(spideretteApi.exec);

function makeTableau(cols: SpideretteTableauCard[][]): SpideretteTableauCard[][] {
  const result: SpideretteTableauCard[][] = [];
  for (let i = 0; i < 7; i++) {
    result.push(cols[i] ?? []);
  }
  return result;
}

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: SpideretteResponse = {
  tableau: makeTableau([
    [{ card: card('SPADE', 13), faceUp: true }],
    [
      { card: null, faceUp: false },
      { card: card('HEART', 5), faceUp: true },
    ],
    [],
    [],
    [],
    [],
    [],
  ]),
  stockCount: 24,
  completedSuits: 0,
  score: 500,
  scoring: { start: 500, movePenalty: 1, suitBonus: 100 },
  phase: 0,
  moveCount: 5,
  canUndo: false,
  isStalemate: false,
  message: '',
};

const gameClearState: SpideretteResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'spiderette.gameClear',
  messageParams: { moveCount: '42', score: '500' },
};

beforeEach(() => {
  mockSend.mockResolvedValue(playingState);
  vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  playSoundMock.mockClear();
});

describe('SpiderettePage', () => {
  it('renders skeleton when no state', () => {
    mockSend.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<SpiderettePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('highlights the hint source card and target column after requesting a hint', async () => {
    renderWithProviders(<SpiderettePage />);
    await screen.findByTestId('spdt-card-1-1');
    // The Hint button fetches a hint: move HEART 5 (col 1, idx 1) onto column 0.
    mockSend.mockResolvedValue({ ...playingState, hint: { fromCol: 1, cardIndex: 1, toCol: 0 } });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    // Hint-suggested source uses ring-ds-info (distinct from user-selected ring-ds-warning).
    await waitFor(() => expect(screen.getByTestId('spdt-card-1-1').className).toContain('ring-ds-info'));
    expect(screen.getByTestId('spdt-card-1-1').className).not.toContain('ring-ds-warning');
    expect(screen.getByTestId('spdt-col-0').className).toContain('ring-ds-success');
    // A non-target column is not highlighted.
    expect(screen.getByTestId('spdt-col-2').className).not.toContain('ring-ds-success');
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
    renderWithProviders(<SpiderettePage />);
    await screen.findByTestId('spdt-card-1-1');
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();
  });

  it('shows the frontend hint tooltip when hints are enabled', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'move', reason: 'frontendHint.buildSameSuit', confidence: 'strong' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<SpiderettePage />);
    const tooltip = await screen.findByTestId('hint-tooltip');
    expect(tooltip).toHaveTextContent('同スートで積み重ねられるカードがあります');
  });

  it('announces the empty-column deal guard to screen readers', async () => {
    renderWithProviders(<SpiderettePage />);
    // playingState has empty columns and stock remaining, so a deal is guarded.
    const dealBtn = (await screen.findAllByRole('button', { name: '配る' }))[0];
    fireEvent.click(dealBtn);
    await waitFor(() => {
      const warn = screen.getByText('空の列をすべて埋めないと配れません');
      expect(warn).toHaveAttribute('role', 'status');
      expect(warn).toHaveAttribute('aria-live', 'assertive');
    });
  });

  it('renders stock count', async () => {
    renderWithProviders(<SpiderettePage />);
    await waitFor(() => expect(screen.getByText(/山札 \(/)).toBeInTheDocument());
    expect(screen.getByText(/\(24\)/)).toBeInTheDocument();
  });

  it('renders move count', async () => {
    renderWithProviders(<SpiderettePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent(/手数: 5/));
  });

  it('renders completed suits 0/4', async () => {
    renderWithProviders(<SpiderettePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent(/完成: 0\/4/));
  });

  it('shows game clear phase label', async () => {
    mockSend.mockResolvedValue(gameClearState);
    renderWithProviders(<SpiderettePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
  });

  it('give up button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    renderWithProviders(<SpiderettePage />);
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
describe('SpiderettePage keyboard shortcuts', () => {
  it.each([
    ['h', 'hint'],
    ['a', 'autocomplete'],
    ['z', 'undo'],
  ])('pressing %s dispatches %s', async (key, command) => {
    mockSend.mockResolvedValue(playingState);
    renderWithProviders(<SpiderettePage />);
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
    renderWithProviders(<SpiderettePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockSend.mockClear();
    fireEvent.keyDown(document, { key: 'g' });
    expect(await screen.findByText('投了確認')).toBeInTheDocument();
    expect(mockSend).not.toHaveBeenCalled();
  });

  it('pressing d does not deal while an empty column blocks it', async () => {
    // handleDealGuarded refuses to deal when a tableau column is empty and stock
    // remains. playingState is exactly that (four empty columns, stock 24), so the
    // key must be inert rather than dispatching. Verified empirically: the same
    // test against Easthaven's fixture DOES dispatch, because its tableau is full.
    mockSend.mockResolvedValue(playingState);
    renderWithProviders(<SpiderettePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    mockSend.mockClear();
    fireEvent.keyDown(document, { key: 'd' });
    await flushPendingDispatch();
    expect(mockSend).not.toHaveBeenCalledWith('deal');
  });

  it('ignores shortcuts once the game has ended', async () => {
    mockSend.mockResolvedValue(gameClearState);
    renderWithProviders(<SpiderettePage />);
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
      renderWithProviders(<SpiderettePage />);
      const score = await screen.findByTestId('spiderette-score');
      const tip = score.getAttribute('title') ?? '';
      expect(tip).toContain('500');
      expect(tip).toContain('100');
    });

    // **数字を焼き込んでいない証拠。**別の決まりを返せばそのまま出る。
    it('renders whatever rule the server sends', async () => {
      mockSend.mockResolvedValue({ ...playingState, scoring: { start: 900, movePenalty: 5, suitBonus: 250 } });
      renderWithProviders(<SpiderettePage />);
      const tip = (await screen.findByTestId('spiderette-score')).getAttribute('title') ?? '';
      expect(tip).toContain('900');
      expect(tip).toContain('250');
      expect(tip).not.toContain('500');
    });

    it('toggles score popover on click and shows breakdown with server figures', async () => {
      // 減点は 7。1 は ja の文言に「1 手ごとに」として必ず出るうえ 100 の部分文字列
      // でもあるので、部分一致では何も証明しない。符号ごと重ならない数字で見る。
      mockSend.mockResolvedValue({ ...playingState, scoring: { start: 500, movePenalty: 7, suitBonus: 100 } });
      renderWithProviders(<SpiderettePage />);
      const scoreBtn = await screen.findByTestId('spiderette-score');

      // 初期状態: ポップオーバーは非表示、aria-expanded は false
      expect(screen.queryByTestId('spiderette-score-popover')).not.toBeInTheDocument();
      expect(scoreBtn.getAttribute('aria-expanded')).toBe('false');

      // クリックで開く
      fireEvent.click(scoreBtn);
      const popover = screen.getByTestId('spiderette-score-popover');
      expect(popover).toBeInTheDocument();
      expect(scoreBtn.getAttribute('aria-expanded')).toBe('true');
      // 領域名は本文の繰り返しであってはならない。
      expect(popover.getAttribute('aria-label')).not.toBe(popover.textContent);
      expect(popover.textContent).toContain('500');
      expect(popover.textContent).toContain('-7');
      expect(popover.textContent).toContain('+100');

      // もう一度クリックで閉じる
      fireEvent.click(scoreBtn);
      expect(screen.queryByTestId('spiderette-score-popover')).not.toBeInTheDocument();
      expect(scoreBtn.getAttribute('aria-expanded')).toBe('false');
    });

    it('renders whatever rule the server sends in popover', async () => {
      mockSend.mockResolvedValue({ ...playingState, scoring: { start: 900, movePenalty: 13, suitBonus: 250 } });
      renderWithProviders(<SpiderettePage />);
      const scoreBtn = await screen.findByTestId('spiderette-score');

      fireEvent.click(scoreBtn);
      const popover = screen.getByTestId('spiderette-score-popover');
      expect(popover.textContent).toContain('900');
      expect(popover.textContent).toContain('-13');
      expect(popover.textContent).toContain('+250');
      expect(popover.textContent).not.toContain('500');
    });

    it('closes the score popover on Escape and on an outside click', async () => {
      mockSend.mockResolvedValue({ ...playingState, scoring: { start: 500, movePenalty: 1, suitBonus: 100 } });
      renderWithProviders(<SpiderettePage />);
      const scoreBtn = await screen.findByTestId('spiderette-score');

      // Escape で閉じる
      fireEvent.click(scoreBtn);
      expect(screen.getByTestId('spiderette-score-popover')).toBeInTheDocument();
      expect(scoreBtn.getAttribute('aria-expanded')).toBe('true');
      fireEvent.keyDown(document, { key: 'Escape' });
      expect(screen.queryByTestId('spiderette-score-popover')).not.toBeInTheDocument();
      expect(scoreBtn.getAttribute('aria-expanded')).toBe('false');

      // 外側クリックで閉じる
      fireEvent.click(scoreBtn);
      expect(screen.getByTestId('spiderette-score-popover')).toBeInTheDocument();
      expect(scoreBtn.getAttribute('aria-expanded')).toBe('true');
      fireEvent.mouseDown(document.body);
      expect(screen.queryByTestId('spiderette-score-popover')).not.toBeInTheDocument();
      expect(scoreBtn.getAttribute('aria-expanded')).toBe('false');
    });

    it('score button meets the 44px tap-target minimum', async () => {
      mockSend.mockResolvedValue(playingState);
      renderWithProviders(<SpiderettePage />);
      const scoreBtn = await screen.findByTestId('spiderette-score');
      expect(scoreBtn.className).toContain('min-h-[44px]');
      expect(scoreBtn.className).toContain('min-w-[44px]');
    });
  });
});
