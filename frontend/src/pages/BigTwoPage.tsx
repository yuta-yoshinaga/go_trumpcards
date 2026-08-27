import { useCallback, useMemo, useState } from 'react';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ReplaySpeedSettingsPanel } from '../components/common/ReplaySpeedSettingsPanel';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useBigTwoGame } from '../hooks/useBigTwoGame';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { badgeErrorColors } from '../styles/badgeStyles';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { activeTurnClass, finishedPlayerClass } from '../styles/gameConstants';
import { gameTheme } from '../styles/gameTheme';
import type { BigTwoResponse } from '../types/card';
import type { TutorialStep } from '../types/tutorial';
import { type BigTwoSortMode, bigTwoPlayTypeKey, classifyBigTwoPlay, sortedBigTwoHand } from '../utils/bigTwoSort';
import {
  BIGTWO_HELP,
  type BigTwoCliArgs,
  formatBigTwoState,
  parseBigTwoCommand,
} from '../utils/cli/commands/bigtwoCommands';
import { hintLocalCommand } from '../utils/cli/hintText';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Hand sort options for the Big Two footer. */
const BIGTWO_SORT_MODES: { mode: BigTwoSortMode; labelKey: string }[] = [
  { mode: 'strength', labelKey: 'sort.strength' },
  { mode: 'suit', labelKey: 'sort.suit' },
  { mode: 'number', labelKey: 'sort.number' },
];

/** Tutorial steps for Big Two. */
const BT_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="bt-cpu-area"]', messageKey: 'tutorial.cpuArea', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="bt-table-cards"]',
    messageKey: 'tutorial.tableCards',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bt-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bt-play-pass"]',
    messageKey: 'tutorial.playPass',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bt-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Big Two (大老二) game page. */
