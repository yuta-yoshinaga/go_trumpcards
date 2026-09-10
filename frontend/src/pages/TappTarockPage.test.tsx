import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { tapptarockApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeTappTarockState } from '../test/stateFactories';
import { TappTarockPage } from './TappTarockPage';

vi.mock('../types/phases', async () => {
  const actual = await vi.importActual<typeof import('../types/phases')>('../types/phases');
  return {
    ...actual,
    TappTarockBid: { PASS: 0, TRISCHAKEN: 1, DREIER: 2, SOLO: 3 },
    TappTarockPhase: { BID: 0, TALON: 1, PLAY: 2, TRICK_END: 3, ROUND_END: 4, GAME_END: 5 },
  };
});

vi.mock('../styles/gameTheme', async () => {
  const actual = await vi.importActual<typeof import('../styles/gameTheme')>('../styles/gameTheme');
  return {
    ...actual,
    gameTheme: { ...actual.gameTheme, tapptarock: { bg: 'bg-test', footer: 'bg-test-footer' } },
  };
});

vi.mock('../api/gameApi', () => ({
  tapptarockApi: { exec: vi.fn() },
  actionLogApi: { tapptarock: vi.fn() },
}));

const mockExec = vi.mocked(tapptarockApi.exec);

/**
 * The shared PlayerHandSection renders each card as a button carrying
 * `aria-pressed` (no test id), so select the hand that way rather than by index.
 */
const handButtons = () => screen.getAllByRole('button').filter((b) => b.hasAttribute('aria-pressed'));

const bidState = makeTappTarockState();
const talonState = makeTappTarockState({
  phase: 1,
  declarerIdx: 0,
  contract: 2,
  contractName: 'dreier',
  discardableIndices: [0, 1, 2, 3, 4, 5, 6, 7, 8],
  players: bidState.players.map((p, i) =>
    i === 0 ? { ...p, isDeclarer: true, cardCount: 9, cards: [...p.cards, ...p.cards, ...p.cards] } : p,
  ),
});
const playState = makeTappTarockState({
  phase: 2,
  trickNumber: 1,
  declarerIdx: 0,
  contract: 2,
  contractName: 'dreier',
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(bidState);
});

