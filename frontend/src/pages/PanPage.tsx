import { useMemo } from 'react';
import type { panApi } from '../api/gameApi';
import { CpuAccordion } from '../components/CpuAccordion';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { useCardDimensions, useIsMobile } from '../hooks/useCardDimensions';
import { CPU_DIFFICULTY_OPTIONS, PLAYER_COUNT_OPTIONS, TARGET_ROUNDS_OPTIONS, usePanGame } from '../hooks/usePanGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { PanPlayer, PanResponse } from '../types/card';
import { PanPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { PAN_HELP, parsePanCommand } from '../utils/cli/commands/panCommands';
import { formatPanState } from '../utils/cli/formatters/panFormatter';
import { hintLocalCommand } from '../utils/cli/hintText';
import { panLayoffIndices, panMeldCandidates } from '../utils/panMeldCandidates';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

const PAN_PHASE_KEYS: Readonly<Record<number, string>> = {
  [PanPhase.DRAW]: 'draw',
  [PanPhase.PLAY]: 'play',
  [PanPhase.ROUND_END]: 'roundEnd',
  [PanPhase.GAME_END]: 'gameEnd',
};

/** Panguingue (Pan) tutorial step definitions. */
const PAN_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="pan-draw-area"]', messageKey: 'tutorial.drawArea', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="pan-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="pan-meld-area"]', messageKey: 'tutorial.meldArea', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="pan-discard-button"]',
    messageKey: 'tutorial.discardButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pan-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="pan-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

type PanGame = ReturnType<typeof usePanGame>;
type PanView = ReturnType<typeof usePanView>;
type PanContext = GamePageContext<PanResponse, PanGame, PanView>;

function usePanView() {
  const isMobile = useIsMobile();
  const { cardWidth } = useCardDimensions();
  const phaseNames = usePhaseNames('pan', PAN_PHASE_KEYS);
  return { isMobile, cardWidth, phaseNames };
}

function PanMessageBox({ state, t }: Pick<PanContext, 'state' | 't'>) {
  const roundEndMessage =
    state.phase === PanPhase.ROUND_END
      ? state.panDeclarerIdx >= 0
        ? t('roundEndByPan', {
            name: playerName(state.panDeclarerIdx, state.players[state.panDeclarerIdx].isHuman),
          })
        : t('roundEndByStockOut')
      : undefined;
  return (
    <GameMessageBox
      message={roundEndMessage ?? state.message}
      messageCode={roundEndMessage ? undefined : state.messageCode}
      messageParams={roundEndMessage ? undefined : state.messageParams}
    />
  );
}

