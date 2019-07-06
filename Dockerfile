FROM scratch

VOLUME ["/data", "/etc/rcon-hub"]
EXPOSE 2222

COPY rcon-hub /

ENTRYPOINT ["/rcon-hub"]
CMD ["--host-key-file", "/data/host_key.pem"]