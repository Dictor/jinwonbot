FROM golang:1.16-alpine

ADD . /jinwonbot
WORKDIR "/jinwonbot"
RUN apk add --no-cache --update bash make git
RUN ["make", "build"]
ENTRYPOINT ["/bin/sh", "-c"]
CMD ["/jinwonbot/jinwonbot -token $TOKEN -store $STORE"]
