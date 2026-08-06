import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { schnapsenApi } from '../api/gameApi';
import { useGameHint } from '../hooks/useGameHint';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, SchnapsenResponse } from '../types/card';
import { SchnapsenPage } from './SchnapsenPage';

vi.mock('../api/gameApi', () => ({
  schnapsenApi: { exec: vi.fn() },
  actionLogApi: { schnapsen: vi.fn() },
}));

vi.mock('../hooks/useGameHint', () => ({
  useGameHint: vi.fn(() => ({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() })),
}));

const mockExec = vi.mocked(schnapsenApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function makeState(overrides: Partial<SchnapsenResponse> = {}): SchnapsenResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 5,
        cards: [card('SPADE', 1), card('HEART', 10), card('DIAMOND', 11), card('CLOVER', 13), card('CLOVER', 12)],
        points: 0,
        trickCount: 0,
      },
      { id: 1, isHuman: false, cardCount: 5, cards: [], points: 0, trickCount: 0 },
    ],
    phase: 0,
    trickNumber: 1,
    currentPlayerIdx: 0,
    currentTrick: [],
    trumpSuit: 1,
    trumpCard: card('SPADE', 13),
    dealerIdx: 0,
    leadPlayerIdx: 0,
    stockRemaining: 9,
    isEndgame: false,
    validPlays: [0, 1, 2, 3, 4],
    marriagePlays: [],
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    config: { cpuDifficulty: 0 },
    ...overrides,
  };
}

beforeEach(() => {
  mockExec.mockResolvedValue(makeState());
  vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled: vi.fn() });
});

