# LLM Observer — Plain-English Guide

## What is LLM Observer?

LLM Observer is a dashboard that shows you exactly what your AI agents are doing — how fast they're running, how much they cost, where they're slow, and when something breaks.

Think of it like a fitness tracker, but for AI models instead of your body.

## Why does this exist?

When you run AI models (like ChatGPT, Claude, or local models through Ollama), you have no visibility into:
- How many tokens (words) each request uses
- How long each step takes
- Which agent in a chain is the bottleneck
- How much money you're spending on API calls

LLM Observer captures all of this in real time and shows it on a live dashboard.

## How does it work?

1. **Your AI app** sends small tracking signals (called "spans") to the Observer
2. **A Go server** collects and batches these signals efficiently
3. **TimescaleDB** (a database built for time-series data) stores everything
4. **A React dashboard** shows charts, traces, and cost breakdowns in real time

## What can you see on the dashboard?

- Token usage per request (input tokens, output tokens)
- Latency heatmaps showing fast vs slow requests
- Cost tracking per model and per agent
- Trace visualization — follow a request through every step of a multi-agent chain
- Anomaly detection — automatic alerts when something looks wrong

## What's it built with?

- **Go** — The collection server (fast, low memory)
- **Python SDK** — Drop into your AI app to start tracking
- **TimescaleDB** — Time-series database (PostgreSQL extension)
- **Redpanda** — Message bus for streaming events
- **React** — The dashboard UI

## Current status

Core pipeline works. Frontend dashboard, anomaly detection, and SDK are functional. Actively maintained.
