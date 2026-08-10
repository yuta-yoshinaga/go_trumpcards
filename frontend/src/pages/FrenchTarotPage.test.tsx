import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { frenchtarotApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeFrenchTarotState } from '../test/stateFactories';
import { FrenchTarotPage } from './FrenchTarotPage';

vi.mock('../api/gameApi', () => ({
  frenchtarotApi: { exec: vi.fn() },
  actionLogApi: { frenchtarot: vi.fn() },
}));

const mockExec = vi.mocked(frenchtarotApi.exec);

const suit = (value: number, design: 'HEART' | 'SPADE' | 'CLOVER' | 'DIAMOND', glyph: string, label: string) => ({
  design,
  value,
  glyph,
  label,
  color: design === 'HEART' || design === 'DIAMOND' ? 'red' : 'black',
  deck: 'tarot',
});

const playPhaseState = makeFrenchTarotState();

const bidPhaseState = makeFrenchTarotState({
  phase: 0,
  isHumanTurn: true,
  isHumanBidTurn: true,
  contract: 0,
  highestBid: 0,
  highestBidder: -1,
  playableIndices: [],
});

const bidPhaseGardeState = makeFrenchTarotState({
  phase: 0,
  isHumanTurn: true,
  isHumanBidTurn: true,
  contract: 0,
  highestBid: 2, // Garde already bid — Petite must be disabled, Garde Sans/Contre enabled
  highestBidder: 1,
  playableIndices: [],
});

const chienPhaseState = makeFrenchTarotState({
  phase: 1,
  isHumanTurn: true,
  isHumanBidTurn: false,
  isHumanDiscard: true,
  contract: 1,
  chienRevealed: true,
  chien: [suit(2, 'HEART', '♥', '2'), suit(3, 'SPADE', '♠', '3')],
  playableIndices: [],
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 24,
      cards: [
        suit(2, 'HEART', '♥', '2'),
        suit(3, 'HEART', '♥', '3'),
        suit(4, 'HEART', '♥', '4'),
        suit(5, 'HEART', '♥', '5'),
        suit(6, 'HEART', '♥', '6'),
        suit(7, 'HEART', '♥', '7'),
        // A King (value 14) and the Excuse must NOT be selectable for the écart.
        suit(14, 'SPADE', '♠', 'R'),
        { design: 'JOKER' as const, value: 0, glyph: '★', label: 'Excuse', color: 'gold', deck: 'tarot' },
      ],
      trickCount: 0,
      cardPoints: 0,
      score: 0,
      isDeclarer: true,
    },
    { id: 1, isHuman: false, cardCount: 18, cards: [], trickCount: 0, cardPoints: 0, score: 0, isDeclarer: false },
    { id: 2, isHuman: false, cardCount: 18, cards: [], trickCount: 0, cardPoints: 0, score: 0, isDeclarer: false },
    { id: 3, isHuman: false, cardCount: 18, cards: [], trickCount: 0, cardPoints: 0, score: 0, isDeclarer: false },
  ],
});

const trickEndState = makeFrenchTarotState({
  phase: 3,
  isHumanTurn: false,
  currentTrick: [
    { playerIdx: 0, card: suit(12, 'HEART', '♥', 'C') },
    { playerIdx: 1, card: suit(13, 'CLOVER', '♣', 'D') },
  ],
});

const roundEndState = makeFrenchTarotState({
  phase: 4,
  isHumanTurn: false,
  outcome: 1,
});

const gameEndState = makeFrenchTarotState({
  phase: 5,
  isHumanTurn: false,
  gameEndFlag: true,
  winnerPlayer: 0,
  message: 'ゲーム終了！ あなたの勝ち！',
});

