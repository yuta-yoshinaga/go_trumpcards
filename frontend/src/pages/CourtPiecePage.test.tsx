import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { courtPieceApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeCourtPieceState } from '../test/stateFactories';
import { CourtPiecePage } from './CourtPiecePage';

vi.mock('../api/gameApi', () => ({
  courtPieceApi: { exec: vi.fn() },
  actionLogApi: { courtpiece: vi.fn() },
}));

const mockExec = vi.mocked(courtPieceApi.exec);

// Default fixture: a human trump-declaration turn (the human at seat 0 is the caller).
const trumpPhaseState = makeCourtPieceState();
// A human play turn (currentPlayerIdx is the human). Trump declared.
const playPhaseState = makeCourtPieceState({
  phase: 1,
  trumpSuit: 3,
  currentPlayerIdx: 0,
  players: [
    {
      id: 0,
      isHuman: true,
      team: 0,
      cardCount: 13,
      cards: [
        { design: 'HEART', value: 12 },
        { design: 'HEART', value: 13 },
        { design: 'SPADE', value: 1 },
      ],
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
    },
    { id: 1, isHuman: false, team: 1, cardCount: 13, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0 },
    { id: 2, isHuman: false, team: 0, cardCount: 13, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0 },
    { id: 3, isHuman: false, team: 1, cardCount: 13, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0 },
  ],
});
// A human play turn where a heart has been led, so the human must follow with a
// heart. ♥Q and ♥K are legal; ♠A is illegal but must stay clickable.
const followSuitState = makeCourtPieceState({
  phase: 1,
  trumpSuit: 3,
  currentPlayerIdx: 0,
  currentTrick: [{ playerIdx: 3, card: { design: 'HEART', value: 7 } }],
  players: [
    {
      id: 0,
      isHuman: true,
      team: 0,
      cardCount: 3,
      cards: [
        { design: 'HEART', value: 12 },
        { design: 'HEART', value: 13 },
        { design: 'SPADE', value: 1 },
      ],
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
    },
    { id: 1, isHuman: false, team: 1, cardCount: 3, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0 },
    { id: 2, isHuman: false, team: 0, cardCount: 3, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0 },
    { id: 3, isHuman: false, team: 1, cardCount: 3, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0 },
  ],
});
const cpuTurnState = makeCourtPieceState({ phase: 1, trumpSuit: 3, currentPlayerIdx: 1 });
const trickEndState = makeCourtPieceState({ phase: 2, trumpSuit: 3 });
const roundEndState = makeCourtPieceState({
  phase: 3,
  trumpSuit: 3,
  lastRoundCourt: true,
  players: [
    { id: 0, isHuman: true, team: 0, cardCount: 0, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 4 },
    { id: 1, isHuman: false, team: 1, cardCount: 0, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 2 },
    { id: 2, isHuman: false, team: 0, cardCount: 0, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 3 },
    { id: 3, isHuman: false, team: 1, cardCount: 0, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 4 },
  ],
});
const gameEndState = makeCourtPieceState({
  phase: 4,
  gameEndFlag: true,
  winnerTeam: 0,
  message: 'ゲーム終了！ あなたのチームの勝ち！',
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(trumpPhaseState);
});

