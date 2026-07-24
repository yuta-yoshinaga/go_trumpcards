import { act, fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { beloteApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BeloteResponse, Card } from '../types/card';
import { BelotePhase } from '../types/phases';
import { BelotePage } from './BelotePage';

vi.mock('../api/gameApi', () => ({
  beloteApi: { exec: vi.fn() },
  actionLogApi: { belote: vi.fn() },
}));

const mockPlaySound = vi.fn();
const mockSoundValue = { playSound: mockPlaySound, muted: false, toggleMute: vi.fn() };
vi.mock('../providers/SoundProvider', () => ({
  SoundProvider: ({ children }: { children: React.ReactNode }) => children,
  useSound: () => mockSoundValue,
  useOptionalSound: () => mockSoundValue,
}));

const mockExec = vi.mocked(beloteApi.exec);

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function makeState(overrides: Partial<BeloteResponse> = {}): BeloteResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 5,
        cards: [card('SPADE', 11), card('HEART', 10), card('CLOVER', 9)],
        team: 0,
        trickCount: 0,
      },
      { id: 1, isHuman: false, cardCount: 5, cards: [], team: 1, trickCount: 0 },
      { id: 2, isHuman: false, cardCount: 5, cards: [], team: 0, trickCount: 0 },
      { id: 3, isHuman: false, cardCount: 5, cards: [], team: 1, trickCount: 0 },
    ],
    phase: BelotePhase.BID_PICK_UP,
    roundNumber: 1,
    trickNumber: 0,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 3,
    trumpSuit: 0,
    faceUpCard: card('HEART', 11),
    makerTeam: 0,
    makerPlayerIdx: -1,
    currentTrick: [],
    teamScores: [0, 0],
    roundPoints: [0, 0],
    roundBeloteBonus: [0, 0],
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: 0,
    message: '',
    config: {
      cpuDifficulty: 1,
      targetScore: 1000,
      dixDeDer: 10,
      enableBeloteRebelote: true,
    },
    ...overrides,
  };
}

const initialState = makeState();
const gameEndState = makeState({
  phase: BelotePhase.GAME_END,
  gameEndFlag: true,
  winnerTeam: 0,
  teamScores: [1010, 600],
});

beforeEach(() => {
  mockExec.mockReset();
  mockPlaySound.mockClear();
  mockExec.mockResolvedValue(initialState);
});

