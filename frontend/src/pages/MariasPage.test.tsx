import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { mariasApi } from '../api/gameApi';
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
  currentTrick: [
    { playerIdx: 0, card: { design: 'HEART', value: 12 } },
    { playerIdx: 1, card: { design: 'CLOVER', value: 13 } },
  ],
});
const roundEndState = makeMariasState({
  phase: 2,
  roundCardPoints: [55, 35, 30],
  roundMarriage: [40, 0, 0],
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

  // **結婚ボーナスは配った時点で確定している (#4759)。**以前このバナーは毎
  // レンダー手札を走査して K と Q の両方を持っているかを見ていたので、どちらかを
  // 出した瞬間に消え、「出したのでボーナスを失った」という誤解を与えていた。
  it('shows the settled marriage bonus during play', async () => {
    mockExec.mockResolvedValue(makeMariasState({ roundMarriage: [40, 0, 0] }));
    renderWithProviders(<MariasPage />);
    const banner = await screen.findByTestId('marias-marriage');
    expect(banner).toHaveTextContent('40');
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
  });

  it('renders round end with the next round button and the round result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<MariasPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
    expect(screen.getByText('ラウンド結果')).toBeInTheDocument();
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
