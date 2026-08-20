import { useCallback, useEffect, useMemo, useState } from 'react';
import type { trogguApi as TrogguApi } from '../api/gameApi';
import { trogguApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { ActionShortcutsPanel } from '../components/ActionShortcutsPanel';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { TrogguResponse } from '../types/card';
import { TrogguPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseTrogguCommand, TROGGU_HELP } from '../utils/cli/commands/trogguCommands';
import { formatTrogguState } from '../utils/cli/formatters/trogguFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Troggu tutorial step definitions. */
const TROGGU_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="tg-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tg-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="tg-actions"]', messageKey: 'tutorial.actions', placement: 'top', advanceOn: 'next' },
];

/**
 * The four contracts, in bidding order.
 *
 * **Each one changes what winning means**, so they are listed with their goal
 * rather than as bare names.
 */
const CONTRACTS = ['trois', 'solo', 'piccolo', 'misere'] as const;

/**
 * Bid values, in the same order the domain ranks them
 * (`internal/domain/Troggu.go`: trois 1 < solo 2 < piccolo 3 < misère 4).
 *
 * **順位は名前から読めない。**ミゼールがソロより上なので、CPU がミゼールを
 * 宣言した配りでは人間のソロは却下される (#5808)。
 */
const CONTRACT_VALUE: Record<(typeof CONTRACTS)[number], number> = {
  trois: 1,
  solo: 2,
  piccolo: 3,
  misere: 4,
};

/**
 * Bid value -> contract key, for naming the bid that is currently winning.
 *
 * 0 は「まだ誰も入札していない」で、訳語は `-`。ここに置いておかないと
 * 呼び出し側に既定値の分岐が要る。
 */
const CONTRACT_NAME: Record<number, string> = {
  0: 'pass',
  1: 'trois',
  2: 'solo',
  3: 'piccolo',
  4: 'misere',
};

/** CPU difficulty options. */
const CPU_DIFFICULTY_OPTIONS = [0, 1, 2] as const;

/** Match-length options, in deals. */
const TARGET_DEALS_OPTIONS = [1, 2, 4, 8, 12] as const;

/** Renders the Troggu (トロッグ) page: a Swiss Valais tarot with four contracts. */
export const TrogguPage = withTutorial(TrogguPageContent, 'troggu', TROGGU_TUTORIAL_STEPS);

