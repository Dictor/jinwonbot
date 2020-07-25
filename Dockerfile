FROM golang:1.14-alpine

ADD ./ /jinwonbot
WORKDIR "/jinwonbot"
RUN ["go", "build"]
ENTRYPOINT ["/jinwonbot/jinwonbot"]
