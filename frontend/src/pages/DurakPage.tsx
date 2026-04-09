import { useMemo } from 'react';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageHeading } from '../components/GamePageHeading';
import { GameResetDialog } from '../components/GameResetDialog';
import { ManualButton } from '../components/ManualButton';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { WinCelebration } from '../components/motion/WinCelebration';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { SkeletonHand } from '../components/skeleton/SkeletonHand';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { TutorialWrapper } from '../components/tutorial/TutorialWrapper';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { CPU_DIFFICULTY_OPTIONS, useDurakGame } from '../hooks/useDurakGame';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useSound } from '../providers/SoundProvider';
import { btnDanger, btnOutline, btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { DurakResponse } from '../types/card';
import type { TutorialStep } from '../types/tutorial';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';

/** Durak phase constants. */
const PHASE_ATTACK = 0;
const PHASE_DEFEND = 1;
const PHASE_BOUT_END = 2;

/** Fallback theme for durak (matching/pass category). */
const theme = gameTheme.durak ?? { bg: 'bg-game-bg-green', footer: 'bg-game-bg-green-dark border-white/20' };

/** Durak tutorial step definitions. */
const DURAK_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="dk-trump-info"]',
    messageKey: 'tutorial.trumpInfo',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dk-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dk-table-cards"]',
    messageKey: 'tutorial.tableCards',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dk-action-buttons"]',
    messageKey: 'tutorial.actionButtons',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Durak game page with attack/defense mechanics. */
export function DurakPage() {
  return (
    <TutorialWrapper gameName="durak" steps={DURAK_TUTORIAL_STEPS}>
      <DurakPageContent />
    </TutorialWrapper>
  );
}

/** Loading skeleton for the Durak page. */
function DurakSkeleton() {
  const { cardWidth, cardHeight } = useCardDimensions();
  return (
    <GameSkeleton
      bgClass={theme.bg}
      footerClassName={`${theme.footer} px-4 py-2.5`}
      footer={
        <>
          <div className="mb-2">
            <SkeletonHand cardWidth={cardWidth} cardHeight={cardHeight} count={6} />
          </div>
          <div className="h-8 w-32 rounded bg-white/10 animate-pulse mx-auto" />
        </>
      }
    >
      <div className="bg-black/30 rounded-[10px] py-2.5 px-3.5 my-2">
        <div className="h-4 w-24 rounded bg-white/10 animate-pulse mb-1" />
        <div className="h-4 w-32 rounded bg-white/10 animate-pulse" />
      </div>
      <div className="flex gap-2 flex-wrap mb-3">
        {Array.from({ length: 3 }, (_, i) => (
          <div key={i} className="p-2 rounded bg-black/30">
            <div className="h-4 w-16 rounded bg-white/10 animate-pulse mb-2" />
            <SkeletonHand cardWidth={cardWidth} cardHeight={cardHeight} count={6} />
          </div>
        ))}
      </div>
    </GameSkeleton>
  );
}

