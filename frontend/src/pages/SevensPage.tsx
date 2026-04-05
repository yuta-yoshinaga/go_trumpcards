import { useCallback, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import type { sevensApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import type { SettingsGroup } from '../components/common/SettingsPanel';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageHeading } from '../components/GamePageHeading';
import { GameResetDialog } from '../components/GameResetDialog';
import { HintTooltip } from '../components/hint/HintTooltip';
import { ManualButton } from '../components/ManualButton';
import { WinCelebration } from '../components/motion/WinCelebration';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { SevensBoard } from '../components/sevens/SevensBoard';
import { SevensCpuArea } from '../components/sevens/SevensCpuArea';
import { SevensHumanArea } from '../components/sevens/SevensHumanArea';
import { SevensSkeleton } from '../components/skeleton/SevensSkeleton';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useSevensGame } from '../hooks/useSevensGame';
import { useSound } from '../providers/SoundProvider';
import { TutorialProvider } from '../providers/TutorialProvider';
import { btnOutline, btnSecondary } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { SevensResponse } from '../types/card';
import type { TutorialConfig, TutorialStep } from '../types/tutorial';
import { parseSevensCommand, SEVENS_HELP } from '../utils/cli/commands/sevensCommands';
import { formatSevensState } from '../utils/cli/formatters/sevensFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';
import { actionDesc } from '../utils/sevensUtils';

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

/** Sevens tutorial configuration. */
const SV_TUTORIAL_CONFIG: TutorialConfig = {
  gameName: 'sevens',
  steps: SV_TUTORIAL_STEPS,
};

/** Renders the Sevens game page with board, player areas, and joker placement. */
export function SevensPage() {
  const { t: tSv } = useTranslation('sevens');
  return (
    <TutorialProvider config={SV_TUTORIAL_CONFIG} translateMessage={tSv}>
      <SevensPageContent />
    </TutorialProvider>
  );
}

