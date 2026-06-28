# Cloudflare Workers (WASM)

Games are deployed to Cloudflare Workers as WASM binaries via TinyGo. Three workers split games by category:

| Worker | Entry point | Games |
|--------|-------------|-------|
| **casino** | `cmd/workers/casino/main.go` | Table & poker games (blackjack, baccarat, poker, holdem, omaha, omahahilo, bigo, bigohilo, shortdeck, pineapple, crazypineapple, irishpoker, indianpoker, videopoker, deuceswild, jokerpoker, threecard, fourcardpoker, caribbeanstud, texasholdembonus, ultimatetexasholdem, mississippistud, sevencardstud, paigow, chinesepoker, letitride, reddog, razz, badugi, deucetoseven, spanish21, casinowar, dragontiger, blackjackswitch, oasispoker, russianpoker, casinoholdem, highcardflush, yaniv, tressette, bourre, napoleon, mighty, bridge, skat, belote, tarneeb, sheepshead, doppelkopf, mus, tute, sueca, fortyfives, twentynine, courtpiece, ecarte, threecardbrag, teenpatti, kemps, pishti, cuarenta, fivecardstud, faro, openfacechinese) |
| **classic** | `cmd/workers/classic/main.go` | Trick-taking, matching & fishing (hearts, spades, pitch, twotenjack, callbreak, briscola, oldmaid, doubt, daifugo, bigtwo, sevens, crazyeights, ohhell, speed, gofish, pinochle, pigtail, durak, war, fiftyone, whist, pageone, trash, president, cassino, spiteandmalice, shithead, nertz, slapjack, egyptianratscrew, tonk, sixcardgolf, truco, klaverjas, manille, marias, sedma, solowhist, knockoutwhist, nap, preference, spoilfive, doudizhu, tichu, scopa, scopone, escoba, bezique, cuckoo, spoons, labellelucie, simplesimon, doubleklondike) |
| **solo** | `cmd/workers/solo/main.go` | Solitaire & rummy (klondike, freecell, seahaventowers, cruel, spider, spiderette, pyramid, tripeaks, memory, ginrummy, conquian, chinchon, threethirteen, canasta, handandfoot, cribbage, golf, clocksolitaire, fortythieves, canfield, yukon, russiansolitaire, scorpion, wasp, accordion, pokersquares, montecarlo, contractrummy, kalooki, calculation, bakersdozen, beleagueredcastle, sevenbridge, crescent, gaps, rummy500, eightoff, penguin, acesup, barbu, macau, mao, thirtyone, tienlen, osmosis, fivehundred, schnapsen, burraco, gongzhu, bristol, bidwhist, easthaven, bakersgame, euchre, piquet, russianbank, blackhole) |

The worker entry points (`cmd/workers/{casino,classic,solo}/main.go`) are thin shells that blank-import the matching `internal/infrastructure/games/<category>` sub-package and call `games.RegisterCategory(mux, games.Category…)`. The registry itself (`internal/infrastructure/games/registry.go`) stores `{Name, Category}` for each game; the human-readable descriptions live in a separate `descriptions.go` map (build-tagged `//go:build !js || !wasm` to keep them out of the WASM binaries). The Web-server factories live in `games_server.go` (excluded from WASM via build tags) and the Worker bindings live in per-category sub-packages — this split is what keeps each Cloudflare Worker binary under the 1 MB gzipped free-tier limit by letting TinyGo dead-code-eliminate the games from the other two categories.

**When adding/modifying a game, always update:**
1. `internal/infrastructure/games/registry.go` — `{Name, Category}` entry (selects the worker), plus the matching CLI display title in `descriptions.go`
2. `internal/infrastructure/games/games_server.go` — `BindWebControllerFor("<name>", …)` for the HTTP server factory
3. `internal/infrastructure/games/{casino,classic,solo}/<category>.go` — `games.RegisterKVGame("<name>", games.Category…, …)` for the KV-backed worker route (must match the `Category`)
4. `frontend/src/api/gameApi.ts` `workerUrl` — must match the `Category`

Build: `make build-worker-{solo,casino,classic}` or `make build-workers` (requires TinyGo).
