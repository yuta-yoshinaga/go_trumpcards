import { useCallback, useEffect, useMemo, useState } from 'react';
import { snapApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { HintTooltip } from '../components/hint/HintTooltip';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { btnDanger, btnPrimary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { SnapResponse } from '../types/card';
import { SnapEventKind, SnapPendingKind, SnapPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseSnapCommand, SNAP_HELP } from '../utils/cli/commands/snapCommands';
import { formatSnapState } from '../utils/cli/formatters/snapFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/**
 * How often to ask the server to advance a booked CPU action.
 *
 * **100ms keeps the reaction distribution intact** — the hardest setting draws
 * around 500ms, so polling any slower would quantise it into something the
 * human could time.
 */
const SNAP_TICK_INTERVAL_MS = 100;

/** Guided tutorial steps (the moving trigger, the single card, the penalty, stocks). */
const SNAP_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="sp-rule"]', messageKey: 'tutorial.rule', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="sp-pile"]', messageKey: 'tutorial.single', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="sp-actions"]', messageKey: 'tutorial.penalty', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="sp-seats"]', messageKey: 'tutorial.stock', placement: 'bottom', advanceOn: 'next' },
];

/**
 * Inner content for the Snap page (wrapped by `withTutorial`).
 *
 * A reflex game made deterministic: the CPUs book their call behind a deadline
 * and the page polls `tick` until it fires, so **claiming before that deadline
 * genuinely wins the pile**. The poll is gated on there being something booked
 * rather than on whose turn it is — a CPU's call is booked during the human's
 * turn too, and gating on the turn would stop the CPUs ever calling.
 */
function SnapPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('snap');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
  } = useGameApi<SnapResponse, Parameters<typeof snapApi.exec>>(snapApi.exec);
  const { cardWidth } = useCardDimensions();
  const { hint, hintEnabled, setHintEnabled } = useGameHint('snap', state);
  const [playerCnt, setPlayerCnt] = useState(2);
  const [difficulty, setDifficulty] = useState(1);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('snap');
  const cliConfig: CliGameConfig<SnapResponse, Parameters<typeof snapApi.exec>> = useMemo(
    () => ({
      gameName: 'snap',
      parseCommand: parseSnapCommand,
      formatResponse: formatSnapState,
      helpText: SNAP_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(dispatch, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useEffect(() => {
    void dispatch('reset');
  }, [dispatch]);

  // **予約でゲートする。手番ではない。** CPU の宣言は人間の手番中にも予約される
  // ので、「CPU の手番だけ」に絞ると CPU が永久に宣言しなくなる。
  const isCpuPending = state !== null && state.pendingKind !== SnapPendingKind.NONE;
  const isRunning = state !== null && !state.gameEndFlag;
  useEffect(() => {
    if (!isRunning || !isCpuPending) return;
    const id = window.setInterval(() => {
      void dispatch('tick');
    }, SNAP_TICK_INTERVAL_MS);
    return () => window.clearInterval(id);
  }, [isRunning, isCpuPending, dispatch]);

  const handleReset = useCallback(() => {
    hideActionLog();
    void dispatch('reset', { playerCnt, cpuDifficulty: difficulty });
  }, [dispatch, hideActionLog, playerCnt, difficulty]);

  const handleStep = useCallback(() => {
    void dispatch('step');
  }, [dispatch]);

  const handleSnap = useCallback(() => {
    void dispatch('snap');
  }, [dispatch]);

  const handleGiveUp = useCallback(() => {
    void dispatch('giveup');
  }, [dispatch]);

  if (!state) {
    return <GameSkeleton gameKey="snap" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 0 }} />;
  }

  const isGameEnd = state.phase === SnapPhase.GAME_END || state.gameEndFlag;

  const eventLine = (() => {
    const name =
      state.lastEventPlayerIdx === 0 ? t('header.you') : t('header.cpu', { idx: String(state.lastEventPlayerIdx) });
    switch (state.lastEventKind) {
      case SnapEventKind.SNAP_CORRECT:
        return t('event.snapCorrect', { name });
      case SnapEventKind.SNAP_WRONG:
        return t('event.snapWrong', { name });
      case SnapEventKind.ELIMINATED:
        return t('event.eliminated', { name });
      case SnapEventKind.STEP:
        return t('event.step', { name });
      default:
        return null;
    }
  })();

  const resultBanner = (() => {
    if (!isGameEnd) return null;
    if (state.winnerIdx === 0) return t('result.you');
    if (state.winnerIdx > 0) return t('result.cpu', { name: t('header.cpu', { idx: String(state.winnerIdx) }) });
    return t('result.none');
  })();

  return (
    <GamePageShell
      title={tc('nav.snap')}
      gameThemeBg={gameTheme.snap.bg}
      phaseName={isGameEnd ? t('phase.gameEnd') : t('phase.play')}
      isHumanTurn={state.isHumanTurn}
      gamePath="/snap"
      gameEndFlag={!!state.gameEndFlag}
      winShow={isGameEnd && state.winnerIdx === 0}
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
          <div className="flex-1 overflow-y-auto pt-3 px-4 lg:px-8">
            {/* **トリガーが動くことが規則そのもの。** 先に出す。 */}
            <div
              className="mb-3 rounded bg-black/30 px-3 py-2 text-ds-text-primary text-center"
              data-testid="sp-rule"
              data-tutorial="sp-rule"
            >
              {t('header.rule')}
            </div>

            <div className="text-center mb-3" data-testid="sp-pile" data-tutorial="sp-pile">
              <div className="text-ds-text-muted text-sm mb-1">
                {state.centerPileSize > 0
                  ? t('header.pile', { n: String(state.centerPileSize) })
                  : t('header.pileEmpty')}
              </div>
              <div className="flex justify-center">
                {state.topCard && <CardImage card={state.topCard} width={cardWidth} />}
              </div>
              {/* **成立しているかは一目で分かる必要がある。** 反射ゲームなので。 */}
              {state.snapAvailable && (
                <div className="mt-2 text-xl font-semibold text-ds-warning" role="status" data-testid="sp-available">
                  {t('header.available')}
                </div>
              )}
            </div>

            <div className="flex flex-wrap justify-center gap-2 mb-4" data-tutorial="sp-seats">
              {state.players.map((p) => (
                <div
                  key={p.id}
                  className="rounded bg-black/30 px-3 py-2 text-sm text-ds-text-muted"
                  data-testid={`sp-seat-${p.id.toString()}`}
                >
                  <span className="text-ds-text-primary">
                    {p.isHuman ? t('header.you') : t('header.cpu', { idx: String(p.id) })}
                  </span>
                  {p.id === state.currentTurnIdx && !isGameEnd && (
                    <span className="ml-1 text-ds-accent">{t('header.yourTurn')}</span>
                  )}
                  {': '}
                  {t('header.stock', { n: String(p.stockSize) })}
                </div>
              ))}
            </div>

            {/* **直近に何が起きたかを出す。** 盤面だけでは誰が取ったのか読めない。 */}
            {eventLine && (
              <div className="text-center mb-3 text-ds-text-muted text-sm" role="status" data-testid="sp-event">
                {eventLine}
              </div>
            )}

            {resultBanner && (
              <div
                className="text-center text-xl my-4 text-ds-accent font-semibold"
                role="status"
                data-testid="sp-result"
              >
                {resultBanner}
              </div>
            )}

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />
            <ErrorAlert message={error} onRetry={retry} />

            {hintEnabled && hint && (
              <div className="mt-3 flex justify-center">
                <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />
              </div>
            )}

            <div className="mt-4 flex flex-wrap gap-2 items-center" data-tutorial="sp-actions">
              {!isGameEnd && (
                <>
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={handleStep}
                    disabled={loading || !state.isHumanTurn}
                    data-testid="sp-step-btn"
                  >
                    {t('actions.step')}
                  </button>
                  {/* **宣言はいつでも押せる。** 成立していなければペナルティ——
                      それがこのゲームの賭けなので、押せなくはしない。 */}
                  <button
                    type="button"
                    className={btnWarning}
                    onClick={handleSnap}
                    disabled={loading}
                    data-testid="sp-snap-btn"
                  >
                    {t('actions.snap')}
                  </button>
                </>
              )}
              <button
                type="button"
                className={btnPrimary}
                onClick={() => requestConfirm(handleReset)}
                disabled={loading}
              >
                {t('actions.reset')}
              </button>
              {!isGameEnd && (
                <button type="button" className={btnDanger} onClick={handleGiveUp} disabled={loading}>
                  {t('actions.giveUp')}
                </button>
              )}
            </div>

            <SettingsPanel
              title={tc('settings.title')}
              groups={[
                {
                  items: [
                    {
                      type: 'select',
                      id: 'snap-players',
                      label: t('actions.players'),
                      value: String(playerCnt),
                      options: [2, 3, 4].map((n) => ({ value: String(n), label: String(n) })),
                      onSelect: (v: string) => setPlayerCnt(Number(v)),
                      testId: 'sp-players-select',
                    },
                    {
                      type: 'select',
                      id: 'snap-difficulty',
                      label: t('actions.difficulty'),
                      value: String(difficulty),
                      options: [
                        { value: '0', label: t('actions.easy') },
                        { value: '1', label: t('actions.normal') },
                        { value: '2', label: t('actions.hard') },
                      ],
                      onSelect: (v: string) => setDifficulty(Number(v)),
                      testId: 'sp-difficulty-select',
                    },
                    hintCheckboxItem(tc, hintEnabled, setHintEnabled),
                  ],
                },
              ]}
            />
          </div>

          <ActionLogSection
            isEndPhase={isGameEnd}
            actionLog={actionLog}
            showActionLog={showActionLog}
            hideActionLog={hideActionLog}
          />
        </>
      )}
    </GamePageShell>
  );
}

/** Snap page wrapped with TutorialProvider. */
export const SnapPage = withTutorial(SnapPageContent, 'snap', SNAP_TUTORIAL_STEPS);
