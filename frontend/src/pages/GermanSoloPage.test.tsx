import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { germansoloApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeGermanSoloState } from '../test/stateFactories';
import { GermanSoloPhase } from '../types/phases';
import { GermanSoloPage } from './GermanSoloPage';

vi.mock('../api/gameApi', () => ({
  germansoloApi: { exec: vi.fn() },
  actionLogApi: { germansolo: vi.fn() },
}));

const mockExec = vi.mocked(germansoloApi.exec);

const playPhaseState = makeGermanSoloState();
const bidPhaseState = makeGermanSoloState({
  phase: GermanSoloPhase.BID,
  currentBidderIdx: 0,
  isHumanTurn: false,
  isHumanBidTurn: true,
  winningBid: 0,
  highestBid: 0,
  biddableBids: [2, 3, 4],
  declarerIdx: -1,
  playsAlone: false,
  trumpSuit: -1,
});
const trickEndState = makeGermanSoloState({
  phase: GermanSoloPhase.TRICK_END,
  currentTrick: [
    { playerIdx: 0, card: { design: 'HEART', value: 12 } },
    { playerIdx: 1, card: { design: 'CLOVER', value: 13 } },
  ],
});
const roundEndState = makeGermanSoloState({
  phase: GermanSoloPhase.ROUND_END,
  outcome: 1,
});
const gameEndState = makeGermanSoloState({
  phase: GermanSoloPhase.GAME_END,
  gameEndFlag: true,
  winnerPlayer: 0,
  message: 'ゲーム終了！ あなたの勝ち！',
});
const cpuTurnState = makeGermanSoloState({ currentPlayerIdx: 1, isHumanTurn: false });

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('GermanSoloPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<GermanSoloPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<GermanSoloPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, targetRounds: 5 },
      }),
    );
  });

  it('renders the play phase with the human cards and the declarer badge', async () => {
    renderWithProviders(<GermanSoloPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♣ Q')).toBeInTheDocument();
      expect(screen.getByAltText('♥ A')).toBeInTheDocument();
    });
    // The human (seat 0) is the default declarer — the badge renders.
    expect(screen.getByText('宣言者')).toBeInTheDocument();
  });

  // **必要トリック数は契約ごとに違う。** 出さないと、5 取って喜んだ Tout が
  // その場で失敗になる理由が読めない。
  it('shows the contract target and the running trick split', async () => {
    mockExec.mockResolvedValue(makeGermanSoloState({ requiredTricks: 8, declarerTricks: 3, defenderTricks: 1 }));
    renderWithProviders(<GermanSoloPage />);
    const line = await screen.findByTestId('germansolo-contract-line');
    expect(line).toHaveTextContent('必要 8 トリック');
    expect(line).toHaveTextContent('宣言側 3');
    expect(line).toHaveTextContent('守備側 1');
  });

  it('renders the bid phase with the three contracts and pass', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<GermanSoloPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'フラーゲ' })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'ソロ' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'トゥー' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'パス' })).toBeInTheDocument();
  });

  // **サーバが弾く選択肢は出さない。** biddableBids は「この席がいま上回れる
  // 契約」そのもので、既に Solo が立っていれば残るのは Tout だけ。
  it('offers only the contracts the server would still accept', async () => {
    mockExec.mockResolvedValue(makeGermanSoloState({ ...bidPhaseState, highestBid: 3, biddableBids: [4] }));
    renderWithProviders(<GermanSoloPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'トゥー' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'フラーゲ' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'ソロ' })).not.toBeInTheDocument();
    // Passing is always allowed.
    expect(screen.getByRole('button', { name: 'パス' })).toBeInTheDocument();
  });

  it('stages Frage → trump selection → confirm and dispatches the bid with the suit', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<GermanSoloPage />);
    // Stage 1: only contract buttons, no trump/confirm yet.
    await screen.findByTestId('germansolo-bid-stage1');
    expect(screen.queryByTestId('germansolo-bid-stage2')).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'スペード' })).not.toBeInTheDocument();

    // Choose Frage → advance to stage 2 (trump + confirm/back).
    fireEvent.click(screen.getByRole('button', { name: 'フラーゲ' }));
    await screen.findByTestId('germansolo-bid-stage2');
    expect(screen.getByTestId('germansolo-bid-confirm')).toBeDisabled();

    // Pick spades (♠) as trump → confirm enabled.
    fireEvent.click(screen.getByRole('button', { name: 'スペード' }));
    expect(screen.getByTestId('germansolo-bid-confirm')).toBeEnabled();

    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(screen.getByTestId('germansolo-bid-confirm'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 2, trumpSuit: 1 }));
  });

  it('stages Tout → trump selection → confirm and dispatches Tout with the suit', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<GermanSoloPage />);
    await screen.findByTestId('germansolo-bid-stage1');
    fireEvent.click(screen.getByRole('button', { name: 'トゥー' }));
    await screen.findByTestId('germansolo-bid-stage2');
    fireEvent.click(screen.getByRole('button', { name: 'ハート' }));
    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(screen.getByTestId('germansolo-bid-confirm'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 4, trumpSuit: 3 }));
  });

  it('back returns from trump selection to contract selection without dispatching', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<GermanSoloPage />);
    await screen.findByTestId('germansolo-bid-stage1');
    fireEvent.click(screen.getByRole('button', { name: 'フラーゲ' }));
    await screen.findByTestId('germansolo-bid-stage2');
    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('germansolo-bid-back'));
    // Back to stage 1; no bid dispatched.
    await screen.findByTestId('germansolo-bid-stage1');
    expect(screen.queryByTestId('germansolo-bid-stage2')).not.toBeInTheDocument();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('passing dispatches bid with bid=0 in one tap and no trump requirement', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<GermanSoloPage />);
    const passBtn = await screen.findByRole('button', { name: 'パス' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(passBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 0, trumpSuit: undefined }));
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<GermanSoloPage />);
    const card = await screen.findByAltText('♣ Q');
    fireEvent.click(card);
    const playBtn = await screen.findByRole('button', { name: '出す' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('renders trick end with the next trick button', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<GermanSoloPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
  });

  it('renders round end with the next deal button and the deal result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<GermanSoloPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のディール' })).toBeInTheDocument());
    expect(screen.getByText('ディール結果')).toBeInTheDocument();
    expect(screen.getByText(/成功（契約達成）/)).toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<GermanSoloPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝ち！')).toBeInTheDocument());
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<GermanSoloPage />);
    await waitFor(() => expect(screen.getByAltText('♣ Q')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('badges Spadille (♣Q) and the trump Manille in the hand once trump is decided', async () => {
    // Default state: trump = spades, hand = [♣Q, ♠7, ♥A] → Spadille + Manille.
    renderWithProviders(<GermanSoloPage />);
    await waitFor(() => expect(screen.getByAltText('♣ Q')).toBeInTheDocument());
    const spadille = screen.getByTestId('card-role-badge-0');
    expect(spadille).toHaveTextContent('1');
    expect(spadille).toHaveAttribute('title', 'スパディーユ (♣Q)');
    expect(screen.getByTestId('card-role-badge-1')).toHaveTextContent('2'); // ♠7 = Manille
    expect(screen.queryByTestId('card-role-badge-2')).not.toBeInTheDocument(); // ♥A is an ordinary card
  });

  it('badges all three matadors including the trump-suit Manille (heart trump → ♥7)', async () => {
    const matadorHand = makeGermanSoloState({
      trumpSuit: 3, // hearts
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 3,
          cards: [
            { design: 'CLOVER', value: 12 }, // Spadille → 1
            { design: 'SPADE', value: 12 }, // Basta → 3
            { design: 'HEART', value: 7 }, // Manille (heart trump) → 2
          ],
          trickCount: 0,
          score: 0,
          isDeclarer: true,
        },
        { id: 1, isHuman: false, cardCount: 3, cards: [], trickCount: 0, score: 0, isDeclarer: false },
        { id: 2, isHuman: false, cardCount: 3, cards: [], trickCount: 0, score: 0, isDeclarer: false },
        { id: 3, isHuman: false, cardCount: 3, cards: [], trickCount: 0, score: 0, isDeclarer: false },
      ],
      playableIndices: [0, 1, 2],
    });
    mockExec.mockResolvedValue(matadorHand);
    renderWithProviders(<GermanSoloPage />);
    await waitFor(() => expect(screen.getByAltText('♣ Q')).toBeInTheDocument());
    expect(screen.getByTestId('card-role-badge-0')).toHaveTextContent('1'); // Spadille
    expect(screen.getByTestId('card-role-badge-1')).toHaveTextContent('3'); // Basta
    expect(screen.getByTestId('card-role-badge-2')).toHaveTextContent('2'); // Manille
  });

  it('shows no matador badge while trump is undecided (bid phase)', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<GermanSoloPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'フラーゲ' })).toBeInTheDocument());
    expect(screen.queryByTestId('card-role-badge-0')).not.toBeInTheDocument();
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` だけを見て描画すると常時表示になる (#4605)。
  it('renders no hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, hint: { cardIndices: [0], reason: 'x' } });
    renderWithProviders(<GermanSoloPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    // バナーは推奨札の位置を `([0])` の形で含む。トグルのラベル (「ヒント表示」)
    // と紛れないよう、そこで判定する。
    expect(screen.queryByText(/\(\[0\]\)/)).not.toBeInTheDocument();
  });

  // **押したときは出る。**押していない側だけを見ていると、`isRequestedHint` を
  // 定数 false にしても通ってしまう。真の分岐も踏んでおく。
  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue({
      ...playPhaseState,
      hint: { cardIndices: [0], reason: 'x' },
      messageCode: 'germansolo.hintRequested',
    });
    renderWithProviders(<GermanSoloPage />);
    expect(await screen.findByText(/\(\[0\]\)/)).toBeInTheDocument();
  });
});

describe('GermanSoloPage ace call', () => {
  // Frage を落札した直後はエース呼びフェーズ。**ここを抜ける操作面が無いと
  // 盤面は固まる。**
  it('offers the callable aces and dispatches the call', async () => {
    mockExec.mockResolvedValue(
      makeGermanSoloState({
        phase: GermanSoloPhase.ACE_CALL,
        winningBid: 2,
        playsAlone: false,
        isHumanTurn: false,
        isHumanAceCallTurn: true,
        calledAceSuit: -1,
        callableAceSuits: [2, 3],
      }),
    );
    renderWithProviders(<GermanSoloPage />);
    await waitFor(() => expect(screen.getByTestId('germansolo-ace-call')).toBeInTheDocument());

    // 呼べるエースだけがボタンになる。
    expect(screen.getByRole('button', { name: 'クラブ' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'スペード' })).not.toBeInTheDocument();

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: 'ハート' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(mockExec).toHaveBeenCalledWith('ace', expect.objectContaining({ aceSuit: 3 }));
  });

  it('hides the ace controls once play has started', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<GermanSoloPage />);
    await waitFor(() => expect(screen.getByTestId('germansolo-ace-line')).toBeInTheDocument());
    expect(screen.queryByTestId('germansolo-ace-call')).not.toBeInTheDocument();
  });

  // **持ち主は呼ばれたエースが場に出るまで伏せる。** サーバが partnerIdx=-1 を
  // 返している間、画面は誰の名前も出してはいけない。
  it('does not name the partner while the called ace is still hidden', async () => {
    mockExec.mockResolvedValue(makeGermanSoloState({ playsAlone: false, calledAceSuit: 3, partnerIdx: -1 }));
    renderWithProviders(<GermanSoloPage />);
    await waitFor(() => expect(screen.getByTestId('germansolo-ace-line')).toBeInTheDocument());
    expect(screen.getByTestId('germansolo-ace-line')).toHaveTextContent('伏せられています');
  });

  it('names the partner once the called ace has been played', async () => {
    mockExec.mockResolvedValue(makeGermanSoloState({ playsAlone: false, calledAceSuit: 3, partnerIdx: 2 }));
    renderWithProviders(<GermanSoloPage />);
    await waitFor(() => expect(screen.getByTestId('germansolo-ace-line')).toBeInTheDocument());
    const line = screen.getByTestId('germansolo-ace-line');
    expect(line).toHaveTextContent('味方');
    expect(line).not.toHaveTextContent('伏せられています');
  });

  it('says playing alone for Solo and Tout', async () => {
    mockExec.mockResolvedValue(makeGermanSoloState({ playsAlone: true, calledAceSuit: -1 }));
    renderWithProviders(<GermanSoloPage />);
    await waitFor(() => expect(screen.getByTestId('germansolo-ace-line')).toBeInTheDocument());
    expect(screen.getByTestId('germansolo-ace-line')).toHaveTextContent('単独プレイ');
  });
});
