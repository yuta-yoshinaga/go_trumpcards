import { fireEvent, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { burracoApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BurracoPlayerData, BurracoResponse } from '../types/card';
import { BurracoPage } from './BurracoPage';

vi.mock('../api/gameApi', () => ({
  burracoApi: { exec: vi.fn() },
  actionLogApi: { burraco: vi.fn() },
}));
const mockExec = vi.mocked(burracoApi.exec);

const playSoundMock = vi.fn();
vi.mock('../providers/SoundProvider', async () => {
  const actual = await vi.importActual<typeof import('../providers/SoundProvider')>('../providers/SoundProvider');
  return {
    ...actual,
    useSound: () => ({ playSound: playSoundMock, muted: false, toggleMute: vi.fn() }),
  };
});

const basePlayers: BurracoPlayerData[] = [
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
    hasBurraco: false,
    hasInitMeld: false,
    tookPozzetto: false,
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
    hasBurraco: false,
    hasInitMeld: false,
    tookPozzetto: false,
  },
];

const drawPhaseState: BurracoResponse = {
  players: basePlayers,
  phase: 0,
  roundNumber: 1,
  currentPlayerIdx: 0,
  discardTop: { design: 'SPADE', value: 5 },
  discardPile: [{ design: 'SPADE', value: 5 }],
  drawPileCount: 67,
  discardPileCount: 1,
  pozzettoCount: 2,
  isFrozen: false,
  gameEndFlag: false,
  winnerIdx: -1,
  message: '',
  messageCode: 'burraco.drawPhase',
  config: { cpuDifficulty: 1, pointLimit: 5000 },
};

const meldPhaseState: BurracoResponse = {
  ...drawPhaseState,
  phase: 1,
  messageCode: 'burraco.meldPhase',
};

const discardPhaseState: BurracoResponse = {
  ...drawPhaseState,
  phase: 2,
  messageCode: 'burraco.discardPhase',
};

const roundEndState: BurracoResponse = {
  ...drawPhaseState,
  phase: 3,
  messageCode: 'burraco.roundEnd',
};

const gameEndState: BurracoResponse = {
  ...drawPhaseState,
  phase: 4,
  gameEndFlag: true,
  winnerIdx: 0,
};

