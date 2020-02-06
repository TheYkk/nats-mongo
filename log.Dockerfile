# Install npm packages
FROM node:12-alpine as builder
WORKDIR /usr/src/app
COPY package.json .
RUN yarn
RUN yarn cache clean

# Push js files
FROM node:12-alpine
WORKDIR /usr/src/app
COPY --from=builder /usr/src/app/ /usr/src/app/
COPY . .
CMD node log.js
