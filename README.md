# urlShortener



steps:

build urlshortener docker image:
docker build -t urlshortener:1.0 . -f Dockerfile

pull mongodb docker image:
docker pull mongo

run mongodb on docker container:
docker run -p 27017:27017 -v $PWD/db:/data/db -d mongo

run urlshortener on docker container:
docker run -p 27017:27017 -p 8080:8080 -it urlshortener:1.0 <reCaptcha public key> <reCaptcha private key>
