# Standard: benchmarking and performance budgets

Platform-agnostic. A performance claim without a method and a recorded result is an opinion. Never publish or write down numbers that were not measured.

## Performance budget (in the spec)

Every spec states, before implementation:

- **Workload model**: request mix, rates, payload sizes, data volume, concurrency, and growth assumption.
- **Targets** as percentiles, not averages (for example p50, p95, p99 latency; sustained throughput; error rate under load).
- **Resource limits**: CPU, memory, connections, storage per unit of load.
- **The benchmark** that will prove it (name and where it lives).

Targets are defaults to tune; record the reasoning (what the user needs, what the budget costs).

## Kinds of benchmark

| Kind | Question | Notes |
|---|---|---|
| Micro | How fast is this function or query? | Isolated, run many iterations; guards hot paths |
| Load | Does the system meet the budget at expected load? | Realistic workload and data |
| Stress | Where does it break, and how? | Ramp past expected load; observe failure mode |
| Soak | Does it stay healthy over hours? | Finds leaks, slow growth, lag build-up |
| Spike | How does it handle sudden bursts and recovery? | Include dependency slowdowns |

## Method

1. **Baseline first.** Measure before changing anything; keep the baseline for comparison.
2. **Representative environment and data**: same configuration class as the target, realistic data volume and distribution, warm caches reported separately from cold.
3. **Warm up**, then measure a steady state; discard the warm-up.
4. **Repeat and report variance** (several runs; show min, median, spread). One run is an anecdote.
5. **Open-loop load** where possible: generate load at a fixed rate independent of responses, otherwise slow responses hide themselves (coordinated omission). Record latency from the intended send time.
6. **Percentiles and histograms**, plus error rates and saturation of the generator itself (the load tool must not be the bottleneck).
7. **Change one variable at a time.**
8. **Observe the system, not only the client**: CPU, memory, garbage/allocations, connection pools, queue depth, consumer lag, storage I/O.
9. **Isolate noise**: no other load on the machine; document the environment.

## Reporting

Each run is recorded in `docs/benchmarks/<date>-<spec id>-<name>.md` with:

- What was tested (version or commit), why, and the spec/budget reference.
- Environment, configuration, dataset, workload model, tool and its settings.
- Results: percentiles, throughput, errors, resource use, with raw data or a link.
- Comparison with the baseline and the budget: pass, fail, or accepted miss (the user decides).
- Findings and next actions.

Do not extrapolate beyond what was measured. State limits ("measured on a single small instance with synthetic data").

## Regression control

- Keep a small, stable micro-benchmark set for hot paths and run it in CI as a trend signal.
- Fail the build only on thresholds you have validated as stable on your CI hardware; otherwise record and review trends.
- Re-run the budgeted benchmark before each spec is marked Implemented and after changes to the hot path.

## Cost awareness

Record the cost of meeting the budget (instances, storage, throughput units). A budget that is expensive to meet is a design signal.

## Review checklist

- [ ] Workload model and percentile targets in the spec?
- [ ] Baseline recorded; environment and data described?
- [ ] Warm-up excluded; several runs; variance shown?
- [ ] Open-loop load, and the generator not saturated?
- [ ] Result recorded in `docs/benchmarks/` and compared with the budget?
- [ ] No claim beyond what was measured?
