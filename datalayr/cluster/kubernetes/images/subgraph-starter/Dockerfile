FROM node:16.15-bullseye-slim as build

ENV NODE_ENV production

RUN apt-get update && apt-get install -y python build-essential git curl

COPY ./wait-for.sh /wait-for

COPY ./data /graph-data-template

# Install dependencies; include dev dependencies
RUN sleep 1
