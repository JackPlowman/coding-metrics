FROM golang:1.25.0-alpine

RUN mkdir /coding-metrics
COPY src /coding-metrics

RUN go build -o /coding-metrics/coding-metrics /coding-metrics && rm -rf /coding-metrics/src
CMD [ "/coding-metrics/coding-metrics" ]
