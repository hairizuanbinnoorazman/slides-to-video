# Tests

Contains integration test of the whole project. This is to test the end to end workflow of requrests going through the various applications in the project as well as to test that it when deployed on various other platforms

The tests will need to target and run tests across the variety of platforms below:

- "Local" setup with docker-compose. This would be priority 0 - meant for use for local development
  - Storage: Minio
  - DB: MySQL
  - Queue: Nats
- Google Serverless Platform. Cloud Run
  - Storage: GCS
  - DB: Cloud Datastore
  - Queue: Google Pubsub
- Kubernetes
  - Storage: Minio, GCS
  - DB: MySQL, Postgreql, Cloud SQL, Datastore, Cassandra
  - Queue: Nats, Google Pubsub, RabbitMQ, Kafka

Types of test to be run:

- Basic functional tests (Across different platforms)
- Upgrade tests (Once a v1 is decided)
- Performance test

# Why python?

The tests written here are meant to target all the various platforms that this solution is meant to target. It would be good if the tests are written using golang; however, writing golang code require more rigorous thought put behind it to code it out. To reduce amount of effort to code this code, python with pytest framework is used instead.