const cpuTurnState = makeFrenchTarotState({ currentPlayerIdx: 1, isHumanTurn: false });

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('FrenchTarotPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<FrenchTarotPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<FrenchTarotPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, targetDeals: 5 },
      }),
    );
  });

  it('renders the play phase with the human cards and the declarer badge', async () => {
    renderWithProviders(<FrenchTarotPage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: 'D ♥' })).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '21 ✦' })).toBeInTheDocument();
    });
    expect(screen.getAllByText('デクレアラー').length).toBeGreaterThan(0);
  });

  it('shows the bouts held in hand (the 21 and the Excuse) and excludes a suit Ace', async () => {
    // The default hand holds the 21 (purple trump) + the Excuse (gold), plus a black CLOVER Ace
    // (value 1) that must NOT be counted as the Petit.
    renderWithProviders(<FrenchTarotPage />);
    const panel = await screen.findByTestId('frenchtarot-bouts');
    expect(panel).toHaveTextContent('保有ブー（2/3）');
    expect(screen.getByTestId('frenchtarot-bout-twentyOne')).toBeInTheDocument();
    expect(screen.getByTestId('frenchtarot-bout-excuse')).toBeInTheDocument();
    expect(screen.queryByTestId('frenchtarot-bout-petit')).not.toBeInTheDocument();
  });

  it('shows no bouts when the hand holds none', async () => {
    mockExec.mockResolvedValue(
      makeFrenchTarotState({
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 3,
            cards: [
              { design: 'HEART' as const, value: 13, glyph: '♥', label: 'D', color: 'red', deck: 'tarot' },
              { design: 'CLOVER' as const, value: 1, glyph: '♣', label: '1', color: 'black', deck: 'tarot' },
              { design: 'JOKER' as const, value: 5, glyph: '✦', label: '5', color: 'purple', deck: 'tarot' },
            ],
            trickCount: 0,
            cardPoints: 0,
            score: 0,
            isDeclarer: true,
          },
          { id: 1, isHuman: false, cardCount: 3, cards: [], trickCount: 0, cardPoints: 0, score: 0, isDeclarer: false },
          { id: 2, isHuman: false, cardCount: 3, cards: [], trickCount: 0, cardPoints: 0, score: 0, isDeclarer: false },
          { id: 3, isHuman: false, cardCount: 3, cards: [], trickCount: 0, cardPoints: 0, score: 0, isDeclarer: false },
        ],
        playableIndices: [0, 1, 2],
      }),
    );
    renderWithProviders(<FrenchTarotPage />);
    const panel = await screen.findByTestId('frenchtarot-bouts');
    expect(panel).toHaveTextContent('保有ブー（0/3）');
    expect(panel).toHaveTextContent('なし');
    // The target only speaks to a bid, so it is not offered mid-play (#4857).
    expect(screen.queryByTestId('frenchtarot-bouts-target')).not.toBeInTheDocument();
    expect(screen.queryByTestId('frenchtarot-bout-twentyOne')).not.toBeInTheDocument();
  });

  it('renders the bid phase with Pass and the four contract buttons', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<FrenchTarotPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'プティット' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'パス' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ガルド' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ガルド・サン' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ガルド・コントル' })).toBeInTheDocument();
  });

  it('declaring Petite dispatches bid with the contract string', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<FrenchTarotPage />);
    const petiteBtn = await screen.findByRole('button', { name: 'プティット' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(petiteBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 'petite' }));
  });

  it('passing dispatches the pass command', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<FrenchTarotPage />);
    const passBtn = await screen.findByRole('button', { name: 'パス' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(passBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass'));
  });

  it('disables contracts not higher than the current highest bid', async () => {
    mockExec.mockResolvedValue(bidPhaseGardeState);
    renderWithProviders(<FrenchTarotPage />);
    // Highest bid is Garde (2): Petite (1) and Garde (2) are disabled, higher ones enabled.
    expect(await screen.findByRole('button', { name: 'プティット' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'ガルド' })).toBeDisabled();
    expect(screen.getByRole('button', { name: 'ガルド・サン' })).toBeEnabled();
    expect(screen.getByRole('button', { name: 'ガルド・コントル' })).toBeEnabled();
  });

  it('does not show bid controls on a CPU/non-bid turn', async () => {
    renderWithProviders(<FrenchTarotPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'D ♥' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'プティット' })).not.toBeInTheDocument();
  });

  it('renders the chien phase and dispatches discard for exactly 6 selected cards', async () => {
    mockExec.mockResolvedValue(chienPhaseState);
    renderWithProviders(<FrenchTarotPage />);
    await screen.findByRole('button', { name: '2 ♥' });
    // Select the six heart cards (indices 0-5).
    for (const label of ['2 ♥', '3 ♥', '4 ♥', '5 ♥', '6 ♥', '7 ♥']) {
      fireEvent.click(screen.getByRole('button', { name: label }));
    }
    const discardBtn = screen.getByRole('button', { name: /捨てる/ });
    expect(discardBtn).toBeEnabled();
    mockExec.mockClear();
    mockExec.mockResolvedValue(chienPhaseState);
    fireEvent.click(discardBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', { cardIndices: [0, 1, 2, 3, 4, 5] }));
  });

  it('keeps the discard button disabled until exactly 6 cards are chosen', async () => {
    mockExec.mockResolvedValue(chienPhaseState);
    renderWithProviders(<FrenchTarotPage />);
    await screen.findByRole('button', { name: '2 ♥' });
    fireEvent.click(screen.getByRole('button', { name: '2 ♥' }));
    expect(screen.getByRole('button', { name: /捨てる/ })).toBeDisabled();
  });

  it('shows the revealed chien during the écart', async () => {
    mockExec.mockResolvedValue(chienPhaseState);
    renderWithProviders(<FrenchTarotPage />);
    await waitFor(() => expect(screen.getByTestId('frenchtarot-chien')).toBeInTheDocument());
  });

  it('surfaces distinct un-buriable reasons for the King and the Excuse during the écart', async () => {
    mockExec.mockResolvedValue(chienPhaseState);
    renderWithProviders(<FrenchTarotPage />);
    // The King card exposes the king-specific reason on its tooltip.
    const kingBtn = await screen.findByRole('button', { name: 'R ♠' });
    expect(kingBtn).toHaveAttribute('title', 'キング（ロワ）はシアンに埋められません。');
    // The Excuse exposes a different, excuse-specific reason.
    const excuseBtn = screen.getByRole('button', { name: 'Excuse ★' });
    expect(excuseBtn).toHaveAttribute('title', 'エクスキューズはシアンに埋められません。');
    // The two reasons differ.
    expect(kingBtn.getAttribute('title')).not.toBe(excuseBtn.getAttribute('title'));
  });

  it('shows no un-buriable tooltip on a freely buriable low card during the écart', async () => {
    mockExec.mockResolvedValue(chienPhaseState);
    renderWithProviders(<FrenchTarotPage />);
    const lowCard = await screen.findByRole('button', { name: '2 ♥' });
    expect(lowCard).not.toHaveAttribute('title');
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<FrenchTarotPage />);
    const card = await screen.findByRole('button', { name: 'D ♥' });
    fireEvent.click(card);
    const playBtn = await screen.findByRole('button', { name: '出す' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('renders trick end with the next trick button', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<FrenchTarotPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('renders round end with the next deal button and the deal result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<FrenchTarotPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のディール' })).toBeInTheDocument());
    expect(screen.getByTestId('frenchtarot-result')).toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<FrenchTarotPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝ち！')).toBeInTheDocument());
  });

  it('the next-game button at game end resets immediately', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<FrenchTarotPage />);
    const nextGame = await screen.findByRole('button', { name: '次のゲーム' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(gameEndState);
    fireEvent.click(nextGame);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', expect.anything()));
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<FrenchTarotPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'D ♥' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('changing the CPU difficulty and target-deals selects updates the config', async () => {
    renderWithProviders(<FrenchTarotPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'D ♥' })).toBeInTheDocument());
    const difficulty = screen.getByLabelText('CPU難易度') as HTMLSelectElement;
    fireEvent.change(difficulty, { target: { value: '2' } });
    expect(difficulty.value).toBe('2');
    const deals = screen.getByLabelText('マッチのディール数') as HTMLSelectElement;
    fireEvent.change(deals, { target: { value: '3' } });
    expect(deals.value).toBe('3');
  });

  it('renders the backend hint banner with its card indices', async () => {
    mockExec.mockResolvedValue(
      makeFrenchTarotState({
        hint: { cardIndices: [0, 2], reason: 'lead_high' },
        messageCode: 'frenchtarot.hintRequested',
      }),
    );
    renderWithProviders(<FrenchTarotPage />);
    await waitFor(() => expect(screen.getByText(/\[0\], \[2\]/)).toBeInTheDocument());
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` だけを見て描画すると常時表示になる (#4605)。
  it('hides the hint when it was not requested', async () => {
    mockExec.mockResolvedValue(
      makeFrenchTarotState({
        hint: { cardIndices: [0, 2], reason: 'lead_high' },
      }),
    );
    renderWithProviders(<FrenchTarotPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByText(/\[0\], \[2\]/)).not.toBeInTheDocument();
  });

  it('computes the bid-phase target from the bouts held (2 bouts → 41)', async () => {
    localStorage.clear();
    mockExec.mockReset();
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<FrenchTarotPage />);
    await waitFor(() => expect(screen.getByTestId('frenchtarot-bouts-target')).toHaveTextContent('41'));
  });

  it('computes the bid-phase target for an empty bout list (0 bouts → 56)', async () => {
    localStorage.clear();
    mockExec.mockReset();
    mockExec.mockResolvedValue(
      makeFrenchTarotState({
        phase: 0,
        isHumanBidTurn: true,
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 1,
            cards: [{ design: 'HEART' as const, value: 13, glyph: '♥', label: 'D', color: 'red', deck: 'tarot' }],
            trickCount: 0,
            cardPoints: 0,
            score: 0,
            isDeclarer: false,
          },
          { id: 1, isHuman: false, cardCount: 1, cards: [], trickCount: 0, cardPoints: 0, score: 0, isDeclarer: false },
          { id: 2, isHuman: false, cardCount: 1, cards: [], trickCount: 0, cardPoints: 0, score: 0, isDeclarer: false },
          { id: 3, isHuman: false, cardCount: 1, cards: [], trickCount: 0, cardPoints: 0, score: 0, isDeclarer: false },
        ],
      }),
    );
    renderWithProviders(<FrenchTarotPage />);
    await waitFor(() => expect(screen.getByTestId('frenchtarot-bouts-target')).toHaveTextContent('56'));
  });

  it('offers the bout target while bidding and withdraws it afterwards', async () => {
    localStorage.clear();
    mockExec.mockReset();
    mockExec.mockResolvedValue(bidPhaseState);
    const { unmount } = renderWithProviders(<FrenchTarotPage />);
    await waitFor(() => expect(screen.getByTestId('frenchtarot-bouts-target')).toBeInTheDocument());
    unmount();

    // In play the held bouts drain as they are played, and for a defender the
    // number never described the contract at all.
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<FrenchTarotPage />);
    await waitFor(() => expect(screen.getByTestId('frenchtarot-bouts-note')).toBeInTheDocument());
    expect(screen.queryByTestId('frenchtarot-bouts-target')).not.toBeInTheDocument();
  });
});
