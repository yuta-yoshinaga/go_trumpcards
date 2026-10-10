import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ombreApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeOmbreState } from '../test/stateFactories';
import { OmbrePhase } from '../types/phases';
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
  highestBid: 0,
  bids: [0, 0, 0],
  bidActed: [false, false, false],
  bidTrump: [-1, -1, -1],
  highestBidderIdx: -1,
  ombreIdx: -1,
  trumpSuit: -1,
});
const trickEndState = makeOmbreState({
  phase: 2,
  lastTrickWinner: 1,
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
  it('shows live auction declarations and the current highest bidder', async () => {
    mockExec.mockResolvedValueOnce(bidPhaseState).mockResolvedValueOnce({
      ...bidPhaseState,
      bids: [0, 1, 2],
      bidActed: [true, true, true],
      bidTrump: [-1, 1, 4],
      highestBid: 2,
      highestBidderIdx: 2,
    });

    renderWithProviders(<OmbrePage />);

    const auction = await screen.findByRole('region', { name: '入札状況' });
    expect(auction).toHaveTextContent('あなた：未宣言');
    fireEvent.click(screen.getByRole('button', { name: 'パス' }));
    await waitFor(() => expect(auction).toHaveTextContent('現在の最高入札: CPU 2 — ソロ'));
    expect(auction).toHaveTextContent('現在の最高入札: CPU 2 — ソロ');
    expect(auction).toHaveTextContent('あなた：パス');
    expect(auction).toHaveTextContent('CPU 1：エントラール（スペード）');
    expect(auction).toHaveTextContent('CPU 2：ソロ（ダイヤ）（最高入札）');
  });
  it.each([
    ['play', OmbrePhase.PLAY],
    ['trick end', OmbrePhase.TRICK_END],
  ])('keeps the completed bid history visible during %s', async (_phaseName, phase) => {
    mockExec.mockResolvedValue(
      makeOmbreState({
        phase,
        bids: [1, 0, 2],
        bidActed: [true, true, true],
        bidTrump: [1, -1, 4],
        highestBid: 2,
        highestBidderIdx: 2,
      }),
    );

    renderWithProviders(<OmbrePage />);

    const auction = await screen.findByRole('region', { name: '入札状況' });
    expect(auction).toHaveTextContent('あなた：エントラール（スペード）');
    expect(auction).toHaveTextContent('CPU 1：パス');
    expect(auction).toHaveTextContent('CPU 2：ソロ（ダイヤ）（最高入札）');
    expect(auction).not.toHaveTextContent('現在の最高入札');
  });
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
    expect(
      within(screen.getByTestId('game-footer-actions')).getByRole('button', { name: 'エントラール' }),
    ).toBeInTheDocument();
  });

  it('omits the pinned action row when the CPU is playing', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<OmbrePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByTestId('game-footer-actions')).not.toBeInTheDocument();
  });

  it('marks bids at or below the highest bid unavailable and keeps pass selectable', async () => {
    mockExec.mockResolvedValue({ ...bidPhaseState, highestBid: 1 });
    renderWithProviders(<OmbrePage />);
    const entrar = await screen.findByRole('button', { name: 'エントラール' });
    const solo = screen.getByRole('button', { name: 'ソロ' });
    const pass = screen.getByRole('button', { name: 'パス' });
    expect(entrar).toHaveAttribute('aria-disabled', 'true');
    expect(entrar).toHaveAttribute('aria-describedby');
    expect(solo).not.toHaveAttribute('aria-disabled', 'true');
    expect(pass).not.toHaveAttribute('aria-disabled', 'true');

    mockExec.mockClear();
    fireEvent.click(entrar);
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
    expect(screen.getByTestId('ombre-bid-stage1')).toBeInTheDocument();
  });

  it('disables both declarations when solo is the current winning bid', async () => {
    mockExec.mockResolvedValue({ ...bidPhaseState, highestBid: 2 });
    renderWithProviders(<OmbrePage />);
    const entrar = await screen.findByRole('button', { name: 'エントラール' });
    const solo = screen.getByRole('button', { name: 'ソロ' });
    const pass = screen.getByRole('button', { name: 'パス' });
    expect(entrar).toHaveAttribute('aria-disabled', 'true');
    expect(solo).toHaveAttribute('aria-disabled', 'true');
    expect(pass).not.toHaveAttribute('aria-disabled', 'true');
    expect(entrar).toHaveAccessibleDescription('現在の入札を上回る必要があります');
    expect(solo).toHaveAccessibleDescription('現在の入札を上回る必要があります');

    mockExec.mockClear();
    fireEvent.click(solo);
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
    expect(screen.getByTestId('ombre-bid-stage1')).toBeInTheDocument();
  });

  it('keeps both declarations selectable when no bid has been made', async () => {
    mockExec.mockResolvedValue({ ...bidPhaseState, highestBid: 0 });
    renderWithProviders(<OmbrePage />);
    const entrar = await screen.findByRole('button', { name: 'エントラール' });
    const solo = screen.getByRole('button', { name: 'ソロ' });
    expect(entrar).not.toHaveAttribute('aria-disabled', 'true');
    expect(solo).not.toHaveAttribute('aria-disabled', 'true');
    fireEvent.click(entrar);
    await waitFor(() => expect(screen.getByTestId('ombre-bid-stage2')).toBeInTheDocument());
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

  it('requests a play hint and marks the recommended card', async () => {
    mockExec.mockResolvedValueOnce(playPhaseState).mockResolvedValueOnce({
      ...playPhaseState,
      hint: { cardIndices: [0], reason: 'lead_high' },
      messageCode: 'ombre.hintRequested',
    });
    renderWithProviders(<OmbrePage />);
    const hintButton = await screen.findByRole('button', { name: 'ヒントを表示' });
    fireEvent.click(hintButton);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
    const hintLive = screen.getByTestId('ombre-hint-live');
    await waitFor(() => {
      expect(hintLive.textContent).toContain('♥ Q');
      expect(hintLive.textContent).toContain('強い札でリード');
    });
    expect(screen.getByAltText('♥ Q').closest('button')).toHaveStyle({ border: '3px solid var(--color-ds-warning)' });
  });

  it('renders trick end with the next trick button', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<OmbrePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
    expect(screen.getByTestId('trick-winner-badge')).toHaveTextContent('CPU 1 が獲得');
  });

  it('does not render a winner badge while a trick is in progress', async () => {
    mockExec.mockResolvedValue(
      makeOmbreState({
        phase: 1,
        lastTrickWinner: 1,
        currentTrick: [{ playerIdx: 1, card: { design: 'HEART', value: 12 } }],
      }),
    );
    renderWithProviders(<OmbrePage />);
    await waitFor(() => expect(screen.getByTestId('trick-display-cards')).toBeInTheDocument());
    expect(screen.queryByTestId('trick-winner-badge')).not.toBeInTheDocument();
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
    expect(screen.getByTestId('ombre-hint-live')).not.toHaveTextContent('♥ Q');
  });

  // **押したときは出る。**押していない側だけを見ていると、`isRequestedHint` を
  // 定数 false にしても通ってしまう。真の分岐も踏んでおく。
  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue({
      ...playPhaseState,
      hint: { cardIndices: [0], reason: 'x' },
      messageCode: 'ombre.hintRequested',
    });
    renderWithProviders(<OmbrePage />);
    await waitFor(() => expect(screen.getByTestId('ombre-hint-live').textContent).toContain('♥ Q'));
  });

  // **全員パスによる強制 Entrar は通常の宣言と区別が付かなかった** ── 差は
  // ombreIdx がディーラーと一致することだけで、それは偶然そうなる局とも
  // 見分けが付かない (#6485)。訳文に解決することを画面で見る。
  it('spells out the forced Entrar after an all-pass auction', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, messageCode: 'ombre.forcedEntrar' });
    renderWithProviders(<OmbrePage />);

    const line = await screen.findByText(/全員パス/);
    expect(line).toBeInTheDocument();
    // 生の識別子も未置換のプレースホルダも残らない。
    expect(line.textContent).not.toContain('ombre.');
    expect(line.textContent).not.toContain('{{');
  });

  // **催促は常設のライブ領域の中にある (#6880)。** フェーズ切り替えで現れる
  // テキストなので、領域が無いとスクリーンリーダには何も届かない。領域を
  // 出現と同時に付けても読み上げられないため、常設にして中身だけ差し替える。
  it('announces the prompt from an always-mounted live region', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<OmbrePage />);

    const live = await screen.findByTestId('ombre-prompt-live');
    expect(live).toHaveAttribute('role', 'status');
    expect(live).toHaveAttribute('aria-live', 'polite');
    // 催促が**その領域の中**にあること。隣に置いただけの実装は属性の検査を通る。
    expect(live).toContainElement(await screen.findByTestId('ombre-bid-prompt'));
  });

  it('shows every player name with the deal score change at round end', async () => {
    mockExec.mockResolvedValue(
      makeOmbreState({
        phase: 3,
        outcome: 3,
        playerScoreDeltas: [-4, 2, 2],
      }),
    );
    renderWithProviders(<OmbrePage />);
    expect(await screen.findByText('今回の増減')).toBeInTheDocument();
    expect(screen.getByText('あなた: -4')).toBeInTheDocument();
    expect(screen.getByText('CPU 1: +2')).toBeInTheDocument();
    expect(screen.getByText('CPU 2: +2')).toBeInTheDocument();
  });
});