/** Inner content of the Sevens page, wrapped by TutorialProvider. */
function SevensPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('sevens');
  const { playSound } = useSound();
  const {
    state,
    loading,
    error,
    exec,
    retry,
    jokerCardIdx,
    setJokerCardIdx,
    cfgTunnel,
    setCfgTunnel,
    cfgTunnelSkipWidth,
    setCfgTunnelSkipWidth,
    cfgJokerCount,
    setCfgJokerCount,
    cfgCpuStrategy,
    setCfgCpuStrategy,
    cfgMaxPasses,
    setCfgMaxPasses,
    cfgNoJokerFinish,
    setCfgNoJokerFinish,
    cfgJokerReclaim,
    setCfgJokerReclaim,
    cfgEndStop,
    setCfgEndStop,
    cfgJokerConsBan,
    setCfgJokerConsBan,
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

  if (!state) return <SevensSkeleton />;

  const tablePlaced = state.tablePlaced;
  const tunnelEnabled = state.config.tunnelEnabled;
  const isHumanTurn = !state.gameEndFlag && !!state.players[state.currentTurn]?.isHuman;
  const humanPlayer = state.players.find((p) => p.isHuman);
  const cpuPlayers = state.players.filter((p) => !p.isHuman);
  const canPass =
    isHumanTurn &&
    humanPlayer != null &&
    (humanPlayer.maxPasses === 0 || humanPlayer.passesUsed < humanPlayer.maxPasses);

  const settingsGroups: SettingsGroup[] = [
    {
      title: t('config.groupRules'),
      items: [
        {
          type: 'checkbox',
          id: 'cfgTunnel',
          label: t('config.tunnel'),
          checked: cfgTunnel,
          onToggle: setCfgTunnel,
        },
        {
          type: 'select',
          id: 'cfgTunnelSkipWidth',
          label: t('config.tunnelSkip'),
          value: cfgTunnelSkipWidth,
          options: [
            { value: 0, label: t('config.tunnelSkipOff') },
            { value: 2, label: '2' },
            { value: 3, label: '3' },
            { value: 4, label: '4' },
            { value: 5, label: '5' },
            { value: 6, label: '6' },
          ],
          onSelect: (v) => setCfgTunnelSkipWidth(Number(v)),
        },
        {
          type: 'checkbox',
          id: 'cfgNoJokerFinish',
          label: t('config.noJokerFinish'),
          checked: cfgNoJokerFinish,
          onToggle: setCfgNoJokerFinish,
        },
        {
          type: 'checkbox',
          id: 'cfgJokerReclaim',
          label: t('config.jokerReclaim'),
          checked: cfgJokerReclaim,
          onToggle: setCfgJokerReclaim,
        },
        {
          type: 'checkbox',
          id: 'cfgEndStop',
          label: t('config.endStop'),
          checked: cfgEndStop,
          onToggle: setCfgEndStop,
        },
        {
          type: 'checkbox',
          id: 'cfgJokerConsBan',
          label: t('config.jokerConsecutiveBanned'),
          checked: cfgJokerConsBan,
          onToggle: setCfgJokerConsBan,
        },
      ],
    },
    {
      title: t('config.groupGame'),
      items: [
        {
          type: 'select',
          id: 'cfgJokerCount',
          label: t('config.joker'),
          value: cfgJokerCount,
          options: [
            { value: 0, label: '0' },
            { value: 1, label: '1' },
            { value: 2, label: '2' },
          ],
          onSelect: (v) => setCfgJokerCount(Number(v)),
        },
        {
          type: 'select',
          id: 'cfgCpuStrategy',
          label: t('config.cpuStrategy'),
          value: cfgCpuStrategy,
          options: [
            { value: 0, label: t('config.cpuStrategyOff') },
            { value: 1, label: t('config.cpuStrategyStrategic') },
            { value: 2, label: t('config.cpuStrategyHarassment') },
          ],
          onSelect: (v) => setCfgCpuStrategy(Number(v)),
        },
        {
          type: 'select',
          id: 'cfgMaxPasses',
          label: t('config.passCount'),
          value: cfgMaxPasses,
          options: [
            { value: 3, label: '3' },
            { value: 5, label: '5' },
            { value: 10, label: '10' },
            { value: 0, label: t('config.passUnlimited') },
          ],
          onSelect: (v) => setCfgMaxPasses(Number(v)),
        },
        {
          type: 'checkbox',
          id: 'frontendHint',
          label: tc('hint.toggle', { ns: 'tutorial' }),
          checked: frontendHintEnabled,
          onToggle: setFrontendHintEnabled,
        },
      ],
    },
  ];

  return (
    <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.sevens.bg}`} aria-busy={loading} aria-live="polite">
      <GamePageHeading title={tc('nav.sevens')} />
      <PhaseIndicator phaseName={state.gameEndFlag ? t('phase.end') : t('phase.play')} isHumanTurn={isHumanTurn}>
        <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        <TutorialButton />
        <ManualButton gamePath="/sevens" />
      </PhaseIndicator>

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
                <div className="bg-black/30 rounded-lg text-yellow-300 py-1.5 px-3 mb-2 text-xs">
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
                className={`rounded-lg py-2 px-3.5 my-2 text-xs ${state.humanAction.forcedPass ? 'bg-red-900/50 text-orange-200 border border-red-500/50' : 'bg-black/40 text-green-200'}`}
              >
                {actionDesc(state.players, state.humanAction, t)}
              </div>
            )}

            {state.cpuActions && state.cpuActions.length > 0 && (
              <div className="bg-black/40 rounded-lg py-2 px-3.5 my-2 text-xs">
                <span className="text-white">{tc('label.cpuActions')}</span>
                {state.cpuActions.map((a, i) => (
                  <div
                    key={`cpu-action-${a.playerIdx}-${i}`}
                    data-testid={a.forcedPass ? `cpu-action-forced-pass-${i}` : `cpu-action-${i}`}
                    className={a.forcedPass ? 'text-orange-200' : 'text-white'}
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
              </div>
            )}

            <div data-tutorial="sv-settings">
              <SettingsPanel title={t('config.title')} groups={settingsGroups} />
            </div>

            <ErrorAlert message={error} onRetry={retry} />

            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

            <div className="text-center">
              <button
                type="button"
                className={`${btnOutline} min-w-[90px]`}
                disabled={loading}
                data-tutorial="sv-reset-button"
                onClick={() =>
                  requestConfirm(() => {
                    hideActionLog();
                    exec('reset', -1, 0, 0, {
                      tunnelEnabled: cfgTunnel,
                      tunnelSkipWidth: cfgTunnelSkipWidth,
                      jokerCount: cfgJokerCount,
                      cpuStrategy: cfgCpuStrategy,
                      maxPasses: cfgMaxPasses,
                      noJokerFinish: cfgNoJokerFinish,
                      jokerReclaim: cfgJokerReclaim,
                      endStop: cfgEndStop,
                      jokerConsecutiveBanned: cfgJokerConsBan,
                    });
                  })
                }
              >
                {tc('button.reset')}
              </button>
              <button
                type="button"
                className={`${btnSecondary} min-w-[90px]`}
                disabled={loading || !canPass}
                onClick={() => exec('play', -1)}
                data-tutorial="sv-play-pass"
              >
                {tc('button.pass')}
              </button>
              {jokerCardIdx !== null && (
                <button type="button" className={`${btnSecondary} min-w-[90px]`} onClick={() => setJokerCardIdx(null)}>
                  {tc('button.cancel')}
                </button>
              )}
            </div>
          </GameFooter>
        </>
      )}
      <WinCelebration show={!!state?.gameEndFlag} onCelebrate={() => playSound('winFanfare')} />
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
    </div>
  );
}
