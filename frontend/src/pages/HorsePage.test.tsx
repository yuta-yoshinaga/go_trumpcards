import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { horseApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeHorseState } from '../test/stateFactories';
import { HorsePage } from './HorsePage';

vi.mock('../api/gameApi', () => ({
  horseApi: { exec: vi.fn() },
  actionLogApi: { horse: vi.fn() },
}));

const mockExec = vi.mocked(horseApi.exec);

const handState = makeHorseState();

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(handState);
});

describe('HorsePage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<HorsePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // **いま何を打っているのかが画面に要る。** ミックスゲームで種目が分からないと、
  // 同じ操作でも規則が違うことに気付けない。
  it('names the current discipline and the hand counters', async () => {
    renderWithProviders(<HorsePage />);
    const info = await screen.findByTestId('ho-discipline');
    expect(info).toHaveTextContent('H');
    expect(info).toHaveTextContent('テキサスホールデム');
    expect(info).toHaveTextContent('ハンド 1/2');
    expect(screen.getByTestId('ho-pot')).toHaveTextContent('30');
  });

  it('shows every seat with its chips', async () => {
    renderWithProviders(<HorsePage />);
    expect(await screen.findByTestId('ho-seat-0')).toBeInTheDocument();
    for (const id of [0, 1, 2, 3]) {
      expect(screen.getByTestId(`ho-seat-${id}-chips`)).toHaveTextContent('1000');
    }
  });

  // **6 つの手をすべて送れる。** 綴りが 1 つ違うと、その手だけが打てなくなる。
  it.each([
    ['コール', 'call'],
    ['フォールド', 'fold'],
    ['オールイン', 'allin'],
  ] as const)('sends %s while a bet is outstanding', async (label, action) => {
    renderWithProviders(<HorsePage />);
    fireEvent.click(await screen.findByRole('button', { name: label }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('action', { action }));
  });

  // **賭けられていなければチェックとベットになる。** toCall を見ずに決め打つと、
  // チェックできる場面でコールしか出せない。
  it.each([
    ['チェック', 'check'],
    ['フォールド', 'fold'],
  ] as const)('sends %s when nothing is outstanding', async (label, action) => {
    mockExec.mockResolvedValue(makeHorseState({ toCall: 0 }));
    renderWithProviders(<HorsePage />);
    fireEvent.click(await screen.findByRole('button', { name: label }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('action', { action }));
  });

  // **ベットとレイズは額を添えて送る。** 添えないとサーバに断られる。
  it('sends bet with an amount when nothing is outstanding', async () => {
    mockExec.mockResolvedValue(makeHorseState({ toCall: 0 }));
    renderWithProviders(<HorsePage />);
    fireEvent.click(await screen.findByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('action', { action: 'bet', amount: expect.any(Number) }));
    const betCall = mockExec.mock.calls.find((c) => c[1] && 'amount' in c[1]);
    expect(betCall?.[1]?.amount).toBeGreaterThan(0);
  });

  it('sends raise with an amount while a bet is outstanding', async () => {
    renderWithProviders(<HorsePage />);
    fireEvent.click(await screen.findByRole('button', { name: 'レイズ' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('action', { action: 'raise', amount: expect.any(Number) }),
    );
  });

  it('shows the community cards once they are dealt', async () => {
    mockExec.mockResolvedValue(
      makeHorseState({
        communityCards: [
          { design: 'SPADE', value: 10, glyph: '♠', label: '10', color: 'black', deck: 'standard' },
          { design: 'HEART', value: 4, glyph: '♥', label: '4', color: 'red', deck: 'standard' },
          { design: 'CLOVER', value: 7, glyph: '♣', label: '7', color: 'black', deck: 'standard' },
        ],
      }),
    );
    renderWithProviders(<HorsePage />);
    const board = await screen.findByTestId('ho-community');
    expect(board).toBeInTheDocument();
  });

  // **共有札はスタッド系には無い。** 常に描くと、無いはずの場が出る。
  it('hides the community row in the stud disciplines', async () => {
    mockExec.mockResolvedValue(
      makeHorseState({ discipline: 2, disciplineLetter: 'R', disciplineName: 'razz', communityCards: [] }),
    );
    renderWithProviders(<HorsePage />);
    expect(await screen.findByTestId('ho-discipline')).toHaveTextContent('ラズ');
    expect(screen.queryByTestId('ho-community')).not.toBeInTheDocument();
  });

  it('offers the next hand once the hand is settled', async () => {
    mockExec.mockResolvedValue(makeHorseState({ phase: 1, isHumanTurn: false }));
    renderWithProviders(<HorsePage />);
    fireEvent.click(await screen.findByTestId('ho-next-hand'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  // **決着していない間はベッティングを出さない。** 出すと押せるのに何も起きない。
  it('hides the betting controls when it is not the human turn', async () => {
    mockExec.mockResolvedValue(makeHorseState({ isHumanTurn: false }));
    renderWithProviders(<HorsePage />);
    await screen.findByTestId('ho-discipline');
    expect(screen.queryByRole('button', { name: 'フォールド' })).not.toBeInTheDocument();
  });

  it('shows the final standings and starts a new match', async () => {
    mockExec.mockResolvedValue(
      makeHorseState({
        phase: 2,
        gameEndFlag: true,
        isHumanTurn: false,
        winnerSeat: 0,
        seats: handState.seats.map((s) => (s.id === 0 ? { ...s, chips: 4000 } : { ...s, chips: 0 })),
      }),
    );
    renderWithProviders(<HorsePage />);
    const result = await screen.findByTestId('ho-result');
    expect(result).toHaveTextContent('あなた');
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '新しいゲーム' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { seats: 4, handsPerDiscipline: 2 } }),
    );
  });
});
