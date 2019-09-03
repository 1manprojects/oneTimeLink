FROM alpine:latest

ENV URL ""
ENV GDPR ""
ENV MAIL ""
ENV LOGO "../static/logo.svg"
ENV PASSWORD "Pa$$word"

RUN mkdir /app
RUN mkdir /app/uploads
RUN mkdir /app/static
RUN mkdir /app/html

ADD main /app/main
ADD html/*.html /app/html/
ADD static/onetime.css /app/static/
ADD static/*.svg /app/static/

WORKDIR /app

EXPOSE 8080

ENTRYPOINT exec /app/main -p $PASSWORD -u $URL -g $GDPR -m $MAIL -l $LOGO