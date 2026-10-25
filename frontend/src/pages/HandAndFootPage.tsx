import type { TFunction } from 'i18next';
import { useCallback, useMemo } from 'react';
import type { handandfootApi } from '../api/gameApi';
import { CardFan } from '../components/CardFan';
import { CpuAccordion } from '../components/CpuAccordion';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { useBreakpoint, useCardDimensions, useIsMobile } from '../hooks/useCardDimensions';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useHandAndFootGame } from '../hooks/useHandAndFootGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeInfoColors } from '../styles/badgeStyles';
import { btnOutline, btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { HandAndFootResponse } from '../types/card';
import { HandAndFootPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { canastaFamilySelectionPoints } from '../utils/canastaFamilyScore';
import { HANDANDFOOT_HELP, parseHandAndFootCommand } from '../utils/cli/commands/handandfootCommands';
import { formatHandAndFootState } from '../utils/cli/formatters/handandfootFormatter';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

const HANDANDFOOT_PHASE_KEYS: Readonly<Record<number, string>> = {
  [HandAndFootPhase.DRAW]: 'draw',
  [HandAndFootPhase.MELD]: 'meld',
  [HandAndFootPhase.DISCARD]: 'discard',
  [HandAndFootPhase.ROUND_END]: 'roundEnd',
  [HandAndFootPhase.GAME_END]: 'gameEnd',
};

/** Hand and Foot tutorial step definitions. */
const HF_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="hf-draw-area"]', messageKey: 'tutorial.drawArea', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="hf-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="hf-meld-area"]', messageKey: 'tutorial.meldArea', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="hf-actions"]', messageKey: 'tutorial.actionButtons', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="hf-draw-area"]', messageKey: 'tutorial.frozenRule', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="hf-player-hand"]', messageKey: 'tutorial.footRule', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="hf-meld-area"]',
    messageKey: 'tutorial.initialMeldRule',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="hf-actions"]', messageKey: 'tutorial.goOutRule', placement: 'top', advanceOn: 'next' },
];

function useGame() {
  const game = useHandAndFootGame();
  return {
    ...game,
    exec: game.gameExec,
  };
}

type HandAndFootGame = ReturnType<typeof useGame>;
type HandAndFootView = ReturnType<typeof useHandAndFootView>;
type HandAndFootContext = GamePageContext<HandAndFootResponse, HandAndFootGame, HandAndFootView>;

