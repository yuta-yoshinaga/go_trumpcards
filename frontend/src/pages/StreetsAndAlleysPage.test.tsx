import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { streetsAndAlleysApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, CardDesign, StreetsAndAlleysResponse, StreetsAndAlleysTableauCard } from '../types/card';
import { StreetsAndAlleysPage } from './StreetsAndAlleysPage';

vi.mock('../api/gameApi', () => ({
  streetsAndAlleysApi: { exec: vi.fn() },
  actionLogApi: { streetsandalleys: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(streetsAndAlleysApi.exec);

function makeTableau(cols: StreetsAndAlleysTableauCard[][]): StreetsAndAlleysTableauCard[][] {
  const result: StreetsAndAlleysTableauCard[][] = [];
  for (let i = 0; i < 8; i++) {
    result.push(cols[i] ?? []);
  }
  return result;
}

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: StreetsAndAlleysResponse = {
  tableau: makeTableau([
    [
      { card: card('SPADE', 13), faceUp: true },
      { card: card('SPADE', 5), faceUp: true },
    ],
    [{ card: card('HEART', 6), faceUp: true }],
    [],
    [],
    [],
    [],
    [],
    [],
  ]),
  foundation: [[card('SPADE', 1)], [card('CLOVER', 1)], [card('HEART', 1)], [card('DIAMOND', 1)]],
  phase: 0,
  moveCount: 3,
  canUndo: false,
  isStalemate: false,
  message: '',
};

const gameClearState: StreetsAndAlleysResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'streetsandalleys.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: StreetsAndAlleysResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'streetsandalleys.gameOver',
};

describe('StreetsAndAlleysPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  });

  it('calls reset on initial render', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<StreetsAndAlleysPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
  });

  it('highlights column tops as drop targets once a source card is selected', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<StreetsAndAlleysPage />);
    // No source selected yet → no target-candidate highlight.
    const heart6 = await screen.findByRole('button', { name: '♥ 6' });
    expect(heart6).not.toHaveAttribute('data-target-candidate');
    // Select the top of column 0 (♠5) as the move source.
    fireEvent.click(screen.getByRole('button', { name: '♠ 5' }));
    // Column 1's top (♥6) is now a drop target with the info ring.
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '♥ 6' })).toHaveAttribute('data-target-candidate', 'true'),
    );
    expect(screen.getByRole('button', { name: '♥ 6' }).className).toContain('ring-ds-info');
    // The selected source card itself is not a target candidate.
    expect(screen.getByRole('button', { name: '♠ 5' })).not.toHaveAttribute('data-target-candidate');
  });

  it('dims a tableau card while it is being dragged', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<StreetsAndAlleysPage />);
    const top = await screen.findByRole('button', { name: '♠ 5' });
    // jsdom doesn't attach a DataTransfer to synthetic drag events, so provide one.
    fireEvent.dragStart(top, { dataTransfer: { setData: () => {}, effectAllowed: '' } });
    await waitFor(() => expect(top.className).toContain('opacity-50'));
  });

  it('renders heading and move count', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<StreetsAndAlleysPage />);
    await waitFor(() => expect(screen.getByText(/ストリート・アンド・アレイズ/)).toBeInTheDocument());
    expect(screen.getByText(/手数: 3/)).toBeInTheDocument();
  });

  it('renders 4 foundation suits', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<StreetsAndAlleysPage />);
    await waitFor(() => expect(screen.getAllByLabelText(/組札 1枚/).length).toBe(4));
  });

  it('renders per-foundation progress counters (n/13) for partially-built piles', async () => {
    const progressState: StreetsAndAlleysResponse = {
      ...playingState,
      foundation: [
        [card('SPADE', 1), card('SPADE', 2), card('SPADE', 3), card('SPADE', 4), card('SPADE', 5)],
        [card('CLOVER', 1)],
        [card('HEART', 1), card('HEART', 2)],
        [card('DIAMOND', 1)],
      ],
    };
    mockExec.mockResolvedValue(progressState);
    renderWithProviders(<StreetsAndAlleysPage />);
    expect(await screen.findByTestId('sa-foundation-progress-0')).toHaveTextContent('5/13');
    expect(screen.getByTestId('sa-foundation-progress-1')).toHaveTextContent('1/13');
    expect(screen.getByTestId('sa-foundation-progress-2')).toHaveTextContent('2/13');
    expect(screen.getByTestId('sa-foundation-progress-3')).toHaveTextContent('1/13');
  });

  it('marks a completed foundation (13/13) with a success color and checkmark', async () => {
    const fullSpades = Array.from({ length: 13 }, (_, i) => card('SPADE', i + 1));
    const completeState: StreetsAndAlleysResponse = {
      ...playingState,
      foundation: [fullSpades, [card('CLOVER', 1)], [card('HEART', 1)], [card('DIAMOND', 1)]],
    };
    mockExec.mockResolvedValue(completeState);
    renderWithProviders(<StreetsAndAlleysPage />);
    const done = await screen.findByTestId('sa-foundation-progress-0');
    expect(done).toHaveTextContent('13/13');
    expect(done.textContent).toContain('✓');
    expect(done.className).toContain('text-ds-success');
  });

  it('renders giveup button when playing', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<StreetsAndAlleysPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
  });

  it('advertises the keyboard shortcuts on the control buttons', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<StreetsAndAlleysPage />);
    const undo = await screen.findByRole('button', { name: '元に戻す' });
    expect(undo).toHaveAttribute('aria-keyshortcuts', 'z');
    expect(undo.querySelector('kbd')?.textContent).toBe('Z');
    const hint = screen.getByRole('button', { name: 'ヒント' });
    expect(hint).toHaveAttribute('aria-keyshortcuts', 'h');
    expect(hint.querySelector('kbd')?.textContent).toBe('H');
    const autoComplete = screen.getByRole('button', { name: '自動完成' });
    expect(autoComplete).toHaveAttribute('aria-keyshortcuts', 'a');
    expect(autoComplete.querySelector('kbd')?.textContent).toBe('A');
    // KbdBadge text is aria-hidden, so the giveup button's accessible name stays clean.
    const giveUp = screen.getByRole('button', { name: 'ギブアップ' });
    expect(giveUp).toHaveAttribute('aria-keyshortcuts', 'g');
    expect(giveUp.querySelector('kbd')?.textContent).toBe('G');
  });

  it('hides giveup button when game cleared', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<StreetsAndAlleysPage />);
    await waitFor(() => expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument());
  });

  it('giveup button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<StreetsAndAlleysPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(gameOverState);
    fireEvent.click(screen.getByRole('button', { name: 'ギブアップ' }));
    expect(mockExec).not.toHaveBeenCalledWith('giveup');
    expect(screen.getByText('投了確認')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('giveup'));
  });

  it('shows phase name in header for game over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<StreetsAndAlleysPage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームオーバー/).length).toBeGreaterThan(0));
  });

  it('shows the foundation progress summary on game over', async () => {
    mockExec.mockResolvedValue(gameOverState); // 4 aces on foundations → 4/52 (8%)
    renderWithProviders(<StreetsAndAlleysPage />);
    const summary = await screen.findByTestId('sa-gameover-summary');
    expect(summary).toHaveTextContent('4/52');
    expect(summary).toHaveTextContent('8%');
  });

  it('does not show the progress summary on game clear', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<StreetsAndAlleysPage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームクリア/).length).toBeGreaterThan(0));
    expect(screen.queryByTestId('sa-gameover-summary')).not.toBeInTheDocument();
  });

  it('disables the auto-complete button and shows a reason when no foundation has progressed', async () => {
    mockExec.mockResolvedValue(playingState); // foundations hold only aces
    renderWithProviders(<StreetsAndAlleysPage />);
    const btn = await screen.findByTestId('autocomplete-button');
    expect(btn).toBeDisabled();
    expect(btn.className).not.toContain('animate-pulse');
    expect(btn).toHaveAttribute('title');
  });

  it('enables and pulses the auto-complete button once a foundation builds past its ace', async () => {
    const readyState: StreetsAndAlleysResponse = {
      ...playingState,
      foundation: [[card('SPADE', 1), card('SPADE', 2)], [card('CLOVER', 1)], [card('HEART', 1)], [card('DIAMOND', 1)]],
    };
    mockExec.mockResolvedValue(readyState);
    renderWithProviders(<StreetsAndAlleysPage />);
    const btn = await screen.findByTestId('autocomplete-button');
    expect(btn).toBeEnabled();
    expect(btn.className).toContain('animate-pulse');
    expect(btn.className).toContain('ring-ds-success');
  });

  it('shows StalemateEscapeButton when stalemate flag is set', async () => {
    const stalemate: StreetsAndAlleysResponse = {
      ...playingState,
      isStalemate: true,
      undoToEscape: 2,
      canUndo: true,
    };
    mockExec.mockResolvedValue(stalemate);
    renderWithProviders(<StreetsAndAlleysPage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
  });

  it('selecting a tableau card marks it as selected', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<StreetsAndAlleysPage />);
    const sourceBtn = await screen.findByRole('button', { name: '♠ 5' });
    fireEvent.click(sourceBtn);
    await waitFor(() => expect(sourceBtn).toHaveAttribute('aria-pressed', 'true'));
  });
});
