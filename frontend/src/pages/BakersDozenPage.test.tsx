import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { bakersDozenApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BakersDozenResponse, BakersDozenTableauCard, Card, CardDesign } from '../types/card';
import { BakersDozenPage } from './BakersDozenPage';

vi.mock('../api/gameApi', () => ({
  bakersDozenApi: { exec: vi.fn() },
  actionLogApi: { bakersdozen: vi.fn() },
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
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
  });

  it('calls reset on initial render', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BakersDozenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec.mock.calls[0]?.[0]).toBe('reset');
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

  it('hides giveup button when game cleared', async () => {
    mockExec.mockResolvedValue(gameClearState);
    renderWithProviders(<BakersDozenPage />);
    await waitFor(() => expect(screen.queryByRole('button', { name: 'ギブアップ' })).not.toBeInTheDocument());
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

  it('forwards reset commands to the API on initial load', async () => {
    mockExec.mockResolvedValue(playingState);
    renderWithProviders(<BakersDozenPage />);
    await waitFor(() => expect(mockExec.mock.calls.some((c) => c[0] === 'reset')).toBe(true));
  });
});