/** Inner content of the Durak page, wrapped by TutorialProvider. */
function DurakPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('durak');
  const {
    state,
    loading,
    error,
    exec: gameExec,
    retry,
    durakConfig,
    selectedCardIdx,
    setSelectedCardIdx,
    selectedAttackIdx,
    setSelectedAttackIdx,
    handleConfigChange,
    handleConfigToggle,
    handleAttack,
    handleDefend,
    handlePass,
    handleTake,
    handleSort,
  } = useDurakGame();

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('durak');
  const durakCliConfig: CliGameConfig<DurakResponse, Parameters<typeof gameExec>> = useMemo(
    () => ({
      gameName: 'durak',
      parseCommand: (_cmd: string) => ({ error: 'CLI not supported' }) as const,
      formatResponse: (_res: DurakResponse): string => '',
      helpText: [] as string[],
    }),
    [],
  );
  const { handleCommand } = useCliGame(gameExec, durakCliConfig, state, { addInput, addOutput, addError, clearLog });

  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();

  if (!state) return <DurakSkeleton />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const cpuPlayers = state.players.filter((p) => !p.isHuman);
  const isAttacker = humanPlayer ? state.attackerIdx === humanPlayer.id : false;
  const isDefender = humanPlayer ? state.defenderIdx === humanPlayer.id : false;
  const isGameEnd = state.gameEndFlag;

  const phaseName = isGameEnd
    ? t('phase.gameEnd')
    : state.phase === PHASE_ATTACK
      ? t('phase.attack')
      : state.phase === PHASE_DEFEND
        ? t('phase.defend')
        : state.phase === PHASE_BOUT_END
          ? t('phase.boutEnd')
          : '';

  const isHumanTurn = !isGameEnd && humanPlayer !== undefined && state.currentTurn === humanPlayer.id;

  const showAttackBtn = !isGameEnd && isAttacker && state.phase === PHASE_ATTACK;
  const showDefendBtn = !isGameEnd && isDefender && state.phase === PHASE_DEFEND;
  const showPassBtn = !isGameEnd && isAttacker && state.phase === PHASE_ATTACK && state.tablePairs.length > 0;
  const showTakeBtn = !isGameEnd && isDefender && state.phase === PHASE_DEFEND;

  const suitSymbol = (suit: string) => {
    switch (suit) {
      case 'SPADE':
        return '\u2660';
      case 'HEART':
        return '\u2665';
      case 'DIAMOND':
        return '\u2666';
      case 'CLUB':
        return '\u2663';
      default:
        return suit;
    }
  };

  return (
    <div className={`flex-1 flex flex-col min-h-0 ${theme.bg}`} aria-busy={loading} aria-live="polite">
      <GamePageHeading title={tc('nav.durak')} />
      <PhaseIndicator phaseName={phaseName} isHumanTurn={isHumanTurn}>
        <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        <TutorialButton />
        <ManualButton gamePath="/durak" />
      </PhaseIndicator>

      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          {/* Settings panel */}
          <SettingsPanel
            title={t('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select',
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: durakConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((opt) => ({
                      value: opt.value,
                      label: t(`settings.difficulty${opt.label.charAt(0).toUpperCase()}${opt.label.slice(1)}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'checkbox',
                    id: 'transferEnabled',
                    label: t('settings.transferEnabled'),
                    checked: durakConfig.transferEnabled,
                    onToggle: (checked) => handleConfigToggle('transferEnabled', checked),
                  },
                ],
              },
            ]}
          />

          {/* Scrollable area */}
          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className={lgTwoColGrid}>
              {/* Left: game play area */}
              <div>
                {/* Trump info */}
                <div className="bg-black/30 rounded-[10px] py-2.5 px-3.5 my-2" data-tutorial="dk-trump-info">
                  <div className="text-white font-bold mb-1">
                    {t('trump')}: {suitSymbol(state.trumpSuit)}
                    {state.trumpCard && (
                      <span className="ml-2">
                        <AnimatedCard card={state.trumpCard} width={cardWidth * 0.6} />
                      </span>
                    )}
                  </div>
                  <div className="text-game-text-muted text-sm">
                    {t('stock')}: {state.stockCount}
                  </div>
                </div>

                {/* Table pairs */}
                <div className="bg-black/30 rounded-[10px] py-2.5 px-3.5 my-2" data-tutorial="dk-table-cards">
                  <div className="text-white font-bold mb-1">{t('table')}</div>
                  {state.tablePairs.length === 0 ? (
                    <div className="text-game-text-muted text-sm">{t('tableEmpty')}</div>
                  ) : (
                    <div className="flex flex-wrap gap-2">
                      {state.tablePairs.map((pair, i) => (
                        <button
                          type="button"
                          key={`pair-${pair.attack.design}-${pair.attack.value}-${i}`}
                          className={`flex flex-col items-center gap-0.5 p-1 rounded cursor-pointer ${
                            selectedAttackIdx === i ? 'ring-2 ring-yellow-400' : ''
                          }`}
                          onClick={() => setSelectedAttackIdx(selectedAttackIdx === i ? null : i)}
                        >
                          <AnimatedCard
                            card={pair.attack}
                            width={cardWidth * 0.8}
                            onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                          />
                          {pair.defense ? (
                            <AnimatedCard card={pair.defense} width={cardWidth * 0.8} />
                          ) : (
                            <div
                              className="border border-dashed border-white/30 rounded"
                              style={{ width: cardWidth * 0.8, height: cardWidth * 0.8 * 1.4 }}
                            />
                          )}
                        </button>
                      ))}
                    </div>
                  )}
                </div>

                {/* CPU actions */}
                {state.cpuActions.length > 0 && (
                  <div className="bg-black/40 rounded-lg text-game-text-muted py-2 px-3.5 my-2 whitespace-pre-line text-xs">
                    {[
                      tc('label.cpuActions'),
                      ...state.cpuActions.map(
                        (a) =>
                          `${playerName(state.players[a.playerIdx]?.id ?? a.playerIdx, false)}: ${t(`action.${['attack', 'defend', 'pass', 'take'][a.actionType] ?? 'pass'}`)}`,
                      ),
                    ].join('\n')}
                  </div>
                )}

                {/* Result message */}
                <GameMessageBox
                  message={state.message}
                  messageCode={state.messageCode}
                  messageParams={state.messageParams}
                />

                {/* Action log */}
                <ActionLogSection
                  isEndPhase={isGameEnd}
                  actionLog={actionLog}
                  showActionLog={showActionLog}
                  hideActionLog={hideActionLog}
                />
              </div>

              {/* Right: CPU info sidebar */}
              <div>
                <div className="flex gap-2 flex-wrap mb-3 lg:flex-col">
                  {cpuPlayers.map((player) => (
                    <div
                      key={player.id}
                      className={`p-2 rounded bg-black/30 ${
                        state.currentTurn === player.id ? 'ring-2 ring-yellow-400' : ''
                      }`}
                    >
                      <div className="text-white text-sm font-bold">
                        {playerName(player.id, false)}
                        {state.attackerIdx === player.id && (
                          <span className="text-red-400 text-xs ml-1">({t('attacker')})</span>
                        )}
                        {state.defenderIdx === player.id && (
                          <span className="text-blue-400 text-xs ml-1">({t('defender')})</span>
                        )}
                      </div>
                      <div className="text-game-text-muted text-xs">{player.cardCount}</div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>

          {/* Sticky footer: human player hand + action buttons */}
          <GameFooter className={`${theme.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <div className="mb-2" data-tutorial="dk-player-hand">
                <div className="text-white font-bold text-sm mb-1">
                  {humanPlayer.cardCount}
                  {isAttacker && <span className="text-red-400 text-xs ml-2">({t('attacker')})</span>}
                  {isDefender && <span className="text-blue-400 text-xs ml-2">({t('defender')})</span>}
                  {isHumanTurn && <span className="text-ds-success text-xs ml-2">{t('selectCard')}</span>}
                </div>
                <div className="flex flex-wrap gap-1">
                  {humanPlayer.cards.map((card, i) => (
                    <button
                      type="button"
                      key={`${card.design}-${card.value}-${i}`}
                      className={`cursor-pointer transition-transform ${selectedCardIdx === i ? '-translate-y-2' : ''}`}
                      onClick={() => setSelectedCardIdx(selectedCardIdx === i ? null : i)}
                    >
                      <AnimatedCard
                        card={card}
                        width={cardWidth}
                        isSelected={selectedCardIdx === i}
                        onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                      />
                    </button>
                  ))}
                </div>
              </div>
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {/* Action buttons */}
            <div className="text-center" data-tutorial="dk-action-buttons">
              <button
                type="button"
                className={`${btnOutline} min-w-[90px]`}
                disabled={loading}
                onClick={() =>
                  requestConfirm(() => {
                    hideActionLog();
                    gameExec('reset', undefined, undefined, durakConfig);
                  })
                }
              >
                {tc('button.reset')}
              </button>
              {showAttackBtn && (
                <button
                  type="button"
                  className={`${btnDanger} min-w-[90px]`}
                  disabled={loading || selectedCardIdx === null}
                  onClick={handleAttack}
                >
                  {t('attackButton')}
                </button>
              )}
              {showDefendBtn && (
                <button
                  type="button"
                  className={`${btnSuccess} min-w-[90px]`}
                  disabled={loading || selectedCardIdx === null || selectedAttackIdx === null}
                  onClick={handleDefend}
                >
                  {t('defendButton')}
                </button>
              )}
              {showPassBtn && (
                <button
                  type="button"
                  className={`${btnSecondary} min-w-[90px]`}
                  disabled={loading}
                  onClick={handlePass}
                >
                  {t('passButton')}
                </button>
              )}
              {showTakeBtn && (
                <button type="button" className={`${btnPrimary} min-w-[90px]`} disabled={loading} onClick={handleTake}>
                  {t('takeButton')}
                </button>
              )}
              {/* Sort buttons */}
              {!isGameEnd && (
                <>
                  <button
                    type="button"
                    className={`${btnOutline} min-w-[70px]`}
                    disabled={loading}
                    onClick={() => handleSort(0)}
                  >
                    {t('sort.suit')}
                  </button>
                  <button
                    type="button"
                    className={`${btnOutline} min-w-[70px]`}
                    disabled={loading}
                    onClick={() => handleSort(1)}
                  >
                    {t('sort.value')}
                  </button>
                </>
              )}
            </div>
          </GameFooter>
        </>
      )}
      <WinCelebration
        show={isGameEnd && humanPlayer !== undefined && state.loserIdx >= 0 && state.loserIdx !== humanPlayer.id}
        onCelebrate={() => playSound('winFanfare')}
      />
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
    </div>
  );
}
