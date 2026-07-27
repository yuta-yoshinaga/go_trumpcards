import { fireEvent, screen, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { bakersDozenApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BakersDozenResponse, BakersDozenTableauCard, Card, CardDesign } from '../types/card';
import { BakersDozenPage } from './BakersDozenPage';

vi.mock('../api/gameApi', () => ({
  bakersDozenApi: { exec: vi.fn() },
  actionLogApi: { bakersdozen: vi.fn() },
}));

const mockPlaySound = vi.fn();
const mockSoundValue = {
  playSound: mockPlaySound,
  muted: false,
  toggleMute: vi.fn(),
  claimExecSound: vi.fn(),
  consumeExecClaim: () => false,
};
vi.mock('../providers/SoundProvider', () => ({
  SoundProvider: ({ children }: { children: ReactNode }) => children,
  useSound: () => mockSoundValue,
  // AnimatedCard AND the central taps (useGameApi / GamePageShell / ErrorAlert)
  // consume useOptionalSound; route it to the same spy and assert on specific
  // sound names so per-card deal sounds don't interfere.
  useOptionalSound: () => mockSoundValue,
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(bakersDozenApi.exec);

function makeTableau(cols: BakersDozenTableauCard[][]): BakersDozenTableauCard[][] {
  const result: BakersDozenTableauCard[][] = [];
  for (let i = 0; i < 13; i++) {
    result.push(cols[i] ?? []);
  }
  return result;
}

const card = (design: CardDesign, value: number): Card => ({ design, value });

const playingState: BakersDozenResponse = {
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
    [],
    [],
    [],
    [],
    [],
  ]),
  foundation: [[], [], [], []],
  phase: 0,
  moveCount: 3,
  canUndo: false,
  isStalemate: false,
  message: '',
};

const gameClearState: BakersDozenResponse = {
  ...playingState,
  phase: 1,
  message: 'ゲームクリア！',
  messageCode: 'bakersdozen.gameClear',
  messageParams: { moveCount: '42' },
};

const gameOverState: BakersDozenResponse = {
  ...playingState,
  phase: 2,
  message: 'ゲームオーバー',
  messageCode: 'bakersdozen.gameOver',
};

describe('BakersDozenPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockPlaySound.mockReset();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  });

  it('calls reset on initial render', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BakersDozenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
  });

  it('marks the last card in a column with a dashed warning ring', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BakersDozenPage />);
    const lastCardBtn = await screen.findByTestId('bd-last-card-1');
    expect(lastCardBtn.className).toContain('ring-dashed');
    // Column 0 has 2 cards so it's NOT the last-card scenario
    expect(screen.queryByTestId('bd-last-card-0')).not.toBeInTheDocument();
  });

  it('persistently marks single-card columns with a dashed warning ring and tooltip', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BakersDozenPage />);
    // Column 1 holds a single card → flagged before any selection.
    const oneCardCol = await screen.findByTestId('bd-onecard-col-1');
    expect(oneCardCol.className).toContain('ring-dashed');
    // The tooltip lives on the (interactive) card button, not the wrapper div.
    expect(screen.getByTestId('bd-last-card-1')).toHaveAttribute('title', 'この列は二度と使えなくなります');
    // Column 0 (2 cards) and empty columns are not flagged.
    expect(screen.queryByTestId('bd-onecard-col-0')).not.toBeInTheDocument();
    expect(screen.queryByTestId('bd-onecard-col-2')).not.toBeInTheDocument();
  });

  it('does not flag any card on a fresh deal with multi-card columns only', async () => {
    const fullState: BakersDozenResponse = {
      ...playingState,
      tableau: makeTableau([
        [
          { card: card('SPADE', 13), faceUp: true },
          { card: card('SPADE', 5), faceUp: true },
        ],
        [
          { card: card('HEART', 6), faceUp: true },
          { card: card('HEART', 7), faceUp: true },
        ],
      ]),
    };
    mockExec.mockResolvedValue(fullState);
    renderWithProviders(<BakersDozenPage />);
    await waitFor(() => expect(screen.getByAltText('♠ 5')).toBeInTheDocument());
    expect(screen.queryByTestId('bd-last-card-0')).not.toBeInTheDocument();
    expect(screen.queryByTestId('bd-last-card-1')).not.toBeInTheDocument();
  });

  it('renders heading and move count', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BakersDozenPage />);
    await waitFor(() => expect(screen.getByText(/ベーカーズ・ダズン/)).toBeInTheDocument());
    expect(screen.getByText(/手数: 3/)).toBeInTheDocument();
  });

  it('renders 4 foundation suits', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BakersDozenPage />);
    await waitFor(() => expect(screen.getAllByText('A').length).toBeGreaterThanOrEqual(4));
  });

  it('renders giveup button when playing', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BakersDozenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ギブアップ' })).toBeInTheDocument());
  });

  it('advertises the keyboard shortcuts on the control buttons', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BakersDozenPage />);
    const giveUp = await screen.findByRole('button', { name: 'ギブアップ' });
    expect(giveUp).toHaveAttribute('aria-keyshortcuts', 'g');
    expect(giveUp.querySelector('kbd')?.textContent).toBe('G');
    // The KbdBadge text is aria-hidden, so the hint button's accessible name stays clean.
    const hint = screen.getByRole('button', { name: 'ヒント' });
    expect(hint).toHaveAttribute('aria-keyshortcuts', 'h');
  });

  it('hides giveup button when game cleared', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<BakersDozenPage />);
    await waitFor(() => expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument());
  });

  it('giveup button opens a confirm dialog and only dispatches giveup after confirm', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BakersDozenPage />);
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

  it('shows phase name in header for game over', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<BakersDozenPage />);
    await waitFor(() => expect(screen.getAllByText(/ゲームオーバー/).length).toBeGreaterThan(0));
  });

  it('renders auto-complete button enabled while playing', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BakersDozenPage />);
    await waitFor(() => expect(screen.getByTestId('autocomplete-button')).toBeEnabled());
  });

  it('shows StalemateEscapeButton when stalemate flag is set', async () => {
    const stalemate: BakersDozenResponse = { ...playingState, isStalemate: true, undoToEscape: 2, canUndo: true };
    mockExec.mockResolvedValue(stalemate);
    renderWithProviders(<BakersDozenPage />);
    await waitFor(() => expect(screen.getByTestId('stalemate-escape-button')).toBeInTheDocument());
  });

  it('renders foundation pile top card', async () => {
    const withFoundation: BakersDozenResponse = {
      ...playingState,
      foundation: [[card('SPADE', 1), card('SPADE', 2)], [], [], []],
    };
    mockExec.mockResolvedValue(withFoundation);
    renderWithProviders(<BakersDozenPage />);
    await waitFor(() => {
      // Foundation top card aria-label uses suit + count (見出しは "♠ 組札 2枚")
      expect(screen.getByLabelText(/♠ 組札 2枚/)).toBeInTheDocument();
    });
  });

  it('renders hint banner when hint state is set', async () => {
    let resolveHint: ((res: BakersDozenResponse) => void) | undefined;
    mockExec.mockImplementation((cmd: string) => {
      if (cmd === 'hint') {
        return new Promise<BakersDozenResponse>((resolve) => {
          resolveHint = resolve;
        });
      }
      return Promise.resolve(playingState);
    });

    renderWithProviders(<BakersDozenPage />);
    const hintBtn = await screen.findByRole('button', { name: 'ヒント' });
    fireEvent.click(hintBtn);

    // Resolve the in-flight hint call with a hint payload.
    resolveHint?.({
      ...playingState,
      hint: { fromCol: 0, cardIndex: 1, toZone: 'tableau', toCol: 1 },
    });

    await waitFor(() => expect(screen.getByText(/ヒントがあります/)).toBeInTheDocument());
  });

  it('selecting a tableau card marks it as selected', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BakersDozenPage />);
    // Pick the top card of column 0 by its aria-label (cardAlt produces "♠ 5").
    const sourceBtn = await screen.findByRole('button', { name: '♠ 5' });
    fireEvent.click(sourceBtn);
    await waitFor(() => expect(sourceBtn).toHaveAttribute('aria-pressed', 'true'));
  });

  it('makes the tableau horizontally scrollable so 13 columns stay legible', async () => {
    mockExec.mockResolvedValue(playingState);
    const { container } = renderWithProviders(<BakersDozenPage />);
    await waitFor(() => expect(container.querySelector('[data-tutorial="bd-tableau"]')).toBeInTheDocument());
    expect(container.querySelector('[data-tutorial="bd-tableau"]')).toHaveClass('overflow-x-auto');
  });

  it('centers the selected tableau column in view on mobile', async () => {
    const originalScrollIntoView = Element.prototype.scrollIntoView;
    const originalWidth = window.innerWidth;
    const scrollIntoView = vi.fn();
    // jsdom lacks scrollIntoView; spy on it and force a mobile viewport so the effect runs.
    Element.prototype.scrollIntoView = scrollIntoView;
    window.innerWidth = 375;
    window.dispatchEvent(new Event('resize'));
    try {
      mockExec.mockResolvedValue(playingState);
      renderWithProviders(<BakersDozenPage />);
      const sourceBtn = await screen.findByRole('button', { name: '♠ 5' });
      fireEvent.click(sourceBtn);
      await waitFor(() => expect(scrollIntoView).toHaveBeenCalledWith(expect.objectContaining({ inline: 'center' })));
    } finally {
      Element.prototype.scrollIntoView = originalScrollIntoView;
      window.innerWidth = originalWidth;
      window.dispatchEvent(new Event('resize'));
    }
  });

  it('forwards reset commands to the API on initial load', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BakersDozenPage />);
    await waitFor(() => expect(mockExec.mock.calls.some((c) => c[0] === 'reset')).toBe(true));
  });

  it('plays the cardPlace sound when a move succeeds (moveCount advances)', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BakersDozenPage />);
    const source = await screen.findByRole('button', { name: '♠ 5' });
    fireEvent.click(source);
    mockPlaySound.mockClear();
    // The move resolves with an advanced moveCount, signalling a server-confirmed move.
    mockExec.mockResolvedValue({ ...playingState, moveCount: 4 });
    fireEvent.click(screen.getByRole('button', { name: '♥ 6' }));
    await waitFor(() => expect(mockPlaySound).toHaveBeenCalledWith('cardPlace'));
  });

  it('stays silent (no cardPlace) and buzzes when a move fails', async () => {
    mockExec.mockResolvedValueOnce(playingState); // initial reset succeeds
    renderWithProviders(<BakersDozenPage />);
    const source = await screen.findByRole('button', { name: '♠ 5' });
    fireEvent.click(source);
    mockPlaySound.mockClear();
    mockExec.mockRejectedValue(new Error('illegal move')); // move rejects → moveCount unchanged
    fireEvent.click(screen.getByRole('button', { name: '♥ 6' }));
    await waitFor(() => expect(mockPlaySound).toHaveBeenCalledWith('errorBuzz'));
    expect(mockPlaySound).not.toHaveBeenCalledWith('cardPlace');
  });

  it('plays the shuffle sound when starting a new game via reset', async () => {
    mockExec.mockResolvedValue(gameOverState);
    renderWithProviders(<BakersDozenPage />);
    const nextBtn = await screen.findByRole('button', { name: '次のゲーム' });
    mockPlaySound.mockClear();
    fireEvent.click(nextBtn);
    // The central tap plays after the reset exec resolves, so await it.
    await waitFor(() => expect(mockPlaySound).toHaveBeenCalledWith('shuffle'));
  });
});
