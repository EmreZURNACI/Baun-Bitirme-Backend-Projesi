
# Intranet Q&A Platform Featuring AI-Powered Question Answering

## Project Overview

Many organizations restrict internet access on employee computers due to security policies. This creates challenges when users face technical issues and need quick solutions.

This project aims to develop an internal (intranet) Q&A platform similar to StackOverflow, enabling employees to ask and answer technical questions within the organization without needing internet access.

Additionally, the platform integrates a Python-based natural language processing (NLP) AI model trained on the accumulated Q&A data to provide automated responses and enhance user experience.



## Technologies

- Frontend: ReactJS

- Backend: Golang (Fiber framework)

- Database: PostgreSQL

- Containerization: Docker & Docker Compose

- AI: Python NLP-based model

- Monitoring: Prometheus for metrics collection and monitoring
## Features

- Internal Q&A platform working fully offline on intranet

- Users can post technical questions and receive answers

- Interaction features: views, likes, dislikes on questions and answers

- AI-powered chatbot for answering frequently asked questions

- Easy deployment using Docker

- Performance monitoring with Prometheus integration


## Installation

To start the project, run the following command:

```bash
    docker-compose -f ./deploy/docker/docker-compose.yml up -d
```
    
If you make any changes and want to rebuild the images before starting the containers, use:

```bash
    docker-compose -f ./deploy/docker/docker-compose.yml up --build -d
```
## Team

Thanks to https://github.com/ozguryurt for valuable contributions on frontend development.

