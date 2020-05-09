# sailmaker-deployer-job
SailMaker Deployer Job queues up deployment using kubectl/kubernetes API and provides update for Release Custom Resource Definition

### Local Generate and Deploy

```
SAILMAKER_ENV=test go run main.go deploy --apps=sample/user/apps \
--releases=sample/releases/release-1.json \
--output=build --resources=sample/provider

```