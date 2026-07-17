import { useCallback, useMemo, useState } from 'react';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { HintTooltip } from '../components/hint/HintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useBarbuGame } from '../hooks/useBarbuGame';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { gameTheme } from '../styles/gameTheme';
import type { BarbuResponse } from '../types/card';
import type { TutorialStep } from '../types/tutorial';
import {
  BARBU_HELP,
  type BarbuCliArgs,
  formatBarbuState,
  parseBarbuCommand,
} from '../utils/cli/commands/barbuCommands';
import type { CliGameConfig } from '../utils/cli/types';

const DIFFICULTY_OPTIONS = [
  { value: '0', label: 'Easy' },
  { value: '1', label: 'Normal' },
  { value: '2', label: 'Hard' },
];

const CONTRACT_COUNT = 7;
const CONTRACT_TRUMPS = 5;
const CONTRACT_DOMINOES = 6;
const SUIT_SYMBOLS = ['', '♠', '♣', '♥', '♦'];

/** Tutorial steps for Barbu. */
const BB_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="bb-deal-info"]', messageKey: 'tutorial.dealInfo', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="bb-contracts"]', messageKey: 'tutorial.contracts', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="bb-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="bb-actions"]', messageKey: 'tutorial.actions', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="bb-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Expands a per-suit bitmask into the list of placed card values (low→high). */
function expandPlaced(mask: number): number[] {
  const values: number[] = [];
  for (let v = 1; v <= 13; v++) {
    if ((mask & (1 << v)) !== 0) values.push(v);
  }
  return values;
}

