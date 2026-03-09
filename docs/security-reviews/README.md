# Mantle Audit Reports

This document provides a summary of the audit reports conducted for Mantle's major versions. Each section includes the audit time, the reviewer, the link to the report, and the audited commit.

---

## Optimism Audit Report

These are upstream OP Stack audit reports inherited from the Optimism codebase. The full set of upstream reports is available in the [optimism/](./optimism/) folder.

| Audit Time | Reviewer                   | Report Link                                                                                                                          | Audited Commit                             |
|------------|----------------------------|--------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------|
| 2020-10    | Trail of Bits              | [2020_10-Rollup-TrailOfBits](./optimism/2020_10-Rollup-TrailOfBits.pdf)                                                              |                                            |
| 2020-11    | Dapphub                    | [2020_11-Dapphub-ECDSA_Wallet](./optimism/2020_11-Dapphub-ECDSA_Wallet.pdf)                                                         |                                            |
| 2021-03    | OpenZeppelin               | [2021_03-OVM_and_Rollup-OpenZeppelin](./optimism/2021_03-OVM_and_Rollup-OpenZeppelin.pdf)                                            |                                            |
| 2021-03    | ConsenSys Diligence        | [2021_03-SafetyChecker-ConsenSysDiligence](./optimism/2021_03-SafetyChecker-ConsenSysDiligence.pdf)                                  |                                            |
| 2022-05    | Zeppelin                   | [2022_05-Bedrock_Contracts-Zeppelin](./optimism/2022_05-Bedrock_Contracts-Zeppelin.pdf)                                              |                                            |
| 2022-05    | Trail of Bits              | [2022_05-OpNode-TrailOfBits](./optimism/2022_05-OpNode-TrailOfBits.pdf)                                                              |                                            |
| 2022-08    | Sigma Prime                | [2022_08-Bedrock_GoLang-SigmaPrime](./optimism/2022_08-Bedrock_GoLang-SigmaPrime.pdf)                                                |                                            |
| 2022-09    | Zeppelin                   | [2022_09-Bedrock_and_Periphery-Zeppelin](./optimism/2022_09-Bedrock_and_Periphery-Zeppelin.pdf)                                      | `93d3bd411a8ae75702539ac9c5fe00bad21d4104` |
| 2022-10    | Spearbit                   | [2022_10-Drippie-Spearbit](./optimism/2022_10-Drippie-Spearbit.pdf)                                                                  | `2a7be367634f147736f960eb2f38a77291cdfcad` |
| 2022-11    | Trail of Bits              | [2022_11-Invariant_Testing-TrailOfBits](./optimism/2022_11-Invariant_Testing-TrailOfBits.pdf)                                        | `b31d35b67755479645dd150e7cc8c6710f0b4a56` |
| 2022-12    | Runtime Verification       | [2022_12-DepositTransaction-RuntimeVerification](./optimism/2022_12-DepositTransaction-RuntimeVerification.pdf)                      |                                            |
| 2023-01    | Trail of Bits              | [2023_01-Bedrock_Updates-TrailOfBits](./optimism/2023_01-Bedrock_Updates-TrailOfBits.pdf)                                            | `ee96ff8585699b054c95c6ff4a2411ee9fedcc87` |
| 2023-12    | Trust                      | [2023_12_SuperchainConfigUpgrade_Trust](./optimism/2023_12_SuperchainConfigUpgrade_Trust.pdf)                                        | `d1651bb22645ebd41ac4bb2ab4786f9a56fc1003` |
| 2024-02    | Cantina                    | [2024_02-MCP_L1-Cantina](./optimism/2024_02-MCP_L1-Cantina.pdf)                                                                      | `e6ef3a900c42c8722e72c2e2314027f85d12ced5` |
| 2024-05    | Sherlock                   | [2024_05-FaultProofs-Sherlock](./optimism/2024_05-FaultProofs-Sherlock.pdf)                                                            |                                            |
| 2024-05    | Cantina                    | [2024_05_SafeLivenessExtensions-Cantina](./optimism/2024_05_SafeLivenessExtensions-Cantina.pdf)                                      |                                            |
| 2024-08    | Cantina                    | [2024_08_Fault-Proofs-MIPS_Cantina](./optimism/2024_08_Fault-Proofs-MIPS_Cantina.pdf)                                                | `71b93116738ee98c9f8713b1a5dfe626ce06c1b2` |
| 2024-08    | Spearbit                   | [2024_08_Fault-Proofs-No-MIPS_Spearbit](./optimism/2024_08_Fault-Proofs-No-MIPS_Spearbit.pdf)                                        | `1f7081798ce2d49b8643514663d10681cb853a3d` |
| 2024-10    | 3Doc Security              | [2024_10-Cannon-FGETFD-3DocSecurity](./optimism/2024_10-Cannon-FGETFD-3DocSecurity.md)                                                | `52d0e60c16498ad4efec8798e3fc1b36b13f46a2` |
| 2024-12    | MiloTruck (independent)    | [2024_12-DPM-MiloTruck](./optimism/2024_12-DPM-MiloTruck.pdf)                                                                        | `2f17e6b67c61de5d8073d556272796d201bc740b` |
| 2024-12    | Radiant Labs               | [2024_12-DPM-RadiantLabs](./optimism/2024_12-DPM-RadiantLabs.pdf)                                                                    | `2f17e6b67c61de5d8073d556272796d201bc740b` |
| 2025-01    | Offbeat Labs               | [2025_01-IRI-OffbeatLabs](./optimism/2025_01-IRI-OffbeatLabs.pdf)                                                                    | `984bae9146398a2997ec13757bfe2438ca8f92eb` |
| 2025-01    | Spearbit                   | [2025_01-MT-Cannon-Spearbit](./optimism/2025_01-MT-Cannon-Spearbit.pdf)                                                              | `cc2715c3d6ebef374451b598f48980ad817e0a0e` |
| 2025-01    | Coinbase Protocol Security | [2025_01-MT-Cannon-Base](./optimism/2025_01-MT-Cannon-Base.pdf)                                                                      | `b8c011f18c79d735e01168345fc1c6f02fac584f` |
| 2025-02    | Spearbit                   | [2025_02-Upgrade13-Spearbit](./optimism/2025_02-Upgrade13-Spearbit.pdf)                                                              | `7d6d15437b7580b022f4c8c1ea9c0cd8d2e587e1` |
| 2025-03    | Spearbit                   | [2025_03-Interop-Contracts-Spearbit](./optimism/2025_03-Interop-Contracts-Spearbit.pdf)                                              | `6c80f23ab3074b5c66ff06e390ae2448bd4d2240` |
| 2025-03    | Wonderland                 | [2025_03-Interop-Portal-Wonderland](./optimism/2025_03-Interop-Portal-Wonderland.pdf)                                                | `9df1fc15d0bf0dc9464db249ce06424607d5f399` |
| 2025-04    | Cantina (contest)          | [2025_04-Interop-Portal-Cantina](./optimism/2025_04-Interop-Portal-Cantina.pdf)                                                      | `e4b921c9dbf8cd3a8db20ef4f15e0e2aa495fcc3` |
| 2025-04    | Aleph_v (independent)      | [2025_04-op-program-blob-handling-aleph_v](./optimism/2025_04-op-program-blob-handling-aleph_v.pdf)                                  | `08d81d98237a3077fbc13fcd4b70f2e8d2e14115` |
| 2025-05    | Coinbase Protocol Security | [2025_05-Cannon-Go-Updates-Coinbase](./optimism/2025_05-Cannon-Go-Updates-Coinbase.pdf)                                              | `4c68444bc9b130e892b52cacf67b31f0424fb6d0` |
| 2025-05    | Spearbit                   | [2025_05-Interop-Portal-Spearbit](./optimism/2025_05-Interop-Portal-Spearbit.pdf)                                                    | `7cd84fed9554193c2dcd683e1ff2d0e2605448f6` |
| 2025-05    | Spearbit                   | [2025_05-Upgrade16-Spearbit](./optimism/2025_05-Upgrade16-Spearbit.pdf)                                                              | `54c19f6acb7a6d3505f884bae601733d3d54a3a6` |
| 2025-06    | Radiant Labs               | [2025_06-Cannon-3DOC](./optimism/2025_06-Cannon-3DOC.pdf)                                                                            | `689111fca9a10e6670ba0b5c7f1a549a212c855b` |
| 2025-06    | Spearbit                   | [2025_06-Spearbit-Cannon-fix-review](./optimism/2025_06-Spearbit-Cannon-fix-review.pdf)                                              | `ffe3d5fed05cabf46a67ea00627a0959c0caa0b5` |
| 2025-07    | Spearbit                   | [2025_07-VerifyOPCM-Spearbit](./optimism/2025_07-VerifyOPCM-Spearbit.pdf)                                                            | `731280c6fc0ad184d252e0fb1d0ad12b5f59fd60` |
| 2025-09    | Spearbit                   | [2025_09-U16a-Spearbit](./optimism/2025_09-U16a-Spearbit.pdf)                                                                        | `475801690f7a451469ee4da87b5fe3c54c92f372` |
| 2025-10    | Spearbit                   | [2025_10-U17-Spearbit](./optimism/2025_10-U17-Spearbit.pdf)                                                                          | `aeed7033f7f739d8ecd4bd70a42ff09013bbc91e` |
| 2025-11    | Spearbit                   | [2025_11-SaferSafes-Spearbit](./optimism/2025_11-SaferSafes-Spearbit.pdf)                                                            | `cb54822c5e18925498f48d8677b71992bf402631` |
| 2025-11    | Spearbit                   | [2025_11-Custom-Gas-Token-Spearbit](./optimism/2025_11-Custom-Gas-Token-Spearbit.pdf)                                                | `1f888ede1940fce20f71db89fc13039fdd96757e` |

