# listbot

Automatically add a checklist to every Pull Request which is opened.
Set the build to failure until the checklist has been completed.

[![Deploy](https://www.herokucdn.com/deploy/button.png)](https://heroku.com/deploy)

## Installation

1. Generate a [personal access](https://github.com/settings/tokens) token from Github
 - It must have `repo` and `repo:status` for scopes
2. Run `torus link` in your checkout of the code
3. Store the token as `GITHUB_TOKEN` using `torus set GITHUB_TOKEN [token]`
4. Create a new machine for your deployment `torus machines create`
5. Click the Heroku deploy button and enter both `TORUS_TOKEN_ID` and `TORUS_TOKEN_SECRET`
6. Add a new file to your repository at `.github/listbot.md`
7. Add a new webhook at `https://github.com/[owner]/[repo]/settings/hooks/new`
 - Set the payload url to `[hostname]/webhook` where `[hostname]` is the deployed service
 - Select `Issue comment` and `Pull request` from "Let me select individual events"
 - Mark it as active and save
