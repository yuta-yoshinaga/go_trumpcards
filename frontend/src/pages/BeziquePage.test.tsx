import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { beziqueApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeBeziqueState } from '../test/stateFactories';
import { BeziquePage } from './BeziquePage';

vi.mock('../api/gameApi', () => ({
  beziqueApi: { exec: vi.fn() },
  actionLogApi: { bezique: vi.fn() },
}));

const mockExec = vi.mocked(beziqueApi.exec);

// Default fixture: a human Play turn (seat 0, phase 1=Play in domain terms = 0 here).
const playPhaseState = makeBeziqueState({ phase: 0, currentPlayerIdx: 0 });
// A CPU turn (currentPlayerIdx is the CPU).
const cpuTurnState = makeBeziqueState({ phase: 0, currentPlayerIdx: 1 });
// A human Meld turn with two declarable melds.
const meldPhaseState = makeBeziqueState({
  phase: 1,
  currentPlayerIdx: 0,
  availableMelds: [
    { type: 0, suit: 1, points: 20 },
    { type: 1, suit: -1, points: 40 },
  ],
});
const roundEndState = makeBeziqueState({
  phase: 2,
  dealPoints: [240, 160],
});
const gameEndState = makeBeziqueState({
  phase: 3,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'ゲーム終了！ あなたの勝利です (1010-820)！',
});
const endgameState = makeBeziqueState({ phase: 0, currentPlayerIdx: 0, stockRemaining: 0, isEndgame: true });
// Endgame human FOLLOW turn: opponent led ♠K; hand is ♠Q / ♦J / ♥A, trump ♥.
// Holding ♠ forces following ♠ (♠Q cannot beat ♠K, so it is the only legal card).
const endgameFollowState = makeBeziqueState({
  phase: 0,
  currentPlayerIdx: 0,
  stockRemaining: 0,
  isEndgame: true,
  trumpSuit: 3,
  currentTrick: [{ playerIdx: 1, card: { design: 'SPADE', value: 13 } }],
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('BeziquePage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<BeziquePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<BeziquePage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, targetScore: 1000 },
      }),
    );
  });

  it('shows the deal-points trick/meld breakdown in the score sidebar', async () => {
    mockExec.mockResolvedValue(
      makeBeziqueState({ phase: 0, currentPlayerIdx: 0, dealPoints: [10, 0], dealMeldPoints: [4, 0] }),
    );
    renderWithProviders(<BeziquePage />);
    const breakdown = await screen.findByTestId('bezique-deal-breakdown-0');
    // trick = deal(10) - meld(4) = 6
    expect(breakdown).toHaveTextContent('6');
    expect(breakdown).toHaveTextContent('4');
  });

  it('renders the play phase with the human cards and the play button', async () => {
    renderWithProviders(<BeziquePage />);
    await waitFor(() => {
      expect(screen.getByAltText('♠ Q')).toBeInTheDocument();
      expect(screen.getByAltText('♦ J')).toBeInTheDocument();
    });
    expect(screen.getByRole('button', { name: '出す' })).toBeInTheDocument();
  });

  it('dispatches play when a card is selected and the play button clicked', async () => {
    renderWithProviders(<BeziquePage />);
    const card = await screen.findByAltText('♠ Q');
    fireEvent.click(card);
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<BeziquePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('shows meld buttons and a skip button on a human meld turn', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<BeziquePage />);
    await waitFor(() => expect(screen.getByTestId('meld-0')).toBeInTheDocument());
    expect(screen.getByTestId('meld-1')).toBeInTheDocument();
    expect(screen.getByTestId('meld-skip')).toBeInTheDocument();
  });

  it('gives each meld button a suit-named aria-label inside a labelled group', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<BeziquePage />);
    // Marriage of ♠ (type 0, suit 1) reads the suit by name, not the glyph.
    const marriage = await screen.findByTestId('meld-0');
    expect(marriage).toHaveAttribute('aria-label', 'スペードの結婚 (K+Q)を宣言 +20点');
    // A non-suited meld (bezique) omits the suit.
    expect(screen.getByTestId('meld-1')).toHaveAttribute('aria-label', 'ベジーク (♠Q+♦J)を宣言 +40点');
    // The melds are bundled in a labelled group.
    const group = screen.getByRole('group', { name: '宣言するメルドを選んでください:' });
    expect(group).toBeInTheDocument();
    expect(group).toContainElement(marriage);
  });

  it('dispatches a meld declaration when a meld button is clicked', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<BeziquePage />);
    const meld1 = await screen.findByTestId('meld-1');
    mockExec.mockClear();
    mockExec.mockResolvedValue(meldPhaseState);
    fireEvent.click(meld1);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('meld', { meldIndex: 1 }));
  });

  it('dispatches skip when the skip-meld button is clicked', async () => {
    mockExec.mockResolvedValue(meldPhaseState);
    renderWithProviders(<BeziquePage />);
    const skip = await screen.findByTestId('meld-skip');
    mockExec.mockClear();
    mockExec.mockResolvedValue(meldPhaseState);
    fireEvent.click(skip);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('skip'));
  });

  it('renders round end with the next deal button and the deal result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<BeziquePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のディール' })).toBeInTheDocument());
    expect(screen.getByText('ディール結果（獲得ポイント）')).toBeInTheDocument();
  });

  it('shows the endgame (phase 2) notice once the stock is empty', async () => {
    mockExec.mockResolvedValue(endgameState);
    renderWithProviders(<BeziquePage />);
    await waitFor(() => expect(screen.getByText(/フェーズ2/)).toBeInTheDocument());
  });

  it('rings only the legal cards during an endgame follow turn', async () => {
    mockExec.mockResolvedValue(endgameFollowState);
    renderWithProviders(<BeziquePage />);
    const legal = (await screen.findByAltText('♠ Q')).closest('button');
    const illegalDiamond = screen.getByAltText('♦ J').closest('button');
    const illegalHeart = screen.getByAltText('♥ A').closest('button');
    // ♠Q is the only legal follow → it alone carries the success ring marker.
    expect(legal).toHaveAttribute('data-legal', 'true');
    expect(illegalDiamond).not.toHaveAttribute('data-legal');
    expect(illegalHeart).not.toHaveAttribute('data-legal');
  });

  it('keeps an illegal card clickable during the endgame (backend still validates)', async () => {
    mockExec.mockResolvedValue(endgameFollowState);
    renderWithProviders(<BeziquePage />);
    const illegal = (await screen.findByAltText('♦ J')).closest('button');
    // Ring-only highlighting must never block clicks: the card stays enabled.
    expect(illegal).not.toHaveAttribute('aria-disabled');
    fireEvent.click(illegal as HTMLElement);
    expect(illegal).toHaveAttribute('aria-pressed', 'true');
    mockExec.mockClear();
    mockExec.mockResolvedValue(endgameFollowState);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    // ♦J is index 1 in the hand; the play still dispatches so the server can reject it.
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 1 }));
  });

  it('does not ring any card before the endgame (all cards legal)', async () => {
    renderWithProviders(<BeziquePage />);
    const card = (await screen.findByAltText('♠ Q')).closest('button');
    expect(card).not.toHaveAttribute('data-legal');
  });

  it('shows the low-stock endgame warning when the stock nears empty', async () => {
    mockExec.mockResolvedValue(makeBeziqueState({ phase: 0, currentPlayerIdx: 0, stockRemaining: 3 }));
    renderWithProviders(<BeziquePage />);
    const warning = await screen.findByTestId('bezique-stock-warning');
    expect(warning).toHaveTextContent('3');
    expect(warning).toHaveAttribute('role', 'status');
  });

  it('does not show the low-stock warning while the stock is plentiful', async () => {
    // Default fixture has stockRemaining 30 (well above the threshold).
    renderWithProviders(<BeziquePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '出す' })).toBeInTheDocument());
    expect(screen.queryByTestId('bezique-stock-warning')).not.toBeInTheDocument();
  });

  it('drops the low-stock warning once the endgame begins', async () => {
    mockExec.mockResolvedValue(endgameState);
    renderWithProviders(<BeziquePage />);
    await waitFor(() => expect(screen.getByText(/フェーズ2/)).toBeInTheDocument());
    expect(screen.queryByTestId('bezique-stock-warning')).not.toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<BeziquePage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝利です (1010-820)！')).toBeInTheDocument());
  });
});
