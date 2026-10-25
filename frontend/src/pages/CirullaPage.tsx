import { useEffect, useState } from 'react';
import type { cirullaApi } from '../api/gameApi';
import { CardImage } from '../components/CardImage';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { CIRULLA_CPU_DIFFICULTY_OPTIONS, CIRULLA_TARGET_OPTIONS, useCirullaGame } from '../hooks/useCirullaGame';
import { useTutorialMessageParams } from '../providers/TutorialProvider';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { CirullaResponse } from '../types/card';
import { CirullaPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { CIRULLA_HELP, parseCirullaCommand } from '../utils/cli/commands/cirullaCommands';
import { formatCirullaState } from '../utils/cli/formatters/cirullaFormatter';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Cirulla tutorial step definitions. */
const CIRULLA_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="cirulla-table"]',
    messageKey: 'tutorial.table',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cirulla-scores"]',
    messageKey: 'tutorial.scores',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cirulla-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cirulla-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cirulla-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

type CirullaGame = ReturnType<typeof useCirullaGame>;
type CirullaView = ReturnType<typeof useCirullaView>;
type CirullaContext = GamePageContext<CirullaResponse, CirullaGame, CirullaView>;

function useCirullaView({ game, t }: { game: CirullaGame; t: CirullaContext['t'] }) {
  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  useTutorialMessageParams({
    targetScore: game.state ? game.state.config.targetScore : game.cirullaConfig.targetScore,
  });

  const { cardWidth, isMobile } = useCardDimensions();

  // **フェーズは文字列。** 共通の usePhaseNames は数値キーを前提にしている。
  const phaseNames: Readonly<Record<string, string>> = {
    [CirullaPhase.PLAY]: t('phase.play'),
    [CirullaPhase.ROUND_END]: t('phase.roundEnd'),
    [CirullaPhase.GAME_END]: t('phase.gameEnd'),
  };

  return { cardWidth, isMobile, phaseNames };
}

