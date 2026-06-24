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

# Detect main branch name (main or master)
MAIN_BRANCH="main"
if git branch -r | grep -q "$REMOTE/master"; then
    MAIN_BRANCH="master"
elif git branch | grep -q "\bmaster\b"; then
    MAIN_BRANCH="master"
fi
echo "Main branch detected: $MAIN_BRANCH"

echo "Fetching all remotes and tags..."
git fetch --all --tags || echo "Warning: Fetch failed, continuing with local state."

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
    if [ "$BRANCH" == "main" ] || [ "$BRANCH" == "master" ] || [ "$BRANCH" == "$INITIAL_BRANCH" ]; then
        continue
    fi

    echo "--- Reconciling Branch: $BRANCH ---"

    # Checkout the feature branch
    if ! git checkout "$BRANCH"; then
        echo "Failed to checkout $BRANCH, skipping."
        continue
    fi

    # Determine the merge source (prefer remote, fallback to local)
    MERGE_SOURCE="$REMOTE/$MAIN_BRANCH"
    if ! git rev-parse --verify "$MERGE_SOURCE" >/dev/null 2>&1; then
        if git rev-parse --verify "$MAIN_BRANCH" >/dev/null 2>&1; then
            MERGE_SOURCE="$MAIN_BRANCH"
        else
            echo "Warning: Could not find merge source for $MAIN_BRANCH."
            MERGE_SOURCE=""
        fi
    fi

    # Reverse Merge: Merging main into feature branch
    if [ -n "$MERGE_SOURCE" ]; then
        echo "Reverse Merge: Merging $MERGE_SOURCE into $BRANCH..."
        if git merge "$MERGE_SOURCE" --no-edit; then
            echo "Successfully caught up $BRANCH with $MERGE_SOURCE."
        else
            echo "CONFLICT or error detected on $BRANCH. Aborting reverse merge."
            git merge --abort || true
        fi

        # Forward Merge
        echo "Forward Merge: Merging $BRANCH back into $MAIN_BRANCH..."
        if git checkout "$MAIN_BRANCH"; then
            if git merge "$BRANCH" --no-edit; then
                echo "Successfully merged $BRANCH into $MAIN_BRANCH."
            else
                echo "CONFLICT detected during forward merge of $BRANCH. Aborting."
                git merge --abort || true
            fi
        else
            echo "Failed to checkout $MAIN_BRANCH for forward merge. Skipping."
        fi
    fi
done

# Return to initial branch and ensure it is caught up
git checkout "$INITIAL_BRANCH"

CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
echo "Finalizing sync for $CURRENT_BRANCH..."
if git merge "$MAIN_BRANCH" --no-edit; then
   echo "Successfully merged $MAIN_BRANCH into $CURRENT_BRANCH."
else
    echo "Final sync merge failed. Manual intervention required."
    git merge --abort || true
fi


echo "Step 3: Workspace Cleanup, Documentation, & Push"

# Governance: Source version
VERSION=$(cat VERSION.md)
echo "Current Version: $VERSION"

# Batch Script Validation: Ensure execution scripts use the correct MAIN_BRANCH
for script in build.sh start.sh; do
    if [ -f "$script" ]; then
        echo "Validating $script..."
        # Targeted replacement using word boundaries where possible to avoid collateral damage
        sed -i "s/\bmain\b/$MAIN_BRANCH/g" "$script"
    fi
done

# Synchronize version across changelog and status
if [ -f "CHANGELOG.md" ]; then
    sed -i "s/v1.0.0-alpha.[0-9]*/$VERSION/g" CHANGELOG.md
fi

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