describe('SchnapsenPage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<SchnapsenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders header info (trick, stock, points, phase)', async () => {
    renderWithProviders(<SchnapsenPage />);
    await waitFor(() => expect(screen.getByText(/トリック: 1/)).toBeInTheDocument());
    expect(screen.getByText(/山札: 9/)).toBeInTheDocument();
    expect(screen.getByTestId('schnapsen-phase')).toHaveTextContent(/第1フェーズ/);
  });

  it('shows the human hand as play buttons', async () => {
    renderWithProviders(<SchnapsenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '♠ A を出す' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '♥ 10 を出す' })).toBeInTheDocument();
  });

  it('fires play with the selected card index when a card is clicked', async () => {
    renderWithProviders(<SchnapsenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '♥ 10 を出す' })).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '♥ 10 を出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 1));
  });

  it('shows a marriage button and dispatches marriage when present', async () => {
    mockExec.mockResolvedValue(makeState({ marriagePlays: [3, 4] })); // ♣ marriage, trump ♠ → 20 pts
    renderWithProviders(<SchnapsenPage />);
    const m3 = await screen.findByTestId('schnapsen-marriage-3');
    expect(m3.getAttribute('aria-label')).toContain('♣');
    expect(m3.getAttribute('aria-label')).toContain('20');
    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('schnapsen-marriage-4'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('marriage', 4));
  });

  it('labels a trump-suit marriage as 40 points', async () => {
    // trump ♣ (2): the ♣ K/Q marriage at idx 3/4 is now worth 40.
    mockExec.mockResolvedValue(makeState({ marriagePlays: [3, 4], trumpSuit: 2 }));
    renderWithProviders(<SchnapsenPage />);
    const m3 = await screen.findByTestId('schnapsen-marriage-3');
    expect(m3.getAttribute('aria-label')).toContain('40');
  });

  it('rings legal cards but keeps illegal cards clickable in the endgame (phase 2)', async () => {
    mockExec.mockResolvedValue(makeState({ isEndgame: true, validPlays: [0] }));
    renderWithProviders(<SchnapsenPage />);
    await waitFor(() => expect(screen.getByTestId('schnapsen-phase')).toHaveTextContent(/第2フェーズ/));
    const legal = screen.getByRole('button', { name: '♠ A を出す' });
    const illegal = screen.getByRole('button', { name: '♥ 10 を出す' });
    // Legal card gets an additive success ring; illegal card does not.
    expect(legal.className).toContain('ring-ds-success');
    expect(illegal.className).not.toContain('ring-ds-success');
    // Illegal card stays clickable (no hard block) — backend still validates.
    expect(illegal).not.toBeDisabled();
    mockExec.mockClear();
    fireEvent.click(illegal);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 1));
  });

  it('shows the must-follow guide banner in phase 2 on the human turn', async () => {
    mockExec.mockResolvedValue(makeState({ isEndgame: true, validPlays: [0] }));
    renderWithProviders(<SchnapsenPage />);
    const banner = await screen.findByTestId('schnapsen-endgame-guide');
    expect(banner).toHaveTextContent(/第2フェーズ/);
  });

  it('shows no guide banner or legal ring in phase 1 (no visual regression)', async () => {
    renderWithProviders(<SchnapsenPage />);
    const btn = await screen.findByRole('button', { name: '♠ A を出す' });
    expect(btn.className).not.toContain('ring-ds-success');
    expect(screen.queryByTestId('schnapsen-endgame-guide')).not.toBeInTheDocument();
  });

  it('shows "Next trick" button on trick-end and dispatches next', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 1 }));
    renderWithProviders(<SchnapsenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリックへ' })).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のトリックへ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('shows youWin banner when winnerIdx is 0', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 2,
        gameEndFlag: true,
        winnerIdx: 0,
        players: [
          { id: 0, isHuman: true, cardCount: 0, cards: [], points: 70, trickCount: 6 },
          { id: 1, isHuman: false, cardCount: 0, cards: [], points: 40, trickCount: 4 },
        ],
      }),
    );
    renderWithProviders(<SchnapsenPage />);
    await waitFor(() => expect(screen.getByText(/あなたの勝ち！.*70.*40/)).toBeInTheDocument());
  });

  it('shows cpuWin banner when winnerIdx is 1', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 2,
        gameEndFlag: true,
        winnerIdx: 1,
        players: [
          { id: 0, isHuman: true, cardCount: 0, cards: [], points: 40, trickCount: 4 },
          { id: 1, isHuman: false, cardCount: 0, cards: [], points: 70, trickCount: 6 },
        ],
      }),
    );
    renderWithProviders(<SchnapsenPage />);
    await waitFor(() => expect(screen.getByText(/CPUの勝ち/)).toBeInTheDocument());
  });

  // **フェーズ2 はマストフォローの厳格ルール。**現物が手札に入っても切り札スートは
  // 出し続ける。汎用の「切り札: 山札なし」だけだと、合法手のリングの意味が
  // 読めなくなっていた (#4810)。
  it('keeps showing the trump suit once the trump card is gone', async () => {
    mockExec.mockResolvedValue(makeState({ trumpCard: undefined, trumpSuit: 3, stockRemaining: 0, isEndgame: true }));
    renderWithProviders(<SchnapsenPage />);
    await waitFor(() => expect(screen.getByText(/切り札: ♥/)).toBeInTheDocument());
    expect(screen.queryByText('切り札: 山札なし')).not.toBeInTheDocument();
  });

  it('disables play buttons when it is not the human turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentPlayerIdx: 1 }));
    renderWithProviders(<SchnapsenPage />);
    await waitFor(() => {
      const btn = screen.getByRole('button', { name: '♠ A を出す' });
      expect(btn).toBeDisabled();
    });
  });

  it('shows confirm dialog on reset click and runs reset on accept', async () => {
    renderWithProviders(<SchnapsenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '確認' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('toggles the frontend hint from the settings panel', async () => {
    const setHintEnabled = vi.fn();
    vi.mocked(useGameHint).mockReturnValue({ hint: null, hintEnabled: false, setHintEnabled });
    renderWithProviders(<SchnapsenPage />);
    await waitFor(() => expect(screen.getByText('設定')).toBeInTheDocument());
    fireEvent.click(screen.getByText('設定')); // open the collapsed settings panel
    fireEvent.click(screen.getByLabelText('ヒント表示'));
    expect(setHintEnabled).toHaveBeenCalledWith(true);
  });

  it('renders the HintTooltip when the hint is enabled and available', async () => {
    vi.mocked(useGameHint).mockReturnValue({
      hint: { targetAction: 'play-0', reason: 'frontendHint.followSuit', confidence: 'moderate' },
      hintEnabled: true,
      setHintEnabled: vi.fn(),
    });
    renderWithProviders(<SchnapsenPage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  it('renders the tutorial anchor elements the guided tour targets', async () => {
    const { container } = renderWithProviders(<SchnapsenPage />);
    await waitFor(() => expect(screen.getByTestId('schnapsen-phase')).toBeInTheDocument());
    for (const anchor of ['schnapsen-trump', 'schnapsen-trick', 'schnapsen-hand', 'schnapsen-actions']) {
      expect(container.querySelector(`[data-tutorial="${anchor}"]`)).toBeInTheDocument();
    }
  });
});
