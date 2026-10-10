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
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useTienLenGame } from '../hooks/useTienLenGame';
import { gameTheme } from '../styles/gameTheme';
import type { TienLenResponse } from '../types/card';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import {
  formatTienLenState,
  parseTienLenCommand,
  TIENLEN_HELP,
  type TienLenCliArgs,
} from '../utils/cli/commands/tienlenCommands';
import { hintLocalCommand } from '../utils/cli/hintText';
import type { CliGameConfig } from '../utils/cli/types';
import { findPlayerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';
import { classifyTienLenCombo, type TienLenCombo, tienLenPlayability } from '../utils/tienLenComboValidator';

// Values match the Go domain constants: 0=Normal, 1=Easy, 2=Hard
// (see TienLenConfig.go / TienLenCuiController help text).
// Values match the Go domain constants: 0=Invalid, 1=Single, 2=Pair,
// 3=Triple, 4=Straight, 5=ThreePairRun, 6=FourOfAKind.
const TIENLEN_PLAY_TYPE_KEYS: Record<number, string> = {
  1: 'single',
  2: 'pair',
  3: 'triple',
  4: 'straight',
  5: 'threePairRun',
  6: 'fourOfAKind',
};

/** Tutorial steps for Tien Len. */
const TL_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="tl-cpu-area"]', messageKey: 'tutorial.cpuArea', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="tl-table-cards"]',
    messageKey: 'tutorial.tableCards',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tl-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tl-play-pass"]',
    messageKey: 'tutorial.playPass',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tl-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Tien Len (ティエンレン) game page. */
