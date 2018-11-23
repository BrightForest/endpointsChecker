FROM golang:alpine
RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN go get gopkg.in/yaml.v2
RUN go build -o checker .
EXPOSE 8080
CMD ["/app/checker"]