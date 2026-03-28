import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { ActionLogSection } from '../components/ActionLogSection';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { DoubtCpuArea } from '../components/doubt/DoubtCpuArea';
import { DoubtHandCard } from '../components/doubt/DoubtHandCard';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageHeading } from '../components/GamePageHeading';
import { GameResetDialog } from '../components/GameResetDialog';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { WinCelebration } from '../components/motion/WinCelebration';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { DoubtSkeleton } from '../components/skeleton/DoubtSkeleton';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import {
  actionDesc,
  CPU_MEMORY_OPTIONS,
  DOUBT_WINDOW_OPTIONS,
  PENALTY_DRAW_LIMIT_OPTIONS,
  useDoubtGame,
} from '../hooks/useDoubtGame';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { TutorialProvider } from '../providers/TutorialProvider';
import { btnDanger, btnOutline, btnPrimary, btnSecondary, btnSuccess, focusRingBlue } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { DoubtCpuAction } from '../types/card';
import type { TutorialConfig, TutorialStep } from '../types/tutorial';
import { valueName } from '../utils/cardUtils';
import { playerName } from '../utils/playerUtils';

/** Doubt tutorial step definitions. */
const DT_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="dt-table-area"]',
    messageKey: 'tutorial.tableArea',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dt-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dt-claim-input"]',
    messageKey: 'tutorial.claimInput',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dt-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dt-doubt-window"]',
    messageKey: 'tutorial.doubtWindow',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dt-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Doubt tutorial configuration. */
const DT_TUTORIAL_CONFIG: TutorialConfig = {
  gameName: 'doubt',
  steps: DT_TUTORIAL_STEPS,
};

/** Renders the Doubt game page with card play, doubt window countdown, and config. */
export function DoubtPage() {
  const { t: tDt } = useTranslation('doubt');
  return (
    <TutorialProvider config={DT_TUTORIAL_CONFIG} translateMessage={tDt}>
      <DoubtPageContent />
    </TutorialProvider>
  );
}

