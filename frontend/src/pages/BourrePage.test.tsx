import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { bourreApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { BourrePlayerData, BourreResponse } from '../types/card';
import { BourrePage } from './BourrePage';

vi.mock('../api/gameApi', () => ({
  bourreApi: { exec: vi.fn() },
  actionLogApi: { bourre: vi.fn() },
}));

const mockExec = vi.mocked(bourreApi.exec);

function player(overrides: Partial<BourrePlayerData>): BourrePlayerData {
  return {
    id: 0,
    isHuman: false,
    isFinished: false,
    folded: false,
    decided: false,
    drawn: false,
    bourreed: false,
    chips: 100,
    tricks: 0,
    cardCount: 5,
    cards: [],
    ...overrides,
  };
}

function makeState(overrides: Partial<BourreResponse> = {}): BourreResponse {
  return {
    players: [
      player({ id: 0, isHuman: true }),
      player({ id: 1 }),
      player({ id: 2 }),
      player({ id: 3 }),
      player({ id: 4 }),
    ],
    phase: 'decide',
    currentPlayerIdx: 0,
    dealerIdx: 4,
    pot: 25,
    carryPot: 0,
    trumpSuit: 'SPADE',
    trumpCard: { design: 'SPADE', value: 5 },
    trickNumber: 0,
    currentTrick: [],
    lastTrick: [],
    lastTrickWinner: -1,
    leadPlayerIdx: -1,
    handNumber: 1,
    gameEndFlag: false,
    winnerIdx: -1,
    validPlays: [],
    results: [],
    config: { cpuDifficulty: 0 },
    message: '',
    ...overrides,
  } as BourreResponse;
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeState());
});

