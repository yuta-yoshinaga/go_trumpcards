import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { jassApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { Card, JassResponse } from '../types/card';
import { JassPhase } from '../types/phases';
import { JassPage } from './JassPage';

vi.mock('../api/gameApi', () => ({
  jassApi: { exec: vi.fn() },
  actionLogApi: { jass: vi.fn() },
}));

const mockPlaySound = vi.fn();
const mockSoundValue = { playSound: mockPlaySound, muted: false, toggleMute: vi.fn() };
vi.mock('../providers/SoundProvider', () => ({
  SoundProvider: ({ children }: { children: React.ReactNode }) => children,
  useSound: () => mockSoundValue,
  useOptionalSound: () => mockSoundValue,
}));

const mockExec = vi.mocked(jassApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function makeState(overrides: Partial<JassResponse> = {}): JassResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 9,
        cards: [card('SPADE', 11), card('HEART', 10), card('CLOVER', 9)],
        team: 0,
        trickCount: 0,
      },
      { id: 1, isHuman: false, cardCount: 9, cards: [], team: 1, trickCount: 0 },
      { id: 2, isHuman: false, cardCount: 9, cards: [], team: 0, trickCount: 0 },
      { id: 3, isHuman: false, cardCount: 9, cards: [], team: 1, trickCount: 0 },
    ],
    phase: JassPhase.BID_TRUMP,
    roundNumber: 1,
    trickNumber: 0,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 3,
    forehandIdx: 0,
    trumpSuit: 0,
    schieben: false,
    makerTeam: -1,
    makerPlayerIdx: -1,
    currentTrick: [],
    lastTrick: [],
    lastTrickWinner: -1,
    teamScores: [0, 0],
    roundPoints: [0, 0],
    roundWeisPoints: [0, 0],
    roundStockPoints: [0, 0],
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: 0,
    message: '',
    config: {
      cpuDifficulty: 1,
      targetScore: 1000,
      lastTrickBonus: 5,
      enableWeis: true,
    },
    ...overrides,
  };
}

const initialState = makeState();
const gameEndState = makeState({
  phase: JassPhase.GAME_END,
  gameEndFlag: true,
  winnerTeam: 0,
  teamScores: [1010, 600],
});

beforeEach(() => {
  mockExec.mockReset();
  mockPlaySound.mockClear();
  mockExec.mockResolvedValue(initialState);
});

