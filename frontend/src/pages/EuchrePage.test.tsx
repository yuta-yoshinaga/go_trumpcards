import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { actionLogApi, euchreApi } from '../api/gameApi';
import { NETWORK_ERROR_MESSAGE } from '../constants/messages';
import { renderWithProviders } from '../test/renderWithProviders';
import type { EuchreResponse } from '../types/card';
import { EuchrePage } from './EuchrePage';

vi.mock('../api/gameApi', () => ({
  euchreApi: { exec: vi.fn() },
  actionLogApi: { euchre: vi.fn() },
}));

const mockExec = vi.mocked(euchreApi.exec);

const playPhaseState: EuchreResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 5,
      cards: [
        { design: 'SPADE', value: 1 },
        { design: 'HEART', value: 11 },
      ],
      team: 0,
      trickCount: 0,
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 5,
      cards: [],
      team: 1,
      trickCount: 1,
    },
    {
      id: 2,
      isHuman: false,
      cardCount: 5,
      cards: [],
      team: 0,
      trickCount: 0,
    },
    {
      id: 3,
      isHuman: false,
      cardCount: 5,
      cards: [],
      team: 1,
      trickCount: 0,
    },
  ],
  phase: 3,
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  bidPlayerIdx: 0,
  dealerIdx: 3,
  trumpSuit: 1,
  faceUpCard: null,
  makerTeam: 0,
  goingAlone: false,
  goingAlonePlayerIdx: -1,
  currentTrick: [],
  teamScores: [0, 0],
  gameEndFlag: false,
  winnerTeam: -1,
  leadPlayerIdx: 0,
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 10 },
};

const pickUpPhaseState: EuchreResponse = {
  ...playPhaseState,
  phase: 0,
  bidPlayerIdx: 0,
  faceUpCard: { design: 'HEART', value: 12 },
  trumpSuit: 0,
  makerTeam: -1,
};

const pickUpPhaseCpuTurnState: EuchreResponse = {
  ...pickUpPhaseState,
  bidPlayerIdx: 1,
};

const callTrumpPhaseState: EuchreResponse = {
  ...playPhaseState,
  phase: 1,
  bidPlayerIdx: 0,
  faceUpCard: { design: 'HEART', value: 12 },
  trumpSuit: 0,
  makerTeam: -1,
};

const discardPhaseState: EuchreResponse = {
  ...playPhaseState,
  phase: 2,
  dealerIdx: 0,
  players: playPhaseState.players.map((p, i) =>
    i === 0
      ? {
          ...p,
          cardCount: 6,
          cards: [
            { design: 'SPADE', value: 1 },
            { design: 'HEART', value: 11 },
            { design: 'HEART', value: 12 },
          ],
        }
      : p,
  ),
};

const trickEndState: EuchreResponse = {
  ...playPhaseState,
  phase: 4,
  currentTrick: [
    { playerIdx: 0, card: { design: 'DIAMOND', value: 3 } },
    { playerIdx: 1, card: { design: 'HEART', value: 5 } },
  ],
};

const roundEndState: EuchreResponse = {
  ...playPhaseState,
  phase: 5,
};

const gameEndState: EuchreResponse = {
  ...playPhaseState,
  phase: 6,
  gameEndFlag: true,
  winnerTeam: 0,
  message: 'Game end!',
};

const gameEndByFlagState: EuchreResponse = {
  ...playPhaseState,
  phase: 3,
  gameEndFlag: true,
  winnerTeam: 0,
  message: 'Game end!',
};

const goingAloneState: EuchreResponse = {
  ...playPhaseState,
  goingAlone: true,
  goingAlonePlayerIdx: 0,
};

const cpuTurnState: EuchreResponse = {
  ...playPhaseState,
  currentPlayerIdx: 1,
};

const noTrumpState: EuchreResponse = {
  ...playPhaseState,
  trumpSuit: 0,
};

beforeEach(() => {
  mockExec.mockResolvedValue(playPhaseState);
});

