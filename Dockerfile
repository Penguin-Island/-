FROM golang:latest
WORKDIR /build
COPY . .
RUN cd be/cmd/ohatori/ && CGO_ENABLED=0 GOOS=linux go build .

FROM node:latest
WORKDIR /build
COPY . .
RUN npm ci
RUN npm run prod

FROM alpine:latest
COPY --from=0 /build/be/cmd/ohatori/ohatori /ohatori
COPY --from=1 /build/dist /dist/
RUN apk --no-cache add curl tzdata
ENTRYPOINT ["/ohatori"]
