# Contributing to raccoon

The aim of this document is to offer guidances for new contributors on how to contribute to 
raccoon. 

## Join the community

For now, we only have a github repo where discussions are opened and welcomed.
In a near future, we could come up with more communication vectors.

## Did you find a bug?
1. Check if the bug hasn't yet been reported [here](https://github.com/BackMarket/raccoon/issues).
2. If no bug has been reported, please open an issue by using the issue template.

## Did you open a PR to fix a bug?
1. Ensure the PR description describes the problem and the solution.
2. Follow the commit messages [template](###commit-messages) and don't forget to 
add `Closes/Fixes #issue_number` to the commit message to link your patch to the corresponding
issue.

## Do you intend to add a new feature?
1. Check if a feature request isn't already opened.
2. If not, suggest your change via an issue, [here](https://github.com/BackMarket/raccoon/issues).
3. Once issue participants are aligned, you can implement it.

## Write a good patch

### Follow our code style

TO LINK ONCE DEFINED.

### Write separate changes

We prefer a commit per fix with its own commit message to know what it corrects so
that it can be selectively applied by a maintainer.
We refuse large PR that makes multiple changes that have nothing to do with each other.

### Rebase, not merge and... autosquash

Crappy, complex, unreadable history is forbidden! 
1. We refuse merge commit in history. Instead use [rebase](https://git-scm.com/book/en/v2/Git-Branching-Rebasing).
2. During a review, when you need to fix something, use 
[fixup commit](https://git-scm.com/docs/git-commit#Documentation/git-commit.txt---fixupamendrewordltcommitgt) 
3. When the patch is validated, you'll be asked to `autosquash` your fixup commits.

### Test cases

TO DEFINE once test cases will be implemented.

## Submit patch

### How to get your patch into raccoon's sources

Your patch will be discussed, as a patch submitter you are the owner of it until it 
gets merged in the main branch.
It means that you have to answer questions about your change, fix nits/flaws that have been 
pointed out.
If no activity or a lack of replies is reported on a PR, we will simply drop/delete it.

### Making quality changes

Make the patch against the most recent version possible.

### Commit messages

Here is a short guide on how to write a commit message in raccoon.

```
[area]: [short line describing the intent of the patch]
-- empty line ---
[full description describing the aim of the patch, why this patch
is made]
-- empty line ---
[Closes/Fixes #issue_number]
[Reported-by: Name of the reporter]
```

The `[area]` can be any part in the system, like `doc`, `strategy`, `internal`. We don't
use __conventional commits__ pattern.


Thanks.
