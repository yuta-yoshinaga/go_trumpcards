import { useEffect, useMemo, useState } from 'react';
import { kempsApi } from '../api/gameApi';
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
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeSuccessColors, badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { KempsResponse } from '../types/card';
import { KempsPhase, KempsSignal } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { KEMPS_HELP, parseKempsCommand } from '../utils/cli/commands/kempsCommands';
import { formatKempsState } from '../utils/cli/formatters/kempsFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** CPU difficulty options for the Kemps settings panel. */
const CPU_DIFFICULTY_OPTIONS = [
  { value: 0, label: 'easy' },
  { value: 1, label: 'normal' },
  { value: 2, label: 'hard' },
];

/** Signal type options for the Kemps signal selector. */
const SIGNAL_OPTIONS = [
  { value: KempsSignal.SOUND, label: 'sound' },
  { value: KempsSignal.BLINK, label: 'blink' },
];

/** Signal-type → i18n label key, for the screen-reader "signal set" announcement. */
const SIGNAL_LABEL_BY_TYPE: Readonly<Record<number, string>> = {
  [KempsSignal.SOUND]: 'sound',
  [KempsSignal.BLINK]: 'blink',
};

/** Kemps tutorial step definitions. */
const KEMPS_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="kemps-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="kemps-field"]',
    messageKey: 'tutorial.field',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="kemps-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="kemps-signal"]',
    messageKey: 'tutorial.signal',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="kemps-declare"]',
    messageKey: 'tutorial.declare',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="kemps-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const KEMPS_PHASE_KEYS: Readonly<Record<number, string>> = {
  [KempsPhase.EXCHANGE]: 'exchange',
  [KempsPhase.DECLARE]: 'declare',
  [KempsPhase.ROUND_END]: 'roundEnd',
  [KempsPhase.GAME_END]: 'gameEnd',
};

/** Renders the Kemps game page: a 4-player, 2-team matching game with secret signals. */
export const KempsPage = withTutorial(KempsPageContent, 'kemps', KEMPS_TUTORIAL_STEPS);

