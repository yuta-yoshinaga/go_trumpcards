import { useCallback, useEffect, useMemo, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { pitchApi } from '../api/gameApi';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageHeading } from '../components/GamePageHeading';
import { GameResetButton } from '../components/GameResetButton';
import { GameResetDialog } from '../components/GameResetDialog';
import { HintTooltip } from '../components/hint/HintTooltip';
import { ManualButton } from '../components/ManualButton';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { isGameRoundActive, useGameRoundGuard } from '../hooks/useGameRoundGuard';
import { useMountReset } from '../hooks/useMountReset';
import { btnDanger, btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { PitchResponse } from '../types/card';
import { PitchPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { PITCH_HELP, parsePitchCommand } from '../utils/cli/commands/pitchCommands';
import { formatPitchState } from '../utils/cli/formatters/pitchFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { findPlayerName, playerName } from '../utils/playerUtils';

const PT_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="pt-bid-controls"]',
    messageKey: 'tutorial.bidControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pt-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pt-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pt-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pt-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pt-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const PHASE_KEYS: Readonly<Record<number, string>> = {
  [PitchPhase.BID]: 'bid',
  [PitchPhase.PLAY]: 'play',
  [PitchPhase.TRICK_END]: 'trickEnd',
  [PitchPhase.ROUND_END]: 'roundEnd',
  [PitchPhase.GAME_END]: 'gameEnd',
};

const SUIT_LABELS: Readonly<Record<number, string>> = {
  0: '—',
  1: '♠',
  2: '♣',
  3: '♥',
  4: '♦',
};

const SUIT_DESIGNS: Readonly<Record<string, string>> = {
  SPADE: '♠',
  CLUB: '♣',
  HEART: '♥',
  DIAMOND: '♦',
};

const VALUE_LABELS: Readonly<Record<number, string>> = {
  1: 'A',
  11: 'J',
  12: 'Q',
  13: 'K',
};

function cardLabel(c: { design: string; value: number }): string {
  const v = VALUE_LABELS[c.value] ?? String(c.value);
  return `${SUIT_DESIGNS[c.design] ?? c.design}${v}`;
}

function isRedSuit(design: string): boolean {
  return design === 'HEART' || design === 'DIAMOND';
}

/** Pitch (Setback) game page. */
export const PitchPage = withTutorial(PitchPageContent, 'pitch', PT_TUTORIAL_STEPS);

