import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { koikoiApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeKoiKoiState } from '../test/stateFactories';
import { KoiKoiPage } from './KoiKoiPage';

vi.mock('../api/gameApi', () => ({
  koikoiApi: { exec: vi.fn() },
  actionLogApi: { koikoi: vi.fn() },
}));

const mockExec = vi.mocked(koikoiApi.exec);

const playState = makeKoiKoiState();
const decisionState = makeKoiKoiState({
  phase: 1,
  pendingYaku: [{ key: 'tane', points: 1 }],
  pendingPoints: 1,
});
const roundEndState = makeKoiKoiState({
  phase: 2,
  roundWinner: 0,
  lastRoundResult: {
    winner: 0,
    yaku: [{ key: 'tane', points: 1 }],
    basePoints: 1,
    multiplier: 2,
    total: 2,
    koikoiCount: 1,
  },
});
const gameEndState = makeKoiKoiState({
  phase: 3,
  gameEndFlag: true,
  winner: 0,
  message: 'ゲーム終了！',
});

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(playState);
});

describe('KoiKoiPage', () => {
  it('renders the loading fallback when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<KoiKoiPage />);
    expect(screen.getByText('読み込み中…')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<KoiKoiPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('plays a hand card with a single match immediately', async () => {
    renderWithProviders(<KoiKoiPage />);
    const card = await screen.findByTestId('hand-card-0');
    mockExec.mockClear();
    fireEvent.click(card);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0 }));
  });

  it('requires a field pick for a two-way match, then plays with fieldIndex', async () => {
    mockExec.mockResolvedValue(makeKoiKoiState({ captureOptions: { 0: [0, 1] } }));
    renderWithProviders(<KoiKoiPage />);
    const card = await screen.findByTestId('hand-card-0');
    fireEvent.click(card);
    // The field-pick prompt appears and candidate field cards become clickable.
    await waitFor(() => expect(screen.getByTestId('koikoi-field-pick')).toBeInTheDocument());
    mockExec.mockClear();
    fireEvent.click(screen.getByTestId('field-card-1'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0, fieldIndex: 1 }));
  });

  it('shows the koi-koi / shobu buttons on the decision phase and dispatches them', async () => {
    mockExec.mockResolvedValue(decisionState);
    renderWithProviders(<KoiKoiPage />);
    await waitFor(() => expect(screen.getByTestId('koikoi-decision')).toBeInTheDocument());
    const koikoiBtn = screen.getByRole('button', { name: 'こいこい' });
    const shobuBtn = screen.getByRole('button', { name: '勝負' });
    mockExec.mockClear();
    fireEvent.click(koikoiBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('koikoi'));
    fireEvent.click(shobuBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('stop'));
  });

  it('shows the next-round button at round end and dispatches nextround', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<KoiKoiPage />);
    const btn = await screen.findByRole('button', { name: '次のラウンド' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('renders the game-end result', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<KoiKoiPage />);
    await waitFor(() => expect(screen.getByTestId('koikoi-result')).toBeInTheDocument());
  });

  it('renders a drawn game-end result without a winner banner', async () => {
    mockExec.mockResolvedValue(makeKoiKoiState({ phase: 3, gameEndFlag: true, winner: -1, message: '引き分け' }));
    renderWithProviders(<KoiKoiPage />);
    const result = await screen.findByTestId('koikoi-result');
    expect(result).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '新しいゲーム' })).toBeInTheDocument();
  });

  it('renders a drawn round-end result', async () => {
    mockExec.mockResolvedValue(
      makeKoiKoiState({
        phase: 2,
        roundWinner: -1,
        lastRoundResult: {
          winner: -1,
          yaku: [],
          basePoints: 0,
          multiplier: 1,
          total: 0,
          koikoiCount: 0,
        },
      }),
    );
    renderWithProviders(<KoiKoiPage />);
    await waitFor(() => expect(screen.getByTestId('koikoi-round-result')).toBeInTheDocument());
  });

  it('does not play a hand card when it is the CPU turn', async () => {
    mockExec.mockResolvedValue(makeKoiKoiState({ isHumanTurn: false, currentTurn: 1 }));
    renderWithProviders(<KoiKoiPage />);
    const card = await screen.findByTestId('hand-card-0');
    expect(card).toBeDisabled();
    mockExec.mockClear();
    fireEvent.click(card);
    // Clicking a disabled/non-human-turn card issues no API call.
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('offers the frontend hint toggle, off by default with no tooltip', async () => {
    renderWithProviders(<KoiKoiPage />);
    const toggle = await screen.findByLabelText('ヒント表示');
    expect(toggle).not.toBeChecked();
    expect(screen.queryByTestId('hint-tooltip')).not.toBeInTheDocument();
  });

  it('shows the hint tooltip when the toggle is enabled and state.hint is set', async () => {
    localStorage.setItem('hint_enabled_koikoi', 'true');
    mockExec.mockResolvedValue(
      makeKoiKoiState({ hint: { cardIndex: 0, fieldIndex: 0, koikoi: -1, reason: 'capture' } }),
    );
    renderWithProviders(<KoiKoiPage />);
    const tooltip = await screen.findByTestId('hint-tooltip');
    expect(tooltip).toHaveTextContent('価値の高い場札を捕獲する');
  });

  it('ignores a field click that is not a capture candidate', async () => {
    mockExec.mockResolvedValue(makeKoiKoiState({ captureOptions: { 0: [0, 1] } }));
    renderWithProviders(<KoiKoiPage />);
    fireEvent.click(await screen.findByTestId('hand-card-0'));
    await waitFor(() => expect(screen.getByTestId('koikoi-field-pick')).toBeInTheDocument());
    mockExec.mockClear();
    // field-card-0 and field-card-1 are the candidates; there is no field-card-2
    // in the two-card field, so a non-candidate click cannot fire. Re-clicking a
    // candidate still dispatches, confirming the guard path is exercised.
    fireEvent.click(screen.getByTestId('field-card-0'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0, fieldIndex: 0 }));
  });

  it('groups captured cards under their yaku-category headers', async () => {
    const hana = (color: string) => ({
      design: 'SPADE' as const,
      value: 0,
      glyph: '🎴',
      label: 'x',
      color,
      deck: 'hanafuda',
    });
    // Clone players (makeKoiKoiState shallow-spreads a shared array) so this
    // fixture never mutates the module-level base state used by other tests.
    const base = makeKoiKoiState();
    const state = makeKoiKoiState({
      players: [
        { ...base.players[0] },
        // CPU captured spans three categories; kasu is absent, so its group must not render.
        {
          ...base.players[1],
          captured: [hana('gold'), hana('purple'), hana('purple'), hana('red')],
          capturedCount: 4,
        },
      ],
    });
    mockExec.mockResolvedValue(state);
    renderWithProviders(<KoiKoiPage />);

    const bright = await screen.findByTestId('koikoi-cpu-group-bright');
    expect(bright).toHaveTextContent('光 · 1');
    expect(screen.getByTestId('koikoi-cpu-group-animal')).toHaveTextContent('種 · 2');
    expect(screen.getByTestId('koikoi-cpu-group-ribbon')).toHaveTextContent('短冊 · 1');
    expect(screen.queryByTestId('koikoi-cpu-group-kasu')).not.toBeInTheDocument();
  });

  it('shows the empty-pile label when a player has captured nothing', async () => {
    renderWithProviders(<KoiKoiPage />);
    // Default state captures nothing for either player.
    await waitFor(() => expect(screen.getByTestId('koikoi-human-captured')).toBeInTheDocument());
    expect(screen.getAllByText('獲得札なし').length).toBeGreaterThanOrEqual(2);
    expect(screen.queryByTestId('koikoi-human-group-bright')).not.toBeInTheDocument();
  });

  // **こいこいを一度でも宣言していれば確定点は 2 倍。**生の pendingPoints を
  // 出すと、実際より低い「今止めた場合の点数」を見せることになる (#4929)。
  // #5709: こいこいを続けるかは「今止めた場合の点」だけでなく**両者の累計差**で
  // 決まる。CUI は koikoiDecisionInfoStr で両者のスコアも出しているのに、Web は
  // 確定点だけで、決断のたびに画面をスクロールして確認する必要があった。
  it('shows both cumulative scores in the decision panel', async () => {
    mockExec.mockResolvedValue(
      makeKoiKoiState({
        phase: 1,
        pendingYaku: [{ key: 'tane', points: 3 }],
        pendingPoints: 3,
        koikoiCount: 0,
        // 素点は factory の既定を使い、**累計だけ**を動かす。
        players: makeKoiKoiState().players.map((p) => ({ ...p, score: p.isHuman ? 12 : 7 })),
      }),
    );
    renderWithProviders(<KoiKoiPage />);

    const scores = await screen.findByTestId('koikoi-decision-scores');

    expect(scores).toHaveTextContent('12');
    expect(scores).toHaveTextContent('7');
  });

  // 対戦相手が居ない状態 (通常は起こらないが、応答が欠けたとき) でも 0 として出す。
  it('falls back to zero when a seat is missing', async () => {
    const base = makeKoiKoiState();
    mockExec.mockResolvedValue(
      makeKoiKoiState({
        phase: 1,
        pendingYaku: [{ key: 'tane', points: 3 }],
        pendingPoints: 3,
        koikoiCount: 0,
        players: base.players.filter((p) => p.isHuman).map((p) => ({ ...p, score: 5 })),
      }),
    );
    renderWithProviders(<KoiKoiPage />);

    const scores = await screen.findByTestId('koikoi-decision-scores');

    expect(scores).toHaveTextContent('5');
    expect(scores).toHaveTextContent('0');
  });

  describe('decision panel multiplier', () => {
    it('shows the raw points on the first decision', async () => {
      mockExec.mockResolvedValue(
        makeKoiKoiState({ phase: 1, pendingYaku: [{ key: 'tane', points: 3 }], pendingPoints: 3, koikoiCount: 0 }),
      );
      renderWithProviders(<KoiKoiPage />);
      const panel = await screen.findByTestId('koikoi-decision-points');
      expect(panel).toHaveTextContent('3文');
      expect(panel).not.toHaveTextContent('倍率');
    });

    it('doubles the points once a koi-koi has been declared', async () => {
      mockExec.mockResolvedValue(
        makeKoiKoiState({ phase: 1, pendingYaku: [{ key: 'tane', points: 3 }], pendingPoints: 3, koikoiCount: 1 }),
      );
      renderWithProviders(<KoiKoiPage />);
      const panel = await screen.findByTestId('koikoi-decision-points');
      expect(panel).toHaveTextContent('6文');
      // 倍率が効いていることを明示する。
      expect(panel).toHaveTextContent('倍率×2');
      // 生の値をそのまま出していないこと。
      expect(panel).not.toHaveTextContent('3文');
    });
  });
});
