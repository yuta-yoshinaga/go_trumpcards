import { useEffect, useMemo, useState } from 'react';
import type { pinochleApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageHeading } from '../components/GamePageHeading';
import { GameResetDialog } from '../components/GameResetDialog';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { WinCelebration } from '../components/motion/WinCelebration';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, usePinochleGame } from '../hooks/usePinochleGame';
import { useSound } from '../providers/SoundProvider';
import { btnOutline, btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { PinochleMeldData, PinochleResponse } from '../types/card';
import { PinochlePhase } from '../types/phases';
import { cardAlt } from '../utils/cardAlt';
import { PINOCHLE_HELP, parsePinochleCommand } from '../utils/cli/commands/pinochleCommands';
import { formatPinochleState } from '../utils/cli/formatters/pinochleFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';

/** Phase name keys for Pinochle. */
const PINOCHLE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [PinochlePhase.BID]: 'bid',
  [PinochlePhase.TRUMP]: 'trump',
  [PinochlePhase.MELD]: 'meld',
  [PinochlePhase.PLAY]: 'play',
  [PinochlePhase.TRICK_END]: 'trickEnd',
  [PinochlePhase.ROUND_END]: 'roundEnd',
  [PinochlePhase.GAME_END]: 'gameEnd',
};

/** Suit labels for display. */
const SUIT_LABELS: Record<number, string> = { 1: '♠', 2: '♣', 3: '♥', 4: '♦' };

/** Pinochle game page. */
export function PinochlePage() {
  return <PinochlePageContent />;
}

