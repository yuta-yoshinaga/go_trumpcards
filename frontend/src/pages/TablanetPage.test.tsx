import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { tablanetApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeTablanetState } from '../test/stateFactories';
import { TablanetPage } from './TablanetPage';

vi.mock('../api/gameApi', () => ({
  tablanetApi: { exec: vi.fn() },
  actionLogApi: { tablanet: vi.fn() },
}));

const mockExec = vi.mocked(tablanetApi.exec);

const playPhaseState = makeTablanetState();
const cpuTurnState = makeTablanetState({ currentTurn: 1, isHumanTurn: false });
const gameEndState = makeTablanetState({
  phase: 1,
  gameEndFlag: true,
  winners: [0],
  players: [
    { id: 0, isHuman: true, cardCount: 0, cards: [], capturedCount: 30, tablaCount: 2, score: 27 },
    { id: 1, isHuman: false, cardCount: 0, cards: [], capturedCount: 10, tablaCount: 0, score: 5 },
    { id: 2, isHuman: false, cardCount: 0, cards: [], capturedCount: 8, tablaCount: 0, score: 4 },
    { id: 3, isHuman: false, cardCount: 0, cards: [], capturedCount: 4, tablaCount: 0, score: 1 },
  ],
  lastDealDetail: {
    cards: { 0: 30, 1: 10, 2: 8, 3: 4 },
    aces: { 0: 2, 1: 1, 2: 1, 3: 0 },
    jacks: { 0: 2, 1: 1, 2: 1, 3: 0 },
    tablas: { 0: 2, 1: 0, 2: 0, 3: 0 },
    hasTenDiamonds: 0,
    hasTwoClubs: 0,
    mostCards: 0,
    gained: { 0: 27, 1: 5, 2: 4, 3: 1 },
  },
});

// A play-phase state where the human's non-Jack card 0 (value 5) can capture
// EVERY table card (both indices), so selecting it makes a tabla possible. Card 1
// is a Jack (value 11) whose sweep must NOT count as a tabla.
const tablaReadyState = makeTablanetState({
  captureOptions: { 0: [0, 1], 1: [0, 1] },
});

beforeEach(() => {
  localStorage.clear();
  mockExec.mockReset();
  mockExec.mockResolvedValue(playPhaseState);
});

