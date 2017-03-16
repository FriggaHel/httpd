FROM scratch
COPY script/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
ADD httpd /httpd
ENTRYPOINT ["/httpd"]
