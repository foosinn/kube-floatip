FROM alpine
RUN apk add --no-cache ca-certificates
ADD floatip /floatip
ENTRYPOINT ["/floatip"]
