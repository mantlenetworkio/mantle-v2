# Mantle Rust Subtree Patches

This file is the authoritative registry of every Mantle modification stacked on top of
the upstream optimism `rust/` subtree in `mantle-v2/rust/`. It is the primary reference
when synchronizing future upstream changes via `git subtree pull`.

**Whenever Mantle changes are added, modified, or removed, update this file.**

## 1. Current baseline

| Item | Value |
|---|---|
| Upstream tracking point | optimism `develop` @ `1ad181f05` (2026-05-11) |
| Bridge tag | `rust-develop-20260511` |
| Bridge repo | https://github.com/mantle-xyz/optimism-rust-bridge |
| `git subtree add` commit | `ba2cc4514` ("Add 'rust/' from commit '1ad181f05...'") |
| Rust toolchain | 1.94 (see `rust/rust-toolchain.toml`) |

## 2. Architecture decisions

### 2.1 revm sourced from mantle-xyz/revm @ mantle-elysium

The `[patch.crates-io]` section in `rust/Cargo.toml` redirects every revm-family crate
to the `mantle-elysium` branch of `mantle-xyz/revm`:

```
revm, revm-bytecode, revm-context, revm-context-interface, revm-database,
revm-database-interface, revm-handler, revm-inspector, revm-interpreter,
revm-precompile, revm-primitives, revm-state, op-revm
```

`mantle-elysium` ships revm v38 plus Mantle protocol changes (ARSIA/JOVIAN hardforks,
BVM_ETH, token_ratio, DA footprint, Arsia fee validation). This avoids re-implementing
those changes inside `rust/op-revm/`; that subtree is excluded from the workspace.

### 2.2 Version skew with mantle-elysium

| Dimension | What develop expects | What mantle-elysium provides | Reconciliation |
|---|---|---|---|
| revm major version | v38 | v38 ✅ | — |
| op-revm major version | v20 | v19 ⚠️ | Adapt Mantle consumers to v19 API |
| `OpSpecId` variants | Includes `KARST` | No `KARST`; includes `OSAKA` + `ARSIA` | Replace KARST references with OSAKA fallbacks or comment them out |

## 3. Mantle changes registry

Every change carries a `[MANTLE]` source comment. Discover all sites with:

```bash
grep -rn "\[MANTLE\]" rust/ --include="*.rs" --include="*.toml"
```

### 3.1 Cargo workspace configuration

| File | Change |
|---|---|
| `Cargo.toml` | `[patch.crates-io]` redirects all 13 revm-family crates to `mantle-xyz/revm@mantle-elysium`. |
| `Cargo.toml` | Workspace `members` drops `"op-revm/"`; `exclude = ["op-revm"]` keeps the orphan subtree out of the build. |

### 3.2 op-alloy — TxDeposit gains BVM_ETH fields

Corresponds to mantle-xyz/op-alloy commits `5f0b879`, `5330f5a`, `79d78a4`, `da4e219`, `6637567`.

| File | Change |
|---|---|
| `op-alloy/crates/consensus/src/transaction/deposit.rs` | Add `eth_value: u128` and `eth_tx_value: Option<u128>` fields with their serde attrs. |
| same | Update `rlp_decode_fields` / `rlp_encode_fields` / `rlp_encoded_fields_length` / `size()` to include the new fields. |
| same | Switch `rlp_decode` to a `split_at(header.payload_length)` strict-boundary form (port of commit 6637567). |
| same | Add the `decode_optional_u128_from_rlp` helper for the trailing optional u128. |
| same (tests) | Add 0/None for both fields in 8 in-file `TxDeposit { ... }` literals; add `_` ignores in 1 alloy-compat destructure. |
| `op-alloy/crates/consensus/src/transaction/envelope.rs` | Add 0/None in 2 test `TxDeposit` literals. |
| `op-alloy/crates/consensus/src/reth_codec.rs` | `From<CompactTxDeposit>` fills 0/None for the new fields. **TODO**: `CompactTxDeposit` itself does not carry the new fields, so reth Compact round-trips drop BVM_ETH data. |
| `op-alloy/crates/consensus/src/transaction/deposit.rs` (`bincode_compat`) | Same situation as `reth_codec`. **TODO**. |
| `op-alloy/crates/consensus/src/nuts/mod.rs` | NutBundle upgrade-tx literal fills 0/None. |
| `op-alloy/crates/rpc-types/src/transaction/request.rs` | OpTransactionRequest destructure adds `_` ignores for the new fields. |

### 3.3 kona — adapts to TxDeposit's BVM_ETH fields

| File | Change |
|---|---|
| `kona/crates/protocol/protocol/src/info/variant.rs` | L1Info deposit literal fills 0/None. |
| `kona/crates/protocol/hardforks/src/{ecotone,fjord,interop,isthmus,jovian}.rs` | 31 OP hardfork upgrade-tx literals filled 0/None via the script in §6. |
| `kona/bin/client/src/fpvm_evm/tx.rs` | `FromTxWithEncoded<TxDeposit>` reads `tx.eth_value` / `tx.eth_tx_value` into `DepositTransactionParts` using the 0→None convention. |