describe('TablanetPage', () => {
  it('renders skeleton fallback when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<TablanetPage />);
    // The skeleton fallback marks its container aria-busy.
    expect(document.querySelector('[aria-busy]')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<TablanetPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('renders the play phase with table and hand cards', async () => {
    renderWithProviders(<TablanetPage />);
    await waitFor(() => {
      expect(screen.getByTestId('hand-card-0')).toBeInTheDocument();
      expect(screen.getByTestId('table-card-0')).toBeInTheDocument();
    });
  });

  it('capturing dispatches play with the selected hand and table indices', async () => {
    renderWithProviders(<TablanetPage />);
    const handCard = await screen.findByTestId('hand-card-0');
    fireEvent.click(handCard);
    fireEvent.click(screen.getByTestId('table-card-0'));
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '捕獲' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 0, tableIndices: [0] }));
  });

  it('announces a capture in the live region attributed to the capturer', async () => {
    renderWithProviders(<TablanetPage />); // mount: captured totals 0
    const handCard = await screen.findByTestId('hand-card-0');
    fireEvent.click(handCard);
    fireEvent.click(screen.getByTestId('table-card-0'));

    // Resolve the play into a state where the human (seat 0) has captured.
    const afterCapture = makeTablanetState({
      lastCaptureIdx: 0,
      players: playPhaseState.players.map((p, i) => (i === 0 ? { ...p, capturedCount: 2 } : p)),
    });
    mockExec.mockResolvedValue(afterCapture);
    fireEvent.click(screen.getByRole('button', { name: '捕獲' }));

    // The region remounts (key={announceNonce}) on each event, so query it fresh.
    await waitFor(() =>
      expect(screen.getByTestId('tablanet-live-region')).toHaveTextContent('あなた が場札を捕獲しました'),
    );
  });

  it('announces a tabla when a player clears the table', async () => {
    renderWithProviders(<TablanetPage />);
    const handCard = await screen.findByTestId('hand-card-0');
    fireEvent.click(handCard);
    fireEvent.click(screen.getByTestId('table-card-0'));

    const afterTabla = makeTablanetState({
      lastCaptureIdx: 0,
      players: playPhaseState.players.map((p, i) => (i === 0 ? { ...p, capturedCount: 2, tablaCount: 1 } : p)),
    });
    mockExec.mockResolvedValue(afterTabla);
    fireEvent.click(screen.getByRole('button', { name: '捕獲' }));

    await waitFor(() =>
      expect(screen.getByTestId('tablanet-live-region')).toHaveTextContent('あなた がタブラ（場を一掃）を達成しました'),
    );
  });

  it('attributes a tabla to the achiever, not merely the last capturer', async () => {
    renderWithProviders(<TablanetPage />);
    const handCard = await screen.findByTestId('hand-card-0');
    fireEvent.click(handCard);
    fireEvent.click(screen.getByTestId('table-card-0'));

    // One response bundles CPU auto-plays: CPU 1 (seat 1) scored the tabla while
    // CPU 2 (seat 2) was the LAST to capture. The tabla must name CPU 1.
    const bundled = makeTablanetState({
      lastCaptureIdx: 2,
      players: playPhaseState.players.map((p, i) => {
        if (i === 1) return { ...p, capturedCount: 3, tablaCount: 1 };
        if (i === 2) return { ...p, capturedCount: 2 };
        return p;
      }),
    });
    mockExec.mockResolvedValue(bundled);
    fireEvent.click(screen.getByRole('button', { name: '捕獲' }));

    await waitFor(() =>
      expect(screen.getByTestId('tablanet-live-region')).toHaveTextContent('CPU 1 がタブラ（場を一掃）を達成しました'),
    );
  });

  it('keeps the live region empty before any capture', async () => {
    renderWithProviders(<TablanetPage />);
    const region = await screen.findByTestId('tablanet-live-region');
    expect(region).toHaveTextContent('');
  });

  it('trailing (no table cards) dispatches play with an empty capture set', async () => {
    renderWithProviders(<TablanetPage />);
    const handCard = await screen.findByTestId('hand-card-1');
    fireEvent.click(handCard);
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '出す' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 1, tableIndices: [] }));
  });

  it('does not show the play button on a CPU turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<TablanetPage />);
    await waitFor(() => expect(screen.getByTestId('tablanet-prompt')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '出す' })).not.toBeInTheDocument();
  });

  it('renders the game-end result with final scores and the new-game button', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<TablanetPage />);
    await waitFor(() => expect(screen.getByTestId('tablanet-result')).toBeInTheDocument());
    expect(screen.getByRole('button', { name: '新しいゲーム' })).toBeInTheDocument();
  });

  it('clicking new game dispatches nextround', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<TablanetPage />);
    const nextBtn = await screen.findByRole('button', { name: '新しいゲーム' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(gameEndState);
    fireEvent.click(nextBtn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('highlights the table and switches the button label when a tabla is possible', async () => {
    mockExec.mockResolvedValue(tablaReadyState);
    renderWithProviders(<TablanetPage />);
    const handCard = await screen.findByTestId('hand-card-0');

    // Before selecting a card, no tabla emphasis.
    expect(screen.queryByTestId('tablanet-tabla-badge')).not.toBeInTheDocument();
    expect(screen.getByTestId('tablanet-play-button')).toHaveTextContent('出す');

    // Selecting the non-Jack sweep card reveals the tabla badge + button label.
    fireEvent.click(handCard);
    expect(screen.getByTestId('tablanet-tabla-badge')).toHaveTextContent('タブラ！');
    expect(screen.getByTestId('tablanet-play-button')).toHaveTextContent('タブラ捕獲！');
  });

  it('does not treat a Jack sweep as a tabla', async () => {
    mockExec.mockResolvedValue(tablaReadyState);
    renderWithProviders(<TablanetPage />);
    // Card 1 is a Jack: even though it can clear the table, it never scores a tabla.
    const jackCard = await screen.findByTestId('hand-card-1');
    fireEvent.click(jackCard);
    expect(screen.queryByTestId('tablanet-tabla-badge')).not.toBeInTheDocument();
    expect(screen.getByTestId('tablanet-play-button')).not.toHaveTextContent('タブラ捕獲！');
  });
});
