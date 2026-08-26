import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { quodlibetApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeQuodlibetState } from '../test/stateFactories';
import { QuodlibetPage } from './QuodlibetPage';

vi.mock('../api/gameApi', () => ({
  quodlibetApi: { exec: vi.fn() },
  actionLogApi: { quodlibet: vi.fn() },
}));

const mockExec = vi.mocked(quodlibetApi.exec);

const contractState = makeQuodlibetState();

/** A Minus deal in progress with the human on turn. */
const playState = makeQuodlibetState({
  phase: 'play',
  isContractPhase: false,
  currentContract: 1,
  currentContractName: 'minus',
  availableContracts: [0, 2, 3],
  availableContractNames: ['plus', 'badNeighbour', 'alarich'],
  currentPlayerIdx: 0,
  playableIndices: [0, 1, 2, 3],
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(contractState);
});

describe('QuodlibetPage', () => {
  it('calls reset on mount with the configured options', async () => {
    renderWithProviders(<QuodlibetPage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { cpuDifficulty: 1, autoSelectContract: false } }),
    );
  });

  it('shows the deal and wheel counters', async () => {
    renderWithProviders(<QuodlibetPage />);
    expect(await screen.findByText('ディール 1/12')).toBeInTheDocument();
    expect(screen.getByText('第 1 の輪')).toBeInTheDocument();
  });

  // **選べるのはこの輪の残りだけ。** 全 12 種目を並べると、押せない選択肢を
  // 勧めることになる。
  it('offers only the contracts left in this wheel', async () => {
    renderWithProviders(<QuodlibetPage />);
    const box = await screen.findByTestId('quodlibet-contract-choices');
    expect(box).toHaveTextContent('プラス');
    expect(box).toHaveTextContent('アラリック');
    // ボタンの数まで見る。名前だけを見ると、12 種目全部を並べても
    // availableContractNames の外は未翻訳キーになるだけで素通りする。
    expect(box.querySelectorAll('button')).toHaveLength(4);
    // 第 2 の輪の種目 (id 4-7) は 1 つも出ない。
    expect(screen.queryByTestId('quodlibet-contract-4')).not.toBeInTheDocument();
    expect(screen.queryByTestId('quodlibet-contract-11')).not.toBeInTheDocument();
  });

  it('sends the chosen contract', async () => {
    renderWithProviders(<QuodlibetPage />);
    fireEvent.click(await screen.findByTestId('quodlibet-contract-2'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('contract', { contract: 2 }));
  });

  it('hides the contract buttons once the deal is under way', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<QuodlibetPage />);
    await screen.findByTestId('quodlibet-play');
    expect(screen.queryByTestId('quodlibet-contract-choices')).not.toBeInTheDocument();
    expect(screen.getByTestId('quodlibet-contract')).toHaveTextContent('マイナス');
  });

  it('plays the selected card', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<QuodlibetPage />);
    const cards = await screen.findAllByRole('button', { name: /♠|♥|♦|♣/ });
    fireEvent.click(cards[0]);
    fireEvent.click(screen.getByTestId('quodlibet-play'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  // **点は罰点で、少ないほうが勝ち。** 向きを書かないと多い人が勝っている
  // ように読める。
  it('labels the score column as penalty points', async () => {
    mockExec.mockResolvedValue(
      makeQuodlibetState({
        players: contractState.players.map((p, i) => ({ ...p, penalty: i === 0 ? 120 : 40 })),
        winners: [1, 2, 3],
      }),
    );
    renderWithProviders(<QuodlibetPage />);
    expect(await screen.findByText('罰点（少ないほうが勝ち）')).toBeInTheDocument();
    expect(screen.getByTestId('quodlibet-scores')).toHaveTextContent('120 点');
  });

  // **パスできるのは出せる札が 1 枚も無いときだけ。**
  it('offers pass only when a shedding contract leaves nothing playable', async () => {
    mockExec.mockResolvedValue(
      makeQuodlibetState({
        ...playState,
        currentContract: 11,
        currentContractName: 'snack',
        isShedding: true,
        playableIndices: [],
        canPass: true,
      }),
    );
    renderWithProviders(<QuodlibetPage />);
    fireEvent.click(await screen.findByTestId('quodlibet-pass'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass'));
  });

  it('hides pass while a card can still be played', async () => {
    mockExec.mockResolvedValue(
      makeQuodlibetState({ ...playState, currentContract: 11, isShedding: true, canPass: false }),
    );
    renderWithProviders(<QuodlibetPage />);
    await screen.findByTestId('quodlibet-play');
    expect(screen.queryByTestId('quodlibet-pass')).not.toBeInTheDocument();
  });

  // **四分と小食いはトリックではない。** トリック欄を出しても何も並ばない。
  it('shows the shed area instead of a trick for shedding contracts', async () => {
    mockExec.mockResolvedValue(
      makeQuodlibetState({
        ...playState,
        currentContract: 10,
        currentContractName: 'quadrature',
        isShedding: true,
        stack: [{ design: 'SPADE' as const, value: 7, color: 'black' }],
      }),
    );
    renderWithProviders(<QuodlibetPage />);
    expect(await screen.findByTestId('quodlibet-shed-area')).toBeInTheDocument();
    // シェディングではトリック番号も出さない。
    expect(screen.queryByText(/トリック 1\/8/)).not.toBeInTheDocument();
  });

  it('advances the deal and shows the breakdown', async () => {
    mockExec.mockResolvedValue(
      makeQuodlibetState({
        phase: 'dealEnd',
        isContractPhase: false,
        isHumanTurn: false,
        currentContract: 1,
        currentContractName: 'minus',
        lastDeal: { contract: 1, contractName: 'minus', round: 0, dealerIdx: 0, points: [30, 0, 20, 30] },
      }),
    );
    renderWithProviders(<QuodlibetPage />);
    const box = await screen.findByTestId('quodlibet-deal-result');
    expect(box).toHaveTextContent('マイナス');
    expect(box).toHaveTextContent('30 点');
    fireEvent.click(screen.getByTestId('quodlibet-next-deal'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextdeal'));
  });

  it('names the seats on the fewest penalty points at the end', async () => {
    mockExec.mockResolvedValue(
      makeQuodlibetState({ phase: 'gameEnd', gameEndFlag: true, isContractPhase: false, winners: [2] }),
    );
    renderWithProviders(<QuodlibetPage />);
    expect(await screen.findByTestId('quodlibet-winner')).toHaveTextContent('CPU 2');
  });

  // ヒントのゲート: 頼んでいないヒントは出さない。
  it('does not render the hint banner unless it was requested', async () => {
    mockExec.mockResolvedValue(
      makeQuodlibetState({ ...playState, hint: { cardIndices: [0], reason: 'avoid_penalty' }, messageCode: '' }),
    );
    renderWithProviders(<QuodlibetPage />);
    await screen.findByTestId('quodlibet-play');
    expect(screen.queryByText(/\(\[0\]\)/)).not.toBeInTheDocument();
  });

  it('renders the hint banner once the hint was requested', async () => {
    mockExec.mockResolvedValue(
      makeQuodlibetState({
        ...playState,
        hint: { cardIndices: [0], reason: 'avoid_penalty' },
        messageCode: 'quodlibet.hintRequested',
      }),
    );
    renderWithProviders(<QuodlibetPage />);
    expect(await screen.findByText(/\(\[0\]\)/)).toBeInTheDocument();
  });
});