### 3.4 alloy-op-evm — Mantle protocol changes

Corresponds to mantle-xyz/evm commits `5f383c5`, `9fe2c85`, `760129f`.

| File | Change |
|---|---|
| `alloy-op-evm/src/tx.rs` | `OpTxTr` impl adds `eth_value()` / `eth_tx_value()` methods (delegated to the wrapped `OpTransaction`). |
| same | `FromTxWithEncoded<TxDeposit>` reads the new BVM_ETH fields into `DepositTransactionParts` (0→None). |
| `alloy-op-evm/src/env.rs` | Comments out the `is_karst_active_at_timestamp => KARST` hook (no KARST on mantle-elysium). |
| same (tests) | Comments out the `OpSpecId::KARST` `test_case`. |
| `alloy-op-evm/src/block/mod.rs` | `deposit_receipt_version = None` (corresponds to commit 760129f). |
| same | Comments out the `ensure_create2_deployer(...)` call and its `use canyon::ensure_create2_deployer;` import. |
| same | Drops the `spec_id` argument from `operator_fee_charge` in two call sites to match mantle-elysium's older 2-arg signature. |
| `alloy-op-evm/src/block/canyon.rs` | Adds `#![allow(dead_code)]` because the function is now unreachable. |

### 3.5 kona-client fpvm — adapts to mantle-elysium op-revm v19

| File | Change |
|---|---|
| `kona/bin/client/src/fpvm_evm/precompiles/provider.rs` | Drop the `karst` import. Collapse the `KARST` match arms into `JOVIAN \| OSAKA \| ARSIA \| INTEROP` and route them to `jovian()` / `accelerated_jovian` so the match remains exhaustive. |

### 3.6 op-reth/rpc — handles the Mantle-specific OpTransactionError variants

| File | Change |
|---|---|
| `op-reth/crates/rpc/src/error.rs` | `TryFrom<OpTxError>` adds an arm for `BvmEth(_) \| TxL1CostOutOfRange`. Placeholder maps to `MissingEnvelopedTx`. **TODO**: extend `OpInvalidTransactionError` with proper variants and dedicated RPC error codes. |

### 3.7 op-core — vendored data (outside the rust/ subtree)

| File | Change |
|---|---|
| `<mantle-v2 root>/op-core/nuts/bundles/karst_nut_bundle.json` | Copied verbatim from optimism develop so `kona-hardforks/build.rs` can find it via its ancestor walk. |

**Note**: this file is *outside* `rust/`, so `git subtree pull` will not sync it. If a
future upstream build.rs looks for additional bundle files, add the corresponding JSONs
under `op-core/nuts/bundles/` manually.

## 4. Sync workflow

### 4.1 Pre-sync dry-run (optional but recommended)

```bash
cd mantle-v2
git checkout -b sync-dryrun-$(date +%Y%m%d)
git subtree pull --prefix=rust/ \
  https://github.com/mantle-xyz/optimism-rust-bridge.git main \
  --no-commit
git diff --name-only --diff-filter=U   # list conflicting files
git merge --abort                       # bail out — this was just a probe
```

### 4.2 Sync run

```bash
git checkout -b rust/sync-$(date +%Y%m)
git subtree pull --prefix=rust/ \
  https://github.com/mantle-xyz/optimism-rust-bridge.git main \
  -m "rust: subtree pull from bridge ($(date +%Y-%m))"

# Resolve each conflict — grep for [MANTLE] markers in conflicted files to make
# sure no Mantle change is dropped.
git diff --name-only --diff-filter=U | xargs grep -l "\[MANTLE\]"
```

### 4.3 Verification

```bash
TOOLCHAIN=$(grep channel rust/rust-toolchain.toml | cut -d'"' -f2)

# 1. Workspace-wide type check.
RUSTUP_TOOLCHAIN=$TOOLCHAIN cargo check --workspace \
  --manifest-path rust/Cargo.toml

# 2. Full build (including tests) for the Mantle-touched crates.
RUSTUP_TOOLCHAIN=$TOOLCHAIN cargo build --tests \
  --manifest-path rust/Cargo.toml \
  -p op-alloy -p op-alloy-consensus -p op-alloy-network \
  -p op-alloy-provider -p op-alloy-rpc-jsonrpsee \
  -p op-alloy-rpc-types -p op-alloy-rpc-types-engine \
  -p alloy-op-evm

# 3. Audit the [MANTLE] markers against this file's §3 registry.
grep -rn "\[MANTLE\]" rust/ --include="*.rs" --include="*.toml" | wc -l
```

### 4.4 Land the sync

```bash
git push -u origin rust/sync-$(date +%Y%m)
# Open a PR and merge into the upgrade branch once review passes.
```