function Board({ game, state, t, busy, view }: PanContext) {
  const { isMobile, cardWidth } = view;
  const loading = busy;
  const humanPlayer = state.players.find((p) => p.isHuman);
  const isDrawPhase = state.phase === PanPhase.DRAW;
  const isPlayPhase = state.phase === PanPhase.PLAY;
  const isRoundEnd = state.phase === PanPhase.ROUND_END;
  const isGameEnd = state.phase === PanPhase.GAME_END || state.gameEndFlag;
  const revealCpu = isRoundEnd || isGameEnd;
  const isHumanTurn = (isDrawPhase || isPlayPhase) && state.players[state.currentPlayerIdx]?.isHuman === true;
  const canLayoff = isPlayPhase && isHumanTurn && game.selectedCardIndices.length === 1 && !loading;

  const renderPlayerMelds = (p: PanPlayer) =>
    p.laidMelds.map((meld, meldIdx) => (
      <fieldset
        key={`meld-${p.id}-${meldIdx}-${meld.cards.map((c) => `${c.design}${c.value}`).join('')}`}
        className={`border-0 p-0 m-0 min-w-0 flex flex-wrap items-center gap-1 ${isMobile ? 'mb-0' : 'mb-1'}`}
        aria-label={t('a11y.meldGroup', {
          name: playerName(p.id, p.isHuman),
          cards: meld.cards.map(cardAlt).join(t('listSeparator')),
        })}
      >
        {meld.cards.map((card, idx) => (
          <span
            key={`meldcard-${p.id}-${meldIdx}-${card.design}-${card.value}-${idx}`}
            className="inline-block"
            style={{ marginLeft: idx > 0 ? -Math.round(cardWidth * (isMobile ? 0.48 : 0.5) * 0.6) : 0 }}
          >
            <AnimatedCard card={card} width={cardWidth * (isMobile ? 0.48 : 0.5)} />
          </span>
        ))}
        {meld.chipUnits > 0 && (
          // A valle deals chips to every player; showing the unit beside its
          // meld makes the source clear on the table (#4853).
          <span
            className={`text-xs font-bold px-1.5 py-0.5 rounded ${badgeWarningColors}`}
            data-testid={`pan-chip-units-${p.id}-${meldIdx}`}
          >
            {t('chipUnits', { count: meld.chipUnits })}
          </span>
        )}
        {canLayoff && (
          <button
            type="button"
            className={btnPrimary}
            onClick={() => game.handleLayoff(p.id, meldIdx)}
            data-testid={`pan-layoff-${p.id}-${meldIdx}`}
            // Owner and cards distinguish this target from other layoff buttons (#6502).
            aria-label={t('a11y.layoffTo', {
              name: playerName(p.id, p.isHuman),
              meld: meld.cards.map(cardAlt).join(t('listSeparator')),
            })}
          >
            {t('layoffButton')}
          </button>
        )}
      </fieldset>
    ));

  // Meld-progress indicator toward the win condition (winMeldCount cards laid on
  // the table), highlighting the final stretch (<=2 cards left) to go out.
  const renderMeldProgress = (p: PanPlayer) => {
    const melded = p.meldedCount;
    const total = state.winMeldCount;
    const remaining = total - melded;
    const isClose = remaining > 0 && remaining <= 2;
    const pct = total > 0 ? Math.min(100, (melded / total) * 100) : 0;
    const label = t('meldProgress', { count: melded, total });
    return (
      <div data-testid="pan-meld-progress" className="mx-auto mb-2 max-w-xs">
        <div className="flex items-center justify-between text-xs mb-0.5">
          <span className="text-ds-text-muted">{label}</span>
          {isClose && <span className="text-ds-warning font-bold">{t('meldRemaining', { count: remaining })}</span>}
        </div>
        <div
          role="progressbar"
          aria-label={label}
          aria-valuemin={0}
          aria-valuemax={total}
          aria-valuenow={melded}
          className="h-1.5 w-full rounded-sm bg-white/15 overflow-hidden"
        >
          <div
            className={`h-full rounded-sm ${isClose ? 'bg-ds-warning' : 'bg-ds-accent'}`}
            style={{ width: `${pct}%` }}
          />
        </div>
      </div>
    );
  };

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('round', { n: state.roundNumber, total: state.targetRounds })}</span>
        {!(isDrawPhase && isHumanTurn) && <span className="mr-4">{t('drawPile', { count: state.drawPileCount })}</span>}
        <span>{t('winMeld', { count: state.winMeldCount })}</span>
      </div>

      {humanPlayer && renderMeldProgress(humanPlayer)}

      <div className={lgTwoColGrid}>
        {/* Left: game play area */}
        <div>
          {/* Discard pile top */}
          {state.discardTop && (
            <div className="my-3 p-3 rounded bg-black/40 flex items-center gap-3">
              <AnimatedCard card={state.discardTop} width={cardWidth} />
              <div className="text-ds-text-muted text-sm">
                <div>{t('discardTop')}</div>
              </div>
            </div>
          )}

          {/* Table melds (all players) with layoff targets */}
          <div className={`${isMobile ? 'my-1 p-1' : 'my-3 p-2'} rounded bg-black/30`} data-tutorial="pan-meld-area">
            <div className="text-ds-text-muted text-sm mb-1">{t('meldsTitle')}</div>
            {!humanPlayer?.laidMelds.length ? (
              <div className="text-ds-text-muted text-sm">{t('noMelds')}</div>
            ) : (
              state.players
                .filter((p) => p.isHuman && p.laidMelds.length > 0)
                .map((p) => (
                  <div key={`melds-${p.id}`} className={isMobile ? 'mb-1' : 'mb-2'}>
                    <div className="text-ds-text-muted text-xs mb-1">
                      {t('meldOwner', { name: playerName(p.id, p.isHuman) })}
                    </div>
                    {renderPlayerMelds(p)}
                  </div>
                ))
            )}
          </div>
        </div>

        {/* Right: info sidebar */}
        <div>
          {/* CPU players */}
          <CpuAccordion
            playerCount={state.players.filter((p) => !p.isHuman).length}
            openOnDesktop={false}
            forceOpen={canLayoff && state.players.some((p) => !p.isHuman && p.laidMelds.length > 0)}
          >
            {state.players
              .filter((p) => !p.isHuman)
              .map((p) => (
                <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
                  <div className="text-ds-text-muted text-sm">
                    {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                    {t('chips', { count: p.chips })} | {t('melded', { count: p.meldedCount })}
                  </div>
                  {p.laidMelds.length > 0 && (
                    <div className="mt-1">
                      <div className="text-ds-text-muted text-xs mb-1">
                        {t('meldOwner', { name: playerName(p.id, p.isHuman) })}
                      </div>
                      {renderPlayerMelds(p)}
                    </div>
                  )}
                  {revealCpu && p.cards.length > 0 && (
                    <div className="flex flex-wrap gap-1 mt-1">
                      {p.cards.map((card, idx) => (
                        <AnimatedCard
                          key={`cpu-${p.id}-${card.design}-${card.value}-${idx}`}
                          card={card}
                          width={cardWidth * 0.8}
                        />
                      ))}
                    </div>
                  )}
                </div>
              ))}
          </CpuAccordion>

          {/* Score table */}
          <details
            className="my-3 p-2 rounded bg-black/30"
            data-tutorial="pan-score-table"
            data-testid="pan-score-table"
            open={!isMobile || undefined}
          >
            <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('scores')}</summary>
            <table className="w-full text-sm text-ds-text-muted">
              <thead>
                <tr>
                  <th scope="col" className="text-left whitespace-nowrap">
                    {t('scoresPlayer')}
                  </th>
                  <th scope="col" className="whitespace-nowrap">
                    {t('scoresChips')}
                  </th>
                  {isRoundEnd && (
                    <th scope="col" className="whitespace-nowrap">
                      {t('scoresHandPoints')}
                    </th>
                  )}
                  <th scope="col" className="whitespace-nowrap">
                    {t('scoresRound')}
                  </th>
                  <th scope="col" className="whitespace-nowrap">
                    {t('scoresTotal')}
                  </th>
                </tr>
              </thead>
              <tbody>
                {state.players.map((p, playerIdx) => (
                  <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                    <td>
                      {playerName(p.id, p.isHuman)}
                      {playerIdx === state.dealerIdx && (
                        <span className="ml-1 rounded-sm bg-ds-surface px-1 text-ds-accent whitespace-nowrap">
                          {t('dealerBadge')}
                        </span>
                      )}
                    </td>
                    <td className="text-center">{p.chips}</td>
                    {isRoundEnd && <td className="text-center">{p.handPoints}</td>}
                    <td className="text-center">{p.roundScore}</td>
                    <td className="text-center">{p.cumulativeScore}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </details>
        </div>
      </div>

      {isMobile && <PanMessageBox state={state} t={t} />}
    </>
  );
}

function FooterHand({ game, state, t, frontendHint, view }: PanContext) {
  const { cardWidth } = view;
  const humanPlayer = state.players.find((p) => p.isHuman);

  // Pre-computed meld-candidate hints for the human's hand (only on their play
  // turn). `candidates` are minimal legal melds to tap-and-select; the two index
  // sets drive the additive ring marks on individual hand cards. Mirrors the
  // domain's meld/layoff rules exactly (see panMeldCandidates.ts).
  const {
    candidates: meldCandidates,
    candidateIndices,
    layoffIndices,
  } = useMemo(() => {
    const empty = {
      candidates: [] as ReturnType<typeof panMeldCandidates>,
      candidateIndices: new Set<number>(),
      layoffIndices: new Set<number>(),
    };
    if (!frontendHint.hintEnabled) return empty;
    const isPlay = state.phase === PanPhase.PLAY;
    const isHuman = state.players[state.currentPlayerIdx]?.isHuman === true;
    if (!humanPlayer || !isPlay || !isHuman) return empty;
    const candidates = panMeldCandidates(humanPlayer.cards);
    const candidateIndices = new Set<number>();
    for (const cand of candidates) for (const i of cand.indices) candidateIndices.add(i);
    const tableMelds = state.players.flatMap((p) => p.laidMelds.map((m) => m.cards));
    const layoffIndices = panLayoffIndices(humanPlayer.cards, tableMelds);
    return { candidates, candidateIndices, layoffIndices };
  }, [state, frontendHint.hintEnabled, humanPlayer]);

  return (
    <>
      {humanPlayer && meldCandidates.length > 0 && (
        <div className="mb-2 p-2 rounded bg-black/30" data-testid="pan-meld-candidates">
          <div className="text-ds-text-muted text-xs mb-1">{t('candidates.title')}</div>
          <div className="flex flex-wrap gap-2">
            {meldCandidates.map((cand) => {
              const cards = cand.indices.map((i) => humanPlayer.cards[i]);
              const kindLabel = t(cand.kind === 'set' ? 'candidates.set' : 'candidates.run');
              return (
                <button
                  type="button"
                  key={`cand-${cand.kind}-${cand.indices.join('-')}`}
                  onClick={() => game.selectCards(cand.indices)}
                  className={`${btnPrimary} flex items-center gap-1`}
                  aria-label={`${kindLabel}: ${cards.map(cardAlt).join(' ')}`}
                  data-testid={`pan-candidate-${cand.indices.join('-')}`}
                >
                  <span className="text-xs font-bold">{kindLabel}</span>
                  {cards.map((card, ci) => (
                    <AnimatedCard
                      key={`cand-card-${card.design}-${card.value}-${ci}`}
                      card={card}
                      width={cardWidth * 0.55}
                    />
                  ))}
                </button>
              );
            })}
          </div>
          <div className="text-ds-text-muted text-xs mt-1">{t('candidates.hint')}</div>
        </div>
      )}
      {humanPlayer && (
        <div className="flex flex-wrap gap-1 mb-2" data-tutorial="pan-player-hand">
          {humanPlayer.cards.map((card, idx) => {
            const isLayoff = layoffIndices.has(idx);
            const isCandidate = candidateIndices.has(idx);
            // Additive outline ring (stacks on the selection border without
            // clobbering it): warning = can lay off a single card onto a
            // table meld; success = part of a legal meld candidate.
            const hintRing = isLayoff
              ? { outline: '2px solid var(--color-ds-warning)', outlineOffset: '1px' }
              : isCandidate
                ? { outline: '2px solid var(--color-ds-success)', outlineOffset: '1px' }
                : {};
            return (
              <button
                type="button"
                key={`${card.design}-${card.value}-${idx}`}
                onClick={() => game.toggleCard(idx)}
                aria-label={cardAlt(card)}
                aria-pressed={game.selectedCardIndices.includes(idx)}
                data-meld-candidate={isCandidate ? 'true' : undefined}
                data-layoff-target={isLayoff ? 'true' : undefined}
                className={`transition-transform ${focusRingCard}`}
                style={{
                  background: 'none',
                  padding: 0,
                  borderRadius: 8,
                  ...selectedCardStyle(game.selectedCardIndices.includes(idx)),
                  ...hintRing,
                  boxSizing: 'border-box',
                }}
              >
                <AnimatedCard card={card} width={cardWidth} />
              </button>
            );
          })}
        </div>
      )}
    </>
  );
}

function FooterStatus({ frontendHint, t }: PanContext) {
  return <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />;
}

function primaryActions({ game, state, t, busy }: PanContext) {
  const isDrawPhase = state.phase === PanPhase.DRAW;
  const isPlayPhase = state.phase === PanPhase.PLAY;
  const isRoundEnd = state.phase === PanPhase.ROUND_END;
  const isHumanTurn = (isDrawPhase || isPlayPhase) && state.players[state.currentPlayerIdx]?.isHuman === true;
  const loading = busy;

  if (!((isDrawPhase && isHumanTurn) || (isPlayPhase && isHumanTurn) || isRoundEnd)) {
    return null;
  }

  return (
    <div className="flex gap-2 items-center justify-center flex-wrap">
      {isDrawPhase && isHumanTurn && (
        <div className="flex gap-2 items-center" data-tutorial="pan-draw-area">
          <span className="text-ds-text-muted text-sm" data-testid="pan-draw-pile-count">
            {t('drawPile', { count: state.drawPileCount })}
          </span>
          {state.drawPileCount === 0 && <span className="text-ds-warning text-sm">{t('drawStockEmpty')}</span>}
          <button type="button" className={btnPrimary} onClick={game.handleDrawStock} disabled={loading}>
            {t('drawStockButton')}
          </button>
          <button
            type="button"
            className={btnPrimary}
            onClick={game.handleDrawDiscard}
            disabled={loading || !state.discardTop}
          >
            {t('drawDiscardButton')}
          </button>
        </div>
      )}
      {isPlayPhase && isHumanTurn && (
        <>
          <button
            type="button"
            className={btnPrimary}
            onClick={game.handleMeld}
            disabled={loading || game.selectedCardIndices.length < 3}
            data-testid="pan-meld-button"
          >
            {t('meldButton')}
          </button>
          <button
            type="button"
            className={btnPrimary}
            onClick={game.handleDiscard}
            disabled={loading || game.selectedCardIndices.length !== 1}
            data-tutorial="pan-discard-button"
            data-testid="pan-discard-button"
          >
            {t('discardButton')}
          </button>
        </>
      )}
      {isRoundEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextRound} disabled={loading}>
          {t('nextRound')}
        </button>
      )}
    </div>
  );
}

