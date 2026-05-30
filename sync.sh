#!/bin/bash
set -e

# THE EXECUTIVE PROTOCOL: REPOSITORY SYNCHRONIZATION & INTELLIGENT MERGE
# This script maintains repository health by reconciling all local feature branches with main.

echo "=== EXECUTIVE PROTOCOL: REPOSITORY SYNC & INTELLIGENT MERGE ==="

# 1. UPSTREAM TRACKING & SUBMODULE SANITIZATION
echo "Step 1: Upstream Tracking & Submodule Sanitization"

# Identify the upstream remote (prefer 'upstream' if it exists, otherwise 'origin')
REMOTE="origin"
if git remote | grep -q "^upstream$"; then
    REMOTE="upstream"
fi
echo "Using remote: $REMOTE"

echo "Fetching all remotes and tags..."
git fetch --all --tags

echo "Updating submodules recursively to their latest tracking commits..."
git submodule update --init --recursive

# Identify current branch to return to it later
INITIAL_BRANCH=$(git rev-parse --abbrev-ref HEAD)
echo "Initial branch: $INITIAL_BRANCH"

# 2. DUAL-DIRECTION INTELLIGENT MERGE ENGINE
echo "Step 2: Dual-Direction Intelligent Merge Engine"

# Get all local branches
ALL_LOCAL_BRANCHES=$(git branch --format='%(refname:short)')

for BRANCH in $ALL_LOCAL_BRANCHES; do
    if [ "$BRANCH" == "main" ]; then
        continue
    fi

    echo "--- Reconciling Branch: $BRANCH ---"

    # Stash if necessary
    STASHED=false
    if [[ $(git status --porcelain) ]]; then
        echo "Stashing local changes on $INITIAL_BRANCH..."
        git stash
        STASHED=true
    fi

    # Checkout the feature branch
    if ! git checkout "$BRANCH"; then
        echo "Failed to checkout $BRANCH, skipping."
        continue
    fi

    # Reverse Merge: Merging main into feature branch
    echo "Reverse Merge: Merging $REMOTE/main into $BRANCH..."
    if git merge "$REMOTE/main" --no-edit; then
        echo "Successfully caught up $BRANCH with $REMOTE/main."
    else
        echo "CONFLICT detected on $BRANCH. Aborting merge."
        git merge --abort
    fi

    # Return to initial branch
    git checkout "$INITIAL_BRANCH"

    if [ "$STASHED" = true ]; then
        echo "Restoring stashed changes..."
        git stash pop
        STASHED=false
    fi
done

# Ensure we are back on the initial branch and it is caught up
if [ "$INITIAL_BRANCH" != "main" ]; then
    echo "Finalizing sync for $INITIAL_BRANCH..."
    git merge "$REMOTE/main" --no-edit || echo "Final sync merge failed. Manual intervention required."
fi

echo "Step 3: Workspace Cleanup & Finalization"
echo "=== EXECUTIVE PROTOCOL COMPLETE ==="
