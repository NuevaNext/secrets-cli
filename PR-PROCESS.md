# Complete Pull Request Process Flow

This document defines the exact step-by-step process for submitting and managing pull requests.

---

## Step 0: PR Prerequisites (BEFORE Submission)

**You MUST complete ALL items before creating the PR. No exceptions.**

### Code Quality
- [ ] Code has been updated and follows DRY principles
- [ ] No code duplication
- [ ] Code is well-commented

### Documentation  
- [ ] Documentation updated (no stale docs)
- [ ] Documentation follows DRY principles (no duplication)
- [ ] SKILL.md updated with new patterns/requirements (if applicable)
- [ ] No ephemeral bug-specific documentation files

### Testing
- [ ] Unit tests added/updated for new functionality
- [ ] All tests pass locally: `go test ./...`
- [ ] E2E tests added/updated (if applicable)
- [ ] Coverage maintained or improved

### Integration
- [ ] Parent Jira issue updated with PR link
- [ ] Related issues linked in PR description

**If ANY checkbox is unchecked, DO NOT submit the PR yet!**

---

## Step 1: Submit PR and Wait for Initial Feedback

1. Create PR using the template (all Step 0 items checked)
2. Wait for CI checks to start
3. Wait for reviewer comments
4. **DO NOT** declare "ready for merge" - you're just starting!

---

## Step 2: Address Failing Tests (ITERATE UNTIL ALL PASS)

**CRITICAL**: You must iterate on this step until ALL tests pass. No exceptions.

### Process:
1. Check test results:
   ```bash
   go test ./...
   ./tests/run-tests.sh
   gh pr checks PR_NUMBER
   ```

2. If tests fail:
   - ✅ **DO**: Write unit tests first to isolate the problem
   - ✅ **DO**: Think about what's CORRECT, not just making tests pass
   - ✅ **DO**: Debug with focused unit tests
   - ❌ **DON'T**: Only rely on e2e tests (slow, hard to debug)
   - ❌ **DON'T**: Make changes just to pass tests without understanding why

3. Fix the issues

4. **GO BACK TO STEP 2** - Test again until ALL pass

### When to proceed:
✅ **ONLY** when: `gh pr checks PR_NUMBER` shows all green ✓

**You will return to this step multiple times during the review process!**

---

## Step 3: Reply to EVERY Review Comment

When review comments arrive, you MUST reply to EVERY SINGLE ONE.

### For Each Comment:

1. **Read and understand** the comment

2. **Add a TODO** to track it:
   - Edit PR description
   - Add checkbox: `- [ ] Address comment #X - [brief description]`

3. **Reply to the comment** with ONE of:

   **Option A: Not Applicable**
   ```
   ℹ️ **Not applicable** - [Clear explanation why this doesn't apply]
   ```

   **Option B: Acknowledged** (most common at this stage)
   ```
   ✅ **Acknowledged** - Will address this. Added TODO to PR description.
   ```

   **Option C: Already Fixed** (only if you happened to fix it already)
   ```
   ✅ **Fixed** - Already implemented [what you did] in commit [hash]
   ```

### Important Notes:
- **DO NOT** say "Fixed" unless it's actually already done
- **DO** add a TODO for every "Acknowledged" comment  
- **DO** reply to ALL comments - no exceptions

### Verification:
```bash
# Check you replied to all comments
gh api repos/OWNER/REPO/pulls/PR_NUMBER/comments | \
  jq '.[] | select(.user.login == "YOUR_USERNAME")'
```

---

## Step 3.5: Work on Acknowledged Review Comments

**Now you actually do the work:**

1. Review all TODOs in PR description
2. For each TODO that's "Acknowledged":
   - Make the necessary code changes
   - Update documentation if needed
   - Add/update tests

3. Commit changes:
   ```bash
   git add -A
   git commit -m "address: [brief description] (re: comment #X)"
   git push origin branch-name
   ```

4. **GO BACK TO STEP 2** - Wait for tests to pass again

5. Once tests pass for that change:
   - Update the comment: Change "Acknowledged" to "Fixed"
   - Add what you did
   - Check off the TODO in PR description

### Example Flow:
```
Initial reply (Step 3):
✅ **Acknowledged** - Will address this. Added TODO to PR description.

After implementing (Step 3.5):
✅ **Fixed** - Implemented regex parsing as suggested (commit abc123)

PR Description:
- [x] Address comment #1 - Use regex parsing ← check this off
```

---

## Step 4: Verify All TODOs Are Complete

Before proceeding, verify:

```bash
# Check for unchecked TODOs
gh pr view PR_NUMBER --comments | grep -E "\- \[ \]"
```

**Output should be empty!** Every checkbox must be `[x]`.

If any unchecked TODOs remain: **GO BACK TO STEP 3.5**

---

## Step 5: Resolve All Review Threads

After ALL comments are addressed and marked "Fixed":

