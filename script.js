import { greeting } from "k6/x/ibmmq";

export default function () {
  console.log(greeting()) // Hello, World!
  console.log(greeting("Joe")) // Hello, Joe!
}
