import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { mariasApi } from '../api/gameApi';
import i18n from '../i18n';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeMariasState } from '../test/stateFactories';
import { MariasPage } from './MariasPage';

vi.mock('../api/gameApi', () => ({
  mariasApi: { exec: vi.fn() },
  actionLogApi: { marias: vi.fn() },
}));

const mockExec = vi.mocked(mariasApi.exec);

const playPhaseState = makeMariasState();
const trickEndState = makeMariasState({
  phase: 1,
  lastTrickWinner: 1,
  currentTrick: [
    { playerIdx: 0, card: { design: 'HEART', value: 12 } },
    { playerIdx: 1, card: { design: 'CLOVER', value: 13 } },
  ],
});
const roundEndState = makeMariasState({
  phase: 2,
  roundCardPoints: [55, 35, 30],
  roundMarriage: [40, 0, 0],
  roundMarriageSuits: [[{ suit: 3, points: 40 }], [], []],
});
const gameEndState = makeMariasState({
  phase: 3,
  gameEndFlag: true,
  winnerPlayer: 0,
  message: 'ゲーム終了！ あなたの勝ちです！',
});
const cpuTurnState = makeMariasState({ currentPlayerIdx: 1, isHumanTurn: false });

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('MariasPage', () => {
  it('renders player roles in English', async () => {
    const previousLanguage = i18n.language;
    await i18n.changeLanguage('en');
    try {
      mockExec.mockResolvedValue(makeMariasState());
      renderWithProviders(<MariasPage />);

      expect(await screen.findByText(/You: Soloist/)).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '♥ Q (trump)' })).toHaveAttribute('data-trump', 'true');
      expect(screen.getByText(/CPU 1: Defender/)).toBeInTheDocument();
      expect(screen.getByText(/CPU 2: Defender/)).toBeInTheDocument();
    } finally {
      await i18n.changeLanguage(previousLanguage);
    }
  });

  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<MariasPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<MariasPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, targetPoints: 10 },
      }),
    );
  });

  it('renders the play phase with the human cards and the Soloist badge', async () => {
    renderWithProviders(<MariasPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♥ Q')).toBeInTheDocument();
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
    });
    // The human (seat 0) is the default Soloist.
    expect(screen.getByText('ソリスト')).toBeInTheDocument();
  });

  it('marks trump cards in the hand and adds the trump label to their accessible names', async () => {
    renderWithProviders(<MariasPage />);
    const queen = await screen.findByRole('button', { name: '♥ Q (切り札)' });
    const king = screen.getByRole('button', { name: '♥ K (切り札)' });
    const ace = screen.getByRole('button', { name: '♠ A' });

    expect(queen).toHaveAttribute('data-trump', 'true');
    expect(king).toHaveAttribute('data-trump', 'true');
    expect(ace).not.toHaveAttribute('data-trump');
  });

  it('renders without a hand when the state has no human player', async () => {
    mockExec.mockResolvedValue(
      makeMariasState({
        players: makeMariasState().players.map((player) => ({ ...player, isHuman: false })),
      }),
    );
    renderWithProviders(<MariasPage />);

    expect(await screen.findByTestId('phase-indicator')).toBeInTheDocument();
  });

  it('shows each player role in the card and trick list and follows the next round Soloist', async () => {
    const nextRoundState = makeMariasState({
      roundNumber: 2,
      players: [
        { id: 0, isHuman: true, cardCount: 10, cards: [], trickCount: 0, score: 0, isSoloist: false },
        { id: 1, isHuman: false, cardCount: 10, cards: [], trickCount: 0, score: 0, isSoloist: true },
        { id: 2, isHuman: false, cardCount: 10, cards: [], trickCount: 0, score: 0, isSoloist: false },
      ],
    });
    mockExec.mockResolvedValueOnce(roundEndState).mockResolvedValueOnce(nextRoundState);
    renderWithProviders(<MariasPage />);

    const playerList = await screen.findByText(/あなた: ソリスト/);
    expect(playerList.parentElement).toHaveTextContent('ソリスト');
    expect(screen.getByText(/CPU 1: ディフェンダー/).parentElement).toHaveTextContent('ディフェンダー');
    expect(screen.getByText(/CPU 2: ディフェンダー/).parentElement).toHaveTextContent('ディフェンダー');

    fireEvent.click(screen.getByRole('button', { name: '次のラウンド' }));
    expect(await screen.findByText(/あなた: ディフェンダー/)).toHaveTextContent('ディフェンダー');
    expect(screen.getByText(/CPU 1: ソリスト/).parentElement).toHaveTextContent('ソリスト');
  });

  // **結婚ボーナスは配った時点で確定している (#4759)。**以前このバナーは毎
  // レンダー手札を走査して K と Q の両方を持っているかを見ていたので、どちらかを
  // 出した瞬間に消え、「出したのでボーナスを失った」という誤解を与えていた。
  it('shows the settled marriage bonus during play', async () => {
    mockExec.mockResolvedValue(
      makeMariasState({ roundMarriage: [40, 0, 0], roundMarriageSuits: [[{ suit: 3, points: 40 }], [], []] }),
    );
    renderWithProviders(<MariasPage />);
    const banner = await screen.findByTestId('marias-marriage');
    expect(banner).toHaveTextContent('ハート: 40');
  });

  // **これがこの issue の本体。**K を場に出しても点数は動かないので、バナーも
  // 消えてはいけない。
  it('keeps the banner after the king has been played', async () => {
    mockExec.mockResolvedValue(
      makeMariasState({
        // 手札から ♥K が消え、♥Q だけが残った状態。roundMarriage は確定のまま。
        roundMarriage: [40, 0, 0],
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 1,
            cards: [{ design: 'HEART', value: 12 }],
            trickCount: 1,
            score: 0,
            isSoloist: true,
          },
          { id: 1, isHuman: false, cardCount: 10, cards: [], trickCount: 0, score: 0, isSoloist: false },
          { id: 2, isHuman: false, cardCount: 10, cards: [], trickCount: 0, score: 0, isSoloist: false },
        ],
      }),
    );
    renderWithProviders(<MariasPage />);
    expect(await screen.findByTestId('marias-marriage')).toHaveTextContent('40');
  });

  it('announces the banner in a live region', async () => {
    mockExec.mockResolvedValue(makeMariasState({ roundMarriage: [40, 0, 0] }));
    renderWithProviders(<MariasPage />);
    const banner = await screen.findByTestId('marias-marriage');
    expect(banner).toHaveAttribute('role', 'status');
    expect(banner).toHaveAttribute('aria-live', 'polite');
  });

  it('shows no marriage banner when no marriage was dealt', async () => {
    mockExec.mockResolvedValue(
      makeMariasState({
        roundMarriage: [0, 0, 0],
        players: [
          {
            id: 0,
            isHuman: true,
            cardCount: 2,
            cards: [
              { design: 'HEART', value: 13 },
              { design: 'SPADE', value: 12 },
            ],
            trickCount: 0,
            score: 0,
            isSoloist: true,
          },
          { id: 1, isHuman: false, cardCount: 10, cards: [], trickCount: 0, score: 0, isSoloist: false },
          { id: 2, isHuman: false, cardCount: 10, cards: [], trickCount: 0, score: 0, isSoloist: false },
        ],
      }),
    );
    renderWithProviders(<MariasPage />);
    await waitFor(() => expect(screen.getByAltText('♥ K')).toBeInTheDocument());
    expect(screen.queryByTestId('marias-marriage')).not.toBeInTheDocument();
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<MariasPage />);
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
    renderWithProviders(<MariasPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のトリック' })).toBeInTheDocument());
    expect(screen.getByTestId('trick-winner-badge')).toHaveTextContent('CPU 1 が獲得');
  });

  it('does not render a winner badge while a trick is in progress', async () => {
    mockExec.mockResolvedValue(
      makeMariasState({ phase: 0, lastTrickWinner: 1, currentTrick: trickEndState.currentTrick.slice(0, 1) }),
    );
    renderWithProviders(<MariasPage />);
    await waitFor(() => expect(screen.getByTestId('trick-display-cards')).toBeInTheDocument());
    expect(screen.queryByTestId('trick-winner-badge')).not.toBeInTheDocument();
  });

  it('names each played trick card with its player and tracks trick updates', async () => {
    mockExec
      .mockResolvedValueOnce(trickEndState)
      .mockResolvedValueOnce(
        makeMariasState({ currentTrick: [{ playerIdx: 2, card: { design: 'SPADE', value: 1 } }] }),
      );
    renderWithProviders(<MariasPage />);

    expect(await screen.findByAltText('あなた: ♥ Q')).toBeInTheDocument();
    expect(screen.getByAltText('CPU 1: ♣ K')).toBeInTheDocument();
    expect(screen.queryByAltText('CPU 2:')).not.toBeInTheDocument();

    fireEvent.click(await screen.findByRole('button', { name: '次のトリック' }));
    expect(await screen.findByAltText('CPU 2: ♠ A')).toBeInTheDocument();
    expect(screen.queryByAltText('あなた: ♥ Q')).not.toBeInTheDocument();
  });

  it('renders round end with the next round button and the round result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<MariasPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.getByText('ラウンド結果')).toBeInTheDocument();
    expect(screen.getByText(/ハート: 40/)).toBeInTheDocument();
  });

  it('shows the Soloist-vs-Defenders total comparison, highlighting the winning Soloist', async () => {
    // Seat 0 is the Soloist: 55 + 40 = 95. Defenders (seats 1,2): 35 + 30 = 65.
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<MariasPage />);
    const row = await screen.findByTestId('marias-side-totals');
    const soloist = screen.getByText('ソリスト: 95点');
    const defenders = screen.getByText('ディフェンダー合計: 65点');
    expect(row).toContainElement(soloist);
    expect(row).toContainElement(defenders);
    // Soloist outscores the Defenders, so the Soloist side is emphasised.
    expect(soloist).toHaveClass('text-ds-warning');
    expect(defenders).not.toHaveClass('text-ds-warning');
  });

  it('shows confirmed per-player points and side totals during play', async () => {
    mockExec.mockResolvedValue(makeMariasState({ roundCardPoints: [25, 15, 10], roundMarriage: [40, 0, 20] }));
    renderWithProviders(<MariasPage />);
    const liveProgress = await screen.findByTestId('marias-round-progress-live');
    expect(liveProgress).toHaveAttribute('role', 'status');
    expect(liveProgress).toHaveAttribute('aria-live', 'polite');
    expect(liveProgress).toHaveAttribute('aria-atomic', 'true');
    expect(liveProgress).toHaveTextContent('あなた カード点: 25');
    expect(liveProgress).toHaveTextContent('ソリスト: 65点');
    const row = await screen.findByTestId('marias-side-totals');
    expect(screen.getByText('ラウンド途中経過（確定済み）')).toBeInTheDocument();
    expect(screen.getByText('あなた カード点: 25')).toBeInTheDocument();
    expect(screen.getByText('あなた マリッジ: 40')).toBeInTheDocument();
    expect(row).toHaveTextContent('ソリスト: 65点');
    expect(row).toHaveTextContent('ディフェンダー合計: 45点');
    expect(screen.queryByText('CPU 1 マリッジ: 0')).not.toBeInTheDocument();
    expect(screen.queryByText('CPU 2 マリッジ: 0')).not.toBeInTheDocument();
    expect(screen.getByText('ソリスト: 65点')).not.toHaveClass('text-ds-warning');
    expect(screen.getByText('ディフェンダー合計: 45点')).not.toHaveClass('text-ds-warning');
  });

  it('shows zero progress when a round has no points yet', async () => {
    mockExec.mockResolvedValue(makeMariasState({ roundCardPoints: [], roundMarriage: [] }));
    renderWithProviders(<MariasPage />);
    const liveProgress = await screen.findByTestId('marias-round-progress-live');
    expect(liveProgress).toHaveTextContent('あなた カード点: 0');
    expect(liveProgress).toHaveTextContent('CPU 1 カード点: 0');
    expect(liveProgress).toHaveTextContent('CPU 2 カード点: 0');
    const sideTotals = screen.getByTestId('marias-side-totals');
    expect(sideTotals).toHaveTextContent('ソリスト: 0点');
    expect(sideTotals).toHaveTextContent('ディフェンダー合計: 0点');
  });

  it('keeps the round result available inside the live region at game end', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<MariasPage />);
    const liveProgress = await screen.findByTestId('marias-round-progress-live');
    expect(liveProgress).toHaveAttribute('role', 'status');
    expect(liveProgress).toHaveTextContent('ラウンド結果');
  });

  it('shows confirmed points and side totals after a trick ends', async () => {
    mockExec.mockResolvedValue(
      makeMariasState({
        phase: 1,
        roundCardPoints: [20, 15, 10],
        roundMarriage: [40, 0, 0],
      }),
    );
    renderWithProviders(<MariasPage />);
    expect(await screen.findByTestId('marias-side-totals')).toHaveTextContent('ソリスト: 60点');
    expect(screen.getByText('あなた カード点: 20')).toBeInTheDocument();
  });

  it('highlights the Defenders when their combined total wins the round', async () => {
    // Soloist (seat 0): 20 + 0 = 20. Defenders (seats 1,2): 50 + 50 = 100.
    mockExec.mockResolvedValue(makeMariasState({ phase: 2, roundCardPoints: [20, 50, 50], roundMarriage: [0, 0, 0] }));
    renderWithProviders(<MariasPage />);
    const defenders = await screen.findByText('ディフェンダー合計: 100点');
    const soloist = screen.getByText('ソリスト: 20点');
    expect(defenders).toHaveClass('text-ds-warning');
    expect(soloist).not.toHaveClass('text-ds-warning');
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<MariasPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝ちです！')).toBeInTheDocument());
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<MariasPage />);
    await waitFor(() => expect(screen.getByAltText('♥ Q')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` だけを見て描画すると常時表示になる (#4605)。
  it('renders no hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, hint: { cardIndices: [0], reason: 'x' } });
    renderWithProviders(<MariasPage />);
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
      messageCode: 'marias.hintRequested',
    });
    renderWithProviders(<MariasPage />);
    expect(await screen.findByText(/\(\[0\]\)/)).toBeInTheDocument();
  });
});

// **切り札の決まり方が見えない。**入札でも固定でもなく、ソリストの手札という
// 非公開情報から機械的に決まるので、ソリスト以外にはブラックボックスだった (#6443)。
describe('MariasPage trump rule', () => {
  it('explains how trump was chosen, in the tooltip and to screen readers', async () => {
    renderWithProviders(<MariasPage />);

    const trump = await screen.findByTestId('marias-trump');
    // 記号の表示はそのまま。
    expect(trump).toHaveTextContent('切り札:');
    // ツールチップと読み上げの両方に規則が載る ── title だけだと
    // キーボードや支援技術には届かない。
    expect(trump).toHaveAttribute('title', '切り札はソリストの最長スートから自動で決まります');
    expect(trump.querySelector('.sr-only')?.textContent).toContain('切り札はソリストの最長スートから自動で決まります');
  });
});
