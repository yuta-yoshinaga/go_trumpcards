import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { sutdaApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeSutdaState } from '../test/stateFactories';
import { SutdaPage } from './SutdaPage';

vi.mock('../api/gameApi', () => ({
  sutdaApi: { exec: vi.fn() },
  actionLogApi: { sutda: vi.fn() },
}));

const mockExec = vi.mocked(sutdaApi.exec);

const betState = makeSutdaState();

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(betState);
});

describe('SutdaPage', () => {
  it('calls reset on mount with the configured table', async () => {
    renderWithProviders(<SutdaPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { cpuDifficulty: 1, seats: 3, startChips: 1000 },
      }),
    );
  });

  it('shows the hand counter and the pot', async () => {
    renderWithProviders(<SutdaPage />);
    expect(await screen.findByText('ハンド 1')).toBeInTheDocument();
    expect(screen.getByTestId('sutda-pot')).toHaveTextContent('ポット 30');
  });

  // **自分の役は常に見える。** 伏せているのは相手の札だけ。
  it('always shows your own two cards and what they make', async () => {
    renderWithProviders(<SutdaPage />);
    expect(await screen.findByTestId('sutda-hand')).toHaveTextContent('38光ッタン');
  });

  it('keeps opponents face down until they are revealed', async () => {
    renderWithProviders(<SutdaPage />);
    await screen.findByTestId('sutda-hand');
    expect(screen.queryByTestId('sutda-cards-1')).not.toBeInTheDocument();
    expect(screen.getByTestId('sutda-cards-0')).toBeInTheDocument();
  });

  it('shows an opponent hand once it is revealed', async () => {
    const base = makeSutdaState();
    mockExec.mockResolvedValue(
      makeSutdaState({
        players: base.players.map((p, i) =>
          i === 1 ? { ...p, revealed: true, cards: base.players[0].cards, handName: 'mangtong', handRank: 600 } : p,
        ),
      }),
    );
    renderWithProviders(<SutdaPage />);
    expect(await screen.findByTestId('sutda-cards-1')).toHaveTextContent('マントン');
  });

  // 差額 0 のときはチェック、要るときはコール。
  it('labels the call button by what is owed', async () => {
    renderWithProviders(<SutdaPage />);
    expect(await screen.findByTestId('sutda-call')).toHaveTextContent('チェック');
    expect(screen.getByTestId('sutda-to-call')).toHaveTextContent('追加は不要');

    mockExec.mockResolvedValue(makeSutdaState({ callAmount: 20 }));
    fireEvent.click(screen.getByTestId('sutda-call'));
    await waitFor(() => expect(screen.getByTestId('sutda-call')).toHaveTextContent('コール（20）'));
    expect(screen.getByTestId('sutda-to-call')).toHaveTextContent('20');
  });

  it.each([
    ['sutda-call', 'call'],
    ['sutda-raise', 'raise'],
    ['sutda-fold', 'fold'],
  ])('%s sends %s', async (testId, command) => {
    renderWithProviders(<SutdaPage />);
    fireEvent.click(await screen.findByTestId(testId));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith(command));
  });

  // **押せる条件はサーバの canRaise がすべて。** 上限とチップの両方を見た
  // 結果なので、画面側で組み直すと判断が食い違う。
  it('hides raise when the server says it is not available', async () => {
    mockExec.mockResolvedValue(makeSutdaState({ canRaise: false }));
    renderWithProviders(<SutdaPage />);
    await screen.findByTestId('sutda-call');
    expect(screen.queryByTestId('sutda-raise')).not.toBeInTheDocument();
  });

  it('offers no betting buttons when it is not your turn', async () => {
    mockExec.mockResolvedValue(makeSutdaState({ isHumanTurn: false, currentPlayerIdx: 1 }));
    renderWithProviders(<SutdaPage />);
    await screen.findByTestId('sutda-hand');
    expect(screen.queryByTestId('sutda-call')).not.toBeInTheDocument();
    expect(screen.queryByTestId('sutda-fold')).not.toBeInTheDocument();
  });

  it('marks a folded seat', async () => {
    const base = makeSutdaState();
    mockExec.mockResolvedValue(
      makeSutdaState({ players: base.players.map((p, i) => (i === 1 ? { ...p, folded: true } : p)) }),
    );
    renderWithProviders(<SutdaPage />);
    expect(await screen.findByTestId('sutda-folded-1')).toBeInTheDocument();
    expect(screen.queryByTestId('sutda-folded-0')).not.toBeInTheDocument();
  });

  it('shows the showdown result and advances the hand', async () => {
    mockExec.mockResolvedValue(
      makeSutdaState({
        phase: 'showdown',
        isShowdown: true,
        isHumanTurn: false,
        lastResult: {
          winners: [0],
          pot: 70,
          handNames: ['gwang38', 'mangtong', 'kkeut5'],
          folded: [false, false, false],
        },
      }),
    );
    renderWithProviders(<SutdaPage />);
    expect(await screen.findByTestId('sutda-result')).toHaveTextContent('70');
    fireEvent.click(screen.getByTestId('sutda-next-hand'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nexthand'));
  });

  it('names the winner at the end of the table', async () => {
    mockExec.mockResolvedValue(
      makeSutdaState({ phase: 'gameEnd', gameEndFlag: true, isHumanTurn: false, winnerIdx: 2 }),
    );
    renderWithProviders(<SutdaPage />);
    expect(await screen.findByTestId('sutda-winner')).toHaveTextContent('CPU 2');
  });

  // ヒントのゲート: 頼んでいないヒントは出さない。
  it('does not render the hint banner unless it was requested', async () => {
    mockExec.mockResolvedValue(makeSutdaState({ hintAction: 'raise', hintReason: 'strong_hand', messageCode: '' }));
    renderWithProviders(<SutdaPage />);
    await screen.findByTestId('sutda-hand');
    expect(screen.queryByText(/強い役なので/)).not.toBeInTheDocument();
  });

  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue(
      makeSutdaState({ hintAction: 'raise', hintReason: 'strong_hand', messageCode: 'sutda.hintRequested' }),
    );
    renderWithProviders(<SutdaPage />);
    expect(await screen.findByText(/強い役なので/)).toBeInTheDocument();
  });
});
