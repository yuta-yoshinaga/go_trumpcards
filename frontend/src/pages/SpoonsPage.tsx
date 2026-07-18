import { useEffect, useMemo, useRef, useState } from 'react';
import { spoonsApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSound } from '../providers/SoundProvider';
import { btnSuccess, btnWarning } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { SpoonsResponse } from '../types/card';
import { SpoonsPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseSpoonsCommand, SPOONS_HELP } from '../utils/cli/commands/spoonsCommands';
import { formatSpoonsState } from '../utils/cli/formatters/spoonsFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { computeSpoonsRankGroups } from '../utils/spoonsRankGroups';

/**
 * Ring color classes for same-rank hand groups, indexed by group color index.
 * Purely a visual aid; assignment is deterministic (by ascending rank) so the
 * same pair always keeps the same color across passes.
 */
const SPOONS_GROUP_RING_CLASSES = [
  'ring-2 ring-ds-info',
  'ring-2 ring-ds-warning',
  'ring-2 ring-ds-success',
  'ring-2 ring-ds-accent',
] as const;

/** CPU difficulty options for the Spoons settings panel. */
const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'easy' },
  { value: 1, label: 'normal' },
  { value: 2, label: 'hard' },
];

/** Spoons tutorial step definitions. */
const SPOONS_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="spoons-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="spoons-players"]',
    messageKey: 'tutorial.players',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="spoons-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="spoons-grab"]',
    messageKey: 'tutorial.grab',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="spoons-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const SPOONS_PHASE_KEYS: Readonly<Record<number, string>> = {
  [SpoonsPhase.PASS]: 'pass',
  [SpoonsPhase.GRAB]: 'grab',
  [SpoonsPhase.ROUND_END]: 'roundEnd',
  [SpoonsPhase.GAME_END]: 'gameEnd',
};

/** Render the S-P-O-O-N-S letter progress for a given count (0–6). */
function lettersText(letters: number): string {
  const word = 'SPOONS';
  const shown = word.slice(0, Math.max(0, Math.min(letters, word.length))).split('');
  return word
    .split('')
    .map((ch, i) => (i < shown.length ? ch : '·'))
    .join(' ');
}

/** Renders the Spoons game page: a 4-player pass-and-grab speed game. */
export const SpoonsPage = withTutorial(SpoonsPageContent, 'spoons', SPOONS_TUTORIAL_STEPS);