describe('BelotePage', () => {
  it('calls reset on mount with default config', async () => {
    renderWithProviders(<BelotePage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        targetScore: 1000,
        dixDeDer: 10,
        enableBeloteRebelote: true,
      }),
    );
  });

  it('shows the orderup and pass buttons in pick-up phase when human is bid turn', async () => {
    renderWithProviders(<BelotePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '取る' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'パス' })).toBeInTheDocument();
  });

  it('renders the face-up card as a card image (not text) during bidding', async () => {
    renderWithProviders(<BelotePage />);
    // faceUpCard is J♥ → AnimatedCard/CardImage exposes it via alt text.
    expect(await screen.findByAltText('♥ J')).toBeInTheDocument();
    // The old plain-text "HEART 11" rendering is gone.
    expect(screen.queryByText(/HEART 11/)).not.toBeInTheDocument();
  });

  it('dispatches orderup when "取る" is clicked', async () => {
    renderWithProviders(<BelotePage />);
    await waitFor(() => screen.getByRole('button', { name: '取る' }));
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '取る' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('orderup'));
  });

  it('shows trump suit buttons during call-trump phase', async () => {
    mockExec.mockResolvedValue(makeState({ phase: BelotePhase.BID_CALL_TRUMP }));
    renderWithProviders(<BelotePage />);
    // 3 callable suits (face-up suit HEART excluded): ♠ ♣ ♦
    await waitFor(() => expect(screen.getByRole('button', { name: '♠ スペード' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '♣ クラブ' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '♦ ダイヤ' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '♥ ハート' })).not.toBeInTheDocument();
  });

  it('dispatches calltrump with the picked suit', async () => {
    mockExec.mockResolvedValue(makeState({ phase: BelotePhase.BID_CALL_TRUMP }));
    renderWithProviders(<BelotePage />);
    await waitFor(() => screen.getByRole('button', { name: '♠ スペード' }));
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '♠ スペード' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('calltrump', undefined, 1));
  });

  it('shows reset button mid-game and opens confirm dialog', async () => {
    renderWithProviders(<BelotePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'リセット' })).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: 'リセット' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('renders bonus trackers (dim by default) during play after trump is set', async () => {
    mockExec.mockResolvedValue(makeState({ phase: BelotePhase.PLAY, trumpSuit: 1, trickNumber: 3, makerTeam: 0 }));
    renderWithProviders(<BelotePage />);
    await waitFor(() => expect(screen.getByTestId('belote-bonus-trackers')).toBeInTheDocument());
    expect(screen.getByTestId('dix-de-der-badge')).not.toHaveAttribute('data-active');
    expect(screen.getByTestId('belote-rebelote-badge')).not.toHaveAttribute('data-active');
  });

  it('activates the dix-de-der badge on the 8th trick', async () => {
    mockExec.mockResolvedValue(makeState({ phase: BelotePhase.PLAY, trumpSuit: 1, trickNumber: 8, makerTeam: 0 }));
    renderWithProviders(<BelotePage />);
    await waitFor(() => expect(screen.getByTestId('dix-de-der-badge')).toHaveAttribute('data-active', 'true'));
  });

  it('activates the belote/rebelote badge once the maker team earns the bonus', async () => {
    mockExec.mockResolvedValue(
      makeState({ phase: BelotePhase.PLAY, trumpSuit: 1, trickNumber: 5, makerTeam: 0, roundBeloteBonus: [20, 0] }),
    );
    renderWithProviders(<BelotePage />);
    await waitFor(() => expect(screen.getByTestId('belote-rebelote-badge')).toHaveAttribute('data-active', 'true'));
  });

  it('activates the belote/rebelote badge when the defender team earns the bonus too', async () => {
    // Backend awards the bonus to whichever team plays K+Q of trump, not the maker.
    mockExec.mockResolvedValue(
      makeState({ phase: BelotePhase.PLAY, trumpSuit: 1, trickNumber: 5, makerTeam: 0, roundBeloteBonus: [0, 20] }),
    );
    renderWithProviders(<BelotePage />);
    await waitFor(() => expect(screen.getByTestId('belote-rebelote-badge')).toHaveAttribute('data-active', 'true'));
  });

  it('chimes and shows a confirmation banner when the belote bonus is freshly earned', async () => {
    mockExec.mockResolvedValue(
      makeState({ phase: BelotePhase.PLAY, trumpSuit: 1, trickNumber: 3, currentPlayerIdx: 0, makerTeam: 0 }),
    );
    renderWithProviders(<BelotePage />);
    const cardBtn = await screen.findByRole('button', { name: '♠ J' });
    fireEvent.click(cardBtn);
    // The play that lands the K+Q-of-trump bonus.
    mockExec.mockResolvedValueOnce(
      makeState({
        phase: BelotePhase.PLAY,
        trumpSuit: 1,
        trickNumber: 3,
        currentPlayerIdx: 1,
        makerTeam: 0,
        roundBeloteBonus: [20, 0],
      }),
    );
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(screen.getByTestId('belote-bonus-confirmed')).toBeInTheDocument());
    expect(mockPlaySound).toHaveBeenCalledWith('winFanfare');
  });

  it('translates a known hint reason', async () => {
    const playState = makeState({ phase: BelotePhase.PLAY, trumpSuit: 1, currentPlayerIdx: 0, makerTeam: 0 });
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<BelotePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    mockExec.mockResolvedValueOnce({ ...playState, hint: { cardIndex: 0, reason: 'trump_cut' } });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getByText(/切り札でカット/)).toBeInTheDocument());
  });

  it('falls back to a generic label for an unknown hint reason', async () => {
    const playState = makeState({ phase: BelotePhase.PLAY, trumpSuit: 1, currentPlayerIdx: 0, makerTeam: 0 });
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<BelotePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ヒント' })).toBeInTheDocument());
    // Omit cardIndex to also exercise the `?? '-'` fallback for a missing index.
    mockExec.mockResolvedValueOnce({ ...playState, hint: { reason: 'brand_new_reason' } });
    fireEvent.click(screen.getByRole('button', { name: 'ヒント' }));
    // Unknown reason -> hintReason.fallback, not the raw identifier.
    await waitFor(() => expect(screen.getByText(/最善手/)).toBeInTheDocument());
    expect(screen.queryByText(/brand_new_reason/)).not.toBeInTheDocument();
  });

  it('auto-hides the confirmation banner after the display window closes', async () => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    try {
      mockExec.mockResolvedValue(
        makeState({ phase: BelotePhase.PLAY, trumpSuit: 1, trickNumber: 3, currentPlayerIdx: 0, makerTeam: 0 }),
      );
      renderWithProviders(<BelotePage />);
      const cardBtn = await screen.findByRole('button', { name: '♠ J' });
      fireEvent.click(cardBtn);
      mockExec.mockResolvedValueOnce(
        makeState({
          phase: BelotePhase.PLAY,
          trumpSuit: 1,
          trickNumber: 3,
          currentPlayerIdx: 1,
          makerTeam: 0,
          roundBeloteBonus: [20, 0],
        }),
      );
      fireEvent.click(screen.getByRole('button', { name: '出す' }));
      await waitFor(() => expect(screen.getByTestId('belote-bonus-confirmed')).toBeInTheDocument());
      await act(async () => {
        vi.advanceTimersByTime(2600);
      });
      await waitFor(() => expect(screen.queryByTestId('belote-bonus-confirmed')).not.toBeInTheDocument());
    } finally {
      vi.useRealTimers();
    }
  });

  it('does not chime on a plain play that earns no new bonus', async () => {
    mockExec.mockResolvedValue(
      makeState({ phase: BelotePhase.PLAY, trumpSuit: 1, trickNumber: 3, currentPlayerIdx: 0, makerTeam: 0 }),
    );
    renderWithProviders(<BelotePage />);
    const cardBtn = await screen.findByRole('button', { name: '♠ J' });
    fireEvent.click(cardBtn);
    mockExec.mockResolvedValueOnce(
      makeState({ phase: BelotePhase.PLAY, trumpSuit: 1, trickNumber: 4, currentPlayerIdx: 1, makerTeam: 0 }),
    );
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, 0));
    await waitFor(() => expect(screen.queryByTestId('belote-bonus-confirmed')).not.toBeInTheDocument());
    expect(mockPlaySound).not.toHaveBeenCalled();
  });

  it('does not chime when loaded into a round that already has the bonus', async () => {
    mockExec.mockResolvedValue(
      makeState({ phase: BelotePhase.PLAY, trumpSuit: 1, trickNumber: 5, makerTeam: 0, roundBeloteBonus: [20, 0] }),
    );
    renderWithProviders(<BelotePage />);
    await waitFor(() => expect(screen.getByTestId('belote-rebelote-badge')).toHaveAttribute('data-active', 'true'));
    expect(mockPlaySound).not.toHaveBeenCalled();
    expect(screen.queryByTestId('belote-bonus-confirmed')).not.toBeInTheDocument();
  });

  it('shows 次のゲーム at game end with no confirm', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<BelotePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のゲーム' })).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, expect.any(Object)));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('shows the hint button during the bid phase and requests a hint', async () => {
    renderWithProviders(<BelotePage />); // default state is BID_PICK_UP, human's bid turn
    const btn = await screen.findByRole('button', { name: 'ヒント' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(initialState);
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('renders an order-up bid hint (take/pass) after requesting it', async () => {
    mockExec.mockResolvedValue(makeState({ hint: { orderUp: true, reason: 'strong' } }));
    renderWithProviders(<BelotePage />);
    fireEvent.click(await screen.findByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getByText(/おすすめ: 取る/)).toBeInTheDocument());
  });

  it('renders a call-trump suit bid hint after requesting it', async () => {
    mockExec.mockResolvedValue(makeState({ phase: BelotePhase.BID_CALL_TRUMP, hint: { suit: 1, reason: 'strong' } }));
    renderWithProviders(<BelotePage />);
    fireEvent.click(await screen.findByRole('button', { name: 'ヒント' }));
    await waitFor(() => expect(screen.getByText(/を宣言/)).toBeInTheDocument());
  });

  it('rings only the legal follow-suit card during the human play turn', async () => {
    // Trump ♠(1). Opponent leads ♥; the human holds ♥10 so must follow ♥.
    // Only ♥10 is legal; ♠J and ♣9 are illegal.
    mockExec.mockResolvedValue(
      makeState({
        phase: BelotePhase.PLAY,
        trumpSuit: 1,
        currentPlayerIdx: 0,
        makerTeam: 0,
        currentTrick: [{ playerIdx: 1, card: card('HEART', 13) }],
      }),
    );
    renderWithProviders(<BelotePage />);
    const legalCard = await screen.findByRole('button', { name: '♥ 10' });
    const illegalCard = screen.getByRole('button', { name: '♠ J' });
    expect(legalCard).toHaveAttribute('data-legal', 'true');
    expect(illegalCard).not.toHaveAttribute('data-legal');
  });

  it('keeps an illegal card clickable so the backend still validates the play', async () => {
    // Same setup: ♠J is illegal (must follow ♥) but must remain selectable —
    // the highlight is additive only and never blocks clicks (see hearts #3977).
    mockExec.mockResolvedValue(
      makeState({
        phase: BelotePhase.PLAY,
        trumpSuit: 1,
        currentPlayerIdx: 0,
        makerTeam: 0,
        currentTrick: [{ playerIdx: 1, card: card('HEART', 13) }],
      }),
    );
    renderWithProviders(<BelotePage />);
    const illegalCard = await screen.findByRole('button', { name: '♠ J' });
    expect(illegalCard).not.toHaveAttribute('aria-disabled');
    // The Play button is disabled until a card is selected.
    expect(screen.getByRole('button', { name: '出す' })).toBeDisabled();
    fireEvent.click(illegalCard);
    // Clicking the illegal card selects it and enables Play — no client-side block.
    expect(illegalCard).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByRole('button', { name: '出す' })).not.toBeDisabled();
  });

  it('does not ring any card during a CPU play turn', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: BelotePhase.PLAY,
        trumpSuit: 1,
        currentPlayerIdx: 1,
        makerTeam: 0,
        currentTrick: [{ playerIdx: 1, card: card('HEART', 13) }],
      }),
    );
    renderWithProviders(<BelotePage />);
    const humanCard = await screen.findByRole('button', { name: '♥ 10' });
    expect(humanCard).not.toHaveAttribute('data-legal');
  });
});
