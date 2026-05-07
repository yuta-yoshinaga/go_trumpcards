import { useCallback, useMemo } from 'react';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ReplaySpeedSettingsPanel } from '../components/common/ReplaySpeedSettingsPanel';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { HintTooltip } from '../components/hint/HintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePresidentGame } from '../hooks/usePresidentGame';
import { gameTheme } from '../styles/gameTheme';
import type { PresidentResponse } from '../types/card';
import type { TutorialStep } from '../types/tutorial';
import {
  formatPresidentState,
  PRESIDENT_HELP,
  type PresidentCliArgs,
  parsePresidentCommand,
} from '../utils/cli/commands/presidentCommands';
import type { CliGameConfig } from '../utils/cli/types';

const DIFFICULTY_OPTIONS = [
  { value: '0', label: 'Easy' },
  { value: '1', label: 'Normal' },
  { value: '2', label: 'Hard' },
];

/** Tutorial steps for President. */
const PR_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="pr-cpu-area"]', messageKey: 'tutorial.cpuArea', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="pr-table-cards"]',
    messageKey: 'tutorial.tableCards',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pr-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pr-play-pass"]',
    messageKey: 'tutorial.playPass',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pr-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const PRESIDENT_RANK_KEYS: Readonly<Record<number, string>> = {
  1: 'rank.president',
  2: 'rank.vicePresident',
  3: 'rank.viceScum',
  4: 'rank.scum',
};

