FROM golang:1.14-alpine

ADD ./ /jinwonbot
WORKDIR "/jinwonbot"
RUN ["go", "build"]
ENTRYPOINT ["/bin/sh", "-c"]
CMD ["/jinwonbot/jinwonbot -t $DISCORD_TOKEN"]
