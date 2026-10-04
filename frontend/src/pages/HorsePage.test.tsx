import { fireEvent, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { eightGameApi, horseApi } from '../api/gameApi';
import { TutorialProvider } from '../providers/TutorialProvider';
import { renderWithProviders } from '../test/renderWithProviders';
import { makeHorseState } from '../test/stateFactories';
import { HorsePage, HorsePageContent } from './HorsePage';

vi.mock('../api/gameApi', () => ({
  horseApi: { exec: vi.fn() },
  eightGameApi: { exec: vi.fn() },
  actionLogApi: { horse: vi.fn() },
}));

const mockExec = vi.mocked(horseApi.exec);
const mockEightExec = vi.mocked(eightGameApi.exec);

const handState = makeHorseState();

beforeEach(() => {
  mockExec.mockReset();
  mockEightExec.mockReset();
  mockExec.mockResolvedValue(handState);
  mockEightExec.mockResolvedValue(handState);
});

describe('HorsePage', () => {
  it('calls reset on mount', async () => {
    renderWithProviders(<HorsePage />);
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // **いま何を打っているのかが画面に要る。** ミックスゲームで種目が分からないと、
  // 同じ操作でも規則が違うことに気付けない。
  it('names the current discipline and the hand counters', async () => {
    renderWithProviders(<HorsePage />);
    const info = await screen.findByTestId('ho-discipline');
    expect(info).toHaveTextContent('H');
    expect(info).toHaveTextContent('テキサスホールデム');
    expect(info).toHaveTextContent('ハンド 1/2');
    expect(info).toHaveTextContent('1種目目/全5種目');
    expect(screen.getByTestId('ho-pot')).toHaveTextContent('30');
  });

  it('puts the game, discipline, round, and action type beside the controls', async () => {
    mockExec.mockResolvedValue(makeHorseState({ tablePhase: 1 }));
    renderWithProviders(<HorsePage />);
    expect(await screen.findByTestId('ho-action-context')).toHaveTextContent('H.O.R.S.E.');
    expect(screen.getByTestId('ho-action-context')).toHaveTextContent('テキサスホールデム');
    expect(screen.getByTestId('ho-action-context')).toHaveTextContent('プリフロップ');
    expect(screen.getByTestId('ho-action-context')).toHaveTextContent('ベッティング');
  });

  it.each([
    ['Hold’em end', { tablePhase: 6 }],
    ['Hold’em rebuy', { tablePhase: 7 }],
    ['Stud end', { discipline: 2, disciplineName: 'stud', tablePhase: 7 }],
    ['Stud rebuy', { discipline: 2, disciplineName: 'stud', tablePhase: 8 }],
    ['Triple Draw end', { discipline: 7, disciplineName: 'tripleDraw', tablePhase: 5 }],
  ])('hides the table round after %s', async (_label, state) => {
    mockExec.mockResolvedValue(makeHorseState(state));
    renderWithProviders(<HorsePage />);
    const context = await screen.findByTestId('ho-action-context');
    expect(context).not.toHaveTextContent('ショーダウン');
    expect(context).not.toHaveTextContent('ベッティング');
  });

  it('identifies Eight-Game Mix separately and marks draw turns', async () => {
    mockEightExec.mockResolvedValue(
      makeHorseState({
        variant: 1,
        discipline: 7,
        disciplineLetter: '2-7',
        disciplineName: 'tripleDraw',
        isDrawPhase: true,
        drawIndex: 2,
        tablePhase: 3,
      }),
    );
    renderWithProviders(
      <TutorialProvider config={{ gameName: 'eightgame', steps: [] }}>
        <HorsePageContent gameKey="eightgame" />
      </TutorialProvider>,
    );
    const context = await screen.findByTestId('ho-action-context');
    expect(context).toHaveTextContent('エイトゲーム・ミックス');
    expect(context).toHaveTextContent('2-7 トリプルドロー');
    expect(context).toHaveTextContent('引き直し');
    expect(context).toHaveTextContent('2');
  });

  it('clears the discipline highlight and announcement on the following unchanged hand', async () => {
    const initial = makeHorseState({ variant: 1, phase: 1 });
    mockEightExec
      .mockResolvedValueOnce(initial)
      .mockResolvedValueOnce(
        makeHorseState({
          variant: 1,
          phase: 1,
          discipline: 5,
          disciplineLetter: 'N',
          disciplineName: 'nlHoldem',
          disciplinePosition: 6,
          handNumber: 2,
        }),
      )
      .mockResolvedValueOnce(
        makeHorseState({
          variant: 1,
          phase: 1,
          discipline: 5,
          disciplineLetter: 'N',
          disciplineName: 'nlHoldem',
          disciplinePosition: 6,
          handNumber: 3,
        }),
      );
    renderWithProviders(
      <TutorialProvider config={{ gameName: 'eightgame', steps: [] }}>
        <HorsePageContent gameKey="eightgame" />
      </TutorialProvider>,
    );

    const discipline = await screen.findByTestId('ho-discipline');
    expect(screen.getByRole('status').textContent).toBe('');
    fireEvent.click(await screen.findByTestId('ho-next-hand'));
    await waitFor(() => expect(discipline).toHaveAttribute('data-discipline-changed', 'true'));
    expect(screen.getByRole('status')).toHaveTextContent('種目が切り替わりました。ノーリミット・ホールデム。');

    fireEvent.click(await screen.findByTestId('ho-next-hand'));
    await waitFor(() => expect(discipline).not.toHaveAttribute('data-discipline-changed', 'true'));
    expect(screen.getByRole('status').textContent).toBe('');
  });

  it('clears the discipline marker when a new game rolls the hand number back', async () => {
    mockEightExec
      .mockResolvedValueOnce(makeHorseState({ variant: 1, phase: 1, handNumber: 7 }))
      .mockResolvedValueOnce(
        makeHorseState({ variant: 1, phase: 1, discipline: 5, disciplineName: 'nlHoldem', handNumber: 8 }),
      )
      .mockResolvedValueOnce(
        makeHorseState({
          variant: 1,
          phase: 2,
          gameEndFlag: true,
          discipline: 5,
          disciplineName: 'nlHoldem',
          handNumber: 8,
        }),
      )
      .mockResolvedValueOnce(makeHorseState({ variant: 1, phase: 0, handNumber: 1 }));
    renderWithProviders(
      <TutorialProvider config={{ gameName: 'eightgame', steps: [] }}>
        <HorsePageContent gameKey="eightgame" />
      </TutorialProvider>,
    );
    const discipline = await screen.findByTestId('ho-discipline');
    fireEvent.click(await screen.findByTestId('ho-next-hand'));
    await waitFor(() => expect(discipline).toHaveAttribute('data-discipline-changed', 'true'));
    // Advance to the completed-game view, then start a fresh game.
    fireEvent.click(await screen.findByTestId('ho-next-hand'));
    await waitFor(() => expect(screen.getByTestId('ho-result')).toBeInTheDocument());
    expect(discipline).toHaveAttribute('data-discipline-changed', 'true');
    fireEvent.click(await screen.findByRole('button', { name: '新しいゲーム' }));
    await waitFor(() => expect(discipline).not.toHaveAttribute('data-discipline-changed', 'true'));
    expect(screen.getByRole('status').textContent).toBe('');
  });

  it('does not render the announcement region in H.O.R.S.E. mode', async () => {
    renderWithProviders(<HorsePage />);
    await screen.findByTestId('ho-discipline');
    expect(screen.queryByRole('status')).not.toBeInTheDocument();
  });

  it('shows the discipline position for both rotations', async () => {
    mockExec.mockResolvedValue(makeHorseState({ disciplinePosition: 3, disciplineTotal: 8 }));
    const { unmount } = renderWithProviders(<HorsePage />);
    expect(await screen.findByTestId('ho-discipline')).toHaveTextContent('3種目目/全8種目');
    unmount();

    mockExec.mockResolvedValue(makeHorseState({ disciplinePosition: 5, disciplineTotal: 5 }));
    renderWithProviders(<HorsePage />);
    expect(await screen.findByTestId('ho-discipline')).toHaveTextContent('5種目目/全5種目');
  });

  it('shows every seat with its chips', async () => {
    renderWithProviders(<HorsePage />);
    expect(await screen.findByTestId('ho-seat-0')).toBeInTheDocument();
    for (const id of [0, 1, 2, 3]) {
      expect(screen.getByTestId(`ho-seat-${id}-chips`)).toHaveTextContent('1000');
    }
  });

  it('shows the amount needed to call, but not when checking is free', async () => {
    mockExec.mockResolvedValueOnce(makeHorseState({ toCall: 25 }));
    const { unmount } = renderWithProviders(<HorsePage />);
    const callButton = await screen.findByRole('button', { name: /コール/ });
    expect(callButton).toHaveTextContent('コール （25 チップ）');

    unmount();
    mockExec.mockResolvedValueOnce(makeHorseState({ toCall: 0 }));
    renderWithProviders(<HorsePage />);
    await waitFor(() => expect(screen.getByRole('button', { name: 'チェック' })).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: /コール/ })).not.toBeInTheDocument();
  });

  it('shows folded and all-in seat statuses', async () => {
    const base = makeHorseState();
    mockExec.mockResolvedValue({
      ...base,
      seats: base.seats.map((seat, i) => ({ ...seat, folded: i === 1, allIn: i === 2 })),
    });
    renderWithProviders(<HorsePage />);
    expect(await screen.findByTestId('ho-seat-1')).toHaveTextContent('フォールド');
    expect(screen.getByTestId('ho-seat-2')).toHaveTextContent('オールイン');
  });

  // **6 つの手をすべて送れる。** 綴りが 1 つ違うと、その手だけが打てなくなる。
  it.each([
    ['コール', 'call'],
    ['フォールド', 'fold'],
    ['オールイン', 'allin'],
  ] as const)('sends %s while a bet is outstanding', async (label, action) => {
    renderWithProviders(<HorsePage />);
    fireEvent.click(await screen.findByRole('button', { name: label === 'コール' ? /^コール/ : label }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('action', { action }));
  });

  // **賭けられていなければチェックとベットになる。** toCall を見ずに決め打つと、
  // チェックできる場面でコールしか出せない。
  it.each([
    ['チェック', 'check'],
    ['フォールド', 'fold'],
  ] as const)('sends %s when nothing is outstanding', async (label, action) => {
    mockExec.mockResolvedValue(makeHorseState({ toCall: 0 }));
    renderWithProviders(<HorsePage />);
    fireEvent.click(await screen.findByRole('button', { name: label }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('action', { action }));
  });

  // **ベットとレイズは額を添えて送る。** 添えないとサーバに断られる。
  it('sends bet with an amount when nothing is outstanding', async () => {
    mockExec.mockResolvedValue(makeHorseState({ toCall: 0 }));
    renderWithProviders(<HorsePage />);
    fireEvent.click(await screen.findByRole('button', { name: 'ベット' }));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('action', { action: 'bet', amount: expect.any(Number) }));
    const betCall = mockExec.mock.calls.find((c) => c[1] && 'amount' in c[1]);
    expect(betCall?.[1]?.amount).toBeGreaterThan(0);
  });

  it('sends raise with an amount while a bet is outstanding', async () => {
    renderWithProviders(<HorsePage />);
    fireEvent.click(await screen.findByRole('button', { name: 'レイズ' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('action', { action: 'raise', amount: expect.any(Number) }),
    );
  });

  it('shows the community cards once they are dealt', async () => {
    mockExec.mockResolvedValue(
      makeHorseState({
        communityCards: [
          { design: 'SPADE', value: 10, glyph: '♠', label: '10', color: 'black', deck: 'standard' },
          { design: 'HEART', value: 4, glyph: '♥', label: '4', color: 'red', deck: 'standard' },
          { design: 'CLOVER', value: 7, glyph: '♣', label: '7', color: 'black', deck: 'standard' },
        ],
      }),
    );
    renderWithProviders(<HorsePage />);
    const board = await screen.findByTestId('ho-community');
    expect(board).toBeInTheDocument();
  });

  // **同じ数字でも種目で意味が違う** (#5788)。ホールデムの 2 はフロップ、
  // スタッドの 2 は 4th street。
  it('names the betting round per discipline', async () => {
    mockExec.mockResolvedValue(makeHorseState({ disciplineName: 'holdem', tablePhase: 2 }));
    const { unmount } = renderWithProviders(<HorsePage />);
    expect(await screen.findByTestId('ho-round')).toHaveTextContent('フロップ');
    unmount();

    mockExec.mockResolvedValue(
      makeHorseState({ discipline: 2, disciplineLetter: 'R', disciplineName: 'razz', tablePhase: 2 }),
    );
    const razz = renderWithProviders(<HorsePage />);
    expect(await screen.findByTestId('ho-round')).toHaveTextContent('フォースストリート');
    razz.unmount();

    // オマハはコミュニティ系。
    mockExec.mockResolvedValue(makeHorseState({ discipline: 1, disciplineName: 'omahaHiLo', tablePhase: 4 }));
    renderWithProviders(<HorsePage />);
    expect(await screen.findByTestId('ho-round')).toHaveTextContent('リバー');
  });

  // **配る前に名前は無い。** 0 は「まだ開始していない」。
  it('shows no round name before the hand starts', async () => {
    mockExec.mockResolvedValue(makeHorseState({ disciplineName: 'holdem', tablePhase: 0 }));
    renderWithProviders(<HorsePage />);
    await screen.findByTestId('ho-discipline');
    expect(screen.queryByTestId('ho-round')).not.toBeInTheDocument();
  });

  // **共有札はスタッド系には無い。** 常に描くと、無いはずの場が出る。
  it('hides the community row in the stud disciplines', async () => {
    mockExec.mockResolvedValue(
      makeHorseState({ discipline: 2, disciplineLetter: 'R', disciplineName: 'razz', communityCards: [] }),
    );
    renderWithProviders(<HorsePage />);
    expect(await screen.findByTestId('ho-discipline')).toHaveTextContent('ラズ');
    expect(screen.queryByTestId('ho-community')).not.toBeInTheDocument();
  });

  it('offers the next hand once the hand is settled', async () => {
    mockExec.mockResolvedValue(makeHorseState({ phase: 1, isHumanTurn: false }));
    renderWithProviders(<HorsePage />);
    fireEvent.click(await screen.findByTestId('ho-next-hand'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  // **決着していない間はベッティングを出さない。** 出すと押せるのに何も起きない。
  it('hides the betting controls when it is not the human turn', async () => {
    mockExec.mockResolvedValue(makeHorseState({ isHumanTurn: false }));
    renderWithProviders(<HorsePage />);
    await screen.findByTestId('ho-discipline');
    expect(screen.queryByRole('button', { name: 'フォールド' })).not.toBeInTheDocument();
  });

  it('shows the final standings and starts a new match', async () => {
    mockExec.mockResolvedValue(
      makeHorseState({
        phase: 2,
        gameEndFlag: true,
        isHumanTurn: false,
        winnerSeat: 0,
        seats: handState.seats.map((s) => (s.id === 0 ? { ...s, chips: 4000 } : { ...s, chips: 0 })),
      }),
    );
    renderWithProviders(<HorsePage />);
    const result = await screen.findByTestId('ho-result');
    expect(result).toHaveTextContent('あなた');
    mockExec.mockClear();
    fireEvent.click(screen.getByRole('button', { name: '新しいゲーム' }));
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { seats: 4, handsPerDiscipline: 2 } }),
    );
  });

  it('shows confirmed showdown hands and omits hands that were not revealed', async () => {
    mockExec.mockResolvedValue(
      makeHorseState({
        phase: 1,
        seats: makeHorseState().seats.map((seat, i) => ({ ...seat, handName: i === 0 ? 'One Pair' : '' })),
      }),
    );
    renderWithProviders(<HorsePage />);
    const result = await screen.findByTestId('ho-result');
    expect(result).toHaveTextContent('役: ワンペア');
    expect(result.querySelectorAll('span')).toHaveLength(1);
  });

  // **PLO ラウンドでポット超過額がボタン操作前に分かる (#6622)。**
  // maxBetAmount が設定されている場合、Maxプリセットが表示され、超過入力で警告が出る。
  it('handles maxBetAmount in pot-limit rounds and keeps negative controls', async () => {
    // 1. maxBetAmount: 120 (PLO ラウンド)
    mockExec.mockResolvedValue(
      makeHorseState({
        disciplineLetter: 'PLO',
        disciplineName: 'plOmaha',
        pot: 100,
        toCall: 20,
        minRaise: 20,
        maxBetAmount: 120,
      }),
    );
    const { unmount } = renderWithProviders(<HorsePage />);

    // Max プリセットボタンが表示される
    const maxBtn = await screen.findByRole('button', { name: 'Max' });
    expect(maxBtn).toBeInTheDocument();

    // Max ボタンをクリックするとベット額が 120 に設定される
    fireEvent.click(maxBtn);
    const input = screen.getByLabelText('ベット額:') as HTMLInputElement;
    expect(input.value).toBe('120');

    // 120 超過の額 (例: 150) を入力すると警告が表示され、レイズボタンが無効化される
    fireEvent.change(input, { target: { value: '150' } });
    expect(screen.getByRole('alert')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'レイズ' })).toBeDisabled();
    unmount();

    // 2. 負のコントロール: maxBetAmount: 250 に変えると Max が 250 になり、150 では警告が出ない
    mockExec.mockResolvedValue(
      makeHorseState({
        disciplineLetter: 'PLO',
        disciplineName: 'plOmaha',
        pot: 200,
        toCall: 50,
        minRaise: 50,
        maxBetAmount: 250,
      }),
    );
    const render250 = renderWithProviders(<HorsePage />);
    const maxBtn250 = await screen.findByRole('button', { name: 'Max' });
    fireEvent.click(maxBtn250);
    const input250 = screen.getByLabelText('ベット額:') as HTMLInputElement;
    expect(input250.value).toBe('250');
    fireEvent.change(input250, { target: { value: '150' } });
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'レイズ' })).not.toBeDisabled();
    render250.unmount();

    // 3. 負のコントロール: maxBetAmount: 0 (HORSE 固定リミット) では Max ボタンが表示されない
    mockExec.mockResolvedValue(
      makeHorseState({
        disciplineLetter: 'H',
        disciplineName: 'holdem',
        pot: 30,
        toCall: 20,
        minRaise: 20,
        maxBetAmount: 0,
      }),
    );
    renderWithProviders(<HorsePage />);
    await screen.findByRole('button', { name: 'レイズ' });
    expect(screen.queryByRole('button', { name: 'Max' })).not.toBeInTheDocument();
  });

  // ドローフェーズの手番では交換カード選択エリアを表示する。
  it('shows draw area during human draw turn', async () => {
    mockExec.mockResolvedValue(makeHorseState({ isDrawPhase: true, drawIndex: 1 }));
    renderWithProviders(<HorsePage />);
    await waitFor(() => expect(screen.getByTestId('ho-draw')).toBeInTheDocument());
  });

  it('names each draw card and exposes its selected state', async () => {
    mockExec.mockResolvedValue(makeHorseState({ isDrawPhase: true, drawIndex: 1 }));
    renderWithProviders(<HorsePage />);

    const ace = await screen.findByRole('button', { name: 'A ♠' });
    expect(ace).toHaveAttribute('aria-pressed', 'false');
    fireEvent.click(ace);
    expect(screen.getByRole('button', { name: 'A ♠' })).toHaveAttribute('aria-pressed', 'true');
    expect(screen.getByRole('button', { name: 'K ♥' })).toHaveAttribute('aria-pressed', 'false');
  });

  // ベットラウンドなどドローフェーズ以外では交換カード選択エリアを表示しない。
  it('hides draw area outside draw turn', async () => {
    mockExec.mockResolvedValue(handState);
    renderWithProviders(<HorsePage />);
    await screen.findByTestId('ho-discipline');
    expect(screen.queryByTestId('ho-draw')).not.toBeInTheDocument();
  });

  // ドローフェーズの手番ではカード交換ボタンを表示する。
  it('shows draw exchange button during human draw turn', async () => {
    mockExec.mockResolvedValue(makeHorseState({ isDrawPhase: true, drawIndex: 1 }));
    renderWithProviders(<HorsePage />);
    await waitFor(() => expect(screen.getByTestId('ho-draw-exchange')).toBeInTheDocument());
  });

  // ドローフェーズ以外ではカード交換ボタンを表示しない。
  it('hides draw exchange button outside draw turn', async () => {
    mockExec.mockResolvedValue(handState);
    renderWithProviders(<HorsePage />);
    await screen.findByTestId('ho-discipline');
    expect(screen.queryByTestId('ho-draw-exchange')).not.toBeInTheDocument();
  });
});
