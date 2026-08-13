import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { zwanzigerrufenApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeZwanzigerrufenState } from '../test/stateFactories';
import { ZwanzigerrufenPage } from './ZwanzigerrufenPage';

vi.mock('../api/gameApi', () => ({
  zwanzigerrufenApi: { exec: vi.fn() },
  actionLogApi: { zwanzigerrufen: vi.fn() },
}));

const mockExec = vi.mocked(zwanzigerrufenApi.exec);

/**
 * The shared PlayerHandSection renders each card as a button carrying
 * `aria-pressed` (no test id), so select the hand that way rather than by index.
 */
const handButtons = () => screen.getAllByRole('button').filter((b) => b.hasAttribute('aria-pressed'));

const bidState = makeZwanzigerrufenState();
const talonState = makeZwanzigerrufenState({
  phase: 1,
  declarerIdx: 0,
  contract: 2,
  contractName: 'rufer',
  calledTrump: 20,
  players: bidState.players.map((p, i) =>
    i === 0 ? { ...p, isDeclarer: true, cardCount: 9, cards: [...p.cards, ...p.cards, ...p.cards] } : p,
  ),
});
const playState = makeZwanzigerrufenState({
  phase: 2,
  trickNumber: 1,
  declarerIdx: 0,
  contract: 2,
  contractName: 'rufer',
  calledTrump: 20,
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(bidState);
});

describe('ZwanzigerrufenPage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<ZwanzigerrufenPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows the deal, trick and contract', async () => {
    renderWithProviders(<ZwanzigerrufenPage />);
    expect(await screen.findByTestId('zw-info')).toHaveTextContent('ディール 1/4');
  });

  // **入札できるのは 20番呼びとソロだけ。** トリシャーケンは全員パスの結果なので
  // ボタンにしない。
  it('offers only rufer, solo and pass in the auction', async () => {
    renderWithProviders(<ZwanzigerrufenPage />);
    expect(await screen.findByTestId('zw-bid-rufer')).toBeInTheDocument();
    expect(screen.getByTestId('zw-bid-solo')).toBeInTheDocument();
    expect(screen.getByTestId('zw-pass')).toBeInTheDocument();
    expect(screen.queryByText('トリシャーケン')).not.toBeInTheDocument();

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('zw-bid-rufer'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 'rufer' }));
  });

  it('sends solo and pass', async () => {
    renderWithProviders(<ZwanzigerrufenPage />);
    fireEvent.click(await screen.findByTestId('zw-bid-solo'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 'solo' }));

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('zw-pass'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass'));
  });

  // **呼び札は公開、持ち主は秘密。**
  it('shows the called trump without naming the partner', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<ZwanzigerrufenPage />);
    const called = await screen.findByTestId('zw-called');
    expect(called).toHaveTextContent('20');
    expect(called).not.toHaveTextContent('パートナー');

    for (const seat of [0, 1, 2, 3]) {
      expect(screen.getByTestId(`zw-seat-${seat}`)).not.toHaveTextContent('（パートナー）');
    }
  });

  it('names the partner once revealed', async () => {
    mockExec.mockResolvedValue(
      makeZwanzigerrufenState({
        ...playState,
        partnerRevealed: true,
        partnerIdx: 2,
        players: playState.players.map((p, i) => (i === 2 ? { ...p, isPartner: true } : p)),
      }),
    );
    renderWithProviders(<ZwanzigerrufenPage />);
    expect(await screen.findByTestId('zw-called')).toHaveTextContent('CPU2');
    expect(screen.getByTestId('zw-seat-2')).toHaveTextContent('（パートナー）');
  });

  // **伏せるのはちょうど 6 枚。** 足りないうちはボタンを押せない。
  it('requires exactly six cards before burying', async () => {
    mockExec.mockResolvedValue(talonState);
    renderWithProviders(<ZwanzigerrufenPage />);
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

  it('plays a card immediately during the play phase', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<ZwanzigerrufenPage />);
    await screen.findByTestId('zw-info');
    mockExec.mockClear();
    fireEvent.click(handButtons()[0]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('advances the trick and the deal', async () => {
    mockExec.mockResolvedValue(makeZwanzigerrufenState({ ...playState, phase: 3, lastTrickWinner: 1 }));
    renderWithProviders(<ZwanzigerrufenPage />);
    fireEvent.click(await screen.findByTestId('zw-next-trick'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));

    mockExec.mockResolvedValue(
      makeZwanzigerrufenState({
        ...playState,
        phase: 4,
        breakdown: {
          contract: 2,
          teamPoints: 60,
          threshold: 52,
          won: true,
          solo: false,
          base: 18,
          seats: [18, -18, 18, -18],
          loser: -1,
          name: 'rufer',
        },
      }),
    );
    renderWithProviders(<ZwanzigerrufenPage />);
    expect(await screen.findByTestId('zw-round-result')).toHaveTextContent('60');
    fireEvent.click(screen.getByTestId('zw-next-round'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  // **トリシャーケンだけ結果の文が逆向き。** 最多得点者が負ける。
  it('reports the Trischaken loser instead of a contract result', async () => {
    mockExec.mockResolvedValue(
      makeZwanzigerrufenState({
        ...playState,
        phase: 4,
        declarerIdx: -1,
        contract: 1,
        contractName: 'trischaken',
        calledTrump: -1,
        breakdown: {
          contract: 1,
          teamPoints: 33,
          threshold: 0,
          won: false,
          solo: false,
          base: 3,
          seats: [1, 1, -3, 1],
          loser: 2,
          name: 'trischaken',
        },
      }),
    );
    renderWithProviders(<ZwanzigerrufenPage />);
    const result = await screen.findByTestId('zw-round-result');
    expect(result).toHaveTextContent('CPU2');
    expect(result).toHaveTextContent('33');
    expect(result).not.toHaveTextContent('成功');
    expect(screen.getByTestId('zw-round-seat-2')).toHaveTextContent('-3');
  });

  it('shows the final scores and restarts with the chosen settings', async () => {
    mockExec.mockResolvedValue(makeZwanzigerrufenState({ ...playState, phase: 5, gameEndFlag: true, winnerPlayer: 0 }));
    renderWithProviders(<ZwanzigerrufenPage />);
    expect(await screen.findByTestId('zw-result')).toHaveTextContent('勝者: あなた');

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '新しいゲーム' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { cpuDifficulty: 1, targetDeals: 4 } }),
    );
  });

  it('reports a tied match', async () => {
    mockExec.mockResolvedValue(
      makeZwanzigerrufenState({ ...playState, phase: 5, gameEndFlag: true, winnerPlayer: -1 }),
    );
    renderWithProviders(<ZwanzigerrufenPage />);
    expect(await screen.findByTestId('zw-result')).toHaveTextContent('引き分け');
  });

  it('surfaces an API error raised after the board is up', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<ZwanzigerrufenPage />);
    await screen.findByTestId('zw-info');

    mockExec.mockRejectedValueOnce(new Error('boom'));
    fireEvent.click(handButtons()[0]);
    expect(await screen.findByRole('alert')).toBeInTheDocument();
  });
});
