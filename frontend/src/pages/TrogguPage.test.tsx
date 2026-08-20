import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { trogguApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeTrogguState } from '../test/stateFactories';
import { TrogguPage } from './TrogguPage';

vi.mock('../api/gameApi', () => ({
  trogguApi: { exec: vi.fn() },
  actionLogApi: { troggu: vi.fn() },
}));

const mockExec = vi.mocked(trogguApi.exec);

/** The hand renders as buttons carrying `aria-pressed` (no test id). */
const handButtons = () => screen.getAllByRole('button').filter((b) => b.hasAttribute('aria-pressed'));

const bidState = makeTrogguState();
const playState = makeTrogguState({
  phase: 1,
  trickNumber: 1,
  declarerIdx: 0,
  contract: 2,
  contractName: 'solo',
  players: bidState.players.map((p, i) => (i === 0 ? { ...p, isDeclarer: true } : p)),
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(bidState);
});

describe('TrogguPage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<TrogguPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows the deal, trick and contract', async () => {
    renderWithProviders(<TrogguPage />);
    expect(await screen.findByTestId('tg-info')).toHaveTextContent('ディール 1/4');
  });

  // **4 契約すべてが打てる。** どれか一つ欠けると、その契約だけが遊べなくなる。
  it('offers all four contracts and pass', async () => {
    renderWithProviders(<TrogguPage />);
    for (const c of ['trois', 'solo', 'piccolo', 'misere']) {
      expect(await screen.findByTestId(`tg-bid-${c}`)).toBeInTheDocument();
    }
    expect(screen.getByTestId('tg-pass')).toBeInTheDocument();

    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('tg-bid-misere'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 'misere' }));
  });

  it('sends pass', async () => {
    renderWithProviders(<TrogguPage />);
    fireEvent.click(await screen.findByTestId('tg-pass'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass'));
  });

  // **キーボードから届かない操作を残さない** (#5787)。
  it('パスと「次へ」をキーからも打てる', async () => {
    const { unmount } = renderWithProviders(<TrogguPage />);
    await screen.findByTestId('tg-pass');

    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: 'p' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass'));
    unmount();

    // トリック終了では next、ディール終了では nextround。
    mockExec.mockResolvedValue(makeTrogguState({ ...playState, phase: 2, lastTrickWinner: 1 }));
    const trick = renderWithProviders(<TrogguPage />);
    await screen.findByTestId('tg-next-trick');
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: 'n' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
    trick.unmount();

    mockExec.mockResolvedValue(makeTrogguState({ ...playState, phase: 3 }));
    renderWithProviders(<TrogguPage />);
    await screen.findByTestId('tg-next-round');
    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: 'n' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  // **押せない場面ではキーも効かない。** ボタンの表示条件と同じ値で門番する。
  it('プレイ中はパスのキーが効かない', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<TrogguPage />);
    await screen.findByTestId('tg-info');

    mockExec.mockClear();
    fireEvent.keyDown(document.body, { key: 'p' });
    fireEvent.keyDown(document.body, { key: 'n' });
    await waitFor(() => expect(screen.getByTestId('tg-info')).toBeInTheDocument());
    expect(mockExec).not.toHaveBeenCalled();
  });

  // キー一覧がフッターに出る（受け入れ条件4）。
  it('キーの一覧をフッターに出す', async () => {
    renderWithProviders(<TrogguPage />);
    const panel = await screen.findByTestId('tg-kbd-shortcuts');
    // 既定は閉じたまま（畳んでいる間は行そのものが mount されない）。
    expect(panel).not.toHaveAttribute('open');
    fireEvent.click(screen.getByText('キーボードショートカット'));
    expect(screen.getByText('パスする')).toBeInTheDocument();
  });

  it('plays a card immediately during the play phase', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<TrogguPage />);
    await screen.findByTestId('tg-info');
    mockExec.mockClear();
    fireEvent.click(handButtons()[0]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('advances the trick', async () => {
    mockExec.mockResolvedValue(makeTrogguState({ ...playState, phase: 2, lastTrickWinner: 1 }));
    renderWithProviders(<TrogguPage />);
    fireEvent.click(await screen.findByTestId('tg-next-trick'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  // **ソロは点数で語る。** 単位を取り違えると読めない結果になる。
  it('reports a Solo deal in card points', async () => {
    mockExec.mockResolvedValue(
      makeTrogguState({
        ...playState,
        phase: 3,
        breakdown: {
          contract: 2,
          contractName: 'solo',
          declarerPoints: 60,
          declarerTricks: 9,
          target: 46,
          targetIsTricks: false,
          won: true,
          base: 20,
          seats: [60, -20, -20, -20],
        },
      }),
    );
    renderWithProviders(<TrogguPage />);
    const result = await screen.findByTestId('tg-round-result');
    expect(result).toHaveTextContent('60点');
    expect(result).not.toHaveTextContent('トリック');
    expect(screen.getByTestId('tg-round-seat-0')).toHaveTextContent('60');
  });

  // **他の契約はトリック数で語る。**
  it('reports a Misere deal in tricks', async () => {
    mockExec.mockResolvedValue(
      makeTrogguState({
        ...playState,
        phase: 3,
        contractName: 'misere',
        breakdown: {
          contract: 4,
          contractName: 'misere',
          declarerPoints: 12,
          declarerTricks: 1,
          target: 0,
          targetIsTricks: true,
          won: false,
          base: 40,
          seats: [-120, 40, 40, 40],
        },
      }),
    );
    renderWithProviders(<TrogguPage />);
    const result = await screen.findByTestId('tg-round-result');
    expect(result).toHaveTextContent('トリック');
    expect(result).toHaveTextContent('失敗');
    expect(result).not.toHaveTextContent('12点');
  });

  // 流局は精算そのものが無い。
  it('reports a thrown-in deal', async () => {
    mockExec.mockResolvedValue(makeTrogguState({ phase: 3, breakdown: null }));
    renderWithProviders(<TrogguPage />);
    expect(await screen.findByTestId('tg-round-result')).toHaveTextContent('流局');
  });

  it('shows the final scores and restarts with the chosen settings', async () => {
    mockExec.mockResolvedValue(makeTrogguState({ ...playState, phase: 4, gameEndFlag: true, winnerPlayer: 0 }));
    renderWithProviders(<TrogguPage />);
    expect(await screen.findByTestId('tg-result')).toHaveTextContent('勝者: あなた');

    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '新しいゲーム' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { cpuDifficulty: 1, targetDeals: 4 } }),
    );
  });

  it('reports a tied match', async () => {
    mockExec.mockResolvedValue(makeTrogguState({ ...playState, phase: 4, gameEndFlag: true, winnerPlayer: -1 }));
    renderWithProviders(<TrogguPage />);
    expect(await screen.findByTestId('tg-result')).toHaveTextContent('引き分け');
  });

  it('surfaces an API error raised after the board is up', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<TrogguPage />);
    await screen.findByTestId('tg-info');

    mockExec.mockRejectedValueOnce(new Error('boom'));
    fireEvent.click(handButtons()[0]);
    expect(await screen.findByRole('alert')).toBeInTheDocument();
  });
});
