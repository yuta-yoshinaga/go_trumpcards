import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { cinchApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeCinchState } from '../test/stateFactories';
import { CinchPage } from './CinchPage';

vi.mock('../api/gameApi', () => ({
  cinchApi: { exec: vi.fn() },
  actionLogApi: { cinch: vi.fn() },
}));

const mockExec = vi.mocked(cinchApi.exec);

const playPhaseState = makeCinchState();
const bidPhaseState = makeCinchState({
  phase: 0,
  bidPlayerIdx: 0,
  bidWinnerIdx: -1,
  currentBid: 0,
  isHumanTurn: true,
});
const nameTrumpState = makeCinchState({
  phase: 1,
  bidWinnerIdx: 0,
  trumpSuit: 0,
  isHumanTurn: true,
});
const roundEndState = makeCinchState({
  phase: 4,
  lastDealDetail: {
    trumpSuit: 1,
    bidderIdx: 0,
    bid: 6,
    setBack: false,
    points: { 0: 8, 1: 2, 2: 2, 3: 2 },
    gained: { 0: 6, 1: -1, 2: -1, 3: -1 },
  },
});
const setBackState = makeCinchState({
  phase: 4,
  lastDealDetail: {
    trumpSuit: 1,
    bidderIdx: 1,
    bid: 8,
    setBack: true,
    points: { 0: 4, 1: 5, 2: 3, 3: 2 },
    gained: { 0: 1, 1: -8, 2: 1, 3: 1 },
  },
});
const gameEndState = makeCinchState({
  phase: 5,
  gameEndFlag: true,
  winnerIdx: 0,
  message: 'ゲーム終了！ あなたの勝ち！',
});
const cpuTurnState = makeCinchState({ currentTurn: 1, isHumanTurn: false });

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('CinchPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<CinchPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount with the default config', async () => {
    renderWithProviders(<CinchPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, pointLimit: 21 },
      }),
    );
  });

  it('renders the play phase with the human cards', async () => {
    renderWithProviders(<CinchPage />);
    await waitFor(() => {
      expect(screen.getByAltText('♥ Q')).toBeInTheDocument();
      expect(screen.getByAltText('♠ A')).toBeInTheDocument();
    });
  });

  it('renders the bid phase with pass and numeric bid buttons', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<CinchPage />);
    await waitFor(() => expect(screen.getByTestId('cinch-bid-prompt')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'パス' })).toBeInTheDocument();
    // A raise button for a legal bid (e.g. 1) is present.
    expect(screen.getByRole('button', { name: '1' })).toBeInTheDocument();
  });

  it('passing dispatches bid with bid=0', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<CinchPage />);
    const passBtn = await screen.findByRole('button', { name: 'パス' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(passBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 0 }));
  });

  it('bidding a number dispatches bid with that value', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<CinchPage />);
    const sixBtn = await screen.findByRole('button', { name: '6' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(bidPhaseState);
    fireEvent.click(sixBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 6 }));
  });

  it('shows the bid-strength guide with the estimated point range and the 14-point legend', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<CinchPage />);
    const guide = await screen.findByTestId('cinch-bid-strength');
    // Default bid hand (K♥, Q♥, A♠) holds 1 point for spades/hearts, 0 otherwise.
    expect(screen.getByTestId('cinch-bid-strength-range')).toHaveTextContent('0〜1点');
    expect(screen.getByTestId('cinch-bid-strength-best')).toHaveTextContent('スペード');
    // The 14-point composition legend is present.
    expect(guide).toHaveTextContent('14点の構成');
    expect(guide).toHaveTextContent('Right Pedro');
  });

  it('does not show the bid-strength guide outside the human bid turn', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<CinchPage />);
    await waitFor(() => expect(screen.getByTestId('cinch-trump-header')).toBeInTheDocument());
    expect(screen.queryByTestId('cinch-bid-strength')).not.toBeInTheDocument();
  });

  it('renders the name-trump phase with four suit buttons', async () => {
    mockExec.mockResolvedValue(nameTrumpState);
    renderWithProviders(<CinchPage />);
    await waitFor(() => expect(screen.getByTestId('cinch-trump-prompt')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: 'スペード' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'ダイヤ' })).toBeInTheDocument();
  });

  it('naming a trump suit dispatches trump with the suit index', async () => {
    mockExec.mockResolvedValue(nameTrumpState);
    renderWithProviders(<CinchPage />);
    const spadeBtn = await screen.findByRole('button', { name: 'スペード' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(nameTrumpState);
    fireEvent.click(spadeBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('trump', { trumpSuit: 1 }));
  });

  it('colors red suits red and keeps black suits default on the trump buttons', async () => {
    mockExec.mockResolvedValue(nameTrumpState);
    renderWithProviders(<CinchPage />);
    const diamondBtn = await screen.findByRole('button', { name: 'ダイヤ' });
    expect(diamondBtn.querySelector('span')?.className).toContain('text-ds-error');
    const spadeBtn = screen.getByRole('button', { name: 'スペード' });
    expect(spadeBtn.querySelector('span')?.className ?? '').not.toContain('text-ds-error');
  });

  it('shows the trump suit name and a red symbol in the header when declared', async () => {
    mockExec.mockResolvedValue(makeCinchState({ trumpSuit: 3, isHumanTurn: false }));
    renderWithProviders(<CinchPage />);
    const header = await screen.findByTestId('cinch-trump-header');
    expect(header).toHaveTextContent('ハート');
    // The ♥ symbol is wrapped in a red span.
    expect(header.querySelector('.text-ds-error')?.textContent).toBe('♥');
  });

  it('selecting a card then playing dispatches play', async () => {
    renderWithProviders(<CinchPage />);
    const card = await screen.findByAltText('♥ Q');
    fireEvent.click(card);
    const playBtn = await screen.findByRole('button', { name: '出す' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(playBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('renders deal end with the next deal button and the deal result', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<CinchPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のディール' })).toBeInTheDocument());
    expect(screen.getByText('ディール結果')).toBeInTheDocument();
  });

  it('shows the bidder detail without a set-back row when the bid is made', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<CinchPage />);
    const detail = await screen.findByTestId('cinch-bidder-detail');
    // Made-bid detail is not styled with the danger color and no set-back row is present.
    expect(detail).not.toHaveClass('text-ds-error');
    expect(screen.queryByTestId('cinch-setback-row')).not.toBeInTheDocument();
  });

  it('emphasizes the bidder detail and set-back row when the bidder is set back', async () => {
    mockExec.mockResolvedValue(setBackState);
    renderWithProviders(<CinchPage />);
    const detail = await screen.findByTestId('cinch-bidder-detail');
    // Set-back bidder detail is emphasized with the danger color.
    expect(detail).toHaveClass('text-ds-error');
    // The bidder's gained row is highlighted as a set-back row.
    expect(screen.getByTestId('cinch-setback-row')).toBeInTheDocument();
  });

  it('renders the game end message', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<CinchPage />);
    await waitFor(() => expect(screen.getByText('ゲーム終了！ あなたの勝ち！')).toBeInTheDocument());
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<CinchPage />);
    await waitFor(() => expect(screen.getByAltText('♥ Q')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  // **押していない人にヒントを見せない。**#4483 以降 `Output()` が毎回
  // ヒントを載せるので、`state.hint` だけを見て描画すると常時表示になる (#4605)。
  it('renders no hint banner when the hint was not requested', async () => {
    mockExec.mockResolvedValue({ ...playPhaseState, hint: { cardIndices: [0], bid: null, reason: 'x' } });
    renderWithProviders(<CinchPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByText(/\(\[0\]\)/)).not.toBeInTheDocument();
  });

  // **押したときは出る。**押していない側だけを見ていると、`isRequestedHint` を
  // 定数 false にしても通ってしまう。真の分岐も踏んでおく。
  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue({
      ...playPhaseState,
      // このページのバナーは `cardIndices` を並べる。`cardIndex` を渡しても
      // 何も出ないので、押した側のテストにならない。
      hint: { cardIndices: [0], reason: 'x' },
      messageCode: 'cinch.hintRequested',
    });
    renderWithProviders(<CinchPage />);
    expect(await screen.findByText(/\(\[0\]\)/)).toBeInTheDocument();
  });

  // **催促は常設のライブ領域の中にある (#6880)。** フェーズ切り替えで現れる
  // テキストなので、領域が無いとスクリーンリーダには何も届かない。領域を
  // 出現と同時に付けても読み上げられないため、常設にして中身だけ差し替える。
  it('announces the prompt from an always-mounted live region', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<CinchPage />);

    const live = await screen.findByTestId('cinch-prompt-live');
    expect(live).toHaveAttribute('role', 'status');
    expect(live).toHaveAttribute('aria-live', 'polite');
    // 催促が**その領域の中**にあること。隣に置いただけの実装は属性の検査を通る。
    expect(live).toContainElement(await screen.findByTestId('cinch-bid-prompt'));
  });

  // 切り札指名の催促もフェーズが変わったときに現れるテキスト。**名前が
  // `-bid-prompt` でないだけ**で領域の外に取り残されていた (#6880 レビュー指摘)。
  it('announces the cinch-trump-prompt from the always-mounted live region', async () => {
    mockExec.mockResolvedValue(nameTrumpState);
    renderWithProviders(<CinchPage />);

    const live = await screen.findByTestId('cinch-prompt-live');
    expect(live).toHaveAttribute('role', 'status');
    expect(live).toHaveAttribute('aria-live', 'polite');
    // 隣に置いただけの実装は属性の検査を通る。**中にあること**を見る。
    expect(live).toContainElement(await screen.findByTestId('cinch-trump-prompt'));
  });

  // **ディール結果が表示される。**
  it('shows the deal result at the end of a round', async () => {
    mockExec.mockResolvedValue(roundEndState);
    const { unmount } = renderWithProviders(<CinchPage />);
    await waitFor(() => expect(screen.getByTestId('cinch-deal-result')).toBeInTheDocument());
    unmount();

    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<CinchPage />);
    await waitFor(() => expect(screen.queryByTestId('cinch-deal-result')).not.toBeInTheDocument());
  });

  // **CPUの入札手番のときに誰がビッドしているか表示される。**
  it('shows the CPU bid notice when it is the CPU turn to bid', async () => {
    mockExec.mockResolvedValue(makeCinchState({ phase: 0, bidPlayerIdx: 1, isHumanTurn: false }));
    const { unmount } = renderWithProviders(<CinchPage />);
    await waitFor(() => expect(screen.getByTestId('cinch-bid-cpu')).toBeInTheDocument());
    unmount();

    mockExec.mockResolvedValue(bidPhaseState);
    renderWithProviders(<CinchPage />);
    await waitFor(() => expect(screen.queryByTestId('cinch-bid-cpu')).not.toBeInTheDocument());
  });

  // **人間の入札手番のときにビッドボタンが表示される。**
  it('shows the bid buttons when it is the human turn to bid', async () => {
    mockExec.mockResolvedValue(bidPhaseState);
    const { unmount } = renderWithProviders(<CinchPage />);
    await waitFor(() => expect(screen.getByTestId('cinch-bid-buttons')).toBeInTheDocument());
    unmount();

    mockExec.mockResolvedValue(makeCinchState({ phase: 0, bidPlayerIdx: 1, isHumanTurn: false }));
    renderWithProviders(<CinchPage />);
    await waitFor(() => expect(screen.queryByTestId('cinch-bid-buttons')).not.toBeInTheDocument());
  });

  // **人間の切札指名手番のときに切札ボタンが表示される。**
  it('shows the trump buttons when it is the human turn to name trump', async () => {
    mockExec.mockResolvedValue(nameTrumpState);
    const { unmount } = renderWithProviders(<CinchPage />);
    await waitFor(() => expect(screen.getByTestId('cinch-trump-buttons')).toBeInTheDocument());
    unmount();

    mockExec.mockResolvedValue(makeCinchState({ phase: 1, bidWinnerIdx: 1, isHumanTurn: false }));
    renderWithProviders(<CinchPage />);
    await waitFor(() => expect(screen.queryByTestId('cinch-trump-buttons')).not.toBeInTheDocument());
  });
});