function Board({ game, state, t, view }: CirullaContext) {
  const { cardWidth } = view;
  const [hoveredCapture, setHoveredCapture] = useState<number[] | null>(null);

  // biome-ignore lint/correctness/useExhaustiveDependencies: Clear the hovered option whenever the selected hand changes.
  useEffect(() => {
    setHoveredCapture(null);
  }, [game.selectedHandIdx]);

  const isPlayPhase = state.phase === CirullaPhase.PLAY;
  const isRoundEnd = state.phase === CirullaPhase.ROUND_END;
  const isGameEnd = state.phase === CirullaPhase.GAME_END || state.gameEndFlag;
  const canPlay = isPlayPhase && state.isHumanTurn;

  // **取れる組はサーバが数えたものだけ。** 3 つの規則が絡むので、画面側で
  // 組み直すと必ずずれる。
  const options = game.selectedHandIdx !== null ? (state.captureOptions[game.selectedHandIdx] ?? []) : [];

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('round', { n: state.roundNumber, target: state.config.targetScore })}</span>
        <span data-testid="cirulla-deck">{t('deck', { n: state.deckRemaining })}</span>
      </div>

      <div className={lgTwoColGrid}>
        <div data-tutorial="cirulla-table">
          <div className="mb-1 text-ds-text-muted text-sm">{t('tableLabel')}</div>
          <div className="mb-2 p-2 rounded bg-black/30 flex flex-wrap gap-1" data-testid="cirulla-table">
            {state.table.length === 0 ? (
              <span className="text-ds-text-muted text-sm">{t('tableEmpty')}</span>
            ) : (
              state.table.map((c, i) => (
                <div
                  key={`${c.design}-${c.value}-${i}`}
                  data-testid={`cirulla-table-card-${i}`}
                  className={`flex flex-col items-center rounded ${hoveredCapture?.includes(i) ? 'ring-2 ring-ds-warning' : ''}`}
                >
                  <CardImage card={c} width={cardWidth} />
                  <span className="text-xs text-ds-text-muted">{i}</span>
                </div>
              ))
            )}
          </div>

          {/* **取れる組はサーバが数えたものだけ。** 同値・合計一致・合計 15・
              アッソの総取りが混ざるので、画面側で組み直すと必ずずれる。 */}
          {canPlay && game.selectedHandIdx !== null && (
            <div className="mb-2 p-2 rounded bg-black/30" data-testid="cirulla-capture-options">
              <div className="text-ds-text-primary text-sm mb-1">
                {options.length > 0 ? t('chooseCapture') : t('noCapture')}
              </div>
              <div className="flex flex-wrap gap-2">
                {options.map((group) => (
                  <button
                    key={group.join('-')}
                    type="button"
                    className={btnPrimary}
                    onClick={() => {
                      setHoveredCapture(null);
                      game.play(group);
                    }}
                    onMouseEnter={() => setHoveredCapture(group)}
                    onMouseLeave={() => setHoveredCapture(null)}
                    onFocus={() => setHoveredCapture(group)}
                    onBlur={() => setHoveredCapture(null)}
                    disabled={game.loading}
                    data-testid={`cirulla-take-${group.join('-')}`}
                    // **取り札ボタンの読み上げには場札の実際の名前が要る。** 索引だけでは
                    // スクリーンリーダー利用者にどの札の組かが分からない (#6628)。
                    aria-label={t('takeGroup', {
                      cards: group.map((i) => cardAlt(state.table[i])).join(t('listSeparator')),
                    })}
                  >
                    {t('takeGroupOption', {
                      indices: group.join(t('listSeparator')),
                      cards: group.map((i) => cardAlt(state.table[i])).join(t('listSeparator')),
                    })}
                  </button>
                ))}
                {options.length === 0 && (
                  <button
                    type="button"
                    className={btnSuccess}
                    onClick={() => game.play([])}
                    disabled={game.loading}
                    data-testid="cirulla-lay-off"
                  >
                    {t('layOffButton')}
                  </button>
                )}
              </div>
            </div>
          )}
        </div>

        <div data-tutorial="cirulla-scores">
          <div className="mb-2 p-2 rounded bg-black/30" data-testid="cirulla-scores">
            {state.players.map((p) => (
              <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                <span className="flex items-center gap-2">
                  <span>
                    {playerName(p.id, p.isHuman)}: {t('score', { n: p.score })} /{' '}
                    {t('taken', { n: p.capturedCount, denari: p.denariCount })} /{' '}
                    <span className="whitespace-nowrap" data-testid="cirulla-scope">
                      {t('scope', { n: p.scope })}
                    </span>
                  </span>
                  {p.isDealer && (
                    <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>{t('dealerBadge')}</span>
                  )}
                </span>
                {/* **配札ボーナスは出た瞬間に見せる。** 集計まで伏せると
                    なぜ点が動いたのか分からない。 */}
                {p.lastBonus && (
                  <span className="text-ds-warning text-xs" data-testid={`cirulla-bonus-${p.id}`}>
                    {t(`bonus.${p.lastBonus}`)}
                  </span>
                )}
              </div>
            ))}
          </div>

          {state.lastResult && (isRoundEnd || isGameEnd) && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="cirulla-round-result">
              <div className="mb-1 text-ds-text-primary">{t('roundResult')}</div>
              {state.lastResult.lines
                .filter((l) => l.points[0] !== 0 || l.points[1] !== 0)
                .map((l) => (
                  <div key={l.key}>
                    {t(`score.${l.key}`)}: {l.points[0]} - {l.points[1]}
                  </div>
                ))}
              <div className="text-ds-text-primary">
                {t('roundTotal', { a: state.lastResult.totals[0], b: state.lastResult.totals[1] })}
              </div>
              {state.lastResult.sweptDenari >= 0 && (
                <div className="text-ds-warning" data-testid="cirulla-swept-denari">
                  {t('sweptDenari', {
                    name: playerName(state.lastResult.sweptDenari, state.lastResult.sweptDenari === 0),
                  })}
                </div>
              )}
            </div>
          )}

          {isGameEnd && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-primary" data-testid="cirulla-winner">
              {t('winner', { name: playerName(state.winnerIdx, state.winnerIdx === 0) })}
            </div>
          )}
        </div>
      </div>

      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterHand({ game, state, t, view }: CirullaContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const isPlayPhase = state.phase === CirullaPhase.PLAY;
  const canPlay = isPlayPhase && state.isHumanTurn;
  const handValidIndices = canPlay ? humanPlayer?.cards.map((_, i) => i) : undefined;
  const captureIndices = canPlay
    ? state.captureOptions.map((cardOptions, i) => (cardOptions.length > 0 ? i : -1)).filter((i) => i >= 0)
    : undefined;

  if (!humanPlayer) return null;

  return (
    <div data-testid="cirulla-capture-markers">
      <PlayerHandSection
        humanPlayer={humanPlayer}
        selectedCardIndices={game.selectedHandIdx === null ? [] : [game.selectedHandIdx]}
        toggleCard={game.selectHand}
        cardWidth={view.cardWidth}
        isMobile={view.isMobile}
        dataTutorialPrefix="cirulla"
        validIndices={handValidIndices}
        legalIndices={handValidIndices}
        restrictedTooltip={t('restrictedTooltip')}
        cardBadgeFor={(idx) => (captureIndices?.includes(idx) ? { glyph: '取', title: t('captureAvailable') } : null)}
      />
    </div>
  );
}

