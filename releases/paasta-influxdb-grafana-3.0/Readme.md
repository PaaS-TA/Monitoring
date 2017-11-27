Please Download follwoing binary and source packages before deploying InfluxDB-Grafana-Boshrelease in their respective folders.

## Git Submodule Update
git submodule update command download git dependency download

```
$ cd PaaSXpert-Monitor
$ git submodule init
$ git submodule update
```


## golan_1.6 binary download 

```
cd PaaSXpert-Monitor/release/InfluxDB-Grafana/src/

mkdir golang_1.6

cd golang_1.6 && wget https://storage.googleapis.com/golang/go1.6.1.linux-amd64.tar.gz
```

## grafana binary download 

```
cd PaaSXpert-Monitor/release/InfluxDB-Grafana/src/

mkdir grafana

cd grafana && wget https://grafanarel.s3.amazonaws.com/builds/grafana-3.1.0-1468321182.linux-x64.tar.gz
```

## bosh create release

```
cd PaaSXpert-Monitor/release/InfluxDB-Grafana/
bosh create release
bosh upload release
```

## bosh deploy (On AWS)
```
cd PaaSXpert-Monitor/deployments/bosh-init-deployments
bosh deployment influxdb-grafana-aws.yml
bosh -n deploy
```

## bosh deploy (On Bosh-lite)
```
cd PaaSXpert-Monitor/deployments/bosh-lite-deployments
bosh deployment influxdb-grafana.yml
bosh -n deploy
```
