# Nexus

Nexus is a cloud-native social media platform written in Go.
It is a long-term project designed to apply and deepen my understanding of
backend architecture, distributed systems, and frontend development.

This repo serves as both a learning project and a portfolio project
demonstrating backend-focused system design.

The project emphasizes real-world engineering practices and principles such as
**Domain-Driven Design**, **CQRS**, **Clean Architecture**, and **containerized deployment**.

The backend follows a modular monolith architecture, where domains are isolated
through clear boundaries but deployed as a single service.

The Nexus API exposes a REST interface for frontend communication, while internal
module communication is planned via gRPC and an event-driven architecture.

> ⚠️ Nexus is under active development. Many features described below are planned
> but not yet implemented.

## Getting Started

> Coming soon. The project will be runnable locally via Docker and Kubernetes.

## Tech Stack

### Backend

- Go (REST API)
- Chi router
- PostgreSQL
- Redis
- OAuth2.0 & OIDC authentication
- Docker & Kubernetes

### Frontend

- React
- Tailwind CSS
- shadcn/ui
- TanStack Router
- TanStack Query

### Infrastructure & Observability

- Kubernetes (local-first deployment)
- Prometheus (metrics & monitoring)
- AWS S3 (image storage)

## Architecture

Nexus is designed as a **modular monolith**, following:

- Domain-Driven Design (DDD)
- CQRS
- Clean Architecture

Inter-module communication is planned using:

- gRPC
- Event-driven messaging (future improvement)

## Project Roadmap

### MVP

- User accounts
- Authentication (OAuth2.0 + email)
- Text-based posts
- Nested comments
- Upvotes & downvotes
- Notifications

### Final Version

- Rich-content posts
- Groups & Pages
- RBAC & PBAC
- Personalized feeds
- Reactions & history
- Moderation features

## Current

The project is in early development. Core domain modeling and architectural
foundations are prioritized before feature implementatin.

## Notes

This project is built from scratch without tutorials and is fully self-designed.
It is primarily a learning project and is not intended for production use,
while still aiming for a level of scalability suitable for small to mid-sized deployments
