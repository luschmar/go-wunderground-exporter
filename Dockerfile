FROM scratch
ADD go-wunderground-exporter /
CMD ["/go-wunderground-exporter"]