export const TienLenPage = withTutorial(TienLenPageContent, 'tienlen', TL_TUTORIAL_STEPS);
function TienLenPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('tienlen');
  const difficultyOptions = [
    { value: '0', label: t('settings.difficultyNormal') },
    { value: '1', label: t('settings.difficultyEasy') },
    { value: '2', label: t('settings.difficultyHard') },
  ];
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
  } = useTienLenGame();
  const { cardWidth } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('tienlen', state);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('tienlen');
  const cliConfig: CliGameConfig<TienLenResponse, TienLenCliArgs> = useMemo(
    () => ({
      gameName: 'tienlen',
      parseCommand: parseTienLenCommand,
      formatResponse: formatTienLenState,
      helpText: [...TIENLEN_HELP],
      localCommand: hintLocalCommand(frontendHint),
    }),
    [frontendHint],
  );
  const { handleCommand } = useCliGame(callApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const onReset = useCallback(() => handleResetWithConfig(), [handleResetWithConfig]);

  if (!state || state.players.length < 4) {
    return (
      <GameSkeleton
        gameKey="tienlen"
        layout={{
          kind: 'trick-taking',
          titleBar: false,
          opponents: 3,
          opponentStyle: 'hand',
          opponentHandSize: 4,
          trickArea: true,
          footerHandSize: 5,
          footerButton: 'wide',
        }}
      />
    );
  }

  const isGameEnd = state.gameEndFlag;
  const humanWon = isGameEnd && state.players[0]?.rank === 1;
  const isHumanTurn = state.currentTurn === 0 && !isGameEnd;
  const canPass = isHumanTurn && state.tableCards.length > 0 && !loading;
  const human = state.players[0];
  const selectedCards = selectedIndices.map((i) => human.cards[i]).filter((c): c is NonNullable<typeof c> => c != null);
  const selectedCombo = classifyTienLenCombo(selectedCards);
  const hasValidCombo = selectedCards.length > 0 && selectedCombo !== 'invalid';
  const playability =
    hasValidCombo && state.tableCards.length > 0
      ? tienLenPlayability(selectedCards, state.tableCards, TIENLEN_PLAY_TYPE_KEYS[state.tablePlayType] as TienLenCombo)
      : 'ok';
  const isBomb = selectedCombo === 'threePairRun' || selectedCombo === 'fourOfAKind';
  const canPlay = isHumanTurn && selectedIndices.length > 0 && hasValidCombo && playability === 'ok';
  const showInvalidCombo = isHumanTurn && selectedIndices.length > 0 && !hasValidCombo;
  const showComboType = isHumanTurn && selectedIndices.length > 0 && hasValidCombo;
  const phaseName = isGameEnd ? t('phase.end') : t('phase.play');

  return (
    <GamePageShell
      title={tc('nav.tienlen')}
      gameThemeBg={gameTheme.tienlen.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/tienlen"
      gameEndFlag={!!isGameEnd}
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
            {error && (
              <button type="button" onClick={retry} className="text-ds-error-text underline">
                {error}
              </button>
            )}

            <section
              aria-label={t('rules.title')}
              data-testid="tl-rules"
              className="rounded-lg bg-ds-surface p-3 text-xs text-ds-text-muted"
            >
              <h2 className="mb-1 font-semibold text-ds-text-primary">{t('rules.title')}</h2>
              <p>{t('rules.strength')}</p>
              <p>{t('rules.follow')}</p>
              <p>{t('rules.lead')}</p>
            </section>

            {/* CPU players */}
            <div className="flex justify-center gap-6 flex-wrap" data-tutorial="tl-cpu-area">
              {state.players
                .filter((p) => !p.isHuman)
                .map((p) => (
                  <div key={p.id} className="text-center">
                    <div className="text-xs text-ds-text-muted mb-1 flex items-center justify-center gap-1">
                      <span>{tc('player.cpu', { id: p.id })}</span>
                      {p.isFinished ? (
                        <span className="font-bold">{t(`rank.${p.rank}`)}</span>
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
            <div className="py-3 bg-black/20 rounded-lg" data-tutorial="tl-table-cards">
              <div className="text-center text-xs text-ds-text-muted mb-2">{t('tableCards')}</div>
              {state.tableCards.length > 0 && state.lastPlayPlayerIdx >= 0 && (
                <div className="text-center text-xs font-semibold text-ds-info mb-2" data-testid="tl-table-owner">
                  {t('tablePlayedBy', { name: findPlayerName(state.players, state.lastPlayPlayerIdx) })}
                </div>
              )}
              <div className="flex justify-center gap-2 min-h-[60px]">
                {state.tableCards.length === 0 ? (
                  <span className="text-ds-text-muted text-sm self-center">{t('tableEmpty')}</span>
                ) : (
                  state.tableCards.map((c, i) => <AnimatedCard key={i} card={c} width={cardWidth * 0.9} />)
                )}
              </div>
              {state.tableCards.length > 0 && TIENLEN_PLAY_TYPE_KEYS[state.tablePlayType] && (
                <div className="text-center text-sm font-semibold text-ds-accent mt-2" data-testid="tl-table-playtype">
                  {`${t('tablePlayLabel')}: ${t(`playType.${TIENLEN_PLAY_TYPE_KEYS[state.tablePlayType]}`)}`}
                </div>
              )}
            </div>

            {/* Human hand */}
            <div className="text-center" data-tutorial="tl-player-hand">
              <div className="text-xs text-ds-text-muted mb-1 flex items-center justify-center gap-1">
                <span>{tc('player.you')}</span>
                {human.isFinished && <span className="font-bold">{t(`rank.${human.rank}`)}</span>}
              </div>
              <div className="flex flex-wrap justify-center gap-2">
                {human.cards.map((c, i) => {
                  const selected = selectedIndices.includes(i);
                  const cardClass = selected
                    ? isHumanTurn
                      ? 'rounded transition-all ring-2 ring-ds-info -translate-y-2 cursor-pointer hover:opacity-90'
                      : 'rounded transition-all ring-2 ring-ds-info -translate-y-2 cursor-default'
                    : isHumanTurn
                      ? 'rounded transition-all cursor-pointer hover:opacity-90'
                      : 'rounded transition-all cursor-default';
                  return (
                    <button
                      key={i}
                      type="button"
                      onClick={() => isHumanTurn && toggleCardSelection(i)}
                      disabled={!isHumanTurn}
                      aria-label={cardAlt(c)}
                      aria-pressed={selected}
                      className={cardClass}
                      data-testid={`hand-card-${i}`}
                    >
                      <AnimatedCard card={c} width={cardWidth} />
                    </button>
                  );
                })}
              </div>
            </div>

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />
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
                    options: difficultyOptions,
                    onSelect: (v: string) => handleConfigChange('cpuDifficulty', Number.parseInt(v, 10)),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />
          <ReplaySpeedSettingsPanel />

          <GameFooter className={`${gameTheme.tienlen.footer} px-4 py-2.5`}>
            {showComboType && (
              <p
                role="status"
                data-testid="tl-combo-type"
                className={`mb-1 text-center text-xs font-semibold ${isBomb ? 'text-ds-warning' : 'text-ds-info'}`}
              >
                {`${t('selectedPlayLabel')}: ${t('comboBadge', {
                  type: t(`playType.${selectedCombo}`),
                  count: selectedCards.length,
                })}`}
                {isBomb && ` · ${t('bombLabel')}`}
              </p>
            )}
            {showInvalidCombo && (
              <p
                role="status"
                data-testid="tl-invalid-combo"
                className="mb-1 text-center font-medium text-ds-warning text-xs"
              >
                {t('invalidCombo')}
              </p>
            )}
            {isHumanTurn && hasValidCombo && playability !== 'ok' && (
              <p
                role="status"
                data-testid="tl-unplayable-reason"
                className="mb-1 text-center font-medium text-ds-warning text-xs"
              >
                {t(`unplayable.${playability}`)}
              </p>
            )}
            <div className="flex gap-2 justify-center flex-wrap" data-tutorial="tl-play-pass">
              <button
                type="button"
                onClick={handlePlay}
                disabled={loading || !canPlay}
                aria-disabled={isHumanTurn && hasValidCombo && playability !== 'ok' ? 'true' : undefined}
                className="px-4 py-2 rounded-lg bg-ds-info hover:bg-ds-info text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed text-sm"
                data-testid="play-button"
              >
                {t('playButton')}
              </button>
              <button
                type="button"
                onClick={() => {
                  if (!canPass) return;
                  handlePass();
                }}
                disabled={loading || !isHumanTurn}
                aria-disabled={isHumanTurn && state.tableCards.length === 0 ? 'true' : undefined}
                aria-describedby={isHumanTurn && state.tableCards.length === 0 ? 'tl-pass-unavailable' : undefined}
                className="px-4 py-2 rounded-lg bg-ds-warning hover:bg-ds-warning text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed aria-disabled:opacity-40 aria-disabled:cursor-not-allowed text-sm"
                data-testid="pass-button"
              >
                {t('passButton')}
              </button>
              {isHumanTurn && state.tableCards.length === 0 && (
                <span id="tl-pass-unavailable" className="sr-only">
                  {t('passUnavailable')}
                </span>
              )}
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={onReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="tl-reset-button"
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
