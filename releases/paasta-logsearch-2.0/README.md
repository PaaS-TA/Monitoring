Please Download follwoing source packages before deploying logsearch in their respective folders.

## Git Submodule Update
git submodule update command download git dependency download

```
$ cd PaaSXpert-Monitor
$ git submodule init
$ git submodule update
```

## bosh create release

```
cd PaaSXpert-Monitor/release/logsearch/
bosh create release
bosh upload release
```

## bosh deploy (On AWS)
```
cd PaaSXpert-Monitor/deployments/bosh-init-deployments
bosh deployment logsearch-aws.yml.yml
bosh -n deploy
```

## bosh deploy (On Bosh-lite)
```
cd PaaSXpert-Monitor/deployments/bosh-lite-deployments
bosh deployment logsearch.yml
bosh -n deploy
```
