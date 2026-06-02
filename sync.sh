#!/bin/bash
# THE EXECUTIVE PROTOCOL: REPOSITORY SYNCHRONIZATION & INTELLIGENT MERGE
# This script maintains repository health by reconciling all local feature branches with main.

echo "=== EXECUTIVE PROTOCOL: REPOSITORY SYNC & INTELLIGENT MERGE ==="

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
fi
echo "Main branch detected: $MAIN_BRANCH"

# IDENTIFY INITIAL STATE
INITIAL_BRANCH=$(git rev-parse --abbrev-ref HEAD)
echo "Initial branch: $INITIAL_BRANCH"

# GLOBAL STASH: Protect uncommitted work before any operations
STASHED_WORK=false
if [[ $(git status --porcelain) ]]; then
    echo "Stashing uncommitted changes across the workspace..."
    git stash
    STASHED_WORK=true
fi

# 1. UPSTREAM TRACKING & SUBMODULE SANITIZATION
echo "Step 1: Upstream Tracking & Submodule Sanitization"
echo "Fetching all remotes and tags..."
git fetch --all --tags

# Ensure local main is updated safely
echo "Updating local $MAIN_BRANCH from $REMOTE..."
git checkout $MAIN_BRANCH
if ! git merge $REMOTE/$MAIN_BRANCH --ff-only; then
    echo "Warning: Fast-forward failed for $MAIN_BRANCH. Manual rebase may be required."
fi

echo "Updating submodules recursively to their latest tracking commits..."
git submodule update --init --recursive

# 2. DUAL-DIRECTION INTELLIGENT MERGE ENGINE
echo "Step 2: Dual-Direction Intelligent Merge Engine"

# Get all local branches
ALL_LOCAL_BRANCHES=$(git branch --format='%(refname:short)')

for BRANCH in $ALL_LOCAL_BRANCHES; do
    if [ "$BRANCH" == "main" ] || [ "$BRANCH" == "master" ]; then
        continue
    fi

    echo "--- Reconciling Branch: $BRANCH ---"

    # Checkout the feature branch
    if ! git checkout "$BRANCH"; then
        echo "Failed to checkout $BRANCH, skipping."
        continue
    fi

    # Reverse Merge: Merging main into feature branch
    echo "Reverse Merge: Merging $REMOTE/$MAIN_BRANCH into $BRANCH..."
    if git merge "$REMOTE/$MAIN_BRANCH" --no-edit; then
        echo "Successfully caught up $BRANCH with $REMOTE/$MAIN_BRANCH."
    else
        echo "CONFLICT detected on $BRANCH. Aborting merge."
        git merge --abort
    fi

    # Forward Merge: Merging feature branch back into main (Optional/Automated)
    # Directive: "Interrogate each active feature branch. If it contains unique development progress... merge it into main."
    # For safety in this script, we only forward merge branches with the 'feat/ready-' prefix.
    if [[ $BRANCH == feat/ready-* ]]; then
        echo "Forward Merge: Merging $BRANCH back into $MAIN_BRANCH..."
        git checkout $MAIN_BRANCH
        if git merge "$BRANCH" --no-edit; then
            echo "Successfully merged $BRANCH into $MAIN_BRANCH."
            git checkout "$BRANCH" # Go back to feature branch for final state
        else
            echo "CONFLICT detected during forward merge of $BRANCH. Aborting."
            git merge --abort
            git checkout "$BRANCH"
        fi
    fi
done

# Ensure we are back on the initial branch
git checkout "$INITIAL_BRANCH"

# RE-APPLY STASHED WORK
if [ "$STASHED_WORK" = true ]; then
    echo "Restoring stashed changes to $INITIAL_BRANCH..."
    git stash pop
fi

# Ensure initial branch is caught up if it's not main
if [ "$INITIAL_BRANCH" != "$MAIN_BRANCH" ]; then
    echo "Finalizing sync for $INITIAL_BRANCH..."
    git merge "$REMOTE/$MAIN_BRANCH" --no-edit || echo "Final sync merge failed. Manual intervention required."
fi

echo "Step 3: Workspace Cleanup, Documentation, & Push"

# Governance: Source version
VERSION=$(cat VERSION.md)
echo "Current Version: $VERSION"

# Batch Script Validation: Ensure execution scripts use the correct MAIN_BRANCH
for script in build.sh start.sh; do
    if [ -f "$script" ]; then
        echo "Validating $script..."
        # Portable sed: use -i '' for BSD/macOS compatibility if needed
        if sed --version >/dev/null 2>&1; then
            sed -i "s/\bmain\b/$MAIN_BRANCH/g" "$script"
        else
            sed -i "" "s/[[:<:]]main[[:>:]]/$MAIN_BRANCH/g" "$script"
        fi
    fi
done

# Targeted version synchronization in CHANGELOG.md (replaces only the first [UNRELEASED] or generic header)
if [ -f "CHANGELOG.md" ]; then
    # Portable sed: find first occurrence of a generic alpha version and update it
    if sed --version >/dev/null 2>&1; then
        sed -i "0,/1.0.0-alpha.[0-9]*/s/1.0.0-alpha.[0-9]*/$VERSION/" CHANGELOG.md
    else
        # macOS/BSD version (doesn't support 0,/pattern/ range easily, so we use a simpler approach for alpha)
        sed -i "" "1,s/1.0.0-alpha.[0-9]*/$VERSION/" CHANGELOG.md
    fi
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
if [ "$HUSTLE_PUSH_ENABLED" = "true" ]; then
    echo "Pushing changes to $REMOTE..."
    git push $REMOTE $(git rev-parse --abbrev-ref HEAD)
    git push $REMOTE --tags
else
    echo "Push disabled. Set HUSTLE_PUSH_ENABLED=true to enable automatic remote updates."
fi

echo "=== EXECUTIVE PROTOCOL COMPLETE ==="
