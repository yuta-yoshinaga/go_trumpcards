import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { spoonsApi } from '../api/gameApi';
import { renderWithProviders } from '../test/renderWithProviders';
import type { SpoonsPlayer, SpoonsResponse } from '../types/card';
import { SpoonsPage } from './SpoonsPage';

vi.mock('../api/gameApi', () => ({
  spoonsApi: { exec: vi.fn() },
  actionLogApi: { spoons: vi.fn() },
}));

const mockPlaySound = vi.fn();
const mockSoundValue = { playSound: mockPlaySound, muted: false, toggleMute: vi.fn() };
vi.mock('../providers/SoundProvider', () => ({
  SoundProvider: ({ children }: { children: React.ReactNode }) => children,
  useSound: () => mockSoundValue,
  useOptionalSound: () => mockSoundValue,
}));

const mockExec = vi.mocked(spoonsApi.exec);

function makePlayer(overrides: Partial<SpoonsPlayer> = {}): SpoonsPlayer {
  return {
    name: 'CPU',
    isHuman: false,
    handSize: 4,
    hand: [],
    letters: 0,
    eliminated: false,
    hasSpoon: false,
    ...overrides,
  };
}

function makeState(overrides: Partial<SpoonsResponse> = {}): SpoonsResponse {
  return {
    phase: 0,
    gameEndFlag: false,
    winnerIdx: -1,
    currentPlayerIdx: 0,
    feederIdx: 0,
    isHumanTurn: true,
    spoonsRemaining: 3,
    grabWindowOpen: false,
    firstGrabberIdx: -1,
    roundLoserIdx: -1,
    roundNumber: 1,
    drawPileSize: 36,
    cpuDifficulty: 1,
    message: '',
    players: [
      makePlayer({
        name: 'You',
        isHuman: true,
        hand: [
          { design: 'SPADE', value: 1 },
          { design: 'HEART', value: 2 },
          { design: 'CLOVER', value: 3 },
          { design: 'DIAMOND', value: 4 },
        ],
      }),
      makePlayer(),
      makePlayer(),
      makePlayer(),
    ],
    ...overrides,
  };
}

const passState = makeState();
const grabState = makeState({ phase: 1, grabWindowOpen: true });
const roundEndState = makeState({ phase: 2, roundLoserIdx: 1, isHumanTurn: false });
const gameEndState = makeState({ phase: 3, gameEndFlag: true, winnerIdx: 0, isHumanTurn: false });

beforeEach(() => {
  mockExec.mockReset();
  mockPlaySound.mockReset();
  mockExec.mockResolvedValue(passState);
});

