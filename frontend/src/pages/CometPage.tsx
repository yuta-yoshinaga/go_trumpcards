import { useEffect, useMemo } from 'react';
import type { cometApi } from '../api/gameApi';
import { CardImage } from '../components/CardImage';
import { CardRoleBadge } from '../components/CardRoleBadge';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { useCardDimensions } from '../hooks/useCardDimensions';
import {
  COMET_CPU_DIFFICULTY_OPTIONS,
  COMET_PLAYER_OPTIONS,
  COMET_TARGET_OPTIONS,
  useCometGame,
} from '../hooks/useCometGame';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { CometResponse } from '../types/card';
import { CometPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { COMET_HELP, parseCometCommand } from '../utils/cli/commands/cometCommands';
import { formatCometState } from '../utils/cli/formatters/cometFormatter';
import { formatSignedDelta } from '../utils/formatSignedDelta';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Cards of the current sequence to show; older ones scroll off. */
const PILE_TAIL = 8;

const isCometWild = (card: { design: string; value: number }) => card.design === 'DIAMOND' && card.value === 9;

/** Comet tutorial step definitions. */
const COMET_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="comet-pile"]',
    messageKey: 'tutorial.pile',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="comet-scores"]',
    messageKey: 'tutorial.scores',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="comet-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="comet-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="comet-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

type CometGame = ReturnType<typeof useCometGame>;
type CometView = ReturnType<typeof useCometView>;
type CometContext = GamePageContext<CometResponse, CometGame, CometView>;

function useCometView({ game, t }: { game: CometGame; t: CometContext['t'] }) {
  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const { cardWidth, isMobile } = useCardDimensions();
  // **フェーズは文字列。** 共通の usePhaseNames は数値キーを前提にしている。
  const phaseNames: Readonly<Record<string, string>> = useMemo(
    () => ({
      [CometPhase.PLAY]: t('phase.play'),
      [CometPhase.ROUND_END]: t('phase.roundEnd'),
      [CometPhase.GAME_END]: t('phase.gameEnd'),
    }),
    [t],
  );

  return { cardWidth, isMobile, phaseNames };
}

function AboveContent({ game, state, t }: CometContext) {
  const { cometConfig } = game;
  return (
    <div role="status" aria-live="polite" data-testid="comet-settings-status">
      {(cometConfig.cpuDifficulty !== state.config.cpuDifficulty ||
        cometConfig.players !== state.config.players ||
        cometConfig.targetScore !== state.config.targetScore) && (
        <p className="mx-4 mt-1 text-sm text-ds-warning">{t('settings.pendingChanges')}</p>
      )}
    </div>
  );
}