function FooterStatus({ state, t, frontendHint }: CirullaContext) {
  return (
    <>
      {/* ライブ領域は常設 (#5955)。 */}
      <div data-testid="cirulla-hint-live" role="status" aria-live="polite">
        {isRequestedHint(state) && state.hintHandIdx >= 0 && (
          <div className="text-ds-warning text-sm mb-2">
            {t('hintAvailable')}: {t(`hint.${state.hintReason}`)} ([{state.hintHandIdx}]
            {state.hintCaptureIdxs.length > 0 && ` → ${state.hintCaptureIdxs.join(', ')}`})
          </div>
        )}
      </div>
      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
    </>
  );
}

function FooterActions({ game, state, t, busy }: CirullaContext) {
  const isRoundEnd = state.phase === CirullaPhase.ROUND_END;
  return (
    <>
      {isRoundEnd && (
        <button
          type="button"
          className={btnSuccess}
          onClick={game.handleNextRound}
          disabled={busy}
          data-testid="cirulla-next-round"
        >
          {t('nextRound')}
        </button>
      )}
    </>
  );
}

const cirullaDefinition: GameDefinition<
  CirullaResponse,
  CirullaGame,
  Parameters<typeof cirullaApi.exec>,
  CirullaView
> = {
  key: 'cirulla',
  useGame: useCirullaGame,
  useView: useCirullaView,
  tutorial: CIRULLA_TUTORIAL_STEPS,
  skeleton: { kind: 'casino-table', sections: [4], footerStyle: 'hand', footerHandSize: 3 },
  cli: () => ({
    parseCommand: parseCirullaCommand,
    formatResponse: formatCirullaState,
    helpText: CIRULLA_HELP,
  }),
  phase: ({ state, view }) => {
    const isGameEnd = state.phase === CirullaPhase.GAME_END || state.gameEndFlag;
    return {
      name: view.phaseNames[state.phase] ?? view.phaseNames[CirullaPhase.PLAY],
      ended: isGameEnd,
      won: isGameEnd && state.winnerIdx === 0,
    };
  },
  shell: ({ state }) => ({
    isHumanTurn: state.isHumanTurn,
  }),
  reset: ({ game }) => {
    game.reset();
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.cirullaConfig.cpuDifficulty,
          options: CIRULLA_CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'targetScore',
          label: t('settings.targetScore'),
          value: game.cirullaConfig.targetScore,
          options: CIRULLA_TARGET_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetScore', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  actionsTutorialId: 'cirulla-action-buttons',
  resetTutorialId: 'cirulla-reset-button',
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};

/**
 * Renders the Cirulla page: the Ligurian fishing game where sum-to-fifteen is
 * added to Scopa's captures and an ace can sweep the table.
 */
export const CirullaPage = createGamePage(cirullaDefinition);