const panDefinition: GameDefinition<PanResponse, PanGame, Parameters<typeof panApi.exec>, PanView> = {
  key: 'pan',
  useGame: usePanGame,
  useView: usePanView,
  tutorial: PAN_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', opponents: 1, centerCard: true, trickArea: true, footerHandSize: 10 },
  cli: ({ frontendHint }) => ({
    parseCommand: parsePanCommand,
    formatResponse: formatPanState,
    helpText: PAN_HELP,
    localCommand: hintLocalCommand(frontendHint.hint),
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase],
    ended: state.phase === PanPhase.GAME_END || state.gameEndFlag,
  }),
  shell: ({ state }) => ({
    isHumanTurn:
      (state.phase === PanPhase.DRAW || state.phase === PanPhase.PLAY) &&
      state.players[state.currentPlayerIdx]?.isHuman === true,
  }),
  reset: ({ game }) => {
    void game.exec('reset', undefined, {
      playerCount: game.panConfig.playerCount,
      cpuDifficulty: game.panConfig.cpuDifficulty,
      targetRounds: game.panConfig.targetRounds,
    });
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'playerCount',
          label: t('settings.playerCount'),
          value: game.panConfig.playerCount,
          options: PLAYER_COUNT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('playerCount', v),
        },
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.panConfig.cpuDifficulty,
          options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'targetRounds',
          label: t('settings.targetRounds'),
          value: game.panConfig.targetRounds,
          options: TARGET_ROUNDS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetRounds', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  primaryActions,
  message: (ctx) => (!ctx.view.isMobile ? <PanMessageBox state={ctx.state} t={ctx.t} /> : null),
  cardNav: ({ game, state }) => {
    const isPlayPhase = state.phase === PanPhase.PLAY;
    const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
    return {
      cardCount: state.players.find((p) => p.isHuman)?.cards?.length ?? 0,
      onToggle: game.toggleCard,
      onConfirm: () => {
        if (isPlayPhase) game.handleDiscard();
      },
      onClear: game.clearSelection,
      enabled: isHumanTurn,
    };
  },
  shortcutsTestId: 'pan-kbd-shortcuts',
  resetTutorialId: 'pan-reset-button',
  actionsClassName: 'flex gap-2 items-center flex-wrap',
  controlsInFooter: true,
  Board,
  FooterHand,
  FooterStatus,
};

/** Renders the Panguingue (Pan) game page with draw, play, and layoff phases. */
export const PanPage = createGamePage(panDefinition);