/** Inner content of the Spoons page, wrapped by TutorialProvider. */
function SpoonsPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('spoons');
  const { state, loading, error, exec, retry } = useGameApi(spoonsApi.exec);

  const [cpuDifficulty, setCpuDifficulty] = useState(1);

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    exec('reset');
  }, []);

  const handleConfigChange = (value: string) => {
    const level = Number(value);
    setCpuDifficulty(level);
    exec('reset', { config: { cpuDifficulty: level } });
  };

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('spoons');
  const cliConfig: CliGameConfig<SpoonsResponse, Parameters<typeof spoonsApi.exec>> = useMemo(
    () => ({
      gameName: 'spoons',
      parseCommand: parseSpoonsCommand,
      formatResponse: formatSpoonsState,
      helpText: SPOONS_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  const phaseNames = usePhaseNames('spoons', SPOONS_PHASE_KEYS);

  // Chime the instant the grab window opens (false→true) — a static "grab now!"
  // text alone was easy to miss in this reflex game.
  const prevGrabOpenRef = useRef(false);
  useEffect(() => {
    const open = state?.grabWindowOpen ?? false;
    if (open && !prevGrabOpenRef.current) {
      playSound('turnTick', { pitchVariation: 0.1 });
    }
    prevGrabOpenRef.current = open;
  }, [state?.grabWindowOpen, playSound]);

  if (!state)
    return <GameSkeleton gameKey="spoons" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 4 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isPassPhase = state.phase === SpoonsPhase.PASS;
  const isRoundEnd = state.phase === SpoonsPhase.ROUND_END;
  const isGameEnd = state.phase === SpoonsPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = state.isHumanTurn;
  const canPass = isPassPhase && isHumanTurn && !state.grabWindowOpen && !isGameEnd;
  const humanWon = isGameEnd && state.winnerIdx === 0;

  const playerLabel = (id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));

  // Build the spoon-icon row: one available icon per remaining spoon, plus one
  // grayed-out icon per player who already grabbed (labeled with the grabber).
  // The full-row total = remaining + grabbed = the number of spoons in play.
  const grabbers = state.players
    .map((p, idx) => ({ idx, isHuman: p.isHuman, hasSpoon: p.hasSpoon }))
    .filter((g) => g.hasSpoon);
  const spoonsInPlay = state.spoonsRemaining + grabbers.length;
  // motion-safe pulse only while spoons can still be grabbed, echoing the grab
  // button's urgency cue; respects prefers-reduced-motion via motion-safe:.
  const spoonPulse = state.grabWindowOpen && !isGameEnd ? ' motion-safe:animate-pulse' : '';

  const handleManualReset = () => {
    hideActionLog();
    exec('reset', { config: { cpuDifficulty } });
  };

  return (
    <GamePageShell
      title={tc('nav.spoons')}
      gameThemeBg={gameTheme.spoons.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn && !isGameEnd}
      gamePath="/spoons"
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
          <SettingsPanel
            title={t('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select',
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label}`),
                    })),
                    onSelect: handleConfigChange,
                  },
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="spoons-info">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span>{t('drawPile', { count: state.drawPileSize })}</span>
            </div>

            {/* Spoons on the table, visualized as icons (grabbed ones grayed out). */}
            <div className="flex flex-wrap items-center justify-center gap-x-3 gap-y-1 mb-3">
              <span className="text-ds-text-muted text-sm">{t('spoonsLabel')}</span>
              {spoonsInPlay > 0 ? (
                <div
                  className="flex flex-wrap items-center gap-1.5"
                  data-testid="spoons-icon-row"
                  role="img"
                  aria-label={t('spoonsRemaining', { count: state.spoonsRemaining })}
                >
                  {Array.from({ length: state.spoonsRemaining }, (_, i) => (
                    <span
                      key={`spoon-avail-${i}`}
                      data-testid="spoons-icon-available"
                      aria-hidden="true"
                      className={`text-lg leading-none${spoonPulse}`}
                    >
                      🥄
                    </span>
                  ))}
                  {grabbers.map((g) => (
                    <span
                      key={`spoon-grabbed-${g.idx}`}
                      data-testid="spoons-icon-grabbed"
                      className="inline-flex items-center gap-0.5 text-ds-text-muted text-xs"
                    >
                      <span aria-hidden="true" className="text-lg leading-none opacity-40 grayscale">
                        🥄
                      </span>
                      {t('spoonGrabbedBy', { name: playerLabel(g.idx, g.isHuman) })}
                    </span>
                  ))}
                </div>
              ) : (
                <span className="text-ds-text-muted text-sm" data-testid="spoons-icon-row">
                  {t('spoonsRemaining', { count: state.spoonsRemaining })}
                </span>
              )}
            </div>

            {/* Players */}
            <div className="mb-2 p-2 rounded bg-black/30" data-tutorial="spoons-players">
              <div className="mb-1 text-ds-text-primary text-sm">{t('playersTitle')}</div>
              {state.players.map((p, idx) => {
                const badges: string[] = [];
                if (p.eliminated) badges.push(t('badge.eliminated'));
                if (p.hasSpoon) badges.push(t('badge.hasSpoon'));
                if (idx === state.feederIdx) badges.push(t('badge.feeder'));
                return (
                  <div
                    key={p.name + idx}
                    className={`text-sm py-0.5 ${idx === state.currentPlayerIdx ? 'text-ds-warning' : 'text-ds-text-muted'} ${p.isHuman ? 'font-semibold' : ''}`}
                  >
                    {playerLabel(idx, p.isHuman)} — {t('lettersLabel')}: {lettersText(p.letters)}
                    {badges.length > 0 ? ` · [${badges.join(', ')}]` : ''}
                  </div>
                );
              })}
            </div>

            {/* Round result */}
            {(isRoundEnd || isGameEnd) && state.roundLoserIdx >= 0 && (
              <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                <div>
                  {t('roundResult.loser', {
                    name: playerLabel(state.roundLoserIdx, state.roundLoserIdx === 0),
                  })}
                </div>
              </div>
            )}

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Footer */}
          <GameFooter className={`${gameTheme.spoons.footer} px-4 py-2.5`}>
            {humanPlayer && humanPlayer.hand.length > 0 ? (
              <div className="mb-2" data-tutorial="spoons-hand">
                <div className="text-ds-text-muted text-xs mb-1">
                  {t('handLabel')}
                  {canPass ? ` — ${t('passNotice')}` : ''}
                </div>
                <div className="flex gap-1 flex-wrap">
                  {(() => {
                    const rankGroups = computeSpoonsRankGroups(humanPlayer.hand);
                    return humanPlayer.hand.map((c, i) => {
                      const group = rankGroups[i];
                      const { colorIndex } = group;
                      const isGrouped = colorIndex !== null;
                      const isReach = isGrouped && group.count >= 3;
                      // Ring only for cards in a same-rank group of 2+; singletons stay neutral.
                      const ringClass = isGrouped
                        ? `${SPOONS_GROUP_RING_CLASSES[colorIndex % SPOONS_GROUP_RING_CLASSES.length]}${isReach ? ' motion-safe:animate-pulse' : ''}`
                        : '';
                      const card = (
                        <CardImage
                          key={`hand-${c.design}-${c.value}-${i}`}
                          card={c}
                          width={cardWidth}
                          className={ringClass}
                        />
                      );
                      const groupProps = {
                        'data-rank-group': isGrouped ? String(group.colorIndex) : 'none',
                        'data-rank-reach': isReach ? 'true' : 'false',
                      };
                      return canPass ? (
                        <button
                          type="button"
                          key={`pass-${c.design}-${c.value}-${i}`}
                          onClick={() => exec('pass', { cardIndex: i })}
                          disabled={loading}
                          className="p-0 bg-transparent border-0 cursor-pointer disabled:cursor-not-allowed"
                          aria-label={t('passCardAria', { card: cardAlt(c) })}
                          data-testid={`spoons-pass-${i}`}
                          {...groupProps}
                        >
                          {card}
                        </button>
                      ) : (
                        <span
                          key={`hand-wrap-${c.design}-${c.value}-${i}`}
                          data-testid={`spoons-hand-${i}`}
                          {...groupProps}
                        >
                          {card}
                        </span>
                      );
                    });
                  })()}
                </div>
              </div>
            ) : (
              <div className="text-ds-text-muted text-sm mb-2" data-tutorial="spoons-hand">
                {t('handLabel')}
              </div>
            )}

            <ErrorAlert message={error} onRetry={retry} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="spoons-grab">
              {state.grabWindowOpen && !isGameEnd && (
                <>
                  <span
                    className="text-ds-warning text-sm font-semibold mr-1"
                    role="alert"
                    data-testid="spoons-grab-notice"
                  >
                    {t('grabNotice')}
                  </span>
                  <button
                    type="button"
                    className={`${btnWarning} motion-safe:animate-pulse`}
                    onClick={() => exec('grab')}
                    disabled={loading}
                    data-testid="spoons-grab-button"
                  >
                    {t('grabButton')}
                  </button>
                </>
              )}

              {isRoundEnd && !isGameEnd && (
                <button type="button" className={btnSuccess} onClick={() => exec('next')} disabled={loading}>
                  {t('nextRound')}
                </button>
              )}

              {isGameEnd && (
                <span className="text-ds-text-primary text-sm font-semibold mr-1">
                  {humanWon ? t('win') : t('lose', { name: playerLabel(state.winnerIdx, state.winnerIdx === 0) })}
                </span>
              )}

              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="spoons-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
