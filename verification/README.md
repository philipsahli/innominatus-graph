# Verification-First Development

## Philosophy

**Verification scripts are NOT unit tests.** They demonstrate that features work in practice by running actual code with real dependencies and capturing real outputs.

## When to Write Verification Scripts

- **Before implementing** a new feature (verification-first approach)
- When adding a feature that produces **artifacts** (exports, database records, API responses)
- When you need to **demonstrate** the feature works end-to-end
- For **integration testing** beyond unit tests

## Verification vs Unit Tests

### Unit Tests
- Test individual functions in isolation
- Use mocks for dependencies
- Fast execution (<1s for full suite)
- Run in CI/CD automatically

### Verification Scripts
- Test entire features end-to-end
- Use **real dependencies** (real database, real GraphViz)
- Save **real outputs** (SVG files, database records, JSON)
- Prove feature works in practice
- Run before declaring feature complete

## Template Usage

Use `template.go` as a starting point for new verification scripts:

```bash
cp verification/template.go verification/verify-my-feature.go
# Edit verify-my-feature.go with your verification logic
go run verification/verify-my-feature.go
```

## Running Verifications

```bash
# Run single verification
go run verification/verify-state-propagation.go

# Run all verifications (if you have multiple)
for f in verification/verify-*.go; do
    echo "Running $f..."
    go run "$f" || exit 1
done
```

## Verification Output

Each verification script should:

1. **Print clear results**: ✅ PASSED or ❌ FAILED
2. **Save artifacts**: Files in `docs/verification/`
3. **Output structured data**: JSON for AI evaluation
4. **Exit with code**: 0 for success, 1 for failure

## Example Artifacts

Saved to `docs/verification/`:
- `feature-name.json`: Verification results
- `graph-export.svg`: Generated visualizations
- `database-dump.json`: Database state
- `execution-log.txt`: Execution traces

## Best Practices

### ✅ Do This
- Test with real dependencies (database, GraphViz)
- Save all outputs for manual inspection
- Print descriptive error messages
- Use structured output (JSON)
- Exit with proper code (0 or 1)

### ❌ Don't Do This
- Don't mock everything (use real dependencies)
- Don't skip saving artifacts
- Don't print unstructured output
- Don't swallow errors

## Verification Workflow

```
1. Write verification script FIRST (before feature)
   ├─ Define expected inputs
   ├─ Define expected outputs
   └─ Define success criteria

2. Run verification (should FAIL)
   └─ Proves verification works

3. Implement feature
   └─ Guided by verification requirements

4. Run verification again (should PASS)
   ├─ Saves artifacts
   └─ Proves feature works

5. Commit both feature code AND verification script
```

## Example Directory Structure

```
verification/
├── README.md                          # This file
├── template.go                        # Template for new verifications
├── verify-state-propagation.go       # Example: State propagation
├── verify-graph-export.go             # Example: Graph export
└── verify-database-persistence.go    # Example: Database ops

docs/verification/                     # Artifacts
├── state-propagation.json
├── state-propagation.svg
├── graph-export-test.svg
└── database-persistence.json
```

## Integration with Claude Code

Claude Code agents (especially QA Engineer) should:
- Request verification scripts for all new features
- Run verification scripts before approving PRs
- Check that artifacts are saved correctly
- Verify structured output is present
