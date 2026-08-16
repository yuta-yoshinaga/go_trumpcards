import { useEffect, useMemo, useRef, useState } from 'react';
import { pishtiApi } from '../api/gameApi';
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
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useSound } from '../providers/SoundProvider';
import { btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, PishtiPlayer, PishtiResponse } from '../types/card';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { PISHTI_HELP, parsePishtiCommand } from '../utils/cli/commands/pishtiCommands';
import { formatPishtiState } from '../utils/cli/formatters/pishtiFormatter';
import { hintCliText, isHintCommand } from '../utils/cli/hintText';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** CPU difficulty options for the Pişti settings panel. */
const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'easy' },
  { value: 1, label: 'normal' },
  { value: 2, label: 'hard' },
];

/** Player-count options for the Pişti settings panel. */
const PLAYER_COUNT_OPTIONS = [2, 3, 4];

/** Pişti tutorial step definitions. */
const PISHTI_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="pishti-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pishti-players"]',
    messageKey: 'tutorial.players',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pishti-pile"]',
    messageKey: 'tutorial.pile',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pishti-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pishti-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Maps the backend Pişti phase strings to i18n phase-label keys. */
const PISHTI_PHASE_KEYS: Readonly<Record<string, string>> = {
  play: 'phase.play',
  roundEnd: 'phase.roundEnd',
  gameEnd: 'phase.gameEnd',
};

/** Renders the Pişti game page: a 2-4 player Turkish capture (fishing) game. */
export const PishtiPage = withTutorial(PishtiPageContent, 'pishti', PISHTI_TUTORIAL_STEPS);

