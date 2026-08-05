import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { nainjauneApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { CardDesign, NainJaunePlayer, NainJauneResponse } from '../types/card';
import { NainJaunePhase } from '../types/phases';
import { NainJaunePage } from './NainJaunePage';

vi.mock('../api/gameApi', () => ({
  nainjauneApi: { exec: vi.fn() },
  actionLogApi: { nainjaune: vi.fn() },
}));

const mockExec = vi.mocked(nainjauneApi.exec);

const card = (design: CardDesign, value: number) => ({ design, value });

const BOXES = [
  { name: 'ten', chips: 4, card: card('DIAMOND', 10) },
  { name: 'jack', chips: 8, card: card('CLOVER', 11) },
  { name: 'queen', chips: 12, card: card('SPADE', 12) },
  { name: 'king', chips: 16, card: card('HEART', 13) },
  { name: 'dwarf', chips: 20, card: card('DIAMOND', 7) },
];

function seat(id: number, isHuman: boolean, overrides?: Partial<NainJaunePlayer>): NainJaunePlayer {
  return {
    id,
    isHuman,
    cardCount: 3,
    cards: isHuman ? [card('SPADE', 3), card('HEART', 9), card('CLOVER', 13)] : [],
    chips: -15,
    points: 22,
    hidden: !isHuman,
    ...overrides,
  };
}

function makeState(overrides?: Partial<NainJauneResponse>): NainJauneResponse {
  return {
    players: [seat(0, true), seat(1, false), seat(2, false), seat(3, false)],
    phase: NainJaunePhase.PLAY,
    validPlays: [],
    currentPlayerIdx: 0,
    boxes: BOXES,
    talonCount: 4,
    awards: [],
    playedPile: [],
    runRank: 0,
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

describe('NainJaunePage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(makeState());
  });

  it('resets on mount', async () => {
    renderWithProviders(<NainJaunePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows both rules permanently', async () => {
    renderWithProviders(<NainJaunePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    expect(screen.getByText(/スート無関係/)).toBeInTheDocument();
    expect(screen.getByText(/枚数ではなく【点数】/)).toBeInTheDocument();
  });

  // **区画はスートまで一致した1枚でしか取れない。**札を出さないと判断できない。
  it('shows all five boxes with the exact card that claims each', async () => {
    renderWithProviders(<NainJaunePage />);
    await waitFor(() => expect(screen.getAllByTestId('nainjaune-box')).toHaveLength(5));
    const dwarf = screen.getAllByTestId('nainjaune-box').find((el) => el.textContent?.includes('黄色い小人'));
    expect(dwarf).toHaveTextContent('20');
  });

  it('plays exactly one card', async () => {
    renderWithProviders(<NainJaunePage />);
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
    renderWithProviders(<NainJaunePage />);
    await waitFor(() => expect(screen.getByText(/好きな札から始められます/)).toBeInTheDocument());
  });

  // **スートを問わない**のが Pope Joan との決定的な違い。案内にも出す。
  it('asks for the next rank of any suit while a run is live', async () => {
    mockExec.mockResolvedValue(makeState({ runRank: 5, playedPile: [card('SPADE', 5)] }));
    renderWithProviders(<NainJaunePage />);
    await waitFor(() => expect(screen.getByText(/6 の札を選んでください（スートは問いません）/)).toBeInTheDocument());
  });

  // 支払いは点数なので、相手の点も見えていないと判断できない。
  it('shows what each hand is worth, not just its size', async () => {
    renderWithProviders(<NainJaunePage />);
    await waitFor(() => expect(screen.getAllByText(/22点/).length).toBeGreaterThan(0));
  });

  it('reports an award', async () => {
    mockExec.mockResolvedValue(makeState({ awards: [{ box: 'dwarf', player: 1, chips: 20 }] }));
    renderWithProviders(<NainJaunePage />);
    await waitFor(() => expect(screen.getByTestId('nainjaune-awards')).toHaveTextContent('20'));
  });

  it('advances to the next deal', async () => {
    mockExec.mockResolvedValue(makeState({ phase: NainJaunePhase.DEAL_END, dealWinner: 2 }));
    renderWithProviders(<NainJaunePage />);
    await waitFor(() => expect(screen.getByTestId('nainjaune-deal-result')).toHaveTextContent('席2'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のディールへ' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('reports each outcome', async () => {
    for (const [winner, code, text] of [
      [0, 'nainjaune.win', '最も多く稼ぎました'],
      [2, 'nainjaune.lose', '及びませんでした'],
    ] as const) {
      mockExec.mockResolvedValue(
        makeState({ phase: NainJaunePhase.GAME_END, gameEndFlag: true, winnerIdx: winner, messageCode: code }),
      );
      renderWithProviders(<NainJaunePage />);
      await waitFor(() => expect(screen.getAllByText(text).length).toBeGreaterThan(0));
    }
  });

  // **並びに従う義務がある。**出せない札を押せてしまうと、サーバに弾かれて
  // 初めて分かる (#4935)。
  describe('playable-card restriction', () => {
    it('disables and dims the cards that cannot legally be played', async () => {
      mockExec.mockResolvedValue(makeState({ validPlays: [0] }));
      renderWithProviders(<NainJaunePage />);
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
      mockExec.mockResolvedValue(makeState({ validPlays: [] }));
      renderWithProviders(<NainJaunePage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalled());

      for (const b of screen.getAllByRole('button').filter((x) => x.hasAttribute('data-hint-action'))) {
        expect(b).not.toBeDisabled();
      }
    });
  });
});