function Board({ state, t, view }: CometContext) {
  const { cardWidth } = view;
  const isRoundEnd = state.phase === CometPhase.ROUND_END;
  const isGameEnd = state.phase === CometPhase.GAME_END || state.gameEndFlag;
  // **クロージャの中で narrow が効くように束ねる。** `state.lastResult` のままだと
  // map のコールバック内で null 可能に戻り、起きない分岐が増える。
  const lastResult = state.lastResult;
  const shownPile = state.pile.slice(-PILE_TAIL);
  const omittedPileCount = state.pile.length - shownPile.length;

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('round', { n: state.roundNumber, target: state.config.targetScore })}</span>
        {/* **死に手の枚数は見せる。** ここに眠った札で連なりが止まる。 */}
        <span data-testid="comet-dead">{t('dead', { n: state.deadCount })}</span>
      </div>

      <div className={lgTwoColGrid}>
        <section data-tutorial="comet-pile" aria-labelledby="comet-pile-label">
          <div id="comet-pile-label" className="mb-1 text-ds-text-muted text-sm">
            {t('pileLabel')}
          </div>
          <div className="mb-2 p-2 rounded bg-black/30 flex flex-wrap gap-1 items-center" data-testid="comet-pile">
            {omittedPileCount > 0 && (
              <span id="comet-pile-omitted" className="text-ds-text-muted text-sm" data-testid="comet-pile-omitted">
                {t('pileOmitted', { count: omittedPileCount })}
              </span>
            )}
            {shownPile.length === 0 ? (
              <span className="text-ds-text-muted text-sm">{t('pileEmpty')}</span>
            ) : (
              <ol
                aria-labelledby="comet-pile-label"
                aria-describedby={omittedPileCount > 0 ? 'comet-pile-omitted' : undefined}
                className="flex flex-wrap gap-1 items-center"
              >
                {shownPile.map((c, i) => (
                  <li key={`${c.design}-${c.value}-${i}`} className="relative list-none">
                    <CardImage
                      card={c}
                      width={cardWidth}
                      ariaLabel={isCometWild(c) ? `${cardAlt(c)} (${t('wildMark')})` : undefined}
                    />
                    {isCometWild(c) && <CardRoleBadge idx={i} glyph="★" title={t('wildMark')} />}
                  </li>
                ))}
              </ol>
            )}
          </div>
          {/* **スートは問わない。** 数字だけで昇ることを毎回書く。 */}
          {state.need > 0 && (
            <div className="text-ds-text-primary text-sm" data-testid="comet-need">
              {t('need', { rank: state.need })}
            </div>
          )}
        </section>

        <div data-tutorial="comet-scores">
          <div className="mb-2 p-2 rounded bg-black/30" data-testid="comet-scores">
            {state.players.map((p) => (
              <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                <span className="flex items-center gap-2">
                  <span>
                    {playerName(p.id, p.isHuman)}: {t('cards', { n: p.cardCount })} / {t('score', { n: p.score })}
                  </span>
                  {p.isDealer && (
                    <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>{t('dealerBadge')}</span>
                  )}
                </span>
              </div>
            ))}
          </div>

          {lastResult && (isRoundEnd || isGameEnd) && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="comet-round-result">
              <div className="mb-1 text-ds-text-primary">
                {t('goOut', {
                  name: playerName(lastResult.winnerIdx, lastResult.winnerIdx === 0),
                  points: lastResult.gained[lastResult.winnerIdx],
                })}
              </div>
              {/* **席の一覧を正本にする。** 残り枚数は席ごとに1つなので、
                  枚数側を回して席を引き当てると「席が無いとき」という
                  起きない分岐が要る。 */}
              {state.players.map((p, idx) => (
                <div key={p.id} data-testid={`comet-cards-left-${idx}`}>
                  {t('cardsLeft', {
                    name: playerName(p.id, p.isHuman),
                    n: lastResult.cardsLeft[idx],
                  })}
                </div>
              ))}
              {state.players.map((p, idx) => (
                <div key={p.id} data-testid={`comet-round-score-${idx}`}>
                  {t('roundScore', {
                    name: playerName(p.id, p.isHuman),
                    gained: formatSignedDelta(lastResult.gained[idx]),
                    total: p.score,
                  })}
                </div>
              ))}
              <div>{t('unplayedKings', { n: lastResult.unplayedKings })}</div>
              {lastResult.heldWildIdx >= 0 && (
                <div className="text-ds-error-text" data-testid="comet-held-wild">
                  {t('heldWild', {
                    name: playerName(lastResult.heldWildIdx, lastResult.heldWildIdx === 0),
                  })}
                </div>
              )}
            </div>
          )}

          {isGameEnd && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-primary" data-testid="comet-winner">
              {t('winner', { name: playerName(state.winnerIdx, state.winnerIdx === 0) })}
            </div>
          )}
        </div>
      </div>

      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterHand({ game, state, t, view }: CometContext) {
  const { cardWidth, isMobile } = view;
  const humanPlayer = state.players.find((p) => p.isHuman);
  if (!humanPlayer) return null;

  const isPlayPhase = state.phase === CometPhase.PLAY;
  const canPlay = isPlayPhase && state.isHumanTurn;
  // **出せる札はサーバが数えたものだけ。** コメットがどのランクの代わりにも
  // なるので、画面側で組み直すと必ずずれる。
  const playable = canPlay ? state.playableIdxs : [];

  return (
    <PlayerHandSection
      humanPlayer={humanPlayer}
      selectedCardIndices={[]}
      toggleCard={game.play}
      cardWidth={cardWidth}
      isMobile={isMobile}
      dataTutorialPrefix="comet"
      validIndices={canPlay ? playable : undefined}
      legalIndices={canPlay ? playable : undefined}
      restrictedTooltip={t('restrictedTooltip')}
      cardBadgeFor={(idx) => {
        const card = humanPlayer.cards[idx];
        return isCometWild(card) ? { glyph: '★', title: t('wildMark') } : null;
      }}
    />
  );
}

