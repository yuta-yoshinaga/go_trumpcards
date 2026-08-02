import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { koenigrufenApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeKoenigrufenState } from '../test/stateFactories';
import { KoenigrufenPage } from './KoenigrufenPage';

vi.mock('../api/gameApi', () => ({
  koenigrufenApi: { exec: vi.fn() },
  actionLogApi: { koenigrufen: vi.fn() },
}));

const mockExec = vi.mocked(koenigrufenApi.exec);

const suit = (value: number, design: 'HEART' | 'SPADE' | 'CLOVER' | 'DIAMOND', glyph: string, label: string) => ({
  design,
  value,
  glyph,
  label,
  color: design === 'HEART' || design === 'DIAMOND' ? 'red' : 'black',
  deck: 'tarot',
});

const playPhaseState = makeKoenigrufenState();

const bidPhaseState = makeKoenigrufenState({
  phase: 0,
  isHumanTurn: true,
  isHumanBidTurn: true,
  contract: 0,
  highestBid: 0,
  highestBidder: -1,
  calledKing: -1,
  playableIndices: [],
});

// Call phase: human declarer holds the King of Hearts, so Hearts must be disabled.
const callPhaseState = makeKoenigrufenState({
  phase: 1,
  isHumanTurn: true,
  isHumanBidTurn: false,
  isHumanCall: true,
  contract: 1,
  calledKing: -1,
  playableIndices: [],
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 12,
      cards: [suit(8, 'HEART', '♥', 'K'), suit(3, 'SPADE', '♠', '3')],
      trickCount: 0,
      cardPoints: 0,
      score: 0,
      isDeclarer: true,
      isPartner: false,
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 12,
      cards: [],
      trickCount: 0,
      cardPoints: 0,
      score: 0,
      isDeclarer: false,
      isPartner: false,
    },
    {
      id: 2,
      isHuman: false,
      cardCount: 12,
      cards: [],
      trickCount: 0,
      cardPoints: 0,
      score: 0,
      isDeclarer: false,
      isPartner: false,
    },
    {
      id: 3,
      isHuman: false,
      cardCount: 12,
      cards: [],
      trickCount: 0,
      cardPoints: 0,
      score: 0,
      isDeclarer: false,
      isPartner: false,
    },
  ],
});

const talonPhaseState = makeKoenigrufenState({
  phase: 2,
  isHumanTurn: true,
  isHumanBidTurn: false,
  isHumanDiscard: true,
  contract: 1,
  calledKing: 4,
  playableIndices: [],
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 18,
      cards: [
        suit(2, 'HEART', '♥', '2'),
        suit(3, 'HEART', '♥', '3'),
        suit(4, 'HEART', '♥', '4'),
        suit(5, 'HEART', '♥', 'J'),
        suit(6, 'HEART', '♥', 'C'),
        suit(7, 'HEART', '♥', 'Q'),
        // A King (value 8) and the Sküs must NOT be selectable for the talon.
        suit(8, 'SPADE', '♠', 'K'),
        { design: 'JOKER' as const, value: 0, glyph: '★', label: 'Sküs', color: 'gold', deck: 'tarot' },
      ],
      trickCount: 0,
      cardPoints: 0,
      score: 0,
      isDeclarer: true,
      isPartner: false,
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 12,
      cards: [],
      trickCount: 0,
      cardPoints: 0,
      score: 0,
      isDeclarer: false,
      isPartner: false,
    },
    {
      id: 2,
      isHuman: false,
      cardCount: 12,
      cards: [],
      trickCount: 0,
      cardPoints: 0,
      score: 0,
      isDeclarer: false,
      isPartner: false,
    },
    {
      id: 3,
      isHuman: false,
      cardCount: 12,
      cards: [],
      trickCount: 0,
      cardPoints: 0,
      score: 0,
      isDeclarer: false,
      isPartner: false,
    },
  ],
});

