import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { costlycoloursApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeCostlyColoursState } from '../test/stateFactories';
import { CostlyColoursPage } from './CostlyColoursPage';

vi.mock('../api/gameApi', () => ({
  costlycoloursApi: { exec: vi.fn() },
  actionLogApi: { costlycolours: vi.fn() },
}));

const mockExec = vi.mocked(costlycoloursApi.exec);

const mogState = makeCostlyColoursState();
const playState = makeCostlyColoursState({ phase: 'play', playableIdxs: [0, 1, 2] });

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(mogState);
});

describe('CostlyColoursPage', () => {
  it('calls reset on mount with the configured target', async () => {
    renderWithProviders(<CostlyColoursPage />);
    // **既定は Cotton の 61 点。**
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { cpuDifficulty: 1, targetScore: 61 } }),
    );
  });

  it('shows the deal and the turn-up', async () => {
    renderWithProviders(<CostlyColoursPage />);
    expect(await screen.findByText('ディール 1（61 点勝負）')).toBeInTheDocument();
    // **表の 1 枚は常に見せる。** ショーの色役も J / 2 の 4 点もこれ次第。
    const turnUp = screen.getByTestId('costlycolours-turnup');
    expect(turnUp.children).toHaveLength(1);
  });

  it('shows the running count', async () => {
    mockExec.mockResolvedValue(makeCostlyColoursState({ phase: 'play', total: 24, playableIdxs: [0] }));
    renderWithProviders(<CostlyColoursPage />);
    expect(await screen.findByTestId('costlycolours-total')).toHaveTextContent('24');
  });

  // **応じる／断るは別のボタン。** 断ると相手に 1 点入るので、片方を既定にしない。
  it('offers both sides of the exchange, and sends each explicitly', async () => {
    renderWithProviders(<CostlyColoursPage />);
    fireEvent.click(await screen.findByTestId('costlycolours-mog'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('mog', { accept: true }));

    mockExec.mockClear();
    renderWithProviders(<CostlyColoursPage />);
    const refusals = await screen.findAllByTestId('costlycolours-nomog');
    fireEvent.click(refusals[refusals.length - 1]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('mog', { accept: false }));
  });

  it('hides the exchange buttons once the count starts', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<CostlyColoursPage />);
    await screen.findByTestId('costlycolours-pile');
    expect(screen.queryByTestId('costlycolours-mog')).not.toBeInTheDocument();
    expect(screen.queryByTestId('costlycolours-nomog')).not.toBeInTheDocument();
  });

  it('plays the card that is clicked', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<CostlyColoursPage />);
    const cards = await screen.findAllByRole('button', { name: /♠|♥|♦|♣/ });
    fireEvent.click(cards[1]);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { handIndex: 1 }));
  });

  it('shows every seat with its score', async () => {
    renderWithProviders(<CostlyColoursPage />);
    expect(await screen.findByTestId('costlycolours-scores')).toHaveTextContent('得点 0');
  });

  // **どの色役が付いたのかを名指す。** 点だけだと梯子のどこか分からない。
  it('names the colour combination in the show', async () => {
    mockExec.mockResolvedValue(
      makeCostlyColoursState({
        phase: 'show',
        isHumanTurn: false,
        lastResult: {
          lines: [
            { key: 'jackDeuce', points: [2, 0] },
            { key: 'rank', points: [0, 0] },
            { key: 'colour', points: [6, 0] },
          ],
          totals: [8, 0],
          combos: ['costlyColours', ''],
        },
      }),
    );
    renderWithProviders(<CostlyColoursPage />);
    const show = await screen.findByTestId('costlycolours-show');
    expect(show).toHaveTextContent('J と 2');
    expect(show).toHaveTextContent('色とスート');
    expect(show).not.toHaveTextContent('同位役', { normalizeWhitespace: true });
    expect(screen.getByTestId('costlycolours-combo-0')).toHaveTextContent('コストリー・カラーズ');
    expect(screen.queryByTestId('costlycolours-combo-1')).not.toBeInTheDocument();

    fireEvent.click(screen.getByTestId('costlycolours-next-deal'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextdeal'));
  });

  it('shows the winner at game end', async () => {
    mockExec.mockResolvedValue(makeCostlyColoursState({ phase: 'gameEnd', gameEndFlag: true, winnerIdx: 0 }));
    renderWithProviders(<CostlyColoursPage />);
    expect(await screen.findByTestId('costlycolours-winner')).toBeInTheDocument();
  });

  // **ヒントのライブ領域は常設。** 出る側と出ない側の両方を見る。
  // 交換フェーズは札を指さないので、hintHandIdx が -1 でも出る必要がある。
  it('announces a requested hint even when no card is named', async () => {
    renderWithProviders(<CostlyColoursPage />);
    expect(await screen.findByTestId('costlycolours-hint-live')).toBeEmptyDOMElement();

    mockExec.mockResolvedValue(
      makeCostlyColoursState({
        messageCode: 'costlycolours.hintRequested',
        hintHandIdx: -1,
        hintReason: 'mog_refuse',
      }),
    );
    renderWithProviders(<CostlyColoursPage />);
    await waitFor(() => {
      const lives = screen.getAllByTestId('costlycolours-hint-live');
      expect(lives[lives.length - 1]).not.toBeEmptyDOMElement();
    });
  });
});