describe('BurracoPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(drawPhaseState);
  });

  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<BurracoPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<BurracoPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        cpuDifficulty: 1,
        pointLimit: 5000,
      }),
    );
  });

  it('shows draw phase buttons', async () => {
    renderWithProviders(<BurracoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeInTheDocument();
  });

  it('renders the discard pile viewer with all pile cards', async () => {
    mockExec.mockResolvedValue({
      ...drawPhaseState,
      discardPile: [
        { design: 'SPADE', value: 3 },
        { design: 'HEART', value: 7 },
        { design: 'CLOVER', value: 10 },
      ],
      discardPileCount: 3,
    });
    renderWithProviders(<BurracoPage />);
    const viewer = await screen.findByTestId('ca-discard-pile-viewer');
    expect(viewer).toBeInTheDocument();
    expect(screen.getByTestId('ca-discard-pile-cards').querySelectorAll('img')).toHaveLength(3);
  });

  it('shows an empty message when the discard pile is empty', async () => {
    mockExec.mockResolvedValue({ ...drawPhaseState, discardTop: null, discardPile: [], discardPileCount: 0 });
    renderWithProviders(<BurracoPage />);
    await screen.findByTestId('ca-discard-pile-viewer');
    expect(screen.getByText('捨て札の山は空です。')).toBeInTheDocument();
    expect(screen.queryByTestId('ca-discard-pile-cards')).not.toBeInTheDocument();
  });

  it('calls drawstock command when button clicked', async () => {
    renderWithProviders(<BurracoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(meldPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '山札から引く' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
  });

  it('shows meld phase buttons', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<BurracoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'メルドする' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'スキップ' })).toBeInTheDocument();
  });

  it('calls skipmeld command when skip button clicked', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<BurracoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'スキップ' })).toBeInTheDocument());
    mockExec.mockClear();
    mockExec.mockResolvedValue(discardPhaseState);
    fireEvent.click(screen.getByRole('button', { name: 'スキップ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('skipmeld'));
  });

  it('shows discard phase buttons', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<BurracoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨てる' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '上がる' })).toBeInTheDocument();
  });

  it('shows next round button at round end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<BurracoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
  });

  it('shows win celebration at game end', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<BurracoPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
  });

  it('renders hint toggle checkbox', async () => {
    renderWithProviders(<BurracoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '山札から引く' })).toBeInTheDocument());
    expect(screen.getByRole('checkbox', { name: 'ヒント表示' })).toBeInTheDocument();
  });

  it('shows HintTooltip when hint is enabled in draw phase', async () => {
    localStorage.setItem('hint_enabled_burraco', 'true');
    // drawPhaseState: human turn (currentPlayerIdx=0), DRAW phase → returns drawStock hint
    renderWithProviders(<BurracoPage />);
    await waitFor(() => expect(screen.getByTestId('hint-tooltip')).toBeInTheDocument());
  });

  // **ヒントは「どの札か」まで持っている** (#5990)。文言だけ出して探させない。
  it('rings the hand cards the hint points at', async () => {
    localStorage.setItem('hint_enabled_burraco', 'true');
    mockExec.mockResolvedValue({
      ...meldPhaseState,
      hint: { action: 'meld', reason: 'meldInitial', indices: [0, 2] },
    });
    renderWithProviders(<BurracoPage />);

    await waitFor(() => expect(screen.getByTestId('bu-hand-card-0')).toHaveAttribute('data-hint-card', 'true'));
    expect(screen.getByTestId('bu-hand-card-2')).toHaveAttribute('data-hint-card', 'true');
    expect(screen.getByTestId('bu-hand-card-1')).not.toHaveAttribute('data-hint-card');
    // **印だけでなく見た目も付く。** data 属性しか見ないと、リングを外しても通る。
    expect(screen.getByTestId('bu-hand-card-0').style.outline).toContain('var(--color-ds-warning)');
    expect(screen.getByTestId('bu-hand-card-1').style.outline).toBe('');
  });

  it('rings nothing while the hint is off', async () => {
    localStorage.setItem('hint_enabled_burraco', 'false');
    mockExec.mockResolvedValue({
      ...meldPhaseState,
      hint: { action: 'meld', reason: 'meldInitial', indices: [0, 2] },
    });
    renderWithProviders(<BurracoPage />);

    await waitFor(() => expect(screen.getByTestId('bu-hand-card-0')).toBeInTheDocument());
    expect(screen.getByTestId('bu-hand-card-0')).not.toHaveAttribute('data-hint-card');
  });

  // **サーバがヒントを返さない場面 (CPU の手番など) では付かない。**
  it('rings nothing when the server sent no hint', async () => {
    localStorage.setItem('hint_enabled_burraco', 'true');
    mockExec.mockResolvedValue({ ...meldPhaseState, hint: undefined });
    renderWithProviders(<BurracoPage />);

    await waitFor(() => expect(screen.getByTestId('bu-hand-card-0')).toBeInTheDocument());
    expect(screen.getByTestId('bu-hand-card-0')).not.toHaveAttribute('data-hint-card');
  });

  it('shows 次のゲーム at game-end and fires reset directly (no confirm)', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<BurracoPage />);
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
    renderWithProviders(<BurracoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeInTheDocument());
    const reason = screen.getByTestId('ca-draw-discard-reason');
    expect(reason).toHaveTextContent('手札からトップカードと同ランクの2枚を選択してください');
  });

  it('renders the frozen badge and reason when the discard pile is frozen', async () => {
    mockExec.mockResolvedValue({ ...drawPhaseState, isFrozen: true });
    renderWithProviders(<BurracoPage />);
    await waitFor(() => expect(screen.getByTestId('ca-frozen-badge')).toBeInTheDocument());
    expect(screen.getByTestId('ca-draw-discard-reason')).toHaveTextContent(
      'フリーズ中はワイルドカードでの代用ができません',
    );
    // Pre-emptive freeze guide shown at the top of the draw controls.
    expect(screen.getByTestId('ca-draw-freeze-guide')).toHaveTextContent(/フリーズ中/);
  });

  it('switches the reason to selectOneMore once the player selects one card', async () => {
    renderWithProviders(<BurracoPage />);
    await waitFor(() => expect(screen.getByTestId('ca-draw-discard-reason')).toBeInTheDocument());
    const handCards = screen.getAllByTestId(/^bu-hand-card-/);
    fireEvent.click(handCards[0]);
    expect(screen.getByTestId('ca-draw-discard-reason')).toHaveTextContent('もう1枚選択してください');
  });

  it('clears the reason and enables the draw button for a legal natural pair', async () => {
    // Two natural cards of the discard top's rank; the fixture hand opens with sevens.
    mockExec.mockResolvedValue({ ...drawPhaseState, discardTop: { design: 'SPADE', value: 7 } });
    renderWithProviders(<BurracoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeInTheDocument());
    const handCards = screen.getAllByTestId(/^bu-hand-card-/);
    fireEvent.click(handCards[0]);
    fireEvent.click(handCards[1]);
    expect(screen.queryByTestId('ca-draw-discard-reason')).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: '捨て札を取る' })).not.toBeDisabled();
    expect(screen.getByRole('button', { name: '捨て札を取る' })).not.toHaveAttribute('aria-describedby');
  });

  it('warns when more than 2 cards are selected', async () => {
    renderWithProviders(<BurracoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeInTheDocument());
    const handCards = screen.getAllByTestId(/^bu-hand-card-/);
    fireEvent.click(handCards[0]);
    fireEvent.click(handCards[1]);
    fireEvent.click(handCards[2]);
    expect(screen.getByTestId('ca-draw-discard-reason')).toHaveTextContent('選択は2枚までです');
    expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeDisabled();
  });

  it('keeps the frozen warning visible while the player has only picked one card', async () => {
    mockExec.mockResolvedValue({ ...drawPhaseState, isFrozen: true });
    renderWithProviders(<BurracoPage />);
    await waitFor(() => expect(screen.getByTestId('ca-frozen-badge')).toBeInTheDocument());
    const handCards = screen.getAllByTestId(/^bu-hand-card-/);
    fireEvent.click(handCards[0]);
    expect(screen.getByTestId('ca-draw-discard-reason')).toHaveTextContent(
      'フリーズ中はワイルドカードでの代用ができません',
    );
  });

  it('shows a banner and plays a sound when a player takes the pozzetto', async () => {
    mockExec.mockResolvedValue(drawPhaseState); // tookPozzetto false for everyone
    renderWithProviders(<BurracoPage />);
    await waitFor(() => expect(screen.getByTestId('bu-pozzetto-count')).toBeInTheDocument());
    expect(screen.queryByTestId('bu-pozzetto-banner')).not.toBeInTheDocument();

    // The next fetch reports the human took the pozzetto (false → true).
    mockExec.mockResolvedValue({
      ...meldPhaseState,
      players: [{ ...basePlayers[0], tookPozzetto: true }, basePlayers[1]],
    });
    fireEvent.click(screen.getByRole('button', { name: '山札から引く' }));
    await waitFor(() => expect(screen.getByTestId('bu-pozzetto-banner')).toBeInTheDocument());
    expect(playSoundMock).toHaveBeenCalledWith('chipClick');
  });

  it('names the CPU in the banner when a CPU takes the pozzetto', async () => {
    mockExec.mockResolvedValue(drawPhaseState);
    renderWithProviders(<BurracoPage />);
    await waitFor(() => expect(screen.getByTestId('bu-pozzetto-count')).toBeInTheDocument());

    mockExec.mockResolvedValue({
      ...meldPhaseState,
      players: [basePlayers[0], { ...basePlayers[1], tookPozzetto: true }],
    });
    fireEvent.click(screen.getByRole('button', { name: '山札から引く' }));
    // ja CPU label is "CPU 1" → banner interpolates the CPU name, not "あなた".
    await waitFor(() => expect(screen.getByTestId('bu-pozzetto-banner')).toHaveTextContent('CPU 1'));
    expect(playSoundMock).toHaveBeenCalledWith('chipClick');
  });

  it('pulses a round-score cell when that score changes', async () => {
    mockExec.mockResolvedValue(drawPhaseState); // roundScore 0
    renderWithProviders(<BurracoPage />);
    await waitFor(() => expect(screen.getByTestId('bu-round-score-0')).toBeInTheDocument());
    expect(screen.getByTestId('bu-round-score-0')).not.toHaveClass('motion-safe:animate-pulse');

    mockExec.mockResolvedValue({
      ...roundEndState,
      players: [{ ...basePlayers[0], roundScore: 120 }, basePlayers[1]],
    });
    fireEvent.click(screen.getByRole('button', { name: '山札から引く' }));
    await waitFor(() => expect(screen.getByTestId('bu-round-score-0')).toHaveClass('motion-safe:animate-pulse'));
  });

  it('reorders the displayed hand when the suit-sort toggle is pressed', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<BurracoPage />);
    await waitFor(() => expect(screen.getByTestId('bu-hand-card-0')).toBeInTheDocument());

    // Original server order: SPADE7(0) CLOVER7(1) HEART7(2) SPADE10(3) CLOVER10(4).
    const originalOrder = screen.getAllByTestId(/^bu-hand-card-/).map((b) => b.getAttribute('data-testid'));
    expect(originalOrder).toEqual([
      'bu-hand-card-0',
      'bu-hand-card-1',
      'bu-hand-card-2',
      'bu-hand-card-3',
      'bu-hand-card-4',
    ]);

    fireEvent.click(screen.getByTestId('bu-sort-suit'));
    expect(screen.getByTestId('bu-sort-suit')).toHaveAttribute('aria-pressed', 'true');

    // Suit-grouped ♠♥♦♣: SPADE7(0) SPADE10(3) HEART7(2) CLOVER7(1) CLOVER10(4).
    const sortedOrder = screen.getAllByTestId(/^bu-hand-card-/).map((b) => b.getAttribute('data-testid'));
    expect(sortedOrder).toEqual([
      'bu-hand-card-0',
      'bu-hand-card-3',
      'bu-hand-card-2',
      'bu-hand-card-1',
      'bu-hand-card-4',
    ]);
  });

  it('melds the correct original indices after the hand is display-sorted', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<BurracoPage />);
    await waitFor(() => expect(screen.getByTestId('bu-hand-card-0')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('bu-sort-suit'));

    // In suit order the three 7s sit at DOM positions 0, 2 and 3. Select them by
    // DOM position and confirm the meld targets their ORIGINAL indices (0,2,1),
    // proving the sort only reorders the display, never the action indices.
    const handButtons = screen.getAllByTestId(/^bu-hand-card-/);
    fireEvent.click(handButtons[0]); // SPADE7  -> index 0
    fireEvent.click(handButtons[2]); // HEART7  -> index 2
    fireEvent.click(handButtons[3]); // CLOVER7 -> index 1

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'メルドする' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('meld', undefined, undefined, undefined, [[0, 2, 1]]));
  });

  it('persists the chosen sort mode to localStorage', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<BurracoPage />);
    await waitFor(() => expect(screen.getByTestId('bu-sort-rank')).toBeInTheDocument());

    fireEvent.click(screen.getByTestId('bu-sort-rank'));
    expect(localStorage.getItem('burraco-sort-mode')).toBe('rank');
  });

  afterEach(() => {
    localStorage.clear();
  });

  it('refuses a two-card selection the shared Canasta rules reject', async () => {
    localStorage.clear();
    mockExec.mockReset();
    // Burraco runs on the Canasta domain, so the same pair rules apply here.
    mockExec.mockResolvedValue(drawPhaseState);
    renderWithProviders(<BurracoPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeInTheDocument());
    const handCards = screen.getAllByTestId(/^bu-hand-card-/);
    fireEvent.click(handCards[0]);
    fireEvent.click(handCards[1]);
    expect(screen.getByRole('button', { name: '捨て札を取る' })).toBeDisabled();
    // Pin the reason too, so this cannot pass merely because nothing got selected.
    expect(screen.getByTestId('ca-draw-discard-reason')).toHaveTextContent('同じランク');
  });
});