describe('JassPage', () => {
  it('calls reset on mount with default config', async () => {
    renderWithProviders(<JassPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        targetScore: 1000,
        lastTrickBonus: 5,
        enableWeis: true,
      }),
    );
  });

  // The phase key map must hold bare keys; usePhaseNames adds the `phase.`
  // prefix itself, so a prefixed key resolved to the literal
  // "phase.phase.bidTrump" on screen. See issue #4374.
  it('renders the translated phase name, not the raw i18n key', async () => {
    renderWithProviders(<JassPage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('切り札を選ぶ'));
    expect(screen.getByTestId('phase-indicator')).not.toHaveTextContent('phase.');
  });

  it('shows the four trump suit buttons and Schieben in the trump-bid phase', async () => {
    renderWithProviders(<JassPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '♠ スペード' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '♣ クラブ' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '♥ ハート' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '♦ ダイヤ' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'シーバー' })).toBeInTheDocument();
  });

  it('dispatches calltrump with the picked suit', async () => {
    renderWithProviders(<JassPage />);
    await waitFor(() => screen.getByRole('button', { name: '♥ ハート' }));
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '♥ ハート' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('calltrump', 3));
  });

  it('dispatches schieben when the Schieben button is clicked', async () => {
    renderWithProviders(<JassPage />);
    await waitFor(() => screen.getByRole('button', { name: 'シーバー' }));
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'シーバー' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('schieben'));
  });

  it('hides Schieben but shows all suits in the partner-bid phase', async () => {
    mockExec.mockResolvedValue(makeState({ phase: JassPhase.BID_PARTNER }));
    renderWithProviders(<JassPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '♠ スペード' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '♥ ハート' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'シーバー' })).not.toBeInTheDocument();
  });

  it('dispatches play with the selected card index during play', async () => {
    mockExec.mockResolvedValue(makeState({ phase: JassPhase.PLAY, trumpSuit: 1, currentPlayerIdx: 0 }));
    renderWithProviders(<JassPage />);
    const cardBtn = await screen.findByRole('button', { name: '♠ J' });
    fireEvent.click(cardBtn);
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, 0));
  });

  it('renders the team score table during play', async () => {
    mockExec.mockResolvedValue(makeState({ phase: JassPhase.PLAY, trumpSuit: 1, teamScores: [40, 25] }));
    renderWithProviders(<JassPage />);
    await waitFor(() => expect(screen.getByText('チームスコア')).toBeInTheDocument());
  });

  it('shows the Weis panel with per-team totals and a counted marker when Weis is declared', async () => {
    mockExec.mockResolvedValue(makeState({ phase: JassPhase.TRICK_END, trumpSuit: 1, roundWeisPoints: [20, 0] }));
    renderWithProviders(<JassPage />);
    const panel = await screen.findByTestId('jass-weis-panel');
    expect(panel).toBeInTheDocument();
    expect(panel).toHaveTextContent('Weis（メルド）宣言');
    // Human is on team 0, so the (You) marker appears on team 0.
    expect(panel).toHaveTextContent('チーム0（あなた）');
    expect(panel).toHaveTextContent('20点');
    // Only the scoring team (team 0) gets the "獲得" marker.
    expect(screen.getAllByText('獲得')).toHaveLength(1);
  });

  it('hides the Weis panel when the feature is disabled', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: JassPhase.TRICK_END,
        trumpSuit: 1,
        roundWeisPoints: [20, 0],
        config: { cpuDifficulty: 1, targetScore: 1000, lastTrickBonus: 5, enableWeis: false },
      }),
    );
    renderWithProviders(<JassPage />);
    await waitFor(() => expect(screen.getByText('チームスコア')).toBeInTheDocument());
    expect(screen.queryByTestId('jass-weis-panel')).not.toBeInTheDocument();
  });

  it('hides the Weis panel when no Weis points were declared', async () => {
    mockExec.mockResolvedValue(makeState({ phase: JassPhase.TRICK_END, trumpSuit: 1, roundWeisPoints: [0, 0] }));
    renderWithProviders(<JassPage />);
    await waitFor(() => expect(screen.getByText('チームスコア')).toBeInTheDocument());
    expect(screen.queryByTestId('jass-weis-panel')).not.toBeInTheDocument();
  });

  it('translates a known hint reason', async () => {
    const playState = makeState({ phase: JassPhase.PLAY, trumpSuit: 1, currentPlayerIdx: 0 });
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<JassPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    mockExec.mockResolvedValueOnce({ ...playState, hint: { cardIndex: 0, reason: 'trump_cut' } });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getByText(/切り札でカット/)).toBeInTheDocument());
  });

  it('falls back to a generic label for an unknown hint reason', async () => {
    const playState = makeState({ phase: JassPhase.PLAY, trumpSuit: 1, currentPlayerIdx: 0 });
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<JassPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    // Omit cardIndex to also exercise the `?? '-'` fallback for a missing index.
    mockExec.mockResolvedValueOnce({ ...playState, hint: { reason: 'brand_new_reason' } });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getByText(/最善手/)).toBeInTheDocument());
    expect(screen.queryByText(/brand_new_reason/)).not.toBeInTheDocument();
  });

  it('shows an empty previous-trick reviewer when no trick has completed yet', async () => {
    mockExec.mockResolvedValue(makeState({ phase: JassPhase.PLAY, trumpSuit: 1, lastTrick: [], lastTrickWinner: -1 }));
    renderWithProviders(<JassPage />);
    const panel = await screen.findByTestId('ja-previous-trick');
    expect(panel).toBeInTheDocument();
    expect(panel).toHaveTextContent('このラウンドにはまだ前のトリックがありません');
  });

  it('renders the previous trick cards and the winner label when a trick has completed', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: JassPhase.PLAY,
        trumpSuit: 1,
        trickNumber: 2,
        lastTrick: [
          { playerIdx: 1, card: card('SPADE', 3) },
          { playerIdx: 2, card: card('SPADE', 1) },
          { playerIdx: 3, card: card('SPADE', 5) },
          { playerIdx: 0, card: card('SPADE', 7) },
        ],
        lastTrickWinner: 2,
      }),
    );
    renderWithProviders(<JassPage />);
    const panel = await screen.findByTestId('ja-previous-trick');
    expect(panel).toBeInTheDocument();
    // Winner label references the winning player's name.
    await waitFor(() => expect(panel).toHaveTextContent('が獲得'));
  });

  it('shows reset button mid-game and opens confirm dialog', async () => {
    renderWithProviders(<JassPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('shows 次のゲーム at game end with no confirm', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<JassPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, expect.any(Object)));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });
});