/** Inner content of the Kemps page, wrapped by TutorialProvider. */
function KempsPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('kemps');
  const { state, loading, error, exec, retry } = useGameApi(kempsApi.exec);

  const [cpuDifficulty, setCpuDifficulty] = useState(1);
  const [selectedHand, setSelectedHand] = useState<number | null>(null);

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    exec('reset');
  }, []);

  const handleDifficultyChange = (value: string) => {
    const level = Number(value);
    setCpuDifficulty(level);
    setSelectedHand(null);
    exec('reset', { config: { cpuDifficulty: level } });
  };

  const handleSignalChange = (value: string) => {
    exec('signal', { signalType: Number(value) });
  };

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('kemps');
  const cliConfig: CliGameConfig<KempsResponse, Parameters<typeof kempsApi.exec>> = useMemo(
    () => ({
      gameName: 'kemps',
      parseCommand: parseKempsCommand,
      formatResponse: formatKempsState,
      helpText: KEMPS_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const { cardWidth } = useCardDimensions();
  const phaseNames = usePhaseNames('kemps', KEMPS_PHASE_KEYS);

  // **フックは早期 return より上。**`if (!state)` の下に置くと、初回レンダー
  // だけフック数が変わってページが骨組みのまま固まる (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('kemps', state);

  if (!state)
    return <GameSkeleton gameKey="kemps" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 4 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  // The backend flags four-of-a-kind for the human only; when set, the Kemps
  // call button is emphasised so the player does not miss the declare chance
  // while focused on field swaps (#3553). The flag clears the moment a swap
  // breaks the quad, so the emphasis follows the current hand automatically.
  const humanHasFour = !!humanPlayer?.hasFourOfAKind;
  const isExchange = state.phase === KempsPhase.EXCHANGE;
  const isDeclare = state.phase === KempsPhase.DECLARE;
  const isRoundEnd = state.phase === KempsPhase.ROUND_END;
  const isGameEnd = state.phase === KempsPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = state.isHumanTurn;
  const canSwap = isExchange && isHumanTurn && !isGameEnd;
  // Rank of the currently selected hand card; field cards sharing this rank are
  // highlighted to help the human spot four-of-a-kind swaps (#3554). Null unless a
  // swap is possible and a hand card is selected, so nothing is highlighted otherwise.
  const selectedRank =
    canSwap && selectedHand !== null && humanPlayer ? (humanPlayer.hand[selectedHand]?.value ?? null) : null;
  const humanTeam = humanPlayer ? humanPlayer.team : 0;
  const humanWon = isGameEnd && state.winnerTeam === humanTeam;

  const teamLabel = (team: number): string => (team === 0 ? t('teamA') : t('teamB'));
  const playerLabel = (id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));

  const handleFieldClick = (fieldIndex: number) => {
    if (!canSwap || selectedHand === null) return;
    const handIndex = selectedHand;
    setSelectedHand(null);
    exec('swap', { handIndex, fieldIndex });
  };

  const handleManualReset = () => {
    hideActionLog();
    setSelectedHand(null);
    exec('reset', { config: { cpuDifficulty } });
  };

  // Opponent seats the human may target with a Counter-Kemps call.
  const opponentSeats = state.players.map((p, idx) => ({ p, idx })).filter(({ p }) => p.team !== humanTeam);

  return (
    <GamePageShell
      title={tc('nav.kemps')}
      gameThemeBg={gameTheme.kemps.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn && !isGameEnd}
      gamePath="/kemps"
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
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="kemps-info">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">
                {teamLabel(0)}: {state.teamScores[0] ?? 0}
              </span>
              <span className="mr-4">
                {teamLabel(1)}: {state.teamScores[1] ?? 0}
              </span>
              <span>{t('targetScore', { count: state.targetScore })}</span>
            </div>

            {/* Players */}
            <div className="mb-2 p-2 rounded bg-black/30" data-tutorial="kemps-players">
              <div className="mb-1 text-ds-text-primary text-sm">{t('playersTitle')}</div>
              {state.players.map((p, idx) => (
                <div
                  key={p.name + idx}
                  className={`text-sm py-0.5 ${idx === state.currentPlayerIdx ? 'text-ds-warning' : 'text-ds-text-muted'} ${p.isHuman ? 'font-semibold' : ''}`}
                >
                  {playerLabel(idx, p.isHuman)} — {teamLabel(p.team)}
                  {p.isHuman && p.hasFourOfAKind ? ` · [${t('badge.fourOfAKind')}]` : ''}
                </div>
              ))}
            </div>

            {/* Shared field */}
            <div className="mb-2 p-2 rounded bg-black/30" data-tutorial="kemps-field">
              <div className="mb-1 text-ds-text-primary text-sm">
                {t('fieldLabel')}
                {canSwap && selectedHand !== null ? ` — ${t('fieldNotice')}` : ''}
              </div>
              {/* Describes the swap-pending state for each field button (sr-only). */}
              <span id="kemps-swap-pending" className="sr-only">
                {t('swapPending')}
              </span>
              <div className="flex gap-1 flex-wrap">
                {state.field.map((c, i) => {
                  const card = <CardImage key={`field-${c.design}-${c.value}-${i}`} card={c} width={cardWidth} />;
                  const rankMatch = selectedRank !== null && c.value === selectedRank;
                  return canSwap && selectedHand !== null ? (
                    <button
                      type="button"
                      key={`field-btn-${c.design}-${c.value}-${i}`}
                      onClick={() => handleFieldClick(i)}
                      disabled={loading}
                      className={`p-0 bg-transparent border-0 cursor-pointer disabled:cursor-not-allowed rounded ${rankMatch ? 'ring-2 ring-ds-success' : 'ring-0'}`}
                      aria-label={
                        rankMatch
                          ? t('swapCardRankMatchAria', { card: cardAlt(c) })
                          : t('swapCardAria', { card: cardAlt(c) })
                      }
                      aria-describedby="kemps-swap-pending"
                      data-testid={`kemps-field-${i}`}
                      data-rank-match={rankMatch ? 'true' : 'false'}
                    >
                      {card}
                    </button>
                  ) : (
                    card
                  );
                })}
              </div>
            </div>

            {/* Signal cues (human only) */}
            {isDeclare && state.partnerSignaling && (
              <div
                className={`my-2 p-2 rounded text-sm font-semibold ${badgeSuccessColors}`}
                role="status"
                aria-live="polite"
                data-testid="kemps-partner-signaling"
              >
                {t('partnerSignaling')}
              </div>
            )}
            {isDeclare && state.opponentSignaling && (
              <div
                className={`my-2 p-2 rounded text-sm ${badgeWarningColors}`}
                role="status"
                aria-live="polite"
                data-testid="kemps-opponent-signaling"
              >
                {t('opponentSignaling')}
              </div>
            )}

            {/* Four-of-a-kind readiness — announced (and shown during EXCHANGE)
                so the human notices the quad before the declare window. */}
            {humanHasFour && !isGameEnd && (
              <div
                className={`my-2 p-2 rounded text-sm font-semibold ${badgeSuccessColors}`}
                role="status"
                aria-live="polite"
                data-testid="kemps-four-ready"
              >
                {t('fourOfAKindReady')}
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
          <GameFooter className={`${gameTheme.kemps.footer} px-4 py-2.5`}>
            {humanPlayer && humanPlayer.hand.length > 0 ? (
              <div className="mb-2" data-tutorial="kemps-hand">
                <div className="text-ds-text-muted text-xs mb-1">
                  {t('handLabel')}
                  {canSwap ? ` — ${t('handNotice')}` : ''}
                </div>
                <div className="flex gap-1 flex-wrap">
                  {humanPlayer.hand.map((c, i) => {
                    const selected = selectedHand === i;
                    const card = <CardImage key={`hand-${c.design}-${c.value}-${i}`} card={c} width={cardWidth} />;
                    return canSwap ? (
                      <button
                        type="button"
                        key={`hand-btn-${c.design}-${c.value}-${i}`}
                        onClick={() => setSelectedHand(selected ? null : i)}
                        disabled={loading}
                        aria-pressed={selected}
                        className={`p-0 bg-transparent border-2 rounded cursor-pointer disabled:cursor-not-allowed ${selected ? 'border-ds-warning' : 'border-transparent'}`}
                        aria-label={t('selectHandCardAria', { card: cardAlt(c) })}
                        data-testid={`kemps-hand-${i}`}
                      >
                        {card}
                      </button>
                    ) : (
                      card
                    );
                  })}
                </div>
              </div>
            ) : (
              <div className="text-ds-text-muted text-sm mb-2" data-tutorial="kemps-hand">
                {t('handLabel')}
              </div>
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {/* Signal selector */}
            <div className="flex flex-wrap gap-2 items-center mb-2" data-tutorial="kemps-signal">
              <span className="text-ds-text-muted text-xs mr-1">{t('signalLabel')}</span>
              {SIGNAL_OPTIONS.map((o) => (
                <button
                  type="button"
                  key={`signal-${o.value}`}
                  onClick={() => handleSignalChange(String(o.value))}
                  disabled={loading || isGameEnd}
                  aria-pressed={state.signalType === o.value}
                  className={`px-3 py-1.5 text-sm rounded min-h-[44px] ${state.signalType === o.value ? 'bg-ds-accent text-ds-text-on-accent' : 'bg-black/30 text-ds-text-muted'}`}
                >
                  {t(`signal.${o.label}`)}
                </button>
              ))}
              {/* Announce the chosen signal — the visual selection is colour-only otherwise (#2690). */}
              <span className="sr-only" role="status" aria-live="polite" data-testid="kemps-signal-sent">
                {t('signalSent', { signal: t(`signal.${SIGNAL_LABEL_BY_TYPE[state.signalType]}`) })}
              </span>
            </div>

            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="kemps-declare">
              {canSwap && (
                <button type="button" className={btnSuccess} onClick={() => exec('pass')} disabled={loading}>
                  {t('passButton')}
                </button>
              )}

              {isDeclare && !isGameEnd && (
                <>
                  <button
                    type="button"
                    className={`${btnWarning}${humanHasFour ? ' ring-2 ring-ds-success motion-safe:animate-pulse' : ''}`}
                    onClick={() => exec('kemps')}
                    disabled={loading}
                    data-testid="kemps-declare-button"
                    data-emphasized={humanHasFour ? 'true' : 'false'}
                  >
                    {t('kempsButton')}
                  </button>
                  {opponentSeats.map(({ p, idx }) => (
                    <button
                      type="button"
                      key={`counter-${idx}`}
                      className={btnPrimary}
                      onClick={() => exec('counter', { targetSeat: idx })}
                      disabled={loading}
                    >
                      {t('counterButton', { name: playerLabel(idx, p.isHuman) })}
                    </button>
                  ))}
                  <button type="button" className={btnSuccess} onClick={() => exec('pass')} disabled={loading}>
                    {t('declineButton')}
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
                  {humanWon ? t('win') : t('lose', { name: teamLabel(state.winnerTeam) })}
                </span>
              )}

              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="kemps-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