```bash
# Find unresolved threads
gh api graphql -f query='
query {
  repository(owner: "OWNER", name: "REPO") {
    pullRequest(number: PR_NUMBER) {
      reviewThreads(first: 20) {
        nodes {
          id
          isResolved
        }
      }
    }
  }
}'

# Resolve each thread
gh api graphql -f query='
mutation {
  resolveReviewThread(input: {threadId: "THREAD_ID"}) {
    thread {
      id
      isResolved
    }
  }
}'
```

Verify: Count of unresolved threads = **0**

---

## Step 6: Clean Up Intermediate Comments

Delete any progress update comments you posted during the review:

```bash
# List your comments
gh api repos/OWNER/REPO/issues/PR_NUMBER/comments | \
  jq -r '.[] | select(.user.login == "YOUR_USERNAME") | "ID: \(.id)\nBody: \(.body[0:60])\n---"'

# Delete intermediate status comments (keep final verification only)
gh api repos/OWNER/REPO/issues/comments/COMMENT_ID -X DELETE
```

**Keep only**: Final verification comment

---

## Step 7: Comprehensive Verification

Run ALL checks explicitly:

```bash
# 1. Tests
go test ./...
./tests/run-tests.sh  
gh pr checks PR_NUMBER  # Must be ✓ ALL GREEN

# 2. Comments
gh api repos/OWNER/REPO/pulls/PR_NUMBER/comments | \
  jq '[.[] | select(.body | contains("✅"))] | length'
# Count should equal number of review comments

# 3. TODOs
gh pr view PR_NUMBER --comments | grep -E "\- \[ \]" | wc -l
# Count must be 0

# 4. Threads
gh api graphql -f query='...' | jq '... | length'
# Count must be 0

# 5. Build
make build  # Must succeed
```

**If ANY check fails: Fix it and start over from relevant step!**

---

## Step 8: Post Final Verification Comment

**Only after Step 7 is 100% complete**, post:

```markdown
## ✅ MERGE READINESS - FINAL VERIFICATION

All items verified and complete:

### ✅ 1. All Tests Passing
- Unit tests: ✓ X/X passing
- E2E tests: ✓ X/X passing  
- CI checks: ✓ All green

### ✅ 2. All Review Comments Addressed
- Total comments: X
- Replied to: X/X
- Status: ✅ ALL REPLIED TO

### ✅ 3. All TODO Items Completed
- Total TODOs: X
- Completed: X/X
- Status: ✅ ALL COMPLETE

### ✅ 4. All Review Threads Resolved
- Unresolved count: 0
- Status: ✅ ALL RESOLVED

### ✅ 5. Code Quality
- Build: ✅ Success
- Linting: ✅ No errors

---

## 🎉 FINAL STATUS: READY FOR MERGE

This PR is ready to be merged. 🚀
```

---

## Step 9: Update Protection Rules (One-time Setup)

**Repository administrators should configure:**

```yaml
Branch Protection Rules for main/master:
  ✓ Require pull request reviews before merging
  ✓ Require status checks to pass before merging:
    - All CI checks
    - No unchecked TODOs in PR description
  ✓ Require conversation resolution before merging
  ✓ Require branches to be up to date before merging
```

---

## Quick Reference: When to Return to Step 2

You return to **Step 2** (wait for tests) after:
- Initial PR submission
- Making changes for acknowledged review comments (from Step 3.5)
- Fixing any failing tests
- Pushing any new commits

**The flow loops**: Step 2 → Step 3 → Step 3.5 → back to Step 2

This continues until all tests pass AND all review comments are addressed.

---

## Summary: The Complete Loop

```
Step 0: Prerequisites ✓
    ↓
Step 1: Submit PR
    ↓
Step 2: Tests pass? ──NO──> Fix tests ──┐
    ↓ YES                                │
Step 3: Reply to comments, add TODOs     │
    ↓                                    │
Step 3.5: Work on TODOs ─────────────────┘
    ↓ (loop until all TODOs done)
Step 4: Verify all TODOs checked
    ↓
Step 5: Resolve threads
    ↓
Step 6: Clean up comments
    ↓
Step 7: Comprehensive verification
    ↓
Step 8: Final verification comment
    ↓
READY FOR MERGE! 🚀
```

---

## Common Mistakes to Avoid

❌ **"Tests passing - ready for merge!"**
- Missing: Steps 3-8

❌ **Replying "Fixed" before actually fixing**
- Should be: "Acknowledged" → do work → then "Fixed"

❌ **Forgetting to add TODOs for acknowledged comments**
- Result: Comments get lost, PR can't track completion

❌ **Not going back to Step 2 after making changes**
- Result: Pushing broken code, tests fail later

❌ **Declaring ready before checking ALL items in Step 7**
- Result: Incomplete PR, wasted reviewer time

---

## Key Principles

1. **No shortcuts** - Follow every step, every time
2. **Tests must pass** - Iterate until green, no exceptions  
3. **Every comment gets a reply** - Shows respect, creates audit trail
4. **Track with TODOs** - Can't merge until all checked
5. **Loop back to tests** - After every change, verify tests pass
6. **Verify explicitly** - Run commands, don't assume
7. **Clean final state** - One final comment, all threads resolved