/** Inner content of the Doubt page, wrapped by TutorialProvider. */
function DoubtPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('doubt');
  const {
    state,
    loading,
    error,
    exec,
    countdown,
    doubtConfig,
    selectedCardIndices,
    toggleCard,
    claimedValue,
    setClaimedValue,
    handleConfigChange,
    handleConfigToggle,
    handlePlay,
    handleDoubt,
    handleSkip,
    handleCpuDoubtConfirm,
    clearSelection,
  } = useDoubtGame();

  const { cardWidth } = useCardDimensions();

  const claimInputRef = useRef<HTMLInputElement>(null);
  const [valWarning, setValWarning] = useState(false);
  const warningTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    return () => {
      if (warningTimeoutRef.current) {
        clearTimeout(warningTimeoutRef.current);
      }
    };
  }, []);

  const isHumanTurn = !state?.gameEndFlag && state?.players[state.currentTurn]?.isHuman === true;
  const showClaimInput = selectedCardIndices.length > 0 && isHumanTurn && state?.phase === 0;

  useEffect(() => {
    if (showClaimInput && claimInputRef.current) {
      claimInputRef.current.focus();
    }
  }, [showClaimInput]);

  const isHumanPlayTurn = isHumanTurn && state?.phase === 0;
  const humanPlayer = state?.players.find((p) => p.isHuman);

  useCardKeyboardNav({
    cardCount: humanPlayer?.cards?.length ?? 0,
    onToggle: toggleCard,
    onConfirm: handlePlay,
    onClear: clearSelection,
    enabled: isHumanPlayTurn && !loading,
  });

  if (!state) return <DoubtSkeleton />;

  const cpuPlayers = state.players.filter((p) => !p.isHuman);
  const isDoubtPhase = state.phase === 1;
  const cpuPlayed = isDoubtPhase && state.lastAction !== null && !state.players[state.lastAction.playerIdx]?.isHuman;

  const cpuTells = new Set(
    [...state.cpuActions, state.lastAction]
      .filter((a): a is DoubtCpuAction => a !== null && a.hasTell === true)
      .map((a) => a.playerIdx),
  );

  return (
    <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.doubt.bg}`} aria-busy={loading} aria-live="polite">
      <GamePageHeading title={tc('nav.doubt')} />
      {/* Phase indicator */}
      <PhaseIndicator
        phaseName={state.gameEndFlag ? t('phase.end') : state.phase === 1 ? t('phase.doubt') : t('phase.play')}
        isHumanTurn={isHumanTurn}
      >
        <TutorialButton />
      </PhaseIndicator>
      {/* Settings panel */}
      <SettingsPanel
        title={t('settings.title')}
        groups={[
          {
            items: [
              {
                type: 'select',
                id: 'doubtWindowSec',
                label: t('settings.doubtTime'),
                value: doubtConfig.doubtWindowSec,
                options: DOUBT_WINDOW_OPTIONS.map((sec) => ({ value: sec, label: t('settings.sec', { sec }) })),
                onSelect: (v) => handleConfigChange('doubtWindowSec', v),
              },
              {
                type: 'select',
                id: 'cpuMemoryLevel',
                label: t('settings.cpuMemory'),
                value: doubtConfig.cpuMemoryLevel,
                options: CPU_MEMORY_OPTIONS.map((opt) => ({ value: opt.value, label: opt.label })),
                onSelect: (v) => handleConfigChange('cpuMemoryLevel', v),
              },
              {
                type: 'select',
                id: 'penaltyDrawLimit',
                label: t('settings.penaltyDrawLimit'),
                value: doubtConfig.penaltyDrawLimit,
                options: PENALTY_DRAW_LIMIT_OPTIONS.map((v) => ({
                  value: v,
                  label: v === 0 ? t('settings.unlimited') : t('settings.cards', { count: v }),
                })),
                onSelect: (v) => handleConfigChange('penaltyDrawLimit', v),
              },
              {
                type: 'checkbox',
                id: 'cpuHesitation',
                label: t('settings.cpuHesitation'),
                checked: doubtConfig.cpuHesitationEnabled,
                onToggle: (checked) => handleConfigToggle('cpuHesitationEnabled', checked),
              },
              {
                type: 'checkbox',
                id: 'cpuMetaAI',
                label: t('settings.cpuMetaAI'),
                checked: doubtConfig.cpuMetaAI,
                onToggle: (checked) => handleConfigToggle('cpuMetaAI', checked),
              },
            ],
          },
        ]}
      />

      {/* Scrollable area */}
      <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
        {/* CPU player areas */}
        <div className="flex gap-2 flex-wrap mb-3">
          {cpuPlayers.map((player) => (
            <DoubtCpuArea
              key={player.id}
              player={player}
              isCurrentTurn={state.currentTurn === player.id}
              hasTell={cpuTells.has(player.id)}
            />
          ))}
        </div>

        {/* Table area */}
        <div className="bg-black/30 rounded-[10px] py-2.5 px-3.5 my-2" data-tutorial="dt-table-area">
          <div className="text-white font-bold mb-1">{t('table')}</div>
          <div className="text-game-text-muted text-sm">{t('tableCards', { count: state.tableCardCount })}</div>
          {state.lastAction && (
            <div className="text-yellow-300 text-xs mt-1">{actionDesc(state.lastAction, state.players, t)}</div>
          )}
        </div>

        {/* Doubt/Skip UI */}
        {isDoubtPhase && !state.gameEndFlag && (
          <div className="bg-black/40 rounded-[10px] py-3 px-4 my-2" data-tutorial="dt-doubt-window">
            {cpuPlayed ? (
              <>
                <div className="text-white font-bold mb-2">{t('doubtQuestion')}</div>
                {countdown !== null && (
                  <div className="text-yellow-300 text-lg font-bold mb-2" aria-live="assertive" aria-atomic="true">
                    {t('countdown', { sec: countdown })}
                  </div>
                )}
                {state.cpuDoubters.length > 0 && (
                  <div className="text-game-text-muted text-xs mb-2">
                    {t('cpuDoubters', { names: state.cpuDoubters.map((idx) => playerName(idx, false)).join(', ') })}
                  </div>
                )}
                <div className="flex gap-2">
                  <button type="button" className={btnDanger} disabled={loading} onClick={handleDoubt}>
                    {t('doubtButton')}
                  </button>
                  <button type="button" className={btnSecondary} disabled={loading} onClick={handleSkip}>
                    {t('skipButton')}
                  </button>
                </div>
              </>
            ) : (
              <>
                <div className="text-white font-bold mb-2">{t('cpuJudging')}</div>
                {state.cpuDoubters.length > 0 && (
                  <div className="text-red-300 text-sm mb-2">
                    {t('cpuDoubtExclaim', { names: state.cpuDoubters.map((idx) => playerName(idx, false)).join(', ') })}
                  </div>
                )}
                <button type="button" className={btnPrimary} disabled={loading} onClick={handleCpuDoubtConfirm}>
                  {t('confirmButton')}
                </button>
              </>
            )}
          </div>
        )}

        {/* Last doubt result */}
        {state.lastDoubtResult && (
          <div className="bg-black/40 rounded-lg py-2 px-3.5 my-2 text-xs">
            <div className="text-white font-bold mb-1">{t('doubtResult.title')}</div>
            <div className={state.lastDoubtResult.wasLying ? 'text-red-300' : 'text-green-300'}>
              {state.lastDoubtResult.wasLying ? t('doubtResult.wasLying') : t('doubtResult.wasTruth')}
            </div>
            <div className="text-game-text-muted">
              {t('doubtResult.loserTook', {
                name: playerName(
                  state.players[state.lastDoubtResult.loserIdx]?.id ?? state.lastDoubtResult.loserIdx,
                  state.players[state.lastDoubtResult.loserIdx]?.isHuman ?? false,
                ),
                count: state.lastDoubtResult.cardCount,
              })}
            </div>
            {state.lastDoubtResult.discardedCount > 0 && (
              <div className="text-yellow-300">
                {t('doubtResult.discarded', { count: state.lastDoubtResult.discardedCount })}
              </div>
            )}
            {state.lastDoubtResult.revealedCards.length > 0 && (
              <div className="flex flex-wrap gap-1 mt-1">
                {state.lastDoubtResult.revealedCards.map((card, i) => (
                  <AnimatedCard key={`${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
                ))}
              </div>
            )}
          </div>
        )}

        {/* Human/CPU action logs */}
        {state.humanAction && !isDoubtPhase && (
          <div className="bg-black/40 rounded-lg text-game-text-highlight py-2 px-3.5 my-2 text-xs">
            {actionDesc(state.humanAction, state.players, t)}
          </div>
        )}
        {state.cpuActions && state.cpuActions.length > 0 && (
          <div className="bg-black/40 rounded-lg text-game-text-muted py-2 px-3.5 my-2 whitespace-pre-line text-xs">
            {[tc('label.cpuActions'), ...state.cpuActions.map((a) => actionDesc(a, state.players, t))].join('\n')}
          </div>
        )}

        {/* Meta-AI info */}
        {state.metaAI?.enabled && (
          <div className="bg-black/40 rounded-lg py-2 px-3.5 my-2 text-xs">
            <div className="text-white font-bold mb-1">{t('metaAI.title')}</div>
            <div className="text-game-text-muted">{t('metaAI.gamesPlayed', { count: state.metaAI.gamesPlayed })}</div>
            <div className="text-game-text-muted">
              {t('metaAI.bluffRate', { rate: `${(state.metaAI.bluffRate * 100).toFixed(0)}%` })}
            </div>
            <div className="text-game-text-muted">
              {t('metaAI.doubtAccuracy', { rate: `${(state.metaAI.doubtAccuracy * 100).toFixed(0)}%` })}
            </div>
            {state.metaAI.hesitationMean > 0 && (
              <div className="text-game-text-muted">
                {t('metaAI.hesitationMean', { ms: Math.round(state.metaAI.hesitationMean) })}
              </div>
            )}
          </div>
        )}

        {/* Result message */}
        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

        {/* Action log */}
        <ActionLogSection
          isEndPhase={state.gameEndFlag}
          actionLog={actionLog}
          showActionLog={showActionLog}
          hideActionLog={hideActionLog}
        />
      </div>

      {/* Sticky footer: human player hand + action buttons */}
      <GameFooter className={`${gameTheme.doubt.footer} px-4 py-2.5`}>
        {/* Human player info */}
        {humanPlayer && (
          <div className="mb-2" data-tutorial="dt-player-hand">
            <div className="text-white font-bold text-sm mb-1">
              {t('yourCards', { count: humanPlayer.cardCount })}
              {isHumanTurn && state.phase === 0 && (
                <span className="text-green-400 text-xs ml-2">{t('selectPrompt')}</span>
              )}
            </div>
            {/* Human cards */}
            <div className="flex flex-wrap gap-1">
              {humanPlayer.cards?.map((card, i) => (
                <DoubtHandCard
                  key={`${card.design}-${card.value}-${i}`}
                  card={card}
                  index={i}
                  selected={selectedCardIndices.includes(i)}
                  selectable={isHumanTurn && state.phase === 0 && !loading}
                  onToggle={toggleCard}
                />
              ))}
            </div>

            {/* Claimed value input (shown when cards are selected) */}
            {showClaimInput && (
              <div className="mt-2 flex items-center gap-2" data-tutorial="dt-claim-input">
                <label htmlFor="claim-input" className="text-white text-sm">
                  {t('claimedValue')}
                </label>
                <input
                  ref={claimInputRef}
                  type="number"
                  min={1}
                  max={13}
                  value={claimedValue}
                  id="claim-input"
                  aria-describedby="claim-range-hint"
                  onChange={(e) => {
                    const num = Number(e.target.value);
                    const clamped = Math.max(1, Math.min(13, num));
                    setClaimedValue(clamped);
                    if (num !== clamped) {
                      if (warningTimeoutRef.current) {
                        clearTimeout(warningTimeoutRef.current);
                      }
                      setValWarning(true);
                      warningTimeoutRef.current = setTimeout(() => setValWarning(false), 2000);
                    }
                  }}
                  className={`bg-black/50 text-white rounded px-2 py-1 w-16 text-sm border border-white/30 ${focusRingBlue}`}
                />
                <span className="text-game-text-muted text-xs">({valueName(claimedValue)})</span>
                <span id="claim-range-hint" className={`text-xs ${valWarning ? 'text-yellow-400' : 'text-gray-400'}`}>
                  {valWarning ? t('claimRangeWarning') : t('claimRangeHint')}
                </span>
              </div>
            )}
          </div>
        )}

        <ErrorAlert message={error} />

        {/* Action buttons */}
        <div className="text-center">
          <button
            type="button"
            className={`${btnOutline} min-w-[90px]`}
            disabled={loading}
            data-tutorial="dt-reset-button"
            onClick={() =>
              requestConfirm(() => {
                hideActionLog();
                exec('reset', undefined, undefined, undefined, doubtConfig);
              })
            }
          >
            {tc('button.reset')}
          </button>
          {isHumanTurn && state.phase === 0 && (
            <button
              type="button"
              className={`${btnSuccess} min-w-[90px]`}
              disabled={loading || selectedCardIndices.length === 0}
              onClick={handlePlay}
              data-tutorial="dt-play-button"
            >
              {t('playButton')}
            </button>
          )}
        </div>
      </GameFooter>
      <WinCelebration show={!!state?.gameEndFlag} />
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
    </div>
  );
}