/** Renders the President (プレジデント) game page. */
export const PresidentPage = withTutorial(PresidentPageContent, 'president', PR_TUTORIAL_STEPS);
function PresidentPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('president');
  const {
    state,
    loading,
    error,
    callApi,
    selectedIndices,
    toggleCardSelection,
    configInput,
    handleConfigChange,
    handlePlay,
    handlePass,
    handleResetWithConfig,
    retry,
  } = usePresidentGame();
  const { cardWidth } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('president', state);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('president');
  const cliConfig: CliGameConfig<PresidentResponse, PresidentCliArgs> = useMemo(
    () => ({
      gameName: 'president',
      parseCommand: parsePresidentCommand,
      formatResponse: formatPresidentState,
      helpText: [...PRESIDENT_HELP],
    }),
    [],
  );
  const { handleCommand } = useCliGame(callApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const onReset = useCallback(() => handleResetWithConfig(), [handleResetWithConfig]);

  if (!state || state.players.length < 4) {
    return (
      <div className={`flex-1 flex items-center justify-center ${gameTheme.president.bg} text-ds-text-muted`} aria-busy>
        {tc('common.loading')}
      </div>
    );
  }

  const isGameEnd = state.gameEndFlag;
  const humanWon = isGameEnd && state.players[0]?.rank === 1;
  const isHumanTurn = state.currentTurn === 0 && !isGameEnd;
  const human = state.players[0];
  const canPlay = isHumanTurn && selectedIndices.length > 0;
  const phaseName = isGameEnd ? t('phase.end') : t('phase.play');

  return (
    <GamePageShell
      title={tc('nav.president')}
      gameThemeBg={gameTheme.president.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/president"
      gameEndFlag={!!isGameEnd}
      winShow={isGameEnd && humanWon}
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
            {error && (
              <button type="button" onClick={retry} className="text-ds-error underline">
                {error}
              </button>
            )}

            {state.revolutionActive && (
              <div className="text-center text-ds-warning font-semibold">{t('badge.revolution')}</div>
            )}

            {/* CPU players */}
            <div className="flex justify-center gap-6 flex-wrap" data-tutorial="pr-cpu-area">
              {state.players
                .filter((p) => !p.isHuman)
                .map((p) => (
                  <div key={p.id} className="text-center">
                    <div className="text-xs text-ds-text-muted mb-1">
                      {tc('player.cpu', { id: p.id })}{' '}
                      {p.isFinished ? (
                        <span className="text-ds-success">({t(PRESIDENT_RANK_KEYS[p.rank] ?? 'rank.unknown')})</span>
                      ) : (
                        <span>— {p.cardCount}</span>
                      )}
                    </div>
                    <div className="flex gap-0.5 justify-center">
                      {Array.from({ length: Math.min(p.cardCount, 13) }, (_, i) => (
                        <AnimatedCardBack key={i} width={cardWidth * 0.45} />
                      ))}
                    </div>
                  </div>
                ))}
            </div>

            {/* Table cards */}
            <div className="py-3 bg-black/20 rounded-lg" data-tutorial="pr-table-cards">
              <div className="text-center text-xs text-ds-text-muted mb-2">{t('label.tableCards')}</div>
              <div className="flex justify-center gap-2 min-h-[60px]">
                {state.tableCards.length === 0 ? (
                  <span className="text-ds-text-muted text-sm self-center">{t('label.tableEmpty')}</span>
                ) : (
                  state.tableCards.map((c, i) => <AnimatedCard key={i} card={c} width={cardWidth * 0.9} />)
                )}
              </div>
            </div>

            {/* Human hand */}
            <div className="text-center" data-tutorial="pr-player-hand">
              <div className="text-xs text-ds-text-muted mb-1">
                {tc('player.you')}
                {human.isFinished && <> — {t(PRESIDENT_RANK_KEYS[human.rank] ?? 'rank.unknown')}</>}
              </div>
              <div className="flex flex-wrap justify-center gap-2">
                {human.cards.map((c, i) => (
                  <button
                    key={i}
                    type="button"
                    onClick={() => isHumanTurn && toggleCardSelection(i)}
                    disabled={!isHumanTurn}
                    className={`rounded transition-all ${
                      selectedIndices.includes(i) ? 'ring-2 ring-ds-info -translate-y-2' : ''
                    } ${isHumanTurn ? 'cursor-pointer hover:opacity-90' : 'cursor-default'}`}
                    data-testid={`hand-card-${i}`}
                  >
                    <AnimatedCard card={c} width={cardWidth} />
                  </button>
                ))}
              </div>
            </div>

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />
            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}
          </div>

          <SettingsPanel
            title={tc('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select' as const,
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: String(configInput.cpuDifficulty ?? 1),
                    options: DIFFICULTY_OPTIONS,
                    onSelect: (v: string) => handleConfigChange('cpuDifficulty', Number.parseInt(v, 10)),
                  },
                  {
                    type: 'checkbox' as const,
                    id: 'revolutionEnabled',
                    label: t('settings.revolution'),
                    checked: configInput.revolutionEnabled ?? true,
                    onToggle: (v: boolean) => handleConfigChange('revolutionEnabled', v),
                  },
                  {
                    type: 'checkbox' as const,
                    id: 'cardExchangeEnabled',
                    label: t('settings.cardExchange'),
                    checked: configInput.cardExchangeEnabled ?? true,
                    onToggle: (v: boolean) => handleConfigChange('cardExchangeEnabled', v),
                  },
                  {
                    type: 'checkbox' as const,
                    id: 'passFieldFlushEnabled',
                    label: t('settings.passFieldFlush'),
                    checked: configInput.passFieldFlushEnabled ?? true,
                    onToggle: (v: boolean) => handleConfigChange('passFieldFlushEnabled', v),
                  },
                  {
                    type: 'checkbox' as const,
                    id: 'frontendHint',
                    label: tc('hint.toggle', { ns: 'tutorial' }),
                    checked: frontendHintEnabled,
                    onToggle: setFrontendHintEnabled,
                  },
                ],
              },
            ]}
          />
          <ReplaySpeedSettingsPanel />

          <GameFooter className={`${gameTheme.president.footer} px-4 py-2.5`}>
            <div className="flex gap-2 justify-center flex-wrap" data-tutorial="pr-play-pass">
              <button
                type="button"
                onClick={handlePlay}
                disabled={loading || !canPlay}
                className="px-4 py-2 rounded-lg bg-ds-info hover:bg-ds-info text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed text-sm"
                data-testid="play-button"
              >
                {t('button.play')}
              </button>
              <button
                type="button"
                onClick={handlePass}
                disabled={loading || !isHumanTurn}
                className="px-4 py-2 rounded-lg bg-ds-warning hover:bg-ds-warning text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed text-sm"
                data-testid="pass-button"
              >
                {t('button.pass')}
              </button>
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={onReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="pr-reset-button"
              />
              <ActionLogSection
                isEndPhase={isGameEnd}
                actionLog={actionLog}
                showActionLog={showActionLog}
                hideActionLog={hideActionLog}
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