---

## Mantle Tectonic Audit Report

| Audit Time | Reviewer     | Report Link                                                                                                                                                                   | Audited Commit |
|------------|--------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------------|
| 2024-03    | OpenZeppelin | [Mantle Node, Batcher, Proposer, and Tooling Incremental Final Audit Report (March 2024)](./mantle-tectonic/OpenZeppelin/Mantle%20Node%2C%20Batcher%2C%20Proposer%2C%20and%20Tooling%20Incremental%20Final%20Audit%20Report%20%28March%202024%29.pdf) |                |
| 2024-03    | OpenZeppelin | [Mantle Op-Geth Audit Final Report (March 2024)](./mantle-tectonic/OpenZeppelin/Mantle%20Op-Geth%20Audit%20Final%20Report%20%28March%202024%29.pdf)                            |                |
| 2024-03    | OpenZeppelin | [Mantle V2 Solidity Contracts Audit Report (March 2024)](./mantle-tectonic/OpenZeppelin/Mantle%20V2%20Solidity%20Contracts%20Audit%20Report%20%28March%202024%29.pdf)          |                |
| 2024-03    | Secure3      | [Mantle_V2_ Secure3 Audit Report](./mantle-tectonic/Secure3/Mantle_V2_%20Secure3%20Audit%20Report.pdf)                                                                       |                |
| 2024-03    | Secure3      | [Mantle_V2_Public_Secure3_Audit_Report](./mantle-tectonic/Secure3/Mantle_V2_Public_Secure3_Audit_Report.pdf)                                                                 |                |
| 2024-04    | Sigma Prime  | [Sigma_Prime_Mantle_L2_Rollup_V2_Security_Assessment_Report_v2_0](./mantle-tectonic/SigmaPrime/Sigma_Prime_Mantle_L2_Rollup_V2_Security_Assessment_Report_v2_0.pdf)           |                |

