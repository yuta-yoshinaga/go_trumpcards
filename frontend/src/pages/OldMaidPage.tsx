import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardBack, CardImage } from '../components/CardImage';
import { ConfirmDialog } from '../components/ConfirmDialog';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { LoadingSpinner } from '../components/LoadingSpinner';
import { OldMaidDiscardedArea } from '../components/oldmaid/OldMaidDiscardedArea';
import { OldMaidDrawHistory } from '../components/oldmaid/OldMaidDrawHistory';
import { OldMaidPlayerArea } from '../components/oldmaid/OldMaidPlayerArea';
import { OldMaidSetupScreen } from '../components/oldmaid/OldMaidSetupScreen';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useActionLog } from '../hooks/useActionLog';
import { useConfirmDialog } from '../hooks/useConfirmDialog';
import { OldMaidMode, useOldMaidGame } from '../hooks/useOldMaidGame';
import { btnPrimary, btnSecondary, btnWarning } from '../styles/buttonStyles';
import type { CpuAction } from '../types/card';
import { cardLabel } from '../utils/cardUtils';
import { findPlayerName } from '../utils/playerUtils';

export function OldMaidPage() {
  const { t } = useTranslation('oldmaid');
  const { t: tc } = useTranslation('common');
  const {
    displayState,
    setupMode,
    setupStrategy,
    setupMemoryAI,
    setupHesitation,
    setupMetaAI,
    gameSettings,
    suspectPins,
    setSuspectPins,
    shakeKey,
    revealedCard,
    loading,
    error,
    exec,
    handleStart,
    handleReset,
    handleReorder,
    setSetupMode,
    setSetupStrategy,
    setSetupMemoryAI,
    setSetupHesitation,
    setSetupMetaAI,
    setGameSettings,
  } = useOldMaidGame();

  const { actionLog, showActionLog, hideActionLog } = useActionLog('oldmaid');
  const { isOpen: confirmOpen, requestConfirm, confirm: confirmReset, cancel: cancelReset } = useConfirmDialog();

  const isHumanTurnForKbd =
    !!displayState && !displayState.gameEndFlag && !!displayState.players[displayState.currentTurn]?.isHuman;

  const actionBindings = useMemo(
    () => [
      { key: 'd', action: () => exec('draw') },
      { key: 's', action: () => exec('shuffle') },
    ],
    [exec],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!gameSettings && isHumanTurnForKbd && !loading,
  });

  if (!gameSettings) {
    return (
      <OldMaidSetupScreen
        mode={setupMode}
        cpuPlacementStrategy={setupStrategy}
        cpuMemoryAI={setupMemoryAI}
        cpuHesitationEnabled={setupHesitation}
        cpuMetaAI={setupMetaAI}
        onModeChange={setSetupMode}
        onStrategyChange={setSetupStrategy}
        onMemoryAIChange={setSetupMemoryAI}
        onHesitationChange={setSetupHesitation}
        onMetaAIChange={setSetupMetaAI}
        onStart={handleStart}
        loading={loading}
      />
    );
  }

  if (!displayState) return null;

  const state = displayState;
  const isHumanTurn = !state.gameEndFlag && !!state.players[state.currentTurn]?.isHuman;
  const cpuPlayers = state.players.filter((p) => !p.isHuman);
  const humanPlayer = state.players.find((p) => p.isHuman);

  const statusLines: string[] = [];
  if (!state.gameEndFlag && state.hasDrawn) {
    const from = findPlayerName(state.players, state.lastDrawPlayerIdx);
    const target = findPlayerName(state.players, state.lastDrawFromIdx);
    let msg = state.lastDrawCard
      ? t('drewCardWithLabel', { from, target, card: cardLabel(state.lastDrawCard) })
      : t('drewCard', { from, target });
    if (state.lastDiscardedPairs > 0) msg += t('discardedPairs', { count: state.lastDiscardedPairs });
    statusLines.push(msg);
  }
  if (isHumanTurn) {
    statusLines.push(t('yourTurn', { target: findPlayerName(state.players, state.nextDrawTargetIdx) }));
  }

  return (
    <div
      key={shakeKey}
      className={`flex-1 flex flex-col min-h-0 bg-[#1a5c1a]${shakeKey > 0 ? ' animate-shake' : ''}`}
      aria-busy={loading}
      aria-live="polite"
    >
      <LoadingSpinner loading={loading} />
      {/* Scrollable: CPU rows + discard + status + logs + result */}
      <div className="flex-1 overflow-y-auto pt-3 px-4">
        {/* Mode badge */}
        {state.mode === OldMaidMode.JijiNuki && (
          <div className="text-center mb-1">
            <span className="inline-block rounded-md bg-red-600 px-2.5 py-0.5 text-sm font-bold text-white">
              {t('badge.jijiNuki')}
            </span>
          </div>
        )}

        {/* CPU row */}
        <div className="flex gap-2 flex-wrap mb-2 justify-center">
          {cpuPlayers.map((player) => (
            <OldMaidPlayerArea
              key={player.id}
              player={player}
              isTarget={state.nextDrawTargetIdx === player.id}
              isHumanTurn={isHumanTurn}
              gameEndFlag={state.gameEndFlag}
              loading={loading}
              highlightedCardIdx={state.nextDrawTargetIdx === player.id ? state.cpuHighlightedCardIdx : -1}
              isSuspect={suspectPins.has(player.id)}
              onToggleSuspect={() =>
                setSuspectPins((prev) => {
                  const next = new Set(prev);
                  if (next.has(player.id)) {
                    next.delete(player.id);
                  } else {
                    next.add(player.id);
                  }
                  return next;
                })
              }
              onDraw={(drawIdx) => exec('draw', drawIdx)}
            />
          ))}
        </div>

        {/* Discarded Area */}
        <OldMaidDiscardedArea cards={state.lastDiscardedCards} />

        {/* Card reveal area */}
        {state.lastDrawCard && !state.gameEndFlag && (
          <div className="flex justify-center my-2" data-testid="card-reveal-area">
            {revealedCard ? (
              <div className="animate-flipIn">
                <CardImage card={revealedCard} width={60} />
              </div>
            ) : (
              <CardBack width={60} />
            )}
          </div>
        )}

        {/* Status */}
        {statusLines.length > 0 && (
          <div className="bg-black/50 rounded-lg text-white py-2 px-3 my-2 whitespace-pre-line text-sm">
            {statusLines.join('\n')}
          </div>
        )}

        {/* CPU log */}
        {state.cpuActions && state.cpuActions.length > 0 && (
          <div className="bg-black/40 rounded-lg text-game-text-muted py-1.5 px-2.5 my-1.5 whitespace-pre-line text-xs max-h-[120px] overflow-y-auto">
            {[
              tc('label.cpuActions'),
              ...state.cpuActions.map((action: CpuAction) => {
                const from = findPlayerName(state.players, action.drawPlayerIdx);
                const target = findPlayerName(state.players, action.drawFromIdx);
                let msg = t('drewCard', { from, target });
                // CPU drawn card is intentionally hidden to preserve game fairness
                if (action.discardedPairs > 0) msg += t('discardedPairs', { count: action.discardedPairs });
                return msg;
              }),
            ].join('\n')}
          </div>
        )}

        {/* Draw History Timeline */}
        <OldMaidDrawHistory entries={state.drawHistory ?? []} players={state.players} />

        {/* Result */}
        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

        {/* JijiNuki: show removed card at game end */}
        {state.gameEndFlag && state.removedCard && (
          <div className="text-center my-2 text-white text-sm">
            {t('removedCard', { card: cardLabel(state.removedCard) })}
          </div>
        )}

        {/* Action log */}
        <ActionLogSection
          isEndPhase={state.gameEndFlag}
          actionLog={actionLog}
          showActionLog={showActionLog}
          hideActionLog={hideActionLog}
        />
      </div>

      {/* Sticky footer: human player hand + buttons */}
      <GameFooter className="bg-[#163e16] border-white/20 px-4 py-2.5">
        {/* Human player */}
        {humanPlayer && (
          <div className="mb-2">
            <OldMaidPlayerArea
              player={humanPlayer}
              isTarget={false}
              isHumanTurn={isHumanTurn}
              gameEndFlag={state.gameEndFlag}
              loading={loading}
              highlightedCardIdx={-1}
              onDraw={(drawIdx) => exec('draw', drawIdx)}
              onReorder={handleReorder}
            />
          </div>
        )}

        <ErrorAlert message={error} />

        {/* Buttons */}
        <div className="text-center">
          <button
            type="button"
            className={`${btnSecondary} min-w-[80px]`}
            disabled={loading}
            onClick={() => {
              setGameSettings(null);
              setSuspectPins(new Set());
            }}
          >
            {t('button.settings')}
          </button>
          <button
            type="button"
            className={`${btnPrimary} min-w-[80px]`}
            disabled={loading}
            onClick={() =>
              requestConfirm(() => {
                hideActionLog();
                handleReset();
              })
            }
          >
            {tc('button.reset')}
          </button>
          <button
            type="button"
            className={`${btnWarning} min-w-[110px]`}
            disabled={loading || !isHumanTurn || state.gameEndFlag}
            onClick={() => exec('draw')}
          >
            {t('button.drawRandom')}
          </button>
          <button
            type="button"
            className={`${btnSecondary} min-w-[110px]`}
            disabled={loading || state.gameEndFlag}
            onClick={() => exec('shuffle')}
          >
            {t('button.shuffle')}
          </button>
        </div>
      </GameFooter>
      <ConfirmDialog
        open={confirmOpen}
        title={tc('button.confirmReset')}
        message={tc('button.confirmResetMessage')}
        confirmLabel={tc('button.confirm')}
        cancelLabel={tc('button.cancel')}
        onConfirm={confirmReset}
        onCancel={cancelReset}
      />
    </div>
  );
}
