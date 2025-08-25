import { Producer } from "k6/x/ibmmq";

const producer = new Producer({
  qmName: "QM1",
  hostname: "localhost",
  portNumber: "1414",
  channelName: "DEV.ADMIN.SVRCONN",
  username: "admin",
  password: "passw0rd"
})

export default function () {
  producer.send("TEST", "test")
  producer.commit()
}

export function teardown(data) {
  producer.close();
}
