import { useCallback, useState } from 'react';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
import { ConfirmDialog } from '../components/ConfirmDialog';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { SpadesSkeleton } from '../components/skeleton/SpadesSkeleton';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useSpadesGame } from '../hooks/useSpadesGame';
import { btnPrimary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { selectedCardStyle } from '../styles/cardStyles';
import { SpadesPhase } from '../types/phases';
import { cardAlt } from '../utils/cardAlt';
import { playerName } from '../utils/playerUtils';

const SPADES_PHASE_KEYS: Readonly<Record<number, string>> = {
  [SpadesPhase.BID]: 'bid',
  [SpadesPhase.PLAY]: 'play',
  [SpadesPhase.TRICK_END]: 'trickEnd',
  [SpadesPhase.ROUND_END]: 'roundEnd',
  [SpadesPhase.GAME_END]: 'gameEnd',
};

export function SpadesPage() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('spades');
  const {
    state,
    loading,
    error,
    exec,
    spadesConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleBid,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  } = useSpadesGame();
  const { cardWidth } = useCardDimensions();
  const [bidValue, setBidValue] = useState(1);

  const isPlayPhaseForKbd = state?.phase === SpadesPhase.PLAY;
  const isHumanTurnForKbd = isPlayPhaseForKbd && state?.players[state.currentPlayerIdx]?.isHuman === true;
  const humanCardCountForKbd = state?.players.find((p) => p.isHuman)?.cards?.length ?? 0;

  const confirmAction = useCallback(() => {
    handlePlay();
  }, [handlePlay]);

  useCardKeyboardNav({
    cardCount: humanCardCountForKbd,
    onToggle: toggleCard,
    onConfirm: confirmAction,
    onClear: clearSelection,
    enabled: !!isHumanTurnForKbd && !loading,
  });

  const phaseNames = usePhaseNames('spades', SPADES_PHASE_KEYS);

  if (!state) return <SpadesSkeleton />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isBidPhase = state.phase === SpadesPhase.BID;
  const isPlayPhase = state.phase === SpadesPhase.PLAY;
  const isTrickEnd = state.phase === SpadesPhase.TRICK_END;
  const isRoundEnd = state.phase === SpadesPhase.ROUND_END;
  const isGameEnd = state.phase === SpadesPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isHumanBidTurn = isBidPhase && state.players[state.bidPlayerIdx]?.isHuman === true;

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-game-bg-blue" aria-busy={loading}>
      {/* Phase indicator */}
      <PhaseIndicator phaseName={phaseNames[state.phase]} isHumanTurn={isHumanBidTurn || isHumanTurn} />

      {/* Settings */}
      <SettingsPanel
        title={t('settings.title')}
        groups={[
          {
            items: [
              {
                type: 'select',
                id: 'cpuDifficulty',
                label: t('settings.cpuDifficulty'),
                value: spadesConfig.cpuDifficulty,
                options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                  value: o.value,
                  label: t(`settings.${o.label.toLowerCase()}`),
                })),
                onSelect: (v) => handleConfigChange('cpuDifficulty', v),
              },
              {
                type: 'select',
                id: 'pointLimit',
                label: t('settings.pointLimit'),
                value: spadesConfig.pointLimit,
                options: POINT_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                onSelect: (v) => handleConfigChange('pointLimit', v),
              },
            ],
          },
        ]}
      />

      {/* Scrollable area */}
      <div className="flex-1 overflow-y-auto pt-3 px-4">
        {/* Round/Trick info */}
        <div className="text-white text-center mb-2">
          <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
          <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
          <span>{state.spadesBroken ? t('spadesBroken') : t('spadesNotBroken')}</span>
        </div>

        {/* Bid phase instruction */}
        {isHumanBidTurn && <div className="text-yellow-300 text-center mb-2">{t('bidPhase')}</div>}

        {/* CPU players */}
        {state.players
          .filter((p) => !p.isHuman)
          .map((p) => (
            <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
              <div className="text-white/70 text-sm">
                {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                {t('cumulativeScore', { score: p.cumulativeScore })} | {t('roundScore', { score: p.roundScore })} |{' '}
                {p.bid >= 0 ? t('bid', { n: p.bid }) : t('bidNone')} | {t('bags', { count: p.bags })}
              </div>
            </div>
          ))}

        {/* Current trick */}
        {state.currentTrick.length > 0 && (
          <div className="my-3 p-3 rounded bg-black/40">
            <div className="text-white/70 text-sm mb-1">{t('currentTrick')}</div>
            <div className="flex gap-2">
              {state.currentTrick.map((trickCard) => (
                <div key={`trick-${trickCard.playerIdx}`} className="text-center">
                  <CardImage card={trickCard.card} width={cardWidth} />
                  <div className="text-white/50 text-xs mt-1">
                    {playerName(
                      state.players[trickCard.playerIdx]?.id ?? trickCard.playerIdx,
                      state.players[trickCard.playerIdx]?.isHuman ?? false,
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Score table */}
        <div className="my-3 p-2 rounded bg-black/30">
          <div className="text-white/70 text-sm mb-1">{t('scores')}</div>
          <table className="w-full text-sm text-white/70">
            <thead>
              <tr>
                <th scope="col" className="text-left">
                  {t('scoresPlayer')}
                </th>
                <th scope="col">{t('scoresBid')}</th>
                <th scope="col">{t('scoresTricks')}</th>
                <th scope="col">{t('scoresBags')}</th>
                <th scope="col">{t('scoresRound')}</th>
                <th scope="col">{t('scoresTotal')}</th>
              </tr>
            </thead>
            <tbody>
              {state.players.map((p) => (
                <tr key={p.id} className={p.isHuman ? 'text-yellow-300' : ''}>
                  <td>{playerName(p.id, p.isHuman)}</td>
                  <td className="text-center">{p.bid >= 0 ? p.bid : '-'}</td>
                  <td className="text-center">{p.trickCount}</td>
                  <td className="text-center">{p.bags}</td>
                  <td className="text-center">{p.roundScore}</td>
                  <td className="text-center">{p.cumulativeScore}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {/* Message */}
        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

        {/* Action log */}
        <ActionLogSection
          isEndPhase={isGameEnd}
          actionLog={actionLog}
          showActionLog={showActionLog}
          hideActionLog={hideActionLog}
        />
      </div>

      {/* Footer */}
      <GameFooter className="bg-game-bg-blue-dark border-white/20 px-4 py-2.5">
        {/* Human cards */}
        {humanPlayer && (
          <div className="flex flex-wrap gap-1 mb-2">
            {humanPlayer.cards.map((card, idx) => (
              <button
                type="button"
                key={`${card.design}-${card.value}-${idx}`}
                onClick={() => toggleCard(idx)}
                aria-label={cardAlt(card)}
                aria-pressed={selectedCardIndices.includes(idx)}
                className="transition-transform"
                style={{
                  background: 'none',
                  padding: 0,
                  borderRadius: 8,
                  ...selectedCardStyle(selectedCardIndices.includes(idx)),
                  boxSizing: 'border-box',
                }}
              >
                <CardImage card={card} width={cardWidth} />
              </button>
            ))}
          </div>
        )}

        <ErrorAlert message={error} />

        <div className="flex gap-2 items-center">
          {isHumanBidTurn && (
            <>
              <input
                type="number"
                min={0}
                max={13}
                value={bidValue}
                onChange={(e) => setBidValue(Number(e.target.value))}
                className="w-16 px-2 py-1 rounded bg-white/20 text-white text-center"
                aria-label="bid-input"
              />
              <button type="button" className={btnPrimary} onClick={() => handleBid(bidValue)} disabled={loading}>
                {t('bidButton')}
              </button>
            </>
          )}
          {isHumanTurn && (
            <button
              type="button"
              className={btnPrimary}
              onClick={handlePlay}
              disabled={loading || selectedCardIndices.length !== 1}
            >
              {t('playButton')}
            </button>
          )}
          {isTrickEnd && (
            <button type="button" className={btnSuccess} onClick={handleNextTrick} disabled={loading}>
              {t('nextTrick')}
            </button>
          )}
          {isRoundEnd && (
            <button type="button" className={btnSuccess} onClick={handleNextRound} disabled={loading}>
              {t('nextRound')}
            </button>
          )}
          <button
            type="button"
            className={btnWarning}
            onClick={() =>
              requestConfirm(() => {
                hideActionLog();
                return exec('reset', undefined, undefined, {
                  cpuDifficulty: spadesConfig.cpuDifficulty,
                  pointLimit: spadesConfig.pointLimit,
                  nilBonus: spadesConfig.nilBonus,
                  bagPenaltyThreshold: spadesConfig.bagPenaltyThreshold,
                });
              })
            }
            disabled={loading}
          >
            {tc('button.reset')}
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
