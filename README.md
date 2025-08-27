# xk6-ibmmq

**k6 extension to interact with IBM MQ**

> [!WARNING]  
> This extension is a proof of concept, isn't supported by the k6 team, and may break in the future. USE AT YOUR OWN RISK!

This k6 extension provide users with the ability to extend k6 functionalities to IBM MQ by wrapping over [
mq-golang-jms20](https://github.com/ibm-messaging/mq-golang-jms20).

```javascript file=script.js
import { Producer, Consumer } from "k6/x/ibmmq";

const QUEUE_NAME = "TEST";

const producer = new Producer({
  qmName: "QM1",
  hostname: "localhost",
  portNumber: 1414,
  channelName: "DEV.ADMIN.SVRCONN",
  username: "admin",
  password: "passw0rd",
  tlsEnabled: true,
  tlsCipherSpec: "TLS_RSA_WITH_AES_128_CBC_SHA256",
  keyRepoPath: "./tls/client/mutual-tls",
  certLabel: "client"
});

const consumer = new Consumer({
  qmName: "QM1",
  hostname: "localhost",
  portNumber: 1414,
  channelName: "DEV.ADMIN.SVRCONN",
  queueName: QUEUE_NAME,
  username: "admin",
  password: "passw0rd",
  timeout: 60,
  msgLimit: 1,
  tlsEnabled: true,
  tlsCipherSpec: "TLS_RSA_WITH_AES_128_CBC_SHA256",
  keyRepoPath: "./tls/client/mutual-tls",
  certLabel: "client"
});

export default function () {
  producer.send(QUEUE_NAME, "test", {"ibmmq": "loadtest"});
  producer.commit();

  let messages = consumer.consume();
  consumer.commit();

  console.log(messages);
}

export function teardown(data) {
  producer.close();
  consumer.close();
}
```

## Quick Start

1. Refer to the [Getting Started](https://github.com/ibm-messaging/mq-golang-jms20?tab=readme-ov-file#getting-started) section in `mq-golang-jms20` repository for in-depth setup guide.

2. For a quick start, run the below commands (inspired by Dockerfile in [openshift-app-sample](https://github.com/ibm-messaging/mq-golang-jms20/blob/main/openshift-app-sample/Dockerfile)):

    ```bash
    # Require elevation to write to /opt
    sudo su

    mkdir -p /opt/mqm && chmod a+rx /opt/mqm

    EXPORT genmqpkg_incnls=1 genmqpkg_incsdk=1 genmqpkg_inctls=1

    # substitute 9.4.3.0 with another version if desired
    cd /opt/mqm \
        && curl -LO "https://public.dhe.ibm.com/ibmdl/export/pub/software/websphere/messaging/mqdev/redist/9.4.3.0-IBM-MQC-Redist-LinuxX64.tar.gz" \
        && tar -zxf ./*.tar.gz \
        && rm -f ./*.tar.gz \
        && bin/genmqpkg.sh -b /opt/mqm
    ```

3. Building the extension

    ```bash
    export CGO_ENABLED=1
    export MQ_INSTALLATION_PATH="/opt/mqm"
    export CGO_CFLAGS="-I/opt/mqm/inc"
    export CGO_LDFLAGS_ALLOW="-Wl,-rpath.*"
    export CGO_LDFLAGS="-L/opt/mqm/lib64 -Wl,-rpath,/opt/mqm/lib64"

    xk6 build \
        --with github.com/RowenTey/xk6-ibmmq=. \
        --cgo=1
    # OR
    make build
    ```

4. Running the sample script

    ```bash
    ./k6 run script.js
    ```

## Download

Building a custom k6 binary with the `xk6-ibmmq` extension is necessary for its use. You can download pre-built k6 binaries from the [Releases page](https://github.com/RowenTey/xk6-ibmmq/releases/).

## Build

Use the [xk6](https://github.com/grafana/xk6) tool to build a custom k6 binary with the `xk6-ibmmq` extension. Refer to the [xk6 documentation](https://github.com/grafana/xk6) for more information.

## Contribute

If you wish to contribute to this project, please start by reading the [Contributing Guidelines](CONTRIBUTING.md).

## Generating certs for TLS

The `gsk8capicmd` command is included in the IBM MQ redistributable client installation in the `./gskit8/bin` directory.

```bash
# Set gskit8 binary in path
export PATH=/opt/mqm/gskit8/bin:$PATH
export LD_LIBRARY_PATH=/opt/mqm/gskit8/lib64:$LD_LIBRARY_PATH

# Create client keystore
gsk8capicmd_64 -keydb -create -db mutual-tls.kdb -pw password -type kdb -expire 0 -stash

# Import server CA
gsk8capicmd_64 -cert -add -db mutual-tls.kdb -file ca.crt -label ServerCertRootCA -stashed -type kdb -format ascii

# Generate keypair
gsk8capicmd_64 -cert -create -db mutual-tls.kdb -pw password -label client -size 2048 -sig_alg SHA256WithRSA -dn "CN=client" -expire 365

# Extract cert
gsk8capicmd_64 -cert -extract -db mutual-tls.kdb -pw password -label client -target client.crt -format ascii
```
