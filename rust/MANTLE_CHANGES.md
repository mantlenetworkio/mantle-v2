# Mantle Rust Subtree Patches

本文件记录 `mantle-v2/rust/` 在 optimism upstream 之上叠加的所有 Mantle 改动,
作为后续从 bridge 同步上游时的对照清单。**任何新增/修改 Mantle 改动都必须在此登记**。

## 1. 当前 baseline

| 项 | 值 |
|---|---|
| upstream 跟踪点 | optimism `develop` @ `1ad181f05` (2026-05-11) |
| bridge tag | `rust-develop-20260511` |
| bridge repo | https://github.com/mantle-xyz/optimism-rust-bridge |
| subtree add commit | `ba2cc4514` ("Add 'rust/' from commit '1ad181f05...'") |
| Rust toolchain | 1.94(见 `rust/rust-toolchain.toml`) |

## 2. 架构决策

### revm 来源:mantle-xyz/revm @ mantle-elysium

`rust/Cargo.toml` 的 `[patch.crates-io]` 段把以下 crate 全部重定向到
`mantle-xyz/revm` 的 `mantle-elysium` 分支:

```
revm, revm-bytecode, revm-context, revm-context-interface, revm-database,
revm-database-interface, revm-handler, revm-inspector, revm-interpreter,
revm-precompile, revm-primitives, revm-state, op-revm
```

mantle-elysium 提供 revm v38 + Mantle 协议改动(ARSIA/JOVIAN hardfork、
BVM_ETH、token_ratio、DA footprint、Arsia fee validation),避免在
`rust/op-revm/` 重做 — 后者已从 workspace.members 排除。

### 版本一致性现状

| 维度 | develop 期望 | mantle-elysium 提供 | 差异处理 |
|---|---|---|---|
| revm 主版本 | v38 | v38 ✅ | — |
| op-revm 主版本 | v20 | v19 ⚠️ | Mantle 改动适配 v19 API |
| OpSpecId variants | 含 `KARST` | 不含 `KARST`,含 `OSAKA + ARSIA` | KARST 引用全部改成 OSAKA 兜底或注释 |

## 3. Mantle 改动注册表

所有改动在源码中带 `[MANTLE]` 注释标记。盘点命令:

```bash
grep -rn "\[MANTLE\]" rust/ --include="*.rs" --include="*.toml"
```

### 3.1 Cargo workspace 配置

| 文件 | 改动 |
|---|---|
| `Cargo.toml` | `[patch.crates-io]` 13 个 revm 系列 crate 重定向到 mantle-elysium |
| `Cargo.toml` | workspace `members` 移除 `"op-revm/"`,加 `exclude = ["op-revm"]` |

### 3.2 op-alloy — TxDeposit 加 BVM_ETH 字段

对应 mantle-xyz/op-alloy commits: `5f0b879`, `5330f5a`, `79d78a4`, `da4e219`, `6637567`。

| 文件 | 改动 |
|---|---|
| `op-alloy/crates/consensus/src/transaction/deposit.rs` | 加 `eth_value: u128`、`eth_tx_value: Option<u128>` 字段 + serde attrs |
| 同上 | `rlp_decode_fields` / `rlp_encode_fields` / `rlp_encoded_fields_length` / `size()` 更新 |
| 同上 | `rlp_decode` 改用 `split_at(header.payload_length)` 严格边界(commit 6637567) |
| 同上 | 新增 `decode_optional_u128_from_rlp` helper |
| 同上 | 8 处 test 用 `TxDeposit { ... }` 字面量补字段;1 处 alloy-compat destructure 加 `_` ignore |
| `op-alloy/crates/consensus/src/transaction/envelope.rs` | 2 处 test 字面量补字段 |
| `op-alloy/crates/consensus/src/reth_codec.rs` | `From<CompactTxDeposit>` 补默认 0/None **(TODO: CompactTxDeposit struct 自身未带这两个字段,reth Compact 存储 round-trip 会丢 BVM_ETH 数据)** |
| `op-alloy/crates/consensus/src/transaction/deposit.rs` `bincode_compat` | 同 reth_codec 情况 **(TODO)** |
| `op-alloy/crates/consensus/src/nuts/mod.rs` | NutBundle upgrade tx 字面量补 0/None |
| `op-alloy/crates/rpc-types/src/transaction/request.rs` | OpTransactionRequest destructure 加 `_` ignore |

### 3.3 kona — 适配 TxDeposit BVM_ETH 字段