function useHandAndFootView({ game, t }: { game: HandAndFootGame; t: TFunction }) {
  const isMobile = useIsMobile();
  const isLargeDesktop = useBreakpoint() === 'largeDesktop';
  const { cardWidth } = useCardDimensions();
  const phaseNames = usePhaseNames('handandfoot', HANDANDFOOT_PHASE_KEYS);

  const state = game.state;
  const humanPlayer = state?.players.find((p) => p.isHuman);
  const humanCardCount = humanPlayer?.cards?.length ?? 0;

  const isDrawPhase = state?.phase === HandAndFootPhase.DRAW;
  const isMeldPhase = state?.phase === HandAndFootPhase.MELD;
  const isDiscardPhase = state?.phase === HandAndFootPhase.DISCARD;
  const isRoundEnd = state?.phase === HandAndFootPhase.ROUND_END;
  const isGameEnd = state?.phase === HandAndFootPhase.GAME_END || !!state?.gameEndFlag;
  const isHumanTurn =
    (isDrawPhase || isMeldPhase || isDiscardPhase) && state?.players[state.currentPlayerIdx]?.isHuman === true;

  const drawDiscardReason = useMemo(() => {
    if (!isDrawPhase || !isHumanTurn) return '';
    const n = game.selectedCardIndices.length;
    if (n > 2) return t('drawDiscardReason.tooMany');
    if (n === 2) return t(state?.isFrozen ? 'drawDiscardReason.frozenReady' : 'drawDiscardReason.ready');
    if (state?.isFrozen) return t('drawDiscardReason.frozen');
    if (n === 1) return t('drawDiscardReason.selectOneMore');
    return t('drawDiscardReason.selectTwo');
  }, [isDrawPhase, isHumanTurn, game.selectedCardIndices.length, state?.isFrozen, t]);

  // Whether the human may currently "go out", plus the first unmet requirement.
  // The web build always uses the default go-out rule (>=1 red/natural and >=1
  // black/mixed team canasta, plus the player having entered their foot); see
  // DefaultHandAndFootConfig — the web config never exposes the rc/bc thresholds.
  const goOutGuidance = useMemo(() => {
    if (!isDiscardPhase || !humanPlayer) return null;
    const team = state?.teams.find((tm) => tm.team === humanPlayer.team);
    const canastas = team?.melds.filter((m) => m.isCanasta) ?? [];
    const redCanastas = canastas.filter((m) => m.isNatural).length;
    const blackCanastas = canastas.filter((m) => !m.isNatural).length;
    if (!humanPlayer.inFoot) return { canGoOut: false, reasonKey: 'goOutReason.needFoot' };
    if (redCanastas < 1) return { canGoOut: false, reasonKey: 'goOutReason.needRedCanasta' };
    if (blackCanastas < 1) return { canGoOut: false, reasonKey: 'goOutReason.needBlackCanasta' };
    return { canGoOut: true, reasonKey: 'goOutReason.ready' };
  }, [isDiscardPhase, humanPlayer, state?.teams]);

  // Meld phase: show the selected cards' running point total and, until the
  // team has opened (no melds yet), the initial-meld minimum so the player can
  // tell if the selection qualifies. Point values mirror the shared
  // Canasta-family table (domain.CanastaFamilyCardValue); the minimum is the
  // server's `minMeld`. The
  // readout also gates the meld button now (#5663): below the minimum the
  // button is disabled, so the player is not sent to a server rejection.
  // **メルドフェーズ以外では 0 点・未達なしを返す。** null を返すと、この値を
  // 読むたびに「無いかもしれない」分岐が増え、押せるかどうかの判断が
  // 追いにくくなる (#5663)。
  const meldPointInfo = useMemo(() => {
    if (!isMeldPhase || !humanPlayer) {
      return { selectedPoints: 0, needInitial: false, minMeld: 0, met: true, below: false };
    }
    const selectedCards = game.selectedCardIndices
      .map((i) => humanPlayer.cards[i])
      .filter((c): c is NonNullable<typeof c> => !!c);
    const selectedPoints = canastaFamilySelectionPoints(selectedCards);
    const team = state?.teams.find((tm) => tm.team === humanPlayer.team);
    const needInitial = (team?.melds.length ?? 0) === 0;
    const minMeld = state.minMeld;
    const met = selectedPoints >= minMeld;
    return { selectedPoints, needInitial, minMeld, met, below: needInitial && !met };
  }, [isMeldPhase, humanPlayer, game.selectedCardIndices, state?.teams, state?.minMeld]);

  const kbdConfirmAction = useCallback(() => {
    if (isDiscardPhase) game.handleDiscard();
    else if (isMeldPhase) game.handleMeldSelected();
  }, [isDiscardPhase, isMeldPhase, game.handleDiscard, game.handleMeldSelected]);

  return {
    isMobile,
    isLargeDesktop,
    cardWidth,
    phaseNames,
    humanPlayer,
    humanCardCount,
    isDrawPhase,
    isMeldPhase,
    isDiscardPhase,
    isRoundEnd,
    isGameEnd,
    isHumanTurn,
    drawDiscardReason,
    goOutGuidance,
    meldPointInfo,
    kbdConfirmAction,
  };
}

function LiveRegion({ t, view }: HandAndFootContext) {
  const { isMeldPhase, isHumanTurn, meldPointInfo } = view;
  return (
    <div
      role="status"
      aria-live="polite"
      aria-atomic="true"
      className="sr-only"
      data-testid="hf-meld-points-announcement"
    >
      {isMeldPhase && isHumanTurn
        ? meldPointInfo.needInitial
          ? t('meldPoints.initial', {
              min: meldPointInfo.minMeld,
              points: meldPointInfo.selectedPoints,
            })
          : t('meldPoints.selected', { points: meldPointInfo.selectedPoints })
        : ''}
    </div>
  );
}

