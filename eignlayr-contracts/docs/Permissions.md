---
title: Permissioning Contracts
description: Overview / Descriptions for the contracts used for permissioning within EigenLayer
---

# Permissioning Contracts

## RepositoryAccess

The RepositoryAccess contract is an *abstract* (i.e. not possible to be deployed by itself) contract designed to be used for access control. It defines an immutable `repository`, which is the <a href=https://hackmd.io/@layr/repository>Repository</a> for a single middleware, and otherwise just defines modifiers and internal getter functions. It is intended to be inerhited from by other middleware contracts, reducing copy-pasting of these modifier and getter function definitions.

###### tags: `docs`