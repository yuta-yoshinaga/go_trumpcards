import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { baccaratbanqueApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeBaccaratBanqueState } from '../test/stateFactories';
import { BaccaratBanquePage } from './BaccaratBanquePage';

vi.mock('../api/gameApi', () => ({
  baccaratbanqueApi: { exec: vi.fn() },
  actionLogApi: { baccaratbanque: vi.fn() },
}));

const mockExec = vi.mocked(baccaratbanqueApi.exec);

const bankerState = makeBaccaratBanqueState();
const resultState = makeBaccaratBanqueState({
  phase: 'result',
  isHumanTurn: false,
  lastResult: {
    bankerTotal: 6,
    sides: [
      { seatIdx: 1, outcome: 'bankerWin', bet: 50, delta: 50 },
      { seatIdx: 2, outcome: 'punterWin', bet: 50, delta: -50 },
    ],
    bankerDelta: 0,
    bankerNatural: false,
  },
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(bankerState);
});

describe('BaccaratBanquePage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<BaccaratBanquePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows all three seats face up with their totals', async () => {
    renderWithProviders(<BaccaratBanquePage />);
    expect(await screen.findByTestId('banque-hand-banker')).toBeInTheDocument();
    // **バカラは全部表向き。** 3 席とも札が描かれていること。
    expect(screen.getByTestId('banque-hand-banker').children).toHaveLength(2);
    expect(screen.getByTestId('banque-hand-right').children).toHaveLength(3);
    expect(screen.getByTestId('banque-hand-left').children).toHaveLength(2);
    expect(screen.getByTestId('banque-total-banker')).toHaveTextContent('6');
    expect(screen.getByTestId('banque-total-left')).toHaveTextContent('7');
  });

  // **負けても席が動かないのがこの形式の要。** 残高からは読めないので盤に書く。
  it('says how long the bank has been held, and that a loss does not end it', async () => {
    renderWithProviders(<BaccaratBanquePage />);
    const line = await screen.findByTestId('banque-coup-line');
    expect(line).toHaveTextContent('2');
    expect(line).toHaveTextContent('負けても続きます');
  });

  it('shows the shoe so the player can see the bank running out', async () => {
    mockExec.mockResolvedValue(makeBaccaratBanqueState({ shoeRemaining: 12 }));
    renderWithProviders(<BaccaratBanquePage />);
    expect(await screen.findByTestId('banque-shoe-line')).toHaveTextContent('12');
  });

  it('marks a natural, and only on a two-card 8 or 9', async () => {
    mockExec.mockResolvedValue(
      makeBaccaratBanqueState({
        // 左は 2 枚で 8 = ナチュラル。右は **3 枚で 8** ── 同じ合計でも印は付かない。
        players: bankerState.players.map((p) => (p.role === 'banker' ? p : { ...p, total: 8 })),
      }),
    );
    renderWithProviders(<BaccaratBanquePage />);
    expect(await screen.findByTestId('banque-natural-left')).toBeInTheDocument();
    // 負のコントロール: 3 枚で 8 に届いた右には印を付けない。
    expect(screen.queryByTestId('banque-natural-right')).not.toBeInTheDocument();
  });

  // **引くと止まるは別のボタンで、別のコマンドとして届く。**
  it('offers both decisions and sends each as its own command', async () => {
    renderWithProviders(<BaccaratBanquePage />);
    fireEvent.click(await screen.findByRole('button', { name: '引く' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
    expect(mockExec).not.toHaveBeenCalledWith('stand');

    mockExec.mockClear();
    mockExec.mockResolvedValue(bankerState);
    renderWithProviders(<BaccaratBanquePage />);
    const stands = await screen.findAllByRole('button', { name: '引かない' });
    fireEvent.click(stands[stands.length - 1]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('stand'));
    expect(mockExec).not.toHaveBeenCalledWith('draw');
  });

  // **固定表が無いことを毎回言う。** バカラで表を覚えている人ほど手が止まる。
  it('says the choice is free on any total', async () => {
    renderWithProviders(<BaccaratBanquePage />);
    expect(await screen.findByTestId('banque-free-notice')).toHaveTextContent('固定表はありません');
  });

  it('hides the draw buttons once the coup is settled', async () => {
    mockExec.mockResolvedValue(resultState);
    renderWithProviders(<BaccaratBanquePage />);
    await screen.findByTestId('banque-result');
    expect(screen.queryByRole('button', { name: '引く' })).not.toBeInTheDocument();
    expect(screen.queryByTestId('banque-free-notice')).not.toBeInTheDocument();
  });

  // **左右は 1 行ずつ。** 差額だけだと、片方に払いもう片方から取ったクーが読めない。
  it('reports each tableau separately and then the bank net', async () => {
    mockExec.mockResolvedValue(resultState);
    renderWithProviders(<BaccaratBanquePage />);
    expect(await screen.findByTestId('banque-side-1')).toHaveTextContent('バンカーの勝ち');
    expect(screen.getByTestId('banque-side-2')).toHaveTextContent('子の勝ち');
    expect(screen.getByTestId('banque-net')).toHaveTextContent('0');
  });

  it('offers the next coup and retiring once settled, and sends each', async () => {
    mockExec.mockResolvedValue(resultState);
    renderWithProviders(<BaccaratBanquePage />);
    fireEvent.click(await screen.findByRole('button', { name: '次のクー' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextcoup'));

    mockExec.mockClear();
    mockExec.mockResolvedValue(resultState);
    renderWithProviders(<BaccaratBanquePage />);
    const retires = await screen.findAllByRole('button', { name: 'バンクを降りる' });
    fireEvent.click(retires[retires.length - 1]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('retire'));
  });

  it('stops offering the next coup once the bank has ended', async () => {
    mockExec.mockResolvedValue(
      makeBaccaratBanqueState({ phase: 'gameEnd', gameEndFlag: true, retired: true, isHumanTurn: false }),
    );
    renderWithProviders(<BaccaratBanquePage />);
    await screen.findByTestId('banque-coup-line');
    expect(screen.queryByRole('button', { name: '次のクー' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'バンクを降りる' })).not.toBeInTheDocument();
  });
});
