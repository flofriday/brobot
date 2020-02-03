# Deploy
This is how I deploy the bot. For me this is the easiest way to get
running, however this may not work for you.

I deploy the bot on a RapberryPi 3 B running Raspbian Linux. The bot runs
inside a docker container and gets deployed automatically by the 
GitLab CI/CD.

## Steps
1. Install Docker
`curl -sSL https://get.docker.com | sh`
2. Install Gitlab Runner
```
curl -L https://packages.gitlab.com/install/repositories/runner/gitlab-runner/script.deb.sh | sudo bash
sudo apt-get install gitlab-runner
```
3. Register a GitLab Runner (Make sure that you use the shell executor and no tags)