## 5. Conflict hot spots and time bombs

### 5.1 High-churn hot spots (likely to conflict every sync)

| Location | Why it churns | Post-sync checks |
|---|---|---|
| `op-alloy/.../deposit.rs` | TxDeposit is a frequently edited struct. | Verify BVM_ETH field positions and RLP order are preserved. |
| `alloy-op-evm/src/block/mod.rs` | The block executor is a high-churn area upstream. | Re-verify `deposit_receipt_version = None`, the commented-out `ensure_create2_deployer`, and the 2-arg `operator_fee_charge` call sites. |
| `kona/bin/client/.../provider.rs` (the `OpSpecId` match arms) | Any new upstream hardfork variant breaks exhaustiveness. | If `cargo check` flags non-exhaustive matches, add the new variant to the appropriate arm. |
| `kona/.../hardforks/src/*.rs` (TxDeposit literals) | Each new hardfork adds new upgrade-tx literals missing BVM_ETH fields. | Run the script in §6 on the newly added files. |

### 5.2 Time bombs (need active monitoring)

| Risk | Trigger | Mitigation |
|---|---|---|
| **revm major-version bump** | Upstream raises revm to v39+. | Coordinate with mantle-xyz/revm to catch up before syncing, or defer the sync. |
| **op-revm v19 → v20+ drift widens** | mantle-elysium does not track upstream op-revm. | Let `cargo check` surface the differences and adapt site-by-site (potentially extending the `OpTxTr` impl, adjusting signatures, etc.). |
| **New OpSpecId variant** | Upstream introduces a new hardfork. | `cargo check` will flag the non-exhaustive match; extend the relevant arm. |
| **mantle-xyz/revm becomes unreachable** | Network, credentials, or repo permission issues. | Temporarily vendor a copy of mantle-elysium under `mantle-v2/` and switch the patch entries from `git = ...` to `path = ...`. |

## 6. Helper script — batch-patch new TxDeposit literals

Whenever upstream introduces a new hardfork upgrade-tx file (e.g. a new module under
`kona-hardforks/src/`), the new `TxDeposit { ... }` literals will not include the Mantle
BVM_ETH fields. Run this Python helper to add `eth_value: 0` and `eth_tx_value: None`
to every struct literal while leaving `impl ... for TxDeposit { ... }` blocks alone.

```python
#!/usr/bin/env python3
"""Inject eth_value/eth_tx_value into TxDeposit { ... } struct literals.
Skips `impl SomeTrait for TxDeposit { ... }` impl blocks (preceded by `for `).
Usage: python3 this.py file1.rs file2.rs ...
"""
import sys

for path in sys.argv[1:]:
    with open(path) as fh:
        content = fh.read()
    out, i, edits = [], 0, 0
    while True:
        idx = content.find('TxDeposit {', i)
        if idx == -1:
            out.append(content[i:])
            break
        # Skip impl blocks: look back over whitespace for the keyword `for`.
        k = idx - 1
        while k >= 0 and content[k] in ' \t\n':
            k -= 1
        if k >= 2 and content[k-2:k+1] == 'for' and (k - 3 < 0 or content[k-3] in ' \t\n'):
            out.append(content[i:idx + len('TxDeposit {')])
            i = idx + len('TxDeposit {')
            continue
        # Brace-track to the matching close.
        depth, j = 1, idx + len('TxDeposit {')
        while j < len(content) and depth > 0:
            depth += {'{': 1, '}': -1}.get(content[j], 0)
            j += 1
        block = content[idx:j]
        if 'eth_value' in block:
            out.append(content[i:j])
            i = j
            continue
        nl = content.rfind('\n', idx, j - 1)
        close_line_start = nl + 1
        close_indent = content[close_line_start:j-1]
        field_indent = close_indent + '    '
        out.append(content[i:close_line_start])
        out.append(f"{field_indent}eth_value: 0,\n")
        out.append(f"{field_indent}eth_tx_value: None,\n")
        out.append(content[close_line_start:j])
        i = j
        edits += 1
    with open(path, 'w') as fh:
        fh.write(''.join(out))
    print(f"{path}: {edits} edits")
```

Example:

```bash
python3 /tmp/fix.py rust/kona/crates/protocol/hardforks/src/new_fork.rs
```

**Caveat**: the script defaults the fields to `0` / `None`, which is correct for OP
upgrade transactions (no BVM_ETH semantics). If a new hardfork introduces literals that
*do* carry BVM_ETH values, patch them manually instead.

## 7. Maintaining this file

When you add, modify, or remove a Mantle change:

1. Add a `[MANTLE]` comment in the source explaining intent.
2. Register the change under the appropriate subsection of §3.
3. If the change is structural (new field, new method, signature change), evaluate
   whether §5.1 needs a new hot spot entry.
4. Reference this file in the commit message so future contributors can find their way back.