/** Inner content of the Troggu page, wrapped by TutorialWrapper. */
function TrogguPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('troggu');
  const [cpuDifficulty, setCpuDifficulty] = useState(1);
  const [targetDeals, setTargetDeals] = useState(4);

  const { loading, error, state, exec: callApi, retry } = useGameApi(trogguApi.exec);
  const { cardWidth, isMobile } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('troggu', state);

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    callApi('reset');
  }, []);

  const resetWithConfig = useCallback(() => {
    callApi('reset', { config: { cpuDifficulty, targetDeals } });
  }, [callApi, cpuDifficulty, targetDeals]);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('troggu');
  const cliConfig: CliGameConfig<TrogguResponse, Parameters<typeof TrogguApi.exec>> = useMemo(
    () => ({
      gameName: 'troggu',
      parseCommand: parseTrogguCommand,
      formatResponse: formatTrogguState,
      helpText: TROGGU_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(callApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  // **キーボードから届かない操作を残さない** (#5787)。有効条件はボタンの
  // 表示条件と同じ値なので、押せないのにキーだけ通ることはない。
  //
  // **フックは早期 return より上。** `if (!state)` の下に置くと初回レンダーだけ
  // フック数が変わる (#4561)。
  const phaseForKbd = state?.phase;
  const gameEndForKbd = !!state?.gameEndFlag;
  const canPassForKbd = phaseForKbd === TrogguPhase.BID && !!state?.isHumanTurn && !gameEndForKbd;
  const isTrickEndForKbd = phaseForKbd === TrogguPhase.TRICK_END;
  const canAdvanceForKbd = isTrickEndForKbd || (phaseForKbd === TrogguPhase.ROUND_END && !gameEndForKbd);
  const advanceAction = useCallback(() => {
    callApi(isTrickEndForKbd ? 'next' : 'nextround');
  }, [callApi, isTrickEndForKbd]);
  const actionBindings = useMemo(
    () => [
      { key: 'p', action: () => callApi('pass'), enabled: canPassForKbd, label: 'pass' },
      { key: 'n', action: advanceAction, enabled: canAdvanceForKbd, label: 'next' },
    ],
    [callApi, canPassForKbd, advanceAction, canAdvanceForKbd],
  );
  useActionKeyboardNav({ bindings: actionBindings, enabled: !loading });

  if (!state) {
    return <GameSkeleton gameKey="troggu" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 18 }} />;
  }

  const human = state.players.find((p) => p.isHuman) ?? state.players[0] ?? null;
  const isBid = state.phase === TrogguPhase.BID;
  const isPlay = state.phase === TrogguPhase.PLAY;
  const isTrickEnd = state.phase === TrogguPhase.TRICK_END;
  const isRoundEnd = state.phase === TrogguPhase.ROUND_END;
  const isGameEnd = state.gameEndFlag;

  const isHumanTurn = state.isHumanTurn && !isGameEnd;
  const humanWon = isGameEnd && state.winnerPlayer === (human?.id ?? 0);
  const phaseName = isGameEnd
    ? t('phase.gameEnd')
    : isRoundEnd
      ? t('phase.roundEnd')
      : isTrickEnd
        ? t('phase.trickEnd')
        : isBid
          ? t('phase.bid')
          : t('phase.play');

  /** Plays the clicked card immediately (there is nothing to select in this game). */
  const toggleCard = (idx: number) => {
    if (!isPlay || !isHumanTurn) return;
    callApi('play', { cardIndex: idx });
  };

  const seatName = (id: number): string =>
    state.players[id]?.isHuman ? t('you') : t('cpu', { n: id, defaultValue: `CPU${id}` });

  /** Renders the deal result in the units the contract is scored in. */
  const roundResultLine = (): string => {
    const bd = state.breakdown;
    if (!bd) return t('roundResult.thrownIn');
    // **ソロだけが点数、他はトリック数。** 単位を取り違えると読めない結果になる。
    const key = bd.targetIsTricks
      ? bd.won
        ? 'roundResult.wonTricks'
        : 'roundResult.lostTricks'
      : bd.won
        ? 'roundResult.wonPoints'
        : 'roundResult.lostPoints';
    return t(key, {
      got: bd.targetIsTricks ? bd.declarerTricks : bd.declarerPoints,
      target: bd.target,
    });
  };

  return (
    <GamePageShell
      title={tc('nav.troggu')}
      gameThemeBg={gameTheme.troggu.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/troggu"
      gameEndFlag={isGameEnd}
      winShow={humanWon}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={<CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />}
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <div className="flex-1 overflow-y-auto px-4 py-2 space-y-3">
            <div
              className="text-center text-xs text-ds-text-muted"
              data-testid="tg-info"
              // **手番かどうかを DOM に出す。** 出さないと E2E は「押しても
              // 何も起きない」を待つことになり、配り次第で落ちる。
              data-human-turn={isHumanTurn || undefined}
            >
              <span className="mr-3">{t('deal', { n: state.roundNumber, total: state.totalRounds })}</span>
              <span className="mr-3">{t('trick', { n: state.trickNumber })}</span>
              <span>{t('contract', { name: t(`contract.${state.contractName}`, { defaultValue: '-' }) })}</span>
            </div>

            <div className="flex flex-wrap justify-center gap-3" data-testid="tg-seats">
              {state.players.map((p) => (
                <div key={p.id} className="text-center text-xs text-ds-text-muted" data-testid={`tg-seat-${p.id}`}>
                  <div className="text-ds-text-primary">
                    {seatName(p.id)}
                    {p.isDeclarer && <span className="ml-1 text-ds-warning">{t('roleDeclarer')}</span>}
                  </div>
                  <div>{t('cards', { count: p.cardCount })}</div>
                  <div>{t('tricks', { count: p.trickCount })}</div>
                  <div data-testid={`tg-seat-${p.id}-points`}>{t('points', { points: p.cardPoints })}</div>
                  <div>{t('score', { score: p.score })}</div>
                </div>
              ))}
            </div>

            <TrickDisplay
              currentTrick={state.currentTrick}
              players={state.players.map((p) => ({ id: p.id, name: seatName(p.id), isHuman: p.isHuman }))}
              cardWidth={cardWidth}
              label={t('currentTrick')}
              dataTutorial="tg-trick-display"
              winnerIdx={isTrickEnd ? state.lastTrickWinner : undefined}
              winnerLabel={t('trickWinner')}
            />

            {human && (
              <PlayerHandSection
                humanPlayer={human}
                selectedCardIndices={[]}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="tg"
                validIndices={isPlay && isHumanTurn ? state.playableIndices : undefined}
                restrictedTooltip={t('restricted')}
              />
            )}

            {isRoundEnd && (
              <div className="my-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="tg-round-result">
                <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                <div className="text-ds-success mb-1">{roundResultLine()}</div>
                {state.breakdown?.seats.map((delta, i) => (
                  <div key={i} data-testid={`tg-round-seat-${i}`}>
                    {t('roundResult.seat', { name: seatName(i), delta })}
                  </div>
                ))}
              </div>
            )}

            {isGameEnd && (
              <div className="my-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="tg-result">
                <div className="mb-1 text-ds-text-primary">{t('result.title')}</div>
                <div className="text-ds-success mb-1">
                  {state.winnerPlayer < 0
                    ? t('result.draw')
                    : t('result.winner', { name: seatName(state.winnerPlayer) })}
                </div>
                {state.players.map((p) => (
                  <div key={p.id}>{t('result.score', { name: seatName(p.id), score: p.score })}</div>
                ))}
              </div>
            )}

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />
            <ErrorAlert message={error} onRetry={retry} />
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />
            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          <SettingsPanel
            title={t('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select' as const,
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: String(cpuDifficulty),
                    options: CPU_DIFFICULTY_OPTIONS.map((v) => ({
                      value: String(v),
                      label: t(`settings.difficulty${v}`),
                    })),
                    onSelect: (v: string) => setCpuDifficulty(Number.parseInt(v, 10)),
                  },
                  {
                    type: 'select' as const,
                    id: 'targetDeals',
                    label: t('settings.targetDeals'),
                    value: String(targetDeals),
                    options: TARGET_DEALS_OPTIONS.map((v) => ({ value: String(v), label: String(v) })),
                    onSelect: (v: string) => setTargetDeals(Number.parseInt(v, 10)),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <GameFooter className={`${gameTheme.troggu.footer} px-4 py-2.5`}>
            <ActionShortcutsPanel bindings={actionBindings} data-testid="tg-kbd-shortcuts" />
            <div className="flex gap-2 justify-center flex-wrap items-center" data-tutorial="tg-actions">
              {isBid &&
                isHumanTurn &&
                CONTRACTS.map((c) => (
                  <button
                    key={c}
                    type="button"
                    className={btnSecondary}
                    onClick={() => callApi('bid', { bid: c })}
                    // **今の最高入札を超えられない契約は押させない。**押せても
                    // サーバーに却下されるだけで、画面は入札のまま動かない (#5808)。
                    disabled={loading || CONTRACT_VALUE[c] <= state.highestBid}
                    title={
                      CONTRACT_VALUE[c] <= state.highestBid
                        ? t('bidTooLow', { high: t(`contract.${CONTRACT_NAME[state.highestBid]}`) })
                        : undefined
                    }
                    data-testid={`tg-bid-${c}`}
                  >
                    {t(`contract.${c}`)}
                  </button>
                ))}
              {isBid && isHumanTurn && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={() => callApi('pass')}
                  disabled={loading}
                  data-testid="tg-pass"
                >
                  {t('pass')}
                </button>
              )}
              {isTrickEnd && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={() => callApi('next')}
                  disabled={loading}
                  data-testid="tg-next-trick"
                >
                  {t('nextTrick')}
                </button>
              )}
              {isRoundEnd && !isGameEnd && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={() => callApi('nextround')}
                  disabled={loading}
                  data-testid="tg-next-round"
                >
                  {t('nextDeal')}
                </button>
              )}
              {isGameEnd && (
                <button type="button" className={btnSuccess} onClick={resetWithConfig} disabled={loading}>
                  {t('newGame')}
                </button>
              )}
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={resetWithConfig}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="tg-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
