import { useState } from 'react';
import type { spadesApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { RoundScoreAnnouncement } from '../components/RoundScoreAnnouncement';
import { ScrollFadeHint } from '../components/ScrollFadeHint';
import { TrickDisplay } from '../components/TrickDisplay';
import { useBreakpoint, useCardDimensions } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import {
  BAG_PENALTY_THRESHOLD_OPTIONS,
  CPU_DIFFICULTY_OPTIONS,
  NIL_BONUS_OPTIONS,
  POINT_LIMIT_OPTIONS,
  useSpadesGame,
} from '../hooks/useSpadesGame';
import { badgeError, badgeInfo, badgeSuccess, badgeWarning } from '../styles/badgeStyles';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { scoreTableClass } from '../styles/scoreTableStyles';
import type { SpadesResponse } from '../types/card';
import { SpadesPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseSpadesCommand, SPADES_HELP } from '../utils/cli/commands/spadesCommands';
import { formatSpadesState } from '../utils/cli/formatters/spadesFormatter';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';
import { spadesBagWarning, spadesBidProgress } from '../utils/spadesBid';

/** Spades tutorial step definitions (exported so tests can assert the rules are covered). */
export const SP_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="sp-bid-controls"]',
    messageKey: 'tutorial.bidControls',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sp-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sp-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sp-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sp-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  // **標準スペードと違う点はスコア表の直後に言う。** この実装は4人が個別に
  // 得点を競うカットスロート方式で、ScoreRound にチームの概念が無い。パートナーが
  // 表示されないことに戸惑うのは、標準ルールを知っているプレイヤーほど強い (#5498)。
  {
    target: '[data-tutorial="sp-score-table"]',
    messageKey: 'tutorial.cutthroat',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sp-bags-info"]',
    messageKey: 'tutorial.bagsInfo',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sp-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const SPADES_PHASE_KEYS: Readonly<Record<number, string>> = {
  [SpadesPhase.BID]: 'bid',
  [SpadesPhase.PLAY]: 'play',
  [SpadesPhase.TRICK_END]: 'trickEnd',
  [SpadesPhase.ROUND_END]: 'roundEnd',
  [SpadesPhase.GAME_END]: 'gameEnd',
};

type SpadesGame = ReturnType<typeof useSpadesGame>;
type SpadesContext = GamePageContext<SpadesResponse, SpadesGame, SpadesView>;

function useSpadesView() {
  const isLargeDesktop = useBreakpoint() === 'largeDesktop';
  const { cardWidth, isMobile } = useCardDimensions();
  const [bidValue, setBidValue] = useState(1);
  const phaseNames = usePhaseNames('spades', SPADES_PHASE_KEYS);
  return { isLargeDesktop, cardWidth, isMobile, bidValue, setBidValue, phaseNames };
}

type SpadesView = ReturnType<typeof useSpadesView>;

function SpadesMessageBox({ state }: Pick<SpadesContext, 'state'>) {
  return (
    <div data-tutorial="sp-bags-info">
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </div>
  );
}

function Board({ state, t, tc, busy, view }: SpadesContext) {
  const { isLargeDesktop, cardWidth, isMobile, bidValue, setBidValue } = view;
  const loading = busy;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isBidPhase = state.phase === SpadesPhase.BID;
  const isPlayPhase = state.phase === SpadesPhase.PLAY;
  const isTrickEnd = state.phase === SpadesPhase.TRICK_END;
  const isRoundEnd = state.phase === SpadesPhase.ROUND_END;
  const isGameEnd = state.phase === SpadesPhase.GAME_END || state.gameEndFlag;
  const isHumanBidTurn = isBidPhase && state.players[state.bidPlayerIdx]?.isHuman === true;
  // Bid-contract progress for the human player, shown during the play phase.
  const bidProgress =
    (isPlayPhase || isTrickEnd) && humanPlayer && humanPlayer.bid >= 0
      ? spadesBidProgress(humanPlayer.bid, humanPlayer.trickCount)
      : null;
  // Bag-penalty proximity warning for the human player.
  const bagThreshold = state.config.bagPenaltyThreshold;
  const humanBagWarning = humanPlayer ? spadesBagWarning(humanPlayer.bags, bagThreshold) : null;
  // Warning color for a score-table bags cell based on that player's own bags.
  const bagCellClass = (bags: number): string => {
    const w = spadesBagWarning(bags, bagThreshold);
    if (!w) return '';
    return w.level === 'danger' ? 'text-ds-warning font-bold' : 'text-ds-warning';
  };

  return (
    <>
      {isHumanBidTurn && (
        <fieldset className="m-0 border-0 p-0" aria-label={t('bidPhase')}>
          <div className="grid grid-cols-7 gap-1 justify-center">
            {[0, ...Array.from({ length: 13 }, (_, i) => i + 1)].map((v) => (
              <button
                key={v}
                type="button"
                className={
                  v === bidValue
                    ? `${btnSecondary} min-w-[44px] ring-2 ring-ds-warning`
                    : `${btnSecondary} min-w-[44px]`
                }
                aria-pressed={bidValue === v}
                onClick={() => setBidValue(v)}
                disabled={loading}
              >
                {v === 0 ? t('nilButton') : v}
              </button>
            ))}
          </div>
        </fieldset>
      )}

      {/* Round/Trick info */}
      <div className="text-ds-text-primary flex flex-wrap justify-center gap-x-4 gap-y-1 text-center mb-2">
        <span className="whitespace-nowrap">{t('round', { n: state.roundNumber })}</span>
        <span className="whitespace-nowrap">{t('trick', { n: state.trickNumber })}</span>
        <span className="whitespace-nowrap">{state.spadesBroken ? t('spadesBroken') : t('spadesNotBroken')}</span>
      </div>

      {(bidProgress || humanBagWarning) && (
        <div className="text-center mb-2 flex flex-wrap gap-2 justify-center items-center">
          {bidProgress && (
            <span
              data-testid="sp-bid-progress"
              className={
                bidProgress.kind === 'nilFail' ? badgeWarning : bidProgress.kind === 'made' ? badgeSuccess : badgeInfo
              }
            >
              {bidProgress.kind === 'remaining'
                ? t('bidProgress.remaining', { n: bidProgress.remaining })
                : bidProgress.kind === 'made'
                  ? t('bidProgress.made', { bags: bidProgress.bags })
                  : bidProgress.kind === 'nilFail'
                    ? t('bidProgress.nilFail')
                    : t('bidProgress.nilOk')}
            </span>
          )}
          {humanBagWarning && (
            <span
              data-testid="sp-bag-warning"
              className={humanBagWarning.level === 'danger' ? `${badgeError} motion-safe:animate-pulse` : badgeWarning}
            >
              {t('bagWarning', { bags: humanBagWarning.bags, threshold: humanBagWarning.threshold })}
            </span>
          )}
        </div>
      )}

      <div className={lgTwoColGrid}>
        {/* Left: game play area */}
        <div>
          {/* Bid phase instruction */}
          {isHumanBidTurn && (
            <div className="text-ds-warning text-center mb-2" data-tutorial="sp-bid-controls">
              {t('bidPhase')}
            </div>
          )}

          {/* Current trick */}
          <TrickDisplay
            currentTrick={state.currentTrick}
            players={state.players}
            cardWidth={cardWidth}
            label={t('currentTrick')}
            winnerIdx={
              state.phase === SpadesPhase.TRICK_END || state.phase === SpadesPhase.ROUND_END
                ? state.leadPlayerIdx
                : undefined
            }
            winnerLabel={t('trickWinnerBadge')}
            dataTutorial="sp-trick-display"
          />
        </div>

        {/* Right: info sidebar */}
        <div>
          {/* CPU players */}
          {isMobile ? (
            <details className="mb-2 p-2 rounded bg-black/30">
              <summary className="cursor-pointer select-none text-ds-text-muted text-sm">
                {tc('label.cpuOpponents', { count: state.players.filter((p) => !p.isHuman).length })}
              </summary>
              <div className="mt-1">
                {state.players
                  .filter((p) => !p.isHuman)
                  .map((p) => (
                    <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                      {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                      {t('cumulativeScore', { score: p.cumulativeScore })} | {t('roundScore', { score: p.roundScore })}{' '}
                      | {p.bid >= 0 ? t('bid', { n: p.bid }) : t('bidNone')} | {t('bags', { count: p.bags })}
                    </div>
                  ))}
              </div>
            </details>
          ) : (
            state.players
              .filter((p) => !p.isHuman)
              .map((p) => (
                <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
                  <div className="flex flex-wrap gap-x-1 text-ds-text-muted text-sm">
                    <span className="whitespace-nowrap">
                      {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })}
                    </span>
                    <span className="whitespace-nowrap">{t('cumulativeScore', { score: p.cumulativeScore })}</span>
                    <span className="whitespace-nowrap">{t('roundScore', { score: p.roundScore })}</span>
                    <span className="whitespace-nowrap">{p.bid >= 0 ? t('bid', { n: p.bid }) : t('bidNone')}</span>
                    <span className="whitespace-nowrap">{t('bags', { count: p.bags })}</span>
                  </div>
                </div>
              ))
          )}

          {/* Score table */}
          {isMobile ? (
            <details
              className="my-3 p-2 rounded bg-black/30 relative"
              data-tutorial="sp-score-table"
              open={isRoundEnd || isGameEnd || undefined}
            >
              <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('scores')}</summary>
              <div className="overflow-x-auto -mx-2 px-2">
                <table
                  className={scoreTableClass(isLargeDesktop, 'w-full text-sm text-ds-text-muted min-w-[360px]', 'mt-1')}
                >
                  <thead>
                    <tr>
                      <th scope="col" className="text-left whitespace-nowrap text-[10px]">
                        {t('scoresPlayer')}
                      </th>
                      <th scope="col" className="whitespace-nowrap text-[10px]">
                        {t('scoresBid')}
                      </th>
                      <th scope="col" className="whitespace-nowrap text-[10px]">
                        {t('scoresTricks')}
                      </th>
                      <th scope="col" className="whitespace-nowrap text-[10px]">
                        {t('scoresBags')}
                      </th>
                      <th scope="col" className="whitespace-nowrap text-[10px]">
                        {t('scoresRound')}
                      </th>
                      <th scope="col" className="whitespace-nowrap text-[10px]">
                        {t('scoresTotal')}
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {state.players.map((p) => (
                      <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                        <th scope="row" className="text-left">
                          {playerName(p.id, p.isHuman)}
                        </th>
                        <td className="text-center">{p.bid >= 0 ? p.bid : '-'}</td>
                        <td className="text-center">{p.trickCount}</td>
                        <td className={`text-center ${bagCellClass(p.bags)}`} data-testid={`sp-bags-cell-${p.id}`}>
                          {p.bags}
                        </td>
                        <td className="text-center">
                          {p.roundScore}
                          {(isRoundEnd || isGameEnd) && (
                            <div className="text-xs" data-testid={`sp-score-breakdown-${p.id}`}>
                              {t('scoreBreakdown', {
                                bid: p.scoreBreakdown.bidScore,
                                overtricks: p.scoreBreakdown.overtrickScore,
                                nil: p.scoreBreakdown.nilScore,
                                penalty: p.scoreBreakdown.bagPenalty,
                              })}
                            </div>
                          )}
                        </td>
                        <td className="text-center">{p.cumulativeScore}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              <ScrollFadeHint />
            </details>
          ) : (
            <div className="my-3 p-2 rounded bg-black/30 relative" data-tutorial="sp-score-table">
              <div className="text-ds-text-muted text-sm mb-1">{t('scores')}</div>
              <div className="overflow-x-auto -mx-2 px-2">
                <table className={scoreTableClass(isLargeDesktop, 'w-full text-sm text-ds-text-muted min-w-[360px]')}>
                  <thead>
                    <tr>
                      <th scope="col" className="text-left whitespace-nowrap text-[10px]">
                        {t('scoresPlayer')}
                      </th>
                      <th scope="col" className="whitespace-nowrap text-[10px]">
                        {t('scoresBid')}
                      </th>
                      <th scope="col" className="whitespace-nowrap text-[10px]">
                        {t('scoresTricks')}
                      </th>
                      <th scope="col" className="whitespace-nowrap text-[10px]">
                        {t('scoresBags')}
                      </th>
                      <th scope="col" className="whitespace-nowrap text-[10px]">
                        {t('scoresRound')}
                      </th>
                      <th scope="col" className="whitespace-nowrap text-[10px]">
                        {t('scoresTotal')}
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {state.players.map((p) => (
                      <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                        <th scope="row" className="text-left">
                          {playerName(p.id, p.isHuman)}
                        </th>
                        <td className="text-center">{p.bid >= 0 ? p.bid : '-'}</td>
                        <td className="text-center">{p.trickCount}</td>
                        <td className={`text-center ${bagCellClass(p.bags)}`} data-testid={`sp-bags-cell-${p.id}`}>
                          {p.bags}
                        </td>
                        <td className="text-center">
                          {p.roundScore}
                          {(isRoundEnd || isGameEnd) && (
                            <div className="text-xs" data-testid={`sp-score-breakdown-${p.id}`}>
                              {t('scoreBreakdown', {
                                bid: p.scoreBreakdown.bidScore,
                                overtricks: p.scoreBreakdown.overtrickScore,
                                nil: p.scoreBreakdown.nilScore,
                                penalty: p.scoreBreakdown.bagPenalty,
                              })}
                            </div>
                          )}
                        </td>
                        <td className="text-center">{p.cumulativeScore}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}
          <RoundScoreAnnouncement
            active={isRoundEnd || isGameEnd}
            entries={state.players.map((p) => ({
              name: playerName(p.id, p.isHuman),
              roundScore: p.roundScore,
              cumulativeScore: p.cumulativeScore,
            }))}
          />
        </div>
      </div>

      {isMobile && <SpadesMessageBox state={state} />}
    </>
  );
}

function FooterHand({ game, state, t, view }: SpadesContext) {
  const humanPlayer = state.players.find((player) => player.isHuman);
  const isHumanTurn = state.phase === SpadesPhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman === true;
  return humanPlayer ? (
    <PlayerHandSection
      humanPlayer={humanPlayer}
      selectedCardIndices={game.selectedCardIndices}
      toggleCard={game.toggleCard}
      cardWidth={view.cardWidth}
      isMobile={view.isMobile}
      dataTutorialPrefix="sp"
      validIndices={isHumanTurn ? state.validPlayIndices : undefined}
      restrictedTooltip={t('restrictedCard')}
    />
  ) : null;
}

function FooterStatus({ game, t, frontendHint }: SpadesContext) {
  return (
    <>
      {/* ライブ領域は**常設**。hint がある間だけ現れる内側の要素に role/aria-live を
          付けると、領域と中身が同じコミットで DOM に入るので変化として扱われず、
          読み上げられないことがある (#5955, #6663)。 */}
      <div data-testid="spades-hint-live" role="status" aria-live="polite">
        {game.hint && (
          <div className="text-ds-warning text-sm mb-2">
            {game.hint.bid != null
              ? `${t('hintBid')}: ${game.hint.bid} (${t(`hintReason.${game.hint.reason}`)})`
              : `${t('hintPlay')}: [${game.hint.cardIndex}] (${t(`hintReason.${game.hint.reason}`)})`}
          </div>
        )}
      </div>
      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
    </>
  );
}

function FooterActions({ game, state, tc, busy }: SpadesContext) {
  const isBidTurn = state.phase === SpadesPhase.BID && state.players[state.bidPlayerIdx]?.isHuman === true;
  const isHumanTurn = state.phase === SpadesPhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman === true;
  if (!isBidTurn && !isHumanTurn) return null;
  return (
    <button type="button" className={btnSuccess} onClick={game.handleHint} disabled={busy || game.hintLoading}>
      {tc('button.hint')}
    </button>
  );
}

function primaryActions({ game, state, t, busy, view }: SpadesContext) {
  const isHumanBidTurn = state.phase === SpadesPhase.BID && state.players[state.bidPlayerIdx]?.isHuman === true;
  const isHumanTurn = state.phase === SpadesPhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isTrickEnd = state.phase === SpadesPhase.TRICK_END;
  const isRoundEnd = state.phase === SpadesPhase.ROUND_END;
  if (!isHumanBidTurn && !isHumanTurn && !isTrickEnd && !isRoundEnd) return null;
  return (
    <div className="flex gap-2 items-center">
      {isHumanBidTurn && (
        <button type="button" className={btnPrimary} onClick={() => game.handleBid(view.bidValue)} disabled={busy}>
          {t('bidButton')}
        </button>
      )}
      {isHumanTurn && (
        <button
          type="button"
          className={btnPrimary}
          onClick={game.handlePlay}
          disabled={busy || game.selectedCardIndices.length !== 1}
        >
          {t('playButton')}
        </button>
      )}
      {isTrickEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextTrick} disabled={busy}>
          {t('nextTrick')}
        </button>
      )}
      {isRoundEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextRound} disabled={busy}>
          {t('nextRound')}
        </button>
      )}
    </div>
  );
}

const spadesDefinition: GameDefinition<SpadesResponse, SpadesGame, Parameters<typeof spadesApi.exec>, SpadesView> = {
  key: 'spades',
  useGame: useSpadesGame,
  useView: useSpadesView,
  tutorial: SP_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 5 },
  cli: () => ({
    parseCommand: parseSpadesCommand,
    formatResponse: formatSpadesState,
    helpText: SPADES_HELP,
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase],
    ended: state.phase === SpadesPhase.GAME_END || state.gameEndFlag,
  }),
  shell: ({ state }) => ({
    isHumanTurn:
      (state.phase === SpadesPhase.BID && state.players[state.bidPlayerIdx]?.isHuman === true) ||
      (state.phase === SpadesPhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman === true),
  }),
  reset: ({ game }) => {
    void game.exec('reset', undefined, undefined, {
      cpuDifficulty: game.spadesConfig.cpuDifficulty,
      pointLimit: game.spadesConfig.pointLimit,
      nilBonus: game.spadesConfig.nilBonus,
      bagPenaltyThreshold: game.spadesConfig.bagPenaltyThreshold,
    });
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.spadesConfig.cpuDifficulty,
          options: CPU_DIFFICULTY_OPTIONS.map((option) => ({
            value: option.value,
            label: t(`settings.${option.label.toLowerCase()}`),
          })),
          onSelect: (value) => game.handleConfigChange('cpuDifficulty', value),
        },
        {
          type: 'select',
          id: 'pointLimit',
          label: t('settings.pointLimit'),
          value: game.spadesConfig.pointLimit,
          options: POINT_LIMIT_OPTIONS.map((value) => ({ value, label: String(value) })),
          onSelect: (value) => game.handleConfigChange('pointLimit', value),
        },
        {
          type: 'select',
          id: 'nilBonus',
          label: t('settings.nilBonus'),
          value: game.spadesConfig.nilBonus,
          options: NIL_BONUS_OPTIONS.map((value) => ({ value, label: String(value) })),
          onSelect: (value) => game.handleConfigChange('nilBonus', value),
          tooltip: t('settings.nextGameNote'),
          testId: 'sp-setting-nil-bonus',
        },
        {
          type: 'select',
          id: 'bagPenaltyThreshold',
          label: t('settings.bagPenaltyThreshold'),
          value: game.spadesConfig.bagPenaltyThreshold,
          options: BAG_PENALTY_THRESHOLD_OPTIONS.map((value) => ({ value, label: String(value) })),
          onSelect: (value) => game.handleConfigChange('bagPenaltyThreshold', value),
          tooltip: t('settings.nextGameNote'),
          testId: 'sp-setting-bag-threshold',
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  primaryActions,
  message: (ctx) => !ctx.view.isMobile && <SpadesMessageBox state={ctx.state} />,
  cardNav: ({ game, state }) => ({
    cardCount: state.players.find((player) => player.isHuman)?.cards?.length ?? 0,
    onToggle: game.toggleCard,
    onConfirm: game.handlePlay,
    onClear: game.clearSelection,
    enabled: state.phase === SpadesPhase.PLAY && state.players[state.currentPlayerIdx]?.isHuman === true,
  }),
  shortcutsTestId: 'spades-kbd-shortcuts',
  resetTutorialId: 'sp-reset-button',
  actionsTutorialId: 'sp-play-button',
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};

/** Renders the Spades game page with bidding, trick play, and scoring. */
export const SpadesPage = createGamePage(spadesDefinition);
