FROM golang:1.11.2
RUN mkdir /app
ADD checker /app/
WORKDIR /app
RUN go get gopkg.in/yaml.v2
RUN go build -o checker .
EXPOSE 8080
CMD ["/app/checker"]
