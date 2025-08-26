import { Producer, Consumer } from "k6/x/ibmmq";

const QUEUE_NAME = "TEST";

const producer = new Producer({
  qmName: "QM1",
  hostname: "localhost",
  portNumber: 1414,
  channelName: "DEV.ADMIN.SVRCONN",
  username: "admin",
  password: "passw0rd"
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
  msgLimit: 1
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