function FooterStatus({ state, t, frontendHint }: CometContext) {
  return (
    <>
      {/* ライブ領域は常設 (#5955)。 */}
      <div data-testid="comet-hint-live" role="status" aria-live="polite">
        {isRequestedHint(state) && state.hintHandIdx >= 0 && (
          <div className="text-ds-warning text-sm mb-2">
            {t('hintAvailable')}: {t(`hint.${state.hintReason}`)} ([{state.hintHandIdx}])
          </div>
        )}
      </div>
      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
    </>
  );
}

function FooterActions({ game, state, t, busy }: CometContext) {
  const isPlayPhase = state.phase === CometPhase.PLAY;
  const isRoundEnd = state.phase === CometPhase.ROUND_END;
  const isGameEnd = state.phase === CometPhase.GAME_END || state.gameEndFlag;
  const canPlay = isPlayPhase && state.isHumanTurn;
  const playable = canPlay ? state.playableIdxs : [];
  const mustPass = canPlay && playable.length === 0;

  return (
    <>
      {isPlayPhase && !state.isHumanTurn && !isGameEnd && (
        <span className="text-ds-text-muted text-sm" data-testid="comet-waiting">
          {t('waitingForTurn')}
        </span>
      )}
      {/* **パスは出せる札が無いときだけ。** 常に出しておくと、出せるのに
          押して弾かれるだけのボタンになる。 */}
      {mustPass && (
        <button type="button" className={btnSecondary} onClick={game.pass} disabled={busy} data-testid="comet-pass">
          {t('passButton')}
        </button>
      )}
      {isRoundEnd && (
        <button
          type="button"
          className={btnSuccess}
          onClick={game.handleNextRound}
          disabled={busy}
          data-testid="comet-next-round"
        >
          {t('nextRound')}
        </button>
      )}
    </>
  );
}

const cometDefinition: GameDefinition<CometResponse, CometGame, Parameters<typeof cometApi.exec>, CometView> = {
  key: 'comet',
  useGame: useCometGame,
  useView: useCometView,
  tutorial: COMET_TUTORIAL_STEPS,
  skeleton: { kind: 'casino-table', sections: [1], footerStyle: 'hand', footerHandSize: 8 },
  cli: () => ({
    parseCommand: parseCometCommand,
    formatResponse: formatCometState,
    helpText: COMET_HELP,
  }),
  phase: ({ state, view }) => ({
    name: view.phaseNames[state.phase] ?? view.phaseNames[CometPhase.PLAY],
    ended: state.phase === CometPhase.GAME_END || state.gameEndFlag,
    won: (state.phase === CometPhase.GAME_END || state.gameEndFlag) && state.winnerIdx === 0,
  }),
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
          value: game.cometConfig.cpuDifficulty,
          options: COMET_CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'players',
          label: t('settings.players'),
          value: game.cometConfig.players,
          options: COMET_PLAYER_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('players', v),
        },
        {
          type: 'select',
          id: 'targetScore',
          label: t('settings.targetScore'),
          value: game.cometConfig.targetScore,
          options: COMET_TARGET_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetScore', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  actionsTutorialId: 'comet-action-buttons',
  resetTutorialId: 'comet-reset-button',
  AboveContent,
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};

/**
 * Renders the Comet page: the ancestor of the stops family, where sequences
 * climb by rank ignoring suit and the 9 of diamonds is a wild Comet.
 */
export const CometPage = createGamePage(cometDefinition);
