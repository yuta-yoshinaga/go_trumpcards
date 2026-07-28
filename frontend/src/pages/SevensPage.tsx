import { useCallback, useEffect, useMemo, useRef } from 'react';
import type { sevensApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardNavShortcutsPanel } from '../components/CardNavShortcutsPanel';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ReplaySpeedSettingsPanel } from '../components/common/ReplaySpeedSettingsPanel';
import type { SettingsGroup } from '../components/common/SettingsPanel';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { SevensBoard } from '../components/sevens/SevensBoard';
import { SevensCpuArea } from '../components/sevens/SevensCpuArea';
import { SevensHumanArea } from '../components/sevens/SevensHumanArea';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useSevensGame } from '../hooks/useSevensGame';
import { badgeSuccess, badgeWarning } from '../styles/badgeStyles';
import { btnSecondary } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { SevensResponse } from '../types/card';
import type { TutorialStep } from '../types/tutorial';
import { parseSevensCommand, SEVENS_HELP } from '../utils/cli/commands/sevensCommands';
import { formatSevensState } from '../utils/cli/formatters/sevensFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';
import { actionDesc, listJokerPlacements } from '../utils/sevensUtils';

/** Sevens tutorial step definitions. */
const SV_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="sv-board"]', messageKey: 'tutorial.board', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="sv-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="sv-play-pass"]', messageKey: 'tutorial.playPass', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="sv-settings"]', messageKey: 'tutorial.settings', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="sv-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Sevens game page with board, player areas, and joker placement. */
export const SevensPage = withTutorial(SevensPageContent, 'sevens', SV_TUTORIAL_STEPS);
/** Inner content of the Sevens page, wrapped by TutorialProvider. */
function SevensPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('sevens');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    jokerCardIdx,
    setJokerCardIdx,
    config,
    handleConfigChange,
    handleToggle,
    handleCardPlay,
    handleJokerPlace,
  } = useSevensGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('sevens', state);
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('sevens');
  const cliConfig: CliGameConfig<SevensResponse, Parameters<typeof sevensApi.exec>> = useMemo(
    () => ({
      gameName: 'sevens',
      parseCommand: parseSevensCommand,
      formatResponse: formatSevensState,
      helpText: SEVENS_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isHumanTurnForKbd = !!state && !state.gameEndFlag && !!state.players[state.currentTurn]?.isHuman;
  const humanCardCount = state?.players.find((p) => p.isHuman)?.cards?.length ?? 0;

  const noop = useCallback(() => {}, []);
  const directPlay = useCallback(
    (idx: number) => {
      handleCardPlay(idx);
    },
    [handleCardPlay],
  );
  useCardKeyboardNav({
    cardCount: humanCardCount,
    onToggle: noop,
    onConfirm: noop,
    onClear: noop,
    enabled: isHumanTurnForKbd && !loading,
    onDirectPlay: directPlay,
  });

  const runAction = exec;
  const handleManualReset = useCallback(() => {
    hideActionLog();
    void runAction('reset', -1, 0, 0, config);
  }, [runAction, hideActionLog, config]);

  // Auto-place: when the player has just selected a joker and the board offers
  // exactly one legal slot, skip the redundant click and dispatch the placement
  // immediately. Anything more than one slot still requires the player to pick
  // (otherwise we'd remove their strategic choice).
  //
  // Important: useSevensGame only clears jokerCardIdx on a *successful* api
  // response. If the placement fails (network error, server-side validation),
  // loading flips back to false but jokerCardIdx is still non-null, which
  // would cause the effect to re-fire and dispatch the same failing placement
  // forever. We guard with a ref that tracks "we already tried auto-placing
  // for this jokerCardIdx", and only reset it once the index changes.
  const tablePlacedForAuto = state?.tablePlaced;
  const tunnelEnabledForAuto = state?.config.tunnelEnabled;
  const endStopEnabledForAuto = state?.config.endStopEnabled;
  const tunnelSkipWidthForAuto = state?.config.tunnelSkipWidth;
  const lastAutoAttemptedRef = useRef<number | null>(null);
  useEffect(() => {
    // Reset the guard whenever the joker selection clears or changes.
    if (jokerCardIdx === null) {
      lastAutoAttemptedRef.current = null;
      return;
    }
    if (lastAutoAttemptedRef.current === jokerCardIdx) return;
    if (loading || tablePlacedForAuto === undefined) return;
    const slots = listJokerPlacements(
      tablePlacedForAuto,
      tunnelEnabledForAuto ?? false,
      endStopEnabledForAuto ?? false,
      tunnelSkipWidthForAuto ?? 0,
    );
    if (slots.length === 1) {
      lastAutoAttemptedRef.current = jokerCardIdx;
      handleJokerPlace(slots[0].suit, slots[0].value);
    }
  }, [
    jokerCardIdx,
    loading,
    tablePlacedForAuto,
    tunnelEnabledForAuto,
    endStopEnabledForAuto,
    tunnelSkipWidthForAuto,
    handleJokerPlace,
  ]);

  if (!state)
    return (
      <GameSkeleton
        gameKey="sevens"
        layout={{
          kind: 'card-grid',
          count: 52,
          cols: 'grid-cols-13',
          aspectRatio: 'aspect-square',
          topPills: 3,
          footerHandSize: 5,
        }}
      />
    );

  const tablePlaced = state.tablePlaced;
  const tunnelEnabled = state.config.tunnelEnabled;
  const isHumanTurn = !state.gameEndFlag && !!state.players[state.currentTurn]?.isHuman;
  const humanPlayer = state.players.find((p) => p.isHuman);
  const cpuPlayers = state.players.filter((p) => !p.isHuman);
  const canPass =
    isHumanTurn &&
    humanPlayer != null &&
    (humanPlayer.maxPasses === 0 || humanPlayer.passesUsed < humanPlayer.maxPasses);

  // Passes remaining before a forced dobon; null when passes are unlimited (maxPasses=0).
  const passesRemaining =
    humanPlayer != null && humanPlayer.maxPasses > 0 ? humanPlayer.maxPasses - humanPlayer.passesUsed : null;

  const settingsGroups: SettingsGroup[] = [
    {
      title: t('config.groupRules'),
      items: [
        {
          type: 'checkbox',
          id: 'tunnelEnabled',
          label: t('config.tunnel'),
          checked: config.tunnelEnabled,
          onToggle: (v) => handleToggle('tunnelEnabled', v),
        },
        {
          type: 'select',
          id: 'tunnelSkipWidth',
          label: t('config.tunnelSkip'),
          value: config.tunnelSkipWidth,
          options: [
            { value: 0, label: t('config.tunnelSkipOff') },
            { value: 2, label: '2' },
            { value: 3, label: '3' },
            { value: 4, label: '4' },
            { value: 5, label: '5' },
            { value: 6, label: '6' },
          ],
          onSelect: (v) => handleConfigChange('tunnelSkipWidth', String(v)),
        },
        {
          type: 'checkbox',
          id: 'noJokerFinish',
          label: t('config.noJokerFinish'),
          checked: config.noJokerFinish,
          onToggle: (v) => handleToggle('noJokerFinish', v),
        },
        {
          type: 'checkbox',
          id: 'jokerReclaim',
          label: t('config.jokerReclaim'),
          checked: config.jokerReclaim,
          onToggle: (v) => handleToggle('jokerReclaim', v),
        },
        {
          type: 'checkbox',
          id: 'endStop',
          label: t('config.endStop'),
          checked: config.endStop,
          onToggle: (v) => handleToggle('endStop', v),
        },
        {
          type: 'checkbox',
          id: 'jokerConsecutiveBanned',
          label: t('config.jokerConsecutiveBanned'),
          checked: config.jokerConsecutiveBanned,
          onToggle: (v) => handleToggle('jokerConsecutiveBanned', v),
        },
      ],
    },
    {
      title: t('config.groupGame'),
      items: [
        {
          type: 'select',
          id: 'jokerCount',
          label: t('config.joker'),
          value: config.jokerCount,
          options: [
            { value: 0, label: '0' },
            { value: 1, label: '1' },
            { value: 2, label: '2' },
          ],
          onSelect: (v) => handleConfigChange('jokerCount', String(v)),
        },
        {
          type: 'select',
          id: 'cpuStrategy',
          label: t('config.cpuStrategy'),
          value: config.cpuStrategy,
          options: [
            { value: 0, label: t('config.cpuStrategyOff') },
            { value: 1, label: t('config.cpuStrategyStrategic') },
            { value: 2, label: t('config.cpuStrategyHarassment') },
          ],
          onSelect: (v) => handleConfigChange('cpuStrategy', String(v)),
        },
        {
          type: 'select',
          id: 'maxPasses',
          label: t('config.passCount'),
          value: config.maxPasses,
          options: [
            { value: 3, label: '3' },
            { value: 5, label: '5' },
            { value: 10, label: '10' },
            { value: 0, label: t('config.passUnlimited') },
          ],
          onSelect: (v) => handleConfigChange('maxPasses', String(v)),
        },
        hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
      ],
    },
  ];

  return (
    <GamePageShell
      title={tc('nav.sevens')}
      gameThemeBg={gameTheme.sevens.bg}
      phaseName={state.gameEndFlag ? t('phase.end') : t('phase.play')}
      isHumanTurn={isHumanTurn}
      gamePath="/sevens"
      gameEndFlag={!!state.gameEndFlag}
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
          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            {state.config &&
              (state.config.tunnelEnabled ||
                state.config.tunnelSkipWidth >= 2 ||
                state.config.jokerCount > 0 ||
                state.config.cpuStrategy !== 0 ||
                state.config.maxPasses !== 5 ||
                state.config.noJokerFinish ||
                state.config.jokerReclaimEnabled ||
                state.config.endStopEnabled ||
                state.config.jokerConsecutiveBanned) && (
                <div className="bg-black/30 rounded-lg text-ds-warning py-1.5 px-3 mb-2 text-xs">
                  {t('rules.title')}
                  {state.config.tunnelEnabled && ` ${t('rules.tunnelTag')}`}
                  {state.config.tunnelSkipWidth >= 2 &&
                    ` ${t('rules.tunnelSkipTag', { width: state.config.tunnelSkipWidth })}`}
                  {state.config.jokerCount > 0 && ` ${t('rules.jokerTag', { count: state.config.jokerCount })}`}
                  {state.config.cpuStrategy === 1 && ` ${t('rules.cpuStrategy')}`}
                  {state.config.cpuStrategy === 2 && ` ${t('rules.cpuHarassment')}`}
                  {state.config.maxPasses === 0 && ` ${t('rules.passUnlimited')}`}
                  {state.config.maxPasses !== 5 &&
                    state.config.maxPasses !== 0 &&
                    ` ${t('rules.passCount', { count: state.config.maxPasses })}`}
                  {state.config.noJokerFinish && ` ${t('rules.noJokerFinish')}`}
                  {state.config.jokerReclaimEnabled && ` ${t('rules.jokerReclaim')}`}
                  {state.config.endStopEnabled && ` ${t('rules.endStop')}`}
                  {state.config.jokerConsecutiveBanned && ` ${t('rules.jokerConsecutiveBanned')}`}
                </div>
              )}

            <div className="flex gap-2.5 flex-wrap mb-2.5">
              {cpuPlayers.map((player) => (
                <SevensCpuArea key={player.id} player={player} isCurrentTurn={state.currentTurn === player.id} />
              ))}
            </div>

            <div data-tutorial="sv-board">
              <SevensBoard
                tablePlaced={tablePlaced}
                tunnelEnabled={tunnelEnabled}
                tunnelSkipWidth={state.config.tunnelSkipWidth}
                endStopEnabled={state.config.endStopEnabled}
                jokerSelecting={jokerCardIdx !== null}
                onJokerPlace={handleJokerPlace}
              />
            </div>

            {state.humanAction && (
              <div
                data-testid={state.humanAction.forcedPass ? 'human-action-forced-pass' : 'human-action'}
                className={`my-2 ${state.humanAction.forcedPass ? badgeWarning : badgeSuccess}`}
              >
                {actionDesc(state.players, state.humanAction, t)}
              </div>
            )}

            {state.cpuActions && state.cpuActions.length > 0 && (
              <div className="bg-black/40 rounded-lg py-2 px-3.5 my-2 text-xs">
                <span className="text-ds-text-primary">{tc('label.cpuActions')}</span>
                {state.cpuActions.map((a, i) => (
                  <div
                    key={`cpu-action-${a.playerIdx}-${i}`}
                    data-testid={a.forcedPass ? `cpu-action-forced-pass-${i}` : `cpu-action-${i}`}
                    className={a.forcedPass ? 'text-ds-warning' : 'text-ds-text-primary'}
                  >
                    {actionDesc(state.players, a, t)}
                  </div>
                ))}
              </div>
            )}

            <GameMessageBox
              message={
                state.gameEndFlag
                  ? `${t('resultPrefix')} ${state.players
                      .filter((p) => p.rank > 0)
                      .sort((a, b) => a.rank - b.rank)
                      .map((p) =>
                        t('resultEntry', { name: playerName(p.id, p.isHuman), rank: t('rankLabel', { rank: p.rank }) }),
                      )
                      .join(' ')}`
                  : state.message
              }
              messageCode={state.gameEndFlag ? undefined : state.messageCode}
              messageParams={state.gameEndFlag ? undefined : state.messageParams}
            />

            <ActionLogSection
              isEndPhase={state.gameEndFlag}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          <GameFooter className={`${gameTheme.sevens.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <div className="mb-2" data-tutorial="sv-player-hand">
                <SevensHumanArea
                  player={humanPlayer}
                  isCurrentTurn={isHumanTurn}
                  tablePlaced={tablePlaced}
                  tunnelEnabled={tunnelEnabled}
                  tunnelSkipWidth={state.config.tunnelSkipWidth}
                  noJokerFinish={state.config.noJokerFinish}
                  endStopEnabled={state.config.endStopEnabled}
                  jokerConsecutiveBanned={state.config.jokerConsecutiveBanned}
                  loading={loading}
                  onPlay={handleCardPlay}
                />
                {/* Number-key shortcut hint, shown while the human's card bindings are
                    active (matches useCardKeyboardNav onDirectPlay above). */}
                {isHumanTurn && (
                  <p className="text-center text-game-text-muted text-xs mt-1" data-testid="sevens-key-hints">
                    {t('keyHints')}
                  </p>
                )}
              </div>
            )}

            <div data-tutorial="sv-settings">
              <SettingsPanel title={t('config.title')} groups={settingsGroups} />
              <ReplaySpeedSettingsPanel />
            </div>

            <ErrorAlert message={error} onRetry={retry} />

            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="text-center">
              <GameResetButton
                isGameEnd={!!state.gameEndFlag}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="sv-reset-button"
                className="min-w-[90px]"
              />
              <button
                type="button"
                className={`${btnSecondary} min-w-[90px] ${passesRemaining === 1 ? 'text-ds-warning' : ''}`}
                disabled={loading || !canPass}
                onClick={() => exec('play', -1)}
                data-tutorial="sv-play-pass"
              >
                {passesRemaining === null ? tc('button.pass') : t('passRemaining', { count: passesRemaining })}
              </button>
              {jokerCardIdx !== null && (
                <button type="button" className={`${btnSecondary} min-w-[90px]`} onClick={() => setJokerCardIdx(null)}>
                  {tc('button.cancel')}
                </button>
              )}
            </div>
            <CardNavShortcutsPanel directPlay data-testid="sevens-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