---

## Mantle Everest Audit Report

| Audit Time | Reviewer     | Report Link                                                                                                                                                                   | Audited Commit |
|------------|--------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------------|
| 2025-02    | OpenZeppelin | [Mantle Op-geth & Op-stack v1.1.1 Diff Audit Report (Feb 2025)](./mantle-everest/Mantle%20Op-geth%20%26%20Op-stack%20v1.1.1%20Diff%20Audit%20Report%20%28Feb%202025%29.pdf)   |                |
| 2025-03    | Sigma Prime  | [Sigma_Prime_Mantle_EigenDA_Integration_Security_Assessment_Report](./mantle-everest/Sigma_Prime_Mantle_EigenDA_Integration_Security_Assessment_Report.pdf)                    |                |

---

## Mantle Euboea Audit Report

| Audit Time | Reviewer     | Report Link                                                                                                                               | Audited Commit |
|------------|--------------|-------------------------------------------------------------------------------------------------------------------------------------------|----------------|
| 2025-03    | Zenith       | [Mantle - Zenith Audit Report](./mantle-euboea/Mantle%20-%20Zenith%20Audit%20Report.pdf)                                                 |                |
| 2025-04    | OpenZeppelin | [Mantle Network Pre-Confirmation Transactions Audit](./mantle-euboea/Mantle%20Network%20Pre-Confirmation%20Transactions%20Audit.pdf)     |                |

---

## Mantle Skadi Audit Report

| Audit Time | Reviewer | Report Link                                                                                                            | Audited Commit |
|------------|----------|------------------------------------------------------------------------------------------------------------------------|----------------|
| 2025-12    | Sherlock | [Sherlock - Mantle Collaborative Audit Report](./mantle-skadi/Sherlock%20-%20Mantle%20Collaborative%20Audit%20Report.pdf) |                |

---

## Notes

- All audit reports are conducted by reputable third-party auditors to ensure the security and robustness of the Mantle network.
- Detailed findings and recommendations can be found in the linked audit reports.

For any questions or further details, feel free to contact the Mantle team.