describe('BourrePage', () => {
  it('renders skeleton before first API response', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<BourrePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset command on mount', async () => {
    renderWithProviders(<BourrePage />);
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith(expect.objectContaining({ command: 'reset' }));
    });
  });

  it('header renders the trump suit as a symbol, not the raw design string', async () => {
    renderWithProviders(<BourrePage />); // default trumpSuit: SPADE
    const trump = await screen.findByTestId('bourre-trump');
    expect(trump).toHaveTextContent('♠');
    expect(trump).not.toHaveTextContent('SPADE');
  });

  it('header colors a red trump suit (hearts) with the error token', async () => {
    mockExec.mockResolvedValue(makeState({ trumpSuit: 'HEART' }));
    renderWithProviders(<BourrePage />);
    const trump = await screen.findByTestId('bourre-trump');
    expect(within(trump).getByText('♥')).toHaveClass('text-ds-error');
  });

  it('header colors a red trump suit (diamonds) with the error token', async () => {
    mockExec.mockResolvedValue(makeState({ trumpSuit: 'DIAMOND' }));
    renderWithProviders(<BourrePage />);
    const trump = await screen.findByTestId('bourre-trump');
    expect(within(trump).getByText('♦')).toHaveClass('text-ds-error');
  });

  it('header does not color a black trump suit (clubs) with the error token', async () => {
    mockExec.mockResolvedValue(makeState({ trumpSuit: 'CLOVER' }));
    renderWithProviders(<BourrePage />);
    const trump = await screen.findByTestId('bourre-trump');
    expect(within(trump).getByText('♣')).not.toHaveClass('text-ds-error');
  });

  it('header shows a dash when the trump suit is unset', async () => {
    mockExec.mockResolvedValue(makeState({ trumpSuit: 'JOKER' }));
    renderWithProviders(<BourrePage />);
    const trump = await screen.findByTestId('bourre-trump');
    expect(trump).toHaveTextContent('-');
    expect(trump).not.toHaveTextContent('JOKER');
  });

  it('renders CPU player areas after load', async () => {
    renderWithProviders(<BourrePage />);
    await waitFor(() => {
      expect(screen.getByText(/CPU 1/)).toBeInTheDocument();
    });
  });

  it('shows localized CPU status for finished, folded, and bourréd players', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [
          player({ id: 0, isHuman: true }),
          player({ id: 1, isFinished: true }),
          player({ id: 2, folded: true }),
          player({ id: 3, bourreed: true }),
        ],
      }),
    );
    renderWithProviders(<BourrePage />);
    // playerStatus resolves each state through i18n (ja).
    await waitFor(() => expect(screen.getByText('脱落')).toBeInTheDocument());
    expect(screen.getAllByText('フォールド').length).toBeGreaterThanOrEqual(1);
    // "ブーレ" is also the sr-only page title, so the bourréd status adds a second.
    expect(screen.getAllByText('ブーレ').length).toBeGreaterThanOrEqual(2);
  });

  it('names the winner via playerName when a CPU wins at game end', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 'roundEnd',
        gameEndFlag: true,
        winnerIdx: 2,
        results: [{ playerIdx: 2, tricks: 3, wonAmount: 0, bourreed: false, folded: false }],
      }),
    );
    renderWithProviders(<BourrePage />);
    // result.youLose interpolates the localized CPU winner name.
    await waitFor(() => expect(screen.getByText(/CPU 2 の勝ち/)).toBeInTheDocument());
  });

  it('decide phase: shows the pot/penalty summary, flagged in warning color when the pot is high', async () => {
    renderWithProviders(<BourrePage />); // default: pot 25, carryPot 0 → penalty 25 (>= threshold)
    const summary = await screen.findByTestId('bourre-decide-summary');
    expect(summary).toHaveTextContent('25');
    expect(summary).toHaveClass('text-ds-warning');
  });

  it('decide phase: pot/penalty summary uses muted color when the pot is small', async () => {
    mockExec.mockResolvedValue(makeState({ pot: 3, carryPot: 0 }));
    renderWithProviders(<BourrePage />);
    const summary = await screen.findByTestId('bourre-decide-summary');
    expect(summary).toHaveClass('text-ds-text-muted');
    expect(summary).not.toHaveClass('text-ds-warning');
  });

  it('decide phase: summary clarifies the penalty only applies when playing, via tooltip', async () => {
    renderWithProviders(<BourrePage />);
    const summary = await screen.findByTestId('bourre-decide-summary');
    // The visible text explains folding avoids the penalty.
    expect(summary).toHaveTextContent(/降りれば罰金なし/);
    expect(summary).toHaveAttribute('title');
  });

  it('decide phase: penalty includes carryPot', async () => {
    mockExec.mockResolvedValue(makeState({ pot: 25, carryPot: 5 }));
    renderWithProviders(<BourrePage />);
    const summary = await screen.findByTestId('bourre-decide-summary');
    expect(summary).toHaveTextContent('30'); // penalty = pot + carryPot
  });

  it('decide phase: carryPot can push a small pot over the warning threshold', async () => {
    mockExec.mockResolvedValue(makeState({ pot: 5, carryPot: 6 }));
    renderWithProviders(<BourrePage />);
    const summary = await screen.findByTestId('bourre-decide-summary');
    expect(summary).toHaveTextContent('11'); // 5 + 6 >= 10
    expect(summary).toHaveClass('text-ds-warning');
  });

  it('does not show the decide summary outside the decide phase', async () => {
    mockExec.mockResolvedValue(makeState({ phase: 'play', currentPlayerIdx: 0 }));
    renderWithProviders(<BourrePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toBeInTheDocument());
    expect(screen.queryByTestId('bourre-decide-summary')).not.toBeInTheDocument();
  });

  it('decide phase: play and fold buttons dispatch decide', async () => {
    renderWithProviders(<BourrePage />);
    fireEvent.click(await screen.findByRole('button', { name: '参加（アンティ）' }));
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith(expect.objectContaining({ command: 'decide', decide: true }));
    });
    fireEvent.click(screen.getByRole('button', { name: 'フォールド' }));
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith(expect.objectContaining({ command: 'decide', decide: false }));
    });
  });

  it('play phase: clicking a legal card dispatches play', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 'play',
        currentPlayerIdx: 0,
        validPlays: [0],
        players: [
          player({
            id: 0,
            isHuman: true,
            cards: [
              { design: 'SPADE', value: 9 },
              { design: 'CLOVER', value: 5 },
            ],
          }),
          player({ id: 1 }),
          player({ id: 2 }),
          player({ id: 3 }),
          player({ id: 4 }),
        ],
      }),
    );
    const { container } = renderWithProviders(<BourrePage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    const cardBtn = container.querySelector('[data-tutorial="bourre-hand"] button');
    expect(cardBtn).not.toBeNull();
    fireEvent.click(cardBtn as Element);
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith(expect.objectContaining({ command: 'p', cardIndex: 0 }));
    });
  });

  it('roundEnd phase: next hand button dispatches next', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 'roundEnd',
        results: [
          { playerIdx: 0, tricks: 3, wonAmount: 25, bourreed: false, folded: false },
          { playerIdx: 1, tricks: 0, wonAmount: 0, bourreed: true, folded: false },
        ],
      }),
    );
    renderWithProviders(<BourrePage />);
    fireEvent.click(await screen.findByRole('button', { name: '次のハンド' }));
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith(expect.objectContaining({ command: 'next' }));
    });
  });

  it('end phase: shows the win banner when the human wins', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 'gameEnd',
        gameEndFlag: true,
        winnerIdx: 0,
        results: [{ playerIdx: 0, tricks: 5, wonAmount: 50, bourreed: false, folded: false }],
      }),
    );
    renderWithProviders(<BourrePage />);
    await waitFor(() => {
      expect(screen.getByText(/あなたの勝ち/)).toBeInTheDocument();
    });
  });

  it('draw phase: selecting a card then discarding dispatches draw with indices', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 'draw',
        currentPlayerIdx: 0,
        players: [
          player({
            id: 0,
            isHuman: true,
            cards: [
              { design: 'SPADE', value: 9 },
              { design: 'CLOVER', value: 5 },
            ],
          }),
          player({ id: 1 }),
          player({ id: 2 }),
          player({ id: 3 }),
          player({ id: 4 }),
        ],
      }),
    );
    const { container } = renderWithProviders(<BourrePage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    const cardBtn = container.querySelector('[data-tutorial="bourre-hand"] button');
    fireEvent.click(cardBtn as Element);
    fireEvent.click(screen.getByRole('button', { name: /交換する/ }));
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith(expect.objectContaining({ command: 'draw', indices: [0] }));
    });
  });

  it('draw phase: keep all dispatches an empty draw', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 'draw',
        currentPlayerIdx: 0,
        players: [
          player({ id: 0, isHuman: true, cards: [{ design: 'SPADE', value: 9 }] }),
          player({ id: 1 }),
          player({ id: 2 }),
          player({ id: 3 }),
          player({ id: 4 }),
        ],
      }),
    );
    renderWithProviders(<BourrePage />);
    fireEvent.click(await screen.findByRole('button', { name: '交換しない' }));
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith(expect.objectContaining({ command: 'draw', indices: [] }));
    });
  });

  it('roundEnd phase shows results and the next hand button', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 'roundEnd',
        results: [
          { playerIdx: 0, tricks: 0, wonAmount: 0, bourreed: true, folded: false },
          { playerIdx: 1, tricks: 3, wonAmount: 25, bourreed: false, folded: false },
        ],
      }),
    );
    renderWithProviders(<BourrePage />);
    fireEvent.click(await screen.findByRole('button', { name: '次のハンド' }));
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith(expect.objectContaining({ command: 'next' }));
    });
  });

  it('end phase: shows the lose message when a CPU wins', async () => {
    mockExec.mockResolvedValue(
      makeState({
        phase: 'gameEnd',
        gameEndFlag: true,
        winnerIdx: 1,
        results: [{ playerIdx: 1, tricks: 5, wonAmount: 50, bourreed: false, folded: false }],
      }),
    );
    renderWithProviders(<BourrePage />);
    await waitFor(() => {
      expect(screen.getByText(/CPU 1 の勝ち/)).toBeInTheDocument();
    });
  });

  it('settings: changing CPU difficulty resets the game with the config', async () => {
    renderWithProviders(<BourrePage />);
    const select = (await screen.findByTestId('bourre-difficulty')) as HTMLSelectElement;
    expect(select.value).toBe('0');
    fireEvent.change(select, { target: { value: '2' } });
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith(
        expect.objectContaining({ command: 'reset', config: { cpuDifficulty: 2 } }),
      );
    });
    expect(select.value).toBe('2');
  });

  it('settings: the next-game button reuses the selected CPU difficulty', async () => {
    // At game end the footer button fires reset directly (no confirm dialog).
    mockExec.mockResolvedValue(
      makeState({
        phase: 'gameEnd',
        gameEndFlag: true,
        winnerIdx: 0,
        results: [{ playerIdx: 0, tricks: 5, wonAmount: 50, bourreed: false, folded: false }],
      }),
    );
    renderWithProviders(<BourrePage />);
    const select = (await screen.findByTestId('bourre-difficulty')) as HTMLSelectElement;
    fireEvent.change(select, { target: { value: '1' } });
    await waitFor(() => expect(select.value).toBe('1'));
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '次のゲーム' }));
    await waitFor(() => {
      expect(mockExec).toHaveBeenCalledWith(
        expect.objectContaining({ command: 'reset', config: { cpuDifficulty: 1 } }),
      );
    });
  });

  it('toggles CLI mode', async () => {
    renderWithProviders(<BourrePage />);
    await waitFor(() => expect(screen.getByText(/CPU 1/)).toBeInTheDocument());
    const cliToggle = screen.getByRole('button', { name: /CLI/i });
    fireEvent.click(cliToggle);
    await waitFor(() => expect(screen.queryByText(/CPU 1/)).not.toBeInTheDocument());
  });

  it('parses and dispatches CLI commands, and formats state output', async () => {
    localStorage.setItem('cli-mode-bourre', 'true');
    mockExec.mockResolvedValue(
      makeState({
        phase: 'play',
        currentTrick: [{ playerIdx: 1, card: { design: 'HEART', value: 7 } }],
        message: 'your turn',
      }),
    );
    renderWithProviders(<BourrePage />);
    const input = await screen.findByRole('textbox');

    const run = async (cmd: string, expected: Record<string, unknown>) => {
      fireEvent.change(input, { target: { value: cmd } });
      fireEvent.keyDown(input, { key: 'Enter' });
      await waitFor(() => {
        expect(mockExec).toHaveBeenCalledWith(expect.objectContaining(expected));
      });
    };

    await run('d 1', { command: 'decide', decide: true });
    await run('d 0', { command: 'decide', decide: false });
    await run('dr 0 2', { command: 'draw', indices: [0, 2] });
    await run('p 3', { command: 'p', cardIndex: 3 });
    await run('n', { command: 'next' });
    await run('r', { command: 'reset' });
    await run('sd 1', { command: 'reset', config: { cpuDifficulty: 1 } });
    await run('log', { command: 'l' });

    // Unknown command surfaces an error and does not dispatch.
    mockExec.mockClear();
    fireEvent.change(input, { target: { value: 'xyz' } });
    fireEvent.keyDown(input, { key: 'Enter' });
    await waitFor(() => {
      expect(screen.getAllByText(/Unknown command/).length).toBeGreaterThan(0);
    });
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('CLI formatter renders the trump suit as a symbol, not the raw design string', async () => {
    localStorage.setItem('cli-mode-bourre', 'true');
    // Distinct objects so useCliGame's [state] effect fires after the command re-renders.
    mockExec.mockResolvedValueOnce(makeState({ phase: 'decide', trumpSuit: 'HEART' })); // mount reset
    mockExec.mockResolvedValue(makeState({ phase: 'play', trumpSuit: 'HEART' })); // after command
    renderWithProviders(<BourrePage />);
    const input = await screen.findByRole('textbox');
    fireEvent.change(input, { target: { value: 'n' } });
    fireEvent.keyDown(input, { key: 'Enter' });
    const output = await screen.findByText(/Trump: ♥/);
    expect(output).toBeInTheDocument();
    expect(output).not.toHaveTextContent('HEART');
  });
});
