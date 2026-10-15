Runs per cell: 1000 · seed: 42 · workers: 16. Latency includes queue wait; percentiles cover billed runs.

| Scenario | Strategy | Completed | Lost | Duplicate effects | Double-billed | Wasted / provider calls | Drain abandoned | p50 | p99 |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|
| baseline | naive | 100.0% | 0 | 0 | 0 | 0 / 6405 | 0 | 159.48s | 170.69s |
| baseline | at-least-once | 100.0% | 0 | 0 | 0 | 0 / 6405 | 0 | 159.48s | 170.69s |
| baseline | exactly-once | 100.0% | 0 | 0 | 0 | 0 / 6405 | 0 | 159.48s | 170.69s |
| crash-5 | naive | 70.3% | 297 | 0 | 0 | 293 / 5997 | 0 | 145.18s | 153.15s |
| crash-5 | at-least-once | 100.0% | 0 | 106 | 17 | 328 / 6733 | 0 | 162.60s | 177.14s |
| crash-5 | exactly-once | 100.0% | 0 | 0 | 0 | 0 / 6405 | 0 | 156.39s | 168.57s |
| dup-5 | naive | 100.0% | 0 | 234 | 39 | 325 / 6730 | 0 | 166.95s | 178.58s |
| dup-5 | at-least-once | 100.0% | 0 | 234 | 39 | 325 / 6730 | 0 | 166.95s | 178.58s |
| dup-5 | exactly-once | 100.0% | 0 | 0 | 0 | 0 / 6405 | 0 | 159.48s | 170.69s |
| hang-5 | naive | 70.6% | 294 | 0 | 0 | 293 / 5975 | 0 | 177.88s | 186.93s |
| hang-5 | at-least-once | 100.0% | 0 | 118 | 24 | 337 / 6742 | 0 | 201.60s | 219.38s |
| hang-5 | exactly-once | 100.0% | 0 | 0 | 0 | 163 / 6568 | 0 | 197.21s | 213.37s |
| error-10 | naive | 45.7% | 543 | 0 | 0 | 568 / 5560 | 0 | 135.22s | 140.43s |
| error-10 | at-least-once | 100.0% | 0 | 0 | 0 | 737 / 7142 | 0 | 173.89s | 190.01s |
| error-10 | exactly-once | 100.0% | 0 | 0 | 0 | 737 / 7142 | 0 | 173.89s | 190.01s |
| drain | naive | 99.4% | 6 | 0 | 0 | 12 / 6399 | 12 | 159.15s | 170.19s |
| drain | at-least-once | 100.0% | 0 | 0 | 0 | 12 / 6417 | 0 | 159.45s | 170.73s |
| drain | exactly-once | 100.0% | 0 | 0 | 0 | 0 / 6405 | 0 | 159.21s | 170.50s |
| storm | naive | 44.8% | 552 | 117 | 8 | 759 / 5707 | 10 | 154.30s | 159.98s |
| storm | at-least-once | 100.0% | 0 | 284 | 49 | 984 / 7389 | 0 | 202.82s | 220.95s |
| storm | exactly-once | 100.0% | 0 | 0 | 0 | 448 / 6853 | 0 | 189.83s | 205.20s |
