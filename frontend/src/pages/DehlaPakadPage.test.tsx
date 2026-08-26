import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { dehlaPakadApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeDehlaPakadState } from '../test/stateFactories';
import { DehlaPakadPage } from './DehlaPakadPage';

vi.mock('../api/gameApi', () => ({
  dehlaPakadApi: { exec: vi.fn() },
  actionLogApi: { dehlapakad: vi.fn() },
}));

const mockExec = vi.mocked(dehlaPakadApi.exec);

const trumpState = makeDehlaPakadState();

/** Hearts called, the human on turn with all 13 cards. */
const playState = makeDehlaPakadState({
  phase: 'play',
  isTrumpPhase: false,
  trumpSuit: 3,
  trumpSuitName: 'heart',
  currentPlayerIdx: 0,
  playableIndices: [0, 1, 2, 3, 4],
  players: trumpState.players.map((p) => ({ ...p, cardCount: 13, isTrumpChooser: false })),
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(trumpState);
});

describe('DehlaPakadPage', () => {
  it('calls reset on mount with the configured match length', async () => {
    renderWithProviders(<DehlaPakadPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { cpuDifficulty: 1, targetKots: 2 } }),
    );
  });

  it('shows the hand counter and the match target', async () => {
    renderWithProviders(<DehlaPakadPage />);
    expect(await screen.findByText('ハンド 1（2 コート先取）')).toBeInTheDocument();
  });

  // **決めるのは親の右隣で、見えているのは最初の 5 枚だけ。**
  it('offers the four suits while the trump is being called', async () => {
    renderWithProviders(<DehlaPakadPage />);
    const box = await screen.findByTestId('dehlapakad-trump-choices');
    expect(box.querySelectorAll('button')).toHaveLength(4);
    fireEvent.click(screen.getByTestId('dehlapakad-trump-3'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('trump', { trumpSuit: 3 }));
  });

  it('hides the suit buttons once the trump is called', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<DehlaPakadPage />);
    await screen.findByTestId('dehlapakad-play');
    expect(screen.queryByTestId('dehlapakad-trump-choices')).not.toBeInTheDocument();
    expect(screen.getByTestId('dehlapakad-trump')).toHaveTextContent('ハート');
  });

  it('plays the selected card', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<DehlaPakadPage />);
    const cards = await screen.findAllByRole('button', { name: /♠|♥|♦|♣/ });
    fireEvent.click(cards[0]);
    fireEvent.click(screen.getByTestId('dehlapakad-play'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  // **これがこのゲームの心臓部。** 取っただけでは札は手に入らない。
  it('shows what is sitting in the centre and who would collect it', async () => {
    mockExec.mockResolvedValue(
      makeDehlaPakadState({ ...playState, centrePileCount: 8, centrePileTens: 2, prevTrickWinner: 1 }),
    );
    renderWithProviders(<DehlaPakadPage />);
    const pile = await screen.findByTestId('dehlapakad-centre-pile');
    expect(pile).toHaveTextContent('中央に 8 枚（うち 10 が 2 枚）');
    expect(screen.getByTestId('dehlapakad-pile-goes-to')).toHaveTextContent('CPU 1');
  });

  it('hides the centre pile while nothing is waiting', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<DehlaPakadPage />);
    await screen.findByTestId('dehlapakad-play');
    expect(screen.queryByTestId('dehlapakad-centre-pile')).not.toBeInTheDocument();
  });

  it('reads the team columns from the human seat', async () => {
    mockExec.mockResolvedValue(makeDehlaPakadState({ ...playState, humanTeam: 1, teamTens: [3, 1], teamKots: [1, 0] }));
    renderWithProviders(<DehlaPakadPage />);
    const box = await screen.findByTestId('dehlapakad-scores');
    expect(box).toHaveTextContent('味方 1 / 相手 3');
    expect(box).toHaveTextContent('味方 0 / 相手 1');
  });

  // **7 連勝もコートになる。** 出さないと、なぜ同じ組が勝ち続けているのかが
  // 数字にならない。
  it('surfaces a winning streak', async () => {
    mockExec.mockResolvedValue(makeDehlaPakadState({ ...playState, streakTeam: 1, streakCount: 4 }));
    renderWithProviders(<DehlaPakadPage />);
    expect(await screen.findByTestId('dehlapakad-streak')).toHaveTextContent('4 連勝');
  });

  it('does not claim a streak on a single win', async () => {
    mockExec.mockResolvedValue(makeDehlaPakadState({ ...playState, streakTeam: 1, streakCount: 1 }));
    renderWithProviders(<DehlaPakadPage />);
    await screen.findByTestId('dehlapakad-play');
    expect(screen.queryByTestId('dehlapakad-streak')).not.toBeInTheDocument();
  });

  it('advances the hand and shows the tens breakdown', async () => {
    mockExec.mockResolvedValue(
      makeDehlaPakadState({
        phase: 'handEnd',
        isTrumpPhase: false,
        isHumanTurn: false,
        lastHand: { winnerTeam: 0, teamTens: [3, 1], kot: false, kotReason: '', dealerIdx: 3, trumpSuit: 3 },
      }),
    );
    renderWithProviders(<DehlaPakadPage />);
    expect(await screen.findByTestId('dehlapakad-hand-result')).toHaveTextContent('10 の枚数 3 対 1');
    expect(screen.queryByTestId('dehlapakad-kot')).not.toBeInTheDocument();
    fireEvent.click(screen.getByTestId('dehlapakad-next-hand'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nexthand'));
  });

  it.each([
    ['allTens', '10 を 4 枚とも取りました'],
    ['streak', '7 連勝を達成しました'],
  ])('names the reason for a kot (%s)', async (kotReason, text) => {
    mockExec.mockResolvedValue(
      makeDehlaPakadState({
        phase: 'handEnd',
        isTrumpPhase: false,
        isHumanTurn: false,
        lastHand: { winnerTeam: 0, teamTens: [4, 0], kot: true, kotReason, dealerIdx: 3, trumpSuit: 3 },
      }),
    );
    renderWithProviders(<DehlaPakadPage />);
    expect(await screen.findByTestId('dehlapakad-kot')).toHaveTextContent(text);
  });

  it('names the winning team at the end of the match', async () => {
    mockExec.mockResolvedValue(
      makeDehlaPakadState({ phase: 'gameEnd', gameEndFlag: true, isTrumpPhase: false, winnerTeam: 1 }),
    );
    renderWithProviders(<DehlaPakadPage />);
    expect(await screen.findByTestId('dehlapakad-winner')).toHaveTextContent('組1 の勝ちです');
  });

  // ヒントのゲート: 頼んでいないヒントは出さない。
  it('does not render the hint banner unless it was requested', async () => {
    mockExec.mockResolvedValue(
      makeDehlaPakadState({ ...playState, hint: { cardIndices: [0], reason: 'take_the_ten' }, messageCode: '' }),
    );
    renderWithProviders(<DehlaPakadPage />);
    await screen.findByTestId('dehlapakad-play');
    expect(screen.queryByText(/\(\[0\]\)/)).not.toBeInTheDocument();
  });

  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue(
      makeDehlaPakadState({
        ...playState,
        hint: { cardIndices: [0], reason: 'take_the_ten' },
        messageCode: 'dehlapakad.hintRequested',
      }),
    );
    renderWithProviders(<DehlaPakadPage />);
    expect(await screen.findByText(/\(\[0\]\)/)).toBeInTheDocument();
  });
});
