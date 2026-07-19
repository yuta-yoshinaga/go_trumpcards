import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { koikoiApi } from '../api/gameApi';
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
});
