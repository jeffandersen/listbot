[![Deploy](https://www.herokucdn.com/deploy/button.png)](https://heroku.com/deploy)

# listbot

Automatically add a checklist (sourced from `.github/listbot.md`) to every Pull Request which is opened. As the checklist is updated the bot will change the build status depending on whether the checklist is completed.

![](./preview.gif)

## Installation

1. Generate a [personal access](https://github.com/settings/tokens) token from Github
 - It must have `repo` and `repo:status` for scopes
 - **Note:** Lists will appear as comments by the user who generates this token.
   - You may want to create a "bot user" and use the token from that account.
2. Signup for [Torus](https://torus.sh) if you haven't already
 - Run `torus link` in your checkout of the code
 - Create an environment for Heroku `torus envs create heroku`
 -  Create a machine role for your deployment `torus machines roles create heroku`
 - Give your machine role read only access `torus allow rl /$org/$project/heroku/default/*/*/* heroku`
 - Create a new machine for your deployment with `torus machines create heroku-instance -r heroku`
 -  Store the token as `GITHUB_TOKEN` using `torus set -e heroku GITHUB_TOKEN [token]`
3. Click the Heroku deploy button
4. Add values for `TORUS_TOKEN_ID`, `TORUS_TOKEN_SECRET`, `TORUS_PROJECT`, `TORUS_ORG` from step 2.
5. Add a new file to your repository at `.github/listbot.md`
6. Add a new webhook at `https://github.com/[owner]/[repo]/settings/hooks/new`
 - Set the payload url to `[hostname]/webhook` where `[hostname]` is the deployed service
 - Select `Issue comment` and `Pull request` from "Let me select individual events"
 - Mark it as active and save
