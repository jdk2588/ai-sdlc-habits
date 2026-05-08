# AI-Enabled SDLC

A reference implementation of the five AI-enabled SDLC habits for team collaboration.

## Who this is for

Experienced engineers who already use AI assistants and want to move from ad-hoc prompting to a disciplined, team-consistent approach.

## Quick start

Start RabbitMQ:

```
docker compose up
```

In a second terminal, start the orders service:

```
make orders
```

In a third terminal, start the fulfillment service:

```
make fulfillment
```

The orders service listens on `http://localhost:8080`. The fulfillment service connects to RabbitMQ and processes orders as they are placed.

## The full story

Read [TUTORIAL.md](TUTORIAL.md) to understand, how two developers built this system independently, where they diverged, and how certain habits brought them back into alignment to collaborate.
