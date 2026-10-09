import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { costlycoloursApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeCostlyColoursState } from '../test/stateFactories';
import type { Card } from '../types/card';
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
  it('reveals both show hands beside their score rows only at show and game end', async () => {
    const base = makeCostlyColoursState();
    const cards: Card[] = [
      { design: 'SPADE', value: 3 },
      { design: 'HEART', value: 5 },
      { design: 'CLOVER', value: 7 },
    ];
    mockExec.mockResolvedValue(
      makeCostlyColoursState({
        players: [base.players[0], { ...base.players[1], cards }],
      }),
    );
    const hidden = renderWithProviders(<CostlyColoursPage />);
    expect(await screen.findByTestId('costlycolours-scores')).toBeInTheDocument();
    expect(screen.queryByTestId('costlycolours-show-hand-1')).not.toBeInTheDocument();
    hidden.unmount();

    for (const state of [
      makeCostlyColoursState({ phase: 'show', players: [base.players[0], { ...base.players[1], cards }] }),
      makeCostlyColoursState({
        phase: 'gameEnd',
        gameEndFlag: true,
        players: [base.players[0], { ...base.players[1], cards }],
      }),
    ]) {
      mockExec.mockResolvedValue(state);
      const page = renderWithProviders(<CostlyColoursPage />);
      const humanHand = await screen.findByTestId('costlycolours-show-hand-0');
      const cpuHand = screen.getByTestId('costlycolours-show-hand-1');
      expect(humanHand).toHaveTextContent('あなたの手札');
      expect(cpuHand).toHaveTextContent('CPU 1の手札');
      expect(cpuHand.querySelectorAll('img')).toHaveLength(4);
      page.unmount();
    }
  });
  it('shows a waiting note only while another player acts', async () => {
    mockExec.mockResolvedValue(makeCostlyColoursState({ phase: 'play', isHumanTurn: false }));
    const { unmount } = renderWithProviders(<CostlyColoursPage />);
    expect(await screen.findByText('相手の手番です。お待ちください。')).toBeInTheDocument();
    unmount();

    mockExec.mockResolvedValue(makeCostlyColoursState({ phase: 'play', isHumanTurn: true }));
    const humanTurn = renderWithProviders(<CostlyColoursPage />);
    expect(await screen.findByTestId('costlycolours-total')).toBeInTheDocument();
    expect(screen.queryByText('相手の手番です。お待ちください。')).not.toBeInTheDocument();
    humanTurn.unmount();

    for (const state of [
      makeCostlyColoursState({ phase: 'mog', isHumanTurn: true }),
      makeCostlyColoursState({ phase: 'show', isHumanTurn: true }),
      makeCostlyColoursState({ phase: 'gameEnd', gameEndFlag: true, isHumanTurn: false }),
    ]) {
      mockExec.mockResolvedValue(state);
      const page = renderWithProviders(<CostlyColoursPage />);
      await screen.findByTestId(state.phase === 'gameEnd' ? 'costlycolours-winner' : 'costlycolours-total');
      expect(screen.queryByText('相手の手番です。お待ちください。')).not.toBeInTheDocument();
      page.unmount();
    }
  });

  it('shows each held J and 2 badge with its score, including when there is no turn-up', async () => {
    mockExec.mockResolvedValue(
      makeCostlyColoursState({
        players: [
          {
            ...makeCostlyColoursState().players[0],
            cards: [
              { design: 'HEART', value: 11 },
              { design: 'CLOVER', value: 2 },
              { design: 'HEART', value: 5 },
            ],
          },
          makeCostlyColoursState().players[1],
        ],
      }),
    );
    const page = renderWithProviders(<CostlyColoursPage />);
    expect(await screen.findByTitle('J・2: 保持すると 4 点')).toBeInTheDocument();
    expect(screen.getByTitle('J・2: 保持すると 2 点')).toBeInTheDocument();
    expect(screen.getByTitle('J・2: 保持すると 4 点')).toHaveTextContent('4');
    expect(screen.getByTitle('J・2: 保持すると 2 点')).toHaveTextContent('2');
    expect(screen.getAllByTitle(/J・2/)).toHaveLength(2);
    page.unmount();

    const defaultState = makeCostlyColoursState();
    mockExec.mockResolvedValue(
      makeCostlyColoursState({
        turnUp: null,
        players: [
          {
            ...defaultState.players[0],
            cards: [
              { design: 'SPADE', value: 11 },
              { design: 'HEART', value: 2 },
            ],
          },
          defaultState.players[1],
        ],
      }),
    );
    renderWithProviders(<CostlyColoursPage />);
    await screen.findByTestId('costlycolours-turnup');
    const noTurnUpBadges = await screen.findAllByTitle(/J・2: 保持すると 2 点/);
    expect(noTurnUpBadges).toHaveLength(2);
  });
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

  it('announces recent plays, including scoring milestones, in its own permanent live region', async () => {
    const card = { design: 'SPADE', value: 7 } as const;
    mockExec.mockResolvedValue(
      makeCostlyColoursState({
        phase: 'play',
        recentPlays: [
          { seat: 0, card, total: 31, points: 2 },
          { seat: 1, card: { design: 'HEART', value: 4 }, total: 9, points: 0 },
        ],
      }),
    );
    renderWithProviders(<CostlyColoursPage />);
    const live = await screen.findByTestId('costlycolours-play-live');
    expect(live).toHaveAttribute('role', 'status');
    expect(live).toHaveAttribute('aria-live', 'polite');
    expect(live).toHaveTextContent('あなた: ♠ 7、累計 31、2 点獲得（31 到達）');
    expect(live).toHaveTextContent('CPU 1: ♥ 4、累計 9');
    expect(live).toHaveClass('sr-only');
    const recent = screen.getByTestId('costlycolours-recent-plays');
    expect(recent).toHaveTextContent('あなた: ♠ 7、累計 31、2 点獲得（31 到達）');
    expect(recent).toHaveTextContent('CPU 1: ♥ 4、累計 9');
    expect(screen.getByTestId('costlycolours-hint-live')).toBeInTheDocument();
  });

  it('announces a 15 point and an ordinary count without adding points', async () => {
    mockExec.mockResolvedValue(
      makeCostlyColoursState({
        phase: 'play',
        recentPlays: [
          { seat: 0, card: { design: 'SPADE', value: 7 }, total: 15, points: 2 },
          { seat: 1, card: { design: 'HEART', value: 10 }, total: 25, points: 2 },
          { seat: 0, card: { design: 'HEART', value: 4 }, total: 19, points: 0 },
        ],
      }),
    );
    renderWithProviders(<CostlyColoursPage />);
    const live = await screen.findByTestId('costlycolours-play-live');
    expect(live).toHaveTextContent('累計 15、2 点獲得（15 到達）');
    expect(live).toHaveTextContent('累計 25、2 点獲得（25 到達）');
    expect(live).toHaveTextContent('累計 19');
    expect(live).not.toHaveTextContent('累計 19、');
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

  // **押す前にどの札が出ていくのか分かる (#6631)。**
  describe('mog discard preview', () => {
    it('names the card that accepting would give away', async () => {
      mockExec.mockResolvedValue(
        makeCostlyColoursState({
          phase: 'mog',
          mogDiscardCard: { design: 'DIAMOND', value: 10 },
        }),
      );
      renderWithProviders(<CostlyColoursPage />);
      const note = await screen.findByTestId('costlycolours-mog-discard');
      expect(note.textContent).toContain('♦');
      expect(note.textContent).not.toContain('{{');
    });

    // **札が変われば表示も変わる。** 定数を出しているだけの実装だと通らない。
    it('follows the served card', async () => {
      mockExec.mockResolvedValue(
        makeCostlyColoursState({ phase: 'mog', mogDiscardCard: { design: 'SPADE', value: 1 } }),
      );
      renderWithProviders(<CostlyColoursPage />);
      const note = await screen.findByTestId('costlycolours-mog-discard');
      expect(note.textContent).toContain('♠');
      expect(note.textContent).not.toContain('♦');
    });

    // **断る側の挙動は変えない** (受け入れ条件)。予告が出ていてもボタンは両方残る。
    it('leaves both buttons in place', async () => {
      mockExec.mockResolvedValue(
        makeCostlyColoursState({ phase: 'mog', mogDiscardCard: { design: 'DIAMOND', value: 10 } }),
      );
      renderWithProviders(<CostlyColoursPage />);
      expect(await screen.findByTestId('costlycolours-mog')).toBeInTheDocument();
      expect(screen.getByTestId('costlycolours-nomog')).toBeInTheDocument();
    });

    // サーバが送らない局面では何も出さない。
    it('shows nothing when the server sends no card', async () => {
      mockExec.mockResolvedValue(makeCostlyColoursState({ phase: 'mog' }));
      renderWithProviders(<CostlyColoursPage />);
      await waitFor(() => expect(mockExec).toHaveBeenCalled());
      expect(screen.queryByTestId('costlycolours-mog-discard')).not.toBeInTheDocument();
    });
  });
});
