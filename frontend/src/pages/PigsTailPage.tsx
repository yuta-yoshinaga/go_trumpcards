import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { pigtailApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { ActionShortcutsPanel } from '../components/ActionShortcutsPanel';
import { CircularDeck } from '../components/CircularDeck';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { HintTooltip } from '../components/hint/HintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { type ActionBinding, useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { usePhaseNames } from '../hooks/usePhaseNames';
import i18n from '../i18n';
import { badgeErrorColors } from '../styles/badgeStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, PigsTailResponse } from '../types/card';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt, isSuitDesign, suitSymbol } from '../utils/cardAlt';
import { valueName } from '../utils/cardUtils';
import { parsePigtailCommand, pigtailHelp } from '../utils/cli/commands/pigtailCommands';
import { formatPigtailState } from '../utils/cli/formatters/pigtailFormatter';
import { hintLocalCommand } from '../utils/cli/hintText';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';

/** Render a center-pile card as suit symbol + rank (e.g. "♠A"), so the rank is visible. */
function centerCardLabel(card: Card): string {
  return `${isSuitDesign(card.design) ? suitSymbol(card.design) : '?'}${valueName(card.value)}`;
}

/** Default participant count (1 human + 3 CPU). Mirrors the domain default. */
const PIGTAIL_DEFAULT_PLAYER_COUNT = 4;

/** Selectable participant counts (1 human + CPUs). Mirrors the domain 2..6 range. */
const PIGTAIL_PLAYER_COUNT_OPTIONS = [2, 3, 4, 5, 6] as const;

const PT_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="pt-circle-area"]',
    messageKey: 'tutorial.circleArea',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pt-center-area"]',
    messageKey: 'tutorial.centerArea',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pt-draw-button"]',
    messageKey: 'tutorial.drawButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pt-player-area"]',
    messageKey: 'tutorial.playerArea',
    placement: 'top',
    advanceOn: 'next',
  },
];

const PIGTAIL_PHASE_PLAY = 0;
const PIGTAIL_PHASE_END = 1;

const PIGTAIL_PHASE_KEYS: Readonly<Record<number, string>> = {
  [PIGTAIL_PHASE_PLAY]: 'play',
  [PIGTAIL_PHASE_END]: 'end',
};