/** Inner content of the Pinochle page. */
function PinochlePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('pinochle');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    pinochleConfig,
    handleConfigChange,
    handleReset,
    handleBid,
    handlePass,
    handleCallTrump,
    handleConfirmMelds,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  } = usePinochleGame();

  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  const phaseNames = usePhaseNames('pinochle', PINOCHLE_PHASE_KEYS);

  const [bidAmount, setBidAmount] = useState(20);

  // Auto-update bid amount when highest bid changes
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('pinochle');
  const cliConfig: CliGameConfig<PinochleResponse, Parameters<typeof pinochleApi.exec>> = useMemo(
    () => ({
      gameName: 'pinochle',
      parseCommand: parsePinochleCommand,
      formatResponse: formatPinochleState,
      helpText: PINOCHLE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useEffect(() => {
    if (state?.highestBid && state.highestBid > 0) {
      setBidAmount(state.highestBid + 1);
    } else {
      setBidAmount(20);
    }
  }, [state?.highestBid]);

  if (!state) {
    return (
      <div className="p-4 text-center text-white">
        <p>{tc('status.thinking')}</p>
      </div>
    );
  }

  const phase = state.phase;
  const humanPlayer = state.players?.find((p) => p.isHuman);
  const isBidTurn = phase === PinochlePhase.BID && state.players?.[state.bidPlayerIdx]?.isHuman;
  const isTrumpTurn = phase === PinochlePhase.TRUMP && state.players?.[state.currentPlayerIdx]?.isHuman;
  const isPlayTurn = phase === PinochlePhase.PLAY && state.players?.[state.currentPlayerIdx]?.isHuman;
  const isGameEnd = phase === PinochlePhase.GAME_END || state.gameEndFlag;

  return (
    <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.pinochle.bg}`} aria-busy={loading} aria-live="polite">
      <GamePageHeading title={tc('nav.pinochle')} />
      <PhaseIndicator phaseName={phaseNames[phase]} isHumanTurn={isBidTurn || isTrumpTurn || isPlayTurn}>
        <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
      </PhaseIndicator>

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
                    type: 'select' as const,
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: pinochleConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({ value: o.value, label: o.label })),
                    onSelect: (v: string) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select' as const,
                    id: 'pointLimit',
                    label: t('settings.pointLimit'),
                    value: pinochleConfig.pointLimit,
                    options: POINT_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v: string) => handleConfigChange('pointLimit', v),
                  },
                ],
              },
            ]}
          />

          <div className="flex-1 overflow-y-auto pt-3 px-4 lg:px-8">
            {/* Game Info */}
            <div className="text-white text-center mb-2 text-sm">
              <span className="mr-4">
                {t('round')}: {state.roundNumber} / {t('trick')}: {state.trickNumber}
              </span>
              <span className="mr-4">
                {t('team')} 0: {state.teamScores[0]} / {t('team')} 1: {state.teamScores[1]}
              </span>
              <span>
                {t('trumpSuit')}: {state.trumpSuit > 0 ? SUIT_LABELS[state.trumpSuit] : '-'}
              </span>
              {state.highestBid > 0 && (
                <span className="ml-4">
                  {t('highestBid')}: {state.highestBid}
                </span>
              )}
            </div>

            {/* Players Info */}
            <div className="grid grid-cols-2 gap-2 mb-3">
              {state.players?.map((p) => (
                <div
                  key={p.id}
                  className={`rounded p-2 text-sm ${p.isHuman ? 'bg-ds-accent/20 text-ds-accent' : 'bg-black/30 text-white/70'}`}
                >
                  <div className="font-bold">{playerName(p.id, p.isHuman)}</div>
                  <div>
                    {t('team')} {p.team} | {t('bid')}: {p.bid} | {t('meldScore')}: {p.meldScore} | T: {p.trickCount}
                  </div>
                </div>
              ))}
            </div>

            {/* Current Trick */}
            {state.currentTrick?.length > 0 && (
              <div className="mb-3 p-2 rounded bg-black/40">
                <div className="text-white/70 text-sm mb-1">{tc('common:table', { defaultValue: 'Table' })}:</div>
                <div className="flex gap-2 justify-center">
                  {state.currentTrick.map((tc, i) => (
                    <div key={i} className="text-center">
                      <AnimatedCard
                        card={tc.card}
                        width={cardWidth * 0.8}
                        onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                      />
                      <div className="text-xs text-white/50 mt-1">P{tc.playerIdx}</div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Melds */}
            {(phase === PinochlePhase.MELD || phase === PinochlePhase.ROUND_END) && state.playerMelds && (
              <div className="mb-3 p-2 rounded bg-purple-900/30">
                <div className="text-white font-bold mb-1">{t('meldScore')}:</div>
                {state.playerMelds.map((melds: PinochleMeldData[], pIdx: number) =>
                  melds.length > 0 ? (
                    <div key={pIdx} className="text-white/70 text-sm mb-1">
                      <span className="font-semibold">{playerName(pIdx, state.players[pIdx]?.isHuman)}: </span>
                      {melds.map((m: PinochleMeldData, mIdx: number) => (
                        <span key={mIdx} className="mr-2">
                          {t(`meldTypes.${m.type}`)} ({m.points})
                        </span>
                      ))}
                    </div>
                  ) : null,
                )}
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

          <GameFooter className={`${gameTheme.pinochle.footer} px-4 py-2.5`}>
            {/* Hand */}
            {humanPlayer && humanPlayer.cards.length > 0 && (
              <div className="flex flex-wrap gap-1 mb-2">
                {humanPlayer.cards.map((card, idx) => {
                  const isValid = state.validPlayIndices?.includes(idx);
                  return (
                    <button
                      type="button"
                      key={`${card.design}-${card.value}-${idx}`}
                      onClick={() => isPlayTurn && isValid && handlePlay(idx)}
                      disabled={loading || !isPlayTurn || !isValid}
                      aria-label={cardAlt(card)}
                      className="transition-transform"
                      style={{
                        background: 'none',
                        padding: 0,
                        borderRadius: 8,
                        opacity: isPlayTurn && !isValid ? 0.5 : 1,
                        boxSizing: 'border-box',
                      }}
                    >
                      <AnimatedCard
                        card={card}
                        width={cardWidth}
                        onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                      />
                    </button>
                  );
                })}
              </div>
            )}

            <ErrorAlert message={error} onRetry={retry} />

            <div className="flex gap-2 items-center flex-wrap">
              {/* Bid */}
              {isBidTurn && (
                <>
                  <input
                    type="number"
                    min={state.highestBid > 0 ? state.highestBid + 1 : 20}
                    value={bidAmount}
                    onChange={(e) => setBidAmount(Number(e.target.value))}
                    className="border rounded px-2 py-1 w-20 text-sm"
                  />
                  <button type="button" className={btnPrimary} onClick={() => handleBid(bidAmount)} disabled={loading}>
                    {t('bid')}
                  </button>
                  <button type="button" className={btnOutline} onClick={handlePass} disabled={loading}>
                    {t('pass')}
                  </button>
                </>
              )}

              {/* Trump */}
              {isTrumpTurn &&
                [1, 2, 3, 4].map((suit) => (
                  <button
                    key={suit}
                    type="button"
                    className={btnPrimary}
                    onClick={() => handleCallTrump(suit)}
                    disabled={loading}
                  >
                    {SUIT_LABELS[suit]}
                  </button>
                ))}

              {/* Meld confirm */}
              {phase === PinochlePhase.MELD && (
                <button type="button" className={btnSuccess} onClick={handleConfirmMelds} disabled={loading}>
                  {t('confirmMelds')}
                </button>
              )}

              {/* Trick End */}
              {phase === PinochlePhase.TRICK_END && (
                <button type="button" className={btnSuccess} onClick={handleNextTrick} disabled={loading}>
                  {t('nextTrick')}
                </button>
              )}

              {/* Round End */}
              {phase === PinochlePhase.ROUND_END && (
                <button type="button" className={btnSuccess} onClick={handleNextRound} disabled={loading}>
                  {t('nextRound')}
                </button>
              )}

              {/* Reset */}
              <button
                type="button"
                className={btnOutline}
                onClick={() =>
                  requestConfirm(() => {
                    hideActionLog();
                    handleReset();
                  })
                }
              >
                {tc('button.reset')}
              </button>
            </div>
          </GameFooter>
        </>
      )}
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
      <WinCelebration show={isGameEnd} onCelebrate={() => playSound('winFanfare')} />
    </div>
  );
}
