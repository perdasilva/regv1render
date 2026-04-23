# Namespace Mode Documentation and Demo

Document all four OLM namespace modes with clear CLI and config file examples. Generate a demo GIF showing the rv1 CLI in action and embed it in the README.

## Task Group 1: Namespace Modes README section (small)

Add a "Namespace Modes" section to the README explaining each mode with CLI examples and config file snippets.

- Add section after "OLMv0 Compatibility" covering:
  - AllNamespaces — default when bundle supports it, no `--watch-namespace` needed
  - OwnNamespace — `--watch-namespace <same-as-install-ns>`
  - SingleNamespace — `--watch-namespace <other-ns>`
  - MultiNamespace — `--watch-namespace ns1 --watch-namespace ns2`
- Include both CLI flag and config file examples for each mode
- Explain the default behavior (AllNamespaces when supported)

## Task Group 2: Help text update (small)

Update `rv1 render --help` to describe namespace behavior.

- Add namespace mode description to the long help text
- Mention default behavior and how to set each mode

## Task Group 3: Demo GIF (small)

Generate an asciinema demo GIF and embed it in the README.

- Record a demo showing rv1 rendering a bundle with different namespace modes
- Convert recording to GIF and save as `assets/demo.gif`
- Embed the GIF in the README near the top

## Task Group 4: Review (small)

Verify all docs are accurate and consistent.

- Verify namespace mode examples work with a real bundle
- Run `make verify`
- Read README end-to-end for coherence
