FROM golang:1.24-alpine AS apiBuilder
WORKDIR /opt
RUN apk update && apk add --no-cache make
COPY . /opt/
RUN go mod download
ARG GIT_TAG=?
RUN make build GIT_TAG=${GIT_TAG}

FROM alpine
ARG ENVIRONMENT=production
WORKDIR /opt
RUN apk update && apk add --no-cache curl ca-certificates
COPY --from=apiBuilder /opt/roflbeacon2 /opt/roflbeacon2
EXPOSE 8080
CMD [ "./roflbeacon2" ]