const trickEndState = makeKoenigrufenState({
  phase: 4,
  isHumanTurn: false,
  currentTrick: [
    { playerIdx: 0, card: suit(7, 'HEART', '♥', 'Q') },
    { playerIdx: 1, card: suit(8, 'CLOVER', '♣', 'K') },
  ],
});

const roundEndState = makeKoenigrufenState({
  phase: 5,
  isHumanTurn: false,
  outcome: 1,
  partnerRevealed: true,
  partnerIdx: 2,
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 0,
      cards: [],
      trickCount: 3,
      cardPoints: 40,
      score: 2,
      isDeclarer: true,
      isPartner: false,
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 0,
      cards: [],
      trickCount: 3,
      cardPoints: 20,
      score: -1,
      isDeclarer: false,
      isPartner: false,
    },
    {
      id: 2,
      isHuman: false,
      cardCount: 0,
      cards: [],
      trickCount: 3,
      cardPoints: 10,
      score: 2,
      isDeclarer: false,
      isPartner: true,
    },
    {
      id: 3,
      isHuman: false,
      cardCount: 0,
      cards: [],
      trickCount: 3,
      cardPoints: 5,
      score: -1,
      isDeclarer: false,
      isPartner: false,
    },
  ],
});

const gameEndState = makeKoenigrufenState({
  phase: 6,
  isHumanTurn: false,
  gameEndFlag: true,
  winnerPlayer: 0,
  message: 'ゲーム終了！ あなたの勝ち！',
});

const gameEndDrawState = makeKoenigrufenState({
  phase: 6,
  isHumanTurn: false,
  gameEndFlag: true,
  winnerPlayer: -1,
  message: 'ゲーム終了！ 引き分け！',
});

