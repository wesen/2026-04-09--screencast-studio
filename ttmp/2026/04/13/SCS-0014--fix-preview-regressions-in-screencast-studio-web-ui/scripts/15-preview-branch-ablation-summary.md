---
Title: Preview branch ablation measurements summary
Ticket: SCS-0014
Status: active
Topics:
    - screencast-studio
    - gstreamer
    - video
    - performance
    - analysis
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Summary of the standalone benchmark that isolates which parts of the preview branch amplify recording cost the most.
LastUpdated: 2026-04-14T00:25:00-04:00
WhatFor: Explain the saved preview-branch ablation benchmark results and their current interpretation.
WhenToUse: Read when deciding whether preview branch transforms, JPEG encoding, or Go-side frame copying are the most promising next optimization targets.
---

# Preview branch ablation measurements summary

This note summarizes the preview-branch ablation benchmark captured under:

- `14-go-preview-branch-ablation-matrix/results/20260414-081748/`

## Goal

Isolate which part of the preview branch contributes most to the elevated cost seen when preview and recording run together on the same shared source.

## Scenarios

The benchmark used the real `2880x960 @ 24 fps` bottom-half region shape and measured:

1. `recorder-only`
2. `preview-fakesink-plus-recorder`
3. `preview-jpeg-discard-plus-recorder`
4. `preview-raw-copy-plus-recorder`
5. `preview-current-plus-recorder`
6. `preview-cheap-plus-recorder`

The preview variants were chosen to separate:

- second branch transform cost without downstream preview work,
- JPEG encode cost without Go frame-copy cost,
- raw Go frame-copy cost without JPEG,
- the current real preview path,
- and a much cheaper recording-mode preview profile.

## Results

Source: `14-go-preview-branch-ablation-matrix/results/20260414-081748/01-summary.md`

- recorder-only: **125.33% avg CPU**, **171.00% max CPU**
- preview-fakesink-plus-recorder: **134.00% avg CPU**, **176.00% max CPU**
- preview-jpeg-discard-plus-recorder: **137.83% avg CPU**, **182.00% max CPU**
- preview-raw-copy-plus-recorder: **141.33% avg CPU**, **197.00% max CPU**
- preview-current-plus-recorder: **152.33% avg CPU**, **204.00% max CPU**
- preview-cheap-plus-recorder: **112.00% avg CPU**, **143.00% max CPU**

Counters of interest:

- `preview-jpeg-discard-plus-recorder` pulled **61** preview samples but copied **0** preview bytes into Go
- `preview-raw-copy-plus-recorder` copied about **133 MB** of raw preview bytes into Go
- `preview-current-plus-recorder` copied about **13 MB** of JPEG preview bytes into Go
- `preview-cheap-plus-recorder` copied only about **1.36 MB** of preview bytes

## Main interpretation

This benchmark gives a more structured decomposition of the preview branch.

### High-confidence observations

1. Adding a preview branch at all increases total cost relative to recorder-only.
2. A branch that only goes to `fakesink` is already somewhat more expensive than recorder-only.
3. JPEG encode work and Go-side frame-copy work each appear to add cost on top of that baseline branch cost.
4. The current real preview path is more expensive than the `fakesink`-only or `jpeg-discard` variants.
5. A much cheaper preview profile can reduce the combined cost substantially.

### Important caveat

The absolute recorder-only baseline in this run (~125%) is higher than the earlier recorder-only benchmark runs (~94%). That means this benchmark should be interpreted more as a **within-run relative ablation** than as a replacement for the earlier reconciled recorder-only baseline.

The safest interpretation is therefore comparative within this run:

- `fakesink` preview branch adds some cost,
- JPEG adds more,
- raw byte copying adds more,
- the full current preview path is the most expensive of the non-cheap preview variants,
- and aggressive preview degradation can significantly reduce the combined cost.

## Practical takeaway

The preview branch problem is not just one thing.

The current best decomposition is:

1. **Having a second active branch already costs something**.
2. **JPEG encoding costs more**.
3. **Go-side frame copying costs more**.
4. **The full current preview path combines enough of those costs to make the combined case expensive**.
5. **A strong recording-mode preview downgrade looks promising as a mitigation**.

This does not yet prove the best final architecture, but it does support a practical next step:

> experiment with dynamically degrading preview while recording before attempting deeper recorder architecture changes.

## Recommended next experiments

1. Re-run the ablation matrix to check stability of the relative ordering.
2. Add more cheap-preview profiles to find a practical “recording mode preview” point.
3. Test a mode that disables continuous preview generation but preserves screenshot capability during recording.
4. Compare branch-local transform placement against more shared upstream transform work.
