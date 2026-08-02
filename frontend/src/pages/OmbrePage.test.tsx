import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ombreApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeOmbreState } from '../test/stateFactories';
import { OmbrePage } from './OmbrePage';

vi.mock('../api/gameApi', () => ({
  ombreApi: { exec: vi.fn() },
  actionLogApi: { ombre: vi.fn() },
}));

const mockExec = vi.mocked(ombreApi.exec);

const playPhaseState = makeOmbreState();
const bidPhaseState = makeOmbreState({
  phase: 0,
  currentBidderIdx: 0,
  isHumanTurn: false,
  isHumanBidTurn: true,
  winningBid: 0,
  ombreIdx: -1,
  trumpSuit: -1,
});
const trickEndState = makeOmbreState({
  phase: 2,
  currentTrick: [
    { playerIdx: 0, card: { design: 'HEART', value: 12 } },
    { playerIdx: 1, card: { design: 'CLOVER', value: 13 } },
  ],
});
const roundEndState = makeOmbreState({
  phase: 3,
  outcome: 1,
});
const gameEndState = makeOmbreState({
  phase: 4,
  gameEndFlag: true,
  winnerPlayer: 0,
  message: 'ゲーム終了！ あなたの勝ち！',
});
const cpuTurnState = makeOmbreState({ currentPlayerIdx: 1, isHumanTurn: false });

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('OmbrePage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<OmbrePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<OmbrePage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, targetRounds: 5 },
      }),
    );
  });

  it('renders the play phase with the human cards and the Ombre badge', async () => {
    renderWithProviders(<OmbrePage />);
    await waitFor(() => {
      expect(screen.getByAltText('♥ Q')).toBeInTheDocument();
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
    });
    // The human (seat 0) is the default Ombre — the badge renders (heading also reads オンブル).
    expect(screen.getAllByText('オンブル').length).toBeGreaterThan(1);
  });

  it('renders the bid phase with entrar, solo and pass buttons', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<OmbrePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'エントラール' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ソロ' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'パス' })).toBeInTheDocument();
  });

  it('stages entrar → trump selection → confirm and dispatches the bid with the suit', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<OmbrePage />);
    // Stage 1: only bid-type buttons, no trump/confirm yet.
    await screen.findByTestId('ombre-bid-stage1');
    expect(screen.queryByTestId('ombre-bid-stage2')).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'スペード' })).not.toBeInTheDocument();

    // Choose entrar → advance to stage 2 (trump + confirm/back).
    fireEvent.click(screen.getByRole('button', { name: 'エントラール' }));
    await screen.findByTestId('ombre-bid-stage2');
    const confirmBtn = screen.getByTestId('ombre-bid-confirm');
    expect(confirmBtn).toBeDisabled();

    // Pick spades (♠) as trump → confirm enabled.
    fireEvent.click(screen.getByRole('button', { name: 'スペード' }));
    expect(screen.getByTestId('ombre-bid-confirm')).toBeEnabled();

    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(screen.getByTestId('ombre-bid-confirm'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 1, trumpSuit: 1 }));
  });

  it('stages solo → trump selection → confirm and dispatches solo with the suit', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<OmbrePage />);
    await screen.findByTestId('ombre-bid-stage1');
    fireEvent.click(screen.getByRole('button', { name: 'ソロ' }));
    await screen.findByTestId('ombre-bid-stage2');
    fireEvent.click(screen.getByRole('button', { name: 'ハート' }));
    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(screen.getByTestId('ombre-bid-confirm'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 2, trumpSuit: 3 }));
  });

  it('back returns from trump selection to bid-type selection without dispatching', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<OmbrePage />);
    await screen.findByTestId('ombre-bid-stage1');
    fireEvent.click(screen.getByRole('button', { name: 'エントラール' }));
    await screen.findByTestId('ombre-bid-stage2');
    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('ombre-bid-back'));
    // Back to stage 1; no bid dispatched.
    await screen.findByTestId('ombre-bid-stage1');
    expect(screen.queryByTestId('ombre-bid-stage2')).not.toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('passing dispatches bid with bid=0 in one tap and no trump requirement', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<OmbrePage />);
    const passBtn = await screen.findByRole('button', { name: 'パス' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(passBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 0, trumpSuit: undefined }));
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<OmbrePage />);
    const card = await screen.findByAltText('♥ Q');
    fireEvent.click(card);
    const playBtn = await screen.findByRole('button', { name: '出す' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('renders trick end with the next trick button', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<OmbrePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('renders round end with the next deal button and the deal result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<OmbrePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のディール' })).toBeInTheDocument());
    expect(screen.getByText('ディール結果')).toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<OmbrePage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝ち！')).toBeInTheDocument());
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<OmbrePage />);
    await waitFor(() => expect(screen.getByAltText('♥ Q')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('badges Spadille (♠A) in the hand when trump is decided', async () => {
    // Default state: trump = spades, hand[2] = ♠A → Spadille (rank 1).
    renderWithProviders(<OmbrePage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    const badge = screen.getByTestId('card-role-badge-2');
    expect(badge).toHaveTextContent('1');
    expect(badge).toHaveAttribute('title', 'スパディーユ (♠A)');
  });

  it('badges all three matadors including the trump-suit Manille (heart trump → ♥7)', async () => {
    const matadorHand = makeOmbreState({
      trumpSuit: 3, // hearts
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 3,
          cards: [
            { design: 'SPADE', value: 1 }, // Spadille → 1
            { design: 'CLOVER', value: 1 }, // Basto → 3
            { design: 'HEART', value: 7 }, // Manille (heart trump) → 2
          ],
          trickCount: 0,
          score: 0,
          isOmbre: true,
        },
        { id: 1, isHuman: false, cardCount: 3, cards: [], trickCount: 0, score: 0, isOmbre: false },
        { id: 2, isHuman: false, cardCount: 3, cards: [], trickCount: 0, score: 0, isOmbre: false },
      ],
      playableIndices: [0, 1, 2],
    });
    mockExec.mockResolvedValue(matadorHand);
    renderWithProviders(<OmbrePage />);
    await waitFor(() => expect(screen.getByAltText('♠ A')).toBeInTheDocument());
    expect(screen.getByTestId('card-role-badge-0')).toHaveTextContent('1'); // Spadille
    expect(screen.getByTestId('card-role-badge-1')).toHaveTextContent('3'); // Basto
    expect(screen.getByTestId('card-role-badge-2')).toHaveTextContent('2'); // Manille
  });

  it('shows no matador badge while trump is undecided (bid phase)', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<OmbrePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'エントラール' })).toBeInTheDocument());
    expect(screen.queryByTestId('card-role-badge-2')).not.toBeInTheDocument();
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` だけを見て描画すると常時表示になる (#4605)。
  it('renders no hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, hint: { cardIndices: [0], reason: 'x' } });
    renderWithProviders(<OmbrePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    // バナーは推奨札の位置を `([0])` の形で含む。トグルのラベル (「ヒント表示」)
    // と紛れないよう、そこで判定する。
    expect(screen.queryByText(/\(\[0\]\)/)).not.toBeInTheDocument();
  });
});
