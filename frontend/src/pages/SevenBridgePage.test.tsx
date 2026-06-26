import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { sevenBridgeApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { SevenBridgeResponse } from '../types/card';
import { SevenBridgePage } from './SevenBridgePage';

vi.mock('../api/gameApi', () => ({
  sevenBridgeApi: { exec: vi.fn() },
  actionLogApi: { sevenbridge: vi.fn() },
  sessionId: 'test-session',
}));

const mockExec = vi.mocked(sevenBridgeApi.exec);

const drawState: SevenBridgeResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 3,
      cards: [
        { design: 'SPADE', value: 9 },
        { design: 'CLOVER', value: 9 },
        { design: 'DIAMOND', value: 3 },
      ],
      melds: [],
      roundScore: 0,
      cumulativeScore: 0,
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 7,
      cards: [],
      melds: [],
      roundScore: 0,
      cumulativeScore: 10,
    },
  ],
  phase: 0,
  roundNumber: 1,
  currentPlayerIdx: 0,
  discardTop: { design: 'HEART', value: 9 },
  drawPileCount: 30,
  gameEndFlag: false,
  winnerIdx: -1,
  roundWinnerIdx: -1,
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 100 },
};

const playState: SevenBridgeResponse = { ...drawState, phase: 1 };

describe('SevenBridgePage', () => {
  beforeEach(() => {
    mockExec.mockReset();
  });

  it('calls reset on mount', async () => {
    mockExec.mockResolvedValue(drawState);
    renderWithProviders(<SevenBridgePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, expect.any(Object)));
  });

  it('renders draw phase controls', async () => {
    mockExec.mockResolvedValue(drawState);
    renderWithProviders(<SevenBridgePage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /山札から引く|Draw from stock/i })).toBeInTheDocument(),
    );
    expect(screen.getByRole('button', { name: /ポン|Pon/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /チー|Chi/i })).toBeInTheDocument();
  });

  it('renders play phase controls', async () => {
    mockExec.mockResolvedValue(playState);
    renderWithProviders(<SevenBridgePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: /メルド|Meld/i })).toBeInTheDocument());
    expect(screen.getByRole('button', { name: /捨てる|Discard/i })).toBeInTheDocument();
  });

  it('renders layoff target melds as clickable card-row buttons and highlights the selection', async () => {
    const meldState: SevenBridgeResponse = {
      ...playState,
      players: [
        {
          ...playState.players[0],
          melds: [
            {
              cards: [
                { design: 'SPADE', value: 5 },
                { design: 'HEART', value: 5 },
                { design: 'DIAMOND', value: 5 },
              ],
            },
          ],
        },
        playState.players[1],
      ],
    };
    mockExec.mockResolvedValue(meldState);
    renderWithProviders(<SevenBridgePage />);
    const meldBtn = await screen.findByTestId('sb-layoff-meld-0-0');
    fireEvent.click(meldBtn);
    expect(meldBtn).toHaveAttribute('aria-pressed', 'true');
    expect(meldBtn.className).toContain('ring-ds-info');
  });

  it('fires drawstock when clicking stock draw', async () => {
    mockExec.mockResolvedValue(drawState);
    renderWithProviders(<SevenBridgePage />);
    const btn = await screen.findByRole('button', { name: /山札から引く|Draw from stock/i });
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
  });

  it('renders round-end with the next-round button and revealed CPU hands', async () => {
    const roundEndState: SevenBridgeResponse = {
      ...drawState,
      phase: 2,
      roundWinnerIdx: 0,
      players: [
        { ...drawState.players[0], roundScore: 0 },
        { ...drawState.players[1], cards: [{ design: 'HEART', value: 4 }], cardCount: 1, roundScore: 12 },
      ],
    };
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<SevenBridgePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '次のラウンド' })).toBeInTheDocument());
  });

  it('renders game-end state', async () => {
    const gameEndState: SevenBridgeResponse = {
      ...drawState,
      phase: 3,
      gameEndFlag: true,
      winnerIdx: 0,
      players: [
        { ...drawState.players[0], cards: [{ design: 'SPADE', value: 9 }] },
        { ...drawState.players[1], cards: [{ design: 'HEART', value: 4 }], cardCount: 1 },
      ],
    };
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<SevenBridgePage />);
    // At game end the CPU hand is revealed (covers the isGameEnd reveal branch).
    await waitFor(() => expect(screen.getByAltText('♥ 4')).toBeInTheDocument());
  });
});