function PitchPageContent() {
  const { t } = useTranslation('pitch');
  const { tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('pitch');
  const { state, loading, error, exec: execApi, retry } = useGameApi(pitchApi.exec);
  const { hint, hintEnabled, setHintEnabled } = useGameHint('pitch', state);
  const [selectedCardIdx, setSelectedCardIdx] = useState<number | null>(null);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('pitch');
  const cliConfig: CliGameConfig<PitchResponse, Parameters<typeof pitchApi.exec>> = useMemo(
    () => ({
      gameName: 'pitch',
      parseCommand: parsePitchCommand,
      formatResponse: formatPitchState,
      helpText: PITCH_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);
  useGameRoundGuard(isGameRoundActive(state));

  // clear card selection when phase or trick changes
  // biome-ignore lint/correctness/useExhaustiveDependencies: phase/trick deps drive re-evaluation
  useEffect(() => {
    setSelectedCardIdx(null);
  }, [state?.trickNumber, state?.phase]);

  const isBidPhase = state?.phase === PitchPhase.BID;
  const isPlayPhase = state?.phase === PitchPhase.PLAY;
  const isTrickEnd = state?.phase === PitchPhase.TRICK_END;
  const isRoundEnd = state?.phase === PitchPhase.ROUND_END;
  const isGameEnd = state?.phase === PitchPhase.GAME_END || state?.gameEndFlag === true;
  const humanIdx = state?.players.findIndex((p) => p.isHuman) ?? -1;
  const isHumanBidTurn = isBidPhase && state?.bidPlayerIdx === humanIdx;
  const isHumanPlayTurn = isPlayPhase && state?.currentPlayerIdx === humanIdx;
  const human = state?.players.find((p) => p.isHuman);

  const cpuDifficulty = state?.config.cpuDifficulty ?? 1;
  const pointLimit = state?.config.pointLimit ?? 7;

  const handleBid = useCallback((bid: number) => execApi('bid', bid), [execApi]);
  const handlePlay = useCallback(() => {
    if (selectedCardIdx === null) return;
    void execApi('play', undefined, selectedCardIdx);
    setSelectedCardIdx(null);
  }, [execApi, selectedCardIdx]);
  const handleNextTrick = useCallback(() => execApi('next'), [execApi]);
  const handleNextRound = useCallback(() => execApi('nextround'), [execApi]);
  const handleManualReset = useCallback(() => {
    hideActionLog();
    void execApi('reset', undefined, undefined, { cpuDifficulty, pointLimit });
  }, [execApi, hideActionLog, cpuDifficulty, pointLimit]);

  if (!state) {
    return (
      <div className={`flex-1 flex items-center justify-center ${gameTheme.pitch.bg}`}>
        <div className="text-ds-text-primary">Loading...</div>
      </div>
    );
  }

  const phaseName = t(`phase.${PHASE_KEYS[state.phase] ?? 'bid'}`);

  return (
    <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.pitch.bg}`} aria-busy={loading}>
      <GamePageHeading title={tc('nav.pitch')} />
      <PhaseIndicator phaseName={phaseName}>
        <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        <TutorialButton />
        <ManualButton gamePath="/pitch" />
      </PhaseIndicator>

      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <div className="overflow-y-auto pt-3 px-4 lg:px-8 flex-1">
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <div className="flex flex-wrap items-center gap-3 text-sm text-ds-text-primary mb-3">
              <span>{t('round', { n: state.roundNumber })}</span>
              <span>{t('trick', { n: state.trickNumber })}</span>
              <span>{t('dealer', { name: findPlayerName(state.players, state.dealerIdx) })}</span>
              <span>
                {t('trumpSuit', {
                  suit: state.trumpSuit === 0 ? t('trumpUnset') : (SUIT_LABELS[state.trumpSuit] ?? '?'),
                })}
              </span>
              <span>{t('currentBid', { n: state.currentBid })}</span>
              {state.bidWinnerIdx >= 0 && (
                <span>{t('bidWinner', { name: findPlayerName(state.players, state.bidWinnerIdx) })}</span>
              )}
            </div>

            <label className="flex items-center gap-1 text-ds-text-primary text-xs justify-center mb-2 cursor-pointer">
              <input type="checkbox" checked={hintEnabled} onChange={(e) => setHintEnabled(e.target.checked)} />
              {tc('hint.toggle', { ns: 'tutorial' })}
            </label>
            {hintEnabled && hint && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}

            {/* Score table */}
            <div data-tutorial="pt-score-table" className="overflow-x-auto mb-3">
              <table className="text-sm w-full border-collapse text-ds-text-primary">
                <thead>
                  <tr className="border-b border-white/20">
                    <th className="text-left p-1">{t('scoresPlayer')}</th>
                    <th className="text-right p-1">{t('scoresBid')}</th>
                    <th className="text-right p-1">{t('scoresTricks')}</th>
                    <th className="text-right p-1">{t('scoresRound')}</th>
                    <th className="text-right p-1">{t('scoresTotal')}</th>
                  </tr>
                </thead>
                <tbody>
                  {state.players.map((p) => {
                    const bidLabel = p.bid === -1 ? t('bidNone') : p.bid === 0 ? t('bidPass') : t('bid', { n: p.bid });
                    return (
                      <tr key={p.id} className={p.isHuman ? 'font-semibold' : ''}>
                        <td className="p-1">{playerName(p.id, p.isHuman)}</td>
                        <td className="text-right p-1">{bidLabel}</td>
                        <td className="text-right p-1">{p.trickCount}</td>
                        <td className="text-right p-1">{p.roundScore}</td>
                        <td className="text-right p-1">{p.cumulativeScore}</td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>

            {/* Current trick */}
            <div
              data-tutorial="pt-trick-display"
              className="border border-white/20 rounded p-2 min-h-[80px] mb-3 text-ds-text-primary"
            >
              <div className="text-xs uppercase opacity-60 mb-1">{t('currentTrick')}</div>
              {state.currentTrick.length === 0 ? (
                <div className="opacity-50">—</div>
              ) : (
                <div className="flex flex-wrap gap-2">
                  {state.currentTrick.map((tc) => (
                    <div
                      key={`${tc.playerIdx}-${tc.card.design}-${tc.card.value}`}
                      className="flex flex-col items-center"
                    >
                      <span className="text-[10px] opacity-60">{findPlayerName(state.players, tc.playerIdx)}</span>
                      <span
                        className={`px-2 py-1 rounded bg-white ${isRedSuit(tc.card.design) ? 'text-red-600' : 'text-gray-900'}`}
                      >
                        {cardLabel(tc.card)}
                      </span>
                    </div>
                  ))}
                </div>
              )}
            </div>

            {/* Player hand (always visible) */}
            {human && human.cards.length > 0 && (
              <div data-tutorial="pt-player-hand" className="flex flex-col gap-2 mb-3 text-ds-text-primary">
                <div className="text-xs uppercase opacity-60">
                  {findPlayerName(state.players, humanIdx)} ({t('cards', { count: human.cards.length })})
                </div>
                <div className="flex flex-wrap gap-2">
                  {human.cards.map((c, idx) => {
                    const isValid = state.validPlayIndices.includes(idx);
                    const isSelected = selectedCardIdx === idx;
                    const canSelect = isHumanPlayTurn && isValid && !loading;
                    return (
                      <button
                        key={`${c.design}-${c.value}-${idx}`}
                        type="button"
                        onClick={() => canSelect && setSelectedCardIdx(idx)}
                        disabled={!canSelect}
                        className={`min-w-[44px] min-h-[44px] px-3 py-2 rounded border-2 transition-all
                          ${isSelected ? 'border-yellow-400 ring-2 ring-yellow-300' : 'border-white/40'}
                          ${canSelect ? 'bg-white opacity-100' : 'bg-white/40 opacity-50 cursor-not-allowed'}
                          ${isRedSuit(c.design) ? 'text-red-600' : 'text-gray-900'}
                        `}
                      >
                        {cardLabel(c)}
                      </button>
                    );
                  })}
                </div>
              </div>
            )}

            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          <GameFooter className={`${gameTheme.pitch.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            <SettingsPanel title={tc('settings.title')} groups={[]} />

            {isBidPhase && (
              <div data-tutorial="pt-bid-controls" className="flex flex-wrap gap-2 items-center justify-center pb-2">
                <span className="text-ds-text-primary text-sm">{t('bidPhase')}</span>
                <button
                  type="button"
                  className={btnSecondary}
                  disabled={!isHumanBidTurn || loading}
                  onClick={() => handleBid(0)}
                >
                  {t('passButton')}
                </button>
                {[2, 3, 4].map((n) => (
                  <button
                    key={n}
                    type="button"
                    className={btnPrimary}
                    disabled={!isHumanBidTurn || loading || n <= state.currentBid}
                    onClick={() => handleBid(n)}
                  >
                    {t('bid', { n })}
                  </button>
                ))}
              </div>
            )}

            {isPlayPhase && isHumanPlayTurn && (
              <div className="flex justify-center pb-2">
                <button
                  type="button"
                  data-tutorial="pt-play-button"
                  className={btnSuccess}
                  onClick={handlePlay}
                  disabled={selectedCardIdx === null || loading}
                >
                  {t('playButton')}
                </button>
              </div>
            )}

            {isTrickEnd && (
              <div className="flex justify-center pb-2">
                <button type="button" className={btnPrimary} onClick={handleNextTrick} disabled={loading}>
                  {t('nextTrick')}
                </button>
              </div>
            )}

            {isRoundEnd && !isGameEnd && (
              <div className="flex justify-center pb-2">
                <button type="button" className={btnPrimary} onClick={handleNextRound} disabled={loading}>
                  {t('nextRound')}
                </button>
              </div>
            )}

            {isGameEnd && (
              <div className="flex flex-col items-center gap-2 pb-2">
                <div className="text-lg font-bold text-ds-text-primary">
                  {state.winnerIdx >= 0 && state.players[state.winnerIdx]?.isHuman
                    ? t('result.humanWin')
                    : t('result.cpuWin', { cpuId: state.winnerIdx })}
                </div>
                <div className="flex justify-center gap-2">
                  <GameResetButton
                    isGameEnd={isGameEnd}
                    onReset={handleManualReset}
                    requestConfirm={requestConfirm}
                    loading={loading}
                  />
                </div>
              </div>
            )}

            <div className="flex justify-center gap-2 pb-2">
              <button
                type="button"
                data-tutorial="pt-reset-button"
                className={btnDanger}
                onClick={() => requestConfirm(handleManualReset)}
                disabled={loading}
              >
                {tc('common.reset')}
              </button>
              <button type="button" className={btnSecondary} onClick={showActionLog} disabled={loading}>
                {tc('actionLog.view')}
              </button>
            </div>
          </GameFooter>
        </>
      )}
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
    </div>
  );
}
