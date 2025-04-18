### plays tcp

A generic TCP server for sending commands. This can be used as a template for any TCP based application.

## TODO

- is it possible to support multiple queues?
- some kind of partition bug around 20 instances
- what happens when a consumer leaves?

## Bugs
 - [ ] when you publish to the queue, it adds it to the connection pool and assigns it partitions. This means that when consuming we wont get all of the messages because the producers partitions arent accounted for.

