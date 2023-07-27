# illa-builder-backend System Design

## Overview

`illa-builder-backend` provides the backend APIs and services for the [illa-builder](https://github.com/illacloud/illa-builder). It is responsible for:

- Managing app state

- Executing actions

- Integrating with other tools

- Exposing REST APIs for the frontend

- Real-time updates via WebSockets

## Components

- REST API Layer 
  - Exposes CRUD APIs for apps, state, actions etc. 
  - Built with Go and gin-gonic framework

- WebSocket Server 
  - Handles real-time communication with clients
  - Broadcasts state updates to clients
  - Uses Gorilla WebSocket library

- Service Layer 
  - Application services for core use cases 
  - Common utilities and helpers 
  - Interfaces with repositories and other layers 

- Repository Layer 
  - Manages persistence using ORM 
  - Database abstraction and models
  
- Database 
  - PostgreSQL used as persistent store 
  - GORM as ORM

## Data Model 
- Apps
- States
  - Key-value state
  - Tree structured state
  - Set based state
- Resources
- Actions

This is the [script](../scripts/postgres-init.sh) that initializes the database tables.

## Backend Interfaces

- [illa-builder-backend HTTP API Documents](https://github.com/illacloud/illa-builder-backend-api-docs)

- [illa-builder-backend WebSocket Message Documents](https://github.com/illacloud/illa-builder-backend-websocket-docs)


## Scalability

Use PostgreSQL clustering for scalable DB

Horizontally scale WebSocket servers

Add caching to reduce load on database

## Monitoring

Log important metrics like request rates, errors

Track WebSocket connection counts

Set alerts for failures, high loads