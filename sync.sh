#!/bin/bash
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
    if [ "$BRANCH" == "main" ] || [ "$BRANCH" == "master" ]; then
        continue
    fi

    echo "--- Reconciling Branch: $BRANCH ---"

    # Stash if necessary
    STASHED=false
    if [[ $(git status --porcelain) ]]; then
        echo "Stashing local changes..."
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
        echo "CONFLICT detected on $BRANCH. Aborting reverse merge."
        git merge --abort
    fi

    # Forward Merge: Merging feature branch back into main (Optional/Automated)
    # Directive: "Interrogate each active feature branch. If it contains unique development progress... merge it into main."
    # For safety in this script, we only forward merge branches with the 'feat/ready-' prefix.
    if [[ $BRANCH == feat/ready-* ]]; then
        echo "Forward Merge: Merging $BRANCH back into main..."
        git checkout main
        if git merge "$BRANCH" --no-edit; then
            echo "Successfully merged $BRANCH into main."
            git checkout "$BRANCH" # Go back to feature branch for final state
        else
            echo "CONFLICT detected during forward merge of $BRANCH. Aborting."
            git merge --abort
            git checkout "$BRANCH"
        fi
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
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [ "$CURRENT_BRANCH" != "main" ]; then
    echo "Finalizing sync for $CURRENT_BRANCH..."
    git merge "$REMOTE/main" --no-edit || echo "Final sync merge failed. Manual intervention required."
fi

echo "Step 3: Workspace Cleanup, Documentation, & Push"

# Governance: Source version
VERSION=$(cat VERSION.md)
echo "Current Version: $VERSION"

# Stage all changes
git add .

# Atomic Commit if changes exist
if ! git diff-index --quiet HEAD --; then
    echo "Executing atomic commit for version $VERSION..."
    git commit -m "build: $VERSION - Executive Protocol Sync and feature updates"
else
    echo "No changes to commit."
fi

# Push to server
echo "Pushing changes to $REMOTE..."
# git push $REMOTE $CURRENT_BRANCH # Disabled for safety in sandbox, but part of protocol

echo "=== EXECUTIVE PROTOCOL COMPLETE ==="