| 文件 | 改动 |
|---|---|
| `kona/crates/protocol/protocol/src/info/variant.rs` | L1Info deposit 字面量补 0/None |
| `kona/crates/protocol/hardforks/src/{ecotone,fjord,interop,isthmus,jovian}.rs` | 31 处 OP 升级交易 TxDeposit 字面量,**脚本批量**(见 §6) |
| `kona/bin/client/src/fpvm_evm/tx.rs` | `FromTxWithEncoded<TxDeposit>` 读 `tx.eth_value` / `tx.eth_tx_value` 填入 `DepositTransactionParts`(0→None 模式) |

### 3.4 alloy-op-evm — Mantle 协议改动

对应 mantle-xyz/evm commits: `5f383c5`, `9fe2c85`, `760129f`。

| 文件 | 改动 |
|---|---|
| `alloy-op-evm/src/tx.rs` | `OpTxTr` impl 加 `eth_value()` / `eth_tx_value()` 方法(delegate to wrapped `OpTransaction`) |
| 同上 | `FromTxWithEncoded<TxDeposit>` 读 BVM_ETH 字段填入 `DepositTransactionParts`(0→None 模式) |
| `alloy-op-evm/src/env.rs` | 注释掉 `is_karst_active_at_timestamp => KARST` hook(mantle-elysium 无 KARST variant) |
| 同上 (test) | 注释掉 `OpSpecId::KARST` test_case |
| `alloy-op-evm/src/block/mod.rs` | `deposit_receipt_version = None`(对应 760129f) |
| 同上 | 注释掉 `ensure_create2_deployer(...)` 调用 + `use canyon::ensure_create2_deployer;` import |
| 同上 | `operator_fee_charge` 调用从 3 参 → 2 参(适配 mantle-elysium 旧签名)x2 处 |
| `alloy-op-evm/src/block/canyon.rs` | 加 `#![allow(dead_code)]`(因为函数被禁用) |

### 3.5 kona-client fpvm — 适配 mantle-elysium op-revm v19

| 文件 | 改动 |
|---|---|
| `kona/bin/client/src/fpvm_evm/precompiles/provider.rs` | 移除 `karst` import;match `KARST` → `JOVIAN \| OSAKA \| ARSIA \| INTEROP` 使用 `jovian()` / `accelerated_jovian` 作兜底 |

### 3.6 op-reth/rpc — 适配 mantle-elysium 错误 enum 扩展

| 文件 | 改动 |
|---|---|
| `op-reth/crates/rpc/src/error.rs` | `TryFrom<OpTxError>` 加 `BvmEth(_) \| TxL1CostOutOfRange` arm,占位映射到 `MissingEnvelopedTx` **(TODO: 后续补正确的 RPC 错误码)** |

### 3.7 op-core — vendor 数据(不在 rust/ 子树内)

| 文件 | 改动 |
|---|---|
| `<mantle-v2 root>/op-core/nuts/bundles/karst_nut_bundle.json` | 从 optimism develop vendor 进来,满足 `kona-hardforks/build.rs` 编译时 ancestor 查找 |

**注:这个文件在 rust/ 之外**,所以 subtree pull 不会自动同步。如果未来上游 build.rs 找其他 bundle,要在 `op-core/nuts/bundles/` 下手动补对应 JSON。

## 4. Sync 工作流

### 4.1 Pre-sync(可选预演)

```bash
# 在临时分支试一遍,看冲突
cd mantle-v2
git checkout -b sync-dryrun-$(date +%Y%m%d)
git subtree pull --prefix=rust/ \
  https://github.com/mantle-xyz/optimism-rust-bridge.git main \
  --no-commit  # 或先 commit 看冲突列表

git diff --name-only --diff-filter=U  # 列出冲突文件
git merge --abort  # 不要,只是预演
```

### 4.2 Sync 执行

```bash
git checkout -b rust/sync-$(date +%Y%m)
git subtree pull --prefix=rust/ \
  https://github.com/mantle-xyz/optimism-rust-bridge.git main \
  -m "rust: subtree pull from bridge ($(date +%Y-%m))"

# 解决冲突 — 每个冲突文件 grep [MANTLE] 确认 Mantle 改动保留
git diff --name-only --diff-filter=U | xargs grep -l "\[MANTLE\]"
```

### 4.3 验证

