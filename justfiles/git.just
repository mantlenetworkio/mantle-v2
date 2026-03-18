import 'default.just'

# Set default values for git info
GITCOMMIT := env('GITCOMMIT', `git rev-parse HEAD 2> /dev/null || true`)
GITDATE := env('GITDATE', `git show -s --format='%ct' 2> /dev/null|| true`)

_ALL_TAGS := shell("git tag --sort=creatordate --points-at $1 2> /dev/null || true", GITCOMMIT)

_VERSION_TAGS := shell("echo $1 | tr ' ' '\n' | grep '^v' | sort -V", _ALL_TAGS)

_PREFERRED_TAG := shell("echo $1 | tr ' ' '\n' | grep -v -- '-rc' | tail -n 1", _VERSION_TAGS)

_LAST_TAG := shell("echo $1 | tr ' ' '\n' | tail -n 1", _VERSION_TAGS)

# Find version tag, prioritizing non-rc release tags
VERSION := if _PREFERRED_TAG != "" {
    _PREFERRED_TAG
} else if _LAST_TAG != "" {
    _LAST_TAG
} else {
    "untagged"
}

VERSION_META := ""