describe('TappTarockPage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<TappTarockPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows the deal, trick and contract', async () => {
    renderWithProviders(<TappTarockPage />);
    expect(await screen.findByTestId('zw-info')).toHaveTextContent('ディール 1/4');
  });

  // **入札できるのはドライヤーとソロだけ。** トリシャーケンは全員パスの結果なので
  // ボタンにしない。
  it('offers only dreier, solo and pass in the auction', async () => {
    renderWithProviders(<TappTarockPage />);
    expect(await screen.findByTestId('tapp-bid-dreier')).toBeInTheDocument();
    expect(screen.getByTestId('zw-bid-solo')).toBeInTheDocument();
    expect(screen.getByTestId('zw-pass')).toBeInTheDocument();
    expect(screen.queryByText('トリシャーケン')).not.toBeInTheDocument();

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('tapp-bid-dreier'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 'dreier' }));
  });

  // **押せるかは規則が決める** (#5786)。上回れない入札を押させると、サーバの
  // エラーで返ってくるだけになる。
  it('disables bids that cannot beat the current highest', async () => {
    mockExec.mockResolvedValue(makeTappTarockState({ highestBid: 2 }));
    const { unmount } = renderWithProviders(<TappTarockPage />);
    // 20番呼び (2) が出ているので、同じ 20番呼びはもう打てない。ソロ (3) は打てる。
    expect(await screen.findByTestId('tapp-bid-dreier')).toBeDisabled();
    expect(screen.getByTestId('zw-bid-solo')).toBeEnabled();
    expect(screen.getByTestId('zw-pass')).toBeEnabled();
    expect(screen.getByTestId('zw-highest-bid')).toHaveTextContent('ドライヤー');
    unmount();

    // ソロまで出ていれば、上回れる入札はもう無い。
    mockExec.mockResolvedValue(makeTappTarockState({ highestBid: 3 }));
    renderWithProviders(<TappTarockPage />);
    expect(await screen.findByTestId('zw-bid-solo')).toBeDisabled();
    expect(screen.getByTestId('tapp-bid-dreier')).toBeDisabled();
    expect(screen.getByTestId('zw-pass')).toBeEnabled();
    expect(screen.getByTestId('zw-highest-bid')).toHaveTextContent('ソロ');
  });

  // **まだ誰も入札していなければ両方押せる。**
  it('keeps both bids live before anyone has bid', async () => {
    renderWithProviders(<TappTarockPage />);
    expect(await screen.findByTestId('tapp-bid-dreier')).toBeEnabled();
    expect(screen.getByTestId('zw-bid-solo')).toBeEnabled();
    expect(screen.getByTestId('zw-highest-bid')).toHaveTextContent('-');
  });

  it('sends solo and pass', async () => {
    renderWithProviders(<TappTarockPage />);
    fireEvent.click(await screen.findByTestId('zw-bid-solo'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 'solo' }));

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('zw-pass'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass'));
  });

  // **タロンに伏せるカードはちょうど 6 枚。** 足りないうちはボタンを押せない。
  it('requires exactly six cards before burying', async () => {
    mockExec.mockResolvedValue(talonState);
    renderWithProviders(<TappTarockPage />);
    const discard = await screen.findByTestId('zw-discard');
    expect(discard).toBeDisabled();

    mockExec.mockClear();
    const cards = handButtons();
    expect(cards.length).toBeGreaterThanOrEqual(6);
    for (let i = 0; i < 6; i++) {
      fireEvent.click(cards[i]);
    }
    expect(await screen.findByTestId('zw-discard')).toBeEnabled();
    fireEvent.click(screen.getByTestId('zw-discard'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', { cardIndices: [0, 1, 2, 3, 4, 5] }));
  });

  // **キングやトゥルルなど伏せられない札はタロンフェーズで無効化される。**
  it('disables cards outside discardableIndices during the talon phase', async () => {
    mockExec.mockResolvedValue(
      makeTappTarockState({
        ...talonState,
        discardableIndices: [1, 2, 3, 4, 5, 6],
      }),
    );
    renderWithProviders(<TappTarockPage />);
    await screen.findByTestId('zw-discard');
    const cards = handButtons();
    expect(cards[0]).toHaveAttribute('aria-disabled', 'true');
    expect(cards[1]).not.toHaveAttribute('aria-disabled', 'true');
  });

  it('plays a card immediately during the play phase', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<TappTarockPage />);
    await screen.findByTestId('zw-info');
    mockExec.mockClear();
    fireEvent.click(handButtons()[0]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('disables cards outside playableIndices during the play phase', async () => {
    mockExec.mockResolvedValue(
      makeTappTarockState({
        ...playState,
        playableIndices: [1, 2],
      }),
    );
    renderWithProviders(<TappTarockPage />);
    await screen.findByTestId('zw-info');
    const cards = handButtons();
    expect(cards[0]).toHaveAttribute('aria-disabled', 'true');
    expect(cards[1]).not.toHaveAttribute('aria-disabled', 'true');
  });

  it('advances the trick and the deal', async () => {
    mockExec.mockResolvedValue(makeTappTarockState({ ...playState, phase: 3, lastTrickWinner: 1 }));
    renderWithProviders(<TappTarockPage />);
    fireEvent.click(await screen.findByTestId('zw-next-trick'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));

    mockExec.mockResolvedValue(
      makeTappTarockState({
        ...playState,
        phase: 4,
        breakdown: {
          contract: 2,
          teamPoints: 60,
          threshold: 52,
          won: true,
          solo: false,
          base: 18,
          seats: [18, -18, 18],
          loser: -1,
          name: 'dreier',
        },
      }),
    );
    renderWithProviders(<TappTarockPage />);
    expect(await screen.findByTestId('zw-round-result')).toHaveTextContent('60');
    fireEvent.click(screen.getByTestId('zw-next-round'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  // **トリシャーケンだけ結果の文が逆向き。** 最多得点者が負ける。
  it('reports the Trischaken loser instead of a contract result', async () => {
    mockExec.mockResolvedValue(
      makeTappTarockState({
        ...playState,
        phase: 4,
        declarerIdx: -1,
        contract: 1,
        contractName: 'trischaken',
        breakdown: {
          contract: 1,
          teamPoints: 33,
          threshold: 0,
          won: false,
          solo: false,
          base: 3,
          seats: [1, 1, -3],
          loser: 2,
          name: 'trischaken',
        },
      }),
    );
    renderWithProviders(<TappTarockPage />);
    const result = await screen.findByTestId('zw-round-result');
    expect(screen.getByTestId('zw-info')).toHaveTextContent('トリシャーケン');
    expect(result).toHaveTextContent('CPU2');
    expect(result).toHaveTextContent('33');
    expect(result).not.toHaveTextContent('成功');
    expect(screen.getByTestId('zw-round-seat-2')).toHaveTextContent('-3');
  });

  it('shows the final scores and restarts with the chosen settings', async () => {
    mockExec.mockResolvedValue(makeTappTarockState({ ...playState, phase: 5, gameEndFlag: true, winnerPlayer: 0 }));
    renderWithProviders(<TappTarockPage />);
    expect(await screen.findByTestId('zw-result')).toHaveTextContent('勝者: あなた');

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '新しいゲーム' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { cpuDifficulty: 1, targetDeals: 4 } }),
    );
  });

  it('reports a tied match', async () => {
    mockExec.mockResolvedValue(makeTappTarockState({ ...playState, phase: 5, gameEndFlag: true, winnerPlayer: -1 }));
    renderWithProviders(<TappTarockPage />);
    expect(await screen.findByTestId('zw-result')).toHaveTextContent('引き分け');
  });

  it('surfaces an API error raised after the board is up', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<TappTarockPage />);
    await screen.findByTestId('zw-info');

    mockExec.mockRejectedValueOnce(new Error('boom'));
    fireEvent.click(handButtons()[0]);
    expect(await screen.findByRole('alert')).toBeInTheDocument();
  });
});