function Board({ state, t, tc, frontendHint, view }: HandAndFootContext) {
  const { isMobile, isLargeDesktop, cardWidth, isRoundEnd, isGameEnd } = view;

  return (
    <>
      <div className="flex flex-wrap justify-center text-ds-text-primary text-center mb-2">
        <span className="mr-4 whitespace-nowrap">{t('round', { n: state.roundNumber })}</span>
        <span className="whitespace-nowrap">
          {t('drawPile', { count: state.drawPileCount })} / {t('discardPile', { count: state.discardPileCount })}
        </span>
        {state.isFrozen && <span className="ml-2 whitespace-nowrap text-ds-info font-bold">[{t('frozen')}]</span>}
      </div>

      <div className={lgTwoColGrid}>
        {/* Left: game play area */}
        <div>
          <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />

          <div className={isLargeDesktop ? 'flex flex-wrap items-start gap-2' : undefined}>
            {/* Discard pile top */}
            {state.discardTop && (
              <div
                className={`my-3 p-3 rounded flex items-center gap-3 relative ${
                  state.isFrozen ? `${badgeInfoColors} ring-2 ring-ds-info` : 'bg-black/40'
                }`}
                data-tutorial="hf-draw-area"
                data-testid="hf-discard-pile"
              >
                <AnimatedCard card={state.discardTop} width={cardWidth} />
                <div className="text-ds-text-muted text-sm">{t('discardTop')}</div>
                {state.isFrozen && (
                  <span
                    className="absolute top-1 right-2 text-ds-info text-xs font-bold"
                    data-testid="hf-frozen-badge"
                    role="img"
                    aria-label={t('frozenIndicator')}
                  >
                    {t('frozenIndicator')}
                  </span>
                )}
              </div>
            )}

            {/* Team melds */}
            {state.teams.map((team, ti) => {
              if (team.melds.length === 0 && team.red3s.length === 0) return null;
              return (
                <div
                  key={team.team}
                  className={`my-2 p-2 rounded bg-black/30 ${isLargeDesktop ? 'flex flex-col' : ''}`}
                  data-tutorial={ti === 0 ? 'hf-meld-area' : undefined}
                >
                  <div className="text-ds-text-muted text-sm mb-1">
                    {t('team', { n: team.team })} - {t('melds')}
                  </div>
                  {team.melds.map((m, mi) => (
                    <div key={mi} className={`flex ${isLargeDesktop ? 'flex-nowrap' : 'flex-wrap'} gap-1 mb-1`}>
                      <span className="text-xs text-ds-text-muted self-center mr-1">
                        {m.isCanasta ? (m.isNatural ? t('naturalCanasta') : t('mixedCanasta')) : `(${m.cards.length})`}
                      </span>
                      {isLargeDesktop ? (
                        <CardFan
                          itemWidth={cardWidth * 0.6}
                          className="min-w-0 flex-1"
                          data-testid={`hf-meld-fan-${ti}-${mi}`}
                        >
                          {m.cards.map((card, ci) => (
                            <AnimatedCard key={`meld-${ti}-${mi}-${ci}`} card={card} width={cardWidth * 0.6} />
                          ))}
                        </CardFan>
                      ) : (
                        m.cards.map((card, ci) => (
                          <AnimatedCard key={`meld-${ti}-${mi}-${ci}`} card={card} width={cardWidth * 0.6} />
                        ))
                      )}
                    </div>
                  ))}
                  {team.red3s.length > 0 && (
                    <div className={`flex ${isLargeDesktop ? 'flex-nowrap' : 'flex-wrap'} gap-1 mt-1`}>
                      <span className="text-xs text-ds-error-text self-center mr-1">{t('red3s')}</span>
                      {team.red3s.map((card, ri) => (
                        <AnimatedCard key={`red3-${ti}-${ri}`} card={card} width={cardWidth * 0.6} />
                      ))}
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        </div>

        {/* Right: info sidebar */}
        <div>
          {/* Score table */}
          <details className="my-3 p-2 rounded bg-black/30" open={!isMobile || undefined}>
            <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('score.table')}</summary>
            <table className="w-full text-sm text-ds-text-muted">
              <thead>
                <tr>
                  <th scope="col" className="text-left">
                    {tc('label.player')}
                  </th>
                  <th scope="col">{t('teamLabel')}</th>
                  <th scope="col">{t('foot')}</th>
                  <th scope="col">{t('score.round')}</th>
                  <th scope="col">{t('score.cumulative')}</th>
                </tr>
              </thead>
              <tbody>
                {state.players.map((p) => (
                  <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                    <th scope="row" className="text-left">
                      {playerName(p.id, p.isHuman)}
                    </th>
                    <td className="text-center">{p.team}</td>
                    <td className="text-center">{p.inFoot ? t('inFoot') : p.footCount}</td>
                    <td className="text-center">{p.roundScore}</td>
                    <td className="text-center">{p.cumulativeScore}</td>
                  </tr>
                ))}
              </tbody>
            </table>
            {(isRoundEnd || isGameEnd) && state.scoreBreakdown.length > 0 && (
              <div data-testid="hf-score-breakdown" className="mt-2 border-t border-white/20 pt-2">
                <h3 className="text-sm font-semibold text-ds-text-primary">{t('score.breakdown')}</h3>
                {state.scoreBreakdown.map((breakdown) => (
                  <div key={breakdown.team} className="mt-1 text-xs text-ds-text-muted">
                    <h4 className="font-medium">{t('team', { n: breakdown.team + 1 })}</h4>
                    <dl className="grid grid-cols-2 gap-x-2">
                      {(
                        [
                          ['meldCards', breakdown.meldCards],
                          ['redCanasta', breakdown.redCanasta],
                          ['blackCanasta', breakdown.blackCanasta],
                          ['redThrees', breakdown.redThrees],
                          ['goingOut', breakdown.goingOut],
                          ['handPenalty', -breakdown.handPenalty],
                          ['footPenalty', -breakdown.footPenalty],
                        ] as const
                      ).map(([key, value]) => (
                        <div key={key} className="flex justify-between gap-2">
                          <dt>{t(`score.${key}`)}</dt>
                          <dd>{value < 0 ? `−${Math.abs(value)}` : value}</dd>
                        </div>
                      ))}
                    </dl>
                  </div>
                ))}
              </div>
            )}
          </details>

          {/* CPU hands (shown at round/game end) */}
          {(isRoundEnd || isGameEnd) && (
            <CpuAccordion playerCount={state.players.filter((p) => !p.isHuman).length}>
              {state.players
                .filter((p) => !p.isHuman)
                .map((p) => (
                  <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
                    <div className="text-ds-text-muted text-sm">
                      {playerName(p.id, p.isHuman)}: {t('handCards', { count: p.cardCount })}
                    </div>
                    {p.cards.length > 0 && (
                      <div className="flex flex-wrap gap-1 mt-1">
                        {p.cards.map((card, idx) => (
                          <AnimatedCard
                            key={`cpu-${card.design}-${card.value}-${idx}`}
                            card={card}
                            width={cardWidth * 0.7}
                          />
                        ))}
                      </div>
                    )}
                  </div>
                ))}
            </CpuAccordion>
          )}
        </div>
      </div>

      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterHand({ game, view }: HandAndFootContext) {
  const { humanPlayer, cardWidth, isMobile, drawDiscardReason } = view;
  const { selectedCardIndices, toggleCard } = game;

  return (
    <>
      {humanPlayer && (
        <PlayerHandSection
          humanPlayer={humanPlayer}
          selectedCardIndices={selectedCardIndices}
          toggleCard={toggleCard}
          cardWidth={cardWidth}
          isMobile={isMobile}
          dataTutorialPrefix="hf"
          stackRows
        />
      )}

      <div
        id="hf-draw-discard-reason"
        data-testid="hf-draw-discard-reason"
        role="status"
        aria-live="polite"
        aria-atomic="true"
      >
        {drawDiscardReason && <span className="text-xs text-ds-text-muted">{drawDiscardReason}</span>}
      </div>
    </>
  );
}

function primaryActions({ game, state, t, busy, view }: HandAndFootContext) {
  const { isDrawPhase, isMeldPhase, isDiscardPhase, isRoundEnd, isHumanTurn, meldPointInfo, goOutGuidance } = view;
  const {
    selectedCardIndices,
    handleDrawStock,
    handleDrawDiscard,
    handleMeldSelected,
    handleSkipMeld,
    handleDiscard,
    handleGoOut,
    handleNextRound,
  } = game;

  if (
    !((isDrawPhase && isHumanTurn) || (isMeldPhase && isHumanTurn) || (isDiscardPhase && isHumanTurn) || isRoundEnd)
  ) {
    return null;
  }

  return (
    <div className="flex gap-2 items-center flex-wrap" data-tutorial="hf-actions">
      {isDrawPhase && isHumanTurn && (
        <div className="flex gap-2 flex-col">
          <div className="flex gap-2">
            <button type="button" className={btnPrimary} onClick={handleDrawStock} disabled={busy}>
              {t('drawStockButton')}
            </button>
            <button
              type="button"
              className={`${btnPrimary} ${
                state.isFrozen && selectedCardIndices.length === 2 && !busy
                  ? 'motion-safe:animate-pulse ring-2 ring-ds-info'
                  : ''
              }`}
              data-frozen={state.isFrozen && selectedCardIndices.length === 2 && !busy ? 'true' : undefined}
              onClick={handleDrawDiscard}
              disabled={busy || selectedCardIndices.length !== 2}
              aria-describedby="hf-draw-discard-reason"
            >
              {t('drawDiscardButton')}
            </button>
          </div>
        </div>
      )}
      {isMeldPhase && isHumanTurn && (
        <>
          <div
            id="hf-meld-points"
            data-testid="hf-meld-points"
            className={`w-full text-xs ${
              meldPointInfo.below
                ? 'text-ds-warning'
                : meldPointInfo.needInitial
                  ? 'text-ds-success'
                  : 'text-ds-text-muted'
            }`}
          >
            {meldPointInfo.needInitial
              ? t('meldPoints.initial', {
                  min: meldPointInfo.minMeld,
                  points: meldPointInfo.selectedPoints,
                })
              : t('meldPoints.selected', { points: meldPointInfo.selectedPoints })}
          </div>
          <button
            type="button"
            className={btnPrimary}
            onClick={handleMeldSelected}
            // **最低点未達なら押させない** (#5663)。警告テキストは出ていた
            // のにボタンはそれを見ておらず、サーバーのバリデーションで
            // 弾かれて初めて気づく形だった。
            disabled={busy || selectedCardIndices.length < 3 || meldPointInfo.below}
          >
            {t('meldButton')}
          </button>
          <button type="button" className={btnOutline} onClick={handleSkipMeld} disabled={busy}>
            {t('skipMeldButton')}
          </button>
        </>
      )}
      {isDiscardPhase && isHumanTurn && (
        <div className="flex gap-2 flex-col">
          <div className="flex gap-2">
            <button
              type="button"
              className={btnPrimary}
              onClick={handleDiscard}
              disabled={busy || selectedCardIndices.length !== 1}
            >
              {t('discardButton')}
            </button>
            <button
              type="button"
              className={btnSuccess}
              onClick={handleGoOut}
              disabled={busy || goOutGuidance?.canGoOut !== true}
              title={goOutGuidance && !goOutGuidance.canGoOut ? t(goOutGuidance.reasonKey) : undefined}
              aria-describedby={goOutGuidance ? 'hf-go-out-guidance' : undefined}
            >
              {t('goOutButton')}
            </button>
          </div>
          {goOutGuidance && (
            <div
              id="hf-go-out-guidance"
              data-testid="hf-go-out-guidance"
              className={`text-xs ${goOutGuidance.canGoOut ? 'text-ds-success' : 'text-ds-text-muted'}`}
            >
              {t(goOutGuidance.reasonKey)}
            </div>
          )}
        </div>
      )}
      {isRoundEnd && (
        <button type="button" className={btnSuccess} onClick={handleNextRound} disabled={busy}>
          {t('nextRound')}
        </button>
      )}
    </div>
  );
}

const handAndFootDefinition: GameDefinition<
  HandAndFootResponse,
  HandAndFootGame,
  Parameters<typeof handandfootApi.exec>,
  HandAndFootView
> = {
  key: 'handandfoot',
  useGame,
  useView: useHandAndFootView,
  tutorial: HF_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', opponents: 3, centerCard: true, trickArea: true, footerHandSize: 11 },
  cli: () => ({
    parseCommand: parseHandAndFootCommand,
    formatResponse: formatHandAndFootState,
    helpText: HANDANDFOOT_HELP,
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase],
    ended: state.phase === HandAndFootPhase.GAME_END || !!state.gameEndFlag,
  }),
  shell: ({ view }) => ({
    isHumanTurn: view.isHumanTurn,
  }),
  reset: ({ game }) => {
    void game.exec('reset', undefined, {
      cpuDifficulty: game.handAndFootConfig.cpuDifficulty,
      pointLimit: game.handAndFootConfig.pointLimit,
    });
  },
  cardNav: ({ game, view }) => ({
    cardCount: view.humanCardCount,
    onToggle: game.toggleCard,
    onConfirm: view.kbdConfirmAction,
    onClear: game.clearSelection,
    enabled: view.isHumanTurn,
  }),
  shortcutsTestId: 'hand-and-foot-kbd-shortcuts',
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.handAndFootConfig.cpuDifficulty,
          options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'pointLimit',
          label: t('settings.pointLimit'),
          value: game.handAndFootConfig.pointLimit,
          options: POINT_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('pointLimit', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  actionsClassName: '',
  primaryActions,
  footerTall: true,
  LiveRegion,
  Board,
  FooterHand,
};

/** Hand and Foot game page. */
export const HandAndFootPage = createGamePage(handAndFootDefinition);
