# Axiom official docs — documentation track (research first)

Official Axiom documentation authored with the docs trio (`foundry` + `foundry-docs` + `cognitive-doc-design`), grounded in reference examples from official Go, Gin, Echo, GORM docs plus other docs worth investigating. Research (web evidence) runs FIRST, then design lock, then docs. Nothing is implemented until the owner picks references.

## Specs

- S1: "Ok, foundry-docs existe ahora, añadila al review packs porque es la skill que se encarga de todo lo relacionado a documentacion, junto a cognitive-doc-design. La vas a usar para crear la documentacion oficial de Axiom, basandote en los ejemplos de la documentacion oficial de Go, Gin, Echo, GORM y cualquier otra documentacion que consideres importante investigar como referencia."
- S2: "Seria foundry, foundry-docs y cognitive-doc-design el trio de implementar para docs"
- S3 (scope, parent proposal — needs owner sign-off): official docs = quickstart (<5min) + guides per family + cookbook recipes + API reference entries, every page with version + prereqs + runnable example per `foundry-docs` output contract; doc comments are the source, prose links never duplicates; standalone in-repo (VISION invariant 8).
- S4 (completion bar, owner-confirmable): quickstart verified <5min by a fresh reader; every guide has concept → steps → example → next step; every recipe standalone runnable with expected output; API entries carry signature + doc comment + example; zero stale versions/counts; drift check documented.

## Tasks (in order; T3+ detailed at design lock)

- T1 (S1): INVESTIGATE — parent-inline web evidence, lanes: (RQ1) Go official docs (go.dev/doc, godoc conventions, Tour of Go) as doc-comment-source model; (RQ2) Gin + Echo official docs (guide/cookbook/reference split, runnable examples); (RQ3) GORM official docs (guides per feature + API ref) + other docs worth investigating (Rust Book/std docs, Kotlin docs/Koans, Swift TSPL/DocC, diátaxis). Route: read-only research, cited findings with pins. No implementation.
- T2 (S3, S4): design lock — doc map (quickstart/guide/cookbook/API per foundry-docs structure + diátaxis), depth per topic (sizing), voice + versioning rules, placement in repo. Route: parent + owner (sign-off).
- T3 (S4): implement docs in bounded surface batches with the trio injected. Route: decided at T2 sign-off under the Writer rule. NO native review for docs-only (docs rule); build gate + independent verifier.
- T4: independent verifier (docs lens) + close-out + delivery. Route: parent commits on feature branch; push/PR/merge owner-gated.

## Log

- L1 (request verbatim): "Ok, foundry-docs existe ahora, añadila al review packs porque es la skill que se encarga de todo lo relacionado a documentacion, junto a cognitive-doc-design. La vas a usar para crear la documentacion oficial de Axiom, basandote en los ejemplos de la documentacion oficial de Go, Gin, Echo, GORM y cualquier otra documentacion que consideres importante investigar como referencia."
- L2 (request verbatim): "Seria foundry, foundry-docs y cognitive-doc-design el trio de implementar para docs"
- L3 (2026-10-21): loop updated for the trio (REVIEW_PACKS docs-units row + tiers + checklist + FOUNDRY-DOCS ref ID; REFERENCES FOUNDRY-DOCS v1.0 pin; config IMPLEMENT skills + trio gate; committed 97459dc on main); ODD prepared with research staged; no implementation until owner picks references.
- L4 (2026-10-21, T1 DONE + T2 signed off): web evidence RQ1 Go (go.dev/doc, godoc blog, Tour) + RQ2 Gin (quickstart + examples repo) + Echo (guide/cookbook/reference, echox) + RQ3 GORM (guides-per-feature + generics API) + prior Rust/Kotlin/Swift/diataxis. Owner lock: modelo completo foundry-docs (quickstart + guías + cookbook + API ref), carpeta docs/ en repo (sitio a futuro), full en los 15, slice grande. Next: T3 implementation.