/** Inner content of the Pişti page, wrapped by TutorialProvider. */
function PishtiPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('pishti');
  const { state, loading, error, exec, retry } = useGameApi(pishtiApi.exec);

  const [cpuDifficulty, setCpuDifficulty] = useState(1);
  const [playerCnt, setPlayerCnt] = useState(4);

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    exec('reset');
  }, []);

  const handleDifficultyChange = (value: string) => {
    const level = Number(value);
    setCpuDifficulty(level);
    exec('reset', { config: { cpuDifficulty: level, playerCnt } });
  };

  const handlePlayerCountChange = (value: string) => {
    const count = Number(value);
    setPlayerCnt(count);
    exec('reset', { config: { cpuDifficulty, playerCnt: count } });
  };

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('pishti');

  // **フックは早期 return より上。**`if (!state)` の下に置くと、初回レンダー
  // だけフック数が変わってページが骨組みのまま固まる (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('pishti', state);
  const cliConfig: CliGameConfig<PishtiResponse, Parameters<typeof pishtiApi.exec>> = useMemo(
    () => ({
      gameName: 'pishti',
      parseCommand: parsePishtiCommand,
      formatResponse: formatPishtiState,
      helpText: PISHTI_HELP,
      localCommand: (input: string) => (isHintCommand(input) ? hintCliText(frontendHint) : null),
    }),
    [frontendHint],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();

  // Celebrate a Pişti the instant a player's bonus rises — the +10/+20 capture is
  // the game's highlight but was easy to miss amid fast CPU turns (#2692). Tracked
  // per-player (not aggregate) so two CPU +10s in one response can't fake a +20 Jack.
  const [pistiCelebration, setPistiCelebration] = useState<{ key: number; jack: boolean } | null>(null);
  const prevBonusesRef = useRef<number[] | null>(null);
  useEffect(() => {
    if (!state) return;
    const current = state.players.map((p) => p.pistiBonus);
    const prev = prevBonusesRef.current;
    prevBonusesRef.current = current;
    if (prev === null) return;
    let anyGain = false;
    let jack = false;
    if (prev.length === current.length) {
      for (let i = 0; i < current.length; i++) {
        const delta = current[i] - prev[i];
        if (delta > 0) {
          anyGain = true;
          if (delta >= 20) jack = true; // a single +20 rise is a Jack Pişti
        }
      }
    }
    if (anyGain) {
      setPistiCelebration((c) => ({ key: (c?.key ?? 0) + 1, jack }));
      playSound('chipClick', { pitchVariation: 0.1 });
    } else if (prev.length !== current.length || current.some((v, i) => v < prev[i])) {
      // A reset / next game / player-count change drops bonuses; clear the stale badge.
      setPistiCelebration(null);
    }
  }, [state, playSound]);

  if (!state)
    return <GameSkeleton gameKey="pishti" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 4 }} />;

  const phaseName = t(PISHTI_PHASE_KEYS[state.phase] ?? '');

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isGameEnd = state.phase === 'gameEnd' || state.gameEndFlag;
  const isHumanTurn = state.phase === 'play' && state.currentTurn === 0 && !isGameEnd;
  const humanWon = isGameEnd && state.winners.includes(0);

  // On the human's turn, hint which cards can capture the pile: a Jack takes the whole
  // pile (accent), and a card matching the pile-top rank captures it (success). Pure
  // client-side derivation from pileTop + the hand.
  const JACK_VALUE = 11;
  const captureRing = (c: Card): string => {
    if (!isHumanTurn) return '';
    if (c.value === JACK_VALUE) return 'ring-2 ring-ds-accent motion-safe:animate-pulse';
    if (state.pileTop && c.value === state.pileTop.value) return 'ring-2 ring-ds-success motion-safe:animate-pulse';
    return '';
  };

  const playerLabel = (id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));

  // Provisional score shown during play so players can gauge where they stand (#3560).
  // The API exposes only capturedCount + pistiBonus (not the captured cards' values),
  // so the true point-card score (A/J/2♣/10♦) is NOT derivable. What IS known:
  //   - Pişti bonuses are already locked into the final score, and
  //   - the most-cards +3 (PishtiScoreMostCards) goes to the sole captured-count leader
  //     (ties award nobody, mirroring the domain's calcFinalScore).
  // We surface only those; the note discloses that card points are still uncounted.
  const MOST_CARDS_BONUS = 3;
  const maxCaptured = Math.max(...state.players.map((p) => p.capturedCount));
  const capturedLeaders = state.players.filter((p) => p.capturedCount === maxCaptured);
  const provisionalLeaderSeat = maxCaptured > 0 && capturedLeaders.length === 1 ? capturedLeaders[0].id : -1;
  const provisionalScore = (p: PishtiPlayer): number =>
    p.pistiBonus + (p.id === provisionalLeaderSeat ? MOST_CARDS_BONUS : 0);

  const handleManualReset = () => {
    hideActionLog();
    exec('reset', { config: { cpuDifficulty, playerCnt } });
  };

  return (
    <GamePageShell
      title={tc('nav.pishti')}
      gameThemeBg={gameTheme.pishti.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn && !isGameEnd}
      gamePath="/pishti"
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
                    onSelect: handleDifficultyChange,
                  },
                  {
                    type: 'select',
                    id: 'playerCnt',
                    label: t('settings.playerCount'),
                    value: playerCnt,
                    options: PLAYER_COUNT_OPTIONS.map((n) => ({ value: n, label: String(n) })),
                    onSelect: handlePlayerCountChange,
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="pishti-info">
              <span>{t('deck', { count: state.remainingDeck })}</span>
            </div>

            {/* Players (hand / captured / Pişti bonus / current turn) */}
            <div className="mb-2 p-2 rounded bg-black/30" data-tutorial="pishti-players">
              <div className="mb-1 text-ds-text-primary text-sm">{t('playersTitle')}</div>
              {!isGameEnd && (
                <div className="mb-1 text-ds-text-muted text-xs" data-testid="pishti-provisional-note">
                  {t('provisionalNote')}
                </div>
              )}
              {state.players.map((p) => (
                <div
                  key={`player-${p.id}`}
                  className={`text-sm py-0.5 flex items-center gap-3 ${
                    p.id === state.currentTurn && !isGameEnd ? 'text-ds-warning' : 'text-ds-text-muted'
                  } ${p.isHuman ? 'font-semibold' : ''}`}
                >
                  <span>{playerLabel(p.id, p.isHuman)}</span>
                  <span>{t('captured', { count: p.capturedCount })}</span>
                  {p.pistiBonus > 0 && <span className="text-ds-accent">{t('pisti', { count: p.pistiBonus })}</span>}
                  {!isGameEnd && (
                    <span
                      className="text-ds-text-primary"
                      data-testid={`pishti-provisional-${p.id}`}
                      title={t('provisionalNote')}
                    >
                      {t('provisional', { score: provisionalScore(p) })}
                      {p.id === provisionalLeaderSeat && <span className="ml-1 text-ds-accent">★</span>}
                    </span>
                  )}
                  {isGameEnd && (
                    <span className="text-ds-text-primary">{t('finalScore', { score: p.finalScore })}</span>
                  )}
                </div>
              ))}
            </div>

            {/* Center pile */}
            <div className="relative mb-2 p-3 rounded bg-black/20 text-center" data-tutorial="pishti-pile">
              {pistiCelebration && (
                <div
                  key={pistiCelebration.key}
                  className="absolute inset-x-0 -top-2 z-10 flex justify-center motion-safe:animate-bounce pointer-events-none"
                  role="status"
                  data-testid="pishti-celebration"
                >
                  <span className="rounded-full bg-ds-accent px-3 py-0.5 text-sm font-bold text-ds-text-on-accent shadow-lg">
                    {pistiCelebration.jack ? t('pistiCelebrationJack') : t('pistiCelebration')}
                  </span>
                </div>
              )}
              <div className="text-ds-text-muted text-xs mb-1">
                {t('pile')} — {t('pileCount', { count: state.pileCount })}
              </div>
              {state.pileTop ? (
                <div className="flex justify-center">
                  <CardImage card={state.pileTop} width={cardWidth} />
                </div>
              ) : (
                <div className="text-ds-text-muted text-sm">{t('pileEmpty')}</div>
              )}
            </div>

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
          <GameFooter className={`${gameTheme.pishti.footer} px-4 py-2.5`}>
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="mb-2" data-tutorial="pishti-hand">
              <div className="text-ds-text-muted text-xs mb-1">{t('yourHand')}</div>
              <div className="flex flex-wrap gap-2">
                {humanPlayer?.cards.map((c, i) => (
                  <button
                    key={`hand-${i}`}
                    type="button"
                    onClick={() => isHumanTurn && exec('play', { handIndex: i })}
                    disabled={!isHumanTurn || loading}
                    className={`rounded transition-all ${
                      isHumanTurn ? 'cursor-pointer hover:opacity-90 hover:-translate-y-1' : 'cursor-default'
                    } ${captureRing(c)}`}
                    data-testid={`hand-card-${i}`}
                    aria-label={
                      captureRing(c)
                        ? `${t('playCardAria', { card: cardAlt(c) })} — ${t('captureHint')}`
                        : t('playCardAria', { card: cardAlt(c) })
                    }
                  >
                    <CardImage card={c} width={cardWidth} />
                  </button>
                ))}
              </div>
            </div>

            <ErrorAlert message={error} onRetry={retry} />

            {isHumanTurn && (
              <div className="text-ds-text-muted text-xs mb-2" role="status" data-testid="pishti-turn-notice">
                {t('turnNotice')}
              </div>
            )}

            <div className="flex flex-wrap gap-2 items-center">
              {isGameEnd && (
                <span className="text-ds-text-primary text-sm font-semibold mr-1">
                  {humanWon
                    ? t('win')
                    : t('lose', { name: playerLabel(state.winners[0] ?? -1, state.winners[0] === 0) })}
                </span>
              )}
              {isGameEnd && (
                <button type="button" className={btnSuccess} onClick={() => exec('next')} disabled={loading}>
                  {t('nextGame')}
                </button>
              )}

              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="pishti-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