const cpuTurnState = makeKoenigrufenState({ currentPlayerIdx: 1, isHumanTurn: false });

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('KoenigrufenPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<KoenigrufenPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<KoenigrufenPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, targetDeals: 5 },
      }),
    );
  });

  it('renders the play phase with the human cards and the declarer badge', async () => {
    renderWithProviders(<KoenigrufenPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'K ♥' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '21 ✦' })).toBeInTheDocument();
    });
    expect(screen.getAllByText('デクレアラー').length).toBeGreaterThan(0);
  });

  it('renders the bid phase with Pass and the Rufer button', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<KoenigrufenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'ルーファー' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'パス' })).toBeInTheDocument();
  });

  it('declaring Rufer dispatches bid with the contract string', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<KoenigrufenPage />);
    const ruferBtn = await screen.findByRole('button', { name: 'ルーファー' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(ruferBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 'rufer' }));
  });

  it('passing dispatches the pass command', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<KoenigrufenPage />);
    const passBtn = await screen.findByRole('button', { name: 'パス' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(passBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass'));
  });

  it('does not show bid controls on a CPU/non-bid turn', async () => {
    renderWithProviders(<KoenigrufenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'K ♥' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'ルーファー' })).not.toBeInTheDocument();
  });

  it('renders the call phase with four suit buttons and disables a held King suit', async () => {
    mockExec.mockResolvedValue(callPhaseState);
    renderWithProviders(<KoenigrufenPage />);
    // Declarer holds the King of Hearts, so Hearts (suit 3) is disabled; the other three are enabled.
    const hearts = await screen.findByTestId('call-king-3');
    expect(hearts).toBeDisabled();
    expect(screen.getByRole('button', { name: 'スペード' })).toBeEnabled();
    expect(screen.getByRole('button', { name: 'クラブ' })).toBeEnabled();
    expect(screen.getByRole('button', { name: 'ダイヤ' })).toBeEnabled();
  });

  it('explains why a held-King suit cannot be called (aria-label + title, strike-through)', async () => {
    mockExec.mockResolvedValue(callPhaseState);
    renderWithProviders(<KoenigrufenPage />);
    const hearts = await screen.findByTestId('call-king-3');
    const reason = '既にこのスートの王を保有しているため呼べません';
    expect(hearts).toHaveAttribute('aria-label', `ハート — ${reason}`);
    expect(hearts.className).toContain('line-through');
    // The tooltip lives on the wrapping span (disabled buttons suppress native tooltips).
    expect(hearts.closest('span')).toHaveAttribute('title', reason);
    // An enabled suit keeps its plain label and no reason.
    const spade = screen.getByTestId('call-king-1');
    expect(spade).toHaveAttribute('aria-label', 'スペード');
    expect(spade.closest('span')).not.toHaveAttribute('title');
  });

  it('calling a King dispatches callking with the suit index', async () => {
    mockExec.mockResolvedValue(callPhaseState);
    renderWithProviders(<KoenigrufenPage />);
    const spadeBtn = await screen.findByRole('button', { name: 'スペード' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(callPhaseState);
    fireEvent.click(spadeBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('callking', { callSuit: 1 }));
  });

  it('renders the talon phase and dispatches discard for exactly 6 selected cards', async () => {
    mockExec.mockResolvedValue(talonPhaseState);
    renderWithProviders(<KoenigrufenPage />);
    await screen.findByRole('button', { name: '2 ♥' });
    for (const label of ['2 ♥', '3 ♥', '4 ♥', 'J ♥', 'C ♥', 'Q ♥']) {
      fireEvent.click(screen.getByRole('button', { name: label }));
    }
    const discardBtn = screen.getByRole('button', { name: /捨てる/ });
    expect(discardBtn).toBeEnabled();
    mockExec.mockClear();
    mockExec.mockResolvedValue(talonPhaseState);
    fireEvent.click(discardBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', { cardIndices: [0, 1, 2, 3, 4, 5] }));
  });

  it('keeps the discard button disabled until exactly 6 cards are chosen', async () => {
    mockExec.mockResolvedValue(talonPhaseState);
    renderWithProviders(<KoenigrufenPage />);
    await screen.findByRole('button', { name: '2 ♥' });
    fireEvent.click(screen.getByRole('button', { name: '2 ♥' }));
    expect(screen.getByRole('button', { name: /捨てる/ })).toBeDisabled();
  });

  it('shows the called King once it is named', async () => {
    mockExec.mockResolvedValue(talonPhaseState);
    renderWithProviders(<KoenigrufenPage />);
    await waitFor(() => expect(screen.getByTestId('koenigrufen-called-king')).toBeInTheDocument());
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<KoenigrufenPage />);
    const card = await screen.findByRole('button', { name: 'K ♥' });
    fireEvent.click(card);
    const playBtn = await screen.findByRole('button', { name: '出す' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('renders trick end with the next trick button', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<KoenigrufenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('renders round end with the next deal button, the deal result, and the revealed partner', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<KoenigrufenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のディール' })).toBeInTheDocument());
    expect(screen.getByTestId('koenigrufen-result')).toBeInTheDocument();
    // Partner is only shown once revealed.
    expect(screen.getAllByText('パートナー').length).toBeGreaterThan(0);
  });

  it('does not reveal the partner before partnerRevealed is true, but shows the called-King clue', async () => {
    renderWithProviders(<KoenigrufenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'K ♥' })).toBeInTheDocument());
    // The revealed-partner label must not appear yet.
    expect(screen.queryByText('パートナー')).not.toBeInTheDocument();
    // Instead, the organized clue tells the human what they legitimately know.
    const clue = screen.getByTestId('koenigrufen-partner-clue');
    expect(clue).toHaveTextContent('の王を持つプレイヤーが宣言者の秘密のパートナー（まだ不明）');
    // The human here is the declarer (does not hold the called King), so no "you are the partner" line.
    expect(screen.queryByTestId('koenigrufen-partner-clue-you')).not.toBeInTheDocument();
  });

  it('tells the human they are the secret partner when they hold the called King', async () => {
    const partnerState = makeKoenigrufenState({
      calledKing: 4,
      partnerRevealed: false,
      isHumanTurn: false,
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 12,
          // The human is a defender-seat holder of the called Diamond King → they are the secret partner.
          cards: [suit(8, 'DIAMOND', '♦', 'K'), suit(3, 'SPADE', '♠', '3')],
          trickCount: 0,
          cardPoints: 0,
          score: 0,
          isDeclarer: false,
          isPartner: false,
        },
        {
          id: 1,
          isHuman: false,
          cardCount: 12,
          cards: [],
          trickCount: 0,
          cardPoints: 0,
          score: 0,
          isDeclarer: true,
          isPartner: false,
        },
        {
          id: 2,
          isHuman: false,
          cardCount: 12,
          cards: [],
          trickCount: 0,
          cardPoints: 0,
          score: 0,
          isDeclarer: false,
          isPartner: false,
        },
        {
          id: 3,
          isHuman: false,
          cardCount: 12,
          cards: [],
          trickCount: 0,
          cardPoints: 0,
          score: 0,
          isDeclarer: false,
          isPartner: false,
        },
      ],
    });
    mockExec.mockResolvedValue(partnerState);
    renderWithProviders(<KoenigrufenPage />);
    await waitFor(() => expect(screen.getByTestId('koenigrufen-partner-clue-you')).toBeInTheDocument());
    expect(screen.getByTestId('koenigrufen-partner-clue-you')).toHaveTextContent('あなたがその王を持っています');
    // Still must not leak the seat index as a revealed partner.
    expect(screen.queryByText('パートナー')).not.toBeInTheDocument();
  });

  it('switches to the revealed-partner label once partnerRevealed is true', async () => {
    const revealedState = makeKoenigrufenState({ partnerRevealed: true, partnerIdx: 2 });
    mockExec.mockResolvedValue(revealedState);
    renderWithProviders(<KoenigrufenPage />);
    await waitFor(() => expect(screen.getByTestId('koenigrufen-called-king')).toBeInTheDocument());
    // The unknown clue is gone; the partner is now named.
    expect(screen.queryByTestId('koenigrufen-partner-clue')).not.toBeInTheDocument();
    expect(screen.getByText(/パートナー: /)).toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<KoenigrufenPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝ち！')).toBeInTheDocument());
  });

  it('renders a draw at game end without a false win celebration', async () => {
    mockExec.mockResolvedValue(gameEndDrawState);
    renderWithProviders(<KoenigrufenPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ 引き分け！')).toBeInTheDocument());
  });

  it('the next-game button at game end resets immediately', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<KoenigrufenPage />);
    const nextGame = await screen.findByRole('button', { name: '次のゲーム' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(gameEndState);
    fireEvent.click(nextGame);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', expect.anything()));
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<KoenigrufenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'K ♥' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('changing the CPU difficulty and target-deals selects updates the config', async () => {
    renderWithProviders(<KoenigrufenPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'K ♥' })).toBeInTheDocument());
    const difficulty = screen.getByLabelText('CPU難易度') as HTMLSelectElement;
    fireEvent.change(difficulty, { target: { value: '2' } });
    expect(difficulty.value).toBe('2');
    const deals = screen.getByLabelText('マッチのディール数') as HTMLSelectElement;
    fireEvent.change(deals, { target: { value: '3' } });
    expect(deals.value).toBe('3');
  });

  it('renders the backend hint banner with its card indices', async () => {
    mockExec.mockResolvedValue(
      makeKoenigrufenState({
        hint: { cardIndices: [0, 2], reason: 'lead_high' },
        messageCode: 'koenigrufen.hintRequested',
      }),
    );
    renderWithProviders(<KoenigrufenPage />);
    await waitFor(() => expect(screen.getByText(/\[0\], \[2\]/)).toBeInTheDocument());
  });
});
