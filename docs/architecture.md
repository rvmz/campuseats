# CampusEats architecture

Session 2 fills this in. Later sessions update it when a failure mode gets a mechanism. Do not leave a heading empty.

## Context

Who is outside the system, and what they are trying to do?

## Containers

Processes and data stores. For each arrow, name the protocol (HTTP, gRPC, SQL, Kafka) and whether the caller waits.

## Place order

Numbered sequence from the student tap to the kitchen seeing the ticket. Mark the step that returns HTTP to the student.

## SLO

One latency SLO and one "kitchen sees the order" SLO. Each needs an endpoint or event and a number.

## Failures

Three failures. For each: what broke, what the student sees, what is still true in Postgres, which session handles it.

1.
2.
3.

## Non-goals

What this course will not build.

## API

| Caller | Call | Sync? |
|---|---|---|
| | | |
