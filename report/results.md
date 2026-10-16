Runs per cell: 1000 · seed: 42 · workers: 16. Counts describe effects and completion; see the README’s cost discussion for batch timing.

| Scenario | Strategy | Completed | Lost | Duplicate effects | Double-billed | Wasted / provider calls | Drain abandoned |
|---|---|---:|---:|---:|---:|---:|---:|
| baseline | naive | 100.0% | 0 | 0 | 0 | 0 / 6405 | 0 |
| baseline | at-least-once | 100.0% | 0 | 0 | 0 | 0 / 6405 | 0 |
| baseline | exactly-once | 100.0% | 0 | 0 | 0 | 0 / 6405 | 0 |
| crash-5 | naive | 70.3% | 297 | 0 | 0 | 293 / 5997 | 0 |
| crash-5 | at-least-once | 100.0% | 0 | 106 | 17 | 328 / 6733 | 0 |
| crash-5 | exactly-once | 100.0% | 0 | 0 | 0 | 0 / 6405 | 0 |
| dup-5 | naive | 100.0% | 0 | 234 | 39 | 325 / 6730 | 0 |
| dup-5 | at-least-once | 100.0% | 0 | 234 | 39 | 325 / 6730 | 0 |
| dup-5 | exactly-once | 100.0% | 0 | 0 | 0 | 0 / 6405 | 0 |
| hang-5 | naive | 70.6% | 294 | 0 | 0 | 293 / 5975 | 0 |
| hang-5 | at-least-once | 100.0% | 0 | 118 | 24 | 337 / 6742 | 0 |
| hang-5 | exactly-once | 100.0% | 0 | 0 | 0 | 163 / 6568 | 0 |
| error-10 | naive | 45.7% | 543 | 0 | 0 | 568 / 5560 | 0 |
| error-10 | at-least-once | 100.0% | 0 | 0 | 0 | 737 / 7142 | 0 |
| error-10 | exactly-once | 100.0% | 0 | 0 | 0 | 737 / 7142 | 0 |
| drain | naive | 99.4% | 6 | 0 | 0 | 12 / 6399 | 12 |
| drain | at-least-once | 100.0% | 0 | 0 | 0 | 12 / 6417 | 0 |
| drain | exactly-once | 100.0% | 0 | 0 | 0 | 0 / 6405 | 0 |
| storm | naive | 44.8% | 552 | 117 | 8 | 759 / 5707 | 10 |
| storm | at-least-once | 100.0% | 0 | 284 | 49 | 984 / 7389 | 0 |
| storm | exactly-once | 100.0% | 0 | 0 | 0 | 448 / 6853 | 0 |