```bash
TOOLCHAIN=$(grep channel rust/rust-toolchain.toml | cut -d'"' -f2)

# 1. 全 workspace 类型检查
RUSTUP_TOOLCHAIN=$TOOLCHAIN cargo check --workspace \
  --manifest-path rust/Cargo.toml

# 2. 重点 crate 完整编译(含 tests)
RUSTUP_TOOLCHAIN=$TOOLCHAIN cargo build --tests \
  --manifest-path rust/Cargo.toml \
  -p op-alloy -p op-alloy-consensus -p op-alloy-network \
  -p op-alloy-provider -p op-alloy-rpc-jsonrpsee \
  -p op-alloy-rpc-types -p op-alloy-rpc-types-engine \
  -p alloy-op-evm

# 3. 盘点 Mantle 标记数量(应与本文件 §3 列表数对得上)
grep -rn "\[MANTLE\]" rust/ --include="*.rs" --include="*.toml" | wc -l
```

### 4.4 提交 sync 结果

```bash
git push -u origin rust/sync-$(date +%Y%m)
# 开 PR,review 合入升级主分支
```

## 5. 高冲突风险与时间炸弹

### 5.1 高 churn 热点(每次 sync 大概率撞)

| 位置 | 风险 | sync 后必查 |
|---|---|---|
| `op-alloy/.../deposit.rs` | TxDeposit 是热点结构 | BVM_ETH 字段顺序、RLP 编码顺序 |
| `alloy-op-evm/src/block/mod.rs` | block executor 高 churn | `deposit_receipt_version = None`、`ensure_create2_deployer` 注释、`operator_fee_charge` 参数 |
| `kona/bin/client/.../provider.rs` 的 `OpSpecId` match | 上游加新 hardfork variant 直接打破穷举 | cargo check 报错则补对应 arm |
| `kona/.../hardforks/src/*.rs` 的 TxDeposit 字面量 | 上游加新 hardfork upgrade tx | 跑 §6 的批量脚本 |

### 5.2 时间炸弹(需要主动监控)

| 风险 | 触发 | 应对 |
|---|---|---|
| **revm 主版本升级** | optimism 升 revm v39+ | 协调 mantle-xyz/revm 同步上游,或推迟 sync |
| **op-revm v19 → v20+ 差异扩大** | mantle-elysium 不跟进 | 编译错误指引适配,可能要补 OpTxTr 方法 |
| **新 OpSpecId variant** | 上游加 hardfork | match 不穷举 → cargo 报错,按指引补 arm |
| **mantle-xyz/revm 仓库失联** | 网络/权限/账号问题 | 临时 vendor mantle-elysium 到本地 path |

## 6. Helper 脚本

### 批量给 TxDeposit 字面量补 BVM_ETH 字段

当上游新增 hardfork upgrade tx(`kona-hardforks` 加新 OP fork 文件)时,
新文件会有 `TxDeposit { ... }` 字面量没有 BVM_ETH 字段。用以下脚本批量补:

```python
#!/usr/bin/env python3
"""Inject eth_value/eth_tx_value into TxDeposit { ... } struct literals.
Skips `impl SomeTrait for TxDeposit { ... }` impl blocks (preceded by `for `).
"""
import os, sys

# 用法: python3 this.py file1.rs file2.rs ...
for path in sys.argv[1:]:
    with open(path) as fh:
        content = fh.read()
    out, i, edits = [], 0, 0
    while True:
        idx = content.find('TxDeposit {', i)
        if idx == -1:
            out.append(content[i:])
            break
        # 跳过 impl 块
        k = idx - 1
        while k >= 0 and content[k] in ' \t\n':
            k -= 1
        if k >= 2 and content[k-2:k+1] == 'for' and (k - 3 < 0 or content[k-3] in ' \t\n'):
            out.append(content[i:idx + len('TxDeposit {')])
            i = idx + len('TxDeposit {')
            continue
        # brace-track 找 matching }
        depth, j = 1, idx + len('TxDeposit {')
        while j < len(content) and depth > 0:
            depth += {'{': 1, '}': -1}.get(content[j], 0)
            j += 1
        block = content[idx:j]
        if 'eth_value' in block:
            out.append(content[i:j]); i = j; continue
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

例:

```bash
python3 /tmp/fix.py \
  rust/kona/crates/protocol/hardforks/src/new_fork.rs
```

注意:**这个脚本只补字段为 `0` / `None`**(适合升级交易这种**无** BVM_ETH 语义的场景);
若新 hardfork 引入了带 BVM_ETH 的字面量,要手动补具体值。

## 7. 维护本文件

每次新增/修改 Mantle 改动:

1. 源码里**必须**加 `[MANTLE]` 注释,说明意图
2. 在 §3 对应小节登记一行
3. 如果是结构性改动(新增字段/新增方法/改 signature),**必须**评估 §5.1 是否要补热点
4. commit message 引用本文件路径,方便回溯