describe('EuchrePage', () => {
  it('renders skeleton when no state', () => {
    mockExec.mockReturnValue(new Promise(() => undefined));
    renderWithProviders(<EuchrePage />);
    expect(screen.getByTestId('skeleton')).toBeInTheDocument();
  });

  it('calls reset on mount', async () => {
    renderWithProviders(<EuchrePage />);
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 10,
      }),
    );
  });

  it('renders play phase with human cards', async () => {
    renderWithProviders(<EuchrePage />);
    await waitFor(() => {
      expect(screen.getByAltText('\u2660 A')).toBeInTheDocument();
      expect(screen.getByAltText('\u2665 J')).toBeInTheDocument();
    });
  });

  it('greys out the going-alone partner row with a sitting-out badge', async () => {
    mockExec.mockResolvedValue(goingAloneState);
    renderWithProviders(<EuchrePage />);
    const partnerRow = await screen.findByTestId('eu-player-row-2');
    expect(partnerRow.className).toContain('grayscale');
    expect(screen.getByTestId('eu-sitting-out-2')).toBeInTheDocument();
    expect(screen.queryByTestId('eu-sitting-out-1')).not.toBeInTheDocument();
  });

  it('does not show sitting-out badge when no one is going alone', async () => {
    renderWithProviders(<EuchrePage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());
    expect(screen.queryByTestId('eu-sitting-out-2')).not.toBeInTheDocument();
  });

  it('renders pick-up phase with order up and pass buttons', async () => {
    mockExec.mockResolvedValue(pickUpPhaseState);
    renderWithProviders(<EuchrePage />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: '\u30aa\u30fc\u30c0\u30fc\u30a2\u30c3\u30d7' })).toBeInTheDocument();
      expect(
        screen.getByRole('button', {
          name: '\u30aa\u30fc\u30c0\u30fc\u30a2\u30c3\u30d7\uff06\u4e00\u4eba\u3067\u52dd\u8ca0',
        }),
      ).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '\u30d1\u30b9' })).toBeInTheDocument();
    });
  });

  it('shows pick-up phase instruction when human bid turn', async () => {
    mockExec.mockResolvedValue(pickUpPhaseState);
    renderWithProviders(<EuchrePage />);
    await waitFor(() => {
      expect(
        screen.getByText(
          '\u30c7\u30a3\u30fc\u30e9\u30fc\u306b\u3081\u304f\u308a\u672d\u3092\u30aa\u30fc\u30c0\u30fc\u30a2\u30c3\u30d7\u3057\u307e\u3059\u304b\uff1f',
        ),
      ).toBeInTheDocument();
    });
  });

  it('does not show pick-up instruction when cpu bid turn', async () => {
    mockExec.mockResolvedValue(pickUpPhaseCpuTurnState);
    renderWithProviders(<EuchrePage />);
    await waitFor(() => expect(screen.getByText('\u30c1\u30fc\u30e0\u30b9\u30b3\u30a2')).toBeInTheDocument());
    expect(
      screen.queryByText(/\u30aa\u30fc\u30c0\u30fc\u30a2\u30c3\u30d7\u3057\u307e\u3059\u304b/),
    ).not.toBeInTheDocument();
  });

  it('calls orderup command when order up button is clicked', async () => {
    mockExec.mockResolvedValue(pickUpPhaseState);
    renderWithProviders(<EuchrePage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '\u30aa\u30fc\u30c0\u30fc\u30a2\u30c3\u30d7' })).toBeInTheDocument(),
    );

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30aa\u30fc\u30c0\u30fc\u30a2\u30c3\u30d7' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('orderup', undefined, undefined, false));
  });

  it('calls orderup with goAlone when order up alone button is clicked', async () => {
    mockExec.mockResolvedValue(pickUpPhaseState);
    renderWithProviders(<EuchrePage />);
    await waitFor(() =>
      expect(
        screen.getByRole('button', {
          name: '\u30aa\u30fc\u30c0\u30fc\u30a2\u30c3\u30d7\uff06\u4e00\u4eba\u3067\u52dd\u8ca0',
        }),
      ).toBeInTheDocument(),
    );

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(
      screen.getByRole('button', {
        name: '\u30aa\u30fc\u30c0\u30fc\u30a2\u30c3\u30d7\uff06\u4e00\u4eba\u3067\u52dd\u8ca0',
      }),
    );

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('orderup', undefined, undefined, true));
  });

  it('calls pass command when pass button is clicked', async () => {
    mockExec.mockResolvedValue(pickUpPhaseState);
    renderWithProviders(<EuchrePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30d1\u30b9' })).toBeInTheDocument());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30d1\u30b9' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass'));
  });

  it('renders call trump phase with suit buttons', async () => {
    mockExec.mockResolvedValue(callTrumpPhaseState);
    renderWithProviders(<EuchrePage />);
    await waitFor(() => {
      expect(
        screen.getByText(
          '\u5207\u308a\u672d\u30b9\u30fc\u30c8\u3092\u9078\u3093\u3067\u304f\u3060\u3055\u3044\uff08\u3081\u304f\u308a\u672d\u306e\u30b9\u30fc\u30c8\u4ee5\u5916\uff09',
        ),
      ).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '\u30d1\u30b9' })).toBeInTheDocument();
    });
  });

  it('shows the turned-down suit as a disabled, explained button', async () => {
    mockExec.mockResolvedValue(callTrumpPhaseState); // faceUpCard is HEART
    renderWithProviders(<EuchrePage />);
    const heart = await screen.findByRole('button', { name: '♥ ハート' });
    expect(heart).toBeDisabled();
    expect(heart).toHaveAttribute('title', 'このスートは選択できません（めくり札のスート）');
    expect(screen.getByRole('button', { name: '♠ スペード' })).toBeEnabled();
  });

  it('marks the red-suit call-trump buttons with a red ring but not the black-suit ones', async () => {
    mockExec.mockResolvedValue(callTrumpPhaseState);
    renderWithProviders(<EuchrePage />);
    const diamond = await screen.findByRole('button', { name: '♦ ダイヤ' });
    const spade = screen.getByRole('button', { name: '♠ スペード' });
    expect(diamond.className).toContain('ring-ds-error');
    expect(spade.className).not.toContain('ring-ds-error');
  });

  it('renders discard phase with discard button', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<EuchrePage />);
    await waitFor(() => {
      expect(
        screen.getByText('1\u679a\u6368\u3066\u308b\u30ab\u30fc\u30c9\u3092\u9078\u3093\u3067\u304f\u3060\u3055\u3044'),
      ).toBeInTheDocument();
      expect(screen.getByRole('button', { name: '\u30c7\u30a3\u30b9\u30ab\u30fc\u30c9' })).toBeInTheDocument();
    });
  });

  it('discard button disabled when not 1 card selected', async () => {
    mockExec.mockResolvedValue(discardPhaseState);
    renderWithProviders(<EuchrePage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '\u30c7\u30a3\u30b9\u30ab\u30fc\u30c9' })).toBeDisabled(),
    );
  });

  it('play button disabled when not 1 card selected', async () => {
    renderWithProviders(<EuchrePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u51fa\u3059' })).toBeDisabled());
  });

  it('play button enabled when 1 card selected', async () => {
    renderWithProviders(<EuchrePage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement);
    expect(screen.getByRole('button', { name: '\u51fa\u3059' })).not.toBeDisabled();
  });

  it('does not show play button when not human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<EuchrePage />);
    await waitFor(() => expect(screen.getByText('\u30c1\u30fc\u30e0\u30b9\u30b3\u30a2')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '\u51fa\u3059' })).not.toBeInTheDocument();
  });

  it('shows next trick button on trick end', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<EuchrePage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '\u6b21\u306e\u30c8\u30ea\u30c3\u30af' })).toBeInTheDocument(),
    );
  });

  it('shows next round button on round end', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<EuchrePage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '\u6b21\u306e\u30e9\u30a6\u30f3\u30c9' })).toBeInTheDocument(),
    );
  });

  it('shows game end with action log button', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<EuchrePage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('\u68cb\u8b5c\u3092\u898b\u308b')).toBeInTheDocument();
    });
  });

  it('shows game end via gameEndFlag with non-6 phase', async () => {
    mockExec.mockResolvedValue(gameEndByFlagState);
    renderWithProviders(<EuchrePage />);
    await waitFor(() => {
      expect(screen.getByText('Game end!')).toBeInTheDocument();
      expect(screen.getByText('\u68cb\u8b5c\u3092\u898b\u308b')).toBeInTheDocument();
    });
  });

  it('shows error alert', async () => {
    renderWithProviders(<EuchrePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    mockExec.mockReset();
    mockExec.mockRejectedValue(new Error('network error'));
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    await waitFor(() => expect(screen.getByText(NETWORK_ERROR_MESSAGE())).toBeInTheDocument());
  });

  it('settings panel changes cpuDifficulty', async () => {
    renderWithProviders(<EuchrePage />);
    await waitFor(() => expect(screen.getByText('\u30c1\u30fc\u30e0\u30b9\u30b3\u30a2')).toBeInTheDocument());

    fireEvent.click(screen.getByText('\u8a2d\u5b9a'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[0], { target: { value: '2' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, {
        cpuDifficulty: 2,
        pointLimit: 10,
      }),
    );
  });

  it('settings panel changes pointLimit', async () => {
    renderWithProviders(<EuchrePage />);
    await waitFor(() => expect(screen.getByText('\u30c1\u30fc\u30e0\u30b9\u30b3\u30a2')).toBeInTheDocument());

    fireEvent.click(screen.getByText('\u8a2d\u5b9a'));
    const selects = screen.getAllByRole('combobox');
    fireEvent.change(selects[1], { target: { value: '21' } });

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 21,
      }),
    );
  });

  it('card selection toggle via aria-pressed', async () => {
    renderWithProviders(<EuchrePage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.click(cardBtn);
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('card buttons have aria-label with card name', async () => {
    renderWithProviders(<EuchrePage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-label', '\u2660 A');

    const cardBtn2 = screen.getByAltText('\u2665 J').closest('button') as HTMLButtonElement;
    expect(cardBtn2).toHaveAttribute('aria-label', '\u2665 J');
  });

  it('reset button calls apiExec', async () => {
    renderWithProviders(<EuchrePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 10,
      }),
    );
  });

  it('team scores shows both teams', async () => {
    renderWithProviders(<EuchrePage />);
    await waitFor(() => {
      expect(screen.getByText('\u30c1\u30fc\u30e0\u30b9\u30b3\u30a2')).toBeInTheDocument();
    });
  });

  it('score table headers have scope="col" for accessibility', async () => {
    const { container } = renderWithProviders(<EuchrePage />);
    await waitFor(() => expect(screen.getByText('\u30c1\u30fc\u30e0\u30b9\u30b3\u30a2')).toBeInTheDocument());
    const ths = container.querySelectorAll('th');
    ths.forEach((th) => {
      expect(th).toHaveAttribute('scope', 'col');
    });
  });

  it('shows trump suit info', async () => {
    renderWithProviders(<EuchrePage />);
    await waitFor(() => {
      expect(screen.getByText(/\u5207\u308a\u672d/)).toBeInTheDocument();
    });
  });

  it('shows no trump text when trump is 0', async () => {
    mockExec.mockResolvedValue(noTrumpState);
    renderWithProviders(<EuchrePage />);
    await waitFor(() => {
      expect(screen.getByText('\u5207\u308a\u672d\u306a\u3057')).toBeInTheDocument();
    });
  });

  it('shows going alone text', async () => {
    mockExec.mockResolvedValue(goingAloneState);
    renderWithProviders(<EuchrePage />);
    await waitFor(() => {
      expect(screen.getByText('\u4e00\u4eba\u3067\u52dd\u8ca0\u4e2d')).toBeInTheDocument();
    });
  });

  it('shows current trick cards', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<EuchrePage />);
    await waitFor(() => {
      expect(screen.getByText('\u73fe\u5728\u306e\u30c8\u30ea\u30c3\u30af')).toBeInTheDocument();
      expect(screen.getByAltText('\u2666 3')).toBeInTheDocument();
      expect(screen.getByAltText('\u2665 5')).toBeInTheDocument();
    });
  });

  it('does not show current trick when empty', async () => {
    renderWithProviders(<EuchrePage />);
    await waitFor(() => expect(screen.getByText('\u30c1\u30fc\u30e0\u30b9\u30b3\u30a2')).toBeInTheDocument());
    expect(screen.queryByText('\u73fe\u5728\u306e\u30c8\u30ea\u30c3\u30af')).not.toBeInTheDocument();
  });

  it('shows CPU player areas', async () => {
    renderWithProviders(<EuchrePage />);
    await waitFor(() => {
      expect(screen.getByText(/CPU 1.*5\u679a/)).toBeInTheDocument();
      expect(screen.getByText(/CPU 2.*5\u679a/)).toBeInTheDocument();
      expect(screen.getByText(/CPU 3.*5\u679a/)).toBeInTheDocument();
    });
  });

  it('shows loading state', async () => {
    renderWithProviders(<EuchrePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    let resolve!: (value: EuchreResponse) => void;
    const slow = new Promise<EuchreResponse>((r) => {
      resolve = r;
    });
    mockExec.mockReturnValueOnce(slow);
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).toBeDisabled();

    resolve(playPhaseState);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());
  });

  it('calls play command when play button is clicked', async () => {
    renderWithProviders(<EuchrePage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    fireEvent.click(screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement);
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u51fa\u3059' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 0));
  });

  it('calls next when next trick button is clicked', async () => {
    mockExec.mockResolvedValue(trickEndState);
    renderWithProviders(<EuchrePage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '\u6b21\u306e\u30c8\u30ea\u30c3\u30af' })).toBeInTheDocument(),
    );

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u6b21\u306e\u30c8\u30ea\u30c3\u30af' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('calls nextround when next round button is clicked', async () => {
    mockExec.mockResolvedValue(roundEndState);
    renderWithProviders(<EuchrePage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '\u6b21\u306e\u30e9\u30a6\u30f3\u30c9' })).toBeInTheDocument(),
    );

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u6b21\u306e\u30e9\u30a6\u30f3\u30c9' }));

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('round and trick info displayed', async () => {
    renderWithProviders(<EuchrePage />);
    await waitFor(() => {
      expect(screen.getByText('\u30e9\u30a6\u30f3\u30c9 1')).toBeInTheDocument();
      expect(screen.getByText('\u30c8\u30ea\u30c3\u30af 1')).toBeInTheDocument();
    });
  });

  it('does not show message when empty', async () => {
    renderWithProviders(<EuchrePage />);
    await waitFor(() => expect(screen.getByText('\u30c1\u30fc\u30e0\u30b9\u30b3\u30a2')).toBeInTheDocument());
    expect(screen.queryByText('Game end!')).not.toBeInTheDocument();
  });

  // -- ConfirmDialog on reset --

  it('shows confirm dialog when reset button is clicked', async () => {
    renderWithProviders(<EuchrePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();
  });

  it('dismisses confirm dialog on cancel', async () => {
    renderWithProviders(<EuchrePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    expect(screen.getByRole('alertdialog')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: '\u30ad\u30e3\u30f3\u30bb\u30eb' }));
    expect(screen.queryByRole('alertdialog')).not.toBeInTheDocument();
  });

  it('executes reset on confirm', async () => {
    renderWithProviders(<EuchrePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);
    fireEvent.click(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' }));
    fireEvent.click(screen.getByRole('button', { name: '\u78ba\u8a8d' }));

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 10,
      }),
    );
  });

  it('handles action log visibility and API fetch', async () => {
    mockExec.mockResolvedValue(gameEndState);
    renderWithProviders(<EuchrePage />);
    await waitFor(() => expect(screen.getByText('\u68cb\u8b5c\u3092\u898b\u308b')).toBeInTheDocument());

    vi.mocked(actionLogApi.euchre).mockResolvedValueOnce({ entries: [] });
    fireEvent.click(screen.getByText('\u68cb\u8b5c\u3092\u898b\u308b'));

    await waitFor(() => expect(actionLogApi.euchre).toHaveBeenCalledTimes(1));
    expect(screen.getByText('\u68cb\u8b5c')).toBeInTheDocument();

    fireEvent.click(screen.getByText('\u9589\u3058\u308b'));
    await waitFor(() => expect(screen.queryByText(/^\u68cb\u8b5c$/)).not.toBeInTheDocument());
    expect(screen.getByText('\u68cb\u8b5c\u3092\u898b\u308b')).toBeInTheDocument();
  });

  it('does not show action log button when not game end', async () => {
    renderWithProviders(<EuchrePage />);
    await waitFor(() => expect(screen.getByText('\u30c1\u30fc\u30e0\u30b9\u30b3\u30a2')).toBeInTheDocument());
    expect(screen.queryByText('\u68cb\u8b5c\u3092\u898b\u308b')).not.toBeInTheDocument();
  });

  it('sets aria-busy on container', async () => {
    renderWithProviders(<EuchrePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })).not.toBeDisabled());

    const container = screen
      .getByRole('button', { name: '\u30ea\u30bb\u30c3\u30c8' })
      .closest('[aria-busy]') as HTMLElement;
    expect(container).toHaveAttribute('aria-busy', 'false');
  });

  it('no human cards renders empty hand area', async () => {
    const noHuman: EuchreResponse = {
      ...playPhaseState,
      players: playPhaseState.players.map((p) => ({ ...p, isHuman: false })),
    };
    mockExec.mockResolvedValue(noHuman);
    renderWithProviders(<EuchrePage />);
    await waitFor(() => expect(screen.getByText('\u30c1\u30fc\u30e0\u30b9\u30b3\u30a2')).toBeInTheDocument());
    expect(screen.queryByAltText('\u2660 A')).not.toBeInTheDocument();
  });

  it('isHumanTurn false when currentPlayerIdx points to cpu', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<EuchrePage />);
    await waitFor(() => expect(screen.getByText('\u30c1\u30fc\u30e0\u30b9\u30b3\u30a2')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: '\u51fa\u3059' })).not.toBeInTheDocument();
  });

  // --- PhaseIndicator coverage ---

  it('phase indicator shows your turn during pick-up phase', async () => {
    mockExec.mockResolvedValue(pickUpPhaseState);
    renderWithProviders(<EuchrePage />);
    await waitFor(() =>
      expect(screen.getByTestId('phase-indicator')).toHaveTextContent('\u3042\u306a\u305f\u306e\u30bf\u30fc\u30f3'),
    );
  });

  it('phase indicator shows your turn when human play turn', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<EuchrePage />);
    await waitFor(() =>
      expect(screen.getByTestId('phase-indicator')).toHaveTextContent('\u3042\u306a\u305f\u306e\u30bf\u30fc\u30f3'),
    );
  });

  it('phase indicator shows waiting when cpu turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<EuchrePage />);
    await waitFor(() => expect(screen.getByTestId('phase-indicator')).toHaveTextContent('\u5f85\u6a5f\u4e2d'));
  });

  // -- Keyboard navigation --

  it('pressing number key toggles card in play phase', async () => {
    renderWithProviders(<EuchrePage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');

    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('Enter key triggers play in play phase', async () => {
    renderWithProviders(<EuchrePage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    fireEvent.keyDown(document, { key: '1' });
    mockExec.mockClear();
    mockExec.mockResolvedValue(playPhaseState);

    fireEvent.keyDown(document, { key: 'Enter' });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 0));
  });

  it('Escape key clears selection', async () => {
    renderWithProviders(<EuchrePage />);
    await waitFor(() => expect(screen.getByAltText('\u2660 A')).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'true');

    fireEvent.keyDown(document, { key: 'Escape' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('keyboard nav disabled when not human turn', async () => {
    mockExec.mockResolvedValue(cpuTurnState);
    renderWithProviders(<EuchrePage />);
    await waitFor(() => expect(screen.getByText(/CPU 1.*5\u679a/)).toBeInTheDocument());

    const cardBtn = screen.getByAltText('\u2660 A').closest('button') as HTMLButtonElement;
    fireEvent.keyDown(document, { key: '1' });
    expect(cardBtn).toHaveAttribute('aria-pressed', 'false');
  });

  it('shows face-up card during pick-up phase', async () => {
    mockExec.mockResolvedValue(pickUpPhaseState);
    renderWithProviders(<EuchrePage />);
    await waitFor(() => {
      expect(screen.getByText('\u3081\u304f\u308a\u672d')).toBeInTheDocument();
      expect(screen.getByAltText('\u2665 Q')).toBeInTheDocument();
    });
  });

  it('renders tutorial button', async () => {
    renderWithProviders(<EuchrePage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '\u30c1\u30e5\u30fc\u30c8\u30ea\u30a2\u30eb' })).toBeInTheDocument(),
    );
  });

  it('starts tutorial when tutorial button is clicked', async () => {
    renderWithProviders(<EuchrePage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '\u30c1\u30e5\u30fc\u30c8\u30ea\u30a2\u30eb' })).toBeInTheDocument(),
    );
    fireEvent.click(screen.getByRole('button', { name: '\u30c1\u30e5\u30fc\u30c8\u30ea\u30a2\u30eb' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
  });

  it('tutorial can be skipped', async () => {
    renderWithProviders(<EuchrePage />);
    await waitFor(() =>
      expect(screen.getByRole('button', { name: '\u30c1\u30e5\u30fc\u30c8\u30ea\u30a2\u30eb' })).toBeInTheDocument(),
    );
    fireEvent.click(screen.getByRole('button', { name: '\u30c1\u30e5\u30fc\u30c8\u30ea\u30a2\u30eb' }));
    await waitFor(() => expect(screen.getByRole('dialog')).toBeInTheDocument());
    fireEvent.click(screen.getByRole('button', { name: '\u30b9\u30ad\u30c3\u30d7' }));
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
  });

  it('renders accessible h1 heading', async () => {
    mockExec.mockResolvedValue(playPhaseState);
    renderWithProviders(<EuchrePage />);
    await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
  });

  it('renders mobile viewport with horizontal scroll hand', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    try {
      mockExec.mockResolvedValue(playPhaseState);
      renderWithProviders(<EuchrePage />);
      await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
      const hand = document.querySelector('[data-tutorial="eu-player-hand"]');
      expect(hand?.className).toContain('overflow-x-auto');
      expect(hand?.className).not.toContain('flex-wrap');
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  it('renders desktop viewport with wrapping hand', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 800 });
    try {
      mockExec.mockResolvedValue(playPhaseState);
      renderWithProviders(<EuchrePage />);
      await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
      const hand = document.querySelector('[data-tutorial="eu-player-hand"]');
      expect(hand?.className).toContain('flex-wrap');
      expect(hand?.className).not.toContain('overflow-x-auto');
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  it('renders CPU info as collapsible details on mobile', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    try {
      mockExec.mockResolvedValue(playPhaseState);
      const { container } = renderWithProviders(<EuchrePage />);
      await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
      const allDetails = container.querySelectorAll('details');
      const cpuDetails = Array.from(allDetails).find((d) =>
        d.querySelector('summary')?.textContent?.includes('CPU対戦相手'),
      );
      expect(cpuDetails).toBeInTheDocument();
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  it('renders score table as collapsible details on mobile', async () => {
    const originalWidth = window.innerWidth;
    Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: 375 });
    try {
      mockExec.mockResolvedValue(playPhaseState);
      const { container } = renderWithProviders(<EuchrePage />);
      await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
      const scoreDetails = container.querySelector('details[data-tutorial="eu-score-table"]');
      expect(scoreDetails).toBeInTheDocument();
      const summary = scoreDetails?.querySelector('summary');
      expect(summary).toHaveTextContent('チームスコア');
    } finally {
      Object.defineProperty(window, 'innerWidth', { writable: true, configurable: true, value: originalWidth });
    }
  });

  describe('bower badges', () => {
    // Human holds the right bower (♠J, trump) and left bower (♣J, same color as ♠).
    const bowerHandState: EuchreResponse = {
      ...playPhaseState,
      trumpSuit: 1, // spades
      players: playPhaseState.players.map((p, i) =>
        i === 0
          ? {
              ...p,
              cardCount: 3,
              cards: [
                { design: 'SPADE', value: 11 }, // right bower
                { design: 'CLOVER', value: 11 }, // left bower (same color)
                { design: 'HEART', value: 11 }, // off-color jack — no badge
              ],
            }
          : p,
      ),
    };

    it('badges the right and left bowers once trump is set', async () => {
      mockExec.mockResolvedValue(bowerHandState);
      renderWithProviders(<EuchrePage />);
      const right = await screen.findByTestId('eu-bower-badge-0');
      const left = await screen.findByTestId('eu-bower-badge-1');
      expect(right).toHaveAttribute('data-bower', 'right');
      expect(right).toHaveTextContent('右');
      expect(left).toHaveAttribute('data-bower', 'left');
      expect(left).toHaveTextContent('左');
      // Off-color jack at index 2 carries no badge.
      expect(screen.queryByTestId('eu-bower-badge-2')).not.toBeInTheDocument();
    });

    it('does not badge bowers while trump is undecided (pick-up phase)', async () => {
      const undecided: EuchreResponse = { ...bowerHandState, phase: 0, trumpSuit: 0 };
      mockExec.mockResolvedValue(undecided);
      renderWithProviders(<EuchrePage />);
      await waitFor(() => expect(screen.getByRole('heading', { level: 1 })).toBeInTheDocument());
      expect(screen.queryByTestId('eu-bower-badge-0')).not.toBeInTheDocument();
      expect(screen.queryByTestId('eu-bower-badge-1')).not.toBeInTheDocument();
    });
  });
});