describe('SpoonsPage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<SpoonsPage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<SpoonsPage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('shows round, spoons remaining and draw pile info', async () => {
    renderWithProviders(<SpoonsPage />);
    await waitFor(() => expect(screen.getByText(/ラウンド 1/)).toBeInTheDocument());
    expect(screen.getByText(/残りスプーン: 3/)).toBeInTheDocument();
    expect(screen.getByText(/山札: 36/)).toBeInTheDocument();
  });

  it('renders the human hand as pass buttons naming each card', async () => {
    renderWithProviders(<SpoonsPage />);
    // Each pass button names the card it would hand over (e.g. ♠A → "♠ A を渡す").
    await waitFor(() => expect(screen.getAllByRole('button', { name: /を渡す$/ }).length).toBe(4));
    expect(screen.getByRole('button', { name: '♠ A を渡す' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '♣ 3 を渡す' })).toBeInTheDocument();
  });

  it('dispatches pass with the clicked card index', async () => {
    renderWithProviders(<SpoonsPage />);
    const clover = await screen.findByRole('button', { name: '♣ 3 を渡す' }); // hand index 2
    mockExec.mockClear();
    fireEvent.click(clover);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass', { cardIndex: 2 }));
  });

  it('color-codes same-rank hand cards into groups and leaves singletons neutral', async () => {
    // Hand: 7♠ 7♥ 3♣ K♦ — the two 7s share a group color; 3 and K are neutral.
    mockExec.mockResolvedValue(
      makeState({
        players: [
          makePlayer({
            name: 'You',
            isHuman: true,
            hand: [
              { design: 'SPADE', value: 7 },
              { design: 'HEART', value: 7 },
              { design: 'CLOVER', value: 3 },
              { design: 'DIAMOND', value: 13 },
            ],
          }),
          makePlayer(),
          makePlayer(),
          makePlayer(),
        ],
      }),
    );
    renderWithProviders(<SpoonsPage />);
    const c0 = await screen.findByTestId('spoons-pass-0');
    const c1 = screen.getByTestId('spoons-pass-1');
    const c2 = screen.getByTestId('spoons-pass-2');
    const c3 = screen.getByTestId('spoons-pass-3');
    // The pair shares a non-"none" group color.
    expect(c0).toHaveAttribute('data-rank-group', c1.getAttribute('data-rank-group') ?? '');
    expect(c0.getAttribute('data-rank-group')).not.toBe('none');
    // Singletons carry no group color and are not reach.
    expect(c2).toHaveAttribute('data-rank-group', 'none');
    expect(c3).toHaveAttribute('data-rank-group', 'none');
    expect(c0).toHaveAttribute('data-rank-reach', 'false');
  });

  it('flags a three-of-a-kind hand as a reach', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [
          makePlayer({
            name: 'You',
            isHuman: true,
            hand: [
              { design: 'SPADE', value: 8 },
              { design: 'HEART', value: 8 },
              { design: 'CLOVER', value: 8 },
              { design: 'DIAMOND', value: 2 },
            ],
          }),
          makePlayer(),
          makePlayer(),
          makePlayer(),
        ],
      }),
    );
    renderWithProviders(<SpoonsPage />);
    const c0 = await screen.findByTestId('spoons-pass-0');
    expect(c0).toHaveAttribute('data-rank-reach', 'true');
    expect(c0.getAttribute('data-rank-group')).not.toBe('none');
    expect(screen.getByTestId('spoons-pass-3')).toHaveAttribute('data-rank-reach', 'false');
  });

  it('does not render pass buttons when it is not the human turn', async () => {
    mockExec.mockResolvedValue(makeState({ currentPlayerIdx: 1, isHumanTurn: false }));
    renderWithProviders(<SpoonsPage />);
    await waitFor(() => expect(screen.getByText(/プレイヤー/)).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: /を渡す$/ })).not.toBeInTheDocument();
  });

  it('announces the grab window opening via role=alert', async () => {
    mockExec.mockResolvedValue(grabState);
    renderWithProviders(<SpoonsPage />);
    const notice = await screen.findByTestId('spoons-grab-notice');
    expect(notice).toHaveAttribute('role', 'alert');
  });

  it('shows the grab button when the grab window is open and dispatches grab', async () => {
    mockExec.mockResolvedValue(grabState);
    renderWithProviders(<SpoonsPage />);
    const btn = await screen.findByRole('button', { name: 'スプーンを取る！' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('grab'));
  });

  it('pulses the grab button while the window is open', async () => {
    mockExec.mockResolvedValue(grabState);
    renderWithProviders(<SpoonsPage />);
    const btn = await screen.findByTestId('spoons-grab-button');
    expect(btn.className).toContain('motion-safe:animate-pulse');
  });

  it('chimes once when the grab window opens (false→true)', async () => {
    // Mount in the pass phase (window closed) → no chime yet.
    renderWithProviders(<SpoonsPage />);
    const buttons = await screen.findAllByRole('button', { name: /を渡す$/ });
    expect(mockPlaySound).not.toHaveBeenCalled();
    // Passing a card opens the grab window in the response → chime fires once.
    mockExec.mockResolvedValueOnce(grabState);
    fireEvent.click(buttons[0]);
    await waitFor(() => expect(mockPlaySound).toHaveBeenCalledWith('turnTick', { pitchVariation: 0.1 }));
    expect(mockPlaySound).toHaveBeenCalledTimes(1);
  });

  it('hides the grab button when the grab window is closed', async () => {
    renderWithProviders(<SpoonsPage />);
    await waitFor(() => expect(screen.getByText(/プレイヤー/)).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: 'スプーンを取る！' })).not.toBeInTheDocument();
  });

  it('does not re-chime while the grab window stays open across updates', async () => {
    renderWithProviders(<SpoonsPage />);
    const buttons = await screen.findAllByRole('button', { name: /を渡す$/ });
    mockExec.mockResolvedValueOnce(grabState);
    fireEvent.click(buttons[0]);
    await waitFor(() => expect(mockPlaySound).toHaveBeenCalledTimes(1));
    // A later update that keeps the window open must not chime again.
    mockExec.mockResolvedValueOnce(makeState({ phase: 1, grabWindowOpen: true, drawPileSize: 35 }));
    fireEvent.click(screen.getByRole('button', { name: 'スプーンを取る！' }));
    await waitFor(() => expect(screen.getByText(/山札: 35/)).toBeInTheDocument());
    expect(mockPlaySound).toHaveBeenCalledTimes(1);
  });

  it('shows the next round button at round end and dispatches next', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<SpoonsPage />);
    const btn = await screen.findByRole('button', { name: '次のラウンドへ' });
    mockExec.mockClear();
    fireEvent.click(btn);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('shows the round loser at round end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<SpoonsPage />);
    await waitFor(() => expect(screen.getByText(/ラウンド結果/)).toBeInTheDocument());
  });

  it('shows the win message on game end when the human wins', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<SpoonsPage />);
    await waitFor(() => expect(screen.getByText('あなたの勝利です！')).toBeInTheDocument());
  });

  it('renders player letters and eliminated badge', async () => {
    mockExec.mockResolvedValue(
      makeState({
        players: [
          makePlayer({ name: 'You', isHuman: true, hand: [{ design: 'SPADE', value: 1 }] }),
          makePlayer({ eliminated: true, letters: 6 }),
          makePlayer({ hasSpoon: true }),
          makePlayer(),
        ],
      }),
    );
    renderWithProviders(<SpoonsPage />);
    await waitFor(() => expect(screen.getByText(/脱落/)).toBeInTheDocument());
    expect(screen.getByText(/スプーン獲得/)).toBeInTheDocument();
  });

  it('changes CPU difficulty via the settings panel and resets', async () => {
    renderWithProviders(<SpoonsPage />);
    await waitFor(() => expect(screen.getByText(/プレイヤー/)).toBeInTheDocument());
    mockExec.mockClear();
    const select = screen.getByLabelText('CPU難易度');
    fireEvent.change(select, { target: { value: '2' } });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: { cpuDifficulty: 2 } }));
  });

  it('toggles the CLI terminal', async () => {
    renderWithProviders(<SpoonsPage />);
    await waitFor(() => expect(screen.getByText(/プレイヤー/)).toBeInTheDocument());
    const toggle = screen.getByRole('button', { name: /CLI/i });
    fireEvent.click(toggle);
    await waitFor(() => expect(screen.getByRole('textbox')).toBeInTheDocument());
  });
});
