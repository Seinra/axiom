# Axiom full audit — feature document

Auditoría completa de Axiom en 7 categorías particionadas (sin solaparse, entre todas cubren el todo), en orden, con foco principal en clean code y legibilidad humana, usando los comandos nuevos de la familia foundry.

## Specs

- S1: "Si tuvieras que hacer una auditoria completa a Axion, pero dividiendo la tarea por categorias para no hacer una sola pasada general, sino varias cubriendo partes, y entre todas cubriendo el todo. Cuales y cuantas serian esas categorias?"
- S2: "Ok, larga en ese orden, con el enfoque principal en clean code y legibilidad para humanos, fijat que la flia foundry de skills se actualizo con algunas funciones extra relacionadas al tema"
- S3 (categorías acordadas): A1 API pública y contratos; A2 correctitud y tests; A3 concurrencia y resiliencia; A4 seguridad y riesgo; A5 stdlib-first y simplicidad; A6 docs y enseñabilidad; A7 gobernanza y delivery.
- S4 (método): auditors read-only (gentle-ai-verify) con foundry-review por lente (gate-clarity/gate-behavior/gate-ops/gate-risk/gate-simplicity/debt) + playbooks foundry ruteados; ledger por batch con severidad + evidence class + causal disposition + path:line; findings en observaciones separadas sin pisado.

## Tasks

- A1 (S2, S3): API pública y contratos — route: bounded read-only audit — commit: n/a (read-only)
- A2 (S2, S3): correctitud y tests — route: bounded read-only audit — commit: n/a
- A3 (S2, S3): concurrencia y resiliencia — route: bounded read-only audit — commit: n/a
- A4 (S2, S3): seguridad y riesgo — route: bounded read-only audit — commit: n/a
- A5 (S2, S3): stdlib-first y simplicidad — route: bounded read-only audit — commit: n/a
- A6 (S2, S3): docs y enseñabilidad — route: bounded read-only audit — commit: n/a
- A7 (S2, S3): gobernanza y delivery — route: bounded read-only audit — commit: n/a
- A8 (S2): síntesis + issues de hallazgos + close — route: parent — commit: docs

## Log

- L1 (request verbatim): "Si tuvieras que hacer una auditoria completa a Axion, pero dividiendo la tarea por categorias para no hacer una sola pasada general, sino varias cubriendo partes, y entre todas cubriendo el todo. Cuales y cuantas serian esas categorias?"
- L2 (request verbatim): "Ok, larga en ese orden, con el enfoque principal en clean code y legibilidad para humanos, fijat que la flia foundry de skills se actualizo con algunas funciones extra relacionadas al tema"
- L3 (2026-10-20): secuencia A1→A7 en orden, un auditor por vez; foco clean-code en todas con R2 Reader-First como lente transversal.
- L4 (2026-10-20): A1-A7 recibidos cada uno en batch propio (A3 con fallo de salida pero contenido completo); síntesis presentada; owner eligió solo-crear-issues — creados #22 (I1 hardening) + #23 (I2 bounds) + #24 (I3 concurrencia) + #25 (I4 docs) + #26 (I5 naming decisión) + #27 (I6 cuts decisión), read-back confirmado.