export const BigTwoPage = withTutorial(BigTwoPageContent, 'bigtwo', BT_TUTORIAL_STEPS);
function BigTwoPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('bigtwo');
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
  } = useBigTwoGame();
  const { cardWidth } = useCardDimensions();
  const [sortMode, setSortMode] = useState<BigTwoSortMode>('strength');
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('bigtwo', state);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('bigtwo');
  const cliConfig: CliGameConfig<BigTwoResponse, BigTwoCliArgs> = useMemo(
    () => ({
      gameName: 'bigtwo',
      parseCommand: parseBigTwoCommand,
      formatResponse: formatBigTwoState,
      helpText: [...BIGTWO_HELP],
      localCommand: hintLocalCommand(frontendHint),
    }),
    [frontendHint],
  );
  const { handleCommand } = useCliGame(callApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const onReset = useCallback(() => handleResetWithConfig(), [handleResetWithConfig]);

  const difficultyOptions = useMemo(
    () => [
      { value: '0', label: t('settings.difficultyEasy') },
      { value: '1', label: t('settings.difficultyNormal') },
      { value: '2', label: t('settings.difficultyHard') },
    ],
    [t],
  );

  if (!state || state.players.length < 4) {
    return (
      <GameSkeleton
        gameKey="bigtwo"
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
  const human = state.players[0];
  // Preview the combo type of the currently-selected cards (client-side, mirrors
  // the domain classifier). 0 = invalid combination; 1-8 map to a play-type key.
  const selectedCards = selectedIndices.map((i) => human.cards[i]).filter(Boolean);
  const selectedPlayType = classifyBigTwoPlay(selectedCards);
  const selectedInvalid = selectedIndices.length > 0 && selectedPlayType === 0;
  const canPlay = isHumanTurn && selectedIndices.length > 0 && !selectedInvalid;
  const phaseName = isGameEnd ? t('phase.end') : t('phase.play');

  return (
    <GamePageShell
      title={tc('nav.bigtwo')}
      gameThemeBg={gameTheme.bigtwo.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/bigtwo"
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
            <ErrorAlert message={error} onRetry={retry} />

            {/* CPU players */}
            <div className="flex justify-center gap-6 flex-wrap" data-tutorial="bt-cpu-area">
              {state.players
                .filter((p) => !p.isHuman)
                .map((p) => (
                  <div
                    key={p.id}
                    data-testid={`bt-cpu-${p.id.toString()}`}
                    // **手番中の CPU を枠線で示す。** currentTurn は届いていたのに
                    // isHumanTurn の判定にしか使われておらず、誰の番か画面に出て
                    // いなかった。Daifugo / Sevens と同じ共有スタイルを使う (#5478)。
                    className={`text-center rounded-[8px] p-1 ${
                      p.isFinished
                        ? `${finishedPlayerClass} border-2 border-transparent`
                        : state.currentTurn === p.id && !isGameEnd
                          ? activeTurnClass
                          : 'border-2 border-transparent'
                    }`}
                  >
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
            <div className="py-3 bg-black/20 rounded-lg" data-tutorial="bt-table-cards">
              <div className="text-center text-xs text-ds-text-muted mb-2 flex items-center justify-center gap-2">
                <span>{t('tableCards')}</span>
                {(() => {
                  const key = bigTwoPlayTypeKey(state.tablePlayType);
                  return key && state.tableCards.length > 0 ? (
                    <span
                      data-testid="bt-table-playtype"
                      className="rounded-full bg-ds-accent/30 px-2 py-0.5 text-ds-text-primary font-semibold"
                    >
                      {`${t('tablePlayLabel')}: ${t(`playType.${key}`)}`}
                    </span>
                  ) : null;
                })()}
              </div>
              <div className="flex justify-center gap-2 min-h-[60px]">
                {state.tableCards.length === 0 ? (
                  <span className="text-ds-text-muted text-sm self-center">{t('tableEmpty')}</span>
                ) : (
                  state.tableCards.map((c, i) => <AnimatedCard key={i} card={c} width={cardWidth * 0.9} />)
                )}
              </div>
            </div>

            {/* Human hand */}
            <div className="text-center" data-tutorial="bt-player-hand">
              <div className="text-xs text-ds-text-muted mb-1 flex items-center justify-center gap-1">
                <span>{tc('player.you')}</span>
                {human.isFinished && <span className="font-bold">{t(`rank.${human.rank}`)}</span>}
              </div>
              {isHumanTurn && selectedIndices.length > 0 && (
                <div className="mb-1.5 flex items-center justify-center">
                  {selectedInvalid ? (
                    <span
                      data-testid="bt-selected-playtype"
                      role="status"
                      className={`rounded-full px-2 py-0.5 text-xs font-semibold ${badgeErrorColors}`}
                    >
                      {t('invalidCombo')}
                    </span>
                  ) : (
                    <span
                      data-testid="bt-selected-playtype"
                      role="status"
                      className="rounded-full bg-ds-info/30 px-2 py-0.5 text-xs font-semibold text-ds-text-primary"
                    >
                      {`${t('selectedPlayLabel')}: ${t(`playType.${bigTwoPlayTypeKey(selectedPlayType)}`)}`}
                    </span>
                  )}
                </div>
              )}
              {human.cards.length > 0 && (
                <fieldset className="flex justify-center gap-1.5 mb-2 border-0 p-0 m-0">
                  <legend className="sr-only">{t('sort.label')}</legend>
                  {BIGTWO_SORT_MODES.map(({ mode, labelKey }) => (
                    <button
                      key={mode}
                      type="button"
                      onClick={() => setSortMode(mode)}
                      className={sortMode === mode ? `${btnPrimary} min-w-[64px]` : `${btnSecondary} min-w-[64px]`}
                      data-testid={`bt-sort-${mode}`}
                    >
                      {t(labelKey)}
                    </button>
                  ))}
                </fieldset>
              )}
              <div className="flex flex-wrap justify-center gap-2">
                {sortedBigTwoHand(human.cards, sortMode).map(({ card, index }) => (
                  <button
                    key={index}
                    type="button"
                    onClick={() => isHumanTurn && toggleCardSelection(index)}
                    disabled={!isHumanTurn}
                    className={`rounded transition-all ${
                      selectedIndices.includes(index) ? 'ring-2 ring-ds-info -translate-y-2' : ''
                    } ${isHumanTurn ? 'cursor-pointer hover:opacity-90' : 'cursor-default'}`}
                    data-testid={`hand-card-${index}`}
                  >
                    <AnimatedCard card={card} width={cardWidth} />
                  </button>
                ))}
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

          <GameFooter className={`${gameTheme.bigtwo.footer} px-4 py-2.5`}>
            <div className="flex gap-2 justify-center flex-wrap" data-tutorial="bt-play-pass">
              <button
                type="button"
                onClick={handlePlay}
                disabled={loading || !canPlay}
                className="px-4 py-2 rounded-lg bg-ds-info hover:bg-ds-info text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed text-sm"
                data-testid="play-button"
              >
                {t('playButton')}
              </button>
              <button
                type="button"
                onClick={handlePass}
                disabled={loading || !isHumanTurn}
                className="px-4 py-2 rounded-lg bg-ds-warning hover:bg-ds-warning text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed text-sm"
                data-testid="pass-button"
              >
                {t('passButton')}
              </button>
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={onReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="bt-reset-button"
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