/** Renders the Barbu (バルブ) game page. */
export const BarbuPage = withTutorial(BarbuPageContent, 'barbu', BB_TUTORIAL_STEPS);
function BarbuPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('barbu');
  const {
    state,
    loading,
    error,
    callApi,
    handIndex,
    setHandIndex,
    configInput,
    handleConfigChange,
    selectContract,
    play,
    pass,
    handleNextDeal,
    handleResetWithConfig,
    retry,
  } = useBarbuGame();
  const { cardWidth } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('barbu', state);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('barbu');
  const cliConfig: CliGameConfig<BarbuResponse, BarbuCliArgs> = useMemo(
    () => ({
      gameName: 'barbu',
      parseCommand: parseBarbuCommand,
      formatResponse: formatBarbuState,
      helpText: [...BARBU_HELP],
    }),
    [],
  );
  const { handleCommand } = useCliGame(callApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const [trumpPickerOpen, setTrumpPickerOpen] = useState(false);
  const [hoveredContract, setHoveredContract] = useState<number | null>(null);
  const onReset = useCallback(() => handleResetWithConfig(), [handleResetWithConfig]);

  const onContractClick = useCallback(
    (contract: number) => {
      if (contract === CONTRACT_TRUMPS) {
        setTrumpPickerOpen(true);
        return;
      }
      selectContract(contract, -1);
    },
    [selectContract],
  );
  const onTrumpClick = useCallback(
    (suit: number) => {
      setTrumpPickerOpen(false);
      selectContract(CONTRACT_TRUMPS, suit);
    },
    [selectContract],
  );

  const human = state && state.players.length >= 1 ? state.players[0] : null;
  const isSelectPhase = !!state && state.phase === 'selectContract';
  // Default the description panel to the first contract that is still selectable.
  const firstAvailableContract = state
    ? Math.max(
        0,
        state.usedContracts.findIndex((used) => !used),
      )
    : 0;
  const isPlayPhase = !!state && state.phase === 'play';
  const isHumanSelect = isSelectPhase && state.dealerIdx === 0 && !state.gameEndFlag;
  const isHumanPlay = isPlayPhase && state.currentTurn === 0 && !state.gameEndFlag;
  const isHumanTurn = isHumanSelect || isHumanPlay;

  if (!state || !human) {
    return (
      <div className={`flex-1 flex items-center justify-center ${gameTheme.barbu.bg} text-ds-text-muted`} aria-busy>
        {tc('skeleton.loading')}
      </div>
    );
  }

  const isGameEnd = state.gameEndFlag;
  const isDealEnd = state.phase === 'dealEnd';
  const humanWon = isGameEnd && state.roundWinners.includes(0);
  const isDominoes = state.currentContract === CONTRACT_DOMINOES;
  const canPlay = isHumanPlay && handIndex !== null;
  const canPass = isHumanPlay && isDominoes && state.dominoPlayable.length === 0;
  const phaseName = isGameEnd ? t('phase.end') : t(`phase.${state.phase}`, t('phase.play'));

  return (
    <GamePageShell
      title={tc('nav.barbu')}
      gameThemeBg={gameTheme.barbu.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/barbu"
      gameEndFlag={isGameEnd}
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
              <button type="button" onClick={retry} className="text-ds-error underline">
                {error}
              </button>
            )}

            <div className="text-center text-xs text-ds-text-muted" data-tutorial="bb-deal-info">
              {t('label.deal', { deal: state.dealNumber + 1, total: state.totalDeals })} —{' '}
              {t('label.dealer', {
                name: state.dealerIdx === 0 ? tc('player.you') : tc('player.cpu', { id: state.dealerIdx }),
              })}
              {state.currentContract >= 0 && (
                <span className="ml-2 text-ds-text">
                  · {t('label.contract')}: {t(`contract.${state.currentContract}`)}
                  {state.trumpSuit >= 1 && ` (${SUIT_SYMBOLS[state.trumpSuit]})`}
                </span>
              )}
            </div>

            {/* CPU players */}
            <div className="flex justify-center gap-6 flex-wrap" data-tutorial="bb-cpu-area">
              {state.players
                .filter((p) => !p.isHuman)
                .map((p) => (
                  <div key={p.id} className="text-center">
                    <div className="text-xs text-ds-text-muted mb-1">
                      {tc('player.cpu', { id: p.id })} — {p.cardCount} / {p.trickCount}t / {p.totalScore}pt
                    </div>
                    <div className="flex gap-0.5 justify-center">
                      {Array.from({ length: Math.min(p.cardCount, 8) }, (_, i) => (
                        <AnimatedCardBack key={i} width={cardWidth * 0.4} />
                      ))}
                    </div>
                  </div>
                ))}
            </div>

            {/* Contract selection */}
            {isHumanSelect && (
              <div className="py-2 bg-black/20 rounded-lg" data-tutorial="bb-contracts">
                {!trumpPickerOpen ? (
                  <>
                    <div className="flex flex-wrap justify-center gap-2">
                      {Array.from({ length: CONTRACT_COUNT }, (_, c) => (
                        <button
                          key={c}
                          type="button"
                          onClick={() => onContractClick(c)}
                          onMouseEnter={() => setHoveredContract(c)}
                          onMouseLeave={() => setHoveredContract(null)}
                          onFocus={() => setHoveredContract(c)}
                          onBlur={() => setHoveredContract(null)}
                          disabled={loading || state.usedContracts[c]}
                          className="px-3 py-2 rounded-lg bg-ds-info text-white text-xs font-medium disabled:opacity-30 disabled:line-through"
                          data-testid={`contract-${c}`}
                          title={t(`contractDesc.${c}`)}
                        >
                          {t(`contract.${c}`)}
                        </button>
                      ))}
                    </div>
                    <p
                      role="status"
                      data-testid="contract-desc"
                      className="mt-2 min-h-[2rem] px-3 text-center text-ds-text-muted text-xs"
                    >
                      {t(`contractDesc.${hoveredContract ?? firstAvailableContract}`)}
                    </p>
                  </>
                ) : (
                  <div className="flex flex-col items-center gap-2">
                    <span className="text-xs text-ds-text-muted">{t('label.selectTrump')}</span>
                    <div className="flex gap-2">
                      {[1, 2, 3, 4].map((suit) => (
                        <button
                          key={suit}
                          type="button"
                          onClick={() => onTrumpClick(suit)}
                          disabled={loading}
                          className="px-3 py-2 rounded-lg bg-ds-warning text-white text-sm font-medium"
                          data-testid={`trump-${suit}`}
                        >
                          {t(`trump.${suit}`)}
                        </button>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            )}

            {/* Board: dominoes layout or current trick */}
            {isPlayPhase && (
              <div className="py-3 bg-black/20 rounded-lg" data-tutorial="bb-board">
                {isDominoes ? (
                  <div className="space-y-1 px-3">
                    {[1, 2, 3, 4].map((suit) => {
                      const vals = expandPlaced(state.tablePlaced[suit] ?? 0);
                      return (
                        <div key={suit} className="text-sm text-ds-text text-center">
                          <span className="mr-2">{SUIT_SYMBOLS[suit]}</span>
                          {vals.length === 0 ? <span className="text-ds-text-muted">—</span> : vals.join(' · ')}
                        </div>
                      );
                    })}
                  </div>
                ) : (
                  <>
                    <div className="text-center text-xs text-ds-text-muted mb-2">{t('label.trick')}</div>
                    <div className="flex justify-center gap-2 min-h-[60px] flex-wrap">
                      {state.currentTrick.length === 0 ? (
                        <span className="text-ds-text-muted text-sm self-center">—</span>
                      ) : (
                        state.currentTrick.map((tcard, i) => {
                          const tp = state.players[tcard.playerIdx];
                          const tIsHuman = tp?.isHuman === true;
                          const isLead = i === 0;
                          return (
                            <div key={i} className="text-center" data-testid="bb-trick-card">
                              <div className={`inline-block rounded ${isLead ? 'ring-2 ring-ds-info' : ''}`}>
                                <AnimatedCard card={tcard.card} width={cardWidth * 0.9} />
                              </div>
                              <div
                                className={`text-xs mt-1 ${tIsHuman ? 'text-ds-accent font-semibold' : 'text-ds-text-muted'}`}
                              >
                                {isLead && (
                                  <span aria-hidden="true" className="mr-0.5" title={t('label.lead')}>
                                    ▸
                                  </span>
                                )}
                                {tIsHuman ? tc('player.you') : tc('player.cpu', { id: tp?.id ?? tcard.playerIdx })}
                              </div>
                            </div>
                          );
                        })
                      )}
                    </div>
                  </>
                )}
              </div>
            )}

            {/* Human hand */}
            <div className="text-center" data-tutorial="bb-player-hand">
              <div className="text-xs text-ds-text-muted mb-1">
                {tc('player.you')} — {human.cardCount} / {t('label.yourTricks')} {human.trickCount} / {human.totalScore}
                pt
              </div>
              <div className="flex flex-wrap justify-center gap-2">
                {human.cards.map((c, i) => {
                  const playable = !isDominoes || state.dominoPlayable.includes(i);
                  return (
                    <button
                      key={i}
                      type="button"
                      onClick={() => isHumanPlay && setHandIndex(handIndex === i ? null : i)}
                      disabled={!isHumanPlay || !playable}
                      className={`rounded transition-all ${
                        handIndex === i ? 'ring-2 ring-ds-info -translate-y-2' : ''
                      } ${isHumanPlay && playable ? 'cursor-pointer hover:opacity-90' : 'opacity-50 cursor-default'}`}
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
                    options: DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v: string) => handleConfigChange('cpuDifficulty', Number.parseInt(v, 10)),
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

          <GameFooter className={`${gameTheme.barbu.footer} px-4 py-2.5`}>
            <div className="flex gap-2 justify-center flex-wrap items-center" data-tutorial="bb-actions">
              <button
                type="button"
                onClick={play}
                disabled={loading || !canPlay}
                className="px-4 py-2 rounded-lg bg-ds-info text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed text-sm"
                data-testid="play-button"
                data-hint-action="play"
              >
                {t('button.play')}
              </button>
              {isDominoes && (
                <button
                  type="button"
                  onClick={pass}
                  disabled={loading || !canPass}
                  className="px-4 py-2 rounded-lg bg-ds-warning text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed text-sm"
                  data-testid="pass-button"
                  data-hint-action="pass"
                >
                  {t('button.pass')}
                </button>
              )}
              {isDealEnd && !isGameEnd && (
                <button
                  type="button"
                  onClick={handleNextDeal}
                  disabled={loading}
                  className="px-4 py-2 rounded-lg bg-ds-success text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed text-sm"
                  data-testid="next-deal-button"
                >
                  {t('button.nextDeal')}
                </button>
              )}
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={onReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="bb-reset-button"
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
