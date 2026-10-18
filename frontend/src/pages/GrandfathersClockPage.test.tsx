import { act, fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
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
  canRedo: false,
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
  it('places settings details inside scrollable play area on mobile and outside on desktop, and tightens mobile margins', async () => {
    const original = window.innerWidth;
    mockExec.mockResolvedValue(playingState);
    try {
      Object.defineProperty(window, 'innerWidth', { configurable: true, value: 375 });
      window.dispatchEvent(new Event('resize'));
      const { container: mobileContainer, unmount } = renderWithProviders(<GrandfathersClockPage />);
      const mobileSettingsDetails = (await screen.findByText('設定')).closest('details');
      const mobileTableau = mobileContainer.querySelector('[data-tutorial="gc-tableau"]');
      const mobilePlayArea = mobileTableau?.closest('.overflow-y-auto');
      const mobileClock = mobileContainer.querySelector('[data-tutorial="gc-clock"]');
      expect(mobilePlayArea).not.toBeNull();
      expect(mobilePlayArea).toHaveClass('pt-1');
      expect(mobileClock).toHaveClass('mb-1');
      expect(mobileSettingsDetails).not.toBeNull();
      expect(mobilePlayArea?.contains(mobileSettingsDetails)).toBe(true);
      expect(
        mobileTableau &&
          mobileSettingsDetails &&
          mobileTableau.compareDocumentPosition(mobileSettingsDetails) & Node.DOCUMENT_POSITION_FOLLOWING,
      ).toBeTruthy();
      unmount();

      Object.defineProperty(window, 'innerWidth', { configurable: true, value: 1024 });
      window.dispatchEvent(new Event('resize'));
      const { container: desktopContainer } = renderWithProviders(<GrandfathersClockPage />);
      const desktopSettingsDetails = (await screen.findByText('設定')).closest('details');
      const desktopTableau = desktopContainer.querySelector('[data-tutorial="gc-tableau"]');
      const desktopPlayArea = desktopTableau?.closest('.overflow-y-auto');
      const desktopClock = desktopContainer.querySelector('[data-tutorial="gc-clock"]');
      expect(desktopPlayArea).not.toBeNull();
      expect(desktopPlayArea).toHaveClass('pt-3');
      expect(desktopClock).toHaveClass('mb-3');
      expect(desktopSettingsDetails).not.toBeNull();
      expect(desktopPlayArea?.contains(desktopSettingsDetails)).toBe(false);
    } finally {
      Object.defineProperty(window, 'innerWidth', { configurable: true, value: original });
      window.dispatchEvent(new Event('resize'));
    }
  });

  it('shows mobile control descriptions only while playing, matching the described controls', async () => {
    const original = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { configurable: true, value: 375 });
    window.dispatchEvent(new Event('resize'));
    try {
      mockExec.mockResolvedValue(playingState);
      const { unmount } = renderWithProviders(<GrandfathersClockPage />);
      const redo = await screen.findByRole('button', { name: 'やり直す' });
      const autoComplete = screen.getByTestId('autocomplete-button');
      expect(document.getElementById('gc-redo-unavailable')).toBeInTheDocument();
      expect(document.getElementById(redo.getAttribute('aria-describedby') ?? '')).toBeInTheDocument();
      expect(document.getElementById(autoComplete.getAttribute('aria-describedby') ?? '')).toBeInTheDocument();
      unmount();

      mockExec.mockResolvedValue(gameOverState);
      renderWithProviders(<GrandfathersClockPage />);
      await screen.findByTestId('gc-gameover-summary');
      expect(screen.queryByRole('button', { name: 'やり直す' })).not.toBeInTheDocument();
      expect(screen.queryByTestId('autocomplete-button')).not.toBeInTheDocument();
      expect(document.getElementById('gc-redo-unavailable')).not.toBeInTheDocument();
      expect(document.getElementById('gc-autocomplete-reason')).not.toBeInTheDocument();
    } finally {
      Object.defineProperty(window, 'innerWidth', { configurable: true, value: original });
      window.dispatchEvent(new Event('resize'));
    }
  });

  it('places control descriptions inside play area on mobile and inside footer on desktop', async () => {
    const original = window.innerWidth;
    mockExec.mockResolvedValue(playingState);
    try {
      Object.defineProperty(window, 'innerWidth', { configurable: true, value: 375 });
      window.dispatchEvent(new Event('resize'));
      const { container: mobileContainer, unmount } = renderWithProviders(<GrandfathersClockPage />);
      const mobileAutoCompleteReason = await screen.findByText('文字盤を配り札より先へ進めると有効になります');
      const mobileRedoUnavailable = screen.getByText('やり直せる手がありません');
      const mobileFooter = mobileContainer.querySelector('footer');
      const mobileTableau = mobileContainer.querySelector('[data-tutorial="gc-tableau"]');
      const mobilePlayArea = mobileTableau?.closest('.overflow-y-auto');

      expect(mobilePlayArea).toBeInstanceOf(HTMLElement);
      expect(mobileFooter).toBeInstanceOf(HTMLElement);

      // Outside footer on mobile
      expect(mobileFooter?.contains(mobileAutoCompleteReason)).toBe(false);
      expect(mobileFooter?.contains(mobileRedoUnavailable)).toBe(false);

      // Inside play area on mobile
      expect(mobilePlayArea?.contains(mobileAutoCompleteReason)).toBe(true);
      expect(mobilePlayArea?.contains(mobileRedoUnavailable)).toBe(true);

      // After tableau, before settings details
      const mobileSettingsDetails = (await screen.findByText('設定')).closest('details');
      expect(mobileSettingsDetails).not.toBeNull();
      expect(
        mobileTableau &&
          mobileAutoCompleteReason &&
          mobileTableau.compareDocumentPosition(mobileAutoCompleteReason) & Node.DOCUMENT_POSITION_FOLLOWING,
      ).toBeTruthy();
      expect(
        mobileAutoCompleteReason &&
          mobileSettingsDetails &&
          mobileAutoCompleteReason.compareDocumentPosition(mobileSettingsDetails) & Node.DOCUMENT_POSITION_FOLLOWING,
      ).toBeTruthy();

      // aria-describedby references preserved
      const mobileAutoCompleteBtn = screen.getByTestId('autocomplete-button');
      expect(mobileAutoCompleteBtn).toHaveAttribute('aria-describedby', 'gc-autocomplete-reason');
      const mobileRedoBtn = screen.getByRole('button', { name: 'やり直す' });
      expect(mobileRedoBtn).toHaveAttribute('aria-describedby', 'gc-redo-unavailable');

      unmount();

      Object.defineProperty(window, 'innerWidth', { configurable: true, value: 1024 });
      window.dispatchEvent(new Event('resize'));
      const { container: desktopContainer } = renderWithProviders(<GrandfathersClockPage />);
      const desktopAutoCompleteReason = await screen.findByText('文字盤を配り札より先へ進めると有効になります');
      const desktopRedoUnavailable = screen.getByText('やり直せる手がありません');
      const desktopFooter = desktopContainer.querySelector('footer');
      const desktopTableau = desktopContainer.querySelector('[data-tutorial="gc-tableau"]');
      const desktopPlayArea = desktopTableau?.closest('.overflow-y-auto');

      expect(desktopPlayArea).toBeInstanceOf(HTMLElement);
      expect(desktopFooter).toBeInstanceOf(HTMLElement);

      // Inside footer on desktop
      expect(desktopFooter?.contains(desktopAutoCompleteReason)).toBe(true);
      expect(desktopFooter?.contains(desktopRedoUnavailable)).toBe(true);

      // Outside play area on desktop
      expect(desktopPlayArea?.contains(desktopAutoCompleteReason)).toBe(false);
      expect(desktopPlayArea?.contains(desktopRedoUnavailable)).toBe(false);

      // aria-describedby references preserved
      const desktopAutoCompleteBtn = screen.getByTestId('autocomplete-button');
      expect(desktopAutoCompleteBtn).toHaveAttribute('aria-describedby', 'gc-autocomplete-reason');
      const desktopRedoBtn = screen.getByRole('button', { name: 'やり直す' });
      expect(desktopRedoBtn).toHaveAttribute('aria-describedby', 'gc-redo-unavailable');
    } finally {
      Object.defineProperty(window, 'innerWidth', { configurable: true, value: original });
      window.dispatchEvent(new Event('resize'));
    }
  });

  afterEach(() => {
    vi.useRealTimers();
  });

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

  it('wraps the twelve compact clock slots at large desktop', async () => {
    const original = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { configurable: true, value: 1280 });
    mockExec.mockResolvedValue(playingState);
    try {
      renderWithProviders(<GrandfathersClockPage />);
      await screen.findByTestId('gc-autocomplete-status');
      const clock = document.querySelector('[data-tutorial="gc-clock"]') as HTMLElement;
      expect(clock).toHaveClass('flex-wrap');
      expect(clock.querySelectorAll('[role="img"], button')).toHaveLength(12);
      expect(clock.querySelector('img')).toHaveStyle({ width: '80px' });
    } finally {
      Object.defineProperty(window, 'innerWidth', { configurable: true, value: original });
    }
  });

  it('keeps develop card sizing and clock markup at desktop width', async () => {
    const original = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { configurable: true, value: 800 });
    mockExec.mockResolvedValue(playingState);
    try {
      renderWithProviders(<GrandfathersClockPage />);
      await screen.findByTestId('gc-autocomplete-status');
      const clock = document.querySelector('[data-tutorial="gc-clock"]') as HTMLElement;
      expect(clock).not.toHaveAttribute('data-testid', 'gc-clock-row');
      expect(clock).toHaveClass('flex-wrap');
      expect(clock.querySelector('img')).toHaveStyle({ width: '60px' });
    } finally {
      Object.defineProperty(window, 'innerWidth', { configurable: true, value: original });
    }
  });

  it('wraps clock slots and sizes them to the mobile viewport', async () => {
    const original = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { configurable: true, value: 375 });
    mockExec.mockResolvedValue(playingState);
    try {
      renderWithProviders(<GrandfathersClockPage />);
      await screen.findByTestId('gc-autocomplete-status');
      const clock = document.querySelector('[data-tutorial="gc-clock"]') as HTMLElement;
      expect(clock).toHaveClass('flex-wrap');
      expect(clock.querySelector('img')).toHaveStyle({ width: '40px' });
    } finally {
      Object.defineProperty(window, 'innerWidth', { configurable: true, value: original });
    }
  });

  it('enables redo only when the server reports redo history', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<GrandfathersClockPage />);
    const redo = await screen.findByRole('button', { name: 'やり直す' });
    expect(redo).toHaveAttribute('aria-disabled', 'true');
    expect(redo).toHaveAttribute('aria-describedby', 'gc-redo-unavailable');
    mockExec.mockClear();
    fireEvent.click(redo);
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    mockExec.mockResolvedValue({ ...playingState, canRedo: true });
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    const enabledRedo = await screen.findByRole('button', { name: 'やり直す' });
    expect(enabledRedo).toHaveAttribute('aria-disabled', 'false');
    expect(enabledRedo).not.toHaveAttribute('aria-describedby');
    fireEvent.click(enabledRedo);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('redo'));
  });

  it('renders all twelve clock faces', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<GrandfathersClockPage />);
    await waitFor(() => expect(screen.getAllByLabelText(/文字盤\d+ \(\d+時\)/).length).toBe(12));
  });

  it('announces the target rank of an empty clock face', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      foundation: faces.map((face, i) => (i === 0 ? { ...face, cards: [] } : face)),
    });
    renderWithProviders(<GrandfathersClockPage />);
    expect(await screen.findByRole('img', { name: '空の文字盤0 (1時)、目標1' })).toBeInTheDocument();
  });

  it('keeps a clock face target focusable and explains the missing source', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<GrandfathersClockPage />);
    const target = await screen.findByLabelText(/文字盤0/);
    expect(target).not.toBeDisabled();
    expect(target).toHaveAttribute('aria-disabled', 'true');
    const hintId = target.getAttribute('aria-describedby');
    expect(hintId).toBeTruthy();
    expect(document.getElementById(hintId ?? '')).toHaveTextContent('先に移動する札を選んでください');
    mockExec.mockClear();
    fireEvent.click(target);
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('move', expect.anything(), expect.anything());
  });

  // The target rank is what the player plans against, so it has to be on screen
  // rather than implied by the clock position.
  it('shows each face target: face 0 wants an Ace, face 11 a Queen', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<GrandfathersClockPage />);
    await waitFor(() => expect(screen.getByLabelText(/文字盤0 \(1時\) 目標1/)).toBeInTheDocument());
    expect(screen.getByLabelText(/文字盤11 \(12時\) 目標12/)).toBeInTheDocument();
  });

  it('marks only legal clock faces after selecting a card', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      tableau: makeTableau([[{ card: card('HEART', 3), faceUp: true }]]),
    });
    renderWithProviders(<GrandfathersClockPage />);
    const source = await screen.findByRole('button', { name: /^♥ 3（/ });
    fireEvent.click(source);
    await waitFor(() => expect(source).toHaveAttribute('aria-pressed', 'true'));

    expect(screen.getByLabelText(/文字盤0 \(1時\).*配置できます/)).toBeInTheDocument();
    expect(screen.getByLabelText(/文字盤1 \(2時\).*配置できません/)).toBeInTheDocument();
  });

  it('labels all eight tableau columns with their 0-based index', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<GrandfathersClockPage />);
    await waitFor(() => expect(screen.getByText('#0')).toBeInTheDocument());
    for (let i = 0; i < 8; i++) {
      expect(screen.getByText(`#${i}`)).toBeInTheDocument();
    }
  });

  it('announces that any card can be placed in an empty tableau column', async () => {
    mockExec.mockResolvedValue({ ...playingState, tableau: makeTableau([]) });
    renderWithProviders(<GrandfathersClockPage />);
    expect(await screen.findByRole('button', { name: '空のタブロー列 0。任意のカードを置けます' })).toBeInTheDocument();
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

  it('explains how to enable auto-complete and keeps its reason associated while unavailable', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<GrandfathersClockPage />);
    const btn = await screen.findByTestId('autocomplete-button');
    expect(btn).toHaveAttribute('aria-disabled', 'true');
    const reason = screen.getByText('文字盤を配り札より先へ進めると有効になります');
    expect(reason).toBeInTheDocument();
    expect(btn).toHaveAttribute('aria-describedby', reason.id);
    fireEvent.click(btn);
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalledWith('autocomplete');
  });

  it('enables and pulses auto-complete once a face builds past its starter', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      foundation: faces.map((f, i) => (i === 0 ? { ...f, cards: [...f.cards, card('HEART', 3)] } : f)),
    });
    renderWithProviders(<GrandfathersClockPage />);
    const btn = await screen.findByTestId('autocomplete-button');
    expect(btn).toBeEnabled();
    expect(btn).not.toHaveAttribute('aria-disabled');
    expect(btn).not.toHaveAttribute('aria-describedby');
    expect(screen.queryByText('文字盤を配り札より先へ進めると有効になります')).not.toBeInTheDocument();
    expect(btn.className).toContain('animate-pulse');
  });

  it('announces auto-complete start and completion and clears the visible status', async () => {
    mockExec.mockResolvedValue({
      ...playingState,
      foundation: faces.map((f, i) => (i === 0 ? { ...f, cards: [...f.cards, card('HEART', 3)] } : f)),
    });
    renderWithProviders(<GrandfathersClockPage />);
    const button = await screen.findByTestId('autocomplete-button');
    await waitFor(() => expect(button).toBeEnabled());

    vi.useFakeTimers();
    fireEvent.click(button);
    const liveRegion = screen.getByTestId('gc-autocomplete-status');
    expect(liveRegion).toHaveClass('sr-only');
    expect(liveRegion).toHaveTextContent('オートコンプリート中');
    expect(screen.getByTestId('gc-autocomplete-visible')).toBeInTheDocument();
    expect(button).toHaveAttribute('aria-disabled', 'true');

    await act(async () => {
      vi.advanceTimersByTime(3000);
    });
    expect(liveRegion).toHaveTextContent('オートコンプリートが完了しました');
    expect(screen.queryByTestId('gc-autocomplete-visible')).not.toBeInTheDocument();
    expect(button).toBeEnabled();
    vi.useRealTimers();
  });

  it('shows StalemateEscapeButton when the stalemate flag is set', async () => {
    mockExec.mockResolvedValue({ ...playingState, isStalemate: true, undoToEscape: 2, canUndo: true });
    renderWithProviders(<GrandfathersClockPage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
  });

  it('labels empty columns 0-based, matching the visible #n headers', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<GrandfathersClockPage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '空のタブロー列 2。任意のカードを置けます' })).toBeInTheDocument(),
    );
    expect(screen.getByRole('button', { name: '空のタブロー列 7。任意のカードを置けます' })).toBeInTheDocument();
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

describe('GrandfathersClockPage compact large-desktop dimensions', () => {
  it('uses 60px cards at 1024px and 80px cards at 1280px', async () => {
    const originalWidth = window.innerWidth;
    try {
      mockExec.mockResolvedValue(playingState);
      Object.defineProperty(window, 'innerWidth', { configurable: true, value: 1024 });
      const compact = renderWithProviders(<GrandfathersClockPage />);
      await screen.findByTestId('gc-autocomplete-status');
      const compactCard = document.querySelector('[data-tutorial="gc-clock"] img');
      expect(compactCard).toHaveStyle({ width: '60px' });
      compact.unmount();

      Object.defineProperty(window, 'innerWidth', { configurable: true, value: 1280 });
      const fullSize = renderWithProviders(<GrandfathersClockPage />);
      await screen.findByTestId('gc-autocomplete-status');
      const fullSizeCard = document.querySelector('[data-tutorial="gc-clock"] img');
      expect(fullSizeCard).toHaveStyle({ width: '80px' });
      fullSize.unmount();
    } finally {
      Object.defineProperty(window, 'innerWidth', { configurable: true, value: originalWidth });
      window.dispatchEvent(new Event('resize'));
    }
  });
});
