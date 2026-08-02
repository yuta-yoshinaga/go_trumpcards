import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { killeApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import type { KillePlayer, KilleResponse } from '../types/card';
import { KillePhase } from '../types/phases';
import { KillePage } from './KillePage';

vi.mock('../api/gameApi', () => ({
  killeApi: { exec: vi.fn() },
  actionLogApi: { kille: vi.fn() },
}));

const mockExec = vi.mocked(killeApi.exec);

/** A Kille card: single suit, so the procedural fields carry the identity. */
const killeCard = (value: number, label: string, color: string) => ({
  design: 'JOKER' as const,
  value,
  glyph: label,
  label,
  color,
  deck: 'kille',
});

function seat(id: number, isHuman: boolean, overrides?: Partial<KillePlayer>): KillePlayer {
  return {
    id,
    isHuman,
    card: isHuman ? killeCard(8, '5', 'black') : null,
    strength: isHuman ? 8 : 0,
    chips: -1,
    reentries: 0,
    reentryCost: 1,
    canReenter: true,
    isOut: false,
    knockedBy: '',
    isSatisfied: false,
    isFinished: false,
    isCurrentTurn: id === 0,
    ...overrides,
  };
}

function makeState(overrides?: Partial<KilleResponse>): KilleResponse {
  return {
    players: [seat(0, true), seat(1, false), seat(2, false), seat(3, false)],
    phase: KillePhase.EXCHANGE,
    roundNumber: 0,
    currentPlayerIdx: 0,
    dealerIdx: 3,
    stockCount: 38,
    pot: 4,
    events: [],
    loserIdxs: [],
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    ...overrides,
  };
}

describe('KillePage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(makeState());
  });

  it('resets on mount', async () => {
    renderWithProviders(<KillePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('exchanges and stands pat', async () => {
    renderWithProviders(<KillePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /交換する/ })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: /交換する/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('exchange'));

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: /交換しない/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('satisfied'));
  });

  // **親は隣ではなく山札と交換する。**ボタンの文言が変わらないと誰と交換するのか判らない。
  it('tells the dealer it swaps with the stock', async () => {
    mockExec.mockResolvedValue(makeState({ dealerIdx: 0 }));
    renderWithProviders(<KillePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /山札と交換/ })).toBeInTheDocument());
    expect(screen.getByTestId('kille-turn-notice')).toHaveTextContent('誰にも仕掛けられません');
  });

  // 交換で渡ってきた道化が最弱になるのはこのゲーム最大の罠なので、常時出す。
  it('states the inverted Harlequin rule permanently', async () => {
    renderWithProviders(<KillePage />);
    await waitFor(() => expect(screen.getByTestId('kille-rules-note')).toBeInTheDocument());
    expect(screen.getByTestId('kille-rules-note')).toHaveTextContent('最強ではなく最弱');
  });

  it('shows the strength ladder', async () => {
    renderWithProviders(<KillePage />);
    await waitFor(() => expect(screen.getByTestId('kille-ladder')).toBeInTheDocument());
    const ladder = screen.getByTestId('kille-ladder');
    for (const name of ['Harlequin', 'Cuckoo', 'Hussar', 'Pig', 'Cavalier', 'Inn', 'Mask']) {
      expect(ladder).toHaveTextContent(name);
    }
  });

  // 何が起きたか判らないと、いきなり落ちた理由が説明できない。
  it('narrates what each exchange did', async () => {
    mockExec.mockResolvedValue(
      makeState({
        events: [
          { kind: 'swap', actor: 0, target: 1 },
          { kind: 'cuckoo', actor: 1, target: 2 },
          { kind: 'hussar', actor: 2, target: 3 },
          { kind: 'stock', actor: 3, target: -1 },
        ],
      }),
    );
    renderWithProviders(<KillePage />);
    await waitFor(() => expect(screen.getByTestId('kille-events')).toBeInTheDocument());
    const events = screen.getByTestId('kille-events');
    expect(events).toHaveTextContent('その場でラウンド終了');
    expect(events).toHaveTextContent('が脱落');
    expect(events).toHaveTextContent('山札と交換');
  });

  // **軽騎兵と豚は手の強さと無関係に落とす。**「最弱だから」と書くと嘘になる。
  it('names why each seat went out', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: KillePhase.SHOWDOWN,
        loserIdxs: [1, 2, 3],
        players: [
          seat(0, true),
          seat(1, false, { isOut: true, knockedBy: 'hussar' }),
          seat(2, false, { isOut: true, knockedBy: 'pig' }),
          seat(3, false, { isOut: true, knockedBy: '' }),
        ],
      }),
    );
    renderWithProviders(<KillePage />);
    await waitFor(() => expect(screen.getByTestId('kille-showdown')).toBeInTheDocument());
    const showdown = screen.getByTestId('kille-showdown');
    expect(showdown).toHaveTextContent('軽騎兵に返り討ち');
    expect(showdown).toHaveTextContent('豚に噛まれた');
    expect(showdown).toHaveTextContent('最弱で脱落');
  });

  it('offers the buy-back at its current price', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: KillePhase.SHOWDOWN,
        loserIdxs: [0],
        players: [
          seat(0, true, { isOut: true, reentries: 1, reentryCost: 6, card: null }),
          seat(1, false),
          seat(2, false),
          seat(3, false),
        ],
      }),
    );
    renderWithProviders(<KillePage />);
    await waitFor(() => expect(screen.getByTestId('kille-reenter-button')).toBeInTheDocument());
    expect(screen.getByTestId('kille-reenter-button')).toHaveTextContent('6');

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('kille-reenter-button'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reenter'));
  });

  // 3 回使い切ったら買い戻しボタン自体を出さない。
  it('withdraws the buy-back once all three are spent', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: KillePhase.SHOWDOWN,
        loserIdxs: [0],
        players: [
          seat(0, true, { isOut: true, reentries: 3, reentryCost: 0, canReenter: false, card: null }),
          seat(1, false),
          seat(2, false),
          seat(3, false),
        ],
      }),
    );
    renderWithProviders(<KillePage />);
    await waitFor(() => expect(screen.getByTestId('kille-reenter-exhausted')).toBeInTheDocument());
    expect(screen.queryByTestId('kille-reenter-button')).not.toBeInTheDocument();
  });

  it('advances to the next round', async () => {
    mockExec.mockResolvedValue(makeState({ phase: KillePhase.SHOWDOWN, loserIdxs: [1] }));
    renderWithProviders(<KillePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /次のラウンドへ/ })).toBeInTheDocument());

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: /次のラウンドへ/ }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  // 交換フェーズでない間は仕掛けられない。
  it('offers no exchange during the showdown', async () => {
    mockExec.mockResolvedValue(makeState({ phase: KillePhase.SHOWDOWN, loserIdxs: [1] }));
    renderWithProviders(<KillePage />);
    await waitFor(() => expect(screen.getByTestId('kille-showdown')).toBeInTheDocument());

    mockExec.mockClear();
    expect(screen.queryByRole('button', { name: /交換する/ })).not.toBeInTheDocument();
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('reports each outcome', async () => {
    for (const [winner, text] of [
      [0, /あなたの勝利です！/],
      [2, /CPU 2 の勝利です。/],
    ] as const) {
      mockExec.mockResolvedValue(
        makeState({ phase: KillePhase.GAME_END, gameEndFlag: true, winnerIdx: winner, loserIdxs: [1, 3] }),
      );
      renderWithProviders(<KillePage />);
      await waitFor(() => expect(screen.getAllByText(text).length).toBeGreaterThan(0));
    }
  });

  // **ヒント経路はページ側からも踏む。**ファクトリ単体テストだけだと
  // `hintFactories` の登録行と、ページのトグル／ツールチップが一度も
  // 実行されない（#4596 / #4600 のレビュー指摘）。
  it('turns the frontend hint on from the checkbox', async () => {
    localStorage.clear();
    mockExec.mockResolvedValue(makeState());
    renderWithProviders(<KillePage />);

    const toggle = await screen.findByRole('checkbox', { name: 'ヒント表示' });
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();

    fireEvent.click(toggle);
    expect(await screen.findByTestId('hint-tooltip')).toBeInTheDocument();
  });
});
