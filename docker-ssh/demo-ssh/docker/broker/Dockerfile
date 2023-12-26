FROM midebian

WORKDIR /app

COPY broker /app/
COPY certificates /app/certificates
COPY users.json /app/cmd/APIRest/

COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh


ENTRYPOINT ["/bin/bash", "/app/entrypoint.sh"]

