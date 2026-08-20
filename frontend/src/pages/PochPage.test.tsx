import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { pochApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { CardDesign, PochPlayer, PochResponse } from '../types/card';
import { PochPhase } from '../types/phases';
import { PochPage } from './PochPage';

vi.mock('../api/gameApi', () => ({
  pochApi: { exec: vi.fn() },
  actionLogApi: { poch: vi.fn() },
}));

const mockExec = vi.mocked(pochApi.exec);

const card = (design: CardDesign, value: number) => ({ design, value });

const POOLS = ['ace', 'king', 'queen', 'jack', 'ten', 'marriage', 'sequence', 'pocher', 'centre'];

function seat(id: number, isHuman: boolean, overrides?: Partial<PochPlayer>): PochPlayer {
  return {
    id,
    isHuman,
    cardCount: 3,
    cards: isHuman ? [card('SPADE', 7), card('HEART', 9), card('CLOVER', 13)] : [],
    chips: -5,
    bet: 0,
    folded: false,
    hidden: !isHuman,
    ...overrides,
  };
}

function makeState(overrides?: Partial<PochResponse>): PochResponse {
  return {
    players: [seat(0, true), seat(1, false), seat(2, false), seat(3, false)],
    phase: PochPhase.POCHEN,
    validPlays: [],
    currentPlayerIdx: 0,
    pools: POOLS.map((name, i) => ({ name, chips: i === 5 ? 12 : 4 })),
    paySuit: 1,
    turnUp: card('SPADE', 9),
    stakingAwards: [{ pool: 'marriage', player: 1, chips: 12 }],
    betTarget: 1,
    yourBestComboSize: 0,
    yourBestComboRank: 0,
    pochenWinner: -1,
    pochenPot: 0,
    playedPile: [],
    stopsSuit: -1,
    stopsRank: 0,
    dealNo: 0,
    targetDeals: 5,
    dealWinner: -1,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    ...overrides,
  };
}

/** Click the hand card at the given index. */
function pickHand(i: number) {
  const hand = screen.getAllByRole('button').filter((b) => b.dataset.hintAction === 'play');
  fireEvent.click(hand[i]);
}

describe('PochPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(makeState());
  });

  it('resets on mount', async () => {
    renderWithProviders(<PochPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows both rules permanently', async () => {
    renderWithProviders(<PochPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.getByText(/めくり札と同じスート/)).toBeInTheDocument();
    expect(screen.getByText(/宣言ではなく同ランクの組の比べ合い/)).toBeInTheDocument();
  });

  // **9 区画すべてが出ていないと、持ち越しがどこに乗っているか読めない。**
  it('shows all nine pools with their chips', async () => {
    renderWithProviders(<PochPage />);
    await waitFor(() => expect(screen.getAllByTestId('poch-pool')).toHaveLength(9));
    const marriage = screen.getAllByTestId('poch-pool').find((el) => el.textContent?.includes('マリッジ'));
    expect(marriage).toHaveTextContent('12');
  });

  // 第 1 段階は自動で解決するので、結果を出さないと何が起きたのか読めない。
  it('reports what stage one paid out', async () => {
    renderWithProviders(<PochPage />);
    await waitFor(() => expect(screen.getByTestId('poch-staking')).toHaveTextContent('マリッジ'));
    expect(screen.getByTestId('poch-staking')).toHaveTextContent('12');
  });

  it('bets and folds during the pochen', async () => {
    renderWithProviders(<PochPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '賭ける' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '降りる' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('fold'));
  });

  it('offers no play control during the pochen', async () => {
    renderWithProviders(<PochPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('plays exactly one card during the stops', async () => {
    mockExec.mockResolvedValue(makeState({ phase: PochPhase.STOPS }));
    renderWithProviders(<PochPage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '出す' })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    pickHand(1);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 1));
  });

  // 止まっているかどうかで出せる札がまるで違うので、案内も変わる。
  it('tells a stopped run apart from one in progress', async () => {
    mockExec.mockResolvedValue(makeState({ phase: PochPhase.STOPS, stopsSuit: -1 }));
    renderWithProviders(<PochPage />);
    await waitFor(() => expect(screen.getByText(/好きな札から始められます/)).toBeInTheDocument());
  });

  it('asks for the next higher card while a run is live', async () => {
    mockExec.mockResolvedValue(
      makeState({ phase: PochPhase.STOPS, stopsSuit: 1, stopsRank: 7, playedPile: [card('SPADE', 7)] }),
    );
    renderWithProviders(<PochPage />);
    await waitFor(() => expect(screen.getByText(/同じスートの次に高い札/)).toBeInTheDocument());
  });

  it('advances to the next deal', async () => {
    mockExec.mockResolvedValue(makeState({ phase: PochPhase.DEAL_END, dealWinner: 2 }));
    renderWithProviders(<PochPage />);
    await waitFor(() => expect(screen.getByTestId('poch-deal-result')).toHaveTextContent('席2'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のディールへ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('reports each outcome', async () => {
    for (const [winner, code, text] of [
      [0, 'poch.win', '最も多く稼ぎました'],
      [2, 'poch.lose', '及びませんでした'],
    ] as const) {
      mockExec.mockResolvedValue(
        makeState({ phase: PochPhase.GAME_END, gameEndFlag: true, winnerIdx: winner, messageCode: code }),
      );
      renderWithProviders(<PochPage />);
      await waitFor(() => expect(screen.getAllByText(text).length).toBeGreaterThan(0));
    }
  });

  // **並びに従う義務がある。**出せない札を押せてしまうと、サーバに弾かれて
  // 初めて分かる (#4933)。
  describe('playable-card restriction', () => {
    it('disables and dims the cards that cannot legally be played', async () => {
      mockExec.mockResolvedValue(makeState({ phase: PochPhase.STOPS, validPlays: [0] }));
      renderWithProviders(<PochPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalled());

      const buttons = screen.getAllByRole('button').filter((b) => b.hasAttribute('data-hint-action'));
      expect(buttons.length).toBeGreaterThan(1);
      expect(buttons[0]).not.toBeDisabled();
      expect(buttons[0]).not.toHaveAttribute('data-unplayable');
      expect(buttons[1]).toBeDisabled();
      expect(buttons[1]).toHaveAttribute('data-unplayable', 'true');
    });

    // **空リストは「一枚も出せない」ではなく「情報が無い」。**空で全部塞ぐと
    // 盤面が操作不能になる。
    it('does not restrict anything when the server sent no list', async () => {
      mockExec.mockResolvedValue(makeState({ phase: PochPhase.STOPS, validPlays: [] }));
      renderWithProviders(<PochPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalled());

      for (const b of screen.getAllByRole('button').filter((x) => x.hasAttribute('data-hint-action'))) {
        expect(b).not.toBeDisabled();
      }
    });
  });

  // #5722: pochen は組の比べ合いなので、賭ける前に自分の組が分かる必要がある。
  it('shows your own strongest set while betting', async () => {
    mockExec.mockResolvedValue(makeState({ yourBestComboSize: 2, yourBestComboRank: 9 }));
    renderWithProviders(<PochPage />);

    const combo = await screen.findByTestId('poch-your-combo');

    expect(combo).toHaveTextContent('9');
    expect(combo).toHaveTextContent('2枚');
  });

  it('says so when you hold no set', async () => {
    mockExec.mockResolvedValue(makeState({ yourBestComboSize: 0, yourBestComboRank: 0 }));
    renderWithProviders(<PochPage />);

    const combo = await screen.findByTestId('poch-your-combo');

    expect(combo).toHaveTextContent('なし');
  });
});
