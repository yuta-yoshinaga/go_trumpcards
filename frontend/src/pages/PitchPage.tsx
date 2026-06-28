import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { pitchApi } from '../api/gameApi';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { CardImage } from '../components/CardImage';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { HintTooltip } from '../components/hint/HintTooltip';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { useSound } from '../providers/SoundProvider';
import { btnDanger, btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { PitchResponse } from '../types/card';
import { PitchPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { PITCH_HELP, parsePitchCommand } from '../utils/cli/commands/pitchCommands';
import { formatPitchState } from '../utils/cli/formatters/pitchFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { pitchHandPipBreakdown, pitchHandPips } from '../utils/pitchUtils';
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
  CLOVER: '♣',
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

function valueLabel(value: number): string {
  return VALUE_LABELS[value] ?? String(value);
}

/** Pitch (Setback) game page. */
export const PitchPage = withTutorial(PitchPageContent, 'pitch', PT_TUTORIAL_STEPS);

function PitchPageContent() {
  const { t } = useTranslation('pitch');
  const { tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('pitch');
  const { playSound } = useSound();
  const { state, loading, error, exec: execApi, retry } = useGameApi(pitchApi.exec);
  const { hint, hintEnabled, setHintEnabled } = useGameHint('pitch', state);
  const [selectedCardIdx, setSelectedCardIdx] = useState<number | null>(null);
  // Game-pips breakdown popover: title-tooltips are unreachable on touch (#2612).
  const [pipsOpen, setPipsOpen] = useState(false);
  const pipsRef = useRef<HTMLSpanElement>(null);
  const { cardWidth } = useCardDimensions();

  // Dismiss the pips popover on Escape or a click/tap outside its wrapper.
  useEffect(() => {
    if (!pipsOpen) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setPipsOpen(false);
    };
    const onPointer = (e: MouseEvent) => {
      if (pipsRef.current && !pipsRef.current.contains(e.target as Node)) setPipsOpen(false);
    };
    document.addEventListener('keydown', onKey);
    document.addEventListener('mousedown', onPointer);
    return () => {
      document.removeEventListener('keydown', onKey);
      document.removeEventListener('mousedown', onPointer);
    };
  }, [pipsOpen]);

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
  const handleConfigChange = useCallback(
    (key: 'cpuDifficulty' | 'pointLimit', value: number) => {
      void execApi('reset', undefined, undefined, {
        cpuDifficulty: key === 'cpuDifficulty' ? value : cpuDifficulty,
        pointLimit: key === 'pointLimit' ? value : pointLimit,
      });
    },
    [execApi, cpuDifficulty, pointLimit],
  );

  if (!state) {
    return (
      <div className={`flex-1 flex items-center justify-center ${gameTheme.pitch.bg}`}>
        <div className="text-ds-text-primary">Loading...</div>
      </div>
    );
  }

  const phaseName = t(`phase.${PHASE_KEYS[state.phase] ?? 'bid'}`);

  return (
    <GamePageShell
      title={tc('nav.pitch')}
      gameThemeBg={gameTheme.pitch.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanBidTurn || isHumanPlayTurn}
      gamePath="/pitch"
      gameEndFlag={isGameEnd}
      winShow={isGameEnd && state.winnerIdx === humanIdx}
      onCelebrate={() => playSound('winFanfare')}
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
              className="border border-ds-border-subtle rounded p-2 min-h-[80px] mb-3 text-ds-text-primary"
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
                      <CardImage card={tc.card} width={cardWidth} />
                    </div>
                  ))}
                </div>
              )}
            </div>

            {/* Player hand (always visible) */}
            {human && human.cards.length > 0 && (
              <div data-tutorial="pt-player-hand" className="flex flex-col gap-2 mb-3 text-ds-text-primary">
                <div className="text-xs uppercase opacity-60 flex items-center gap-2 flex-wrap">
                  <span>
                    {findPlayerName(state.players, humanIdx)} ({t('cards', { count: human.cards.length })})
                  </span>
                  <span className="relative inline-flex" ref={pipsRef}>
                    <button
                      type="button"
                      data-testid="pitch-game-pips-badge"
                      className="normal-case inline-flex items-center justify-center min-h-[44px] min-w-[44px] rounded-full bg-ds-accent/20 text-ds-accent px-2 py-0.5 text-[11px] font-bold"
                      aria-expanded={pipsOpen}
                      aria-haspopup="true"
                      onClick={() => setPipsOpen((o) => !o)}
                      title={t('gamePipsTooltip')}
                    >
                      {t('gamePips', { pips: pitchHandPips(human.cards) })}
                    </button>
                    {pipsOpen && (
                      <section
                        data-testid="pitch-game-pips-popover"
                        aria-label={t('gamePipsTooltip')}
                        className="absolute top-full left-0 z-10 mt-1 w-max max-w-[16rem] rounded-lg bg-ds-surface border border-ds-accent/40 p-2 text-[11px] text-ds-text-primary shadow-lg"
                      >
                        <div className="font-bold mb-1">{t('gamePipsTooltip')}</div>
                        <ul className="space-y-0.5">
                          {pitchHandPipBreakdown(human.cards)
                            .filter((b) => b.pips > 0)
                            .map((b) => (
                              <li key={`pip-${b.value}`} className="flex justify-between gap-3">
                                <span>{valueLabel(b.value)}</span>
                                <span className="font-mono">+{b.pips}</span>
                              </li>
                            ))}
                        </ul>
                        <div className="mt-1 border-t border-ds-accent/30 pt-1 flex justify-between gap-3 font-bold">
                          <span>{t('gamePipsTotal')}</span>
                          <span className="font-mono">{pitchHandPips(human.cards)}</span>
                        </div>
                      </section>
                    )}
                  </span>
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
                        aria-label={cardLabel(c)}
                        className={`min-w-[44px] min-h-[44px] rounded transition-all
                          ${isSelected ? 'ring-2 ring-ds-accent' : ''}
                          ${canSelect ? 'opacity-100' : 'opacity-50 cursor-not-allowed'}
                        `}
                      >
                        <CardImage card={c} width={cardWidth} />
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
            <SettingsPanel
              title={tc('settings.title')}
              groups={[
                {
                  items: [
                    {
                      type: 'select',
                      id: 'cpuDifficulty',
                      label: t('settings.cpuDifficulty'),
                      value: cpuDifficulty,
                      options: [
                        { value: 0, label: t('settings.easy') },
                        { value: 1, label: t('settings.normal') },
                        { value: 2, label: t('settings.hard') },
                      ],
                      onSelect: (v) => handleConfigChange('cpuDifficulty', Number(v)),
                    },
                    {
                      type: 'select',
                      id: 'pointLimit',
                      label: t('settings.pointLimit'),
                      value: pointLimit,
                      options: [5, 7, 9, 11, 15, 21].map((v) => ({ value: v, label: String(v) })),
                      onSelect: (v) => handleConfigChange('pointLimit', Number(v)),
                    },
                  ],
                },
              ]}
            />

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
                {tc('button.reset')}
              </button>
              <button type="button" className={btnSecondary} onClick={showActionLog} disabled={loading}>
                {tc('actionLog.view')}
              </button>
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