describe('CourtPiecePage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<CourtPiecePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<CourtPiecePage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, pointLimit: 7 },
      }),
    );
  });

  it('shows trump suit buttons on a human trump-declaration turn', async () => {
    renderWithProviders(<CourtPiecePage />);
    await waitFor(() => expect(screen.getByTestId('trump-1')).toBeInTheDocument());
    expect(screen.getByTestId('trump-2')).toBeInTheDocument();
    expect(screen.getByTestId('trump-3')).toBeInTheDocument();
    expect(screen.getByTestId('trump-4')).toBeInTheDocument();
  });

  it('shows the undeclared status and accessible labels on suit buttons during declaration', async () => {
    renderWithProviders(<CourtPiecePage />);
    await waitFor(() => expect(screen.getByTestId('cp-trump-status')).toBeInTheDocument());
    expect(screen.getByTestId('cp-trump-status')).toHaveTextContent('切り札未宣言');
    // Each suit button carries a declare aria-label with the interpolated suit name.
    for (const [testId, label] of [
      ['trump-1', '♠ スペードを切り札に宣言'],
      ['trump-2', '♣ クラブを切り札に宣言'],
      ['trump-3', '♥ ハートを切り札に宣言'],
      ['trump-4', '♦ ダイヤを切り札に宣言'],
    ] as const) {
      expect(screen.getByTestId(testId)).toHaveAttribute('aria-label', label);
    }
  });

  it('dispatches a trump declaration when a suit button is clicked', async () => {
    renderWithProviders(<CourtPiecePage />);
    const trump3 = await screen.findByTestId('trump-3');
    mockExec.mockClear();
    mockExec.mockResolvedValue(trumpPhaseState);
    fireEvent.click(trump3);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('trump', { trumpSuit: 3 }));
  });

  it('renders the play phase with the human cards and the caller badge', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<CourtPiecePage />);
    await waitFor(() => {
      expect(screen.getByAltText('♥ Q')).toBeInTheDocument();
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
    });
    expect(screen.getByText('コーラー')).toBeInTheDocument();
  });

  it('rings the follow-suit-legal cards during a human play turn', async () => {
    mockExec.mockResolvedValue(followSuitState);
    renderWithProviders(<CourtPiecePage />);
    const legalCard = (await screen.findByAltText('♥ Q')).closest('button');
    // Legal cards carry the additive success ring.
    expect(legalCard?.className).toContain('ring-ds-success');
    // The other heart is also legal.
    expect((await screen.findByAltText('♥ K')).closest('button')?.className).toContain('ring-ds-success');
    // The illegal spade gets no success ring.
    const illegalCard = (await screen.findByAltText('♠ A')).closest('button');
    expect(illegalCard?.className).not.toContain('ring-ds-success');
  });

  it('keeps an illegal card clickable (ring-only, no hard block)', async () => {
    mockExec.mockResolvedValue(followSuitState);
    renderWithProviders(<CourtPiecePage />);
    const illegalCard = (await screen.findByAltText('♠ A')).closest('button');
    // The card must NOT be disabled — the backend still validates the play.
    expect(illegalCard).not.toHaveAttribute('aria-disabled');
    expect(illegalCard?.className).not.toContain('cursor-not-allowed');
    // Clicking it selects the card (does not throw / is not blocked).
    if (illegalCard) fireEvent.click(illegalCard);
    await waitFor(() => expect(screen.getByRole('button', { name: '出す' })).toBeEnabled());
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<CourtPiecePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
    expect(screen.queryByTestId('trump-1')).not.toBeInTheDocument();
  });

  it('renders trick end with the next trick button', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<CourtPiecePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('shows the live team-trick tally toward the 7-trick target during play', async () => {
    mockExec.mockResolvedValue(
      makeCourtPieceState({
        phase: 1,
        trumpSuit: 3,
        currentPlayerIdx: 0,
        players: [
          { id: 0, isHuman: true, team: 0, cardCount: 9, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 3 },
          { id: 1, isHuman: false, team: 1, cardCount: 9, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 1 },
          { id: 2, isHuman: false, team: 0, cardCount: 9, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 1 },
          { id: 3, isHuman: false, team: 1, cardCount: 9, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 1 },
        ],
      }),
    );
    renderWithProviders(<CourtPiecePage />);
    // Team 0 = seats 0(3) + 2(1) = 4 tricks; team 1 = seats 1(1) + 3(1) = 2 tricks.
    const teamA = await screen.findByTestId('cp-live-tricks-team-0');
    expect(teamA).toHaveTextContent('チームA 4/7');
    expect(screen.getByTestId('cp-live-tricks-team-1')).toHaveTextContent('チームB 2/7');
    // Neither team has reached the target, so no accent emphasis.
    expect(teamA.className).not.toContain('text-ds-accent');
  });

  it('emphasizes a team that has reached the 7-trick target during play', async () => {
    mockExec.mockResolvedValue(
      makeCourtPieceState({
        phase: 1,
        trumpSuit: 3,
        currentPlayerIdx: 0,
        players: [
          { id: 0, isHuman: true, team: 0, cardCount: 2, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 4 },
          { id: 1, isHuman: false, team: 1, cardCount: 2, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 2 },
          { id: 2, isHuman: false, team: 0, cardCount: 2, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 3 },
          { id: 3, isHuman: false, team: 1, cardCount: 2, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 2 },
        ],
      }),
    );
    renderWithProviders(<CourtPiecePage />);
    // Team 0 = 4 + 3 = 7 tricks (target reached); team 1 = 4.
    const teamA = await screen.findByTestId('cp-live-tricks-team-0');
    expect(teamA).toHaveTextContent('チームA 7/7');
    expect(teamA.className).toContain('text-ds-accent');
    expect(screen.getByTestId('cp-live-tricks-team-1').className).not.toContain('text-ds-accent');
  });

  it('hides the live team-trick tally once the round ends', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<CourtPiecePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.queryByTestId('cp-live-tricks')).not.toBeInTheDocument();
  });

  it('renders round end with the next round button and the round result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<CourtPiecePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.getByText('ラウンド結果（獲得トリック）')).toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<CourtPiecePage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたのチームの勝ち！')).toBeInTheDocument());
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` だけを見て描画すると常時表示になる (#4605)。
  it('renders no hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue({ ...trumpPhaseState, hint: { cardIndex: 0, reason: 'x' } });
    renderWithProviders(<CourtPiecePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByText(/\(\[0\]\)/)).not.toBeInTheDocument();
  });
});
