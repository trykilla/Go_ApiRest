FROM midebian:try

WORKDIR /app

COPY broker /app/
COPY certificates /app/certificates

RUN chmod +x /app/broker 

CMD ["/app/broker"]