/** Renders the Pig's Tail game page. */
export const PigsTailPage = withTutorial(PigsTailPageContent, 'pigtail', PT_TUTORIAL_STEPS);
/** Inner content of the Pig's Tail page. */
function PigsTailPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('pigtail');
  const { state, loading, exec: execApi } = useGameApi(pigtailApi.exec);
  const [playerCount, setPlayerCount] = useState<number>(PIGTAIL_DEFAULT_PLAYER_COUNT);
  const handleDraw = useCallback(() => execApi('draw'), [execApi]);
  const handleReset = useCallback(() => execApi('reset', undefined, playerCount), [execApi, playerCount]);
  const { hint, hintEnabled, setHintEnabled } = useGameHint('pigtail', state);

  useMountReset(execApi);

  const phaseNames = usePhaseNames('pigtail', PIGTAIL_PHASE_KEYS);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('pigtail');
  const canDraw =
    !!state &&
    !loading &&
    !cliEnabled &&
    actionLog === null &&
    !state.gameEndFlag &&
    state.players[state.currentTurn]?.isHuman === true;
  const actionBindings = useMemo<ActionBinding[]>(
    () => [{ key: 'd', action: handleDraw, label: 'draw', enabled: canDraw }],
    [handleDraw, canDraw],
  );
  useActionKeyboardNav({ bindings: actionBindings, enabled: canDraw });
  type PtArgs = Parameters<typeof pigtailApi.exec>;
  // pigtailHelp() reads i18n internally, so depend on i18n.language to
  // re-localize the CLI help after a runtime language switch.
  // biome-ignore lint/correctness/useExhaustiveDependencies: i18n.language drives help re-localization
  const cliConfig: CliGameConfig<PigsTailResponse, PtArgs> = useMemo(
    () => ({
      gameName: 'pigtail',
      parseCommand: parsePigtailCommand,
      formatResponse: formatPigtailState,
      helpText: pigtailHelp(),
      localCommand: hintLocalCommand(hint),
    }),
    [i18n.language, hint],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  // Penalty screen-flash: fire whenever lastPenalty transitions to true on a
  // new lastDrawCard. We key on centerCount+circleCount to detect fresh draws,
  // not just spurious re-renders.
  const [penaltyFlash, setPenaltyFlash] = useState(0);
  const prevDrawSigRef = useRef<string>('');
  useEffect(() => {
    if (!state) return;
    const sig = `${state.circleCount}-${state.centerCount}-${state.lastDrawCard ? `${state.lastDrawCard.design}${state.lastDrawCard.value}` : 'none'}`;
    if (sig !== prevDrawSigRef.current) {
      if (state.lastPenalty) {
        setPenaltyFlash(Date.now());
      }
    }
    prevDrawSigRef.current = sig;
  }, [state]);
  useEffect(() => {
    if (penaltyFlash === 0) return;
    const id = window.setTimeout(() => setPenaltyFlash(0), 600);
    return () => window.clearTimeout(id);
  }, [penaltyFlash]);

  if (!state)
    return (
      <>
        <div role="status" aria-live="polite" className="sr-only" data-testid="pigtail-action-announcement" />
        <GameSkeleton gameKey="pigtail" layout={{ kind: 'centered', rows: [2], shape: 'circle', bars: 4 }} />
      </>
    );

  const announcement = state.humanAction
    ? t('announcement.summary', {
        human: t('announcement.action', {
          player: playerName(state.humanAction.drawPlayerIdx, true),
          card: state.humanAction.drawnCard ? cardAlt(state.humanAction.drawnCard) : '?',
          result: state.humanAction.penaltyFlag
            ? t('announcement.penalty', { count: state.humanAction.penaltyCount })
            : t('announcement.safe'),
        }),
        cpu:
          state.cpuActions.length > 0
            ? t('announcement.cpuPrefix', {
                cpu: state.cpuActions
                  .map((action) =>
                    t('announcement.action', {
                      player: playerName(action.drawPlayerIdx, false),
                      card: action.drawnCard ? cardAlt(action.drawnCard) : '?',
                      result: action.penaltyFlag
                        ? t('announcement.penalty', { count: action.penaltyCount })
                        : t('announcement.safe'),
                    }),
                  )
                  .join(t('listSeparator')),
              })
            : '',
      })
    : '';

  const isGameEnd = state.gameEndFlag;
  const isHumanTurn = !isGameEnd && state.players[state.currentTurn]?.isHuman === true;
  const currentPhaseName = isGameEnd ? phaseNames[PIGTAIL_PHASE_END] : phaseNames[PIGTAIL_PHASE_PLAY];
  const loserIsHuman = isGameEnd && state.loserIdx >= 0 && state.players[state.loserIdx]?.isHuman === true;
  // Signature of the current draw; changing it remounts the reveal card so the
  // flip animation re-fires on every new draw (each draw changes circleCount).
  const drawSig = `${state.circleCount}-${state.centerCount}-${state.lastDrawCard ? `${state.lastDrawCard.design}${state.lastDrawCard.value}` : 'none'}`;

  return (
    <GamePageShell
      title={tc('nav.pigtail')}
      gameThemeBg={gameTheme.pigtail.bg}
      phaseName={currentPhaseName ?? ''}
      gamePath="/pigtail"
      gameEndFlag={!!isGameEnd}
      winShow={isGameEnd && !loserIsHuman}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={<CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />}
    >
      <div role="status" aria-live="polite" className="sr-only" data-testid="pigtail-action-announcement">
        {announcement}
      </div>
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <SettingsPanel
            title={t('setup.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select' as const,
                    id: 'playerCount',
                    label: t('setup.playerCount'),
                    description:
                      playerCount !== state.players.length ? t('setup.playerCountAppliesAfterReset') : undefined,
                    value: playerCount,
                    options: PIGTAIL_PLAYER_COUNT_OPTIONS.map((n) => ({
                      value: n,
                      label: t('setup.playerCountUnit', { count: n }),
                    })),
                    onSelect: (v) => setPlayerCount(Number(v)),
                    testId: 'pigtail-player-count',
                  },
                ],
              },
            ]}
          />
          {penaltyFlash > 0 && (
            <div
              aria-hidden="true"
              data-testid="pigtail-penalty-flash"
              className="pointer-events-none fixed inset-0 z-40 bg-ds-error/40 motion-safe:animate-[pulse-once_0.6s_ease-out]"
            />
          )}
          <div className="flex-1 overflow-y-auto px-4 py-2 space-y-3">
            {/* Circle & Center area */}
            <div className="flex flex-col items-center gap-3" data-tutorial="pt-circle-area">
              <div className="text-xs text-ds-text-muted">
                {t('label.circle')}: {state.circleCount}
              </div>
              <CircularDeck
                count={state.circleCount}
                cardWidth={32}
                diameter={160}
                onDrawCard={handleDraw}
                disabled={loading || isGameEnd || !isHumanTurn}
                drawAriaLabel={t('button.draw')}
              />
              <div className="text-center" data-tutorial="pt-center-area">
                <div className="text-xs text-ds-text-muted mb-1">
                  {t('label.center')} ({state.centerCount})
                </div>
                <div
                  className="w-16 h-16 rounded-lg bg-ds-warning/60 border-2 border-ds-warning/40 flex items-center justify-center text-lg font-bold text-white"
                  data-testid="pt-center-top"
                >
                  {state.centerTop ? centerCardLabel(state.centerTop) : '-'}
                </div>
                {state.centerHistory.length > 0 && (
                  <div className="mt-1.5">
                    <div className="text-[10px] text-ds-text-muted mb-0.5">{t('label.recentCenter')}</div>
                    <div className="flex items-center justify-center gap-1" data-testid="pt-center-history">
                      {state.centerHistory.map((c, i) => (
                        <span
                          key={`${c.design}-${c.value}-${i}`}
                          className={`px-1 py-0.5 rounded text-xs font-semibold ${
                            i === state.centerHistory.length - 1
                              ? 'bg-ds-warning/50 text-white'
                              : 'bg-black/20 text-ds-text-muted'
                          }`}
                        >
                          {centerCardLabel(c)}
                        </span>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            </div>

            {/* Drawn-card reveal: flips the just-drawn card face-up onto the board
                so the player can see what was drawn, with a red highlight on penalty. */}
            {state.lastDrawCard && (
              <div className="flex flex-col items-center gap-1" data-testid="pt-draw-reveal">
                <div key={drawSig} className="motion-safe:animate-flipIn">
                  <div
                    className={
                      state.lastPenalty ? 'rounded-lg ring-2 ring-ds-error shadow-lg shadow-ds-error/50' : undefined
                    }
                  >
                    <AnimatedCard card={state.lastDrawCard} width={48} silent />
                  </div>
                </div>
                <div className={`text-sm font-medium ${state.lastPenalty ? 'text-ds-error-text' : 'text-ds-success'}`}>
                  {state.lastPenalty ? t('label.penalty') : t('label.safe')}
                </div>
              </div>
            )}

            {/* Human action */}
            {state.humanAction && (
              <div
                data-testid="pt-human-action"
                className={`text-xs px-2 py-1 rounded ${state.humanAction.penaltyFlag ? badgeErrorColors : 'bg-black/30 text-ds-text-muted'}`}
              >
                {playerName(state.humanAction.drawPlayerIdx, true)}:{' '}
                {state.humanAction.drawnCard
                  ? (isSuitDesign(state.humanAction.drawnCard.design)
                      ? suitSymbol(state.humanAction.drawnCard.design)
                      : '?') + state.humanAction.drawnCard.value
                  : '?'}
                {state.humanAction.penaltyFlag
                  ? ` — ${t('label.penalty')} (+${state.humanAction.penaltyCount})`
                  : ` — ${t('label.safe')}`}
              </div>
            )}

            {/* CPU actions */}
            {state.cpuActions.length > 0 && (
              <div className="space-y-1">
                {state.cpuActions.map((action, i) => (
                  <div
                    key={i}
                    className={`text-xs px-2 py-1 rounded ${action.penaltyFlag ? badgeErrorColors : 'bg-black/30 text-ds-text-muted'}`}
                  >
                    {playerName(action.drawPlayerIdx, false)}:{' '}
                    {action.drawnCard
                      ? (isSuitDesign(action.drawnCard.design) ? suitSymbol(action.drawnCard.design) : '?') +
                        action.drawnCard.value
                      : '?'}
                    {action.penaltyFlag
                      ? ` — ${t('label.penalty')} (+${action.penaltyCount})`
                      : ` — ${t('label.safe')}`}
                  </div>
                ))}
              </div>
            )}

            {/* Players */}
            <div className="space-y-2" data-tutorial="pt-player-area">
              {state.players.map((player, idx) => (
                <div
                  key={player.id}
                  className={`flex items-center justify-between px-3 py-2 rounded ${
                    !isGameEnd && state.currentTurn === idx
                      ? 'bg-ds-warning/30 border border-ds-warning/50'
                      : 'bg-black/30'
                  } ${isGameEnd && state.loserIdx === idx ? 'bg-ds-error/40 border border-ds-error/50' : ''}`}
                >
                  <span className="text-ds-text-primary text-sm font-medium">
                    {playerName(player.id, player.isHuman)}
                  </span>
                  <span className="text-ds-text-primary text-sm">
                    {player.cardCount} {t('label.cards')}
                  </span>
                </div>
              ))}
            </div>

            {/* Message */}
            {state.message && (
              <GameMessageBox
                message={state.message}
                messageCode={state.messageCode}
                messageParams={state.messageParams}
              />
            )}

            {/* Action log */}
            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {hintEnabled && hint && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}

          <GameFooter className={`${gameTheme.pigtail.footer} px-4 py-2.5`}>
            <ActionShortcutsPanel bindings={actionBindings} data-testid="pigtail-kbd-shortcuts" />
            <div data-testid="pigtail-draw-guidance" className="mb-2 text-center text-xs text-ds-text-primary">
              {state.centerCount > 0 && <p>{t('drawGuidance.penalty')}</p>}
              <p>
                {state.circleCount > 0
                  ? t('drawGuidance.stock', { count: state.circleCount })
                  : t('drawGuidance.emptyStock')}
              </p>
              <p>
                {/* CPU turns run to completion inside the same request, so outside game end
                    the response is always the human's turn. */}
                {isGameEnd ? t('drawGuidance.gameEnded') : t('drawGuidance.yourTurn')}
              </p>
            </div>
            <div className="flex gap-2 justify-center items-center flex-wrap">
              <label className="flex items-center gap-1 text-ds-text-primary text-xs min-h-[44px]">
                <input
                  type="checkbox"
                  checked={hintEnabled}
                  onChange={(e) => setHintEnabled(e.target.checked)}
                  aria-label={tc('hint.toggle', { ns: 'tutorial' })}
                />
                {tc('hint.toggle', { ns: 'tutorial' })}
              </label>
              <button
                type="button"
                className="px-6 py-2 rounded-lg bg-ds-info hover:bg-ds-info text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
                onClick={handleDraw}
                disabled={loading || isGameEnd || !isHumanTurn}
                data-tutorial="pt-draw-button"
              >
                {t('button.draw')}
              </button>
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleReset}
                requestConfirm={requestConfirm}
                loading={loading}
              />
              <button
                type="button"
                className="px-4 py-2 rounded-lg bg-ds-surface-elevated hover:bg-ds-surface-elevated text-ds-text-primary text-sm transition-colors"
                onClick={showActionLog}
              >
                {tc('actionLog.view')}
              </button>
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
