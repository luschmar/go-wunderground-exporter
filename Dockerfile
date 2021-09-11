FROM scratch
COPY go-wunderground-exporter app/
CMD ["/app/go-wunderground-exporter"]
