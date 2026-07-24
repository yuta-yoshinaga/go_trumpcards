import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { canastaApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { CanastaPlayerData, CanastaResponse } from '../types/card';
import { CanastaPage } from './CanastaPage';

vi.mock('../api/gameApi', () => ({
  canastaApi: { exec: vi.fn() },
  actionLogApi: { canasta: vi.fn() },
}));
const mockExec = vi.mocked(canastaApi.exec);

const basePlayers: CanastaPlayerData[] = [
  {
    id: 0,
    isHuman: true,
    cardCount: 15,
    cards: [
      { design: 'SPADE', value: 7 },
      { design: 'CLOVER', value: 7 },
      { design: 'HEART', value: 7 },
      { design: 'SPADE', value: 10 },
      { design: 'CLOVER', value: 10 },
    ],
    melds: [],
    red3Count: 0,
    red3s: [],
    roundScore: 0,
    cumulativeScore: 0,
    hasCanasta: false,
    hasInitMeld: false,
  },
  {
    id: 1,
    isHuman: false,
    cardCount: 15,
    cards: [],
    melds: [],
    red3Count: 0,
    red3s: [],
    roundScore: 0,
    cumulativeScore: 0,
    hasCanasta: false,
    hasInitMeld: false,
  },
];

const drawPhaseState: CanastaResponse = {
  players: basePlayers,
  phase: 0,
  roundNumber: 1,
  currentPlayerIdx: 0,
  discardTop: { design: 'SPADE', value: 5 },
  drawPileCount: 67,
  discardPileCount: 1,
  isFrozen: false,
  gameEndFlag: false,
  winnerIdx: -1,
  message: '',
  messageCode: 'canasta.drawPhase',
  config: { cpuDifficulty: 1, pointLimit: 5000 },
};

const meldPhaseState: CanastaResponse = {
  ...drawPhaseState,
  phase: 1,
  messageCode: 'canasta.meldPhase',
};

const discardPhaseState: CanastaResponse = {
  ...drawPhaseState,
  phase: 2,
  messageCode: 'canasta.discardPhase',
};

const roundEndState: CanastaResponse = {
  ...drawPhaseState,
  phase: 3,
  messageCode: 'canasta.roundEnd',
};

const gameEndState: CanastaResponse = {
  ...drawPhaseState,
  phase: 4,
  gameEndFlag: true,
  winnerIdx: 0,
};

// Human holds one completed natural canasta (7 cards) and one incomplete meld (3 cards)
// so the compact stack + count badge rendering can be exercised.
const meldDisplayState: CanastaResponse = {
  ...meldPhaseState,
  players: [
    {
      ...basePlayers[0],
      hasCanasta: true,
      melds: [
        {
          rank: 7,
          isNatural: true,
          isCanasta: true,
          cards: Array.from({ length: 7 }, (_, i) => ({ design: i % 2 ? 'HEART' : 'SPADE', value: 7 })),
        },
        {
          rank: 10,
          isNatural: false,
          isCanasta: false,
          cards: [
            { design: 'SPADE', value: 10 },
            { design: 'CLOVER', value: 10 },
            { design: 'HEART', value: 10 },
          ],
        },
      ],
    },
    basePlayers[1],
  ],
};

describe('CanastaPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(drawPhaseState);
  });

  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<CanastaPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<CanastaPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        cpuDifficulty: 1,
        pointLimit: 5000,
      }),
    );
  });

  it('shows draw phase buttons', async () => {
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeInTheDocument();
  });

  it('calls drawstock command when button clicked', async () => {
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(meldPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '山札から引く' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
  });

  it('shows meld phase buttons', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'メルドする' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'スキップ' })).toBeInTheDocument();
  });

  it('shows the initial-meld minimum and selected total in the meld phase', async () => {
    mockExec.mockResolvedValue(meldPhaseState); // score 0 → min 50; hasInitMeld false
    renderWithProviders(<CanastaPage />);
    const info = await screen.findByTestId('ca-meld-points');
    expect(info).toHaveTextContent('初回メルド最低点: 50');
    expect(info).toHaveTextContent('選択合計: 0');
  });

  it('shows only the selected total once the initial meld is made', async () => {
    mockExec.mockResolvedValue({
      ...meldPhaseState,
      players: [{ ...basePlayers[0], hasInitMeld: true }, basePlayers[1]],
    });
    renderWithProviders(<CanastaPage />);
    const info = await screen.findByTestId('ca-meld-points');
    expect(info).toHaveTextContent('選択合計: 0');
    expect(info).not.toHaveTextContent('初回メルド最低点');
  });

  it('calls skipmeld command when skip button clicked', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'スキップ' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(discardPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('skipmeld'));
  });

  it('shows discard phase buttons', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨てる' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '上がる' })).toBeInTheDocument();
  });

  it('disables go out and shows the reason while the player has no canasta', async () => {
    mockExec.mockResolvedValue(discardPhaseState); // hasCanasta: false
    renderWithProviders(<CanastaPage />);
    const goOut = await screen.findByTestId('ca-go-out-button');
    expect(goOut).toBeDisabled();
    const reason = screen.getByTestId('ca-go-out-reason');
    expect(reason).toHaveTextContent('ゴーアウトするには完成したカナスタ（7枚メルド）が1組以上必要です');
    expect(goOut).toHaveAttribute('aria-describedby', 'ca-go-out-reason');
  });

  it('enables go out and hides the reason once a canasta is completed', async () => {
    mockExec.mockResolvedValue({
      ...discardPhaseState,
      players: [{ ...basePlayers[0], hasCanasta: true }, basePlayers[1]],
    });
    renderWithProviders(<CanastaPage />);
    const goOut = await screen.findByTestId('ca-go-out-button');
    expect(goOut).toBeEnabled();
    expect(screen.queryByTestId('ca-go-out-reason')).not.toBeInTheDocument();
    expect(goOut).not.toHaveAttribute('aria-describedby');
  });

  it('shows next round button at round end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
  });

  it('shows win celebration at game end', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
  });

  it('renders hint toggle checkbox', async () => {
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument());
    expect(screen.getByRole('checkbox', { name: 'ヒント表示' })).toBeInTheDocument();
  });

  it('shows HintTooltip when hint is enabled in draw phase', async () => {
    localStorage.setItem('hint_enabled_canasta', 'true');
    // drawPhaseState: human turn (currentPlayerIdx=0), DRAW phase → returns drawStock hint
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  it('shows 次のゲーム at game-end and fires reset directly (no confirm)', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        cpuDifficulty: 1,
        pointLimit: 5000,
      }),
    );
    expect(screen.queryByRole('button', { name: '確認' })).not.toBeInTheDocument();
  });

  it('shows the disabled reason for draw-from-discard when no cards are selected', async () => {
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeInTheDocument());
    const reason = screen.getByTestId('ca-draw-discard-reason');
    expect(reason).toHaveTextContent('手札からトップカードと同ランクの2枚を選択してください');
  });

  it('renders the frozen badge and reason when the discard pile is frozen', async () => {
    mockExec.mockResolvedValue({ ...drawPhaseState, isFrozen: true });
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(screen.getByTestId('ca-frozen-badge')).toBeInTheDocument());
    expect(screen.getByTestId('ca-draw-discard-reason')).toHaveTextContent(
      'フリーズ中はワイルドカードでの代用ができません',
    );
  });

  it('switches the reason to selectOneMore once the player selects one card', async () => {
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(screen.getByTestId('ca-draw-discard-reason')).toBeInTheDocument());
    const handCards = screen.getAllByRole('button', { pressed: false }).filter((b) => b.hasAttribute('aria-pressed'));
    fireEvent.click(handCards[0]);
    expect(screen.getByTestId('ca-draw-discard-reason')).toHaveTextContent('もう1枚選択してください');
  });

  it('clears the reason and enables the draw button when exactly 2 cards are selected', async () => {
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeInTheDocument());
    const handCards = screen.getAllByRole('button', { pressed: false }).filter((b) => b.hasAttribute('aria-pressed'));
    fireEvent.click(handCards[0]);
    fireEvent.click(handCards[1]);
    expect(screen.queryByTestId('ca-draw-discard-reason')).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: '捨て札を取る' })).not.toBeDisabled();
    expect(screen.getByRole('button', { name: '捨て札を取る' })).not.toHaveAttribute('aria-describedby');
  });

  it('warns when more than 2 cards are selected', async () => {
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeInTheDocument());
    const handCards = screen.getAllByRole('button', { pressed: false }).filter((b) => b.hasAttribute('aria-pressed'));
    fireEvent.click(handCards[0]);
    fireEvent.click(handCards[1]);
    fireEvent.click(handCards[2]);
    expect(screen.getByTestId('ca-draw-discard-reason')).toHaveTextContent('選択は2枚までです');
    expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeDisabled();
  });

  it('keeps the frozen warning visible while the player has only picked one card', async () => {
    mockExec.mockResolvedValue({ ...drawPhaseState, isFrozen: true });
    renderWithProviders(<CanastaPage />);
    await waitFor(() => expect(screen.getByTestId('ca-frozen-badge')).toBeInTheDocument());
    const handCards = screen.getAllByRole('button', { pressed: false }).filter((b) => b.hasAttribute('aria-pressed'));
    fireEvent.click(handCards[0]);
    expect(screen.getByTestId('ca-draw-discard-reason')).toHaveTextContent(
      'フリーズ中はワイルドカードでの代用ができません',
    );
  });

  it('renders melds compactly with a count badge and canasta label, collapsed by default', async () => {
    mockExec.mockResolvedValue(meldDisplayState);
    renderWithProviders(<CanastaPage />);

    const canastaMeld = await screen.findByTestId('ca-meld-0-0');
    // Collapsed by default so a meld occupies ~one card of height.
    expect(canastaMeld).not.toHaveAttribute('open');
    // Count badge shows the number of cards in the meld.
    expect(screen.getByTestId('ca-meld-badge-0-0')).toHaveTextContent('7');
    expect(screen.getByTestId('ca-meld-badge-0-1')).toHaveTextContent('3');
    // Canasta type (with the ★) is visible at a glance for the completed canasta.
    expect(within(canastaMeld).getByText(/ナチュラルカナスタ/)).toBeInTheDocument();
  });

  it('expands a meld to reveal every constituent card when the summary is clicked', async () => {
    mockExec.mockResolvedValue(meldDisplayState);
    renderWithProviders(<CanastaPage />);

    const canastaMeld = (await screen.findByTestId('ca-meld-0-0')) as HTMLDetailsElement;
    expect(canastaMeld.open).toBe(false);
    const summary = canastaMeld.querySelector('summary');
    if (!summary) throw new Error('meld summary not found');
    fireEvent.click(summary);
    expect(canastaMeld.open).toBe(true);
  });

  afterEach(() => {
    localStorage.clear();
  });
});
