import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { piedmonteseTarotApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makePiedmonteseTarotState } from '../test/stateFactories';
import { PiedmonteseTarotPage } from './PiedmonteseTarotPage';

vi.mock('../api/gameApi', () => ({
  piedmonteseTarotApi: { exec: vi.fn() },
  actionLogApi: { piedmontesetarot: vi.fn() },
}));

const mockExec = vi.mocked(piedmonteseTarotApi.exec);

const suit = (value: number, design: 'HEART' | 'SPADE' | 'CLOVER' | 'DIAMOND', glyph: string, label: string) => ({
  design,
  value,
  glyph,
  label,
  color: design === 'HEART' || design === 'DIAMOND' ? 'red' : 'black',
  deck: 'tarot',
});

const playState = makePiedmonteseTarotState();

/** The human deals: two cards to bury, a hand mixing pips, a Roi and the Matto. */
const scartoState = makePiedmonteseTarotState({
  phase: 0,
  isHumanTurn: false,
  isHumanScarto: true,
  scartoCount: 0,
  playableIndices: [],
  dealerIdx: 0,
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 21,
      cards: [
        suit(2, 'HEART', '♥', '2'),
        suit(3, 'HEART', '♥', '3'),
        suit(14, 'SPADE', '♠', 'R'),
        { design: 'JOKER' as const, value: 0, glyph: '★', label: 'Matto', color: 'gold', deck: 'tarot' },
      ],
      trickCount: 0,
      cardThirds: 0,
      cardPoints: '0',
      score: 0,
      isDealer: true,
    },
    ...makePiedmonteseTarotState()
      .players.slice(1)
      .map((p) => ({ ...p, isDealer: false })),
  ],
});

/** Three-handed table: the talon is three cards, not two. */
const threeHandedScartoState = makePiedmonteseTarotState({
  ...scartoState,
  talonSize: 3,
  trickCount: 25,
  config: { seats: 3, cpuDifficulty: 1, targetDeals: 4 },
});

const roundEndState = makePiedmonteseTarotState({
  phase: 3,
  isHumanTurn: false,
  playableIndices: [],
  outcome: 1,
  dealScores: [12, -4, -4, -4],
  players: makePiedmonteseTarotState().players.map((p, i) => ({
    ...p,
    cardThirds: i === 0 ? 90 : 48,
    cardPoints: i === 0 ? '30' : '16',
    score: i === 0 ? 12 : -4,
  })),
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playState);
});

describe('PiedmonteseTarotPage', () => {
  it('calls reset on mount with the configured table', async () => {
    renderWithProviders(<PiedmonteseTarotPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { seats: 4, cpuDifficulty: 1, targetDeals: 4 } }),
    );
  });

  it('shows the deal and trick counters', async () => {
    renderWithProviders(<PiedmonteseTarotPage />);
    expect(await screen.findByText('ディール 1')).toBeInTheDocument();
    expect(screen.getByText('トリック 1/19')).toBeInTheDocument();
  });

  it('plays the selected card', async () => {
    renderWithProviders(<PiedmonteseTarotPage />);
    const cards = await screen.findAllByRole('button', { name: /♥|♠|♣|✦|★/ });
    fireEvent.click(cards[0]);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  // **捨てる枚数は卓が決める。** 定数にすると 3 人卓でボタンが押せない。
  it('asks for two cards at a four-handed table', async () => {
    mockExec.mockResolvedValue(scartoState);
    renderWithProviders(<PiedmonteseTarotPage />);
    const prompt = await screen.findByTestId('piedmontesetarot-discard-prompt');
    expect(prompt).toHaveTextContent('0/2');

    const button = screen.getByRole('button', { name: /捨てる/ });
    expect(button).toBeDisabled();
    const cards = await screen.findAllByRole('button', { name: /♥|♠|★/ });
    fireEvent.click(cards[0]);
    fireEvent.click(cards[1]);
    expect(button).toBeEnabled();
    fireEvent.click(button);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('scarto', { cardIndices: [0, 1] }));
  });

  it('asks for three cards at a three-handed table', async () => {
    mockExec.mockResolvedValue(threeHandedScartoState);
    renderWithProviders(<PiedmonteseTarotPage />);
    expect(await screen.findByTestId('piedmontesetarot-discard-prompt')).toHaveTextContent('0/3');
    const cards = await screen.findAllByRole('button', { name: /♥|♠|★/ });
    fireEvent.click(cards[0]);
    fireEvent.click(cards[1]);
    // 2 枚では足りない ── 4 人卓の枚数で判定していれば、ここで押せてしまう。
    expect(screen.getByRole('button', { name: /捨てる/ })).toBeDisabled();
  });

  it('waits while a CPU deals', async () => {
    mockExec.mockResolvedValue(
      makePiedmonteseTarotState({ phase: 0, isHumanTurn: false, isHumanScarto: false, playableIndices: [] }),
    );
    renderWithProviders(<PiedmonteseTarotPage />);
    expect(await screen.findByTestId('piedmontesetarot-waiting')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /捨てる/ })).not.toBeInTheDocument();
  });

  it('advances the trick', async () => {
    mockExec.mockResolvedValue(makePiedmonteseTarotState({ phase: 2, isHumanTurn: false, playableIndices: [] }));
    renderWithProviders(<PiedmonteseTarotPage />);
    fireEvent.click(await screen.findByRole('button', { name: '次のトリック' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('advances the deal', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<PiedmonteseTarotPage />);
    fireEvent.click(await screen.findByRole('button', { name: '次のディール' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  // **精算の式まで出す。** 取り分と変動は席数倍ちがうので、並べただけでは
  // 計算が合わないように見える。
  it('breaks the settlement down at the end of a deal', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<PiedmonteseTarotPage />);
    const box = await screen.findByTestId('piedmontesetarot-result');
    expect(box).toHaveTextContent('平均より上');
    const breakdown = screen.getByTestId('piedmontesetarot-breakdown');
    expect(breakdown).toHaveTextContent('78');
    expect(breakdown).toHaveTextContent('+12');
    expect(screen.getByTestId('piedmontesetarot-formula')).toHaveTextContent('4');
  });

  // ヒントのゲート: 頼んでいないヒントは出さない。
  it('does not render the hint banner unless it was requested', async () => {
    mockExec.mockResolvedValue(
      makePiedmonteseTarotState({ hint: { cardIndices: [0], reason: 'lead_low' }, messageCode: '' }),
    );
    renderWithProviders(<PiedmonteseTarotPage />);
    await screen.findByText('ディール 1');
    expect(screen.queryByText(/\(\[0\]\)/)).not.toBeInTheDocument();
  });

  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue(
      makePiedmonteseTarotState({
        hint: { cardIndices: [0], reason: 'lead_low' },
        messageCode: 'piedmontesetarot.hintRequested',
      }),
    );
    renderWithProviders(<PiedmonteseTarotPage />);
    expect(await screen.findByText(/\(\[0\]\)/)).toBeInTheDocument();
  });
});
